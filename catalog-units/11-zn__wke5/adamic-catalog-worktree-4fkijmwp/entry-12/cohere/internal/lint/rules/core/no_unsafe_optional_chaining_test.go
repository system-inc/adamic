package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// unsafeOptionalChainingFile is where the fixtures pretend to live.
//
// A `.ts` extension rather than `.js`, because four upstream cases carry type arguments
// (`x?.f<T>()`, `a?.c?.b<c>`) that only parse as TypeScript. Upstream snapshots this rule under
// `.tsx` for the same reason.
const unsafeOptionalChainingFile = "/repository/source/Chain.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_unsafe_optional_chaining.rs`: 64 pass
// and 18 fail, against a snapshot recording 20 diagnostics. The extractor flagged that
// discrepancy, and the two inputs reporting twice were recovered by aligning each snapshot entry
// on the source line it prints rather than by walking the list in order. They are
// `(obj?.foo && obj?.baz).bar` and `(foo ? obj?.foo : obj?.bar).bar`, and both are in
// TestNoUnsafeOptionalChainingReportsPerChain below rather than here, because this test asserts
// one finding per input.
func TestNoUnsafeOptionalChainingFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// Property access on a parenthesized chain. The parentheses end the chain, so `.bar`
		// runs against `undefined` rather than short-circuiting with it.
		{"property access on the result", "(obj?.foo).bar"},
		{"a call of the result", "(obj?.foo)();"},
		{"construction from the result", "new (obj?.foo)();"},
		{"a template tag", "(obj?.foo)`template`"},
		{"a heritage clause", "class A extends obj?.foo {}"},
		{"destructuring a declaration", "const {foo} = obj?.bar;"},
		{"destructuring an assignment", "({foo} = obj?.bar);"},
		{"the right of in", "foo in obj?.bar;"},
		{"the right of for-of", "for (foo of obj?.bar) {}"},
		{"a with statement", "with (obj?.foo) {};"},
		{"a with statement through await", "async function foo() { with ( await obj?.foo) {}; }"},
		{"array spread", "const a = [...obj?.foo];"},
		{"array spread after another", "const b = [...c, ...obj?.foo];"},
		{"array spread in a second declarator", "const s = [], t = [...obj?.foo];"},
		{"array spread of a parenthesized chain", "const c = () => ([...(obj?.foo)]);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoUnsafeOptionalChaining, unsafeOptionalChainingFile,
					testCase.sourceText), "unsafeOptionalChain")
		})
	}
}

// The two inputs the snapshot records twice, which is the whole reason 18 inputs make 20
// diagnostics.
//
// Both are operators that propagate the short circuit through more than one branch: `&&` reaches
// its left and its right, and a conditional reaches its consequent and its alternate. Each branch
// holding a chain is separately unsafe, so the rule reports per chain rather than per unsafe use.
// A port anchoring on the enclosing member expression would report each of these once and pass a
// fixture that asserted one finding.
func TestNoUnsafeOptionalChainingReportsPerChain(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"both sides of a logical and", "(obj?.foo && obj?.baz).bar"},
		{"both branches of a conditional", "(foo ? obj?.foo : obj?.bar).bar"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoUnsafeOptionalChaining, unsafeOptionalChainingFile,
					testCase.sourceText), "unsafeOptionalChain", "unsafeOptionalChain")
		})
	}
}

