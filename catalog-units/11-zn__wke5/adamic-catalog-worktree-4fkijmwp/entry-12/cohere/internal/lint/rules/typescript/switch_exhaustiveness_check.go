package typescript

import (
	"regexp"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// SwitchExhaustivenessCheck flags a switch over a union of literal types that does not handle every
// member, and, under one option, a default case on a switch that already handles every member.
//
//	valid:   declare const day: 'a' | 'b'; switch (day) { case 'a': break; case 'b': break; }
//	valid:   declare const day: 'a' | 'b'; switch (day) { case 'a': break; default: break; }
//	invalid: declare const day: 'a' | 'b'; switch (day) { case 'a': break; }
//	invalid: declare const value: number; switch (value) { case 0: break; }   // requireDefaultForNonUnion
//
// A switch over `'a' | 'b'` that handles only `'a'` compiles and then falls through at runtime, and
// the compiler will not tell you when a third member is added to the union later. This rule is the
// thing that does tell you, which is why it is the one type-aware rule in this family whose value
// grows with the codebase rather than being constant per site.
//
// # Absorbed from tsgolint, which is the source of record for this rule
//
// Provenance: tsgolint `internal/rules/switch_exhaustiveness_check/switch_exhaustiveness_check.go`,
// vendored at commit `05b7fbc` and absorbed onto cohere's own rule interface here. It reaches for
// `TypeRecurser`, `GetConstrainedTypeAtLocation`, `UnionTypeParts`, `IntersectionTypeParts`,
// `IsTypeFlagSet`, `Some`, `Every` and `Ref` because that is what upstream reaches for, and this
// note is why a reader finds those helpers in a file that otherwise looks native.
//
// tsgolint is not re-synced, so this file is now the only copy of the algorithm rather than a
// translation layer over a vendored one. The checker logic below is byte-identical to upstream's;
// what changed is the interface it speaks: `rule.RuleContext` became `rule.Context`,
// `RuleListeners` became `Listeners`, and `rule.RuleMessage` became `rule.Message`. No predicate,
// no flag set, no defaulting block and no traversal was touched, and the options struct came across
// field for field rather than being re-declared — see the note on its pointer fields below.
//
// oxc declares this rule as `SwitchExhaustivenessCheck(tsgolint)` and carries no algorithm and no
// corpus, so the behavior oxlint exhibits IS tsgolint's: the release binary shells out to a
// `tsgolint` executable and refuses to run the rule when that binary is absent, which is what it
// does on this machine. That refusal is itself the proof, so oxlint cannot serve as ground truth
// here and tsgolint's own test file is the corpus instead.
//
// # What decides a finding
//
// The anchor is `KindSwitchStatement`, one listener, and every case runs all three checks against
// one shared `SwitchMetadata` computed once. The metadata is three facts about the switch:
//
// MISSING LITERAL BRANCHES. The discriminant's type is resolved through
// `GetConstrainedTypeAtLocation` (so a type parameter walks to its base constraint) and then
// recursed with `TypeRecurser`. A recursed member counts as missing when it is literal-like and is
// not pointer-identical to any case expression's type. Literal-like means the flags
// `Literal | Undefined | Null | UniqueESSymbol`, which is why `case undefined` and `case null` are
// real branches here and why an enum member is one too.
//
// CONTAINS A NON-LITERAL TYPE. True when SOME union member has EVERY intersection member
// non-literal-like. The nesting is the part a reader gets wrong: it is some-over-union of
// every-over-intersection, so `'foo' & { bar: 1 }` contains a non-literal type because the
// intersection is not wholly literal, while `'foo' | number` contains one because of the `number`
// arm. This is what makes a default case never superfluous on a switch that admits `string`.
//
// THE DEFAULT CASE, or nil when there is none.
//
// From those, three independent checks, and they can BOTH fire on one switch: four of upstream's
// invalid cases report twice on the same statement, once for a missing branch and once for the
// missing default. A fixture asserting one id per input would have passed while dropping half the
// findings, which is why the fixtures here assert an ordered id list rather than a set.
//
// # The undefined special case, which no amount of reading the message text would suggest
//
// `missing`, `optional` and `undefined` are three different runtime type objects that all carry
// `TypeFlagsUndefined`. So once ANY case expression's type carries that flag, EVERY undefined-ish
// member of the discriminant is treated as covered, rather than matching them pairwise. Without it,
// `switch (x)` over `string | undefined` with `case undefined` would report the optional-flavored
// undefined as still missing.
//
// # The option surface is THREE live options, and the inventory says zero
//
// A captured rule inventory carried `"options": "no"` for this rule. That is wrong, and it is the
// eighth inventory error found in this lane and the most consequential, because each of these
// options flips a whole class of input rather than tuning an edge:
//
//	allowDefaultCaseForExhaustiveSwitch  default TRUE   false turns dangerousDefaultCase ON
//	considerDefaultExhaustiveForUnions   default FALSE  true makes any default case satisfy the union
//	requireDefaultForNonUnion            default FALSE  true reports a non-union switch with no default
//
// Sixty-eight of tsgolint's own hundred and six cases set at least one of them, and the same source
// text flips between reporting and silent depending only on which is set. The entry is corrected in
// this commit. The parity guard reads only rule presence from that file and never the options
// column, so the wrong value was inert rather than enforced — worth saying, because "the guard is
// green" was not evidence the column was right.
//
// # And a FOURTH option, which tsgolint declared and never read, now implemented
//
// `defaultCaseCommentPattern` makes a comment stand in for a default clause. tsgolint declared the
// field and never read it, and all six of its `Skip: true` cases were the ones that needed it. A
// config writing it decoded clean and changed nothing, and worse, typescript-eslint applies its
// DEFAULT pattern `/^no default$/iu` with the option unset, so a codebase that adopted the
// `// no default` convention under typescript-eslint lit up here. Implemented in #6esg2nx from
// 8.71's getCommentDefaultCase, found by the registry guard that fails on a decoded option nothing
// reads; see defaultCaseComment for the shapes, each replayed through the installed 8.71.0.
//
// The comment is the default case for all three checks: it satisfies `requireDefaultForNonUnion`
// and `considerDefaultExhaustiveForUnions`, and on an exhaustive switch under
// `allowDefaultCaseForExhaustiveSwitch: false` it is itself reported as `dangerousDefaultCase`, on
// the comment. A real default clause wins over it, as upstream's `defaultCase ?? comment` does.
//
// # There are no fixes and no suggestions, and the brief expected some
//
// This rule is the one in the family that would ship a repair which WRITES NEW CODE — adding a
// missing `case` clause, not deleting or rewriting an existing one. tsgolint ships neither. Every
// one of the fifty-two expected suggestion outputs in upstream's corpus is COMMENTED OUT under
// `TODO(port): add support for suggestions`, and `checkSwitchNoUnionDefaultCase` carries its own
// `// TODO(port): missing suggestion` at the site, which is carried across below.
//
// So `addMissingCases` is not a message id this rule can emit, and there is nothing for `rule_testing`
// to apply and nothing needing a hand-rolled suggestion applier. That is a real behavioral gap
// against `@typescript-eslint`, and it is upstream's gap, reproduced rather than improved on.
//
// Upstream declares a `buildAddMissingCasesMessage` builder that nothing calls. It is the one thing
// this absorption did NOT carry across, because a message builder no call site reaches is dead
// weight rather than behavior: keeping it would move zero findings and add a function a reader has
// to chase to discover it is unreachable. If a later change implements the suggestion, the builder
// comes back with the call site that needs it.
//
// # The message names the missing branches, which tsgolint's did not
//
// tsgolint's `buildSwitchIsNotExhaustiveMessage` took the missing-branch list and threw it away,
// rendering the bare sentence with the interpolating half of the format string commented out. That
// was reproduced while the differential harness compared against oxlint, which runs tsgolint. Parity
// is now measured against `@typescript-eslint` (`s l --linter both`), which renders
// `Switch is not exhaustive. Cases not matched: "b" | "c"`, and the list is most of what makes the
// finding actionable, so it is restored (#21011kd). See renderMissingBranches.
//
// # Where the two references disagree, counted rather than assumed
//
// Both register exactly ONE listener, on the switch statement, so there is no missing-arm drift of
// the kind `await-thenable` had — counted on both sides rather than assumed. Both carry the same
// three message ids and the same four schema keys. The disagreements are all in what is
// IMPLEMENTED behind them, and all three were measured by driving `@typescript-eslint`'s Linter API
// over a real program rather than by reading its source:
//
//	MESSAGE TEXT.  On `'a' | 'b' | 'c'` covering only 'a', it renders
//	               `Switch is not exhaustive. Cases not matched: "b" | "c"`. We now render the same
//	               text (#21011kd); tsgolint's builder discarded the list.
//
//	SUGGESTIONS.   The same input carries one SUGGESTION and no fix there. We carry neither.
//
//	COMMENT PATTERN. `// no default` under `requireDefaultForNonUnion` produces an EMPTY diagnostic
//	               list there, because its default pattern is `/^no default$/i`. We reported until
//	               #6esg2nx and are silent now.
//
// The comment pattern was the sharpest of the three for a user, and it is closed (#6esg2nx); the
// suggestion is still reproduced from tsgolint rather than repaired. It is a gap rather than a
// contradiction: no input makes the two report a DIFFERENT id.
//
// # The checker, and the nil guard that now lives here
//
// The listener reads `ctx.TypeChecker` unconditionally, so this needs the checker and its fixtures
// use `RunTyped`. While this rule was adapted, the standing `if ctx.TypeChecker == nil { return }`
// could not be written, because the listener was upstream's and editing it would have turned a
// re-sync into a merge. `upstream.Adapt` set `NeedsTypeChecker` on every rule it wrapped, so the
// nil case was unreachable through it. Absorbing the rule removes that constraint AND makes the nil
// case reachable, so the guard is now written where the advice always wanted it, and it is the one
// addition to the body.
//
// That direction matters here more than usual. Under a checker-less Context this rule does not
// panic — the shim's type queries return nil rather than crashing — it goes SILENT, and silence
// makes every clean fixture pass having proven nothing, which is forty-nine of this rule's
// ninety-seven cases. A test in this package pins the declaration and the guard together so a later
// revert fails loudly instead of going vacuously green.
//
// # Cost
//
// `KindSwitchStatement` is an uncommon anchor, but unlike the rest of this family the listener has
// no cheap syntactic exit: it consults the checker for the discriminant and for every case
// expression on every switch it sees. A switch with N cases costs N+1 type resolutions plus a
// recursion over the discriminant's union. That is the honest price of the only question this rule
// can ask.
var SwitchExhaustivenessCheck = rule.Rule{
	Name: "@typescript-eslint/switch-exhaustiveness-check",

	// The listener consults the checker for the discriminant and for every case expression on every
	// switch it sees, with no cheap syntactic exit, so the checker is required.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		opts, ok := rule.OptionsAs[SwitchExhaustivenessCheckOptions](options)
		if !ok {
			opts = SwitchExhaustivenessCheckOptions{}
		}
		if opts.AllowDefaultCaseForExhaustiveSwitch == nil {
			opts.AllowDefaultCaseForExhaustiveSwitch = type_checking.Ref(true)
		}
		if opts.ConsiderDefaultExhaustiveForUnions == nil {
			opts.ConsiderDefaultExhaustiveForUnions = type_checking.Ref(false)
		}
		if opts.RequireDefaultForNonUnion == nil {
			opts.RequireDefaultForNonUnion = type_checking.Ref(false)
		}
		// Upstream compiles the pattern with `new RegExp(pattern, 'u')`, so an invalid one throws
		// when the rule is created. Here it matches no comment, which leaves the real default clause
		// the only default, the reading that reports rather than the one that silences.
		commentPattern := defaultCaseCommentPattern
		if opts.DefaultCaseCommentPattern != nil {
			commentPattern, _ = regexp.Compile(*opts.DefaultCaseCommentPattern)
		}

		isLiteralLikeType := func(t *checker.Type) bool {
			return type_checking.IsTypeFlagSet(
				t,
				checker.TypeFlagsLiteral|checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsUniqueESSymbol,
			)
		}

		/**
		 * For example:
		 *
		 * - `"foo" | "bar"` is a type with all literal types.
		 * - `"foo" | number` is a type that contains non-literal types.
		 * - `"foo" & { bar: 1 }` is a type that contains non-literal types.
		 *
		 * Default cases are never superfluous in switches with non-literal types.
		 */
		doesTypeContainNonLiteralType := func(t *checker.Type) bool {
			return type_checking.Some(
				type_checking.UnionTypeParts(t),
				func(t *checker.Type) bool {
					return type_checking.Every(
						type_checking.IntersectionTypeParts(t),
						func(t *checker.Type) bool {
							return !isLiteralLikeType(t)
						},
					)
				},
			)
		}

		getSwitchMetadata := func(node *ast.SwitchStatement) *switchMetadata {
			cases := node.CaseBlock.AsCaseBlock().Clauses.Nodes
			defaultCaseIndex := slices.IndexFunc(cases, func(clause *ast.Node) bool {
				return clause.Kind == ast.KindDefaultClause
			})
			var defaultCase *ast.CaseOrDefaultClause
			if defaultCaseIndex > -1 {
				defaultCase = cases[defaultCaseIndex].AsCaseOrDefaultClause()
			}

			discriminantType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, node.Expression)

			caseTypes := make([]*checker.Type, 0, len(cases))
			for _, c := range cases {
				if c.Kind == ast.KindDefaultClause {
					continue
				}

				caseTypes = append(caseTypes, type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, c.AsCaseOrDefaultClause().Expression))
			}

			containsNonLiteralType := doesTypeContainNonLiteralType(discriminantType)

			missingLiteralBranchTypes := make([]*checker.Type, 0, 10)
			type_checking.TypeRecurser(discriminantType, func(t *checker.Type) bool {
				if slices.Contains(caseTypes, t) || !isLiteralLikeType(t) {
					return false
				}

				// "missing", "optional" and "undefined" types are different runtime objects,
				// but all of them have TypeFlags.Undefined type flag
				if slices.ContainsFunc(caseTypes, func(t *checker.Type) bool {
					return type_checking.IsTypeFlagSet(t, checker.TypeFlagsUndefined)
				}) && type_checking.IsTypeFlagSet(t, checker.TypeFlagsUndefined) {
					return false
				}

				missingLiteralBranchTypes = append(missingLiteralBranchTypes, t)

				return false
			})

			metadata := &switchMetadata{
				ContainsNonLiteralType:    containsNonLiteralType,
				DefaultCase:               defaultCase,
				MissingLiteralBranchTypes: missingLiteralBranchTypes,
			}
			if defaultCase == nil {
				metadata.DefaultCaseComment = defaultCaseComment(ctx, node, commentPattern)
			}
			return metadata
		}

		checkSwitchExhaustive := func(node *ast.SwitchStatement, metadata *switchMetadata) {
			// If considerDefaultExhaustiveForUnions is enabled, the presence of a default case
			// always makes the switch exhaustive.
			if *opts.ConsiderDefaultExhaustiveForUnions && metadata.hasDefaultCase() {
				return
			}

			if len(metadata.MissingLiteralBranchTypes) > 0 {
				ctx.ReportNode(node.Expression, buildSwitchIsNotExhaustiveMessage(
					renderMissingBranches(ctx.TypeChecker, metadata.MissingLiteralBranchTypes)))
			}
		}

		checkSwitchUnnecessaryDefaultCase := func(metadata *switchMetadata) {
			if *opts.AllowDefaultCaseForExhaustiveSwitch {
				return
			}

			if len(metadata.MissingLiteralBranchTypes) == 0 &&
				metadata.hasDefaultCase() &&
				!metadata.ContainsNonLiteralType {
				if metadata.DefaultCase != nil {
					ctx.ReportNode(&metadata.DefaultCase.Node, buildDangerousDefaultCaseMessage())
				} else {
					ctx.ReportRange(metadata.DefaultCaseComment.Range, buildDangerousDefaultCaseMessage())
				}
			}
		}
		checkSwitchNoUnionDefaultCase := func(node *ast.SwitchStatement, metadata *switchMetadata) {
			if !*opts.RequireDefaultForNonUnion {
				return
			}

			if metadata.ContainsNonLiteralType && !metadata.hasDefaultCase() {
				ctx.ReportNode(node.Expression, buildSwitchIsNotExhaustiveMessage("default"))
				// TODO(port): missing suggestion
			}
		}

		return rule.Listeners{
			ast.KindSwitchStatement: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				stmt := node.AsSwitchStatement()

				metadata := getSwitchMetadata(stmt)
				checkSwitchExhaustive(stmt, metadata)
				checkSwitchUnnecessaryDefaultCase(metadata)
				checkSwitchNoUnionDefaultCase(stmt, metadata)
			},
		}
	},
}

