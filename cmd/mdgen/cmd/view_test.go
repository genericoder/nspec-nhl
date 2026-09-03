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

func TestViewCmd_LiveUpdatesTerminalOnBrowserSave(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "a.yaml")
	content := "title: My Doc\nblocks:\n  - type: paragraph\n    text: original text\n"
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

	done := make(chan error, 1)
	go func() { done <- root.ExecuteContext(ctx) }()

	editURL := waitForURL(t, &out)
	if !strings.Contains(out.String(), "original text") {
		t.Fatalf("expected initial render to contain original text, got:\n%s", out.String())
	}

	saveURL := strings.TrimSuffix(editURL, "/edit/a") + "/api/save/a"
	newMD := "# My Doc\n\nupdated live text\n"
	resp, err := http.Post(saveURL, "text/plain", strings.NewReader(newMD))
	if err != nil {
		t.Fatalf("POST %s: %v", saveURL, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST %s: status = %d, want 200", saveURL, resp.StatusCode)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) && !strings.Contains(out.String(), "updated live text") {
		time.Sleep(10 * time.Millisecond)
	}
	final := out.String()
	if !strings.Contains(final, "updated live text") {
		t.Fatalf("expected terminal output to refresh with saved content, got:\n%s", final)
	}
	// The refreshed render should have replaced the on-screen view (via
	// a clear-screen escape), not merely appended to old output.
	if !strings.Contains(final, "\x1b[H\x1b[2J") {
		t.Errorf("expected output to contain a clear-screen escape sequence on refresh")
	}

	// The yaml file itself should also reflect the save.
	yamlBytes, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("reading yaml after save: %v", err)
	}
	if !strings.Contains(string(yamlBytes), "updated live text") {
		t.Errorf("expected yaml file to contain saved text, got:\n%s", yamlBytes)
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

func TestViewCmd_DoesNotRefreshForUnrelatedFile(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "a.yaml")
	if err := os.WriteFile(yamlPath, []byte("title: T\nblocks:\n  - type: paragraph\n    text: original\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.yaml"), []byte("title: B\nblocks:\n  - type: paragraph\n    text: other doc\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	root := newRootCmd()
	var out syncBuffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", yamlPath, "--port", "0"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- root.ExecuteContext(ctx) }()

	editURL := waitForURL(t, &out)
	saveURL := strings.TrimSuffix(editURL, "/edit/a") + "/api/save/b"

	resp, err := http.Post(saveURL, "text/plain", strings.NewReader("# B\n\nbrand new content for b\n"))
	if err != nil {
		t.Fatalf("POST %s: %v", saveURL, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST %s: status = %d, want 200", saveURL, resp.StatusCode)
	}

	// Give any (incorrect) refresh a moment to happen, then confirm it
	// didn't: the terminal should still show only the original content
	// for a.yaml, with no clear-screen redraw triggered.
	time.Sleep(200 * time.Millisecond)
	final := out.String()
	if strings.Contains(final, "brand new content for b") {
		t.Errorf("view of a.yaml should not refresh when b.yaml is saved, got:\n%s", final)
	}
	if strings.Contains(final, "\x1b[H\x1b[2J") {
		t.Errorf("no clear-screen redraw should occur for an unrelated file's save")
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
