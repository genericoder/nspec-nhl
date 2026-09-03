package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mdgen/internal/generate"
)

func newGenerateCmd() *cobra.Command {
	var dir, out string

	c := &cobra.Command{
		Use:   "generate",
		Short: "Generate Markdown files from a directory of YAML sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := generate.Dir(dir, out)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "no .yaml/.yml files found in %s\n", dir)
				return nil
			}

			var failed int
			for _, r := range results {
				if r.Err != nil {
					failed++
					fmt.Fprintf(cmd.ErrOrStderr(), "FAIL %s: %v\n", r.YAMLPath, r.Err)
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "OK   %s -> %s\n", r.YAMLPath, r.MDPath)
			}
			if failed > 0 {
				return fmt.Errorf("%d of %d file(s) failed to generate", failed, len(results))
			}
			return nil
		},
	}

	c.Flags().StringVar(&dir, "dir", "yamls", "directory containing .yaml/.yml source files")
	c.Flags().StringVar(&out, "out", "", "output directory for generated .md files (default: alongside each source file)")
	return c
}
