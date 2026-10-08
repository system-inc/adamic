package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, name := range []string{"scanner_array_notyet", "scanner_typed_array_notyet"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{
			"internal/oracle/testdata/new_expression_" + name + ".a", false, false,
		})
	}
}

// Source Node pins the scanner witness. Typing its elements exposes the separate
// sparse-array requirement instead of silently treating length as one element.
func TestNewExpressionScannerArrayDependencies(t *testing.T) {
	for _, fixture := range []struct{ name, stop string }{
		{"scanner_array_notyet", "an array of any"},
		{"scanner_typed_array_notyet", "an Array constructor creating holes"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/new_expression_"+fixture.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != "2\n" || len(node.stderr) != 0 {
				t.Fatalf("Node scanner Array: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
			}
			_, err = lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || notYet.What != fixture.stop {
				t.Fatalf("want %s NotYet, got %v", fixture.stop, err)
			}
		})
	}
}
