package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const consistentTypeExportsFile = "main.ts"

// consistentTypeExportsCaseName numbers a row so a failure names which one, since many rows differ
// only in which name is a type.
func consistentTypeExportsCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// consistentTypeExportsOnDisk is what the harness actually writes for a fixture.
//
// `RunTypedFiles` writes each file as `strings.TrimSpace(contents)+"\n"`, so a case copied from an
// upstream tester carries a leading newline the file on disk does not have. A span sliced from the
// Go literal is therefore off by one, and an expected fix output compared against the untrimmed
// literal fails on a trailing newline while the repair is byte correct.
func consistentTypeExportsOnDisk(sourceText string) string {
	return strings.TrimSpace(sourceText) + "\n"
}

// consistentTypeExportsFixtureFiles is the plugin's own fixture directory, transcribed.
//
// Upstream's star cases import from `./consistent-type-exports/<name>`, and the verdict depends
// entirely on what those files export, so the rule cannot be fixtured without them. They are written
// here as a map rather than copied from the clone, so the fixture is self-contained and a missing
// clone cannot turn a real failure into a silent pass.
func consistentTypeExportsFixtureFiles(subject string) map[string]string {
	files := map[string]string{
		"consistent-type-exports/type-only-exports.ts": "export type TypeFoo = 1;\n" +
			"export interface InterfaceFoo { foo: 'bar' }\n" +
			"class LocalClass {}\n",
		"consistent-type-exports/type-only-reexport.ts": "export type { TypeFoo } from './type-only-exports';\n",
		"consistent-type-exports/value-reexport.ts":     "export { valueBar } from './index';\n",
		"consistent-type-exports/index.ts": "export type Type1 = 1;\n" +
			"export type Type2 = 1;\n" +
			"export const value1 = 2;\n" +
			"export const value2 = 2;\n" +
			"export class Class1 {}\n" +
			"export const valueBar = 3;\n",
		"consistent-type-exports/reexport-1.ts": "export type A = 1;\n",
		"consistent-type-exports/reexport-2-named.ts": "import { A } from './reexport-1';\n" +
			"const A = 1;\nexport { A };\n",
	}
	// Several upstream cases import from the DIRECTORY rather than from a file in it, which
	// resolves to its index. Written under both names so the resolution works whichever the
	// program prefers, and because one upstream case exports a name declared twice there, as a
	// type and as a value, which is the shape its `NAME as Foo` case turns on.
	files["consistent-type-exports/index.ts"] = files["consistent-type-exports/index.ts"] +
		"export type NAME = 'name';\nexport const NAME2 = 'name';\n"
	files["main.ts"] = subject
	return files
}

// decodeConsistentTypeExportsOptions runs a configuration through the real decoder.
func decodeConsistentTypeExportsOptions(t *testing.T, configuration string) any {
	t.Helper()

	if configuration == "" {
		// A rule configured as a bare `"error"` is handed nil, not an empty struct.
		return nil
	}
	options, err := DecodeConsistentTypeExportsOptions([]byte(configuration))
	if err != nil {
		t.Fatalf("could not decode %s: %v", configuration, err)
	}
	return options
}

