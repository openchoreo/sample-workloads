// Copyright 2026 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// --- Chat API types ---

type ChatRequest struct {
	Message string `json:"message"`
}

type ChatResponse struct {
	Reply     string   `json:"reply"`
	ToolsUsed []string `json:"tools_used,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// --- OpenAI API types ---

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
	Tools    []openAITool    `json:"tools,omitempty"`
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openAITool struct {
	Type     string         `json:"type"`
	Function openAIFunction `json:"function"`
}

type openAIFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type openAIToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type openAIResponse struct {
	Choices []struct {
		Message      openAIMessage `json:"message"`
		FinishReason string        `json:"finish_reason"`
	} `json:"choices"`
}

// --- MCP JSON-RPC types ---

type jsonrpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
	ID      int    `json:"id,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type mcpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type mcpToolsListResult struct {
	Tools []mcpTool `json:"tools"`
}

type mcpToolCallResult struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	IsError bool `json:"isError"`
}

// --- Server ---

type Server struct {
	llmBaseURL  string
	llmAPIKey   string
	llmModel    string
	mcpGateway  string
	componentID string
	httpClient  *http.Client
	logger      *slog.Logger

	toolsMu      sync.RWMutex
	mcpTools     []mcpTool
	oaiTools     []openAITool
	mcpSessionID string
	reqID        atomic.Int64
}

const maxToolRounds = 10

const systemPrompt = `You are a helpful AI reading assistant. You help users manage their reading list and take notes about books.

You have access to tools that let you:
- Manage a reading list (add, list, update, delete books)
- Take and search reading notes

Use the available tools to help users with their requests. When listing books or notes, present the information in a clear, readable format.`

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(os.Getenv("LOG_LEVEL")),
	}))

	llmBaseURL := os.Getenv("OPENAI_BASE_URL")
	if llmBaseURL == "" {
		llmBaseURL = "https://api.openai.com"
	}

	llmModel := os.Getenv("OPENAI_MODEL")
	if llmModel == "" {
		llmModel = "gpt-4o-mini"
	}

	mcpGateway := os.Getenv("MCP_GATEWAY_URL")
	if mcpGateway == "" {
		logger.Error("MCP_GATEWAY_URL is required")
		os.Exit(1)
	}

	srv := &Server{
		llmBaseURL:  strings.TrimRight(llmBaseURL, "/"),
		llmAPIKey:   os.Getenv("OPENAI_API_KEY"),
		llmModel:    llmModel,
		mcpGateway:  strings.TrimRight(mcpGateway, "/"),
		componentID: os.Getenv("X_OPENCHOREO_COMPONENT"),
		httpClient:  &http.Client{Timeout: 60 * time.Second},
		logger:      logger,
	}

	// Discover tools from the MCP gateway on startup.
	if err := srv.discoverTools(context.Background()); err != nil {
		logger.Warn("failed to discover tools on startup, will retry on first request", "error", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.handleHealth)
	mux.HandleFunc("GET /readyz", srv.handleReady)
	mux.HandleFunc("POST /chat", srv.handleChat)
	mux.HandleFunc("POST /refresh-tools", srv.handleRefreshTools)

	addr := fmt.Sprintf(":%d", *port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("starting AI reading assistant",
			"addr", addr, "model", llmModel, "mcp_gateway", mcpGateway)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	httpServer.Shutdown(shutdownCtx)
}

// =============================================================================
// Tool Discovery — connects to MCP gateway and discovers all available tools
// =============================================================================

func (s *Server) discoverTools(ctx context.Context) error {
	// Step 1: Initialize MCP session
	if err := s.mcpCall(ctx, "initialize", map[string]any{
		"protocolVersion": "2025-03-26",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]any{
			"name":    "ai-reading-assistant",
			"version": "1.0.0",
		},
	}, nil); err != nil {
		return fmt.Errorf("mcp initialize: %w", err)
	}

	// Step 2: Send initialized notification
	s.mcpNotify(ctx, "notifications/initialized")

	// Step 3: List all tools (from all federated sources)
	var result mcpToolsListResult
	if err := s.mcpCall(ctx, "tools/list", nil, &result); err != nil {
		return fmt.Errorf("mcp tools/list: %w", err)
	}

	// Step 4: Convert MCP tool definitions to OpenAI function-calling format
	oaiTools := make([]openAITool, len(result.Tools))
	for i, t := range result.Tools {
		oaiTools[i] = openAITool{
			Type: "function",
			Function: openAIFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.InputSchema,
			},
		}
	}

	s.toolsMu.Lock()
	s.mcpTools = result.Tools
	s.oaiTools = oaiTools
	s.toolsMu.Unlock()

	s.logger.Info("discovered tools", "count", len(result.Tools), "tools", toolNames(result.Tools))
	return nil
}

// =============================================================================
// MCP Client — JSON-RPC over Streamable HTTP
// =============================================================================

