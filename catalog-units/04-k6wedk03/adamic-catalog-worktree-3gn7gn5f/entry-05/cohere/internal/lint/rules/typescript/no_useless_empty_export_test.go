package typescript

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus writes no file name, so one is chosen here. It must not end in `.d.ts`, which is
// the rule's own exemption and would silence every case below.
const uselessEmptyExportFile = "/repository/source/Thing.ts"

// TestNoUselessEmptyExportFires carries every failing input of oxc's first tester block verbatim.
//
// The snapshot records eleven diagnostics against these eleven inputs, so each reports exactly
// once and a fixture asserting one finding per input is the right shape.
func TestNoUselessEmptyExportFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an export modifier before the empty export", "\n            export const _ = {};\n            export {};\n        "},
		{"a star re-export before the empty export", "\n            export * from '_';\n            export {};\n        "},
		{"a star re-export after the empty export", "\n            export {};\n            export * from '_';\n        "},
		{"a default export before the empty export", "\n            const _ = {};\n            export default _;\n            export {};\n        "},
		{"a default export after the empty export", "\n            export {};\n            const _ = {};\n            export default _;\n        "},
		{"a named export clause with specifiers", "\n            const _ = {};\n            export { _ };\n            export {};\n        "},
		{"an empty import clause", "\n            import {} from '_';\n            export {};"},
		{"a default import", "\n            import _ from '_';\n            export {};"},
		{"a bare side-effect import", "\n            import '_';\n            export {};"},
		{"a namespace import", "\n            import * as all from '_';\n            export {};"},
		{"an external import-equals", "\n            import _ = require('_')\n            export {};"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessEmptyExport, uselessEmptyExportFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "uselessEmptyExport")
		})
	}
}

// TestNoUselessEmptyExportStaysSilent carries every passing input of oxc's first tester block
// verbatim, then the cases the probes added.
//
// Upstream's clean cases are the load-bearing half of this corpus. `declare module '_'` and
// `export = 3;` both read as exports and neither makes the file a module for this rule's purposes,
// so a predicate written from the phrase "any top-level import or export" breaks precisely here.
func TestNoUselessEmptyExportStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an ambient module declaration alone", "declare module '_'"},
		{"an empty import clause alone", "import {} from '_';"},
		{"a namespace import alone", "import * as _ from '_';"},
		{"an export assignment of an object", "export = {};"},
		{"an export assignment of a number", "export = 3;"},
		{"an export modifier alone", "export const _ = {};"},
		{"a default export alone", "\n            const _ = {};\n            export default _;\n        "},
		{"a star re-export beside an export assignment", "\n            export * from '_';\n            export = {};\n        "},
		{"the empty export alone, which is the whole point of the form", "export {};"},
		{"an internal import-equals", "import x = ns.value; export {};"},
		{"a nested import-equals inside a namespace", "namespace Foo { import Bar = Baz; } export {};"},

		// Added from probing the release binary. Each is a statement that reads as an export or an
		// import and is neither, so each one breaks a looser predicate that the imported corpus
		// alone would let through.
		{"an ambient const declaration", "declare const a = 2;\nexport {};\n"},
		{"an ambient global augmentation", "declare global { interface W {} }\nexport {};\n"},
		{"a bare interface declaration", "interface I {}\nexport {};\n"},
		{"an ambient module with a body", "declare module 'x' { export const y = 1; }\nexport {};\n"},
		{"a non-exported namespace", "namespace N { const q = 1; }\nexport {};\n"},
		{"an export assignment beside the empty export", "export = 3;\nexport {};\n"},
		{"an empty re-export carrying a module specifier", "export const a = 1;\nexport type {} from './x';\n"},
		{"an external import-equals nested inside a namespace", "namespace Foo { import Bar = require('_'); }\nexport {};\n"},
		{"two empty exports and nothing else", "export {};\nexport {};\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoUselessEmptyExport, uselessEmptyExportFile, testCase.sourceText))
		})
	}
}

// TestNoUselessEmptyExportFiresOnFormsTheCorpusOmits covers the disarming statements upstream
// never wrote a failing case for. Each was measured against the release binary before it was
// written here, and each pairs with a silent case above that differs by one token.
func TestNoUselessEmptyExportFiresOnFormsTheCorpusOmits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an exported type alias, against the bare interface above", "export type A = 1;\nexport {};\n"},
		{"an exported declare const, against the ambient const above", "export declare const a = 2;\nexport {};\n"},
		{"an exported namespace, against the bare namespace above", "export namespace N { }\nexport {};\n"},
		{"an exported interface", "export interface I {}\nexport {};\n"},
		{"an exported class", "export class C {}\nexport {};\n"},
		{"an exported enum", "export enum E { A }\nexport {};\n"},
		{"a type-only import", "import type { A } from '_';\nexport {};\n"},
		{"a star export with a namespace binding", "export * as ns from '_';\nexport {};\n"},
		{"a type-only empty export, which is still empty", "import '_';\nexport type {};\n"},
		{"an import beside an export assignment, which the corpus case cannot show", "import '_';\nexport = 3;\nexport {};\n"},
		{"an internal import-equals beside a real export", "import x = ns.value;\nexport const a = 1;\nexport {};\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoUselessEmptyExport, uselessEmptyExportFile, testCase.sourceText),
				"uselessEmptyExport")
		})
	}
}

