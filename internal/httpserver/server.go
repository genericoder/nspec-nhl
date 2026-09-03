// Package httpserver serves a minimal browser editor for the YAML/Markdown
// documents in a directory. Saving in the browser writes both the .md and
// the .yaml file, keeping them in sync with the filesystem.
package httpserver

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"

	"mdgen/internal/doc"
	"mdgen/internal/generate"
	"mdgen/internal/mdparse"
	"mdgen/internal/mdrender"
)

//go:embed assets/edit.html.tmpl assets/index.html.tmpl
var templateFS embed.FS

//go:embed assets/app.js assets/style.css
var staticFS embed.FS

var (
	editTmpl  = template.Must(template.ParseFS(templateFS, "assets/edit.html.tmpl"))
	indexTmpl = template.Must(template.ParseFS(templateFS, "assets/index.html.tmpl"))

	previewMD = goldmark.New(goldmark.WithExtensions(extension.GFM))

	// validName matches the document names accepted in URLs: no path
	// separators or traversal sequences, just a plain file base name.
	validName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
)

// Server serves the browser editor for every .yaml/.yml file directly
// inside Dir.
type Server struct {
	Dir string
}

// New returns a Server rooted at dir.
func New(dir string) *Server {
	return &Server{Dir: dir}
}

// Handler builds the server's HTTP routes.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /edit/{name}", s.handleEditPage)
	mux.HandleFunc("POST /api/save/{name}", s.handleSave)
	mux.HandleFunc("POST /api/preview", s.handlePreview)

	static, err := fs.Sub(staticFS, "assets")
	if err != nil {
		panic(err) // embedded FS is compiled in; this cannot fail at runtime
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static)))
	return mux
}

// yamlPath resolves a URL document name to its source file on disk,
// rejecting anything that isn't a plain, existing file name.
func (s *Server) yamlPath(name string) (string, error) {
	if !validName.MatchString(name) {
		return "", fmt.Errorf("invalid document name %q", name)
	}
	for _, ext := range []string{".yaml", ".yml"} {
		p := filepath.Join(s.Dir, name+ext)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no such document %q", name)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())))
	}
	sort.Strings(names)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := indexTmpl.Execute(w, names); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleEditPage(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	yamlPath, err := s.yamlPath(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	content, err := currentMarkdown(yamlPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Name    string
		Content string
	}{Name: name, Content: content}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := editTmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// currentMarkdown returns the document's current Markdown: the .md file
// if one already exists next to yamlPath, otherwise freshly rendered
// from the YAML.
func currentMarkdown(yamlPath string) (string, error) {
	mdPath := generate.MDPath(yamlPath, "")
	if b, err := os.ReadFile(mdPath); err == nil {
		return string(b), nil
	}
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return "", err
	}
	d, err := doc.Parse(data)
	if err != nil {
		return "", err
	}
	return mdrender.Render(d), nil
}

func (s *Server) handleSave(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	yamlPath, err := s.yamlPath(name)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	mdPath := generate.MDPath(yamlPath, "")
	if err := os.WriteFile(mdPath, body, 0o644); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	d := mdparse.Parse(body)
	yamlBytes, err := d.Marshal()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	if err := os.WriteFile(yamlPath, yamlBytes, 0o644); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handlePreview(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var buf bytes.Buffer
	if err := previewMD.Convert(body, &buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"ok": false, "error": err.Error()})
}
