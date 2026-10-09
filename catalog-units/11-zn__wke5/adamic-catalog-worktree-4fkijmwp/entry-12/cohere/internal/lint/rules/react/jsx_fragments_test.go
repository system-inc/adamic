package react

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is upstream's, extracted by stubbing its RuleTester and capturing the case objects the
// test file hands it, so no case was retyped and no escape sequence was typed by hand.
//
// One mechanical transformation was applied and it is the reason this file needs a note. Upstream
// configures `settings.react.pragma` to `Act` and `settings.react.fragment` to `Frag`, and every
// one of its cases writes those names. cohere reads no React settings, so the two names were
// rewritten to `React` and `Fragment` by a whole-word substitution, and then ALL TWENTY TWO cases
// were replayed against the installed build under default settings before anything was written
// down. Nineteen reproduced their upstream verdict exactly, including the fixer output on all
// thirteen `output` assertions. The three that did not are upstream's `fragmentsNotSupported`
// cases, which depend on `settings.react.version` being 16.1; under default settings the installed
// build answers `preferFragment` for the two named forms and silence for the shorthand, and those
// are the verdicts recorded below.
//
// So every row here is a MEASURED verdict from the installed build rather than a copied assertion,
// which is what makes the pragma substitution safe to have made.
const jsxFragmentsFile = "/repository/source/JsxFragments.tsx"

func TestJsxFragmentsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		options          JsxFragmentsOptions
	}{
		{"upstream valid 0", "<><Foo /></>", JsxFragmentsOptions{Mode: JsxFragmentsSyntax}},
		{"upstream valid 1", "<React.Fragment><Foo /></React.Fragment>", JsxFragmentsOptions{Mode: JsxFragmentsElement}},
		{"upstream valid 2", "<React.Fragment />", JsxFragmentsOptions{Mode: JsxFragmentsElement}},
		{"upstream valid 3", "\n        import React, { Fragment as F } from 'react';\n        <F><Foo /></F>;\n      ", JsxFragmentsOptions{Mode: JsxFragmentsElement}},
		{"upstream valid 4", "\n        const F = React.Fragment;\n        <F><Foo /></F>;\n      ", JsxFragmentsOptions{Mode: JsxFragmentsElement}},
		{"upstream valid 5", "\n        const { Fragment } = React;\n        <Fragment><Foo /></Fragment>;\n      ", JsxFragmentsOptions{Mode: JsxFragmentsElement}},
		{"upstream valid 6", "\n        const { Fragment } = require('react');\n        <Fragment><Foo /></Fragment>;\n      ", JsxFragmentsOptions{Mode: JsxFragmentsElement}},
		{"upstream valid 7", "<React.Fragment key=\"key\"><Foo /></React.Fragment>", JsxFragmentsOptions{Mode: JsxFragmentsSyntax}},
		{"upstream valid 8", "<React.Fragment key=\"key\" />", JsxFragmentsOptions{Mode: JsxFragmentsSyntax}},
		{"upstream invalid 0", "<><Foo /></>", JsxFragmentsOptions{Mode: JsxFragmentsSyntax}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, JsxFragments, jsxFragmentsFile, testCase.sourceText, testCase.options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

func TestJsxFragmentsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		options          JsxFragmentsOptions
		messageIds       []string
	}{
		{"upstream invalid 1", "<React.Fragment><Foo /></React.Fragment>", JsxFragmentsOptions{Mode: JsxFragmentsSyntax}, []string{"preferFragment"}},
		{"upstream invalid 2", "<React.Fragment />", JsxFragmentsOptions{Mode: JsxFragmentsSyntax}, []string{"preferFragment"}},
		{"upstream invalid 3", "<><Foo /></>", JsxFragmentsOptions{Mode: JsxFragmentsElement}, []string{"preferPragma"}},
		{"upstream invalid 4", "<><Foo /></>", JsxFragmentsOptions{Mode: JsxFragmentsElement}, []string{"preferPragma"}},
		{"upstream invalid 5", "<React.Fragment><Foo /></React.Fragment>", JsxFragmentsOptions{Mode: JsxFragmentsSyntax}, []string{"preferFragment"}},
		{"upstream invalid 6", "<React.Fragment />", JsxFragmentsOptions{Mode: JsxFragmentsSyntax}, []string{"preferFragment"}},
		{"upstream invalid 7", "\n        import React, { Fragment as F } from 'react';\n        <F />;\n      ", JsxFragmentsOptions{Mode: JsxFragmentsSyntax}, []string{"preferFragment"}},
		{"upstream invalid 8", "\n        import React, { Fragment as F } from 'react';\n        <F><Foo /></F>;\n      ", JsxFragmentsOptions{Mode: JsxFragmentsSyntax}, []string{"preferFragment"}},
		{"upstream invalid 9", "\n        import React, { Fragment } from 'react';\n        <Fragment><Foo /></Fragment>;\n      ", JsxFragmentsOptions{Mode: JsxFragmentsSyntax}, []string{"preferFragment"}},
		{"upstream invalid 10", "\n        const F = React.Fragment;\n        <F><Foo /></F>;\n      ", JsxFragmentsOptions{Mode: JsxFragmentsSyntax}, []string{"preferFragment"}},
		{"upstream invalid 11", "\n        const { Fragment } = React;\n        <Fragment><Foo /></Fragment>;\n      ", JsxFragmentsOptions{Mode: JsxFragmentsSyntax}, []string{"preferFragment"}},
		{"upstream invalid 12", "\n        const { Fragment } = require('react');\n        <Fragment><Foo /></Fragment>;\n      ", JsxFragmentsOptions{Mode: JsxFragmentsSyntax}, []string{"preferFragment"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, JsxFragments, jsxFragmentsFile, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)
		})
	}
}

