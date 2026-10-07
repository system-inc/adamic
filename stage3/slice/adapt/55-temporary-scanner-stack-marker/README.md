Temporary: comes out when function widening lands

Reconciled with adaptation 80: retain TypeScript's exact `stackCrawlMark?:
AnyFunction` declarations. The former `{}` rewrite is withdrawn because the
pinned Node declarations require Function for captureStackTrace's second
argument. No runtime statement or argument changes. New slices need no source
edit from 55; its adapter restores AnyFunction when replayed on an old opaque
slice and checks that the expected callable marker declarations are present.

This clears TS2740 without claiming that it closes the original function
widening refusal. Validation and the next compiler stop are recorded in the
scanner driver's BLOCKERS.md and evidence/marker-reconciliation.json.

Historical validation of the withdrawn opaque marker follows:


Plan published before implementation. Slice only: treat Debug.fail's optional
stackCrawlMark as opaque metadata (`{}` rather than AnyFunction), including the
same marker parameter on reached assert and assertEqual. Nothing calls this
marker; V8 receives exactly the same function object as before. This answers
the method-signature-style refusal of stackCrawlMark || fail, where AnyFunction
has a never[] parameter list. Keep captureStackTrace and all runtime statements.

Validate all 509,014 Node tokens, failure behavior, and the complete upstream
baseline. Record the next actual compiler diagnostic before claiming it closed.

Validation: full upstream oracle 106,367 passing, zero failing, zero pending,
zero baseline differences (429.269 seconds). All 509,014 Node tokens match.
The marker probe preserves messages and explicit marker trimming; replacing
stackCrawlMark || fail with fail makes that probe exit 1. Idempotence: zero edits.
The next actual refusal is Debug.assert's assertion predicate, debug.ts:18:133.
