package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

func buildNoMixedEnumsMessage() rule.Message {
	return rule.Message{
		Id: "mixed",
		Description: "This enum has both number members and string members. A reader reaches for an " +
			"enum expecting one kind of value, and a mixed one silently changes what comparison, " +
			"serialization, and reverse lookup mean from member to member: numeric members get a " +
			"reverse mapping and string members do not. Pick one kind for the whole enum.",
	}
}

// noMixedEnumsMemberType is the three-way classification upstream calls AllowedType.
//
// `Unknown` is not "we could not tell". It is a real third verdict that ABORTS the whole enum: a
// member whose initializer is a boolean literal makes upstream stop scanning rather than treat the
// member as either kind, so `enum E { A = 1, B = false, C = 'c' }` is silent even though the first
// and third members plainly disagree. Measured on the installed rule, which is the only way to
// learn it, because reading the source shows a `return` inside a loop and not what it costs.
type noMixedEnumsMemberType int

const (
	noMixedEnumsNumber noMixedEnumsMemberType = iota
	noMixedEnumsString
	noMixedEnumsUnknown
)

// NoMixedEnums flags an enum whose members are not all of one kind.
//
//	valid:   enum E { A, B }
//	valid:   enum E { A = 0, B = 1 }
//	valid:   enum E { A = 'a', B = 'b' }
//	valid:   enum E { A = 1, B = false }
//	valid:   enum E { A = 1 } enum E { B = 2 }
//	invalid: enum E { A, B = 'b' }
//	invalid: enum E { A = 'a', B = 1 }
//	invalid: enum E { A = 1 } enum E { B = 'b' }
//
// A numeric enum member gets a reverse mapping in the emitted object and a string member does not,
// so a mixed enum makes `E[E.A]` work for some members and return undefined for others. The rule is
// about that asymmetry rather than about taste.
//
// # What decides a finding
//
// One listener on `KindEnumDeclaration`. An empty body is declined first. Then a DESIRED type is
// established for the whole declaration, and the members are scanned in order; the first member
// disagreeing with the desired type reports, and the scan stops. Measured: an enum with three
// disagreeing members reports ONCE, on the first of them, not three times.
//
// The desired type comes from the first of these that answers, and the order is upstream's:
//
//	an imported merge     a `declare module` augmenting an enum imported from elsewhere takes the
//	                      kind from THAT enum's first member
//	an earlier sibling    a previous non-empty `enum E` in the same scope, textually before this one
//	a namespace merge     an exported enum inside a namespace block, merged with another block
//	this enum itself      the kind of its own first member
//
// # The `Unknown` verdict aborts, and it is not a fallback
//
// A member classifies as Unknown only when its initializer is a literal that is neither a number
// nor a string, which in practice means a boolean. Reaching one abandons the entire declaration.
// Measured, holding everything else fixed:
//
//	enum E { A = 1, B = false }              CLEAN
//	enum E { A = 1, B = false, C = 'c' }     CLEAN   (the string never gets judged)
//	enum E { A = 1, C = 'c' }                REPORTS (control: the pair alone does disagree)
//
// The third row is the control. Without it the first two read as "booleans are numbers", which is
// the intuitive reading and the wrong one.
//
// # Number is the type that can be ADOPTED, and string is not
//
// Upstream writes `desiredType ??= currentType` guarded by `currentType === Number`. That asymmetry
// only bites when the desired type is still unset, which happens when the first member's
// initializer is a call or another expression the literal switch does not name. Measured:
//
//	declare const f: () => any;
//	enum E { A = f(), B = 1, C = 'c' }       REPORTS on `'c'`
//
// `A` takes the checker arm and comes back Number, `B` is Number, and `C` disagrees. A port that
// skipped the adoption line would have no desired type at `C` and go silent.
//
// # Parentheses, and why this port skips them while upstream has no code for it
//
// ESTree does not represent a parenthesis as a node, so upstream's `member.initializer.type`
// switch is handed the INNER node for `A = (1)` and its Literal arm fires. Our parser produces a
// real `KindParenthesizedExpression`, so a literal port of that switch would fall through to the
// checker arm on every parenthesized initializer and classify a parenthesized boolean as Number
// instead of Unknown.
//
// That is a parser difference rather than a rule difference. Measured on the installed rule, with
// the bare forms as controls on each row:
//
//	enum E { A = 'a', B = (1) }        REPORTS on `1`, span EXCLUDES the parens
//	enum E { A = 'a', B = (false) }    CLEAN    -- same as the bare `false`
//	enum E { A = 'a', B = false }      CLEAN    (control)
//
// The second row is the one that settles it: if parentheses routed to the checker arm, a
// parenthesized boolean would classify as Number and REPORT. It does not, so upstream is seeing
// through the parenthesis to the literal, and skipping is what reproduces the decision. Confirmed
// directly against the parser as well: `A = (1)` yields a `Literal` node whose range is the inner
// `1`, which is also why the span excludes the wrapper.
//
// # A template literal is a string even with a substitution
//
// Upstream's switch names `TemplateLiteral` and returns String without inspecting it, so a member
// initialized to a template carrying a substitution is a string member regardless of what the
// substituted expression's type is. Reproduced rather than improved on: asking the checker would
// be more careful and would disagree with the rule being ported.
//
// # The merge cases, and what replaces the scope manager
//
// Upstream walks `context.sourceCode.getScope(node)` upward, reading the resolved variable's
// definitions to find both an earlier sibling declaration and any import bindings. We have no
// scope manager, and the checker answers the same question: the enum's own name resolves to a
// symbol whose `Declarations` is the merged list, which is exactly what upstream reconstructs.
//
// Three properties of that list were measured rather than assumed, because the brief's warning
// about `Declarations[0]` applies directly here:
//
//	enum E { A = 1 } enum E { B = 'b' }                  REPORTS on `'b'`  (earlier sibling wins)
//	enum E { A = 'a' } enum E { B = 1 }                  REPORTS on `1`
//	enum E {} enum E { A = 1, B = 'b' }                  REPORTS on `'b'`  (empty one is skipped)
//	enum E { A = 1 } enum E { B = 'b' } enum E { C = 2 } REPORTS on `'b'` ONLY
//
// The fourth row is the one a reader would not predict: the third declaration is judged against
// the FIRST, not against the second, because upstream takes the earliest non-empty declaration
// positioned before it and stops. It agrees with the first, so it is clean, and only the middle
// one reports.
//
// A non-exported enum inside a namespace does not merge and is judged on its own; measured CLEAN
// for a pair that would otherwise disagree.
//
// # Cost
//
// `KindEnumDeclaration` is a rare anchor. The checker is consulted once for the declaration's own
// name, and once per member whose initializer is not a literal the switch already names.
var NoMixedEnums = rule.Rule{
	Name: "@typescript-eslint/no-mixed-enums",

	// The classification of a non-literal initializer is a type question, and so is finding the
	// merged declaration list.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindEnumDeclaration: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				declaration := node.AsEnumDeclaration()
				if declaration == nil || declaration.Members == nil || len(declaration.Members.Nodes) == 0 {
					return
				}
				members := declaration.Members.Nodes

				// A pointer rather than a value, because "no desired type yet" is a real state
				// distinct from "the desired type is Number". Upstream spells it as `undefined`
				// and its `??=` adoption line only fires in that state.
				desiredType := noMixedEnumsDesiredType(ctx, node, members)

				// Upstream's `if (desiredType === ts.TypeFlags.Unknown) return;`, which sits above
				// the member scan and abandons the declaration outright. It is reachable only when
				// a MERGE REFERENCE classified as Unknown, because this enum's own first member
				// being Unknown would abort on the first iteration anyway. Measured: the pair
				// `enum E { Z = false }` followed by `enum E { A = 1, B = 'b' }` is clean upstream,
				// and without this test it reports.
				if desiredType != nil && *desiredType == noMixedEnumsUnknown {
					return
				}

				for _, member := range members {
					currentType := noMixedEnumsClassify(ctx, member)

					// Not "unclassifiable, keep going". Reaching one abandons the whole enum.
					if currentType == noMixedEnumsUnknown {
						return
					}

					// Upstream's `desiredType ??= currentType`, guarded on Number.
					//
					// It is UNREACHABLE here and kept anyway, with the measurement rather than an
					// argument, because the two implementations arrive at "no desired type" by
					// different routes and only upstream's route can still be scanning.
					//
					// noMixedEnumsDesiredType returns nil only when the declaration has no members,
					// and the listener has already declined that case above. Every other path
					// returns a classified pointer. So `desiredType` is non-nil at this line for
					// every input that reaches it.
					//
					// Measured rather than argued: replacing this branch's body with a panic and
					// running the whole suite -- 51 imported cases plus every probe below them --
					// produced no panic. A mutation neutralizing the branch also survives, and that
					// survival is CORRECT rather than a fixture gap.
					//
					// The verdict names its callers, so it expires if any of them changes: the only
					// nil-producing path is noMixedEnumsClassifyFirst's empty-member-list return. If
					// a future change makes that function return nil for some other shape, or adds
					// a return of nil to noMixedEnumsDesiredType, this line becomes live again and
					// the reasoning here has to be re-taken.
					if desiredType == nil && currentType == noMixedEnumsNumber {
						adopted := currentType
						desiredType = &adopted
					}

					if desiredType != nil && currentType != *desiredType {
						// A member with no initializer has no initializer node to point at, so the
						// finding lands on the member itself. Measured: `enum E { A = 'a', B }`
						// reports on `B`, spanning just the name.
						//
						// The parenthesis skip is repeated here rather than only in the classifier,
						// and it is a SECOND decision rather than a restatement of the first: ESTree
						// has no parenthesis node, so upstream's report node for `B = (1)` is the
						// inner literal and its span excludes the wrapper. Reporting the outer node
						// would classify correctly and still point at text upstream never
						// highlights, which the count assertions cannot see. Caught by the span
						// fixture, not by any imported case.
						target := member.AsEnumMember().Initializer
						if target == nil {
							target = member
						} else if skipped := ast.SkipParentheses(target); skipped != nil {
							target = skipped
						}
						ctx.ReportNode(target, buildNoMixedEnumsMessage())
						return
					}
				}
			},
		}
	},
}

