package typescript

import (
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoRequireImportsOptions tunes which `require` forms this rule permits.
type NoRequireImportsOptions struct {
	// Allow holds patterns tested against the required path. A path matching any of them is
	// permitted.
	//
	// The usual reason to set this is `package.json`. That file commonly lives outside the
	// TypeScript root directory, so importing it statically drags the root directory outward and
	// pulls the whole package tree into the program. The same escape hatch covers any JSON in a
	// toolchain without JSON modules.
	//
	// Upstream compiles each string with the `u` flag; see compileAllowPatterns for what that
	// costs us and what it does not.
	Allow []*regexp.Regexp

	// AllowAsImport permits `import name = require('path')` while still reporting a bare call.
	//
	// False by default, which is upstream's default on both implementations: oxc derives `Default`
	// on a plain `bool` field, and `@typescript-eslint`'s `defaultOptions` writes
	// `allowAsImport: false` outright. The option exists for a codebase pinned to CommonJS interop
	// semantics, where the import-equals form is the only spelling that produces the right runtime
	// shape, and where a bare `require` call is still a defect.
	AllowAsImport bool
}

// noRequireImportsRawOptions is the wire shape, kept separate because Allow arrives as strings and
// is compiled once at decode time rather than per file.
type noRequireImportsRawOptions struct {
	Allow         []string `json:"allow"`
	AllowAsImport bool     `json:"allowAsImport"`
}

var messageNoRequireImports = rule.Message{
	Id: "noRequireImports",
	Description: "This loads a module with a CommonJS `require` call. A `require` is an ordinary " +
		"function call evaluated at runtime, so a missing module or a mistyped path is a crash " +
		"when the line executes rather than an error the compiler reports, and a bundler cannot " +
		"see through it to drop the parts of the module nothing uses. Scattered calls also hide " +
		"what a file depends on, which an import list at the top makes plain. Use an `import` " +
		"declaration, or `import()` where the load genuinely has to be deferred.",
}

// NoRequireImports flags CommonJS `require`, both as a call and as an import-equals declaration.
//
//	valid:   import { l } from 'lib';
//	valid:   var lib4 = lib2.subImport;
//	valid:   import lib9 = lib2.anotherSubImport;
//	valid:   import { createRequire } from 'module'; const require = createRequire(); require(m);
//	valid:   function foo() { let require = bazz; require(someModule); }
//	invalid: var lib = require('lib');
//	invalid: var lib = require?.('lib');
//	invalid: import lib8 = require('lib8');
//	invalid: class Foo { require(m: string) { return require(m); } }
//
// Ported from oxc's `no_require_imports`, which is what the gate runs. `@typescript-eslint` ships
// the same rule and the two disagree in one visible way, recorded at the report site below.
//
// # Two arms, one rule, and they cannot both fire on one node
//
// A bare call and `import name = require('path')` are the same finding upstream and share a message,
// but they are different nodes and the split is not cosmetic. In an import-equals the `require`
// token is not an identifier at all: the parser folds the whole `require('path')` into a single
// `KindExternalModuleReference` whose only child is the path expression. Probed on our own parser,
// an identifier listener over `import lib8 = require('lib8')` sees no `require` at all. So the two
// arms partition the inputs rather than overlapping, and no input can report twice from one node.
//
// The same probe settled the other half: `import lib9 = lib2.anotherSubImport` produces no
// `KindExternalModuleReference` node whatsoever, so the qualified-name form upstream explicitly
// ignores is declined by the listener kind rather than by a test inside it.
//
// # Why this reads the checker, and why resolvesToAGlobal is the wrong helper
//
// The rule must separate the CommonJS `require` from a local binding somebody named `require`.
// Upstream asks `is_global_reference_name`, which is name resolution rather than a scope flag, and
// there is no structural answer: the shadow can be a `const` from `createRequire()`, a `let` in an
// enclosing function, a parameter, or an import, at any depth.
//
// `resolvesToAGlobal` is the shipped helper for "global or a local shadowing it" and it is the
// wrong test here, in the direction that silences the rule rather than the one that over-reports.
// It answers **false** for a symbol the checker cannot resolve, and that is the common case for
// `require`: measured on our own harness, a bare `require('lib')` in a program without `@types/node`
// resolves to a **nil symbol**. Using that helper would have made the rule report nothing at all in
// exactly the trees it exists for, and every fixture would still have been green, because the
// fixture harness has no `@types/node` either and so agrees with the real tree only by accident.
//
// The predicate is therefore the inverse and it is written to answer the same way in both worlds.
// Measured four ways against the harness:
//
//	no @types/node, bare require        symbol is nil                         report
//	@types/node present, bare require   one KindFunctionDeclaration, in a .d.ts  report
//	shadowed by a local binding         one KindVariableDeclaration, in source   silent
//	class method named `require`        nil, or the .d.ts declaration            report
//
// The last row is upstream's own corpus and it is the one that shows the predicate is about scope
// bindings rather than about the spelling: a method named `require` is a member of the class, not a
// binding in the scope the call sits in, so the call still reaches the global and still reports.
// Upstream ships two such cases and both report.
//
// So a source declaration is the only thing that buys silence, and the loop is over every
// declaration rather than the first: a `require` symbol genuinely carries more than one declaration
// when two declaration files declare it, which was measured rather than assumed.
//
// # Parentheses, measured rather than guessed
//
// oxc reaches the callee through `get_identifier_reference()`, which skips parentheses. Our AST
// keeps the parenthesis as a real node, so reproducing upstream means **adding** a skip that is not
// in upstream's text. Pinned against the release binary: `(require)('lib')` reports with the span
// starting at column 1, `((require))('lib')` reports, and `(require('lib'))` reports with the span
// starting at column 2, which is the inner call rather than the wrapper. The corpus writes no
// parenthesized form, so guessing either way would have cost nothing at fixture time and shipped a
// divergence.
//
// # The `allow` option matches a template on RAW text and a string on COOKED text
//
// One option, two answers, split by literal kind, and this is invisible from the corpus because
// every corpus path is escape-free. oxc reads `quasi.value.raw` for a template and
// `string_literal.value` for a string. Our `.Text` is the cooked value for both, so the template
// arm has to slice the source instead. Pinned against the release binary with `allow: ["^abc$"]`:
//
//	require(`abc`)   raw is abc, no match, REPORTS
//	require("abc")   cooked is abc,   matches,  silent
//
// A port reading `.Text` for both would have gone silent on the first, and no imported fixture
// could have seen it.
//
// # Only a string literal participates in `allow` on the import-equals arm
//
// oxc's `ExternalModuleReference` arm reads `mod_ref.expression.value`, a field that exists only on
// a string literal, so a template there can never match an allow pattern. In oxc that arm is
// unreachable: `import a = require(`lib`)` is a parse error, confirmed by running the
// release binary, which reports `Unexpected token` and no rule diagnostic. Our parser recovers from
// it and hands the arm a `KindNoSubstitutionTemplateLiteral`, so the branch is reachable here and
// upstream has no answer to copy. Following upstream's code, a non-string expression simply does
// not match, and the declaration reports. Recorded because the intuitive reading is the opposite
// one, and the next reader would otherwise helpfully "fix" it.
//
// # No fix
//
// Upstream marks the fixer `pending` and ships none, and it should not. Rewriting `const x =
// require('y')` into an import moves a statement to the top of the file, changes when the module is
// evaluated, and has no correct answer when the call sits inside a function or a conditional. That
// is a change of meaning at a site the rule does not own, so it is not even a suggestion here.
var NoRequireImports = rule.Rule{
	Name: "@typescript-eslint/no-require-imports",

	// See the doc above: telling the CommonJS `require` from a local binding of that name is name
	// resolution, and the four-way probe that established it is recorded there.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, _ := rule.OptionsAs[NoRequireImportsOptions](options)

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				// A typed rule handed no checker must decline rather than guess. Without this the
				// rule does not crash, it goes quiet: GetSymbolAtLocation on a nil checker returns
				// nil, which this rule reads as "not locally bound" and would report on every
				// `require` in the file regardless of shadowing.
				if ctx.TypeChecker == nil {
					return
				}

				call := node.AsCallExpression()
				// Skip parentheses to match upstream's `get_identifier_reference`; see the doc.
				callee := ast.SkipParentheses(call.Expression)
				if callee == nil || callee.Kind != ast.KindIdentifier {
					return
				}
				if callee.Text() != "require" {
					return
				}
				if isLocallyBound(ctx, callee) {
					return
				}

				// The emptiness test is a short-circuit rather than a decision, and a mutation sweep
				// scores it as a survivor for that reason. No input distinguishes the two versions:
				// matchesAnyAllowPattern loops over the slice and returns false when it is empty,
				// so the only thing the test changes is whether requiredPathForAllow slices the
				// source on a file with no `allow` configured, which is the common case. Confirmed
				// by the inverse mutation: making the whole check unreachable fails twelve lines,
				// while relaxing this test to `>= 0` fails none.
				if len(settings.Allow) > 0 && call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
					if path, isStatic := requiredPathForAllow(ctx, call.Arguments.Nodes[0]); isStatic {
						if matchesAnyAllowPattern(settings.Allow, path) {
							return
						}
					}
				}

				ctx.ReportNode(node, messageNoRequireImports)
			},

			ast.KindExternalModuleReference: func(node *ast.Node) {
				// Present for symmetry with the call arm and because the registry guard reads this
				// file textually; this arm asks the checker nothing, since `require` here is not an
				// identifier that anything could shadow.
				if ctx.TypeChecker == nil {
					return
				}

				// Upstream reports the whole declaration rather than this node, and the two spans
				// differ by `import name = ` and the trailing semicolon. Pinned against the
				// snapshot: `import lib8 = require('lib8');` underlines all 30 characters from
				// column 1. `@typescript-eslint` reports the reference node instead, so this is the
				// one place the two upstreams visibly disagree and the span is the only assertion
				// that records which one was ported.
				// The parent is always the import-equals declaration, so this is the report target
				// rather than a filter. A kind test here would be unreachable: the grammar has one
				// production for this node, and a probe over ten inputs, four of them malformed
				// enough to exercise error recovery, found no other parent kind. The nil test
				// stays because it is crash protection rather than a decision.
				declaration := node.Parent
				if declaration == nil {
					return
				}

				expression := node.AsExternalModuleReference().Expression
				if expression != nil && expression.Kind == ast.KindStringLiteral {
					if len(settings.Allow) > 0 &&
						matchesAnyAllowPattern(settings.Allow, expression.AsStringLiteral().Text) {
						return
					}
				}

				// The allow check runs first, matching upstream's order: an allowed path is silent
				// whether or not `allowAsImport` is set, and upstream's own corpus pins that with
				// five allow-permitted import-equals cases carrying no `allowAsImport`.
				if settings.AllowAsImport {
					return
				}

				ctx.ReportNode(declaration, messageNoRequireImports)
			},
		}
	},
}

