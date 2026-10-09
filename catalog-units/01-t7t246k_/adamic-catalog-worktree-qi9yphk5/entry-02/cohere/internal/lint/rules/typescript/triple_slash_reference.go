package typescript

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messageTripleSlashReference names the referenced module in its own text.
//
// `rule.Message` is `{Id, Description}` with no interpolation layer, so upstream's format verb
// becomes string concatenation at the report site rather than a rendered template. The module name
// is worth carrying: a file with three directives produces three findings whose spans are the only
// other thing separating them, and upstream's own text names it.
//
// The value is reproduced exactly as upstream computes it, quirks included. `path='foo'` renders
// with the single quotes still attached, because only the double quote is trimmed, and
// `path=""` renders an empty name. Both are measured against the release binary and both are
// upstream's output rather than ours.
func messageTripleSlashReference(referenceName string) rule.Message {
	return rule.Message{
		Id: "tripleSlashReference",
		Description: "This file pulls in `" + referenceName + "` with a triple-slash reference " +
			"directive. A directive is a global instruction to the compiler with no binding and no " +
			"local name, so nothing in the file records that the dependency is used here and no " +
			"tool can follow it the way it follows an import. Write an `import` instead, which " +
			"names what it brings in and participates in module resolution.",
	}
}

// TripleSlashReference flags a triple-slash reference directive that should be an import.
//
//	valid:   // <reference path="foo" />                     a double slash is not a directive
//	valid:   /// <reference lib="foo" />                     lib defaults to always
//	valid:   /// <reference types="foo" />                   without an import of foo
//	valid:   const a = 1;\n/// <reference path="foo" />      not at the top of the file
//	invalid: /// <reference path="foo" />
//	invalid: /// <reference types="foo" />                   with `import * as foo from 'foo'`
//
// Ported from `typescript/triple-slash-reference`, which oxc in turn ports from
// `@typescript-eslint/triple-slash-reference`.
//
// # A comment, not a node, which decides the whole shape of the port
//
// A triple-slash directive is trivia. There is no node to anchor a listener on, so the rule runs
// once per file out of a `KindSourceFile` listener rather than reacting to a kind, matching
// upstream's `run_once`. The comments come from `comments.ForFile`, the shared per-file scan, so a
// second rule asking for them in the same file pays nothing.
//
// # Position is load-bearing and it is not "anywhere in the file"
//
// A triple-slash comment is only a directive to TypeScript when it precedes the first statement,
// and upstream reproduces that: it scans comments in `0..first statement's span start`, falling
// back to the end of the program when there is no statement. So a directive written between two
// statements is not scanned at all.
//
// The subtlety is which position bounds the range. oxc's `span().start` is the statement's first
// **token**, so a directive sitting in the statement's own leading trivia is inside the range and
// reports. typescript-go's `Pos()` is where leading trivia *begins*, which would place the cutoff
// before the directive and silence it. `rule.TokenRange` is the translation, and the fixture pair
// that pins it is a directive above `const a = 1;` reporting while the same directive below it is
// silent. Both measured against the release binary.
//
// # The content test is `starts_with('/')`, and it is not a test for a line comment
//
// Upstream reads `comment.content_span()`, which strips exactly two characters from the front of
// any comment. For `/// x` that leaves `/ x`, which starts with a slash; for `// x` it leaves ` x`
// and for `/* x */` it leaves ` x `, both of which do not. So one slash test does the work of
// distinguishing a triple slash from a double one, and it looks like a test for comment kind while
// being a test on the third character.
//
// It is not equivalent to one, and the difference is reachable. A block comment written `/*/ ... */`
// has content beginning with a slash and **reports** on the release binary, as does `//// ...`.
// The corpus's block-comment passing case reads as evidence that block comments are exempt and is
// not: its content begins with a newline. Measured by changing that one character, which moves the
// verdict from silent to reporting. ESLint's implementation does gate on comment kind, so this is a
// place the two disagree and oxc is what the gate runs.
//
// # The attribute scan reads the first matching part and stops
//
// Upstream splits the text between `<reference ` and `/>` on whitespace, keeps the parts beginning
// `types=`, `path=` or `lib=`, and reads only the **first** one. So
// `/// <reference lib="foo" path="bar" />` is silent under the defaults because `lib` is read and
// `lib` defaults to always, while the same two attributes in the other order report `bar`. Both
// measured. A value containing an `=` declines entirely, because the split produces three parts and
// upstream requires exactly two.
//
// `no-default-lib` is not in the filter set and is not an option. A
// `/// <reference no-default-lib="true" />` never reports under any configuration, measured against
// the release binary rather than assumed from the directive's existence in TypeScript.
//
// # The prefer-import arm counts imports, not directives
//
// Under `types: prefer-import` a directive is recorded and reported only when a top-level import of
// the same specifier exists. The loop is over statements, so **two imports of one module produce
// two findings against one directive**, both with the same span. That is upstream's structure and
// there is no corpus case for it.
//
// The recording side is a map keyed by the reference name, so two directives naming the same module
// leave only the **last** span. Probed with a triple, which reports once against the third copy.
// This is the counting decision the brief warns hides inside a data-structure choice, and no
// message-id fixture can see it.
//
// Only a top-level `import ... from` and a top-level `import x = require(...)` count. A dynamic
// `import()`, a bare `require()` call, an `export * from`, and an import nested inside a
// `declare module` are all silent, each measured. That is why this does not reach for
// `imports.SourceVisitors`, which would widen the rule to the two call shapes upstream ignores.
var TripleSlashReference = rule.Rule{
	// No namespace prefix. The config writes `typescript/triple-slash-reference` and the parity
	// guard strips the namespace on a `/` boundary.
	Name: "@typescript-eslint/triple-slash-reference",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// Upstream's `should_run` reads `ctx.source_type().is_typescript()`. A directive in a
		// JavaScript file is silent there, so the gate is fidelity rather than an optimization.
		if !isTypeScriptSourceFile(ctx.SourceFile.FileName().AsString()) {
			return nil
		}

		// The zero value is not the default, so a failed assertion has to fall back explicitly
		// rather than ride on Go's zero. This is not defensive coding: it is the normal path for a
		// rule configured as a bare `"error"`. The config layer hands `Run` a nil `options` in that
		// case, because `rule.DecodeOptionsInto` reports an error on empty input and
		// `configuration.OptionsRegistry.Decode` translates that into nil for a rule whose options merely
		// tune it. A bare type assertion then yields three empty settings, which match no arm, and
		// the rule goes silent on every file while every fixture stays green because fixtures reach
		// the rule through the decoder. Found by the dry run rather than by the suite: 3,407 files,
		// 3,407 registrations, zero findings, and a seeded probe tree that also reported nothing.
		parsed, decoded := rule.OptionsAs[TripleSlashReferenceOptions](options)
		if !decoded {
			parsed = DefaultTripleSlashReferenceOptions()
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				reportTripleSlashReferences(ctx, parsed)
			},
		}
	},
}

