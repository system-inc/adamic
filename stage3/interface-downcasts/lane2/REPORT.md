Built standalone array-contract and callable-admission hooks, JavaScript read/call helpers, and seven source witnesses; compiler dispatch integration is incomplete.
Base: 4dbf3b6a; this report and its implementation are committed together on codex/views-arrays-callables.
Commands: touched-package tests, vet, focused Node helper tests, and the existing checked-view oracle pass; source witnesses still fail compiler admission.
Mutants: array kind, selected element validation, and unproven signature acceptance each fail a semantic assertion, with no build-warning kill.
Not covered: native array metadata/emission, shared dispatch/IR wiring, implementation certificates, tuples, iteration/callback/mutator integration, and whole-corpus lowering.

## Observations

The lane plan at `4dbf3b6a` reserves shared IR, dispatch, readiness, runtime headers,
object runtime and counts for lane 1. Its proposed `ir.ViewContractID` and
`internal/ir/views.go` do not exist on that tip. This branch adds no edits in those
territories. Helpers are intentionally not enabled by the current compiler.

`viewArrayContract(node, target, buildElement)` identifies mutable/readonly arrays
and delegates their element type to the shared recursive builder exactly once.
Nested builder errors propagate. Tuples return NotYet because positional,
optional and rest contracts need a descriptor that this hook cannot yet express.
The callback currently returns error only; lane 1 can capture its interned ID in
an adapter without assigning a second ID namespace here.

JavaScript `emitViewArrayFieldRead` uses shared presence/readiness state, then
checks array kind without scanning. `emitViewArrayRead` calls the supplied
contract checker for the selected value on every read. That checker returns the
value or corresponding transitive view. No proxy, copy, or cached element proof
hides writes by another alias. Missing indexes pass undefined to the contract
checker so an optional element contract can preserve undefined semantics.

Callable reads check readiness and function kind, including AdamicClosure's
backend representation. Method lookup preserves the backend's explicit receiver
convention; field closures use adamicCall. `viewCallableCall` refuses an
unproven signature, naming the member and explaining that runtime checking cannot
establish it. Its boolean evidence input must be supplied by the implementation
certificate, not an asserted type or typeof-function test. These helpers do not
produce that certificate. A class method test verifies calling convention only,
not class-tag proof erasure or free compiler emission.

Tests execute the helper runtime on Node. Positive selected reads, object element
identity, closures and methods also have plain Node controls. Negative cases pin
exit 70 and stderr. They cover lazy malformed second elements, readonly access
with a mutable alias, sparse elements, optional absent indexes, nested malformed
payload, single evaluation, missing/uninitialized array fields and callables, and
non-function/non-array kinds. They are helper tests, not end-to-end new compiler
oracles. No count rows are supplied because the witnesses do not lower yet.

## Integration handoff

1. In the common contract builder, route arrays to viewArrayContract and capture
   the recursive element contract ID. Add positional tuple descriptors separately.
2. Carry element contracts and source expression text through field reads, aliases,
   parameters, returns, captures, element reads, iteration and callback dispatch.
   Include viewArraysRuntime/viewCallablesRuntime with fieldReadinessRuntime.
   Wire the emitViewArrayFieldRead/Read and emitViewCallableRead/Call helpers.
3. Native arrays currently have untagged adamic_value elements and a boolean
   references flag. That cannot distinguish number from boolean, or certify a
   per-element union discriminant. Shared semantic metadata must describe element
   storage and conversion before lane 2 can implement a sound native check.
   A references-only guard must not be advertised as an element type check.
   Metadata must survive stores, mutation, copying and alias effects.
4. Supply class/closure/intrinsic implementation certificates. Do not pass true
   to viewCallableCall based on a function tag or asserted signature. Proven
   class methods can keep existing direct dispatch; opaque callbacks stay refused.
5. Promote the seven .a witnesses to lane-owned end-to-end oracle tests after
   admission is wired; shared count regeneration remains lane 1's responsibility.

The required merge of origin/main (71d7e49) was attempted and aborted. It conflicts
in internal/native/emit_statements.go and internal/oracle/counts.md, both reserved
for lane 1. This branch is not claimed to have landed on current main.

## Site share

The unchanged per-site audit has 2,936 entries: 1,758 tagged and 1,178 untagged,
all undecidable. The plan's primary lane 2 partition is 25 tagged + 1,046 untagged
= 1,071. Untagged dependencies include 1,010 arrays and 71 callable contracts,
with 35 overlapping. No shared dispatch behavior changes, so this checkpoint
unlocks **zero additional sites**. Before and after successful-lowering totals
were not remeasured over the TypeScript compiler corpus. Contract eligibility is
not substituted for that measurement; certified conformance remains zero in the
unchanged audit. No erasure or speedup is claimed.

## Validation

All commands source /workspace/adamic-tools/env.sh and set
GOPROXY='https://proxy.golang.org|direct'. nproc reports 5, cpu.max 400000/100000.
Final setup reports Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s,
build cache warm 253s, done 253s. Go 1.27.1, Node 24.19.0, clang 20.1.8.

The initial setup overlapped my checkout and failed cache warming: it reported
missing internal/fuzz/overrides.go, undefined l.interfaceCast/l.view, missing
CheckedFields/Uninitialized IR fields and readiness helpers, then too many errors.
I reran setup on the stable branch; it completed. This was an invalid setup run,
not evidence of a stable branch build defect.

- go test ./internal/lower ./internal/javascript -count=1 -timeout 30m:
  lower 26.044s, javascript 0.993s, both pass.
- go vet ./internal/lower ./internal/javascript: exit 0, no output.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
  TestDefaultTaggedSourceViews -count=1 -timeout 30m: pass, 23.416s.
- Focused helper tests and mutants: see committed logs. Mutant array-kind fails
  TestViewArraysNode/kind; element fails TestViewArraysNode/second; signature fails
  TestViewCallableSignature. Each mutation is restored in a finally block.
- Source Node runs all seven witnesses, exit 0: array field 2, selected element 7,
  object element ok, class method 10, function field 8, non-array undefined,
  opaque signature 7. Compiler probes reach array NotYet, class-cast refusal,
  and callable-field refusal. These are unresolved integration blockers.

No full gate, native lane 2 test, updated audit lowering census, or benchmark ran.
Logs are in this directory's logs/ subtree. The reproducible mutant runner writes
its individual logs under /tmp/adamic-view-lane2-mutants/.