// TestConsistentTypeExportsStaysSilentOnUpstreamPassCases is the imported clean corpus.
//
// All twenty-four of upstream's passing inputs, extracted from the clone's test file by parsing it
// with the TypeScript compiler rather than by reading it, so no escape sequence passed through a
// shell or a keyboard on the way here. Every one was additionally run through the installed 8.67.0
// build over a real program with the plugin's own fixture files on disk, which reported nothing on
// all twenty-four.
//
// Ten of them import from another module and are silent because of what THAT module exports, which
// is the half of this rule no single-file fixture can reach.
func TestConsistentTypeExportsStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		sourceText    string
	}{
		{
			configuration: "",
			sourceText:    "export { Foo } from 'foo';",
		},
		{
			configuration: "",
			sourceText:    "export type { Type1 } from './consistent-type-exports';",
		},
		{
			configuration: "",
			sourceText:    "export { value1 } from './consistent-type-exports';",
		},
		{
			configuration: "",
			sourceText:    "export { value1 as \"🍎\" } from './consistent-type-exports';",
		},
		{
			configuration: "",
			sourceText:    "export type { value1 } from './consistent-type-exports';",
		},
		{
			configuration: "",
			sourceText:    "\nconst variable = 1;\nclass Class {}\nenum Enum {}\nfunction Func() {}\nnamespace ValueNS {\n  export const x = 1;\n}\n\nexport { variable, Class, Enum, Func, ValueNS };\n    ",
		},
		{
			configuration: "",
			sourceText:    "\ntype Alias = 1;\ninterface IFace {}\nnamespace TypeNS {\n  export type x = 1;\n}\n\nexport type { Alias, IFace, TypeNS };\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nconst foo = 1;\nexport type { foo };\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nnamespace NonTypeNS {\n  export const x = 1;\n}\n\nexport { NonTypeNS };\n    ",
		},
		{
			configuration: "",
			sourceText:    "export * from './unknown-module';",
		},
		{
			configuration: "",
			sourceText:    "export * from './consistent-type-exports';",
		},
		{
			configuration: "",
			sourceText:    "export * as foo from './consistent-type-exports';",
		},
		{
			configuration: "",
			sourceText:    "\nimport * as Foo from './consistent-type-exports';\ntype Foo = 1;\nexport { Foo }\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nimport { Type1 } from './consistent-type-exports';\nconst Type1 = 1;\nexport { Type1 };\n    ",
		},
		{
			configuration: "",
			sourceText:    "export type * from './consistent-type-exports/type-only-exports';",
		},
		{
			configuration: "",
			sourceText:    "export type * from './consistent-type-exports/type-only-reexport';",
		},
		{
			configuration: "",
			sourceText:    "export * from './consistent-type-exports/value-reexport';",
		},
		{
			configuration: "",
			sourceText:    "export type * as foo from './consistent-type-exports/type-only-exports';",
		},
		{
			configuration: "",
			sourceText:    "export type * as foo from './consistent-type-exports/type-only-reexport';",
		},
		{
			configuration: "",
			sourceText:    "export * as foo from './consistent-type-exports/value-reexport';",
		},
		{
			configuration: "",
			sourceText:    "\nexport { A } from './consistent-type-exports/reexport-2-named';\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nimport { A } from './consistent-type-exports/reexport-2-named';\nexport { A };\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nexport { A } from './consistent-type-exports/reexport-2-namespace';\n    ",
		},
		{
			configuration: "",
			sourceText:    "\nimport { A } from './consistent-type-exports/reexport-2-namespace';\nexport { A };\n    ",
		},
	}
	for index, testCase := range cases {
		t.Run(consistentTypeExportsCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedFilesWithOptions(t,
				ConsistentTypeExports, consistentTypeExportsFixtureFiles(testCase.sourceText),
				consistentTypeExportsFile,
				decodeConsistentTypeExportsOptions(t, testCase.configuration)))
		})
	}
}

