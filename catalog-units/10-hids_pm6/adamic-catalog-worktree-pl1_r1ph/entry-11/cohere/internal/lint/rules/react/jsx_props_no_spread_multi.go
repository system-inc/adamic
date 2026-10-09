package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageJsxPropsNoSpreadMultiIdentifier = rule.Message{
	Id: "jsxPropsNoSpreadMultiIdentifier",
	Description: "This value is spread into the same element more than once. React builds the " +
		"props object by applying each spread in order, so any prop written between the two " +
		"spreads is overwritten by the second one and never reaches the component. Even where " +
		"nothing sits between them, the second spread recomputes work the first already did. " +
		"Remove all but one spread.",
}

var messageJsxPropsNoSpreadMultiMemberExpression = rule.Message{
	Id: "jsxPropsNoSpreadMultiMemberExpression",
	Description: "This property is spread into the same element more than once. React builds the " +
		"props object by applying each spread in order, so any prop written between the two " +
		"spreads is overwritten by the second one and never reaches the component. Even where " +
		"nothing sits between them, the second spread recomputes work the first already did. " +
		"Remove all but one spread.",
}

// JsxPropsNoSpreadMulti flags a JSX element that spreads the same expression more than once.
//
//	valid:   <App {...a} />
//	valid:   <App {...a} {...b} />
//	valid:   <App {...props.x} {...props.foo} />
//	invalid: <App {...props} {...props} />
//	invalid: <App {...props.foo} {...props.foo} />
//	invalid: <div {...props} a="a" {...props} />
//
// Ported from `react/jsx-props-no-spread-multi`, read against oxc's
// `jsx_props_no_spread_multi.rs` and against `eslint-plugin-react`'s own
// `jsx-props-no-spread-multi.js`. Neither source carries an option surface: oxc declares no
// `from_configuration` and ESLint's `meta` block has no `schema` key at all, so the inventory's
// `options: "no"` is right here and was checked rather than taken.
//
// # Two judgments, not one, and they count differently
//
// The name suggests one question and the implementation asks two, against two disjoint sets of
// spreads, with different reporting arithmetic. Establishing this was the whole first hour, because
// a port that collapses them passes every imported fixture: upstream's fail corpus has at most one
// repeated expression per element, so the two readings agree on all five.
//
// **Identifier spreads are grouped.** `{...props}` is keyed by the identifier's name into a map, and
// the whole group produces exactly one finding however many copies there are. So
// `<div {...props} {...props} {...props} />` reports ONCE.
//
// **Member-expression spreads are paired.** `{...props.foo}` is pushed onto a list and every
// unordered PAIR is compared, so N copies produce N*(N-1)/2 findings.
// `<App {...props.foo} {...props.foo} {...props.foo} />` reports THREE times. Measured on the
// release oxlint binary rather than inferred from `array_combinations`, and pinned by a fixture
// below that upstream does not have, since upstream never writes a third copy of a member
// expression.
//
// The two sets do not overlap and neither is a fallback for the other: an identifier is never a
// member expression, so `<App {...props} {...props.foo} />` is silent in both halves.
//
// # Where ESLint and oxc disagree, and which one this follows
//
// ESLint implements only the identifier half. Its filter is
// `attr.argument.type === 'Identifier'`, so `<App {...props.foo} {...props.foo} />` is silent there
// and reports here. It also reports once per extra occurrence rather than once per group, so a
// third copy is a second finding under ESLint and still one finding here. oxc is followed on both
// counts, because oxc is what the gate being replaced actually runs, and because the member half is
// the larger part of what the rule is for. The divergence is stated rather than quietly resolved.
//
// # Parentheses and TypeScript wrappers, which upstream does strip
//
// This is the one place the brief warns that guessing is free at fixture time, so it was measured
// on the release binary in both directions rather than assumed. oxc reaches
// `without_parentheses()` and then `get_identifier_reference()`, which itself calls
// `get_inner_expression()`, so the peeled set is parentheses plus the TypeScript wrappers `as`,
// `satisfies`, `!`, instantiation and type assertion. `ast.SkipOuterExpressions` with
// `OEKParentheses|OEKAssertions` is the same set, less `OEKPartiallyEmittedExpressions`, which the
// parser never produces from source text.
//
// Measured, all against the release binary:
//
//	<App {...(props)} {...props} />                       REPORTS   parens peel
//	<App {...props!} {...props!} />                       REPORTS   non-null peels
//	<App {...props as any} {...props as any} />           REPORTS   as peels
//	<App {...(props.foo)} {...(props.foo)} />             REPORTS   and names `props.foo`
//	<App {...(props.foo).baz} {...(props.foo.baz)} />     REPORTS   inner parens peel in the compare
//	<App {...(props.foo).baz} {...(props.y.baz)} />       SILENT    different receiver
//
// The fourth line is the subtle one and it is why the finding is reported on the spread attribute
// rather than on the expression: after peeling, the two arguments are the same node shape but their
// source ranges differ from the written text, so a rule reporting the expression would point at a
// range that does not include the parentheses the reader wrote.
//
// # Optional chaining, which is the hardest thing in this port and is entirely measured
//
// `<App {...props?.foo} {...props?.foo} />` is SILENT in oxc and is silent here, and the two are
// silent for completely different reasons, which is why this needs a section rather than a line.
//
// oxc wraps an optional chain in a `ChainExpression`, a sibling variant of `MemberExpression`, so
// `as_member_expression()` on the spread argument returns `None` and the spread never enters either
// collection. Our parser has no such wrapper: `props?.foo` parses to an ordinary
// `KindPropertyAccessExpression` carrying a question-dot token, which a literal translation of
// oxc's code would happily compare against a plain `props.foo` and report. So this is the one place
// in the port where reproducing upstream takes an extra condition rather than one fewer.
//
// The extra condition has a boundary, and the boundary is not where anybody would guess. Six inputs
// were run against the release binary to find it, none of which any imported fixture writes:
//
//	<App {...props?.foo} {...props?.foo} />    SILENT    the plain case
//	<App {...a?.b.c}     {...a?.b.c}     />    SILENT    a question-dot one link DOWN still
//	                                                     disqualifies: the ChainExpression wraps the
//	                                                     whole chain, not the one optional link
//	<App {...(a?.b.c)}   {...(a?.b.c)}   />    SILENT    parenthesizing the whole chain wraps the
//	                                                     ChainExpression and changes nothing
//	<App {...(a?.b).c}   {...(a?.b).c}   />    REPORTS   parenthesizing a PREFIX terminates the
//	                                                     chain, so the outer access is ordinary
//	<App {...(a?.b).c}   {...(a.b).c}    />    REPORTS   and a terminated chain compares EQUAL to
//	                                                     the non-optional spelling
//	<App {...props[a?.b]} {...props[a.b]} />   REPORTS   a chain NESTED in a subscript is ordinary
//	                                                     structure and compares equal too
//
// Two rules fall out and both are load-bearing. The exclusion applies only to the spread ARGUMENT,
// never to an expression nested inside one, which is why `isSameExpression` reaches for a plain
// access check while the argument loop reaches for `isSpreadableMemberExpression`. And the walk
// looking for a question-dot descends through receivers but STOPS at a parenthesis or a TypeScript
// assertion, because those terminate the chain rather than being transparent to it.
//
// The first version of this rule had both halves wrong: it excluded nested chains and it descended
// through parentheses, so lines four, five and six above were silently missed. Neither error was
// visible from any fixture. What exposed them was mutating the descent, finding the mutant survived,
// and going to the binary to ask why rather than adding a fixture to kill it.
//
// The intuitive reading is that `props?.foo` twice IS the same double spread and should report, and
// that reading is recorded here beside the actual one precisely so the next reader does not
// helpfully correct the rule back to it.
//
// # A variable subscript compares as equal here, which the shelf declines to do
//
// `<App {...props[k]} {...props[k]} />` REPORTS in oxc, measured. That is worth stating because it
// runs against the grain of every other same-reference judgment in this tree:
// `no_self_assign.go`'s `isSameReference` returns false for `a[i]` even against itself, and
// `property.AccessedName` declines a variable subscript by design, its doc saying a comparison
// involving one is "false rather than optimistic".
//
// Both of those are right for their own question and wrong for this one. They ask whether two reads
// hit the same memory, where an intervening write to `k` makes the answer unknowable. This rule asks
// whether the author wrote the same spread twice inside a single JSX element, where nothing can run
// between them, so `k` cannot change and the two spreads are genuinely redundant. oxc's
// `is_same_member_expression` falls through to a structural comparison of the subscript expressions
// for exactly this reason. So `AccessedName` is used for the static-key case it answers well and the
// comparison falls back to structure when it declines, rather than treating its decline as a verdict.
//
// # What a call expression contributes, which is nothing
//
// `<App {...props()} {...props()} />` is silent, measured. A call is neither an identifier nor a
// member expression, so it enters neither collection. This is upstream's judgment and it is
// correct rather than a gap: two calls may return different objects, so the second spread is not
// redundant.
//
// # No fix, and the reason is stated rather than left as an omission
//
// oxc declares `fix` and ships one, deleting all but the last spread. This port reports without it.
// The identifier half's fixer is defensible on its own, but the member half emits one fix per
// PAIR, so a third copy produces three findings whose fixes delete overlapping ranges, and the
// merged result on `<App {...props.foo} {...props.foo} {...props.foo} />` was measured on the
// release binary as `<App   {...props.foo} />`: correct output arrived at by three separate
// deletions that happen not to conflict. Reproducing that arithmetic faithfully is possible and
// reproducing it wrongly deletes a spread the author needed, unattended, and a wrong fix is worse
// than no fix. The subset shown correct here is the detection, and it is complete. If the fix is
// wanted later it should come with the three-copy case asserted through `ExpectFixedSource`, which
// is the only assertion that can see a repair writing the right text over the wrong span.
//
// # Two element kinds where oxc has one
//
// oxc listens on `JSXOpeningElement` and gets both shapes, because its parser gives a self-closing
// element an opening element carrying a `self_closing` flag. Our parser gives a self-closing
// element a distinct `JsxSelfClosingElement` with no opening element inside it, so listening on the
// literal translation of oxc's node would make this rule silent on every fail fixture upstream
// ships: all five are self-closing. Both kinds are listened on, which `jsx_no_duplicate_props.go`
// established here first.
var JsxPropsNoSpreadMulti = rule.Rule{
	// No namespace prefix. The config writes `react/jsx-props-no-spread-multi` and the parity guard
	// strips the namespace on a `/` boundary, so a rule named `react-jsx-props-no-spread-multi`
	// would match no inventory entry, lint no files, and still pass every fixture in this package.
	Name: "react/jsx-props-no-spread-multi",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		report := func(node *ast.Node) {
			_, attributes := jsx.ElementParts(node)
			if attributes == nil {
				return
			}
			properties := attributes.AsJsxAttributes().Properties
			if properties == nil {
				return
			}

			// The two collections upstream keeps, and they are deliberately separate rather than one
			// list with a kind tag: the identifier half is grouped by name and reports once per
			// group, while the member half is compared pairwise and reports once per pair. Merging
			// them would force one arithmetic onto both.
			//
			// firstIdentifierSpread carries the earliest spread of each name, which is what gets
			// reported, and reportedIdentifiers keeps a group from reporting twice when a third copy
			// arrives.
			firstIdentifierSpread := make(map[string]*ast.Node, len(properties.Nodes))
			reportedIdentifiers := make(map[string]bool, len(properties.Nodes))

			// Parallel slices rather than a struct slice, because both are read only in the pair
			// loop below and never handed anywhere.
			var memberExpressions []*ast.Node
			var memberSpreads []*ast.Node

			for _, jsxProperty := range properties.Nodes {
				// A spread is a node rather than an absence: it is a KindJsxSpreadAttribute sitting
				// in the same Properties list as ordinary attributes. A rule listening only for
				// KindJsxAttribute sees nothing here, which is why the kind is matched directly.
				if jsxProperty.Kind != ast.KindJsxSpreadAttribute {
					continue
				}

				argument := ast.SkipOuterExpressions(
					jsxProperty.AsJsxSpreadAttribute().Expression,
					ast.OEKParentheses|ast.OEKAssertions,
				)
				if argument == nil {
					continue
				}

				if argument.Kind == ast.KindIdentifier {
					name := argument.Text()
					if earlier, seen := firstIdentifierSpread[name]; seen {
						// Once per group, not once per extra copy. A third spread of the same name
						// finds the group already reported and adds nothing, which is what makes
						// `<div {...props} {...props} {...props} />` a single finding.
						if !reportedIdentifiers[name] {
							reportedIdentifiers[name] = true
							ctx.ReportNode(earlier, messageJsxPropsNoSpreadMultiIdentifier)
						}
						continue
					}
					firstIdentifierSpread[name] = jsxProperty
					continue
				}

				if isSpreadableMemberExpression(argument) {
					memberExpressions = append(memberExpressions, argument)
					memberSpreads = append(memberSpreads, jsxProperty)
				}
			}

			// Every unordered pair, which is upstream's `array_combinations`. Reporting on the left
			// member of each pair matches where oxc's diagnostic points: it passes both spans as
			// labels and the first is the primary position, checked against the snapshot and
			// against the release binary's json output, whose first label is the earlier spread.
			for leftIndex := 0; leftIndex < len(memberExpressions); leftIndex++ {
				for rightIndex := leftIndex + 1; rightIndex < len(memberExpressions); rightIndex++ {
					if isSameMemberExpression(memberExpressions[leftIndex], memberExpressions[rightIndex]) {
						ctx.ReportNode(memberSpreads[leftIndex], messageJsxPropsNoSpreadMultiMemberExpression)
					}
				}
			}
		}

		return rule.Listeners{
			// Both kinds, and neither is redundant. Every one of upstream's five fail fixtures is
			// self-closing, so dropping the second arm would silence the entire imported corpus
			// while leaving the first arm's fixtures green.
			ast.KindJsxOpeningElement:     report,
			ast.KindJsxSelfClosingElement: report,
		}
	},
}

