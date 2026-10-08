package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// An empty, implicitly typed Map construction takes its arguments from the
// checker's contextual type, including the asserted Map view. Explicit arguments
// are never replaced; any written by the source remains an adaptation.
func (l *lowering) mapTypeArguments(node *ast.Node) []*checker.Type {
	arguments := l.typeArguments(l.concrete(l.checker.GetTypeAtLocation(node)))
	if node.Kind != ast.KindNewExpression {
		return arguments
	}
	created := node.AsNewExpression()
	if created.TypeArguments != nil && len(created.TypeArguments.Nodes) != 0 {
		return arguments
	}
	if created.Arguments != nil && len(created.Arguments.Nodes) != 0 {
		return arguments
	}
	contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	if contextual == nil {
		return arguments
	}
	inferred := l.contextualMapArguments(l.concrete(contextual), map[*checker.Type]bool{})
	if len(inferred) != 2 {
		return arguments
	}
	for _, argument := range inferred {
		if argument.Flags()&checker.TypeFlagsAny != 0 {
			return arguments
		}
	}
	return inferred
}

// An asserted interface may inherit Map's storage while refining its methods,
// as PragmaMap does. Obtain storage arguments from the actual library base.
func (l *lowering) contextualMapArguments(proven *checker.Type, seen map[*checker.Type]bool) []*checker.Type {
	if proven == nil || seen[proven] {
		return nil
	}
	seen[proven] = true
	if l.isLibraryType(proven, "Map", "ReadonlyMap") {
		return l.typeArguments(proven)
	}
	if proven.Flags()&checker.TypeFlagsObject == 0 || proven.ObjectFlags()&checker.ObjectFlagsInterface == 0 {
		return nil
	}
	var found []*checker.Type
	for _, base := range l.checker.GetBaseTypes(proven) {
		arguments := l.contextualMapArguments(l.concrete(base), seen)
		if len(arguments) != 2 {
			continue
		}
		if found != nil && (found[0] != arguments[0] || found[1] != arguments[1]) {
			return nil
		}
		found = arguments
	}
	return found
}
