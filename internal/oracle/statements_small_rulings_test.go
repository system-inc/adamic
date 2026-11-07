package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

// These source witnesses are held to Node; their design refusals prevent backend runs.
func TestStatementsSmallRulings(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct {
		name, stdout, reason string
		refused              bool
	}{
		{
			"template",
			">=1.2.3\n",
			"a template interpolating an object, an array, a map, a function or undefined",
			false,
		},
		{"nonnull", "2\n", "the non-null assertion !", true},
		{
			"structural_error",
			"false\n",
			"throwing an Error that isn't made where it's thrown or caught by the catch around it",
			false,
		},
		{"capture", "7\n", "a function value that captures the variable its own initializer declares", false},
	} {
		test.Run(probe.name, func(test *testing.T) {
			path, pathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/statements_small_stopped", probe.name+".a"),
			)
			if pathError != nil {
				test.Fatal(pathError)
			}
			node := onNode(test, path)
			if node.exitCode != 0 || string(node.stdout) != probe.stdout || len(node.stderr) != 0 {
				test.Fatalf("source Node: exit %d stdout %q stderr %q", node.exitCode, node.stdout, node.stderr)
			}
			program, loweredError := lowered(test, path)
			if loweredError == nil {
				native, _ := natively(test, program)
				backend := onJavaScriptBackend(test, program)
				test.Fatalf(
					"want preserved stop %q; admitted program: native %s; JavaScript %s",
					probe.reason,
					disagreement(node, native),
					disagreement(node, backend),
				)
			}
			var refusal *lower.Refused
			var notYet *lower.NotYet
			rightKind := errors.As(loweredError, &notYet)
			if probe.refused {
				rightKind = errors.As(loweredError, &refusal)
			}
			if !rightKind || !strings.Contains(loweredError.Error(), probe.reason) {
				test.Fatalf("want preserved stop %q, got %v", probe.reason, loweredError)
			}
		})
	}
}
