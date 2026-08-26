// Package blender provides a TCP client for communicating with the
// Blender MCP add-on's socket server.
//
// Protocol: null-byte-delimited JSON over TCP.
// Request:  {"type": "execute", "code": "...", "strict_json": true}\0
// Response: {"status": "ok"|"error", "result": {...}, "stdout": "...", "stderr": "..."}\0
//
// The Blender add-on closes the connection after each request, so each
// ExecuteCode call opens a fresh TCP connection.
package blender

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

const (
	defaultHost      = "localhost"
	defaultPort      = 9876
	connectTimeout   = 5 * time.Second
	requestTimeout   = 120 * time.Second
	maxResponseBytes = 50 * 1024 * 1024 // 50 MiB
)

// ExecResult is the response from Blender after executing code.
type ExecResult struct {
	Status  string         `json:"status"`
	Message string         `json:"message,omitempty"`
	Result  map[string]any `json:"result,omitempty"`
	Stdout  string         `json:"stdout,omitempty"`
	Stderr  string         `json:"stderr,omitempty"`
}

// Client connects to Blender's TCP socket server.
type Client struct {
	host string
	port int
}

// NewClient creates a client targeting the given address.
func NewClient(host string, port int) *Client {
	if host == "" {
		host = defaultHost
	}
	if port <= 0 {
		port = defaultPort
	}
	return &Client{host: host, port: port}
}

// Disconnect is a no-op kept for API compatibility (each request is a fresh connection).
func (c *Client) Disconnect() {}

// Connected always returns true — connections are per-request.
func (c *Client) Connected() bool { return true }

// ExecuteCode sends Python code to Blender for execution and returns the result.
func (c *Client) ExecuteCode(code string, strictJson bool) (*ExecResult, error) {
	addr := net.JoinHostPort(c.host, fmt.Sprintf("%d", c.port))
	conn, err := net.DialTimeout("tcp", addr, connectTimeout)
	if err != nil {
		return nil, fmt.Errorf("connect to blender at %s: %w", addr, err)
	}
	defer conn.Close()

	// Build request.
	req := map[string]any{
		"type":        "execute",
		"code":        code,
		"strict_json": strictJson,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	reqBytes = append(reqBytes, 0) // null-byte delimiter

	// Set deadline.
	conn.SetDeadline(time.Now().Add(requestTimeout))

	// Send request.
	if _, err := conn.Write(reqBytes); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	// Read response until null byte.
	reader := bufio.NewReaderSize(conn, 4096)
	var buf []byte
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}
		if b == 0 {
			break
		}
		buf = append(buf, b)
		if len(buf) > maxResponseBytes {
			return nil, fmt.Errorf("response exceeds %d byte limit", maxResponseBytes)
		}
	}

	var result ExecResult
	if err := json.Unmarshal(buf, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &result, nil
}

// Execute is a convenience wrapper that sends code and returns the combined output.
func (c *Client) Execute(code string) (string, error) {
	result, err := c.ExecuteCode(code, true)
	if err != nil {
		return "", err
	}
	if result.Status == "error" {
		return "", fmt.Errorf("blender error: %s", result.Message)
	}
	out := ""
	if result.Stdout != "" {
		out += result.Stdout
	}
	if result.Stderr != "" {
		if out != "" {
			out += "\n"
		}
		out += result.Stderr
	}
	return out, nil
}
