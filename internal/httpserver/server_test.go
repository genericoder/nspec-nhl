package httpserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mdgen/internal/doc"
)

const sampleYAML = `title: "Sample"
blocks:
  - type: paragraph
    text: "hello"
`

func newTestServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.yaml"), []byte(sampleYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	srv := New(dir)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts, dir
}

func TestHandleIndex_ListsYAMLFiles(t *testing.T) {
	ts, dir := newTestServer(t)
	if err := os.WriteFile(filepath.Join(dir, "b.yml"), []byte(sampleYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignore me"), 0o644); err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body := readAll(t, resp)
	if !strings.Contains(body, "/edit/a") {
		t.Errorf("index missing link to a, body:\n%s", body)
	}
	if !strings.Contains(body, "/edit/b") {
		t.Errorf("index missing link to b, body:\n%s", body)
	}
	if strings.Contains(body, "notes") {
		t.Errorf("index should not list non-yaml files, body:\n%s", body)
	}
}

func TestHandleEditPage_ExistingFile(t *testing.T) {
	ts, _ := newTestServer(t)
	resp, err := http.Get(ts.URL + "/edit/a")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body := readAll(t, resp)
	if !strings.Contains(body, "hello") {
		t.Errorf("edit page missing rendered markdown content, body:\n%s", body)
	}
	if !strings.Contains(body, "# Sample") {
		t.Errorf("edit page missing title heading, body:\n%s", body)
	}
}

func TestHandleEditPage_PrefersExistingMDFile(t *testing.T) {
	ts, dir := newTestServer(t)
	// Simulate a previously-generated/edited .md that differs from what
	// would be freshly rendered from the yaml.
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("# Sample\n\nEDITED CONTENT\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get(ts.URL + "/edit/a")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body := readAll(t, resp)
	if !strings.Contains(body, "EDITED CONTENT") {
		t.Errorf("expected edit page to prefer existing .md content, body:\n%s", body)
	}
}

func TestHandleEditPage_MissingFile(t *testing.T) {
	ts, _ := newTestServer(t)
	resp, err := http.Get(ts.URL + "/edit/nope")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestHandleEditPage_RejectsPathTraversal(t *testing.T) {
	ts, _ := newTestServer(t)
	// Plain ".." is normalized away by the HTTP client/ServeMux before it
	// ever reaches our handler as name=".." (it resolves to "/"), so the
	// meaningful cases here are ones smuggled past that normalization via
	// percent-encoded separators.
	for _, name := range []string{"..%2Fa", "a%2Fb", "a%2F..%2F..%2Fetc"} {
		resp, err := http.Get(ts.URL + "/edit/" + name)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			t.Errorf("name %q: expected non-200 status, got 200", name)
		}
	}
}

func TestHandleSave_WritesMDAndYAML(t *testing.T) {
	ts, dir := newTestServer(t)

	newMD := "# Sample\n\n## New Section\n\nUpdated body text.\n\n- one\n- two\n"
	resp, err := http.Post(ts.URL+"/api/save/a", "text/plain", strings.NewReader(newMD))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", resp.StatusCode, readAll(t, resp))
	}

	mdBytes, err := os.ReadFile(filepath.Join(dir, "a.md"))
	if err != nil {
		t.Fatalf("reading saved .md: %v", err)
	}
	if string(mdBytes) != newMD {
		t.Errorf(".md content = %q, want %q", mdBytes, newMD)
	}

	yamlBytes, err := os.ReadFile(filepath.Join(dir, "a.yaml"))
	if err != nil {
		t.Fatalf("reading saved .yaml: %v", err)
	}
	d, err := doc.Parse(yamlBytes)
	if err != nil {
		t.Fatalf("parsing saved yaml: %v", err)
	}
	if d.Title != "Sample" {
		t.Errorf("Title = %q, want %q", d.Title, "Sample")
	}
	if len(d.Blocks) != 3 {
		t.Fatalf("len(Blocks) = %d, want 3 (heading, paragraph, list), got %+v", len(d.Blocks), d.Blocks)
	}
	if d.Blocks[0].Type != doc.BlockHeading || d.Blocks[0].Text != "New Section" {
		t.Errorf("Blocks[0] = %+v", d.Blocks[0])
	}
	if d.Blocks[2].Type != doc.BlockList || len(d.Blocks[2].Items) != 2 {
		t.Errorf("Blocks[2] = %+v", d.Blocks[2])
	}
}

func TestHandleSave_MissingDocument(t *testing.T) {
	ts, _ := newTestServer(t)
	resp, err := http.Post(ts.URL+"/api/save/nope", "text/plain", strings.NewReader("# x\n"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestHandleSave_RejectsPathTraversal(t *testing.T) {
	ts, dir := newTestServer(t)
	resp, err := http.Post(ts.URL+"/api/save/a%2F..%2F..%2Fpwned", "text/plain", strings.NewReader("# x\n"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatal("expected traversal attempt to be rejected")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(dir), "pwned.yaml")); err == nil {
		t.Fatal("traversal attempt wrote a file outside the served directory")
	}
}

func TestHandlePreview_RendersHTML(t *testing.T) {
	ts, _ := newTestServer(t)
	resp, err := http.Post(ts.URL+"/api/preview", "text/plain", strings.NewReader("# Hi\n\n- one\n- two\n"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body := readAll(t, resp)
	if !strings.Contains(body, "<h1") || !strings.Contains(body, "<li>") {
		t.Errorf("expected rendered HTML with h1/li tags, got:\n%s", body)
	}
}

func TestStaticAssets_AreServed(t *testing.T) {
	ts, _ := newTestServer(t)
	for _, path := range []string{"/static/app.js", "/static/style.css"} {
		resp, err := http.Get(ts.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s: status = %d, want 200", path, resp.StatusCode)
		}
	}
}

func readAll(t *testing.T, resp *http.Response) string {
	t.Helper()
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if err != nil {
			break
		}
	}
	return string(buf)
}
