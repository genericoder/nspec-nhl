package mdparse

import (
	"reflect"
	"testing"

	"mdgen/internal/doc"
	"mdgen/internal/mdrender"
)

func TestParse_TitleFromFirstH1(t *testing.T) {
	d := Parse([]byte("# My Title\n\nSome text.\n"))
	if d.Title != "My Title" {
		t.Errorf("Title = %q, want %q", d.Title, "My Title")
	}
	if len(d.Blocks) != 1 || d.Blocks[0].Type != doc.BlockParagraph || d.Blocks[0].Text != "Some text." {
		t.Errorf("Blocks = %+v, want single paragraph %q", d.Blocks, "Some text.")
	}
}

func TestParse_NoLeadingH1MeansNoTitle(t *testing.T) {
	d := Parse([]byte("Some text.\n"))
	if d.Title != "" {
		t.Errorf("Title = %q, want empty", d.Title)
	}
	if len(d.Blocks) != 1 || d.Blocks[0].Text != "Some text." {
		t.Errorf("Blocks = %+v", d.Blocks)
	}
}

func TestParse_SecondH1IsHeadingBlockNotTitle(t *testing.T) {
	d := Parse([]byte("# First\n\nbody\n\n# Second\n"))
	if d.Title != "First" {
		t.Errorf("Title = %q, want %q", d.Title, "First")
	}
	if len(d.Blocks) != 2 {
		t.Fatalf("len(Blocks) = %d, want 2", len(d.Blocks))
	}
	if d.Blocks[1].Type != doc.BlockHeading || d.Blocks[1].Level != 1 || d.Blocks[1].Text != "Second" {
		t.Errorf("Blocks[1] = %+v, want heading level 1 %q", d.Blocks[1], "Second")
	}
}

func TestParse_Heading(t *testing.T) {
	d := Parse([]byte("## Section Two\n"))
	want := doc.Block{Type: doc.BlockHeading, Level: 2, Text: "Section Two"}
	if len(d.Blocks) != 1 || !reflect.DeepEqual(d.Blocks[0], want) {
		t.Errorf("Blocks[0] = %+v, want %+v", d.Blocks[0], want)
	}
}

func TestParse_ParagraphSoftWrapJoinsWithSpace(t *testing.T) {
	d := Parse([]byte("line one\nline two\n"))
	want := "line one line two"
	if len(d.Blocks) != 1 || d.Blocks[0].Text != want {
		t.Errorf("Blocks[0].Text = %q, want %q", d.Blocks[0].Text, want)
	}
}

func TestParse_ListUnordered(t *testing.T) {
	d := Parse([]byte("- one\n- two\n- three\n"))
	if len(d.Blocks) != 1 {
		t.Fatalf("len(Blocks) = %d, want 1", len(d.Blocks))
	}
	b := d.Blocks[0]
	if b.Type != doc.BlockList || b.Ordered {
		t.Fatalf("Blocks[0] = %+v, want unordered list", b)
	}
	if !reflect.DeepEqual(b.Items, []string{"one", "two", "three"}) {
		t.Errorf("Items = %#v", b.Items)
	}
}

func TestParse_ListOrdered(t *testing.T) {
	d := Parse([]byte("1. one\n2. two\n"))
	b := d.Blocks[0]
	if !b.Ordered {
		t.Fatalf("Blocks[0] = %+v, want ordered list", b)
	}
	if !reflect.DeepEqual(b.Items, []string{"one", "two"}) {
		t.Errorf("Items = %#v", b.Items)
	}
}

func TestParse_FencedCodeBlock(t *testing.T) {
	d := Parse([]byte("```go\nfmt.Println(1)\nfmt.Println(2)\n```\n"))
	b := d.Blocks[0]
	if b.Type != doc.BlockCode || b.Language != "go" {
		t.Fatalf("Blocks[0] = %+v, want code block lang go", b)
	}
	want := "fmt.Println(1)\nfmt.Println(2)"
	if b.Text != want {
		t.Errorf("Text = %q, want %q", b.Text, want)
	}
}

func TestParse_FencedCodeBlockNoLanguage(t *testing.T) {
	d := Parse([]byte("```\nplain\n```\n"))
	b := d.Blocks[0]
	if b.Language != "" {
		t.Errorf("Language = %q, want empty", b.Language)
	}
	if b.Text != "plain" {
		t.Errorf("Text = %q, want %q", b.Text, "plain")
	}
}