// TestJsxFragmentsTypeArgumentsAreReportedAndNotRewritten pins the deliberate decline of the fixer.
//
// Upstream ships a repair for both directions and it DELETES TYPE ARGUMENTS. A JSX tag can carry
// them, they live inside the opening tag but outside the attributes list, and upstream's only guard
// is `attrs && attrs.length > 0`, which cannot see them. Measured against the installed build with
// the TypeScript parser on 2026-08-27:
//
//	<React.Fragment<T>><Foo /></React.Fragment>   fixed to <><Foo /></>
//	<React.Fragment<T> />                         fixed to <></>
//
// Upstream's corpus is JavaScript and cannot express either shape, so all thirteen of its `output`
// assertions pass while the repair destroys type information on a TypeScript tree.
//
// This rule reports both and proposes nothing. The assertion that no fix is attached is the point:
// a message-id assertion alone cannot see a repair, which is exactly how the two rules this tree
// has already lost type information to got through review.
func TestJsxFragmentsTypeArgumentsAreReportedAndNotRewritten(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, sourceText string }{
		{"type argument on a paired tag", "<React.Fragment<T>><Foo /></React.Fragment>;"},
		{"type argument on a self closing tag", "<React.Fragment<T> />;"},
		{"no type argument, the ordinary shape", "<React.Fragment><Foo /></React.Fragment>;"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, JsxFragments, jsxFragmentsFile,
				testCase.sourceText, JsxFragmentsOptions{Mode: JsxFragmentsSyntax})
			rule_testing.ExpectFindings(t, result, "preferFragment")
			for _, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) != 0 {
					t.Fatalf("this rule must propose no fix, got %d", len(diagnostic.Fixes))
				}
				if len(diagnostic.Suggestions) != 0 {
					t.Fatalf("this rule must propose no suggestion, got %d",
						len(diagnostic.Suggestions))
				}
			}
		})
	}
}