// TestConsistentTypeExportsFiresOnUpstreamFailCases is the imported reporting corpus.
//
// All twenty-five of upstream's reporting inputs, each carrying its repair. Four of them are
// export-star statements whose verdict is read out of another module.
//
// Three things are asserted per row. The ids say which of the three arms ran, and they differ in
// what they name. The spans say where the finding points, which is the whole statement. And the
// applied source says what the edit engine will write unattended, which for the mixed cases is a
// statement SPLIT into two and is the thing a message-id assertion cannot see at all.
func TestConsistentTypeExportsFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		sourceText    string
		wantIds       []string
		wantSpans     []string
		wantFixed     string
	}{
		{
			configuration: "",
			sourceText:    "export { Type1 } from './consistent-type-exports';",
			wantIds:       []string{"typeOverValue"},
			wantSpans:     []string{"export { Type1 } from './consistent-type-exports';"},
			wantFixed:     "export type { Type1 } from './consistent-type-exports';",
		},
		{
			configuration: "",
			sourceText:    "export { Type1 as \"🍎\" } from './consistent-type-exports';",
			wantIds:       []string{"typeOverValue"},
			wantSpans:     []string{"export { Type1 as \"🍎\" } from './consistent-type-exports';"},
			wantFixed:     "export type { Type1 as \"🍎\" } from './consistent-type-exports';",
		},
		{
			configuration: "",
			sourceText:    "export { Type1, value1 } from './consistent-type-exports';",
			wantIds:       []string{"singleExportIsType"},
			wantSpans:     []string{"export { Type1, value1 } from './consistent-type-exports';"},
			wantFixed:     "export type { Type1 } from './consistent-type-exports';\nexport { value1 } from './consistent-type-exports';",
		},
		{
			configuration: "",
			sourceText:    "\nexport { Type1, value1, value2 } from './consistent-type-exports';\n      ",
			wantIds:       []string{"singleExportIsType"},
			wantSpans:     []string{"export { Type1, value1, value2 } from './consistent-type-exports';"},
			wantFixed:     "\nexport type { Type1 } from './consistent-type-exports';\nexport { value1, value2 } from './consistent-type-exports';\n      ",
		},
		{
			configuration: "",
			sourceText:    "\nexport { Type1, value1, Type2, value2 } from './consistent-type-exports';\n      ",
			wantIds:       []string{"multipleExportsAreTypes"},
			wantSpans:     []string{"export { Type1, value1, Type2, value2 } from './consistent-type-exports';"},
			wantFixed:     "\nexport type { Type1, Type2 } from './consistent-type-exports';\nexport { value1, value2 } from './consistent-type-exports';\n      ",
		},
		{
			configuration: "",
			sourceText:    "export { Type2 as Foo } from './consistent-type-exports';",
			wantIds:       []string{"typeOverValue"},
			wantSpans:     []string{"export { Type2 as Foo } from './consistent-type-exports';"},
			wantFixed:     "export type { Type2 as Foo } from './consistent-type-exports';",
		},
		{
			configuration: "",
			sourceText:    "\nexport { Type2 as Foo, value1 } from './consistent-type-exports';\n      ",
			wantIds:       []string{"singleExportIsType"},
			wantSpans:     []string{"export { Type2 as Foo, value1 } from './consistent-type-exports';"},
			wantFixed:     "\nexport type { Type2 as Foo } from './consistent-type-exports';\nexport { value1 } from './consistent-type-exports';\n      ",
		},
		{
			configuration: "",
			sourceText:    "\nexport {\n  Type2 as Foo,\n  value1 as BScope,\n  value2 as CScope,\n} from './consistent-type-exports';\n      ",
			wantIds:       []string{"singleExportIsType"},
			wantSpans:     []string{"export {\n  Type2 as Foo,\n  value1 as BScope,\n  value2 as CScope,\n} from './consistent-type-exports';"},
			wantFixed:     "\nexport type { Type2 as Foo } from './consistent-type-exports';\nexport { value1 as BScope, value2 as CScope } from './consistent-type-exports';\n      ",
		},
		{
			configuration: "",
			sourceText:    "\nimport { Type2 } from './consistent-type-exports';\nexport { Type2 };\n      ",
			wantIds:       []string{"typeOverValue"},
			wantSpans:     []string{"export { Type2 };"},
			wantFixed:     "\nimport { Type2 } from './consistent-type-exports';\nexport type { Type2 };\n      ",
		},
		{
			configuration: "",
			sourceText:    "\nimport { value2, Type2 } from './consistent-type-exports';\nexport { value2, Type2 };\n      ",
			wantIds:       []string{"singleExportIsType"},
			wantSpans:     []string{"export { value2, Type2 };"},
			wantFixed:     "\nimport { value2, Type2 } from './consistent-type-exports';\nexport type { Type2 };\nexport { value2 };\n      ",
		},
		{
			configuration: "",
			sourceText:    "\ntype Alias = 1;\ninterface IFace {}\nnamespace TypeNS {\n  export type x = 1;\n  export const f = 1;\n}\n\nexport { Alias, IFace, TypeNS };\n      ",
			wantIds:       []string{"multipleExportsAreTypes"},
			wantSpans:     []string{"export { Alias, IFace, TypeNS };"},
			wantFixed:     "\ntype Alias = 1;\ninterface IFace {}\nnamespace TypeNS {\n  export type x = 1;\n  export const f = 1;\n}\n\nexport type { Alias, IFace };\nexport { TypeNS };\n      ",
		},
		{
			configuration: "",
			sourceText:    "\nnamespace TypeNS {\n  export interface Foo {}\n}\n\nexport { TypeNS };\n      ",
			wantIds:       []string{"typeOverValue"},
			wantSpans:     []string{"export { TypeNS };"},
			wantFixed:     "\nnamespace TypeNS {\n  export interface Foo {}\n}\n\nexport type { TypeNS };\n      ",
		},
		{
			configuration: "",
			sourceText:    "\ntype T = 1;\nexport { type T, T };\n      ",
			wantIds:       []string{"typeOverValue"},
			wantSpans:     []string{"export { type T, T };"},
			wantFixed:     "\ntype T = 1;\nexport type { T, T };\n      ",
		},
		{
			configuration: "",
			sourceText:    "\ntype T = 1;\nexport { type/* */T, type     /* */T, T };\n      ",
			wantIds:       []string{"typeOverValue"},
			wantSpans:     []string{"export { type/* */T, type     /* */T, T };"},
			wantFixed:     "\ntype T = 1;\nexport type { /* */T, /* */T, T };\n      ",
		},
		{
			configuration: "",
			sourceText:    "\ntype T = 1;\nconst x = 1;\nexport { type T, T, x };\n      ",
			wantIds:       []string{"singleExportIsType"},
			wantSpans:     []string{"export { type T, T, x };"},
			wantFixed:     "\ntype T = 1;\nconst x = 1;\nexport type { T, T };\nexport { x };\n      ",
		},
		{
			configuration: "{\"fixMixedExportsWithInlineTypeSpecifier\":true}",
			sourceText:    "\ntype T = 1;\nconst x = 1;\nexport { T, x };\n      ",
			wantIds:       []string{"singleExportIsType"},
			wantSpans:     []string{"export { T, x };"},
			wantFixed:     "\ntype T = 1;\nconst x = 1;\nexport { type T, x };\n      ",
		},
		{
			configuration: "{\"fixMixedExportsWithInlineTypeSpecifier\":true}",
			sourceText:    "\ntype T = 1;\nexport { type T, T };\n      ",
			wantIds:       []string{"typeOverValue"},
			wantSpans:     []string{"export { type T, T };"},
			wantFixed:     "\ntype T = 1;\nexport type { T, T };\n      ",
		},
		{
			configuration: "{\"fixMixedExportsWithInlineTypeSpecifier\":false}",
			sourceText:    "\nexport {\n  Type1,\n  Type2 as Foo,\n  type value1 as BScope,\n  value2 as CScope,\n} from './consistent-type-exports';\n      ",
			wantIds:       []string{"multipleExportsAreTypes"},
			wantSpans:     []string{"export {\n  Type1,\n  Type2 as Foo,\n  type value1 as BScope,\n  value2 as CScope,\n} from './consistent-type-exports';"},
			wantFixed:     "\nexport type { Type1, Type2 as Foo, value1 as BScope } from './consistent-type-exports';\nexport { value2 as CScope } from './consistent-type-exports';\n      ",
		},
		{
			configuration: "{\"fixMixedExportsWithInlineTypeSpecifier\":true}",
			sourceText:    "\nexport {\n  Type1,\n  Type2 as Foo,\n  type value1 as BScope,\n  value2 as CScope,\n} from './consistent-type-exports';\n      ",
			wantIds:       []string{"multipleExportsAreTypes"},
			wantSpans:     []string{"export {\n  Type1,\n  Type2 as Foo,\n  type value1 as BScope,\n  value2 as CScope,\n} from './consistent-type-exports';"},
			wantFixed:     "\nexport {\n  type Type1,\n  type Type2 as Foo,\n  type value1 as BScope,\n  value2 as CScope,\n} from './consistent-type-exports';\n      ",
		},
		{
			configuration: "",
			sourceText:    "\n        import type * as Foo from './consistent-type-exports';\n        type Foo = 1;\n        export { Foo };\n      ",
			wantIds:       []string{"typeOverValue"},
			wantSpans:     []string{"export { Foo };"},
			wantFixed:     "\n        import type * as Foo from './consistent-type-exports';\n        type Foo = 1;\n        export type { Foo };\n      ",
		},
		{
			configuration: "",
			sourceText:    "\n        import { type NAME as Foo } from './consistent-type-exports';\n        export { Foo };\n      ",
			wantIds:       []string{"typeOverValue"},
			wantSpans:     []string{"export { Foo };"},
			wantFixed:     "\n        import { type NAME as Foo } from './consistent-type-exports';\n        export type { Foo };\n      ",
		},
		{
			configuration: "",
			sourceText:    "\n        export * from './consistent-type-exports/type-only-exports';\n      ",
			wantIds:       []string{"typeOverValue"},
			wantSpans:     []string{"export * from './consistent-type-exports/type-only-exports';"},
			wantFixed:     "\n        export type * from './consistent-type-exports/type-only-exports';\n      ",
		},
		{
			configuration: "",
			sourceText:    "\n        /* comment 1 */ export\n          /* comment 2 */ *\n            // comment 3\n            from './consistent-type-exports/type-only-exports';\n      ",
			wantIds:       []string{"typeOverValue"},
			wantSpans:     []string{"export\n          /* comment 2 */ *\n            // comment 3\n            from './consistent-type-exports/type-only-exports';"},
			wantFixed:     "\n        /* comment 1 */ export\n          /* comment 2 */ type *\n            // comment 3\n            from './consistent-type-exports/type-only-exports';\n      ",
		},
		{
			configuration: "",
			sourceText:    "\n        export * from './consistent-type-exports/type-only-reexport';\n      ",
			wantIds:       []string{"typeOverValue"},
			wantSpans:     []string{"export * from './consistent-type-exports/type-only-reexport';"},
			wantFixed:     "\n        export type * from './consistent-type-exports/type-only-reexport';\n      ",
		},
		{
			configuration: "",
			sourceText:    "\n        export * as foo from './consistent-type-exports/type-only-reexport';\n      ",
			wantIds:       []string{"typeOverValue"},
			wantSpans:     []string{"export * as foo from './consistent-type-exports/type-only-reexport';"},
			wantFixed:     "\n        export type * as foo from './consistent-type-exports/type-only-reexport';\n      ",
		},
	}
	for index, testCase := range cases {
		t.Run(consistentTypeExportsCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFilesWithOptions(t, ConsistentTypeExports,
				consistentTypeExportsFixtureFiles(testCase.sourceText),
				consistentTypeExportsFile,
				decodeConsistentTypeExportsOptions(t, testCase.configuration))
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)

			onDisk := consistentTypeExportsOnDisk(testCase.sourceText)
			for index, wantSpan := range testCase.wantSpans {
				reported := onDisk[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
				if reported != wantSpan {
					t.Errorf("finding %d covers %q, want %q", index, reported, wantSpan)
				}
			}

			if testCase.wantFixed == "" {
				for index, diagnostic := range result.Diagnostics {
					if len(diagnostic.Fixes) != 0 {
						t.Errorf("finding %d carries %d fixes, want none", index, len(diagnostic.Fixes))
					}
				}
				return
			}
			rule_testing.ExpectFixedSource(t, result,
				consistentTypeExportsOnDisk(testCase.wantFixed))
		})
	}
}

