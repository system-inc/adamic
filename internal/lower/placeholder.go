package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// placeholderDeclaration admits only literal slot initializers. A shadowed
// undefined binding is an ordinary assertion, as are calls and standalone values.
func (l *lowering) placeholderDeclaration(node *ast.Node) bool {
	if !l.uninitializedInitializer(node) {
		return false
	}
	for node.Parent != nil && (node.Parent.Kind == ast.KindParenthesizedExpression || node.Parent.Kind == ast.KindAsExpression) {
		node = node.Parent
	}
	if node.Parent == nil {
		return false
	}
	switch parent := node.Parent; parent.Kind {
	case ast.KindVariableDeclaration:
		return parent.AsVariableDeclaration().Initializer == node
	case ast.KindPropertyDeclaration:
		return parent.AsPropertyDeclaration().Initializer == node
	case ast.KindParameter:
		return parent.AsParameterDeclaration().Initializer == node
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Initializer == node
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		return binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Right == node
	}
	return false
}

// placeholderZero is the observable undefined payload of an unset slot.
func placeholderZero(of ir.Type) ir.Expression {
	if of.IsReference() {
		return ir.Undefined{Of: of}
	}
	if of == ir.MaybeBoolean {
		return ir.MaybeOf{Of: of}
	}
	return zeroValue(of)
}

// Placeholder storage uses the existing counted union: NULL is undefined,
// adamic_null is null, and each present arm retains its ordinary ownership.
func placeholderStorage(of ir.Type) ir.Type {
	if of == 0 {
		return 0
	}
	return ir.Union
}

func (l *lowering) placeholderInitialValue(node *ast.Node, of ir.Type) ir.Expression {
	if node != nil && l.uninitializedInitializer(node) {
		node = ast.SkipParentheses(node)
		for node.Kind == ast.KindAsExpression {
			node = ast.SkipParentheses(node.AsAsExpression().Expression)
		}
		if ast.SkipParentheses(node.AsNonNullExpression().Expression).Kind == ast.KindNullKeyword {
			return fit(ir.Null{}, of)
		}
	}
	return placeholderZero(of)
}

func (l *lowering) placeholderOptional(of *checker.Type) bool {
	return l.includesUndefined(of) || l.includesNull(of)
}

// Only a placeholder compared with literal null gets the ruling's loose
// nullish comparison. General coercing equality retains its refusal.
func (l *lowering) placeholderNullishTest(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind != ast.KindBinaryExpression {
		node = node.Parent
	}
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := node.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindEqualsEqualsToken && binary.OperatorToken.Kind != ast.KindExclamationEqualsToken {
		return false
	}
	left, right := ast.SkipParentheses(binary.Left), ast.SkipParentheses(binary.Right)
	return left.Kind == ast.KindNullKeyword && l.placeholderReadOrigin(right) != "" || right.Kind == ast.KindNullKeyword && l.placeholderReadOrigin(left) != ""
}

func (l *lowering) placeholderKey(symbol *ast.Symbol) string {
	if symbol == nil || len(symbol.Declarations) == 0 {
		return ""
	}
	declaration := symbol.Declarations[0]
	return l.program.Where(declaration) + ":" + symbol.Name
}

func (l *lowering) placeholderSymbolOrigin(symbol *ast.Symbol) string {
	return l.result.PlaceholderSources[l.placeholderKey(symbol)]
}

func (l *lowering) placeholderOrigin(node *ast.Node) string {
	if node == nil {
		return ""
	}
	return l.placeholderSymbolOrigin(l.symbol(ast.SkipParentheses(node)))
}

