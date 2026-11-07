package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

var elementAccessBoundaries = []string{
	"nodearray",
	"paths",
	"template",
	"sorted",
	"jsdoc",
	"numeric_record",
	"string_dictionary",
}

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
		}{
			"internal/oracle/testdata/notyet_element_access/" + name + ".a",
			name == "nodearray" || name == "sorted",
			false,
		})
	}
}

// Each compiler-area boundary witness is held to Node before pinning its explicit
// representation stop. String dictionaries now have a record representation, and tsc's
// SortedReadonlyArray, its brand typed undefined as adaptation 40 types it, is held as its
// base ReadonlyArray (#gb0fqaa), so it lowers and is held to Node as nodearray is.
func TestElementAccessCompilerAreaBoundaries(test *testing.T) {
	test.Parallel()
	for _, name := range elementAccessBoundaries {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/notyet_element_access/"+name+".a"),
			)
			if err != nil {
				test.Fatal(err)
			}
			truth := onNode(test, path)
			if truth.exitCode != 0 || string(truth.stdout) != "witness\n" || len(truth.stderr) != 0 {
				test.Fatalf("Node: %#v", truth)
			}
			program, err := lowered(test, path)
			if name == "nodearray" || name == "sorted" {
				if err != nil {
					test.Fatal(err)
				}
				sanitized, binary := nativelyUncached(test, program)
				for mode, result := range map[string]run{
					"JavaScript":       onJavaScriptBackend(test, program),
					"native sanitized": sanitized,
					"native release":   releasedUncached(test, program),
				} {
					if difference := disagreement(truth, result); difference != "" {
						test.Fatalf("%s: %s", mode, difference)
					}
				}
				if report := leaksUncached(test, program, binary); report != "" {
					test.Fatal(report)
				}
				return
			}
			if name == "string_dictionary" {
				if err != nil {
					test.Fatal(err)
				}
				actual, binary := natively(test, program)
				for _, got := range []run{actual, releasedUncached(test, program), onJavaScriptBackend(test, program)} {
					if difference := disagreement(truth, got); difference != "" {
						test.Fatal(difference)
					}
				}
				if report := leaks(test, program, binary); report != "" {
					test.Fatal(report)
				}
				return
			}
			var stop *lower.NotYet
			// Phantom primitive brands represent paths' return type.
			want := map[string]string{
				"jsdoc":          "a value of type JSDocArray",
				"numeric_record": "a value of type Record<number, FlowGraphNode>",
				"template":       "a value of type TemplateStringsArray",
				"paths":          "a value of type Readonly<PathPathComponents>",
			}[name]
			if !errors.As(err, &stop) || stop.What != want {
				test.Fatalf("expected explicit NotYet representation/read boundary %q, got %v", want, err)
			}
			test.Logf("%s", err)
		})
	}
}
