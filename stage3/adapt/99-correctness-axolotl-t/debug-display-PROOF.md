> Nov 10 Change 2: current behavior and complete evidence are in
> [debug-display-2-README.md](debug-display-2-README.md). The graph module
> home also amends 99n; earlier Change 1 records below remain historical.

# Completed debugger-display proof

Ready to land. Final main: 7f2c852dffe31163e9699f0c91a158596757b888.
The full amended apply validated main 95afe50a6e94dbff360a399458b5f45f43cfccad.
The entire stage3 Git tree is byte-identical between those main tips; the final
main update changes only the native runtime and its validation. See
debug-display-final-main-input-identity.json.

| Gate | Result |
| --- | --- |
| Owner audit | Amendment in 99t because enableDebugInfo has an active whole-owner pin |
| Complete baseline lane | 398 accepted units; guarded counts flow / array / enable = 0 / 0 / 0 |
| Forced-debug control | 6801 / 45429 / 1; ordinary control 0 / 0 / 0; both exit 0 |
| Paired Node CLI witness | 200 cases, raw stdout/stderr/exit byte-identical with and without inspector options |
| Counted witness replays | Identical output; 1312455 / 9088154 / 200 executions |
| Emitted hunk classification | 3204 class (a), 2 class (b), 0 class (c) |
| Runtime exports | Sorted module keys 2258, separate ts/CommonJS keys 2256; byte-identical |
| Normal candidate build | Exit 0; both public declaration files byte-identical |
| Full amended apply | Exit 0 |
| Applied source equivalence | All 746 source files byte-identical to the normal, classified candidate |
| Final adapter replay | Exit 0; zero files and edits |
| Complete final lane | All 398 ordinary guards pass; 106366 passes, one sanctioned API failure, zero skips; controller exit 0 |
| FlowGraphNode | Entire formatControlFlowGraph owner byte-identical to baseline |

The restoring mutant changes exactly one line and reproduces Refused at
debug.ts:525:37: a cast the runtime can't check. Without it, binder.ts and
core.ts both first stop at debug.ts:1021:29, where FlowGraphNode.edges can
write FlowGraphEdge[] where the literal inferred never[]. That code is
untouched. debug-display-census.json retains the ordinary per-file results.

## Mechanical bundle fallout

Only terminal decimal suffixes on AST Identifier nodes are normalized;
strings, comments, numeric literals, punctuation and whitespace are untouched.
Every class-(a) hunk compares byte-identically after normalization. Every
residual byte of all three bundles is equal after normalization and deletion
of exactly the three inventoried functions. The complete normal emitted patch
is in debug-display-all-emitted-javascript-hunks.patch.gz; each residual hunk's
class and comparison hashes are in debug-display-bundle-classification.json.gz.

A removed declaration is class (b) when every reference is inside one of the
five owners or inside another removed declaration that is itself class (b),
closed transitively. The classifier computes the least fixed point and retains
the edges and chains. Unanchored cycles and external references fail closed;
six controls execute the actual classifier block.

Hunk 55 records getJSDocModifierFlagsNoCache ->
getEffectiveModifierFlagsNoCache -> enableDebugInfo. Its original reference
is _tsc.js:17073:51; the parent helper's reference is _tsc.js:1791:42.
isArrayTypeNode's only reference is _tsc.js:1775:1231 inside enableDebugInfo.
There are no other removed declarations. The earlier raw-byte-only comparison
and direct-reference stops are historical and superseded by this authorized
transitive rule; their source adaptation and pins have not changed.

## Evidence

- debug-display-final-apply.json and its compressed log record the successful full apply, all-source equality, unchanged graph owner and normal JavaScript hashes.
- debug-display-final-lane.json.gz and debug-display-final-lane-reports.jsonl.gz retain every ordinary accepted shard and its own child exit. The run has no observer instrumentation.
- debug-display-runtime-export-keys.json.gz retains both complete sorted runtime key arrays and hashes from actual Node import and require loads.
- debug-display-contract.json, debug-display-final-replay.json and debug-display-transitive-controls.json retain the strict pin, idempotence and closure controls.
- debug-display-count.json.gz and debug-display-witness.json.gz retain the complete baseline count and the paired CLI witness. Preliminary metadata/ignoreConfig attempts and their passing corrections remain recorded there.

The final lane uses the normal candidate build. The successful full apply's
entire source is byte-identical to that candidate, and the lane's ten normal
JavaScript files match the classified build hashes. No GraphNode repair,
export-set normalization, output normalization, changed lane expectation or
native lowering overlay is part of this adaptation.
