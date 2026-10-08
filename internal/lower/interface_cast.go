package lower

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A tag chooses the interface; its fields remain checked at every read.
func (l *lowering) interfaceCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	if source.Flags()&checker.TypeFlagsObject == 0 || target.Flags()&checker.TypeFlagsObject == 0 || value.Type() != ir.Object || isClassInstance(target) {
		return nil, nil
	}
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
		if _, err := l.view(node, value, target); err != nil {
			return nil, err
		}
		return l.certifiedCheckedCast(node, ir.CheckedCast{Value: value, Field: field, FieldType: fieldType, Allowed: []ir.Expression{allowed}, CheckedFields: true, Message: "cast failed: this " + l.checker.TypeToString(source) + " is not a " + l.checker.TypeToString(target)}, target)
	}
	return nil, nil
}

// view is the shared entry point. Checking by field name throughout the program is
// conservative: aliases and function boundaries cannot lose a checked read.
// More precise view propagation and erasure can reduce that set without trusting casts.
func (l *lowering) view(node *ast.Node, value ir.Expression, target *checker.Type) (ir.Expression, error) {
	fields, err := l.viewSchema(node, target)
	if err != nil {
		return nil, err
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
		if part.Kind == ast.KindObjectBindingPattern {
			for _, binding := range part.AsBindingPattern().Elements.Nodes {
				name := binding.Name()
				if binding.AsBindingElement().PropertyName != nil {
					name = binding.AsBindingElement().PropertyName
				}
				if name != nil && fields[name.Text()] {
					declared := l.checker.GetTypeAtLocation(binding.Name())
					of, known := l.representation(declared)
					if !known || (of < ir.Number || of > ir.Array) && !l.viewPrimitiveUnionRead(declared) || !l.viewDataType(declared) {
						found = l.lazyReadRefusal(binding, name.Text(), "destructuring representation conversion")
					}
				}
			}
		}
		if part.Kind == ast.KindPropertyAccessExpression && fields[part.Name().Text()] {
			access := part.AsPropertyAccessExpression()
			if base, _ := l.representation(l.checker.GetTypeAtLocation(access.Expression)); base == ir.Object {
				field := l.checker.GetSymbolAtLocation(part.Name())
				if field != nil && !l.callableViewContract(l.checker.GetTypeOfSymbol(field)) {
					of, known := l.representation(l.checker.GetTypeOfSymbol(field))
					if of == ir.Object && ast.IsAssignmentTarget(part) && !l.result.OptionalViewFields[l.fieldName(part.Name())] {
						found = l.notYet(part, "writing a checked object field without its source-slot type certificate")
					}

					if (!l.viewDataType(l.checker.GetTypeOfSymbol(field)) || !known || (of < ir.Number || of > ir.Array) && of != ir.MaybeNumber && of != ir.MaybeBoolean && !(of == ir.Union && (l.includesNull(l.checker.GetTypeOfSymbol(field)) || l.includesUndefined(l.checker.GetTypeOfSymbol(field)))) && !(of == ir.Closure && l.callableViewContract(l.checker.GetTypeOfSymbol(field))) && !(of == ir.Map && l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeOfSymbol(field)), "Map", "ReadonlyMap"))) && !l.objectPrimitiveViewType(l.checker.GetTypeOfSymbol(field)) && !l.viewPrimitiveUnionRead(l.checker.GetTypeOfSymbol(field)) || accessorSymbol(field) {
						family := l.unsupportedViewFamily(l.checker.GetNonNullableType(l.concrete(l.checker.GetTypeOfSymbol(field))))
						if family == "" {
							family = "representation conversion"
						}
						if accessorSymbol(field) {
							family = "accessor"
						}
						if of == ir.Union {
							family = "mixed representation union"
						}
						found = l.lazyReadRefusal(part, part.Name().Text(), family)
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
	if l.result.CheckedFields == nil {
		l.result.CheckedFields = map[string]bool{}
	}
	for field := range fields {
		l.result.CheckedFields[field] = true
	}
	l.result.ViewOrigins = append(l.result.ViewOrigins, value)
	return value, nil
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
	return proven.Flags()&checker.TypeFlagsUndefined != 0 || proven.Flags()&(checker.TypeFlagsString|checker.TypeFlagsStringLiteral|checker.TypeFlagsNumber|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBoolean|checker.TypeFlagsBooleanLiteral) != 0 && proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsIntersection|checker.TypeFlagsTypeParameter) == 0
}

func (l *lowering) interfaceScalarShape(proven *checker.Type) bool {
	if isClassInstance(proven) || len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) != 0 {
		return false
	}
	for _, property := range l.checker.GetPropertiesOfType(proven) {
		if property.Flags&ast.SymbolFlagsOptional != 0 || !l.checker.IsReadonlySymbol(property) || !interfaceScalar(l.checker.GetTypeOfSymbol(property)) {
			return false
		}
	}
	return len(l.checker.GetPropertiesOfType(proven)) > 0
}

// Look at initializer types, never an asserted/contextual interface as evidence of payload.
func interfaceLiteralFields(node *ast.Node) (map[string]*ast.Node, bool) {
	fields := map[string]*ast.Node{}
	for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind != ast.KindPropertyAssignment && property.Kind != ast.KindShorthandPropertyAssignment {
			return nil, false
		}
		name := property.Name()
		if name == nil || (name.Kind != ast.KindIdentifier && name.Kind != ast.KindStringLiteral) {
			return nil, false
		}
		value := name
		if property.Kind == ast.KindPropertyAssignment {
			value = property.AsPropertyAssignment().Initializer
		}
		fields[name.Text()] = value
	}
	return fields, true
}

