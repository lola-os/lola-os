// Package jsonrpc implements a minimal JSON-RPC 2.0 server used as
// lola-core's integration surface for the Python and TypeScript SDKs
// (which spawn lola-core as a subprocess and talk to it over stdin/stdout)
// and optionally over a local TCP socket (for the Go SDK or any other
// direct-connect consumer).
//
// This package intentionally implements only the slice of JSON-RPC 2.0
// LOLA needs — single requests, no batching, no notifications — to keep
// the wire protocol simple and easy for SDK authors to reason about.
package jsonrpc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
)

// Request is a JSON-RPC 2.0 request object.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response is a JSON-RPC 2.0 response object.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// Error is a JSON-RPC 2.0 error object. Code follows the conventions in
// the blueprint's error taxonomy so SDKs can map codes to typed
// exceptions (BudgetExceededError, ABIMismatchError, RPCConnectionError).
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Standard JSON-RPC error codes.
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
)

// LOLA-specific error codes, starting at -33000 to avoid colliding with
// the JSON-RPC reserved range.
const (
	CodeBudgetExceeded = -33001
	CodeABIMismatch    = -33002
	CodeRPCConnection  = -33003
	CodeApprovalDenied = -33004
	CodeUnknownChain   = -33005
)

// HandlerFunc handles a single method call. params is the raw JSON
// `params` field; implementations should unmarshal it into their expected
// type. Returning (nil, err) where err is a *Error preserves the LOLA
// error code; any other error is wrapped as CodeInternalError.
type HandlerFunc func(ctx context.Context, params json.RawMessage) (interface{}, error)

// Server dispatches JSON-RPC requests to registered method handlers.
type Server struct {
	mu       sync.RWMutex
	handlers map[string]HandlerFunc
}

// NewServer constructs an empty Server. Register methods with Handle.
func NewServer() *Server {
	return &Server{handlers: map[string]HandlerFunc{}}
}

// Handle registers a handler for method.
func (s *Server) Handle(method string, fn HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[method] = fn
}

// ServeStdio reads newline-delimited JSON-RPC requests from r and writes
// responses to w, one per line — the transport used by the Python and
// TypeScript SDKs when they spawn lola-core as a subprocess.
func (s *Server) ServeStdio(ctx context.Context, r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024) // allow large params (e.g. ABIs)
	var writeMu sync.Mutex

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		// Copy the line since scanner.Bytes() is reused on next Scan().
		data := make([]byte, len(line))
		copy(data, line)

		go func(raw []byte) {
			resp := s.dispatch(ctx, raw)
			out, err := json.Marshal(resp)
			if err != nil {
				return
			}
			writeMu.Lock()
			defer writeMu.Unlock()
			w.Write(out)
			w.Write([]byte("\n"))
		}(data)
	}
	return scanner.Err()
}

// ServeTCP listens on addr and serves each connection as an independent
// newline-delimited JSON-RPC stream (same framing as ServeStdio).
func (s *Server) ServeTCP(ctx context.Context, addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("jsonrpc: listening on %s: %w", addr, err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				continue
			}
		}
		go func(c net.Conn) {
			defer c.Close()
			_ = s.ServeStdio(ctx, c, c)
		}(conn)
	}
}

func (s *Server) dispatch(ctx context.Context, raw []byte) Response {
	var req Request
	if err := json.Unmarshal(raw, &req); err != nil {
		return Response{JSONRPC: "2.0", Error: &Error{Code: CodeParseError, Message: "invalid JSON: " + err.Error()}}
	}
	if req.Method == "" {
		return Response{JSONRPC: "2.0", ID: req.ID, Error: &Error{Code: CodeInvalidRequest, Message: "missing method"}}
	}

	s.mu.RLock()
	handler, ok := s.handlers[req.Method]
	s.mu.RUnlock()
	if !ok {
		return Response{JSONRPC: "2.0", ID: req.ID, Error: &Error{Code: CodeMethodNotFound, Message: "method not found: " + req.Method}}
	}

	result, err := handler(ctx, req.Params)
	if err != nil {
		if rpcErr, ok := err.(*Error); ok {
			return Response{JSONRPC: "2.0", ID: req.ID, Error: rpcErr}
		}
		return Response{JSONRPC: "2.0", ID: req.ID, Error: &Error{Code: CodeInternalError, Message: err.Error()}}
	}
	return Response{JSONRPC: "2.0", ID: req.ID, Result: result}
}

// NewError constructs a *Error, satisfying the standard `error` interface
// so handlers can `return nil, jsonrpc.NewError(...)`.
func NewError(code int, message string, data interface{}) *Error {
	return &Error{Code: code, Message: message, Data: data}
}

func (e *Error) Error() string {
	return fmt.Sprintf("jsonrpc error %d: %s", e.Code, e.Message)
}
