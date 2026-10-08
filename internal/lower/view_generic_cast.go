package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Preflight can defer a generic object/array assertion, but cannot certify it
// from its constraint. The instantiated body runs castProof again with concrete
// types and must produce an ordinary checked view. An unmapped parameter keeps
// the original refusal instead of becoming a trusted assertion.
func (l *lowering) genericViewCastCandidate(_ *ast.Node, source, target *checker.Type) bool {
	unresolved := target.Flags()&checker.TypeFlagsTypeParameter != 0
	if l.checker.IsArrayType(target) {
		unresolved = l.checker.GetElementTypeOfArrayType(target).Flags()&checker.TypeFlagsTypeParameter != 0
	}
	return unresolved && (source.Flags()&checker.TypeFlagsObject != 0 || l.checker.IsArrayType(l.withoutUndefined(source)))
}

func (l *lowering) unresolvedGenericViewCast(_ *ast.Node, _ ir.Expression, _, _ *checker.Type) (ir.Expression, error) {
	return nil, nil
}
