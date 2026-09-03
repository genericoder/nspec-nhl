package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"mdgen/internal/doc"
	"mdgen/internal/mdrender"
	"mdgen/internal/tui"
)

func newViewCmd() *cobra.Command {
	var width int
	var port int
	var open bool
	var noServe bool

	c := &cobra.Command{
		Use:   "view <file.yaml>",
		Short: "Preview a YAML document's rendered Markdown in the terminal",
		Long: `View renders a YAML document as syntax-highlighted Markdown in the
terminal. Unless --no-serve is given, it also starts a local browser
editor for the file and prints its URL; the terminal view refreshes
automatically whenever the file is saved in the browser. Press Ctrl+C
to stop serving.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			out, err := renderYAMLFile(path, width)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), out)

			if noServe {
				return nil
			}

			base := filepath.Base(path)
			name := strings.TrimSuffix(base, filepath.Ext(base))

			viewedPath := path
			if abs, err := filepath.Abs(path); err == nil {
				viewedPath = abs
			}

			onSaved := func(savedYAMLPath, url string) {
				savedAbs := savedYAMLPath
				if abs, err := filepath.Abs(savedYAMLPath); err == nil {
					savedAbs = abs
				}
				if savedAbs != viewedPath {
					return
				}

				refreshed, err := renderYAMLFile(path, width)
				w := cmd.OutOrStdout()
				if err != nil {
					// Leave the previous view on screen; report the
					// problem without wiping out something readable.
					fmt.Fprintf(cmd.ErrOrStderr(), "error re-rendering %s after save: %v\n", path, err)
					return
				}
				fmt.Fprint(w, tui.ClearScreen())
				fmt.Fprint(w, refreshed)
				fmt.Fprintln(w, tui.EditLine(url))
				fmt.Fprintln(w, tui.UpdatedLine(time.Now()))
			}

			return serveAndBlock(cmd, filepath.Dir(path), port, open, "/edit/"+name, onSaved)
		},
	}

	c.Flags().IntVar(&width, "width", 0, "wrap width for rendered output (default: 80)")
	c.Flags().IntVar(&port, "port", 0, "port to listen on (default: automatically chosen free port)")
	c.Flags().BoolVar(&open, "open", false, "open the browser editor automatically")
	c.Flags().BoolVar(&noServe, "no-serve", false, "print the terminal preview only; don't start the browser editor")
	return c
}

// renderYAMLFile reads, parses, and renders a YAML document for terminal
// display. It's used both for the initial view and for live redraws
// triggered by a browser save.
func renderYAMLFile(path string, width int) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	d, err := doc.Parse(data)
	if err != nil {
		return "", fmt.Errorf("parsing %s: %w", path, err)
	}
	md := mdrender.Render(d)
	return tui.Render(path, d, md, width)
}
