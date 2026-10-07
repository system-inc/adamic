package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Whole numeric enum tags can overlap. A matching tag proves the selected
// payload only after its actual storage has been checked. Unchanged fields
// retain the union's existing promises.
func (l *lowering) enumTagViewFields(node *ast.Node, declared, observed *checker.Type) ([]*ast.Symbol, error) {
	fields := []*ast.Symbol{}
	for _, field := range l.checker.GetPropertiesOfType(observed) {
		target := l.checker.GetTypeOfSymbol(field)
		changed := false
		for _, member := range declared.Types() {
			actual := l.checker.GetPropertyOfType(member, field.Name)
			if actual == nil || !l.checker.IsTypeAssignableTo(l.checker.GetTypeOfSymbol(actual), target) || !l.enumAssignable(l.checker.GetTypeOfSymbol(actual), target) {
				changed = true
			}
		}
		if !changed {
			continue
		}
		of, known := l.representation(target)
		if !known || (of != ir.Number && of != ir.String && of != ir.Boolean) || field.Flags&ast.SymbolFlagsOptional != 0 || l.includesUndefined(target) || l.includesNull(target) {
			return nil, l.notYet(node, "a checked numeric enum object view with an incompatible structured or optional payload field "+field.Name)
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func (l *lowering) enumTagView(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	if remainder, _, _ := l.enumObjectRemainder(node); remainder != nil {
		return value, nil
	}
	symbol := l.flagValueSymbol(node)
	if symbol == nil || value.Type() != ir.Object {
		return value, nil
	}
	declared, observed := l.checker.GetTypeOfSymbol(symbol), l.checker.GetTypeAtLocation(node)
	// Comparing the tag observes a number; it does not read a variant payload.
	// Excluding one whole-enum member must not check a payload before the next
	// equality has even run.
	if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.AsPropertyAccessExpression().Expression == node && declared.Flags()&checker.TypeFlagsUnion != 0 && observed.Flags()&checker.TypeFlagsObject != 0 && l.enumTagDiscriminant(declared, observed, node.Parent.Name().Text()) {
		return value, nil
	}
	return l.enumTagViewAs(node, value, declared, observed)
}

func (l *lowering) enumTagViewAs(node *ast.Node, value ir.Expression, declared, observed *checker.Type) (ir.Expression, error) {
	if declared.Flags()&checker.TypeFlagsUnion == 0 || observed.Flags()&checker.TypeFlagsObject == 0 || declared == observed {
		return value, nil
	}
	needed := false
	for _, field := range l.checker.GetPropertiesOfType(observed) {
		needed = needed || l.enumTagDiscriminant(declared, observed, field.Name)
	}
	if !needed {
		return value, nil
	}
	if isClassInstance(observed) {
		return l.checkedClassCast(node, value, []*checker.Type{observed}, "numeric enum object view failed: class")
	}
	fields, err := l.enumTagViewFields(node, declared, observed)
	if err != nil || len(fields) == 0 {
		return value, err
	}
	b := l.libraryArrayBuilder([]ir.Expression{value})
	object := b.read(b.parameters[0])
	for _, field := range fields {
		target := l.checker.GetTypeOfSymbol(field)
		of, _ := l.representation(target)
		message := "numeric enum object view failed: field " + field.Name
		read := ir.Property{Object: object, Name: field.Name, Of: of, CheckMessage: message}
		// Checking storage precedes checking literal values. A singleton whole enum
		// is open, so its one declared value is never a closed payload promise.
		if !l.openNumericEnumType(target) {
			parts := castMembers(target)
			var matches ir.Expression
			for _, part := range parts {
				constant, _, known := l.literalConstant(part)
				if !known {
					matches = nil
					break
				}
				test := ir.Binary{Operator: ir.Equal, Left: read, Right: constant}
				if matches == nil {
					matches = test
				} else {
					matches = ir.Binary{Operator: ir.Or, Left: matches, Right: test}
				}
			}
			if matches != nil {
				b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}})
				continue
			}
		}
		b.body = append(b.body, ir.Evaluate{Value: read})
	}
	return b.finish("enum_tag_view", object), nil
}

