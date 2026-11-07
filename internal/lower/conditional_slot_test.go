package lower

import (
	"errors"
	"strings"
	"testing"
)

const conditionalSlotAnimals = "class Animal { name = 'animal'; }\nclass Dog extends Animal { bark(): string { return 'woof'; } }\n"

const conditionalSlotPrelude = "const emptyArray: never[] = [];\nfunction parse(spec: string): number[][] { return [[spec.length]]; }\n"

// tsc's semver.ts:211 (#2vjqseg) writes a conditional of a number[][] and a shared never[] straight into
// a readonly field. The branches are judged against the field, which can't write, and lower. With no
// slot, the variable holds the reduced union number[][] and a push through it writes into emptyArray,
// so that stays refused by adamic/invariant-mutable. Mutant: conditionalBranchInSlot returning false;
// the readonly rows read Refused ("never[] seen as number[][]").
func TestConditionalBranchesAreJudgedAgainstTheirSlot(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		conditionalSlotPrelude +
			"class Range { private readonly alternatives: readonly (readonly number[])[]; constructor(spec: string) { this.alternatives = spec ? parse(spec) : emptyArray; } }\nconsole.log(String(new Range('ab') !== undefined));",
		conditionalSlotPrelude +
			"function count(spec: string): number { const all: readonly (readonly number[])[] = spec ? parse(spec) : emptyArray; return all.length; }\nconsole.log(String(count('ab')));",
		conditionalSlotPrelude +
			"function count(spec: string): number { const all: readonly (readonly number[])[] = (spec ? (parse(spec)) : (emptyArray)); return all.length; }\nconsole.log(String(count('ab')));",
	} {
		if _, lowerSourceError := lowerSource(test, source); lowerSourceError != nil {
			test.Errorf("%s: %v", source, lowerSourceError)
		}
	}
	for _, source := range []string{
		conditionalSlotPrelude +
			"function count(spec: string): number { const all = spec ? parse(spec) : emptyArray; all.push([1]); return all.length; }\nconsole.log(String(count('ab')));",
		conditionalSlotPrelude +
			"function count(spec: string): number { const all: number[][] = spec ? parse(spec) : emptyArray; all.push([1]); return all.length; }\nconsole.log(String(count('ab')));",
		// The skip hands each branch to the conditional's own judgment, which is per branch, never the
		// reduced union (Lower, Nov 5): dogs is still judged against Animal[] and refused.
		conditionalSlotAnimals +
			"function pick(flag: boolean, dogs: Dog[], animals: Animal[]): number { const pets: Animal[] = flag ? dogs : animals; pets.push(new Animal()); return pets.length; }\nconsole.log(String(pick(true, [new Dog()], [])));",
		// A generic argument's slot is its instantiated parameter, the reduced union, so the branch is
		// judged against it and refused.
		conditionalSlotAnimals +
			"function keep<T>(xs: T): T { return xs; }\nfunction pick(flag: boolean, dogs: Dog[], animals: Animal[]): number { const pets = keep(flag ? dogs : animals); pets.push(new Animal()); return pets.length; }\nconsole.log(String(pick(true, [new Dog()], [])));",
		// Language's d4: the conditional is an assignment's right side inside a readonly initializer, so
		// its own slot is the mutable held, where it's refused.
		conditionalSlotPrelude +
			"function count(spec: string): number { let held: number[][] = []; const all: readonly (readonly number[])[] = (held = spec ? parse(spec) : emptyArray); held.push([1]); return all.length; }\nconsole.log(String(count('ab')));",
		// Language's d6: one object literal with a readonly property and a mutable sibling, each holding
		// the conditional. The mutable sibling's slot refuses it.
		conditionalSlotPrelude +
			"interface Pair { readonly kept: readonly (readonly number[])[]; open: number[][]; }\nfunction count(spec: string): number { const pair: Pair = { kept: spec ? parse(spec) : emptyArray, open: spec ? parse(spec) : emptyArray }; pair.open.push([1]); return pair.kept.length; }\nconsole.log(String(count('ab')));",
		// A nested conditional's inner branches keep their implied target, since the inner conditional
		// isn't at a view site of its own. That over-refuses a readonly slot, and it's deliberate: it
		// fails closed. Lifting it needs the same per-branch reasoning carried through the nesting, not
		// a deleted row (Compiler, Nov 5).
		conditionalSlotPrelude +
			"class Range { private readonly alternatives: readonly (readonly number[])[]; constructor(spec: string, nested: boolean) { this.alternatives = spec ? (nested ? parse(spec) : emptyArray) : emptyArray; } }\nconsole.log(String(new Range('ab', true) !== undefined));",
	} {
		_, lowerSourceError := lowerSource(test, source)
		var refused *Refused
		if !errors.As(lowerSourceError, &refused) || !strings.Contains(refused.Error(), "adamic/invariant-mutable") {
			test.Errorf("%s: got %v, want adamic/invariant-mutable", source, lowerSourceError)
		}
	}
}
