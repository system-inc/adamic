package lower

import (
	"context"
	"errors"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestNestedFunctionGapsAreLoud(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{
			"block",
			`function run(): number { if (true) { function inner(): number { return 1; } return inner(); } return 0; } console.log(String(run()));`,
			"block-scoped nested",
		},
		{
			"generic",
			`function run(): number { function inner<T>(value: T): T { return value; } const stored = inner; return stored(1); } console.log(String(run()));`,
			"generic function as a value",
		},
		{
			"dynamic this",
			`function run(): number { function a(this: { x: number }): number { return this.x; } return 1; } console.log(String(run()));`,
			"dynamic this",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			_, lowerSourceError := lowerSource(test, probe.source)
			var notYet *NotYet
			if !errors.As(lowerSourceError, &notYet) || !strings.Contains(lowerSourceError.Error(), probe.want) {
				test.Fatalf("want NotYet %q, got %v", probe.want, lowerSourceError)
			}
		})
	}
}

func TestNestedFunctionCycleUsesRegions(test *testing.T) {
	test.Parallel()
	program := graphAccepted(test, `function make(): () => number {
 let saved: (() => number) | undefined = undefined;
 let remaining = 3;
 function read(): number { if (remaining <= 0) return 0; remaining -= 1; return saved === undefined ? 0 : saved(); }
 saved = read;
 return read;
}
console.log(String(make()()));`)
	if len(program.GraphTypes) == 0 {
		test.Fatal("cyclic environment stayed counted")
	}
}

func TestNestedEnvironmentHasOneAllocationSite(test *testing.T) {
	test.Parallel()
	// JavaScript cannot observe environment allocation counts or ownership layout.
	program := lowersAndAgreesWithNode(test, `function make(start: number): () => number {
 let count = start;
 function first(): number { count += start; return count; }
 function second(): number { return first(); }
 return second;
} console.log(String(make(2)()));`)
	sites := 0
	for _, function := range program.Functions {
		for _, statement := range function.Body {
			allocation, ok := statement.(ir.AllocateEnvironment)
			if !ok {
				continue
			}
			sites++
			if len(allocation.Cells) != 2 || !slices.Equal(allocation.Cells, function.FrameEnvironment) {
				test.Fatal("expected one complete parameter/local layout")
			}
			for _, local := range allocation.Cells {
				if !program.Locals[local].EnvironmentCell {
					test.Fatal("slot not linked to its environment")
				}
			}
		}
	}
	if sites != 1 {
		test.Fatalf("want one environment allocation, got %d", sites)
	}
}

func TestNestedEnvironmentCycleIncludesDisjointSlots(test *testing.T) {
	test.Parallel()
	program := graphAccepted(test, `function make(): () => number {
 let saved: (() => number) | undefined = undefined;
 let count = 1;
 function read(): number { return count; }
 const observe = (): string => saved === undefined ? "empty" : "held";
 saved = read;
 console.log(observe());
 return read;
} const held = make(); console.log(String(held()));`)
	if len(program.GraphTypes) == 0 {
		test.Fatal("missing graph ownership")
	}
}

func TestNestedCapturedParametersAreOwned(test *testing.T) {
	test.Parallel()
	// JavaScript garbage collection cannot observe captured parameter ownership.
	program := lowersAndAgreesWithNode(
		test,
		`function make(text: string): () => string { function read(): string { return text; } return read; } console.log(make("hello")());`,
	)
	found := false
	for _, local := range program.Locals {
		if local.EnvironmentCell && local.Type == ir.String {
			found = true
			if local.Borrowed {
				test.Fatal("captured parameter is borrowed")
			}
		}
	}
	if !found {
		test.Fatal("missing captured parameter")
	}
}