// The transfer graph is conservative across branches and loops. A copied local
// never loses its origin when another path writes it. Readiness still proves
// individual writes using the existing CFG; every remaining use as T is checked.
func (l *lowering) notePlaceholderSlots() {
	// The refusal pass is also used independently of IR construction.
	if l.result == nil {
		l.result = &ir.Program{}
	}
	if l.result.PlaceholderSources != nil {
		return
	}
	l.result.PlaceholderSources = map[string]string{}
	l.result.PlaceholderViews = map[string]string{}
	type transfer struct {
		to   string
		from *ast.Node
	}
	var transfers []transfer
	var views []transfer
	inferred := map[string]bool{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsPartOfTypeNode(node) {
			return false
		}
		var target, value *ast.Node
		switch node.Kind {
		case ast.KindVariableDeclaration:
			target, value = node.Name(), node.AsVariableDeclaration().Initializer
			if ast.IsIdentifier(target) {
				key := l.placeholderKey(l.symbol(target))
				inferred[key] = node.AsVariableDeclaration().Type == nil || l.placeholderOptional(l.checker.GetTypeOfSymbol(l.symbol(target)))
				if value != nil {
					transfers = append(transfers, transfer{key, value})
				}
			}
		case ast.KindPropertyDeclaration:
			target, value = node.Name(), node.AsPropertyDeclaration().Initializer
			key := l.placeholderKey(l.symbol(target))
			inferred[key] = l.placeholderOptional(l.checker.GetTypeOfSymbol(l.symbol(target)))
			if value != nil {
				transfers = append(transfers, transfer{key, value})
			}
		case ast.KindParameter:
			target, value = node.Name(), node.AsParameterDeclaration().Initializer
		case ast.KindCallExpression:
			if signature := l.checker.GetResolvedSignature(node); signature != nil {
				parameters := signature.Parameters()
				for index, argument := range nodesOf(node.AsCallExpression().Arguments) {
					if index < len(parameters) && l.placeholderOptional(l.checker.GetTypeOfSymbol(parameters[index])) {
						key := l.placeholderKey(parameters[index])
						inferred[key] = true
						transfers = append(transfers, transfer{key, argument})
					}
				}
			}
		case ast.KindPropertyAssignment:
			target, value = node.Name(), node.AsPropertyAssignment().Initializer
			key := l.placeholderKey(l.symbol(target))
			inferred[key] = l.placeholderOptional(l.checker.GetTypeOfSymbol(l.symbol(target)))
			transfers = append(transfers, transfer{key, value})
		case ast.KindBinaryExpression:
			binary := node.AsBinaryExpression()
			if binary.OperatorToken.Kind == ast.KindEqualsToken {
				target, value = ast.SkipParentheses(binary.Left), binary.Right
				if ast.IsIdentifier(target) || target.Kind == ast.KindPropertyAccessExpression {
					if target.Kind == ast.KindPropertyAccessExpression {
						key := l.placeholderKey(l.symbol(target))
						inferred[key] = l.placeholderOptional(l.checker.GetTypeOfSymbol(l.symbol(target)))
					}
					transfers = append(transfers, transfer{l.placeholderKey(l.symbol(target)), value})
				}
			}
		}
		if node.Kind == ast.KindObjectLiteralExpression {
			properties := node.AsObjectLiteralExpression().Properties.Nodes
			if len(properties) > 0 && properties[0].Kind == ast.KindSpreadAssignment {
				source := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(properties[0].AsSpreadAssignment().Expression))
				target := l.checker.GetTypeAtLocation(node)
				for _, field := range l.checker.GetPropertiesOfType(target) {
					if from := l.checker.GetPropertyOfType(source, field.Name); from != nil && len(from.Declarations) > 0 {
						views = append(views, transfer{l.placeholderKey(field), from.Declarations[0].Name()})
					}
				}
			}
		}
		if node.Parent != nil && ast.IsExpression(node) && !ast.IsDeclarationName(node) && !ast.IsPartOfTypeNode(node) {
			if targetType := l.checker.GetContextualType(node, checker.ContextFlagsNone); targetType != nil {
				sourceType := l.checker.GetTypeAtLocation(node)
				if sourceType.Flags()&checker.TypeFlagsObject != 0 && targetType.Flags()&checker.TypeFlagsObject != 0 {
					for _, field := range l.checker.GetPropertiesOfType(targetType) {
						if source := l.checker.GetPropertyOfType(sourceType, field.Name); source != nil && len(source.Declarations) > 0 {
							views = append(views, transfer{l.placeholderKey(field), source.Declarations[0].Name()})
						}
					}
				}
			}
		}
		if target != nil && l.uninitializedInitializer(value) {
			key := l.placeholderKey(l.symbol(target))
			if key != "" && l.result.PlaceholderSources[key] == "" {
				l.result.PlaceholderSources[key] = sourceExpression(target)
			}
		}
		return node.ForEachChild(visit)
	}
	for _, file := range l.program.CompilerProgram().GetSourceFiles() {
		if !load.IsLibrary(file) && !load.IsPrelude(file) {
			visit(file.AsNode())
		}
	}
	var origins func(*ast.Node) string
	origins = func(node *ast.Node) string {
		if node == nil {
			return ""
		}
		node = ast.SkipParentheses(node)
		if node.Kind == ast.KindConditionalExpression {
			conditional := node.AsConditionalExpression()
			if origin := origins(conditional.WhenTrue); origin != "" {
				return origin
			}
			return origins(conditional.WhenFalse)
		}
		if node.Kind == ast.KindAsExpression {
			return origins(node.AsAsExpression().Expression)
		}
		if ast.IsIdentifier(node) || node.Kind == ast.KindPropertyAccessExpression {
			return l.placeholderReadOrigin(node)
		}
		return ""
	}
	for changed := true; changed; {
		changed = false
		for _, edge := range views {
			if edge.to == "" || l.result.PlaceholderSources[edge.to] != "" || l.result.PlaceholderViews[edge.to] != "" {
				continue
			}
			if origin := origins(edge.from); origin != "" {
				l.result.PlaceholderViews[edge.to] = origin
				changed = true
			}
		}
		for _, edge := range transfers {
			if edge.to == "" || l.result.PlaceholderSources[edge.to] != "" || !inferred[edge.to] {
				continue
			}
			if origin := origins(edge.from); origin != "" {
				l.result.PlaceholderSources[edge.to] = origin
				changed = true
			}
		}
	}
}

