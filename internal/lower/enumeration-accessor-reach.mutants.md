# Enumeration accessor reach (#1atvfp5, family 17, Adamic's ruling of Nov 22)

"Object enumeration with getter or accessor descriptors" used to refuse
Object.keys, values and entries, and encodeJson, whenever any get or set
accessor existed anywhere in the program. enumerationDescriptors now asks
enumerationAccessorReaches of each accessor (enumeration_accessor_reach.go):
it refuses only when the accessor's owner (an object literal's regular type, a
class's instance or static side and every class inheriting the accessor, an
interface's or type literal's declared type) and a type the enumeration reads
are related by assignability, either way. Those are the operand's type, a
union both as itself and by its members, and for encodeJson, which reads every field at any depth,
recursively their property types, type arguments and element types. A read
type that can't be classified (any, unknown, object, {}, a type parameter, an
index signature whose values can be objects, a type nested past depth 16) keeps
the program-wide scan, and an owner inside type parameters, or one whose name
doesn't lead back to its accessor, counts as reaching.

TestEnumerationReachUnknownKeepsProgramScan pins the unknown causes an
enumeration reaches: object, {}, a readonly string index whose values are
objects (by Object.values), an empty interface and a chain past depth 16 (by
encodeJson). The rest are fail-closed and unreachable today, each stopped first
elsewhere: any is refused as "a value of type any" (refuseAnyValue); a type
parameter is mapped to its instantiation by l.concrete before the guard reads
it, since a generic body is lowered per instantiation; encodeJson's schema
(jsonDecodeSchema) refuses every non-array index signature; and a mutable
string index is taken by recordObjectCallArguments as a record call without
asking the guard, as at main.

Each producer is a top-level TestProduct_EnumerationReach<Name>Mutant:
t.Parallel(), one buildcache.GoTestMutated build of
internal/lower/enumeration_accessor_reach.go, no subtests and no other
assertions. TestEnumerationReachMutantsAreCaught runs each pin in its mutant
binary, requires exit 1, and requires each listed row's witness line, so a pin
that fails for some other reason doesn't count as a catch. The pins also run
unmutated in enumeration_accessor_reach_test.go and in the whole internal/lower.

| Mutant | Mutation | Catching test | Catch line in enumeration_accessor_reach_test.go |
| --- | --- | --- | --- |
| WholeGraph: every accessor reaches, the scan of before | `if reach.unknown {` to `if true \|\| reach.unknown {` | TestEnumerationReachUnrelatedGetterLowers | 70, `enumeration reach row unrelated was refused` |
| UnionMembers: a union's members are left unread | `case flags&checker.TypeFlagsUnion != 0:` to `case false && ...` | TestEnumerationReachUnionLowers | 70, `enumeration reach row union was refused` |
| DeepReach: encodeJson reads only the operand's own type | `if !deep {` to `if true {` | TestEnumerationReachDeepGetterRefused/(field\|element) | 90, `enumeration reach row <row> was admitted` |
| DiscriminatedUnion: a union is read by its members alone | `types = append(types, next.proven) // the union itself` to `// the union itself is left unread` | TestEnumerationReachReachingGetterRefused/discriminated | 90, `enumeration reach row discriminated was admitted` |
| OneWay: an owner reaches only when assignable to a read type, not the other way | `l.checker.IsTypeAssignableTo(owner, read) \|\| l.checker.IsTypeAssignableTo(read, owner)` to `l.checker.IsTypeAssignableTo(owner, read)` | TestEnumerationReachDowncastRefused | 90, `enumeration reach row downcast was admitted` |

UnionMembers fails closed rather than open: a union left unread falls to the
classifier's last case, a type that isn't an object type, and makes reach
unknown, so the union pin, which must lower, is refused. The members are what
encodeJson's deep read needs, for a field one member alone declares, which
TestEnumerationReachDeepUnionGetterRefused pins unmutated.

DiscriminatedUnion: tsc relates a source to a discriminated union as a whole by
decomposing its discriminant, so a class whose kind is 1 | 2 is assignable to
{ kind: 1; g: number } | { kind: 2; g: number } while relating to neither member
either way (Lower's witness on bb242562). The union is therefore read as itself
too; under the mutant it's read by its members alone and the discriminated row
is admitted.

OneWay: a checked downcast in .ts keeps the same object at a slot its owner
type is not assignable to, so reach is judged both ways (the hand's soundness
doubt on 6ac68068). The pin's class with a getter sits in held.ts, seen in
main.ts as a narrower shape and cast down to a wider one the class is not
assignable to; under the mutant the guard admits it, and the pin fails with its
witness whatever stops the program after.

Every pin reaches the guard through Object.values, Object.entries or encodeJson,
and Object.keys only in the unrelated pin's shape with an optional field:
objectKeys (class_features.go) lowers a class instance, a static side and a
plain view straight to ir.ObjectKeys, and calls objectCall, and so the guard,
only for a node array, a RegExp, a possibly undefined operand or a shape with an
optional field. Object.values of a class instance, a constructor or a union with
an interface member is any[] to tsc, and read as join's receiver it stops at "an
array of any" before the call is lowered, so those refused rows return the call
from a function typed number[].

Product order: cloud/fast-gate/enumeration-reach-product-order.py writes the
five producer rows, each waiting on nothing.
