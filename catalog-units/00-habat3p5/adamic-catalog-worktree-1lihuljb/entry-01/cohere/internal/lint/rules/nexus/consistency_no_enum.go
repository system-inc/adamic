package nexus

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// consistencyNoEnumText is the rule's message, whose wording lives in `policy/messages/consistency-no-enum.json`.
var consistencyNoEnumText = policy.MessageOf("nexus/consistency-no-enum", "noEnum")

func messageNoEnum() rule.Message {
	return rule.Message{Id: "noEnum", Description: consistencyNoEnumText.Render(nil)}
}

// ConsistencyNoEnum bans TypeScript's enum keyword, including const enum.
//
// An enum compiles to a runtime object with a reverse mapping, which does not tree-shake and is
// larger than the union it replaces. Its type narrows to an opaque branded value rather than to the
// literal string, so a value that is obviously "A" stops being usable as "A" at a boundary. The
// `as const` pattern has neither problem and matches the shape generated GraphQL already uses.
//
// No fix, deliberately. Migrating an enum means choosing the replacement name and updating every
// reference, which is a judgment the rule cannot make from one declaration.
var ConsistencyNoEnum = rule.Rule{
	Name: "nexus/consistency-no-enum",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindEnumDeclaration: func(node *ast.Node) {
				declaration := node.AsEnumDeclaration()
				if declaration == nil {
					return
				}
				// Report on the name rather than the whole declaration: an enum body can be
				// hundreds of lines, and a finding that highlights all of them says less than one
				// pointing at the word that has to change.
				target := declaration.Name()
				if target == nil {
					target = node
				}
				ctx.ReportNode(target, messageNoEnum())
			},
		}
	},
}
