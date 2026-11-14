# Binding views

Base: 7534c1d7575db02bb7b6512119e0685987b7e333.
TypeScript: 050880ce59e30b356b686bd3144efe24f875ebc8 (v6.0.3).
No Loom submission or upstream issue filing occurred.

## Declaration bisect and accepted form

Resolution-only reproduces the initial 1,302 export-modifier changes in
typescript.d.ts. Program-only leaves both declaration files identical. The
neighbouring convention, `/** @internal */ export interface ResolutionReadView`,
leaves both files byte-identical with both source halves applied. Each build
uses upstream Hereby dts; baseline and candidate both run npm ci. Bisect outputs
are under continuation/bisect; the initial stop remains in FIRST-STOP.md and the
original exact diff remains in typescript.d.ts.diff.

| Public output | Identical baseline/candidate SHA-256 |
| --- | --- |
| typescript.d.ts | f03adcf915df559d616fcafd75181d3423eaa0f272237482cedbd025525c9516 |
| tsserverlibrary.d.ts | cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd |

## Source proof results

| Proof | Result |
| --- | --- |
| Full resolutionCache.ts census | 221 -> 217 raw; 133 -> 129 unique; 4 gone, 0 new, 0 changed |
| program.ts census — scan only | 81 -> 79; 2 gone, 0 new, 0 changed; zero body-lowering units |
| Restoring mutant | Both original files restored byte-for-byte; all six exact refusals return; baseline/mutant deltas empty |
| Stock tsc 6.0.3, 79 roots | 0 -> 0 diagnostics with scout Node ambient roots and ledger require shim; 0 new |
| Bare-root supplementary stock check | 64 -> 64 identical ambient-related diagnostics, 0 new |
| Adamic types | 79 roots; exit 0, 0 diagnostics |
| JavaScript | All 604 source files, 1,208 emits in both comment modes, byte-identical to baseline |
| Drift, duplicate, partial-tree and mixed-line-ending controls | All rejected before writes |
| Replay | New adapter and amended 95h both write zero files; deliberate 95h partial owner rejected with zero writes |
| Normal apply | Pass; its 747 source files match the measured candidate byte-for-byte |
| Generated files | Canonical inputs/hashes match; no regeneration |
| Ledger | One type-only line citing #ty5myjt and Compiler's accept |

The complete guarded CommonJSBundle streams and all deltas are retained under
continuation/proofs. Baseline target locations are resolutionCache.ts:774,
1053, 1056 and 1471; program.ts:2796 and 2829. Program observations are **scan
only**, not full body lowering. The scan wrapper matches the one on
promise-discard-sites at 2dc0593b937f8fbaf6c32f8107a899f78c223966.

Ownership: 95-checker-axolotl-h pins createProgram whole, so the helper overloads
are an amendment there. No prior whole-owner pin claims the edited resolution
declarations. Complete text/hash pins and stricter final-file pins protect the
new work. The original ResolutionWithResolvedFileName interface is unchanged.

## Whole lane and pushed implementation

PASS: 106,366 passing, one sanctioned public-API-row failure, zero pending.
The API guard accepts exactly 273 sanctioned declarations. Normal apply exits 0;
upstream install/build exit 0, tests/oracle exit 1 solely for the sanctioned row.
Runners are all, tests is null, workers is 8; no filter, timeout override,
expectation change or sanction change was made. Full run and reporter logs are
in continuation/lane. The run takes 1,071.031 seconds.

Implementation: typescript/binding-views-1, `5ead54d316a4a6606aa61a370adea226d66f3db9`, exactly one commit on the
requested base. The lane records the base commit because implementation is
committed only after acceptance. Its tested product key matches the committed
inputs: `319b74bb713134075597eea81d9617fbb2ae2419afce9f118f244cb6339792ba`.
No source or adapter input changed after the passing run.
