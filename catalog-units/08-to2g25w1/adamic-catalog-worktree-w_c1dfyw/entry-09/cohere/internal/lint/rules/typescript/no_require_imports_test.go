package typescript

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// requireImportsFile is where the fixtures pretend to live. The extension matters: upstream runs
// this corpus through `change_rule_path_extension("ts")`, and the import-equals form is TypeScript
// syntax that a `.js` file cannot hold.
const requireImportsFile = "/repository/source/RequireImports.ts"

// requireImportsCase is one upstream row: the source, the options as they appear in the config, and
// how many findings the snapshot records for it.
type requireImportsCase struct {
	sourceText string
	options    string
	findings   int
}

// decodeRequireImportsOptions routes a case's options through the rule's own exported decoder
// rather than building the struct, so that the string-to-regexp compilation and the `allowAsImport`
// default are themselves under test. A case with no options is handed nil, which is what a rule
// configured as a bare "error" actually receives.
func decodeRequireImportsOptions(t *testing.T, optionsText string) any {
	t.Helper()
	if optionsText == "" {
		return nil
	}
	decoded, err := DecodeNoRequireImportsOptions(json.RawMessage(optionsText))
	if err != nil {
		t.Fatalf("decoding the options %s: %v", optionsText, err)
	}
	return decoded
}

