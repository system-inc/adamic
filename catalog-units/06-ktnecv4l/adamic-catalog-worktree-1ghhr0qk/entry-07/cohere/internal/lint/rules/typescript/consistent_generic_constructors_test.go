package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const consistentGenericConstructorsFile = "/repository/source/Constructing.ts"

func consistentGenericConstructorsCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// consistentGenericConstructorsSettingsFor routes a mode through the rule's own exported decoder.
//
// Building the options struct directly would leave the decoder untested, and the decoder is where
// the two lines most likely to be wrong live: the bare-string wire shape cohere actually delivers,
// and the fallback that keeps a bare `"error"` meaning upstream's default rather than meaning an
// empty mode that matches no arm.
func consistentGenericConstructorsSettingsFor(t *testing.T, wire string) any {
	t.Helper()
	decoded, err := DecodeConsistentGenericConstructorsOptions([]byte(wire))
	if err != nil {
		t.Fatalf("decoding %q: %v", wire, err)
	}
	return decoded
}

// TestConsistentGenericConstructorsStaysSilent is upstream's passing cases verbatim, in both modes.
//
// One of upstream's forty six is NOT here and its absence is deliberate: it is clean only because
// `isolatedDeclarations` is on, which this harness cannot express. It is recorded in
// TestConsistentGenericConstructorsCannotExpressIsolatedDeclarations below rather than quietly
// dropped or forced green.
func TestConsistentGenericConstructorsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wire       string
	}{
		{sourceText: "const a = new Foo();", wire: "\"constructor\""},
		{sourceText: "const a = new Foo<string>();", wire: "\"constructor\""},
		{sourceText: "const a: Foo<string> = new Foo<string>();", wire: "\"constructor\""},
		{sourceText: "const a: Foo = new Foo();", wire: "\"constructor\""},
		{sourceText: "const a: Bar<string> = new Foo();", wire: "\"constructor\""},
		{sourceText: "const a: Foo = new Foo<string>();", wire: "\"constructor\""},
		{sourceText: "const a: Bar = new Foo<string>();", wire: "\"constructor\""},
		{sourceText: "const a: Bar<string> = new Foo<string>();", wire: "\"constructor\""},
		{sourceText: "const a: Foo<string> = Foo<string>();", wire: "\"constructor\""},
		{sourceText: "const a: Foo<string> = Foo();", wire: "\"constructor\""},
		{sourceText: "const a: Foo = Foo<string>();", wire: "\"constructor\""},
		{sourceText: "\nclass Foo {\n  a = new Foo<string>();\n}\n    ", wire: "\"constructor\""},
		{sourceText: "\nclass Foo {\n  accessor a = new Foo<string>();\n}\n    ", wire: "\"constructor\""},
		{sourceText: "\nfunction foo(a: Foo = new Foo<string>()) {}\n    ", wire: "\"constructor\""},
		{sourceText: "\nfunction foo({ a }: Foo = new Foo<string>()) {}\n    ", wire: "\"constructor\""},
		{sourceText: "\nfunction foo([a]: Foo = new Foo<string>()) {}\n    ", wire: "\"constructor\""},
		{sourceText: "\nclass A {\n  constructor(a: Foo = new Foo<string>()) {}\n}\n    ", wire: "\"constructor\""},
		{sourceText: "\nconst a = function (a: Foo = new Foo<string>()) {};\n    ", wire: "\"constructor\""},
		{sourceText: "\nconst a: Float32Array<ArrayBufferLike> = new Float32Array();\nexport {};\n    ", wire: "\"constructor\""},
		{sourceText: "\nconst a: Float64Array<ArrayBufferLike> = new Float64Array();\nexport {};\n    ", wire: "\"constructor\""},
		{sourceText: "\nconst a: Int16Array<ArrayBufferLike> = new Int16Array();\nexport {};\n    ", wire: "\"constructor\""},
		{sourceText: "\nconst a: Int8Array<ArrayBufferLike> = new Int8Array();\nexport {};\n    ", wire: "\"constructor\""},
		{sourceText: "\nconst a: Uint16Array<ArrayBufferLike> = new Uint16Array();\nexport {};\n    ", wire: "\"constructor\""},
		{sourceText: "\nconst a: Uint32Array<ArrayBufferLike> = new Uint32Array();\nexport {};\n    ", wire: "\"constructor\""},
		{sourceText: "\nconst a: Uint8Array<ArrayBufferLike> = new Uint8Array();\nexport {};\n    ", wire: "\"constructor\""},
		{sourceText: "\nconst a: Uint8ClampedArray<ArrayBufferLike> = new Uint8ClampedArray();\nexport {};\n    ", wire: "\"constructor\""},
		// The isolatedDeclarations case lives in its own test: see the note above.
		{sourceText: "const a = new Foo();", wire: "\"type-annotation\""},
		{sourceText: "const a: Foo<string> = new Foo();", wire: "\"type-annotation\""},
		{sourceText: "const a: Foo<string> = new Foo<string>();", wire: "\"type-annotation\""},
		{sourceText: "const a: Foo = new Foo();", wire: "\"type-annotation\""},
		{sourceText: "const a: Bar = new Foo<string>();", wire: "\"type-annotation\""},
		{sourceText: "const a: Bar<string> = new Foo<string>();", wire: "\"type-annotation\""},
		{sourceText: "const a: Foo<string> = Foo<string>();", wire: "\"type-annotation\""},
		{sourceText: "const a: Foo<string> = Foo();", wire: "\"type-annotation\""},
		{sourceText: "const a: Foo = Foo<string>();", wire: "\"type-annotation\""},
		{sourceText: "const a = new (class C<T> {})<string>();", wire: "\"type-annotation\""},
		{sourceText: "\nclass Foo {\n  a: Foo<string> = new Foo();\n}\n      ", wire: "\"type-annotation\""},
		{sourceText: "\nclass Foo {\n  accessor a: Foo<string> = new Foo();\n}\n      ", wire: "\"type-annotation\""},
		{sourceText: "\nfunction foo(a: Foo<string> = new Foo()) {}\n      ", wire: "\"type-annotation\""},
		{sourceText: "\nfunction foo({ a }: Foo<string> = new Foo()) {}\n      ", wire: "\"type-annotation\""},
		{sourceText: "\nfunction foo([a]: Foo<string> = new Foo()) {}\n      ", wire: "\"type-annotation\""},
		{sourceText: "\nclass A {\n  constructor(a: Foo<string> = new Foo()) {}\n}\n      ", wire: "\"type-annotation\""},
		{sourceText: "\nconst a = function (a: Foo<string> = new Foo()) {};\n      ", wire: "\"type-annotation\""},
		{sourceText: "\nconst [a = new Foo<string>()] = [];\n      ", wire: "\"type-annotation\""},
		{sourceText: "\nfunction a([a = new Foo<string>()]) {}\n      ", wire: "\"type-annotation\""},

		// The ARROW function default parameter, which upstream's corpus does not write in either
		// mode and which its selector excludes. `:matches(FunctionDeclaration,FunctionExpression) >
		// AssignmentPattern` never reaches an arrow, so both spellings are silent, measured on the
		// installed build. A port widening the parameter gate to every parent passes the whole
		// imported corpus and reports these two, which is how this pair came to exist: the mutant
		// that widened the gate survived until they did.
		{sourceText: "const foo = (a: Foo<string> = new Foo()) => {};\n", wire: "\"constructor\""},
		{sourceText: "const foo = (a = new Foo<string>()) => {};\n", wire: "\"type-annotation\""},

		// A getter is not a construction site at all, so the return annotation is out of scope.
		{sourceText: "class A {\n  get foo(): Foo<string> {\n    return new Foo();\n  }\n}\n", wire: "\"constructor\""},

		// A function TYPE's parameter annotation has no initializer, so there is nothing to move.
		{sourceText: "declare const f: (a: Foo<string>) => void;\n", wire: "\"constructor\""},

		// Annotations that are NOT a plain type reference, so there is no single name to compare
		// against the callee and nothing safe to move. Upstream's corpus writes none of these, and
		// a port dropping the type-reference gate reports all four while passing every imported
		// case. Each measured silent on the installed build.
		{sourceText: "const a: Foo<string>[] = new Foo();\n", wire: "\"constructor\""},
		{sourceText: "const a: N.Foo<string> = new Foo();\n", wire: "\"constructor\""},
		{sourceText: "const a: Foo<string> | null = new Foo();\n", wire: "\"constructor\""},
		{sourceText: "const a: readonly Foo<string>[] = new Foo();\n", wire: "\"constructor\""},
	}
	for index, testCase := range cases {
		t.Run(consistentGenericConstructorsCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t,
				ConsistentGenericConstructors, consistentGenericConstructorsFile, testCase.sourceText,
				consistentGenericConstructorsSettingsFor(t, testCase.wire)))
		})
	}
}

