package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	root.SetArgs([]string{"view", yamlPath})

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
	root.SetArgs([]string{"view", yamlPath})

	if err := root.Execute(); err == nil {
		t.Fatal("expected error for invalid document")
	}
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
