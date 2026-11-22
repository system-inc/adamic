package lower

import (
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A visit whose callee is parenthesized, (xs.forEach)(f) or (xs.find)(f), is xs.forEach(f) or xs.find(f): it lowers,
// and to the same ArrayVisit (#x490w4a). Reading the receiver for the view check off the unskipped callee panicked on
// the ParenthesizedExpression. The callback's type is the checker's number, so it's required named in both, not equal.
func TestArrayVisitParenthesizedCallee(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		method, source string
	}{
		{"forEach", "function each(f: (n: number) => void): void { VISIT; }\neach((n: number): void => { console.log(String(n)); });\n"},
		{"find", "function first(f: (n: number) => boolean): number | undefined { return VISIT; }\nconst found = first((n: number): boolean => n > 1);\nif (found !== undefined) { console.log(String(found)); }\n"},
	} {
		plainCall, parenthesizedCall := "xs."+probe.method+"(f)", "(xs."+probe.method+")(f)"
		plain := arrayVisitLowered(t, plainCall, "const xs: number[] = [1 + 0, 2, 3];\n"+strings.ReplaceAll(probe.source, "VISIT", plainCall))
		parenthesized := arrayVisitLowered(t, parenthesizedCall, "const xs: number[] = [1 + 0, 2, 3];\n"+strings.ReplaceAll(probe.source, "VISIT", parenthesizedCall))
		if plain.CallbackType == 0 || parenthesized.CallbackType == 0 {
			t.Errorf("%s: callback types %d and %d, want both named", probe.method, plain.CallbackType, parenthesized.CallbackType)
		}
		plain.CallbackType, parenthesized.CallbackType = 0, 0
		if !reflect.DeepEqual(plain, parenthesized) {
			t.Errorf("%s lowers to %#v, want %s's %#v", parenthesizedCall, parenthesized, plainCall, plain)
		}
	}
}

// arrayVisitLowered lowers source and gives back the one ArrayVisit in its functions, stopping the test on a panic, an
// error, or other than one visit.
func arrayVisitLowered(t *testing.T, call, source string) (visit ir.ArrayVisit) {
	t.Helper()
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("%s panicked: %v", call, recovered)
		}
	}()
	program, err := lowerSource(t, source)
	if err != nil {
		t.Fatalf("%s: %v", call, err)
	}
	visits := 0
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			if found, isVisit := node.(ir.ArrayVisit); isVisit {
				visits++
				visit = found
			}
			return true
		})
	}
	if visits != 1 {
		t.Fatalf("%s: found %d visits, want one", call, visits)
	}
	return visit
}