// TestConsistentGenericConstructorsFires is upstream's thirty five reporting cases verbatim, with
// their message ids, spans, and the exact text their repair writes.
//
// Eight of them are noFormat cases upstream wrote to pin whitespace handling, which is what makes
// the fix assertions worth more than the id assertions here: `new Map <string, number> ()` and
// `new \n Foo<string> \n ()` both report, and a fixer computing ranges by arithmetic rather than by
// scanning for the brackets writes the wrong text on both while satisfying every id.
func TestConsistentGenericConstructorsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wire       string
		wantId     string
		wantSpan   string
		wantFixed  string
	}{
		{
			sourceText: "const a: Foo<string> = new Foo();",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "a: Foo<string> = new Foo()",
			wantFixed:  "const a = new Foo<string>();",
		},
		{
			sourceText: "const a: Map<string, number> = new Map();",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "a: Map<string, number> = new Map()",
			wantFixed:  "const a = new Map<string, number>();",
		},
		{
			sourceText: "const a: Map <string, number> = new Map();",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "a: Map <string, number> = new Map()",
			wantFixed:  "const a = new Map<string, number>();",
		},
		{
			sourceText: "const a: Map< string, number > = new Map();",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "a: Map< string, number > = new Map()",
			wantFixed:  "const a = new Map< string, number >();",
		},
		{
			sourceText: "const a: Map<string, number> = new Map ();",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "a: Map<string, number> = new Map ()",
			wantFixed:  "const a = new Map<string, number> ();",
		},
		{
			sourceText: "const a: Foo<number> = new Foo;",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "a: Foo<number> = new Foo",
			wantFixed:  "const a = new Foo<number>();",
		},
		{
			sourceText: "const a: /* comment */ Foo/* another */ <string> = new Foo();",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "a: /* comment */ Foo/* another */ <string> = new Foo()",
			wantFixed:  "const a = new Foo/* comment *//* another */<string>();",
		},
		{
			sourceText: "const a: Foo/* comment */ <string> = new Foo /* another */();",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "a: Foo/* comment */ <string> = new Foo /* another */()",
			wantFixed:  "const a = new Foo/* comment */<string> /* another */();",
		},
		{
			sourceText: "const a: Foo<string> = new \n Foo \n ();",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "a: Foo<string> = new \n Foo \n ()",
			wantFixed:  "const a = new \n Foo<string> \n ();",
		},
		{
			sourceText: "\nclass Foo {\n  a: Foo<string> = new Foo();\n}\n      ",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "a: Foo<string> = new Foo();",
			wantFixed:  "\nclass Foo {\n  a = new Foo<string>();\n}\n      ",
		},
		{
			sourceText: "\nclass Foo {\n  [a]: Foo<string> = new Foo();\n}\n      ",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "[a]: Foo<string> = new Foo();",
			wantFixed:  "\nclass Foo {\n  [a] = new Foo<string>();\n}\n      ",
		},
		{
			sourceText: "\nclass Foo {\n  accessor a: Foo<string> = new Foo();\n}\n      ",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "accessor a: Foo<string> = new Foo();",
			wantFixed:  "\nclass Foo {\n  accessor a = new Foo<string>();\n}\n      ",
		},
		{
			sourceText: "\nclass Foo {\n  accessor a = new Foo<string>();\n}\n      ",
			wire:       "\"type-annotation\"",
			wantId:     "preferTypeAnnotation",
			wantSpan:   "accessor a = new Foo<string>();",
			wantFixed:  "\nclass Foo {\n  accessor a: Foo<string> = new Foo();\n}\n      ",
		},
		{
			sourceText: "\nclass Foo {\n  accessor [a]: Foo<string> = new Foo();\n}\n      ",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "accessor [a]: Foo<string> = new Foo();",
			wantFixed:  "\nclass Foo {\n  accessor [a] = new Foo<string>();\n}\n      ",
		},
		{
			sourceText: "\nfunction foo(a: Foo<string> = new Foo()) {}\n      ",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "a: Foo<string> = new Foo()",
			wantFixed:  "\nfunction foo(a = new Foo<string>()) {}\n      ",
		},
		{
			sourceText: "\nfunction foo({ a }: Foo<string> = new Foo()) {}\n      ",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "{ a }: Foo<string> = new Foo()",
			wantFixed:  "\nfunction foo({ a } = new Foo<string>()) {}\n      ",
		},
		{
			sourceText: "\nfunction foo([a]: Foo<string> = new Foo()) {}\n      ",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "[a]: Foo<string> = new Foo()",
			wantFixed:  "\nfunction foo([a] = new Foo<string>()) {}\n      ",
		},
		{
			sourceText: "\nclass A {\n  constructor(a: Foo<string> = new Foo()) {}\n}\n      ",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "a: Foo<string> = new Foo()",
			wantFixed:  "\nclass A {\n  constructor(a = new Foo<string>()) {}\n}\n      ",
		},
		{
			sourceText: "\nconst a = function (a: Foo<string> = new Foo()) {};\n      ",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "a: Foo<string> = new Foo()",
			wantFixed:  "\nconst a = function (a = new Foo<string>()) {};\n      ",
		},
		{
			sourceText: "const a = new Foo<string>();",
			wire:       "\"type-annotation\"",
			wantId:     "preferTypeAnnotation",
			wantSpan:   "a = new Foo<string>()",
			wantFixed:  "const a: Foo<string> = new Foo();",
		},
		{
			sourceText: "const a = new Map<string, number>();",
			wire:       "\"type-annotation\"",
			wantId:     "preferTypeAnnotation",
			wantSpan:   "a = new Map<string, number>()",
			wantFixed:  "const a: Map<string, number> = new Map();",
		},
		{
			sourceText: "const a = new Map <string, number> ();",
			wire:       "\"type-annotation\"",
			wantId:     "preferTypeAnnotation",
			wantSpan:   "a = new Map <string, number> ()",
			wantFixed:  "const a: Map<string, number> = new Map  ();",
		},
		{
			sourceText: "const a = new Map< string, number >();",
			wire:       "\"type-annotation\"",
			wantId:     "preferTypeAnnotation",
			wantSpan:   "a = new Map< string, number >()",
			wantFixed:  "const a: Map< string, number > = new Map();",
		},
		{
			sourceText: "const a = new \n Foo<string> \n ();",
			wire:       "\"type-annotation\"",
			wantId:     "preferTypeAnnotation",
			wantSpan:   "a = new \n Foo<string> \n ()",
			wantFixed:  "const a: Foo<string> = new \n Foo \n ();",
		},
		{
			sourceText: "const a = new Foo/* comment */ <string> /* another */();",
			wire:       "\"type-annotation\"",
			wantId:     "preferTypeAnnotation",
			wantSpan:   "a = new Foo/* comment */ <string> /* another */()",
			wantFixed:  "const a: Foo<string> = new Foo/* comment */  /* another */();",
		},
		{
			sourceText: "const a = new Foo</* comment */ string, /* another */ number>();",
			wire:       "\"type-annotation\"",
			wantId:     "preferTypeAnnotation",
			wantSpan:   "a = new Foo</* comment */ string, /* another */ number>()",
			wantFixed:  "const a: Foo</* comment */ string, /* another */ number> = new Foo();",
		},
		{
			sourceText: "\nclass Foo {\n  a = new Foo<string>();\n}\n      ",
			wire:       "\"type-annotation\"",
			wantId:     "preferTypeAnnotation",
			wantSpan:   "a = new Foo<string>();",
			wantFixed:  "\nclass Foo {\n  a: Foo<string> = new Foo();\n}\n      ",
		},
		{
			sourceText: "\nclass Foo {\n  [a] = new Foo<string>();\n}\n      ",
			wire:       "\"type-annotation\"",
			wantId:     "preferTypeAnnotation",
			wantSpan:   "[a] = new Foo<string>();",
			wantFixed:  "\nclass Foo {\n  [a]: Foo<string> = new Foo();\n}\n      ",
		},
		{
			sourceText: "\nclass Foo {\n  [a + b] = new Foo<string>();\n}\n      ",
			wire:       "\"type-annotation\"",
			wantId:     "preferTypeAnnotation",
			wantSpan:   "[a + b] = new Foo<string>();",
			wantFixed:  "\nclass Foo {\n  [a + b]: Foo<string> = new Foo();\n}\n      ",
		},
		{
			sourceText: "\nfunction foo(a = new Foo<string>()) {}\n      ",
			wire:       "\"type-annotation\"",
			wantId:     "preferTypeAnnotation",
			wantSpan:   "a = new Foo<string>()",
			wantFixed:  "\nfunction foo(a: Foo<string> = new Foo()) {}\n      ",
		},
		{
			sourceText: "\nfunction foo({ a } = new Foo<string>()) {}\n      ",
			wire:       "\"type-annotation\"",
			wantId:     "preferTypeAnnotation",
			wantSpan:   "{ a } = new Foo<string>()",
			wantFixed:  "\nfunction foo({ a }: Foo<string> = new Foo()) {}\n      ",
		},
		{
			sourceText: "\nfunction foo([a] = new Foo<string>()) {}\n      ",
			wire:       "\"type-annotation\"",
			wantId:     "preferTypeAnnotation",
			wantSpan:   "[a] = new Foo<string>()",
			wantFixed:  "\nfunction foo([a]: Foo<string> = new Foo()) {}\n      ",
		},
		{
			sourceText: "\nclass A {\n  constructor(a = new Foo<string>()) {}\n}\n      ",
			wire:       "\"type-annotation\"",
			wantId:     "preferTypeAnnotation",
			wantSpan:   "a = new Foo<string>()",
			wantFixed:  "\nclass A {\n  constructor(a: Foo<string> = new Foo()) {}\n}\n      ",
		},
		{
			sourceText: "\nconst a = function (a = new Foo<string>()) {};\n      ",
			wire:       "\"type-annotation\"",
			wantId:     "preferTypeAnnotation",
			wantSpan:   "a = new Foo<string>()",
			wantFixed:  "\nconst a = function (a: Foo<string> = new Foo()) {};\n      ",
		},
		{
			sourceText: "\nclass Float32Array<T> {}\nconst a: Float32Array<ArrayBufferLike> = new Float32Array();\nexport {};\n      ",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "a: Float32Array<ArrayBufferLike> = new Float32Array()",
			wantFixed:  "\nclass Float32Array<T> {}\nconst a = new Float32Array<ArrayBufferLike>();\nexport {};\n      ",
		},

		// A class METHOD's default parameter, which reports where the arrow above does not. In
		// estree a method is a MethodDefinition whose value is a FunctionExpression, so upstream's
		// selector reaches it through its second branch. This case and the arrow ones above are the
		// pair that holds the parameter gate in place from both sides. Upstream writes neither.
		{
			sourceText: "class A {\n  foo(a: Foo<string> = new Foo()) {}\n}\n",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "a: Foo<string> = new Foo()",
			wantFixed:  "class A {\n  foo(a = new Foo<string>()) {}\n}\n",
		},

		// A PARENTHESIZED annotation, which reports upstream and was silent here until the type
		// position unwrap existed. estree has no parenthesized-type node so upstream's rule never
		// sees one; our parser produces a real KindParenthesizedType and the type-reference gate
		// declined it. The corpus writes no parenthesized annotation anywhere, which is why this
		// had to be found by probing shapes rather than by running imported cases.
		//
		// Note what the repair writes: the parentheses go with the annotation, because the whole
		// annotation including them is what gets deleted.
		{
			sourceText: "const a: (Foo<string>) = new Foo();\n",
			wire:       "\"constructor\"",
			wantId:     "preferConstructor",
			wantSpan:   "a: (Foo<string>) = new Foo()",
			wantFixed:  "const a = new Foo<string>();\n",
		},
	}
	for index, testCase := range cases {
		t.Run(consistentGenericConstructorsCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, ConsistentGenericConstructors,
				consistentGenericConstructorsFile, testCase.sourceText,
				consistentGenericConstructorsSettingsFor(t, testCase.wire))

			rule_testing.ExpectFindings(t, result, testCase.wantId)

			// RunTyped writes the fixture trimmed, so both the span slice and the expected rewrite
			// are taken against that text rather than against the Go literal above.
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			diagnostic := result.Diagnostics[0]

			gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Fatalf("span: expected %q, got %q", testCase.wantSpan, gotSpan)
			}

			rule_testing.ExpectFixedSource(t, result, strings.TrimSpace(testCase.wantFixed)+"\n")
		})
	}
}

