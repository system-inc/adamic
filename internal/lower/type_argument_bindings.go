package lower

import (
	"reflect"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A type argument bound to a constrained type parameter is seen as its constraint wherever the generic's
// body sees the type parameter, and no view site outside the body shows it. Two rules judge each binding:
// the declared-key rule in .a (records_declared_keys.go), and the relation rule (constraint_relation.go,
// #xywvn06), which judges it by the registry with the constraint as its slot. Both walk the bindings
// typeArgumentBindings finds, so there is one list of the places a type argument is bound.

// typeArgumentBinding is one type argument a site binds to a constrained type parameter.
type typeArgumentBinding struct {
	// site is the call, new or tagged template, the class's extends clause, or the generic function or
	// class written as a value its slot instantiates.
	site      *ast.Node
	parameter *checker.Type

	// argument is what the parameter is bound to, and constraint its constraint as instantiated there.
	// argument is nil when the binding can't be read, and constraint is then the declared one.
	argument, constraint *checker.Type

	// written is the type argument written for the parameter at the site, or nil when it's inferred or
	// defaulted.
	written *ast.Node

	// generic and mapper are the generic signature a call's, a new's or a value's binding instantiates
	// and its instantiation, so the arguments the binding was inferred from can be found. A class's
	// extends clause has no generic signature.
	generic *checker.Signature
	mapper  *typeMapper
}

// typeArgumentBindings visits each type argument node binds to a constrained type parameter. The
// bindings are a call's, a new's or a tagged template's, inferred or written, read from its resolved
// signature, which reaches generic arrows and methods too; a class's extends clause's; and a generic
// function's or class's, instantiated as the value its slot takes (valueTypeArgumentBindings), a call's
// result included. A binding whose type argument can't be read is visited with no argument, so each rule
// fails closed on it as its own rule needs.
func (lowering *lowering) typeArgumentBindings(node *ast.Node, visit func(typeArgumentBinding) error) error {
	switch node.Kind {
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression:
		// The bindings the call makes, then the value it yields, which may itself be a generic function
		// its slot instantiates: const put: (target: Options) => void = maker() (g05).
		if signatureTypeArgumentBindingsError := lowering.signatureTypeArgumentBindings(
			node,
			lowering.checker.GetResolvedSignature(node),
			visit,
		); signatureTypeArgumentBindingsError != nil {
			return signatureTypeArgumentBindingsError
		}
	case ast.KindExpressionWithTypeArguments:
		clause := node.Parent
		if clause == nil ||
			clause.Kind != ast.KindHeritageClause ||
			clause.AsHeritageClause().Token != ast.KindExtendsKeyword ||
			clause.Parent == nil ||
			!ast.IsClassLike(clause.Parent) {
			return nil
		}
		return lowering.baseTypeArgumentBindings(node, clause.Parent, visit)
	}
	if node.Kind == ast.KindParenthesizedExpression || !lowering.isExpression(node) || ast.IsDeclarationName(node) {
		return nil
	}
	if node.Parent != nil &&
		node.Parent.Kind == ast.KindPropertyAccessExpression &&
		node.Parent.AsPropertyAccessExpression().Name() == node {
		return nil // The access is the value, never the name it reads.
	}
	return lowering.valueTypeArgumentBindings(node, visit)
}

// valueTypeArgumentBindings visits a generic function or class written as a value where its slot's type
// instantiates it, as contextualGenericSignature instantiates a function value: const make: new (held:
// Options) => Store<Options> = Store binds Store's T to Options with no call or new to judge it at. One
// left generic, with no slot or a generic one, binds nothing until it's called or constructed, and is
// judged then. One whose instantiation can't be read, since it or its slot has more than one signature,
// is visited unread.
func (lowering *lowering) valueTypeArgumentBindings(node *ast.Node, visit func(typeArgumentBinding) error) error {
	own := lowering.checker.GetTypeAtLocation(node)
	for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
		signatures := lowering.checker.GetSignaturesOfType(own, kind)
		var parameters []*checker.Type
		for _, signature := range signatures {
			parameters = append(parameters, signature.TypeParameters()...)
		}
		if len(parameters) == 0 {
			continue
		}
		contextual := lowering.concrete(lowering.slotType(node))
		if contextual == nil {
			continue
		}
		slots := lowering.checker.GetSignaturesOfType(lowering.withoutUndefined(contextual), kind)
		if len(slots) == 0 || len(slots) == 1 && len(slots[0].TypeParameters()) > 0 {
			continue
		}
		if len(signatures) == 1 && len(slots) == 1 {
			if resolved := instantiateSignatureInContextOf(
				lowering.checker,
				signatures[0],
				slots[0],
				nil,
				nil,
			); resolved != nil {
				if signatureTypeArgumentBindingsError := lowering.signatureTypeArgumentBindings(
					node,
					resolved,
					visit,
				); signatureTypeArgumentBindingsError != nil {
					return signatureTypeArgumentBindingsError
				}
				continue
			}
		}
		if unreadTypeArgumentBindingsError := lowering.unreadTypeArgumentBindings(
			node,
			parameters,
			visit,
		); unreadTypeArgumentBindingsError != nil {
			return unreadTypeArgumentBindingsError
		}
	}
	return nil
}

