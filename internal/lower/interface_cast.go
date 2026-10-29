package lower

import (
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A tag chooses the interface; its fields remain checked at every read. In .a the tag must tell the
// target from the source: a test every source value passes proves nothing, so .a takes the first test
// a source value can fail and refuses the cast when there is none. .ts keeps the first tag, checked.
func (l *lowering) interfaceCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	if source.Flags()&checker.TypeFlagsObject == 0 || target.Flags()&checker.TypeFlagsObject == 0 || value.Type() != ir.Object || isClassInstance(target) {
		return nil, nil
	}
	var vacuous *Refused
	for _, property := range l.checker.GetPropertiesOfType(source) {
		field := property.Name
		literal := l.fieldLiteral(target, field)
		if literal == nil || !l.checker.IsTypeAssignableTo(target, source) || l.widened(target, source, map[[2]*checker.Type]bool{}) != nil {
			continue
		}
		allowed, fieldType, constant := l.literalConstant(literal)
		if !constant {
			continue
		}
		if !l.checkedAssertionSource(node) && l.withinLiterals(l.checker.GetTypeOfSymbol(property), []*checker.Type{literal}) {
			if vacuous == nil {
				vacuous = &Refused{Where: l.program.Where(node), What: "a downcast to " + l.checker.TypeToString(target) + " whose tag test is vacuous: every " + l.checker.TypeToString(source) + " has " + field + " " + l.checker.TypeToString(literal), Fix: "give " + l.checker.TypeToString(target) + " a tag of its own in a discriminated union, or build the value as a " + l.checker.TypeToString(target) + " so the compiler can prove it"}
			}
			continue
		}
		if _, err := l.view(node, value, target); err != nil {
			return nil, err
		}
		return ir.CheckedCast{Value: value, Field: field, FieldType: fieldType, Allowed: []ir.Expression{allowed}, CheckedFields: true, Message: "cast failed: this " + l.checker.TypeToString(source) + " is not a " + l.checker.TypeToString(target)}, nil
	}
	if vacuous != nil {
		return nil, vacuous
	}
	return nil, nil
}

// withinLiterals is whether every value declared's type admits is one of the allowed literals,
// compared by runtime value as a cast's tag test compares, so the test passes for every one of them: a
// union whose members all are, a literal equal to one (two enums' members sharing a value are equal
// here), an intersection with a member that is, or a type parameter whose constraint is. Any other
// type (string, undefined, an object, unknown, a template) admits a value the test can fail.
func (l *lowering) withinLiterals(declared *checker.Type, allowed []*checker.Type) bool {
	switch {
	case declared.Flags()&checker.TypeFlagsUnion != 0:
		for _, member := range declared.Types() {
			if !l.withinLiterals(member, allowed) {
				return false
			}
		}
		return true
	case declared.Flags()&checker.TypeFlagsIntersection != 0:
		for _, member := range declared.Types() {
			if l.withinLiterals(member, allowed) {
				return true
			}
		}
		return false
	case declared.Flags()&checker.TypeFlagsTypeParameter != 0:
		constraint := l.checker.GetBaseConstraintOfType(declared)
		return constraint != nil && constraint != declared && l.withinLiterals(constraint, allowed)
	}
	mask := checker.TypeFlagsStringLiteral | checker.TypeFlagsNumberLiteral | checker.TypeFlagsBooleanLiteral
	if declared.Flags()&mask == 0 {
		return false
	}
	for _, literal := range allowed {
		if literal.Flags()&mask == declared.Flags()&mask && reflect.DeepEqual(literal.AsLiteralType().Value(), declared.AsLiteralType().Value()) {
			return true
		}
	}
	return false
}

// view registers checks by receiver type and member, including nested contracts.
// An unrelated receiver with the same member name does not become a view.
func (l *lowering) view(node *ast.Node, value ir.Expression, target *checker.Type) (ir.Expression, error) {
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		return l.legacyView(node, value, target)
	}
	fields, err := l.viewSchema(node, target)
	if err != nil {
		return nil, err
	}
	if err := l.refuseViewElementReads(fields); err != nil {
		return nil, err
	}
	if l.result.CheckedFields == nil {
		l.result.CheckedFields = map[string]bool{}
	}

	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return nil, err
	}
	var refused error
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		if refused != nil {
			return true
		}
		// An accessor slot still needs a conversion the checked read boundary can't emit. An
		// optional slot or receiver is emission's to judge, at the read (unemittableCheckedViewRead).
		if part.Kind == ast.KindPropertyAccessExpression && l.checkedViewReceiverField(l.checker.GetTypeAtLocation(part.AsPropertyAccessExpression().Expression), part.Name().Text()) {
			access := part.AsPropertyAccessExpression()
			if base, _ := l.representation(l.checker.GetTypeAtLocation(access.Expression)); base == ir.Object {
				field := l.checker.GetSymbolAtLocation(part.Name())
				if field != nil && len(l.checker.GetSignaturesOfType(l.checker.GetTypeOfSymbol(field), checker.SignatureKindCall)) == 0 && accessorSymbol(field) {
					refused = l.notYet(part, checkedFieldAliasAccessor)
				}
			}
		}
		if part.Kind == ast.KindPropertyAccessExpression && l.checkedViewReceiverField(l.checker.GetTypeAtLocation(part.AsPropertyAccessExpression().Expression), part.Name().Text()) && ast.IsAssignmentTarget(part) {
			if symbol := l.checker.GetSymbolAtLocation(part.Name()); symbol != nil {
				of, _ := l.representation(l.checker.GetTypeOfSymbol(symbol))
				if of == ir.Object {
					refused = l.notYet(part, "writing a checked object field without its source-slot type certificate")
				}
			}
		}
		if part.Kind == ast.KindGetAccessor && part.Name() != nil && fields[part.Name().Text()] {
			refused = l.notYet(part, "a getter in a checked field contract")
		}
		if refused == nil {
			part.ForEachChild(visit)
		}
		return refused != nil
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	if refused != nil {
		return nil, refused
	}
	l.result.ViewOrigins = append(l.result.ViewOrigins, value)
	l.graphViewSite = node
	return value, nil
}

