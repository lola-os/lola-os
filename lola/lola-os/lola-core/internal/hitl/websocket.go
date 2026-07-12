package hitl

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// wsMessage is the wire format sent to/from connected WebSocket clients.
type wsMessage struct {
	Type     string  `json:"type"` // "approval_request" | "approval_response"
	ID       string  `json:"id"`
	Request  *Request `json:"request,omitempty"`
	Decision string  `json:"decision,omitempty"`
}

// WebSocketServer hosts a localhost-only HITL approval channel. It is
// intentionally unauthenticated, per the blueprint: it binds to 127.0.0.1
// only and is meant to be consumed by a UI running on the same machine as
// the user.
type WebSocketServer struct {
	addr     string
	upgrader websocket.Upgrader
	server   *http.Server

	mu      sync.Mutex
	clients map[*websocket.Conn]bool
	pending map[string]chan Decision
}

// NewWebSocketServer constructs a server bound to addr (must be a
// loopback address, e.g. "127.0.0.1:8765"). It refuses to bind to any
// non-loopback address as a safety guard, since the protocol has no auth.
func NewWebSocketServer(addr string) (*WebSocketServer, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("hitl: invalid websocket address %q: %w", addr, err)
	}
	ip := net.ParseIP(host)
	if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return nil, fmt.Errorf("hitl: websocket server must bind to a loopback address, got %q", addr)
	}
	return &WebSocketServer{
		addr:     addr,
		upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		clients:  map[*websocket.Conn]bool{},
		pending:  map[string]chan Decision{},
	}, nil
}

// Start begins listening in the background. Call Stop to shut down.
func (s *WebSocketServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleConn)
	s.server = &http.Server{Addr: s.addr, Handler: mux}

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("hitl: binding websocket server to %s: %w", s.addr, err)
	}
	go func() {
		_ = s.server.Serve(ln)
	}()
	return nil
}

// Stop gracefully shuts down the server.
func (s *WebSocketServer) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

func (s *WebSocketServer) handleConn(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		conn.Close()
	}()

	for {
		var msg wsMessage
		if err := conn.ReadJSON(&msg); err != nil {
			return
		}
		if msg.Type == "approval_response" {
			s.mu.Lock()
			ch, ok := s.pending[msg.ID]
			s.mu.Unlock()
			if ok {
				select {
				case ch <- Decision(msg.Decision):
				default:
				}
			}
		}
	}
}

// RequestApproval broadcasts req to all connected clients and waits (up to
// timeout) for any one of them to respond. Implements the Approver
// interface so it is interchangeable with ConsoleApprover.
func (s *WebSocketServer) RequestApproval(ctx context.Context, req Request, timeout time.Duration) (Decision, error) {
	ch := make(chan Decision, 1)
	s.mu.Lock()
	s.pending[req.ID] = ch
	clients := make([]*websocket.Conn, 0, len(s.clients))
	for c := range s.clients {
		clients = append(clients, c)
	}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.pending, req.ID)
		s.mu.Unlock()
	}()

	if len(clients) == 0 {
		return DecisionTimeout, fmt.Errorf("hitl: no WebSocket clients connected to approve request %s", req.ID)
	}

	msg := wsMessage{Type: "approval_request", ID: req.ID, Request: &req}
	payload, err := json.Marshal(msg)
	if err != nil {
		return DecisionTimeout, fmt.Errorf("hitl: encoding approval request: %w", err)
	}
	for _, c := range clients {
		_ = c.WriteMessage(websocket.TextMessage, payload)
	}

	select {
	case <-ctx.Done():
		return DecisionTimeout, ctx.Err()
	case <-time.After(timeout):
		return DecisionTimeout, nil
	case d := <-ch:
		return d, nil
	}
}

// ConnectedClients returns the number of currently connected WebSocket
// clients, useful for `lola doctor`.
func (s *WebSocketServer) ConnectedClients() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.clients)
}
