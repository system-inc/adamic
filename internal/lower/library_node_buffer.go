package lower

import (
	"math"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func (l *lowering) nodeBufferSymbol(node *ast.Node, name string) bool {
	symbol := l.symbol(node)
	return symbol != nil && symbol.Name == name && len(symbol.Declarations) > 0 && load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0]))
}
func (l *lowering) nodeBufferType(proven *checker.Type, name string) bool {
	if proven == nil {
		return false
	}
	symbol := proven.Symbol()
	return symbol != nil && symbol.Name == name && len(symbol.Declarations) > 0 && load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0]))
}
func (l *lowering) nodeBufferReadArgument(node *ast.Node) bool {
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	call := parent.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	return len(call.Arguments.Nodes) > 0 && call.Arguments.Nodes[0] == outer && callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "from" && l.nodeBufferSymbol(callee.AsPropertyAccessExpression().Expression, "Buffer")
}
func (l *lowering) nodeBufferView(from, to *checker.Type) bool {
	for _, name := range []string{"Buffer", "Hash"} {
		if l.nodeBufferType(from, name) != l.nodeBufferType(to, name) {
			return false
		}
	}
	return true
}

func (l *lowering) nodeBufferEncoding(node *ast.Node, fallback int) (int, error) {
	if node == nil {
		return fallback, nil
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindStringLiteral {
		switch node.Text() {
		case "utf8", "utf-8":
			return 0, nil
		case "utf16le", "utf-16le", "ucs2", "ucs-2":
			return 1, nil
		case "base64":
			return 2, nil
		case "hex":
			return 3, nil
		}
	}
	return 0, l.notYet(node, "a Buffer encoding other than a census literal")
}
func (l *lowering) libraryNodeBuffer(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	args := call.Arguments.Nodes
	operation, encoding, result := "", 0, ir.Array
	var receiver *ast.Node
	if l.nodeBufferSymbol(callee, "createHash") {
		if len(args) != 1 || ast.SkipParentheses(args[0]).Kind != ast.KindStringLiteral || args[0].Text() != "sha256" {
			return nil, true, l.notYet(node, "crypto.createHash outside the census's sha256 algorithm")
		}
		return ir.NodeBufferCall{Function: "hash_new", Returns: ir.Object}, true, nil
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	access := callee.AsPropertyAccessExpression()
	name := callee.Name().Text()
	switch {
	case l.nodeBufferSymbol(access.Expression, "Buffer"):
		if name != "from" || len(args) < 1 || len(args) > 2 {
			return nil, true, l.notYet(node, "Buffer constructor outside from(input, encoding)")
		}
		operation = "buffer_from"
	case l.nodeBufferType(l.checker.GetTypeAtLocation(access.Expression), "Buffer"):
		if name != "toString" || len(args) > 3 {
			return nil, true, l.notYet(node, "Buffer."+name+" outside the host census")
		}
		operation, result, receiver = "buffer_string", ir.String, access.Expression
	case l.nodeBufferType(l.checker.GetTypeAtLocation(access.Expression), "Hash"):
		receiver = access.Expression
		switch name {
		case "update":
			if len(args) < 1 || len(args) > 2 {
				return nil, true, l.notYet(node, "Hash.update with these arguments")
			}
			operation, result = "hash_update", ir.Object
		case "digest":
			if len(args) != 1 || args[0].Kind != ast.KindStringLiteral || args[0].Text() != "hex" {
				return nil, true, l.notYet(node, "Hash.digest outside hex")
			}
			operation, result, args = "hash_digest", ir.String, nil
		default:
			return nil, true, l.notYet(node, "Hash."+name+" outside the host census")
		}
	default:
		return nil, false, nil
	}
	var encodingNode *ast.Node
	if operation == "buffer_string" {
		if len(args) > 0 {
			encodingNode = args[0]
		}
	} else if len(args) > 1 {
		encodingNode = args[1]
	}
	var err error
	if encoding, err = l.nodeBufferEncoding(encodingNode, 0); err != nil {
		return nil, true, err
	}
	values := []ir.Expression{}
	if receiver != nil {
		value, err := l.expression(receiver)
		if err != nil {
			return nil, true, err
		}
		values = append(values, value)
	}
	if operation == "buffer_string" {
		for index := 1; index <= 2; index++ {
			value := ir.Expression(ir.NumberConstant{Value: 0})
			if index == 2 {
				value = ir.NumberConstant{Value: math.Inf(1)}
			}
			if len(args) > index {
				value, err = l.expression(args[index])
				if err != nil {
					return nil, true, err
				}
				if value.Type() != ir.Number {
					return nil, true, l.notYet(args[index], "a non-number Buffer byte offset")
				}
			}
			values = append(values, value)
		}
	} else {
		if operation != "hash_digest" {
			value, err := l.expression(args[0])
			if err != nil {
				return nil, true, err
			}
			if operation == "buffer_from" && value.Type() == ir.Array {
				if len(args) != 1 {
					return nil, true, l.notYet(node, "Buffer.from(array) with an encoding")
				}
				element, err := l.elementType(args[0])
				if err != nil {
					return nil, true, err
				}
				if element != ir.Number {
					return nil, true, l.notYet(node, "Buffer.from with non-number elements")
				}
				operation = "buffer_copy"
			} else if value.Type() != ir.String {
				return nil, true, l.notYet(node, "Buffer or Hash input other than a string or numeric array")
			}
			values = append(values, value)
		}
	}
	return ir.NodeBufferCall{Function: operation, Arguments: values, Encoding: encoding, Returns: result}, true, nil
}
