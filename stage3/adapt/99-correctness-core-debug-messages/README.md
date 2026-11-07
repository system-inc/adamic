# debug.ts lazy assertion failure messages

Task b5w3ycg, Adamic's Nov 7 19:16Z road (b), ruled JavaScript-changing.
Base: origin/main d15620712c459baf72b3d359c1829df8d0808fda.
TypeScript 6.0.3: 050880ce59e30b356b686bd3144efe24f875ebc8.

This adaptation omits the test function name from four lazy failure messages.
It leaves getFunctionName itself unchanged. The chosen messages are:

| Owner | Adapted line | Upstream line | Message |
| --- | --- | --- | --- |
| assertEachNode | 267 | 288 | Node array did not pass test. |
| assertNode | 280 | 301 | Node ${formatSyntaxKind(node?.kind)} did not pass test. |
| assertNotNode | 293 | 314 | Node ${formatSyntaxKind(node!.kind)} should not have passed test. |
| assertOptionalNode | 307 | 328 | Node ${formatSyntaxKind(node?.kind)} did not pass test. |

All four rows are JavaScript-changing. Their callbacks remain lazy. Each row
loses only the space, quoted function name and call to getFunctionName; the
final period stays. The node kind interpolation stays where it was. The
assertion conditions and the call into assert are byte-identical. The
implementations of assert, fail, shouldAssertFunction and getFunctionName
are pinned unchanged, including new Error and throw e.

## Ownership

The pre-edit inspection is recorded in ownership-audit.json. No active
whole-owner or exact-message pin covers these four owners. 99-cli-debug-assertion-level
pins the helper callers, 76 pins the finite key set and cache, 40 pins other
debug expressions, and 45 pins getFunctionName's capture. 99t's three whole
owners are different. 99n marks these four proposals dropped; its only active
owner is formatControlFlowGraph.hasNode. Historical whole-file evidence is
not an active input pin. Therefore this is a new adaptation, not an amendment.

## Adapter

The adapter finds the four implementations in namespace Debug by name. Each
whole owner must be exactly the before or after pin, and all four must have
the same state. Partial restoration and unreviewed owner changes are refused
before writing. A second run writes nothing. Stock 6.0.3 emission, with and
without comments, must differ only by the four reviewed message expressions.
The two node?.kind messages have identical text, so the emission proof checks
their multiplicity together. Non-null assertions are erased by the stock
emitter before the JavaScript comparison. Declaration emission must match
with and without stripInternal.

## Verification

The full applies, whole-lane measurement, complete-tree witnesses, paired
replay and census proofs are in progress. No unfinished check is claimed.
The finished proof and raw archive index will be recorded here before push.
