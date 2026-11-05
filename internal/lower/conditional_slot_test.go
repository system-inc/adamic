package lower

import (
	"errors"
	"strings"
	"testing"
)

const conditionalSlotPrelude = "const emptyArray: never[] = [];\nfunction parse(spec: string): number[][] { return [[spec.length]]; }\n"

// tsc's semver.ts:211 (#2vjqseg) writes a conditional of a number[][] and a shared never[] straight into
// a readonly field. The branches are judged against the field, which can't write, and lower. With no
// slot, the variable holds the reduced union number[][] and a push through it writes into emptyArray,
// so that stays refused by adamic/invariant-mutable. Mutant: conditionalBranchInSlot returning false;
// the readonly rows read Refused ("never[] seen as number[][]").
func TestConditionalBranchesAreJudgedAgainstTheirSlot(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		conditionalSlotPrelude + "class Range { private readonly alternatives: readonly (readonly number[])[]; constructor(spec: string) { this.alternatives = spec ? parse(spec) : emptyArray; } }\nconsole.log(String(new Range('ab') !== undefined));",
		conditionalSlotPrelude + "function count(spec: string): number { const all: readonly (readonly number[])[] = spec ? parse(spec) : emptyArray; return all.length; }\nconsole.log(String(count('ab')));",
		conditionalSlotPrelude + "function count(spec: string): number { const all: readonly (readonly number[])[] = (spec ? (parse(spec)) : (emptyArray)); return all.length; }\nconsole.log(String(count('ab')));",
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Errorf("%s: %v", source, err)
		}
	}
	for _, source := range []string{
		conditionalSlotPrelude + "function count(spec: string): number { const all = spec ? parse(spec) : emptyArray; all.push([1]); return all.length; }\nconsole.log(String(count('ab')));",
		conditionalSlotPrelude + "function count(spec: string): number { const all: number[][] = spec ? parse(spec) : emptyArray; all.push([1]); return all.length; }\nconsole.log(String(count('ab')));",
	} {
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(refused.Error(), "adamic/invariant-mutable") {
			t.Errorf("%s: got %v, want adamic/invariant-mutable", source, err)
		}
	}
}
