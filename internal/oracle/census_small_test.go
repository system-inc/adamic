package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/census_small_rest.a", true, false}, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/census_small_boolean.a", true, false}, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/census_small_overload.a", true, false}, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/census_small_stopped/census_small_optional_stopped.a", false, false}, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/census_small_stopped/census_small_default_stopped.a", false, false})
}

// Stopped families still have an independent oracle for the untouched source bodies.
// These programs are not asserted to compile natively.
func TestCensusSmallStoppedSourceOnNode(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, stdout string }{
		{"census_small_optional_stopped.a", "receiver\nundefined\n"},
		{"census_small_default_stopped.a", "one/two\nmissing\none/[^/]*\n(?:one(?:/[^./])?)?\none/(?:[^./][^/]*)?\n"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/census_small_stopped", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			result := onNode(t, path)
			if result.exitCode != 0 || string(result.stdout) != probe.stdout || len(result.stderr) != 0 {
				t.Fatalf("original source on Node: %+v; want stdout %q", result, probe.stdout)
			}
		})
	}
}

// This witness must stop at its registered construct, with the same reason.
func TestCensusSmallDynamicReceiverRefusalPin(t *testing.T) {
	t.Parallel()
	path := "internal/oracle/testdata/census_small_stopped/census_small_optional_stopped.a"
	absolute, err := filepath.Abs(filepath.Join(repository, path))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, absolute)
	const reason = "a function expression with a this parameter (dynamic receivers are not implemented)"
	const location = "census_small_optional_stopped.a:11:"
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), reason) || !strings.Contains(notYet.Where, location) {
		t.Fatalf("want located refusal %q at %s: %v", reason, location, err)
	}
	t.Logf("%v", err)
}
