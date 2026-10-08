package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"strings"
)

func (l *lowering) reflectionMembers(node *ast.Node, t *checker.Type) ([]ir.ReflectionMember, error) {
	if l.openNumericEnumType(t) {
		return []ir.ReflectionMember{{Kind: ir.Number}}, nil
	}
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		result := []ir.ReflectionMember{}
		for _, member := range t.Types() {
			values, err := l.reflectionMembers(node, member)
			if err != nil {
				return nil, err
			}
			result = append(result, values...)
		}
		return result, nil
	}
	member := ir.ReflectionMember{}
	switch {
	case t.Flags()&checker.TypeFlagsNumberLike != 0:
		member.Kind = ir.Number
		if t.Flags()&checker.TypeFlagsNumberLiteral != 0 {
			member.Literal = true
			member.Number = t.AsLiteralType().Value().(float64)
		}
	case t.Flags()&checker.TypeFlagsStringLike != 0:
		member.Kind = ir.String
		if t.Flags()&checker.TypeFlagsStringLiteral != 0 {
			member.Literal = true
			member.Text = t.AsLiteralType().Value().(string)
		}
	case t.Flags()&checker.TypeFlagsBooleanLike != 0:
		member.Kind = ir.Boolean
		if t.Flags()&checker.TypeFlagsBooleanLiteral != 0 {
			member.Literal = true
			member.Boolean = l.checker.TypeToString(t) == "true"
		}
	case t.Flags()&checker.TypeFlagsUndefined != 0:
		return nil, l.notYet(node, "Object reflection of undefined members without distinct null metadata")
	default:
		return nil, l.notYet(node, "Object reflection checking a non-primitive member "+l.checker.TypeToString(t))
	}
	return []ir.ReflectionMember{member}, nil
}

func (l *lowering) reflectionFields(node *ast.Node) ([]ir.ReflectionField, error) {
	fields := []ir.ReflectionField{}
	for _, field := range l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(node)) {
		members, err := l.reflectionMembers(node, l.checker.GetTypeOfSymbol(field))
		if err != nil {
			return nil, err
		}
		fields = append(fields, ir.ReflectionField{Name: field.Name, Members: members})
	}
	if indexed := l.checker.GetIndexTypeOfType(l.checker.GetTypeAtLocation(node), l.checker.GetStringType()); indexed != nil {
		members, err := l.reflectionMembers(node, indexed)
		if err != nil {
			return nil, err
		}
		fields = append(fields, ir.ReflectionField{Index: true, Members: members})
	}
	return fields, nil
}

