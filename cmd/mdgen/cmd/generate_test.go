package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateCmd_Success(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "a.yaml")
	if err := os.WriteFile(yamlPath, []byte("title: T\nblocks:\n  - type: paragraph\n    text: hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"generate", "--dir", dir})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v\noutput: %s", err, out.String())
	}
	if !strings.Contains(out.String(), "OK") {
		t.Errorf("expected success output, got: %s", out.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "a.md")); err != nil {
		t.Errorf("expected a.md to be created: %v", err)
	}
}

func TestGenerateCmd_MissingDir(t *testing.T) {
	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"generate", "--dir", "/nonexistent-dir-xyz"})

	if err := root.Execute(); err == nil {
		t.Fatal("expected error for missing directory")
	}
}

func TestGenerateCmd_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"generate", "--dir", dir})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
	if !strings.Contains(out.String(), "no .yaml/.yml files found") {
		t.Errorf("expected empty-dir message, got: %s", out.String())
	}
}

func TestGenerateCmd_PartialFailureReturnsError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "good.yaml"), []byte("title: T\nblocks:\n  - type: paragraph\n    text: hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("title: T\nblocks:\n  - type: paragraph\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"generate", "--dir", dir})

	if err := root.Execute(); err == nil {
		t.Fatal("expected error when one of two files fails to generate")
	}
	if !strings.Contains(out.String(), "FAIL") {
		t.Errorf("expected FAIL line in output, got: %s", out.String())
	}
}

func TestGenerateCmd_OutFlag(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	if err := os.WriteFile(filepath.Join(dir, "a.yaml"), []byte("title: T\nblocks:\n  - type: paragraph\n    text: hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"generate", "--dir", dir, "--out", outDir})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "a.md")); err != nil {
		t.Errorf("expected a.md in --out directory: %v", err)
	}
}
