package javascript

// print/ternary-old.js, and printTernary from print/ternary.js. experimentalTernaries is always false
// in our configs, so printTernary goes straight to printTernaryOld and the experimental printer in
// ternary.js is not ported.

// printTernary is upstream's printTernary. The experimental-ternaries branch is dropped; args is only
// read there, so it is unused here.
func printTernary(path *Path, options *Options, print PrintFunc, args *printArguments) Doc {
	return printTernaryOld(path, options, print)
}

// If we have nested conditional expressions, we want to print them in JSX mode
// if there's at least one JSXElement somewhere in the tree.
//
// A conditional expression chain like this should be printed in normal mode,
// because there aren't JSXElements anywhere in it:
//
// isA ? "A" : isB ? "B" : isC ? "C" : "Unknown";
//
// But a conditional expression chain like this should be printed in JSX mode,
// because there is a JSXElement in the last ConditionalExpression:
//
// isA ? "A" : isB ? "B" : isC ? "C" : <span className="warning">Unknown</span>;
//
// This type of ConditionalExpression chain is structured like this in the AST:
//
//	ConditionalExpression {
//	  test: ...,
//	  consequent: ...,
//	  alternate: ConditionalExpression {
//	    test: ...,
//	    consequent: ...,
//	    alternate: ConditionalExpression {
//	      test: ...,
//	      consequent: ...,
//	      alternate: ...,
//	    }
//	  }
//	}
func conditionalExpressionChainContainsJsx(node Node) bool {
	// Given this code:
	//
	// // Using a ConditionalExpression as the consequent is uncommon, but should
	// // be handled.
	// A ? B : C ? D : E ? F ? G : H : I
	//
	// which has this AST:
	//
	// ConditionalExpression {
	//   test: Identifier(A),
	//   consequent: Identifier(B),
	//   alternate: ConditionalExpression {
	//     test: Identifier(C),
	//     consequent: Identifier(D),
	//     alternate: ConditionalExpression {
	//       test: Identifier(E),
	//       consequent: ConditionalExpression {
	//         test: Identifier(F),
	//         consequent: Identifier(G),
	//         alternate: Identifier(H),
	//       },
	//       alternate: Identifier(I),
	//     }
	//   }
	// }
	//
	// We don't care about whether each node was the test, consequent, or alternate
	// We are only checking if there's any JSXElements inside.
	conditionalExpressions := []Node{node}
	for index := 0; index < len(conditionalExpressions); index++ {
		conditionalExpression := conditionalExpressions[index]
		for _, property := range []string{"test", "consequent", "alternate"} {
			node := conditionalExpression.Child(property)

			if isJsxElement(node) {
				return true
			}

			if node.Is("ConditionalExpression") {
				conditionalExpressions = append(conditionalExpressions, node)
			}
		}
	}

	return false
}

// printTernaryTest is upstream's printTernaryTest.
func printTernaryTest(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	isConditionalExpression := current.Is("ConditionalExpression")
	alternateNodePropertyName := "falseType"
	if isConditionalExpression {
		alternateNodePropertyName = "alternate"
	}

	parent := parentOf(path)

	var printed Doc
	if isConditionalExpression {
		printed = print("test", nil)
	} else {
		printed = concatIn(path, print("checkType", nil), " ", "extends", " ", print("extendsType", nil))
	}
	/**
	 *     a
	 *       ? b
	 *       : multiline
	 *         test
	 *         node
	 *       ^^ align(2)
	 *       ? d
	 *       : e
	 */
	if parent.Is(current.Type()) && parent.Child(alternateNodePropertyName) == current {
		return align(2, printed)
	}
	return printed
}

var ancestorNameMap = map[string]string{
	"AssignmentExpression": "right",
	"VariableDeclarator":   "init",
	"ReturnStatement":      "argument",
	"ThrowStatement":       "argument",
	"UnaryExpression":      "argument",
	"YieldExpression":      "argument",
	"AwaitExpression":      "argument",
}

