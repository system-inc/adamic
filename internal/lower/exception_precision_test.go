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