// The clean cases, which are the whole discrimination and outnumber the firing ones four to one.
//
// Each encodes a distinction that is easy to get wrong in the other direction: a chain whose
// result is consumed by another `?.` is safe, because the second link short-circuits too. A chain
// behind `??` or `||` is safe on the right, because the operator supplies the fallback. Arithmetic
// is safe under the default options, which is what the option exists to change.
func TestNoUnsafeOptionalChainingStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"no chain at all", "var foo;"},
		{"a class declaration", "class Foo {}"},
		{"a boolean coercion", "!!obj?.foo"},
		// The call is part of the chain, so it short-circuits rather than throwing.
		{"an optional call in the chain", "obj?.foo();"},
		{"an explicitly optional call", "obj?.foo?.();"},
		{"a nullish fallback before a call", "(obj?.foo ?? bar)();"},
		{"an optional call of a parenthesized chain", "(obj?.foo)?.()"},
		{"a nullish fallback of two chains", "(obj?.foo ?? bar?.baz)?.()"},
		{"an optional call on a plain member", "(obj.foo)?.();"},
		// The trap case. `.bar` here is inside the chain, so it never runs on `undefined`.
		// A port reading IsOptionalChain for the parent's optionality reports this.
		{"a continued chain", "obj?.foo.bar;"},
		{"an explicitly optional continuation", "obj?.foo?.bar;"},
		{"an optional access on a parenthesized chain", "(obj?.foo)?.bar;"},
		{"an optional access continued", "(obj?.foo)?.bar.baz;"},
		{"an optional call then access", "(obj?.foo)?.().bar"},
		{"a nullish fallback before access", "(obj?.foo ?? bar).baz;"},
		{"a nullish fallback before a template tag", "(obj?.foo ?? val)`template`"},
		{"a nullish fallback before construction", "new (obj?.foo ?? val)()"},
		{"plain construction", "new bar();"},
		{"a chained call of a call", "obj?.foo?.()();"},
		{"destructuring a defaulted value", "const {foo} = obj?.baz || {};"},
		// Assignment to a plain binding is safe: `undefined` is a legal value to hold.
		{"a plain declaration", "const foo = obj?.bar"},
		{"a plain assignment", "foo = obj?.bar"},
		{"assignment to a member", "foo.bar = obj?.bar"},
		{"call spread with a fallback", "bar(...obj?.foo ?? []);"},
		// Object spread of `undefined` is legal and yields nothing, unlike array spread.
		{"object spread", "var bar = {...foo?.bar};"},
		{"a chain on the left of in", "foo?.bar in {};"},
		{"a chain on the left of a relational", "foo?.bar < foo?.baz;"},
		{"a chain on the left of less-or-equal", "foo?.bar <= foo?.baz;"},
		{"a chain on the left of greater", "foo?.bar > foo?.baz;"},
		{"a chain on the left of greater-or-equal", "foo?.bar >= foo?.baz;"},
		{"an array pattern default", "[foo = obj?.bar] = [];"},
		{"an array pattern default on a member", "[foo.bar = obj?.bar] = [];"},
		{"an object pattern shorthand default", "({foo = obj?.bar} = obj);"},
		{"an object pattern default on a member", "({foo: obj.bar = obj?.baz} = obj);"},
		// Only the last expression of a sequence is the value, so a chain earlier is discarded.
		{"a chain that is not the sequence result", "(foo?.bar, bar)();"},
		{"a logical or fallback before access", "(obj?.foo || bar).baz;"},
		{"a conditional with no chain in either branch", "(foo?.bar ? baz : qux)();"},
		{"awaited chains in safe positions", "\n        async function func() {\n          await obj?.foo();\n          await obj?.foo?.();\n          (await obj?.foo)?.();\n          (await obj?.foo)?.bar;\n          await bar?.baz;\n          await (foo ?? obj?.foo.baz);\n          (await bar?.baz ?? bar).baz;\n          (await bar?.baz ?? await bar).baz;\n          await (foo?.bar ? baz : qux);\n        }\n        "},
		{"a chain of nullish fallbacks", "(obj?.foo ?? bar?.baz ?? qux)();"},
		{"nullish inside logical or", "((obj?.foo ?? bar?.baz) || qux)();"},
		{"logical or inside logical or", "((obj?.foo || bar?.baz) || qux)();"},
		{"logical and inside logical or", "((obj?.foo && bar?.baz) || qux)();"},
		// Arithmetic is clean under the default options. Each operator is listed separately
		// because the option's operator set is the thing a port drops silently.
		{"subtraction", "obj?.foo - bar;"},
		{"addition", "obj?.foo + bar;"},
		{"multiplication", "obj?.foo * bar;"},
		{"division", "obj?.foo / bar;"},
		{"remainder", "obj?.foo % bar;"},
		{"exponentiation", "obj?.foo ** bar;"},
		{"unary plus", "+obj?.foo;"},
		{"unary minus", "-obj?.foo;"},
		{"addition assignment", "bar += obj?.foo;"},
		{"subtraction assignment", "bar -= obj?.foo;"},
		{"remainder assignment", "bar %= obj?.foo;"},
		{"exponentiation assignment", "bar **= obj?.foo;"},
		{"multiplication assignment", "bar *= obj?.boo"},
		{"division assignment", "bar /= obj?.boo"},
		{"awaited arithmetic", "async function func() {\n            await obj?.foo + await obj?.bar;\n            await obj?.foo - await obj?.bar;\n            await obj?.foo * await obj?.bar;\n            +await obj?.foo;\n            -await obj?.foo;\n            bar += await obj?.foo;\n            bar -= await obj?.foo;\n            bar %= await obj?.foo;\n            bar **= await obj?.foo;\n            bar *= await obj?.boo;\n            bar /= await obj?.boo;\n        }\n        "},
		// Type arguments on a chain link. These parse only as TypeScript and each is a chain
		// whose result is consumed inside the chain, so none is unsafe.
		{"a typed optional call", "x?.f<T>();"},
		{"an explicitly optional typed call", "x?.f?.<T>();"},
		{"a typed optional call on an identifier", "f?.<Q>();"},
		{"a type argument on a chain", "a?.c?.b<c>"},
		{"object spread in a declaration", "const baz = {...obj?.foo };"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoUnsafeOptionalChaining, unsafeOptionalChainingFile,
					testCase.sourceText))
		})
	}
}

