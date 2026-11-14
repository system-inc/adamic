# Canonical Path result amendment (#xn6fj0p)

All requested proofs pass. No Loom submission was made.

- Implementation: `61a2277e9fdd4108ed89fb5c9487cee7b959d47c`, exactly one commit on `e531d51d66c45368afa7a9598bc6baa8602b4b41`.
- Branch: `typescript/path-mint-618-1`.
- Evidence branch: `typescript/path-mint-618-evidence`, this directory.
- Scout: `13b63e39bb0b304c83d4ee78ed5dd0ed9f8e3e1e`, road (b).
- Upstream TypeScript 6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`.

## Amendment

Only the whole createDirectoryWatcherSupportingRecursive after text and SHA-256
in 99-correctness-timer-closures are amended. The string-returning canonicalizer
has no function cast. The directory key, both recursive invokeCallbacks branches,
and child-watcher cache lookup each mint the result with `as Path`, exactly as
road (b). The function-value and includes readers retain strings. No helper,
normalization, guard or runtime operation is introduced. The before pin and all
other owner pins remain unchanged. The README records the amendment and the
ledger has one type-only line citing #xn6fj0p. `owner.patch` and `source.patch`
show every edit; only sys.ts changes among 747 source files.

## Results

| Proof | Result |
| --- | --- |
| Whole-owner contract | PASS: 13 owners / 9 files, drift/duplicate/moved/half-owner refusals, LF/CRLF, unrelated edits, no-write byte/mtime replay |
| Final adapted-tree replay | 0 files / 0 edits; bytes and mtimes preserved |
| Stock tsc 6.0.3 | Main 0 diagnostics, amendment 0 diagnostics, 79 recursive roots; 0 new |
| Adamic types | Main exit 0, amendment exit 0, all 79 recursive roots |
| Full apply | Main and amendment exit 0; lane receipt identifies implementation commit |
| Emitted JavaScript | All 604 non-declaration source outputs SHA-256 identical; 0 new/gone/changed files (including all 79 compiler roots) |
| Upstream hereby dts | Both commands exit 0; typescript.d.ts and tsserverlibrary.d.ts byte-identical |
| Whole unfiltered lane | PASS: 106,366 passing, one sanctioned Public APIs failure, 0 pending; 8 workers, all runners, no filter |
| Existing timer witnesses | Main and amendment stdout byte-identical |
| Existing restoring-slot mutant | Expected exit 1, restored-variadic-slot / System owner drift |

The lane's only failure is
`unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`;
its only baseline difference is `api/typescript.d.ts`. Apply/install/build exit 0;
tests/oracle exit 1 as the lane requires. Timeout, exception, missing baseline and
other failure counts are all zero. No sanction, expectation, compiler, harness,
lane implementation or generated manifest was changed.

## Complete full-census delta

All three runs use the scratch-only census overlay and
`LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_ONLY_FILE=<exact sys.ts>` with
CommonJSBundle. Each stream has its relation-report trailer, zero checker
diagnostics, and no errors, panics or relation overflow. The fixed measurement
label says checker-rejected; every actual header says `checker_rejected:false`.

| Run | Unique NotYet | Refused | Boundary |
| --- | ---: | ---: | ---: |
| Main | 41 | 19 | 55 |
| Amendment | 41 | 18 | 55 |
| Restoring function cast | 41 | 19 | 55 |

Deduplicate `(kind, where, reason, text)` across attempts. No lines are inserted
and no coordinate normalization beyond removing scratch tree prefixes is used.
`census-delta.json` retains full semantic and phase/attempt context deltas.

Gone findings (complete list, one):

- Refused, `src/compiler/sys.ts:618:33`, refusal_scan, attempting owner
  `src/compiler/sys.ts:590:1`: **a cast the runtime can't check**.
  Full text: `src/compiler/sys.ts:618:33: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)`.

New findings: **none**. Changed findings (including kind, reason, diagnostic text,
phase and attempting-unit context): **none**. The only context removed is the
618:33 refusal_scan record for the 590:1 owner.

The restoring mutant puts only the original function cast back, preserving CRLF
and retaining the four result casts. The complete reverse delta adds only the
618:33 refusal above; 0 other new, gone or changed findings. Its full census
stream, sys.ts and one-line restoring patch are retained.

Coverage limit: the owner still stops in its prologue at sys.ts:594:5, refusing
a method in object destructuring, with its owner boundary at 590:1. This proves
the measured cast refusal is removed without new census findings; it does not
claim end-to-end native lowering of this watcher or that these four new mints
are reached by ordinary lowering. No usable IR is produced by the census.

## Declarations and provenance

Both upstream tasks are `node node_modules/hereby/bin/hereby.js dts` in their own
fresh main/amendment trees. Direct byte comparisons pass:

- typescript.d.ts: `f03adcf915df559d616fcafd75181d3423eaa0f272237482cedbd025525c9516`
- tsserverlibrary.d.ts: `cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd`

`provenance.json` pins both binaries, upstream and submodule commits, and every
source file in the two full apply trees. `acceptance.json` gathers the verified
results. Full raw censuses, logs, source snapshots, emitted-file hash ledgers,
declarations, contract/replay receipts and lane evidence are retained here.
`SHA256SUMS` hashes the complete final evidence payload except itself.

Scripts use `/workspace/scratch/path-mint-618`; adjust that scratch prefix to
reproduce elsewhere. Start from a new detached worktree of the base, initialize
submodules recursively with depth 1, clone upstream v6.0.3 with depth 1, and run
stage3/api npm ci. env.sh declares the pinned source, Go/tool caches and temp dirs.
Run baseline apply before amendment; then amend.py/amend-docs.py, contract-seed.py
and contract.cjs; commit the three implementation files; run the whole lane and
baseline-proofs.sh/after-proofs.sh/declaration-proofs.sh with their recorded trees.
The final report/assertions are generated by final-report.py.

Setup deviation: the initial stage3/api npm ci and main baseline apply used their
default cache locations. Subsequent commands redirect npm, stage3, Go and temp
caches to scratch. Repository source edits and build trees are in the isolated
scratch worktree; the shared checkout's source files were not modified. No force
push, baseline acceptance, PR submission or Loom submission was performed.