// The corpus is oxc's, copied rather than rewritten, from
// `oxc/crates/oxc_linter/src/rules/typescript/no_require_imports.rs`: 26 pass and 27 fail across one
// tester block. The snapshot records 30 diagnostics from those 27 inputs, so one finding per input
// would be wrong here. Three inputs report twice and each is a genuine second call rather than a
// doubled report: the two `lib5`/`lib6` cases hold two declarators with a `require` in each, and the
// `configValidator` case holds two statements with a `require` in each. Every other input reports
// once, recovered by aligning each snapshot entry on the source line it prints.
func TestNoRequireImportsFires(t *testing.T) {
	t.Parallel()

	cases := []requireImportsCase{
		{sourceText: "var lib = require('lib');", options: "", findings: 1},
		{sourceText: "let lib2 = require('lib2');", options: "", findings: 1},
		{sourceText: "\n            var lib5 = require('lib5'),\n              lib6 = require('lib6');\n                  ", options: "", findings: 2},
		{sourceText: "import lib8 = require('lib8');", options: "", findings: 1},
		{sourceText: "var lib = require?.('lib');", options: "", findings: 1},
		{sourceText: "let lib2 = require?.('lib2');", options: "", findings: 1},
		{sourceText: "\n            var lib5 = require?.('lib5'),\n              lib6 = require?.('lib6');\n                  ", options: "", findings: 2},
		{sourceText: "const pkg = require('./package.json');", options: "", findings: 1},
		{sourceText: "const pkg = require('./package.jsonc');", options: "{ \"allow\": [\"/package\\\\.json$\"] }", findings: 1},
		{sourceText: "const pkg = require(`./package.jsonc`);", options: "{ \"allow\": [\"/package\\\\.json$\"] }", findings: 1},
		{sourceText: "import pkg = require('./package.json');", options: "", findings: 1},
		{sourceText: "import pkg = require('./package.jsonc');", options: "{ \"allow\": [\"/package\\\\.json$\"] }", findings: 1},
		{sourceText: "import pkg = require('./package.json');", options: "{ \"allow\": [\"^some-package$\"] }", findings: 1},
		{sourceText: "var foo = require?.('foo');", options: "{ \"allowAsImport\": true }", findings: 1},
		{sourceText: "let foo = trick(require?.('foo'));", options: "{ \"allowAsImport\": true }", findings: 1},
		{sourceText: "trick(require('foo'));", options: "{ \"allowAsImport\": true }", findings: 1},
		{sourceText: "const foo = require('./foo.json') as Foo;", options: "{ \"allowAsImport\": true }", findings: 1},
		{sourceText: "const foo: Foo = require('./foo.json').default;", options: "{ \"allowAsImport\": true }", findings: 1},
		{sourceText: "const foo = <Foo>require('./foo.json');", options: "{ \"allowAsImport\": true }", findings: 1},
		{sourceText: "\n            const configValidator = new Validator(require('./a.json'));\n            configValidator.addSchema(require('./a.json'));\n                  ", options: "{ \"allowAsImport\": true }", findings: 2},
		{sourceText: "require('foo');", options: "{ \"allowAsImport\": true }", findings: 1},
		{sourceText: "require?.('foo');", options: "{ \"allowAsImport\": true }", findings: 1},
		{sourceText: "function foo() {\n            require('foo')\n            }", options: "", findings: 1},
		{sourceText: "const m = require(someVariable);", options: "", findings: 1},
		{sourceText: "const m = require(path.join(dir, file));", options: "", findings: 1},
		{sourceText: "class Foo {\n            require(module: string) {\n                return require(module);\n            }\n            }", options: "", findings: 1},
		{sourceText: "class Foo {\n            require(module: string) {\n                return require('foo');\n            }\n            }", options: "", findings: 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoRequireImports, requireImportsFile,
				testCase.sourceText, decodeRequireImportsOptions(t, testCase.options))
			expected := make([]string, testCase.findings)
			for index := range expected {
				expected[index] = "noRequireImports"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// The clean cases carry the whole discrimination and they fail in four distinct ways. `load(...)`
// and `lib2.subImport` are not `require` at all. `import lib9 = lib2.anotherSubImport` is the
// qualified-name form of import-equals, which produces no external module reference node. The
// `createRequire` and `let require = bazz` cases each declare a real local binding, which is the
// only thing that makes a genuine `require` call silent. The remainder exercise the two options.
func TestNoRequireImportsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []requireImportsCase{
		{sourceText: "import { l } from 'lib';", options: ""},
		{sourceText: "var lib3 = load('not_an_import');", options: ""},
		{sourceText: "var lib4 = lib2.subImport;", options: ""},
		{sourceText: "var lib7 = 700;", options: ""},
		{sourceText: "import lib9 = lib2.anotherSubImport;", options: ""},
		{sourceText: "import lib10 from 'lib10';", options: ""},
		{sourceText: "var lib3 = load?.('not_an_import');", options: ""},
		{sourceText: "\n            import { createRequire } from 'module';\n            const require = createRequire();\n            require(someModule);\n                ", options: ""},
		{sourceText: "\n            function foo() {\n                let require = bazz;\n                require(someModule);\n            }\n                ", options: ""},
		{sourceText: "\n            import { createRequire } from 'module';\n            const require = createRequire();\n            require('remark-preset-prettier');\n                ", options: ""},
		{sourceText: "const pkg = require('./package.json');", options: "{ \"allow\": [\"/package\\\\.json$\"] }"},
		{sourceText: "const pkg = require('../package.json');", options: "{ \"allow\": [\"/package\\\\.json$\"] }"},
		{sourceText: "const pkg = require(`./package.json`);", options: "{ \"allow\": [\"/package\\\\.json$\"] }"},
		{sourceText: "const pkg = require('../packages/package.json');", options: "{ \"allow\": [\"/package\\\\.json$\"] }"},
		{sourceText: "import pkg = require('../packages/package.json');", options: "{ \"allow\": [\"/package\\\\.json$\"] }"},
		{sourceText: "import pkg = require('data.json');", options: "{ \"allow\": [\"\\\\.json$\"] }"},
		{sourceText: "import pkg = require('some-package');", options: "{ \"allow\": [\"^some-package$\"] }"},
		{sourceText: "import foo = require('foo');", options: "{ \"allowAsImport\": true }"},
		{sourceText: "\n            let require = bazz;\n            trick(require('foo'));\n                  ", options: "{ \"allowAsImport\": true }"},
		{sourceText: "\n            let require = bazz;\n            const foo = require('./foo.json') as Foo;\n                  ", options: "{ \"allowAsImport\": true }"},
		{sourceText: "\n            let require = bazz;\n            const foo: Foo = require('./foo.json').default;\n                  ", options: "{ \"allowAsImport\": true }"},
		{sourceText: "\n            let require = bazz;\n            const foo = <Foo>require('./foo.json');\n                  ", options: "{ \"allowAsImport\": true }"},
		{sourceText: "\n            let require = bazz;\n            const configValidator = new Validator(require('./a.json'));\n            configValidator.addSchema(require('./a.json'));\n                  ", options: "{ \"allowAsImport\": true }"},
		{sourceText: "\n            let require = bazz;\n            require('foo');\n                  ", options: "{ \"allowAsImport\": true }"},
		{sourceText: "\n            let require = bazz;\n            require?.('foo');\n                  ", options: "{ \"allowAsImport\": true }"},
		{sourceText: "\n            import { createRequire } from 'module';\n            const require = createRequire();\n            require('remark-preset-prettier');\n                  ", options: "{ \"allowAsImport\": true }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoRequireImports,
				requireImportsFile, testCase.sourceText,
				decodeRequireImportsOptions(t, testCase.options)))
		})
	}
}