// TestJsxFragmentsSpans asserts where each finding points.
//
// Upstream reports on the whole JSX element rather than on its opening tag, measured: the installed
// build spans column 1 through 41 for `<React.Fragment><Foo /></React.Fragment>`, which is the
// entire element. A message-id assertion cannot see this, and the two spellings of the named form
// are different node kinds here, so an anchor mistake on one would be invisible from the other.
func TestJsxFragmentsSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText, reported string
		options                    JsxFragmentsOptions
	}{
		{
			"the paired named form spans the whole element",
			"const x = <React.Fragment><Foo /></React.Fragment>;",
			"<React.Fragment><Foo /></React.Fragment>",
			JsxFragmentsOptions{Mode: JsxFragmentsSyntax},
		},
		{
			"the self closing named form spans the whole element",
			"const x = <React.Fragment />;",
			"<React.Fragment />",
			JsxFragmentsOptions{Mode: JsxFragmentsSyntax},
		},
		{
			"the shorthand spans the whole fragment",
			"const x = <><Foo /></>;",
			"<><Foo /></>",
			JsxFragmentsOptions{Mode: JsxFragmentsElement},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, JsxFragments, jsxFragmentsFile,
				testCase.sourceText, testCase.options)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			// The typed harness writes `strings.TrimSpace(contents)+"\n"`, so slicing the Go
			// literal would be off by one against the file on disk for any fixture with a leading
			// newline. These have none, and the trim is applied here anyway so the transformation
			// matches the one the harness made.
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			reported := onDisk[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.reported {
				t.Fatalf("finding points at %q, wanted %q", reported, testCase.reported)
			}
		})
	}
}

// TestJsxFragmentsMessageText asserts the rendered text of both messages exactly.
//
// Neither message interpolates, so there is no format string to guard, but asserting the text
// against a literal typed here rather than against the rule's own constant is what keeps a message
// mutation visible: comparing a finding to the constant it was reported with moves both sides
// together and passes either way.
func TestJsxFragmentsMessageText(t *testing.T) {
	t.Parallel()

	shorthand := rule_testing.RunTypedWithOptions(t, JsxFragments, jsxFragmentsFile,
		"<React.Fragment><Foo /></React.Fragment>;", JsxFragmentsOptions{Mode: JsxFragmentsSyntax})
	if len(shorthand.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(shorthand.Diagnostics))
	}
	wantShorthand := "Prefer fragment shorthand over React.Fragment. The shorthand `<>` says the " +
		"same thing with less to read, and a named fragment that carries no props is only the " +
		"long spelling of it."
	if shorthand.Diagnostics[0].Message.Description != wantShorthand {
		t.Fatalf("shorthand message is %q", shorthand.Diagnostics[0].Message.Description)
	}
	if shorthand.Diagnostics[0].Message.Id != "preferFragment" {
		t.Fatalf("shorthand id is %q", shorthand.Diagnostics[0].Message.Id)
	}

	pragma := rule_testing.RunTypedWithOptions(t, JsxFragments, jsxFragmentsFile,
		"<><Foo /></>;", JsxFragmentsOptions{Mode: JsxFragmentsElement})
	if len(pragma.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(pragma.Diagnostics))
	}
	wantPragma := "Prefer React.Fragment over fragment shorthand. The named form is the one this " +
		"file is configured to use, and it is the only spelling that can carry a `key`."
	if pragma.Diagnostics[0].Message.Description != wantPragma {
		t.Fatalf("pragma message is %q", pragma.Diagnostics[0].Message.Description)
	}
	if pragma.Diagnostics[0].Message.Id != "preferPragma" {
		t.Fatalf("pragma id is %q", pragma.Diagnostics[0].Message.Id)
	}
}

// TestDecodeJsxFragmentsOptions routes configuration through the rule's own decoder.
//
// The option is a bare enum STRING rather than an object, which is the line most likely to have no
// upstream counterpart: our config layer unwraps the severity tuple, so what arrives is `"element"`
// on its own rather than upstream's `["element"]`. Building the options struct directly in a
// fixture would leave that untested.
//
// The nil case matters separately. A rule configured as a bare `"error"` is handed empty input, and
// a decoder that did not fall back would hand the rule a zero-value mode matching neither arm, so
// the rule would register on every file and report nothing.
func TestDecodeJsxFragmentsOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
		want JsxFragmentsMode
	}{
		{"absent configuration falls back to the default", "", JsxFragmentsSyntax},
		{"the explicit default", `"syntax"`, JsxFragmentsSyntax},
		{"the element mode", `"element"`, JsxFragmentsElement},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeJsxFragmentsOptions([]byte(testCase.raw))
			if err != nil {
				t.Fatalf("decode returned %v", err)
			}
			options, isOptions := decoded.(JsxFragmentsOptions)
			if !isOptions {
				t.Fatalf("decode returned %T", decoded)
			}
			if options.Mode != testCase.want {
				t.Fatalf("mode is %q, wanted %q", options.Mode, testCase.want)
			}
		})
	}
}

