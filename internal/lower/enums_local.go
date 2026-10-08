package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

func localConstEnum(declaration *ast.Node) bool {
	return declaration.Parent != nil && declaration.Parent.Kind == ast.KindBlock && ast.HasSyntacticModifier(declaration, ast.ModifierFlagsConst)
}

// Node transforms even local const enums into hoisted runtime vars. Inlining is
// sound only when the declaration dominates the read. A direct later read in
// its lexical block has that proof; deferred bodies need a separate call proof.
func (l *lowering) localConstEnumRead(node, declaration *ast.Node) error {
	if !localConstEnum(declaration) {
		return nil
	}
	parent := node
	for parent != nil && parent != declaration.Parent {
		if parent == declaration {
			return nil
		}
		if ast.IsFunctionLike(parent) {
			return l.notYet(node, "a local const enum read from a deferred body; its per-call initialization is not proven")
		}
		parent = parent.Parent
	}
	if parent == nil || node.Pos() < declaration.End() {
		return l.notYet(node, "a local const enum read before its lexical initialization; move the read after the declaration")
	}
	return nil
}
