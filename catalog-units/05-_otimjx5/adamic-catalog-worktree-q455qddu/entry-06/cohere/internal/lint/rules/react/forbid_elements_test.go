package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// forbidElementsFile is where the fixtures pretend to live.
//
// A .tsx extension because most cases hold JSX. The rule has no suffix gate, which a case below
// pins by writing reporting source to a plain `.ts` file.
const forbidElementsFile = "/repository/source/ForbidElements.tsx"

// The corpus is upstream's, extracted mechanically rather than retyped.
//
// Every case in the two tables below comes from
// /tmp/lint-sources/eslint-plugin-react/tests/lib/rules/forbid-elements.js, read by evaluating the
// two arrays in the tester with a stubbed RuleTester and emitting Go raw strings from the decoded
// values, so no escape sequence was typed on the way here. The option bodies are the decoded
// objects re-serialised, which is what the config layer would hand the decoder. Upstream carries 14
// valid and 16 invalid cases, and the invalid ones name 22 findings between them.
//
// All 30 were replayed against the installed build, eslint-plugin-react 7.37.5, through the ESLint
// Linter API before any Go was written, and the corpus and the running rule agreed on every one.
//
// The tester carries a bare `require('babel-eslint')` for its side effect, which the extractor
// stubs. Nothing in the corpus depends on that parser.

// forbidElementsOptions decodes a raw option body the way the config layer does.
//
// Fixtures route through the exported decoder rather than building the struct, which is what puts
// the string-or-object union under test. An empty body is the nil-options case: upstream reads
// `configuration.forbid || []`, so the answer is an empty list and the rule declines everything.
func forbidElementsOptions(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeForbidElementsOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding %q: %v", raw, err)
	}
	return decoded
}