func (l *lowering) ruledObjectCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	if name != "entries" && name != "assign" {
		return nil, false, nil
	}
	args := node.AsCallExpression().Arguments.Nodes
	if name == "entries" && len(args) != 1 || name == "assign" && len(args) < 1 {
		return nil, false, nil
	}
	typedSource := strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts")
	if !typedSource {
		allExact := true
		for _, arg := range args {
			if l.exactObject(arg, 0) {
				continue
			}
			allExact = false
			if !l.reflectionProducerKnown(arg, 0) {
				return nil, true, &Refused{Where: l.program.Where(arg), What: "Object." + name + " on an unproven shape", Fix: "use a fresh literal or a Record/index-signature value whose producers are all known"}
			}
		}
		if allExact && name == "assign" {
			return nil, false, nil
		}
	}
	for _, arg := range args {
		if err := l.reflectionProducerMetadata(arg); err != nil {
			return nil, true, err
		}
	}
	call := ir.ObjectCall{Method: name, Returns: ir.Array, Reflection: &ir.ObjectReflection{Message: "Object." + name + " slot check failed at " + l.program.Where(node), OwnPropertyOrder: true}}
	for _, arg := range args {
		value, err := l.expression(arg)
		if err != nil {
			return nil, true, err
		}
		if value.Type() != ir.Object || l.includesUndefined(l.checker.GetTypeAtLocation(arg)) || isClassInstance(l.checker.GetTypeAtLocation(arg)) {
			return nil, true, l.notYet(arg, "checked Object reflection on other than a present plain object")
		}
		call.Arguments = append(call.Arguments, value)
	}
	if name == "entries" {
		resultArgs := l.checker.GetTypeArguments(l.checker.GetTypeAtLocation(node))
		if len(resultArgs) != 1 {
			return nil, true, l.notYet(node, "Object.entries without a represented result")
		}
		pair := l.checker.GetTypeArguments(resultArgs[0])
		if len(pair) != 2 {
			return nil, true, l.notYet(node, "Object.entries without a represented pair")
		}
		members, err := l.reflectionMembers(node, pair[1])
		if err != nil {
			return nil, true, err
		}
		call.Reflection.Members = members
		element, known := l.representation(pair[1])
		if !known {
			return nil, true, l.notYet(node, "Object.entries value representation")
		}
		call.Element = element
	} else {
		call.Returns = ir.Object
		if l.checker.GetIndexTypeOfType(l.checker.GetTypeAtLocation(args[0]), l.checker.GetStringType()) != nil {
			return nil, true, l.notYet(args[0], "Object.assign to an open index-signature target needs growable own-property storage")
		}
		for _, field := range l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(args[0])) {
			of, known := l.representation(l.checker.GetTypeOfSymbol(field))
			if field.Flags&ast.SymbolFlagsOptional != 0 || !known || (of != ir.Number && of != ir.Boolean && of != ir.String) {
				return nil, true, l.notYet(args[0], "Object.assign target requires present scalar slots with stable representations")
			}
		}
		targets, err := l.reflectionFields(args[0])
		if err != nil {
			return nil, true, err
		}
		call.Reflection.Targets = targets
		for _, arg := range args[1:] {
			fields, err := l.reflectionFields(arg)
			if err != nil {
				return nil, true, err
			}
			call.Reflection.Sources = append(call.Reflection.Sources, fields)
		}
	}
	l.result.ReflectionChecks++
	return call, true, nil
}

// Explicit index/Record annotations do not erase the proof of an immutable literal producer.
func (l *lowering) reflectionProducerKnown(node *ast.Node, depth int) bool {
	if depth > 32 {
		return false
	}
	node = ast.SkipParentheses(node)
	if l.exactObject(node, 0) || l.reflectionLiteralShapeKnown(node) {
		return true
	}
	if node.Kind == ast.KindAsExpression {
		assertion := node.AsAsExpression()
		return assertion.Type.Kind == ast.KindTypeReference && assertion.Type.AsTypeReferenceNode().TypeName.Text() == "const" && l.reflectionProducerKnown(assertion.Expression, depth+1)
	}
	if node.Kind != ast.KindIdentifier {
		return false
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}
	variable := declaration.AsVariableDeclaration()
	// Restrict this extra proof to an index/Record view. Ordinary structural annotations retain refusal.
	if variable.Initializer == nil {
		return false
	}
	if variable.Type == nil {
		return l.reflectionProducerKnown(variable.Initializer, depth+1)
	}
	t := l.checker.GetTypeAtLocation(variable.Type)
	if l.checker.GetIndexTypeOfType(t, l.checker.GetStringType()) == nil {
		return false
	}
	return l.reflectionProducerKnown(variable.Initializer, depth+1)
}

func (l *lowering) reflectionIndexSignatureAllowed(node *ast.Node) bool {
	if strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts") {
		return true
	}
	if node.Parent.Kind != ast.KindTypeLiteral || node.Parent.Parent.Kind != ast.KindVariableDeclaration {
		return false
	}
	declaration := node.Parent.Parent
	variable := declaration.AsVariableDeclaration()
	return declaration.Parent.Flags&ast.NodeFlagsConst != 0 && variable.Type == node.Parent && variable.Initializer != nil && l.reflectionProducerKnown(variable.Initializer, 0)
}

