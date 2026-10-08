package oracle

import (
	"bytes"
	"path/filepath"
	"testing"
)

// Preserve all eight previously reported abort locations against source Node.
// Current integration also closes the earlier symbols array-write frontier.
func TestShapeGraphReportedFrontiers(t *testing.T) {
	for _, name := range []string{"cache", "entries", "literals", "regression_06", "regression_07", "regression_08", "regression_09", "symbols"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/graph_regions_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			reference := onNode(t, path)
			if reference.exitCode != 0 {
				t.Fatalf("Node: %+v", reference)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			sanitized, binary := nativelyUncached(t, program)
			runs := []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)}
			for _, actual := range runs {
				if actual.exitCode != reference.exitCode || !bytes.Equal(actual.stdout, reference.stdout) || len(actual.stderr) != 0 {
					t.Fatalf("disagreement: Node=%+v backend=%+v", reference, actual)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}
