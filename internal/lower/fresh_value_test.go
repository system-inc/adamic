package lower

import (
	"errors"
	"strings"
	"testing"
)

// freshValueInvariant is the invariant rule's words for a number[] seen as (string | number)[] and a
// { a: number } seen as { a: string | number }, after the value's own type, which differs by the node
// that's refused first.
const freshValueInvariant = ", which can write string | number where number is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"

// #k8ed1f2, applying #7n03qzg: under ?. the receiver of a native copy is read without its undefined, so
// maybe?.map(f) is fresh as maybe.map(f) is, and seen as a wider array it lowers and agrees with Node: out
// takes a cat that reaches no Dog[]. The elements are Dogs seen as Animals, interfaces with one
// representation on both sides, so only freshness decides it. Mutant: the non-nullable read removed, `if access.QuestionDotToken
// != nil {` in freshValue made `if false {`, and the receiver reads Dog[] | undefined, no array, so the view
// of a Dog[] as a writable Animal[] refuses.
func TestFreshValueOptionalCopyLowers(t *testing.T) {
	t.Parallel()
	source := "interface Animal { name: string }\n" +
		"interface Dog extends Animal { bark: string }\n" +
		"function pick(n: number): Dog[] | undefined { return n > 0 ? [{ name: \"Rex\", bark: \"woof\" }] : undefined; }\n" +
		"const maybe = pick(1);\n" +
		"const out: Animal[] = maybe?.map(d => d) ?? [];\n" +
		"out.push({ name: \"Tom\" });\n" +
		"const none: Animal[] = pick(0)?.map(d => d) ?? [];\n" +
		"console.log(String(out.length));\n" +
		"console.log(String(maybe?.length));\n" +
		"console.log(String(none.length));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, source, extension)
		})
	}
}

// A receiver that isn't an Array or a ReadonlyArray under ?. is no native copy: Pack's own map returns
// the array it holds, so out writes a string among pack's numbers, and the view stays refused. Mutant:
// under ?. freshValue answers true without the array test, and this lowers.
func TestFreshValueOptionalNonArrayStaysRefused(t *testing.T) {
	t.Parallel()
	source := "class Pack {\n" +
		"\theld: number[] = [1, 2];\n" +
		"\tmap(): number[] {\n" +
		"\t\treturn this.held;\n" +
		"\t}\n" +
		"}\n" +
		"function pick(n: number): Pack | undefined { return n > 0 ? new Pack() : undefined; }\n" +
		"const pack = pick(1);\n" +
		"const out: (number | string)[] = pack?.map() ?? [];\n" +
		"out.push('s');\n" +
		"console.log(String(out.length));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, source, extension)
			where := "main" + extension + ":9:34: Adamic 0.1 refuses a value of type "
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), where) || !strings.HasSuffix(err.Error(), freshValueInvariant) {
				t.Fatalf("got %v, want a refusal at %q ending %q", err, where, freshValueInvariant)
			}
		})
	}
}

// #k8ed1f2, applying #7n03qzg: an && is its left's falsy arms or its right, so a literal made right there
// on the right keeps its freshness, as an operand of || does, and moduleNameResolver.ts's
// `resolved && { ... }` shape lowers and agrees with Node. Mutant: the && arm removed, the
// `case ast.KindAmpersandAmpersandToken:` arm in freshOrWidened deleted, and the && is judged by its
// own type, { a: number } seen as { a: string | number }, which refuses.
func TestFreshValueAndRightLiteralLowers(t *testing.T) {
	t.Parallel()
	source := "function pick(n: number): { on: boolean } | undefined { return n > 0 ? { on: true } : undefined; }\n" +
		"const r: { a: number | string } | undefined = pick(1) && { a: 1 };\n" +
		"const s: { a: number | string } | undefined = pick(0) && { a: 2 };\n" +
		"if (r !== undefined) r.a = 'x';\n" +
		"console.log(String(r === undefined ? 'none' : r.a));\n" +
		"console.log(String(s === undefined ? 'none' : s.a));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, source, extension)
		})
	}
}

// A held object on the right of && is the same object under both names: r.a = 'x' writes a string where
// held.a reads a number, so the view stays refused in the invariant words. Mutant: the && arm returns nil
// without judging its right, and this lowers.
func TestFreshValueAndRightHeldStaysRefused(t *testing.T) {
	t.Parallel()
	source := "function pick(n: number): { on: boolean } | undefined { return n > 0 ? { on: true } : undefined; }\n" +
		"const held = { a: 1 };\n" +
		"const r: { a: number | string } | undefined = pick(1) && held;\n" +
		"if (r !== undefined) r.a = 'x';\n" +
		"console.log(typeof held.a);\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, source, extension)
			where := "main" + extension + ":3:47: Adamic 0.1 refuses a value of type "
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), where) || !strings.HasSuffix(err.Error(), freshValueInvariant) {
				t.Fatalf("got %v, want a refusal at %q ending %q", err, where, freshValueInvariant)
			}
		})
	}
}

// #k8ed1f2, the && gate: tsc's getEmitFlags shape, `emitNode && emitNode.flags || 0` returned into a numeric enum,
// lowers on main, since today's walk of the whole && value lets it through. freshOrWidened's && arm judges each
// operand on its own, which reaches the enum read emitNode.flags, a check the whole walk never makes, so the arm is
// consulted only where that walk refuses. Mutant: the gate dropped, `if l.widened(own, contextual, ...) == nil {`
// made `if false {`, and the read is refused as an unproven value in an enum slot (the census measured 14 such
// stops across tsc without the gate).
func TestFreshValueAndGateKeepsTheWholeWalk(t *testing.T) {
	t.Parallel()
	source := "enum Flags { None = 0, A = 1, B = 2 }\n" +
		"interface EmitNode { flags: Flags }\n" +
		"interface Holder { emitNode?: EmitNode }\n" +
		"function getFlags(node: Holder): Flags {\n" +
		"\tconst emitNode = node.emitNode;\n" +
		"\treturn emitNode && emitNode.flags || 0;\n" +
		"}\n" +
		"console.log(String(getFlags({ emitNode: { flags: Flags.A } })));\n" +
		"console.log(String(getFlags({})));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, source, extension)
		})
	}
}
