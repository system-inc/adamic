# Publication status

This is the evidence-only stop report. Automatic approval review blocked exporting the raw generated type log, the candidate source diff, and the compressed census payload. Explicit export approval is pending. This published checkpoint contains summaries, full phased finding deltas, the read inventory and proof scripts/receipts; the raw artifacts described below remain in `/workspace/scratch/watch-directory-1533/work/stage3/census/watch-directory-1533/`. `PUBLICATION.json` lists exactly the artifacts published in this checkpoint. `SHA256.json` inventories the complete local bundle, including files not yet published. No implementation is submitted and no Loom action was taken.

# watch-directory 1533: stop at the census gate

Base: `8e3a784aa5a601f57342c07b68f145fe56252c21`.
Upstream: TypeScript v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`.
Task: type-only watchDirectory 1533 adaptation, excluding watchFile 1532. No Loom submission.

**Decision: STOP.** The read gate holds, but a consistent type-only widening creates two new NotYet findings and three associated Boundary records. The requested zero-new-finding gate fails. No implementation branch, adaptation amendment, ledger line, or landing claim is supplied. Only the authorized evidence branch is published.

The final scratch experiment widens `recursive: boolean` to `recursive: boolean | undefined` at 12 signature positions: nine in sys.ts, one in virtualFileSystemWithWatch.ts, and two in watchUtils.ts. This includes the explicitly returned HostWatchDirectory alias and the concrete downstream contracts. It changes no runtime expression. `candidate.diff` and the compressed complete source snapshots retain every byte.

## Every recursive read

All reads either test truthiness, negate/coerce the boolean, or forward to another reader listed here. No read distinguishes undefined from false. `reads.json` retains the exact source text on both trees. Parameter declarations, comments and child-watcher calls with literal false are not incoming-value reads.

| File | Adapted line | Upstream line | Reader |
| --- | ---: | ---: | --- |
| src/compiler/sys.ts | 1111 | 1108 | pass to fsWatch |
| src/compiler/sys.ts | 1129 | 1126 | pass to hostRecursiveDirectoryWatcher |
| src/compiler/sys.ts | 621 | 623 | truthiness branch |
| src/compiler/sys.ts | 623 | 625 | pass to nonRecursiveWatchDirectory |
| src/compiler/sys.ts | 1133 | 1130 | !recursive assertion |
| src/compiler/sys.ts | 1163 | 1160 | pass to fsWatch |
| src/compiler/sys.ts | 1211 | 1207 | truthiness selects watcher map |
| src/compiler/sys.ts | 1215 | 1211 | pass to fsWatchHandlingExistenceOnHost |
| src/compiler/sys.ts | 1269 | 1264 | pass to worker or timestamp worker |
| src/compiler/sys.ts | 1367 | 1362 | pass to worker |
| src/compiler/sys.ts | 1776 | 1782 | !!recursive normalizes before Node |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts | 889 | 889 | pass to watchUtils.fsWatch |
| src/harness/watchUtils.ts | 184 | 182 | truthiness selects watcher map |

The only concrete createSystemWatchFunctions callers are getNodeSystem and TestServerHost. At the Node boundary `!!recursive` is exactly false in both cases. The test worker eventually chooses its map with a ternary. Missing-entry/polling fallbacks do not read recursive. The complete required owners/targets/adapt.cjs search is `owner-search.json.gz`; the factory is pinned by 99-correctness-single-watcher-forwarder, TestServerHost by 99-correctness-timer-closures, and getNodeSystem by 99-wasm-crypto-import (with preceding 93/41 changes). Since the census fails, no pins are amended.

## Proof results

- Full single-file sys.ts census: `LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_ONLY_FILE=<absolute sys.ts>`, all 79 compiler roots, CommonJSBundle. All three final runs exited 0 and have the final relation trailer. Main has 185 records; candidate 189. There are **2 new NotYet, 3 new Boundary, 1 gone Refused, 0 changed**. Including every kind, 5 new / 1 gone / 0 changed. The gone finding is refusal_scan 1533:13. Every new/gone record and its phase appears in full in delta.json; no other finding disappears. Main and candidate have zero checker diagnostics.
- New NotYet, phase `lowering`, sys.ts:1215:82: `a boolean | undefined variable a function value captures`.
- New NotYet, phase `lowering`, sys.ts:1216:208: `reading callbacks`.
- New Boundary, phase `lowering`, sys.ts:1210:9 and 1216:170: `failed statement; state rolled back; continued at next statement or declaration`, with the captured-optional-boolean failure text above.
- New Boundary, phase `lowering`, sys.ts:1216:204: the same boundary reason, with the reading-callbacks failure text above.
- 1532:13 remains byte-for-byte the same Refused finding (apart from normalized tree paths) in main, candidate, and mutant. It is not modified.
- Restoring mutant: all three source files are restored to their original exact bytes. It returns 1533:13 and its complete contextual findings equal main. 1532 is still present. Complete streams are main.jsonl.gz, change.jsonl.gz, mutant.jsonl.gz. Comparisons normalize only tree paths and exclude shifted numeric byte offsets from identity; original offsets remain in every raw stream and reported delta record. Unit, phase, kind, location, reason and text remain in the contextual comparison.
- Stock tsc 6.0.3: 79 roots, 0 main diagnostics, 0 candidate diagnostics, 0 new. The script enables the pinned Node ambient types and otherwise follows production checker options. stock-delta.json retains the result.
- Ordinary `adamic types`: all 79 roots, exit 0. Complete declaration/type output is retained losslessly in adamic-types.log.gz.parts/ (concatenate the numerically ordered parts, then gunzip).
- JavaScript: 604 non-declaration TypeScript files emitted independently with stock 6.0.3; every corresponding output is byte-identical. All 747 source files were compared; only the three signature-only source files differ. javascript.json.gz retains each output hash.
- stage3/apply.sh --build-product: unmodified main passed, exit 0; its patch table is baseline-patch-set.md. No candidate adapter was installed because the census gate failed, so candidate apply and adapter contracts are not claimed.
- Declarations: upstream's own npm ci and node_modules/.bin/hereby dts passed, exit 0, with their logs retained. A main/candidate declaration comparison was not run after the stop; there is no fabricated empty diff.
- Whole unfiltered stage3/lane: not run after the census stop; 106,366 passing is not claimed.
- Ledger: unchanged because no adaptation is built.

## Reproduction and provenance

All work took place in the scratch detached worktree and scratch source trees, not the shared checkout. Recursive shallow submodules were initialized. GOTMPDIR, TMPDIR and GOCACHE were scratch paths for Go builds; setup.json records commits and Node version. make_overlay.py generated the guarded measurement binary, with CommonJSBundle selected by the repository's own unchanged census templates. The raw run logs and build logs are included. Every Go source written by the measurement overlay was outside stage3; no new .go file is committed here.

Commands, with R=/workspace/scratch/watch-directory-1533 and W=$R/work:

```sh
ADAMIC_TYPESCRIPT_SOURCE=$R/TypeScript XDG_CACHE_HOME=$R/cache TMPDIR=$R GOTMPDIR=$R GOCACHE=$R/go-cache $W/stage3/apply.sh --build-product $R/baseline
python3 $W/stage3/census/latent/make_overlay.py $W $R/overlay
go build -buildvcs=false -overlay=$R/overlay/overlay.json -o $R/census ./stage3/census/latent/tool
# For each T in baseline, candidate, mutant; O in main, change, mutant respectively:
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_ONLY_FILE=$R/$T/src/compiler/sys.ts $R/census $R/$T/src/compiler $R/$O.jsonl
NODE_PATH=$W/stage3/api/node_modules node $R/stock.cjs $R/baseline $R/stock-main.json
NODE_PATH=$W/stage3/api/node_modules node $R/stock.cjs $R/candidate $R/stock-change.json
NODE_PATH=$W/stage3/api/node_modules node $R/emit.cjs
python3 $R/delta.py
```

`experiment.py SCRATCH_PARENT` reproduces the final candidate from a completed baseline. The initial one-parameter pilot is isolated under pilot/: it widened the implementation parameter but left the explicit result alias unchanged, producing two diagnostics while leaving 1533 present. It was corrected before the final census by consistently widening the already-traced signature contracts. The final exact candidate is the diff and source snapshots; experiment.json records its scope. Once the corrected experiment produced new findings, no further widening or runtime change was attempted.

The first shell push lacked GitHub credentials. The connected GitHub tools published checkpoint c1954286b0fe5c8118177604ea4ba1e541a2463c and publish the final evidence by a fast-forward ref update, never force. No PR, message, or Loom submission is created.
