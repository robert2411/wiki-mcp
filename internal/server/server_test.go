package server_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/robertstevens/wiki-mcp/internal/config"
	"github.com/robertstevens/wiki-mcp/internal/server"
)

func testConfig(t *testing.T) *config.Config {
	t.Helper()
	return &config.Config{
		WikiPath: t.TempDir(),
	}
}

// jsonRPCRequest builds a JSON-RPC 2.0 request string with newline terminator.
func jsonRPCRequest(id int, method string, params any) string {
	p, _ := json.Marshal(params)
	return fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"%s","params":%s}`, id, method, p) + "\n"
}

func TestHandshake(t *testing.T) {
	cfg := testConfig(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := server.New(cfg, "test", logger)

	// Pipe for stdin: we write requests, server reads them.
	stdinR, stdinW := io.Pipe()
	// Pipe for stdout: server writes responses, we read them.
	stdoutR, stdoutW := io.Pipe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.RunStdio(ctx, stdinR, stdoutW)
	}()

	initReq := jsonRPCRequest(1, "initialize", map[string]any{
		"protocolVersion": "2025-03-26",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]any{
			"name":    "test-client",
			"version": "1.0",
		},
	})

	_, err := stdinW.Write([]byte(initReq))
	if err != nil {
		t.Fatalf("write initialize request: %v", err)
	}

	scanner := bufio.NewScanner(stdoutR)
	if !scanner.Scan() {
		t.Fatal("no response from server")
	}
	resp := scanner.Text()

	var result map[string]any
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		t.Fatalf("unmarshal response: %v\nraw: %s", err, resp)
	}

	// Verify it's a valid JSON-RPC response with server info.
	if result["id"] != float64(1) {
		t.Errorf("expected id=1, got %v", result["id"])
	}
	res, ok := result["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got %v", result)
	}
	info, ok := res["serverInfo"].(map[string]any)
	if !ok {
		t.Fatalf("expected serverInfo in result, got %v", res)
	}
	if info["name"] != "wiki-mcp" {
		t.Errorf("expected serverInfo.name=wiki-mcp, got %v", info["name"])
	}

	// Verify capabilities include tools, resources, prompts.
	caps, ok := res["capabilities"].(map[string]any)
	if !ok {
		t.Fatalf("expected capabilities in result, got %v", res)
	}
	for _, cap := range []string{"tools", "resources", "prompts"} {
		if _, exists := caps[cap]; !exists {
			t.Errorf("expected capability %q in response", cap)
		}
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil && !strings.Contains(err.Error(), "context canceled") {
			t.Errorf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down within 2s")
	}
}

func TestContextCancellationShutdown(t *testing.T) {
	cfg := testConfig(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := server.New(cfg, "test", logger)

	stdinR, stdinW := io.Pipe()
	_, stdoutW := io.Pipe()

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.RunStdio(ctx, stdinR, stdoutW)
	}()

	// Give server moment to start, then cancel.
	time.Sleep(50 * time.Millisecond)
	cancel()
	stdinW.Close()

	start := time.Now()
	select {
	case <-errCh:
		elapsed := time.Since(start)
		if elapsed > 1*time.Second {
			t.Errorf("shutdown took %v, expected <1s", elapsed)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down within 2s after context cancellation")
	}
}

func TestSSENotImplemented(t *testing.T) {
	if server.ErrSSENotImplemented == nil {
		t.Fatal("ErrSSENotImplemented should not be nil")
	}
	msg := server.ErrSSENotImplemented.Error()
	if !strings.Contains(msg, "not implemented") {
		t.Errorf("expected 'not implemented' in error, got: %s", msg)
	}
	if !strings.Contains(msg, "TASK-22") {
		t.Errorf("expected 'TASK-22' in error, got: %s", msg)
	}
}
