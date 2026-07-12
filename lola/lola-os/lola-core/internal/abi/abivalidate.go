// Package abivalidate implements LOLA OS's pre-flight ABI validation:
// before a contract call is simulated or broadcast, it checks that the
// requested method exists in the ABI and that the supplied arguments are
// type-compatible with what the ABI declares. This catches mistakes — a
// wrong method name, or a string where a uint256 is expected — before they
// cost gas or produce a revert.
package abivalidate

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// ABIMismatchError describes exactly which argument failed validation and
// why, so SDKs can surface an actionable TypeError to the calling agent.
type ABIMismatchError struct {
	Method     string
	ArgIndex   int    // -1 if the problem is the method itself or arg count
	Expected   string
	GotValue   interface{}
	Reason     string
}

func (e *ABIMismatchError) Error() string {
	if e.ArgIndex < 0 {
		return fmt.Sprintf("ABI mismatch for method %q: %s", e.Method, e.Reason)
	}
	return fmt.Sprintf("ABI mismatch for method %q, argument %d: expected %s, got %v (%s)",
		e.Method, e.ArgIndex, e.Expected, e.GotValue, e.Reason)
}

// Validate checks that method exists in abiJSON and that args are
// compatible with the method's declared input types. Returns nil if valid,
// or an *ABIMismatchError describing the first problem found.
func Validate(abiJSON, method string, args []interface{}) error {
	parsed, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return fmt.Errorf("abivalidate: parsing ABI: %w", err)
	}

	m, ok := parsed.Methods[method]
	if !ok {
		available := make([]string, 0, len(parsed.Methods))
		for name := range parsed.Methods {
			available = append(available, name)
		}
		return &ABIMismatchError{
			Method: method, ArgIndex: -1,
			Reason: fmt.Sprintf("method not found in ABI; available methods: %s", strings.Join(available, ", ")),
		}
	}

	if len(args) != len(m.Inputs) {
		return &ABIMismatchError{
			Method: method, ArgIndex: -1,
			Reason: fmt.Sprintf("expected %d argument(s), got %d", len(m.Inputs), len(args)),
		}
	}

	for i, input := range m.Inputs {
		if err := checkType(input.Type, args[i]); err != nil {
			return &ABIMismatchError{
				Method: method, ArgIndex: i,
				Expected: input.Type.String(),
				GotValue: args[i],
				Reason:   err.Error(),
			}
		}
	}

	// Finally, coerce the arguments into the concrete Go types the ABI
	// encoder requires and attempt the real ABI packing — this catches any
	// subtler mismatches (e.g. fixed-size array length, struct/tuple field
	// counts) that the per-argument type check above doesn't fully cover.
	coerced, err := CoerceArgs(m.Inputs, args)
	if err != nil {
		return &ABIMismatchError{
			Method: method, ArgIndex: -1,
			Reason: fmt.Sprintf("arguments failed ABI encoding: %v", err),
		}
	}
	if _, err := parsed.Pack(method, coerced...); err != nil {
		return &ABIMismatchError{
			Method: method, ArgIndex: -1,
			Reason: fmt.Sprintf("arguments failed ABI encoding: %v", err),
		}
	}

	return nil
}

// CoerceArgs converts JSON-decoded argument values (hex strings, numbers,
// slices) into the concrete Go types go-ethereum's ABI encoder requires —
// common.Address for addresses, *big.Int or a fixed-width integer for
// numeric types, [N]byte for fixed bytes, and correctly-typed slices/arrays
// for collections. SDK callers send arguments as plain JSON, so this bridge
// is what lets an agent pass "0xabc…" and 100 and have them packed exactly
// as the contract expects. The returned slice is safe to hand to abi.Pack.
func CoerceArgs(inputs abi.Arguments, args []interface{}) ([]interface{}, error) {
	if len(args) != len(inputs) {
		return nil, fmt.Errorf("expected %d argument(s), got %d", len(inputs), len(args))
	}
	out := make([]interface{}, len(args))
	for i, in := range inputs {
		v, err := coerceValue(in.Type, args[i])
		if err != nil {
			return nil, fmt.Errorf("argument %d (%s): %w", i, in.Type.String(), err)
		}
		out[i] = v
	}
	return out, nil
}

func coerceValue(t abi.Type, v interface{}) (interface{}, error) {
	switch t.T {
	case abi.AddressTy:
		switch val := v.(type) {
		case common.Address:
			return val, nil
		case string:
			if !common.IsHexAddress(val) {
				return nil, fmt.Errorf("%q is not a valid hex address", val)
			}
			return common.HexToAddress(val), nil
		}
		return nil, fmt.Errorf("expected an address, got %T", v)
	case abi.UintTy, abi.IntTy:
		bi, err := toBigInt(v)
		if err != nil {
			return nil, err
		}
		return fitInteger(t, bi), nil
	case abi.BoolTy:
		switch val := v.(type) {
		case bool:
			return val, nil
		case string:
			return strings.EqualFold(val, "true"), nil
		}
		return nil, fmt.Errorf("expected a bool, got %T", v)
	case abi.StringTy:
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("expected a string, got %T", v)
		}
		return s, nil
	case abi.BytesTy:
		return toBytes(v)
	case abi.FixedBytesTy:
		b, err := toBytes(v)
		if err != nil {
			return nil, err
		}
		arr := reflect.New(t.GetType()).Elem()
		if len(b) > arr.Len() {
			return nil, fmt.Errorf("bytes%d: got %d bytes, too long", arr.Len(), len(b))
		}
		for i := 0; i < len(b); i++ {
			arr.Index(i).Set(reflect.ValueOf(b[i]))
		}
		return arr.Interface(), nil
	case abi.SliceTy, abi.ArrayTy:
		return coerceSlice(t, v)
	case abi.TupleTy:
		// Tuple/struct arguments are passed through; abi.Pack validates
		// their field shape. Building arbitrary structs from JSON maps is
		// intentionally out of scope here.
		return v, nil
	default:
		return v, nil
	}
}