// TestConsistentTypeExportsSplitKeepsWhatTheSpecifierCarries is the type-safety fixture.
//
// This rule's repair SPLITS a statement, which is the shape the coordinator flagged as the riskiest
// in this batch: the rewrite has to decide what text each specifier keeps, and upstream re-renders
// each one from its names rather than copying its source text.
//
// It is safe because a specifier holds only names. There is no annotation, no generic, no modifier
// and no assertion inside one, so re-rendering can lose nothing a copy would have kept. What it CAN
// lose is spelling, and these rows pin the two spellings that carry more than a bare identifier: an
// alias, and a string-literal export name whose quotes and escapes must survive verbatim.
//
// Every expectation is what the installed 8.67.0 build produced for that exact input.
func TestConsistentTypeExportsSplitKeepsWhatTheSpecifierCarries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantFixed  string
		reason     string
	}{
		{
			sourceText: "import { Type1, Type2, value1 } from './consistent-type-exports/index';\nexport { Type1 as Renamed, value1 };",
			wantFixed:  "import { Type1, Type2, value1 } from './consistent-type-exports/index';\nexport type { Type1 as Renamed };\nexport { value1 };",
			reason:     "an alias survives the split",
		},
		{
			sourceText: "import { Type1, Type2, value1 } from './consistent-type-exports/index';\nexport { Type1 as 'string-name', value1 };",
			wantFixed:  "import { Type1, Type2, value1 } from './consistent-type-exports/index';\nexport type { Type1 as 'string-name' };\nexport { value1 };",
			reason:     "a string-literal export name keeps its quotes",
		},
		{
			sourceText: "import { Type1, Type2, value1 } from './consistent-type-exports/index';\nexport { Type1 as \"double\", value1 };",
			wantFixed:  "import { Type1, Type2, value1 } from './consistent-type-exports/index';\nexport type { Type1 as \"double\" };\nexport { value1 };",
			reason:     "and a double-quoted one keeps its own",
		},
		{
			sourceText: "import { Type1, Type2, value1 } from './consistent-type-exports/index';\nexport { Type1, Type2, value1 };",
			wantFixed:  "import { Type1, Type2, value1 } from './consistent-type-exports/index';\nexport type { Type1, Type2 };\nexport { value1 };",
			reason:     "two type names move together",
		},
		{
			sourceText: "import { Type1, Type2, value1 } from './consistent-type-exports/index';\nexport { type Type1, Type2, value1 };",
			wantFixed:  "import { Type1, Type2, value1 } from './consistent-type-exports/index';\nexport type { Type2, Type1 };\nexport { value1 };",
			reason:     "an inline specifier joins the moved group",
		},
		{
			sourceText: "export { Type1 as Renamed, value1 } from './consistent-type-exports/index';",
			wantFixed:  "export type { Type1 as Renamed } from './consistent-type-exports/index';\nexport { value1 } from './consistent-type-exports/index';",
			reason:     "the source is carried onto the new statement",
		},
	}
	for index, testCase := range cases {
		t.Run(consistentTypeExportsCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFilesWithOptions(t, ConsistentTypeExports,
				consistentTypeExportsFixtureFiles(testCase.sourceText),
				consistentTypeExportsFile, nil)
			if len(result.Diagnostics) == 0 {
				t.Fatalf("expected a finding on %q", testCase.sourceText)
			}
			rule_testing.ExpectFixedSource(t, result,
				consistentTypeExportsOnDisk(testCase.wantFixed))
		})
	}
}

