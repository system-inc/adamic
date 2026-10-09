package base

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// boundaryNoGlobalContainerText is the rule's message, whose wording lives in
// `policy/messages/boundary-no-global-container.json`.
var boundaryNoGlobalContainerText = policy.MessageOf("base/boundary-no-global-container", "boundaryNoGlobalContainer")

// messageNoGlobalContainer is the finding, rendered when it is reported so the text comes from the current catalog.
func messageNoGlobalContainer() rule.Message {
	return rule.Message{Id: boundaryNoGlobalContainerText.Id, Description: boundaryNoGlobalContainerText.Render(nil)}
}

// BoundaryNoGlobalContainer bans every reference to the name `getGlobalContainer`.
//
//	valid:   const c = resolveFromInjectedContainer();
//	valid:   const s = 'getGlobalContainer';            a string is not a reference
//	invalid: const c = getGlobalContainer();
//	invalid: getGlobalContainer().resolve('thing');
//	invalid: import { getGlobalContainer } from './c';
//
// Reaching for the global container hides a dependency and defeats the injection graph, so the ban
// is on the NAME rather than on any particular use of it. The original says this outright: one
// identifier visitor catches the call, member access on its result, and the import alike, where
// separate call and member visitors would report the same usage twice.
//
// # It is a name test, not a resolution test, and that is deliberate
//
// Nothing here asks the checker what `getGlobalContainer` binds to. A locally declared function of
// that name reports, a property key of that name reports, and a shadow would report too. Measured
// against the source rule across thirteen shapes rather than assumed: it reports on a property key,
// on a member access, on a local function declaration, and on a type query, and stays silent only
// where there is no identifier at all, which is a string literal, a comment, and a longer name that
// merely starts with it.
//
// That breadth is the point. A rule that resolved the name would exempt a local shadow, and a local
// shadow of the global container is exactly as bad as the global container.
//
// # Where this project is exempted is CONFIG, not rule
//
// The source repository turns this off for its own dependency-injection folders, because that is
// where the container hierarchy is built. That exemption lives in BaseLintConfiguration.ts as a
// files-scoped override, so it is deliberately not reproduced here: baking a path into the rule
// would make the exemption invisible to whoever configures it.
//
// # One divergence, and it is a parser difference in our favour
//
// The source rule reports an unaliased import specifier TWICE, at the identical span. Its parser
// gives `import { getGlobalContainer }` two Identifier nodes in one place, the imported name and the
// local binding, and the visitor sees both. Measured: `import { getGlobalContainer } from './c';`
// alone produces two findings at columns 10 through 28, and the same import followed by a call
// produces three.
//
// That contradicts the rule's own stated reason for using a single visitor, which is to avoid
// duplicate diagnostics on one line. Our parser produces one identifier for that specifier, so this
// port reports once. The divergence is recorded rather than reproduced: emitting a second finding at
// a span already reported would mean inventing a duplicate to match a defect.
var BoundaryNoGlobalContainer = rule.Rule{
	Name: "base/boundary-no-global-container",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindIdentifier: func(node *ast.Node) {
				if node.AsIdentifier().Text != "getGlobalContainer" {
					return
				}
				ctx.ReportNode(node, messageNoGlobalContainer())
			},
		}
	},
}
