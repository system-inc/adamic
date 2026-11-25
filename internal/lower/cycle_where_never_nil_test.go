package lower

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A holder the cycle finder met with no node is refused with a Where, never handed to Program.Where as nil,
// which on tsc panicked in slotsOf (#ew207ss). An array and a tuple of Node, whose elements reach back through
// kids, are recorded as met with a nil node, and one unproven element write is planted, so slotsOf refuses
// each: the refusal's Where is whereOf's, a declaration or the first file.
// Mutant: whereOf returning f.where[proven] raw, and slotsOf panics in Program.Where.
func TestCycleRefusalOfAHolderMetWithNoNodeHasAWhere(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"list", "pair"} {
		f, types := cycleIndexTypes(t, `interface Node { kids: Node[] } declare const list: Node[]; declare const pair: [Node];`)
		holder := types[name]
		if holder == nil {
			t.Fatalf("no type for %s", name)
		}
		f.where[holder] = nil
		f.writes = []fresh.Write{{Kind: fresh.WriteElement, Function: -1, Why: "a write the test plants"}}
		var refused *Refused
		func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Errorf("slotsOf(%s) panicked with the holder met with no node: %v", name, recovered)
				}
			}()
			if err := f.slotsOf(holder); !errors.As(err, &refused) {
				t.Errorf("slotsOf(%s): got %v, want the cycle refusal", name, err)
			}
		}()
		if refused == nil {
			continue
		}
		if !refused.cycle || refused.Where == "" {
			t.Errorf("slotsOf(%s): got %+v, want a cycle refusal with a Where", name, refused)
		}
		t.Logf("%s refused at %s: %s", name, refused.Where, refused.What)
	}
}

// cycleWhereProbes are programs whose cycle holder may be met only through an instantiation with no node.
var cycleWhereProbes = []struct{ name, source string }{
	{"union argument", "interface Sized {\n\tlength: number;\n}\nfunction keep<T>(value: T): number {\n\treturn 0;\n}\nfunction add(lists: number[][], item: number[]): void {\n\tlists.push(item);\n}\nconst lists: number[][] = [];\nadd(lists, [1]);\nconsole.log(`${keep<Sized[] | undefined>(undefined) + lists.length}`);\n"},
	{"tuple argument", "interface Sized {\n\tlength: number;\n}\nfunction keep<T>(value: T | undefined): number {\n\treturn 0;\n}\nfunction add(lists: number[][], item: number[]): void {\n\tlists.push(item);\n}\nconst lists: number[][] = [];\nadd(lists, [1]);\nconsole.log(`${keep<[Sized[], number]>(undefined) + lists.length}`);\n"},
}

// Probe: what each candidate program lowers to, plain and with first-link traps.
func TestCycleWhereProbe(t *testing.T) {
	t.Parallel()
	for _, probe := range cycleWhereProbes {
		for _, traps := range []bool{false, true} {
			outcome := func() (text string) {
				defer func() {
					if recovered := recover(); recovered != nil {
						text = fmt.Sprintf("panic: %v", recovered)
					}
				}()
				checked, err := load.Load([]string{writeAgreementSource(t, "main.a", probe.source)})
				if err != nil {
					return "Load: " + err.Error()
				}
				program, err := LowerWithOptions(context.Background(), checked, Options{FirstLinkTraps: traps})
				if err != nil {
					return "error: " + err.Error()
				}
				rows := []ir.Trap{}
				if program != nil {
					rows = program.Traps
				}
				return fmt.Sprintf("lowered, traps %+v", rows)
			}()
			t.Logf("PROBE %s traps=%t: %s", probe.name, traps, outcome)
		}
	}
	for _, probe := range cycleWhereProbes {
		checked, err := load.Load([]string{writeAgreementSource(t, "main.a", probe.source)})
		if err != nil {
			t.Logf("PROBE %s Load: %v", probe.name, err)
			continue
		}
		entry := checked.Files()[0]
		check, release := checked.Checker(context.Background(), entry)
		lowered, err := lowerChecked(checked, check, entry, nil, false, false)
		if lowered != nil {
			for _, proven := range lowered.instantiated {
				t.Logf("PROBE %s instantiated %s, node %v", probe.name, check.TypeToString(proven), lowered.classNodeFor(proven) != nil)
			}
		}
		t.Logf("PROBE %s lowerChecked: %v", probe.name, err)
		release()
	}
}
