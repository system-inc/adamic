package javascript

import (
	"regexp"
	"slices"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// print/member-chain.js.

// printedNode is upstream's PrintedNode typedef. Upstream's optional shouldInline is never read in
// this file, so it is not carried.
type printedNode struct {
	node                 Node
	printed              Doc
	hasTrailingEmptyLine bool
}

// memberChainFactory is the regular expression upstream's isFactory tests.
var memberChainFactory = regexp.MustCompile(`^[A-Z]|^[$_]+$`)

// We detect calls on member expressions specially to format a
// common pattern better. The pattern we are looking for is this:
//
// arr
//   .map(x => x + 1)
//   .filter(x => x > 10)
//   .some(x => x % 2)
//
// The way it is structured in the AST is via a nested sequence of
// MemberExpression and CallExpression. We need to traverse the AST
// and make groups out of it to print it in the desired way.
/*
- `BindExpression`
- `MemberExpression`
- `OptionalMemberExpression`
*/

// printMemberChain is upstream's printMemberChain.
func printMemberChain(path *Path, options *Options, print PrintFunc) Doc {
	statementParent := parentOf(path)
	if statementParent.Is("ChainExpression") {
		statementParent = grandparentOf(path)
	}
	isExpressionStatement := statementParent.Is("ExpressionStatement")

	// The first phase is to linearize the AST by traversing it down.
	//
	//   a().b()
	// has the following AST structure:
	//   CallExpression(MemberExpression(CallExpression(Identifier)))
	// and we transform it into
	//   [Identifier, CallExpression, MemberExpression, CallExpression]
	var printedNodes []printedNode

	// unshift is upstream's printedNodes.unshift. Nothing reads printedNodes until the walk down is done, so
	// each node is appended and the list reversed once afterwards, rather than copied whole on every unshift:
	// a new slice per node was 0.1M of the printer's allocations a pass on ahra (#fyw36kf).
	unshift := func(printed printedNode) {
		printedNodes = append(printedNodes, printed)
	}

	// Here we try to retain one typed empty line after each call expression or
	// the first group whether it is in parentheses or not
	shouldInsertEmptyLineAfter := func(node Node) bool {
		originalText := options.OriginalText
		nextCharIndex := getNextNonSpaceNonCommentCharacterIndex(originalText, locEnd(node))
		nextChar := charAt(originalText, nextCharIndex)

		// if it is cut off by a parenthesis, we only account for one typed empty
		// line after that parenthesis
		if nextChar == ")" {
			// upstream's `nextCharIndex !== false`: the core's false position is notFound.
			return nextCharIndex != notFound &&
				isNextLineEmpty(originalText, nextCharIndex+1)
		}

		return isNextLineEmptyAfter(node, options)
	}

	// rec returns nothing upstream; struct{} satisfies call's result type.
	var rec func(path *Path) struct{}
	rec = func(path *Path) struct{} {
		current := node(path)

		if isCallExpression(current) &&
			(isMemberish(current.Child("callee")) || isCallExpression(current.Child("callee"))) &&
			!needsParentheses(path, options) {
			hasTrailingEmptyLine := shouldInsertEmptyLineAfter(current)
			var trailingLine Doc = emptyDoc
			if hasTrailingEmptyLine {
				trailingLine = hardline
			}
			unshift(printedNode{
				node:                 current,
				hasTrailingEmptyLine: hasTrailingEmptyLine,
				printed: concatIn(path, printing.PrintComments(
					path,
					concatIn(path, printOptionalToken(path),
						print("typeArguments", nil),
						printCallArguments(path, options, print),
					),
					options,
					nil,
				),
					trailingLine,
				),
			})
			call(path, rec, "callee")
		} else if isMemberish(current) && !needsParentheses(path, options) {
			// upstream's `isMemberExpression(node) ? printMemberLookup(...) : printBindExpressionCallee(...)`:
			// BindExpression is Babel-only, so a TypeScript tree always takes the member lookup.
			unshift(printedNode{
				node:    current,
				printed: printing.PrintComments(path, printMemberLookup(path, options, print), options, nil),
			})
			call(path, rec, "object")
		} else if current.Is("ChainExpression") && !needsParentheses(path, options) {
			call(path, rec, "expression")
		} else if current.Is("TSNonNullExpression") && !needsParentheses(path, options) {
			unshift(printedNode{
				node:    current,
				printed: printing.PrintComments(path, concatIn(path, "!"), options, nil),
			})
			call(path, rec, "expression")
		} else {
			unshift(printedNode{
				node:    current,
				printed: print(nil, nil),
			})
		}
		return struct{}{}
	}
	// Note: the comments of the root node have already been printed, so we
	// need to extract this first call without printing them as they would
	// if handled inside of the recursive call.
	current := node(path)
	unshift(printedNode{
		node: current,
		printed: concatIn(path, printOptionalToken(path),
			print("typeArguments", nil),
			printCallArguments(path, options, print),
		),
	})

	if current.Child("callee") != nil {
		call(path, rec, "callee")
	}
	slices.Reverse(printedNodes)

	// Once we have a linear list of printed nodes, we want to create groups out
	// of it.
	//
	//   a().b.c().d().e
	// will be grouped as
	//   [
	//     [Identifier, CallExpression],
	//     [MemberExpression, MemberExpression, CallExpression],
	//     [MemberExpression, CallExpression],
	//     [MemberExpression],
	//   ]
	// so that we can print it as
	//   a()
	//     .b.c()
	//     .d()
	//     .e

	// The first group is the first node followed by
	//   - as many CallExpression as possible
	//       < fn()()() >.something()
	//   - as many array accessors as possible
	//       < fn()[0][1][2] >.something()
	//   - then, as many MemberExpression as possible but the last one
	//       < this.items >.something()
	groups := make([][]printedNode, 0, len(printedNodes))
	currentGroup := []printedNode{printedNodes[0]}
	i := 1
	for ; i < len(printedNodes); i++ {
		if printedNodes[i].node.Is("TSNonNullExpression") ||
			printedNodes[i].node.Is("ChainExpression") ||
			isCallExpression(printedNodes[i].node) ||
			(isMemberExpression(printedNodes[i].node) &&
				printedNodes[i].node.Bool("computed") &&
				isNumericLiteral(printedNodes[i].node.Child("property"))) {
			currentGroup = append(currentGroup, printedNodes[i])
		} else {
			break
		}
	}
	if !isCallExpression(printedNodes[0].node) {
		for ; i+1 < len(printedNodes); i++ {
			if isMemberish(printedNodes[i].node) &&
				isMemberish(printedNodes[i+1].node) {
				currentGroup = append(currentGroup, printedNodes[i])
			} else {
				break
			}
		}
	}
	groups = append(groups, currentGroup)
	currentGroup = nil

	// Then, each following group is a sequence of MemberExpression followed by
	// a sequence of CallExpression. To compute it, we keep adding things to the
	// group until we has seen a CallExpression in the past and reach a
	// MemberExpression
	hasSeenCallExpression := false
	for ; i < len(printedNodes); i++ {
		if hasSeenCallExpression && isMemberish(printedNodes[i].node) {
			// [0] should be appended at the end of the group instead of the
			// beginning of the next one
			if printedNodes[i].node.Bool("computed") &&
				isNumericLiteral(printedNodes[i].node.Child("property")) {
				currentGroup = append(currentGroup, printedNodes[i])
				continue
			}

			groups = append(groups, currentGroup)
			currentGroup = nil
			hasSeenCallExpression = false
		}

		if isCallExpression(printedNodes[i].node) ||
			printedNodes[i].node.Is("ImportExpression") {
			hasSeenCallExpression = true
		}
		currentGroup = append(currentGroup, printedNodes[i])

		if hasComment(printedNodes[i].node, commentTrailing, nil) {
			groups = append(groups, currentGroup)
			currentGroup = nil
			hasSeenCallExpression = false
		}
	}
	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}

	// There are cases like Object.keys(), Observable.of(), _.values() where
	// they are the subject of all the chained calls and therefore should
	// be kept on the same line:
	//
	//   Object.keys(items)
	//     .filter(x => x)
	//     .map(x => x)
	//
	// In order to detect those cases, we use an heuristic: if the first
	// node is an identifier with the name starting with a capital
	// letter or just a sequence of _$. The rationale is that they are
	// likely to be factories.
	isFactory := func(name string) bool {
		return memberChainFactory.MatchString(name)
	}

	// In case the Identifier is shorter than tab width, we can keep the
	// first call in a single line, if it's an ExpressionStatement.
	//
	//   d3.scaleLinear()
	//     .domain([0, 100])
	//     .range([0, width]);
	//
	// upstream's `name.length` is UTF-16 code units of an identifier name, not a display width, so
	// it is utf16Length rather than getStringWidth.
	isShort := func(name string) bool {
		return utf16Length(name) <= settingsOf(options).TabWidth
	}

	shouldNotWrap := func(groups [][]printedNode) bool {
		// upstream's `groups[1][0]?.node.computed`
		hasComputed := len(groups[1]) > 0 && groups[1][0].node.Bool("computed")

		if len(groups[0]) == 1 {
			firstNode := groups[0][0].node
			return firstNode.Is("ThisExpression") ||
				(firstNode.Is("Identifier") &&
					(isFactory(firstNode.String("name")) ||
						(isExpressionStatement && isShort(firstNode.String("name"))) ||
						hasComputed))
		}

		lastNode := groups[0][len(groups[0])-1].node
		return isMemberExpression(lastNode) &&
			lastNode.Child("property").Is("Identifier") &&
			(isFactory(lastNode.Child("property").String("name")) || hasComputed)
	}

	shouldMerge := len(groups) >= 2 &&
		!hasComment(groups[1][0].node, 0, nil) &&
		shouldNotWrap(groups)

	printGroup := func(printedGroup []printedNode) Doc {
		parts := partsIn(path, len(printedGroup))
		for index, tuple := range printedGroup {
			parts[index] = tuple.printed
		}
		return sequenceIn(path, parts)
	}

	printIndentedGroup := func(groups [][]printedNode) Doc {
		/* c8 ignore next 3 */
		if len(groups) == 0 {
			return emptyDoc
		}
		printed := partsIn(path, len(groups))
		for index, printedGroup := range groups {
			printed[index] = printGroup(printedGroup)
		}
		return indentIn(path, concatIn(path, hardline, join(hardline, printed)))
	}

	printedGroups := partsIn(path, len(groups))
	for index, printedGroup := range groups {
		printedGroups[index] = printGroup(printedGroup)
	}
	oneLine := sequenceIn(path, printedGroups)

	cutoff := 2
	if shouldMerge {
		cutoff = 3
	}
	// The groups partition printedNodes, so flattening them refills a list of the same length.
	flatGroups := make([]printedNode, 0, len(printedNodes))
	for _, printedGroup := range groups {
		flatGroups = append(flatGroups, printedGroup...)
	}

	nodeHasComment := false
	// upstream's flatGroups.slice(1, -1)
	for index := 1; index < len(flatGroups)-1; index++ {
		if hasComment(flatGroups[index].node, commentLeading, nil) {
			nodeHasComment = true
			break
		}
	}
	if !nodeHasComment {
		// upstream's flatGroups.slice(0, -1)
		for index := 0; index < len(flatGroups)-1; index++ {
			if hasComment(flatGroups[index].node, commentTrailing, nil) {
				nodeHasComment = true
				break
			}
		}
	}
	if !nodeHasComment {
		nodeHasComment = cutoff < len(groups) &&
			hasComment(groups[cutoff][0].node, commentLeading, nil)
	}

	// If we only have a single `.`, we shouldn't do anything fancy and just
	// render everything concatenated together.
	everyGroupLacksTrailingEmptyLine := true
	for _, printedGroup := range groups {
		if printedGroup[len(printedGroup)-1].hasTrailingEmptyLine {
			everyGroupLacksTrailingEmptyLine = false
			break
		}
	}
	if len(groups) <= cutoff &&
		!nodeHasComment &&
		everyGroupLacksTrailingEmptyLine {
		if isLongCurriedCallExpression(path) {
			return oneLine
		}
		return groupIn(path, oneLine)
	}

	// Find out the last node in the first group and check if it has an
	// empty line after
	lastGroupBeforeIndent := groups[0]
	if shouldMerge {
		lastGroupBeforeIndent = groups[1]
	}
	lastNodeBeforeIndent := lastGroupBeforeIndent[len(lastGroupBeforeIndent)-1].node
	shouldHaveEmptyLineBeforeIndent := !isCallExpression(lastNodeBeforeIndent) &&
		shouldInsertEmptyLineAfter(lastNodeBeforeIndent)

	var mergedGroup Doc = emptyDoc
	if shouldMerge {
		// upstream's groups.slice(1, 2).map(printGroup)
		mergedGroup = concatIn(path, printGroup(groups[1]))
	}
	var emptyLineBeforeIndent Doc = emptyDoc
	if shouldHaveEmptyLineBeforeIndent {
		emptyLineBeforeIndent = hardline
	}
	indentedGroups := groups[1:]
	if shouldMerge {
		indentedGroups = groups[2:]
	}
	expanded := concatIn(path, printGroup(groups[0]),
		mergedGroup,
		emptyLineBeforeIndent,
		printIndentedGroup(indentedGroups),
	)

	var callExpressions []Node
	for _, printed := range printedNodes {
		if isCallExpression(printed.node) {
			callExpressions = append(callExpressions, printed.node)
		}
	}

	lastGroupWillBreakAndOtherCallsHaveFunctionArguments := func() bool {
		lastGroup := groups[len(groups)-1]
		lastGroupNode := lastGroup[len(lastGroup)-1].node
		lastGroupDoc := printedGroups[len(printedGroups)-1]
		if !isCallExpression(lastGroupNode) || !doc.WillBreak(lastGroupDoc) {
			return false
		}
		// upstream's callExpressions.slice(0, -1)
		for index := 0; index < len(callExpressions)-1; index++ {
			for _, argument := range callExpressions[index].List("arguments") {
				if isFunctionOrArrowExpression(argument) {
					return true
				}
			}
		}
		return false
	}

	// upstream's callExpressions.some((expr) => expr.arguments.some((arg) => !isSimpleCallArgument(arg)))
	someCallHasNonSimpleArgument := func() bool {
		for _, expression := range callExpressions {
			for _, argument := range expression.List("arguments") {
				if !isSimpleCallArgument(argument, 2) {
					return true
				}
			}
		}
		return false
	}

	// upstream's printedGroups.slice(0, -1).some(willBreak)
	someGroupButTheLastWillBreak := func() bool {
		for index := 0; index < len(printedGroups)-1; index++ {
			if doc.WillBreak(printedGroups[index]) {
				return true
			}
		}
		return false
	}

	var result Doc

	// We don't want to print in one line if at least one of these conditions occurs:
	//  * the chain has comments,
	//  * the chain is an expression statement and all the arguments are literal-like ("fluent configuration" pattern),
	//  * the chain is longer than 2 calls and has non-trivial arguments or more than 2 arguments in any call but the first one,
	//  * any group but the last one has a hard line,
	//  * the last call's arguments have a hard line and other calls have non-trivial arguments.
	if nodeHasComment ||
		(len(callExpressions) > 2 && someCallHasNonSimpleArgument()) ||
		someGroupButTheLastWillBreak() ||
		lastGroupWillBreakAndOtherCallsHaveFunctionArguments() {
		result = groupIn(path, expanded)
	} else {
		var breakParentDoc Doc = emptyDoc
		// We only need to check `oneLine` because if `expanded` is chosen
		// that means that the parent group has already been broken
		// naturally
		if doc.WillBreak(oneLine) || shouldHaveEmptyLineBeforeIndent {
			breakParentDoc = breakParent
		}
		result = concatIn(path, breakParentDoc,
			conditionalGroup([]Doc{oneLine, expanded}, doc.GroupOptions{}),
		)
	}

	// upstream's label({ memberChain: true }, result): Go labels are strings.
	return label("member-chain", result)
}
