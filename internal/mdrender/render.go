// Package mdrender renders a doc.Document into Markdown text.
package mdrender

import (
	"fmt"
	"strconv"
	"strings"

	"mdgen/internal/doc"
)

// Render converts a document into a Markdown string. The title, if set,
// becomes the document's only H1; all blocks follow it in order, each
// separated by a blank line.
func Render(d *doc.Document) string {
	var parts []string
	if d.Title != "" {
		parts = append(parts, "# "+d.Title)
	}
	for _, b := range d.Blocks {
		parts = append(parts, renderBlock(b))
	}
	return strings.Join(parts, "\n\n") + "\n"
}

func renderBlock(b doc.Block) string {
	switch b.Type {
	case doc.BlockHeading:
		return strings.Repeat("#", b.Level) + " " + b.Text
	case doc.BlockParagraph:
		return b.Text
	case doc.BlockList:
		return renderList(b)
	case doc.BlockCode:
		return renderCode(b)
	case doc.BlockQuote:
		return renderQuote(b)
	case doc.BlockTable:
		return renderTable(b)
	case doc.BlockHR:
		return "---"
	default:
		// Validated documents never reach here; render nothing rather
		// than panic on data that slipped past validation.
		return ""
	}
}

func renderList(b doc.Block) string {
	lines := make([]string, len(b.Items))
	for i, item := range b.Items {
		if b.Ordered {
			lines[i] = strconv.Itoa(i+1) + ". " + item
		} else {
			lines[i] = "- " + item
		}
	}
	return strings.Join(lines, "\n")
}

func renderCode(b doc.Block) string {
	text := strings.TrimSuffix(b.Text, "\n")
	return "```" + b.Language + "\n" + text + "\n```"
}

func renderQuote(b doc.Block) string {
	lines := strings.Split(b.Text, "\n")
	for i, l := range lines {
		lines[i] = "> " + l
	}
	return strings.Join(lines, "\n")
}

func renderTable(b doc.Block) string {
	var sb strings.Builder
	writeRow(&sb, b.Headers)
	sep := make([]string, len(b.Headers))
	for i := range sep {
		sep[i] = "---"
	}
	writeRow(&sb, sep)
	for _, row := range b.Rows {
		writeRow(&sb, row)
	}
	return strings.TrimSuffix(sb.String(), "\n")
}

func writeRow(sb *strings.Builder, cells []string) {
	fmt.Fprintf(sb, "| %s |\n", strings.Join(cells, " | "))
}
