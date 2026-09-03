package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
editor for the file and prints its URL; press Ctrl+C to stop serving.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("reading %s: %w", path, err)
			}
			d, err := doc.Parse(data)
			if err != nil {
				return fmt.Errorf("parsing %s: %w", path, err)
			}
			md := mdrender.Render(d)
			out, err := tui.Render(path, d, md, width)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), out)

			if noServe {
				return nil
			}

			base := filepath.Base(path)
			name := strings.TrimSuffix(base, filepath.Ext(base))
			return serveAndBlock(cmd, filepath.Dir(path), port, open, "/edit/"+name)
		},
	}

	c.Flags().IntVar(&width, "width", 0, "wrap width for rendered output (default: 80)")
	c.Flags().IntVar(&port, "port", 0, "port to listen on (default: automatically chosen free port)")
	c.Flags().BoolVar(&open, "open", false, "open the browser editor automatically")
	c.Flags().BoolVar(&noServe, "no-serve", false, "print the terminal preview only; don't start the browser editor")
	return c
}