// TestJsxFragmentsNilOptionsUsesTheDefault bypasses the decoder entirely.
//
// A rule handed nil options reaches `options.(JsxFragmentsOptions)`, which fails, and the fallback
// is what keeps the rule working. Every other fixture in this file reaches the rule through an
// options struct, so nothing else here can see this line.
func TestJsxFragmentsNilOptionsUsesTheDefault(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, JsxFragments, jsxFragmentsFile,
		"<React.Fragment><Foo /></React.Fragment>;")
	rule_testing.ExpectFindings(t, result, "preferFragment")
}

// TestJsxFragmentsRequiresTheTypedHarness asserts the checker is genuinely needed.
//
// `GetSymbolAtLocation` on a nil checker returns nil rather than crashing, so a rule that lost its
// guard would go quietly narrower rather than announcing itself. This pins the direction: the
// namespaced spelling is matched structurally and survives, while the bare-identifier route needs
// resolution and goes silent, so a later revert to the untyped harness fails here rather than
// passing vacuously.
func TestJsxFragmentsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !JsxFragments.NeedsTypeChecker {
		t.Fatal("this rule resolves a bare tag name through the checker and must declare it")
	}
	untyped := rule_testing.RunWithOptions(t, JsxFragments, jsxFragmentsFile,
		"import React, { Fragment } from 'react';\n<Fragment><Foo /></Fragment>;",
		JsxFragmentsOptions{Mode: JsxFragmentsSyntax})
	rule_testing.ExpectClean(t, untyped)

	typed := rule_testing.RunTypedWithOptions(t, JsxFragments, jsxFragmentsFile,
		"import React, { Fragment } from 'react';\n<Fragment><Foo /></Fragment>;",
		JsxFragmentsOptions{Mode: JsxFragmentsSyntax})
	rule_testing.ExpectFindings(t, typed, "preferFragment")
}

// TestJsxFragmentsShapesUpstreamDoesNotWrite covers inputs measured against the installed build
// that upstream's corpus has no case for.
//
// Every verdict below was measured on 2026-08-27 rather than reasoned about. The nested pair is the
// one that matters most: upstream's corpus never nests two fragments, and the installed build
// reports BOTH, so a rule that stopped descending after the outer one would pass every imported
// case.
func TestJsxFragmentsShapesUpstreamDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		options          JsxFragmentsOptions
		messageIds       []string
	}{
		{
			"two nested named fragments both report",
			"<React.Fragment><React.Fragment><Foo /></React.Fragment></React.Fragment>;",
			JsxFragmentsOptions{Mode: JsxFragmentsSyntax},
			[]string{"preferFragment", "preferFragment"},
		},
		{
			"a named fragment nested inside an ordinary element reports",
			"<div><React.Fragment><Foo /></React.Fragment></div>;",
			JsxFragmentsOptions{Mode: JsxFragmentsSyntax},
			[]string{"preferFragment"},
		},
		{
			"an empty named fragment reports",
			"<React.Fragment></React.Fragment>;",
			JsxFragmentsOptions{Mode: JsxFragmentsSyntax},
			[]string{"preferFragment"},
		},
		{
			"a spread attribute counts as a prop and is left alone",
			"<React.Fragment {...props}><Foo /></React.Fragment>;",
			JsxFragmentsOptions{Mode: JsxFragmentsSyntax},
			nil,
		},
		{
			"a three segment member access is not the pragma",
			"<A.React.Fragment><Foo /></A.React.Fragment>;",
			JsxFragmentsOptions{Mode: JsxFragmentsSyntax},
			nil,
		},
		{
			"a fragment imported from another module is not the pragma",
			"import { Fragment } from 'preact';\n<Fragment><Foo /></Fragment>;",
			JsxFragmentsOptions{Mode: JsxFragmentsSyntax},
			nil,
		},
		{
			"an unresolved bare tag name is not the pragma",
			"<Fragment><Foo /></Fragment>;",
			JsxFragmentsOptions{Mode: JsxFragmentsSyntax},
			nil,
		},
		{
			"the shorthand is silent under the default",
			"<><Foo /></>;",
			JsxFragmentsOptions{Mode: JsxFragmentsSyntax},
			nil,
		},
		{
			"the named form is silent under the element mode",
			"<React.Fragment><Foo /></React.Fragment>;",
			JsxFragmentsOptions{Mode: JsxFragmentsElement},
			nil,
		},
		{
			"a nested shorthand pair both report under the element mode",
			"<><><Foo /></></>;",
			JsxFragmentsOptions{Mode: JsxFragmentsElement},
			[]string{"preferPragma", "preferPragma"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, JsxFragments, jsxFragmentsFile,
				testCase.sourceText, testCase.options)
			if len(testCase.messageIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)
		})
	}
}