// shouldExtraIndentForConditionalExpression is upstream's shouldExtraIndentForConditionalExpression.
func shouldExtraIndentForConditionalExpression(path *Path) bool {
	current := node(path)
	if !current.Is("ConditionalExpression") {
		return false
	}

	var parent Node
	child := current
	for ancestorCount := 0; parent == nil; ancestorCount++ {
		node, _ := path.GetParentNode(ancestorCount)
		// Upstream would throw reading `.type` past the root; Go's nil-safe predicates would loop
		// forever instead, so stop at the root.
		if node == nil {
			return false
		}

		if isChainElementWrapper(node) && node.Child("expression") == child ||
			isCallExpression(node) && node.Child("callee") == child ||
			isMemberExpression(node) && node.Child("object") == child {
			child = node
			continue
		}

		// Reached chain root

		if node.Is("NewExpression") && node.Child("callee") == child ||
			isBinaryCastExpression(node) && node.Child("expression") == child {
			parent, _ = path.GetParentNode(ancestorCount + 1)
			child = node
			// Same root guard as above: a nil parent would keep the loop going.
			if parent == nil {
				return false
			}
		} else {
			parent = node
		}
	}

	// Do not add indent to direct `ConditionalExpression`
	if child == current {
		return false
	}

	// upstream's `parent[ancestorNameMap.get(parent.type)] === child`: a type missing from the map
	// reads parent[undefined], which is never child.
	name, inMap := ancestorNameMap[parent.Type()]
	return inMap && parent.Child(name) == child
}