// TestConsistentTypeExportsDiscriminatesOnCasesUpstreamDoesNotWrite covers the rest.
//
// Each row was run through the installed 8.67.0 build and carries the verdict that build produced,
// so a row asserting silence asserts upstream's silence rather than this port's.
func TestConsistentTypeExportsDiscriminatesOnCasesUpstreamDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		sourceText    string
		wantIds       []string
		wantFixed     string
		reason        string
	}{
		{
			configuration: "",
			sourceText:    "import { Type1, Type2, value1 } from './consistent-type-exports/index';\nexport type { Type1 };",
			wantIds:       nil,
			wantFixed:     "",
			reason:        "an already-type export is clean",
		},
		{
			configuration: "",
			sourceText:    "import { Type1, Type2, value1 } from './consistent-type-exports/index';\nexport { value1 };",
			wantIds:       nil,
			wantFixed:     "",
			reason:        "a value export is clean",
		},
		{
			// An unresolvable name is upstream's `undefined` case and it is LEFT ALONE rather than
			// guessed at. The whole imported corpus declares every name it exports, so a mutant
			// bucketing an unresolved specifier as type-based survived all forty-nine cases; these
			// rows are what separates the arm from its neighbours. Measured clean upstream, against
			// the two controls below that report and stay silent respectively.
			configuration: "",
			sourceText:    "export { NotDeclaredAnywhere };",
			wantIds:       nil,
			wantFixed:     "",
			reason:        "an unresolvable specifier is left alone rather than assumed a type",
		},
		{
			configuration: "",
			sourceText:    "const v = 1;\nexport { NotDeclaredAnywhere, v };",
			wantIds:       nil,
			wantFixed:     "",
			reason:        "and it does not drag a resolvable value into a finding either",
		},
		{
			configuration: "",
			sourceText:    "type T = 1;\nexport { T };",
			wantIds:       []string{"typeOverValue"},
			wantFixed:     "type T = 1;\nexport type { T };",
			reason:        "the control: a resolvable type in the same position does report",
		},
		{
			configuration: "",
			sourceText:    "const v = 1;\nexport { v };",
			wantIds:       nil,
			wantFixed:     "",
			reason:        "and a resolvable value is silent",
		},
		{
			configuration: "",
			sourceText:    "import { Type1, Type2, value1 } from './consistent-type-exports/index';\nexport { type Type1, value1 };",
			wantIds:       nil,
			wantFixed:     "",
			reason:        "an inline specifier is already separated",
		},
		{
			configuration: "{\"fixMixedExportsWithInlineTypeSpecifier\": true}",
			sourceText:    "import { Type1, Type2, value1 } from './consistent-type-exports/index';\nexport { Type1, value1 };",
			wantIds:       []string{"singleExportIsType"},
			wantFixed:     "import { Type1, Type2, value1 } from './consistent-type-exports/index';\nexport { type Type1, value1 };",
			reason:        "the inline option keeps one statement",
		},
		{
			configuration: "{\"fixMixedExportsWithInlineTypeSpecifier\": true}",
			sourceText:    "import { Type1, Type2, value1 } from './consistent-type-exports/index';\nexport { Type1, Type2, value1 };",
			wantIds:       []string{"multipleExportsAreTypes"},
			wantFixed:     "import { Type1, Type2, value1 } from './consistent-type-exports/index';\nexport { type Type1, type Type2, value1 };",
			reason:        "and marks each type name",
		},
		{
			configuration: "",
			sourceText:    "export * from './consistent-type-exports/type-only-exports';",
			wantIds:       []string{"typeOverValue"},
			wantFixed:     "export type * from './consistent-type-exports/type-only-exports';",
			reason:        "a type-only star reports",
		},
		{
			configuration: "",
			sourceText:    "export * from './consistent-type-exports/index';",
			wantIds:       nil,
			wantFixed:     "",
			reason:        "a star over a module with values is clean",
		},
		{
			configuration: "",
			sourceText:    "export type * from './consistent-type-exports/type-only-exports';",
			wantIds:       nil,
			wantFixed:     "",
			reason:        "an already-type star is clean",
		},
	}
	for index, testCase := range cases {
		t.Run(consistentTypeExportsCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFilesWithOptions(t, ConsistentTypeExports,
				consistentTypeExportsFixtureFiles(testCase.sourceText),
				consistentTypeExportsFile,
				decodeConsistentTypeExportsOptions(t, testCase.configuration))
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if testCase.wantFixed != "" {
				rule_testing.ExpectFixedSource(t, result,
					consistentTypeExportsOnDisk(testCase.wantFixed))
			}
		})
	}
}

