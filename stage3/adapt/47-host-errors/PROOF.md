# Metadata spread candidate: stopped

Task ff5eddz. Base: origin/main
`1bb9663d4a9cbf1466a758c41c2293c9c869e6ab`. TypeScript 6.0.3:
`050880ce59e30b356b686bd3144efe24f875ebc8`. Node v24.19.0.

**Do not land this candidate.** Replacing the three spreads removes the requested
single-spread refusal, but reveals a new finding in startTracing:

```text
tracing.ts:104:46: stage 0 can't lower JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) yet
```

That violates the requested condition that nothing new appear in startTracing.
The lowerer and JSON.stringify call are unchanged. Addressing this additional
boundary would exceed the three-literal scope. The complete sharded lane and
its startTracing count were not run, and no lane count or landing is claimed.
No ledger entry was added for an unlanded candidate.

## Completed checks

| Check | Result |
| --- | --- |
| Current-main cloud setup | child exit 0 |
| Pinned API install | child exit 0 |
| Unchanged full apply | child exit 0 |
| Amended full apply | child exit 0 |
| Both compiler and harness builds | all four child exits 0 |
| Actual adamic types, all 79 recursive compiler roots | child exit 0, zero diagnostics |
| Real Node trace files | byte-identical, no timestamp normalization |
| Runtime ts export keys | all 2,256 identical |
| Emitted JavaScript audit | only the three literals, zero rename hunks |
| Restoring mutant | original 103:67 refusal returns |
| No new startTracing findings | **fails: JSON.stringify at 104:46** |

The first setup attempt began on the initial older checkout, and failed when
its older cohere pin no longer matched current main's imports. Setup was rerun
on current main, selected its correct submodule pin, and passed. Both logs are
retained. No production dependency or compiler option was changed.

## Exact source and JavaScript scope

The entire before/after src inventories have the same names. Only
src/compiler/tracing.ts differs. Replacing the three after literals with their
pinned original text recovers that entire file byte for byte. The meta local,
single timestamp() evaluation, require("fs"), and every other source byte
remain unchanged. No other adaptation pins the exact metadata literals.

All ten emitted .js files under built/ were audited. Every JavaScript diff hunk
is listed here; each is a single compact output line containing all three
metadata objects:

| File | Changed line |
| --- | ---: |
| built/local/_tsc.js | 2228 |
| built/local/run.js | 2618 |
| built/local/typescript.js | 4824 |

For each changed file, parsing the JavaScript, replacing only its three after
object nodes with the corresponding before nodes, and comparing whole-file
bytes recovers the before file exactly. There are no pure esbuild renames and
no other emitted JavaScript changes. The other seven files are byte-identical.
`meta-spread-emission.json` records inventories, lines, and both SHA-256 values.
The complete hunks are on the evidence branch in raw/all-emitted-javascript-hunks.patch.

## Node witness

Run `meta-spread-witness.cjs BEFORE_TREE AFTER_TREE OUTPUT_JSON` with Node 24.
Both complete built/local/typescript.js bundles run in separate VM contexts
with a fixed performance clock of 123.25, and the real Node fs builtin. Each
starts server-mode tracing into its own temporary directory, then stops it.
Server mode accepts an empty type catalog. Each start calls the clock once.
The actual complete trace files contain exactly the three metadata records and
are byte-identical without normalization. SHA-256:
`9a18cde59efa0695cec3d711f43d6264b8e764820343fc92ea01eb6739dab38c`.

`meta-spread-witness.json` preserves the exact trace bytes, clock counts, and
runtime export-key count and digest. The process and thread key order is
name,args,cat,ph,ts,pid,tid; the browser order is name,cat,ph,ts,pid,tid.

## Full census and restoring mutant

Each run loads the complete 79-root compiler program, selects only tracing.ts
for attempts, sets LATENT_FULL=1 and LATENT_ASSERT_NO_OUTPUT=1, and sets
LATENT_ONLY_FILE to that run's absolute adapted tracing.ts path. All three
census children exit 0 and write complete relation-report trailers. Their
headers report zero checker diagnostics and checker_rejected=false; the tool's
historical measurement label still says "measured on a checker-rejected program".
The binary is a scratch no-output overlay built by the repository's unchanged
make_overlay.py. Production compiler source is unchanged.

These are all startTracing NotYet/Refused findings, grouped by source address;
Boundary records are retained in the raw streams:

| Address | Before | After | Restoring mutant |
| --- | --- | --- | --- |
| 64:22 | NotYet: a call returning any | same | same |
| 72:9 | NotYet: assigning a field of a value | same | same |
| 96:19, refusal_scan | NotYet: a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports | same | same |
| 96:19, lowering | NotYet: reading tracingEnabled | same | same |
| 103:67 | Refused: a spread after the first field (adamic/single-spread) | absent | restored |
| 104:46 | absent | NotYet: JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) | absent |

The before and restoring-mutant scans report only the first spread in the
statement, at 103:67. They do not separately report the original columns 125
and 171. The after source contains none of the three spread nodes, and its
startTracing findings contain no single-spread refusal. The mutant restores
all three original literals in a private source copy, recovers the before
tracing.ts bytes exactly, and reproduces the before startTracing findings
exactly after path normalization. `meta-spread-census.json` records every
NotYet/Refused finding above and the failing no-new-findings comparison.

A separate attempt to replay all of 47 on the full finished tree fails at its
existing sys.ts helper-import guard, before reaching startTracing. This is
recorded separately from the census stop; it is not a successful replay claim.
Untouched origin/main adaptation 47 reproduces the same helper-import failure
(child exit 1). The focused contract script and both failure logs are retained
only on the evidence branch. No unrelated import guard was weakened.

## Evidence and disposition

Raw logs, child-exit records, streams, source and emitted patches, and scratch
measurement scripts are on
[typescript/meta-spread-evidence](https://github.com/system-inc/adamic/tree/typescript/meta-spread-evidence/raw).
The candidate and this stop report are saved to typescript/meta-spread-checkpoint.
There is no final commit on typescript/meta-spread and no main push.