// isSpreadableMemberExpression reports whether a spread ARGUMENT is what oxc's
// `as_member_expression` accepts on the spread argument itself.
//
// A property access and an element access, and nothing else. A call, an identifier, a literal and an
// object all decline, which is why `<App {...props()} {...props()} />` and
// `<App {...{a: 1}} {...{a: 1}} />` are both silent upstream.
//
// An optional chain declines too, and getting the boundary of that decline right took four
// measurements against the release binary, because the intuitive reading is wrong in both
// directions at once. See `isOptionalChain` for the boundary. The decline belongs HERE, on the
// argument, rather than inside the comparison: declining it in the comparison would leave the
// spread in the collection where it would pair against a plain `props.foo`, giving a pair count
// that is right for the wrong reason.
//
// The name says `Spreadable` rather than plain `isMemberExpression` because the chain exclusion is
// specific to the argument position. An identical node NESTED inside a subscript is not excluded,
// and `isSameExpression` deliberately does not call this. That distinction is measured:
// `<App {...props[a?.b]} {...props[a.b]} />` REPORTS.
func isSpreadableMemberExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind != ast.KindPropertyAccessExpression && node.Kind != ast.KindElementAccessExpression {
		return false
	}
	return !isOptionalChain(node)
}

// isOptionalChain reports whether a node is what oxc's parser would wrap in a `ChainExpression`.
//
// Four measurements against the release oxlint binary define this boundary, and every one of them
// contradicts a reading somebody would arrive at by thinking about it:
//
//	<App {...a?.b.c} {...a?.b.c} />        SILENT   the question-dot is one link DOWN and still
//	                                               disqualifies, because oxc's ChainExpression wraps
//	                                               the WHOLE chain rather than the one optional link
//	<App {...(a?.b.c)} {...(a?.b.c)} />    SILENT   parenthesizing the whole chain wraps the
//	                                               ChainExpression and changes nothing
//	<App {...(a?.b).c} {...(a?.b).c} />    REPORTS  parenthesizing a PREFIX terminates the chain, so
//	                                               the outer `.c` is an ordinary member expression
//	                                               whose receiver happens to be a ChainExpression
//	<App {...(a?.b).c} {...(a.b).c} />     REPORTS  and the terminated chain compares EQUAL to the
//	                                               non-optional spelling, because the comparison
//	                                               unwraps the ChainExpression
//
// So the walk descends through receivers, which the first two lines require, and stops at a
// parenthesis, which the third and fourth require. Descending through parentheses was this rule's
// first version: it made line three silent, and a mutation of the descent is what exposed it, since
// no imported fixture writes any of these four shapes.
//
// TypeScript's non-null and assertion wrappers are NOT walked through either, for the same reason as
// the parenthesis: `a?.b!.c` terminates the chain at the `!` in oxc exactly as `(a?.b).c` does.
func isOptionalChain(node *ast.Node) bool {
	for node != nil {
		switch node.Kind {
		case ast.KindPropertyAccessExpression:
			access := node.AsPropertyAccessExpression()
			if access.QuestionDotToken != nil {
				return true
			}
			node = access.Expression
		case ast.KindElementAccessExpression:
			access := node.AsElementAccessExpression()
			if access.QuestionDotToken != nil {
				return true
			}
			node = access.Expression
		case ast.KindCallExpression:
			call := node.AsCallExpression()
			if call.QuestionDotToken != nil {
				return true
			}
			node = call.Expression
		default:
			// A parenthesis, a non-null assertion or a type assertion TERMINATES the chain rather
			// than being seen through. That is the third measured line above, and descending here
			// instead is the defect this comment exists to prevent being reintroduced.
			return false
		}
	}
	return false
}

