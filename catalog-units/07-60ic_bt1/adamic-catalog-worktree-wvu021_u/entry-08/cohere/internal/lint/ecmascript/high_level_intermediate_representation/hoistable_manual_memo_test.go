package high_level_intermediate_representation

import (
	"strings"
	"testing"
)

func TestManualMemoDependenciesProveOnlyNonOptionalPrefixes(t *testing.T) {
	t.Parallel()
	const source = `import {useState, useMemo} from 'react';
function usePreview(options) {
const [selected] = useState(null);
const result = useMemo(() => {
if (selected === null) return null;
return options.rows.find(row => row.id === selected) ?? null;
}, [options.rows, selected]);
return {result};
}`
	for _, testCase := range []struct {
		name   string
		source string
		fires  bool
	}{
		{"non-optional dependency", source, false},
		{"optional dependency", strings.Replace(source, "[options.rows, selected]", "[options?.rows, selected]", 1), true},
		{"unconditional read", strings.Replace(source, "if (selected === null) return null;", "", 1), false},
		{"unpreserved dependency", `import {useMemo} from 'react'; function Component({propA}) {return useMemo(() => propA.x(), [propA.x]);}`, true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			findings, lowered := findingsForSource(t, testCase.source)
			if !lowered {
				t.Fatal("fixture did not lower")
			}
			if (len(findings) != 0) != testCase.fires {
				t.Fatalf("findings=%v, want fires=%t", findings, testCase.fires)
			}
		})
	}
}
