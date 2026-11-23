package lower

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
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

func (l *lowering) mergedFieldRulings(module *ast.SourceFile) error {
	if strings.HasSuffix(l.program.FileName(module), ".a") {
		return nil
	}
	var found error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil {
			return true
		}
		found = l.mergedFieldRead(node)
		if found != nil {
			return true
		}
		return node.ForEachChild(visit)
	}
	module.AsNode().ForEachChild(visit)
	return found
}

// classMemberSlot is the storage decision shared by class layout and optional
// write analysis. A merge with unknown storage remembers its declaring class.
type classMemberSlot struct {
	holds, declared bool
	merge           *ast.Node
	of              ir.Type
	// known says a merge's member has a representation, so a merge that holds no slot was refused one
	// because it can point onward (mergedLeaf), not because its storage is unknown.
	known bool
}

// mergedLeaf reports whether a merged member's slot holds nothing that points onward. A merged slot is
// reserved Absent, which sends the class's construction down emit_objects.go's constructionFields, and a
// construction object is never adopted into a graph region: that was sound while construction slots held
// only scalars (construction.go's reserveFactoryFields), and it stays sound only while a merged slot can't
// reach another object, since a cycle through it would never be freed (rulings.md's no collector, Language's
// c01 on #86cmqrp). A string is immutable, so it's a leaf; a union may hold an object, so it isn't.
func mergedLeaf(of ir.Type) bool {
	switch of {
	case ir.Number, ir.Boolean, ir.String, ir.MaybeNumber, ir.MaybeBoolean, ir.NullRepresentation, ir.UndefinedRepresentation:
		return true
	}
	return false
}

func (l *lowering) classSlot(class *checker.Type, name string) classMemberSlot {
	field := l.checker.GetPropertyOfType(class, name)
	if field == nil {
		return classMemberSlot{}
	}
	// The checker can report a class's field via a merged interface's
	// extends clause. Only the runtime class and its actual base own storage.
	symbol := class.Symbol()
	if symbol == nil {
		return classMemberSlot{}
	}
	for _, declaration := range field.Declarations {
		owner := declaration.Parent
		if declaration.Kind == ast.KindParameter && parameterProperty(declaration) {
			if owner != nil && slices.Contains(symbol.Declarations, owner.Parent) {
				return classMemberSlot{holds: true, declared: true}
			}
			continue
		}
		if owner != nil && slices.Contains(symbol.Declarations, owner) &&
			(owner.Kind == ast.KindClassDeclaration || owner.Kind == ast.KindClassExpression) {
			return classMemberSlot{holds: declaration.Kind == ast.KindPropertyDeclaration, declared: true}
		}
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindClassDeclaration && declaration.Kind != ast.KindClassExpression {
			continue
		}
		if runtimeBase := l.staticBase(declaration); runtimeBase != nil {
			for _, base := range l.classBases(class) {
				if l.classBaseDeclaration(base.Symbol()) != runtimeBase {
					continue
				}
				if slot := l.classSlot(base, name); slot.declared || slot.merge != nil {
					return slot
				}
			}
		}
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindClassDeclaration {
			continue
		}
		file := ast.GetSourceFileOfNode(declaration)
		// A library class's declaration merge describes host storage. Only a
		// program-owned class has an instance layout that can reserve a slot.
		if load.IsLibrary(file) || load.IsNodeLibrary(file) ||
			!strings.HasSuffix(l.program.FileName(file), ".ts") || l.mergedClass(declaration) == nil {
			continue
		}
		of, known := l.representation(l.checker.GetTypeOfSymbol(field))
		return classMemberSlot{holds: known && mergedLeaf(of), merge: declaration, of: of, known: known}
	}
	return classMemberSlot{}
}

// reserveMergedFields adds storage for the instance members introduced by this
// interface merge. The constructor has no initializer for these slots.
func (l *lowering) reserveMergedFields(declaration *ast.Node, fields []ir.Field) []ir.Field {
	if l.mergedClass(declaration) == nil {
		return fields
	}
	for _, field := range l.checker.GetPropertiesOfType(l.classType) {
		if slices.ContainsFunc(fields, func(slot ir.Field) bool { return slot.Name == field.Name }) {
			continue
		}
		slot := l.classSlot(l.classType, field.Name)
		if slot.merge != declaration {
			continue
		}
		if !slot.holds {
			if l.missingMergedSlots == nil {
				l.missingMergedSlots = map[int]map[string]string{}
			}
			class := l.instance.class
			if l.missingMergedSlots[class] == nil {
				l.missingMergedSlots[class] = map[string]string{}
			}
			l.missingMergedSlots[class][field.Name] = missingMergedSlot(declaration, field.Name, slot.known)
			continue
		}
		value := zeroValue(slot.of)
		if slot.of.IsReference() {
			value = ir.Undefined{Of: slot.of}
		}
		fields = append(fields, ir.Field{Name: field.Name, Value: value, Absent: true, Uninitialized: true})
	}
	return fields
}