func (l *lowering) interfaceConstructions(cast *ast.Node, target *checker.Type, field string, literal *checker.Type) error {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	checked := map[string]*checker.Type{}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		checked[property.Name] = l.checker.GetTypeOfSymbol(property)
	}
	var found error
	fail := func(node *ast.Node, reason string) {
		found = &Refused{Where: l.program.Where(cast), What: "an unproven base interface construction", Fix: fmt.Sprintf("%s at %s; build the complete target in an immutable literal and preserve its fields (adamic/interface-construction)", reason, l.program.Where(node))}
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil {
			return true
		}
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsAmbient) {
			fail(node, "an ambient value has no construction body")
			return true
		}
		switch node.Kind {
		case ast.KindAnyKeyword, ast.KindTypeAssertionExpression:
			fail(node, "an unproven type could forge a construction")
		case ast.KindIdentifier:
			if node.Text() == "Object" || node.Text() == "Reflect" || node.Text() == "JSON" {
				fail(node, "reflection or host construction is outside the prototype")
			}
		case ast.KindClassDeclaration, ast.KindNewExpression, ast.KindSpreadAssignment, ast.KindGetAccessor, ast.KindSetAccessor:
			fail(node, "opaque, staged or spread construction is outside the prototype")
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration:
			if len(node.TypeParameters()) > 0 || node.Body() == nil {
				fail(node, "a generic or bodyless factory is outside the prototype")
			}
		case ast.KindImportDeclaration, ast.KindExportDeclaration:
			if specifier := node.ModuleSpecifier(); specifier != nil && !strings.HasPrefix(specifier.Text(), ".") {
				fail(node, "an external module has no closed construction proof")
			}
		case ast.KindCallExpression:
			callee := ast.SkipParentheses(node.AsCallExpression().Expression)
			if callee.Kind == ast.KindPropertyAccessExpression {
				base := callee.AsPropertyAccessExpression().Expression
				if ast.IsIdentifier(base) && (base.Text() == "Object" || base.Text() == "Reflect" || base.Text() == "JSON") {
					fail(node, "reflection or host construction is outside the prototype")
				}
			}
		case ast.KindBinaryExpression:
			binary := node.AsBinaryExpression()
			if ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
				l.interfaceWrite(binary.Left, checked, fail)
			}
		case ast.KindPrefixUnaryExpression:
			unary := node.AsPrefixUnaryExpression()
			if unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken {
				l.interfaceWrite(unary.Operand, checked, fail)
			}
		case ast.KindPostfixUnaryExpression:
			l.interfaceWrite(node.AsPostfixUnaryExpression().Operand, checked, fail)
		case ast.KindAsExpression:
			as := node.AsAsExpression()
			if as.Type.Kind == ast.KindTypeReference && as.Type.AsTypeReferenceNode().TypeName.Text() == "const" {
				break
			}
			from, to := l.checker.GetTypeAtLocation(as.Expression), l.checker.GetTypeAtLocation(node)
			// Another base cast is checked by this same pass at its own lowering site.
			baseCast := l.interfaceScalarShape(from) && l.interfaceScalarShape(to) && l.fieldLiteral(to, field) != nil && l.checker.IsTypeAssignableTo(to, from)
			scalarUpcast := interfaceScalar(from) && interfaceScalar(to) && l.checker.IsTypeAssignableTo(from, to)
			objectUpcast := l.interfaceScalarShape(to) && l.checker.IsTypeAssignableTo(from, to) && l.widened(from, to, map[[2]*checker.Type]bool{}) == nil
			if !baseCast && !scalarUpcast && !objectUpcast {
				fail(node, "an assertion could forge a construction or scalar payload")
			}
		case ast.KindObjectLiteralExpression:
			fields, complete := interfaceLiteralFields(node)
			if !complete {
				fail(node, "a literal has unproven field names or behavior")
				break
			}
			kind := fields[field]
			if kind == nil {
				break // No write or spread can add a tag later in an admitted program.
			}
			kindType := l.checker.GetTypeAtLocation(kind)
			if !interfaceScalar(kindType) {
				fail(kind, "a tag has an unproven scalar type")
				break
			}
			if !l.checker.IsTypeAssignableTo(literal, kindType) {
				break // A precise different tag cannot select this target.
			}
			for name, wanted := range checked {
				value := fields[name]
				if value == nil {
					fail(node, "matching kind lacks required field "+name)
					break
				}
				actual := l.checker.GetTypeAtLocation(value)
				// A broad tag needs the runtime comparison, so it need not be a target literal.
				if !interfaceScalar(actual) || (name != field && !l.checker.IsTypeAssignableTo(actual, wanted)) {
					fail(value, "matching kind has incompatible field "+name)
					break
				}
			}
		}
		if found == nil {
			node.ForEachChild(visit)
		}
		return found != nil
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
		if found != nil {
			return found
		}
	}
	return nil
}

