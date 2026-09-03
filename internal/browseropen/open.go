// Package browseropen opens a URL in the user's default browser.
package browseropen

import (
	"os/exec"
	"runtime"
)

// Open launches the default browser at url.
func Open(url string) error {
	name, args := commandFor(runtime.GOOS, url)
	return exec.Command(name, args...).Start()
}

// commandFor returns the OS command and arguments used to open url,
// split out from Open so the mapping can be tested without actually
// launching a process.
func commandFor(goos, url string) (string, []string) {
	switch goos {
	case "darwin":
		return "open", []string{url}
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", url}
	default:
		return "xdg-open", []string{url}
	}
}
