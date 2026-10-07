# Discriminant writes

Decision by @system_adamic, October 6, refined October 7.

A discriminant is a property accepted by TypeScript's own narrowing predicate,
`isDiscriminantProperty`. After construction, a write is accepted only when the
written value's type is assignable to the property's declared type in every
union member that can inhabit the receiver's static view. Compound updates use
the result type; destructuring follows the source property domain. Structural
base interfaces are included. This check is shared by both backends.

A member's own literal remains writable. A declared domain wider than one
literal remains writable within that domain, including an open numeric enum.
Fresh construction uses the existing holder-confinement proof; assigning to an
escaped constructor receiver is no longer construction. Locals and allocation
helper results are provisional until that proof runs. A fresh scalar value is
not proof that its holder is fresh. A constructor increment that violates the
holder's own declared literal is independently refused: construction cannot
return a value that already contradicts its declared type.

Refused programs (executable witnesses are in the lower/oracle discriminant tests):

```typescript
type Leaf = { kind: 'leaf'; value: number };
type Branch = { kind: 'branch'; label: string };
type Node = Leaf | Branch;
function change(node: Node): void { node.kind = 'branch'; }
change({ kind: 'leaf', value: 4 });
```

Exact refusal: `Adamic 0.1 refuses a write to discriminant field 'kind' that
could move the object to variant "branch"; changing variant means building a
new object`. The original mutable_kind_guard witness prints `branch undefined`
on Node. Adamic refuses its write before lowering. The native missing-field
backstop remains unchanged.

```typescript
// Flags is an ambient numeric enum with A = 1 and B = 2.
type Node = { flags: Flags.B; kind: 'b' } | { flags: Flags.A; kind: 'a' };
function change(node: Node): void { node.flags = 2; }
```

Exact refusal: `Adamic 0.1 refuses a write to discriminant field 'flags' that
could move the object to variant Flags.B; changing variant means building a
new object`. Checking only the first union member loses the across-variant
refusal; the required mutant is held by a checker-level test.

```typescript
let escaped: Base | undefined;
class Base {
    kind: string = 'leaf';
    constructor() { escaped = this; this.kind = 'branch'; }
}
type Node = { kind: 'leaf' } | { kind: 'branch' };
console.log(new Base().kind);
```

This has the same exact `kind` refusal after the existing proof finds the holder
published. Setting a Leaf's own kind to 'leaf' is accepted even after publication,
because it passes the ruled per-member test.

The TypeScript 6.0.3 census and every event comparison are in
[the census report](../notes/discriminant-writes/typescript-writes.md).
