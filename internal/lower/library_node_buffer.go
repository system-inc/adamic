package lower

import (
	"math"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The shared loader supplies the pinned @types/node tree. Ambient modules in
// application files are not host declarations, even when their names match.
func nodeBufferModule(symbol *ast.Symbol) string {
	if symbol == nil {
		return ""
	}
	for _, declaration := range symbol.Declarations {
		source := ast.GetSourceFileOfNode(declaration)
		if source == nil || !strings.Contains(strings.ReplaceAll(source.FileName().AsString(), "\\", "/"), "/@types/node/") {
			continue
		}
		for parent := declaration; parent != nil; parent = parent.Parent {
			if parent.Kind == ast.KindModuleDeclaration && parent.Name() != nil && parent.Name().Kind == ast.KindStringLiteral {
				switch name := parent.Name().Text(); name {
				case "node:buffer", "node:crypto":
					return name
				}
			}
		}
	}
	return ""
}
func (l *lowering) nodeBufferSymbol(node *ast.Node, name string) bool {
	symbol := l.symbol(node)
	return symbol != nil && symbol.Name == name && nodeBufferModule(symbol) != ""
}
func (l *lowering) nodeBufferType(proven *checker.Type, name string) bool {
	if proven == nil {
		return false
	}
	symbol := proven.Symbol()
	return symbol != nil && symbol.Name == name && nodeBufferModule(symbol) != ""
}

// nodeBufferAccessRefusal checks a member read that isn't called.
func (l *lowering) nodeBufferAccessRefusal(node *ast.Node) error {
	if (node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression) && !called(node) {
		if err := l.nodeBufferUnsupportedUse(node); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	refusalRules.register(2700, "nodeBufferAccessRefusal", (*lowering).nodeBufferAccessRefusal)
}

// Reading a host member as a value must not fall through to ordinary object
// slots. Hash's private runtime layout is not its inherited stream layout.
func (l *lowering) nodeBufferUnsupportedUse(node *ast.Node) error {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindPropertyAccessExpression {
		access := node.AsPropertyAccessExpression()
		name := node.Name().Text()
		receiver := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression))
		switch {
		case l.nodeBufferSymbol(access.Expression, "Buffer"):
			return l.notYet(node, "Buffer."+name+" read as a value")
		case l.nodeBufferType(receiver, "Buffer"):
			if name != "length" {
				return l.notYet(node, "Buffer."+name+" outside the census value reads")
			}
		case l.nodeBufferType(receiver, "Hash"):
			return l.notYet(node, "Hash."+name+" outside the census value reads")
		}
	}
	if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		member := ast.SkipParentheses(access.ArgumentExpression)
		receiver := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression))
		if member.Kind == ast.KindStringLiteral {
			for _, name := range []string{"Buffer", "Hash"} {
				if l.nodeBufferType(receiver, name) || (name == "Buffer" && l.nodeBufferSymbol(access.Expression, name)) {
					return l.notYet(node, name+"."+member.Text()+" accessed through a string index")
				}
			}
		}
	}
	if node.Kind == ast.KindIdentifier || node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression {
		symbol := l.symbol(node)
		if module := nodeBufferModule(symbol); module != "" {
			if symbol.Name == "Buffer" && node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && (node.Parent.Name().Text() == "from" || node.Parent.Name().Text() == "allocUnsafe") && called(node.Parent) {
				return nil
			}
			return l.notYet(node, module+"."+symbol.Name+" read as a value")
		}
	}
	if node.Kind == ast.KindNewExpression {
		symbol := l.symbol(node.AsNewExpression().Expression)
		if module := nodeBufferModule(symbol); module != "" {
			return l.notYet(node, module+"."+symbol.Name+" outside the host census")
		}
	}
	return nil
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
	// A Buffer is held as a Uint8Array (representationReason), so buffer_from and buffer_copy make one.
	operation, encoding, result := "", 0, ir.Uint8Array
	var receiver *ast.Node
	if l.nodeBufferSymbol(callee, "createHash") {
		if len(args) != 1 || ast.SkipParentheses(args[0]).Kind != ast.KindStringLiteral || args[0].Text() != "sha256" {
			return nil, true, l.notYet(node, "crypto.createHash outside the census's sha256 algorithm")
		}
		return ir.NodeBufferCall{Function: "hash_new", Returns: ir.Object}, true, nil
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		if symbol := l.symbol(callee); nodeBufferModule(symbol) != "" {
			return nil, true, l.notYet(node, nodeBufferModule(symbol)+"."+symbol.Name+" outside the host census")
		}
		return nil, false, nil
	}
	access := callee.AsPropertyAccessExpression()
	name := callee.Name().Text()
	switch {
	case l.nodeBufferSymbol(access.Expression, "Buffer") && name == "allocUnsafe":
		return l.nodeBufferAllocUnsafe(node)
	case l.nodeBufferSymbol(access.Expression, "Buffer"):
		if name != "from" || len(args) < 1 || len(args) > 2 {
			return nil, true, l.notYet(node, "Buffer."+name+" outside from(input, encoding)")
		}
		operation = "buffer_from"
	case l.nodeBufferType(l.checker.GetTypeAtLocation(access.Expression), "Buffer") && (name == "slice" || name == "subarray"):
		return l.nodeBufferSubarray(node, access.Expression, name)
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
		if symbol := l.symbol(callee); nodeBufferModule(symbol) != "" {
			return nil, true, l.notYet(node, nodeBufferModule(symbol)+"."+symbol.Name+" outside the host census")
		}
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
			} else if operation == "buffer_from" && value.Type() == ir.Uint8Array && l.nodeBufferType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(args[0])), "Buffer") {
				// Buffer.from(aBuffer) copies its bytes. Only a Buffer: Buffer.from(aUint8Array) isn't in the census.
				if len(args) != 1 {
					return nil, true, l.notYet(node, "Buffer.from(buffer) with an encoding")
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

// nodeBufferAllocUnsafe lowers Buffer.allocUnsafe(size) of one number to a Buffer of that many bytes.
// Node checks the size before it truncates it, so -0.5 throws ERR_OUT_OF_RANGE, and a size it can't
// allocate throws ERR_MEMORY_ALLOCATION_FAILED: both are the runtime's (node_buffer.c), and the call is
// one that can throw (ir.NodeBufferCall's MayThrow). Its bytes are unspecified on Node, and zero natively.
func (l *lowering) nodeBufferAllocUnsafe(node *ast.Node) (ir.Expression, bool, error) {
	args := node.AsCallExpression().Arguments.Nodes
	if len(args) != 1 || args[0].Kind == ast.KindSpreadElement {
		return nil, true, l.notYet(node, "Buffer.allocUnsafe outside allocUnsafe(size)")
	}
	size, err := l.expression(args[0])
	if err != nil {
		return nil, true, err
	}
	if size.Type() != ir.Number {
		return nil, true, l.notYet(node, "Buffer.allocUnsafe of a size that isn't a number")
	}
	return ir.NodeBufferCall{Function: "buffer_alloc_unsafe", Arguments: []ir.Expression{size}, Returns: ir.Uint8Array}, true, nil
}

// nodeBufferSubarray lowers buffer.slice(start, end) and buffer.subarray(start, end) to a view sharing
// the Buffer's bytes, as Node's do: a Buffer's slice is its subarray, where a Uint8Array's slice copies
// (typedArrayExpression refuses that one, and a Uint8Array never reaches here). Each argument clamps as
// a typed array's subarray's does, NaN as 0, a fraction truncated, a negative from the end, and
// undefined as the default, so each is held as number | undefined, as typedArrayExpression holds
// subarray's: the native emitter passes its present flag (typedArrayOffset). The view is a Buffer, an
// ir.Uint8Array as every Buffer is, whose data points into its owner's (typed_array.c).
func (l *lowering) nodeBufferSubarray(node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	args := node.AsCallExpression().Arguments.Nodes
	if len(args) > 2 {
		return nil, true, l.notYet(node, "Buffer."+name+" outside "+name+"(start, end)")
	}
	for _, argument := range args {
		if argument.Kind == ast.KindSpreadElement {
			return nil, true, l.notYet(node, "Buffer."+name+" outside "+name+"(start, end)")
		}
	}
	// A ?. call meets optional_host_call.go's stop for a host method first (#3625x53); this one is the
	// guard behind it.
	if err := l.optionalCall(node); err != nil {
		return nil, true, err
	}
	array, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	lowered := []ir.Expression{}
	for _, argument := range args {
		value, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		value = fit(value, ir.MaybeNumber)
		if value.Type() != ir.MaybeNumber {
			return nil, true, l.notYet(argument, "a Buffer."+name+" offset that isn't number or undefined")
		}
		lowered = append(lowered, value)
	}
	return ir.TypedArraySubarray{Array: array, Arguments: lowered}, true, nil
}
