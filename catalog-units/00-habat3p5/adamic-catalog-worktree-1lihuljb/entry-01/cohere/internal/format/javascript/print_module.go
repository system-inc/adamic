package javascript

import (
	"fmt"
	"strings"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

// print/module.js: import and export declarations, their specifiers, sources and attributes.

/*
- `ImportDeclaration`
*/
// printImportDeclaration is upstream's printImportDeclaration.
func printImportDeclaration(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	phase := ""
	if current.Truthy("phase") {
		phase = " " + current.String("phase")
	}
	return concatIn(path, "import",
		phase,
		// Upstream's printImportKind(node) leaves spaceBeforeKind undefined, which defaults to true.
		printImportKind(current, true),
		printModuleSpecifiers(path, options, print),
		printModuleSource(path, options, print),
		printImportAttributes(path, options, print),
		printSemicolon(options),
	)
}

// isDefaultExport is upstream's isDefaultExport.
func isDefaultExport(node Node) bool {
	return node.Is("ExportDefaultDeclaration") ||
		(node.Is("DeclareExportDeclaration") && node.Truthy("default"))
}

/*
- `ExportDefaultDeclaration`
- `ExportNamedDeclaration`
- `ExportAllDeclaration`
- `DeclareExportDeclaration`(flow)
- `DeclareExportAllDeclaration`(flow)
*/
// printExportDeclaration is upstream's printExportDeclaration.
func printExportDeclaration(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)

	defaultKeyword := ""
	if isDefaultExport(current) {
		defaultKeyword = " default"
	}
	parts := []any{
		printDecoratorsBeforeExport(path, options, print),
		printDeclareToken(path),
		"export",
		defaultKeyword,
	}

	declaration := current.Child("declaration")
	exported := current.Child("exported")

	if hasComment(current, commentDangling, nil) {
		parts = append(parts, " ", printing.PrintDanglingComments(path, options, printing.DanglingOptions[Node]{}))

		if needsHardlineAfterDanglingComment(current) {
			parts = append(parts, hardline)
		}
	}

	if declaration != nil {
		parts = append(parts, " ", print("declaration", nil))
	} else {
		parts = append(parts, printExportKind(current))

		if current.Is("ExportAllDeclaration", "DeclareExportAllDeclaration") {
			parts = append(parts, " *")
			if exported != nil {
				parts = append(parts, " as ", print("exported", nil))
			}
		} else {
			parts = append(parts, printModuleSpecifiers(path, options, print))
		}

		parts = append(parts,
			printModuleSource(path, options, print),
			printImportAttributes(path, options, print),
		)
	}

	parts = append(parts, printSemicolonAfterExportDeclaration(current, options))

	return concatIn(path, parts...)
}

// shouldOmitSemicolon is upstream's shouldOmitSemicolon, a createTypeCheckFunction.
func shouldOmitSemicolon(node Node) bool {
	return node.Is(
		"ClassDeclaration",
		"ComponentDeclaration",
		"FunctionDeclaration",
		"TSInterfaceDeclaration",
		"DeclareClass",
		"DeclareComponent",
		"DeclareFunction",
		"DeclareHook",
		"HookDeclaration",
		"TSDeclareFunction",
		"EnumDeclaration",
	)
}

// printSemicolonAfterExportDeclaration is upstream's printSemicolonAfterExportDeclaration.
func printSemicolonAfterExportDeclaration(node Node, options *Options) Doc {
	declaration := node.Child("declaration")
	if declaration == nil ||
		(isDefaultExport(node) && !shouldOmitSemicolon(declaration)) {
		return printSemicolon(options)
	}

	return emptyDoc
}

// printImportOrExportKind is upstream's printImportOrExportKind. Upstream's spaceBeforeKind defaults to
// true; Go has no default parameters, so callers pass it.
func printImportOrExportKind(kind string, spaceBeforeKind bool) string {
	if kind == "" || kind == "value" {
		return ""
	}
	before, after := "", " "
	if spaceBeforeKind {
		before, after = " ", ""
	}
	return before + kind + after
}

// printImportKind is upstream's printImportKind.
func printImportKind(node Node, spaceBeforeKind bool) string {
	return printImportOrExportKind(node.String("importKind"), spaceBeforeKind)
}

// printExportKind is upstream's printExportKind.
func printExportKind(node Node) string {
	return printImportOrExportKind(node.String("exportKind"), true)
}

// printModuleSource is upstream's printModuleSource.
func printModuleSource(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)

	if current.Child("source") == nil {
		return emptyDoc
	}

	from := ""
	if shouldPrintSpecifiers(current, options) {
		from = " from"
	}
	return concatIn(path, from, " ", print("source", nil))
}

