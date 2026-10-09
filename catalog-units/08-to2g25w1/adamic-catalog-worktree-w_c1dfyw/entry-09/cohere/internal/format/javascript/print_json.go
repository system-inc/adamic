package javascript

import (
	"fmt"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

// language-json/print/json.js and language-json/printers.js: the estree-json printer, which the
// json-stringify parser selects. It prints the way JSON.stringify(value, null, indent) does: every
// object and array broken one entry per line, whatever the print width.
//
// The json parser does not come here. It prints with the JavaScript printer (estreePrinter), the way
// upstream's createParser defaults its astFormat to estree.

// estreeJSONPrinter is upstream's estree-json printer. The json-stringify parser refuses comments,
// so none of the comment hooks are needed.
var estreeJSONPrinter = &printing.Printer[*estree.Node]{
	Print:       printJSON,
	LocStart:    estree.LocStart,
	LocEnd:      estree.LocEnd,
	VisitorKeys: jsonVisitorKeys,
}

// jsonVisitorKeys is language-json/visitor-keys.evaluate.js.
func jsonVisitorKeys(node Node) []string {
	switch node.Type() {
	case "JsonRoot":
		return []string{"node"}
	case "ArrayExpression":
		return []string{"elements"}
	case "ObjectExpression":
		return []string{"properties"}
	case "ObjectProperty":
		return []string{"key", "value"}
	case "UnaryExpression":
		return []string{"argument"}
	case "TemplateLiteral":
		return []string{"quasis"}
	}
	return nil
}

// printJSON is upstream's printJson.
func printJSON(path *Path, options *Options, print PrintFunc, _ any) Doc {
	current := node(path)
	switch current.Type() {
	case "JsonRoot":
		return concatIn(path, print("node", nil), hardline)
	case "ArrayExpression":
		elements := current.List("elements")
		if len(elements) == 0 {
			return doc.Text("[]")
		}
		printed := mapPath(path, func(path *Path, _ int) Doc {
			if node(path) == nil {
				return doc.Text("null")
			}
			return print(nil, nil)
		}, "elements")
		return concatIn(path, "[", indentIn(path, concatIn(path, hardline, join(concatIn(path, ",", hardline), printed))), hardline, "]")
	case "ObjectExpression":
		if len(current.List("properties")) == 0 {
			return doc.Text("{}")
		}
		return concatIn(path, "{", indentIn(path, concatIn(path, hardline, join(concatIn(path, ",", hardline), printAll(path, print, "properties")))), hardline, "}")
	case "ObjectProperty":
		return concatIn(path, print("key", nil), ": ", print("value", nil))
	case "UnaryExpression":
		operator := current.String("operator")
		if operator == "+" {
			operator = ""
		}
		return concatIn(path, operator, print("argument", nil))
	case "NullLiteral":
		return doc.Text("null")
	case "BooleanLiteral":
		if current.Bool("value") {
			return doc.Text("true")
		}
		return doc.Text("false")
	case "StringLiteral":
		return doc.Text(printString(getRaw(current), options))
	case "NumericLiteral":
		// Intentionally to not use `printNumber`
		// We may start stop support number normalization like the string print
		// Upstream compares String(Number(raw)) with raw. The literal's value stands in for Number(raw):
		// the two differ only where raw has a numeric separator, and then neither string equals raw.
		raw := getRaw(current)
		if isJSONObjectKey(path) && javaScriptNumberString(current.Get("value").(float64)) == raw {
			return doc.Text(`"` + raw + `"`)
		}
		return doc.Text(raw)
	case "Identifier":
		if isJSONObjectKey(path) {
			return doc.Text(jsonStringify(current.String("name")))
		}
		return doc.Text(current.String("name"))
	case "TemplateLiteral":
		// There is only one `TemplateElement`
		return print([]any{"quasis", 0}, nil)
	case "TemplateElement":
		return doc.Text(jsonStringify(*current.Get("value").(*estree.TemplateValue).Cooked))
	}
	panic(fmt.Sprintf("unknown JSON node type %q", current.Type()))
}

// isJSONObjectKey is upstream's isObjectKey in print/json.js.
func isJSONObjectKey(path *Path) bool {
	return keyOf(path) == "key" && parentOf(path).Is("ObjectProperty")
}
