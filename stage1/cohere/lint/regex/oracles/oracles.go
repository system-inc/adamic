// Package oracles owns the compiler-free Go oracle products for this port.
package oracles

import (
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"path/filepath"
	"testing"
)

var producer bool

func Program(t *testing.T, file, output string) string {
	t.Helper()
	path, err := filepath.Abs(file)
	if err != nil {
		t.Fatal(err)
	}
	return stock(t, output+"-owner-v1", path, nil)
}

// stock is the identical stock request in owners and finders; only the owner builds.
func stock(t testing.TB, output, pkg string, arguments []string, environment ...string) string {
	t.Helper()
	if producer {
		return buildcache.GoBuildIn(t, "cohere", output, pkg, arguments, environment...)
	}
	inputs, err := buildcache.GoBuildInInputs("cohere", output, pkg, arguments, nil, environment...)
	if err != nil {
		t.Fatal(err)
	}
	directory := buildcache.BuiltElsewhere(t, "Go oracle", inputs, "stage1/cohere/lint/regex/oracles owners", func(string) error { return fmt.Errorf("settle the oracle owner first") })
	return filepath.Join(directory, output)
}
