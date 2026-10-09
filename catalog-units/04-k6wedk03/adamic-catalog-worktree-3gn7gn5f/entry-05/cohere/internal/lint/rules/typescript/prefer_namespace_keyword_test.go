package typescript

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const preferNamespaceKeywordFile = "/repository/source/Namespaces.ts"

// TestPreferNamespaceKeywordFires holds every failing case in the upstream corpus, verbatim,
// with the finding count recovered from the snapshot rather than assumed to be one per input.
//
// The snapshot carries six diagnostics across five inputs. The extra one is the brace-nested
// case, which reports twice: nesting a module declaration inside another module BLOCK is not
// exempt. Only the dotted form is, and the two look alike in source while parsing differently.
func TestPreferNamespaceKeywordFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		count      int
	}{
		{"a bare module declaration", "module foo {}", 1},
		{"a dotted module declaration reports once for the whole chain", "module A.B {}", 1},
		{"an ambient module declaration with an identifier name", "declare module foo {}", 1},
		{"a module nested in a module block reports for each", "\n        declare module foo {\n          module bar {}\n        }\n        ", 2},
		{"a module inside a global augmentation", "\n        declare global {\n            module foo {}\n        }\n        ", 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, PreferNamespaceKeyword, preferNamespaceKeywordFile, testCase.sourceText)
			expected := make([]string, testCase.count)
			for index := range expected {
				expected[index] = "preferNamespaceKeyword"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestPreferNamespaceKeywordStaysSilent holds every passing case in the upstream corpus,
// verbatim, plus cases drawn from reading our own parse shape.
//
// The upstream five are cheap for oxc and expensive for us. oxc gives a string-named module
// (TSExternalModuleDeclaration) and a global augmentation (TSGlobalDeclaration) their own node
// types, so its rule never sees them and needs no guard. typescript-go folds all of it into one
// KindModuleDeclaration, so every exemption oxc gets from its grammar is a test this rule has to
// perform by hand. These five are the only fixtures standing over those hand-written guards.
func TestPreferNamespaceKeywordStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an ambient module declared with a quoted name and no body", "declare module 'foo';"},
		{"an ambient module declared with a quoted name", "declare module 'foo' {}"},
		{"the namespace keyword", "namespace foo {}"},
		{"an ambient namespace", "declare namespace foo {}"},
		{"a global augmentation", "declare global {}"},

		// Ours, from reading the parse rather than the corpus.
		//
		// A dotted namespace parses as a chain of ModuleDeclaration nodes, and every link after
		// the first carries the namespace keyword nowhere in its own text: the child of
		// `namespace A.B {}` spans only `B {}`. A keyword test that scans the child node in
		// isolation finds no `namespace` there and, if it defaulted to reporting, would report
		// this clean input. The corpus has no dotted namespace case, so nothing upstream covers
		// this and it is exactly the shape our AST creates and oxc's does not.
		{"a dotted namespace declaration", "namespace A.B {}"},
		{"a deeply dotted namespace declaration", "namespace A.B.C {}"},

		// A module whose quoted name has no `declare`. Upstream is silent here, and it is
		// silent for the name rather than for the modifier: measured on the release binary,
		// `declare module "foo" {}` is clean while `declare module foo {}` reports. The guard
		// that matters is the name kind, so a fixture that only ever pairs a quoted name with
		// `declare` cannot tell which of the two did the work.
		{"a module with a quoted name and no declare modifier", "module \"foo\" {}"},

		// `global` is a keyword form here, not a name. `module global {}` REPORTS upstream, so
		// the exemption cannot be keyed on the identifier text.
		{"a global augmentation with statements inside", "declare global {\n  interface Window {\n    x: string;\n  }\n}"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, PreferNamespaceKeyword, preferNamespaceKeywordFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestPreferNamespaceKeywordFixesWriteWhatTheyClaim applies the repair and compares the whole
// resulting source, which is the only assertion that can see where the rewrite landed.
//
// A message id proves a finding appeared. It cannot distinguish a correct replacement written
// over the correct range from the same correct replacement written over the wrong one, and the
// second is applied unattended to real files. These seven pairs are upstream's own fix vectors,
// which the extractor reports separately from the corpus for exactly this reason.
//
// Two of them are the whole reason the fix cannot be a naive search for the text `module`:
// `declare /* module */ module foo {}` puts the word inside a comment ahead of the keyword, and
// `declare module X.Y.module { let x: 'module'; }` puts it in a name segment and a string after
// it. A rule that rewrote the first hit would corrupt both while still producing a finding whose
// id and span are correct.
func TestPreferNamespaceKeywordFixesWriteWhatTheyClaim(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSource string
	}{
		{"a bare module declaration", "module foo {}", "namespace foo {}"},
		{"a dotted module declaration rewrites only the leading keyword", "module A.B {}", "namespace A.B {}"},
		{"nested module blocks each get their own keyword rewritten", "\n            module A {\n              module B {}\n            }\n            ", "\n            namespace A {\n              namespace B {}\n            }\n            "},
		{"an ambient module declaration keeps its declare modifier", "declare module foo {}", "declare namespace foo {}"},
		{"an ambient module with a nested module", "\n            declare module foo {\n              module bar {}\n            }\n            ", "\n            declare namespace foo {\n              namespace bar {}\n            }\n            "},
		{"a comment holding the word module ahead of the keyword", "declare /* module */ module foo {}", "declare /* module */ namespace foo {}"},
		{"a name segment and a string both holding the word module", "declare module X.Y.module { let x: 'module'; }", "declare namespace X.Y.module { let x: 'module'; }"},

		// Ours. The export modifier is outside the reported span (measured on the release
		// binary: `export module foo {}` reports at column 8, not column 1), but the fix still
		// has to find the keyword past it.
		{"an exported module declaration", "export module foo {}", "export namespace foo {}"},
		{"an exported ambient module declaration", "export declare module foo {}", "export declare namespace foo {}"},

		// A dotted chain deeper than the corpus goes, to pin that the single rewrite is at the
		// head of the chain rather than once per link.
		{"a three-segment dotted module declaration", "module A.B.C {}", "namespace A.B.C {}"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, PreferNamespaceKeyword, preferNamespaceKeywordFile, testCase.sourceText)
			rule_testing.ExpectFixedSource(t, result, testCase.wantSource)
		})
	}
}