const checkedFieldAliasAccessor = "a checked field alias of an accessor"

const checkedFieldAliasNotScalar = "a checked field alias whose field isn't a scalar"

// checkedFieldAliasGap names the first conversion a checked field alias lacks,
// in one order at every site: an optional slot, then an accessor, then a field
// whose representation isn't a plain scalar. Empty means the alias needs none.
func checkedFieldAliasGap(access *ast.PropertyAccessExpression, field *ast.Symbol, scalar bool) string {
	switch {
	case field.Flags&ast.SymbolFlagsOptional != 0 || access.QuestionDotToken != nil:
		return "a checked field alias of an optional field or through ?."
	case accessorSymbol(field):
		return checkedFieldAliasAccessor
	case !scalar:
		return checkedFieldAliasNotScalar
	}
	return ""
}

func interfaceScalar(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if !interfaceScalar(member) {
				return false
			}
		}
		return true
	}
	return proven.Flags()&(checker.TypeFlagsString|checker.TypeFlagsStringLiteral|checker.TypeFlagsNumber|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBoolean|checker.TypeFlagsBooleanLiteral) != 0 && proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsIntersection|checker.TypeFlagsTypeParameter) == 0
}

// A finite literal contract is checked as well as its primitive representation.
func (l *lowering) viewLiterals(declared *checker.Type) []ir.Expression {
	if declared.Flags()&checker.TypeFlagsUnion != 0 {
		var allowed []ir.Expression
		for _, member := range declared.Types() {
			values := l.viewLiterals(member)
			if len(values) == 0 {
				return nil
			}
			allowed = append(allowed, values...)
		}
		return allowed
	}
	value, _, known := l.literalConstant(declared)
	if known {
		return []ir.Expression{value}
	}
	return nil
}

func (l *lowering) callableViewContract(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if l.callableViewContract(member) {
				return true
			}
		}
	}
	return len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) != 0
}

func (l *lowering) viewSchema(node *ast.Node, target *checker.Type) (map[string]bool, error) {
	l.graphViewSite = node
	id, err := l.viewContract(node, target)
	if err != nil {
		return nil, err
	}
	fields := map[string]bool{}
	// A failed whole-object adapter can leave an unsupported descriptor without
	// its fields. Keep those names registered so demanded reads reach the lazy
	// refusal instead of silently reading unchecked storage.
	if contract := l.result.ViewContracts[id-1]; contract.Unsupported != "" {
		if l.result.CheckedFields == nil {
			l.result.CheckedFields = map[string]bool{}
		}
		for _, property := range l.checker.GetPropertiesOfType(target) {
			fields[property.Name] = true
			// Preserve the boundary under receiver-specific registration too.
			l.result.CheckedFields[ir.CheckedFieldKey(int(target.Id()), property.Name)] = true
		}
	}
	seen := map[ir.ViewContractID]bool{}
	var visit func(ir.ViewContractID)
	visit = func(id ir.ViewContractID) {
		if id == 0 || seen[id] {
			return
		}
		seen[id] = true
		contract := l.result.ViewContracts[id-1]
		for _, field := range contract.Fields {
			fields[field.Name] = true
			for typeID, contractID := range l.result.ViewContractTypes {
				if contractID == id {
					if l.result.CheckedFields == nil {
						l.result.CheckedFields = map[string]bool{}
					}
					l.result.CheckedFields[ir.CheckedFieldKey(typeID, field.Name)] = true
				}
			}
			visit(field.Contract)
		}
		for _, member := range contract.Members {
			visit(member)
		}
	}
	visit(id)
	return fields, nil
}

