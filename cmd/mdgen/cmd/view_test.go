package cmd

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestViewCmd_Success(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "a.yaml")
	content := "title: My Doc\nblocks:\n  - type: paragraph\n    text: hello world\n"
	if err := os.WriteFile(yamlPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", yamlPath, "--no-serve"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v\noutput: %s", err, out.String())
	}
	if !strings.Contains(out.String(), "a.yaml") {
		t.Errorf("expected output to mention the file name, got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "1 paragraph") {
		t.Errorf("expected output to include block summary, got:\n%s", out.String())
	}
}

func TestViewCmd_MissingFile(t *testing.T) {
	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", "/nonexistent.yaml"})

	if err := root.Execute(); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestViewCmd_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(yamlPath, []byte("blocks:\n  - type: paragraph\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", yamlPath, "--no-serve"})

	if err := root.Execute(); err == nil {
		t.Fatal("expected error for invalid document")
	}
}

func TestViewCmd_ServesUntilContextCanceled(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "a.yaml")
	content := "title: My Doc\nblocks:\n  - type: paragraph\n    text: hello world\n"
	if err := os.WriteFile(yamlPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	root := newRootCmd()
	var out syncBuffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", yamlPath, "--port", "0"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	root.SetContext(ctx)

	done := make(chan error, 1)
	go func() { done <- root.ExecuteContext(ctx) }()

	url := waitForURL(t, &out)

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET %s: status = %d, want 200", url, resp.StatusCode)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Execute() returned error after cancel: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("view command did not shut down after context cancellation")
	}
}

// syncBuffer is a bytes.Buffer safe for concurrent writes (from the
// command goroutine) and reads (from the polling test goroutine).
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

// waitForURL polls buf for the "Edit in your browser: <url>" line printed
// once the server starts listening, up to a short timeout.
func waitForURL(t *testing.T, buf *syncBuffer) string {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if idx := strings.Index(buf.String(), "http://"); idx != -1 {
			line := buf.String()[idx:]
			if nl := strings.IndexAny(line, "\n\r"); nl != -1 {
				line = line[:nl]
			}
			return strings.TrimSpace(line)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out waiting for server URL in output")
	return ""
}

func TestViewCmd_RequiresExactlyOneArg(t *testing.T) {
	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view"})

	if err := root.Execute(); err == nil {
		t.Fatal("expected error when no file argument given")
	}
}
