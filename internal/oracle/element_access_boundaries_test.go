package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

var elementAccessBoundaries = []string{"nodearray", "paths", "template", "sorted", "jsdoc", "numeric_record", "string_dictionary"}

func init() {
	for _, name := range elementAccessBoundaries {
		// The shared fixture list reads a non-lowering fixture as a NotYet stop. string_dictionary
		// is refused (an index signature), which TestElementAccessCompilerAreaBoundaries pins.
		if name == "string_dictionary" {
			continue
		}
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/notyet_element_access/" + name + ".a", false, false})
	}
}

// These are compiler-area boundary witnesses, not runnable backend certificates.
// Each valid source is held to Node before pinning its explicit lowering stop.
func TestElementAccessCompilerAreaBoundaries(t *testing.T) {
	t.Parallel()
	for _, name := range elementAccessBoundaries {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/notyet_element_access/"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != "witness\n" || len(truth.stderr) != 0 {
				t.Fatalf("Node: %#v", truth)
			}
			_, err = lowered(t, path)
			if name == "string_dictionary" {
				var refusal *lower.Refused
				if !errors.As(err, &refusal) || refusal.What != "an index signature" {
					t.Fatalf("expected index-signature ruling boundary, got %v", err)
				}
			} else {
				var stop *lower.NotYet
				want := "an ElementAccessExpression"
				if name == "paths" {
					want = "a function returning Path | undefined"
				}
				if !errors.As(err, &stop) || stop.What != want {
					t.Fatalf("expected explicit NotYet representation/read boundary, got %v", err)
				}
			}
			t.Logf("%s", err)
		})
	}
}
