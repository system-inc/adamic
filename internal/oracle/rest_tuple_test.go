package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	for _, path := range []string{"internal/oracle/testdata/rest_tuple_create.a", "internal/oracle/testdata/rest_tuple_update.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}

func TestRestTupleStorageBoundaries(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, output, reason string }{
		{"rest_tuple_mutate.a", "2\n", "a tuple rest escaping read-only fixed-call storage"},
		{"rest_tuple_escape.a", "2\n", "a tuple rest escaping read-only fixed-call storage"},
		{"rest_tuple_optional.a", "1 2\n", "a rest parameter other than an array"},
		{"rest_tuple_ambiguous.a", "2 2\n", "a rest parameter other than an array"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			source := onNode(t, path)
			if source.exitCode != 0 || string(source.stdout) != probe.output || len(source.stderr) != 0 {
				t.Fatalf("source Node: %+v", source)
			}
			_, err = lowered(t, path)
			var stopped *lower.NotYet
			if !errors.As(err, &stopped) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want tuple storage boundary, got %v", err)
			}
		})
	}
}
