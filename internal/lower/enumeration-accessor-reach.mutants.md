# Enumeration accessor reach (#1atvfp5, family 17, Adamic's ruling of Nov 22)

"Object enumeration with getter or accessor descriptors" used to refuse
Object.keys, values and entries, and encodeJson, whenever any get or set
accessor existed anywhere in the program. enumerationDescriptors now asks
enumerationAccessorReaches of each accessor (enumeration_accessor_reach.go):
it refuses only when the accessor's owner (an object literal's regular type, a
class's instance or static side and every class inheriting the accessor, an
interface's or type literal's declared type) is assignable to a type the
enumeration reads. Those are the operand's type and its union members, and for
encodeJson, which reads every field at any depth, recursively their property
types, type arguments and element types. A read type that can't be classified
(any, unknown, object, {}, a type parameter, an index signature whose values can
be objects, a type nested past depth 16) keeps the program-wide scan, and an
owner inside type parameters, or one whose name doesn't lead back to its
accessor, counts as reaching.

Each producer is a top-level TestProduct_EnumerationReach<Name>Mutant:
t.Parallel(), one buildcache.GoTestMutated build of
internal/lower/enumeration_accessor_reach.go, no subtests and no other
assertions. TestEnumerationReachMutantsAreCaught runs each pin in its mutant
binary, requires exit 1, and requires each listed row's witness line, so a pin
that fails for some other reason doesn't count as a catch. The pins also run
unmutated in enumeration_accessor_reach_test.go and in the whole internal/lower.

| Mutant | Mutation | Catching test | Catch line in enumeration_accessor_reach_test.go |
| --- | --- | --- | --- |
| WholeGraph: every accessor reaches, the scan of before | `if reach.unknown {` to `if true \|\| reach.unknown {` | TestEnumerationReachUnrelatedGetterLowers | 53, `enumeration reach row unrelated was refused` |
| UnionMembers: a union's members are left unread | `case flags&checker.TypeFlagsUnion != 0:` to `case false && ...` | TestEnumerationReachUnionLowers | 53, `enumeration reach row union was refused` |
| DeepReach: encodeJson reads only the operand's own type | `if !deep {` to `if true {` | TestEnumerationReachDeepGetterRefused/(field\|element) | 67, `enumeration reach row <row> was admitted` |

UnionMembers fails closed rather than open: a union left unread falls to the
classifier's last case, a type that isn't an object type, and makes reach
unknown, so the union pin, which must lower, is refused. A union that is read
only as itself would not fail open for Object.keys, values and entries either,
since an owner is assignable to a union when it is to a member; only encodeJson's
deep read needs the members, for a field one member alone declares, which
TestEnumerationReachDeepUnionGetterRefused pins unmutated.

Product order: cloud/fast-gate/enumeration-reach-product-order.py writes the
three producer rows, each waiting on nothing.
