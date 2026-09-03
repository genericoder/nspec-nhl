// Package cmd defines mdgen's cobra command tree.
package cmd

import (
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "mdgen",
		Short: "Generate, view, and edit Markdown from YAML sources",
		Long: `mdgen turns YAML documents into Markdown files, lets you preview
them in the terminal, and serves a browser editor that keeps the
Markdown and YAML in sync.`,
		SilenceUsage: true,
	}
	root.AddCommand(newGenerateCmd())
	root.AddCommand(newViewCmd())
	root.AddCommand(newServeCmd())
	return root
}

// Execute runs the mdgen root command.
func Execute() error {
	return newRootCmd().Execute()
}
