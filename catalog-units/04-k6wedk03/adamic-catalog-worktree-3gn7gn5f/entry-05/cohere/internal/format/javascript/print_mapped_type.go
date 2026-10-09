package javascript

import (
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

// print/mapped-type.js. printFlowMappedTypeOptionalModifier and printFlowMappedTypeProperty print
// Flow's ObjectTypeMappedTypeProperty and are not ported.

// printTypeScriptMappedTypeModifier is upstream's printTypeScriptMappedTypeModifier. The token is
// `true`, "+" or "-" in the tree, so it arrives as the raw value.
func printTypeScriptMappedTypeModifier(tokenNode any, keyword string) string {
	if token, isString := tokenNode.(string); isString && (token == "+" || token == "-") {
		return token + keyword
	}

	return keyword
}

// printTypeScriptMappedType is upstream's printTypeScriptMappedType.
func printTypeScriptMappedType(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	// Break after `{` like `printObject`
	shouldBreak := false
	// objectWrap is always "preserve".
	{
		text := stripComments(options)
		// Skip `{`
		start := locStart(current) + 1
		textAfter := text[start:]
		// upstream's textAfter.search(/\S/): the offset of the first non-whitespace character, or -1.
		search := len(textAfter) - len(estree.TrimStartJavaScript(textAfter))
		if search == len(textAfter) {
			search = -1
		}
		nextTokenIndex := start + search
		if hasNewlineInRange(options.OriginalText, start, nextTokenIndex) {
			shouldBreak = true
		}
	}

	var danglingCommentsDoc []any
	danglingComments := getComments(current, commentDangling, nil)
	if len(danglingComments) > 0 {
		lastComment := danglingComments[len(danglingComments)-1]
		// upstream's printDanglingComments without indent returns join(hardline, parts), an array.
		parts, _ := doc.Parts(printing.PrintDanglingComments(path, options, printing.DanglingOptions[Node]{}))
		if len(parts) > 0 {
			for _, part := range parts[:len(parts)-1] {
				danglingCommentsDoc = append(danglingCommentsDoc, part)
			}
			var separator Doc = line
			if isLineComment(lastComment) ||
				hasNewline(options.OriginalText, locEnd(lastComment)) {
				separator = hardline
			}
			danglingCommentsDoc = append(danglingCommentsDoc, groupIn(path, concatIn(path, parts[len(parts)-1], separator)))
		}
	}

	var bracketLine Doc = softline
	if settingsOf(options).BracketSpacing {
		bracketLine = line
	}

	var readonly Doc = emptyDoc
	if current.Truthy("readonly") {
		readonly = concatIn(path, printTypeScriptMappedTypeModifier(current.Get("readonly"), "readonly"), " ")
	}

	var nameType Doc = emptyDoc
	if current.Truthy("nameType") {
		nameType = concatIn(path, " as ", print("nameType", nil))
	}

	optional := ""
	if current.Truthy("optional") {
		optional = printTypeScriptMappedTypeModifier(current.Get("optional"), "?")
	}

	colon := ""
	if current.Truthy("typeAnnotation") {
		colon = ": "
	}

	var semicolon Doc = emptyDoc
	if settingsOf(options).Semi {
		semicolon = ifBreak(";", "")
	}

	indented := []any{bracketLine}
	indented = append(indented, danglingCommentsDoc...)
	indented = append(indented,
		readonly,
		groupIn(path, concatIn(path, "[",
			indentIn(path, concatIn(path, softline,
				print("key", nil),
				" in ",
				print("constraint", nil),
				nameType,
			)),
			softline,
			"]",
		)),
		optional,
		colon,
		print("typeAnnotation", nil),
		semicolon,
	)

	return groupWithIn(path, concatIn(path, "{",
		indentIn(path, concatIn(path, indented...)),
		bracketLine,
		"}",
	),
		doc.GroupOptions{ShouldBreak: shouldBreak},
	)
}
