package tui

import (
	"strings"
	"testing"
	"time"

	"mdgen/internal/doc"
)

func TestSummarize_EmptyDocument(t *testing.T) {
	got := Summarize(&doc.Document{})
	if got != "empty document" {
		t.Errorf("Summarize() = %q, want %q", got, "empty document")
	}
}

func TestSummarize_SingularAndPlural(t *testing.T) {
	d := &doc.Document{Blocks: []doc.Block{
		{Type: doc.BlockHeading}, {Type: doc.BlockHeading},
		{Type: doc.BlockParagraph},
	}}
	got := Summarize(d)
	if !strings.Contains(got, "2 headings") {
		t.Errorf("Summarize() = %q, want it to contain %q", got, "2 headings")
	}
	if !strings.Contains(got, "1 paragraph") || strings.Contains(got, "1 paragraphs") {
		t.Errorf("Summarize() = %q, want singular %q", got, "1 paragraph")
	}
}

func TestSummarize_StableOrder(t *testing.T) {
	d := &doc.Document{Blocks: []doc.Block{
		{Type: doc.BlockHR}, {Type: doc.BlockHeading}, {Type: doc.BlockTable, Headers: []string{"a"}},
	}}
	got := Summarize(d)
	iHeading := strings.Index(got, "heading")
	iTable := strings.Index(got, "table")
	iHR := strings.Index(got, "hr")
	if !(iHeading < iTable && iTable < iHR) {
		t.Errorf("Summarize() = %q, want heading before table before hr", got)
	}
}

func TestRender_ProducesNonEmptyOutput(t *testing.T) {
	d := &doc.Document{Title: "T", Blocks: []doc.Block{{Type: doc.BlockParagraph, Text: "hi"}}}
	out, err := Render("file.yaml", d, "# T\n\nhi\n", 80)
	if err != nil {
		t.Fatalf("Render() returned error: %v", err)
	}
	if !strings.Contains(out, "file.yaml") {
		t.Errorf("Render() output missing path, got:\n%s", out)
	}
	if !strings.Contains(out, "1 paragraph") {
		t.Errorf("Render() output missing summary, got:\n%s", out)
	}
}

func TestRender_DefaultsWidthWhenNonPositive(t *testing.T) {
	d := &doc.Document{}
	if _, err := Render("f.yaml", d, "hi\n", 0); err != nil {
		t.Errorf("Render() with width 0 returned error: %v", err)
	}
	if _, err := Render("f.yaml", d, "hi\n", -5); err != nil {
		t.Errorf("Render() with negative width returned error: %v", err)
	}
}

func TestEditLine_ContainsURL(t *testing.T) {
	got := EditLine("http://localhost:8080/edit/a")
	if !strings.Contains(got, "http://localhost:8080/edit/a") {
		t.Errorf("EditLine() = %q, want it to contain the URL", got)
	}
}

func TestClearScreen(t *testing.T) {
	got := ClearScreen()
	if got != "\x1b[H\x1b[2J" {
		t.Errorf("ClearScreen() = %q, want the cursor-home + clear-screen ANSI sequence", got)
	}
}

func TestUpdatedLine_ContainsTimestamp(t *testing.T) {
	ts := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)
	got := UpdatedLine(ts)
	if !strings.Contains(got, "15:04:05") {
		t.Errorf("UpdatedLine() = %q, want it to contain the formatted time", got)
	}
}