// Field names are deliberately global, including writes through wider or unrelated views.
// Destructuring and computed destinations fail closed rather than guess their alias effects.
func (l *lowering) interfaceWrite(node *ast.Node, checked map[string]*checker.Type, fail func(*ast.Node, string)) {
	node = ast.SkipParentheses(node)
	if ast.IsIdentifier(node) {
		return
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		if _, needed := checked[node.AsPropertyAccessExpression().Name().Text()]; !needed {
			return
		}
	}
	fail(node, "a write may invalidate the kind-to-shape invariant")
}

// A finite literal contract is checked as well as its primitive representation.
func (l *lowering) viewLiterals(declared *checker.Type) []ir.Expression {
	if base := l.phantomBase(declared); base != nil {
		return l.viewLiterals(base)
	}
	// A whole numeric enum admits numbers outside its declared members.
	if l.openNumericEnumType(declared) {
		return nil
	}
	if declared.Flags()&checker.TypeFlagsUnion != 0 {
		var allowed []ir.Expression
		for _, member := range declared.Types() {
			if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
				continue
			}
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

// viewSchema is the exact contract-admission portion of the shared entry point.
// Inventories can audit it without claiming program lowering or alias admission.
func (l *lowering) viewSchema(node *ast.Node, target *checker.Type) (map[string]bool, error) {
	id, err := l.viewContract(node, target)
	if err != nil {
		return nil, err
	}
	fields := map[string]bool{}
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
			if field.Optional {
				l.optionalViewWriteField(field.Name)
			}
			visit(field.Contract)
		}
		for _, member := range contract.Members {
			visit(member)
		}
		visit(contract.Key)
		visit(contract.Element)
	}
	visit(id)
	if ir.HasArrayViews(l.result) {
		if err := l.viewArrayUnsupportedUses(node); err != nil {
			return nil, err
		}
	}
	return fields, nil
}
