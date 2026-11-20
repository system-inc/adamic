package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// unionFieldStorageDiffers names how the object members of a receiver's union store field name, or is empty
// when they all store it one way (#svtsk2x). A field read takes one representation, the read's own type's,
// and reads every object's slot at it, but each object stores the field as its own type declares it: XA's
// a: string a string, YA's a: number an unboxed number. A read of s.a through XA | YA is string | number,
// held as one Union, so it read YA's number as a boxed pointer and retained it, a native crash where Node
// prints 3. Two members that share a literal tag (both Kind.A) are what lets a switch leave such a union
// standing, but the hazard is the storage, so any union reaching the read is judged. A member storing the
// field in a representation this can't name fails closed. An intersection of object types is judged as an
// object is: (XA & { tag: 1 }) | YA stores a as XA's string and YA's number all the same (Compiler's must). A
// primitive, a branded primitive (string & { brand }), an array and a tuple are another read's, and are
// skipped; any other member, one this can't name, fails closed.
func (l *lowering) unionFieldStorageDiffers(receiver *checker.Type, name string) string {
	receiver = l.checker.GetNonNullableType(l.concrete(receiver))
	if receiver.Flags()&checker.TypeFlagsUnion == 0 {
		return ""
	}
	var first ir.Type
	var firstMember *checker.Type
	for _, member := range receiver.Types() {
		switch {
		case member.Flags()&checker.TypeFlagsPrimitive != 0 || unionFieldBranded(member):
			continue
		case member.Flags()&checker.TypeFlagsObject != 0 && (l.isArrayType(member) || checker.IsTupleType(member)):
			continue
		case member.Flags()&checker.TypeFlagsObject != 0, unionFieldObjectIntersection(member):
		default:
			return l.checker.TypeToString(member) + " can't be judged here"
		}
		property := l.checker.GetPropertyOfType(member, name)
		if property == nil {
			continue
		}
		stored, known := l.representation(l.checker.GetTypeOfSymbol(property))
		if !known {
			return l.checker.TypeToString(member) + " stores it in a representation not known here"
		}
		if firstMember == nil {
			first, firstMember = stored, member
			continue
		}
		if stored != first {
			return l.checker.TypeToString(firstMember) + " as " + typeName(first) + ", " + l.checker.TypeToString(member) + " as " + typeName(stored)
		}
	}
	return ""
}

// unionFieldObjectIntersection says whether a type is an intersection whose every member is an object type.
func unionFieldObjectIntersection(member *checker.Type) bool {
	if member.Flags()&checker.TypeFlagsIntersection == 0 {
		return false
	}
	for _, part := range member.Types() {
		if part.Flags()&checker.TypeFlagsObject == 0 {
			return false
		}
	}
	return true
}

// unionFieldBranded says whether a type is an intersection holding a primitive, a branded primitive, which is
// that primitive at run time.
func unionFieldBranded(member *checker.Type) bool {
	if member.Flags()&checker.TypeFlagsIntersection == 0 {
		return false
	}
	for _, part := range member.Types() {
		if part.Flags()&checker.TypeFlagsPrimitive != 0 {
			return true
		}
	}
	return false
}