func (s *Server) mcpCall(ctx context.Context, method string, params any, result any) error {
	id := int(s.reqID.Add(1))
	req := jsonrpcRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.mcpGateway+"/mcp", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	if s.componentID != "" {
		httpReq.Header.Set("x-openchoreo-component", s.componentID)
	}
	s.toolsMu.RLock()
	sessionID := s.mcpSessionID
	s.toolsMu.RUnlock()
	if sessionID != "" {
		httpReq.Header.Set("mcp-session-id", sessionID)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Capture session ID from initialize response
	if sid := resp.Header.Get("mcp-session-id"); sid != "" {
		s.toolsMu.Lock()
		s.mcpSessionID = sid
		s.toolsMu.Unlock()
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Handle SSE format: strip "data: " prefix if present
	parsed := respBody
	for _, line := range bytes.Split(parsed, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if bytes.HasPrefix(line, []byte("data: ")) {
			parsed = bytes.TrimPrefix(line, []byte("data: "))
			break
		}
	}

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(parsed, &rpcResp); err != nil {
		return fmt.Errorf("unmarshal response: %w (body: %s)", err, string(respBody))
	}

	if rpcResp.Error != nil {
		return fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	if result != nil {
		if err := json.Unmarshal(rpcResp.Result, result); err != nil {
			return fmt.Errorf("unmarshal result: %w", err)
		}
	}
	return nil
}

func (s *Server) mcpNotify(ctx context.Context, method string) {
	req := jsonrpcRequest{JSONRPC: "2.0", Method: method}
	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.mcpGateway+"/mcp", bytes.NewReader(body))
	if err != nil {
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	if s.componentID != "" {
		httpReq.Header.Set("x-openchoreo-component", s.componentID)
	}
	s.toolsMu.RLock()
	sessionID := s.mcpSessionID
	s.toolsMu.RUnlock()
	if sessionID != "" {
		httpReq.Header.Set("mcp-session-id", sessionID)
	}
	resp, err := s.httpClient.Do(httpReq)
	if err == nil {
		resp.Body.Close()
	}
}

func (s *Server) mcpToolCall(ctx context.Context, name string, arguments map[string]any) (string, error) {
	var result mcpToolCallResult
	if err := s.mcpCall(ctx, "tools/call", map[string]any{
		"name":      name,
		"arguments": arguments,
	}, &result); err != nil {
		return "", err
	}

	var texts []string
	for _, c := range result.Content {
		if c.Type == "text" {
			texts = append(texts, c.Text)
		}
	}
	return strings.Join(texts, "\n"), nil
}

// =============================================================================
// LLM Client — OpenAI-compatible chat completions with function calling
// =============================================================================

func (s *Server) callLLM(ctx context.Context, messages []openAIMessage, tools []openAITool) (*openAIResponse, error) {
	reqBody := openAIRequest{
		Model:    s.llmModel,
		Messages: messages,
	}
	if len(tools) > 0 {
		reqBody.Tools = tools
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.llmBaseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if s.llmAPIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+s.llmAPIKey)
	}
	if s.componentID != "" {
		httpReq.Header.Set("x-openchoreo-component", s.componentID)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &openAIResp, nil
}

// =============================================================================
// HTTP Handlers
// =============================================================================

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(w, "ok")
}

func (s *Server) handleReady(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(w, "ready")
}

func (s *Server) handleRefreshTools(w http.ResponseWriter, r *http.Request) {
	if err := s.discoverTools(r.Context()); err != nil {
		writeError(w, http.StatusBadGateway, "failed to refresh tools: "+err.Error())
		return
	}
	s.toolsMu.RLock()
	names := toolNames(s.mcpTools)
	s.toolsMu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"tools": names})
}

// handleChat implements the agentic tool-calling loop:
//  1. Send user message + available tools to LLM
//  2. If LLM returns tool_calls -> execute each via MCP gateway -> feed results back
//  3. Repeat until LLM returns a text response (or max rounds reached)
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.Message) == "" {
		writeError(w, http.StatusBadRequest, "message is required")
		return
	}

	s.logger.Info("chat request", "message_length", len(req.Message))

	// Get cached tools
	s.toolsMu.RLock()
	oaiTools := s.oaiTools
	s.toolsMu.RUnlock()

	if len(oaiTools) == 0 {
		// Try discovering tools if we don't have any yet
		if err := s.discoverTools(r.Context()); err != nil {
			s.logger.Error("tool discovery failed", "error", err)
		}
		s.toolsMu.RLock()
		oaiTools = s.oaiTools
		s.toolsMu.RUnlock()
	}

	messages := []openAIMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: req.Message},
	}

	var toolsUsed []string

	// Agentic loop
	for round := 0; round < maxToolRounds; round++ {
		resp, err := s.callLLM(r.Context(), messages, oaiTools)
		if err != nil {
			s.logger.Error("LLM call failed", "error", err, "round", round)
			writeError(w, http.StatusBadGateway, "AI provider error")
			return
		}

		if len(resp.Choices) == 0 {
			writeError(w, http.StatusBadGateway, "no response from AI provider")
			return
		}

		choice := resp.Choices[0]

		// No tool calls — return the text response
		if len(choice.Message.ToolCalls) == 0 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ChatResponse{
				Reply:     choice.Message.Content,
				ToolsUsed: toolsUsed,
			})
			return
		}

		// Add assistant message with tool calls to the conversation
		messages = append(messages, choice.Message)

		// Execute each tool call via the MCP gateway
		for _, tc := range choice.Message.ToolCalls {
			s.logger.Info("calling tool", "tool", tc.Function.Name, "round", round)
			toolsUsed = append(toolsUsed, tc.Function.Name)

			var args map[string]any
			json.Unmarshal([]byte(tc.Function.Arguments), &args)

			result, err := s.mcpToolCall(r.Context(), tc.Function.Name, args)
			if err != nil {
				result = fmt.Sprintf("Error calling tool: %s", err.Error())
			}

			// Add tool result to conversation
			messages = append(messages, openAIMessage{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
	}

	writeError(w, http.StatusInternalServerError, "too many tool call rounds")
}

// =============================================================================
// Helpers
// =============================================================================

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}

func toolNames(tools []mcpTool) []string {
	names := make([]string, len(tools))
	for i, t := range tools {
		names[i] = t.Name
	}
	return names
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
