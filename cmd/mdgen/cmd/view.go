package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"mdgen/internal/doc"
	"mdgen/internal/mdrender"
	"mdgen/internal/tui"
)

func newViewCmd() *cobra.Command {
	var width int

	c := &cobra.Command{
		Use:   "view <file.yaml>",
		Short: "Preview a YAML document's rendered Markdown in the terminal",
		Args:  cobra.ExactArgs(1),
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
			return nil
		},
	}

	c.Flags().IntVar(&width, "width", 0, "wrap width for rendered output (default: 80)")
	return c
}