// TestNoRequireImportsSpans pins where each arm points, which no message-id assertion can see.
//
// The import-equals span is the one place oxc and `@typescript-eslint` visibly disagree: oxc reports
// the whole `KindImportEqualsDeclaration`, semicolon included, and `@typescript-eslint` reports the
// inner `TSExternalModuleReference`. Both carry the same message id, so every id fixture above stays
// green over the wrong choice. The expected strings here are typed literals rather than anything
// derived from the rule, so a mutation cannot move both sides together.
func TestNoRequireImportsSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    string
		reported   []string
	}{
		{
			name:       "a call reports the call and not the declarator",
			sourceText: "var lib = require('lib');",
			reported:   []string{"require('lib')"},
		},
		{
			name:       "an optional call includes the optional token",
			sourceText: "var lib = require?.('lib');",
			reported:   []string{"require?.('lib')"},
		},
		{
			name:       "an import-equals reports the whole declaration including the semicolon",
			sourceText: "import lib8 = require('lib8');",
			reported:   []string{"import lib8 = require('lib8');"},
		},
		{
			name:       "an import-equals with no semicolon ends at the reference",
			sourceText: "import lib8 = require('lib8')",
			reported:   []string{"import lib8 = require('lib8')"},
		},
		{
			name:       "two declarators report separately",
			sourceText: "var a = require('x'), b = require('y');",
			reported:   []string{"require('x')", "require('y')"},
		},
		{
			name:       "a call inside a type assertion reports only the call",
			sourceText: "const foo = require('./foo.json') as Foo;",
			reported:   []string{"require('./foo.json')"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoRequireImports, requireImportsFile,
				testCase.sourceText, decodeRequireImportsOptions(t, testCase.options))
			if len(result.Diagnostics) != len(testCase.reported) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.reported))
			}
			for index, diagnostic := range result.Diagnostics {
				got := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if got != testCase.reported[index] {
					t.Errorf("finding %d points at %q, want %q", index, got, testCase.reported[index])
				}
			}
		})
	}
}

// TestNoRequireImportsMessage asserts the message a reader sees, against literals typed here rather
// than against the rule's own constant. Comparing to the constant is equality that looks correct and
// moves with the rule under mutation, so it proves nothing.
func TestNoRequireImportsMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTypedWithOptions(t, NoRequireImports, requireImportsFile,
		"var lib = require('lib');", nil)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
	}
	if result.Diagnostics[0].Message.Id != "noRequireImports" {
		t.Errorf("message id is %q, want %q", result.Diagnostics[0].Message.Id, "noRequireImports")
	}
	if !strings.HasPrefix(result.Diagnostics[0].Message.Description,
		"This loads a module with a CommonJS `require` call.") {
		t.Errorf("message description begins %q", result.Diagnostics[0].Message.Description)
	}
	if len(result.Diagnostics[0].Fixes) != 0 || len(result.Diagnostics[0].Suggestions) != 0 {
		t.Errorf("the rule offers a repair; upstream marks its fixer pending and ships none")
	}
}

