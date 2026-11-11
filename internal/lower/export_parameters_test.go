package lower

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// The entry module's exported functions are called by outside code (the JavaScript backend's
// ModuleExports, wasm32's handleRequest) with anything, so the flow graph holds every parameter of
// one unknown, whatever the program's own calls pass it. export_parameters.mutants.md names the edit
// that turns it back.

// exportedKeep is a program whose keep(box) the program calls once with a literal; prefix is written
// before its function keyword.
func exportedKeep(prefix string) string {
	return "interface Box { tag: string }\n" + prefix + `function keep(box: Box): Box {
	return box;
}
const kept = keep({ tag: 'inside' });
console.log(kept.tag);
`
}

// exportedKeptReach lowers source as main.ts and asks the flow graph which allocations the
// module-level kept can hold, and returns that with the allocation site of the literal tagged inside.
func exportedKeptReach(t *testing.T, source string) (ir.AllocationSet, int) {
	t.Helper()
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatalf("lowering: %v", err)
	}
	assignAllocationSites(program)
	kept := -1
	for index, declared := range program.Locals {
		if declared.Name == "kept" && declared.Global {
			kept = index
		}
	}
	if kept < 0 {
		t.Fatal("no module-level variable is named kept")
	}
	sites := []int{}
	visit := func(node any) bool {
		if literal, isLiteral := node.(ir.ObjectLiteral); isLiteral && len(literal.GraphTypes) > 0 {
			for _, field := range literal.Fields {
				if field.Name != "tag" {
					continue
				}
				walk(field.Value, func(inner any) bool {
					if constant, isString := inner.(ir.StringConstant); isString && program.Strings[constant.Index] == "inside" {
						sites = append(sites, literal.GraphTypes[len(literal.GraphTypes)-1])
					}
					return true
				})
			}
		}
		return true
	}
	walk(program.Main, visit)
	for _, function := range program.Functions {
		walk(function.Body, visit)
	}
	if len(sites) != 1 || sites[0] >= 0 {
		t.Fatalf("the literal tagged inside has sites %v, want one numbered site", sites)
	}
	return newAllocationFlowGraph(program).ReachingAllocations(ir.Read{Local: kept, Of: program.Locals[kept].Type}), sites[0]
}

// Not exported, keep's parameter is the one literal the program passes it; exported from the entry
// module, it's unknown, though the program's own call is the same. Mutant 1.
func TestExportedEntryFunctionParameterIsUnknown(t *testing.T) {
	t.Parallel()
	t.Run("not exported", func(t *testing.T) {
		t.Parallel()
		reach, inside := exportedKeptReach(t, exportedKeep(""))
		if reach.Unknown || !slices.Equal(reach.Sites, []int{inside}) {
			t.Fatalf("reach is %+v, want exactly the site %d", reach, inside)
		}
	})
	t.Run("exported from the entry module", func(t *testing.T) {
		t.Parallel()
		reach, _ := exportedKeptReach(t, exportedKeep("export "))
		if !reach.Unknown || !slices.Contains(reach.Reasons, "parameter may receive untracked or omitted arguments") {
			t.Fatalf("reach is %+v, want Unknown for a parameter that may receive untracked arguments", reach)
		}
	})
}

// The gap was live. An Iterable<number> parameter the program hands only a Set's values() lowers as
// that iterator's own loop, stepping it by next() (iterable_views.go, libraryIteratorLoop), proven by
// ReachingAllocations to be handed nothing else. Exported from the entry module, main lowered it the
// same way, since the program's one call was all the graph saw; outside code calling largest([1, 2, 3])
// on the JavaScript backend's ModuleExports would have had items.next called on an array, a TypeError
// where Node prints 3. It is NotYet now, and the same function not exported still lowers. Mutant 1.
func TestExportedIterableViewParameterIsNotProven(t *testing.T) {
	t.Parallel()
	const program = `function largest(items: Iterable<number>): number {
    let best = -1;
    for(const value of items) {
        if(value > best) {
            best = value;
        }
    }
    return best;
}
const sizes = new Set<number>();
sizes.add(3);
sizes.add(9);
console.log(String(largest(sizes.values())));
`
	t.Run("not exported", func(t *testing.T) {
		t.Parallel()
		if _, err := lowerSource(t, program); err != nil {
			t.Fatalf("a view handed only a Set's values(): %v", err)
		}
	})
	t.Run("exported from the entry module", func(t *testing.T) {
		t.Parallel()
		_, err := lowerSource(t, "export "+program)
		var gap *NotYet
		if !errors.As(err, &gap) || !strings.Contains(gap.What, "a view over an iterable of unproven origin") || !strings.Contains(gap.What, "parameter may receive untracked or omitted arguments") {
			t.Fatalf("got %v, want a NotYet saying the exported parameter's origin is unproven", err)
		}
	})
}