// placeholderAllowsUnset is deliberately narrower than acceptsUndefined: a
// template or a container element is a use as T, even when Node tolerates it.
func (l *lowering) placeholderAllowsUnset(node *ast.Node) bool {
	if comparedWithUndefined(node) || l.placeholderNullishTest(node.Parent) {
		return true
	}
	for node.Parent != nil && (node.Parent.Kind == ast.KindParenthesizedExpression || node.Parent.Kind == ast.KindAsExpression) {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindIfStatement:
		return parent.AsIfStatement().Expression == node
	case ast.KindWhileStatement:
		return parent.AsWhileStatement().Expression == node
	case ast.KindDoStatement:
		return parent.AsDoStatement().Expression == node
	case ast.KindTypeOfExpression:
		return true
	case ast.KindPrefixUnaryExpression:
		return parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	case ast.KindConditionalExpression:
		if parent.AsConditionalExpression().Condition == node {
			return true
		}
		return l.placeholderAllowsUnset(parent)
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Right == node && ast.SkipParentheses(binary.Left).Kind == ast.KindElementAccessExpression {
			return false
		}
		if binary.Left == node && (binary.OperatorToken.Kind == ast.KindQuestionQuestionToken || binary.OperatorToken.Kind == ast.KindAmpersandAmpersandToken || binary.OperatorToken.Kind == ast.KindBarBarToken) {
			return true
		}
		if binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Right == node && l.placeholderOrigin(binary.Left) != "" {
			return true
		}
		if binary.OperatorToken.Kind == ast.KindAmpersandAmpersandToken || binary.OperatorToken.Kind == ast.KindBarBarToken {
			return l.placeholderAllowsUnset(parent)
		}
	case ast.KindVariableDeclaration:
		return parent.AsVariableDeclaration().Initializer == node && l.placeholderOrigin(parent.Name()) != ""
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Initializer == node && l.placeholderOrigin(parent.Name()) != ""
	case ast.KindCallExpression:
		callee := ast.SkipParentheses(parent.AsCallExpression().Expression)
		if callee.Kind == ast.KindPropertyAccessExpression {
			receiver := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(callee.AsPropertyAccessExpression().Expression))
			if l.checker.IsArrayType(receiver) || l.isLibraryType(receiver, "Map", "Set") {
				switch callee.Name().Text() {
				case "set", "add", "push", "unshift", "splice", "fill":
					return false
				}
			}
		}
	case ast.KindArrayLiteralExpression, ast.KindSpreadElement:
		return false
	}
	contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	return contextual != nil && l.placeholderOptional(contextual)
}

