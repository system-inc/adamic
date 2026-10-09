package base

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// consistencyNoHandBuiltDeclaredErrorText is the rule's message, whose wording lives in
// `policy/messages/consistency-no-hand-built-declared-error.json`.
var consistencyNoHandBuiltDeclaredErrorText = policy.MessageOf("base/consistency-no-hand-built-declared-error", "consistencyNoHandBuiltDeclaredError")

// messageNoHandBuiltDeclaredError is the rule's single finding.
//
// The wording is the original's, which names the three tiers by their real spellings rather than
// describing them, because the reader's next action is to type one of them.
func messageNoHandBuiltDeclaredError() rule.Message {
	return rule.Message{
		Id:          consistencyNoHandBuiltDeclaredErrorText.Id,
		Description: consistencyNoHandBuiltDeclaredErrorText.Render(nil),
	}
}

// ConsistencyNoHandBuiltDeclaredError flags `new BaseError(..., { identifier })`.
//
//	valid:   new BaseError('boom')
//	valid:   new BaseError('boom', { statusCode: 500 })
//	valid:   AccountModule.error('AccountNotFound')
//	invalid: new BaseError('boom', { identifier: 'AccountNotFound' })
//
// An error that names a declared failure is built by the tier that declares it. The constructor
// takes `statusCode` and `identifier` as independent options, so nothing checks that the status
// matches the one the declaration states, and the two exist separately precisely because they
// answer different statuses. The original records the case that motivated it:
// `AccountSessionContextProvider` wrote both by hand for `AccountNotFound` and `AccountNotActive`
// one line apart, where signing in again cannot fix a suspended account, so a client retrying on a
// 401 loops forever. They happened to be right, and nothing was positioned to notice if they had
// not been.
//
// It also escapes the tier check: `Tier.error()` narrows to that tier's own declarations, which is
// what keeps a module's failure from being raised as the framework's, while a hand-built error
// takes any identifier the composed type admits.
//
// # The key is `identifier`, not the constructor
//
// A `BaseError` carrying no identifier is a different thing. `TaskCanceledError`,
// `GraphQlValidationError`, and the normalizing constructors inside `BaseError` itself wrap
// something that was never a declared failure, so the rule keys on the option rather than on the
// class.
//
// # Two files are exempt by name
//
// `BaseError.ts` normalizes one arriving from the wire and `CreateBaseErrors.ts` builds one from a
// declaration, so constructing these is what those two files are for. The original keys the
// exemption on the filename rather than an inline disable, so it is stated once where it is true
// and cannot travel to a call site by someone pasting a comment. That reasoning is the original's
// and is reproduced rather than re-decided.
//
// # No repair
//
// Reaching the right tier means knowing which one declares the identifier, and choosing wrong is
// the failure this rule exists to prevent. The original ships no fixer for that reason and neither
// does this.
//
// # Cost
//
// One kind test per `new` expression, and the checker is never consulted: the judgment is entirely
// syntactic. The filename exemption is two suffix comparisons per file.
var ConsistencyNoHandBuiltDeclaredError = rule.Rule{
	Name: "base/consistency-no-hand-built-declared-error",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}
		if isNoHandBuiltDeclaredErrorExemptFile(ctx.SourceFile.FileName().AsString()) {
			return nil
		}

		return rule.Listeners{
			ast.KindNewExpression: func(node *ast.Node) {
				newExpression := node.AsNewExpression()

				// A bare `BaseError` identifier, matching the original's `callee.type ===
				// 'Identifier' && callee.name === 'BaseError'`. A namespaced `Errors.BaseError` is
				// deliberately not matched, because the original does not match it either.
				callee := newExpression.Expression
				if callee == nil || callee.Kind != ast.KindIdentifier || callee.Text() != "BaseError" {
					return
				}

				// The options object is the second argument, when there is one.
				if newExpression.Arguments == nil || len(newExpression.Arguments.Nodes) < 2 {
					return
				}
				options := newExpression.Arguments.Nodes[1]
				if options.Kind != ast.KindObjectLiteralExpression {
					return
				}

				if !objectLiteralNamesIdentifier(options) {
					return
				}

				// The finding points at the whole `new` expression, which is the original's
				// `node`, rather than at the option that triggered it.
				ctx.ReportNode(node, messageNoHandBuiltDeclaredError())
			},
		}
	},
}

// isNoHandBuiltDeclaredErrorExemptFile reproduces the original's two filename exemptions.
//
// The original compares `context.filename` with `endsWith`, and a normalized path is compared the
// same way here. The suffix includes the separator so a file named `MyBaseError.ts` is not exempt,
// which the original's `'/BaseError.ts'` also guarantees.
func isNoHandBuiltDeclaredErrorExemptFile(fileName string) bool {
	return strings.HasSuffix(fileName, "/BaseError.ts") ||
		strings.HasSuffix(fileName, "/CreateBaseErrors.ts")
}

// objectLiteralNamesIdentifier answers whether an options object has an `identifier` property.
//
// The original tests `property.type === 'Property' && property.key.type === 'Identifier' &&
// property.key.name === 'identifier'`, so three shapes it does NOT count are reproduced here: a
// spread carrying the option, a computed key, and a string-literal key. Each is a place the rule
// knowingly misses, and widening any of them would report inputs the original is silent on.
func objectLiteralNamesIdentifier(options *ast.Node) bool {
	properties := options.AsObjectLiteralExpression().Properties
	if properties == nil {
		return false
	}
	for _, property := range properties.Nodes {
		// A shorthand `{ identifier }` is a different node kind here and IS counted, because in
		// estree it is a `Property` whose key is an Identifier named `identifier`, exactly what the
		// original tests. Our parser splits the two spellings across two kinds.
		switch property.Kind {
		case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
		default:
			continue
		}
		name := property.Name()
		if name != nil && name.Kind == ast.KindIdentifier && name.Text() == "identifier" {
			return true
		}
	}
	return false
}