// isSameMemberExpression reports whether two member accesses spread the same property of the same
// receiver, reproducing oxc's `is_same_member_expression`.
//
// Shape taken from `no_self_assign.go`'s `isSameReference`, which asks a structurally similar
// question, and its semantics deliberately NOT taken: that function treats `a.b` and `a?.b` as one
// reference and returns false for `a[i]` against itself, and this rule needs the opposite answer on
// both. See the rule's doc comment for why the two questions differ.
//
// The static-key path and the structural path are both needed and neither subsumes the other.
// `AccessedName` settles `.foo` against `["foo"]`, which is how `<App {...props["foo"]}
// {...props.foo} />` reports, measured. It declines a variable subscript, and where BOTH decline the
// comparison falls through to comparing the subscript expressions structurally, which is how
// `<App {...props[k]} {...props[k]} />` reports, also measured. Where exactly one declines the two
// are different shapes and the answer is false, which is oxc's `(Some, None)` arm.
// Both arguments arrive already peeled and this function deliberately does not peel them again.
// The two call sites are the pair loop, whose entries were peeled when they were collected, and
// `isSameExpression`, which peels at its own top before dispatching here. A peel here as well was
// in the first version and a mutation removing it SURVIVED, correctly: no input can reach this
// function with an unpeeled node, so the line was unreachable defence rather than a discrimination.
// It is deleted rather than covered by a fixture, since a fixture over an unreachable branch
// asserts nothing.
func isSameMemberExpression(left *ast.Node, right *ast.Node) bool {
	if left == nil || right == nil {
		return false
	}
	// A plain access check rather than the spreadable one, because this function is reached both
	// from the argument position and from inside `isSameExpression`, where a chain is transparent.
	// The argument position has already applied the chain exclusion before anything gets here.
	if !ast.IsAccessExpression(left) || !ast.IsAccessExpression(right) {
		return false
	}

	// TAGGED rather than plain, and the tag is load-bearing. `property.AccessedName` renders a
	// numeric key and a string key with the same digits as one name, so `o[1]` and `o["1"]` compare
	// equal under it, and `<App {...props[o[1]]} {...props[o["1"]]} />` then reports where upstream
	// is SILENT. oxc keys its literal comparison on matched pairs of variants, so a NumericLiteral
	// and a StringLiteral never meet whatever their text.
	//
	// This is exactly the caller `property.NameTagged`'s doc says it exists for, and it was found by
	// a mutation of the literal arm rather than by any fixture: no imported case writes two keys
	// whose text agrees across types.
	leftName, leftStatic := accessedNameTagged(left)
	rightName, rightStatic := accessedNameTagged(right)
	switch {
	case leftStatic && rightStatic:
		if leftName != rightName {
			return false
		}
	case leftStatic != rightStatic:
		// One names a property the syntax settles and the other does not, so they are different
		// shapes whatever the receivers are. oxc's `(Some(_), None) | (None, Some(_))` arm.
		return false
	default:
		// Neither key is static, so the two subscripts are compared as expressions. Only reached
		// when both sides are element accesses, since a property access always has a static name.
		if left.Kind != ast.KindElementAccessExpression || right.Kind != ast.KindElementAccessExpression {
			return false
		}
		if !isSameExpression(
			left.AsElementAccessExpression().ArgumentExpression,
			right.AsElementAccessExpression().ArgumentExpression,
		) {
			return false
		}
	}

	return isSameExpression(accessedReceiver(left), accessedReceiver(right))
}