func (l *lowering) placeholderRead(node *ast.Node, value ir.Expression, origin string) ir.Expression {
	if l.placeholderAllowsUnset(node) {
		return value
	}
	use := "use as T"
	at := node
	for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
		at = at.Parent
	}
	if parent := at.Parent; parent != nil {
		switch parent.Kind {
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			use = "property read"
		case ast.KindCallExpression:
			use = "argument"
			if parent.AsCallExpression().Expression == at {
				use = "call"
			}
		case ast.KindReturnStatement:
			use = "return"
		case ast.KindVariableDeclaration, ast.KindBinaryExpression, ast.KindPropertyAssignment:
			use = "assignment"
		case ast.KindArrayLiteralExpression:
			use = "container element"
		}
	}
	of, known := l.representation(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node)))
	if !known {
		of = ir.Union
	}
	checkArm := false
	if symbol := l.symbol(node); symbol != nil {
		declared, _ := l.representation(l.checker.GetNonNullableType(l.checker.GetTypeOfSymbol(symbol)))
		checkArm = declared == ir.Union && of != ir.Union
		if checkArm && (of == ir.Weak || of.IsTypedArray()) && l.unlowerable == nil {
			l.unlowerable = l.notYet(node, "a placeholder narrowing without a present-arm storage check")
		}
	}
	return ir.PlaceholderUse{Value: value, Origin: origin, Use: use, Path: sourceExpression(node), Where: l.program.Where(node), CheckArm: checkArm, Of: of}
}

func placeholderProven(value ir.Expression, program *ir.Program, ready []bool, fields map[fieldReadiness]int) bool {
	if value == nil {
		return false
	}
	switch value := value.(type) {
	case ir.Read:
		return program.Locals[value.Local].Placeholder != "" && ready != nil && ready[value.Local]
	case ir.Property:
		if local, ok := readinessObject(value.Object); ok {
			if slot, found := fields[fieldReadiness{local, value.Name}]; found {
				return ready[slot]
			}
		}
		return false
	case ir.Box:
		return placeholderProven(value.Value, program, ready, fields)
	case ir.MaybeOf:
		return value.Value != nil
	case ir.Undefined, ir.Null:
		return false
	case ir.PlaceholderUse:
		return true // Its checked boundary produces T or stops.
	case ir.Conditional:
		return placeholderProven(value.WhenTrue, program, ready, fields) && placeholderProven(value.WhenNot, program, ready, fields)
	case ir.Coalesce:
		return value.Panic != nil || placeholderProven(value.Fallback, program, ready, fields)
	}
	return !value.Type().IsMaybe() && value.Type() != ir.Union
}

func placeholderUseValue(use ir.PlaceholderUse, program *ir.Program, ready []bool, fields map[fieldReadiness]int) ir.Expression {
	proven := placeholderProven(use.Value, program, ready, fields)
	status := "checked"
	if proven {
		status = "proven"
	}
	program.PlaceholderChecks = append(program.PlaceholderChecks, ir.PlaceholderCheck{Origin: use.Origin, Use: use.Use, Path: use.Path, Where: use.Where, Status: status})
	armMessage := "placeholder '" + use.Origin + "' has the wrong present arm at " + use.Use + " via " + use.Path
	if use.CheckArm {
		program.PlaceholderChecks = append(program.PlaceholderChecks, ir.PlaceholderCheck{Origin: use.Origin, Use: "present arm at " + use.Use, Path: use.Path, Where: use.Where, Status: "checked"})
	}
	presentValue := func(value ir.Expression) ir.Expression {
		if value.Type() == ir.Union && use.Of != ir.Union {
			return ir.Narrow{Value: value, To: use.Of, Checked: use.CheckArm, Message: armMessage}
		}
		return value
	}
	if proven {
		if use.Value.Type().IsMaybe() {
			return ir.Unwrap{Value: use.Value, Proven: true}
		}
		return presentValue(use.Value)
	}
	message := "placeholder '" + use.Origin + "' is unset at " + use.Use + " via " + use.Path
	index := len(program.Strings)
	program.Strings = append(program.Strings, message)
	return presentValue(ir.Coalesce{Value: use.Value, Panic: ir.StringConstant{Index: index}, Of: use.Value.Type().Present()})
}

// Structural views preserve the source slot's optional storage but do not make
// the target's ordinary T fields legal destinations for an unset write.
func (l *lowering) placeholderReadOrigin(node *ast.Node) string {
	if origin := l.placeholderOrigin(node); origin != "" {
		return origin
	}
	return l.result.PlaceholderViews[l.placeholderKey(l.symbol(node))]
}