func TestClosedFrameInputRejectsMutation(test *testing.T) {
	test.Parallel()
	// Behavior cannot observe the closed-frame ownership proof or its invalidation.
	program := lowersAndAgreesWithNode(test, `function parser(scanner: { scan: () => number }): () => number {
 let token = 0;
 function read(): number { return token; }
 function next(): number { return token = scanner.scan(); }
 next(); return read;
 } console.log(String(parser({scan: () => 1})()));`)
	lowering := lowering{result: program}
	input := -1
	for index, local := range program.Locals {
		if local.Name == "scanner" {
			input = index
		}
	}
	if input < 0 || !lowering.closedFrameInput(input) {
		test.Fatal("closed literal input was not proven")
	}
	owner := program.Locals[input].Function
	// A field store invalidates the closed graph even if all initial literals were safe.
	program.Functions[owner].Body = append(program.Functions[owner].Body, ir.SetProperty{})
	if lowering.closedFrameInput(input) {
		test.Fatal("mutable graph accepted as closed")
	}
}

// Bypass only the suppression-directive gate to exercise the lowering guard
// behind TypeScript's earlier TS2630 diagnostic. The public fixture pins TS2630.
func TestNestedRebindingNotYet(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("../oracle/refusals/nested_rebinding.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	source, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	text := strings.Replace(string(source), "    inner =", "    // @ts-expect-error\n    inner =", 1)
	program, loadOverlayError := load.LoadOverlay([]string{path}, map[string]string{path: text})
	if loadOverlayError != nil {
		test.Fatal(loadOverlayError)
	}
	file := program.Files()[0]
	check, release := program.Checker(context.Background(), file)
	defer release()
	{
		lowering := &lowering{program: program, checker: check, result: &ir.Program{}, this: -1, functionIndex: -1}
		err := lowering.declareModule(file.Statements.Nodes)
		// declareModule lowers the body, so the ordinary marker reaches the guard.
		if err == nil {
			_, err = lowering.statements(file.Statements.Nodes)
		}
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "rebinding a nested function declaration") {
			test.Fatalf("want canonical NotYet, got %v", err)
		}
	}
}

func TestNestedCallbackCycleIsRefused(test *testing.T) {
	test.Parallel()
	_, lowerSourceError := lowerSource(test, `function make(): () => number {
 let saved: (() => number) | undefined = undefined;
 function read(): number { return saved === undefined ? 1 : saved(); }
 function factory(): () => number { return () => read(); }
 saved = factory();
 return saved;
} console.log(String(make()()));`)
	var refused *Refused
	if !errors.As(lowerSourceError, &refused) || !strings.Contains(lowerSourceError.Error(), "adamic/cycle-capable") {
		test.Fatalf("want callback environment cycle refusal, got %v", lowerSourceError)
	}
}

func TestNestedBodylessDeclarationsAreLoud(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct {
		name string
		run  func(*lowering, *ast.Node) error
		want string
	}{
		{"missing implementation", func(lowering *lowering, signature *ast.Node) error {
			_, nestedDeclarationsError := lowering.nestedDeclarations([]*ast.Node{signature})
			return nestedDeclarationsError
		}, "a nested function declaration without an implementation"},
		{
			"body lowering",
			func(lowering *lowering, signature *ast.Node) error {
				return lowering.lowerBody(0, signature, -1, nil, nil)
			},
			"a function without a body",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs("../oracle/testdata/scanner_nested_overload.a")
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loadError := load.Load([]string{path})
			if loadError != nil {
				test.Fatal(loadError)
			}
			file := program.Files()[0]
			check, release := program.Checker(context.Background(), file)
			defer release()
			outer := file.Statements.Nodes[0]
			signature := outer.Body().AsBlock().Statements.Nodes[0]
			if signature.Body() != nil {
				test.Fatal("probe must contain a bodyless overload")
			}
			lowering := &lowering{
				program:       program,
				checker:       check,
				result:        &ir.Program{Functions: []ir.Function{{Name: "outer"}}},
				function:      &ir.Function{},
				functionIndex: 0,
				this:          -1,
			}
			runError := probe.run(lowering, signature)
			var notYet *NotYet
			if !errors.As(runError, &notYet) || !strings.Contains(runError.Error(), probe.want) {
				test.Fatalf("want NotYet %q, got %v", probe.want, runError)
			}
		})
	}
}

func TestNestedRestIsSupported(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(
		test,
		`function run(): number { function inner(...values: number[]): number { return values.length; } return inner(1); } console.log(String(run()));`,
	)
}
