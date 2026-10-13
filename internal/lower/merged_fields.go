package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// The merged symbol exposes interface promises, not evidence of runtime storage.
func (l *lowering) mergedClass(node *ast.Node) *ast.Node {
	if node == nil || node.Name() == nil {
		return nil
	}
	symbol := l.checker.GetSymbolAtLocation(node.Name())
	if symbol == nil {
		return nil
	}
	var class *ast.Node
	merged := false
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindClassDeclaration {
			class = declaration
		}
		if declaration.Kind == ast.KindInterfaceDeclaration {
			merged = true
		}
	}
	if merged {
		return class
	}
	return nil
}

// Ordinary required class fields are subject to the checker's strict property
// initialization proof. declare and ! provide no such proof. Parameter properties
// and implemented accessors/methods are actual class members, including inherited ones.
func initializedClassMember(field *ast.Symbol) bool {
	for _, declaration := range field.Declarations {
		parent := declaration.Parent
		if parent == nil {
			continue
		}
		if declaration.Kind == ast.KindParameter && parameterProperty(declaration) {
			return true
		}
		if parent.Kind != ast.KindClassDeclaration && parent.Kind != ast.KindClassExpression {
			continue
		}
		if ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAmbient|ast.ModifierFlagsAbstract|ast.ModifierFlagsStatic) {
			continue
		}
		switch declaration.Kind {
		case ast.KindPropertyDeclaration:
			property := declaration.AsPropertyDeclaration()
			if property.Initializer != nil || (declaration.PostfixToken() == nil || declaration.PostfixToken().Kind != ast.KindExclamationToken) {
				return true
			}
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			if declaration.Body() != nil {
				return true
			}
		}
	}
	return false
}

func (l *lowering) mergedDeclarationRefusal(node *ast.Node) error {
	if node.Kind != ast.KindClassDeclaration && node.Kind != ast.KindInterfaceDeclaration {
		return nil
	}
	class := l.mergedClass(node)
	if class == nil {
		return nil
	}
	for _, field := range l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(class.Name())) {
		if field.Flags&ast.SymbolFlagsOptional != 0 || initializedClassMember(field) {
			continue
		}
		return &Refused{Where: l.program.Where(node), What: "class " + class.Name().Text() + " merged with an interface promises field " + field.Name + " that the class does not initialize", Fix: "declare and initialize " + field.Name + " in class " + class.Name().Text() + ", or use a separate interface with a checked view"}
	}
	return nil
}

func (l *lowering) mergedFieldRulings(module *ast.SourceFile) error {
	adamic := strings.HasSuffix(l.program.FileName(module), ".a")
	var found error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil {
			return true
		}
		if adamic {
			found = l.mergedDeclarationRefusal(node)
		} else {
			found = l.mergedFieldRead(node)
		}
		if found != nil {
			return true
		}
		return node.ForEachChild(visit)
	}
	module.AsNode().ForEachChild(visit)
	return found
}

// Reuse the checked-view presence, readiness and runtime type contract at each
// actual read. Holding or declaring the merged type needs no runtime promise.
func (l *lowering) mergedFieldRead(node *ast.Node) error {
	var field *ast.Symbol
	if node.Kind == ast.KindPropertyAccessExpression {
		field = l.checker.GetSymbolAtLocation(node.Name())
	} else if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		key := ast.SkipParentheses(access.ArgumentExpression)
		if key.Kind == ast.KindStringLiteral {
			field = l.checker.GetPropertyOfType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression)), key.Text())
		}
	}
	if field == nil || field.Flags&ast.SymbolFlagsOptional != 0 || initializedClassMember(field) {
		return nil
	}
	merged := false
	for _, declaration := range field.Declarations {
		if declaration.Parent != nil && declaration.Parent.Kind == ast.KindInterfaceDeclaration && l.mergedClass(declaration.Parent) != nil {
			merged = true
		}
	}
	if !merged {
		return nil
	}
	if _, err := l.strictViewContract(node, l.checker.GetTypeOfSymbol(field)); err != nil {
		return err
	}
	if l.result.CheckedFields == nil {
		l.result.CheckedFields = map[string]bool{}
	}
	l.result.CheckedFields[field.Name] = true
	return nil
}

// Constant bracket reads need the same contract as dotted reads. The ordinary
// finite-key path has no declaration-name node from which to recover that contract.
func (l *lowering) mergedComputedRead(node *ast.Node) (ir.Expression, bool, error) {
	if node.Kind != ast.KindElementAccessExpression {
		return nil, false, nil
	}
	access := node.AsElementAccessExpression()
	key := ast.SkipParentheses(access.ArgumentExpression)
	if key.Kind != ast.KindStringLiteral || !l.result.CheckedFields[key.Text()] {
		return nil, false, nil
	}
	field := l.checker.GetPropertyOfType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression)), key.Text())
	if field == nil || initializedClassMember(field) {
		return nil, false, nil
	}
	merged := false
	for _, declaration := range field.Declarations {
		if declaration.Parent != nil && declaration.Parent.Kind == ast.KindInterfaceDeclaration && l.mergedClass(declaration.Parent) != nil {
			merged = true
		}
	}
	if !merged {
		return nil, false, nil
	}
	object, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	declared := l.checker.GetTypeOfSymbol(field)
	of, known := l.representation(declared)
	if !known {
		return nil, true, l.notYet(node, "a merged field without runtime storage")
	}
	return ir.Property{Object: object, Name: field.Name, Of: of, Optional: access.QuestionDotToken != nil,
		View: sourceExpression(node), ViewType: l.checker.TypeToString(declared), ViewWhere: l.program.Where(node),
		ViewTypeID: int(declared.Id()), ViewContract: l.result.ViewContractTypes[int(declared.Id())],
		ViewReceiverTypeID: int(l.checker.GetTypeAtLocation(access.Expression).Id())}, true, nil
}
