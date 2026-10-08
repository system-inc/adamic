package lower

import "github.com/microsoft/TypeScript/tsc/shim/checker"

// tupleReadNarrowingType preserves JavaScript array identity when a primitive
// branch leaves a closed union of supported tuples. Reuse the tuple lane's
// complete, disjoint-arity proof; an ordinary object union gets no tuple marker.
func tupleReadNarrowingType(target *checker.Type) bool {
	return checker.IsTupleType(target) || tupleAlternativesType(target)
}
