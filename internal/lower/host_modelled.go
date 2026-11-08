package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// hostModelledNever is docs/axioms/host-any.md's Modelled rows typed never, in code: whether node's
// value is one the host audit models as never, so it's never had. The cast rule (castProof) and the
// any-relation rule (refuseAnyRelation) read it and nothing else, so neither names the declarations
// behind a row, and a later Modelled row typed never is added here, nowhere else.
//
// Today that's the require row (module.d.ts:707): a value whose evaluation starts with one of
// nodeCommonJSRequireThrows' two forms, through parentheses, assertions, non-nulls and the member
// accesses and calls made on it, which nodeCommonJSRequireRoot already finds. The require throws before
// anything after it in the chain runs, so no part of the chain is ever had.
func (l *lowering) hostModelledNever(node *ast.Node) bool {
	return l.nodeCommonJSRequireRoot(node) != nil
}