// TestForbidElementsFires runs the sixteen failing cases from upstream.
func TestForbidElementsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
		wantIds    []string
	}{
		{"invalid 0 a forbidden intrinsic", `<button />`, `{"forbid":["button"]}`, []string{"forbiddenElement"}},
		{"invalid 1 two forbidden elements in one array", `[<Modal />, <button />]`, `{"forbid":["button","Modal"]}`, []string{"forbiddenElement", "forbiddenElement"}},
		{"invalid 2 a dotted lowercase name", `<dotted.component />`, `{"forbid":["dotted.component"]}`, []string{"forbiddenElement"}},
		{"invalid 3 a dotted capitalised name with a note", `<dotted.Component />`, `{"forbid":[{"element":"dotted.Component","message":"that ain't cool"}]}`, []string{"forbiddenElement_message"}},
		{"invalid 4 an intrinsic with a note", `<button />`, `{"forbid":[{"element":"button","message":"use <Button> instead"}]}`, []string{"forbiddenElement_message"}},
		{"invalid 5 nested elements, both objects", `<button><input /></button>`, `{"forbid":[{"element":"button"},{"element":"input"}]}`, []string{"forbiddenElement", "forbiddenElement"}},
		{"invalid 6 nested elements, object then string", `<button><input /></button>`, `{"forbid":[{"element":"button"},"input"]}`, []string{"forbiddenElement", "forbiddenElement"}},
		{"invalid 7 nested elements, string then object", `<button><input /></button>`, `{"forbid":["input",{"element":"button"}]}`, []string{"forbiddenElement", "forbiddenElement"}},
		// The corpus's one statement that later entries win. The finding carries the SECOND note,
		// which the message-text assertion below pins, since both entries share a message id.
		{"invalid 8 the same name twice keeps the last note", `<button />`, `{"forbid":[{"element":"button","message":"use <Button> instead"},{"element":"button","message":"use <Button2> instead"}]}`, []string{"forbiddenElement_message"}},
		{"invalid 9 a createElement literal with children", `React.createElement("button", {}, child)`, `{"forbid":["button"]}`, []string{"forbiddenElement"}},
		{"invalid 10 an identifier and a literal in one array", `[React.createElement(Modal), React.createElement("button")]`, `{"forbid":["button","Modal"]}`, []string{"forbiddenElement", "forbiddenElement"}},
		{"invalid 11 a createElement member with a note", `React.createElement(dotted.Component)`, `{"forbid":[{"element":"dotted.Component","message":"that ain't cool"}]}`, []string{"forbiddenElement_message"}},
		{"invalid 12 a createElement member, lowercase", `React.createElement(dotted.component)`, `{"forbid":["dotted.component"]}`, []string{"forbiddenElement"}},
		{"invalid 13 a createElement identifier starting with an underscore", `React.createElement(_comp)`, `{"forbid":["_comp"]}`, []string{"forbiddenElement"}},
		{"invalid 14 a createElement literal with a note", `React.createElement("button")`, `{"forbid":[{"element":"button","message":"use <Button> instead"}]}`, []string{"forbiddenElement_message"}},
		{"invalid 15 nested createElement calls", `React.createElement("button", {}, React.createElement("input"))`, `{"forbid":[{"element":"button"},{"element":"input"}]}`, []string{"forbiddenElement", "forbiddenElement"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(
				t,
				ForbidElements,
				forbidElementsFile,
				testCase.sourceText,
				forbidElementsOptions(t, testCase.rawOptions),
			)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestForbidElementsStaysSilent runs the fourteen passing cases from upstream.
func TestForbidElementsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
	}{
		{"valid 0 no forbid list at all", `<button />`, `{}`},
		{"valid 1 an empty forbid list", `<button />`, `{"forbid":[]}`},
		{"valid 2 a capitalised component is a different name", `<Button />`, `{"forbid":["button"]}`},
		{"valid 3 the same, with the object spelling", `<Button />`, `{"forbid":[{"element":"button"}]}`},
		// A lowercase identifier fails upstream's `/^[A-Z_]/`, so nothing is looked up at all.
		{"valid 4 a lowercase createElement identifier", `React.createElement(button)`, `{"forbid":["button"]}`},
		// A bare call with no react import does not resolve to createElement.
		{"valid 5 a bare createElement call", `createElement("button")`, `{"forbid":["button"]}`},
		{"valid 6 a namespace that is not React", `NotReact.createElement("button")`, `{"forbid":["button"]}`},
		// The literal regex is `/^[a-z][^.]*$/`, so an underscore, a capital and a dot each fail it.
		{"valid 7 a literal starting with an underscore", `React.createElement("_thing")`, `{"forbid":["_thing"]}`},
		{"valid 8 a capitalised literal", `React.createElement("Modal")`, `{"forbid":["Modal"]}`},
		{"valid 9 a literal holding a dot", `React.createElement("dotted.component")`, `{"forbid":["dotted.component"]}`},
		{"valid 10 a function expression argument", `React.createElement(function() {})`, `{"forbid":["button"]}`},
		{"valid 11 an object argument", `React.createElement({})`, `{"forbid":["button"]}`},
		{"valid 12 a numeric literal argument", `React.createElement(1)`, `{"forbid":["button"]}`},
		{"valid 13 no arguments at all", `React.createElement()`, ``},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(
				t,
				ForbidElements,
				forbidElementsFile,
				testCase.sourceText,
				forbidElementsOptions(t, testCase.rawOptions),
			)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestForbidElementsHasNoFileSuffixGate pins that the rule reads a plain `.ts` file.
//
// Three siblings in this package gate on `.tsx`/`.jsx`, inherited from oxc. This rule is ported
// from the authority, which has no such gate, and the createElement arm needs no JSX at all.
//
// `.jsx` and `.js` are not covered because the typed program's tsconfig includes only TypeScript
// extensions, which is a fact about the harness rather than about the rule. `.ts` is the extension
// that separates gated from ungated behaviour and it is the one the tree is mostly made of.
func TestForbidElementsHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	for _, suffix := range []string{".tsx", ".ts"} {
		t.Run(suffix, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(
				t,
				ForbidElements,
				"/repository/source/SuffixProbe"+suffix,
				`React.createElement("button");`,
				forbidElementsOptions(t, `{"forbid":["button"]}`),
			)
			rule_testing.ExpectFindings(t, result, "forbiddenElement")
		})
	}
}

// TestForbidElementsAnchorsOnTheName asserts where each finding points.
//
// The corpus asserts message ids only, so nothing in it can see a rule reporting the right judgment
// in the wrong place. Upstream reports on the NAME node for JSX and on the ARGUMENT node for a
// createElement call, so neither finding covers the whole element or the whole call. Both were read
// from upstream's reported columns rather than from its source.
func TestForbidElementsAnchorsOnTheName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
		wantText   string
	}{
		{"a jsx name, not the tag", `<button />`, `{"forbid":["button"]}`, `button`},
		{"a dotted jsx name in full", `<dotted.component />`, `{"forbid":["dotted.component"]}`, `dotted.component`},
		{"a createElement literal with its quotes", `React.createElement("button")`, `{"forbid":["button"]}`, `"button"`},
		{"a createElement identifier", `React.createElement(Modal)`, `{"forbid":["Modal"]}`, `Modal`},
		{"a createElement member in full", `React.createElement(dotted.component)`, `{"forbid":["dotted.component"]}`, `dotted.component`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// RunTyped writes strings.TrimSpace(source)+"\n" to disk, so the expectation is
			// transformed the same way rather than sliced from the untransformed literal.
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			result := rule_testing.RunTypedWithOptions(
				t, ForbidElements, forbidElementsFile, testCase.sourceText,
				forbidElementsOptions(t, testCase.rawOptions),
			)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
			}
			got := onDisk[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if got != testCase.wantText {
				t.Errorf("finding spans %q, want %q", got, testCase.wantText)
			}
		})
	}
}

