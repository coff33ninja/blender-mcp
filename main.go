// blender-mcp is a Go MCP server that bridges MCP clients (via stdio)
// to Blender's TCP socket server (the blender-mcp add-on).
//
// Usage:
//
//	go-blender-mcp [--host localhost] [--port 9876] [--verbose]
//
// The server reads JSON-RPC 2.0 messages from stdin and writes responses
// to stdout, conforming to the Model Context Protocol (MCP) specification.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/coff33ninja/blender-mcp/blender"
	"github.com/coff33ninja/blender-mcp/mcp"
)

const (
	serverName    = "blender-mcp"
	serverVersion = "1.0.0"
)

func main() {
	host := flag.String("host", "localhost", "Blender TCP server host")
	port := flag.Int("port", 9876, "Blender TCP server port")
	verbose := flag.Bool("verbose", false, "Enable debug logging to stderr")
	flag.Parse()

	if *verbose {
		log.SetOutput(os.Stderr)
		log.SetFlags(log.Ltime | log.Lmicroseconds)
	} else {
		log.SetOutput(os.Stderr)
		log.SetFlags(0)
	}

	log.Printf("starting %s v%s", serverName, serverVersion)
	log.Printf("blender target: %s:%d", *host, *port)

	// Create Blender TCP client (connects per-request).
	bc := blender.NewClient(*host, *port)

	// Register all Blender tools into a registry.
	registry := mcp.RegisterTools(bc)

	// Create MCP server over stdio.
	s := mcp.NewServer(os.Stdin, os.Stdout)

	// --- MCP protocol handlers ---

	s.Handle("initialize", func(params json.RawMessage) (any, *mcp.Error) {
		var p mcp.InitializeParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &mcp.Error{Code: mcp.CodeInvalidParams, Message: "invalid params"}
		}
		log.Printf("client: %s v%s (protocol %s)", p.ClientInfo.Name, p.ClientInfo.Version, p.ProtocolVersion)
		return mcp.InitializeResult{
			ProtocolVersion: "2025-03-26",
			Capabilities: mcp.ServerCaps{
				Tools: &mcp.ToolsCapability{ListChanged: false},
			},
			ServerInfo: mcp.ServerInfo{
				Name:    serverName,
				Version: serverVersion,
			},
		}, nil
	})

	s.Handle("tools/list", func(params json.RawMessage) (any, *mcp.Error) {
		return mcp.ToolsListResult{Tools: registry.List()}, nil
	})

	s.Handle("tools/call", func(params json.RawMessage) (any, *mcp.Error) {
		var p mcp.ToolCallParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &mcp.Error{Code: mcp.CodeInvalidParams, Message: "invalid params"}
		}
		return registry.Call(p.Name, p.Arguments)
	})

	// Handle graceful shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("received signal %v, shutting down", sig)
		bc.Disconnect()
		os.Exit(0)
	}()

	// Run the MCP server (blocks until stdin EOF).
	if err := s.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}

	log.Println("stdin closed, exiting")
	bc.Disconnect()
}
