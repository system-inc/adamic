package typescript

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// PreferEnumInitializers flags an enum member written without an explicit value.
//
//	valid:   enum Direction {}
//	valid:   enum Direction { Up = 1 }
//	valid:   enum Direction { Up = 'Up', Down = 'Down' }
//	invalid: enum Direction { Up }
//	invalid: enum Direction { Up = 'Up', Down }
//
// An uninitialized member takes its value from its POSITION, so inserting a member above it or
// reordering two silently renumbers everything below. That is invisible in a diff of the enum and
// catastrophic anywhere the number was persisted: a value written to a database or sent over a wire
// last week now means a different member. Writing the value down makes the enum's contract explicit
// and makes a reorder a no-op.
//
// # The repair is three SUGGESTIONS, never a fix
//
// Upstream sets `hasSuggestions` and offers three, and the reason it cannot be a fix is the whole
// point of the rule: the correct value is not recoverable from the source. The rule does not know
// whether the author wanted the positional number, the next number, or the member's own name as a
// string, and picking one unattended would bake a guess into a value that may already have been
// persisted. So all three are offered and a human chooses:
//
//	Name = <index>       the value it has today, made explicit
//	Name = <index + 1>   the value it would have if the enum were one-based
//	Name = 'Name'        a string enum, where position cannot matter at all
//
// `index` is the member's POSITION in the enum rather than the value TypeScript would compute for
// it. Those differ whenever an earlier member carries an initializer, and upstream uses the position
// in both the suggestion text and the applied edit. Measured on `enum D { A, B = 1, C }`, where `C`
// is offered `C = 2` and `C = 3` while its actual computed value is 2: the first suggestion happens
// to be right here and would not be if `B = 10`. That is upstream's behavior and it is reproduced
// rather than corrected, because a port that computed the real value would write a different edit
// than the corpus asserts and the differential harness would read it as a disagreement.
//
// # The member's own text is the interpolated name, and upstream doubles a quote
//
// The name in both messages is `sourceCode.getText(member)`, which for an uninitialized member is
// the member's whole text and therefore just its key. For a QUOTED key that text includes the
// quotes, so upstream renders them twice over:
//
//	enum D { 'quoted' }   The value of the member ''quoted'' should be explicitly defined.
//	                      Can be fixed to 'quoted' = ''quoted''
//
// Measured against the installed 8.x build rather than inferred. Both the doubled message and the
// syntactically broken third suggestion are upstream's, and both are reproduced: the message is what
// the differential compares, and the suggestion is offered rather than applied, so a human sees it
// before it can do harm. Recording it here because it reads exactly like a defect in this port.
//
// `rule.TokenRange` supplies that text, and it is the right accessor rather than a near one:
// `node.Pos()` includes leading trivia, so a member indented on its own line would render a name
// beginning at the end of the previous line and would anchor the replacement over the whitespace
// too, which is a fix hazard as much as a span one.
//
// # Where our parser and upstream's differ, and it is a SEMANTIC ERROR rather than a rule decision
//
// A computed enum key is illegal TypeScript. `@typescript-eslint/parser` refuses it outright:
//
//	enum D { ['computed'] }   Parsing error: Computed property names are not allowed in enums.
//
// The rule never runs on that input, so upstream's silence says nothing about what the rule believes
// and everything about what its parser accepts. The tell is that the finding does not merely vanish,
// every other diagnostic vanishes with it. Our parser recovers from the illegal source and hands
// back a real enum member, so this port reports it, with the member's whole computed text as the
// name. That is a divergence, it is stated rather than silent, and it can only be reached by source
// that does not compile. Establishing it took running the parser directly, since a bare zero from
// the linter looks identical to a clean verdict.
//
// # Cost
//
// One listener on `KindEnumDeclaration`, which is a rare anchor, and the body is a single pass over
// the member list. No checker, no program, no per-file state.
var PreferEnumInitializers = rule.Rule{
	Name: "@typescript-eslint/prefer-enum-initializers",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindEnumDeclaration: func(node *ast.Node) {
				// Members is never nil, so there is no guard here.
				//
				// Measured over eight parses rather than assumed, including source the parser has
				// to recover from: `enum D {}`, `enum D`, `enum`, `enum D {`, `enum D { ,, }`,
				// `const enum D`, `declare enum D;` and a well-formed enum. Every one returns a
				// non-nil list, empty where the source has no members. Error recovery synthesizes
				// the list rather than omitting it.
				//
				// Written down because a nil check here would be indistinguishable from a
				// load-bearing one and nothing would ever delete it. If a future parser change
				// makes the list optional, this loop is where it surfaces.
				for index, member := range node.AsEnumDeclaration().Members.Nodes {
					if member.AsEnumMember().Initializer != nil {
						continue
					}

					// Upstream interpolates `getText(member)`, which for a member with no
					// initializer is the key and nothing else. The same text anchors the
					// replacement, so the two cannot drift apart.
					memberRange := rule.TokenRange(ctx.SourceFile, member)
					name := ctx.SourceFile.Text()[memberRange.Pos():memberRange.End()]

					ctx.ReportNodeWithSuggestions(member, buildDefineInitializerMessage(name),
						buildEnumInitializerSuggestion(memberRange, name, strconv.Itoa(index)),
						buildEnumInitializerSuggestion(memberRange, name, strconv.Itoa(index+1)),
						buildEnumInitializerSuggestion(memberRange, name, "'"+name+"'"),
					)
				}
			},
		}
	},
}

// buildEnumInitializerSuggestion offers one of the three repairs, replacing the member's own text.
//
// The suggested value arrives already rendered, because upstream's third form quotes the name while
// the first two are bare numbers, and deciding that here would put the quoting two call sites away
// from the message that has to agree with it.
func buildEnumInitializerSuggestion(memberRange core.TextRange, name string, suggested string) rule.Suggestion {
	return rule.Suggestion{
		Message: rule.Message{
			Id:          "defineInitializerSuggestion",
			Description: "Can be fixed to " + name + " = " + suggested,
		},
		Fixes: []rule.Fix{rule.ReplaceRange(memberRange, name+" = "+suggested)},
	}
}

func buildDefineInitializerMessage(name string) rule.Message {
	return rule.Message{
		Id:          "defineInitializer",
		Description: "The value of the member '" + name + "' should be explicitly defined.",
	}
}
