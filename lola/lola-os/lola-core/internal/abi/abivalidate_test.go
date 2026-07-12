package abivalidate

import (
	"strings"
	"testing"
)

const testABI = `[
	{"constant":false,"inputs":[{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"type":"function"},
	{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"type":"function"},
	{"constant":false,"inputs":[{"name":"flag","type":"bool"}],"name":"setActive","outputs":[],"type":"function"}
]`

func TestValidate_CorrectArgsPass(t *testing.T) {
	err := Validate(testABI, "transfer", []interface{}{"0x1111111111111111111111111111111111111111", 100})
	if err != nil {
		t.Fatalf("expected valid call to pass, got %v", err)
	}
}

func TestValidate_UnknownMethodFails(t *testing.T) {
	err := Validate(testABI, "trasnfer", []interface{}{"0x1111111111111111111111111111111111111111", 100})
	if err == nil {
		t.Fatalf("expected unknown method (typo) to fail validation")
	}
	if !strings.Contains(err.Error(), "method not found") {
		t.Fatalf("expected a clear 'method not found' message, got: %v", err)
	}
}

func TestValidate_WrongArgCountFails(t *testing.T) {
	err := Validate(testABI, "transfer", []interface{}{"0x1111111111111111111111111111111111111111"})
	if err == nil {
		t.Fatalf("expected wrong argument count to fail validation")
	}
}

func TestValidate_WrongArgTypeFails(t *testing.T) {
	// "amount" should be numeric, not a string — a common mistake when a
	// caller passes a stringified number.
	err := Validate(testABI, "transfer", []interface{}{"0x1111111111111111111111111111111111111111", "one hundred"})
	if err == nil {
		t.Fatalf("expected non-numeric amount to fail validation")
	}
	var mismatch *ABIMismatchError
	if !asABIMismatch(err, &mismatch) {
		t.Fatalf("expected an *ABIMismatchError, got %T: %v", err, err)
	}
	if mismatch.ArgIndex != 1 {
		t.Fatalf("expected mismatch on argument index 1, got %d", mismatch.ArgIndex)
	}
}

func TestValidate_InvalidAddressFails(t *testing.T) {
	err := Validate(testABI, "transfer", []interface{}{"not-an-address", 100})
	if err == nil {
		t.Fatalf("expected invalid address string to fail validation")
	}
}

func TestValidate_BoolArgPasses(t *testing.T) {
	if err := Validate(testABI, "setActive", []interface{}{true}); err != nil {
		t.Fatalf("expected valid bool arg to pass, got %v", err)
	}
}

func TestValidate_NoArgMethodPasses(t *testing.T) {
	if err := Validate(testABI, "totalSupply", []interface{}{}); err != nil {
		t.Fatalf("expected no-arg method to pass with empty args, got %v", err)
	}
}

func asABIMismatch(err error, out **ABIMismatchError) bool {
	if m, ok := err.(*ABIMismatchError); ok {
		*out = m
		return true
	}
	return false
}
