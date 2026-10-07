package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (l *lowering) jsonParseRefusal(node *ast.Node) error {
	return &Refused{Where: l.program.Where(node), What: "JSON.parse: its result's type can't be proven from the text", Fix: "a checked parse against a declared type is a later design; use a proven literal scalar, discard the result, or stringify a bounded literal parse directly"}
}

// jsonLiteralDepth proves an upper bound, not validity. The runtime alone validates the grammar.
func jsonLiteralDepth(text string) int {
	depth, maximum := 0, 0
	quoted, escaped := false, false
	for _, c := range text {
		if quoted {
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == '"' {
				quoted = false
			}
		} else if c == '"' {
			quoted = true
		} else if c == '{' || c == '[' {
			depth++
			if depth > maximum {
				maximum = depth
			}
		} else if c == '}' || c == ']' {
			depth--
		}
	}
	return maximum
}
func jsonLiteralKind(text string) (string, ir.Type) {
	text = strings.Trim(text, " \t\r\n")
	if len(text) == 0 {
		return "", 0
	}
	switch text[0] {
	case '"':
		return "string", ir.String
	case 't', 'f':
		return "boolean", ir.Boolean
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return "number", ir.Number
	}
	return "", 0
}
func (l *lowering) jsonParse(node *ast.Node, canonical bool) (ir.Expression, error) {
	args := node.AsCallExpression().Arguments.Nodes
	if len(args) > 2 || hasSpread(node) {
		return nil, l.notYet(node, "JSON.parse with spread or extra arguments")
	}
	p := ir.JSONParse{Text: ir.StringConstant{Index: l.constant("undefined")}, ReviverKind: "undefined", Mode: "discard", Of: ir.Number}
	literal, isLiteral := "", false
	rootKind, rootType := "", ir.Type(0)
	if len(args) > 0 {
		n := ast.SkipParentheses(args[0])
		if n.Kind == ast.KindStringLiteral {
			literal, isLiteral = n.Text(), true
			rootKind, rootType = jsonLiteralKind(literal)
		}
		if n.Kind == ast.KindNullKeyword {
			p.Text = ir.StringConstant{Index: l.constant("null")}
		} else {
			value, err := l.expression(args[0])
			if err != nil {
				return nil, err
			}
			switch value.Type() {
			case ir.String:
				p.Text = value
			case ir.Number:
				p.Text = ir.NumberToString{Value: value}
				if !isLiteral {
					rootKind, rootType = "number", ir.Number
				}
			case ir.Boolean:
				p.Text = ir.BooleanToString{Value: value}
				rootKind, rootType = "boolean", ir.Boolean
			default:
				if _, ok := value.(ir.Undefined); ok {
					p.Text = ir.StringConstant{Index: l.constant("undefined")}
				} else {
					return nil, l.notYet(args[0], "JSON.parse text requiring unproved object or union coercion")
				}
			}
		}
	}
	if len(args) > 1 {
		n := ast.SkipParentheses(args[1])
		signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(n), checker.SignatureKindCall)
		if len(signatures) != 0 {
			if !isLiteral || jsonLiteralDepth(literal) > 64 {
				return nil, l.notYet(node, "JSON.parse reviver without a literal nesting bound of 64 (Node's recursive reviver can overflow)")
			}
			if len(signatures) != 1 {
				return nil, l.notYet(n, "JSON.parse overloaded reviver")
			}
			parameters := signatures[0].Parameters()
			if len(parameters) > 2 {
				return nil, l.notYet(n, "JSON.parse reviver holder, this or context without proven metadata")
			}
			if len(parameters) > 0 {
				of, known := l.representation(l.checker.GetTypeOfSymbol(parameters[0]))
				if !known || of != ir.String || !l.jsonScalarArgument(ir.String, l.checker.GetTypeOfSymbol(parameters[0])) {
					return nil, l.notYet(n, "JSON.parse reviver key without a proven string type")
				}
			}
			if len(parameters) > 1 {
				of, known := l.representation(l.checker.GetTypeOfSymbol(parameters[1]))
				if rootType == 0 || !known || of != rootType || !l.jsonScalarArgument(rootType, l.checker.GetTypeOfSymbol(parameters[1])) {
					return nil, l.notYet(n, "JSON.parse reviver value without a proven scalar input for every visit")
				}
			}
			returned := l.checker.GetReturnTypeOfSignature(signatures[0])
			if returned.Flags()&checker.TypeFlagsVoid != 0 && !l.jsonCallbackReturnsUndefined(n) {
				return nil, l.notYet(n, "JSON.parse void reviver without proof that its actual return is undefined")
			}
			if returned.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsUndefined|checker.TypeFlagsNever) != 0 {
				p.ReviverKind = "undefined"
				rootKind, rootType = "", 0
			} else {
				if l.includesUndefined(returned) {
					return nil, l.notYet(n, "JSON.parse reviver return possibly undefined without a proven result union")
				}
				of, known := l.representation(returned)
				kinds := map[ir.Type]string{ir.Number: "number", ir.String: "string", ir.Boolean: "boolean"}
				kind, scalar := kinds[of]
				if !known || !scalar {
					return nil, l.notYet(n, "JSON.parse reviver return without a proven scalar representation")
				}
				p.ReviverKind = kind
				rootKind, rootType = kind, of
			}
			value, err := l.expression(n)
			if err != nil {
				return nil, err
			}
			p.Reviver = value
			p.Takes = len(parameters)
		} else if n.Kind != ast.KindNullKeyword {
			value, err := l.expression(n)
			if err != nil {
				return nil, err
			}
			if value.Type() != ir.Number && value.Type() != ir.String && value.Type() != ir.Boolean {
				if _, ok := value.(ir.Undefined); !ok {
					return nil, l.notYet(n, "JSON.parse non-callable reviver without proven primitive identity")
				}
			}
			p.IgnoredReviver = value
		}
	}
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	discard := parent != nil && parent.Kind == ast.KindExpressionStatement
	if canonical {
		if !isLiteral || jsonLiteralDepth(literal) > 64 {
			return nil, l.notYet(node, "JSON.stringify a parsed value without a literal nesting bound of 64")
		}
		p.Mode, p.Of = "canonical", ir.Object
	} else if !discard {
		if rootType == 0 || (!isLiteral && p.Reviver == nil && rootType == ir.String) {
			return nil, l.jsonParseRefusal(node)
		}
		p.Mode, p.Of = rootKind, rootType
		// JSON.parse is declared any by TypeScript. Its contextual type is no license to trust a
		// number as a string, or a broad scalar as a narrower literal type.
		if contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone); contextual != nil && contextual.Flags()&checker.TypeFlagsAny == 0 {
			var source *checker.Type
			switch rootType {
			case ir.Number:
				source = l.checker.GetNumberType()
			case ir.String:
				source = l.checker.GetStringType()
			case ir.Boolean:
				source = l.checker.GetBooleanType()
			}
			if source != nil && !l.checker.IsTypeAssignableTo(source, l.concrete(contextual)) {
				return nil, l.notYet(node, "JSON.parse scalar incompatible with its contextual type")
			}
		}
	}
	return p, nil
}