// TestNoRequireImportsMeasuredAgainstTheReleaseBinary covers branches upstream's corpus never
// exercises, each pinned by running `oxlint` on the input rather than by reading oxc's source. The
// corpus writes no parenthesized callee and no escaped path, so every one of these would have been a
// silent divergence in either direction.
func TestNoRequireImportsMeasuredAgainstTheReleaseBinary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    string
		findings   int
		// reasoning records the measurement, so a later reader does not have to re-run it.
		reasoning string
	}{
		{
			name:       "a parenthesized callee still reports",
			sourceText: "(require)('lib');",
			findings:   1,
			reasoning: "oxc reaches the callee through get_identifier_reference, which skips " +
				"parentheses; our AST keeps the paren node, so the skip had to be added rather " +
				"than omitted. Release binary reports this at column 1.",
		},
		{
			name:       "a doubly parenthesized callee still reports",
			sourceText: "((require))('lib');",
			findings:   1,
			reasoning:  "the skip is transitive upstream. Release binary reports.",
		},
		{
			name:       "a template path is matched on raw text, so an escape defeats allow",
			sourceText: "require(`a\\u0062c`);",
			options:    `{ "allow": ["^abc$"] }`,
			findings:   1,
			reasoning: "oxc reads quasi.value.raw for a template. The raw text is a\\u0062c, which " +
				"does not match ^abc$. Release binary reports. Our .Text is cooked, so reading it " +
				"here would have gone silent.",
		},
		{
			name:       "a string path is matched on cooked text, so the same escape is allowed",
			sourceText: `require("abc");`,
			options:    `{ "allow": ["^abc$"] }`,
			findings:   0,
			reasoning: "oxc reads string_literal.value, which is cooked. The cooked value is abc, " +
				"which matches. Release binary is silent. One option, two answers, split by " +
				"literal kind.",
		},
		{
			name:       "a call with no arguments reports",
			sourceText: "require();",
			options:    `{ "allow": ["^abc$"] }`,
			findings:   1,
			reasoning:  "there is no argument to match, so the allow check cannot fire. Release binary reports.",
		},
		{
			name:       "a spread argument is not a static path and reports",
			sourceText: "require(...args);",
			options:    `{ "allow": ["^abc$"] }`,
			findings:   1,
			reasoning:  "a spread element is neither literal kind. Release binary reports.",
		},
		{
			name:       "a template with a substitution compares only its first quasi",
			sourceText: "require(`./a${b}c`);",
			options:    `{ "allow": ["^\\./a"] }`,
			findings:   0,
			reasoning: "upstream takes quasis.first() with no static check, so the prefix ./a " +
				"matches and the call is permitted even though the assembled path is unknown. " +
				"This reads as a defect and this port first implemented the opposite; the release " +
				"binary is silent on this exact input, so the intuition was wrong.",
		},
		{
			name:       "the first quasi ends at the substitution rather than including its opener",
			sourceText: "require(`./a${b}c`);",
			options:    `{ "allow": ["^\\./a$"] }`,
			findings:   0,
			reasoning: "the head's raw source runs from the backtick to the ${ that opens the " +
				"substitution, so three delimiter characters come off rather than two. An " +
				"end-anchored pattern is the only kind that can tell ./a from ./a$, which is why " +
				"the prefix-anchored case above cannot see this. Release binary is silent.",
		},
		{
			name:       "a first quasi that does not match still reports",
			sourceText: "require(`./z${b}c`);",
			options:    `{ "allow": ["^\\./a"] }`,
			findings:   1,
			reasoning: "the companion to the case above: the first-quasi comparison is a real " +
				"comparison rather than an unconditional permit.",
		},
		{
			name:       "allowAsImport does not exempt a bare call",
			sourceText: "require('foo');",
			options:    `{ "allowAsImport": true }`,
			findings:   1,
			reasoning: "upstream's own corpus pins this: the same input is a fail case under " +
				"allowAsImport, and becomes a pass only when a local binding is added above it.",
		},
		{
			name:       "allow beats allowAsImport being unset on an import-equals",
			sourceText: "import pkg = require('./package.json');",
			options:    `{ "allow": ["/package\\.json$"] }`,
			findings:   0,
			reasoning:  "the allow check runs before the allowAsImport check, matching upstream's order.",
		},
		{
			name:       "a qualified-name import-equals is not an external module reference",
			sourceText: "import lib9 = lib2.anotherSubImport;",
			findings:   0,
			reasoning: "declined by the listener kind rather than by a test: the parser produces no " +
				"KindExternalModuleReference node for this form at all.",
		},
		{
			name:       "a method named require does not shadow the call inside it",
			sourceText: "class Foo {\n  require(module: string) {\n    return require('foo');\n  }\n}",
			findings:   1,
			reasoning: "a method is a member of the class rather than a binding in the scope, so " +
				"the call still reaches the global. Upstream ships two such fail cases.",
		},
		{
			name:       "a local binding one scope out still shadows",
			sourceText: "let require = bazz;\nfunction foo() {\n  return require('x');\n}",
			findings:   0,
			reasoning: "the checker resolves through enclosing scopes, which is why there is no " +
				"bounded syntactic answer and the rule reads the checker at all.",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoRequireImports, requireImportsFile,
				testCase.sourceText, decodeRequireImportsOptions(t, testCase.options))
			if len(result.Diagnostics) != testCase.findings {
				t.Fatalf("got %d findings, want %d (%s)",
					len(result.Diagnostics), testCase.findings, testCase.reasoning)
			}
		})
	}
}

