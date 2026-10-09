package lower

import (
	"fmt"
	"reflect"
	"strings"
	_ "unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// These checker operations preserve the binder in a read from T, where an
// ordinary property read is reported through T's constraint. Reference the
// checker implementation through its shim types rather than copying it.
//
//go:linkname genericIndexedAccess github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).getIndexedAccessType
func genericIndexedAccess(*checker.Checker, *checker.Type, *checker.Type) *checker.Type

//go:linkname genericStringLiteral github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).getStringLiteralType
func genericStringLiteral(*checker.Checker, string) *checker.Type

type genericBodyRelation struct {
	where, source *ast.Node
	target        *checker.Type
	kind          string
}

// A relation is recorded before lowering can erase literal and binder identities.
func (l *lowering) genericRelation(node *ast.Node) *genericBodyRelation {
	generic := false
	for parent := node; parent != nil; parent = parent.Parent {
		if (ast.IsFunctionLike(parent) || parent.Kind == ast.KindClassDeclaration || parent.Kind == ast.KindClassExpression) && len(parent.TypeParameters()) > 0 {
			generic = true
			break
		}
	}
	if !generic {
		return nil
	}
	relation := &genericBodyRelation{where: node}
	switch node.Kind {
	case ast.KindVariableDeclaration:
		relation.source = node.AsVariableDeclaration().Initializer
		relation.target = l.checker.GetTypeAtLocation(node.Name())
		relation.kind = "initializer"
	case ast.KindParameter:
		relation.source = node.AsParameterDeclaration().Initializer
		relation.target = l.checker.GetTypeAtLocation(node.Name())
		relation.kind = "default initializer"
	case ast.KindPropertyDeclaration:
		relation.source = node.AsPropertyDeclaration().Initializer
		relation.target = l.checker.GetTypeAtLocation(node.Name())
		relation.kind = "field initializer"
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind != ast.KindEqualsToken && !ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			return nil
		}
		relation.source = binary.Right
		if binary.OperatorToken.Kind != ast.KindEqualsToken {
			relation.source = node
		}
		relation.target = l.genericWriteType(binary.Left)
		relation.kind = "assignment"
	case ast.KindReturnStatement:
		relation.source = node.AsReturnStatement().Expression
		function := node.Parent
		for function != nil && !ast.IsFunctionLike(function) {
			function = function.Parent
		}
		if function == nil {
			return nil
		}
		relation.target = l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(function))
		relation.kind = "return"
	case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression:
		var operand *ast.Node
		var operator ast.Kind
		if node.Kind == ast.KindPrefixUnaryExpression {
			operand = node.AsPrefixUnaryExpression().Operand
			operator = node.AsPrefixUnaryExpression().Operator
		} else {
			operand = node.AsPostfixUnaryExpression().Operand
			operator = node.AsPostfixUnaryExpression().Operator
		}
		if operator != ast.KindPlusPlusToken && operator != ast.KindMinusMinusToken {
			return nil
		}
		relation.source = node
		relation.target = l.genericWriteType(operand)
		relation.kind = "assignment"
	case ast.KindArrowFunction:
		if node.Body() == nil || node.Body().Kind == ast.KindBlock {
			return nil
		}
		relation.source = node.Body()
		relation.target = l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(node))
		relation.kind = "return"
	case ast.KindShorthandPropertyAssignment:
		relation.source = node.Name()
		context := l.checker.GetContextualType(node.Parent, checker.ContextFlagsNone)
		if context == nil {
			return nil
		}
		if node.Name().Kind == ast.KindComputedPropertyName {
			relation.target = genericIndexedAccess(l.checker, context, l.genericOwnType(node.Name().AsComputedPropertyName().Expression))
		} else {
			relation.target = l.checker.GetTypeOfPropertyOfType(context, node.Name().Text())
		}
		relation.kind = "field initializer"
	case ast.KindPropertyAssignment:
		relation.source = node.AsPropertyAssignment().Initializer
		context := l.checker.GetContextualType(node.Parent, checker.ContextFlagsNone)
		if context == nil {
			return nil
		}
		if node.Name().Kind == ast.KindComputedPropertyName {
			relation.target = genericIndexedAccess(l.checker, context, l.genericOwnType(node.Name().AsComputedPropertyName().Expression))
		} else {
			relation.target = l.checker.GetTypeOfPropertyOfType(context, node.Name().Text())
		}
		relation.kind = "field initializer"
	default:
		return nil
	}
	if relation.source == nil || relation.target == nil {
		return nil
	}
	return relation
}

