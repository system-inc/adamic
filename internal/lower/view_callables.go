package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// signatureProven must come from the implementation certificate, never from a
// runtime function tag or the asserted target type. Reads still check readiness
// and kind; a certificate only authorizes using the declared calling convention.
func (l *lowering) viewCallableCall(node *ast.Node, member string, signatureProven bool) error {
	return checkViewCallableSignature(l.program.Where(node), member, signatureProven)
}

func checkViewCallableSignature(where, member string, signatureProven bool) error {
	if signatureProven {
		return nil
	}
	return &Refused{
		Where: where,
		What:  "a checked view call to member " + member,
		Fix:   "its signature cannot be checked at runtime and no compatible implementation is proven; prove the callable implementation before calling through this view",
	}
}
