package lower

import (
	"errors"
	"strings"
	"testing"
)

// Every generator form this slice doesn't lower stops at lowering, where lowering reaches its
// declaration, in words that name the form, never in the refusal walk and never compiled as an
// ordinary function (Lower, open question 7 on #qk8rztp). Each is lowered with a yield and with
// none (a function* that only returns, which as an ordinary function would return undefined where
// Node returns a generator, and which yieldStatement's backstop can't see), used inside a plain
// function and inside an instantiated generic one. A class's methods are declared at the top level,
// since a class inside a function stops as that first.
//
// Mutants, read by testing. signature's NotYet for a form it doesn't lower dropped: every method
// row goes red (object literal, class, static and generic class), the ones with no yield compiling
// into a program checkGeneratorRoads lets through, the ones with a yield stopping only at the
// for-of, in other words; the other forms stop earlier, each where its declaration is reached.
// nestedDeclarations' generator stop dropped: the block rows stop as a block-scoped function. The
// block statement's stop dropped: the top-level block row compiles. The module's stop dropped: the
// unnamed default export crashes declareModule. functionExpression's own stop dropped is
// equivalent: its signature then says the same words.
//
// A plain nested declaration lowers since S3a, and its rows are TestNestedGeneratorsLower's and
// TestNestedGeneratorRoadsAreNotYet's (generators_nested_test.go); a generic one still stops here.
// A generator written in place and called lowers since S3b in a plain function, and in a generic one
// stops in its own words, both in generators_in_place_test.go.
func TestGeneratorFormsStopAtLowering(test *testing.T) {
	test.Parallel()
	const generator = "Generator<number, void, unknown>"
	// Each form's use: a for-of over its call when its body yields, and its call dropped as a
	// statement when it doesn't, which as an ordinary function would run its body and return.
	forms := []struct{ name, declared, consumed, dropped, words, yields string }{
		{
			"nested generic declaration",
			"",
			"function* inner<U>(seed: U): Generator<U, void, unknown> { BODY }\nfor (const value of inner(1)) console.log(String(value));",
			"function* inner<U>(seed: U): Generator<U, void, unknown> { BODY }\ninner(1);",
			"declared generic inside a function",
			"yield seed;",
		},
		{
			"block declaration",
			"",
			"if (String(1).length > 0) { function* inner(): " +
				generator +
				" { BODY } for (const value of inner()) console.log(String(value)); }",
			"if (String(1).length > 0) { function* inner(): " + generator + " { BODY } inner(); }",
			"declared inside a function or a block",
			"yield 1;",
		},
		{
			"function expression",
			"",
			"const make = function* (): " +
				generator +
				" { BODY };\nfor (const value of make()) console.log(String(value));",
			"const make = function* (): " + generator + " { BODY };\nmake();",
			"written as an expression",
			"yield 1;",
		},
		{
			"object literal method",
			"",
			"const box = { *entries(): " +
				generator +
				" { BODY } };\nfor (const value of box.entries()) console.log(String(value));",
			"const box = { *entries(): " + generator + " { BODY } };\nbox.entries();",
			"written as an object literal's method",
			"yield 1;",
		},
		{
			"class method",
			"class Box { *each(): " + generator + " { BODY } }\n",
			"for (const value of new Box().each()) console.log(String(value));",
			"new Box().each();",
			"written as a method",
			"yield 1;",
		},
		{
			"static method",
			"class Box { static *each(): " + generator + " { BODY } }\n",
			"for (const value of Box.each()) console.log(String(value));",
			"Box.each();",
			"written as a static method",
			"yield 1;",
		},
		{
			"generic class method",
			"class Box<T> { constructor(readonly seed: T) {} *each(): Generator<T, void, unknown> { BODY } }\n",
			"for (const value of new Box(1).each()) console.log(String(value));",
			"new Box(1).each();",
			"written as a method",
			"yield this.seed;",
		},
		{
			"namespace declaration",
			"namespace Numbers { export function* each(): " + generator + " { BODY } }\n",
			"for (const value of Numbers.each()) console.log(String(value));",
			"Numbers.each();",
			"declared inside a namespace",
			"yield 1;",
		},
	}
	hosts := map[string]func(declared, used string) string{
		"in a plain function": func(declared, used string) string {
			return declared + "function host(): void {\n" + used + "\n}\nhost();\n"
		},
		"in an instantiated generic function": func(declared, used string) string {
			return declared + "function host<T>(seed: T): void {\n" + used + "\n}\nhost('seed');\n"
		},
	}
	for _, form := range forms {
		for host, wrap := range hosts {
			for body, use := range map[string][2]string{
				"with a yield":  {form.yields, form.consumed},
				"with no yield": {"return;", form.dropped},
			} {
				test.Run(form.name+" "+host+" "+body, func(test *testing.T) {
					test.Parallel()
					source := wrap(
						strings.ReplaceAll(form.declared, "BODY", use[0]),
						strings.ReplaceAll(use[1], "BODY", use[0]),
					)
					_, err := lowerSource(test, source)
					var gap *NotYet
					if !errors.As(err, &gap) || !strings.Contains(gap.What, "a generator function "+form.words) {
						test.Fatalf("got %v, want a NotYet saying %q", err, form.words)
					}
				})
			}
		}
	}
	// A declaration lowering reaches stops whether or not anything calls it: a nested generic one,
	// which is otherwise put by to instantiate at a call, and one in a block at the top level, which
	// no module declares. export default function* () has no name, and stops where the module
	// declares its functions.
	for _, probe := range []struct{ name, source, words string }{
		{
			"a nested generic generator nothing calls",
			"function host(): void {\nfunction* inner<U>(seed: U): Generator<U, void, unknown> { yield seed; }\nconsole.log('control');\n}\nhost();\n",
			"declared generic inside a function",
		},
		{"a top-level block's generator nothing calls", "if (String(1).length > 0) {\nfunction* inner(): " +
			generator +
			" { return; }\n}\nconsole.log('control');\n", "declared inside a function or a block"},
		{
			"an unnamed default export",
			"export default function* (): " + generator + " { yield 1; }\nconsole.log('control');\n",
			"declared without a name",
		},
	} {
		_, err := lowerSource(test, probe.source)
		var gap *NotYet
		if !errors.As(err, &gap) || !strings.Contains(gap.What, "a generator function "+probe.words) {
			test.Errorf("%s: got %v, want a NotYet saying %q", probe.name, err, probe.words)
		}
	}
}

