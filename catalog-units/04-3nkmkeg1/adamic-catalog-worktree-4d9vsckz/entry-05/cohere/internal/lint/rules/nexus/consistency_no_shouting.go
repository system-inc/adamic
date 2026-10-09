package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// consistencyNoShoutingText is the rule's message, whose wording lives in
// `policy/messages/consistency-no-shouting.json`.
var consistencyNoShoutingText = policy.MessageOf("nexus/consistency-no-shouting", "shoutingInComment")

// messageShoutingInComment names the tokens it saw, because a finding that says only "this comment
// shouts" makes the reader rescan a paragraph to find which word tripped it.
func messageShoutingInComment(tokens []string) rule.Message {
	return rule.Message{
		Id:          consistencyNoShoutingText.Id,
		Description: consistencyNoShoutingText.Render(map[string]string{"tokens": quotedTokenList(tokens)}),
	}
}

// ConsistencyNoShoutingOptions lets a project name uppercase tokens it legitimately writes, beyond
// the shared allowlist.
type ConsistencyNoShoutingOptions struct {
	Allow []string `json:"allow"`
}

// ConsistencyNoShouting bans all-caps words in comments.
//
//	valid:   // the `NOT_FOUND` branch returns early
//	valid:   // parse the JSON body before validating it
//	invalid: // NEVER cache this
//	invalid: // TODO: this is REALLY important
//
// Why a lint rule and not a style note: comments are the register the next reader inherits. A mind
// that opens a file already shouting comes up braced and writes its own comments a little louder to
// match, so the volume ratchets file by file until emphasis means nothing because everything has it.
//
// There is a craft argument underneath the tone one. "NEVER cache this" is a writer hoping volume
// will carry a claim the sentence does not. "Caching this returns a stale session id after a fork"
// needs no capitals, because it gives the reason. Turning the volume down tends to force the reason
// out, which is the actual repair.
//
// Backticks are the escape hatch, and they are the whole discipline: the rule is not "no capitals",
// it is "tell me which kind of capitals this is." Code wears backticks and is never read as
// shouting. Everything outside them is you talking.
//
// No fix, deliberately. Lowercasing the word leaves the sentence that needed shouting still weak,
// and the repair is usually to state the reason instead, which a rule cannot write.
var ConsistencyNoShouting = rule.Rule{
	Name: "nexus/consistency-no-shouting",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		additionalAllowed := map[string]bool{}
		if settings, hasSettings := rule.OptionsAs[ConsistencyNoShoutingOptions](options); hasSettings {
			for _, token := range settings.Allow {
				additionalAllowed[token] = true
			}
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				for _, comment := range comments.ForFile(ctx) {
					tokens := shoutedTokensIn(comment.Text)

					kept := tokens[:0]
					for _, token := range tokens {
						if !additionalAllowed[token] {
							kept = append(kept, token)
						}
					}
					if len(kept) == 0 {
						continue
					}

					ctx.ReportRange(comment.Range, messageShoutingInComment(kept))
				}
			},
		}
	},
}

// quotedTokenList renders the first few shouted tokens for a reader, so the finding names what to
// look at rather than making them rescan the comment.
//
// Capped because a banner line can hold a dozen, and a message that lists all of them is harder to
// act on than one that names the first few.
func quotedTokenList(tokens []string) string {
	const maximumListed = 4

	listed := tokens
	if len(listed) > maximumListed {
		listed = listed[:maximumListed]
	}

	quoted := make([]string, 0, len(listed))
	for _, token := range listed {
		quoted = append(quoted, `"`+token+`"`)
	}
	return strings.Join(quoted, ", ")
}
