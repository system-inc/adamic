package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	_ "unsafe"
)

// Use the checker through its submodule, as the existing instantiation bridges do.
// The generated shim does not yet expose this normalization entry point.
//
//go:linkname reducedOptionalSource github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).getReducedType
func reducedOptionalSource(receiver *checker.Checker, source *checker.Type) *checker.Type

// optionalWidening names the relation whose optional field could be hidden by a structural view.
// This is independent of mutable invariance: even a readonly optional field can lie on a read.
type optionalWidening struct {
	property       string
	conflict       *checker.Type
	declaration    *ast.Node
	source, target *checker.Type
}

func (l *lowering) optionalWidened(source, target *checker.Type, skip map[string]bool, visited map[[2]*checker.Type]bool) *optionalWidening {
	if source == nil || target == nil {
		return nil
	}
	// An impossible relation constituent cannot hide an optional property. This
	// reduction says nothing about whether the outer expression is unreachable.
	source = reducedOptionalSource(l.checker, source)
	if source.Flags()&checker.TypeFlagsNever != 0 {
		return nil
	}
	if weak := l.weakTarget(source); weak != nil {
		source = weak
	}
	if weak := l.weakTarget(target); weak != nil {
		target = weak
	}
	// Only an accepted assignment is a widening. Function rest parameters and checked
	// downcasts can pair unrelated types; those belong to their own relation rules.
	if skip == nil && !l.checker.IsTypeAssignableTo(source, target) {
		return nil
	}
	if source == target || visited[[2]*checker.Type{source, target}] {
		return nil
	}
	visited[[2]*checker.Type{source, target}] = true
	if source.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range source.Types() {
			if found := l.optionalWidened(member, target, skip, visited); found != nil {
				return found
			}
		}
		return nil
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if l.checker.IsTypeAssignableTo(source, member) {
				if found := l.optionalWidened(source, member, skip, visited); found != nil {
					return found
				}
			}
		}
		return nil
	}
	if source.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return l.optionalWidened(l.checker.GetBaseConstraintOfType(source), target, skip, visited)
	}
	if target.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return l.optionalWidened(source, l.checker.GetBaseConstraintOfType(target), skip, visited)
	}
	if source.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 || target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
		return nil
	}
	// Container elements are views too. Avoid walking library methods, whose generic signatures
	// describe operations rather than value fields.
	if (l.checker.IsArrayType(source) || checker.IsTupleType(source)) && (l.checker.IsArrayType(target) || checker.IsTupleType(target)) {
		sources, targets := l.checker.GetTypeArguments(source), l.checker.GetTypeArguments(target)
		for index, inside := range sources {
			at := index
			if !checker.IsTupleType(target) {
				at = 0
			}
			if at < len(targets) {
				if found := l.optionalWidened(inside, targets[at], nil, visited); found != nil {
					return found
				}
			}
		}
		return nil
	}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		if skip[property.Name] {
			continue
		}
		declared := l.checker.GetPropertyOfType(source, property.Name)
		if declared == nil {
			if property.Flags&ast.SymbolFlagsOptional != 0 {
				l.optionalViewWriteField(property.Name)
				found := &optionalWidening{property: property.Name, source: source, target: target}
				if !isClassInstance(source) {
					return found
				}
				if conflict, declaration := l.optionalClassConflict(source, property); conflict != nil {
					found.conflict, found.declaration = conflict, declaration
					return found
				}
			}
			continue
		}
		if found := l.optionalWidened(l.checker.GetTypeOfSymbol(declared), l.checker.GetTypeOfSymbol(property), nil, visited); found != nil {
			return found
		}
	}
	sources := l.checker.GetSignaturesOfType(source, checker.SignatureKindCall)
	targets := l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(sources) == 1 && len(targets) == 1 && len(sources[0].TypeParameters()) == 0 && len(targets[0].TypeParameters()) == 0 {
		if found := l.optionalWidened(l.checker.GetReturnTypeOfSignature(sources[0]), l.checker.GetReturnTypeOfSignature(targets[0]), nil, visited); found != nil {
			return found
		}
		takes, gives := sources[0].Parameters(), targets[0].Parameters()
		for index := 0; index < len(takes) && index < len(gives); index++ {
			if found := l.optionalWidened(l.checker.GetTypeOfSymbol(gives[index]), l.checker.GetTypeOfSymbol(takes[index]), nil, visited); found != nil {
				return found
			}
		}
	}
	return nil
}

