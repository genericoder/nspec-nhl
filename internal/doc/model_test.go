package doc

import (
	"strings"
	"testing"
)

func TestParse_Valid(t *testing.T) {
	src := `
title: "Sample"
blocks:
  - type: heading
    level: 2
    text: "Intro"
  - type: paragraph
    text: "Hello world"
  - type: list
    ordered: true
    items: ["a", "b"]
  - type: code
    language: go
    text: "fmt.Println(1)"
  - type: quote
    text: "quoted"
  - type: table
    headers: ["A", "B"]
    rows:
      - ["1", "2"]
  - type: hr
`
	d, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if d.Title != "Sample" {
		t.Errorf("Title = %q, want %q", d.Title, "Sample")
	}
	if len(d.Blocks) != 7 {
		t.Fatalf("len(Blocks) = %d, want 7", len(d.Blocks))
	}
	if d.Blocks[0].Type != BlockHeading || d.Blocks[0].Level != 2 {
		t.Errorf("Blocks[0] = %+v, want heading level 2", d.Blocks[0])
	}
}

func TestParse_InvalidYAML(t *testing.T) {
	_, err := Parse([]byte("title: [unterminated"))
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
}

func TestParse_EmptyDocument(t *testing.T) {
	d, err := Parse([]byte(""))
	if err != nil {
		t.Fatalf("Parse returned error for empty doc: %v", err)
	}
	if d.Title != "" || len(d.Blocks) != 0 {
		t.Errorf("expected zero-value document, got %+v", d)
	}
}

func TestValidate_TableCases(t *testing.T) {
	cases := []struct {
		name    string
		block   Block
		wantErr bool
	}{
		{"valid table", Block{Type: BlockTable, Headers: []string{"A", "B"}, Rows: [][]string{{"1", "2"}}}, false},
		{"missing headers", Block{Type: BlockTable, Rows: [][]string{{"1", "2"}}}, true},
		{"row length mismatch", Block{Type: BlockTable, Headers: []string{"A", "B"}, Rows: [][]string{{"1"}}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := &Document{Blocks: []Block{tc.block}}
			err := d.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidate_HeadingLevelBounds(t *testing.T) {
	for _, level := range []int{0, 7, -1} {
		d := &Document{Blocks: []Block{{Type: BlockHeading, Level: level, Text: "x"}}}
		if err := d.Validate(); err == nil {
			t.Errorf("Validate() with level %d: expected error, got nil", level)
		}
	}
	for _, level := range []int{1, 3, 6} {
		d := &Document{Blocks: []Block{{Type: BlockHeading, Level: level, Text: "x"}}}
		if err := d.Validate(); err != nil {
			t.Errorf("Validate() with level %d: unexpected error: %v", level, err)
		}
	}
}

func TestValidate_RequiredText(t *testing.T) {
	types := []BlockType{BlockParagraph, BlockQuote}
	for _, bt := range types {
		d := &Document{Blocks: []Block{{Type: bt}}}
		if err := d.Validate(); err == nil {
			t.Errorf("Validate() for %s with empty text: expected error, got nil", bt)
		}
	}
}

func TestValidate_ListRequiresItems(t *testing.T) {
	d := &Document{Blocks: []Block{{Type: BlockList}}}
	if err := d.Validate(); err == nil {
		t.Fatal("expected error for list with no items")
	}
}

func TestValidate_UnknownBlockType(t *testing.T) {
	d := &Document{Blocks: []Block{{Type: "bogus"}}}
	err := d.Validate()
	if err == nil {
		t.Fatal("expected error for unknown block type")
	}
	if !strings.Contains(err.Error(), "unknown block type") {
		t.Errorf("error = %v, want it to mention unknown block type", err)
	}
}

func TestValidate_HRAndEmptyCodeAreValid(t *testing.T) {
	d := &Document{Blocks: []Block{
		{Type: BlockHR},
		{Type: BlockCode, Language: "go"},
	}}
	if err := d.Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

func TestMarshal_RoundTrip(t *testing.T) {
	d := &Document{
		Title: "Round Trip",
		Blocks: []Block{
			{Type: BlockHeading, Level: 1, Text: "H"},
			{Type: BlockList, Ordered: false, Items: []string{"x", "y"}},
			{Type: BlockTable, Headers: []string{"A"}, Rows: [][]string{{"1"}}},
		},
	}
	out, err := d.Marshal()
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	d2, err := Parse(out)
	if err != nil {
		t.Fatalf("re-parsing marshaled output failed: %v", err)
	}
	if d2.Title != d.Title {
		t.Errorf("Title = %q, want %q", d2.Title, d.Title)
	}
	if len(d2.Blocks) != len(d.Blocks) {
		t.Fatalf("len(Blocks) = %d, want %d", len(d2.Blocks), len(d.Blocks))
	}
	for i := range d.Blocks {
		if d2.Blocks[i].Type != d.Blocks[i].Type {
			t.Errorf("Blocks[%d].Type = %q, want %q", i, d2.Blocks[i].Type, d.Blocks[i].Type)
		}
	}
}

func TestMarshal_InvalidDocumentStillMarshals(t *testing.T) {
	// Marshal itself does not validate; that is Parse's job on the way in.
	d := &Document{Blocks: []Block{{Type: "bogus"}}}
	if _, err := d.Marshal(); err != nil {
		t.Errorf("Marshal returned unexpected error: %v", err)
	}
}
