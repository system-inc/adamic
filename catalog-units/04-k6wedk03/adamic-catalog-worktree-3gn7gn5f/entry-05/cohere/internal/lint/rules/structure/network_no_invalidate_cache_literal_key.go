package structure

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// The rule's messages, one handle per id, whose wording lives in
// `policy/messages/network-no-invalidate-cache-literal-key.json`.
var (
	networkNoInvalidateCacheLiteralKeyNoStringLiteralInvalidateCacheText          = policy.MessageOf("structure/network-no-invalidate-cache-literal-key", "noStringLiteralInvalidateCache")
	networkNoInvalidateCacheLiteralKeyNoTemplateLiteralInvalidateCacheText        = policy.MessageOf("structure/network-no-invalidate-cache-literal-key", "noTemplateLiteralInvalidateCache")
	networkNoInvalidateCacheLiteralKeyNoArrayWithStringLiteralInvalidateCacheText = policy.MessageOf("structure/network-no-invalidate-cache-literal-key", "noArrayWithStringLiteralInvalidateCache")
)

// messageNoStringLiteralInvalidateCache is the finding, rendered when it is reported so the text comes from the current catalog.
func messageNoStringLiteralInvalidateCache() rule.Message {
	return rule.Message{
		Id:          networkNoInvalidateCacheLiteralKeyNoStringLiteralInvalidateCacheText.Id,
		Description: networkNoInvalidateCacheLiteralKeyNoStringLiteralInvalidateCacheText.Render(nil),
	}
}

// messageNoTemplateLiteralInvalidateCache is the finding, rendered when it is reported so the text comes from the current catalog.
func messageNoTemplateLiteralInvalidateCache() rule.Message {
	return rule.Message{
		Id:          networkNoInvalidateCacheLiteralKeyNoTemplateLiteralInvalidateCacheText.Id,
		Description: networkNoInvalidateCacheLiteralKeyNoTemplateLiteralInvalidateCacheText.Render(nil),
	}
}

// messageNoArrayWithStringLiteralInvalidateCache is the finding, rendered when it is reported so the text comes from the current catalog.
func messageNoArrayWithStringLiteralInvalidateCache() rule.Message {
	return rule.Message{
		Id:          networkNoInvalidateCacheLiteralKeyNoArrayWithStringLiteralInvalidateCacheText.Id,
		Description: networkNoInvalidateCacheLiteralKeyNoArrayWithStringLiteralInvalidateCacheText.Render(nil),
	}
}

// NetworkNoInvalidateCacheLiteralKey flags a literal cache key passed to cache.invalidate.
//
//	valid:   networkService.cache.invalidate(UserCacheKey)
//	valid:   networkService.cache.invalidate([UserCacheKey, PostCacheKey])
//	invalid: networkService.cache.invalidate('users')
//	invalid: networkService.cache.invalidate(`users-${id}`)
//	invalid: networkService.cache.invalidate(['users', PostCacheKey])
//
// Three message ids rather than one, because the three shapes want different repairs and the array
// case reports per element. A caller passing an array of five keys with one literal among them
// should see the finding on that element rather than on the array.
//
// Deliberately narrower than its sibling network-no-invalidate-cache-in-on-success: that rule tests
// the receiver name to decide whether the cache belongs to a service, while this one does not. The
// difference is what each is protecting. That rule is about where an invalidation lives, so it has
// to know it is looking at the network's cache; this one is about how any cache key is written, and
// a typo'd key is a defect whatever the cache belongs to.
//
// No fix. The repair is to export a key from wherever the cache is created and import it, which
// means creating a binding in another file.
var NetworkNoInvalidateCacheLiteralKey = rule.Rule{
	Name: "structure/network-no-invalidate-cache-literal-key",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if !isCacheInvalidateCall(callExpressionCallee(call)) {
					return
				}
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}

				firstArgument := ast.SkipParentheses(call.Arguments.Nodes[0])
				if firstArgument == nil {
					return
				}

				switch firstArgument.Kind {
				case ast.KindStringLiteral:
					ctx.ReportNode(firstArgument, messageNoStringLiteralInvalidateCache())

				case ast.KindTemplateExpression, ast.KindNoSubstitutionTemplateLiteral:
					ctx.ReportNode(firstArgument, messageNoTemplateLiteralInvalidateCache())

				case ast.KindArrayLiteralExpression:
					// Reported per element, so an array of five keys with one literal among them
					// points at the literal rather than at the array.
					for _, element := range firstArgument.AsArrayLiteralExpression().Elements.Nodes {
						switch ast.SkipParentheses(element).Kind {
						case ast.KindStringLiteral:
							ctx.ReportNode(element, messageNoArrayWithStringLiteralInvalidateCache())
						case ast.KindTemplateExpression, ast.KindNoSubstitutionTemplateLiteral:
							ctx.ReportNode(element, messageNoTemplateLiteralInvalidateCache())
						}
					}
				}
			},
		}
	},
}

// callExpressionCallee returns a call's callee with grouping removed.
func callExpressionCallee(call *ast.CallExpression) *ast.Node {
	if call == nil {
		return nil
	}
	return ast.SkipParentheses(call.Expression)
}

// isCacheInvalidateCall reports a `<something>.cache.invalidate` callee, without judging the
// receiver. The receiver test lives in network-no-invalidate-cache-in-on-success, which needs to
// know whose cache it is; this shape check is the part both rules share.
func isCacheInvalidateCall(callee *ast.Node) bool {
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}

	invalidate := callee.AsPropertyAccessExpression()
	if name := invalidate.Name(); name == nil || name.Text() != "invalidate" {
		return false
	}

	cacheAccess := ast.SkipParentheses(invalidate.Expression)
	if cacheAccess == nil || cacheAccess.Kind != ast.KindPropertyAccessExpression {
		return false
	}

	name := cacheAccess.AsPropertyAccessExpression().Name()
	return name != nil && name.Text() == "cache"
}
