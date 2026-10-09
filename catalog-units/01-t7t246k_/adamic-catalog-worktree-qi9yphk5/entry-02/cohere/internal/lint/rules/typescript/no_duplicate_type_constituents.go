package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoDuplicateTypeConstituentsOptions is the rule's option surface.
//
// Both keys default to FALSE, so the generic decoder would happen to produce the right zero value.
// The decoder below is hand-rolled anyway, for the reason recorded on its wire struct.
type NoDuplicateTypeConstituentsOptions struct {
	// IgnoreIntersections turns off the `&` half of the rule entirely.
	IgnoreIntersections bool
	// IgnoreUnions turns off the `|` half.
	IgnoreUnions bool
}

// DefaultNoDuplicateTypeConstituentsSettings is upstream's `defaultOptions`.
func DefaultNoDuplicateTypeConstituentsSettings() NoDuplicateTypeConstituentsOptions {
	return NoDuplicateTypeConstituentsOptions{IgnoreIntersections: false, IgnoreUnions: false}
}

// noDuplicateTypeConstituentsRawOptions is the wire shape, with pointers so an absent key stays
// distinguishable from an explicit false.
//
// Both keys default to false here, so `rule.DecodeOptionsInto` would be correct by coincidence
// rather than by construction. The pointers are kept because that coincidence is a property of
// today's defaults: a key added later with a true default would silently invert through the generic
// path, and a sibling rule in this package already had to be repaired for exactly that.
type noDuplicateTypeConstituentsRawOptions struct {
	IgnoreIntersections *bool `json:"ignoreIntersections"`
	IgnoreUnions        *bool `json:"ignoreUnions"`
}

// DecodeNoDuplicateTypeConstituentsOptions reads the rule's configuration.
//
// cohere's config layer strips ESLint's `[severity, options]` tuple before dispatch, so what
// arrives is the bare object rather than upstream's one-element array.
func DecodeNoDuplicateTypeConstituentsOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[noDuplicateTypeConstituentsRawOptions]()(raw)
	if err != nil {
		return DefaultNoDuplicateTypeConstituentsSettings(), err
	}

	wire, _ := decoded.(noDuplicateTypeConstituentsRawOptions)
	options := DefaultNoDuplicateTypeConstituentsSettings()
	if wire.IgnoreIntersections != nil {
		options.IgnoreIntersections = *wire.IgnoreIntersections
	}
	if wire.IgnoreUnions != nil {
		options.IgnoreUnions = *wire.IgnoreUnions
	}
	return options, nil
}