// TestConsistentGenericConstructorsCannotExpressIsolatedDeclarations records a case from upstream's
// corpus that this harness cannot reproduce, with the controls that establish why.
//
// Upstream's forty sixth passing case is `const foo: Foo<string> = new Foo();` under
// `parserOptions.isolatedDeclarations`. It is clean ONLY because of that setting: without it the
// same source reports, which the first assertion below shows. Under isolated declarations the
// annotation has to stay explicit, so the repair that deletes it would break the file, and upstream
// declines rather than proposing it.
//
// The rule reads the real compiler option off the program rather than a lint-config field, which is
// the right instrument here and is why no fixture can set it: rule_testing writes its tsconfig AFTER
// the setup hook runs and always with the default settings, so a hook writing its own tsconfig is
// overwritten. Measured, with a hook writing `"isolatedDeclarations": true` and the rule still
// observing the option unset.
//
// So the case is recorded as a fact about the harness rather than hidden by relaxing the rule to
// make it green. The branch it guards is the one part of this port no fixture covers, and saying so
// here is more useful than a mutation score that quietly counts it as unreachable.
func TestConsistentGenericConstructorsCannotExpressIsolatedDeclarations(t *testing.T) {
	t.Parallel()

	sourceText := "const foo: Foo<string> = new Foo();\n"

	// Control one: without the option, this source reports. So the case upstream lists as clean is
	// clean because of the setting rather than because of anything in the source.
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t,
		ConsistentGenericConstructors, consistentGenericConstructorsFile, sourceText,
		consistentGenericConstructorsSettingsFor(t, "\"constructor\"")), "preferConstructor")

	// Control two: the same source is clean in the other mode, which shows the finding above is the
	// constructor-mode judgment rather than the rule reporting on anything it sees.
	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t,
		ConsistentGenericConstructors, consistentGenericConstructorsFile, sourceText,
		consistentGenericConstructorsSettingsFor(t, "\"type-annotation\"")))
}

