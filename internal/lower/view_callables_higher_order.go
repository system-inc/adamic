package lower

import "github.com/microsoft/TypeScript/tsc/shim/checker"

// A closure representation says nothing about its parameter or result domains.
// Higher-order reads require the complete logical producer certificate, selected
// by the existing closure-record registry after all function bodies are lowered.
func (l *lowering) viewCallableNeedsLogicalProducer(target *checker.Type) bool {
	for _, signature := range l.checker.GetSignaturesOfType(target, checker.SignatureKindCall) {
		children := []*checker.Type{l.checker.GetReturnTypeOfSignature(signature)}
		for _, parameter := range signature.Parameters() {
			children = append(children, l.checker.GetTypeOfSymbol(parameter))
		}
		for _, child := range children {
			if l.callableViewContract(l.checker.GetNonNullableType(child)) {
				return true
			}
		}
	}
	return false
}

func (l *lowering) viewCallableNestedShape(target *checker.Type, active map[*checker.Type]bool) bool {
	// This adapter uses optional closure pointers. Null-bearing callbacks need
	// their nullable adapter rather than having null stripped from the contract.
	if l.includesNull(target) {
		return false
	}
	target = l.checker.GetNonNullableType(target)
	if active[target] {
		return false
	}
	active[target] = true
	defer delete(active, target)
	signatures := l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(signatures) != 1 || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)) != 0 {
		return false
	}
	signature := signatures[0]
	if len(signature.TypeParameters()) != 0 || signature.HasRestParameter() || predicateOfSignature(l.checker, signature) != nil {
		return false
	}
	children := []*checker.Type{l.checker.GetReturnTypeOfSignature(signature)}
	for _, parameter := range signature.Parameters() {
		children = append(children, l.checker.GetTypeOfSymbol(parameter))
	}
	for _, child := range children {
		if child.Flags()&checker.TypeFlagsVoid != 0 {
			continue
		}
		if l.callableViewContract(l.checker.GetNonNullableType(child)) {
			if !l.viewCallableNestedShape(child, active) {
				return false
			}
		} else if !l.viewCallableBoxedRepresentation(child) {
			return false
		}
	}
	return true
}
