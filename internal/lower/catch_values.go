package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// prepareCatchValues follows catch values into record fields before any body is
// lowered. A project may call those fields Error-like, but storing a thrown
// primitive must preserve its tag until a use checks the promised type.
func (l *lowering) prepareCatchValues() {
	if l.caught != nil {
		return
	}
	l.caught = map[*ast.Symbol]bool{}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		panic("lower: catch preparation lost the already checked module order")
	}
	visitAll := func(action func(*ast.Node)) {
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			action(node)
			node.ForEachChild(visit)
			return false
		}
		for _, module := range modules {
			module.AsNode().ForEachChild(visit)
		}
	}
	visitAll(func(node *ast.Node) {
		if node.Kind == ast.KindCatchClause {
			if declaration := node.AsCatchClause().VariableDeclaration; declaration != nil && ast.IsIdentifier(declaration.Name()) {
				l.caught[l.symbol(declaration.Name())] = true
			}
		}
	})
	for changed := true; changed; {
		changed = false
		visitAll(func(node *ast.Node) {
			if node.Kind != ast.KindObjectLiteralExpression {
				return
			}
			for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
				var value *ast.Node
				origin := false
				switch property.Kind {
				case ast.KindPropertyAssignment:
					value = property.AsPropertyAssignment().Initializer
				case ast.KindShorthandPropertyAssignment:
					value = property.Name()
					origin = l.catchSymbolOrigin(l.checker.GetShorthandAssignmentValueSymbol(property), map[*ast.Symbol]bool{})
				default:
					continue
				}
				if !origin && !l.catchOrigin(value, map[*ast.Symbol]bool{}) {
					continue
				}
				for _, object := range []*checker.Type{l.checker.GetTypeAtLocation(node), l.checker.GetContextualType(node, checker.ContextFlagsNone)} {
					if object == nil {
						continue
					}
					if field := l.checker.GetPropertyOfType(object, property.Name().Text()); field != nil && !l.caught[field] {
						if of, known := l.representation(l.checker.GetTypeOfSymbol(field)); known && (of == ir.String || of == ir.Number || of == ir.Boolean || of.IsMaybe()) {
							// Definite primitive fields check their initializer and retain
							// their declared layout, including later ordinary writes.
							continue
						}
						l.caught[field] = true
						changed = true
					}
				}
			}
		})
	}
}

func (l *lowering) catchOrigin(node *ast.Node, visited map[*ast.Symbol]bool) bool {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindPropertyAccessExpression {
		return l.caught[l.symbol(node.Name())] || l.catchOrigin(node.AsPropertyAccessExpression().Expression, visited)
	}
	if !ast.IsIdentifier(node) {
		return false
	}
	return l.catchSymbolOrigin(l.symbol(node), visited)
}

func (l *lowering) catchSymbolOrigin(symbol *ast.Symbol, visited map[*ast.Symbol]bool) bool {
	if symbol == nil || visited[symbol] {
		return false
	}
	if l.caught[symbol] {
		return true
	}
	visited[symbol] = true
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindVariableDeclaration {
			if value := declaration.AsVariableDeclaration().Initializer; value != nil && l.catchOrigin(value, visited) {
				return true
			}
		}
	}
	return false
}

