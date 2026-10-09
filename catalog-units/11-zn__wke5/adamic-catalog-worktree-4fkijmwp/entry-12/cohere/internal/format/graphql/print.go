package graphql

// src/language-graphql/printer-graphql.js: genericPrint and the printer object's hooks. The object itself
// is graphqlPrinter in format.go.

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

func genericPrint(path *astPath, options *printerOptions, print printing.PrintFunc, _ any) doc.Doc {
	node := currentNode(path)

	switch node.Type() {
	case "Document":
		return append(
			join(hardline, printSequence(path, options, print, "definitions")),
			hardline,
		)

	case "OperationDefinition":
		hasOperation := options.OriginalText[locStart(node)] != '{'
		hasName := node.Child("name") != nil
		return doc.Concat{
			printDescription(path, options, print),
			choose(hasOperation, text(node.String("operation"))),
			choose(hasOperation && hasName, doc.Concat{text(" "), print("name", nil)}),
			choose(hasOperation && !hasName && isNonEmptyArray(node, "variableDefinitions"), text(" ")),
			printVariableDefinitions(path, print),
			printDirectives(path, print),
			choose(!(!hasOperation && !hasName), text(" ")),
			print("selectionSet", nil),
		}

	case "FragmentDefinition":
		return doc.Concat{
			printDescription(path, options, print),
			text("fragment "),
			print("name", nil),
			printVariableDefinitions(path, print),
			text(" on "),
			print("typeCondition", nil),
			printDirectives(path, print),
			text(" "),
			print("selectionSet", nil),
		}

	case "SelectionSet":
		return doc.Concat{
			text("{"),
			indent(
				hardline,
				join(hardline, printSequence(path, options, print, "selections")),
			),
			hardline,
			text("}"),
		}

	case "Field":
		return group(
			choose(node.Child("alias") != nil, doc.Concat{print("alias", nil), text(": ")}),
			print("name", nil),
			choose(isNonEmptyArray(node, "arguments"), group(
				text("("),
				indent(
					softline,
					join(
						doc.Concat{ifBreak(empty, text(", ")), softline},
						printSequence(path, options, print, "arguments"),
					),
				),
				softline,
				text(")"),
			)),
			printDirectives(path, print),
			choose(node.Child("selectionSet") != nil, text(" ")),
			print("selectionSet", nil),
		)

	case "Name":
		return text(node.String("value"))

	case "StringValue":
		if node.Bool("block") {
			lines := strings.Split(strings.ReplaceAll(node.String("value"), `"""`, `\"""`), "\n")
			if len(lines) == 1 {
				lines[0] = trim(lines[0])
			}

			allEmpty := true
			for _, line := range lines {
				if line != "" {
					allEmpty = false
					break
				}
			}
			if allEmpty {
				lines = lines[:0]
			}

			parts := []doc.Doc{text(`"""`)}
			for _, line := range lines {
				parts = append(parts, text(line))
			}
			parts = append(parts, text(`"""`))
			return join(hardline, parts)
		}
		return doc.Concat{
			text(`"`),
			text(strings.ReplaceAll(escapeQuoteAndBackslash.Replace(node.String("value")), "\n", `\n`)),
			text(`"`),
		}

	case "IntValue", "FloatValue", "EnumValue":
		return text(node.String("value"))

	case "BooleanValue":
		if node.Bool("value") {
			return text("true")
		}
		return text("false")

	case "NullValue":
		return text("null")

	case "Variable":
		return doc.Concat{text("$"), print("name", nil)}

	case "ListValue":
		isEmpty := !isNonEmptyArray(node, "values")
		return group(
			text("["),
			printing.PrintDanglingComments(path, options, printing.DanglingOptions[*estree.Node]{Indent: true}),
			choose(!isEmpty, indent(
				softline,
				join(doc.Concat{ifBreak(empty, text(", ")), softline}, printAll(path, print, "values")),
			)),
			softline,
			text("]"),
		)

	case "ObjectValue":
		isEmpty := !isNonEmptyArray(node, "fields")
		bracketSpace := choose(settingsOf(options).BracketSpacing && !isEmpty, text(" "))
		return group(
			text("{"),
			bracketSpace,
			printing.PrintDanglingComments(path, options, printing.DanglingOptions[*estree.Node]{Indent: true}),
			choose(!isEmpty, doc.Concat{
				indent(
					softline,
					join(doc.Concat{ifBreak(empty, text(", ")), softline}, printAll(path, print, "fields")),
				),
			}),
			softline,
			ifBreak(empty, bracketSpace),
			text("}"),
		)

	case "ObjectField", "Argument", "FragmentArgument":
		return doc.Concat{print("name", nil), text(": "), print("value", nil)}

	case "Directive":
		return doc.Concat{text("@"), print("name", nil), printArguments(path, options, print)}

	case "NamedType":
		return print("name", nil)

	case "VariableDefinition":
		return doc.Concat{
			printDescription(path, options, print),
			print("variable", nil),
			text(": "),
			print("type", nil),
			choose(node.Child("defaultValue") != nil, doc.Concat{text(" = "), print("defaultValue", nil)}),
			printDirectives(path, print),
		}

	case "ObjectTypeExtension",
		"ObjectTypeDefinition",
		"InputObjectTypeExtension",
		"InputObjectTypeDefinition",
		"InterfaceTypeExtension",
		"InterfaceTypeDefinition":
		kind := node.Type()
		parts := doc.Concat{}

		if strings.HasSuffix(kind, "TypeDefinition") {
			parts = append(parts, printDescription(path, options, print))
		} else {
			parts = append(parts, text("extend "))
		}

		if strings.HasPrefix(kind, "ObjectType") {
			parts = append(parts, text("type"))
		} else if strings.HasPrefix(kind, "InputObjectType") {
			parts = append(parts, text("input"))
		} else {
			parts = append(parts, text("interface"))
		}
		parts = append(parts, text(" "), print("name", nil))

		if !strings.HasPrefix(kind, "InputObjectType") &&
			isNonEmptyArray(node, "interfaces") {
			parts = append(parts,
				text(" implements "),
				indent(group(join(doc.Concat{text(" &"), line}, printAll(path, print, "interfaces")))),
			)
		}

		parts = append(parts, printDirectives(path, print))

		if isNonEmptyArray(node, "fields") {
			parts = append(parts, doc.Concat{
				text(" {"),
				indent(
					hardline,
					join(hardline, printSequence(path, options, print, "fields")),
				),
				hardline,
				text("}"),
			})
		}

		return parts

	case "FieldDefinition":
		return doc.Concat{
			printDescription(path, options, print),
			print("name", nil),
			choose(isNonEmptyArray(node, "arguments"), group(
				text("("),
				indent(
					softline,
					join(
						doc.Concat{ifBreak(empty, text(", ")), softline},
						printSequence(path, options, print, "arguments"),
					),
				),
				softline,
				text(")"),
			)),
			text(": "),
			print("type", nil),
			printDirectives(path, print),
		}

	case "DirectiveDefinition":
		return append(doc.Concat{
			printDescription(path, options, print),
			text("directive "),
			text("@"),
			print("name", nil),
			choose(isNonEmptyArray(node, "arguments"), group(
				text("("),
				indent(
					softline,
					join(
						doc.Concat{ifBreak(empty, text(", ")), softline},
						printSequence(path, options, print, "arguments"),
					),
				),
				softline,
				text(")"),
			)),
			printDirectives(path, print),
			choose(node.Bool("repeatable"), text(" repeatable")),
			text(" on "),
		}, join(text(" | "), printAll(path, print, "locations"))...)

	case "DirectiveExtension":
		return doc.Concat{
			text("extend directive @"),
			print("name", nil),
			printDirectives(path, print),
		}

	case "EnumTypeExtension", "EnumTypeDefinition":
		return doc.Concat{
			printDescription(path, options, print),
			choose(node.Is("EnumTypeExtension"), text("extend ")),
			text("enum "),
			print("name", nil),
			printDirectives(path, print),
			choose(isNonEmptyArray(node, "values"), doc.Concat{
				text(" {"),
				indent(
					hardline,
					join(hardline, printSequence(path, options, print, "values")),
				),
				hardline,
				text("}"),
			}),
		}

	case "EnumValueDefinition":
		return doc.Concat{
			printDescription(path, options, print),
			print("name", nil),
			printDirectives(path, print),
		}

	case "InputValueDefinition":
		return doc.Concat{
			printDescription(path, options, print),
			print("name", nil),
			text(": "),
			print("type", nil),
			choose(node.Child("defaultValue") != nil, doc.Concat{text(" = "), print("defaultValue", nil)}),
			printDirectives(path, print),
		}

	case "SchemaExtension":
		parts := doc.Concat{
			text("extend schema"),
			printDirectives(path, print),
		}
		if isNonEmptyArray(node, "operationTypes") {
			parts = append(parts,
				text(" {"),
				indent(
					hardline,
					join(
						hardline,
						printSequence(path, options, print, "operationTypes"),
					),
				),
				hardline,
				text("}"),
			)
		}
		return parts

	case "SchemaDefinition":
		return doc.Concat{
			printDescription(path, options, print),
			text("schema"),
			printDirectives(path, print),
			text(" {"),
			choose(isNonEmptyArray(node, "operationTypes"), indent(
				hardline,
				join(
					hardline,
					printSequence(path, options, print, "operationTypes"),
				),
			)),
			hardline,
			text("}"),
		}

	case "OperationTypeDefinition":
		return doc.Concat{text(node.String("operation")), text(": "), print("type", nil)}

	case "FragmentSpread":
		return doc.Concat{
			text("..."),
			print("name", nil),
			printArguments(path, options, print),
			printDirectives(path, print),
		}

	case "InlineFragment":
		return doc.Concat{
			text("..."),
			choose(node.Child("typeCondition") != nil, doc.Concat{text(" on "), print("typeCondition", nil)}),
			printDirectives(path, print),
			text(" "),
			print("selectionSet", nil),
		}

	case "UnionTypeExtension", "UnionTypeDefinition":
		return group(
			printDescription(path, options, print),
			group(
				choose(node.Is("UnionTypeExtension"), text("extend ")),
				text("union "),
				print("name", nil),
				printDirectives(path, print),
				choose(isNonEmptyArray(node, "types"), doc.Concat{
					text(" ="),
					ifBreak(empty, text(" ")),
					indent(
						ifBreak(doc.Concat{line, text("| ")}, nil),
						join(doc.Concat{line, text("| ")}, printAll(path, print, "types")),
					),
				}),
			),
		)

	case "ScalarTypeExtension", "ScalarTypeDefinition":
		return doc.Concat{
			printDescription(path, options, print),
			choose(node.Is("ScalarTypeExtension"), text("extend ")),
			text("scalar "),
			print("name", nil),
			printDirectives(path, print),
		}

	case "NonNullType":
		return doc.Concat{print("type", nil), text("!")}

	case "ListType":
		return doc.Concat{text("["), print("type", nil), text("]")}

	default:
		/* c8 ignore next */
		panic(fmt.Sprintf("Unexpected Graphql node kind: '%s'.", node.Type()))
	}
}

