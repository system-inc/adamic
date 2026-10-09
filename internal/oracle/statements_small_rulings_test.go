package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
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
		{"nonnull", "2\n", "the non-null assertion !", true},
		{"structural_error", "false\n", "throwing an Error that isn't made where it's thrown or caught by the catch around it", false},
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
			program, err := lowered(t, path)
			if err == nil {
				native, _ := natively(t, program)
				backend := onJavaScriptBackend(t, program)
				t.Fatalf("want preserved stop %q; admitted program: native %s; JavaScript %s", probe.reason, disagreement(node, native), disagreement(node, backend))
			}
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

// Proven ToPrimitive lowering closed this template gap; the actual bytes, not
// admission alone, must match Node on every backend.
func TestStatementsSmallTemplateAgreesWithNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/statements_small_stopped/template.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != ">=1.2.3\n" {
		t.Fatalf("source Node: %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	actual, binary := natively(t, program)
	if difference := disagreement(truth, actual); difference != "" {
		t.Fatalf("native: %s", difference)
	}
	if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatalf("JavaScript: %s", difference)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		if difference := disagreement(truth, onWASI(t, native.C(program))); difference != "" {
			t.Fatalf("WASI: %s", difference)
		}
	}
	t.Logf("Node and compiled template agree: %q", truth.stdout)
	changed := false
	for index, text := range program.Strings {
		if text == ">=" {
			program.Strings[index] = "<="
			changed = true
		}
	}
	if !changed {
		t.Fatal("template mutant changed no input")
	}
	scout22MutantMatchesOnlyNode(t, path, program)
}