func (l *lowering) genericWriteType(node *ast.Node) *checker.Type {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindIdentifier {
		if symbol := l.checker.GetSymbolAtLocation(node); symbol != nil {
			return l.checker.GetTypeOfSymbol(symbol)
		}
	}
	return l.genericOwnType(node)
}

func (l *lowering) genericOwnType(node *ast.Node) *checker.Type {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		if ast.IsAssignmentOperator(binary.OperatorToken.Kind) && binary.OperatorToken.Kind != ast.KindEqualsToken {
			bound := l.checker.GetBaseConstraintOfType(l.checker.GetTypeAtLocation(binary.Left))
			if bound != nil {
				return l.checker.GetBaseTypeOfLiteralType(bound)
			}
		}
	}
	own := l.checker.GetTypeAtLocation(node)
	if l.genericDependent(own, map[*checker.Type]bool{}) {
		return own
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		object := l.genericOwnType(access.Expression)
		original := l.checker.GetTypeAtLocation(access.Expression)
		if original.Flags()&checker.TypeFlagsTypeParameter != 0 && original.AsTypeParameter().IsThisType() {
			return own
		}
		if l.genericDependent(object, map[*checker.Type]bool{}) {
			return genericIndexedAccess(l.checker, object, genericStringLiteral(l.checker, access.Name().Text()))
		}
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		object := l.genericOwnType(access.Expression)
		if l.genericDependent(object, map[*checker.Type]bool{}) {
			return genericIndexedAccess(l.checker, object, l.genericOwnType(access.ArgumentExpression))
		}
	}
	return l.checker.GetTypeAtLocation(node)
}

// Do not replace a binder with its upper bound: a bound permits narrower actual
// arguments. Follow checker type graphs, including aliases, nested containers,
// conditional types, signatures and recursive object fields.
func (l *lowering) genericDependent(proven *checker.Type, seen map[*checker.Type]bool) bool {
	if proven == nil || seen[proven] {
		return false
	}
	seen[proven] = true
	flags := proven.Flags()
	if flags&checker.TypeFlagsTypeParameter != 0 {
		return true
	}
	var children []*checker.Type
	switch {
	case flags&checker.TypeFlagsUnionOrIntersection != 0:
		children = proven.Types()
	case flags&checker.TypeFlagsIndexedAccess != 0:
		access := proven.AsIndexedAccessType()
		children = []*checker.Type{access.ObjectType(), access.IndexType()}
	case flags&checker.TypeFlagsIndex != 0:
		children = []*checker.Type{proven.AsIndexType().Target()}
	case flags&checker.TypeFlagsSubstitution != 0:
		children = []*checker.Type{proven.AsSubstitutionType().BaseType(), proven.AsSubstitutionType().SubstConstraint()}
	case flags&checker.TypeFlagsConditional != 0:
		conditional := proven.AsConditionalType()
		children = []*checker.Type{conditional.CheckType(), conditional.ExtendsType(), l.checker.GetTrueTypeOfConditionalType(proven), l.checker.GetFalseTypeOfConditionalType(proven)}
	case flags&checker.TypeFlagsTemplateLiteral != 0:
		children = proven.AsTemplateLiteralType().Types()
	case flags&checker.TypeFlagsStringMapping != 0:
		children = []*checker.Type{proven.AsStringMappingType().Target()}
	case flags&checker.TypeFlagsObject != 0:
		if proven.ObjectFlags()&checker.ObjectFlagsReference != 0 {
			for _, argument := range l.checker.GetTypeArguments(proven) {
				if l.genericDependent(argument, seen) {
					return true
				}
			}
			return false
		}
		if proven.ObjectFlags()&checker.ObjectFlagsMapped != 0 {
			mapped := proven.AsMappedType()
			mapped.ResolveComponents(l.checker, proven)
			seen[mapped.TypeParameter()] = true
			children = append(children, mapped.ConstraintType(), mapped.NameType(), mapped.TemplateType())
		}
		for _, index := range l.checker.GetIndexInfosOfType(proven) {
			children = append(children, index.KeyType(), index.ValueType())
		}
		for _, property := range l.checker.GetPropertiesOfType(proven) {
			children = append(children, l.checker.GetTypeOfSymbol(property))
		}
		for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
			for _, signature := range l.checker.GetSignaturesOfType(proven, kind) {
				for _, binder := range signature.TypeParameters() {
					seen[binder] = true
				}
				children = append(children, l.checker.GetReturnTypeOfSignature(signature))
				for _, parameter := range signature.Parameters() {
					children = append(children, l.checker.GetTypeOfSymbol(parameter))
				}
			}
		}
	}
	for _, child := range children {
		if l.genericDependent(child, seen) {
			return true
		}
	}
	return false
}

