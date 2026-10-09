package adamic

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// noTypePredicateText is the rule's message, whose wording lives in `policy/messages/no-type-predicate.json`.
var noTypePredicateText = policy.MessageOf("adamic/no-type-predicate", "typePredicate")

/*
 * NoTypePredicate reports every written type predicate (#drbrp8c).
 *
 *     invalid: function isCat(pet: Pet): pet is Cat { return pet.kind === 'Dog'; }
 *     invalid: function assertCat(pet: Pet): asserts pet is Cat { ... }
 *     invalid: function assertPresent(value: unknown): asserts value { ... }
 *     invalid: class Node { isLeaf(): this is Leaf { ... } }
 *     valid:   pets.filter((pet) => pet.kind === 'Cat')     an inferred predicate, which the body proves
 *
 * # The hole
 *
 * tsc narrows a value on a predicate's word and never checks the body against it, so the body can say
 * anything: adamic's refusal 4 returns `pet.kind === 'Dog'` from `isCat`, and `pet.meow()` is then a
 * TypeError on Node. A predicate tsc infers (5.5 and later, from a body like `x !== undefined`) is proven
 * by the body, so it is never written and never reported.
 *
 * Every written one is reported wherever it is written: a function, a method, an overload, a declared
 * function, a function type or an interface's member, `.d.ts` included. A declaration without a body
 * still promises the narrowing to every caller, so the trust is extended there.
 *
 * Adamic may bring back a predicate whose body the compiler can verify is exactly the narrowing; until
 * then, none.
 *
 * # No fix
 *
 * Removing the predicate changes how every caller narrows, which can break their compile.
 */
var NoTypePredicate = rule.Rule{
	Name: "adamic/no-type-predicate",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindTypePredicate: func(node *ast.Node) {
				ctx.ReportNode(node, rule.Message{Id: "typePredicate", Description: noTypePredicateText.Render(nil)})
			},
		}
	},
}
