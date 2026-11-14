# Recordreplay amendment and paired censuses

| Pair | Every moved existing unit first stop (all units) | New process findings (all-unit first-error census) | 79-root tables |
|---|---|---|---|
| A: main 9c304036 → CLI 7ccad5b5 | None | None | Not run by decision |
| B: original main → amended CLI | None | None | Not run by decision |

The revised requested census work is complete. A covers all **4,563 units**.
B covers the same **4,563 existing units**, plus one new erased import unit at
`src/compiler/sys.ts:56:1` with no finding. No existing unit first stop moved;
there are no new all-unit first-error findings naming process. Every moved unit
and every new process finding is therefore listed above (empty in both pairs).
`first-error/B-native-every-change.md` records the source-changing comparison;
`first-error/B-native-source.json` retains every common-unit comparison and the
added import. The six same-input B pairs contain 4,564 units, paired by read with
`stage3/census/pair`; they are not mislabeled as identical-source native pairs.
All twelve pieces completed, with disjoint file groups covering all 81 entry
implementation sources. The former interruption is historical (`STOPPED.json`);
`COMPLETE.json` and `first-error/summary.json` describe the final census state.

The final B piece was rerun fresh on both sides, serially, then committed and
pushed as `ff98b0e62`. It took **926.028 seconds** total, below its 60-minute
bound. Baseline peak RSS: **1,557,716 KiB**; head: **1,541,608 KiB**. Both exited 0.
Earlier partial evidence is retained under that piece's `interrupted/` directory
and excluded from accepted results. All-unit maximum accepted child RSS remains
**2,527,056 KiB**. Per-child resource records retain exact commands and peaks.

**The 79-root tables in both formats are not run by decision**, as explicitly
requested: the earlier `sys.ts:790:87` first wall masks the read at 1588. The ten
historical exploratory main/CommonJS controls all observed that 790 wall; they
are not a completed paired table or a measurement of all 79 roots. No further
ordinary root census was run. This report does not claim the omitted tables
verify Lower's root-level condition.

The accepted source amendment remains the one commit `e00ce47c6` on `7ccad5b5`.
Follow-up `e731288ba` repairs only its own inventory proof code/records; adapted
source and accepted gate bytes remain unchanged. Both are pushed on
`typescript/recordreplay-amend-1`. Census evidence is on
`typescript/recordreplay-census`. All pushes are new commits, never force.
Never submitted to Loom.

The read gate holds for the declaration-file form. The one-commit amendment is
pushed to `typescript/recordreplay-amend-1` at `e00ce47c6`, directly on `7ccad5b5`.
It changes only 99-wasm-crypto-import. Its scoped getNodeSystem before pin and
both node: import pins stay fixed; the after owner pin is recut. 93 also pins getNodeSystem whole; it runs before
99 and is unchanged. `gate/owner-pins.json` verifies both owners and import pins
against the fresh official snapshots, and verifies that 99 is the last adapter. The fresh official apply's pre-99 snapshot, replayed through
the amended adapter, matches the gate source byte for byte. Source scope,
no-write rerun and owner/declaration/out-of-scope mutation checks pass.

The first candidate placed the global augmentation in sys.ts. Stock tsc 6.0.3
and Adamic types both accepted it, all 79 transpiled JavaScript products matched,
and both declaration bundles matched. Its full sys.ts census nevertheless added
`Refused: a global augmentation` at 1977:1; this form was not selected.

The selected form creates `src/recordreplay.d.ts` and loads it with
`import type {} from "../recordreplay.js";`, occupying a reserved blank import
line in sys.ts. It removes `(process as any)` in favor of the direct `process`
receiver. The declaration contains only an optional unknown Process property;
no value or later view is added. All later sys.ts line numbers and all 79
compiler implementation roots are preserved. Stock TypeScript **6.0.3** with
pinned @types/node reports **0 before and 0 after diagnostics**; Adamic types
exits **0**. All 79 compiler JavaScript products are byte-identical by SHA-256.
Fresh `hereby dts` builds both exit 0. The complete declaration diffs are empty:

