package oracle

import (
	"path/filepath"
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
// Not parallel: oracle helpers write the shared os.UserCacheDir()/adamic/gate and adamic/runtime directories.
func TestCensusSmallStoppedSourceOnNode(t *testing.T) {
	for _, probe := range []struct{ file, stdout string }{
		{"census_small_optional_stopped.a", "receiver\nundefined\n"},
		{"census_small_default_stopped.a", "one/two\nmissing\none/[^/]*\n(?:one(?:/[^./])?)?\none/(?:[^./][^/]*)?\n"},
	} {
		// Not parallel: oracle helpers write the shared os.UserCacheDir()/adamic/gate and adamic/runtime directories.
		t.Run(probe.file, func(t *testing.T) {
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
