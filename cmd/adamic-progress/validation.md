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
