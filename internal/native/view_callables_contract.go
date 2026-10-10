package native

import "fmt"

// The caller owns presence/readiness and the value's existing lifetime. Shape
// checking returns the same pointer and takes no extra reference. recorded must
// select producer metadata, never an asserted target descriptor.
func emitViewCallableShape(value, recorded, expected, expression string, optional bool) string {
	return fmt.Sprintf("((adamic_closure *)adamic_view_callable_shape((const adamic_heap *)(%s), %s, %s, %s, %t))", value, recorded, expected, cString(expression), optional)
}
