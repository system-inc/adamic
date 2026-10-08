package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/for_in_options_refused.a", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/for_in_array_refused.a", false, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/for_in_runtime.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/for_in_static.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/for_in_primitives.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/for_in_storage_name.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/for_in_array_properties.a", true, false})
}

// A missing property and a present undefined property have different key sets.
// Runtime enumeration preserves actual property presence through structural views.
func TestForInOptionsRuling(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/for_in_options_refused.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != "stable\nstable|missing\n" {
		t.Fatalf("Node property presence witness: %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	native, binary := natively(t, program)
	backend := onJavaScriptBackend(t, program)
	if difference := disagreement(truth, native); difference != "" {
		t.Fatal(difference)
	}
	if difference := disagreement(truth, backend); difference != "" {
		t.Fatal(difference)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}

}

// for-in copies present indices, including undefined values, but skips holes.
func TestForInArrayRuling(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/for_in_array_refused.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != "0|1|2\n0|2\n" {
		t.Fatalf("Node array presence witness: %+v", truth)
	}
	_, err = lowered(t, path)
	var stop *lower.NotYet
	const reason = "an OmittedExpression in an array literal"
	if !errors.As(err, &stop) || stop.What != reason {
		t.Fatalf("want exact array stop %q, got %v", reason, err)
	}
	t.Logf("Node %q; retained %v", truth.stdout, stop)
}
