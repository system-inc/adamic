package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// arrayConstructorFile is where the fixtures pretend to live.
//
// `.tsx` because upstream snapshots this corpus as `no_array_constructor.tsx`, and the corpus leans
// on it: `Array<Foo>(1, 2, 3)` is a type-argument call in TS and would parse as a comparison chain
// in a plain script, so a `.ts` fixture file would exercise a different tree for those cases.
const arrayConstructorFile = "/repository/source/ArrayConstructor.tsx"

// The corpus is oxc's, copied rather than rewritten.
//
// Verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_array_constructor.rs`: 32 pass and 31
// fail. The extractor reports one Tester block, and reports the snapshot holding 33 diagnostics from
// those 31 fail inputs, so one finding per input would be wrong here. Two inputs report twice and
// they are the last two, each containing two separate `Array()` calls; those are asserted with two
// message ids below rather than one.
//
// Copied because a fixture a porter invents encodes the same belief as the port, and the case that
// catches a bug is the one nobody would think to write.
func TestNoArrayConstructorFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"an empty constructor", "new Array()", []string{"noArrayConstructor"}},
		{"a constructor with no parentheses", "new Array", []string{"noArrayConstructor"}},
		{"two arguments", "new Array(x, y)", []string{"noArrayConstructor"}},
		{"three arguments", "new Array(0, 1, 2)", []string{"noArrayConstructor"}},

		// The spread family. Each of these counts as one, two, or three arguments syntactically
		// while its runtime length is unknown, which is why a bare count is not the predicate.
		{"a lone spread", "const array = Array(...args);", []string{"noArrayConstructor"}},
		{"two spreads", "const array = Array(...foo, ...bar);", []string{"noArrayConstructor"}},
		{"a lone spread under new", "const array = new Array(...args);", []string{"noArrayConstructor"}},
		{"a value then a spread", "const array = Array(5, ...args);", []string{"noArrayConstructor"}},
		{"two values then a spread", "const array = Array(5, 6, ...args);", []string{"noArrayConstructor"}},
		// Not upstream's. Every spread case it ships has the spread last, so a rule keying on
		// position rather than on count would pass its whole corpus. This one fires on the count
		// alone and pins that the spread is never load bearing above one argument.
		{"a spread that is not last", "const array = Array(...args, 5);", []string{"noArrayConstructor"}},

		// The comment family, which exists to pin where the finding points rather than whether it
		// fires. Leading trivia is not part of the node's own text, so a rule reporting from
		// `node.Pos()` would anchor these at the comment.
		{"a leading comment", "/*a*/Array()", []string{"noArrayConstructor"}},
		{"comments on both sides", "/*a*/Array()/*b*/", []string{"noArrayConstructor"}},
		{"a comment inside the callee", "Array/*a*/()", []string{"noArrayConstructor"}},
		{"comments everywhere", "/*a*//*b*/Array/*c*//*d*/()/*e*//*f*/;/*g*//*h*/", []string{"noArrayConstructor"}},
		{"comments as the only arguments", "Array(/*a*/ /*b*/)", []string{"noArrayConstructor"}},
		{"comments among the arguments", "Array(/*a*/ x /*b*/, /*c*/ y /*d*/)", []string{"noArrayConstructor"}},
		{"comments among arguments and outside", "/*a*/Array(/*b*/ x /*c*/, /*d*/ y /*e*/)/*f*/;/*g*/", []string{"noArrayConstructor"}},
		{"a comment before new", "/*a*/new Array", []string{"noArrayConstructor"}},
		{"comments around a parenless new", "/*a*/new Array/*b*/", []string{"noArrayConstructor"}},
		{"a comment inside new", "new/*a*/Array", []string{"noArrayConstructor"}},
		{"comments everywhere under new", "new/*a*//*b*/Array/*c*//*d*/()/*e*//*f*/;/*g*//*h*/", []string{"noArrayConstructor"}},
		{"new with comments as arguments", "new Array(/*a*/ /*b*/)", []string{"noArrayConstructor"}},
		{"new with comments among arguments", "new Array(/*a*/ x /*b*/, /*c*/ y /*d*/)", []string{"noArrayConstructor"}},
		{"new with comments throughout", "new/*a*/Array(/*b*/ x /*c*/, /*d*/ y /*e*/)/*f*/;/*g*/", []string{"noArrayConstructor"}},

		// The same shapes terminated by a semicolon, which upstream carries separately because the
		// statement boundary is what the fixer has to reason about.
		{"an empty constructor as a statement", "new Array();", []string{"noArrayConstructor"}},
		{"an empty call as a statement", "Array();", []string{"noArrayConstructor"}},
		{"two arguments as a statement", "new Array(x, y);", []string{"noArrayConstructor"}},
		{"a two argument call as a statement", "Array(x, y);", []string{"noArrayConstructor"}},
		{"three arguments as a statement", "new Array(0, 1, 2);", []string{"noArrayConstructor"}},
		{"a three argument call as a statement", "Array(0, 1, 2);", []string{"noArrayConstructor"}},

		// The two inputs behind the extractor's discrepancy line. Each holds two `Array()` calls,
		// one inside the function or object and one after the `as` expression, so each reports
		// twice. Asserting a single finding here would pass on a rule that missed the second.
		{"two calls around an as expression", `
                        (function () {
                            Fn
                            Array() // ";" required
                        }) as Fn
                        Array() // ";" not required
                        `, []string{"noArrayConstructor", "noArrayConstructor"}},
		{"two calls around an as expression in a method", `
                        ({
                            foo() {
                                Object
                                Array() // ";" required
                            }
                        }) as Object
                        Array() // ";" not required
                        `, []string{"noArrayConstructor", "noArrayConstructor"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoArrayConstructor, arrayConstructorFile, testCase.sourceText),
				testCase.wantIds...)
		})
	}
}