// printModuleSpecifiers is upstream's printModuleSpecifiers.
func printModuleSpecifiers(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)

	if !shouldPrintSpecifiers(current, options) {
		return emptyDoc
	}

	parts := []any{" "}

	if len(current.List("specifiers")) > 0 {
		var standaloneSpecifiers []Doc
		var groupedSpecifiers []Doc

		each(path, func(path *Path, _ int) {
			specifierType := node(path).Type()
			if specifierType == "ExportNamespaceSpecifier" ||
				specifierType == "ExportDefaultSpecifier" ||
				specifierType == "ImportNamespaceSpecifier" ||
				specifierType == "ImportDefaultSpecifier" {
				standaloneSpecifiers = append(standaloneSpecifiers, print(nil, nil))
			} else if specifierType == "ExportSpecifier" ||
				specifierType == "ImportSpecifier" {
				groupedSpecifiers = append(groupedSpecifiers, print(nil, nil))
			} else {
				// Upstream throws UnexpectedNodeError(node, "specifier").
				panic(fmt.Sprintf("javascript: unexpected specifier %q", specifierType))
			}
		}, "specifiers")

		parts = append(parts, join(", ", standaloneSpecifiers))

		if len(groupedSpecifiers) > 0 {
			if len(standaloneSpecifiers) > 0 {
				parts = append(parts, ", ")
			}

			canBreak := len(groupedSpecifiers) > 1 ||
				len(standaloneSpecifiers) > 0 ||
				specifiersHaveComment(current.List("specifiers"))

			bracketSpacing := settingsOf(options).BracketSpacing
			if canBreak {
				var spacingLine Doc = softline
				if bracketSpacing {
					spacingLine = line
				}
				parts = append(parts,
					groupIn(path, concatIn(path, "{",
						indentIn(path, concatIn(path, spacingLine,
							join(concatIn(path, ",", line), groupedSpecifiers),
						)),
						printTrailingComma(options, ""),
						spacingLine,
						"}",
					)),
				)
			} else {
				spacing := ""
				if bracketSpacing {
					spacing = " "
				}
				single := []any{"{", spacing}
				for _, specifier := range groupedSpecifiers {
					single = append(single, specifier)
				}
				single = append(single, spacing, "}")
				parts = append(parts, concatIn(path, single...))
			}
		}
	} else {
		parts = append(parts, "{}")
	}
	return concatIn(path, parts...)
}

// specifiersHaveComment is upstream's inline `node.specifiers.some((node) => hasComment(node))`.
func specifiersHaveComment(specifiers []Node) bool {
	for _, specifier := range specifiers {
		if hasAnyComment(specifier) {
			return true
		}
	}
	return false
}

// shouldPrintSpecifiers is upstream's shouldPrintSpecifiers.
func shouldPrintSpecifiers(node Node, options *Options) bool {
	if !node.Is("ImportDeclaration") ||
		len(node.List("specifiers")) > 0 ||
		node.String("importKind") == "type" {
		return true
	}

	text := stripComments(options)[locStart(node):locStart(node.Child("source"))]

	return strings.HasSuffix(estree.TrimEndJavaScript(text), "from")
}

// getImportAttributesKeyword is upstream's getImportAttributesKeyword. It returns "" where upstream
// returns undefined.
func getImportAttributesKeyword(node Node, options *Options) string {
	attributes := node.List("attributes")
	end := locEnd(node)
	if len(attributes) > 0 && attributes[0] != nil {
		end = locStart(attributes[0])
	}
	textBetweenSourceAndAttributes := estree.TrimStartJavaScript(stripComments(options)[locEnd(node.Child("source")):end])

	if strings.HasPrefix(textBetweenSourceAndAttributes, "assert") {
		return "assert"
	}

	if strings.HasPrefix(textBetweenSourceAndAttributes, "with") {
		return "with"
	}

	if len(attributes) > 0 {
		return "with"
	}
	return ""
}

// isSingleTypeImportAttributes is upstream's isSingleTypeImportAttributes.
func isSingleTypeImportAttributes(node Node) bool {
	attributes := node.List("attributes")

	if len(attributes) != 1 {
		return false
	}

	attribute := attributes[0]
	key := attribute.Child("key")
	value := attribute.Child("value")
	return attribute.Is("ImportAttribute") &&
		((key.Is("Identifier") && key.String("name") == "type") ||
			(isStringLiteral(key) && key.String("value") == "type")) &&
		isStringLiteral(value) &&
		!hasAnyComment(attribute) &&
		!hasAnyComment(key) &&
		!hasAnyComment(value)
}

/*
- `ImportDeclaration`
- `ExportDefaultDeclaration`
- `ExportNamedDeclaration`
- `ExportAllDeclaration`
- `DeclareExportDeclaration` (Flow)
- `DeclareExportAllDeclaration` (Flow)
*/
// printImportAttributes is upstream's printImportAttributes.
func printImportAttributes(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)

	if current.Child("source") == nil {
		return emptyDoc
	}

	keyword := getImportAttributesKeyword(current, options)
	if keyword == "" {
		return emptyDoc
	}

	attributesDoc := printObject(path, options, print)
	if isSingleTypeImportAttributes(current) {
		attributesDoc = doc.RemoveLines(attributesDoc)
	}

	return concatIn(path, " "+keyword+" ", attributesDoc)
}

// printModuleSpecifier is upstream's printModuleSpecifier.
func printModuleSpecifier(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	specifierType := current.Type()

	isImportSpecifier := strings.HasPrefix(specifierType, "Import")
	leftSideProperty := "local"
	rightSideProperty := "exported"
	if isImportSpecifier {
		leftSideProperty = "imported"
		rightSideProperty = "local"
	}
	leftSideNode := current.Child(leftSideProperty)
	rightSideNode := current.Child(rightSideProperty)
	var left Doc = emptyDoc
	var right Doc = emptyDoc
	if specifierType == "ExportNamespaceSpecifier" ||
		specifierType == "ImportNamespaceSpecifier" {
		left = doc.Text("*")
	} else if leftSideNode != nil {
		left = print(leftSideProperty, nil)
	}

	if rightSideNode != nil && !isShorthandSpecifier(current) {
		right = print(rightSideProperty, nil)
	}

	kind := current.String("exportKind")
	if specifierType == "ImportSpecifier" {
		kind = current.String("importKind")
	}
	// Upstream's `left && right` tests the docs' truthiness: "" is falsy, any printed doc is truthy.
	as := ""
	if !isEmptyString(left) && !isEmptyString(right) {
		as = " as "
	}
	return concatIn(path, printImportOrExportKind(kind, false /* spaceBeforeKind */),
		left,
		as,
		right,
	)
}