// isSameExpression compares two expressions structurally, reproducing the subset of oxc's
// `is_same_expression` this rule can reach.
//
// The subset is deliberate rather than an abbreviation. A spread argument reaches here only as the
// receiver of a member access or as a computed subscript, so the shapes upstream handles that
// cannot occur in that position are not reproduced: an arrow function, a class expression and an
// await are all syntactically reachable but would compare as false under upstream's own default
// arm, which is what returning false here gives them. What is reproduced is every shape a real
// spread receiver takes.
func isSameExpression(left *ast.Node, right *ast.Node) bool {
	left = ast.SkipOuterExpressions(left, ast.OEKParentheses|ast.OEKAssertions)
	right = ast.SkipOuterExpressions(right, ast.OEKParentheses|ast.OEKAssertions)
	if left == nil || right == nil {
		return false
	}

	// No chain exclusion here, deliberately. Inside a subscript or a receiver an optional chain is
	// ordinary structure, and oxc's `is_same_expression` unwraps a `ChainExpression` to compare the
	// member expression inside it. Measured, both directions:
	//
	//	<App {...props[a?.b]} {...props[a.b]} />    REPORTS
	//	<App {...(a?.b).c}    {...(a.b).c}    />    REPORTS
	//
	// So `a?.b` compares EQUAL to `a.b` once it is nested, which is why this reaches for the plain
	// access check rather than `isSpreadableMemberExpression`.
	if ast.IsAccessExpression(left) && ast.IsAccessExpression(right) {
		return isSameMemberExpression(left, right)
	}

	if left.Kind != right.Kind {
		return false
	}

	switch left.Kind {
	case ast.KindIdentifier, ast.KindPrivateIdentifier:
		return left.Text() == right.Text()
	case ast.KindThisKeyword, ast.KindSuperKeyword, ast.KindNullKeyword:
		return true
	case ast.KindStringLiteral, ast.KindNumericLiteral,
		ast.KindNoSubstitutionTemplateLiteral, ast.KindRegularExpressionLiteral:
		return left.Text() == right.Text()
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	}

	return false
}

