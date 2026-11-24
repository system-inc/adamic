package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// nullishOnlyStep reports whether a step that names no field only removes or keeps undefined and
// null, and so passes outside an access reference's local tests counting no proof (Adamic's ruling
// on #s8gy89r, Nov 22, family 1's U8). handled says the step is one of the two shapes; passes is its
// answer, which stepAt returns as is.
//
// The two shapes: after is before less undefined and null (a truthiness or `!= null` test's true
// side), or after is exactly before's undefined and null members (its false side), where before
// also holds a member that is neither. Inside an access reference's local tests, and for a name,
// the true side still counts its proof as it always has; outside them it passes and counts none.
// The false side never counts one: on it the reference holds no object, so no tag comes from it.
func (w *enumTagWalk) nullishOnlyStep(at *ast.Node, before, after *checker.Type) (handled bool, passes bool) {
	c := w.l.checker
	if checker.Checker_isTypeIdenticalTo(c, after, c.GetNonNullableType(before)) {
		if w.tests == nil || w.inLocalTests(at) {
			return true, w.proves(at)
		}
		return true, true
	}
	if before.Flags()&checker.TypeFlagsUnion == 0 {
		return false, false
	}
	var nullish []*checker.Type
	others := 0
	for _, member := range before.Types() {
		if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
			nullish = append(nullish, member)
		} else {
			others++
		}
	}
	if len(nullish) == 0 || others == 0 {
		return false, false
	}
	var left []*checker.Type
	if after.Flags()&checker.TypeFlagsUnion != 0 {
		left = after.Types()
	} else {
		left = []*checker.Type{after}
	}
	if len(left) != len(nullish) {
		return false, false
	}
	for _, member := range left {
		found := false
		for _, want := range nullish {
			if member == want {
				found = true
				break
			}
		}
		if !found {
			return false, false
		}
	}
	return true, true
}

// inLocalTests reports whether node sits in one of an access reference's local tests.
func (w *enumTagWalk) inLocalTests(node *ast.Node) bool {
	if node == nil {
		return false
	}
	for _, test := range w.tests {
		if node.Pos() >= test.Pos() && node.End() <= test.End() {
			return true
		}
	}
	return false
}