// switchMetadata is the three facts about a switch that all three checks read, computed once per
// statement.
//
// Upstream names this `SwitchMetadata` and exports it. Nothing outside the rule reads it, so it is
// unexported here, and that forces the second half of the same rename: upstream's three check
// closures each take a PARAMETER also named `switchMetadata`, shadowing the type inside their own
// bodies, which compiles only while the type's name differs in case. Those parameters are `metadata`
// here. Both renames are naming rather than behavior — no expression changed, only what it is
// spelled — and they are the only edits in this file beyond the interface renames and the guard.
type switchMetadata struct {
	ContainsNonLiteralType bool
	// nil if there is no default case
	DefaultCase *ast.CaseOrDefaultClause
	// DefaultCaseComment is the comment standing in for a default clause, when there is no clause and
	// the last comment after the last case matches the pattern; nil otherwise.
	DefaultCaseComment        *comments.Comment
	MissingLiteralBranchTypes []*checker.Type
	// TODO: add support for fixed (symbolname is used only for fixes)
	// SymbolName string
}

// hasDefaultCase is upstream's `defaultCase != null`, which a matching comment satisfies too.
func (metadata *switchMetadata) hasDefaultCase() bool {
	return metadata.DefaultCase != nil || metadata.DefaultCaseComment != nil
}

