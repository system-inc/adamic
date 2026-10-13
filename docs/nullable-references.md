# Nullable references

A single reference type R, combined with exactly one empty case, uses one nullable
pointer. R can be a string, array, object, class, Map, Set or function. NULL means
null in R | null and undefined in R | undefined. Locals, fields, array elements,
parameters, returns, closures and generic instantiations use the same representation.
The empty case belongs to the instantiated static type, not to the pointer bits.

String observation helpers, templates and concatenation spell the empty case as
`null` or `undefined`. typeof null is `object`; typeof undefined is `undefined`.
Number converts a nullable string's null to zero and its undefined to NaN.
An empty string remains a present string: its numeric conversion is zero and its
negation is true, while ?? preserves it. Negation of other references tests absence.
Non-null assertions and loose equality remain refused in `.a` programs.

Strict comparisons with null and undefined consult the instantiated static type.
A null pointer in a string | null parameter is never strictly equal to undefined.
A null pointer in a string | undefined parameter is never strictly equal to null.
Generic instance keys distinguish these types, including collection arguments.
A generic typeof function also distinguishes instantiations at null and undefined.
Observation helpers use ordinary IR and evaluate their operand once.

JSON schemas carry the empty case separately from their present representation.
Null serializes as `null` at the top level, as a field and as an element. Undefined
returns undefined at the top level, omits the object field and writes `null` as an
array element. Present function values retain JSON's own omission behavior.

## Why two empty cases wait

R | null | undefined has three cases. One pointer distinguishes present from absent,
but cannot distinguish two absent values. It needs an empty-case tag and a separate
design for slots, ownership and views. Lowering says NotYet with the named reason
`nullable reference needs an empty-case tag`, rather than guessing a meaning.

These programs are refused, also when their types arise through monomorphization:

```ts
function observe(v: string | null | undefined): string { return typeof v; }
function optional(v?: string | null): string { return typeof v; }
function defaulted(v: string | null = 'default'): string { return typeof v; }
function wider(v: string | null): void {
    const viewed: string | null | undefined = v;
}
function widerField(v: { readonly x: string | null }): void {
    const viewed: { readonly x?: string | null } = v;
}
function generic<T>(v: T): void {
    const viewed: T | undefined = v;
}
generic<string | null>(null);
```

The same widening from R | undefined to R | null | undefined is NotYet. Arrays,
map arguments and function parameters and returns are checked recursively. A
R | null to R | undefined view is rejected by TypeScript; generic views that add
the other empty case are rejected after substitution. Optional reads introduce
undefined, so indexing an array of nullable references or getting a nullable map
value can require the tag too. Iterate over such elements directly instead.
A defaulted nullable parameter also needs to distinguish an omitted argument from
an explicitly supplied null, so it waits for that design.

## Current limits

Console declarations currently accept a single string argument. Witnesses that
pass nullable strings directly wait for host-5; templates and coalescing provide
string arguments for the observations retained here. It does not format arrays
or objects. Object and class JSON
references remain NotYet: a structural type can hide additional fields or toJSON,
and the existing JSON design only proves complete shapes for direct literals.
ToPrimitive conversion of array, object, class, map, set and function references,
and template interpolation or concatenation of these references, remains NotYet.
Their proven empty values do lower after narrowing, including Number and JSON;
an unnarrowed nullable reference still needs its present-value operation proved. This unit does not invent a
present-value coercion to make an empty-case fixture pass.

The oracle fixtures cover nullable string operations, generic functions and classes
across reference kinds, reference negation, optional reads, narrowing, coalescing,
closures, and JSON of strings, arrays, maps and functions. Lowering tests hold the
refusals above, and assertions and loose equality remain refused.

## Inline reads and narrowing

The tag refusal applies to every expression, before comparison lowering or other
observations. A missing `Map<K, R | null>.get()` and an index read of `(R | null)[]`
have type R | null | undefined, even when their immediate use is a strict equality,
template or argument. They are NotYet with the same named tag reason.
An immediate typeof is accepted for array index/at/pop and Map.get lookups:
main's typeof-null-2 fix retains the container presence slot until classification,
so a present NULL means null and a missing slot means undefined. The nullable
pointer representation preserves that distinction. The former typeof refusal
predated typeof-null-2; typeof_null_slots.a agrees with source Node in both backends.
`TestNullableReferenceReadsNeedATagInEveryExpression` holds these cases, including
the integration match-or-null probe.

The current Map declaration's has() returns boolean and does not narrow get().
TypeScript also does not retain narrowing across repeated get() calls. Neither is
an exception to the expression guard. A local triple-empty value remains refused,
including before an undefined test. The sound neighbor therefore tests a
Map<K, R> read (one undefined case) narrowed with !== undefined, and direct
iteration over Map<K, R | null> values (one null case). These compile and agree
with Node. Supporting a nullable get() with a proven presence refinement needs a
checker-visible refinement; it is not inferred merely from has().

## Method presence

Method truthiness and strict null/undefined comparisons observe the receiver once
without constructing a detached closure. Own closure fields shadow prototype
methods, including an undefined own field. Method names remain observable when
the method signature cannot use a callable closure slot. Stored method values,
callbacks and returned detached methods stay refused. Ambient host methods and
library methods without runtime descriptors stay NotYet.

The Node mutant checks cover empty-value spelling, typeof, Number, JSON, distinct
generic bodies, observation evaluation count and method presence. Each mutant
compiles and runs cleanly and differs only in stdout from the original source on
Node; wasm32 runs the same comparison when `ADAMIC_ORACLE_WASI=1`. Refusal mutants
separately hold defaulted nullable parameters, widening views, triple-empty reads
and shadowed undefined names.