// TestNoUselessEmptyExportReportsEachEmptyExportSeparately pins that the gather-then-report shape
// reports per statement rather than once per file. Measured: two diagnostics upstream.
func TestNoUselessEmptyExportReportsEachEmptyExportSeparately(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoUselessEmptyExport, uselessEmptyExportFile,
		"export const _ = {};\nexport {};\nexport {};\n")
	rule_testing.ExpectFindings(t, result, "uselessEmptyExport", "uselessEmptyExport")
}

// TestNoUselessEmptyExportFixesWriteWhatTheyClaim carries all eleven of oxc's fix vectors.
//
// This assertion is the reason the rule can ship a fix at all. An id assertion is satisfied by a
// correct detection carrying a deletion over the wrong range, and this fix deletes a whole
// statement, so a range off by one node removes real code while every count above stays green.
// The vectors are upstream's own, which is what makes them worth more than invented pairs: they
// pin that the deletion takes the statement and not the whitespace around it, in both orders.
func TestNoUselessEmptyExportFixesWriteWhatTheyClaim(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSource string
	}{
		{"an export modifier before it", "export const _ = {};export {};", "export const _ = {};"},
		{"a star re-export before it", "export * from '_';export {};", "export * from '_';"},
		{"a star re-export after it", "export {};export * from '_';", "export * from '_';"},
		{"a default export before it", "const _ = {};export default _;export {};", "const _ = {};export default _;"},
		{"a default export after it", "export {};const _ = {};export default _;", "const _ = {};export default _;"},
		{"a named export clause before it", "const _ = {};export { _ };export {};", "const _ = {};export { _ };"},
		{"an empty import clause across lines", "import {} from '_';\n\nexport {};", "import {} from '_';\n\n"},
		{"a default import", "import _ from '_';export {};", "import _ from '_';"},
		{"a bare side-effect import", "import '_';export {};", "import '_';"},
		{"a namespace import", "import * as all from '_';export {};", "import * as all from '_';"},
		{"an external import-equals across lines", "import _ = require('_')\n\nexport {};", "import _ = require('_')\n\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t,
				rule_testing.Run(t, NoUselessEmptyExport, uselessEmptyExportFile, testCase.sourceText),
				testCase.wantSource)
		})
	}
}

// TestNoUselessEmptyExportPointsAtTheEmptyExport slices the source with the finding's own range.
//
// The message id cannot see where a finding points, and the two statements in each input below are
// both exports, so a rule anchoring on the wrong one produces an identical id and count. The
// second case puts the empty export first so that pointing at "the first export" fails.
func TestNoUselessEmptyExportPointsAtTheEmptyExport(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantText   string
	}{
		{"the empty export second", "export const _ = {};\nexport {};\n", "export {};"},
		{"the empty export first", "export {};\nexport * from '_';\n", "export {};"},
		{"a type-only empty export", "import '_';\nexport type {};\n", "export type {};"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessEmptyExport, uselessEmptyExportFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.wantText {
				t.Errorf("finding points at %q, wanted %q", reported, testCase.wantText)
			}
		})
	}
}

// TestNoUselessEmptyExportReportsTheMessageItClaims asserts against literals typed here rather
// than against the rule's own message constant. Comparing to the constant is equality that looks
// correct and cannot fail, because a mutation moves both sides together.
func TestNoUselessEmptyExportReportsTheMessageItClaims(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoUselessEmptyExport, uselessEmptyExportFile,
		"export const _ = {};\nexport {};\n")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "uselessEmptyExport" {
		t.Errorf("message id is %q", got)
	}
	const wantDescription = "An `export {}` exists to make a script into a module, and this file " +
		"is already a module because something else in it imports or exports. The statement " +
		"therefore changes nothing, and it reads as though it were load-bearing to anyone deciding " +
		"whether they may delete it. Remove it."
	if got := result.Diagnostics[0].Message.Description; got != wantDescription {
		t.Errorf("message description is %q", got)
	}
}

// TestNoUselessEmptyExportExemptsDeclarationFiles carries oxc's second tester block.
//
// Those four inputs are pass cases only because the block runs them as `.d.ts`. Each is asserted
// twice here, silent under the declaration extension and reporting under `.ts`, because a gate
// tested only on its silent side passes just as well when the rule is broken outright.
func TestNoUselessEmptyExportExemptsDeclarationFiles(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an exported type alias", "\n            export type A = 1;\n            export {};\n        "},
		{"an exported declare const", "\n            export declare const a = 2;\n            export {};\n        "},
		{"a type-only import", "\n            import type { A } from '_';\n            export {};\n        "},
		{"a named import", "\n            import { A } from '_';\n            export {};\n        "},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoUselessEmptyExport,
				"/repository/source/Thing.d.ts", testCase.sourceText))
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoUselessEmptyExport,
				uselessEmptyExportFile, testCase.sourceText), "uselessEmptyExport")
		})
	}
}