// The clean cases, verbatim from upstream. These are the discrimination and each fails differently.
//
// The single-argument forms are the exemption the rule is built around. The dotted forms are a
// different function entirely. The type-argument forms are TypeScript rather than the untyped
// pitfall. The optional forms are upstream declining to judge a case, not judging it safe.
func TestNoArrayConstructorStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// One argument is the length form, which is the whole reason the rule is not "no Array".
		{"one argument", "new Array(x)"},
		{"one argument called", "Array(x)"},
		{"a literal length", "new Array(9)"},
		{"a literal length called", "Array(9)"},

		// A property access is a different function.
		{"a property of foo", "new foo.Array()"},
		{"a called property of foo", "foo.Array()"},
		{"a property of Array", "new Array.foo"},
		{"a called property of Array", "Array.foo()"},
		{"a property of globalThis", "new globalThis.Array"},

		// The two shadowing cases, and the reason this rule reads the checker at all.
		//
		// Both are zero-argument `new Array`, which is the most-reported shape there is, so nothing
		// syntactic declines them. They are clean only because `Array` names a local binding rather
		// than the global. A first draft of this rule matched on the name alone and reported both;
		// these two lines are what caught it, which is the entire argument for copying upstream's
		// clean corpus verbatim instead of writing the cases a porter thinks are interesting.
		{"a parameter shadowing Array", "const createArray = Array => new Array()"},
		{"a var shadowing Array", "var Array; new Array;"},

		// Upstream omits its `globals: { Array: off }` case with the comment "We do not support
		// globals config in tests", so there is nothing to copy for it and nothing here asserts it.

		{"one argument as a statement", "new Array(x);"},
		{"one argument called as a statement", "Array(x);"},
		{"a literal length as a statement", "new Array(9);"},
		{"a literal length called as a statement", "Array(9);"},
		{"a property of foo as a statement", "new foo.Array();"},
		{"a called property of foo as a statement", "foo.Array();"},
		{"a property of Array as a statement", "new Array.foo();"},
		{"a called property of Array as a statement", "Array.foo();"},

		// Explicit type arguments mean TypeScript is being asked for a typed array. Note the third
		// and fourth of these carry two or three arguments, so they would fire on argument count
		// alone: they are the cases that prove the type-argument check is load bearing.
		{"type arguments with three arguments", "new Array<Foo>(1, 2, 3);"},
		{"type arguments with none", "new Array<Foo>();"},
		{"a typed call with three arguments", "Array<Foo>(1, 2, 3);"},
		{"a typed call with none", "Array<Foo>();"},
		{"a typed call with a length", "Array<Foo>(3);"},

		// The optional family. `Array?.(x)` and `Array?.(9)` would pass on argument count anyway,
		// but `Array?.<Foo>(1, 2, 3)` and the bare optional forms are exempt only because upstream
		// marks optional chaining "TODO: Catch optional chaining cases".
		{"an optional call with one argument", "Array?.(x);"},
		{"an optional call with a length", "Array?.(9);"},
		{"an optional property of foo", "foo?.Array();"},
		{"an optional property of Array", "Array?.foo();"},
		{"an optional call on a property of foo", "foo.Array?.();"},
		{"an optional call on a property of Array", "Array.foo?.();"},
		{"an optional typed call with three arguments", "Array?.<Foo>(1, 2, 3);"},
		{"an optional typed call with none", "Array?.<Foo>();"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoArrayConstructor, arrayConstructorFile, testCase.sourceText))
		})
	}
}

