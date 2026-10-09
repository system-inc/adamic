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

func TestNestedFunctionGapsAreLoud(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"block", `function run(): number { if (true) { function inner(): number { return 1; } return inner(); } return 0; } console.log(String(run()));`, "block-scoped nested"},
		{"generic", `function run(): number { function inner<T>(value: T): T { return value; } const stored = inner; return stored(1); } console.log(String(run()));`, "generic function as a value"},
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

func TestNestedFunctionCycleIsRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function make(): () => number {
 let saved: (() => number) | undefined = undefined;
 function read(): number { return saved === undefined ? 0 : saved(); }
 saved = read;
 return read;
}
console.log(String(make()()));`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/cycle-capable") {
		t.Fatalf("want cycle refusal, got %v", err)
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
	_, err := lowerSource(t, `function make(): () => number {
 let saved: (() => number) | undefined = undefined;
 let count = 1;
 function read(): number { return count; }
 const observe = (): string => saved === undefined ? "empty" : "held";
 saved = read;
 console.log(observe());
 return read;
} const held = make(); console.log(String(held()));`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/cycle-capable") {
		t.Fatalf("want complete environment cycle refusal, got %v", err)
	}
}

func TestNestedCapturedParametersAreOwned(t *testing.T) {
	program, err := lowerSource(t, `function make(text: string): () => string { function read(): string { return text; } return read; } console.log(make("hello")());`)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, local := range program.Locals {
		if local.EnvironmentCell && local.Type == ir.String {
			found = true
			if local.Borrowed {
				t.Fatal("captured parameter is borrowed")
			}
		}
	}
	if !found {
		t.Fatal("missing captured parameter")
	}
}

func TestClosedFrameInputRejectsMutation(t *testing.T) {
	program, err := lowerSource(t, `function parser(scanner: { scan: () => number }): () => number {
 let token = 0;
 function read(): number { return token; }
 function next(): number { return token = scanner.scan(); }
 next(); return read;
 } console.log(String(parser({scan: () => 1})()));`)
	if err != nil {
		t.Fatal(err)
	}
	l := lowering{result: program}
	input := -1
	for i, local := range program.Locals {
		if local.Name == "scanner" {
			input = i
		}
	}
	if input < 0 || !l.closedFrameInput(input) {
		t.Fatal("closed literal input was not proven")
	}
	owner := program.Locals[input].Function
	// A field store invalidates the closed graph even if all initial literals were safe.
	program.Functions[owner].Body = append(program.Functions[owner].Body, ir.SetProperty{})
	if l.closedFrameInput(input) {
		t.Fatal("mutable graph accepted as closed")
	}
}

// Bypass only the suppression-directive gate to exercise the lowering guard
// behind TypeScript's earlier TS2630 diagnostic. The public fixture pins TS2630.
func TestNestedRebindingNotYet(t *testing.T) {
	path, err := filepath.Abs("../oracle/refusals/nested_rebinding.a")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Replace(string(source), "    inner =", "    // @ts-expect-error\n    inner =", 1)
	program, err := load.LoadOverlay([]string{path}, map[string]string{path: text})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	check, release := program.Checker(context.Background(), file)
	defer release()
	{
		l := &lowering{program: program, checker: check, result: &ir.Program{}, this: -1, functionIndex: -1}
		err := l.declareModule(file.Statements.Nodes)
		// declareModule lowers the body, so the ordinary marker reaches the guard.
		if err == nil {
			_, err = l.statements(file.Statements.Nodes)
		}
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "rebinding a nested function declaration") {
			t.Fatalf("want canonical NotYet, got %v", err)
		}
	}
}

func TestNestedCallbackCycleIsRefused(t *testing.T) {
	_, err := lowerSource(t, `function make(): () => number {
 let saved: (() => number) | undefined = undefined;
 function read(): number { return saved === undefined ? 1 : saved(); }
 function factory(): () => number { return () => read(); }
 saved = factory();
 return saved;
} console.log(String(make()()));`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/cycle-capable") {
		t.Fatalf("want callback environment cycle refusal, got %v", err)
	}
}

func TestNestedBodylessDeclarationsAreLoud(t *testing.T) {
	path, err := filepath.Abs("../oracle/testdata/scanner_nested_overload.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	check, release := program.Checker(context.Background(), file)
	defer release()
	outer := file.Statements.Nodes[0]
	signature := outer.Body().AsBlock().Statements.Nodes[0]
	if signature.Body() != nil {
		t.Fatal("probe must contain a bodyless overload")
	}
	l := &lowering{program: program, checker: check, result: &ir.Program{Functions: []ir.Function{{Name: "outer"}}}, function: &ir.Function{}, functionIndex: 0, this: -1}
	for _, probe := range []struct {
		name string
		run  func() error
		want string
	}{
		{"missing implementation", func() error { _, err := l.nestedDeclarations([]*ast.Node{signature}); return err }, "a nested function declaration without an implementation"},
		{"body lowering", func() error { return l.lowerBody(0, signature, -1, nil, nil) }, "a function without a body"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			err := probe.run()
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("want NotYet %q, got %v", probe.want, err)
			}
		})
	}
}

func TestNestedRestIsSupported(t *testing.T) {
	t.Parallel()
	if _, err := lowerSource(t, `function run(): number { function inner(...values: number[]): number { return values.length; } return inner(1); } console.log(String(run()));`); err != nil {
		t.Fatal(err)
	}
}