// A contextual () => void can wrap a function that actually returns a value. JSON observes
// that value, so a void annotation alone is insufficient evidence for undefined.
func (l *lowering) jsonCallbackReturnsUndefined(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if ast.IsIdentifier(node) {
		symbol := l.symbol(node)
		if symbol == nil || len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindFunctionDeclaration {
			return false
		}
		node = symbol.Declarations[0]
	}
	if !ast.IsFunctionLike(node) || node.Body() == nil || node.Body().Kind != ast.KindBlock {
		return false
	}
	invalid := false
	var visit ast.Visitor
	visit = func(inner *ast.Node) bool {
		if ast.IsFunctionLike(inner) {
			return false
		}
		if inner.Kind == ast.KindReturnStatement && inner.AsReturnStatement().Expression != nil {
			invalid = true
			return true
		}
		return inner.ForEachChild(visit)
	}
	node.Body().ForEachChild(visit)
	return !invalid
}

// A matching representation does not prove a narrower literal parameter. TypeScript's any
// value parameter in JSON callbacks allows such declarations; the library must prove the call.
func (l *lowering) jsonScalarArgument(of ir.Type, target *checker.Type) bool {
	var source *checker.Type
	switch of {
	case ir.Number:
		source = l.checker.GetNumberType()
	case ir.String:
		source = l.checker.GetStringType()
	case ir.Boolean:
		source = l.checker.GetBooleanType()
	default:
		return false
	}
	return l.checker.IsTypeAssignableTo(source, l.concrete(target))
}
