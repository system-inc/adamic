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
		// The shared fixture list reads a non-lowering fixture as a NotYet stop.
		// string_dictionary now lowers as a record and is held to Node below.
		if name == "string_dictionary" {
			continue
		}
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/notyet_element_access/" + name + ".a", name == "nodearray", false})
	}
}

// Each compiler-area boundary witness is held to Node before pinning its explicit
// representation stop. String dictionaries now have a record representation.
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
			program, err := lowered(t, path)
			if name == "nodearray" {
				if err != nil {
					t.Fatal(err)
				}
				sanitized, binary := nativelyUncached(t, program)
				for mode, result := range map[string]run{"JavaScript": onJavaScriptBackend(t, program), "native sanitized": sanitized, "native release": releasedUncached(t, program)} {
					if difference := disagreement(truth, result); difference != "" {
						t.Fatalf("%s: %s", mode, difference)
					}
				}
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
				return
			}
			if name == "string_dictionary" {
				if err != nil {
					t.Fatal(err)
				}
				actual, binary := natively(t, program)
				for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if difference := disagreement(truth, got); difference != "" {
						t.Fatal(difference)
					}
				}
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
				return
			}
			var stop *lower.NotYet
			want := map[string]string{
				"jsdoc":          "a value of type JSDocArray",
				"numeric_record": "a value of type Record<number, FlowGraphNode>",
				"template":       "a value of type TemplateStringsArray",
				"sorted":         "a value of type SortedReadonlyArray<MappedPosition>",
				"paths":          "a function returning Path | undefined",
			}[name]
			if !errors.As(err, &stop) || stop.What != want {
				t.Fatalf("expected explicit NotYet representation/read boundary %q, got %v", want, err)
			}
			t.Logf("%s", err)
		})
	}
}