// The option, in all three of its states.
//
// Upstream carries the same source under an empty object and under an explicit false, which is
// how it pins that an options object that omits the key still gets the default rather than the
// zero value of whatever the decoder produced.
func TestNoUnsafeOptionalChainingArithmeticOption(t *testing.T) {
	t.Parallel()

	// Omitting the option entirely leaves arithmetic clean.
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoUnsafeOptionalChaining,
		unsafeOptionalChainingFile, "obj?.foo - bar;"))

	// An options object that says nothing about it, upstream's `[{}]`.
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoUnsafeOptionalChaining,
		unsafeOptionalChainingFile, "obj?.foo - bar;", NoUnsafeOptionalChainingOptions{}))

	// Explicit false, upstream's `[{ "disallowArithmeticOperators": false }]`.
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoUnsafeOptionalChaining,
		unsafeOptionalChainingFile, "obj?.foo - bar;",
		NoUnsafeOptionalChainingOptions{DisallowArithmeticOperators: false}))

	// Explicit true, upstream's only options-carrying fail case. Note the distinct message id:
	// arithmetic on `undefined` yields NaN rather than throwing, so it is a different judgment.
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoUnsafeOptionalChaining,
		unsafeOptionalChainingFile, "bar + obj?.foo;",
		NoUnsafeOptionalChainingOptions{DisallowArithmeticOperators: true}), "unsafeArithmetic")
}

// Where the finding points, which the message-id assertions above cannot see.
//
// This rule carries no fix, and that is exactly why the spans are asserted rather than assumed:
// a rule whose defect is where it points passes a complete fixture pair while being wrong, and
// gating span checks on "does it have a repair" is how that ships.
//
// The wanted text is upstream's, read off the snapshot's own underline rather than derived from
// this port. Every one of the twenty diagnostics underlines the chain expression itself, never
// the enclosing member access or call that made it unsafe. That is a real design decision: the
// chain is the thing the author must change, and the enclosing operator is only the evidence.
func TestNoUnsafeOptionalChainingReportsTheChainSpan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantTexts  []string
	}{
		{"property access on the result", "(obj?.foo).bar", []string{"obj?.foo"}},
		{"construction from the result", "new (obj?.foo)();", []string{"obj?.foo"}},
		{"a heritage clause", "class A extends obj?.foo {}", []string{"obj?.foo"}},
		{"the right of in", "foo in obj?.bar;", []string{"obj?.bar"}},
		{"the right of for-of", "for (foo of obj?.bar) {}", []string{"obj?.bar"}},
		{"a with statement", "with (obj?.foo) {};", []string{"obj?.foo"}},
		{"destructuring a declaration", "const {foo} = obj?.bar;", []string{"obj?.bar"}},
		{"array spread", "const a = [...obj?.foo];", []string{"obj?.foo"}},
		// Reported through await: the span is the chain, not the await expression that wraps it.
		{"a with statement through await",
			"async function foo() { with ( await obj?.foo) {}; }", []string{"obj?.foo"}},
		// The spread of a parenthesized chain, where a port anchoring on the spread argument
		// would underline `(obj?.foo)` including its parentheses.
		{"array spread of a parenthesized chain",
			"const c = () => ([...(obj?.foo)]);", []string{"obj?.foo"}},
		// The two-finding inputs, where order and both spans matter. A port reporting the
		// enclosing expression twice would produce two identical spans here.
		{"both sides of a logical and", "(obj?.foo && obj?.baz).bar",
			[]string{"obj?.foo", "obj?.baz"}},
		{"both branches of a conditional", "(foo ? obj?.foo : obj?.bar).bar",
			[]string{"obj?.foo", "obj?.bar"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnsafeOptionalChaining, unsafeOptionalChainingFile,
				testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantTexts) {
				t.Fatalf("wanted %d diagnostics, got %d", len(testCase.wantTexts),
					len(result.Diagnostics))
			}
			for index, wantText := range testCase.wantTexts {
				diagnostic := result.Diagnostics[index]
				reported := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if reported != wantText {
					t.Fatalf("diagnostic %d underlined %q, wanted %q", index, reported, wantText)
				}
			}
		})
	}
}