// Calls to something that is simply not `Array`, which upstream's clean corpus never contains.
//
// Every one of upstream's 32 pass cases that involves a name is either `Array` itself or a *member*
// access such as `foo.Array()`, and those are declined a step earlier by the callee-kind check. So
// nothing upstream ships can see the name comparison at all: a mutation replacing `callee.Text() !=
// "Array"` with a test that never fires left the whole corpus green while turning the rule into one
// that reports every zero-argument constructor call in the tree.
//
// That is the most damaging failure this rule has available, and upstream's corpus is blind to it,
// which is a good demonstration that copying the corpus is a floor rather than a ceiling.
func TestNoArrayConstructorDeclinesOtherConstructors(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{
		"foo();",
		"new Foo(1, 2);",
		"new Date();",
		"Object();",
		"String(1, 2);",
		// The near miss, because a prefix or contains comparison would report it.
		"ArrayBuffer(1, 2);",
		"new ArrayLike();",
	} {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoArrayConstructor, arrayConstructorFile, sourceText))
		})
	}
}

// Cases written from reading our code rather than upstream's.
//
// Upstream's optional corpus never pairs `?.` with a shape that would otherwise fire, because every
// optional pass case either carries one argument or carries type arguments, both of which are
// refused a step earlier. So deleting the optional check entirely leaves upstream's whole corpus
// green. `Array?.()` and `Array?.(1, 2)` are the inputs that can see it, and neither is upstream's.
func TestNoArrayConstructorDeclinesBareOptionalCalls(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{"Array?.();", "Array?.(1, 2);", "Array?.(...args);"} {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoArrayConstructor, arrayConstructorFile, sourceText))
		})
	}
}

// A shadow upstream's corpus does not carry, written from reading our code.
//
// Upstream's two shadow cases are both zero-argument. This one passes two arguments through a
// parameter named `Array`, so it exercises the same resolution on a different reported shape, and it
// is the case a reader is most likely to hit in real code: a function that takes a constructor.
func TestNoArrayConstructorDeclinesAShadowedArray(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{
		"function build(Array) { return new Array(1, 2); }",
		"function build(Array) { return Array(); }",
		"import { Array } from './Shim';\nexport const a = new Array(1, 2);\n",
	} {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoArrayConstructor, arrayConstructorFile, sourceText))
		})
	}
}