// noMixedEnumsDesiredType establishes the kind the whole declaration is expected to be, or nil when
// nothing has settled it yet and the scan is left to adopt one.
//
// Upstream's three merge cases collapse into one question here: does this enum's name resolve to a
// symbol carrying an enum declaration OTHER than this one that should set the kind? The checker's
// merged declaration list answers all three, so the import walk, the sibling scan, and the
// namespace special case are one loop rather than three branches.
//
// The ordering upstream encodes is preserved: a declaration from ANOTHER FILE wins over a local
// sibling, because an ambient module augmentation is upstream's first case and its sibling scan is
// the second.
func noMixedEnumsDesiredType(ctx rule.Context, node *ast.Node, members []*ast.Node) *noMixedEnumsMemberType {
	name := node.AsEnumDeclaration().Name()
	if name == nil {
		return noMixedEnumsClassifyFirst(ctx, members)
	}

	symbol := ctx.TypeChecker.GetSymbolAtLocation(name)

	// An EXPORTED declaration is handed a symbol carrying only itself, and the merged list lives on
	// the declaration's LocalSymbol. Without this fallback every exported enum in a namespace would
	// look unmerged, which is exactly the namespace-merging case upstream has a dedicated branch
	// for. `LocalSymbol` is nil on a non-exported declaration, so it is a fallback and not a
	// replacement, and the longer list is the merged one.
	if local := node.LocalSymbol(); local != nil {
		if symbol == nil || len(local.Declarations) > len(symbol.Declarations) {
			symbol = local
		}
	}

	if symbol != nil {
		var fromOtherFile *ast.Node
		var earlierSibling *ast.Node

		for _, candidate := range symbol.Declarations {
			if candidate == nil || candidate == node || !ast.IsEnumDeclaration(candidate) {
				continue
			}
			candidateMembers := candidate.AsEnumDeclaration().Members
			// An EMPTY declaration is skipped rather than treated as agreeing with anything.
			// Measured: `enum E {} enum E { A = 1, B = 'b' }` reports, so the empty one did not
			// become the reference.
			//
			// The length half of this test is SUBSUMED and kept deliberately. Mutating it away
			// survives the whole suite, and that verdict is correct rather than a fixture gap:
			// noMixedEnumsClassifyFirst returns nil for an empty member list, so selecting an
			// empty declaration as the reference yields the same nil desired type that skipping it
			// yields. Confirmed by applying the mutation and running the two tests that cover the
			// path, both of which stayed green. Kept because it states the intent at the point the
			// candidate is chosen rather than leaving it to a caller three functions away, and
			// because the nil-return in that caller is not obviously load-bearing from here.
			if candidateMembers == nil || len(candidateMembers.Nodes) == 0 {
				continue
			}

			if ast.GetSourceFileOfNode(candidate) != ast.GetSourceFileOfNode(node) {
				if fromOtherFile == nil {
					fromOtherFile = candidate
				}
				continue
			}

			// Strictly BEFORE this one, and the earliest such. Upstream breaks out of its loop on
			// the first match in scope order, and the observable consequence is that a third
			// declaration is judged against the FIRST rather than the second. Measured.
			if candidate.Pos() < node.Pos() && (earlierSibling == nil || candidate.Pos() < earlierSibling.Pos()) {
				earlierSibling = candidate
			}
		}

		if reference := fromOtherFile; reference != nil {
			return noMixedEnumsClassifyFirst(ctx, reference.AsEnumDeclaration().Members.Nodes)
		}
		if earlierSibling != nil {
			return noMixedEnumsClassifyFirst(ctx, earlierSibling.AsEnumDeclaration().Members.Nodes)
		}
	}

	return noMixedEnumsClassifyFirst(ctx, members)
}