// choose is upstream's `condition ? value : ""`.
func choose(condition bool, value doc.Doc) doc.Doc {
	if condition {
		return value
	}
	return empty
}

// escapeQuoteAndBackslash is upstream's replaceAll(/["\\]/g, String.raw`\$&`).
var escapeQuoteAndBackslash = strings.NewReplacer(`"`, `\"`, `\`, `\\`)

func canAttachComment(node *estree.Node, _ []*estree.Node) bool {
	return !node.Is("Comment")
}

func printComment(path *astPath, _ *printerOptions) doc.Doc {
	comment := currentNode(path)
	if comment.Is("Comment") {
		return text("#" + trimEnd(comment.String("value")))
	}

	/* c8 ignore next */
	encoded, _ := json.Marshal(comment.Type())
	panic("Not a comment: " + string(encoded))
}

func hasPrettierIgnore(path *astPath) bool {
	node := currentNode(path)
	if node == nil {
		return false
	}
	for _, comment := range node.Comments {
		if trim(comment.String("value")) == "prettier-ignore" {
			return true
		}
	}
	return false
}

// trim and trimEnd are JavaScript's String.prototype.trim and trimEnd: they strip ECMAScript WhiteSpace
// and LineTerminator, which is not the set Go's strings.TrimSpace strips (U+FEFF is in it, U+0085 not).
func trim(value string) string    { return strings.TrimFunc(value, isJavaScriptWhitespace) }
func trimEnd(value string) string { return strings.TrimRightFunc(value, isJavaScriptWhitespace) }

func isJavaScriptWhitespace(character rune) bool {
	switch character {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0x00A0, 0x1680, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000, 0xFEFF:
		return true
	}
	return character >= 0x2000 && character <= 0x200A
}
