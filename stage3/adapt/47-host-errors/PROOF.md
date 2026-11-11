# startTracing metadata fields, ff5eddz

The three metadata literals build from explicit reads of the fresh local meta.
They retain the same keys, insertion order and values. The meta declaration,
its single timestamp() call, the require("fs") producer, and all other source
bytes are unchanged. This is a JavaScript-changing adaptation, with actual
Node byte and emitted-output witnesses below.

The only newly visible census finding is the explicitly accepted next stop:

```text
tracing.ts:104:46: stage 0 can't lower JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) yet
```

Before the edit, lowering stops the statement at its first spread. Removing
the spreads exposes the unchanged map callback's JSON.stringify boundary.
This belongs to Compiler's #d9ep7de head 1 and is the same reason already
reported at tracing.ts:194:67. The user's follow-up accepts only this finding.
A whole-file comparison checks every tracing.ts finding, including boundaries:
no other finding is added, and the restoring mutant exactly recovers before.

## Pins and source scope

The final validation base is origin/main at
`16429a30452d027809f0ea85bc395479423e493e`. The initial measurement base was
`1bb9663d4a9cbf1466a758c41c2293c9c869e6ab`; all adaptation inputs are identical
between those main commits. TypeScript is pinned to 6.0.3,
`050880ce59e30b356b686bd3144efe24f875ebc8`. Node is v24.19.0.

47 previously checked the builtin loader producer, without a whole-function
pin. meta-spread-pins.json now admits exactly the incoming function and the
completed function: the latter includes 47's existing errorMessage rewrite
and the three explicit-read literals. CRLF/LF normalization alone is allowed.
The existing fs producer check is unchanged. No other adaptation pins these
exact literals. No later node:fs import amendment is included.

The pin witness accepts both exact states, proves no-write finished replay,
and refuses six other states before writing: the two mixed catch/literal
forms, mixed metadata literals, meta-value drift, loader-producer drift, and
unrelated function drift. It uses 47's own immediate sys, tracing and helper
snapshots so later import merging and helper edits do not bypass 47's guards.
The earlier full-tree replay attempt failed at an unrelated sys import guard;
untouched origin/main 47 fails the same way. Both attempts remain in raw
historical evidence, and no unrelated guard was weakened.

All before/after src paths match. Only src/compiler/tracing.ts differs, and
reversing the three literal replacements recovers its before bytes exactly.
The completed source's only function diff is those three literal nodes.

## Completed checks

| Check | Result |
| --- | --- |
| Current-main setup and pinned API install | child exits 0 |
| Unchanged and amended full applies | child exits 0 |
| Final two-state-pin full apply | child exit 0; all 747 src files byte-identical to the audited build inputs |
| Both compiler and harness builds | all four child exits 0 |
| Actual current-main adamic types, all 79 compiler roots | exit 0, zero diagnostics |
| Node trace bytes and runtime ts export keys | identical |
| Every emitted JavaScript hunk | only the three literals, zero renames |
| Before, after and mutant full censuses | all child exits 0 |
| Whole-file census comparison | only the accepted next stop is newly visible |
| Complete sharded lane and start/meta counts | 398 shards pass; 0 starts, 0 meta builds; positive control 1 start and 1 meta build |

The initial setup attempt selected the old checkout's cohere pin and failed
after the branch changed. Setup was rerun with current main's proper submodule
and passed; setup also passed after main advanced. All logs are retained.
No dependency, compiler option, expected lane count, or lane script is changed.

## Node and emission witnesses

meta-spread-witness.cjs loads both complete built/local/typescript.js bundles
in separate VM contexts with the real fs builtin and the same fixed performance
clock, 123.25. Each starts server-mode tracing in a fresh temporary directory,
then stops it; server mode accepts an empty type catalog. The clock fires once
per start. The actual complete trace files are byte-identical, without timestamp
normalization. SHA-256:
`9a18cde59efa0695cec3d711f43d6264b8e764820343fc92ea01eb6739dab38c`.
All 2,256 runtime ts export keys match. The exact trace and key digest are in
meta-spread-witness.json.

The process/thread record key order is name,args,cat,ph,ts,pid,tid. The browser
record order is name,cat,ph,ts,pid,tid: its explicit timeline category retains
the position introduced by meta's cat. The meta local itself has the ordered
own enumerable keys cat,ph,ts,pid,tid.

All ten emitted .js files under built/ were audited. These are every emitted
JavaScript diff hunk, each a single compact line containing all three objects:

