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

## Bounded backlog and partial reports

The previous all-diff pipe and shared eight-second deadline could kill backlog
hashing on a real checkout. Hashing now runs in batches of at most 16 commits,
with SHA-keyed entries atomically checkpointed under the Git directory after
each completed batch. A warm run never hashes cached commits. Its independent
60-second context and bounded pipe cleanup leave other sections available.
Backlog failure is `null` in JSON with its error and an explained missing line in
text, not zero or a nonzero command exit. Individual record/section failures also
leave the rest of the report intact. Repository bootstrap/output failures remain
errors. Existing rebase and shared-patch deduplication tests still pass.

Tests create 300 independent remote branches and 3,000 off-main commits
(3,003 total) with Git fast-import, inside a 60-second context. They verify actual
counts, persisted SHA entries, 16-commit bounds, a warm run forbidden to invoke
patch-id, hashing only one newly added commit, and the full native CLI. Both
backends of the progress output (text/JSON), plus --history, exit zero when a
fixture kills patch-id. Additional checks cover abbreviated deadlines, the
independent timeout, corrupted caches, an interrupted later batch retaining the
first checkpoint, and other failed sections preserving available Apple evidence.

Commands and results (test output redirected to logs before reading):

* `go test ./cmd/adamic-progress -count=1 -v -timeout=90s`: passed; scale test
  completed in 2.255s including fixture construction, both backlog runs,
  full report, native CLI build and invocation.
* `go test -race -count=1 ./cmd/adamic-progress ./internal/stage1progress ./cmd/adamic-meter ./cmd/adamic-stage1-progress ./documentation/progress`: passed.
* `go test -race ./cmd/adamic-progress -count=1 -v -timeout=90s`: passed on the
  final code; its instrumented scale test completed in 11.051s, still under its
  60-second context (cold backlog 7.002s, warm 0.175s, native CLI 0.224s).
* `go vet ./...`, formatting and `git diff --check`: passed.
* `go run ./cmd/adamic-progress --json` on the actual fetched cloud checkout:
  cold 9.167s, warm 2.245s, exit 0 with no section errors.

| Repository/run | Observed wall time | Instrument |
| --- | --- | --- |
| Actual cloud checkout: 386 remote refs, 1,702 commits off main, cold full go run | 9.167s | Bash time wall clock |
| Same checkout, warm full go run | 2.245s | Bash time wall clock |
| Same checkout, cold backlog: 1,623 SHAs, 102 batches | 8.343s | Go time.Now/time.Since monotonic clock, JSON stats |
| Same checkout, warm backlog: zero hashes, 1,623 cache hits | 0.194s | Go time.Now/time.Since monotonic clock, JSON stats |
| Generated 300-branch/3,000-commit fixture, cold backlog | 1.215s | Go time.Now/time.Since monotonic clock |
| Generated fixture, warm backlog | 0.060s | Go time.Now/time.Since monotonic clock |
| Generated fixture, complete cached native CLI | 0.076s | Go time.Now/time.Since monotonic clock |

Both observed actual-repo counts remained labelled: 1,702 commit identities and
1,545 pending patches. Go compilation is included in Bash go run wall time;
fixture CLI timings use the compiled executable. Timings vary with CPU/cache
load. These measurements are Linux x86_64 observations; Kirk's Mac was not
accessible in this execution environment.

All new mutants compiled and then failed the named assertion checks:

| Mutant | Check that caught it |
| --- | --- |
| bounded-unbounded-batch | `TestPatchBatchBoundAndIncrementalCache` |
| bounded-ignore-cache | `TestPatchBatchBoundAndIncrementalCache` |
| bounded-no-checkpoint | `TestCompletedPatchBatchesSurviveFailure` |
| bounded-fatal-section | `TestBacklogFailureKeepsReport` |
| bounded-invent-zero | `TestBacklogFailureKeepsReport` |
| bounded-shared-timeout | `TestBacklogBudgetIndependentOfReportReads` |
| bounded-ignore-deadline | `TestBacklogFailureKeepsReport` |
| bounded-trust-invalid-cache | `TestCompletedPatchBatchesSurviveFailure` |
| bounded-fatal-track | `TestFailedSectionKeepsOtherTracks` |

The original 15 bar/count/rate/deadline mutants were rerun against this final
implementation and caught by their previously listed tests. Logs:
`/tmp/progress-bounded-tests.log`, `/tmp/progress-bounded-tests-final.log`,
`/tmp/progress-bounded-race.log`, `/tmp/progress-bounded-progress-race.log`,
`/tmp/progress-bounded-vet.log`, `/tmp/progress-bounded-mutants.log`,
`/tmp/progress-previous-mutants-rerun.log`, and individual
`/tmp/progress-mutants/bounded-*-final.log`.
Actual-repo outputs: `/tmp/progress-bounded-real-{cold,warm}.json`,
`/tmp/progress-bounded-real-warm.time`.

Not covered: an actual macOS run, full compiler gate, or peak-memory profiling.
Neither gates nor fixtures are executed by the progress command. This correction
is to recorded-evidence reads, patch hashing, cache persistence and reporting.