// The guard that makes every fixture in this file mean something.
//
// This rule declares NeedsTypeChecker, so `rule_testing.Run` hands it a nil checker,
// `resolvesToAGlobal` answers false, and the rule reports nothing at all. Every Fires case above
// would still pass under the syntax-only harness if it were used, because the assertions there
// would be comparing empty against empty only if they expected nothing, and every StaysSilent case
// would pass vacuously. That is the exact shape `TestSyntaxOnlyHarnessCannotProveATypeAwareRule`
// describes in the harness package.
//
// So this states the dependency as an executable claim: the syntax-only harness cannot see this
// rule, and anyone who switches a fixture back to `Run` gets a failure here rather than a green
// suite that proves nothing.
func TestNoArrayConstructorNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoArrayConstructor.NeedsTypeChecker {
		t.Fatal("the rule stopped declaring NeedsTypeChecker; the shadow fixtures no longer mean anything")
	}
	result := rule_testing.Run(t, NoArrayConstructor, arrayConstructorFile, "new Array()")
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected the syntax-only harness to see nothing, got %v", result.MessageIds())
	}
}

// What the suggestion rewrites, which the message-id fixtures above cannot see.
//
// Asserted by applying the repair and comparing the resulting source, never by comparing the fix
// text: a fix writing the right string over the wrong span passes a text comparison. Every pair here
// is upstream's own `fix` table, which is the part of the corpus the extractor's counts do not
// cover.
func TestNoArrayConstructorSuggestsAnArrayLiteral(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSource string
	}{
		{"an empty constructor", "new Array()", "[]"},
		{"no parentheses at all", "new Array", "[]"},
		{"two arguments", "new Array(x, y)", "[x, y]"},
		{"three arguments", "new Array(0, 1, 2)", "[0, 1, 2]"},
		{"values and a spread", "const array = Array(5, 6, ...args);", "const array = [5, 6, ...args];"},

		// The trivia pairs. Each one fails if the replaced span is taken from `node.Pos()` rather
		// than from the start of the node's own text, because that span swallows the leading
		// comment.
		{"a leading comment", "/*a*/Array()", "/*a*/[]"},
		{"comments on both sides", "/*a*/Array()/*b*/", "/*a*/[]/*b*/"},
		{"a comment before new", "/*a*/new Array", "/*a*/[]"},
		{"comments around a parenless new", "/*a*/new Array/*b*/", "/*a*/[]/*b*/"},

		{"an empty constructor as a statement", "new Array();", "[];"},
		{"an empty call as a statement", "Array();", "[];"},
		{"two arguments as a statement", "new Array(x, y);", "[x, y];"},
		{"a two argument call as a statement", "Array(x, y);", "[x, y];"},
		{"three arguments as a statement", "new Array(0, 1, 2);", "[0, 1, 2];"},
		{"a three argument call as a statement", "Array(0, 1, 2);", "[0, 1, 2];"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoArrayConstructor, arrayConstructorFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			suggestions := result.Diagnostics[0].Suggestions
			if len(suggestions) != 1 || len(suggestions[0].Fixes) != 1 {
				t.Fatalf("wanted one suggestion carrying one fix, got %d suggestions", len(suggestions))
			}

			fix := suggestions[0].Fixes[0]
			rewritten := testCase.sourceText[:fix.Range.Pos()] + fix.Text +
				testCase.sourceText[fix.Range.End():]
			if rewritten != testCase.wantSource {
				t.Fatalf("applying the suggestion gave %q, wanted %q", rewritten, testCase.wantSource)
			}
		})
	}
}