| File | Changed line |
| --- | ---: |
| built/local/_tsc.js | 2228 |
| built/local/run.js | 2618 |
| built/local/typescript.js | 4824 |

For each changed file, parsing JavaScript and replacing only its three after
object nodes with the corresponding before nodes recovers whole-file before
bytes exactly. No esbuild rename or other JavaScript change occurs. The other
seven files are byte-identical. meta-spread-emission.json records both hashes;
the complete hunks are in raw/all-emitted-javascript-hunks.patch on the evidence
branch. The audited bundles are reused only after the final apply's complete
source inventory is proved identical.

## Full census and restoring mutant

Each run loads all 79 compiler roots and selects tracing.ts for attempts with
LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, and LATENT_ONLY_FILE set to that run's
absolute adapted tracing.ts path. The current-main scratch no-output binary
comes from the unchanged make_overlay.py. All streams have complete trailers
and zero checker diagnostics. Their historical measurement label still says
"measured on a checker-rejected program", while checker_rejected is false.

These are all startTracing NotYet/Refused findings; Boundary records are kept
and compared in the complete streams:

| Address and phase | Before | After | Restoring mutant |
| --- | --- | --- | --- |
| 64:22, lowering | NotYet: a call returning any | same | same |
| 72:9, lowering | NotYet: assigning a field of a value | same | same |
| 96:19, refusal_scan | NotYet: a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports | same | same |
| 96:19, lowering | NotYet: reading tracingEnabled | same | same |
| 103:67, lowering | Refused: a spread after the first field (adamic/single-spread) | absent | restored |
| 104:46, lowering | absent | NotYet: JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) | absent |

The before and restoring-mutant scans report the first spread at 103:67,
not separate findings for original columns 125 and 171. The after source
contains none of the three spreads and has no single-spread finding in
startTracing. The private mutant restores all three originals, recovers before
tracing.ts bytes, and reproduces all before findings after path normalization.
meta-spread-census.json records the exact comparison over all tracing.ts sites,
the sole permitted addition, the removed refusal, and the matching 194:67 stop.

## Lane runtime count

The unchanged measure_shards.py runs all 398 manifest shards with concurrency 2
through STAGE3_ADAPTED_TREE, the gate's product path. Only private counting
copies of built/local/_tsc.js, run.js and typescript.js are instrumented.
Each start records entry, then records Object.keys(meta), Reflect.ownKeys(meta),
the enumerable own-key order, and actual field values immediately after the
original meta declaration. Each bundle also records its load.

The complete lane passes: 106,366 passing observations, the one unchanged
sanctioned Public APIs failure, and zero pending, for 106,367 observations.
Every shard has its expected count, names digest and verdict. Wall time is
909.660 seconds; the maximum shard wall is 9.083 seconds, below 30 seconds.
All 398 private harness copies hash identically, and all 398 shards produce
counter-load records from 398 instrumented processes.

**startTracing entries: 0. Actual in-lane meta builds: 0.** Consequently there
are no in-lane meta key sets to enumerate and no differing key sets. The
positive control loads the same instrumented compiler, starts and stops a real
trace, and records exactly one entry and one meta build. Its actual Object.keys,
Reflect.ownKeys and enumerable own-key lists are all exactly:

```text
cat, ph, ts, pid, tid
```

Its values have cat="__metadata", ph="M", finite ts, pid=1 and tid=1. The
separate fixed-clock Node witness proves the single timestamp call and exact
trace bytes. meta-spread-lane.json preserves the reconciled lane result and
positive control. Raw per-shard reports, observations, logs and diffs are in
raw/lane-evidence.tar.gz; raw/lane-events.jsonl retains every bundle-load event.

The first private copy mistakenly excluded nested node_modules fixtures and
was stopped with child exit -15 after a project-fixture failure. It is archived
separately, never counted as the final lane. The corrected copy was compared
against all 85,763 inventory entries: all 70 nested module-fixture entries are
retained, and only the three instrumented bundles differ. The entire final
lane restarted from scratch. No production lane code, fixture or expectation
was changed.

## Evidence

Raw logs, child exit records, streams, source and JavaScript patches, controls,
private instrumentation, and lane observations are on
[typescript/meta-spread-evidence](https://github.com/system-inc/adamic/tree/typescript/meta-spread-evidence/raw).
The original stopped attempt is retained there as historical evidence. Only
the user's specified next JSON.stringify stop is now allowed. Checkpoints go
to typescript/meta-spread-checkpoint; main is never pushed or force-pushed.