// TestConsistentGenericConstructorsDecoderReadsEveryWireShape pins the option surface, including the
// two spellings that have no upstream counterpart and the empty input that a bare severity delivers.
//
// The empty case is the one worth having. A rule configured as `"error"` reaches its decoder with no
// bytes, and a decoder returning a zero-value struct there gives Mode the empty string, which
// matches neither arm and makes the rule silently inert while every fixture built from an explicit
// struct still passes.
func TestConsistentGenericConstructorsDecoderReadsEveryWireShape(t *testing.T) {
	t.Parallel()

	cases := []struct {
		wire string
		want ConsistentGenericConstructorsMode
	}{
		{wire: "", want: ConsistentGenericConstructorsConstructor},
		{wire: `"constructor"`, want: ConsistentGenericConstructorsConstructor},
		{wire: `"type-annotation"`, want: ConsistentGenericConstructorsTypeAnnotation},

		// The array spelling an ESLint config is written in, accepted so a copied setting is read
		// rather than refused. No upstream counterpart, because upstream never sees the unwrapped
		// form cohere delivers.
		{wire: `["type-annotation"]`, want: ConsistentGenericConstructorsTypeAnnotation},
		{wire: `["constructor"]`, want: ConsistentGenericConstructorsConstructor},

		// `null` is a bare severity too, and gets the default.
		{wire: `null`, want: ConsistentGenericConstructorsConstructor},
	}

	for _, testCase := range cases {
		t.Run(testCase.wire, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeConsistentGenericConstructorsOptions([]byte(testCase.wire))
			if err != nil {
				t.Fatalf("decoding %q: %v", testCase.wire, err)
			}
			settings, isSettings := decoded.(ConsistentGenericConstructorsOptions)
			if !isSettings {
				t.Fatalf("decoding %q produced %T rather than the options struct", testCase.wire, decoded)
			}
			if settings.Mode != testCase.want {
				t.Fatalf("decoding %q: expected mode %q, got %q", testCase.wire, testCase.want,
					settings.Mode)
			}
		})
	}

	// And the nil-options path the rule itself takes when the config layer hands it nothing, which
	// is a different route than the decoder and has its own fallback.
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t,
		ConsistentGenericConstructors, consistentGenericConstructorsFile,
		"const a: Map<string, number> = new Map();\n", nil), "preferConstructor")
}

