package lower

import (
	"strings"
	"testing"
)

func TestLibraryTscCensusCompilerBoundaries(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, source, reason string }{
		{"optional find", "const declarations: {kind:number}[] | undefined = [{kind:1}]; console.log(`${declarations?.find(d => d.kind === 1)?.kind}`);", "a call through ?."},
		{"spread push", "const values = new Set([1,2]); const target = [0]; target.push(...values, 3);", "compiler iterator argument dispatch"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, test.source)
			if err == nil || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("want named compiler boundary %q, got %v", test.reason, err)
			}
		})
	}
}