// isLocallyBound reports whether an identifier spelled `require` resolves to a binding written in
// source rather than to the CommonJS global.
//
// The question is whether any declaration actually introduces a binding. A `let`, `const`, `var`,
// parameter or function somebody wrote does; an ambient declaration does not, whether it sits in a
// declaration file or inside a `declare global` block in source; and in a program with no
// `@types/node` the name resolves to nothing at all. An unresolved name is therefore treated as the
// global rather than as a local, which is the direction that keeps the rule working in a tree that
// does not carry those types.
//
// # Why the ambient flag is the whole test, and not "which file declares it"
//
// The first version of this asked `IsDeclarationFile` on the declaration's own source file, which
// is what `resolvesToAGlobal` does. That is wrong in one measurable place and a mutation sweep
// found it: `declare global { function require(...) }` puts a declaration in a SOURCE file without
// introducing a binding, and upstream reports the call beside it while staying silent beside a
// plain `function require(...)`. Both were pinned on the release binary.
//
// With the ambient test in place the file test became unreachable rather than merely redundant.
// Probed four ways, every declaration reachable inside a `.d.ts` carries the ambient flag, because
// TypeScript marks the whole declaration file ambient, so the `continue` above declines every input
// a file test could have seen. It was removed rather than left as dead reassurance.
//
// # Every declaration, not the first
//
// The symbol genuinely carries more than one, and the orderings run both ways. Two declaration
// files declaring `require` give two ambient declarations. An ambient overload written above a real
// implementation in one source file gives `[ambient, real]`, putting the only genuine binding at
// index 1, and upstream is silent on it. An index-zero read reports that input and is wrong; both
// orderings are pinned by fixtures.
func isLocallyBound(ctx rule.Context, identifier *ast.Node) bool {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		// An ambient declaration declares a name that already exists rather than introducing a
		// binding, so it is not a shadow however it is spelled. This matters because `declare
		// global { function require(...) }` sits in a source file and would otherwise read as one:
		// measured on the release binary, upstream REPORTS a `require('lib')` beside such a block
		// and is silent beside a plain `function require(...)`, and the ambient flag is what
		// separates them. Every genuine shadow probed here carries ambient=false and both halves of
		// the declare-global pair carry ambient=true.
		if declaration.Flags&ast.NodeFlagsAmbient != 0 {
			continue
		}
		return true
	}
	return false
}

