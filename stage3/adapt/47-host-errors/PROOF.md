# startTracing metadata fields and catalog reset, ff5eddz

The three metadata literals build from explicit reads of the fresh local meta.
They retain the same keys, insertion order and values. The meta declaration,
its single timestamp() call and the require("fs") producer are unchanged.
The additional routed edit is typeCatalog.length = 0 to typeCatalog.splice(0),
under rulings.md:306-307. All bytes outside these four source nodes are unchanged. This is a JavaScript-changing adaptation, with actual
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
no other finding is added. The 72:9 length-write finding and its rollback
boundary disappear. A length-only restoring mutant brings 72:9 back; a full
four-edit restoring mutant exactly recovers before.

## Pins and source scope

The final validation base is origin/main at
`16429a30452d027809f0ea85bc395479423e493e`. The initial measurement base was
`1bb9663d4a9cbf1466a758c41c2293c9c869e6ab`; all adaptation inputs are identical
between those main commits. TypeScript is pinned to 6.0.3,
`050880ce59e30b356b686bd3144efe24f875ebc8`. Node is v24.19.0.

47 previously checked the builtin loader producer, without a whole-function
pin. meta-spread-pins.json now admits exactly the incoming function and the
completed function: the latter includes 47's existing errorMessage rewrite
the three explicit-read literals and the splice reset. CRLF/LF normalization alone is allowed.
The existing fs producer check is unchanged. No other adaptation pins these
exact literals. No later node:fs import amendment is included.

The pin witness accepts both exact states, proves no-write finished replay,
and refuses seven other states before writing: the two mixed catch/literal
forms, mixed metadata literals, meta-value drift, loader-producer drift, and
a wrong splice index, and restoring the old length write as an intermediate state. It uses 47's own immediate sys, tracing and helper
snapshots so later import merging and helper edits do not bypass 47's guards.
The earlier full-tree replay attempt failed at an unrelated sys import guard;
untouched origin/main 47 fails the same way. Both attempts remain in raw
historical evidence, and no unrelated guard was weakened.

All before/after src paths match. Only src/compiler/tracing.ts differs, and
reversing the three literal replacements and the reset recovers its before
bytes exactly. The fresh combined apply and audited build have all 747 source
files byte-identical.
The completed source's only function diff is the three literal nodes and
catalog reset.

## Completed checks

| Check | Result |
| --- | --- |
| Current-main setup and pinned API install | child exits 0 |
| Unchanged and amended full applies | child exits 0 |
| Final two-state-pin full apply | child exit 0; all 747 src files byte-identical to the audited build inputs |
| Both compiler and harness builds | all four child exits 0 |
| Actual current-main adamic types, all 79 compiler roots | exit 0, zero diagnostics |
| Node trace bytes and runtime ts export keys | identical |
| Every emitted JavaScript hunk | only the three literals and reset, zero renames |
| Before, after and mutant full censuses | all child exits 0 |
| Whole-file census comparison | only the accepted next stop is newly visible |
| Combined sharded lane and catalog-length counts | running; the metadata-only lane is retained as historical evidence |

The initial setup attempt selected the old checkout's cohere pin and failed
after the branch changed. Setup was rerun with current main's proper submodule
and passed; setup also passed after main advanced. All logs are retained.
No dependency, compiler option, expected lane count, or lane script is changed.

## Node and emission witnesses

meta-spread-witness.cjs loads both complete built/local/typescript.js bundles
in separate VM contexts with the real fs builtin and the same fixed performance
clock, 123.25. Each starts server-mode tracing in a fresh temporary directory,
then stops it; server mode accepts an empty type catalog. The clock fires once
per start. The witness exposes only the actual private catalog at its
declaration in each complete loaded bundle, seeds it with zero or three object
entries, then invokes the actual startTracing. After each start the same array
has length zero, no entries, and only its own length key. Both before and after
observe these states. This witnesses clearing, while acknowledging splice
allocates a discarded removed-elements array unlike the length write. The actual complete trace files are byte-identical, without timestamp
normalization. SHA-256:
`9a18cde59efa0695cec3d711f43d6264b8e764820343fc92ea01eb6739dab38c`.
All 2,256 runtime ts export keys match. The exact trace and key digest are in
meta-spread-witness.json.

The process/thread record key order is name,args,cat,ph,ts,pid,tid. The browser
record order is name,cat,ph,ts,pid,tid: its explicit timeline category retains
the position introduced by meta's cat. The meta local itself has the ordered
own enumerable keys cat,ph,ts,pid,tid.

All ten emitted .js files under built/ were audited. These are every emitted
JavaScript diff hunk: one reset line and one compact metadata line per file:

| File | Reset line | Metadata line |
| --- | ---: | ---: |
| built/local/_tsc.js | 2208 | 2228 |
| built/local/run.js | 2598 | 2618 |
| built/local/typescript.js | 4804 | 4824 |

For each changed file, parsing JavaScript and replacing only its three after
object nodes and its reset statement with the corresponding before nodes
recovers whole-file before bytes exactly. No esbuild rename or other JavaScript change occurs. The other
seven files are byte-identical. meta-spread-emission.json records both hashes;
the complete hunks are in raw/catalog-all-emitted-javascript-hunks.patch on the evidence
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
| 72:9, lowering | NotYet: assigning a field of a value | absent | restored |
| 96:19, refusal_scan | NotYet: a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports | same | same |
| 96:19, lowering | NotYet: reading tracingEnabled | same | same |
| 103:67, lowering | Refused: a spread after the first field (adamic/single-spread) | absent | restored |
| 104:46, lowering | absent | NotYet: JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) | absent |

The before and restoring-mutant scans report the first spread at 103:67,
not separate findings for original columns 125 and 171. The after source
contains none of the three spreads and has no single-spread finding in
startTracing. The full private mutant restores all three originals and the length write,
recovers before tracing.ts bytes, and reproduces all before findings after
path normalization. A separate length-only mutant retains explicit metadata
and restores 72:9 while leaving the accepted 104:46 next stop visible.
meta-spread-census.json records the exact comparison over all tracing.ts sites,
the sole permitted addition, the removed refusal, and the matching 194:67 stop.

## Lane runtime count

The combined lane is running afresh through the gate product path. Its private
bundles record each entry, actual catalog length immediately before and after
the reset, and actual meta keys and values after construction. The populated
positive control records lengths 3 before reset and 0 after reset, plus meta's
own ordered keys cat,ph,ts,pid,tid. Final full-lane results are required before
publishing the combined amendment. The earlier metadata-only lane and its
398 passing shards remain historical evidence, not a substitute for this run.

## Evidence

Raw logs, child exit records, streams, source and JavaScript patches, controls,
private instrumentation, and lane observations are on
[typescript/meta-spread-evidence](https://github.com/system-inc/adamic/tree/typescript/meta-spread-evidence/raw).
The original stopped attempt is retained there as historical evidence. Only
the user's specified next JSON.stringify stop is now allowed. Checkpoints go
to typescript/meta-spread-checkpoint; main is never pushed or force-pushed.
