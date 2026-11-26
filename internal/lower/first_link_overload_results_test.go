package lower

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// lowerFirstLinkTS is lowerFirstLink for a .ts program, as the overload result shapes are written.
func lowerFirstLinkTS(t *testing.T, source string, traps bool) (*ir.Program, error) {
	t.Helper()
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return LowerWithOptions(context.Background(), program, Options{FirstLinkTraps: traps})
}

// A call to an overload accepted narrow, inside a unit the refusal pass trapped before anything lowered, never
// gets its result check and needs none: the trap replaced the unit's code. checkOverloadResults passes over it
// in a first-link lowering, where it panicked before, as tsc's transformers/classFields.ts:2223:59 did on the
// harness's first run (#1061tyd). The unit is a function in one row and a top-level statement in the other;
// each lowers with the refusal as its one trap, and without the option each is refused at compile time.
// Mutant: checkOverloadResults without its two firstLinkSkips guards, and both rows panic again,
// "a call to an overload accepted narrow ... lowered without its result check".
func TestFirstLinkOverloadCallInATrappedUnitIsPassedOver(t *testing.T) {
	t.Parallel()
	shape := strings.Replace(overloadLiteralShape, "BODY", "true", 1)
	for _, row := range []struct{ name, source, unit string }{
		{"function", shape + "function second(n: number): boolean {\n\tconst ok = canDecorate(true, declaration);\n\treturn ok == true;\n}\nconsole.log('before');\nconsole.log(`${second(1)}`);\n", "main.ts second"},
		{"statement", shape + "console.log('before');\nconst flag = canDecorate(true, declaration) == true;\nconsole.log(`${flag}`);\n", "main.ts module"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerFirstLinkTS(t, row.source, true)
			if err != nil {
				t.Fatal(err)
			}
			if len(program.Traps) != 1 || program.Traps[0].Kind != "Refused" || program.Traps[0].Phase != "Refusal" || program.Traps[0].Unit != row.unit {
				t.Fatalf("Traps: got %+v, want the one refusal of == in %s", program.Traps, row.unit)
			}
			_, off := lowerFirstLinkTS(t, row.source, false)
			var refused *Refused
			if !errors.As(off, &refused) {
				t.Fatalf("without the option: got %v, want the refusal of ==", off)
			}
		})
	}
}

// The same narrow call in a unit that lowers is still held to its result check: the guard keeps its promise
// everywhere the code is real, and a unit that lowered and then stopped (Math.random, a NotYet) has its call
// checked before the stop, so its trap leaves the count as the untrapped program's.
func TestFirstLinkOverloadCallInALoweredUnitIsStillChecked(t *testing.T) {
	t.Parallel()
	shape := strings.Replace(overloadLiteralShape, "BODY", "true", 1)
	program, err := lowerFirstLinkTS(t, shape+"function second(): number {\n\tconst ok = canDecorate(true, declaration);\n\tconsole.log(String(ok));\n\treturn Math.random();\n}\nconsole.log('before');\nconsole.log(`${second()}`);\n", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Traps) != 1 || program.Traps[0].Kind != "NotYet" || program.OverloadResultChecks.Checked != 2 {
		t.Fatalf("got traps %+v and checks %+v, want one NotYet trap and both calls checked", program.Traps, program.OverloadResultChecks)
	}
}
