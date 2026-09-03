package mdrender

import (
	"strings"
	"testing"

	"mdgen/internal/doc"
)

func TestRender_TitleOnly(t *testing.T) {
	got := Render(&doc.Document{Title: "Hello"})
	want := "# Hello\n"
	if got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
}

func TestRender_NoTitle(t *testing.T) {
	got := Render(&doc.Document{Blocks: []doc.Block{{Type: doc.BlockParagraph, Text: "hi"}}})
	want := "hi\n"
	if got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
}

func TestRender_Heading(t *testing.T) {
	got := renderBlock(doc.Block{Type: doc.BlockHeading, Level: 3, Text: "Section"})
	if got != "### Section" {
		t.Errorf("renderBlock(heading) = %q, want %q", got, "### Section")
	}
}

func TestRender_ListUnordered(t *testing.T) {
	got := renderBlock(doc.Block{Type: doc.BlockList, Items: []string{"a", "b"}})
	want := "- a\n- b"
	if got != want {
		t.Errorf("renderBlock(list) = %q, want %q", got, want)
	}
}

func TestRender_ListOrdered(t *testing.T) {
	got := renderBlock(doc.Block{Type: doc.BlockList, Ordered: true, Items: []string{"a", "b", "c"}})
	want := "1. a\n2. b\n3. c"
	if got != want {
		t.Errorf("renderBlock(ordered list) = %q, want %q", got, want)
	}
}

func TestRender_Code(t *testing.T) {
	got := renderBlock(doc.Block{Type: doc.BlockCode, Language: "go", Text: "fmt.Println(1)\n"})
	want := "```go\nfmt.Println(1)\n```"
	if got != want {
		t.Errorf("renderBlock(code) = %q, want %q", got, want)
	}
}

func TestRender_CodeNoLanguage(t *testing.T) {
	got := renderBlock(doc.Block{Type: doc.BlockCode, Text: "plain"})
	want := "```\nplain\n```"
	if got != want {
		t.Errorf("renderBlock(code) = %q, want %q", got, want)
	}
}

func TestRender_Quote(t *testing.T) {
	got := renderBlock(doc.Block{Type: doc.BlockQuote, Text: "line1\nline2"})
	want := "> line1\n> line2"
	if got != want {
		t.Errorf("renderBlock(quote) = %q, want %q", got, want)
	}
}

func TestRender_Table(t *testing.T) {
	got := renderBlock(doc.Block{
		Type:    doc.BlockTable,
		Headers: []string{"A", "B"},
		Rows:    [][]string{{"1", "2"}, {"3", "4"}},
	})
	want := "| A | B |\n| --- | --- |\n| 1 | 2 |\n| 3 | 4 |"
	if got != want {
		t.Errorf("renderBlock(table) = %q, want %q", got, want)
	}
}

func TestRender_HR(t *testing.T) {
	if got := renderBlock(doc.Block{Type: doc.BlockHR}); got != "---" {
		t.Errorf("renderBlock(hr) = %q, want %q", got, "---")
	}
}

func TestRender_UnknownTypeRendersEmpty(t *testing.T) {
	if got := renderBlock(doc.Block{Type: "bogus"}); got != "" {
		t.Errorf("renderBlock(bogus) = %q, want empty string", got)
	}
}

func TestRender_FullDocument(t *testing.T) {
	d := &doc.Document{
		Title: "Doc",
		Blocks: []doc.Block{
			{Type: doc.BlockHeading, Level: 2, Text: "Intro"},
			{Type: doc.BlockParagraph, Text: "Body text."},
			{Type: doc.BlockList, Items: []string{"x"}},
			{Type: doc.BlockHR},
		},
	}
	got := Render(d)
	for _, want := range []string{"# Doc", "## Intro", "Body text.", "- x", "---"} {
		if !strings.Contains(got, want) {
			t.Errorf("Render() missing %q in output:\n%s", want, got)
		}
	}
	// Blocks are separated by a blank line.
	if !strings.Contains(got, "# Doc\n\n## Intro") {
		t.Errorf("Render() blocks not blank-line separated:\n%s", got)
	}
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("Render() output should end with a single trailing newline")
	}
}