// TestNoRequireImportsRequiresTheTypedHarness asserts that the rule genuinely declines when handed
// no checker, so a later revert of the nil guard fails loudly rather than reporting on every
// `require` in the file regardless of shadowing.
func TestNoRequireImportsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoRequireImports.NeedsTypeChecker {
		t.Fatal("the rule stopped declaring NeedsTypeChecker; the shadow test cannot work without it")
	}
	result := rule_testing.RunWithOptions(t, NoRequireImports, requireImportsFile,
		"var lib = require('lib');", nil)
	if len(result.Diagnostics) != 0 {
		t.Errorf("the untyped harness produced %d findings; the nil-checker guard is missing",
			len(result.Diagnostics))
	}
}

// TestDecodeNoRequireImportsOptions pins the decoder itself, which is where the default inversion
// and the pattern compilation live and where upstream has no counterpart to compare against.
func TestDecodeNoRequireImportsOptions(t *testing.T) {
	t.Parallel()

	t.Run("an absent allowAsImport is false", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoRequireImportsOptions(json.RawMessage(`{}`))
		if err != nil {
			t.Fatalf("decoding: %v", err)
		}
		options, _ := decoded.(NoRequireImportsOptions)
		if options.AllowAsImport {
			t.Error("allowAsImport defaulted to true; both upstreams default it to false")
		}
		if len(options.Allow) != 0 {
			t.Errorf("allow decoded %d patterns from an empty object", len(options.Allow))
		}
	})

	t.Run("patterns compile", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoRequireImportsOptions(
			json.RawMessage(`{ "allow": ["/package\\.json$", "^some-package$"] }`))
		if err != nil {
			t.Fatalf("decoding: %v", err)
		}
		options, _ := decoded.(NoRequireImportsOptions)
		if len(options.Allow) != 2 {
			t.Fatalf("compiled %d patterns, want 2", len(options.Allow))
		}
		if !options.Allow[0].MatchString("../packages/package.json") {
			t.Error("the escaped-dot pattern did not match a path upstream permits")
		}
		if options.Allow[0].MatchString("./package.jsonc") {
			t.Error("the anchored pattern matched a path upstream reports")
		}
	})

	t.Run("a pattern Go cannot compile is dropped rather than failing the run", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoRequireImportsOptions(
			json.RawMessage(`{ "allow": ["(?<=x)y", "^ok$"] }`))
		if err != nil {
			t.Fatalf("decoding: %v", err)
		}
		options, _ := decoded.(NoRequireImportsOptions)
		if len(options.Allow) != 1 {
			t.Fatalf("kept %d patterns, want 1; RE2 has no lookbehind", len(options.Allow))
		}
		if !options.Allow[0].MatchString("ok") {
			t.Error("the surviving pattern is not the compilable one")
		}
	})
}

// TestNoRequireImportsSurvivesMalformedImportEquals pins two guards that are crash protection
// rather than behavioral filters, which no ExpectFindings fixture can see: a mutation sweep reports
// SURVIVED whether such a guard matters or not, because a panic is not a finding.
//
// `import a = require(`lib`)` is a parse error in oxc, confirmed by running the release binary,
// which prints `Unexpected token` and no rule diagnostic. Our parser recovers from it and hands the
// arm a KindNoSubstitutionTemplateLiteral, so an unguarded AsStringLiteral is an unchecked interface
// conversion that takes the whole linter down. Measured by removing the kind test and watching this
// input panic with `ast.nodeData is *ast.NoSubstitutionTemplateLiteral, not *ast.StringLiteral`.
func TestNoRequireImportsSurvivesMalformedImportEquals(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    string
	}{
		{
			name:       "a template in an import-equals, which only error recovery produces",
			sourceText: "import a = require(`lib`);",
			options:    `{ "allow": ["^lib$"] }`,
		},
		{
			name:       "an identifier in an import-equals",
			sourceText: "import a = require(someVariable);",
			options:    `{ "allow": ["^lib$"] }`,
		},
		{
			name:       "an import-equals with no argument at all",
			sourceText: "import a = require();",
			options:    `{ "allow": ["^lib$"] }`,
		},
		{
			name:       "an unterminated import-equals",
			sourceText: "import a = require('a', 'b');",
			options:    `{ "allow": ["^lib$"] }`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// The assertion is that this returns at all. A panic here fails the test by crashing it,
			// which is the only signal available for this class of guard.
			rule_testing.RunTypedWithOptions(t, NoRequireImports, requireImportsFile,
				testCase.sourceText, decodeRequireImportsOptions(t, testCase.options))
		})
	}
}