// defaultCaseCommentPattern is typescript-eslint's DEFAULT_COMMENT_PATTERN, `/^no default$/iu`.
var defaultCaseCommentPattern = regexp.MustCompile(`(?i)^no default$`)

// defaultCaseComment is upstream's getCommentDefaultCase: the LAST comment between the end of the last
// case clause and the switch's closing brace, when its text, delimiters stripped and trimmed, matches
// the pattern. A nil pattern matches nothing.
//
// Measured on the installed 8.71.0 under `requireDefaultForNonUnion`, each of these follows:
//
//	case 0: break;  // no default            a default    so is `/* no default */`, `// No Default`,
//	case 0: break; // no default              a default    `//no default` and a block over three lines
//	case 0: { break; // no default }          not          inside the last case's block, not after it
//	// no default                             not          before the first case, or with no case at all
//	  // something else  (after it)           not          only the last comment is asked
//	/** no default */                         not          its trimmed text starts with `*`
//	// no default here                        not          the default pattern is anchored at both ends
//
// An empty pattern matches every comment, as `new RegExp(”)` does upstream.
func defaultCaseComment(ctx rule.Context, statement *ast.SwitchStatement, pattern *regexp.Regexp) *comments.Comment {
	if pattern == nil {
		return nil
	}
	caseBlock := statement.CaseBlock
	clauses := caseBlock.AsCaseBlock().Clauses.Nodes
	if len(clauses) == 0 {
		return nil
	}
	after, closingBrace := clauses[len(clauses)-1].End(), caseBlock.End()-1
	var last *comments.Comment
	all := comments.ForFile(ctx)
	for index := range all {
		if all[index].Range.Pos() >= after && all[index].Range.End() <= closingBrace {
			last = &all[index]
		}
	}
	if last == nil {
		return nil
	}
	value := last.Text[2:]
	if last.IsBlock {
		value = strings.TrimSuffix(value, "*/")
	}
	if !pattern.MatchString(text.TrimWhitespace(value)) {
		return nil
	}
	return last
}