// checkedCatchUse checks a catch-derived unknown only when its context requires
// a primitive or a nominal Error. Observations and storage keep the tagged value.
func (l *lowering) checkedCatchUse(node *ast.Node, value ir.Expression) ir.Expression {
	if value.Type() != ir.Union {
		return value
	}
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if comparedWithUndefined(node) || parent != nil && (parent.Kind == ast.KindTypeOfExpression || parent.Kind == ast.KindPropertyAccessExpression || parent.Kind == ast.KindShorthandPropertyAssignment) {
		return value
	}
	target := l.checker.GetTypeAtLocation(node)
	if target.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 || l.caught[l.symbol(node)] {
		if contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone); contextual != nil && contextual.Flags()&checker.TypeFlagsAny == 0 {
			target = contextual
		}
	}
	if l.isLibraryType(l.checker.GetNonNullableType(target), "Error") {
		b := l.libraryArrayBuilder([]ir.Expression{value})
		held := b.read(b.parameters[0])
		matches := ir.Expression(ir.InstanceOf{Value: held, Class: ir.ErrorClass})
		if l.includesUndefined(target) {
			matches = ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: held}, Right: matches}
		}
		b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant("adamic/catch-type: caught value is not an Error")}}}})
		return b.finish("checked_catch_error", ir.Narrow{Value: held, To: ir.Object})
	}
	if to, known := l.representation(target); known && (to == ir.String || to == ir.Number || to == ir.Boolean || to.IsMaybe()) {
		return ir.Narrow{Value: value, To: to, Checked: true, Optional: l.includesUndefined(target), Where: l.program.Where(node)}
	}
	return value
}

func (l *lowering) catchProperty(node *ast.Node) (ir.Expression, bool, error) {
	access := node.AsPropertyAccessExpression()
	if !l.catchOrigin(access.Expression, map[*ast.Symbol]bool{}) {
		return nil, false, nil
	}
	object, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	if object.Type() != ir.Union {
		return nil, false, nil
	}
	name := access.Name().Text()
	// The admitted dynamic fields are ordinary Error/host-record data fields.
	// Other names may denote primitive metadata or prototype methods; the
	// native own-field reader cannot implement those without the receiver type.
	if !supportedCaughtMember(name, l.caughtMemberNeedsLayout(name)) {
		return nil, true, l.notYet(node, "an unclassified caught property outside ordinary message or code data fields; narrow to the receiver's declared type before reading it")
	}
	if l.accessorNames[name] {
		return nil, true, l.notYet(node, "a dynamic catch property that may reach an accessor; narrow to its declared class before reading the accessor")
	}
	if access.QuestionDotToken != nil {
		return nil, true, l.notYet(node, "optional property access on an unclassified caught value; check for undefined first")
	}
	value := caughtMemberRead(object, ir.StringConstant{Index: l.constant(name)})
	return l.checkedCatchUse(node, value), true, nil
}

// caughtMemberRead is the policy boundary: reads preserve absence and only nullish
// receivers throw. Definite typed uses are checked separately by checkedCatchUse.
func caughtMemberRead(object, name ir.Expression) ir.Expression {
	return ir.ObjectCall{Method: "catchProperty", Arguments: []ir.Expression{object, name}, Returns: ir.Union}
}

// boxCaughtValue preserves the NULL meaning of a nullable source view when it
// enters a generic carrier. Undefined still uses NULL; actual null gets its tag.
func (l *lowering) boxCaughtValue(node *ast.Node, value ir.Expression) ir.Expression {
	if value.Type() == ir.Union {
		return value
	}
	proven := l.checker.GetTypeAtLocation(node)
	return ir.Box{Value: value, Null: l.includesNull(proven) && !l.includesUndefined(proven)}
}

func supportedCaughtMember(name string, method bool) bool {
	return (name == "message" || name == "code") && !method
}

// A class method lives in the dispatch table rather than an ordinary own-field
// slot. Static fields use a separate constructor layout too. Until these
// dynamic layouts can be read, refuse a possible target.
func (l *lowering) caughtMemberNeedsLayout(name string) bool {
	found := false
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if (node.Kind == ast.KindMethodDeclaration || node.Kind == ast.KindPropertyDeclaration && ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic)) && node.Name() != nil && (ast.IsIdentifier(node.Name()) || node.Name().Kind == ast.KindStringLiteral) && node.Name().Text() == name {
			found = true
			return true
		}
		return node.ForEachChild(visit)
	}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		panic("lower: catch member policy lost the already checked module order")
	}
	for _, file := range modules {
		file.AsNode().ForEachChild(visit)
		if found {
			break
		}
	}
	return found
}