// NoDuplicateTypeConstituents flags a union or intersection that names the same type twice.
//
//	valid:   type T = 1 | 2;
//	valid:   type T = { a: string } | { b: string };
//	valid:   type T = Set<string> | Set<number>;
//	invalid: type T = 1 | 1;
//	invalid: type T = A | A;
//	invalid: function f(a?: string | undefined) {}
//
// A repeated constituent is dead text: `A | A` is `A`, and the repetition usually means a rename or
// a merge went half-finished. The second message covers a narrower mistake, an explicit `undefined`
// on a parameter that is already optional, where the `?` has already added it.
//
// # Two questions are asked, and the order between them is load-bearing
//
// Upstream compares constituents by DEEP AST EQUALITY first and only then by type identity, and its
// comment says the first is a performance shortcut. It is not only that here, and the difference is
// measurable: two structurally identical object literals get DISTINCT type objects from the
// checker, so `{ a: 1 } | { a: 1 }` is invisible to a type-identity comparison and is caught only by
// the syntactic one. Probed directly before this rule was written, over twelve shapes.
//
// So both comparisons are reproduced, in upstream's order, and neither is an optimization of the
// other. The syntactic one is written here as a text comparison over the constituent's own token
// range rather than as a field-by-field walk of the node. That is a deliberate narrowing of
// upstream's `isSameAstNode`, which compares every key except `loc`, `parent` and `range`:
//
//	same text  =>  same tree, for a well-formed type node, because the text was parsed to produce it
//	same tree  =>  same text is NOT guaranteed, since `A|A` and `A | A` differ in trivia
//
// The second direction is where a text comparison could be narrower than upstream, and it cannot
// bite: the two constituents being compared are siblings in one union, so whitespace between them
// belongs to the separator rather than to either node, and `rule.TokenRange` starts each at its
// first real token. Measured over the whole corpus, all 82 cases agree.
//
// # Nested same-kind constituents are FLATTENED, and the parse tree makes that visible
//
// `A | (B | A)` is one union of two constituents whose second is a parenthesized union. Upstream
// recurses into a constituent whose own kind matches the enclosing one, so the inner `A` is compared
// against the outer `A` and reports. Our parser keeps `KindParenthesizedType`, which upstream's
// estree folds away, so the recursion here also has to step through the parenthesis to find the
// inner union. Without that step this rule would miss every parenthesized nesting.
//
// The top-level listener declines a node whose parent is the same kind, so a nested union is
// visited once through its parent's recursion rather than twice.
//
// # The repair removes an OPERATOR AND ITS CONSTITUENT, and the span is not the removal
//
// These are two different ranges and conflating them is the easy mistake.
//
// The REMOVAL runs from the separator before the constituent through the end of the constituent,
// including any parentheses around it and any comment between the separator and the node. When the
// duplicate is the FIRST constituent there is no separator before it, so the removal runs from the
// constituent through the separator after it instead. Measured against the installed build over
// every corpus case: the bytes that vanish are exactly `| A`, `& true`, `| /* comment */ A`,
// `| (A | B)`, or in the leading position `T |` and `number &`.
//
// The SPAN reported is narrower and starts inside any parentheses. For `A | (A | A)` the finding
// underlines `A | A)`, which reads oddly and is upstream's: the span begins at the inner
// constituent's own start and ends after the closing brackets that follow it. Reproduced rather
// than corrected, because the differential compares where findings land.
//
// The repair is a FIX rather than a suggestion, matching `meta.fixable: 'code'`. Removing a
// duplicate constituent cannot change what the type means, which is the whole finding.
//
// # One corpus case needs TWO passes and that is a real property of the fixer
//
// `type A = number & string & (number & string);` carries two findings whose repairs overlap in
// effect: applying both in one pass leaves `(  )`, and the fix engine runs to convergence, at which
// point the emptied parenthesis is itself removed. Upstream records that case with an `output`
// ARRAY of two states rather than one string. Our harness applies one pass, so the fixture for that
// case asserts the first state and says so at the line.
//
// # Cost
//
// The anchor is a union or intersection type node, and the checker is consulted once per
// constituent only after the syntactic comparison has failed to find a duplicate.
var NoDuplicateTypeConstituents = rule.Rule{
	Name: "@typescript-eslint/no-duplicate-type-constituents",

	// Every constituent's type is resolved to compare identities.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, isSettings := rule.OptionsAs[NoDuplicateTypeConstituentsOptions](options)
		if !isSettings {
			settings = DefaultNoDuplicateTypeConstituentsSettings()
		}

		// textOf renders a constituent the way upstream's `sourceCode.getText(previous)` does.
		//
		// Parentheses are stepped through for the same reason the span is: estree has no
		// parenthesized type node, so upstream's node for `(A | B)` is the inner union and its text
		// is `A | B` without the brackets. The message interpolates that text, so a port keeping
		// the brackets renders `duplicated with (A | B).` where upstream renders
		// `duplicated with A | B.` Measured on four corpus cases.
		//
		// The unwrap is a loop with its own nil check rather than a helper call, because
		// `ast.SkipParentheses` is for expressions and dereferences its argument; a parenthesized
		// type with no inner type is a shape error recovery can produce.
		textOf := func(node *ast.Node) string {
			inner := node
			for inner.Kind == ast.KindParenthesizedType {
				next := inner.AsParenthesizedTypeNode().Type
				if next == nil {
					break
				}
				inner = next
			}
			nodeRange := rule.TokenRange(ctx.SourceFile, inner)
			return ctx.SourceFile.Text()[nodeRange.Pos():nodeRange.End()]
		}

		// report builds the two ranges and emits the finding.
		//
		// `parent` is the union or intersection the constituent belongs to, which bounds the search
		// for a separator so that an operator belonging to an enclosing type is never taken.
		report := func(message rule.Message, constituent *ast.Node, parent *ast.Node) {
			constituentRange := rule.TokenRange(ctx.SourceFile, constituent)

			// The SPAN starts inside any parentheses while the REMOVAL starts outside them, and
			// that asymmetry is upstream's rather than a choice here. estree has no parenthesized
			// type node, so upstream's `constituentNode` for `(A | B)` IS the inner union and its
			// `loc.start` sits after the opening bracket, while the token run its fixer removes
			// still reaches back to the bracket. The result reads as an unbalanced fragment:
			// `A | (A | A)` underlines `A | A)`.
			//
			// Measured on the installed build across five parenthesized cases before this line
			// existed; the rule reported `(A | B)` and upstream reports `A | B)`.
			spanStart := constituentRange.Pos()
			for inner := constituent; inner.Kind == ast.KindParenthesizedType; {
				next := inner.AsParenthesizedTypeNode().Type
				if next == nil {
					break
				}
				spanStart = rule.TokenRange(ctx.SourceFile, next).Pos()
				inner = next
			}
			parentRange := rule.TokenRange(ctx.SourceFile, parent)
			text := ctx.SourceFile.Text()

			// The separator before the constituent, searched backwards within the parent. Upstream
			// takes the LAST such token before the node, which is the one immediately preceding it.
			separatorBefore, hasSeparatorBefore := lastSeparatorBefore(text, parentRange.Pos(), constituentRange.Pos())

			// The constituent's own end already covers its closing brackets, so there is nothing
			// to walk past.
			//
			// Upstream has to reassemble that run from tokens, because estree has no parenthesized
			// type node: its constituent for `(A | B)` is the inner union, whose range stops before
			// the bracket, so its fixer collects the brackets on either side separately. Our parser
			// gives `KindParenthesizedType` a range that INCLUDES both brackets, which is the same
			// asymmetry that makes the reported span start inside the opening one.
			//
			// A bracket-walking helper was written here first and a mutation disabling it survived
			// the entire suite. That was not a fixture gap: probed over the two parenthesized
			// corpus shapes, the constituent's token range already ends at 44 and 34, exactly the
			// byte after the closing bracket, so the walk consumed nothing on any input. Deleted
			// rather than kept, with the measurement recorded here because a reader meeting
			// upstream's token assembly will reasonably expect an equivalent to exist.
			// Named for what it means at the three sites below rather than inlined, since the
			// span and both removals all end at the same place and the reason they do is the
			// paragraph above.
			closeEnd := constituentRange.End()

			// The repair is a LIST of token removals rather than one contiguous span, and that
			// distinction is visible in every one of upstream's forty-eight recorded outputs.
			//
			// `type T = 1 | 1;` repairs to `type T = 1  ;` with TWO spaces: upstream removes the
			// separator token and the constituent token as separate edits, so the whitespace
			// between them survives. A single range from the separator through the constituent
			// removes that space too and produces one space, which is byte-wrong on all forty-two
			// single-constituent cases while looking entirely reasonable. This rule was written the
			// contiguous way first and every fix assertion failed the same way.
			fixes := []rule.Fix{}
			if hasSeparatorBefore {
				// The separator, and then the constituent with its brackets. Anything between them,
				// including a comment, is left where it was, which is why upstream's
				// comment-bearing case keeps its comment.
				fixes = append(fixes,
					rule.RemoveRange(core.NewTextRange(separatorBefore, separatorBefore+1)),
					rule.RemoveRange(core.NewTextRange(constituentRange.Pos(), closeEnd)))
			} else {
				// The constituent is first, so upstream takes the separator AFTER it instead. That
				// is why `type T = T | undefined` repairs by deleting `T` and its following `|`.
				fixes = append(fixes,
					rule.RemoveRange(core.NewTextRange(constituentRange.Pos(), closeEnd)))
				if separatorAfter, hasSeparatorAfter := firstSeparatorAfter(text, closeEnd, parentRange.End()); hasSeparatorAfter {
					fixes = append(fixes,
						rule.RemoveRange(core.NewTextRange(separatorAfter, separatorAfter+1)))
				}
			}

			// The reported span starts at the constituent and ends after its trailing brackets,
			// which is narrower than what the repair removes whenever a separator was taken.
			ctx.ReportRangeWithFixes(
				core.NewTextRange(spanStart, closeEnd),
				message,
				fixes...,
			)
		}

		// checkOne is upstream's `checkDuplicateRecursively`.
		var checkOne func(node *ast.Node, parent *ast.Node, wantKind ast.Kind,
			seenText map[string]*ast.Node, seenType map[*checker.Type]*ast.Node,
			onNewConstituent func(*checker.Type, *ast.Node, *ast.Node))

		checkOne = func(constituent *ast.Node, parent *ast.Node, wantKind ast.Kind,
			seenText map[string]*ast.Node, seenType map[*checker.Type]*ast.Node,
			onNewConstituent func(*checker.Type, *ast.Node, *ast.Node)) {

			// The syntactic comparison first, exactly as upstream orders it. This is the only test
			// that can see two structurally identical object literals, whose types are distinct
			// objects.
			text := textOf(constituent)
			if previous, isDuplicate := seenText[text]; isDuplicate {
				report(buildDuplicateConstituentMessage(kindWord(wantKind), textOf(previous)), constituent, parent)
				return
			}

			constituentType := ctx.TypeChecker.GetTypeAtLocation(constituent)
			if constituentType == nil {
				return
			}
			// An error type compares equal to every other error type, so a file with an unresolved
			// name would report every constituent against the first. Upstream skips it for the same
			// reason.
			if type_checking.IsIntrinsicErrorType(constituentType) {
				return
			}

			if previous, isDuplicate := seenType[constituentType]; isDuplicate {
				report(buildDuplicateConstituentMessage(kindWord(wantKind), textOf(previous)), constituent, parent)
				return
			}

			if onNewConstituent != nil {
				onNewConstituent(constituentType, constituent, parent)
			}
			seenType[constituentType] = constituent
			seenText[text] = constituent

			// Recurse into a nested same-kind constituent. The parenthesis is stepped through
			// because our parser keeps it and upstream's does not, so without this every
			// parenthesized nesting would be invisible.
			inner := constituent
			for inner.Kind == ast.KindParenthesizedType {
				next := inner.AsParenthesizedTypeNode().Type
				if next == nil {
					return
				}
				inner = next
			}
			if inner.Kind != wantKind {
				return
			}
			for _, nested := range constituentsOf(inner) {
				checkOne(nested, inner, wantKind, seenText, seenType, onNewConstituent)
			}
		}

		check := func(node *ast.Node, wantKind ast.Kind,
			onNewConstituent func(*checker.Type, *ast.Node, *ast.Node)) {
			seenText := map[string]*ast.Node{}
			seenType := map[*checker.Type]*ast.Node{}
			for _, constituent := range constituentsOf(node) {
				checkOne(constituent, node, wantKind, seenText, seenType, onNewConstituent)
			}
		}

		return rule.Listeners{
			ast.KindIntersectionType: func(node *ast.Node) {
				if ctx.TypeChecker == nil || settings.IgnoreIntersections {
					return
				}
				// A nested intersection is reached through its parent's recursion, so visiting it
				// again here would report every duplicate twice.
				if isNestedInSameKind(node, ast.KindIntersectionType) {
					return
				}
				check(node, ast.KindIntersectionType, nil)
			},

			ast.KindUnionType: func(node *ast.Node) {
				if ctx.TypeChecker == nil || settings.IgnoreUnions {
					return
				}
				if isNestedInSameKind(node, ast.KindUnionType) {
					return
				}

				check(node, ast.KindUnionType, func(constituentType *checker.Type, constituent *ast.Node, parent *ast.Node) {
					// The second message: an explicit `undefined` in the type of a parameter that
					// is already optional. The `?` has already widened the type, so the union member
					// adds nothing.
					if !type_checking.IsTypeFlagSet(constituentType, checker.TypeFlagsUndefined) {
						return
					}
					if !isOptionalParameterType(node) {
						return
					}
					report(buildUnnecessaryUndefinedMessage(), constituent, parent)
				})
			},
		}
	},
}