// SwitchExhaustivenessCheckOptions is the configuration surface, named once so the registration and
// a config reader do not name two types.
//
// It is upstream's struct field for field rather than a translation of it, which is what makes
// `rule.DecodeOptionsInto` sufficient here where `no-this-alias` needed a hand-written decoder:
// nothing is inverted and nothing is renamed, so each tag is upstream's key exactly. The fields were
// untagged until #4a4yse4 and bound through `encoding/json`'s case-insensitive matching, which also
// let any casing of a key through; the tags declare the spelling and rule.UnmarshalOptions holds a
// config to it.
//
// The pointer fields are load-bearing and must not be flattened to plain bools. Two of the three
// live options default to a value that is not the zero value — `allowDefaultCaseForExhaustiveSwitch`
// defaults to TRUE — so a `bool` field could not tell "the user wrote false" from "the user wrote
// nothing", and the defaulting block at the top of `Run` reads exactly that distinction.
type SwitchExhaustivenessCheckOptions struct {
	AllowDefaultCaseForExhaustiveSwitch *bool   `json:"allowDefaultCaseForExhaustiveSwitch"`
	ConsiderDefaultExhaustiveForUnions  *bool   `json:"considerDefaultExhaustiveForUnions"`
	DefaultCaseCommentPattern           *string `json:"defaultCaseCommentPattern"`
	RequireDefaultForNonUnion           *bool   `json:"requireDefaultForNonUnion"`
}