// Adamic's condition on Lower's ruling (#qk8rztp): with rule 2500 gone, "the walk found nothing"
// must still mean the walk ran over a generic generator nothing instantiates. Its body holding a
// construct the walk refuses for good reads Refused; the same generator with tsc's flatMapIterator's
// shape (yield*, a narrowing continue), and a generic function with createSet's (a nested function*
// and an object literal's *entries()), still uninstantiated, lower with no stop, since nothing
// lowers their bodies. Mutant: the refusal walk pruned at a generic generator's declaration; the
// Refused control lowers, and the test goes red.
func TestUninstantiatedGenericGeneratorsAreWalkedButNotLowered(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(
		test,
		"function* each<T>(items: readonly T[]): Generator<T, void, unknown> { for (const item of items) { yield (item as any); } }\nconsole.log('control');\n",
	)
	var refused *Refused
	if !errors.As(err, &refused) {
		test.Errorf("an uninstantiated generic generator yielding an any cast: got %v, want the walk's refusal", err)
	}
	for name, source := range map[string]string{
		"flatMapIterator's shape": `function* flatMapIterator<T, U>(iter: Iterable<T>, mapfn: (x: T) => readonly U[] | Iterable<U> | undefined): Generator<U, void, unknown> {
    for (const x of iter) {
        const iter2 = mapfn(x);
        if (!iter2) continue;
        yield* iter2;
    }
}
console.log('control');
`,
		"createSet's shape": `function createBag<T>(items: readonly T[]) {
    function* each(): Generator<T, void, unknown> {
        for (const item of items) yield item;
    }
    return {
        size: items.length,
        each,
        *entries(): Generator<[T, T], void, unknown> {
            for (const item of items) yield [item, item];
        },
    };
}
console.log('control');
`,
	} {
		if _, err := lowerSource(test, source); err != nil {
			test.Errorf("%s, uninstantiated: %v", name, err)
		}
	}
}
