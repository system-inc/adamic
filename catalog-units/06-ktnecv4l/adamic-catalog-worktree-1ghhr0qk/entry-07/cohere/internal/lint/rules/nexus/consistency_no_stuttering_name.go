package nexus

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// defaultGenericNames are the words that count as saying nothing when they stutter.
//
// Kept short on purpose. These already read as placeholders everywhere, and the list matches the
// suffixes the hook-result naming rule bans, so the two agree on what counts as abdication.
var defaultGenericNames = []string{
	"outcome", "result", "data", "value", "output", "response", "state", "item", "thing",
}

// ConsistencyNoStutteringNameOptions replaces the set of names that count as generic.
type ConsistencyNoStutteringNameOptions struct {
	GenericNames []string `json:"genericNames"`
}

// messageStutteringName names the word that stuttered, because the repair is to pick a different
// one and a message that will not say which word is the problem is asking the reader to guess.
// consistencyNoStutteringNameText is the rule's message, whose wording lives in
// `policy/messages/consistency-no-stuttering-name.json`.
var consistencyNoStutteringNameText = policy.MessageOf("nexus/consistency-no-stuttering-name", "stutteringName")

func messageStutteringName(name string) rule.Message {
	return rule.Message{
		Id:          consistencyNoStutteringNameText.Id,
		Description: consistencyNoStutteringNameText.Render(map[string]string{"name": name}),
	}
}

// ConsistencyNoStutteringName rejects a placeholder name that stutters against its own field.
//
//	valid:   response.json()
//	valid:   parsed.value
//	invalid: outcome.outcome
//	invalid: result.result
//
// Why the stutter specifically, and not every generic name: a name is read where it is used, not
// where it is declared, so the question is never "is this word generic" but "does this word still
// carry anything by the time someone reads it." Most generic names are fine, because the value dies
// three lines later. `const response = await fetch(url); return response.json()` needs no better
// word; response is exactly what it is.
//
// The stutter is the case where the name provably carries nothing, and it is self-evident from the
// text alone. `outcome.outcome === 'Unreadable'` reads as a value whose own name had nothing to say,
// so it borrowed the field's word and said it twice.
//
// This started as a ban on generic names at the declaration and measured 1,355 hits across the repo,
// of which roughly forty were real. The rest were HTTP responses named response and process results
// named result, living four lines and correct as written. The stutter is the part of the idea that
// survived contact with the codebase, and that ratio is the reason this rule is narrow rather than
// thorough.
//
// No fix. The right noun is the judgment this rule exists to demand, and a rule cannot pick it.
var ConsistencyNoStutteringName = rule.Rule{
	Name: "nexus/consistency-no-stuttering-name",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		genericNames := map[string]bool{}
		names := defaultGenericNames
		if settings, hasSettings := rule.OptionsAs[ConsistencyNoStutteringNameOptions](options); hasSettings && len(settings.GenericNames) > 0 {
			names = settings.GenericNames
		}
		for _, name := range names {
			genericNames[name] = true
		}

		return rule.Listeners{
			// Optional chaining parses as the same kind, so `outcome?.outcome` is caught here too.
			// An element access like `outcome['outcome']` is deliberately not: a computed key is
			// usually a variable, and only the written-out form is provably a stutter.
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				access := node.AsPropertyAccessExpression()
				if access == nil || access.Expression == nil {
					return
				}
				if access.Expression.Kind != ast.KindIdentifier {
					return
				}

				name := access.Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					return
				}
				if access.Expression.Text() != name.Text() {
					return
				}
				if !genericNames[name.Text()] {
					return
				}

				ctx.ReportNode(name, messageStutteringName(name.Text()))
			},
		}
	},
}
