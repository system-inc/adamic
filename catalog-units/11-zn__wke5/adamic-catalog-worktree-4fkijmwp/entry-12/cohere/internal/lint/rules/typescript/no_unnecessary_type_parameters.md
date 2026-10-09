# `@typescript-eslint/no-unnecessary-type-parameters`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **27** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Disallow type parameters that aren't used multiple times

## Where cohere is deliberately quieter than ESLint

The exact type-equality idiom reports in ESLint and is silent here:

```ts
type IsExactlyType<TLeft, TRight> =
    (<T>() => T extends TLeft ? 1 : 2) extends <T>() => T extends TRight ? 1 : 2 ? true : false;
```

`T` used once is the mechanism, not decoration. The checker defers a conditional type whose check
type is an unresolved type parameter and relates two deferred conditionals only when their extends
types are identical, which is how this asks "is L exactly R" where assignability would let `any`
through. The rule's suggestion, replacing `T` with its constraint, resolves both sides eagerly and
destroys the comparison, so both findings are false.

The principle (`isTypeWitnessParameter`, Kirk's ruling of 2026-10-01): a type parameter is a witness,
not a disguised cast, when its owner has no value parameters (a `this` parameter counts as one) and the
parameter is referenced only inside the return type annotation. A cast needs something to cast from,
and a function nobody hands a value has nothing. The exact-equality idiom above meets it, and so does
`ObjectTypes.ts:93` `typeOnly<Shape>(): Shape`, the phantom-type witness behind about 228 call sites.
Both are silent here and reported by ESLint.

This replaced a narrower recognizer that matched the equality idiom by shape. Every case it silenced
has no value parameters and uses `T` only in its return, so the principle subsumes it.

A witness never produces a value of `T` (Kirk's ruling of 2026-10-02). When the function has a body,
every value it returns must be `null`, `undefined` or `void`, read through `as`, `!`, `satisfies` and
parentheses. A parameterless function that reads the world and returns it as `T` (`readConfig<T>(): T`
over `JSON.parse`, or api-phi-health's `ormDrizzleCredentialsFromEnvironment<CredentialsType>()`) is a
cast from I/O, and it reports, as ESLint does. A signature with no body, a function type or a
`declare`, has nothing to read and stays silent.

The real sites are `libraries/structure/libraries/nexus/source/types/UnionFromClasses.test.ts:23`
(columns 7 and 49) and `ObjectTypes.ts:93`. Upstream's own corpus pins the idiom as reporting
(invalid57, `Equal<X, Y>` over `Compute`); that case is asserted silent here. Fixtures live in
`no_unnecessary_type_parameters_test.go`: the equality idiom's silent rows, the witness rows, and
reporting controls for a value parameter, a `this` parameter, and a parameter that escapes the return
type.

### A phantom brand on a `declare` property

A class type parameter whose use is the type of a `declare` property is silent here and reported by
ESLint (measured on 8.67.0). `isPhantomBrandParameter`, by @system_cohere's ruling of 2026-10-02
(#ye9s2jx), from Base:

```ts
export class WorkerQueueBinding<MessageType = unknown> extends TypedBinding {
    declare private readonly __workerQueue: MessageType;
}
```

A `declare` property emits nothing, so the parameter is not a value the class holds and hands back
unchecked. It ties `new WorkerQueueBinding<StripeEvent>(name)` to its message type so a decorator can
thread `WorkerQueue<StripeEvent>` through, and the brand is what makes it part of the class's
structure at all; a second use would be invented to quiet the rule. `RpcClientBinding<RpcInterface>`
is the same shape. An ordinary property, a member of an ambient `declare class` (which carries no
`declare` modifier of its own), and a brand naming some other type still report, as fixtures in
`TestNoUnnecessaryTypeParametersPhantomBrands`.

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

27 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/command-line/CommandArguments.ts:354`**

```
export function jsonFlag<ParsedType>(flags: CommandArgumentFlagsType, key: string): ParsedType | undefined {
```

> Type parameter ParsedType is used only once in the function signature

**`libraries/structure/libraries/nexus/source/structured-text/json/Json.ts:42`**

```
export function convertRowsToJson<RowType extends Record<string, unknown>>(
```

> Type parameter RowType is used only once in the function signature

**`libraries/structure/libraries/nexus/source/structured-text/tables/Csv.ts:21`**

```
export function convertRowsToCsv<TRow extends Record<string, unknown>>(
```

> Type parameter TRow is used only once in the function signature

**`libraries/structure/libraries/nexus/source/time/TimeSeriesProcessors.ts:434`**

```
export function getTopBarDataKey<T extends Record<string, number | string>>(
```

> Type parameter T is used only once in the function signature

**`libraries/structure/libraries/nexus/source/types/Constructor.test.ts:441`**

```
function createInstance<T>(classType: unknown, ...constructorArguments: unknown[]): T | null {
```

> Type parameter T is used only once in the function signature

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

