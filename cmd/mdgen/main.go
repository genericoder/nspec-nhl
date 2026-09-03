// Command mdgen generates Markdown files from YAML sources and lets you
// view and edit them.
package main

import (
	"os"

	"mdgen/cmd/mdgen/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
