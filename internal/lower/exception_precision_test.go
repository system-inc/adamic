package lower

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// These assertions hold exception precision independently of runtime counts.
// Marking every call as possibly throwing must fail even if nobody regenerates
// or compares the counts table.
func TestMayThrowPrecision(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `
interface Node { readonly label: string; }
const label = 'ready';
function read(node: Node | undefined): string {
    if (node === undefined) { return label; }
    return node.label;
}
function forward(node: Node | undefined): string { return read(node); }
console.log(forward({ label: 'present' }));
console.log(forward(undefined));
`)
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, function := range program.Functions {
		if function.Name == "read" || function.Name == "forward" {
			found++
			if function.MayThrow {
				t.Errorf("%s cannot throw on any reachable path", function.Name)
			}
		}
	}
	if found != 2 {
		t.Fatalf("found %d source functions, want 2", found)
	}
	calls := 0
	walk(program.Main, func(node any) bool {
		if call, ok := node.(ir.Call); ok {
			calls++
			if program.CallMayThrow(call) {
				t.Errorf("call to %s cannot throw", program.Functions[call.Function].Name)
			}
		}
		return true
	})
	if calls != 2 {
		t.Fatalf("found %d main calls, want 2", calls)
	}
}

func TestReadinessIncludesEarlyIndirectCalls(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `
function read(): string { return label; }
const indirect = (): string => read();
try { console.log(indirect()); } catch (error) { console.log('early'); }
const label = 'ready';
console.log(indirect());
`)
	if err != nil {
		t.Fatal(err)
	}
	for _, function := range program.Functions {
		if function.Name == "read" {
			if !function.MayThrow {
				t.Fatal("early indirect global read must throw")
			}
			return
		}
	}
	t.Fatal("missing read function")
}

func TestGeneratedGuardPrecision(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `
interface Sample { readonly figures: number; readonly value: number; }
const samples: readonly Sample[] = [{ figures: 3, value: 1.25 }, { figures: 5, value: 12 }];
function format(): string {
    let result = '';
    for (const sample of samples) { result += sample.value.toPrecision(sample.figures); }
    for (let digits = 0; digits <= 20; digits += 1) { result += (1.25).toFixed(digits); }
    return result;
}
console.log(format());
`)
	if err != nil {
		t.Fatal(err)
	}
	for _, function := range program.Functions {
		if function.Name == "format" {
			if function.MayThrow {
				t.Fatal("bounded formats cannot throw")
			}
			return
		}
	}
	t.Fatal("missing format function")
}

func TestGeneratedGuardFactsIncludeUnknownStores(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `
interface Sample { figures: number; readonly value: number; }
const sample: Sample = { figures: 3, value: 1.25 };
function change(target: Sample, figures: number): void { target.figures = figures; }
function format(): string { return sample.value.toPrecision(sample.figures); }
change(sample, -1);
console.log(format());
`)
	if err != nil {
		t.Fatal(err)
	}
	for _, function := range program.Functions {
		if function.Name == "format" {
			if !function.MayThrow {
				t.Fatal("unknown field store must remain throwing")
			}
			return
		}
	}
	t.Fatal("missing format function")
}

func TestGeneratedGuardCounterWritesRemainThrowing(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `
function format(): string {
    let result = '';
    for (let digits = 0; digits <= 2; digits += 1) {
        if (digits === 0) { digits = -1; }
        result += (1.25).toFixed(digits);
    }
    return result;
}
console.log(format());
`)
	if err != nil {
		t.Fatal(err)
	}
	for _, function := range program.Functions {
		if function.Name == "format" {
			if !function.MayThrow {
				t.Fatal("counter modified in body must remain throwing")
			}
			return
		}
	}
	t.Fatal("missing format function")
}

func TestProtectedStringGuardPrecision(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `
function bounded(value: number): string {
    try { return 'value ' + String(value); } catch { return 'caught'; }
}
function finished(): string {
    let result = '';
    for (let index = 0; index < 6; index++) {
        try { console.log(bounded(index)); } finally { result += String(index) + ' '; }
    }
    return result;
}
console.log(finished());
`)
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, function := range program.Functions {
		if function.Name == "bounded" || function.Name == "finished" {
			found++
			if function.MayThrow {
				t.Errorf("%s cannot throw on this bounded protected path", function.Name)
			}
		}
	}
	if found != 2 {
		t.Fatalf("found %d source functions, want 2", found)
	}
}

func TestStringLengthBoundsIncludeUnknownObjects(t *testing.T) {
	t.Parallel()
	for _, producer := range []ir.Expression{
		ir.ObjectLiteral{Spread: ir.Read{Local: 0, Of: ir.Object}},
		ir.ObjectCall{Method: "fromEntries", Returns: ir.Object},
	} {
		t.Run("unknown producer", func(t *testing.T) {
			program := &ir.Program{Strings: []string{"small"}, Main: []ir.Statement{
				ir.Evaluate{Value: ir.ObjectLiteral{Fields: []ir.Field{{Name: "text", Value: ir.StringConstant{Index: 0}}}}},
				ir.Evaluate{Value: producer},
			}}
			lowering := &lowering{result: program}
			bound := lowering.stringLengthBounds()(ir.Property{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "text", Of: ir.String})
			if bound != maximumStringLength {
				t.Fatalf("unknown object must retain the full string bound, got %v", bound)
			}
		})
	}
}

func TestRepeatRefusalChecksResultLength(t *testing.T) {
	t.Parallel()
	lowering := &lowering{result: &ir.Program{Strings: []string{"ab"}}}
	operation := ir.StringCall{Method: "repeat", Value: ir.StringConstant{Index: 0}, Arguments: []ir.Expression{ir.NumberConstant{Value: 268435456}}}
	if failing := lowering.libraryFailure([]ir.Statement{ir.Evaluate{Value: operation}}, map[int]bool{}); failing != "repeat length or count" {
		t.Fatalf("oversized constant repeat must remain a named refusal without its guard, got %q", failing)
	}
	operation.Arguments[0] = ir.NumberConstant{Value: 2}
	if failing := lowering.libraryFailure([]ir.Statement{ir.Evaluate{Value: operation}}, map[int]bool{}); failing != "" {
		t.Fatalf("small constant repeat is safe, got %q", failing)
	}
}

func TestLoopPresenceSurvivesRuntimeMutation(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `
interface Link { readonly label: string; readonly pattern: RegExp; next: Link | undefined; }
const root: Link = { label: 'root', pattern: /a/g, next: undefined };
let cursor: Link | undefined = root;
const labels: string[] = [];
while (cursor !== undefined) {
    labels.push(cursor.pattern.test('a') ? cursor.label : 'absent');
    cursor = cursor.next;
}
console.log(labels.join('|'));
`)
	if err != nil {
		t.Fatal(err)
	}
	walk(program.Main, func(node any) bool {
		if call, ok := node.(ir.Call); ok && program.Functions[call.Function].Name == "error_defined" {
			t.Error("loop condition proves cursor present; runtime operations cannot reassign it")
		}
		return true
	})
}

func TestCheckedUnionReferenceCannotThrow(t *testing.T) {
	program, err := lowerSource(t, `let held: string | number | undefined = 'ready'; function keep(): void { held = 'still ready'; } function read(): number { if (typeof held === 'string') { keep(); return held.length; } return 0; } console.log(String(read()));`)
	if err != nil {
		t.Fatal(err)
	}
	for _, function := range program.Functions {
		if function.Name == "read" {
			if function.MayThrow {
				t.Fatal("the checked string-tag return is present; its second guard cannot throw")
			}
			return
		}
	}
	t.Fatal("missing read function")
}