// TestConsistentTypeExportsNamesTheExportsUpstreamNames pins the interpolated message.
//
// Two of the three messages carry the offending names, joined by upstream's `formatWordList`: one
// name alone, two joined by ` and `, and three or more joined by commas with ` and ` before the
// last, with no comma before the conjunction. A message-id assertion cannot see any of that, and a
// mutant collapsing the join to plain commas survived every other fixture in this file.
//
// Each expectation is the name list the installed 8.67.0 build printed for that exact input,
// recovered from its message text, and compared against a literal typed here rather than against
// the rule's own constant, since a constant moves with the mutation.
func TestConsistentTypeExportsNamesTheExportsUpstreamNames(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantId     string
		wantNames  string
	}{
		{"type A = 1;\nconst v = 4;\nexport { A, v };", "singleExportIsType", "A"},
		{"type A = 1;\ntype B = 2;\nconst v = 4;\nexport { A, B, v };", "multipleExportsAreTypes", "A and B"},
		{"type A = 1;\ntype B = 2;\ntype C = 3;\nconst v = 4;\nexport { A, B, C, v };", "multipleExportsAreTypes", "A, B and C"},
	}
	for index, testCase := range cases {
		t.Run(consistentTypeExportsCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFilesWithOptions(t, ConsistentTypeExports,
				consistentTypeExportsFixtureFiles(testCase.sourceText),
				consistentTypeExportsFile, nil)
			rule_testing.ExpectFindings(t, result, testCase.wantId)

			var want string
			switch testCase.wantId {
			case "singleExportIsType":
				want = messageConsistentTypeExportsSingleExportIsType(testCase.wantNames).Description
			default:
				want = messageConsistentTypeExportsMultipleExportsAreTypes(testCase.wantNames).Description
			}
			if result.Diagnostics[0].Message.Description != want {
				t.Errorf("the description is %q, want %q",
					result.Diagnostics[0].Message.Description, want)
			}
		})
	}
}