func (l *lowering) legacyView(node *ast.Node, value ir.Expression, target *checker.Type) (ir.Expression, error) {
	if l.callableViewContract(target) {
		return nil, &Refused{Where: l.program.Where(node), What: "a checked view with a callable contract", Fix: "prove the callable body rather than asserting its signature"}
	}
	// Diagnose unreifiable contracts before temporary backend limitations.
	for _, property := range l.checker.GetPropertiesOfType(target) {
		declared := l.checker.GetTypeOfSymbol(property)
		if l.callableViewContract(declared) {
			return nil, &Refused{Where: l.program.Where(node), What: "a checked view with callable field " + property.Name, Fix: "prove the callable body rather than asserting its signature"}
		}
	}
	fields := map[string]bool{}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		declared := l.checker.GetTypeOfSymbol(property)
		if l.callableViewContract(declared) {
			return nil, &Refused{Where: l.program.Where(node), What: "a checked view with callable field " + property.Name, Fix: "prove the callable body rather than asserting its signature"}
		}
		of, known := l.representation(declared)
		if property.Flags&ast.SymbolFlagsOptional != 0 || !interfaceScalar(declared) || !known || of < ir.Number || of > ir.String {
			return nil, l.notYet(node, "checked view field "+property.Name+" of type "+l.checker.TypeToString(declared))
		}
		fields[property.Name] = true
	}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return nil, err
	}
	var found error
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		if found != nil {
			return true
		}
		if part.Kind == ast.KindGetAccessor && part.Name() != nil && fields[part.Name().Text()] {
			found = l.notYet(part, "a getter in a checked field contract")
		}
		if part.Kind == ast.KindObjectBindingPattern {
			for _, binding := range part.AsBindingPattern().Elements.Nodes {
				name := binding.Name()
				if binding.AsBindingElement().PropertyName != nil {
					name = binding.AsBindingElement().PropertyName
				}
				if name != nil && fields[name.Text()] {
					declared := l.checker.GetTypeAtLocation(binding.Name())
					of, known := l.representation(declared)
					if !known || of < ir.Number || of > ir.String || !interfaceScalar(declared) {
						found = l.notYet(binding, "a checked destructured alias requiring a representation conversion")
					}
				}
			}
		}
		if part.Kind == ast.KindPropertyAccessExpression && l.checkedViewReceiverField(l.checker.GetTypeAtLocation(part.AsPropertyAccessExpression().Expression), part.Name().Text()) {
			access := part.AsPropertyAccessExpression()
			if base, _ := l.representation(l.checker.GetTypeAtLocation(access.Expression)); base == ir.Object {
				field := l.checker.GetSymbolAtLocation(part.Name())
				if field != nil && len(l.checker.GetSignaturesOfType(l.checker.GetTypeOfSymbol(field), checker.SignatureKindCall)) == 0 {
					of, known := l.representation(l.checker.GetTypeOfSymbol(field))
					if gap := checkedFieldAliasGap(access, field, interfaceScalar(l.checker.GetTypeOfSymbol(field)) && known && of >= ir.Number && of <= ir.String); gap != "" {
						if gap == checkedFieldAliasNotScalar {
							gap += " (" + l.checker.TypeToString(l.checker.GetTypeOfSymbol(field)) + ")"
						}
						found = l.notYet(part, gap)
					}
				}
			}
		}
		if found == nil {
			part.ForEachChild(visit)
		}
		return found != nil
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	if found != nil {
		return nil, found
	}
	if err := l.refuseViewElementReads(fields); err != nil {
		return nil, err
	}
	if l.result.CheckedFields == nil {
		l.result.CheckedFields = map[string]bool{}
	}
	for field := range fields {
		l.result.CheckedFields[ir.CheckedFieldKey(int(target.Id()), field)] = true
	}
	l.graphViewSite = node
	return value, nil
}

// Nullable receivers read the present object's contract. A generic receiver
// uses its instantiated type, or its constraint when the binder remains rigid.
// Neither operation relates unrelated objects merely because their names match.
func (l *lowering) viewReceiverTypeID(receiver *checker.Type) int {
	receiver = l.concrete(receiver)
	if receiver.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if constraint := l.checker.GetBaseConstraintOfType(receiver); constraint != nil {
			receiver = l.concrete(constraint)
		}
	}
	return int(l.checker.GetNonNullableType(receiver).Id())
}

// A union receiver can carry a checked member's allocation. Preserve that
// boundary for its common field without marking unrelated receiver types.
func (l *lowering) checkedViewReceiverField(receiver *checker.Type, name string) bool {
	id := l.viewReceiverTypeID(receiver)
	if l.result.CheckedField(id, name) {
		return true
	}
	receiver = l.checker.GetNonNullableType(l.concrete(receiver))
	if receiver.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range receiver.Types() {
			if l.checkedViewReceiverField(member, name) {
				l.result.CheckedFields[ir.CheckedFieldKey(id, name)] = true
				return true
			}
		}
	}
	return false
}
