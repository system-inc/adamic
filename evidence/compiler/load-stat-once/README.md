# Evidence: a load stats each .a file once (compiler/load-stat-once, #npd568j)

Tested for compiler Performance. The change was not edited; mutants were applied, run and reverted.

- Change: `compiler/load-stat-once` at `173dcdca6d449322749960336518c2fcdf7852e9`, one commit on main `5f8e15794ce3bd0bc6bfda913c77756cdfb37bac`. Touches only `internal/load/source_fs.go` and the new `internal/load/source_fs_stats_test.go`.
- Box (`logs/02-box.log`, `logs/01-setup.log`): Linux 6.18.44 x86_64 cloud VM, `nproc` 4, 15.7 GB, `go version go1.27.1 linux/amd64`, clang 20.1.8, node v24.22.2, strace at /usr/bin/strace (ptrace allowed).
- Setup: `git submodule update --init --recursive --depth 1` (exit 0, `logs/00-submodule.log`), `bash cloud/setup.sh --wasi-sdk` (exit 0, 564.6 s), `source /opt/adamic-tools/env.sh`.
- Every log ends in `exit=<code>`. Scripts used for runs 4 and 5 are in `scripts/`.

## Results, observed

| Run | Command | Exit | Key lines |
| --- | --- | --- | --- |
| 1 | `go test ./internal/load -run '^(TestALoadStatsAnAdamicFileOnce\|TestAFileMadeBetweenTwoLoadsIsFound)$' -count=1 -v` | 0 | both `--- PASS` (`logs/03-run1-tests.log`) |
| 2a | same, with the `if kept, ok := s.adamicFiles.Load(path); ok { ... }` block deleted (`logs/04-mutant-a.diff`) | 1 | `source_fs_stats_test.go:49: one load asked .../shapes.a.ts 1000 times, want once`; `--- FAIL: TestALoadStatsAnAdamicFileOnce`; TestAFileMadeBetweenTwoLoadsIsFound passed (`logs/04-mutant-a.log`) |
| 2a restore | `git checkout -- internal/load/source_fs.go; git diff` | | empty diff, 0 bytes (`logs/04-mutant-a-restored.log`) |
| 2b | same, with package-level `var sharedAdamicFiles sync.Map` used for both the Load and the Store (`logs/05-mutant-b.diff`) | 1 | `source_fs_stats_test.go:72: the second load, after late.a was written, gave .../main.a:1:23: error TS2307: Cannot find module './late.a' ...; want it found`; `--- FAIL: TestAFileMadeBetweenTwoLoadsIsFound`; TestALoadStatsAnAdamicFileOnce passed (`logs/05-mutant-b.log`) |
| 2b restore | `git checkout -- internal/load/source_fs.go; git diff` | | empty diff, 0 bytes (`logs/05-mutant-b-restored.log`) |
| 3 | `go test ./internal/load -count=1` | 0 | `ok github.com/system-inc/adamic/internal/load 23.377s` (`logs/06-load-whole.log`) |
| 3 | `go vet ./internal/load` | 0 | no output (`logs/07-vet.log`) |
| 4 | `go build -o <dir>/adamic ./cmd/adamic` at 5f8e1579 (a worktree) and at 173dcdca | 0, 0 | `logs/08-build-*.log` |
| 4 | `adamic c <file>` with both binaries, for every `.a` under `internal/oracle/testdata` (recursive, `logs/09-files.txt`), stdout, stderr and exit code to separate files, then `diff -r` | diff 0 | 2077 files; 6231 output files per side; 0 differ (`logs/09-summary.txt`, `logs/09-diff.txt` is empty) |
| 5 | `strace -f -c -e trace=newfstatat,statx,stat,lstat,fstatat64` around the prebuilt `lower.test -test.run '^TestCycleFinderEquality_00[0-9]$' -test.count=1 -test.v`, two rounds, alternating main and change | 0 x4 | see below |

Run 4 detail: in both sides' outputs, 1637 files exited 0 with C on stdout and 440 exited 1 with a message on stderr (refusals, NotYet, expected type errors in the refused fixtures). Those 440 are byte-identical too, exit codes included.

### Run 5, syscall count

Test binaries were built with `go test -c ./internal/lower` in each sha's own tree (exit 0, `logs/10-testc-*.log`) and each ran with its own tree's `internal/lower` as working directory. `internal/lower` testdata and the fixtures are the same at both shas (the commit touches only `internal/load`). The pattern runs 10 shards of TestCycleFinderEquality, 312 fixture subtests, 322 PASS and 0 FAIL in every run.

| Run | stat-like calls (all `newfstatat`) | of which errors (ENOENT) | wall, under strace |
| --- | ---: | ---: | ---: |
| main 5f8e1579, round 1 | 681,087 | 340,446 | 79.5 s |
| main 5f8e1579, round 2 | 681,087 | 340,446 | 83.3 s |
| change 173dcdca, round 1 | 32,092 | 16,035 | 49.2 s |
| change 173dcdca, round 2 | 32,092 | 16,035 | 60.5 s |

No `statx`, `stat`, `lstat` or `fstatat64` calls were recorded; only `newfstatat`. Files: `logs/11-strace-{main,change}-{1,2}.{strace.txt,summary.txt,test.log}`.

The same pattern without strace, alternating (`logs/12-nostrace-wall.txt`): main 42.6 s and 42.4 s, change 40.5 s and 41.4 s, all exit 0.

## Inferred, not measured

- The change removes 649,000 of 681,000 stat calls in this pattern, about 21 times fewer, and the count is deterministic across rounds. The errors halve with the calls in the same ratio, consistent with the `.a.ts` probe (which fails for every `.a` file) being asked once instead of on every `Where`.
- The wall gain under strace (about 25 to 30 s) is mostly ptrace cost per syscall, so it overstates the real gain. Without strace the difference is about 1 to 2 s of 42 s on this 4 core box, within what two samples per side can resolve; I would not claim a wall time win from these numbers.
- The byte-identical `adamic c` output over 2077 fixtures shows the cache doesn't change what one load produces. It doesn't exercise a file appearing between two loads in one process; that is what TestAFileMadeBetweenTwoLoadsIsFound and mutant 2b cover.

## Not covered

- The full gate (products, `go test ./...`) was not run. Only `internal/load` whole, and the one `internal/lower` pattern above.
- macOS, and any process that keeps one sourceFS across several loads (none was looked for).

## Exploratory runs, kept for honesty

While choosing the run 5 pattern I ran the change's whole `lower.test` minus `^TestProduct_` once (785 s, exit 1). Its two top-level failures were `TestCycleFinderCensusEquality` and `TestProgramRegionCensusMembership`, both saying `ADAMIC_TYPESCRIPT_SOURCE must name the pinned upstream TypeScript checkout`: an environment variable this ad hoc run didn't set, not the change (`logs/13-explore-lower-whole-failures.txt`). A nested `--- FAIL: TestNodeRealpathNativeClosedUses/two_arguments.a` in that log is printed inside TestNodeRealpathNativeUseMutantIsCaught, which passed; it's that test's own mutant failing as it should. That run was not repeated at main. An earlier `^TestClassFeatures` probe failed with `'sanitizer/asan_interface.h' file not found` because I had not sourced `env.sh` in that shell (`logs/14-explore-classfeatures-without-env.log`); every run reported above sourced it.