// TestJsxFragmentsVersionArmIsNotPorted records a divergence rather than asserting a behaviour.
//
// Upstream carries a third message, `fragmentsNotSupported`, which fires when
// `settings.react.version` is below 16.2. cohere reads no React settings, and with no version
// configured upstream's own `version.js` sets its default to `999.999.999`, so every
// `testReactVersion` comparison passes and the arm is unreachable.
//
// Measured on the installed build with no react settings, which is upstream's three version cases
// replayed: `<><Foo /></>` is SILENT and both named forms answer `preferFragment` rather than
// `fragmentsNotSupported`. Those are the verdicts the imported rows above record, and this test
// exists so the substitution is stated where a reader will look rather than inferred from three
// rows whose upstream ids do not match.
//
// The `json` import is used here to spell the decoder's wire shape once, which keeps this test
// honest about what the config layer actually delivers.
func TestJsxFragmentsVersionArmIsNotPorted(t *testing.T) {
	t.Parallel()

	for _, messageId := range []string{"fragmentsNotSupported"} {
		for _, source := range []string{
			"<><Foo /></>;",
			"<React.Fragment><Foo /></React.Fragment>;",
			"<React.Fragment />;",
		} {
			for _, mode := range []JsxFragmentsMode{JsxFragmentsSyntax, JsxFragmentsElement} {
				result := rule_testing.RunTypedWithOptions(t, JsxFragments, jsxFragmentsFile,
					source, JsxFragmentsOptions{Mode: mode})
				for _, diagnostic := range result.Diagnostics {
					if diagnostic.Message.Id == messageId {
						t.Fatalf("this rule must not produce %q; it has no version input",
							messageId)
					}
				}
			}
		}
	}

	// The decoder receives the bare enum, not upstream's array. Spelled through the JSON layer so
	// the assertion is about the wire shape rather than about a Go literal.
	raw, err := json.Marshal("element")
	if err != nil {
		t.Fatalf("marshal returned %v", err)
	}
	decoded, err := DecodeJsxFragmentsOptions(raw)
	if err != nil {
		t.Fatalf("decode returned %v", err)
	}
	if decoded.(JsxFragmentsOptions).Mode != JsxFragmentsElement {
		t.Fatal("the bare enum did not decode")
	}
}

