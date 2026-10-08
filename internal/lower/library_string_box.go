package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
)

// The slot is compiler-private. Reflection/spreads and prototype views are refused;
// admitted boxes preserve object identity and expose only length and indexed units.
const stringBoxSlot = "#StringData"

func (l *lowering) stringBoxType(node *ast.Node) bool {
	t := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node))
	return t.Flags()&checker.TypeFlagsObject != 0 && l.isLibraryType(t, "String")
}

func (l *lowering) stringBoxOrigin(node *ast.Node, depth int) bool {
	if depth > 32 {
		return false
	}
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindNewExpression:
		return l.isLibraryGlobal(node.AsNewExpression().Expression, "String")
	case ast.KindIdentifier:
		symbol := l.symbol(node)
		if symbol != nil && len(symbol.Declarations) == 1 {
			d := symbol.Declarations[0]
			if d.Kind == ast.KindVariableDeclaration && d.AsVariableDeclaration().Initializer != nil && l.regexStableBinding(symbol, d) {
				return l.stringBoxOrigin(d.AsVariableDeclaration().Initializer, depth+1)
			}
		}
	case ast.KindAsExpression:
		return l.stringBoxOrigin(node.AsAsExpression().Expression, depth+1)
	}
	return false
}

func (l *lowering) stringBoxPrimitive(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	if !l.stringBoxOrigin(node, 0) || l.includesUndefined(l.checker.GetTypeAtLocation(node)) {
		return nil, l.notYet(node, "a String box without a proven unchanged intrinsic constructor origin")
	}
	return ir.Property{Object: value, Name: stringBoxSlot, Of: ir.String}, nil
}

func (l *lowering) newStringBox(node *ast.Node) (ir.Expression, error) {
	var args []*ast.Node
	if node.AsNewExpression().Arguments != nil {
		args = node.AsNewExpression().Arguments.Nodes
	}
	if len(args) > 1 || (len(args) == 1 && args[0].Kind == ast.KindSpreadElement) {
		return nil, l.notYet(node, "new String with spread or extra arguments")
	}
	values := []ir.Expression{}
	text := ir.Expression(ir.StringConstant{Index: l.constant("")})
	if len(args) == 1 {
		v, err := l.expression(args[0])
		if err != nil {
			return nil, err
		}
		if _, missing := v.(ir.Undefined); missing {
			text = ir.StringConstant{Index: l.constant("undefined")}
		} else {
			values = append(values, v)
		}
	}
	f, reads := l.stringHelper("box", values)
	if len(reads) > 0 {
		var err error
		text, err = l.stringConversionValue(args[0], reads[0])
		if err != nil {
			return nil, err
		}
	}
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "text", Type: ir.String, Function: f})
	read := ir.Read{Local: local, Of: ir.String}
	l.result.Functions[f].Returns = ir.Object
	l.result.Functions[f].Body = []ir.Statement{ir.Declare{Local: local, Value: text}, ir.Return{Value: ir.ObjectLiteral{Fields: []ir.Field{
		{Name: stringBoxSlot, Value: read}, {Name: "length", Value: ir.StringLength{Value: read}},
	}}}}
	return ir.Call{Function: f, Arguments: values, Returns: ir.Object}, nil
}

