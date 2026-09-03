package browseropen

import (
	"reflect"
	"testing"
)

func TestCommandFor(t *testing.T) {
	cases := []struct {
		goos     string
		wantName string
		wantArgs []string
	}{
		{"darwin", "open", []string{"http://localhost:8080/"}},
		{"windows", "rundll32", []string{"url.dll,FileProtocolHandler", "http://localhost:8080/"}},
		{"linux", "xdg-open", []string{"http://localhost:8080/"}},
		{"freebsd", "xdg-open", []string{"http://localhost:8080/"}},
	}
	for _, tc := range cases {
		t.Run(tc.goos, func(t *testing.T) {
			name, args := commandFor(tc.goos, "http://localhost:8080/")
			if name != tc.wantName {
				t.Errorf("commandFor(%q) name = %q, want %q", tc.goos, name, tc.wantName)
			}
			if !reflect.DeepEqual(args, tc.wantArgs) {
				t.Errorf("commandFor(%q) args = %#v, want %#v", tc.goos, args, tc.wantArgs)
			}
		})
	}
}
