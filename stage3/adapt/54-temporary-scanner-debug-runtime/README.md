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