// TestConsistentTypeExportsDecoderResolvesTheOption puts the decoder under test.
func TestConsistentTypeExportsDecoderResolvesTheOption(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		wantInline    bool
	}{
		{"{}", false},
		{"{\"fixMixedExportsWithInlineTypeSpecifier\": true}", true},
		{"{\"fixMixedExportsWithInlineTypeSpecifier\": false}", false},
	}
	for index, testCase := range cases {
		t.Run(consistentTypeExportsCaseName(index), func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeConsistentTypeExportsOptions([]byte(testCase.configuration))
			if err != nil {
				t.Fatalf("decoding %s: %v", testCase.configuration, err)
			}
			options, ok := decoded.(ConsistentTypeExportsOptions)
			if !ok {
				t.Fatalf("the decoder returned %T rather than the options struct", decoded)
			}
			if options.FixMixedExportsWithInlineTypeSpecifier != testCase.wantInline {
				t.Errorf("the inline option resolved to %v, want %v",
					options.FixMixedExportsWithInlineTypeSpecifier, testCase.wantInline)
			}
		})
	}
}

// TestConsistentTypeExportsNilOptionsUsesTheSplittingRepair bypasses the decoder.
//
// A rule configured as a bare `"error"` is handed nil, and the zero value happens to match
// upstream's default here. Asserted anyway so a later default change fails loudly rather than
// silently switching which repair the tree gets.
func TestConsistentTypeExportsNilOptionsUsesTheSplittingRepair(t *testing.T) {
	t.Parallel()

	source := "import { Type1, value1 } from './consistent-type-exports/index';\nexport { Type1, value1 };"
	result := rule_testing.RunTypedFilesWithOptions(t, ConsistentTypeExports,
		consistentTypeExportsFixtureFiles(source), consistentTypeExportsFile, nil)
	rule_testing.ExpectFindings(t, result, "singleExportIsType")
	rule_testing.ExpectFixedSource(t, result, consistentTypeExportsOnDisk(
		"import { Type1, value1 } from './consistent-type-exports/index';\nexport type { Type1 };\nexport { value1 };"))
}

// TestConsistentTypeExportsRequiresTheTypedHarness pins the checker declaration.
//
// A typed rule handed the plain harness gets a nil checker, and the guard at the top of the listener
// turns that into silence rather than a panic. Silence is the more dangerous failure: every clean
// case passes vacuously and every reporting case fails in a way that reads as a rule bug.
func TestConsistentTypeExportsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	source := "import { Type1 } from './consistent-type-exports/index';\nexport { Type1 };"
	rule_testing.ExpectClean(t, rule_testing.Run(t, ConsistentTypeExports,
		consistentTypeExportsFile, source))
	rule_testing.ExpectFindings(t, rule_testing.RunTypedFiles(t, ConsistentTypeExports,
		consistentTypeExportsFixtureFiles(source), consistentTypeExportsFile), "typeOverValue")
}