// TestConsistentGenericConstructorsDecoderRefusesAValueUpstreamRefuses pins the other half (#p9s1131).
//
// Each of these used to decode to the default constructor mode, so a config naming one loaded clean
// and ran a mode nobody wrote. Upstream's schema is a string enum and refuses every one at load. The
// error must name the value, because the config layer prefixes the rule's name and the value is the
// half the author needs to find the line.
func TestConsistentGenericConstructorsDecoderRefusesAValueUpstreamRefuses(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		wire, named string
	}{
		{wire: `"type-anotation"`, named: "type-anotation"},
		{wire: `"Constructor"`, named: "Constructor"},
		{wire: `{"anything": 1}`, named: "anything"},
		{wire: `1`, named: "1"},
		{wire: `true`, named: "true"},
		{wire: `["type-anotation"]`, named: "type-anotation"},
		{wire: `["constructor", "type-annotation"]`, named: "type-annotation"},
		{wire: `[]`, named: "[]"},
	} {
		t.Run(testCase.wire, func(t *testing.T) {
			t.Parallel()
			_, err := DecodeConsistentGenericConstructorsOptions([]byte(testCase.wire))
			if err == nil {
				t.Fatalf("%s decoded, so the configured mode is silently not the one written", testCase.wire)
			}
			if !strings.Contains(err.Error(), testCase.named) {
				t.Fatalf("%s was refused without naming %q: %v", testCase.wire, testCase.named, err)
			}
		})
	}
}
