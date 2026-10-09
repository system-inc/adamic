package base

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/sourcename"
	"github.com/system-inc/cohere/policy"
)

// consistencyNoBareThrowMessage is the rule's one message, shared with its Swift twin in
// `policy/messages/consistency-no-bare-throw.json`, where the wording lives.
var consistencyNoBareThrowMessage = policy.MessageOf("base/consistency-no-bare-throw", "bareThrow")

// messageNoBareThrow names the constructor, because the message is the fix and the fix differs.
//
// The name is worth carrying rather than dropping: a file throwing an `Error` and a `TypeError`
// produces two findings under one id, and the constructor is the only thing in the text that
// separates them.
func messageNoBareThrow(constructorName string) rule.Message {
	return rule.Message{
		Id:          consistencyNoBareThrowMessage.Id,
		Description: consistencyNoBareThrowMessage.Render(map[string]string{"constructor": constructorName}),
	}
}

// consistencyNoBareThrowConstructors is the original's `bareErrorConstructors`.
//
// The seven built-ins that carry a message and no identifier. `AggregateError` is deliberately
// absent and the original says why: it is a container rather than a failure, its `errors` array
// holds the real ones, and replacing it with a single identifier would discard which of the fifty
// listeners failed. Measured against the running rule, `throw new AggregateError([], 'x')` is clean.
var consistencyNoBareThrowConstructors = map[string]bool{
	"Error":          true,
	"TypeError":      true,
	"RangeError":     true,
	"SyntaxError":    true,
	"ReferenceError": true,
	"EvalError":      true,
	"URIError":       true,
}

// consistencyNoBareThrowCapturePathFiles is the original's `capturePathFiles`.
//
// The capture path cannot report through a vocabulary, because the vocabulary is what it would be
// reporting through: `BaseError` builds the errors this rule points at, and the envelope writers run
// at the moment something is already broken. Keyed on the filename rather than an inline disable so
// the exemption is stated once, where it is true, and cannot travel to a call site by someone
// pasting a comment.
//
// The leading slash is load-bearing rather than decorative. Measured against the running rule:
// `NotBaseError.ts` REPORTS, so the suffix test is on the path separator and not on the bare name.
var consistencyNoBareThrowCapturePathFiles = []string{
	"/BaseError.ts",
	"/BaseLog.ts",
	"/CreateBaseErrors.ts",
	"/WriteUnhandledErrorEnvelope.ts",
	"/IsolateErrorHandlers.ts",
}

// consistencyNoBareThrowDeclarationTimePaths is the original's `declarationTimePaths`.
//
// Declaration time, where no caller is waiting: the schema builder walking entity metadata, the
// manifest an initialize forces before the first request, the decorator registries that fill at
// import. What they catch is a developer having declared something contradictory, and the message
// they carry is the fix. There is no status to answer and no row worth writing, because a worker
// whose declaration is wrong fails identically on every request and never serves one.
//
// The original records two corrections to this list and both are carried across, because a reader
// here would otherwise reach for the wider spelling that was already tried and reverted. The
// metadata directories are named individually rather than matched as `/metadata/`, which was the
// first spelling and was too wide: a registry holds both the guards that fire while decorators fill
// it and the lookups that read it later, and the blanket path exempted the second kind. And
// `/foundation/configuration/` was on this list and is gone, because the manifest is built lazily
// inside request handling, so everything it validates surfaces on the first request a deployed
// worker takes.
var consistencyNoBareThrowDeclarationTimePaths = []string{
	"/command-line/",
	"/foundation/orm/schema/",
	"/foundation/orm/metadata/",
	"/foundation/internal/metadata/",
	"/account/metadata/",
}

// consistencyNoBareThrowBelowVocabularyPaths is the original's `belowVocabularyPaths`.
//
// Below the vocabulary in the package boundary. `client` and `api` sit under `foundation`, and all
// of nexus sits under everything: they are the floor both repositories stand on, so they cannot
// import a tier to raise through, and reaching the framework object would invert the layering rather
// than fix a name. What they throw is a primitive refusing its own input, and the layer above
// catches and names it, which is where the caller and the status live. Nexus has two code folders,
// `source/` and `code-quality/`, and both are named, because the whole library is the floor: its
// lint tooling sits under base as much as its source does.
var consistencyNoBareThrowBelowVocabularyPaths = []string{
	"/base/source/client/",
	"/base/source/api/",
	"/nexus/source/",
	"/nexus/code-quality/",
}

// consistencyNoBareThrowTestPaths is the original's `testPaths`.
//
// A test throwing to assert on it is not a failure anybody reports, and neither is the scaffolding
// it calls. Their messages are the diagnosis, which an identifier would replace with a name nothing
// branches on.
//
// The slashes on both sides matter and are measured: `source/latest/Thing.ts` and
// `source/protest/Thing.ts` both contain the letters `test` and both REPORT, because the test is on
// a path segment rather than on a substring.
var consistencyNoBareThrowTestPaths = []string{"/test/", "/tests/", "/testing/"}

