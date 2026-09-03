package cmd

import (
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var dir string
	var port int
	var open bool

	c := &cobra.Command{
		Use:   "serve",
		Short: "Serve a browser editor for a directory of YAML/Markdown documents",
		Long: `Serve starts a local, loopback-only HTTP server with an editor page for
every .yaml/.yml file in --dir. Saving a document in the browser writes
both its .md and .yaml files. Press Ctrl+C to stop.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return serveAndBlock(cmd, dir, port, open, "/")
		},
	}

	c.Flags().StringVar(&dir, "dir", "yamls", "directory containing .yaml/.yml source files")
	c.Flags().IntVar(&port, "port", 0, "port to listen on (default: automatically chosen free port)")
	c.Flags().BoolVar(&open, "open", false, "open the index page in your default browser")
	return c
}
