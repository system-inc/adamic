package lower

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

type overloadRelationFailure struct{ path, relation string }

// Explain the same forward result relation used for admission. Readonly fields
// are covariant; writable fields additionally need the reverse relation. The
// explanation never changes whether a contract is served.
func (l *lowering) overloadResultFailure(produced, promised *checker.Type, path string, visited map[[2]*checker.Type]bool) overloadRelationFailure {
	produced, promised = l.concrete(produced), l.concrete(promised)
	describe := func(rule string, source, target *checker.Type) string {
		return rule + " requires " + l.checker.TypeToString(source) + " to fit " + l.checker.TypeToString(target)
	}
	fallback := overloadRelationFailure{path, describe("result covariance", produced, promised)}
	pair := [2]*checker.Type{produced, promised}
	if visited[pair] {
		return fallback
	}
	visited[pair] = true
	if produced.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range produced.Types() {
			if !l.censusRelated(member, promised) {
				return l.overloadResultFailure(member, promised, path, visited)
			}
		}
	}
	if promised.Flags()&checker.TypeFlagsObject != 0 {
		for _, property := range l.checker.GetPropertiesOfType(promised) {
			own := l.checker.GetPropertyOfType(produced, property.Name)
			next := path + "." + property.Name
			if own == nil {
				return overloadRelationFailure{next, "required field is absent from " + l.checker.TypeToString(produced)}
			}
			if own.Flags&ast.SymbolFlagsOptional != 0 && property.Flags&ast.SymbolFlagsOptional == 0 {
				return overloadRelationFailure{next, "required field cannot be served by an optional field"}
			}
			source, target := l.checker.GetTypeOfSymbol(own), l.checker.GetTypeOfSymbol(property)
			if !l.checker.IsReadonlySymbol(property) {
				if l.checker.IsReadonlySymbol(own) {
					return overloadRelationFailure{next, "writable field cannot be served by a readonly field"}
				}
				if !l.censusRelated(target, source) {
					return overloadRelationFailure{next, describe("writable invariance (reverse direction)", target, source)}
				}
			}
			if !l.censusRelated(source, target) {
				failure := l.overloadResultFailure(source, target, next, visited)
				if failure.path == next {
					rule := "writable invariance (forward direction)"
					if l.checker.IsReadonlySymbol(property) {
						rule = "readonly covariance"
					}
					failure.relation = describe(rule, source, target)
				}
				return failure
			}
		}
	}
	return fallback
}

func (l *lowering) overloadResultRelation(produced, promised *checker.Type) string {
	return l.overloadResultFailure(produced, promised, "result", map[[2]*checker.Type]bool{}).relation
}

// A visitor value is passed to the implementation as a callback. The
// implementation's invocation domain must fit the supplied callback's input,
// even when the outer overload's parameter relation is otherwise covariant.
func (l *lowering) overloadCallbackParameterRelation(given, takes *checker.Type) string {
	supplied := l.checker.GetSignaturesOfType(l.concrete(given), checker.SignatureKindCall)
	invoked := l.checker.GetSignaturesOfType(l.concrete(takes), checker.SignatureKindCall)
	if len(supplied) != 1 || len(invoked) != 1 {
		return ""
	}
	for index, parameter := range invoked[0].Parameters() {
		if index >= len(supplied[0].Parameters()) {
			break
		}
		domain := l.censusCallableParameterType(parameter)
		input := l.censusCallableParameterType(supplied[0].Parameters()[index])
		if !l.censusRelated(domain, input) {
			return fmt.Sprintf("; callback contravariance at callback.parameter%d requires %s to fit %s", index+1, l.checker.TypeToString(domain), l.checker.TypeToString(input))
		}
	}
	return ""
}