// Adamic compiles the whole program together: this closed world includes every runtime
// subclass, even class expressions. Separate compilation must revisit this exemption.
func (l *lowering) optionalClassConflict(source *checker.Type, property *ast.Symbol) (*checker.Type, *ast.Node) {
	// The standalone census has no loaded program and inventories structural relations only.
	if l.program == nil {
		return nil, nil
	}
	seenFiles := map[*ast.SourceFile]bool{}
	var conflict *checker.Type
	var declaration *ast.Node
	for _, root := range l.program.Files() {
		modules, err := l.moduleOrder(root)
		if err != nil {
			// Import cycles are refused by module ordering before the relation pass.
			continue
		}
		for _, file := range modules {
			if seenFiles[file] {
				continue
			}
			seenFiles[file] = true
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if conflict != nil {
					return true
				}
				if node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression {
					var candidate *checker.Type
					if node.Kind == ast.KindClassDeclaration && node.Name() != nil {
						candidate = l.checker.GetTypeAtLocation(node.Name())
					} else {
						signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node), checker.SignatureKindConstruct)
						if len(signatures) > 0 {
							candidate = l.checker.GetReturnTypeOfSignature(signatures[0])
						}
					}
					if candidate != nil {
						base := l.optionalClassBase(candidate, source.Symbol(), map[*ast.Symbol]bool{})
						if base == nil {
							return node.ForEachChild(visit)
						}
						patterns, actual := l.checker.GetTypeArguments(base), l.checker.GetTypeArguments(source)
						bindings := map[*checker.Type]*checker.Type{}
						possible := len(patterns) == len(actual)
						for index := range patterns {
							if !possible || !l.optionalClassArguments(patterns[index], actual[index], bindings) {
								possible = false
								break
							}
						}
						if !possible {
							return node.ForEachChild(visit)
						}
						if len(bindings) > 0 {
							from, to := []*checker.Type{}, []*checker.Type{}
							for parameter, argument := range bindings {
								from = append(from, parameter)
								to = append(to, argument)
							}
							candidate = instantiateType(l.checker, candidate, newTypeMapper(from, to))
						}
						field := l.checker.GetPropertyOfType(candidate, property.Name)
						if field != nil && !l.checker.IsTypeAssignableTo(l.checker.GetTypeOfSymbol(field), l.checker.GetTypeOfSymbol(property)) {
							conflict = candidate
							if len(field.Declarations) > 0 {
								declaration = field.Declarations[0]
							}
							return true
						}
					}
				}
				return node.ForEachChild(visit)
			}
			file.AsNode().ForEachChild(visit)
			if conflict != nil {
				return conflict, declaration
			}
		}
	}
	return nil, nil
}

func (l *lowering) optionalClassBase(candidate *checker.Type, source *ast.Symbol, seen map[*ast.Symbol]bool) *checker.Type {
	if candidate.Symbol() == source {
		return candidate
	}
	if seen[candidate.Symbol()] {
		return nil
	}
	seen[candidate.Symbol()] = true
	for _, base := range l.classBases(candidate) {
		if found := l.optionalClassBase(base, source, seen); found != nil {
			return found
		}
	}
	return nil
}

// Class type arguments are invariant. Solve inherited type parameters before checking a
// descendant's property, so Derived<T> extends Base<T> is judged as Derived<number>
// for a Base<number> source, and a Base<string> descendant cannot poison that view.
func (l *lowering) optionalClassArguments(pattern, actual *checker.Type, bindings map[*checker.Type]*checker.Type) bool {
	if pattern.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if prior := bindings[pattern]; prior != nil {
			return l.checker.IsTypeAssignableTo(prior, actual) && l.checker.IsTypeAssignableTo(actual, prior)
		}
		if constraint := l.checker.GetBaseConstraintOfType(pattern); constraint != nil && !l.checker.IsTypeAssignableTo(actual, constraint) {
			return false
		}
		bindings[pattern] = actual
		return true
	}
	if pattern.ObjectFlags()&checker.ObjectFlagsReference != 0 && actual.ObjectFlags()&checker.ObjectFlagsReference != 0 && pattern.Target() == actual.Target() {
		from, to := l.checker.GetTypeArguments(pattern), l.checker.GetTypeArguments(actual)
		if len(from) != len(to) {
			return false
		}
		for index := range from {
			if !l.optionalClassArguments(from[index], to[index], bindings) {
				return false
			}
		}
		return true
	}
	// An unresolved compound generic argument cannot prove that a descendant is impossible.
	// Include it conservatively; an unproven property type then refuses the exemption.
	if l.optionalUnresolvedArgument(pattern, map[*checker.Type]bool{}) {
		return true
	}
	return l.checker.IsTypeAssignableTo(pattern, actual) && l.checker.IsTypeAssignableTo(actual, pattern)
}

