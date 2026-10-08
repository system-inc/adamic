package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, name := range []string{"binder_432", "binder_1296"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path: "internal/oracle/testdata/for_of_object_stopped/" + name + ".a"})
	}
}

// NodeArray carries array elements and metadata. Until both have a native representation,
// these source reductions must stop before either backend receives a usable program.
func TestForOfObjectNodeArrayStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, stdout string }{
		{"binder_432", "2\n"},
		{"binder_1296", "true\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/for_of_object_stopped", probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			observed := onNode(t, path)
			if observed.exitCode != 0 || string(observed.stdout) != probe.stdout || len(observed.stderr) != 0 {
				t.Fatalf("Node: %+v; want stdout %q", observed, probe.stdout)
			}
			program, err := lowered(t, path)
			var stop *lower.NotYet
			if program != nil || !errors.As(err, &stop) || stop.What != "for...of over an object" {
				t.Fatalf("want the NodeArray representation stop and no IR, got program %v, error %v", program != nil, err)
			}
		})
	}
}
