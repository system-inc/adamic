# Nullable references

A single reference type R, meaning a string, array, object or class, joined with exactly null or exactly undefined has one nullable pointer representation. The same rule applies in generic instantiations. This is the language contract; the scout does not implement it or claim every current lowering path conforms.

## Representation and ownership

In C, the present value is the ordinary typed reference pointer and the empty value is NULL. Empty strings and empty arrays remain present allocations or immortal values; neither is the empty pointer. Retain and release must accept NULL without allocating a box. The declared union determines whether NULL means null or undefined, and lowering must preserve that information through locals, fields, calls, returns, captures and containers.

The JavaScript backend preserves native JavaScript null and undefined. It must not translate one to the other. R | null | undefined remains NotYet until its own ruling: a single pointer cannot distinguish both empty values without additional information. Current main has some boxed-union support beyond this contract; that does not establish a general ruling for three-member nullable references.

## Observable empty cases

Node decides these results, independently of the pointer encoding.

| Operation on the empty value | null | undefined |
| --- | --- | --- |
| typeof | "object" | "undefined" |
| Template interpolation | "null" | "undefined" |
| String concatenation, "x" + value | "xnull" | "xundefined" |
| Numeric addition, 1 + value | 1 | NaN |
| Number(value) | 0 | NaN |
| JSON.stringify(value) | string "null" | undefined, not a string |
| JSON.stringify({ value }) | '{"value":null}' | '{}' |
| JSON.stringify([value]) | '[null]' | '[null]' |
| console.log(value) | null followed by newline | undefined followed by newline |

The + operator uses JavaScript conversion and evaluates each operand once in source order. Present object conversions may invoke valueOf, toString or Symbol.toPrimitive; pointer support alone does not implement those protocols. JSON serialization similarly needs complete shape information and toJSON semantics for present objects. Current unsupported protocols must remain explicit NotYet stops. Console formatting with multiple arguments retains Node's formatting rules.

## Generics and soundness

Monomorphization must prove the concrete reference type before selecting this representation. An unconstrained T can also instantiate to number or boolean. Those optional scalars require their own present-and-value representation. A cache key must retain the concrete type and empty-case meaning; sharing a pointer layout does not prove interchangeable field layouts, ownership or conversion behavior. Branded strings such as __String retain their proof while using string storage; an arbitrary intersection is not automatically a string.

Nullable references do not license unchecked casts, unproved predicates, mutable covariance, hidden optional fields or ownership cycles. Programs that claim a missing value is present without a proof remain subject to soundness checks. For example, widening a mutable holder of Dog | undefined to a holder of Animal | undefined still permits a write that violates the original type. Overload implementations must still prove each promised result. Use checked narrowing, an explicit fallback, readonly views or a copy as appropriate.

## Verification and limits

The scout map is `review/compiler/nullable-scout/map.md`. Its counts describe the pinned census and distinguish conditional representation candidates from soundness refusals and scalar work. The reductions carry upstream file and line provenance. No native or JavaScript lowering was changed by this unit. The empty-case table is independently exercised on Node; current support is recorded separately in the scout evidence.