// reportTripleSlashReferences is upstream's `run_once` body.
//
// It is a function rather than a closure so the two halves of the rule, the directive scan and the
// import match, read as the two passes they are.
func reportTripleSlashReferences(ctx rule.Context, options TripleSlashReferenceOptions) {
	cutoff := directiveScanCutoff(ctx)

	// Keyed by reference name, holding the directive to blame. A map rather than a list, matching
	// upstream's `FxHashMap`, so a repeated name keeps only the last position. See the doc comment.
	referencesAwaitingImport := make(map[string]comments.Comment)

	for _, comment := range comments.ForFile(ctx) {
		// `break` rather than `continue`, and the two are equivalent here rather than one being a
		// cost optimization. They differ only on an input where a comment starting before the
		// cutoff appears in the slice after one starting at or past it, and `comments.All` sorts by
		// `Range.Pos()` on its only return path (`comments.go:178`), so the slice is monotonic in
		// position and no such input exists for any source text. Recorded because a mutation
		// swapping them survives the sweep, correctly, and the next reader should find the argument
		// rather than a missing fixture. What is pinned by a fixture instead is the property the
		// early exit depends on: four directives above a statement all report and one below it does
		// not, which is what a break firing too early or never firing would each break.
		if comment.Range.Pos() >= cutoff {
			break
		}

		attributeKey, attributeValue, found := referenceAttribute(comment.Text)
		if !found {
			continue
		}

		if (attributeKey == "types" && options.Types == TripleSlashReferenceNever) ||
			(attributeKey == "path" && options.Path == TripleSlashReferenceNever) ||
			(attributeKey == "lib" && options.Lib == TripleSlashReferenceNever) {
			ctx.ReportRange(comment.Range, messageTripleSlashReference(attributeValue))
		}

		if attributeKey == "types" && options.Types == TripleSlashReferencePreferImport {
			referencesAwaitingImport[attributeValue] = comment
		}
	}

	if len(referencesAwaitingImport) == 0 {
		return
	}

	// Top-level statements only. Upstream iterates `program.body`, so an import nested inside a
	// namespace never matches, which is measured in the doc comment.
	for _, statement := range ctx.SourceFile.Statements.Nodes {
		specifier, isImport := topLevelImportSpecifier(statement)
		if !isImport {
			continue
		}
		if comment, awaiting := referencesAwaitingImport[specifier]; awaiting {
			ctx.ReportRange(comment.Range, messageTripleSlashReference(specifier))
		}
	}
}

