Preserved checker-context witness; fixture 25 remains blocked at 1156:33.
Commits: nullable merge cc73ab53; report and guard repair follow on this branch.
Checks: lower 28.057s, ir 18.217s; nullable dependency witnesses 127.010s; JavaScript package has no tests.
Mutant: dropping nullable-reference refusal caught by TestNullableReferenceLookupNeedsTag (exit 1).
Not covered: contextual map fix, its Node/backend proof and mutant, reference-kind sentinel generalization, full gate or next fixture stop.

Date: 2026-10-08 UTC.

Nullable support 1a83ba1ae65fe0e010dd6840c232a136a6caad37 merged as cc73ab53. This dependency migrates strings only; nullable.c returns NULL for other reference kinds. No claim of a completed reference-kind migration.

Fixture 25 remains blocked at 1156:33. The exact checker observations for a minimal generic witness are:

```typescript
function rows<T>(xs: T[]): T[][] { return xs.map(() => []); }
const numbers = rows([1, 2]);
const objects = rows([{ name: "x" }]);
```

- map call: own never[][], contextual T[][].
- callback: own () => never[], contextual (value: T, index: number, array: T[]) => never[].
- empty literal: own never[], contextual never[].

Thus the destination element context does not currently reach the empty literal. The diagnostic witness in internal/lower/contextual_empty_map_test.go records these observations and the existing refusal. This is not a completed fix or a Node/backend equivalence proof. A syntax-based exploratory fresh-return exception made this witness lower, but was discarded because the requested approach uses checker contextual types.

Pending choice: extend the existing fresh-value proof across map callback returns, or fix contextual propagation in the checker shim. No feature mutant, fixture-25 pass, or next-stop advancement is claimed.

Setup log /tmp/host-walk-setup.log: total 30.366s, clang .186s, build 30.106s, cache 30.340s; nproc 5. Focused lower controls passed in .109s (/tmp/contextual-empty-map-controls.log). Nullable witness and affected package runs are recorded in /tmp/host-walk-nullable-witnesses.log and /tmp/host-walk-nullable-packages.log.

Final affected-package results: /tmp/host-walk-nullable-packages-final.log. Explicit restored guard mutant: /tmp/host-walk-nullable-guard-mutant.log. Nullable dependency witnesses passed (127.010s); this run overlapped the guard mutant briefly, so the affected package results and separate focused controls are the authoritative guard verification.
