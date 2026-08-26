package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
)

// Handler processes a method name and params, returning a result or error.
type Handler func(params json.RawMessage) (any, *Error)

// Server reads JSON-RPC from Reader and writes to Writer.
type Server struct {
	reader   io.Reader
	writer   io.Writer
	mu       sync.Mutex
	handlers map[string]Handler

	// Set after initialize handshake.
	initialized bool
}

// NewServer creates a server that reads from r and writes to w.
func NewServer(r io.Reader, w io.Writer) *Server {
	return &Server{
		reader:   r,
		writer:   w,
		handlers: make(map[string]Handler),
	}
}

// Handle registers a handler for the given MCP method.
func (s *Server) Handle(method string, h Handler) {
	s.handlers[method] = h
}

// Run reads requests from stdin and writes responses to stdout until EOF or error.
func (s *Server) Run() error {
	scanner := bufio.NewScanner(s.reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 1MB max message

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		s.handleMessage(line)
	}
	return scanner.Err()
}

// SendMessage writes a raw JSON-RPC message to the transport (for notifications).
func (s *Server) SendMessage(msg Response) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}
	data = append(data, '\n')
	_, err = s.writer.Write(data)
	return err
}

func (s *Server) handleRequest(req Request) {
	// Check if handler exists.
	h, ok := s.handlers[req.Method]
	if !ok {
		s.sendError(req.ID, CodeMethodNotFound, fmt.Sprintf("method not found: %s", req.Method))
		return
	}

	// Enforce initialize-first for all methods except initialize.
	if req.Method != "initialize" && !s.initialized {
		s.sendError(req.ID, CodeInternalError, "server not initialized")
		return
	}

	result, rpcErr := h(req.Params)
	if rpcErr != nil {
		s.sendError(req.ID, rpcErr.Code, rpcErr.Message)
		return
	}

	s.sendResult(req.ID, result)

	// Mark initialized after successful initialize response.
	if req.Method == "initialize" {
		s.mu.Lock()
		s.initialized = true
		s.mu.Unlock()
	}
}

func (s *Server) handleMessage(data []byte) {
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		log.Printf("invalid JSON-RPC request: %v", err)
		return
	}

	// Notifications have no ID — we don't respond.
	if req.ID == nil {
		// Handle "notifications/initialized" as a no-op.
		if req.Method == "notifications/initialized" {
			return
		}
		log.Printf("ignoring notification: %s", req.Method)
		return
	}

	s.handleRequest(req)
}

func (s *Server) sendResult(id json.RawMessage, result any) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("failed to marshal result: %v", err)
		return
	}
	data = append(data, '\n')
	if _, err := s.writer.Write(data); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func (s *Server) sendError(id json.RawMessage, code int, message string) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("failed to marshal error: %v", err)
		return
	}
	data = append(data, '\n')
	if _, err := s.writer.Write(data); err != nil {
		log.Printf("failed to write error response: %v", err)
	}
}