// directiveScanCutoff is the position past which a comment is no longer a directive.
//
// Upstream is `program.body.first().map_or(program.span.end, |v| v.span().start)`. The mapping of
// `span().start` onto `rule.TokenRange(...).Pos()` rather than `Pos()` is the whole content of this
// function, and it is the difference between reporting and silence for every directive written
// above a statement. See the doc comment on the rule.
func directiveScanCutoff(ctx rule.Context) int {
	statements := ctx.SourceFile.Statements
	if statements == nil || len(statements.Nodes) == 0 {
		return ctx.SourceFile.End()
	}
	return rule.TokenRange(ctx.SourceFile, statements.Nodes[0]).Pos()
}

// topLevelImportSpecifier returns the module specifier a statement imports, when it is one of the
// two shapes upstream matches.
//
// `import ... from 'x'` and `import x = require('x')` and nothing else. An `import x = y.z` carries
// a qualified name rather than an external module reference and is declined, which is upstream's
// two silent arms written as one.
func topLevelImportSpecifier(statement *ast.Node) (string, bool) {
	switch statement.Kind {
	case ast.KindImportDeclaration:
		declaration := statement.AsImportDeclaration()
		if declaration == nil || declaration.ModuleSpecifier == nil {
			return "", false
		}
		if !ast.IsStringLiteralLike(declaration.ModuleSpecifier) {
			return "", false
		}
		return declaration.ModuleSpecifier.Text(), true

	case ast.KindImportEqualsDeclaration:
		declaration := statement.AsImportEqualsDeclaration()
		if declaration == nil || declaration.ModuleReference == nil {
			return "", false
		}
		if declaration.ModuleReference.Kind != ast.KindExternalModuleReference {
			return "", false
		}
		reference := declaration.ModuleReference.AsExternalModuleReference()
		if reference == nil || reference.Expression == nil {
			return "", false
		}
		if !ast.IsStringLiteralLike(reference.Expression) {
			return "", false
		}
		return reference.Expression.Text(), true
	}
	return "", false
}

// referenceAttribute is upstream's `get_attr_key_and_value`, reading the directive's one attribute.
//
// The parameter is the comment's full source text including its delimiters, where upstream receives
// the text with exactly two leading characters already removed. Dropping two characters here rather
// than asking the comment its kind is what reproduces the slash test described on the rule, and it
// is deliberately not spelled as "is this a line comment": the two are not the same predicate and
// the difference reports on real input.
//
// Every trimming step below is upstream's, in upstream's order, and several of them are lossy in
// ways a reader would call a defect. They are reproduced rather than improved because they decide
// what the message says, and the fixtures pin each one.
func referenceAttribute(commentText string) (string, string, bool) {
	if len(commentText) < 2 {
		return "", "", false
	}
	content := commentText[2:]

	// Upstream's `raw.starts_with('/')`. This is the third character of the comment, so it accepts
	// `///`, `////` and `/*/` alike, and rejects `//` and `/*`.
	if !strings.HasPrefix(content, "/") {
		return "", "", false
	}

	const referenceStart = "<reference "
	const referenceEnd = "/>"

	startIndex := strings.Index(content, referenceStart)
	if startIndex < 0 {
		return "", "", false
	}
	// The end is searched from the start index onward, so a `/>` written before `<reference ` does
	// not close it. Upstream indexes into the same suffix.
	endOffset := strings.Index(content[startIndex:], referenceEnd)
	if endOffset < 0 {
		return "", "", false
	}
	inner := content[startIndex+len(referenceStart) : startIndex+endOffset]

	// Go whitespace: oxc's split_whitespace and trim, Go's set, until #jjfa7qb ports typescript-eslint's regex.
	for _, part := range strings.Fields(inner) {
		if !strings.HasPrefix(part, "types=") &&
			!strings.HasPrefix(part, "path=") &&
			!strings.HasPrefix(part, "lib=") {
			continue
		}

		// Only the first matching part is read, and a part is read whether or not it yields a
		// value: a `path="a=b"` splits into three and declines the whole comment rather than
		// falling through to a later attribute. Measured silent on the release binary.
		attributeParts := strings.Split(part, "=")
		if len(attributeParts) != 2 {
			return "", "", false
		}
		// Go whitespace: oxc's split_whitespace and trim, Go's set, until #jjfa7qb ports typescript-eslint's regex.
		key := strings.Trim(strings.TrimSpace(attributeParts[0]), "\"")
		value := strings.TrimRight(strings.Trim(attributeParts[1], "\""), "/")
		return key, value, true
	}
	return "", "", false
}

// TripleSlashReferenceSetting is what to enforce for one kind of reference directive.
//
// Spelled as a string type rather than a bool pair because the surface has three values and only
// one of the three keys accepts all of them, which a bool cannot express and which the decoder has
// to be able to refuse.
type TripleSlashReferenceSetting string