// The arithmetic finding's span, asserted separately because it is produced by a different arm
// and could point at the operator or the whole binary expression without any fixture noticing.
func TestNoUnsafeOptionalChainingReportsTheArithmeticSpan(t *testing.T) {
	t.Parallel()

	source := "bar + obj?.foo;"
	result := rule_testing.RunWithOptions(t, NoUnsafeOptionalChaining, unsafeOptionalChainingFile,
		source, NoUnsafeOptionalChainingOptions{DisallowArithmeticOperators: true})
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
	}
	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "obj?.foo" {
		t.Fatalf("underlined %q, wanted %q", reported, "obj?.foo")
	}
}

// Cases written from reading our own code rather than upstream's, each covering a context the
// imported corpus exercises thinly or not at all.
//
// The named failure shape for this rule is covering a subset of the unsafe contexts silently: a
// port that handles member access and calls but forgets `instanceof`, or handles array spread but
// not a call's spread, passes most of a 64-case clean corpus and reports nothing on the gap.
// Upstream has a fail case for `in` but none for `instanceof`, none for a computed access, none
// for a call argument's spread, and none for a class expression's heritage. Each is below.
func TestNoUnsafeOptionalChainingCoversEveryUnsafeContext(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantId     string
	}{
		// `instanceof` is in ESLint's relational set beside `in`, and upstream tests only `in`.
		// A port reading the operator set from the `in` fixture alone drops this.
		{"the right of instanceof", "foo instanceof obj?.bar;", "unsafeOptionalChain"},
		// A computed access is a different node kind from a dotted one and needs its own arm.
		{"computed access on the result", "(obj?.foo)[key];", "unsafeOptionalChain"},
		// Spread into a call, which reaches the SpreadElement through a different parent than the
		// array-literal cases upstream covers. ESLint's guard is "parent is not an object
		// expression", so a call's argument spread is unsafe.
		{"spread into a call argument", "bar(...obj?.foo);", "unsafeOptionalChain"},
		// A class expression rather than a declaration. ESLint registers both; a port listening
		// only for the declaration reports nothing here.
		{"a class expression heritage", "const A = class extends obj?.foo {};", "unsafeOptionalChain"},
		// Array destructuring rather than object destructuring, the other pattern kind.
		{"array destructuring a declaration", "const [foo] = obj?.bar;", "unsafeOptionalChain"},
		{"array destructuring an assignment", "[foo] = obj?.bar;", "unsafeOptionalChain"},
		// A destructuring default, ESLint's AssignmentPattern arm. Upstream's clean cases cover
		// defaults whose target is an identifier; this one's target is a pattern.
		{"a destructuring pattern default", "function f({a} = obj?.bar) {}", "unsafeOptionalChain"},
		// Arithmetic reached through each operator the option names, on the left and the right.
		// Upstream's single fail case is `bar + obj?.foo`, so five of the six operators and the
		// left-hand position are untested there.
		{"arithmetic subtraction under the option", "obj?.foo - bar;", "unsafeArithmetic"},
		{"arithmetic exponentiation under the option", "bar ** obj?.foo;", "unsafeArithmetic"},
		{"unary minus under the option", "-obj?.foo;", "unsafeArithmetic"},
		{"an arithmetic assignment under the option", "bar += obj?.foo;", "unsafeArithmetic"},
		// Each compound assignment operator gets its own case, because the set is a map and
		// dropping one entry is invisible to a fixture that exercises a different key. A sweep
		// removing `**=` survived the whole suite until these were added; `+=` alone was
		// standing in for six operators.
		{"a subtraction assignment under the option", "bar -= obj?.foo;", "unsafeArithmetic"},
		{"a multiplication assignment under the option", "bar *= obj?.foo;", "unsafeArithmetic"},
		{"a division assignment under the option", "bar /= obj?.foo;", "unsafeArithmetic"},
		{"a remainder assignment under the option", "bar %= obj?.foo;", "unsafeArithmetic"},
		{"an exponentiation assignment under the option", "bar **= obj?.foo;", "unsafeArithmetic"},
		// The same coverage for the binary operators, where upstream's only fail case is `+`.
		{"arithmetic multiplication under the option", "obj?.foo * bar;", "unsafeArithmetic"},
		{"arithmetic division under the option", "obj?.foo / bar;", "unsafeArithmetic"},
		{"arithmetic remainder under the option", "obj?.foo % bar;", "unsafeArithmetic"},
		// Unary plus as well as unary minus, the other half of that condition.
		{"unary plus under the option", "+obj?.foo;", "unsafeArithmetic"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Every case runs with the option on, which changes nothing for the usage cases
			// (their contexts are unsafe regardless) and enables the arithmetic ones.
			result := rule_testing.RunWithOptions(t, NoUnsafeOptionalChaining,
				unsafeOptionalChainingFile, testCase.sourceText,
				NoUnsafeOptionalChainingOptions{DisallowArithmeticOperators: true})
			rule_testing.ExpectFindings(t, result, testCase.wantId)
		})
	}
}

