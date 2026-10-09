package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Instantiate producing binders in the receiving signature's context. A receiving
// generic binder remains rigid: its constraint is not a chosen runtime argument.
func (l *lowering) checkedSignaturePair(source, target *checker.Signature) (*checker.Signature, *checker.Signature) {
	if len(source.TypeParameters()) != 0 {
		source = checker.Checker_instantiateSignatureInContextOf(l.checker, source, target, nil, func(from, to *checker.Type, _ bool) checker.Ternary {
			if l.checker.IsTypeAssignableTo(from, to) {
				return checker.TernaryTrue
			}
			return checker.TernaryFalse
		})
	}
	return source, target
}

// An object method's contextual signature supplies the receiving binders for its
// return expressions. Pair by declaration identity, rather than a printed T name.
func (l *lowering) checkedSiteMapper(node *ast.Node) *typeMapper {
	for function := node.Parent; function != nil; function = function.Parent {
		if !ast.IsFunctionLike(function) {
			continue
		}
		actual := l.checker.GetSignatureFromDeclaration(function)
		if actual == nil || len(actual.TypeParameters()) == 0 {
			return l.typeMapper
		}
		contextual := l.checker.GetContextualType(function, checker.ContextFlagsNone)
		if function.Kind == ast.KindMethodDeclaration && function.Parent != nil && function.Parent.Kind == ast.KindObjectLiteralExpression {
			object := l.checker.GetContextualType(function.Parent, checker.ContextFlagsNone)
			if object != nil {
				contextual = l.checker.GetTypeOfPropertyOfType(object, function.Name().Text())
			}
		}
		if contextual == nil {
			return l.typeMapper
		}
		signatures := l.checker.GetSignaturesOfType(contextual, checker.SignatureKindCall)
		if len(signatures) != 1 {
			return l.typeMapper
		}
		instantiated, _ := l.checkedSignaturePair(actual, signatures[0])
		mapping := map[*checker.Type]*checker.Type{}
		for index, parameter := range actual.Parameters() {
			if index < len(instantiated.Parameters()) {
				l.inferTypes(l.checker.GetTypeOfSymbol(parameter), l.checker.GetTypeOfSymbol(instantiated.Parameters()[index]), mapping)
			}
		}
		l.inferTypes(l.checker.GetReturnTypeOfSignature(actual), l.checker.GetReturnTypeOfSignature(instantiated), mapping)
		sources, targets := []*checker.Type{}, []*checker.Type{}
		for _, parameter := range actual.TypeParameters() {
			if given := mapping[parameter]; given != nil {
				sources = append(sources, parameter)
				targets = append(targets, given)
			}
		}
		if len(sources) != 0 {
			return newTypeMapper(sources, targets)
		}
		return l.typeMapper
	}
	return l.typeMapper
}

// Removing readonly preserves the exact values and optionality of every key of
// the same rigid binder. Other mapped transformations do not establish this proof.
func (l *lowering) checkedMutableIdentity(from, to *checker.Type) bool {
	if from.Flags()&checker.TypeFlagsObject == 0 || from.ObjectFlags()&checker.ObjectFlagsMapped == 0 || from.Alias().Symbol() == nil {
		return false
	}
	unchanged := false
	for _, declaration := range from.Alias().Symbol().Declarations {
		if declaration.Kind != ast.KindTypeAliasDeclaration {
			continue
		}
		node := declaration.AsTypeAliasDeclaration().Type
		if node.Kind != ast.KindMappedType {
			continue
		}
		mapped := node.AsMappedTypeNode()
		unchanged = mapped.NameType == nil && mapped.QuestionToken == nil && mapped.ReadonlyToken != nil && mapped.ReadonlyToken.Kind == ast.KindMinusToken
	}
	if !unchanged {
		return false
	}
	mapped := from.AsMappedType()
	mapped.ResolveComponents(l.checker, from)
	keys, value := mapped.ConstraintType(), mapped.TemplateType()
	return keys != nil && keys.Flags()&checker.TypeFlagsIndex != 0 && keys.AsIndexType().Target() == to && value != nil && value.Flags()&checker.TypeFlagsIndexedAccess != 0 && value.AsIndexedAccessType().ObjectType() == to && value.AsIndexedAccessType().IndexType() == mapped.TypeParameter()
}

// Compare one overload without assuming another overload accepts its arguments.
func (l *lowering) checkedSignatureWidening(producing, receiving *checker.Signature, visited map[[2]*checker.Type]bool) *widening {
	if l.censusNeverRestSignature(receiving) {
		if l.censusDiscardedMarkerPredicate(producing, receiving) {
			return nil
		}
		source, target := l.checker.GetReturnTypeOfSignature(producing), l.checker.GetReturnTypeOfSignature(receiving)
		if !l.checker.IsTypeAssignableTo(source, target) {
			return &widening{source: source, target: target}
		}
		return l.widened(source, target, visited)
	}
	producing, receiving = l.checkedSignaturePair(producing, receiving)
	fromParameters, toParameters := producing.Parameters(), receiving.Parameters()
	for index, parameter := range fromParameters {
		takes := l.censusCallableParameterType(parameter)
		given := l.checker.GetUndefinedType()
		if index < len(toParameters) {
			given = l.censusCallableParameterType(toParameters[index])
		}
		if !l.enumAssignable(given, takes) || !l.checker.IsTypeAssignableTo(given, takes) {
			return &widening{source: takes, target: given, parameter: true}
		}
		if found := l.widened(given, takes, visited); found != nil {
			return found
		}
	}
	source, target := l.checker.GetReturnTypeOfSignature(producing), l.checker.GetReturnTypeOfSignature(receiving)
	// A void callback discards its result. It does not expose a writable slot
	// or a value typed void to its caller.
	if target.Flags() == checker.TypeFlagsVoid {
		return nil
	}
	if !l.enumAssignable(source, target) || !l.checker.IsTypeAssignableTo(source, target) {
		return &widening{source: source, target: target}
	}
	return l.widened(source, target, visited)
}