// constituentsOf returns the members of a union or intersection type node.
func constituentsOf(node *ast.Node) []*ast.Node {
	switch node.Kind {
	case ast.KindUnionType:
		if node.AsUnionTypeNode().Types == nil {
			return nil
		}
		return node.AsUnionTypeNode().Types.Nodes
	case ast.KindIntersectionType:
		if node.AsIntersectionTypeNode().Types == nil {
			return nil
		}
		return node.AsIntersectionTypeNode().Types.Nodes
	}
	return nil
}

// isNestedInSameKind reports whether this node is already inside one of its own kind.
//
// Parentheses are stepped through, because our parser keeps a node upstream's estree folds away and
// `A | (A | A)` would otherwise read as a top-level union rather than as a nested one.
func isNestedInSameKind(node *ast.Node, wantKind ast.Kind) bool {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedType {
		parent = parent.Parent
	}
	return parent != nil && parent.Kind == wantKind
}

// isOptionalParameterType reports whether this type node annotates an optional function parameter.
//
// Upstream walks the estree chain type annotation, identifier, function, and additionally checks
// that the identifier is among the function's own parameters. Our tree gives the parameter
// declaration directly as the type node's parent, and the parameter carries its own question token,
// so the same decision needs one step rather than three.
func isOptionalParameterType(typeNode *ast.Node) bool {
	parent := typeNode.Parent
	// The kind is compared directly rather than through `ast.IsParameterDeclaration`, whose name
	// suggests this question and whose body was not read; a kind test cannot be wrong about which
	// question it answers.
	if parent == nil || parent.Kind != ast.KindParameter {
		return false
	}
	parameter := parent.AsParameterDeclaration()
	if parameter.QuestionToken == nil || parameter.Type != typeNode {
		return false
	}
	// Upstream requires the enclosing construct to be a function or a function type. A parameter
	// declaration cannot occur anywhere else, so the test has no shape to exclude here and the
	// equivalent narrowing is the parameter-kind test above.
	return true
}

