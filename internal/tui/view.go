// Package tui renders a document as a crisp terminal view: a styled
// header summarizing the source file, followed by syntax-highlighted
// Markdown.
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"mdgen/internal/doc"
)

var (
	pathStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	metaStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	boxStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).Padding(0, 1)
	urlStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Underline(true)
)

// blockOrder fixes a stable, human-meaningful ordering for the block
// summary line regardless of map iteration order.
var blockOrder = []doc.BlockType{
	doc.BlockHeading, doc.BlockParagraph, doc.BlockList,
	doc.BlockCode, doc.BlockQuote, doc.BlockTable, doc.BlockHR,
}

// Render produces the full terminal view: a header box (source path plus
// a block-count summary) followed by the document rendered as
// syntax-highlighted Markdown. width is the target wrap width; if it is
// not positive, 80 is used.
func Render(path string, d *doc.Document, markdown string, width int) (string, error) {
	if width <= 0 {
		width = 80
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return "", fmt.Errorf("creating terminal renderer: %w", err)
	}
	body, err := r.Render(markdown)
	if err != nil {
		return "", fmt.Errorf("rendering markdown: %w", err)
	}

	header := boxStyle.Width(width - 2).Render(
		pathStyle.Render(path) + "\n" + metaStyle.Render(Summarize(d)),
	)

	return header + "\n" + strings.TrimRight(body, "\n") + "\n", nil
}

// EditLine renders the "edit this in your browser" hint printed below the
// view when a browser editor is available for the file.
func EditLine(url string) string {
	return metaStyle.Render("Edit in your browser: ") + urlStyle.Render(url)
}

// ClearScreen returns the ANSI escape sequence that clears the terminal
// and moves the cursor to the top-left, for redrawing a live view in
// place rather than scrolling.
func ClearScreen() string {
	return "\x1b[H\x1b[2J"
}

// UpdatedLine renders the footer shown after a live view redraws in
// response to a browser save.
func UpdatedLine(t time.Time) string {
	return metaStyle.Render(fmt.Sprintf("Updated %s — refreshes automatically when you save in the browser.", t.Format("15:04:05")))
}

// Summarize produces a one-line, human-readable block count summary, e.g.
// "2 headings · 3 paragraphs · 1 table".
func Summarize(d *doc.Document) string {
	counts := make(map[doc.BlockType]int, len(blockOrder))
	for _, b := range d.Blocks {
		counts[b.Type]++
	}

	var parts []string
	for _, t := range blockOrder {
		if n := counts[t]; n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, pluralize(string(t), n)))
		}
	}
	if len(parts) == 0 {
		return "empty document"
	}
	return strings.Join(parts, " · ")
}

func pluralize(word string, n int) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