| Public declaration | Bytes | SHA-256, identical before and after |
|---|---:|---|
| typescript.d.ts | 593536 | f03adcf915df559d616fcafd75181d3423eaa0f272237482cedbd025525c9516 |
| tsserverlibrary.d.ts | 866 | cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd |

`gate/` retains every candidate's stock diagnostics, JavaScript hashes, types
output, full sys stream, source diff, public declarations and empty declaration
diffs, together with exact commands, exits and peak RSS. The accepted selected
full sys run peaks at **868000 KiB**. One short diagnostic run accidentally
overlapped that census; both overlapping observations were discarded and rerun
serially. `gate/discarded-overlap.json` records this correction. The shared
runner now holds a cross-process lock for the entire child run.

The full CommonJSBundle sys.ts pairs use `LATENT_FULL=1` and
`LATENT_ASSERT_NO_OUTPUT=1`. `full-sys/A-every-change.md` and
`full-sys/B-every-change.md` retain every gone, moved and new raw finding,
including continuation boundaries, in full. Their JSON also retains distinct
read pairing and complete per-reason totals. Both pairs have zero unit
first-stop changes and zero regressions under the pairing tool's definitions.

A changes the lowering reason at sys.ts:1588:191 from
`process.recordreplay as a value (use its supported process operation directly)`
to `process.recordreplay with an unrepresented type`. The separate namespace
scan finding at 1588:192 remains. This is one newly exposed process finding.

B clears the recordreplay lowering finding and its old failed-statement boundary.
It exposes `sys.ts:1613:46: reading modulePath` and two continuation boundaries,
with **zero new findings naming process**. No unit's earlier first stop moves.
The main compiler's B control is run on the same amended bytes as the head:
`stage3/census/pair` requires identical adapted source and unit eligibility.
No ordinary B table was run. The native source-changing full comparison is retained in
`full-sys/B-native-every-change.md`. The same-input B latent comparison is an
additional controlled observation, not a replacement for the source-changing
observations. A uses the original adapted source on both compilers. The adaptation inputs at
main and CLI head are identical, so the one fresh official control tree supplies
both sides of A. The amended adapter's exact output supplies both sides of B.

Actual `hereby local` builds of both trees also exit 0. Every emitted root-level
JavaScript bundle, including `_tsc.js` and `typescript.js`, is byte-identical;
`gate/built-javascript.json` records each size and SHA-256.

All worktrees, generated source, measurement outputs, upstream source and
configured measurement/build caches are under `/workspace/scratch/recordreplay`.
Setup deviation: the initial API `npm ci` ran before `npm_config_cache` was
redirected and wrote to the default `/home/agent/.npm` cache. Subsequent installs
and measurement commands use the scratch cache. The default cache was left
untouched after discovery; no source checkout outside scratch was edited. Upstream is v6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8`. Neither requested branch was present
before creation; all pushes are ordinary non-force pushes.

The complete stock tsc entry is also checked with TypeScript 6.0.3 and pinned
Node types: all 81 resolved implementation files are covered, with zero diagnostics
for the control and both candidates and byte-identical JavaScript for all 81.
See `gate/complete-entry-stock.json` and the complete diagnostic/hash records.


Inventory audit: the amended tree has 748 source files against the current
747-file control. The old 746-file count was already stale. The repaired 99
checker and scoped inventory segment pass; lane/apply identities, source
roundtrips and fresh speculative manifests include the added declaration.
Historical 91/92/93 scope checks and two foreign 746-count guards reject it;
only 99's own records/code were repaired. See [inventory/README.md](inventory/README.md)
for every tested contract, negative controls, per-reader index and execution
limitations. Static-only historical candidates were not run as whole drivers;
this is not a claim of exhaustive historical-reader execution.
