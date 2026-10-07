package lower

import "testing"

func TestMayThrowBoundsOnlyProvenClosureCalls(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `
const fixed = (): number => 1;
let mutable = (): number => 2;
function unrelated(): number { throw new Error('unrelated'); }
function literal(): number { return (() => 3)(); }
function constant(): number { return fixed(); }
function unknown(): number { return mutable(); }
console.log(String(literal() + constant() + unknown()));
`)
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, function := range program.Functions {
		switch function.Name {
		case "literal", "constant":
			found++
			if function.MayThrow {
				t.Errorf("%s cannot throw", function.Name)
			}
		case "unknown":
			found++
			if !function.MayThrow {
				t.Error("Unknown must include the unrelated throwing function")
			}
		}
	}
	if found != 3 {
		t.Fatalf("found %d functions", found)
	}
}
