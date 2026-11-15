# Verification — #et209ny

Records refusal-scan stops only, plus the first declaration, body or tail stop
reached; blind past that first stop. Zero body-phase or tail-rule counts are never
a pass. The two example reports repeat this warning above every table.

Base main: `fe5a1a54f045e9bce4e5aa32ca74885a317fbebd`. The original scout is
`0d078a8da915d349e3fc70831fd213287afcf017`; its four independent rule heads are
carried byte for byte and identified in every report header. They are not claimed
to be merged into main. The tool's orchestration uses this base's production
lowering after its continued scan.

Setup used a fresh detached main worktree at `/tmp/et209ny/adamic`, initialized
recursive shallow submodules, and TypeScript v6.0.3 confirmed at
`050880ce59e30b356b686bd3144efe24f875ebc8`. `npm ci` ran in `stage3/api`.
`GOTMPDIR`, `TMPDIR`, `GOCACHE`, `ADAMIC_BUILD_CACHE_DIR` and `STAGE3_CACHE` all
pointed into `/tmp/et209ny`. Main advanced from `35ef35961` during verification;
the draft was moved to `fe5a1a54f`, the adapted tree rebuilt with its
`stage3/apply.py`, and both example measurements repeated. The earlier
measurements are not included here.

## Complete package and mutants

`go test -overlay "$scratch/overlay/overlay.json" -count=1 -timeout=60s -v ./stage3/census/scan-all/...`
passed all **12 tests**, with **no skips**, in **0.217 s**. The generator and tool
subpackages contain no tests. Ordinary package tests also passed (the eight
measurement controls explicitly skip without the overlay, as documented).
Preparation builds ran outside tests. No new test runs a build subprocess; each
calls `t.Parallel()`.

The following controls each passed normally. Each was then run separately with
its named first-error mutant using the precompiled overlay test binary and
`-test.run=^NAME$ -test.timeout=60s`. Every mutant selected production's untouched
`refuse`, exited 1 and turned **that same named test red**. Only the earlier
`@ts-check` scan stop remained: positive targets disappeared, and clean cases no
longer completed. Positive controls also require the exact tail phase and the
stop's order after the earlier scan stop.

Records refusal-scan stops only, plus the first declaration, body or tail stop
reached; blind past that first stop. Zero body-phase or tail-rule counts are never
a pass.

| Test | `SCAN_ALL_MUTANT_FIRST_ERROR` | Normal | Mutant |
| --- | --- | --- | --- |
| `TestScanAllCensusBindStop` | `bind-stop` | PASS | RED, exit 1 |
| `TestScanAllCensusBindClean` | `bind-clean` | PASS | RED, exit 1 |
| `TestScanAllCensusRecordStop` | `record-stop` | PASS | RED, exit 1 |
| `TestScanAllCensusRecordClean` | `record-clean` | PASS | RED, exit 1 |
| `TestScanAllCensusCombinedBindStopRecordStop` | `combined-bind-stop-record-stop` | PASS | RED, exit 1 |
| `TestScanAllCensusCombinedBindStopRecordClean` | `combined-bind-stop-record-clean` | PASS | RED, exit 1 |
| `TestScanAllCensusCombinedBindCleanRecordStop` | `combined-bind-clean-record-stop` | PASS | RED, exit 1 |
| `TestScanAllCensusCombinedBindCleanRecordClean` | `combined-bind-clean-record-clean` | PASS | RED, exit 1 |

Additional tests passed:

- `TestScanAllCensusMainAnchorsStillMatch`: current-main source anchors match and the generated refusal clone has no diagnosed-body exclusions.
- `TestScanAllCensusDriftNamesItsAnchor`: either orchestration/refusal snapshot drifting fails with its named anchor and no partial overlay.
- `TestScanAllCensusRuleHeadBaseDriftFailsLoudly`: changing any one of the four whole-file replacement bases fails with that file's named anchor.
- `TestScanAllCensusScratchCannotBeProduction`: production destinations fail; an outside scratch destination passes.

## Final closure measurements

Each row is **one root import-closure run**. Counts are events and distinct
`(rule, site, words)` triples, separately. Per-file tables in the linked reports
come only from recorded sites. Linux `wait4` measures wall time and peak process
RSS, excluding builds and adaptation and including load/check/measurement.

Records refusal-scan stops only, plus the first declaration, body or tail stop
reached; blind past that first stop. Zero body-phase or tail-rule counts are never
a pass.

| Root | Events | Distinct triples | Wall seconds | Peak RSS, KiB | Pipeline completed |
| --- | ---: | ---: | ---: | ---: | --- |
| [sys.ts](sys.md) | 5538 | 4022 | 195.508535 | 1,917,916 | false |
| [corePublic.ts](corePublic.md) | 0 | 0 | 0.041657 | 32,248 | true |

The sys.ts declaration stop leaves body and tail outcomes unknown. CorePublic's
completed measurement with zero events is not a compiler acceptance result.
No raw observation streams are committed.

## Required gate packages and isolation

`go test -count=1 -timeout=60s ./cmd/adamic-gate/... ./internal/compilercensus/...`
passed the complete trees:

- `cmd/adamic-gate`: PASS, 1.503 s.
- `cmd/adamic-gate/parallelcensus`: PASS, 2.442 s.
- `internal/compilercensus`: PASS, 2.092 s.
- `internal/compilercensus/cmd/dependencies`: no test files.

The final ordinary scan-all package tree was included in that run too.
Both the generator package and generated driver declare their compiler
consumption, with explicit markers for imports hidden in `.go.txt` templates.

`git diff --exit-code -- internal/lower stage3/meter stage3/fixtures` was empty.
`rg -n 'scan-all|ScanAll|SCAN_ALL' cmd cloud internal stage3/meter stage3/fixtures`
found only the two package entries in
`cloud/fast-gate/compiler-dependencies.json`: **no gate reads the output**.
All generated copies, the overlay map, binaries, temporary files, raw receipts
and caches stayed in scratch. The driver and all eight controls checked that
both public lowering APIs return nil IR plus an error.

No Loom submission was made.