// Record values excluded on this path, rather than trusting the checker's
// closed-enum remainder. Aliased members contribute the same numeric value.
func (l *lowering) enumTagValuesExcluded(node *ast.Node, symbol *ast.Symbol, field string, stored *checker.Type) bool {
	if symbol == nil {
		return false
	}
	excluded := map[any]bool{}
	add := func(condition *ast.Node, equalExcluded bool) {
		condition = ast.SkipParentheses(condition)
		if condition.Kind != ast.KindBinaryExpression {
			return
		}
		binary := condition.AsBinaryExpression()
		operator := binary.OperatorToken.Kind
		if (equalExcluded && operator != ast.KindEqualsEqualsEqualsToken) || (!equalExcluded && operator != ast.KindExclamationEqualsEqualsToken) {
			return
		}
		access, other := ast.SkipParentheses(binary.Left), binary.Right
		if access.Kind != ast.KindPropertyAccessExpression {
			access, other = ast.SkipParentheses(binary.Right), binary.Left
		}
		if access.Kind != ast.KindPropertyAccessExpression || access.Name().Text() != field || !l.sameEnumSubject(access.AsPropertyAccessExpression().Expression, node) {
			return
		}
		if member := l.enumMember(other); member != nil {
			excluded[l.checker.GetConstantValue(member)] = true
		} else if literal := l.checker.GetTypeAtLocation(other); literal.Flags()&checker.TypeFlagsNumberLiteral != 0 {
			excluded[literal.AsLiteralType().Value()] = true
		}
	}
	for child := node; child.Parent != nil; child = child.Parent {
		parent := child.Parent
		if ast.IsFunctionLike(parent) {
			break
		}
		if parent.Kind == ast.KindIfStatement {
			branch := parent.AsIfStatement()
			if branch.ElseStatement == child {
				add(branch.Expression, true)
			}
			if branch.ThenStatement == child {
				add(branch.Expression, false)
			}
		}
		if parent.Kind == ast.KindBlock {
			for _, previous := range parent.AsBlock().Statements.Nodes {
				if previous == child {
					break
				}
				if previous.Kind == ast.KindIfStatement && enumBranchReturns(previous.AsIfStatement().ThenStatement) {
					add(previous.AsIfStatement().Expression, true)
				}
			}
		}
		if parent.Kind == ast.KindDefaultClause {
			switched := parent.Parent.Parent
			if switched.Kind == ast.KindSwitchStatement {
				expression := ast.SkipParentheses(switched.AsSwitchStatement().Expression)
				if expression.Kind == ast.KindPropertyAccessExpression && expression.Name().Text() == field && l.sameEnumSubject(expression.AsPropertyAccessExpression().Expression, node) {
					for _, clause := range switched.AsSwitchStatement().CaseBlock.AsCaseBlock().Clauses.Nodes {
						if clause.Kind == ast.KindCaseClause {
							test := clause.AsCaseOrDefaultClause().Expression
							if member := l.enumMember(test); member != nil {
								excluded[l.checker.GetConstantValue(member)] = true
							}
						}
					}
				}
			}
		}
	}
	for _, member := range stored.Types() {
		property := l.checker.GetPropertyOfType(member, field)
		if property == nil {
			return false
		}
		identity := l.enumIdentity(l.checker.GetTypeOfSymbol(property))
		if !l.numericEnum(identity) {
			return false
		}
		for _, value := range identity.ValueDeclaration.AsEnumDeclaration().Members.Nodes {
			if !excluded[l.checker.GetConstantValue(value)] {
				return false
			}
		}
	}
	return len(excluded) > 0
}

func enumBranchReturns(node *ast.Node) bool {
	if node.Kind == ast.KindReturnStatement || node.Kind == ast.KindThrowStatement {
		return true
	}
	if node.Kind == ast.KindBlock {
		statements := node.AsBlock().Statements.Nodes
		return len(statements) > 0 && enumBranchReturns(statements[len(statements)-1])
	}
	return false
}

