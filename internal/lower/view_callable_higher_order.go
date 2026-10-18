package lower

import "github.com/microsoft/TypeScript/tsc/shim/checker"

// The receiving declaration determines the callback ABI. The closure's own
// declarations still determine the domains checked when that callback is called.
func (l *lowering) viewCallableHasCallback(target *checker.Type) bool {
	for _, signature := range l.checker.GetSignaturesOfType(l.checker.GetNonNullableType(target), checker.SignatureKindCall) {
		for _, parameter := range signature.Parameters() {
			if l.callableViewContract(l.checker.GetTypeOfSymbol(parameter)) {
				return true
			}
		}
		if l.callableViewContract(l.checker.GetReturnTypeOfSignature(signature)) {
			return true
		}
	}
	return false
}