// The contexts that must stay clean even with the option on, paired with the cases above.
//
// The sharpest is the object spread: `{...undefined}` is legal and yields an empty object, while
// `[...undefined]` throws. A port that treats every spread alike reports this, and upstream's
// two object-spread clean cases are the only thing pinning the distinction.
func TestNoUnsafeOptionalChainingDeclinesSafeContextsUnderTheOption(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"object spread stays legal", "const a = {...obj?.foo};"},
		// The left of a relational operator is not the thing being reached into.
		{"the left of instanceof", "obj?.foo instanceof Bar;"},
		// A comparison is not arithmetic, so the option does not reach it.
		{"a comparison under the option", "obj?.foo === bar;"},
		// Bitwise and shift operators are absent from the option's set: ESLint names exactly
		// `+ - / * % **`. A port matching "any binary operator that is not relational" reports
		// these.
		{"a bitwise or under the option", "obj?.foo | bar;"},
		{"a left shift under the option", "obj?.foo << bar;"},
		// `typeof` and `!` are unary but not arithmetic, and both are safe on `undefined`.
		{"typeof under the option", "typeof obj?.foo;"},
		{"logical not under the option", "!obj?.foo;"},
		// A string concatenation reached by `+=` on a plain binding is still arithmetic to this
		// rule, but a plain `=` is not, and that boundary is what this pins.
		{"a plain assignment under the option", "bar = obj?.foo;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoUnsafeOptionalChaining,
				unsafeOptionalChainingFile, testCase.sourceText,
				NoUnsafeOptionalChainingOptions{DisallowArithmeticOperators: true}))
		})
	}
}

// The trap this port was most likely to fall into, pinned as its own test so a regression names
// itself rather than arriving as one failure among eighty.
//
// typescript-go propagates NodeFlagsOptionalChain down every link of a chain, so
// ast.IsOptionalChain is true for `obj?.foo.bar` as a whole even though `.bar` carries no `?.`.
// Translating upstream's per-node `optional` boolean as IsOptionalChain therefore treats `.bar`
// as optional and reports nothing where it should, or treats the outer access as a fresh unsafe
// context and reports where it should not. IsOptionalChainRoot is the faithful translation, and
// a probe over this corpus confirmed it tracks QuestionDotToken() exactly.
//
// The same distinction, from the other side: parentheses end a chain, so `(obj?.foo).bar` has a
// genuinely non-optional access and does report. The pair is what makes the discrimination
// visible; either case alone passes under the wrong predicate.
func TestNoUnsafeOptionalChainingDistinguishesChainRootFromChainMembership(t *testing.T) {
	t.Parallel()

	// Inside one chain: `.bar` short-circuits with the rest and is safe.
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoUnsafeOptionalChaining,
		unsafeOptionalChainingFile, "obj?.foo.bar;"))
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoUnsafeOptionalChaining,
		unsafeOptionalChainingFile, "obj?.foo.bar.baz;"))
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoUnsafeOptionalChaining,
		unsafeOptionalChainingFile, "obj?.[key].bar;"))
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoUnsafeOptionalChaining,
		unsafeOptionalChainingFile, "obj?.foo.bar();"))

	// Parenthesized, which ends the chain: the same access is now unsafe.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoUnsafeOptionalChaining,
		unsafeOptionalChainingFile, "(obj?.foo).bar;"), "unsafeOptionalChain")
	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoUnsafeOptionalChaining,
		unsafeOptionalChainingFile, "(obj?.[key]).bar;"), "unsafeOptionalChain")
}