// The checker uses the same Type for a singleton whole enum and its member.
// Recover an explicitly written member slot from its annotation so a literal
// tag promise cannot admit arbitrary numbers during construction.
func (l *lowering) enumMemberSlot(node *ast.Node) *ast.Node {
	child := node
	for child.Parent != nil && child.Parent.Kind == ast.KindParenthesizedExpression {
		child = child.Parent
	}
	parent := child.Parent
	if parent == nil {
		return nil
	}
	var annotation *ast.Node
	switch parent.Kind {
	case ast.KindVariableDeclaration, ast.KindPropertyDeclaration, ast.KindParameter:
		annotation = parent.Type()
	case ast.KindPropertyAssignment:
		if target := l.checker.GetContextualType(parent.Parent, checker.ContextFlagsNone); target != nil {
			name, known := l.methodName(parent)
			if !known {
				return nil
			}
			if field := l.checker.GetPropertyOfType(target, name); field != nil && len(field.Declarations) > 0 {
				annotation = field.Declarations[0].Type()
			}
		}
	case ast.KindBinaryExpression:
		if binary := parent.AsBinaryExpression(); binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Right == child {
			if target := l.flagValueSymbol(binary.Left); target != nil && len(target.Declarations) > 0 {
				annotation = target.Declarations[0].Type()
			}
		}
	case ast.KindAsExpression:
		if parent.AsAsExpression().Expression == child {
			annotation = parent.AsAsExpression().Type
		}
	case ast.KindReturnStatement:
		for function := parent.Parent; function != nil; function = function.Parent {
			if ast.IsFunctionLike(function) {
				annotation = function.Type()
				break
			}
		}
	}
	if annotation == nil || annotation.Kind != ast.KindTypeReference {
		return nil
	}
	symbol := l.symbol(annotation.AsTypeReferenceNode().TypeName)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsEnumMember != 0 {
		return symbol.ValueDeclaration
	}
	return nil
}

// Annotation identity also survives structural views of objects and container
// elements. A mutable container must preserve the promise in both directions.
func (l *lowering) enumMemberPromises(from, to *checker.Type, seen map[[2]*checker.Type]bool) bool {
	if from == nil || to == nil || from == to || seen[[2]*checker.Type{from, to}] {
		return true
	}
	seen[[2]*checker.Type{from, to}] = true
	defer delete(seen, [2]*checker.Type{from, to})
	if from.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range from.Types() {
			if !l.enumMemberPromises(member, to, seen) {
				return false
			}
		}
		return true
	}
	if to.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range to.Types() {
			if l.checker.IsTypeAssignableTo(from, member) && l.enumMemberPromises(from, member, seen) {
				return true
			}
		}
		return false
	}
	for _, target := range l.containers(to) {
		for _, source := range l.containers(from) {
			if !l.sameContainer(source, target) {
				continue
			}
			a, b := l.typeArguments(source), l.typeArguments(target)
			for i := 0; i < len(a) && i < len(b); i++ {
				if !l.enumMemberPromises(a[i], b[i], seen) {
					return false
				}
			}
		}
	}
	if from.Flags()&checker.TypeFlagsObject == 0 || to.Flags()&checker.TypeFlagsObject == 0 {
		return true
	}
	for _, field := range l.checker.GetPropertiesOfType(to) {
		actual := l.checker.GetPropertyOfType(from, field.Name)
		if actual == nil {
			continue
		}
		target := l.enumAnnotatedMember(field)
		if target != nil && l.openNumericEnumType(l.checker.GetTypeOfSymbol(field)) && l.enumAnnotatedMember(actual) != target {
			return false
		}
		if !l.enumMemberPromises(l.checker.GetTypeOfSymbol(actual), l.checker.GetTypeOfSymbol(field), seen) {
			return false
		}
	}
	return true
}

// Property symbols belong to declarations, so holderA.node and holderB.node
// can share a symbol without being the same discriminant read.
func (l *lowering) sameEnumSubject(a, b *ast.Node) bool {
	a, b = ast.SkipParentheses(a), ast.SkipParentheses(b)
	if a.Kind != b.Kind {
		return false
	}
	if ast.IsIdentifier(a) {
		return l.flagValueSymbol(a) == l.flagValueSymbol(b)
	}
	if a.Kind == ast.KindPropertyAccessExpression {
		return a.Name().Text() == b.Name().Text() && l.sameEnumSubject(a.AsPropertyAccessExpression().Expression, b.AsPropertyAccessExpression().Expression)
	}
	return false
}