// TestForbidElementsLaterEntriesWin pins the lookup's assignment order.
//
// Upstream builds a map keyed by element name and assigns in list order, so a name written twice
// keeps the LAST entry. The corpus states this once, for two objects sharing a message id, which
// means an id assertion cannot see which of the two won. The rendered text can, and the mixed
// spellings were measured against the installed build because the corpus writes neither.
func TestForbidElementsLaterEntriesWin(t *testing.T) {
	t.Parallel()

	t.Run("the last of two notes is the one reported", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTypedWithOptions(
			t, ForbidElements, forbidElementsFile, `<button />`,
			forbidElementsOptions(t, `{"forbid":[{"element":"button","message":"use <Button> instead"},{"element":"button","message":"use <Button2> instead"}]}`),
		)
		rule_testing.ExpectFindings(t, result, "forbiddenElement_message")
		got := result.Diagnostics[0].Message.Description
		if !strings.HasSuffix(got, ": use <Button2> instead") {
			t.Errorf("reported the wrong note: %q", got)
		}
	})

	// A string entry after an object entry ERASES the note, which is the direction a port storing
	// the first match would get backwards.
	t.Run("a bare string after an object drops the note", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTypedWithOptions(
			t, ForbidElements, forbidElementsFile, `<button />`,
			forbidElementsOptions(t, `{"forbid":[{"element":"button","message":"m"},"button"]}`),
		)
		rule_testing.ExpectFindings(t, result, "forbiddenElement")
	})

	t.Run("an object after a bare string adds the note", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTypedWithOptions(
			t, ForbidElements, forbidElementsFile, `<button />`,
			forbidElementsOptions(t, `{"forbid":["button",{"element":"button","message":"m"}]}`),
		)
		rule_testing.ExpectFindings(t, result, "forbiddenElement_message")
	})
}

