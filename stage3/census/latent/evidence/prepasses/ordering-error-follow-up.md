## Ordering-error follow-up and measurement provenance

The census numbers and cost samples above were taken before the ordering-error
fix, on compiler base `0816e18511b5e1ef4383543ee81c526fc36acfaf` with the measured
code checkpoint `a2d4b5783`. They have not been rerun or relabelled as measurements
of the recut base `fd8b74b9bd4fa5d8012dfcc0c0252c3f1964cd4d`. The complete change is now squashed on that newer
main tip as `typescript/census-prepasses-2`.

In this API ESMModuleOrder returns `(modules, cyclic)`, not an error. Its second
result was the cycle flag. Production moduleOrder adds the error-producing
preflight, including dynamicImportGap. Both latentPrepassModules copies now call
that production path on a disposable lowering and return `([]*ast.SourceFile,
error)`. A failed root returns nil and the original error, discarding even an order
accumulated from earlier successful roots. No prepass, registration or fact capture
runs over a partial result. Cycle/read-proof state stays on the disposable probe,
so the successful unit's existing first-query effects remain in place.

First-error and predicates units use latentRecord and return from the unit.
Full mode records registration failure at file scope even if every unit is deferred,
and records the error and an explicit boundary for each otherwise eligible unit.
Speculative mode records file and unit failures with latentSpecFailure, preserves
failed boundaries and skips both unit and whole-file walks. Errors are never cached
as a successful empty module list.

A real dynamic-import fixture exercises production's moduleOrder refusal. The first
root succeeds and a later root fails; even the first root's clean units must emit
NotYet at the failing import. The separately built overlay passes this test in
first-error, full and speculative modes, plus a full-mode deferred-only file case.
The pre-fix binary fails all four assertions. ESMModuleOrder cannot itself produce
an error in this signature; no injected error is needed because the production
moduleOrder error is exercised directly.

The historical predicates textual refusal generator is already incompatible with
current refuse syntax. Its scratch validation overlay uses the current tested AST
refusal rewriter and unused-import cleaner, with first-error-only compatibility
stubs. Its actual predicates driver passes the same real ordering-error test. This
scratch compatibility change is not applied to the checked-in historical generator.

The preserved 81-file manifest was hash-verified and scanned with stock TypeScript
for the exact import-call AST condition used by dynamicImportGap. There are no such
calls, and the module-order/error-preflight source is unchanged between the measured
and recut bases. Thus the new error branch cannot fire for the measured inputs; its
successful module list and unit state are unchanged. The original numbers remain
valid evidence for that measured success path. This does not claim a new census
of unrelated compiler changes on the newer main tip. See ordering-fix-validation.json
and the ordering-fix logs for the checks run after this correction.