// signatureTypeArgumentBindings visits the bindings an instantiated signature made: its target's type
// parameters, each as its mapper maps it, which the checker keeps unexported and classGenericCall reads
// the same way. A signature that isn't an instantiation binds nothing. One whose bindings can't be read
// is visited unread.
func (lowering *lowering) signatureTypeArgumentBindings(
	node *ast.Node,
	signature *checker.Signature,
	visit func(typeArgumentBinding) error,
) error {
	if signature == nil {
		return nil
	}
	if len(signature.TypeParameters()) > 0 {
		return lowering.unreadTypeArgumentBindings(node, signature.TypeParameters(), visit)
	}
	target := signature.Target()
	if target == nil || len(target.TypeParameters()) == 0 {
		return nil
	}
	field := reflect.ValueOf(signature).Elem().FieldByName("mapper")
	if !field.IsValid() || field.Kind() != reflect.Pointer || field.IsNil() {
		return lowering.unreadTypeArgumentBindings(node, target.TypeParameters(), visit)
	}
	mapper := (*typeMapper)(field.UnsafePointer())
	var written []*ast.Node
	switch node.Kind {
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression:
		written = node.TypeArguments()
	}
	for index, parameter := range target.TypeParameters() {
		constraint := lowering.checker.GetConstraintOfTypeParameter(parameter)
		if constraint == nil {
			continue
		}
		binding := typeArgumentBinding{
			site:       node,
			parameter:  parameter,
			argument:   instantiateType(lowering.checker, parameter, mapper),
			constraint: instantiateType(lowering.checker, constraint, mapper),
			generic:    target,
			mapper:     mapper,
		}
		if index < len(written) {
			binding.written = written[index]
		}
		if visitError := visit(binding); visitError != nil {
			return visitError
		}
	}
	return nil
}

// baseTypeArgumentBindings visits the type arguments a class's extends clause (node) binds to its base
// class's type parameters, as the base type the checker resolved for the class carries them.
func (lowering *lowering) baseTypeArgumentBindings(
	node *ast.Node,
	class *ast.Node,
	visit func(typeArgumentBinding) error,
) error {
	symbol := lowering.checker.GetSymbolOfDeclaration(class)
	if symbol == nil {
		return nil
	}
	for _, base := range lowering.checker.GetBaseTypes(lowering.checker.GetDeclaredTypeOfSymbol(symbol)) {
		if base.Symbol() == nil {
			continue
		}
		parameters := lowering.checker.GetLocalTypeParametersOfClassOrInterfaceOrTypeAlias(base.Symbol())
		if len(parameters) == 0 {
			continue
		}
		if base.ObjectFlags()&checker.ObjectFlagsReference == 0 || base.Target() == nil {
			return lowering.unreadTypeArgumentBindings(node, parameters, visit)
		}
		// The generic class's own reference carries its type parameters where the base carries what
		// they're bound to, outer ones first, so each local parameter is found by its place there.
		declared, given := lowering.checker.GetTypeArguments(base.Target()), lowering.checker.GetTypeArguments(base)
		arguments := make([]*checker.Type, len(parameters))
		for index, parameter := range parameters {
			place := slices.Index(declared, parameter)
			if place < 0 || place >= len(given) {
				return lowering.unreadTypeArgumentBindings(node, parameters, visit)
			}
			arguments[index] = given[place]
		}
		mapper := newTypeMapper(parameters, arguments)
		written := node.TypeArguments()
		for index, parameter := range parameters {
			constraint := lowering.checker.GetConstraintOfTypeParameter(parameter)
			if constraint == nil {
				continue
			}
			binding := typeArgumentBinding{
				site:       node,
				parameter:  parameter,
				argument:   arguments[index],
				constraint: instantiateType(lowering.checker, constraint, mapper),
				mapper:     mapper,
			}
			if index < len(written) {
				binding.written = written[index]
			}
			if visitError := visit(binding); visitError != nil {
				return visitError
			}
		}
	}
	return nil
}

// unreadTypeArgumentBindings visits each constrained one of parameters as a binding at node whose type
// argument can't be read.
func (lowering *lowering) unreadTypeArgumentBindings(
	node *ast.Node,
	parameters []*checker.Type,
	visit func(typeArgumentBinding) error,
) error {
	for _, parameter := range parameters {
		if constraint := lowering.checker.GetConstraintOfTypeParameter(parameter); constraint != nil {
			if visitError := visit(
				typeArgumentBinding{site: node, parameter: parameter, constraint: constraint},
			); visitError != nil {
				return visitError
			}
		}
	}
	return nil
}