// TestForbidElementsEmptyMessageIsNotAMessage pins upstream's truthiness test.
//
// `message ? messages.forbiddenElement_message : messages.forbiddenElement` takes the else branch on
// an empty string, so an entry carrying `message: ""` reports the PLAIN id. Measured against the
// installed build. A port testing for the key's presence rather than its truthiness would report a
// message id with an empty tail, and no corpus case writes an empty message.
func TestForbidElementsEmptyMessageIsNotAMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTypedWithOptions(
		t, ForbidElements, forbidElementsFile, `<button />`,
		forbidElementsOptions(t, `{"forbid":[{"element":"button","message":""}]}`),
	)
	rule_testing.ExpectFindings(t, result, "forbiddenElement")
}

// TestForbidElementsArgumentShapes covers the two regexes and the member arm.
//
// This is where nearly all of upstream's judgment sits and the corpus states only part of it. Every
// case here was measured against the installed build before it was written.
func TestForbidElementsArgumentShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
		wantIds    []string
	}{
		// The identifier regex is `/^[A-Z_]/`, tested at both accepted characters.
		{"an uppercase identifier reports", `React.createElement(Modal)`, `{"forbid":["Modal"]}`, []string{"forbiddenElement"}},
		{"an underscore identifier reports", `React.createElement(_x)`, `{"forbid":["_x"]}`, []string{"forbiddenElement"}},
		{"a lowercase identifier declines", `React.createElement(button)`, `{"forbid":["button"]}`, nil},
		// The literal regex is `/^[a-z][^.]*$/`, anchored at both ends.
		{"a dash inside a literal is fine", `React.createElement("my-el")`, `{"forbid":["my-el"]}`, []string{"forbiddenElement"}},
		{"a digit first fails the literal regex", `React.createElement("1x")`, `{"forbid":["1x"]}`, nil},
		{"a template literal is not a Literal upstream", "React.createElement(`button`)", `{"forbid":["button"]}`, nil},
		// The one value in upstream's Literal set whose stringification matches the lowercase
		// regex. Nothing in the corpus writes it, and it is reproduced rather than corrected: the
		// alternative is a port that quietly disagrees with the running rule.
		{"a boolean keyword reaches the lookup as its text", `React.createElement(true)`, `{"forbid":["true"]}`, []string{"forbiddenElement"}},
		{"a false keyword likewise", `React.createElement(false)`, `{"forbid":["false"]}`, []string{"forbiddenElement"}},
		// The member arm reads SOURCE TEXT, so the spelling survives into the lookup key. A port
		// rebuilding the name from the tree would answer "a.b" for all three of these.
		{"a deep member reports its whole path", `React.createElement(a.b.c)`, `{"forbid":["a.b.c"]}`, []string{"forbiddenElement"}},
		{"a computed member keeps its brackets and quotes", `React.createElement(a["b"])`, `{"forbid":["a[\"b\"]"]}`, []string{"forbiddenElement"}},
		{"a spaced member keeps its spaces", `React.createElement(a . b)`, `{"forbid":["a . b"]}`, []string{"forbiddenElement"}},
		{"a spaced member does not match the tight spelling", `React.createElement(a . b)`, `{"forbid":["a.b"]}`, nil},
		// The JSX arm reads source text too, which is why a namespaced name arrives whole.
		{"a namespaced jsx name reports with its colon", `<svg:rect />`, `{"forbid":["svg:rect"]}`, []string{"forbiddenElement"}},
		{"a deep jsx member name reports whole", `<a.b.c />`, `{"forbid":["a.b.c"]}`, []string{"forbiddenElement"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(
				t, ForbidElements, forbidElementsFile, testCase.sourceText,
				forbidElementsOptions(t, testCase.rawOptions),
			)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestForbidElementsCreateElementResolution covers which callees count.
//
// The corpus pins two declines directly, and the accepting shapes below were measured. This is the
// same resolution three sibling rules already share, so the value here is pinning that this rule
// uses it rather than the shelf's wider `react.IsCreateElementCall`, which accepts a bare call with
// no import and would report on upstream's own clean case.
func TestForbidElementsCreateElementResolution(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a bare call with no import declines", `createElement("button");`, nil},
		{"a foreign namespace declines", `NotReact.createElement("button");`, nil},
		{"a computed member declines", `React['createElement']("button");`, nil},
		{"a destructured react import reports", "import {createElement} from 'react';\ncreateElement(\"button\");", []string{"forbiddenElement"}},
		{"a destructured preact import declines", "import {createElement} from 'preact';\ncreateElement(\"button\");", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(
				t, ForbidElements, forbidElementsFile, testCase.sourceText,
				forbidElementsOptions(t, `{"forbid":["button"]}`),
			)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestForbidElementsRequiresTheTypedHarness pins the checker declaration.
//
// The bare-identifier branch of the createElement resolution asks the checker, so under the plain
// harness that arm guards and goes silent while the JSX arm keeps reporting. A later revert of
// `NeedsTypeChecker` fails here rather than producing a vacuous green, and the JSX half is what
// makes this a discrimination rather than a blanket silence.
func TestForbidElementsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	source := "import {createElement} from 'react';\ncreateElement(\"button\");"
	options := forbidElementsOptions(t, `{"forbid":["button"]}`)

	untyped := rule_testing.RunWithOptions(t, ForbidElements, forbidElementsFile, source, options)
	rule_testing.ExpectClean(t, untyped)

	typed := rule_testing.RunTypedWithOptions(t, ForbidElements, forbidElementsFile, source, options)
	rule_testing.ExpectFindings(t, typed, "forbiddenElement")

	jsxUntyped := rule_testing.RunWithOptions(t, ForbidElements, forbidElementsFile, `<button />`, options)
	rule_testing.ExpectFindings(t, jsxUntyped, "forbiddenElement")
}

// TestDecodeForbidElementsOptions covers the union the schema declares and Go has no shape for.
//
// The items are `anyOf: [string, object]`, so the decoder tries a string first and falls back to
// the object. Both arms and the nil path are asserted here, because a fixture handing the rule a
// pre-built struct would leave the whole decoder untested, and the union is the one line most
// likely to have no upstream counterpart.
func TestDecodeForbidElementsOptions(t *testing.T) {
	t.Parallel()

	t.Run("an empty body yields an empty list rather than an error", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeForbidElementsOptions(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		options, ok := decoded.(ForbidElementsOptions)
		if !ok {
			t.Fatalf("decoded to %T", decoded)
		}
		if len(options.Forbid) != 0 {
			t.Errorf("want an empty list, got %d entries", len(options.Forbid))
		}
	})

	t.Run("both arms of the union decode", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeForbidElementsOptions([]byte(`{"forbid":["a",{"element":"b","message":"m"}]}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		options := decoded.(ForbidElementsOptions)
		if len(options.Forbid) != 2 {
			t.Fatalf("want 2 entries, got %d", len(options.Forbid))
		}
		if options.Forbid[0].Element != "a" || options.Forbid[0].Message != "" {
			t.Errorf("string arm decoded to %+v", options.Forbid[0])
		}
		if options.Forbid[1].Element != "b" || options.Forbid[1].Message != "m" {
			t.Errorf("object arm decoded to %+v", options.Forbid[1])
		}
	})

	t.Run("order is preserved, which is what makes the last entry win", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeForbidElementsOptions([]byte(`{"forbid":["x","y","z"]}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		options := decoded.(ForbidElementsOptions)
		got := []string{options.Forbid[0].Element, options.Forbid[1].Element, options.Forbid[2].Element}
		if got[0] != "x" || got[1] != "y" || got[2] != "z" {
			t.Errorf("order is %v", got)
		}
	})

	t.Run("a malformed body errors rather than silently emptying the list", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeForbidElementsOptions([]byte(`{"forbid":`)); err == nil {
			t.Error("want an error for a truncated body")
		}
		if _, err := DecodeForbidElementsOptions([]byte(`{"forbid":[1]}`)); err == nil {
			t.Error("want an error for an entry that is neither a string nor an object")
		}
	})
}
