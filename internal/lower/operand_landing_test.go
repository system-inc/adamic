package lower

import (
	"errors"
	"strings"
	"testing"
)

// Adamic's ruling on #r11wppy (applying #7n03qzg, rulings.md:71 and :93): an operand of ||, ?? or ?: is
// judged where the expression lands (operandLanding). A declared slot judges it against the slot's type, a
// spread or a for...of judges it as a read, and anything else keeps the operator's implied type.

// operandLandingInvariant is the words refuseWidening refuses a mutable widening in, after the position.
func operandLandingInvariant(own, seen, write, read string) string {
	return " Adamic 0.1 refuses a value of type " + own + " seen as " + seen + ", which can write " + write + " where " + read + " is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"
}

// operandLandingMap is a map the programs below read from, with one key set and one missing, so each
// operand is taken once.
const operandLandingMap = "const m = new Map<string, string[]>();\nm.set('a', ['x', 'y']);\n"

func TestOperandLandingLowers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		// A shared array, not a fresh literal (item 2's rule covers that), returned under a declared
		// readonly result. Green on fb0c7074 too: impliedTarget already takes an array-typed slot for an
		// operand of || (invariance.go, its KindBinaryExpression case), so no mutant of operandLanding
		// alone turns it red. The nested row below is the one that pins the landing.
		{"declared readonly", "const none: never[] = [];\nfunction names(lists: Map<string, string[]>, k: string): readonly string[] { return lists.get(k) || none; }\n" + operandLandingMap + "console.log(String(names(m, 'a').length + names(m, 'b').length));\n"},
		// The same slot reached through a ?: nested in the ||, whose branch tsc types by the conditional's
		// own string[]. Mutant: operandLanding's KindReturnStatement case answering no landing, or the walk
		// not passing the landing up through the parentheses and the ||, which refuses none as before.
		{"declared readonly nested", "const none: never[] = [];\nfunction names(lists: Map<string, string[]>, k: string, flag: boolean): readonly string[] { return lists.get(k) || (flag ? none : ['z']); }\n" + operandLandingMap + "console.log(String(names(m, 'a', true).length + names(m, 'b', true).length + names(m, 'b', false).length));\n"},
		// A spread reads the operand's elements out once and keeps none of its slots. Mutant:
		// operandLanding's KindSpreadElement case answering no landing, which refuses none as never[]
		// seen as string[].
		{"spread", "const none: never[] = [];\nfunction count(lists: Map<string, string[]>, k: string): number {\n\tconst all = [...(lists.get(k) ?? none)];\n\treturn all.length;\n}\n" + operandLandingMap + "console.log(String(count(m, 'a') + count(m, 'b')));\n"},
		// A for...of reads its iterable's elements out once. Mutant: operandLanding's KindForOfStatement case
		// answering no landing.
		{"for of", "const none: never[] = [];\nfunction show(lists: Map<string, string[]>, k: string): void {\n\tfor (const name of lists.get(k) || none) {\n\t\tconsole.log(name);\n\t}\n}\n" + operandLandingMap + "show(m, 'a');\nshow(m, 'b');\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				lowersAndAgreesWithNodeAs(t, probe.source, extension)
			})
		}
	}
}

func TestOperandLandingRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, where, words string }{
		// The attack: xs, an unannotated binding, holds the conditional's inferred number[], and a push
		// through it writes a number into emptyArray, which tsc reads as never. Refused at the operand, as
		// on fb0c7074. Its annotated twin, `const xs: readonly number[] = ...`, isn't a fixture: tsc itself
		// refuses the push. Mutant: an unannotated binding taken as a landing that skips the operand,
		// operandLanding's KindVariableDeclaration case answering read (or a declared landing whose slot is
		// nil judged nowhere), which lowers it. Taking its inferred type as a declared slot can't turn it
		// green: that type is number[], and never[] seen as number[] is the same refusal.
		{"attack", "const emptyArray: never[] = [];\nfunction pick(flag: boolean) {\n\tconst xs = flag ? [1] : emptyArray;\n\txs.push(2);\n}\npick(false);\nconst seen = emptyArray[0];\nconsole.log(typeof seen);\n", ":3:26:", operandLandingInvariant("never[]", "number[]", "number", "never")},
		// A receiver keeps the operator's type: the push writes through string[] into none. Mutant: a
		// property access's object taken as a read landing in operandLanding.
		{"receiver", "const none: never[] = [];\nfunction add(lists: Map<string, string[]>, k: string): void {\n\t(lists.get(k) || none).push('x');\n}\n" + operandLandingMap + "add(m, 'b');\nconsole.log(String(none.length));\n", ":3:19:", operandLandingInvariant("never[]", "string[]", "string", "never")},
		// A read still walks what the operand holds: each list is bound to the loop's name, which pushes a
		// number into a list of words. Mutant: readWidened answering nil for a paired container instead of
		// walking its elements.
		{"read walks inside", "const words: string[][] = [['a']];\nconst mixed: (string | number)[][] = [];\nfor (const list of words.length > 0 ? words : mixed) {\n\tlist.push(7);\n}\nconsole.log(typeof words[0]?.[1]);\n", ":3:39:", operandLandingInvariant("string[][]", "(string | number)[][]", "string | number", "string")},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				want := "main" + extension + probe.where + probe.words
				var refused *Refused
				if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
					t.Fatalf("got %v, want a refusal ending %q", err, want)
				}
			})
		}
	}
}

// A declared mutable landing stays refused: string[] can write a string into none. The returned value is
// judged against the slot at the return before its operands are visited (freshOrWidened's || arm), so the
// words are the whole value's, as on fb0c7074, and the position isn't pinned. Mutant: that arm skipping
// its right operand, with operandLanding's declared slot taken as a read, which lowers it.
func TestOperandLandingDeclaredMutableRefused(t *testing.T) {
	t.Parallel()
	source := "const none: never[] = [];\nfunction names(lists: Map<string, string[]>, k: string): string[] { return lists.get(k) || none; }\n" + operandLandingMap + "names(m, 'b').push('x');\nconsole.log(String(none.length));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, source, extension)
			want := operandLandingInvariant("string[]", "string[]", "string", "never")
			var refused *Refused
			if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
				t.Fatalf("got %v, want a refusal ending %q", err, want)
			}
		})
	}
}

// Lower's controls on 8b42425d (#r11wppy): a landing that is itself an enum literal or a closed member union
// still meets the enum check, now with the landing as its target, and the open enum is the control that
// lowers (J-nc480ba read all three at main and at the head). An operand of a spread or a for...of whose two
// sides are different containers (an array and a Set) never pairs in readWidened and keeps failing closed,
// at stage 0's union-of-members stop. A change that skipped the enum check for a target operandLanding
// gave would lower the two closed rows.
func TestOperandLandingLowerControls(t *testing.T) {
	t.Parallel()
	const enum = "enum E { A = 0, B = 1, C = 2 }\nconst holder: { k: E } = { k: E.C };\nconst open: E = holder.k;\n"
	for _, probe := range []struct{ name, source, words string }{
		{"closed binding", enum + "const k: E.B | E.C = open || E.B;\nconsole.log(String(k));\n", "an unproven value assigned to a numeric literal or enum member slot E.B | E.C"},
		{"closed argument", enum + "function f(k: E.B | E.C): void {\n\tconsole.log(String(k));\n}\nf(open || E.B);\n", "an unproven value assigned to a numeric literal or enum member slot E.B | E.C"},
		{"spread of array or set", "const flag = Date.now() > 0;\nconst list: number[] = [1, 2];\nconst set: Set<number> = new Set([3]);\nconst all = [...(flag ? list : set)];\nconsole.log(String(all.length));\n", "iterating a union of differently held members"},
		{"for of array or set", "const flag = Date.now() > 0;\nconst list: number[] = [1, 2];\nconst set: Set<number> = new Set([3]);\nfor (const value of flag ? list : set) {\n\tconsole.log(String(value));\n}\n", "for...of over a union of differently held members"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				if err == nil || !strings.Contains(err.Error(), probe.words) {
					t.Fatalf("got %v, want it stopped in %q", err, probe.words)
				}
			})
		}
	}
	for _, extension := range []string{".a", ".ts"} {
		t.Run("open control"+extension, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, enum+"const k: E = open || E.B;\nconsole.log(String(k));\n", extension)
		})
	}
}
