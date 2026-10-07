package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// predicateRefusal is the admission seam for the syntax refusal visitor. Keep
// tag-only proofs behind the checked-view handoff: a kind test does not prove
// initialization or representation of the target's other fields.
func (l *lowering) predicateRefusal(node *ast.Node) error {
	proof, err := l.provePredicate(node)
	if err != nil {
		return err
	}
	if proof.TaggedView {
		return l.notYet(node, "checked field reads for a predicate-narrowed interface; the body proves its kind, but the checked-view handoff is not available")
	}
	return nil
}
