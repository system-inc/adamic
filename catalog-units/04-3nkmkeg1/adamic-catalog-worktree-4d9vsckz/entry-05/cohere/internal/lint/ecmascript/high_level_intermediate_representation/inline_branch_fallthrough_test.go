package high_level_intermediate_representation

import (
	"strings"
	"testing"
)

func TestCopyNestedBodyRemapsBranchFallthrough(t *testing.T) {
	t.Parallel()
	parent, nested, captures := loweredParentAndNested(t, `
function outer(value) {
if (value) visit(value);
const callback = () => value ? {value} : {};
return callback;
}`)
	remap, ok := CopyNestedBodyInto(parent, nested, captures)
	if !ok {
		t.Fatal("copy declined")
	}
	branches := 0
	for _, block := range nested.Blocks {
		branch, ok := block.Terminal.(*Branch)
		if !ok || !HasBlock(branch.Fallthrough) {
			continue
		}
		branches++
		copiedBlock, found := parent.Block(remap.Blocks[block.Id])
		if !found {
			t.Fatal("copied branch block is missing")
		}
		copied, ok := copiedBlock.Terminal.(*Branch)
		if !ok {
			t.Fatal("copied terminal is not a branch")
		}
		want := remap.Blocks[branch.Fallthrough]
		if want == branch.Fallthrough {
			t.Fatal("fixture's block IDs coincide, so a skipped remap cannot be detected")
		}
		if copied.Fallthrough != want {
			t.Errorf("copied fallthrough=%d, want %d; nested fallthrough=%d", copied.Fallthrough, want, branch.Fallthrough)
		}
	}
	if branches != 1 {
		t.Fatalf("checked %d branches, want one", branches)
	}
}

func TestConditionalManualMemosKeepSeparateDependencies(t *testing.T) {
	t.Parallel()
	const source = `import {useMemo} from 'react'; import {useData, useMounted} from './hooks';
function Component() {
const [first] = useData('first', []);
const [second] = useData('second', []);
const mounted = useMounted();
const firstSet = useMemo(() => mounted ? new Set(first) : new Set(), [mounted, first]);
const secondSet = useMemo(() => mounted ? new Set(second) : new Set(), [mounted, second]);
return <div first={firstSet} second={secondSet}/>;
}`
	for _, testCase := range []struct {
		name   string
		source string
		fires  bool
	}{
		{"conditional", source, false},
		{"array allocation", strings.ReplaceAll(strings.ReplaceAll(source, "new Set(first)", "[first]"), "new Set(second)", "[second]"), false},
		{"unconditional", strings.ReplaceAll(strings.ReplaceAll(source, "mounted ? new Set(first) : new Set()", "new Set(first)"), "mounted ? new Set(second) : new Set()", "new Set(second)"), false},
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
