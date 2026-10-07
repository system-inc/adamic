Temporary: comes out when native debugger and V8 Error stack capture support land

Slice-only plan, published before implementation. Require slice.json. Restrict
edits to the reached Debug.fail member in debug.ts. Remove its debugger statement
(which has no effect without an attached debugger). Probe V8-specific stack
capture separately; do not remove or rewrite assertion control flow, throw,
error message construction, or the assertion's type contract without recording
the next blocker and revising the plan first.

Validate stock Node token equality and the full stage 3 baseline before claiming
completion. Explicitly report that attached-debugger behavior is outside this
scanner token proof. Each next refusal remains in BLOCKERS.md.

## Revised validation

Full unfiltered upstream baseline passes: 106,367 tests, zero failures, zero
pending, zero baseline differences. Exact slice edits were mapped by recorded
source spans onto complete upstream declarations for validation only; the
adaptation scripts themselves still reject full-tree input. Install, build and
tests all exit 0. Combined wall time 259.105 seconds, nproc 5, four workers.
Node scanner output matches all 509,014 full-tree tokens. Idempotence passes.
The earlier rejected implementations are historical probes, not this result.
See scanner/evidence/member-baseline-revised-report.json and its test log.