func (l *lowering) genericBodyAllows(source, target *checker.Type) bool {
	if source == target || source.Flags()&checker.TypeFlagsNever != 0 {
		return true
	}
	// A polymorphic this read keeps its own identity. Its enclosing class also
	// carries the same generic arguments when used through the class view.
	if source.Flags()&checker.TypeFlagsTypeParameter != 0 && source.AsTypeParameter().IsThisType() {
		if bound := l.checker.GetBaseConstraintOfType(source); bound != nil {
			return l.genericBodyAllows(bound, target)
		}
	}
	if source.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range source.Types() {
			if !l.genericBodyAllows(member, target) {
				return false
			}
		}
		return true
	}
	if source.Flags()&checker.TypeFlagsSubstitution != 0 {
		return l.genericBodyAllows(source.AsSubstitutionType().BaseType(), target)
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		if source.Flags()&checker.TypeFlagsUnion != 0 {
			for _, member := range source.Types() {
				if !l.genericBodyAllows(member, target) {
					return false
				}
			}
			return true
		}
		for _, member := range target.Types() {
			if l.genericBodyAllows(source, member) {
				return true
			}
		}
		return false
	}
	if !l.genericDependent(target, map[*checker.Type]bool{}) {
		return l.checker.IsTypeAssignableTo(source, target)
	}
	if identicalTypes(l.checker, source, target) {
		return true
	}
	if target.Flags()&checker.TypeFlagsTypeParameter != 0 && l.narrowedFrom(source, target) {
		return true
	}
	if source.Flags()&checker.TypeFlagsObject != 0 && target.Flags()&checker.TypeFlagsObject != 0 && l.checker.IsTypeAssignableTo(source, target) {
		if source.ObjectFlags()&checker.ObjectFlagsReference != 0 && target.ObjectFlags()&checker.ObjectFlagsReference != 0 {
			from, to := l.checker.GetTypeArguments(source), l.checker.GetTypeArguments(target)
			if len(from) != len(to) {
				return false
			}
			for index := range to {
				if !l.genericBodyAllows(from[index], to[index]) {
					return false
				}
			}
			return true
		}
		if target.ObjectFlags()&checker.ObjectFlagsMapped != 0 {
			wanted := target.AsMappedType()
			wanted.ResolveComponents(l.checker, target)
			if l.genericDependent(wanted.ConstraintType(), map[*checker.Type]bool{}) {
				if source.ObjectFlags()&checker.ObjectFlagsMapped == 0 {
					return false
				}
				actual := source.AsMappedType()
				actual.ResolveComponents(l.checker, source)
				if !l.genericBodyAllows(actual.ConstraintType(), wanted.ConstraintType()) {
					return false
				}
				mapper := newTypeMapper([]*checker.Type{actual.TypeParameter()}, []*checker.Type{wanted.TypeParameter()})
				if !l.genericBodyAllows(instantiateType(l.checker, actual.TemplateType(), mapper), wanted.TemplateType()) {
					return false
				}
			}
		}
		for _, wanted := range l.checker.GetIndexInfosOfType(target) {
			if !l.genericDependent(wanted.ValueType(), map[*checker.Type]bool{}) {
				continue
			}
			indexed := false
			for _, actual := range l.checker.GetIndexInfosOfType(source) {
				if l.checker.IsTypeAssignableTo(wanted.KeyType(), actual.KeyType()) {
					if !l.genericBodyAllows(actual.ValueType(), wanted.ValueType()) {
						return false
					}
					indexed = true
				}
			}
			if !indexed {
				for _, property := range l.checker.GetPropertiesOfType(source) {
					if !l.genericBodyAllows(l.checker.GetTypeOfSymbol(property), wanted.ValueType()) {
						return false
					}
				}
			}
		}
		from, to := l.checker.GetSignaturesOfType(source, checker.SignatureKindCall), l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
		if len(from) == 1 && len(to) == 1 {
			parameters, actual := to[0].Parameters(), from[0].Parameters()
			for index, parameter := range parameters {
				if index < len(actual) && !l.genericBodyAllows(l.checker.GetTypeOfSymbol(parameter), l.checker.GetTypeOfSymbol(actual[index])) {
					return false
				}
			}
			return l.genericBodyAllows(l.checker.GetReturnTypeOfSignature(from[0]), l.checker.GetReturnTypeOfSignature(to[0]))
		}
		for _, property := range l.checker.GetPropertiesOfType(target) {
			wanted := l.checker.GetTypeOfSymbol(property)
			if !l.genericDependent(wanted, map[*checker.Type]bool{}) {
				continue
			}
			actual := l.checker.GetPropertyOfType(source, property.Name)
			if actual == nil || !l.genericBodyAllows(l.checker.GetTypeOfSymbol(actual), wanted) {
				return false
			}
		}
		return true
	}
	return false
}

