// Package doc defines the document model shared between the YAML source
// files and the Markdown files generated from them.
package doc

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// BlockType identifies the kind of content a Block holds.
type BlockType string

const (
	BlockHeading   BlockType = "heading"
	BlockParagraph BlockType = "paragraph"
	BlockList      BlockType = "list"
	BlockCode      BlockType = "code"
	BlockQuote     BlockType = "quote"
	BlockTable     BlockType = "table"
	BlockHR        BlockType = "hr"
)

// Block is a single unit of Markdown content. Which fields are meaningful
// depends on Type; see the package doc comment in doc.go for the schema.
type Block struct {
	Type     BlockType  `yaml:"type"`
	Level    int        `yaml:"level,omitempty"`
	Text     string     `yaml:"text,omitempty"`
	Ordered  bool       `yaml:"ordered,omitempty"`
	Items    []string   `yaml:"items,omitempty"`
	Language string     `yaml:"language,omitempty"`
	Headers  []string   `yaml:"headers,omitempty"`
	Rows     [][]string `yaml:"rows,omitempty"`
}

// Document is the root object stored in a YAML source file. Title is
// rendered as the document's sole top-level (H1) heading; everything else
// is rendered from Blocks in order.
type Document struct {
	Title  string  `yaml:"title"`
	Blocks []Block `yaml:"blocks"`
}

// Parse decodes a YAML document and validates it.
func Parse(data []byte) (*Document, error) {
	var d Document
	if err := yaml.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}
	if err := d.Validate(); err != nil {
		return nil, err
	}
	return &d, nil
}

// Marshal encodes the document back to YAML.
func (d *Document) Marshal() ([]byte, error) {
	out, err := yaml.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("marshaling yaml: %w", err)
	}
	return out, nil
}

// Validate checks that the document is internally consistent, returning a
// descriptive error identifying the offending block on failure.
func (d *Document) Validate() error {
	for i, b := range d.Blocks {
		if err := b.validate(); err != nil {
			return fmt.Errorf("block %d: %w", i, err)
		}
	}
	return nil
}

func (b *Block) validate() error {
	switch b.Type {
	case BlockHeading:
		if b.Level < 1 || b.Level > 6 {
			return fmt.Errorf("heading level must be between 1 and 6, got %d", b.Level)
		}
		if b.Text == "" {
			return fmt.Errorf("heading requires text")
		}
	case BlockParagraph:
		if b.Text == "" {
			return fmt.Errorf("paragraph requires text")
		}
	case BlockList:
		if len(b.Items) == 0 {
			return fmt.Errorf("list requires at least one item")
		}
	case BlockCode:
		// Empty code blocks are valid (e.g. a placeholder).
	case BlockQuote:
		if b.Text == "" {
			return fmt.Errorf("quote requires text")
		}
	case BlockTable:
		if len(b.Headers) == 0 {
			return fmt.Errorf("table requires headers")
		}
		for i, row := range b.Rows {
			if len(row) != len(b.Headers) {
				return fmt.Errorf("table row %d has %d cells, want %d", i, len(row), len(b.Headers))
			}
		}
	case BlockHR:
		// No fields required.
	default:
		return fmt.Errorf("unknown block type %q", b.Type)
	}
	return nil
}