func (l *lowering) optionalUnresolvedArgument(proven *checker.Type, seen map[*checker.Type]bool) bool {
	if seen[proven] {
		return false
	}
	seen[proven] = true
	if proven.Flags()&(checker.TypeFlagsTypeParameter|checker.TypeFlagsIndexedAccess|checker.TypeFlagsConditional|checker.TypeFlagsSubstitution|checker.TypeFlagsIndex|checker.TypeFlagsTemplateLiteral|checker.TypeFlagsStringMapping) != 0 {
		return true
	}
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			if l.optionalUnresolvedArgument(member, seen) {
				return true
			}
		}
	}
	if proven.ObjectFlags()&checker.ObjectFlagsReference != 0 {
		for _, argument := range l.checker.GetTypeArguments(proven) {
			if l.optionalUnresolvedArgument(argument, seen) {
				return true
			}
		}
	}
	if proven.Flags()&checker.TypeFlagsObject != 0 {
		for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
			for _, signature := range l.checker.GetSignaturesOfType(proven, kind) {
				if l.optionalUnresolvedArgument(l.checker.GetReturnTypeOfSignature(signature), seen) {
					return true
				}
				for _, parameter := range signature.Parameters() {
					if l.optionalUnresolvedArgument(l.checker.GetTypeOfSymbol(parameter), seen) {
						return true
					}
				}
			}
		}
		for _, property := range l.checker.GetPropertiesOfType(proven) {
			if l.optionalUnresolvedArgument(l.checker.GetTypeOfSymbol(property), seen) {
				return true
			}
		}
	}
	return false
}

func (l *lowering) optionalValue(node *ast.Node, target *checker.Type) *optionalWidening {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
		// The literal's fields, spreads and elements are checked separately at their own sites.
		return nil
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		for _, branch := range []*ast.Node{conditional.WhenTrue, conditional.WhenFalse} {
			if found := l.optionalValue(branch, target); found != nil {
				return found
			}
		}
		return nil
	}
	return l.optionalWidened(l.checker.GetTypeAtLocation(node), target, nil, map[[2]*checker.Type]bool{})
}

// optionalAtSite uses the same relation sites as the other view rules, before lowering.
// A leading spread is not proof of absence: it copies fields hidden by the source's type too.
func (l *lowering) optionalAtSite(node *ast.Node) *optionalWidening {
	if l.nodeFSFileReadOnlyArgument(node) {
		return nil
	}
	// Only argument positions of a directly imported node:* call can use the host exemption.
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindCallExpression && l.optionalNodeHostCall(parent) {
		for index, argument := range parent.AsCallExpression().Arguments.Nodes {
			if argument == node && l.nodeHostConsumesArgument(parent, index) {
				return nil
			}
		}
	}
	var found *optionalWidening
	switch node.Kind {
	case ast.KindAsExpression:
		found = l.optionalValue(node.AsAsExpression().Expression, l.checker.GetTypeAtLocation(node))
	case ast.KindShorthandPropertyAssignment:
		target := l.checker.GetContextualType(node.Parent, checker.ContextFlagsNone)
		if target != nil {
			found = l.optionalValue(node.Name(), l.checker.GetTypeOfPropertyOfType(target, node.Name().Text()))
		}
	case ast.KindSpreadAssignment:
		target := l.checker.GetContextualType(node.Parent, checker.ContextFlagsNone)
		if target == nil {
			return nil
		}
		skip := map[string]bool{}
		after := false
		for _, property := range node.Parent.AsObjectLiteralExpression().Properties.Nodes {
			if property == node {
				after = true
			} else if after && property.Name() != nil {
				skip[property.Name().Text()] = true
			}
		}
		expression := ast.SkipParentheses(node.AsSpreadAssignment().Expression)
		if expression.Kind != ast.KindObjectLiteralExpression {
			found = l.optionalWidened(l.checker.GetTypeAtLocation(expression), target, skip, map[[2]*checker.Type]bool{})
		}
	default:
		if node.Kind == ast.KindParenthesizedExpression || !l.isExpression(node) || l.genericFunction(node) {
			return nil
		}
		var target *checker.Type
		if viewSite(node) {
			target = l.checker.GetContextualType(node, checker.ContextFlagsNone)
		}
		if target == nil {
			target = l.impliedTarget(node)
		}
		if target == nil {
			return nil
		}
		found = l.optionalValue(node, target)
	}
	return found
}

func (l *lowering) refuseOptionalWidening(node *ast.Node) error {
	found := l.optionalAtSite(node)
	if found == nil {
		return nil
	}
	// The same view machinery used by interface downcasts preserves checks through
	// aliases and descendant fields. A class conflict now requires a view, too.
	_, err := l.view(node, nil, found.target)
	return err
}

// Resolve the import binding rather than trusting the callee's spelling.
func (l *lowering) optionalNodeHostCall(call *ast.Node) bool {
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	if callee.Kind == ast.KindPropertyAccessExpression {
		callee = ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	}
	symbol := l.checker.GetSymbolAtLocation(callee)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		for ancestor := declaration; ancestor != nil; ancestor = ancestor.Parent {
			if ancestor.Kind == ast.KindImportDeclaration {
				specifier := ancestor.ModuleSpecifier()
				return specifier != nil && strings.HasPrefix(specifier.Text(), "node:")
			}
		}
	}
	return false
}
