// Copyright 2026 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
)

// --- JSON-RPC 2.0 types ---

type jsonrpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      any             `json:"id,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
	ID      any       `json:"id"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// --- MCP protocol types ---

type mcpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type toolResult struct {
	Content []contentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// --- Note model ---

type Note struct {
	ID        int    `json:"id"`
	BookTitle string `json:"book_title"`
	Text      string `json:"text"`
}

// --- MCP Server ---

type MCPServer struct {
	mu     sync.RWMutex
	notes  []Note
	nextID int
	logger *slog.Logger
}

var tools = []mcpTool{
	{
		Name:        "add_note",
		Description: "Add a reading note for a book. Use this to save thoughts, quotes, or observations about a book.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"book_title": map[string]any{
					"type":        "string",
					"description": "Title of the book this note is about",
				},
				"text": map[string]any{
					"type":        "string",
					"description": "The note content",
				},
			},
			"required": []string{"book_title", "text"},
		},
	},
	{
		Name:        "list_notes",
		Description: "List all reading notes, optionally filtered by book title.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"book_title": map[string]any{
					"type":        "string",
					"description": "Filter notes by book title (optional, case-insensitive partial match)",
				},
			},
		},
	},
	{
		Name:        "search_notes",
		Description: "Search reading notes by keyword. Searches in both book titles and note text.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Search keyword",
				},
			},
			"required": []string{"query"},
		},
	},
}

func main() {
	port := flag.Int("port", 8081, "HTTP server port")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	server := &MCPServer{logger: logger}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /mcp", server.handleMCP)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	addr := fmt.Sprintf(":%d", *port)
	logger.Info("starting reading-notes MCP server", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func (s *MCPServer) handleMCP(w http.ResponseWriter, r *http.Request) {
	var req jsonrpcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, jsonrpcResponse{
			JSONRPC: "2.0",
			Error:   &rpcError{Code: -32700, Message: "parse error"},
		})
		return
	}

	s.logger.Info("mcp request", "method", req.Method)

	switch req.Method {
	case "initialize":
		writeJSON(w, jsonrpcResponse{
			JSONRPC: "2.0",
			Result: map[string]any{
				"protocolVersion": "2025-03-26",
				"capabilities": map[string]any{
					"tools": map[string]any{},
				},
				"serverInfo": map[string]any{
					"name":    "reading-notes",
					"version": "1.0.0",
				},
			},
			ID: req.ID,
		})

	case "notifications/initialized":
		w.WriteHeader(http.StatusAccepted)

	case "tools/list":
		writeJSON(w, jsonrpcResponse{
			JSONRPC: "2.0",
			Result:  map[string]any{"tools": tools},
			ID:      req.ID,
		})

	case "tools/call":
		result := s.handleToolCall(req.Params)
		writeJSON(w, jsonrpcResponse{
			JSONRPC: "2.0",
			Result:  result,
			ID:      req.ID,
		})

	default:
		writeJSON(w, jsonrpcResponse{
			JSONRPC: "2.0",
			Error:   &rpcError{Code: -32601, Message: "method not found: " + req.Method},
			ID:      req.ID,
		})
	}
}

func (s *MCPServer) handleToolCall(params json.RawMessage) toolResult {
	var tc toolCallParams
	if err := json.Unmarshal(params, &tc); err != nil {
		return toolResult{
			Content: []contentBlock{{Type: "text", Text: "invalid tool call params: " + err.Error()}},
			IsError: true,
		}
	}

	switch tc.Name {
	case "add_note":
		return s.addNote(tc.Arguments)
	case "list_notes":
		return s.listNotes(tc.Arguments)
	case "search_notes":
		return s.searchNotes(tc.Arguments)
	default:
		return toolResult{
			Content: []contentBlock{{Type: "text", Text: "unknown tool: " + tc.Name}},
			IsError: true,
		}
	}
}

func (s *MCPServer) addNote(args map[string]any) toolResult {
	bookTitle, _ := args["book_title"].(string)
	text, _ := args["text"].(string)

	if bookTitle == "" || text == "" {
		return toolResult{
			Content: []contentBlock{{Type: "text", Text: "book_title and text are required"}},
			IsError: true,
		}
	}

	s.mu.Lock()
	s.nextID++
	note := Note{ID: s.nextID, BookTitle: bookTitle, Text: text}
	s.notes = append(s.notes, note)
	s.mu.Unlock()

	data, _ := json.Marshal(note)
	return toolResult{
		Content: []contentBlock{{Type: "text", Text: fmt.Sprintf("Note added: %s", string(data))}},
	}
}

func (s *MCPServer) listNotes(args map[string]any) toolResult {
	bookTitle, _ := args["book_title"].(string)

	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []Note
	for _, n := range s.notes {
		if bookTitle == "" || strings.Contains(strings.ToLower(n.BookTitle), strings.ToLower(bookTitle)) {
			filtered = append(filtered, n)
		}
	}

	if len(filtered) == 0 {
		return toolResult{
			Content: []contentBlock{{Type: "text", Text: "No notes found."}},
		}
	}

	data, _ := json.MarshalIndent(filtered, "", "  ")
	return toolResult{
		Content: []contentBlock{{Type: "text", Text: string(data)}},
	}
}

func (s *MCPServer) searchNotes(args map[string]any) toolResult {
	query, _ := args["query"].(string)
	if query == "" {
		return toolResult{
			Content: []contentBlock{{Type: "text", Text: "query is required"}},
			IsError: true,
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	q := strings.ToLower(query)
	var matches []Note
	for _, n := range s.notes {
		if strings.Contains(strings.ToLower(n.BookTitle), q) ||
			strings.Contains(strings.ToLower(n.Text), q) {
			matches = append(matches, n)
		}
	}

	if len(matches) == 0 {
		return toolResult{
			Content: []contentBlock{{Type: "text", Text: fmt.Sprintf("No notes matching '%s'.", query)}},
		}
	}

	data, _ := json.MarshalIndent(matches, "", "  ")
	return toolResult{
		Content: []contentBlock{{Type: "text", Text: string(data)}},
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