// ConsistencyNoBareThrow refuses a thrown built-in Error, which names no declared failure.
//
//	valid:   throw new AggregateError([], 'x');     a container, not a failure
//	valid:   throw error;                           rethrowing is not constructing
//	valid:   throw new Errors.Validation('x');      a qualified callee is not a built-in name
//	valid:   any throw inside /nexus/source/        below the vocabulary
//	invalid: throw new Error('x');
//	invalid: throw new TypeError('x');
//
// Ported from `base/consistency-no-bare-throw` in `api-phi-health`, at
// `libraries/base/code-quality/lint/rules/ConsistencyNoBareThrowRule.ts`. This is one of OUR rules rather than
// an upstream one, so there is no corpus to import and the source file is the specification. Every
// verdict below was measured by driving that rule through the ESLint 10.8.1 Linter API over
// thirty-nine inputs, twenty-two of which exist to prove a gate declines rather than that the rule
// never looked.
//
// # No autofix, and that is the point rather than an omission
//
// The original says it outright: choosing the identifier is the work, and choosing the tier wrong is
// the failure this rule exists to prevent. A repair would have to pick one of three tiers and invent
// a name, which is exactly the judgment being asked for. Reported without a fix, deliberately.
//
// # Four gates, all on the filename, and the near misses are the interesting part
//
// A substring test on a path is easy to write and easy to get wrong in the widening direction, so
// each gate carries a measured near miss rather than an argument. `source/latest/Thing.ts` and
// `source/protest/Thing.ts` both contain `test` and both report. `source/modules/orm/schema-tools/`
// contains neither `/foundation/orm/schema/` nor anything else on the list and reports.
// `NotBaseError.ts` reports because the capture-path test keeps its leading slash. All four measured
// against the running rule.
//
// # The span is the NEW EXPRESSION rather than the throw statement
//
// The original reports `node.argument`, so the finding covers `new Error('x')` and not the `throw`
// keyword. Measured on three constructors. That is worth stating because reporting the statement
// would read as equally correct and would put the finding one token to the left, where a reader
// looking for the constructor name would not find it underlined.
var ConsistencyNoBareThrow = rule.Rule{
	Name: "base/consistency-no-bare-throw",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// The separator is normalized because every gate below is a substring or a suffix test on a
		// path written with forward slashes, and the original ran on a filename ESLint had already
		// normalized. Without this a Windows-shaped path would decline every gate and the rule would
		// report inside the capture path.
		fileName := imports.NormalizedFileName(ctx.SourceFile)

		if consistencyNoBareThrowFileIsExempt(fileName) {
			// Returning no listeners rather than guarding at each report, matching the original's
			// early `return {}`. It also means an exempt file costs nothing beyond four substring
			// scans, which is the cheapest a rule can decline.
			return nil
		}

		return rule.Listeners{
			ast.KindThrowStatement: func(node *ast.Node) {
				reportNoBareThrow(ctx, node)
			},
		}
	},
}

// consistencyNoBareThrowFileIsExempt answers the original's four early returns, in the original's order.
//
// The order does not decide any verdict, since each gate independently exempts, and it is preserved
// because the original's comments are keyed to it: a reader comparing the two files should find the
// same four paragraphs in the same sequence.
func consistencyNoBareThrowFileIsExempt(fileName string) bool {
	for _, capturePathFile := range consistencyNoBareThrowCapturePathFiles {
		if strings.HasSuffix(fileName, capturePathFile) {
			return true
		}
	}
	for _, declarationTimePath := range consistencyNoBareThrowDeclarationTimePaths {
		if strings.Contains(fileName, declarationTimePath) {
			return true
		}
	}
	for _, belowVocabularyPath := range consistencyNoBareThrowBelowVocabularyPaths {
		if strings.Contains(fileName, belowVocabularyPath) {
			return true
		}
	}
	// `Dog.test.a`, Adamic's test file, is one too (#kwt1htp).
	if strings.HasSuffix(sourcename.TreatedAs(fileName), ".test.ts") {
		return true
	}
	for _, testPath := range consistencyNoBareThrowTestPaths {
		if strings.Contains(fileName, testPath) {
			return true
		}
	}
	return false
}

// reportNoBareThrow judges one throw statement.
func reportNoBareThrow(ctx rule.Context, node *ast.Node) {
	thrown := node.AsThrowStatement().Expression
	if thrown == nil {
		// Error recovery produces a throw with nothing thrown, from source as ordinary as a
		// half-typed `throw`. The original cannot meet this shape because its parser refuses the
		// file outright; ours recovers and hands the walk a statement with a nil expression.
		return
	}

	// Only a `new` expression whose callee is a bare identifier. The original tests
	// `thrown.type !== 'NewExpression' || thrown.callee.type !== 'Identifier'`, which declines a
	// rethrow, a call, a property access and a qualified constructor in one condition.
	if thrown.Kind != ast.KindNewExpression {
		return
	}
	callee := thrown.AsNewExpression().Expression
	if callee == nil || callee.Kind != ast.KindIdentifier {
		return
	}

	// `Text()` is read only after the kind is known to be an identifier, which is the whole reason
	// the kind test sits above it rather than beside it. `ast.Node.Text()` PANICS on a property
	// access and on a call expression, and both are shapes this rule meets: `throw someError.message`
	// and `throw makeError()` are ordinary source. The walk recovers per FILE rather than per rule,
	// so one such node would cost every rule in the package every finding in that file while the run
	// still printed a plausible summary.
	constructorName := callee.Text()
	if !consistencyNoBareThrowConstructors[constructorName] {
		return
	}

	// The new expression rather than the statement, matching the original's `node: thrown`.
	ctx.ReportNode(thrown, messageNoBareThrow(constructorName))
}