// missingMergedSlot is the stop for a merged member the class's layout holds no slot for: one whose storage
// is unknown, or one that can point onward (mergedLeaf), which the member declared on the class can hold.
func missingMergedSlot(class *ast.Node, name string, known bool) string {
	if known {
		return "accessing " + name + " from the interface merge of class " + class.Name().Text() +
			", whose slot could hold a reference to another object, and a merged slot can't hold one yet, since a cycle through it would never be freed; declare " +
			name + " on class " + class.Name().Text() + " itself"
	}
	return "accessing " + name + " from the interface merge of class " + class.Name().Text() + " without a known slot representation"
}

// mergedFieldAccess uses the receiver's layout to identify a merge even
// when the checker reports an original declaration outside that merge.
func (l *lowering) mergedFieldAccess(receiver *ast.Node, field *ast.Symbol) classMemberSlot {
	if receiver != nil && field != nil {
		class := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver))
		if isClassInstance(class) {
			if slot := l.classSlot(class, field.Name); slot.merge != nil {
				return slot
			}
		}
	}
	if field == nil || initializedClassMember(field) {
		return classMemberSlot{}
	}
	for _, declaration := range field.Declarations {
		if declaration.Parent != nil && declaration.Parent.Kind == ast.KindInterfaceDeclaration {
			if class := l.mergedClass(declaration.Parent); class != nil {
				return classMemberSlot{holds: true, merge: class}
			}
		}
	}
	return classMemberSlot{}
}

func mergedFieldView(node, class *ast.Node) string {
	return sourceExpression(node) + " (class " + class.Name().Text() + ", interface merge)"
}

// Reuse the checked-view presence, readiness and runtime type contract at each
// actual read. Holding or declaring the merged type needs no runtime promise.
func (l *lowering) mergedFieldRead(node *ast.Node) error {
	var field *ast.Symbol
	var receiver *ast.Node
	if node.Kind == ast.KindPropertyAccessExpression {
		field = l.checker.GetSymbolAtLocation(node.Name())
		receiver = node.AsPropertyAccessExpression().Expression
	} else if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		key := ast.SkipParentheses(access.ArgumentExpression)
		if key.Kind == ast.KindStringLiteral {
			receiver = access.Expression
			field = l.checker.GetPropertyOfType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression)), key.Text())
		}
	}
	slot := l.mergedFieldAccess(receiver, field)
	if slot.merge == nil {
		return nil
	}
	if !slot.holds && !slot.known {
		return l.notYet(node, missingMergedSlot(slot.merge, field.Name, false))
	}
	// A merged member that could point onward has no slot either (mergedLeaf), but a refusal that names its
	// own fix (a callable contract) still answers first, so the stop below is only what's left.
	if field.Flags&ast.SymbolFlagsOptional != 0 {
		if !slot.holds {
			return l.notYet(node, missingMergedSlot(slot.merge, field.Name, true))
		}
		return nil
	}
	if err := l.mergedRequiredContract(node, field); err != nil || !slot.holds {
		if err == nil {
			err = l.notYet(node, missingMergedSlot(slot.merge, field.Name, true))
		}
		return err
	}
	return nil
}

// mergedRequiredContract is a required merged member's read contract: an unbuilt Node host member's own
// NotYet, or the checked view's contract proof, or nil.
func (l *lowering) mergedRequiredContract(node *ast.Node, field *ast.Symbol) error {
	// This module scan precedes rule 800. An unbuilt Node host member needs its
	// own NotYet, not a checked-view contract for storage we do not build. A nil
	// answer keeps the contract proof for implemented host and program members.
	if err := l.nodeLibraryRefusal(node); err != nil {
		return err
	}
	if root, _ := l.inspectorDoomedUse(node); root != nil {
		return nil // Syntax only: actual lowering records the guarded obligation.
	}
	if _, err := l.strictViewContract(node, l.checker.GetTypeOfSymbol(field)); err != nil {
		return err
	}
	// A bare name has no view and no source, so it keeps the record guard's stop.
	l.markCheckedField(field.Name, field.Name, false)
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
	class := l.mergedFieldAccess(access.Expression, field).merge
	if class == nil {
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
	return ir.Property{Object: object, Name: field.Name, Of: of, Optional: l.questionDot(node),
		View: mergedFieldView(node, class), ViewType: l.checker.TypeToString(declared), ViewWhere: l.program.Where(node),
		ViewTypeID: int(declared.Id()), ViewContract: l.result.ViewContractTypes[int(declared.Id())],
		ViewReceiverTypeID: int(l.checker.GetTypeAtLocation(access.Expression).Id())}, true, nil
}
