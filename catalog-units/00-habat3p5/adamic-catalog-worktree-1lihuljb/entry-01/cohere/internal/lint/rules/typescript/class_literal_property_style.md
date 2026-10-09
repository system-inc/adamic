# `@typescript-eslint/class-literal-property-style`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **8** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce that literals on classes are exposed in a consistent style

## Where cohere is deliberately quieter than ESLint

Upstream skips a member written `override`, because converting an overriding member changes what it
overrides, but it reads only the keyword. Without `noImplicitOverride` an override carries no
keyword, and upstream then reports a conversion the compiler rejects: a field overriding a concrete
base accessor is TS2610, and a getter overriding a concrete base property is TS2611. cohere asks the
checker for the base member and stays silent exactly where the conversion would not compile
(`classLiteralConversionBreaksAnOverride`), using the vendored checker's own conditions from
`checkKindsOfPropertyMemberOverrides`: instance members only, neither side private, the base not
from a mapped type, and the base not abstract or from an interface (any declaration for an
intersection property, all of them otherwise). A field may implement an abstract accessor, so the
seven schema `typeName` getters over `abstract get typeName()` stay true positives.

The rule now declares `NeedsTypeChecker` and `ReadsProgram`, for this check alone: the base class is
usually in another file.

The real site that went silent is
`libraries/structure/libraries/nexus/source/validation/schema/StringSchema.ts:37`, a `typeDefault`
getter returning the empty string over `BaseSchema`'s concrete `typeDefault` getter, where ESLint's
suggested field is TS2610 (checked by the scout with the compiler). 8 findings before, 7 after.

Fixtures: `TestClassLiteralPropertyStyleDeclinesAConversionThatCannotCompile` in
`class_literal_property_style_test.go`. Four rows go silent (the StringSchema shape beside its true
positive, a concrete getter-and-setter base, a concrete getter two classes up, and a field over a
concrete property in getters style). The rest are reporting controls: abstract getter, abstract
property, a getter over a base property (already TS2611, which the field repairs), no base
counterpart, statics, interface-merged members, a mapped-type base, an intersection with an interface
side, and three already-invalid inheritances the checker skips. As of 2026-10-04 the site no longer
differs: both engines read 0 on `StringSchema.ts` (#cn8sthd).

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

8 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/validation/schema/ArraySchema.ts:23`**

```
get typeName(): string {
```

> Literals should be exposed using readonly fields

**`libraries/structure/libraries/nexus/source/validation/schema/BooleanSchema.ts:16`**

```
get typeName(): string {
```

> Literals should be exposed using readonly fields

**`libraries/structure/libraries/nexus/source/validation/schema/DateSchema.ts:23`**

```
get typeName(): string {
```

> Literals should be exposed using readonly fields

**`libraries/structure/libraries/nexus/source/validation/schema/FileSchema.ts:20`**

```
get typeName(): string {
```

> Literals should be exposed using readonly fields

**`libraries/structure/libraries/nexus/source/validation/schema/NumberSchema.ts:16`**

```
get typeName(): string {
```

> Literals should be exposed using readonly fields

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