// Only the library's own indexed interface is exempted. Numeric property keys
// must be nonnegative integers; fractional keys are absent, rather than truncated.
func (l *lowering) stringBoxIndex(node *ast.Node) (ir.Expression, error) {
	a := node.AsElementAccessExpression()
	if a.QuestionDotToken != nil {
		return nil, l.notYet(node, "optional indexing of a String box")
	}
	object, err := l.expression(a.Expression)
	if err != nil {
		return nil, err
	}
	text, err := l.stringBoxPrimitive(a.Expression, object)
	if err != nil {
		return nil, err
	}
	key := ast.SkipParentheses(a.ArgumentExpression)
	var index ir.Expression
	if key.Kind == ast.KindStringLiteral {
		n, e := strconv.ParseUint(key.Text(), 10, 32)
		if e != nil || strconv.FormatUint(n, 10) != key.Text() {
			return nil, l.notYet(node, "a noncanonical String box index key")
		}
		index = ir.NumberConstant{Value: float64(n)}
	} else {
		index, err = l.expression(key)
		if err != nil {
			return nil, err
		}
	}
	if index.Type() != ir.Number {
		return nil, l.notYet(node, "a String box index other than a number")
	}
	f, reads := l.stringHelper("box_index", []ir.Expression{text, index})
	valid := ir.Binary{Operator: ir.Equal, Left: reads[1], Right: ir.MathCall{Function: "trunc", Arguments: []ir.Expression{reads[1]}}}
	result := ir.Conditional{Condition: valid, WhenTrue: ir.StringIndex{Value: reads[0], Index: reads[1]}, WhenNot: ir.Undefined{}, Of: ir.String}
	l.result.Functions[f].Body = []ir.Statement{ir.Return{Value: result}}
	return ir.Call{Function: f, Arguments: []ir.Expression{text, index}, Returns: ir.String}, nil
}

// This is a library representation guard, not general structural/prototype support.
func (l *lowering) stringBoxViewRefusal(node *ast.Node) error {
	switch node.Kind {
	case ast.KindSpreadAssignment:
		if l.stringBoxType(node.AsSpreadAssignment().Expression) {
			return l.notYet(node, "spreading String indexed exotic properties")
		}
	case ast.KindVariableDeclaration:
		d := node.AsVariableDeclaration()
		if d.Initializer != nil && d.Name().Kind == ast.KindObjectBindingPattern && l.stringBoxType(d.Initializer) {
			return l.notYet(node, "destructuring String prototype fields")
		}
	case ast.KindBinaryExpression:
		b := node.AsBinaryExpression()
		target := ast.SkipParentheses(b.Left)
		if b.OperatorToken.Kind == ast.KindEqualsToken && target.Kind == ast.KindPropertyAccessExpression && l.stringBoxType(target.AsPropertyAccessExpression().Expression) {
			return l.notYet(node, "overwriting a String box prototype member")
		}
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindNewExpression, ast.KindCallExpression, ast.KindObjectLiteralExpression, ast.KindPropertyAccessExpression, ast.KindAsExpression, ast.KindConditionalExpression, ast.KindStringLiteral:
	default:
		return nil
	}
	actual := l.checker.GetTypeAtLocation(node)
	view := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	if view == nil {
		return nil
	}
	boxed := func(t *checker.Type) bool {
		t = l.checker.GetNonNullableType(t)
		return t.Flags()&checker.TypeFlagsObject != 0 && l.isLibraryType(t, "String")
	}
	if boxed(view) && !boxed(actual) {
		return l.notYet(node, "a primitive or structural object viewed as a String box")
	}
	if boxed(actual) && !boxed(view) {
		outer := node
		for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
			outer = outer.Parent
		}
		parent := outer.Parent
		// JSON stringify consumes fresh intrinsic boxes using its proven primitive slot.
		if parent != nil && parent.Kind == ast.KindCallExpression && node.Kind == ast.KindNewExpression {
			callee := ast.SkipParentheses(parent.AsCallExpression().Expression)
			if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "stringify" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "JSON") {
				return nil
			}
		}
		if parent != nil && parent.Kind == ast.KindCallExpression && len(parent.AsCallExpression().Arguments.Nodes) > 0 && parent.AsCallExpression().Arguments.Nodes[0] == outer {
			callee := ast.SkipParentheses(parent.AsCallExpression().Expression)
			if l.isLibraryGlobal(callee, "String") {
				return nil
			}
			if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "call" {
				if _, known := l.stringPrototypeMethod(callee.AsPropertyAccessExpression().Expression); known {
					return nil
				}
			}
		}
		if parent != nil && parent.Kind == ast.KindNewExpression && l.isLibraryGlobal(parent.AsNewExpression().Expression, "String") {
			return nil
		}
		return l.notYet(node, "a String box viewed through erased, mutable or prototype fields")
	}

	if !boxed(actual) && l.stringBoxViewLosesSlot(actual, view, 0) {
		return l.notYet(node, "a nested String box viewed through erased fields")
	}

	return nil
}

