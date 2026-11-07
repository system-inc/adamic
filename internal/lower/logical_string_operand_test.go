package lower

import (
	"errors"
	"strings"
	"testing"
)

// A falsy string is "" (or undefined), which only a string or a union result can hold. Where a brand
// hides the string's falsy part (tsc's __String), the checker types name && x as x alone, and Node's
// "" would reach a slot that can't say it, so it stops as not yet (#v5k6t3k). Before this, each panicked
// in emission. Mutant: the stop removed, and each program reaches native's "no conversion" panic.
func TestLogicalStringOperandHeldOtherwiseIsNotYet(test *testing.T) {
	test.Parallel()
	brand := "type Name = (string & { __brand: void }) | (void & { __brand: void });\nfunction nameOf(text: string): Name { return text as Name; }\n"
	for _, body := range []string{
		"function known(name: Name): boolean { return name === 'a'; }\nfunction test(text: string): boolean { const name = nameOf(text); if (name && !known(name)) { return true; } return false; }\nconsole.log(String(test('b')));",
		"interface Item { readonly label: string }\nconst table = new Map<string, Item>();\nfunction find(text: string): Item | undefined { const name = nameOf(text); return name && table.get(text); }\nconsole.log(String(find('a') === undefined));",
		"const tables = new Map<string, Map<string, number>>();\nfunction find(text: string): Map<string, number> | undefined { const name = nameOf(text); return name && tables.get(text); }\nconsole.log(String(find('a') === undefined));",
	} {
		_, lowerSourceError := lowerSource(test, brand+body)
		var gap *NotYet
		if !errors.As(lowerSourceError, &gap) || !strings.Contains(gap.What, "a logical string operand") {
			test.Errorf("got %v, want the logical string operand stop", lowerSourceError)
		}
	}
}
