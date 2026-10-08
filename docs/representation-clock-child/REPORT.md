# Child representation clock ruling

Delivery baseline: origin/area/compiler ca59a31af5d649eedb39d9f060c3498727fa6959.
Only this contribution's commits are carried on that baseline. The brief's compiler
3cead3fdfeaf8dca6920d748f016a0b0ef1823f2 is a different lineage; replay tooling
from it is used only as scratch measurement infrastructure.

Assumption: use the brief's explicit census-context ruling option when an
unresolved structural constraint cannot prove a unique ABI. Cancel the reserved
`NotYet: a value of type Child` kind as a census-only unresolved-specialization
request, retaining the production refusal. This does not claim that the generic
census body now lowers or that an unchecked entry-root program compiles.

`Child extends NodeView` does not uniquely select an object header. A checked
`CallableNode extends NodeView` with a call signature is assignable to NodeView,
but the two checked types have Object and Closure representations respectively.
The direct checked-type test probes Child without a substitution and requires the
named NotYet, then supplies Object and Closure substitutions and requires each
exact representation. No constraint fallback, identity rewrite, new IR kind,
backend change, or separate runtime helper is introduced.

The exact minimal fixture is registered in its own oracle test file. Its named
`clock-child-ignore-argument` mutant replaces the render specialization body with
`return "mutant"`; the test requires a successful, sanitizer-clean, leak-clean
native run differing from Node only in stdout. A scratch-only admission mutant
returns Object for unsubstituted parameters; the checked-type test must fail.

Source reconstruction uses the exact stage3 archive at
9d534d3a31814f1a192a528e701f6c2ea7c910bc. All 81 source byte lengths and SHA-256
hashes match the original tsc source manifest. The guarded replay disables
ordinary loading and IR output. Evidence below records its result on the
current delivery baseline, rather than transferring the brief's observation.

## Results

- The exact replay exits 0 and still reports the Child stop at emitter.ts:4754:19,
  in emitNodeListItems (unit emitter.ts:4729:5). `replay.json.gz` retains the entire
  guarded replay stdout and `replay.log` its stderr. It is a measurement of a
  checker-rejected entry-root program, without a concrete Child substitution.
  **Cancel this clock kind as a census echo of that unresolved context; keep the
  production refusal.** No representation stop is claimed retired.
- `TestRepresentationClockChildUnresolvedConstraint` passes. The scratch Go
  overlay mutant inserts `if !isKnown { return ir.Object, true }` before
  `return substituted, isKnown` in the type-parameter branch. Its targeted test
  exits 1 with `unresolved Child admitted as 4` (Object), retained in
  `admission-mutant.log`. The production helper was never edited.
- The registered fixture passes `TestNativeAgreesWithNode` for JavaScript,
  sanitized native, release native, and leak checks. Node stdout is `1:leaf\n`.
- `clock-child-ignore-argument` passes its mutant-killing test: exit 0, empty
  stderr, clean sanitizers and leak check, stdout `mutant\n`, and exactly the
  expected stdout disagreement with Node.

Focused commands (all stdout/stderr written directly to logs):

```sh
go test ./internal/lower -run TestRepresentationClockChild -v
go test ./internal/oracle -run 'TestRepresentationClockChildMutant|TestNativeAgreesWithNode/internal/oracle/testdata/representation_clock_child.a' -v
go test -overlay=/tmp/clock-child-admission-mutant/overlay.json ./internal/lower -run TestRepresentationClockChildUnresolvedConstraint -v
go run ./stage3/census/latent/replay -project /tmp/notyet-representations-adapted/src/tsc/tsc.ts -where /tmp/notyet-representations-adapted/src/compiler/emitter.ts:4754:19 -kind NotYet -reason 'a value of type Child'
```

The required counts refresh passes:
`go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts`.
Only the new fixture row is added to counts.md (retained output in counts.log).
No new separate runtime helper exists. Ownership history was checked with
`git log --remotes=origin/codex/notyet-* -- internal/lower/expression.go` before
work; no other worker's lowering function, compiler source, or other clock kind
was changed.
