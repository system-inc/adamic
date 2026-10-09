package structure

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// networkNoInvalidateCacheInOnSuccessText is the rule's message, whose wording lives in
// `policy/messages/network-no-invalidate-cache-in-on-success.json`.
var networkNoInvalidateCacheInOnSuccessText = policy.MessageOf("structure/network-no-invalidate-cache-in-on-success", "noInvalidateCacheInOnSuccess")

// messageNoInvalidateCacheInOnSuccess is the finding, rendered when it is reported so the text comes from the current catalog.
func messageNoInvalidateCacheInOnSuccess() rule.Message {
	return rule.Message{
		Id:          networkNoInvalidateCacheInOnSuccessText.Id,
		Description: networkNoInvalidateCacheInOnSuccessText.Render(nil),
	}
}

// NetworkNoInvalidateCacheInOnSuccess flags a cache invalidation inside an onSuccess handler.
//
//	valid:   useMutation({ invalidateOnSuccess: ['users'] })
//	valid:   networkService.cache.invalidate(key)                (outside onSuccess)
//	invalid: useMutation({ onSuccess: () => networkService.cache.invalidate(key) })
//	invalid: useMutation({ onSuccess() { apiService.cache.invalidate(key); } })
//
// Three things must line up before this reports: an `onSuccess` property, a `.cache.invalidate(...)`
// call somewhere inside it, and a receiver that reads as a service. The original matches the
// receiver by name, accepting `networkService` exactly or anything containing "service" or
// "Service", and this reproduces that rather than tightening it, since a rule that reported every
// `.cache.invalidate` inside an `onSuccess` would catch caches that are not the network's.
//
// The property is matched in both spellings, `onSuccess: () => {}` and the method shorthand
// `onSuccess() {}`, which reach different node kinds. Enumerated before the listener: the original's
// ESLint selector matches a Property, and the shorthand is a Property in ESTree while it is a method
// declaration here, so a port reading only the assignment form would miss half the call sites.
//
// No fix. Moving an invalidation into the option means deciding which keys belong there and whether
// anything else in the handler depended on running at that moment.
var NetworkNoInvalidateCacheInOnSuccess = rule.Rule{
	Name: "structure/network-no-invalidate-cache-in-on-success",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		reportInvalidations := func(body *ast.Node) {
			if body == nil {
				return
			}
			var visit func(*ast.Node)
			visit = func(current *ast.Node) {
				if current == nil {
					return
				}
				if current.Kind == ast.KindCallExpression && isServiceCacheInvalidate(current) {
					ctx.ReportNode(current, messageNoInvalidateCacheInOnSuccess())
				}
				current.ForEachChild(func(child *ast.Node) bool {
					visit(child)
					return false
				})
			}
			visit(body)
		}

		return rule.Listeners{
			// `onSuccess: () => { ... }` and `onSuccess: function () { ... }`.
			ast.KindPropertyAssignment: func(node *ast.Node) {
				assignment := node.AsPropertyAssignment()
				if !isNamedOnSuccess(assignment.Name()) {
					return
				}
				reportInvalidations(assignment.Initializer)
			},

			// The method shorthand `onSuccess() { ... }`, which is a Property in ESTree and a
			// method declaration here, so it needs its own listener rather than falling out of the
			// one above.
			ast.KindMethodDeclaration: func(node *ast.Node) {
				method := node.AsMethodDeclaration()
				if !isNamedOnSuccess(method.Name()) {
					return
				}
				reportInvalidations(method.Body)
			},
		}
	},
}

// isNamedOnSuccess reports a property named onSuccess, in either the identifier or string spelling.
//
// `property.Textual` rather than the wider set, and the narrowness is preserved rather than widened
// on adoption: this compares against one fixed non-numeric name, so a numeric or computed key could
// never answer `onSuccess` and accepting them would change no verdict while widening the rule.
func isNamedOnSuccess(name *ast.Node) bool {
	text, named := property.Name(name, property.Textual)
	return named && text == "onSuccess"
}

// isServiceCacheInvalidate reports a `<something>.cache.invalidate(...)` whose receiver reads as a
// service.
//
// The shape half is shared with network-no-invalidate-cache-literal-key, since both rules are about
// the same call and two matchers that almost agree would drift. What is not shared is the receiver
// test, and the difference is deliberate: this rule is about where an invalidation lives, so it has
// to know it is looking at the network's cache; that rule is about how a key is written, and a
// typo'd key is a defect whatever the cache belongs to.
//
// The name test is the original's and it is deliberately loose: `networkService` exactly, or any
// name containing "service" in either casing. Tightening it would be a different rule, and loosening
// it further would catch caches that have nothing to do with the network layer.
func isServiceCacheInvalidate(call *ast.Node) bool {
	callee := callExpressionCallee(call.AsCallExpression())
	if !isCacheInvalidateCall(callee) {
		return false
	}

	cacheAccess := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	receiver := ast.SkipParentheses(cacheAccess.AsPropertyAccessExpression().Expression)
	if receiver == nil || receiver.Kind != ast.KindIdentifier {
		return false
	}

	receiverName := receiver.Text()
	return receiverName == "networkService" || strings.Contains(strings.ToLower(receiverName), "service")
}
