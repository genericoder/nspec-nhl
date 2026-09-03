package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"mdgen/internal/browseropen"
	"mdgen/internal/httpserver"
	"mdgen/internal/tui"
)

// serveAndBlock starts the browser editor server for dir, prints its URL
// (pointing at path, or the index if path is empty), optionally opens it
// in the default browser, and blocks until the command's context is
// canceled (e.g. by Ctrl+C), then shuts down gracefully.
//
// If onSaved is non-nil, it is called (with the saved .yaml path and the
// server's base URL) synchronously whenever a browser save completes, so
// a caller can refresh a live view of the affected document.
func serveAndBlock(cmd *cobra.Command, dir string, port int, openBrowser bool, path string, onSaved func(yamlPath, url string)) error {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fmt.Errorf("starting listener: %w", err)
	}

	url := fmt.Sprintf("http://%s%s", ln.Addr().String(), path)
	fmt.Fprintln(cmd.OutOrStdout(), tui.EditLine(url))

	if openBrowser {
		if err := browseropen.Open(url); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "could not open browser automatically: %v\n", err)
		}
	}

	srv := httpserver.New(dir)
	if onSaved != nil {
		srv.OnSaved = func(yamlPath string) { onSaved(yamlPath, url) }
	}
	httpSrv := &http.Server{Handler: srv.Handler()}

	ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() { errCh <- httpSrv.Serve(ln) }()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return httpSrv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	}
}