// What the suggestion does to comments, measured rather than assumed.
//
// Upstream carries these three under "TODO: Preserve comments around callee" and does not assert
// them, so this is the part of the corpus with no expected answer to copy and the part a porter is
// most likely to get wrong by writing down what they expected.
//
// The first draft of this file asserted all three preserved their comments, on the reasoning that
// copying source text from the first argument to the closing paren carries whatever sits between
// them. Two of the three do. The first does not, and the fixture caught it: comments are trivia, so
// `Array(/*a*/ /*b*/)` has an *empty* argument list, the code takes its no-arguments branch, and the
// comments are dropped. That is upstream's behavior too, which is why it is recorded here as the
// real answer rather than repaired.
//
// The other comment loss is between `Array` and `(`, which no span anchored on the first argument
// can reach, so `Array/*a*/()` also drops to `[]`.
//
// None of this is a correctness problem while the rewrite is a suggestion, because a human sees the
// result before it lands. It would be one if this were ever promoted to a fix.
func TestNoArrayConstructorKeepsCommentsAmongArguments(t *testing.T) {
	t.Parallel()

	cases := []struct{ sourceText, wantSource string }{
		// Comments as the only "arguments" are lost, because they are not arguments.
		{"Array(/*a*/ /*b*/)", "[]"},
		{"Array/*a*/()", "[]"},
		// Comments between real arguments survive, carried by the copied source span.
		{"Array(/*a*/ x /*b*/, /*c*/ y /*d*/)", "[/*a*/ x /*b*/, /*c*/ y /*d*/]"},
		{
			"/*a*/Array(/*b*/ x /*c*/, /*d*/ y /*e*/)/*f*/;/*g*/",
			"/*a*/[/*b*/ x /*c*/, /*d*/ y /*e*/]/*f*/;/*g*/",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoArrayConstructor, arrayConstructorFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			fix := result.Diagnostics[0].Suggestions[0].Fixes[0]
			rewritten := testCase.sourceText[:fix.Range.Pos()] + fix.Text +
				testCase.sourceText[fix.Range.End():]
			if rewritten != testCase.wantSource {
				t.Fatalf("applying the suggestion gave %q, wanted %q", rewritten, testCase.wantSource)
			}
		})
	}
}

// The reason this rewrite is a suggestion rather than a fix, as an executable claim.
//
// This is upstream's own input, from the pair it commented out of its fix table with "These
// currently produce invalid syntax, need to fix the fixer" while still declaring the rule `fix`. So
// oxc applies this rewrite unattended on an input it knows it gets wrong.
//
// The rewrite turns two statements into one. `Fn` followed by `Array()` is an identifier statement
// and a call statement; `Fn` followed by `[]` is the single element access `Fn[]`. The output
// parses, so the edit engine's syntax guard cannot refuse it, which is exactly the failure the
// engine is structurally unable to catch. ESLint reaches the same conclusion from the other side and
// emits `;[...]` for this case under a distinct message id.
//
// The claim under test is not that the rewrite is wrong everywhere. It is that the rule cannot tell
// this case from the ordinary one without looking at what precedes the statement, so nothing here
// may be applied unattended.
func TestNoArrayConstructorNeverProposesAnAutomaticFix(t *testing.T) {
	t.Parallel()

	sourceText := "const value = 1;\nFn\nArray()\n"
	result := rule_testing.RunTyped(t, NoArrayConstructor, arrayConstructorFile, sourceText)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
	}
	if len(result.Diagnostics[0].Fixes) != 0 {
		t.Fatalf("the rule proposed %d unattended fixes; every rewrite here must be a suggestion",
			len(result.Diagnostics[0].Fixes))
	}

	// And the suggestion, if a human took it blindly, would be the meaning change described above.
	// Recorded rather than repaired, so that a later change adding an ASI guard has a fixture that
	// visibly flips instead of a comment nobody reads.
	fix := result.Diagnostics[0].Suggestions[0].Fixes[0]
	rewritten := sourceText[:fix.Range.Pos()] + fix.Text + sourceText[fix.Range.End():]
	if rewritten != "const value = 1;\nFn\n[]\n" {
		t.Fatalf("unexpected rewrite %q", rewritten)
	}
}