// requiredPathForAllow returns the path text an `allow` pattern is tested against, and whether the
// argument is a literal at all.
//
// The two literal kinds are read differently on purpose and that is upstream's behavior rather than
// an oversight: a string is compared on its cooked value and a template on its raw source text,
// which diverge the moment either carries an escape. See the rule doc for the pinned measurement.
//
// A template carrying a substitution compares only its FIRST QUASI, which is upstream's behavior
// and reads as a defect until it is measured. oxc takes `template_literal.quasis.first()` with no
// check that the template is static, so `require(`./a${b}c`)` under `allow: ["^\\./a"]` compares
// `./a` and is permitted even though the assembled path is unknown. The intuitive reading is that a
// dynamic path should never be permitted by a prefix, and this port originally implemented that
// intuition; the release binary is silent on that exact input, so the intuition was wrong and this
// reproduces upstream instead. Recorded at the line because the next reader will have the same
// instinct.
func requiredPathForAllow(ctx rule.Context, argument *ast.Node) (string, bool) {
	if argument == nil {
		return "", false
	}
	sourceText := ctx.SourceFile.Text()
	switch argument.Kind {
	case ast.KindStringLiteral:
		return argument.AsStringLiteral().Text, true
	case ast.KindNoSubstitutionTemplateLiteral:
		// The raw source, with the backticks stripped. Slicing the file rather than reading `.Text`
		// is the whole point; `.Text` is cooked and upstream compares raw.
		literalRange := rule.TokenRange(ctx.SourceFile, argument)
		raw := sourceText[literalRange.Pos():literalRange.End()]
		if len(raw) < 2 {
			return "", false
		}
		return raw[1 : len(raw)-1], true
	case ast.KindTemplateExpression:
		// The head is the first quasi. Its raw text runs from the opening backtick to the `${` that
		// begins the first substitution, so three delimiter characters come off rather than two.
		head := argument.AsTemplateExpression().Head
		if head == nil {
			return "", false
		}
		headRange := rule.TokenRange(ctx.SourceFile, head)
		raw := sourceText[headRange.Pos():headRange.End()]
		if len(raw) < 3 {
			return "", false
		}
		return raw[1 : len(raw)-2], true
	default:
		return "", false
	}
}

