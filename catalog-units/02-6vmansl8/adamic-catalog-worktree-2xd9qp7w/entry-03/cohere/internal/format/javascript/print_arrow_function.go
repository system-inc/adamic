package javascript

import (
	"slices"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// print/arrow-function.js.

// In order to avoid confusion between
// a => a ? a : a
// a <= a ? a : a
//
// shouldAddParensIfNotBreak is upstream's shouldAddParensIfNotBreak. Upstream memoizes it in a WeakMap
// (shouldAddParensIfNotBreakCache); the result depends only on the node, so the Go computes it each
// time.
func shouldAddParensIfNotBreak(node Node) bool {
	return node.Is("ConditionalExpression") &&
		!startsWithNoLookaheadToken(
			node,
			func(node Node) bool { return node.Is("ObjectExpression") },
		)
}

// We handle sequence expressions as the body of arrows specially,
// so that the required parentheses end up on their own lines.
//
// shouldAlwaysAddParens is upstream's shouldAlwaysAddParens.
func shouldAlwaysAddParens(node Node) bool { return node.Is("SequenceExpression") }

/*
- `ArrowFunctionExpression`
*/
// printArrowFunction is upstream's printArrowFunction. Upstream defaults args to {}; a nil args is
// replaced with an empty one, and that empty one is what reaches print("body", args), as upstream's {}
// does.
func printArrowFunction(path *Path, options *Options, print PrintFunc, args *printArguments) Doc {
	if args == nil {
		args = &printArguments{}
	}
	var signatureDocs []Doc
	var bodyDoc Doc
	var bodyComments []Doc
	shouldBreakChain := false
	shouldPrintAsChain :=
		!args.expandLastArg && node(path).Child("body").Is("ArrowFunctionExpression")
	var functionBody Node

	// upstream's immediately invoked `(function rec() { ... })()`, with path.call(rec, "body").
	var rec func(path *Path)
	rec = func(path *Path) {
		current := node(path)
		signatureDoc := printArrowFunctionSignature(
			path,
			options,
			print,
			args,
		)
		if len(signatureDocs) == 0 {
			signatureDocs = append(signatureDocs, signatureDoc)
		} else {
			leading, trailing := printing.PrintCommentsSeparately(path, options, nil)
			signatureDocs = append(signatureDocs, concatIn(path, leading, signatureDoc))
			bodyComments = append([]Doc{trailing}, bodyComments...)
		}

		if shouldPrintAsChain {
			shouldBreakChain = shouldBreakChain ||
				// Always break the chain if:
				(current.Truthy("returnType") && len(getFunctionParameters(current)) > 0) ||
				current.Truthy("typeParameters") ||
				slices.ContainsFunc(getFunctionParameters(current), func(param Node) bool {
					return !param.Is("Identifier")
				})
		}

		if !shouldPrintAsChain || !current.Child("body").Is("ArrowFunctionExpression") {
			bodyDoc = print("body", args)
			functionBody = current.Child("body")
		} else {
			call(path, func(path *Path) any { rec(path); return nil }, "body")
		}
	}
	rec(path)

	// We want to always keep these types of nodes on the same line
	// as the arrow.
	shouldPutBodyOnSameLine :=
		!hasLeadingOwnLineComment(originalText(options), functionBody) &&
			(shouldAlwaysAddParens(functionBody) ||
				mayBreakAfterShortPrefix(functionBody, bodyDoc, options) ||
				(!shouldBreakChain && shouldAddParensIfNotBreak(functionBody)))

	isCallee := keyOf(path) == "callee" && isCallLikeExpression(parentOf(path))
	chainGroupID := newGroupID("arrow-chain")

	signaturesDoc := printArrowFunctionSignatures(path, args, signatureDocs, shouldBreakChain)
	shouldBreakSignatures := false
	shouldIndentSignatures := false
	shouldPrintSoftlineInIndent := false
	if shouldPrintAsChain &&
		(isCallee ||
			// isAssignmentRhs
			args.assignmentLayout != "") {
		shouldIndentSignatures = true
		// If the arrow function has a leading line comment, there should be a hardline above it
		// so we should not print a softline in indent call
		// https://github.com/prettier/prettier/issues/16067
		//
		// upstream's `CommentCheckFlags.Leading & CommentCheckFlags.Line` is a bitwise and of two
		// distinct flags, which is 0, so this is true when the node has no comment at all. Ported as is.
		shouldPrintSoftlineInIndent = !hasComment(
			node(path),
			commentLeading&commentLine,
			nil,
		)
		shouldBreakSignatures =
			args.assignmentLayout == "chain-tail-arrow-chain" ||
				(isCallee && !shouldPutBodyOnSameLine)
	}

	bodyDoc = printArrowFunctionBody(path, options, args, bodyDoc, bodyComments, functionBody, shouldPutBodyOnSameLine)

	var signatures Doc = signaturesDoc
	if shouldIndentSignatures {
		var softlineDoc Doc = emptyDoc
		if shouldPrintSoftlineInIndent {
			softlineDoc = softline
		}
		signatures = indentIn(path, concatIn(path, softlineDoc, signaturesDoc))
	}
	var body Doc
	if shouldPrintAsChain {
		body = indentIfBreak(bodyDoc, chainGroupID, false)
	} else {
		body = groupIn(path, bodyDoc)
	}
	var trailing Doc = emptyDoc
	if shouldPrintAsChain && isCallee {
		trailing = ifBreakWithGroup(softline, "", chainGroupID)
	}
	return groupIn(path, concatIn(path, groupWithIn(path, signatures, doc.GroupOptions{ShouldBreak: shouldBreakSignatures, ID: chainGroupID}),
		" =>",
		body,
		trailing,
	))
}

// printArrowFunctionSignature is upstream's printArrowFunctionSignature.
func printArrowFunctionSignature(path *Path, options *Options, print PrintFunc, args *printArguments) Doc {
	current := node(path)
	var parts []any

	if current.Truthy("async") {
		parts = append(parts, "async ")
	}

	if shouldPrintParamsWithoutParens(path, options) {
		parts = append(parts, print([]any{"params", 0}, nil))
	} else {
		shouldExpandParameters := args.expandLastArg || args.expandFirstArg
		returnTypeDoc := printReturnType(path, print)
		if shouldExpandParameters {
			if doc.WillBreak(returnTypeDoc) {
				panic(argExpansionBailout{})
			}
			returnTypeDoc = groupIn(path, doc.RemoveLines(returnTypeDoc))
		}
		parts = append(parts,
			groupIn(path, concatIn(path, printFunctionParameters(
				path,
				options,
				print,
				shouldExpandParameters,
				/* shouldPrintTypeParameters */ true,
			),
				returnTypeDoc,
			)),
		)
	}

	dangling := printing.PrintDanglingComments(path, options, printing.DanglingOptions[Node]{
		Marker: "commentBeforeArrow",
	})
	if !isEmptyString(dangling) {
		parts = append(parts, " ", dangling)
	}
	return concatIn(path, parts...)
}

// mayBreakAfterShortPrefix is upstream's mayBreakAfterShortPrefix.
//
// upstream's `bodyDoc.label?.hug !== false && (bodyDoc.label?.embed || ...)`: Go labels are strings,
// and upstream's embed label is `{ embed: true, ...doc.label }` from language-js/embed/index.js, read
// here as the label "embed". Only the html embed sets `hug: false`, and no Go label carries it, so the
// hug test always passes.
func mayBreakAfterShortPrefix(functionBody Node, bodyDoc Doc, options *Options) bool {
	return isArrayExpression(functionBody) ||
		isObjectExpression(functionBody) ||
		functionBody.Is("ArrowFunctionExpression") ||
		functionBody.Is("DoExpression") ||
		functionBody.Is("BlockStatement") ||
		isJsxElement(functionBody) ||
		(labelOf(bodyDoc) == "embed" ||
			isTemplateOnItsOwnLine(functionBody, originalText(options)))
}

// printArrowFunctionSignatures is upstream's printArrowFunctionSignatures. Upstream's options object
// { signatureDocs, shouldBreak } is two parameters.
func printArrowFunctionSignatures(
	path *Path,
	args *printArguments,
	signatureDocs []Doc,
	shouldBreak bool,
) Doc {
	if len(signatureDocs) == 1 {
		return signatureDocs[0]
	}

	parent := parentOf(path)
	key := keyOf(path)
	if (key != "callee" && isCallLikeExpression(parent)) ||
		isBinaryish(parent) {
		return groupWithIn(path, concatIn(path, signatureDocs[0],
			" =>",
			indentIn(path, concatIn(path, line, join(concatIn(path, " =>", line), signatureDocs[1:]))),
		),
			doc.GroupOptions{ShouldBreak: shouldBreak},
		)
	}

	if (key == "callee" && isCallLikeExpression(parent)) ||
		// isAssignmentRhs
		args.assignmentLayout != "" {
		return groupWithIn(path, join(concatIn(path, " =>", line), signatureDocs), doc.GroupOptions{ShouldBreak: shouldBreak})
	}

	return groupWithIn(path, indentIn(path, join(concatIn(path, " =>", line), signatureDocs)), doc.GroupOptions{ShouldBreak: shouldBreak})
}

// printArrowFunctionBody is upstream's printArrowFunctionBody. Upstream's options object { bodyDoc,
// bodyComments, functionBody, shouldPutBodyOnSameLine } is four parameters.
func printArrowFunctionBody(
	path *Path,
	options *Options,
	args *printArguments,
	bodyDoc Doc,
	bodyComments []Doc,
	functionBody Node,
	shouldPutBodyOnSameLine bool,
) Doc {
	current := node(path)
	parent := parentOf(path)

	var trailingComma Doc = emptyDoc
	if args.expandLastArg {
		trailingComma = printTrailingComma(options, "all")
	}

	// if the arrow function is expanded as last argument, we are adding a
	// level of indentation and need to add a softline to align the closing )
	// with the opening (, or if it's inside a JSXExpression (e.g. an attribute)
	// we should align the expression's closing } with the line with the opening {.
	var trailingSpace Doc = emptyDoc
	if (args.expandLastArg || parent.Is("JSXExpressionContainer")) &&
		!hasAnyComment(current) {
		trailingSpace = softline
	}

	if shouldPutBodyOnSameLine && shouldAddParensIfNotBreak(functionBody) {
		return concatIn(path, " ",
			groupIn(path, concatIn(path, ifBreak("", "("),
				indentIn(path, concatIn(path, softline, bodyDoc)),
				ifBreak("", ")"),
				trailingComma,
				trailingSpace,
			)),
			bodyComments,
		)
	}

	if shouldPutBodyOnSameLine {
		return concatIn(path, " ", bodyDoc, bodyComments)
	}
	return concatIn(path, indentIn(path, concatIn(path, line, bodyDoc, bodyComments)), trailingComma, trailingSpace)
}