// Runtime-created objects have their own presence/enumerability protocol. Until those
// owners publish reflection metadata, do not treat their physical slots as own keys.
func (l *lowering) reflectionProducerMetadata(node *ast.Node) error {
	target := l.checker.GetTypeAtLocation(node)
	var blocked *ast.Node
	var visit ast.Visitor
	visit = func(candidate *ast.Node) bool {
		if blocked != nil {
			return true
		}
		unsupported := false
		switch candidate.Kind {
		case ast.KindNewExpression, ast.KindRegularExpressionLiteral:
			unsupported = true
		case ast.KindObjectLiteralExpression:
			for _, property := range candidate.AsObjectLiteralExpression().Properties.Nodes {
				if property.Kind == ast.KindGetAccessor || property.Kind == ast.KindSetAccessor {
					unsupported = true
				}
			}
		case ast.KindCallExpression:
			callee := ast.SkipParentheses(candidate.AsCallExpression().Expression)
			if callee.Kind == ast.KindPropertyAccessExpression {
				access := callee.AsPropertyAccessExpression()
				if l.isLibraryGlobal(access.Expression, "Object") && (access.Name().Text() == "assign" || access.Name().Text() == "freeze") {
					candidate.ForEachChild(visit)
					return false
				}
			}
			if symbol := l.checker.GetSymbolAtLocation(callee); symbol != nil {
				for _, declaration := range symbol.Declarations {
					source := ast.GetSourceFileOfNode(declaration)
					if load.IsLibrary(source) || load.IsPrelude(source) {
						unsupported = true
					}
				}
			}
		}
		producerType := l.checker.GetTypeAtLocation(candidate)
		producerRepresentation, known := l.representation(producerType)
		if unsupported && known && producerRepresentation == ir.Object && l.checker.IsTypeAssignableTo(producerType, target) {
			blocked = candidate
			return true
		}
		candidate.ForEachChild(visit)
		return false
	}
	for _, file := range l.program.Files() {
		file.AsNode().ForEachChild(visit)
	}
	if blocked != nil {
		return l.notYet(node, "Object reflection requires own-property metadata for a compatible producer or shape-changing write at "+l.program.Where(blocked))
	}
	return nil
}

func (l *lowering) reflectionOwnNameKnown(node *ast.Node, key string, depth int) bool {
	if depth > 32 {
		return false
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindObjectLiteralExpression && l.exactObject(node, 0) {
		for _, field := range node.AsObjectLiteralExpression().Properties.Nodes {
			if field.Name().Text() == key {
				return true
			}
		}
		return false
	}
	if node.Kind == ast.KindAsExpression {
		assertion := node.AsAsExpression()
		return assertion.Type.Kind == ast.KindTypeReference && assertion.Type.AsTypeReferenceNode().TypeName.Text() == "const" && l.reflectionOwnNameKnown(assertion.Expression, key, depth+1)
	}
	if node.Kind != ast.KindIdentifier {
		return false
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	return declaration.Kind == ast.KindVariableDeclaration && declaration.Parent.Flags&ast.NodeFlagsConst != 0 && declaration.AsVariableDeclaration().Initializer != nil && l.reflectionOwnNameKnown(declaration.AsVariableDeclaration().Initializer, key, depth+1)
}

// Length-bearing reflection names admit NUL and # in ordinary literal string keys.
func (l *lowering) reflectionLiteralShapeKnown(node *ast.Node) bool {
	if node.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	for _, field := range node.AsObjectLiteralExpression().Properties.Nodes {
		if field.Kind != ast.KindPropertyAssignment && field.Kind != ast.KindShorthandPropertyAssignment {
			return false
		}
		if field.Name().Kind != ast.KindIdentifier && field.Name().Kind != ast.KindStringLiteral {
			return false
		}
		if field.Name().Text() == "__proto__" {
			return false
		}
	}
	return true
}

// Admitting the index-signature type for reflection must not admit missing native slots.
func (l *lowering) reflectionIndexAccess(node *ast.Node) error {
	var receiver *ast.Node
	key := ""
	if node.Kind == ast.KindPropertyAccessExpression {
		access := node.AsPropertyAccessExpression()
		receiver = access.Expression
		key = access.Name().Text()
	}
	if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		receiver = access.Expression
		if access.ArgumentExpression != nil && access.ArgumentExpression.Kind == ast.KindStringLiteral {
			key = access.ArgumentExpression.Text()
		}
	}
	if receiver == nil || l.checker.GetIndexTypeOfType(l.checker.GetTypeAtLocation(receiver), l.checker.GetStringType()) == nil {
		return nil
	}
	// RegExp named groups already have a checked dictionary protocol.
	if l.regexGroups(receiver) || l.reflectionOwnNameKnown(receiver, key, 0) {
		return nil
	}
	return l.notYet(node, "an index-signature slot without a known own literal producer (growable own-property storage is not represented)")
}