// Follow container fields as well: an unknown view of {inner: new String(...) }
// would otherwise expose the non-exotic backing object through a later checked cast.
func (l *lowering) stringBoxViewLosesSlot(actual, view *checker.Type, depth int) bool {
	return l.stringBoxViewLosesSlotSeen(actual, view, depth, make(map[[2]*checker.Type]bool))
}

func (l *lowering) stringBoxViewLosesSlotSeen(actual, view *checker.Type, depth int, seen map[[2]*checker.Type]bool) bool {
	if actual == nil || actual == view {
		return false
	}
	if depth > 64 {
		return true
	}
	pair := [2]*checker.Type{actual, view}
	if seen[pair] {
		return false
	}
	seen[pair] = true
	actual = l.checker.GetNonNullableType(actual)
	if view != nil {
		view = l.checker.GetNonNullableType(view)
	}
	if actual.Flags()&checker.TypeFlagsObject != 0 && l.isLibraryType(actual, "String") {
		return view == nil || view.Flags()&checker.TypeFlagsObject == 0 || !l.isLibraryType(view, "String")
	}
	if actual.Flags()&checker.TypeFlagsUnion != 0 {
		for _, part := range actual.Types() {
			if l.stringBoxViewLosesSlotSeen(part, view, depth+1, seen) {
				return true
			}
		}
		return false
	}
	if actual.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	if l.checker.IsArrayType(actual) || checker.IsTupleType(actual) || l.isLibraryType(actual, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
		values := l.checker.GetTypeArguments(actual)
		var views []*checker.Type
		if view != nil && (l.checker.IsArrayType(view) || checker.IsTupleType(view) || l.isLibraryType(view, "Map", "ReadonlyMap", "Set", "ReadonlySet")) {
			views = l.checker.GetTypeArguments(view)
		}
		for index, value := range values {
			var target *checker.Type
			if index < len(views) {
				target = views[index]
			}
			if l.stringBoxViewLosesSlotSeen(value, target, depth+1, seen) {
				return true
			}
		}
		return false
	}
	// Callable library prototypes are not stored container fields.
	if len(l.checker.GetSignaturesOfType(actual, checker.SignatureKindCall)) != 0 {
		return false
	}
	for _, field := range l.checker.GetPropertiesOfType(actual) {
		var target *checker.Type
		if view != nil && view.Flags()&checker.TypeFlagsObject != 0 {
			target = l.checker.GetTypeOfPropertyOfType(view, field.Name)
		}
		if l.stringBoxViewLosesSlotSeen(l.checker.GetTypeOfSymbol(field), target, depth+1, seen) {
			return true
		}
	}
	return false
}

// The indexed exotic own properties can be answered without exposing the backing
// object's private slot. Dynamic ToPropertyKey and prototype mutation stay refused.
func (l *lowering) stringBoxHasOwn(node *ast.Node, text ir.Expression, args []*ast.Node) (ir.Expression, bool, error) {
	if len(args) != 1 || ast.SkipParentheses(args[0]).Kind != ast.KindStringLiteral {
		return nil, true, l.notYet(node, "String box hasOwnProperty with a nonliteral key")
	}
	key := ast.SkipParentheses(args[0]).Text()
	f, reads := l.stringHelper("box_hasOwn", []ir.Expression{text})
	result := ir.Expression(ir.BooleanConstant{Value: key == "length"})
	if index, err := strconv.ParseUint(key, 10, 32); err == nil && strconv.FormatUint(index, 10) == key {
		result = ir.Binary{Operator: ir.Less, Left: ir.NumberConstant{Value: float64(index)}, Right: ir.StringLength{Value: reads[0]}}
	}
	l.result.Functions[f].Returns = ir.Boolean
	l.result.Functions[f].Body = []ir.Statement{ir.Return{Value: result}}
	return ir.Call{Function: f, Arguments: []ir.Expression{text}, Returns: ir.Boolean}, true, nil
}