// TestNoRequireImportsAcrossDeclarationOrderings pins the loop over symbol.Declarations, which an
// index-zero read passes every other fixture in this file while being wrong.
//
// The distinguishing shape is a symbol carrying a declaration-file declaration at index 0 and a
// source declaration at index 1. `declare global { function require(...) }` in a module produces
// exactly that, measured on the harness: two declarations, the ambient one first.
//
// It also pins a divergence that reading the code alone would have got backwards. A source
// declaration is not automatically a shadow: measured on the release binary, upstream REPORTS the
// call beside a `declare global` block and is SILENT beside a plain `function require(...)`, even
// though both put a declaration in a source file. An ambient declaration declares a name that
// already exists rather than introducing a binding, and the ambient flag is what separates them.
func TestNoRequireImportsAcrossDeclarationOrderings(t *testing.T) {
	t.Parallel()

	const ambientTypes = "declare function require(id: string): any;\n"

	cases := []struct {
		name      string
		files     map[string]string
		findings  int
		reasoning string
	}{
		{
			name: "a declare-global block does not shadow, and its source declaration is at index 1",
			files: map[string]string{
				"globals.d.ts": ambientTypes,
				"Subject.ts": "export {};\n" +
					"declare global { function require(id: boolean): any; }\n" +
					"require('lib');",
			},
			findings: 1,
			reasoning: "two declarations, .d.ts first and source second, both ambient. An " +
				"index-zero read would see only the .d.ts one and happen to report; the loop sees " +
				"both and reports because neither is a real binding. Release binary reports.",
		},
		{
			name: "a plain source function declaration does shadow",
			files: map[string]string{
				"globals.d.ts": ambientTypes,
				"Subject.ts":   "function require(id: number): any { return id; }\nrequire('lib');",
			},
			findings: 0,
			reasoning: "one declaration, in source, not ambient. Release binary is silent. This is " +
				"the companion that makes the case above a real discrimination rather than a " +
				"blanket report.",
		},
		{
			name: "an ambient overload above a real implementation puts the shadow at index 1",
			files: map[string]string{
				"Subject.ts": "declare function require(id: string): any;\n" +
					"function require(id: any): any { return id; }\n" +
					"require('lib');",
			},
			findings: 0,
			reasoning: "two declarations in one source file: the ambient overload at index 0 and " +
				"the real implementation at index 1. An index-zero read sees only the ambient one, " +
				"skips it, and reports. The loop reaches the implementation and stays silent, " +
				"which is what the release binary does. This is the input that separates the two " +
				"after the ambient exclusion removed the declare-global one.",
		},
		{
			name: "a source binding shadows the ambient require from @types/node",
			files: map[string]string{
				"globals.d.ts": ambientTypes,
				"Subject.ts":   "const require = createRequire();\nrequire('lib');",
			},
			findings: 0,
			reasoning: "the shadow wins outright: one source declaration and the ambient one is " +
				"not in the symbol at all.",
		},
		{
			name: "the ambient require alone reports",
			files: map[string]string{
				"globals.d.ts": ambientTypes,
				"Subject.ts":   "require('lib');",
			},
			findings: 1,
			reasoning: "one declaration, in a declaration file. This is the real tree's shape and " +
				"the one the fixture harness cannot reach without a second file.",
		},
		{
			name: "two declaration files each declaring require still report",
			files: map[string]string{
				"globals.d.ts":  ambientTypes,
				"globals2.d.ts": "declare function require(id: number): any;\n",
				"Subject.ts":    "require('lib');",
			},
			findings: 1,
			reasoning: "two declarations, both in declaration files. This is why the loop cannot " +
				"stop at the first match in the other direction either.",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFiles(t, NoRequireImports, testCase.files, "Subject.ts")
			if len(result.Diagnostics) != testCase.findings {
				t.Fatalf("got %d findings, want %d (%s)",
					len(result.Diagnostics), testCase.findings, testCase.reasoning)
			}
		})
	}
}
