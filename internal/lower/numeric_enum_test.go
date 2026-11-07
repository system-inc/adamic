package lower

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// numericEnum is computed once per enum symbol, however many times void widening asks, and a held answer is the
// checker's: a numeric enum, a string-valued member, and one computed from another (#e0k9s2t). Mutant, red: the held
// answer dropped, so every ask walks the members again.
func TestNumericEnumIsComputedOncePerSymbol(t *testing.T) {
	t.Parallel()
	finder, _ := cycleIndexTypes(t, `enum Numeric { A, B = 4, C }
enum Mixed { X = 1, Y = "y" }
enum Derived { P = 2, Q = P * 3 }
declare const n: Numeric; declare const m: Mixed; declare const d: Derived;`)
	held := finder.l
	symbols := map[string]*ast.Symbol{}
	for _, statement := range held.program.Files()[0].Statements.Nodes {
		if statement.Kind == ast.KindEnumDeclaration {
			symbols[statement.Name().Text()] = held.checker.GetSymbolAtLocation(statement.Name())
		}
	}
	want := map[string]bool{"Numeric": true, "Mixed": false, "Derived": true}
	for name, numeric := range want {
		for asked := 1; asked <= 3; asked++ {
			if got := held.numericEnum(symbols[name]); got != numeric {
				t.Fatalf("%s, ask %d: numericEnum %t, want %t", name, asked, got, numeric)
			}
		}
	}
	if held.numericEnumWalks != len(want) {
		t.Fatalf("numericEnum walked members %d times for %d enums asked three times each", held.numericEnumWalks, len(want))
	}
	fresh := &lowering{program: held.program, checker: held.checker, result: &ir.Program{}, this: -1, functionIndex: -1}
	for name, numeric := range want {
		if fresh.numericEnum(symbols[name]) != numeric {
			t.Fatalf("%s: a fresh lowering's answer isn't the held one", name)
		}
	}
}
