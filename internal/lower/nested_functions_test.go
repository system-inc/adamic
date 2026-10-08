package lower

import (
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"slices"
	"strings"
	"testing"
)

func TestNestedFunctionGapsAreLoud(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"optional", `function run(): number { function inner(value?: number): number { return value ?? 1; } return inner(); } console.log(String(run()));`, "function value with an optional parameter"},
		{"default", `function run(): number { function inner(value = 1): number { return value; } return inner(); } console.log(String(run()));`, "function value with an optional parameter"},
		{"rest", `function run(): number { function inner(...values: number[]): number { return values.length; } return inner(1); } console.log(String(run()));`, "parameter that isn't a plain name"},
		{"self value", `function run(): () => number { function inner(): number { const self: () => number = inner; return 1; } return inner; } console.log(String(run()()));`, "first-class nested function reference"},
		{"block", `function run(): number { if (true) { function inner(): number { return 1; } return inner(); } return 0; } console.log(String(run()));`, "block-scoped nested"},
		{"generic", `function run(): number { function inner<T>(value: T): T { return value; } return inner(1); } console.log(String(run()));`, "generic or unnamed nested"},
		{"sibling value", `function run(): () => number { function a(): number { return 1; } function b(): () => number { return a; } return b(); } console.log(String(run()()));`, "first-class nested function reference"},
		{"ancestor call", `function run(): number { function a(): number { return 1; } function b(): number { function c(): number { return a(); } return c(); } return b(); } console.log(String(run()));`, "first-class nested function reference"},
		{"dynamic this", `function run(): number { function a(this: { x: number }): number { return this.x; } return 1; } console.log(String(run()));`, "dynamic this"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("want NotYet %q, got %v", probe.want, err)
			}
		})
	}
}

func TestNestedFunctionCycleUsesRegions(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `function make(): () => number {
 let saved: (() => number) | undefined = undefined;
 function read(): number { return saved === undefined ? 0 : saved(); }
 saved = read;
 return read;
}
console.log(String(make()()));`)
	if err != nil {
		t.Fatal(err)
	}
	cells, closures := 0, 0
	for _, local := range program.Locals {
		if local.GraphCell {
			cells++
		}
	}
	for _, function := range program.Functions {
		if function.GraphClosure {
			closures++
		}
	}
	if cells == 0 || closures == 0 {
		t.Fatal("cyclic environment has no graph cells or closures")
	}

}

func TestNestedEnvironmentHasOneAllocationSite(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `function make(start: number): () => number {
 let count = start;
 function first(): number { count += start; return count; }
 function second(): number { return first(); }
 return second;
} console.log(String(make(2)()));`)
	if err != nil {
		t.Fatal(err)
	}
	sites := 0
	for _, function := range program.Functions {
		for _, statement := range function.Body {
			allocation, ok := statement.(ir.AllocateEnvironment)
			if !ok {
				continue
			}
			sites++
			if len(allocation.Cells) != 2 || !slices.Equal(allocation.Cells, function.FrameEnvironment) {
				t.Fatal("expected one complete parameter/local layout")
			}
			for _, local := range allocation.Cells {
				if !program.Locals[local].EnvironmentCell {
					t.Fatal("slot not linked to its environment")
				}
			}
		}
	}
	if sites != 1 {
		t.Fatalf("want one environment allocation, got %d", sites)
	}
}

func TestNestedEnvironmentCycleIncludesDisjointSlots(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `function make(): () => number {
 let saved: (() => number) | undefined = undefined;
 let count = 1;
 function read(): number { return count; }
 const observe = (): string => saved === undefined ? "empty" : "held";
 saved = read;
 console.log(observe());
 return read;
} const held = make(); console.log(String(held()));`)
	if err != nil {
		t.Fatal(err)
	}
	cells, closures := 0, 0
	for _, local := range program.Locals {
		if local.GraphCell {
			cells++
		}
	}
	for _, function := range program.Functions {
		if function.GraphClosure {
			closures++
		}
	}
	if cells == 0 || closures == 0 {
		t.Fatal("cyclic environment has no graph cells or closures")
	}

}
