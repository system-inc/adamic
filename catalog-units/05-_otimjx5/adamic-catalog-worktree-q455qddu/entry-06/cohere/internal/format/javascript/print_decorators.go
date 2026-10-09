package javascript

import "github.com/system-inc/cohere/internal/format/estree"

// print/decorators.js and print/ignored.js.

// printClassMemberDecorators is upstream's printClassMemberDecorators.
func printClassMemberDecorators(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	separator := line
	if hasNewlineBetweenOrAfterDecorators(current, options) {
		separator = hardline
	}
	return groupIn(path, concatIn(path, join(line, printAll(path, print, "decorators")), separator))
}

// printDecoratorsBeforeExport is upstream's printDecoratorsBeforeExport.
func printDecoratorsBeforeExport(path *Path, options *Options, print PrintFunc) Doc {
	if !hasDecoratorsBeforeExport(node(path)) {
		return emptyDoc
	}
	return concatIn(path, join(hardline, printAll(path, print, "declaration", "decorators")), hardline)
}

// printDecorators is upstream's printDecorators.
func printDecorators(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	parent := parentOf(path)
	decorators := current.List("decorators")
	if len(decorators) == 0 || hasDecoratorsBeforeExport(parent) || isIgnored(path) {
		return emptyDoc
	}
	shouldBreak := current.Is("ClassExpression", "ClassDeclaration") || hasNewlineBetweenOrAfterDecorators(current, options)
	var leading Doc = emptyDoc
	switch {
	case keyOf(path) == "declaration" && isExportDeclaration(parent):
		leading = hardline
	case shouldBreak:
		leading = breakParent
	}
	return concatIn(path, leading, join(line, printAll(path, print, "decorators")), line)
}

// hasNewlineBetweenOrAfterDecorators is upstream's hasNewlineBetweenOrAfterDecorators.
func hasNewlineBetweenOrAfterDecorators(node Node, options *Options) bool {
	for _, decorator := range node.List("decorators") {
		if hasNewline(options.OriginalText, locEnd(decorator)) {
			return true
		}
	}
	return false
}

// hasDecoratorsBeforeExport is upstream's hasDecoratorsBeforeExport.
func hasDecoratorsBeforeExport(node Node) bool {
	if !node.Is("ExportDefaultDeclaration", "ExportNamedDeclaration", "DeclareExportDeclaration") {
		return false
	}
	decorators := node.Child("declaration").List("decorators")
	return len(decorators) > 0 && locStart(node) == locStart(decorators[0])
}

// printIgnored is upstream's printIgnored, print/ignored.js.
func printIgnored(path *Path, options *Options) Doc {
	current := node(path)
	text := options.OriginalText[locStart(current):locEnd(current)]
	if settingsOf(options).Semi && estree.ShouldIgnoredNodePrintSemicolon(current) {
		text += ";"
	} else if shouldExpressionStatementPrintLeadingSemicolon(path, options) {
		text = ";" + text
	}
	if current.Is("ClassExpression") && len(current.List("decorators")) > 0 {
		return concatIn(path, indentIn(path, concatIn(path, softline, text)), softline)
	}
	return concatIn(path, text)
}
