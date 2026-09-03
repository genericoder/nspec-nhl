// Package generate turns a directory of YAML documents into Markdown files.
package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mdgen/internal/doc"
	"mdgen/internal/mdrender"
)

// Result records the outcome of generating one file.
type Result struct {
	YAMLPath string
	MDPath   string
	Err      error
}

// Dir generates a .md file next to (or under outDir, if non-empty) every
// .yaml/.yml file directly inside srcDir. It does not recurse into
// subdirectories. Processing continues past per-file errors; callers should
// inspect each Result's Err.
func Dir(srcDir, outDir string) ([]Result, error) {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", srcDir, err)
	}

	var results []Result
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		yamlPath := filepath.Join(srcDir, e.Name())
		results = append(results, File(yamlPath, outDir))
	}
	return results, nil
}

// File generates the .md file for a single YAML file. If outDir is empty,
// the .md file is written alongside the source file; otherwise it is
// written to outDir using the source file's base name.
func File(yamlPath, outDir string) Result {
	res := Result{YAMLPath: yamlPath}

	data, err := os.ReadFile(yamlPath)
	if err != nil {
		res.Err = fmt.Errorf("reading %s: %w", yamlPath, err)
		return res
	}

	d, err := doc.Parse(data)
	if err != nil {
		res.Err = fmt.Errorf("parsing %s: %w", yamlPath, err)
		return res
	}

	mdPath := MDPath(yamlPath, outDir)
	res.MDPath = mdPath

	if err := os.MkdirAll(filepath.Dir(mdPath), 0o755); err != nil {
		res.Err = fmt.Errorf("creating output directory for %s: %w", mdPath, err)
		return res
	}

	md := mdrender.Render(d)
	if err := os.WriteFile(mdPath, []byte(md), 0o644); err != nil {
		res.Err = fmt.Errorf("writing %s: %w", mdPath, err)
		return res
	}
	return res
}

// MDPath computes the output .md path for a given YAML source path. If
// outDir is empty, the result sits alongside yamlPath with the extension
// swapped; otherwise it is base(yamlPath) with a .md extension inside
// outDir.
func MDPath(yamlPath, outDir string) string {
	base := filepath.Base(yamlPath)
	base = strings.TrimSuffix(base, filepath.Ext(base)) + ".md"
	if outDir == "" {
		return filepath.Join(filepath.Dir(yamlPath), base)
	}
	return filepath.Join(outDir, base)
}