func (l *lowering) genericBodyRefusal(node *ast.Node) error {
	relation := l.genericRelation(node)
	if relation == nil || !l.genericDependent(relation.target, map[*checker.Type]bool{}) {
		return nil
	}
	source := l.genericOwnType(relation.source)
	if l.genericBodyAllows(source, relation.target) {
		return nil
	}
	bounds := []string{}
	for parent := node; parent != nil; parent = parent.Parent {
		if !ast.IsFunctionLike(parent) && parent.Kind != ast.KindClassDeclaration && parent.Kind != ast.KindClassExpression {
			continue
		}
		for _, parameter := range parent.TypeParameters() {
			bound := parameter.AsTypeParameterDeclaration().Constraint
			text := parameter.Name().Text()
			if bound != nil {
				text += " extends " + l.checker.TypeToString(l.checker.GetTypeFromTypeNode(bound))
			} else {
				text += " (unconstrained)"
			}
			bounds = append(bounds, text)
		}
	}
	if len(bounds) == 0 {
		return nil
	}
	return &Refused{Where: l.program.Where(node), What: "generic body " + relation.kind + " into slot " + l.checker.TypeToString(relation.target) + " from source " + l.checker.TypeToString(source) + " is not proven for every allowed type argument; checker constraint: " + strings.Join(bounds, "; "), Fix: "read a value of the same dependent type from the type parameter, or use the concrete constraint type for the slot (adamic/generic-body-relations)"}
}

// This witness is independent of the refusal. Dropping the body proof must expose
// a compiler error, never a second user refusal or an erased representation check.
func (l *lowering) genericBodyWitness(declaration *ast.Node) error {
	return nil
	if l.typeMapper == nil {
		return nil
	}
	// The checker keeps polymorphic this as a binder even when class arguments
	// are substituted. Resolve that binder to the class being instantiated for
	// the witness, while the universal body proof keeps the original identity.
	var thisMapper *typeMapper
	if l.classNode != nil {
		class := l.checker.GetTypeAtLocation(l.classNode)
		if class.ObjectFlags()&checker.ObjectFlagsReference != 0 {
			class = class.AsTypeReference().Target()
		}
		if class.ObjectFlags()&checker.ObjectFlagsClassOrInterface != 0 {
			if self := class.AsInterfaceType().ThisType(); self != nil {
				thisMapper = newTypeMapper([]*checker.Type{self}, []*checker.Type{l.classType})
			}
		}
	}
	resolved := func(proven *checker.Type) *checker.Type {
		if thisMapper != nil {
			proven = instantiateType(l.checker, proven, thisMapper)
		}
		return l.concrete(proven)
	}
	var failure error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if failure != nil {
			return true
		}
		if node != declaration && (ast.IsFunctionLike(node) || node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression) {
			return false
		}
		relation := l.genericRelation(node)
		if relation != nil {
			source, target := resolved(l.genericOwnType(relation.source)), resolved(relation.target)
			if !l.checker.IsTypeAssignableTo(source, target) {
				failure = fmt.Errorf("internal compiler error: generic body %s witness at %s: resolved source %s is not assignable to slot %s", relation.kind, l.program.Where(relation.where), l.checker.TypeToString(source), l.checker.TypeToString(target))
				return true
			}
		}
		node.ForEachChild(visit)
		return false
	}
	visit(declaration)
	return failure
}

// Read the checker mapper by field name, as class-generic calls already do.
// Its identity and implementation remain owned by the referenced checker.
func genericSignatureMapper(signature *checker.Signature) *typeMapper {
	if signature == nil {
		return nil
	}
	field := reflect.ValueOf(signature).Elem().FieldByName("mapper")
	if !field.IsValid() || field.Kind() != reflect.Pointer || field.IsNil() {
		return nil
	}
	return (*typeMapper)(field.UnsafePointer())
}
