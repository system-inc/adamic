package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

// These source witnesses are held to Node; their design refusals prevent backend runs.
func TestStatementsSmallRulings(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, stdout, reason string
		refused              bool
	}{
		{"template", ">=1.2.3\n", "a template interpolating an object, an array, a map, a function or undefined", false},
		{"nonnull", "2\n", "the non-null assertion !", true},
		{"capture", "7\n", "a function value that captures the variable its own initializer declares", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, pathErr := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/statements_small_stopped", probe.name+".a"))
			if pathErr != nil {
				t.Fatal(pathErr)
			}
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != probe.stdout || len(node.stderr) != 0 {
				t.Fatalf("source Node: exit %d stdout %q stderr %q", node.exitCode, node.stdout, node.stderr)
			}
			_, err := lowered(t, path)
			var refusal *lower.Refused
			var notYet *lower.NotYet
			rightKind := errors.As(err, &notYet)
			if probe.refused {
				rightKind = errors.As(err, &refusal)
			}
			if !rightKind || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want preserved stop %q, got %v", probe.reason, err)
			}
		})
	}
}