// matchesAnyAllowPattern reports whether a path matches any configured allow pattern.
func matchesAnyAllowPattern(allow []*regexp.Regexp, path string) bool {
	for _, pattern := range allow {
		if pattern.MatchString(path) {
			return true
		}
	}
	return false
}

// DecodeNoRequireImportsOptions compiles the `allow` strings once, at configuration time.
//
// Upstream compiles each pattern with JavaScript's `u` flag. Go's `regexp` is RE2 and always treats
// the pattern as UTF-8, so the Unicode half of that flag is already the behavior here. What RE2 does
// not have is backtracking, so a pattern using a backreference or a lookaround fails to compile.
// Such a pattern is dropped rather than failing the run, because a linter that refuses to start over
// one malformed entry in a configuration file reports a clean tree, which is the failure this tool
// exists to remove. The dropped pattern simply permits nothing, so the rule stays strict rather than
// silently going permissive.
func DecodeNoRequireImportsOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[noRequireImportsRawOptions]()(raw)
	if err != nil {
		return NoRequireImportsOptions{}, err
	}

	wire, _ := decoded.(noRequireImportsRawOptions)

	options := NoRequireImportsOptions{AllowAsImport: wire.AllowAsImport}
	for _, pattern := range wire.Allow {
		compiled, compileError := regexp.Compile(pattern)
		if compileError != nil {
			continue
		}
		options.Allow = append(options.Allow, compiled)
	}

	return options, nil
}