// accessedNameTagged is `property.AccessedName` with the key's type carried alongside it.
//
// The shelf has `property.NameTagged` for a property KEY and `property.AccessedName` for a member
// ACCESS, but no tagged variant of the access reader, so this composes the two: it takes the
// subscript or name node the way `AccessedName` does and then reads it through `NameTagged` rather
// than `Name`.
//
// It lives here rather than on the shelf because exactly one caller wants it. `NameTagged`'s own doc
// makes that argument for the key version, saying one caller needs it and the rest must not have it,
// and the same holds one level up. If a second rule ever needs this, it should move to
// `internal/utilities/ecmascript/property/` rather than be copied, since two helpers doing almost the
// same thing is worse than one slightly wrong shape.
//
// The identifier-subscript decline is `AccessedName`'s and is kept deliberately: `o[k]` answers
// nothing here, and the caller falls through to comparing the subscripts structurally, which is what
// makes `<App {...props[k]} {...props[k]} />` report.
func accessedNameTagged(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return property.NameTagged(node.AsPropertyAccessExpression().Name(), property.Static)
	case ast.KindElementAccessExpression:
		argument := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		if argument == nil || argument.Kind == ast.KindIdentifier ||
			argument.Kind == ast.KindPrivateIdentifier {
			return "", false
		}
		return property.NameTagged(argument, property.Static)
	}
	return "", false
}

// accessedReceiver reads the object a member access reads from.
func accessedReceiver(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		return node.AsElementAccessExpression().Expression
	}
	return nil
}
