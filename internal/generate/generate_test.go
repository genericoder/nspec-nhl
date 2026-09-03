package generate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validYAML = `
title: "Doc"
blocks:
  - type: paragraph
    text: "hello"
`

const invalidYAML = `
title: "Bad"
blocks:
  - type: paragraph
`

func TestFile_Success(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "a.yaml")
	if err := os.WriteFile(yamlPath, []byte(validYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	res := File(yamlPath, "")
	if res.Err != nil {
		t.Fatalf("File() returned error: %v", res.Err)
	}
	wantMD := filepath.Join(dir, "a.md")
	if res.MDPath != wantMD {
		t.Errorf("MDPath = %q, want %q", res.MDPath, wantMD)
	}
	content, err := os.ReadFile(wantMD)
	if err != nil {
		t.Fatalf("reading generated md: %v", err)
	}
	if !strings.Contains(string(content), "# Doc") {
		t.Errorf("generated md missing title, got: %s", content)
	}
	if !strings.Contains(string(content), "hello") {
		t.Errorf("generated md missing body, got: %s", content)
	}
}

func TestFile_SeparateOutDir(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	yamlPath := filepath.Join(dir, "a.yaml")
	if err := os.WriteFile(yamlPath, []byte(validYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	res := File(yamlPath, outDir)
	if res.Err != nil {
		t.Fatalf("File() returned error: %v", res.Err)
	}
	if _, err := os.Stat(res.MDPath); err != nil {
		t.Errorf("expected md file to exist at %s: %v", res.MDPath, err)
	}
	if filepath.Dir(res.MDPath) != outDir {
		t.Errorf("md written to %s, want dir %s", res.MDPath, outDir)
	}
}

func TestFile_MissingSource(t *testing.T) {
	res := File("/nonexistent/a.yaml", "")
	if res.Err == nil {
		t.Fatal("expected error for missing source file")
	}
}

func TestFile_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(yamlPath, []byte(invalidYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	res := File(yamlPath, "")
	if res.Err == nil {
		t.Fatal("expected error for invalid document")
	}
}

func TestDir_ProcessesOnlyYAMLFiles(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"a.yaml": validYAML,
		"b.yml":  validYAML,
		"c.txt":  "not yaml",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	results, err := Dir(dir, "")
	if err != nil {
		t.Fatalf("Dir() returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2 (a.yaml, b.yml)", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected per-file error for %s: %v", r.YAMLPath, r.Err)
		}
	}
}

func TestDir_ContinuesPastPerFileErrors(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "good.yaml"), []byte(validYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte(invalidYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	results, err := Dir(dir, "")
	if err != nil {
		t.Fatalf("Dir() returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	var sawGood, sawBadErr bool
	for _, r := range results {
		if strings.Contains(r.YAMLPath, "good.yaml") && r.Err == nil {
			sawGood = true
		}
		if strings.Contains(r.YAMLPath, "bad.yaml") && r.Err != nil {
			sawBadErr = true
		}
	}
	if !sawGood {
		t.Error("expected good.yaml to succeed")
	}
	if !sawBadErr {
		t.Error("expected bad.yaml to report an error")
	}
}

func TestDir_MissingDirectory(t *testing.T) {
	if _, err := Dir("/nonexistent/dir", ""); err == nil {
		t.Fatal("expected error for missing source directory")
	}
}

func TestMDPath(t *testing.T) {
	cases := []struct {
		yamlPath, outDir, want string
	}{
		{"yamls/a.yaml", "", "yamls/a.md"},
		{"yamls/a.yml", "", "yamls/a.md"},
		{"yamls/a.yaml", "out", "out/a.md"},
	}
	for _, tc := range cases {
		got := MDPath(tc.yamlPath, tc.outDir)
		want := filepath.FromSlash(tc.want)
		if got != want {
			t.Errorf("MDPath(%q, %q) = %q, want %q", tc.yamlPath, tc.outDir, got, want)
		}
	}
}