const (
	// TripleSlashReferenceAlways permits the directive.
	TripleSlashReferenceAlways TripleSlashReferenceSetting = "always"

	// TripleSlashReferenceNever reports the directive outright.
	TripleSlashReferenceNever TripleSlashReferenceSetting = "never"

	// TripleSlashReferencePreferImport reports a `types` directive only when the same module is
	// also imported normally. It is accepted for `types` alone: the release binary refuses a
	// configuration naming it for `path` or `lib`, which is what settles that this is one value of
	// a three-valued key rather than a shared vocabulary.
	TripleSlashReferencePreferImport TripleSlashReferenceSetting = "prefer-import"
)

// TripleSlashReferenceOptions configures which directives report.
//
// The authoritative surface is oxc's `TripleSlashReferenceConfig`, which derives `JsonSchema` under
// `serde(rename_all = "camelCase", deny_unknown_fields)` with exactly three fields, and
// `@typescript-eslint`'s `meta.schema` agrees key for key. Our rule inventory records
// `"options": "no"` for this rule, which is wrong: three keys are accepted and each changes the
// verdict. Measured against the release binary rather than read off that column.
//
// The zero value is not the default. Upstream's defaults are `path: never`, `types: prefer-import`
// and `lib: always`, none of which is an empty string, so `DecodeTripleSlashReferenceOptions` fills
// them and the rule is never handed a bare struct through the config path. `defaultTripleSlash`
// is what both the decoder and a hand-built options value should start from.
type TripleSlashReferenceOptions struct {
	// Lib is upstream's `lib`, defaulting to always. Accepts always and never.
	Lib TripleSlashReferenceSetting

	// Path is upstream's `path`, defaulting to never. Accepts always and never.
	Path TripleSlashReferenceSetting

	// Types is upstream's `types`, defaulting to prefer-import. The only key of the three that
	// accepts prefer-import.
	Types TripleSlashReferenceSetting
}

// DefaultTripleSlashReferenceOptions is upstream's configured-nothing behavior.
//
// Exported because a fixture asserting the default has to be able to name it, and because a caller
// building options by hand that started from the Go zero value would get three empty strings, which
// match no arm and silence the rule entirely.
func DefaultTripleSlashReferenceOptions() TripleSlashReferenceOptions {
	return TripleSlashReferenceOptions{
		Lib:   TripleSlashReferenceAlways,
		Path:  TripleSlashReferenceNever,
		Types: TripleSlashReferencePreferImport,
	}
}

// tripleSlashReferenceRawOptions is the wire shape.
//
// Pointers rather than values because "the key was absent" and "the key was the empty string" have
// to stay distinguishable long enough for the default to be applied.
type tripleSlashReferenceRawOptions struct {
	Lib   *string `json:"lib"`
	Path  *string `json:"path"`
	Types *string `json:"types"`
}

// DecodeTripleSlashReferenceOptions maps the wire keys onto the settings the rule reads, applying
// upstream's per-key defaults for anything absent.
//
// A hand-written decoder rather than `rule.DecodeOptionsInto` because the defaults are per-key and
// non-zero, and because an unrecognized value has to fall back rather than silently disable the
// key. Upstream refuses such a configuration outright; there is no error channel that reaches a
// user here, so the default is used and the rule keeps its documented behavior instead of going
// quiet on a typo.
func DecodeTripleSlashReferenceOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[tripleSlashReferenceRawOptions]()(raw)
	if err != nil {
		return DefaultTripleSlashReferenceOptions(), err
	}

	wire, _ := decoded.(tripleSlashReferenceRawOptions)
	options := DefaultTripleSlashReferenceOptions()

	options.Lib = settingOrDefault(wire.Lib, options.Lib, false)
	options.Path = settingOrDefault(wire.Path, options.Path, false)
	options.Types = settingOrDefault(wire.Types, options.Types, true)

	return options, nil
}

// settingOrDefault reads one wire value, keeping the default when it is absent or not a value this
// key accepts.
//
// allowPreferImport is what encodes the asymmetry: `types` takes three values and the other two
// take two. Reproducing that matters because `path: prefer-import` would otherwise silently behave
// like a fourth state the rule never checks for, where upstream refuses the config outright.
func settingOrDefault(wire *string, fallback TripleSlashReferenceSetting, allowPreferImport bool) TripleSlashReferenceSetting {
	if wire == nil {
		return fallback
	}
	switch TripleSlashReferenceSetting(*wire) {
	case TripleSlashReferenceAlways:
		return TripleSlashReferenceAlways
	case TripleSlashReferenceNever:
		return TripleSlashReferenceNever
	case TripleSlashReferencePreferImport:
		if allowPreferImport {
			return TripleSlashReferencePreferImport
		}
	}
	return fallback
}
