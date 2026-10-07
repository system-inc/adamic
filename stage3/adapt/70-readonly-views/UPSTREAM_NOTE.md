# Public readonly API compatibility note

## 1. setTextRange reads its location without mutating it

Status: type-only adaptation 70, sanctioned by @system_adamic; draft upstream
compatibility note, not filed. Microsoft could accept or refuse this change.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Location: src/compiler/factory/utilitiesPublic.ts:10, input location parameter.
The emitted declaration is at api/typescript.d.ts:9185 in the pristine snapshot.

The implementation reads the input's pos and end and assigns those numbers to
the destination parameter. Readonly describes the input view, not object freezing. The stock-checker symbol audit, including every resolved
consumer body, finds no input writes or retained input aliases. Its owner proof
is in evidence/wave5/adapt.json. The destination parameter remains mutable.

Proposed published declaration difference, on top of adaptation 20's separately
sanctioned 189 optional additions:

```diff
- function setTextRange<T extends TextRange>(range: T, location: TextRange | undefined): T;
+ function setTextRange<T extends TextRange>(range: T, location: Readonly<TextRange> | undefined): T;
```

This makes the input contract truthful, but it is a source compatibility change
for downstream users deriving a mutable view from the published parameter type.
The stock TypeScript 6.0.3 checker proves this consumer compiles before the change
and reports two TS2540 diagnostics after it:

```typescript
import type * as ts from "typescript";
type View = NonNullable<Parameters<typeof ts.setTextRange>[1]>;
function adjust(view: View): void {
    view.pos = 1;
    view.end = 2;
}
```

Expected after: writes to pos and end fail because the view is readonly.
Observed: TS2540 at both assignments; zero diagnostics against the old API.
See evidence/wave5/consumer.json. This does not claim that TypeScript prevents
all deliberate readonly erasure or broader mutable aliasing downstream.

### Draft GitHub issue

Title: Describe setTextRange's non-mutating location input with a readonly view

Version: TypeScript 6.0.3, stock checker and Node 24.19.0.

The location argument is read-only throughout the compiler, services, server
and test source program. Readonly<TextRange> expresses that behavior and permits
sounder callers with narrower readonly positions. The change adds no JavaScript
and preserves every emitted JavaScript byte and all computational baselines.
The API snapshot change is mechanically exactly this one readonly owner addition
plus the separately measured 189 optional additions; all 60,930 other reference
baselines match. Microsoft may prefer the existing downstream mutable contract;
that compatibility choice is explicit, rather than treating readonly as having
no public effect. Nothing has been sent upstream and the global ledger is untouched.
