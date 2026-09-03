// Package mdparse converts Markdown text back into a doc.Document by
// walking goldmark's AST. It is the inverse of mdrender: the browser
// editor posts raw Markdown, and this package turns it back into the
// block structure that gets written to YAML.
package mdparse

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	gtext "github.com/yuin/goldmark/text"

	"mdgen/internal/doc"
)

var md = goldmark.New(goldmark.WithExtensions(extension.GFM))

// Parse converts raw Markdown source into a Document.
//
// The first top-level node, if it is an H1 heading, becomes the
// document's Title; every other top-level node becomes a Block. Markdown
// constructs mdgen doesn't model (raw HTML, nested blockquotes, images,
// etc.) are skipped rather than causing a failure, so arbitrary
// hand-edited Markdown always parses to *something* rather than erroring.
func Parse(source []byte) *doc.Document {
	reader := gtext.NewReader(source)
	root := md.Parser().Parse(reader)

	d := &doc.Document{}
	first := true
	for n := root.FirstChild(); n != nil; n = n.NextSibling() {
		if first {
			first = false
			if h, ok := n.(*ast.Heading); ok && h.Level == 1 {
				d.Title = strings.TrimSpace(linesText(n, source))
				continue
			}
		}
		if b, ok := convertBlock(n, source); ok {
			d.Blocks = append(d.Blocks, b)
		}
	}
	return d
}

func convertBlock(n ast.Node, source []byte) (doc.Block, bool) {
	switch node := n.(type) {
	case *ast.Heading:
		return doc.Block{Type: doc.BlockHeading, Level: node.Level, Text: strings.TrimSpace(linesText(node, source))}, true
	case *ast.Paragraph:
		return doc.Block{Type: doc.BlockParagraph, Text: strings.TrimSpace(linesText(node, source))}, true
	case *ast.List:
		var items []string
		for item := node.FirstChild(); item != nil; item = item.NextSibling() {
			items = append(items, strings.TrimSpace(blockText(item, source)))
		}
		return doc.Block{Type: doc.BlockList, Ordered: node.IsOrdered(), Items: items}, true
	case *ast.FencedCodeBlock:
		lang := ""
		if node.Info != nil {
			lang = strings.TrimSpace(string(node.Info.Text(source)))
		}
		return doc.Block{Type: doc.BlockCode, Language: lang, Text: linesTextRaw(node, source)}, true
	case *ast.CodeBlock:
		return doc.Block{Type: doc.BlockCode, Text: linesTextRaw(node, source)}, true
	case *ast.Blockquote:
		return doc.Block{Type: doc.BlockQuote, Text: strings.TrimSpace(blockText(node, source))}, true
	case *east.Table:
		return convertTable(node, source), true
	case *ast.ThematicBreak:
		return doc.Block{Type: doc.BlockHR}, true
	default:
		return doc.Block{}, false
	}
}

func convertTable(t *east.Table, source []byte) doc.Block {
	b := doc.Block{Type: doc.BlockTable}
	for n := t.FirstChild(); n != nil; n = n.NextSibling() {
		switch row := n.(type) {
		case *east.TableHeader:
			b.Headers = tableRowCells(row, source)
		case *east.TableRow:
			b.Rows = append(b.Rows, tableRowCells(row, source))
		}
	}
	return b
}

func tableRowCells(row ast.Node, source []byte) []string {
	var cells []string
	for c := row.FirstChild(); c != nil; c = c.NextSibling() {
		cells = append(cells, strings.TrimSpace(inlineText(c, source)))
	}
	return cells
}

// linesNode is implemented by goldmark block nodes (Paragraph, Heading,
// CodeBlock, FencedCodeBlock, ...) that expose their raw source lines.
type linesNode interface {
	Lines() *gtext.Segments
}

// linesText joins a node's raw source lines with a single space,
// reconstituting soft-wrapped Markdown as one logical line of text.
func linesText(n ast.Node, source []byte) string {
	ln, ok := n.(linesNode)
	if !ok {
		return blockText(n, source)
	}
	lines := ln.Lines()
	parts := make([]string, 0, lines.Len())
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		parts = append(parts, strings.TrimRight(string(seg.Value(source)), "\n"))
	}
	return strings.Join(parts, " ")
}

// linesTextRaw joins a node's raw source lines with newlines preserved,
// for content (code blocks) where whitespace is significant.
func linesTextRaw(n ast.Node, source []byte) string {
	ln, ok := n.(linesNode)
	if !ok {
		return ""
	}
	lines := ln.Lines()
	var sb strings.Builder
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		sb.Write(seg.Value(source))
	}
	return strings.TrimSuffix(sb.String(), "\n")
}

// blockText concatenates the text of a node's child blocks, one per line.
// It is the fallback for nodes (list items, blockquotes) whose content
// lives in child blocks rather than in their own Lines().
func blockText(n ast.Node, source []byte) string {
	var parts []string
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		parts = append(parts, linesText(c, source))
	}
	return strings.Join(parts, "\n")
}

// inlineText concatenates the literal text of a node's inline
// descendants (used for table cells, whose content is inline nodes
// rather than Lines()).
func inlineText(n ast.Node, source []byte) string {
	var sb strings.Builder
	var walk func(ast.Node)
	walk = func(n ast.Node) {
		if n == nil {
			return
		}
		if t, ok := n.(*ast.Text); ok {
			sb.Write(t.Value(source))
		}
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			walk(c)
		}
	}
	walk(n)
	return sb.String()
}