// noMixedEnumsClassifyFirst classifies a declaration's first member, which is what every merge case
// reads.
//
// An Unknown first member is returned AS Unknown rather than collapsed into nil, and that
// distinction is the whole reason this function exists rather than being inlined. Upstream's
// desired-type function returns `getMemberType(...)` straight out of its sibling branch, so an
// Unknown coming back from a MERGE REFERENCE reaches the listener's `desiredType === Unknown` test
// and abandons the enum before a single member is scanned.
//
// Collapsing it to nil instead lets the scan continue and adopt a kind from a later member, which
// costs a false positive on a shape no imported fixture writes:
//
//	enum E { Z = false }
//	enum E { A = 1, B = 'b' }        upstream CLEAN, the nil-collapsing port REPORTED
//
// Found by a fixture written for a mutant that survived, which is the only reason it surfaced: the
// mutant was equivalent, and chasing why produced the probe that caught this.
func noMixedEnumsClassifyFirst(ctx rule.Context, members []*ast.Node) *noMixedEnumsMemberType {
	if len(members) == 0 {
		return nil
	}
	classified := noMixedEnumsClassify(ctx, members[0])
	return &classified
}

// noMixedEnumsClassify answers what kind one member is.
//
// The literal arms come first and answer without the checker, matching upstream's switch on the
// initializer's ESTree node type. Everything else asks the checker whether the type is StringLike,
// and a negative answer is NUMBER rather than unknown -- upstream's ternary has no third branch,
// which is why a call returning a boolean classifies as Number rather than aborting.
func noMixedEnumsClassify(ctx rule.Context, member *ast.Node) noMixedEnumsMemberType {
	initializer := member.AsEnumMember().Initializer

	// No initializer is an implicit number.
	if initializer == nil {
		return noMixedEnumsNumber
	}

	// ESTree has no parenthesis node, so upstream's switch sees the inner node. Ours does, so the
	// skip is what reproduces both the classification and the span. The doc comment on the rule
	// records the measurement that settles this, including the parenthesized-boolean control.
	initializer = ast.SkipParentheses(initializer)
	if initializer == nil {
		return noMixedEnumsUnknown
	}

	switch initializer.Kind {
	case ast.KindNumericLiteral:
		return noMixedEnumsNumber

	case ast.KindStringLiteral:
		return noMixedEnumsString

	// A template literal is a string whether or not it has a substitution. Upstream returns String
	// from the node kind alone without consulting the checker, and this reproduces that rather than
	// improving on it.
	case ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression:
		return noMixedEnumsString

	// A boolean literal is upstream's `default` inside the Literal arm: `typeof value` is neither
	// 'number' nor 'string', so it returns Unknown and the caller abandons the enum. This is the
	// one arm whose effect is not local to the member.
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return noMixedEnumsUnknown

	// A `null` literal is the same third case: `typeof null` is 'object'.
	case ast.KindNullKeyword:
		return noMixedEnumsUnknown
	}

	if ctx.TypeChecker == nil {
		return noMixedEnumsNumber
	}
	initializerType := ctx.TypeChecker.GetTypeAtLocation(initializer)
	if initializerType == nil {
		return noMixedEnumsNumber
	}
	if type_checking.IsTypeFlagSet(initializerType, checker.TypeFlagsStringLike) {
		return noMixedEnumsString
	}
	return noMixedEnumsNumber
}
