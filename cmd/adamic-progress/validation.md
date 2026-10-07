# Progress validation

Setup: 118s total; Go 0s, clang 1s, Node 1s, submodules 1s,
build-cache warm 118s. `nproc`: 5; cgroup quota: 4 CPUs.

Commands (output redirected to logs before reading):

* `go test -count=1 ./cmd/adamic-progress ./internal/stage1progress ./cmd/adamic-meter ./cmd/adamic-stage1-progress ./documentation/progress`: passed.
* `go vet ./...`: passed.
* `go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(strings|collections|exceptions)\.a$' -count=1 -timeout 30m`: passed in 11.219s.
* `go run ./cmd/adamic-progress --json`: 2.351s.
* `go run ./cmd/adamic-progress --history`: 2.218s.
* `git diff --check`: passed.

Every mutant below compiled and failed its named test through a Go overlay.
Production files were unchanged by mutation runs.

| Mutant | Test that caught it |
| --- | --- |
| checker-is-lowering | `TestRecordedMeterAndTrackHistories` |
| wrong-deadline | `TestGitSnapshotMissingRecordsAndDistinctBacklog` |
| invent-missing-meter | `TestGitSnapshotMissingRecordsAndDistinctBacklog` |
| wrong-lint-count | `TestGitSnapshotMissingRecordsAndDistinctBacklog` |
| ignore-invalid-meter | `TestMeterRefusesInconsistentRecord` |
| backlog-doubled | `TestGitSnapshotMissingRecordsAndDistinctBacklog` |
| inverted-bar | `TestBarsAgainstLiteralOracle` |
| reverse-rate | `TestSixHourRateNeededAndSparkline` |
| inverse-needed | `TestSixHourRateNeededAndSparkline` |
| reverse-pace-color | `TestSixHourRateNeededAndSparkline` |
| reverse-sparkline | `TestSixHourRateNeededAndSparkline` |
| ignore-javascript | `TestHostBothBackendsAndMissingResults` |
| ignore-missing-backend | `TestHostBothBackendsAndMissingResults` |
| allow-flat-backlog | `TestBacklogMustFallEveryElapsedHour` |
| reverse-overdue | `TestMicroDeadlineStatusesAndOverdueFirst` |

Logs: `/tmp/progress-tests-final.log`, `/tmp/progress-vet-final.log`,
`/tmp/progress-oracle.log`, `/tmp/progress-mutants-final.log`, and individual
`/tmp/progress-mutants/*-final.log`. Timing depends on the installed Go cache.
Go compilation happens before the executable's eight-second read budget.

Not covered: full compiler gate, new execution of port corpora or host fixtures,
macOS GUI execution, unrecorded historical branch refs, and actual remote push
times (Git records commit timestamps). Missing sources are explicit unknowns.
No successful main meter run or area/library checkpoint was available in this
checkout; the command does not invent those results.

## Patch backlog correction

Both counts remain displayed: commit identities (legacy comparison) and stable
patch IDs. Main candidates are safely narrowed by changed-path sets, with ASCII
whitespace normalized as patch-id does. Patch hashing and hourly reconstruction
run concurrently with isolated caches. Merge and empty commits have no standalone
patch. Squashes and conflict edits may change patch identity.

`go test -race ./cmd/adamic-progress -count=1` passed in 1.414s;
`go vet ./...` and `git diff --check` passed. Real-repo JSON took 7.280s;
history output took 7.875s before the final parallel candidate-scan optimization.
The fetched refs contained 1,702 commit identities versus 1,545 pending patches.
Fixture histories independently use git cherry to establish rebased equivalence,
then assert four commit IDs represent one patch, falling to zero patches after
that change lands with a new ID. Commit and patch archives remain separate.
Old main slices missing their required GAPS.md recording remain unknown.

| Correction mutant | Check that caught it |
| --- | --- |
| patch-ignore-main | `TestPatchBacklogRebasedAndSharedChanges` |
| patch-use-commit-identity | `TestPatchBacklogRebasedAndSharedChanges` |
| patch-skip-main-candidates | `TestPatchBacklogRebasedAndSharedChanges` |
| patch-credit-commit-archive | `TestHistoricalBacklogUnitsStaySeparate` |
| patch-invent-old-inventory | `TestHistoricalMissingInventoryIsUnknown` |

Logs: `/tmp/progress-patches-tests.log`, `/tmp/progress-patches-vet.log`,
`/tmp/progress-patch-mutants.log` and `/tmp/progress-mutants/patch-*-final.log`.
No compiler code changed; the full compiler gate was not rerun for this correction.
Past patch counts require `documentation/velocity/patch-backlog.csv`; historical
commit counts never supply the patch backlog or its falling-hourly milestone.

## Batch 8 parse instruction milestones

`go test ./cmd/adamic-progress -count=1` passed in 0.448s;
`go vet ./cmd/adamic-progress` and `git diff --check` passed.
`TestBatch8ParseInstructionMilestones` verifies the shipped plan's exact MDT
noon deadlines, runtime owner, source #93z4yv7, both inclusive thresholds,
complete batch, parse-only scope, integer counts and zero-denominator rejection.
All five source mutants compiled, then failed that test:

| Mutant | Check that caught it |
| --- | --- |
| parse-reverse-ratio | `TestBatch8ParseInstructionMilestones` |
| parse-credit-whole-run | `TestBatch8ParseInstructionMilestones` |
| parse-credit-short-batch | `TestBatch8ParseInstructionMilestones` |
| parse-reject-equality | `TestBatch8ParseInstructionMilestones` |
| parse-ignore-zero-count | `TestBatch8ParseInstructionMilestones` |

Logs: `/tmp/progress-parse-tests.log`, `/tmp/progress-parse-vet.log`,
`/tmp/progress-parse-mutants.log`, individual `/tmp/progress-mutants/parse-*-final.log`.
No parse benchmark or full compiler gate was run. Reported whole-run Go counts
are retained only as contextual observations, never credited as parse alone.