// buildDangerousDefaultCaseMessage is upstream's message, text unchanged.
func buildDangerousDefaultCaseMessage() rule.Message {
	return rule.Message{
		Id:          "dangerousDefaultCase",
		Description: "The switch statement is exhaustive, so the default case is unnecessary.",
	}
}

// buildSwitchIsNotExhaustiveMessage is typescript-eslint's message, naming the branches still missing.
func buildSwitchIsNotExhaustiveMessage(missingBranches string) rule.Message {
	return rule.Message{
		Id:          "switchIsNotExhaustive",
		Description: "Switch is not exhaustive. Cases not matched: " + missingBranches,
	}
}

// renderMissingBranches is typescript-eslint's `missingBranches`: each missing member as the checker
// prints it, joined with ` | ` in the order the rule found them. A unique symbol prints as
// `typeof <name>`, the type a case would have to name, since the checker's own rendering of a unique
// symbol type is not something a reader can write in a case clause. Measured against the installed
// 8.67.0 build on seven kinds: `"b" | "c"`, `2 | 3`, `false`, `Color.Green | Color.Blue`,
// `undefined | null`, `typeof second`, and an optional property's `undefined | "q"`.
func renderMissingBranches(typeChecker *checker.Checker, missingTypes []*checker.Type) string {
	rendered := make([]string, 0, len(missingTypes))
	for _, missingType := range missingTypes {
		if symbol := missingType.Symbol(); symbol != nil &&
			checker.Type_flags(missingType)&checker.TypeFlagsESSymbolLike != 0 {
			rendered = append(rendered, "typeof "+symbol.Name)
			continue
		}
		rendered = append(rendered, typeChecker.TypeToString(missingType))
	}
	return strings.Join(rendered, " | ")
}
