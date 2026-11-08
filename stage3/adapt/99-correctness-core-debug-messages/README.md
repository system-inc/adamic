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

All required checks passed. [PROOF.md](PROOF.md) gives the measured results,
all emitted JavaScript hunks, the remaining getFunctionName handoff, controller
history and reproduction. [PROOF.json](PROOF.json) is the machine-readable gate
record. [ARCHIVES.json](ARCHIVES.json) lists the raw archives and hashes.

| Required check | Result |
| --- | --- |
| Whole baseline lane oracle | 398 accepted units, 106366 passes, 1 expected failure, 0 skips |
| Callback executions, each counted separately | assertEachNode 0; assertNode 0; assertNotNode 0; assertOptionalNode 0 |
| Forced-failure counter controls | Each counter is 1 before and 1 after |
| Node throw/message witness | All 32 paired cases pass; same Error class, conditions and throw point |
| Both full applies and normal builds | Exit 0 |
| Both adamic types runs | All 79 roots together, exit 0, zero diagnostics |
| Paired replay | 96 adapters; 95 identical; only this adapter differs, 4 edits before and 0 after |
| Every normal emitted JavaScript file | 10 checked; only _tsc.js, typescript.js and run.js differ, each only at the four messages |
| Public declarations | typescript.d.ts and tsserverlibrary.d.ts byte-identical |
| debug.ts full census | 105 units each; exactly the four class-4 sites removed, no sites added |
| Restoring mutant | The original refusal returns at 267:72 |
| getFunctionName | Source and emitted body unchanged; its own census unit remains for Adamic |

The lane count is a complete baseline measurement, not a second candidate
whole-oracle run. The candidate has the completed-tree Node witnesses, positive
emitted counter controls, both declaration checks, full build/type checks,
paired replay and full census proof above. The oracle's existing expected
public-API acknowledgement failure is retained. No lane expectation changed.

## Stacked move: condition-4 guard follows the helper

The authorized 99-correctness-core-getfunctionname-move applies after road (b).
Road (b) itself still leaves getFunctionName unchanged. On a finished moved
tree, its condition-4 guard follows the function to deprecate.ts, pinning the
same body byte for byte. A drifted or half-applied local body, duplicate homes
or a missing helper still refuses on replay. The original Debug owner remains
accepted unchanged before the move. getfunctionname-home.json records both
exact declarations and the identical body; pins.json retains the original
whole-owner guard and adds the unchanged body hash for the moved home.

This amendment changes recognition only. The four assertion message edits and
all other guards remain as proved above. The move directory records the new
paired replays, drift controls and normal-build proofs. Historical stop logs
are preserved there as history, not current implementation status.