// printTernaryOld is upstream's printTernaryOld: the shared logic for ternary operators, namely
// ConditionalExpression, ConditionalTypeAnnotation and TSConditionalType.
func printTernaryOld(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	isConditionalExpression := current.Is("ConditionalExpression")
	consequentNodePropertyName := "trueType"
	alternateNodePropertyName := "falseType"
	testNodePropertyNames := []string{"checkType", "extendsType"}
	if isConditionalExpression {
		consequentNodePropertyName = "consequent"
		alternateNodePropertyName = "alternate"
		testNodePropertyNames = []string{"test"}
	}
	consequentNode := current.Child(consequentNodePropertyName)
	alternateNode := current.Child(alternateNodePropertyName)
	var parts []any

	// We print a ConditionalExpression in either "JSX mode" or "normal mode".
	// See `tests/format/jsx/conditional-expression.js` for more info.
	jsxMode := false
	parent := parentOf(path)
	isParentTest := false
	if parent.Is(current.Type()) {
		for _, prop := range testNodePropertyNames {
			if parent.Child(prop) == current {
				isParentTest = true
				break
			}
		}
	}
	forceNoIndent := parent.Is(current.Type()) && !isParentTest

	// Find the outermost non-ConditionalExpression parent, and the outermost
	// ConditionalExpression parent. We'll use these to determine if we should
	// print in JSX mode.
	var currentParent Node
	var previousParent Node
	// upstream's `testNodePropertyNames.every((prop) => currentParent[prop] !== previousParent)`.
	everyPropertyIsNot := func() bool {
		for _, prop := range testNodePropertyNames {
			if currentParent.Child(prop) == previousParent {
				return false
			}
		}
		return true
	}
	i := 0
	// upstream's do { ... } while (...).
	for {
		previousParent = currentParent
		if previousParent == nil {
			previousParent = current
		}
		currentParent, _ = path.GetParentNode(i)
		i++
		if !(currentParent != nil && currentParent.Is(current.Type()) &&
			everyPropertyIsNot()) {
			break
		}
	}
	firstNonConditionalParent := currentParent
	if firstNonConditionalParent == nil {
		firstNonConditionalParent = parent
	}
	lastConditionalParent := previousParent

	if isConditionalExpression &&
		(isJsxElement(current.Child(testNodePropertyNames[0])) ||
			isJsxElement(consequentNode) ||
			isJsxElement(alternateNode) ||
			conditionalExpressionChainContainsJsx(lastConditionalParent)) {
		jsxMode = true
		forceNoIndent = true

		// Even though they don't need parens, we wrap (almost) everything in
		// parens when using ?: within JSX, because the parens are analogous to
		// curly braces in an if statement.
		wrap := func(document Doc) Doc {
			return concatIn(path, ifBreak("(", ""),
				indentIn(path, concatIn(path, softline, document)),
				softline,
				ifBreak(")", ""),
			)
		}

		// The only things we don't wrap are:
		// * Nested conditional expressions in alternates
		// * null
		// * undefined
		isNil := func(node Node) bool {
			// upstream's `node.value === null`: the Go tree also leaves value nil on bigint and regex
			// Literals, which upstream gives a BigInt or RegExp value, so those are ruled out here.
			return node.Is("NullLiteral") ||
				(node.Is("Literal") && node.Get("value") == nil && node.Get("regex") == nil &&
					!node.Truthy("bigint")) ||
				(node.Is("Identifier") && node.String("name") == "undefined")
		}

		var printedConsequent Doc
		if isNil(consequentNode) {
			printedConsequent = print(consequentNodePropertyName, nil)
		} else {
			printedConsequent = wrap(print(consequentNodePropertyName, nil))
		}
		var printedAlternate Doc
		if alternateNode.Is(current.Type()) || isNil(alternateNode) {
			printedAlternate = print(alternateNodePropertyName, nil)
		} else {
			printedAlternate = wrap(print(alternateNodePropertyName, nil))
		}
		parts = append(parts,
			" ? ",
			printedConsequent,
			" : ",
			printedAlternate,
		)
	} else {
		/*
		   This does not mean to indent, but make the doc aligned with the first character after `? ` or `: `,
		   so we use `2` instead of `options.tabWidth` here.

		   ```js
		   test
		    ? {
		        consequent
		      }
		    : alternate
		   ```

		   instead of

		   ```js
		   test
		    ? {
		      consequent
		    }
		    : alternate
		   ```
		*/
		printBranch := func(nodePropertyName string) Doc {
			if settingsOf(options).UseTabs {
				return indentIn(path, print(nodePropertyName, nil))
			}
			return align(2, print(nodePropertyName, nil))
		}
		// normal mode
		consequentIsTernary := consequentNode.Is(current.Type())
		var openParen, closeParen Doc = emptyDoc, emptyDoc
		if consequentIsTernary {
			openParen = ifBreak("", "(")
		}
		printedConsequent := printBranch(consequentNodePropertyName)
		if consequentIsTernary {
			closeParen = ifBreak("", ")")
		}
		part := concatIn(path, line,
			"? ",
			openParen,
			printedConsequent,
			closeParen,
			line,
			": ",
			printBranch(alternateNodePropertyName),
		)
		if !parent.Is(current.Type()) ||
			parent.Child(alternateNodePropertyName) == current ||
			isParentTest {
			parts = append(parts, part)
		} else if settingsOf(options).UseTabs {
			parts = append(parts, dedent(indentIn(path, part)))
		} else {
			parts = append(parts, align(max(0, settingsOf(options).TabWidth-2), part))
		}
	}

	maybeGroup := func(document Doc) Doc {
		if parent == firstNonConditionalParent {
			return groupIn(path, document)
		}
		return document
	}

	// Break the closing paren to keep the chain right after it:
	// (a
	//   ? b
	//   : c
	// ).call()
	// upstream's `parent.type === "NGPipeExpression" && parent.left === node` is Angular-only, dropped.
	breakClosingParen := !jsxMode &&
		isMemberExpression(parent) &&
		!parent.Bool("computed")

	shouldExtraIndent := shouldExtraIndentForConditionalExpression(path)

	printedTest := printTernaryTest(path, options, print)
	var printedParts Doc
	if forceNoIndent {
		printedParts = concatIn(path, parts...)
	} else {
		printedParts = indentIn(path, concatIn(path, parts...))
	}
	var closing Doc = emptyDoc
	if isConditionalExpression && breakClosingParen && !shouldExtraIndent {
		closing = softline
	}
	result := maybeGroup(concatIn(path, printedTest, printedParts, closing))

	if isParentTest || shouldExtraIndent {
		return groupIn(path, concatIn(path, indentIn(path, concatIn(path, softline, result)), softline))
	}
	return result
}
