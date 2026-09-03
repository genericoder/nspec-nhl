package cmd

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestServeCmd_ServesUntilContextCanceled(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.yaml"), []byte("title: T\nblocks:\n  - type: paragraph\n    text: hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	root := newRootCmd()
	var out syncBuffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"serve", "--dir", dir, "--port", "0"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- root.ExecuteContext(ctx) }()

	base := waitForURL(t, &out)

	resp, err := http.Get(base)
	if err != nil {
		t.Fatalf("GET %s: %v", base, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET %s: status = %d, want 200", base, resp.StatusCode)
	}

	editResp, err := http.Get(strings.TrimSuffix(base, "/") + "/edit/a")
	if err != nil {
		t.Fatalf("GET edit page: %v", err)
	}
	editResp.Body.Close()
	if editResp.StatusCode != http.StatusOK {
		t.Errorf("GET /edit/a: status = %d, want 200", editResp.StatusCode)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Execute() returned error after cancel: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve command did not shut down after context cancellation")
	}
}

func TestServeCmd_RejectsExtraArgs(t *testing.T) {
	root := newRootCmd()
	var out syncBuffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"serve", "unexpected-arg"})

	if err := root.Execute(); err == nil {
		t.Fatal("expected error for unexpected positional argument")
	}
}
