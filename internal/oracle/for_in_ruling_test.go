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
	}{"internal/oracle/testdata/for_in_options_refused.a", false, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/for_in_array_refused.a", false, false})
}

// A missing property and a present undefined property have different key sets.
// These witnesses stay stopped until property presence has a sound representation.
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
	_, err = lowered(t, path)
	var stop *lower.NotYet
	const reason = "for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly)"
	if !errors.As(err, &stop) || stop.What != reason {
		t.Fatalf("want exact origin stop %q, got %v", reason, err)
	}
	t.Logf("Node %q; retained %v", truth.stdout, stop)
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
	const reason = "for...in over an array (holes and own enumerable properties are not represented; use for...of for elements)"
	if !errors.As(err, &stop) || stop.What != reason {
		t.Fatalf("want exact array stop %q, got %v", reason, err)
	}
	t.Logf("Node %q; retained %v", truth.stdout, stop)
}
