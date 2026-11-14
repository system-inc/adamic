# Recordreplay amendment and paired censuses

| Pair | 79-root first stops, both formats | Full sys.ts unit first stops |
|---|---|---|
| A: main 9c304036 against CLI 7ccad5b5 | Running; no complete paired table yet | None moved |
| B: main 9c304036 against amendment e00ce47c6 | Pending | None moved |

This is an incomplete checkpoint. The paired 79-root tables are running serially;
the all-tsc latent first-error pieces are running in six verified disjoint groups. No unmeasured result is
claimed. Never submitted to Loom.

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
The ordinary B table retains original main source on its baseline side; the
native source-changing full comparison is also retained in
`full-sys/B-native-every-change.md`. The same-input B latent comparison is an
additional controlled observation, not a replacement for the source-changing
observations. A uses the original adapted source on both compilers. The adaptation inputs at
main and CLI head are identical, so the one fresh official control tree supplies
both sides of A. The amended adapter's exact output supplies both sides of B.

Actual `hereby local` builds of both trees also exit 0. Every emitted root-level
JavaScript bundle, including `_tsc.js` and `typescript.js`, is byte-identical;
`gate/built-javascript.json` records each size and SHA-256.

The requested next-stop scripts are retained verbatim under `next-stops/` and
retargeted through environment paths. Their group scheduler is serial, and each
root is observed twice by the unchanged ordinary tool. The census binaries have
no backend and the latent binaries retain the no-output guards. All worktrees,
temporary files, caches, upstream source and builds used for measurement are
under `/workspace/scratch/recordreplay`. Upstream is v6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8`. Neither requested branch was present
before creation; all pushes are ordinary non-force pushes.
