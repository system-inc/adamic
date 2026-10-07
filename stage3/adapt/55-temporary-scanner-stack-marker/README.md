Temporary: comes out when function widening lands

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
