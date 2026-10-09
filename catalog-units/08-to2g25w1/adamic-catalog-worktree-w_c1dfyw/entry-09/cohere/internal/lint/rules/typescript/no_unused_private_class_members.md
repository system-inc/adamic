# `@typescript-eslint/no-unused-private-class-members`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | **4** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow unused private class members

## Deliberate divergence: a `declare` field is never unused

Upstream reports `declare private readonly __brand: X` when nothing reads it (measured on the
installed 8.67.0 build, instance and static alike). Cohere does not, by @system_cohere's ruling of
2026-10-02 (#r2fbv77). A `declare` field emits nothing: no slot, no initializer, no runtime
existence. It is there to make the class nominal, so never being read is its purpose, and deleting it
would change what the type checker accepts while removing no code. A real `private` field beside the
brand still reports.

So cohere is silent where the gate reports on api-phi-health's `TypedBinding` subclasses, which brand
each binding this way.

## Why this recommendation

Catches a defect rather than a preference, and the cleanup is small enough to do in one pass.

## Violations

4 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/numbers/Decimal.ts:385`**

```
private signedCoefficient(): bigint {
```

> Private class member 'signedCoefficient' is defined but never used

**`libraries/structure/libraries/nexus/source/numbers/Decimal.ts:757`**

```
private toFixedPointStringPadded(decimalPlaces: number): string {
```

> Private class member 'toFixedPointStringPadded' is defined but never used

**`libraries/structure/source/api/web-sockets/shared-worker/WebSocketConnection.ts:58`**

```
private boundHandleInternetAvailable: EventListener | null = null;
```

> Private class member 'boundHandleInternetAvailable' is defined but never used

**`libraries/structure/source/api/web-sockets/shared-worker/WebSocketConnection.ts:59`**

```
private boundHandleInternetUnavailable: EventListener | null = null;
```

> Private class member 'boundHandleInternetUnavailable' is defined but never used

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