// TestJsxFragmentsSurvivorShapes covers two guards that survived the mutation sweep.
//
// Neither shape appears in upstream's corpus, and each was measured against the installed build
// before the fixture was written rather than after.
//
// The MERGED case is the index-zero trap this tree names for six other rules. Declaration merging
// puts an `interface Fragment` written above the import at declaration index 0, so a rule reading
// `symbol.Declarations[0]` instead of looping goes silent on it. Measured: the installed build
// reports `preferFragment` for both orderings, so the loop is what reproduces upstream and the
// index would be a silent false negative. Both orderings are pinned because only one of them can
// distinguish the two versions.
//
// The MEMBER ACCESS case pins the pragma name itself. `<Foo.Fragment>` is silent upstream, measured,
// because upstream compares the rendered element type against the string `React.Fragment` rather
// than checking only the property half. A guard that dropped the object comparison would report
// every `Something.Fragment` in the tree.
func TestJsxFragmentsSurvivorShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		messageIds       []string
	}{
		{
			"a merged interface written above the import still resolves",
			"interface Fragment { a: number }\nimport React, { Fragment } from 'react';\n<Fragment><Foo /></Fragment>;",
			[]string{"preferFragment"},
		},
		{
			"a merged interface written below the import still resolves",
			"import React, { Fragment } from 'react';\ninterface Fragment { a: number }\n<Fragment><Foo /></Fragment>;",
			[]string{"preferFragment"},
		},
		{
			"a member access on a name other than the pragma is silent",
			"<Foo.Fragment><Bar /></Foo.Fragment>;",
			nil,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, JsxFragments, jsxFragmentsFile,
				testCase.sourceText, JsxFragmentsOptions{Mode: JsxFragmentsSyntax})
			if len(testCase.messageIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)
		})
	}
}

// TestJsxFragmentsNonReactSources covers the three initializer shapes with the wrong module or the
// wrong pragma, which upstream's corpus writes only in their accepting spelling.
//
// The `require` row is the one that caught a survivor: without the module-name comparison, any
// `require(...)` destructuring a `Fragment` would report. Measured against the installed build,
// `require('preact')` is silent, so the comparison is load-bearing rather than defensive.
//
// The other two rows pin the pragma half of the same question for the identifier and member-access
// initializers, both measured silent.
func TestJsxFragmentsNonReactSources(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, sourceText string }{
		{
			"require of another module is not the fragment source",
			"const { Fragment } = require('preact');\n<Fragment><Foo /></Fragment>;",
		},
		{
			"a member access on another namespace is not the fragment source",
			"const F = Preact.Fragment;\n<F><Foo /></F>;",
		},
		{
			"destructuring another namespace is not the fragment source",
			"const { Fragment } = Preact;\n<Fragment><Foo /></Fragment>;",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, JsxFragments, jsxFragmentsFile,
				testCase.sourceText, JsxFragmentsOptions{Mode: JsxFragmentsSyntax})
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The decoder reads empty input and `null` as upstream's default and each of upstream's two modes as
// itself, and refuses anything else, naming the value (#rfbha44). Upstream's schema is
// `[{ enum: ["syntax", "element"] }]` on the installed build and refuses each of these at load, the
// empty string included; they used to decode to the default here.
func TestJsxFragmentsDecoderRefusesWhatUpstreamRefuses(t *testing.T) {
	t.Parallel()

	for _, accepted := range []struct {
		wire string
		want JsxFragmentsMode
	}{
		{"", JsxFragmentsSyntax},
		{"null", JsxFragmentsSyntax},
		{`"syntax"`, JsxFragmentsSyntax},
		{`"element"`, JsxFragmentsElement},
	} {
		decoded, err := DecodeJsxFragmentsOptions([]byte(accepted.wire))
		if err != nil {
			t.Errorf("%s was refused: %v", accepted.wire, err)
			continue
		}
		if got := decoded.(JsxFragmentsOptions).Mode; got != accepted.want {
			t.Errorf("%s decoded to %s, want %s", accepted.wire, got, accepted.want)
		}
	}
	for _, refused := range []struct{ wire, named string }{
		{`"elements"`, "elements"},
		{`"Syntax"`, "Syntax"},
		{`""`, `""`},
		{`{"mode": "element"}`, "mode"},
		{`true`, "true"},
	} {
		_, err := DecodeJsxFragmentsOptions([]byte(refused.wire))
		if err == nil {
			t.Errorf("%s decoded; it must be refused", refused.wire)
			continue
		}
		if !strings.Contains(err.Error(), refused.named) {
			t.Errorf("%s was refused without naming %s: %v", refused.wire, refused.named, err)
		}
	}
}
