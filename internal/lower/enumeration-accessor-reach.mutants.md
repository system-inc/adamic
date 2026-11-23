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
| GenericLiteral: any enclosing type parameters make a literal's owner unknown | `if holder == nil {` to `if holder == nil \|\| enumerationGenericContext(holder) {` | TestEnumerationReachGenericLiteralLowers | 70, `enumeration reach row generic literal was refused` |
| LiteralTypeParameter: a literal whose own type mentions T is owned by that type | `if l.eachReached(owner, 0,` to `if false && l.eachReached(owner, 0,` | TestEnumerationReachReachingGetterRefused/held | 90, `enumeration reach row held was admitted` |
| DeclaredOwner: an annotated literal's declared type is never taken | `if declared, named := ...; named {` to `...; false && named {` | TestEnumerationReachDeclaredLiteralLowers | 70, `enumeration reach row declared literal was refused` |
| GenericClass: a class's own type parameters make its owner unknown | `if enumerationGenericContext(class.Parent) {` to `if enumerationGenericContext(class) {` | TestEnumerationReachGenericClassLowers | 70, `enumeration reach row generic class was refused` |
| AnyInstance: a generic class is read uninstantiated | `owners = append(owners, l.enumerationAnyInstance(side, static))` to `owners = append(owners, side)` | TestEnumerationReachReachingGetterRefused/generic | 90, `enumeration reach row generic was admitted` |
| Contextual: a literal's contextual type is never taken | `if contextual := ...; contextual != nil {` to `...; false && contextual != nil {` | TestEnumerationReachContextualLiteralLowers | 70, `enumeration reach row contextual literal was refused` |
| ContextualScope: a contextual type naming an enclosing type parameter is taken | `if len(enclosing) != 0 && l.eachReached(contextual,` to `if false && ...` | TestEnumerationReachReachingGetterRefused/held | 90, `enumeration reach row held was admitted` |
| AssertedLiteral: an asserted literal is owned by its slot type, the asserted type | `around.Kind == ast.KindSatisfiesExpression \|\| around.Kind == ast.KindAsExpression \|\| around.Kind == ast.KindTypeAssertionExpression` to `around.Kind == ast.KindSatisfiesExpression` | TestEnumerationReachAssertedLiteralRefused | 90, `enumeration reach row asserted literal was admitted` |
| OneWay: an owner reaches only when assignable to a read type, not the other way | `l.checker.IsTypeAssignableTo(owner, read) \|\| l.checker.IsTypeAssignableTo(read, owner)` to `l.checker.IsTypeAssignableTo(owner, read)` | TestEnumerationReachDowncastRefused | 90, `enumeration reach row downcast was admitted` |
| ViewPass: no admitted enumeration is held against the checked views | `func (l *lowering) checkEnumerationReachViews() error {` to the same line and `return nil` | TestEnumerationReachViewRefused/sibling | 90, `enumeration reach row sibling was admitted` |

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

ViewPass: a checked view checks at a read, never at its cast, so a .ts
downcast keeps a getter-bearing object at a view whose type relates to the
object's neither way, and the guard alone admits Object.values of it (Lower's
probe J-m9g0n72: Node printed 2,1,4, native exited 70 on a field the checker
proved there). Each enumeration the guard admits beside an accessor is recorded
(enumerationReachSites, snapshot-owned in statecopy), and
checkEnumerationReachViews, after checkSlotWrites (pass ledger row, scanning
reference), widens its read types by the source of every view that can put a
value at one (viewsReaching, view_reach.go, shared with #9z2ksv7), classified as
an operand is and walked as deep, until nothing new is added, so a chain of casts
is followed link by link. A view with no source, or a source that can't be
classified, makes the site unknown. Every accessor is then asked again against
the widened types, and one that reaches, or an unknown site, stops in the guard's
words. TestEnumerationReachViewRefused pins the sibling ({ b; d: 4 }), Language's
sibling ({ b; c }) and a chain of two casts; the control with no cast is the
unrelated row. Under the mutant the post-pass returns at its top and the sibling
row is admitted.

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
fourteen producer rows, each waiting on nothing.

A literal with no annotated variable is owned by its contextual type (and each
member of a union one) when the checker gives one that names no type parameter
of a declaration around it: a declared return, a parameter, a declared object's
property. satisfies and a type assertion give none, since the literal keeps its own type there
(Lower's should on 5eacc825: an asserted type can relate to the literal's own type neither way).

A class's own type parameters are taken at any (enumerationAnyInstance), so its
instance side relates to a read type whenever some instantiation could, since
any is assignable to and from every type; a static side needs none, since no
static member names them. Type parameters around a class, or around an
interface or type literal, still leave the owner unknown.

A literal's owner: when the literal initializes a variable whose annotation
names no type parameter and nothing declared inside type parameters, the
declared type is its owner, since every slot holding the literal is reached
through that variable. Otherwise its regular type is, unless that type mentions
a type parameter (eachReached walks member signatures uninstantiated and stops
past depth 16, which is why tsc's createTransformationContext literal needed the
declared road, J-6b476xq). An enclosing generic function alone no longer makes a
literal's owner unknown; a class's accessor keeps that rule.
