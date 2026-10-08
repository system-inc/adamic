package oracle

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"testing"
)

// Node witnesses the five shapes. Existing move restrictions still refuse the
// nested/region shapes: a query result cannot authorize a new graph protocol.
func TestOwnershipQueryNodeShapes(t *testing.T) {
	for _, c := range []struct {
		name, want string
		move       bool
	}{{"flat", "2\n", true}, {"array", "2\n3\n4\n", true}, {"nested", "2\n", false}, {"ring", "2 2\n", false}, {"outside-ring", "2 2\n", false}} {
		t.Run(c.name, func(t *testing.T) {
			path := filepath.Join(repository, "internal/oracle/testdata/ownership-query", c.name+".a")
			path, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			source := onNode(t, path)
			if source.exitCode != 0 || string(source.stdout) != c.want {
				t.Fatalf("Node: %+v", source)
			}
			checked, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.LowerWithOptions(context.Background(), checked, lower.Options{OwnershipQuery: true})
			if !c.move {
				if err == nil {
					t.Fatal("graph runtime gate bypassed")
				}
				t.Logf("Node agrees; runtime gate: %v", err)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			checkParallelVariants(t, program, source)
		})
	}
	// Every currently admitted move fixture also runs with the flag enabled and
	// the existing parallel sanitizer variants, including TSan on Linux.
	paths, err := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata/moves/accepted/*.a"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		path, err = filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Run(filepath.Base(path), func(t *testing.T) {
			checked, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.LowerWithOptions(context.Background(), checked, lower.Options{OwnershipQuery: true})
			if err != nil {
				t.Fatal(err)
			}
			checkParallelVariants(t, program, onNode(t, path))
		})
	}
}