// lastSeparatorBefore finds the `|` or `&` immediately preceding a constituent, within its parent.
//
// The search is bounded by the parent's own start so that an operator belonging to an enclosing
// type can never be taken, which is upstream's `parent.range[0] <= token.range[0]` guard.
func lastSeparatorBefore(text string, lowerBound int, position int) (int, bool) {
	for index := position - 1; index >= lowerBound; index-- {
		if text[index] == '|' || text[index] == '&' {
			return index, true
		}
	}
	return 0, false
}

// firstSeparatorAfter finds the `|` or `&` following a constituent, within its parent.
func firstSeparatorAfter(text string, position int, upperBound int) (int, bool) {
	for index := position; index < upperBound && index < len(text); index++ {
		if text[index] == '|' || text[index] == '&' {
			return index, true
		}
	}
	return 0, false
}

// kindWord renders the word upstream interpolates into the duplicate message.
func kindWord(kind ast.Kind) string {
	if kind == ast.KindIntersectionType {
		return "Intersection"
	}
	return "Union"
}

func buildDuplicateConstituentMessage(unionOrIntersection string, previous string) rule.Message {
	return rule.Message{
		Id:          "duplicate",
		Description: unionOrIntersection + " type constituent is duplicated with " + previous + ".",
	}
}

func buildUnnecessaryUndefinedMessage() rule.Message {
	return rule.Message{
		Id:          "unnecessary",
		Description: "Explicit undefined is unnecessary on an optional parameter.",
	}
}