func coerceSlice(t abi.Type, v interface{}) (interface{}, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, fmt.Errorf("expected a slice/array for %s, got %T", t.String(), v)
	}
	target := reflect.New(t.GetType()).Elem()
	if t.T == abi.SliceTy {
		target.Set(reflect.MakeSlice(t.GetType(), rv.Len(), rv.Len()))
	} else if rv.Len() != target.Len() {
		return nil, fmt.Errorf("%s expects %d element(s), got %d", t.String(), target.Len(), rv.Len())
	}
	for i := 0; i < rv.Len(); i++ {
		cv, err := coerceValue(*t.Elem, rv.Index(i).Interface())
		if err != nil {
			return nil, fmt.Errorf("element %d: %w", i, err)
		}
		target.Index(i).Set(reflect.ValueOf(cv))
	}
	return target.Interface(), nil
}

// fitInteger returns bi as the exact Go integer type the ABI encoder wants:
// *big.Int for sizes above 64 bits, and the matching fixed-width type
// (uint8…uint64 / int8…int64) otherwise.
func fitInteger(t abi.Type, bi *big.Int) interface{} {
	if t.Size > 64 {
		return bi
	}
	if t.T == abi.UintTy {
		u := bi.Uint64()
		switch {
		case t.Size <= 8:
			return uint8(u)
		case t.Size <= 16:
			return uint16(u)
		case t.Size <= 32:
			return uint32(u)
		default:
			return u
		}
	}
	n := bi.Int64()
	switch {
	case t.Size <= 8:
		return int8(n)
	case t.Size <= 16:
		return int16(n)
	case t.Size <= 32:
		return int32(n)
	default:
		return n
	}
}

func toBigInt(v interface{}) (*big.Int, error) {
	switch val := v.(type) {
	case *big.Int:
		return val, nil
	case int:
		return big.NewInt(int64(val)), nil
	case int64:
		return big.NewInt(val), nil
	case uint64:
		return new(big.Int).SetUint64(val), nil
	case uint:
		return new(big.Int).SetUint64(uint64(val)), nil
	case float64:
		// JSON numbers decode as float64; require an integral value.
		if val != math.Trunc(val) {
			return nil, fmt.Errorf("expected an integer, got fractional %v", val)
		}
		bi, _ := big.NewFloat(val).Int(nil)
		return bi, nil
	case string:
		s := strings.TrimSpace(val)
		base := 10
		if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
			base, s = 16, s[2:]
		}
		bi, ok := new(big.Int).SetString(s, base)
		if !ok {
			return nil, fmt.Errorf("%q is not a valid integer", val)
		}
		return bi, nil
	default:
		return nil, fmt.Errorf("expected a numeric value (int/uint/*big.Int/string), got %T", v)
	}
}

func toBytes(v interface{}) ([]byte, error) {
	switch val := v.(type) {
	case []byte:
		return val, nil
	case string:
		b, err := hex.DecodeString(strings.TrimPrefix(val, "0x"))
		if err != nil {
			return nil, fmt.Errorf("invalid hex bytes: %w", err)
		}
		return b, nil
	default:
		return nil, fmt.Errorf("expected bytes or a hex string, got %T", v)
	}
}

// checkType performs a best-effort compatibility check between an ABI
// type and a Go value, before the (authoritative but less specific) call
// to abi.Pack. This exists mainly to produce a clear, per-argument error
// message rather than relying solely on go-ethereum's pack error text.
func checkType(t abi.Type, v interface{}) error {
	switch t.T {
	case abi.IntTy, abi.UintTy:
		switch v.(type) {
		case *big.Int, int, int64, uint64, uint, float64:
			return nil
		default:
			return fmt.Errorf("expected a numeric value (int/uint/*big.Int), got %T", v)
		}
	case abi.BoolTy:
		if _, ok := v.(bool); !ok {
			return fmt.Errorf("expected a bool, got %T", v)
		}
		return nil
	case abi.StringTy:
		if _, ok := v.(string); !ok {
			return fmt.Errorf("expected a string, got %T", v)
		}
		return nil
	case abi.AddressTy:
		switch val := v.(type) {
		case common.Address:
			return nil
		case string:
			if !common.IsHexAddress(val) {
				return fmt.Errorf("string %q is not a valid hex address", val)
			}
			return nil
		default:
			return fmt.Errorf("expected an address (hex string or common.Address), got %T", v)
		}
	case abi.BytesTy, abi.FixedBytesTy:
		switch v.(type) {
		case []byte, string:
			return nil
		default:
			return fmt.Errorf("expected bytes or a hex string, got %T", v)
		}
	case abi.SliceTy, abi.ArrayTy:
		// Arrays/slices: defer detailed element checking to abi.Pack, but
		// at least confirm we were given something slice-like.
		switch v.(type) {
		case []interface{}, []*big.Int, []string, []common.Address, []bool:
			return nil
		default:
			return fmt.Errorf("expected a slice/array for type %s, got %T", t.String(), v)
		}
	case abi.TupleTy:
		// Struct/tuple args: defer to abi.Pack for field-level validation.
		return nil
	default:
		// Unknown/uncommon types: don't block on something we can't
		// confidently validate; let abi.Pack be the final authority.
		return nil
	}
}