func TestParse_Blockquote(t *testing.T) {
	d := Parse([]byte("> quoted text\n"))
	b := d.Blocks[0]
	if b.Type != doc.BlockQuote || b.Text != "quoted text" {
		t.Errorf("Blocks[0] = %+v, want quote %q", b, "quoted text")
	}
}

func TestParse_MultilineBlockquote(t *testing.T) {
	d := Parse([]byte("> line one\n> line two\n"))
	b := d.Blocks[0]
	if b.Type != doc.BlockQuote {
		t.Fatalf("Blocks[0].Type = %v, want quote", b.Type)
	}
	want := "line one line two"
	if b.Text != want {
		t.Errorf("Text = %q, want %q", b.Text, want)
	}
}

func TestParse_Table(t *testing.T) {
	md := "| A | B |\n| --- | --- |\n| 1 | 2 |\n| 3 | 4 |\n"
	d := Parse([]byte(md))
	if len(d.Blocks) != 1 {
		t.Fatalf("len(Blocks) = %d, want 1", len(d.Blocks))
	}
	b := d.Blocks[0]
	if b.Type != doc.BlockTable {
		t.Fatalf("Blocks[0].Type = %v, want table", b.Type)
	}
	if !reflect.DeepEqual(b.Headers, []string{"A", "B"}) {
		t.Errorf("Headers = %#v", b.Headers)
	}
	wantRows := [][]string{{"1", "2"}, {"3", "4"}}
	if !reflect.DeepEqual(b.Rows, wantRows) {
		t.Errorf("Rows = %#v, want %#v", b.Rows, wantRows)
	}
}

func TestParse_ThematicBreak(t *testing.T) {
	d := Parse([]byte("---\n"))
	if len(d.Blocks) != 1 || d.Blocks[0].Type != doc.BlockHR {
		t.Errorf("Blocks = %+v, want single hr block", d.Blocks)
	}
}

func TestParse_UnsupportedNodeIsSkipped(t *testing.T) {
	d := Parse([]byte("<div>raw html</div>\n\nfollowed by a paragraph\n"))
	if len(d.Blocks) != 1 || d.Blocks[0].Text != "followed by a paragraph" {
		t.Errorf("Blocks = %+v, want raw html skipped and paragraph kept", d.Blocks)
	}
}

func TestParse_EmptyInput(t *testing.T) {
	d := Parse([]byte(""))
	if d.Title != "" || len(d.Blocks) != 0 {
		t.Errorf("Parse(\"\") = %+v, want zero-value document", d)
	}
}

func TestRoundTrip_RenderThenParseReproducesDocument(t *testing.T) {
	original := &doc.Document{
		Title: "Round Trip Doc",
		Blocks: []doc.Block{
			{Type: doc.BlockHeading, Level: 2, Text: "Intro"},
			{Type: doc.BlockParagraph, Text: "Some paragraph text."},
			{Type: doc.BlockList, Ordered: false, Items: []string{"a", "b", "c"}},
			{Type: doc.BlockList, Ordered: true, Items: []string{"first", "second"}},
			{Type: doc.BlockCode, Language: "go", Text: "fmt.Println(1)"},
			{Type: doc.BlockQuote, Text: "a quote"},
			{Type: doc.BlockTable, Headers: []string{"A", "B"}, Rows: [][]string{{"1", "2"}, {"3", "4"}}},
			{Type: doc.BlockHR},
		},
	}

	md := mdrender.Render(original)
	got := Parse([]byte(md))

	if got.Title != original.Title {
		t.Errorf("Title = %q, want %q", got.Title, original.Title)
	}
	if !reflect.DeepEqual(got.Blocks, original.Blocks) {
		t.Errorf("round-tripped blocks differ.\ngot:  %+v\nwant: %+v\nmarkdown was:\n%s", got.Blocks, original.Blocks, md)
	}
}

func TestParse_ResultAlwaysValidates(t *testing.T) {
	d := Parse([]byte("# T\n\nHi\n\n- a\n- b\n\n```go\ncode\n```\n\n> q\n\n| A |\n| --- |\n| 1 |\n\n---\n"))
	if err := d.Validate(); err != nil {
		t.Errorf("parsed document failed validation: %v\ndoc: %+v", err, d)
	}
}