// TestPreferNamespaceKeywordReportsTheRightSpanAndText pins where the finding points and what
// it says.
//
// The span is not the obvious one and the corpus does not contain the input that shows it. oxc
// reports the TSNamespaceDeclaration span, which includes a `declare` modifier and excludes an
// `export` one, because `export` wraps the declaration in a separate node there. typescript-go
// puts both modifiers inside the declaration, so reproducing that means skipping export and only
// export. Measured on the release binary: `declare module foo {}` reports at column 1 and
// `export module foo {}` reports at column 8.
//
// The message strings below are literals typed here rather than references to the rule's own
// constants. Comparing a diagnostic against the constant it was reported with is an equality
// that cannot fail, because a mutation moves both sides together.
func TestPreferNamespaceKeywordReportsTheRightSpanAndText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSpans  []string
	}{
		{"a bare declaration spans the whole statement", "module foo {}", []string{"module foo {}"}},
		{"the declare modifier is inside the span", "declare module foo {}", []string{"declare module foo {}"}},
		{"the export modifier is outside the span", "export module foo {}", []string{"module foo {}"}},
		{"export is skipped and declare is kept", "export declare module foo {}", []string{"declare module foo {}"}},
		{"a dotted declaration spans the whole chain once", "module A.B {}", []string{"module A.B {}"}},
		{"a nested declaration points at the inner module", "declare module foo {\n  module bar {}\n}", []string{"declare module foo {\n  module bar {}\n}", "module bar {}"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, PreferNamespaceKeyword, preferNamespaceKeywordFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.wantSpans))
			}
			for index, want := range testCase.wantSpans {
				diagnostic := result.Diagnostics[index]
				got := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if got != want {
					t.Errorf("finding %d spans %q, want %q", index, got, want)
				}
				if diagnostic.Message.Id != "preferNamespaceKeyword" {
					t.Errorf("finding %d has id %q", index, diagnostic.Message.Id)
				}
				if diagnostic.Message.Description != "This declares a namespace with the `module` "+
					"keyword, which TypeScript has deprecated and plans to make a parse error. It "+
					"means the same thing as `namespace` and reads as an ECMAScript module, which "+
					"it is not. Write `namespace` instead." {
					t.Errorf("finding %d has description %q", index, diagnostic.Message.Description)
				}
			}
		})
	}
}
