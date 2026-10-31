# Evidence: compiler/repository-once (#npd568j)

Tested for compiler Performance. The change as asked: `compiler/repository-once` at
9bc0fc5227e07c60f288c571b746ddd838fd9a67, one commit on main 2a032b3ce59a023dcc80a58d7b9d743ff7737bb6. This branch
sits on 9bc0fc52 and adds only `evidence/`. Every log ends with the command's `exit=` line.

Partway through, Kirk passed on the reviewer's must and should, to be fixed in a second commit on the same branch.
That commit is 1f6fc7575ec69c14937574de09a51b1c8f2709fc, pushed with a plain push (section 5).

## Box (02-box.log)

- go version go1.27.1 linux/amd64; nproc 4; Linux 6.18.44 x86_64; git 2.43.0; strace 6.8
- Running as uid 0 (root), TMPDIR=/tmp/adamic-gate
- Setup: `git submodule update --init --recursive --depth 1` (00-submodules.log, exit 0), then
  `bash cloud/setup.sh --wasi-sdk` (01-setup.log, exit 0, 372 s), then `source /opt/adamic-tools/env.sh`.
  cohere at 06af2e399c6c1fb93fd570b774027942747400b4.

## 1. The three new tests, at 9bc0fc52 (10-new-tests.log)

`go test ./internal/buildcache -run '^(TestOneCheckoutIsListedOnceHoweverManyAsk|TestOneCheckoutByThreeNamesIsListedOnce|TestAFailedListingIsNeverKept)$' -count=1 -v -race`

exit=0, and all three PASS.

## 2. Mutants, at 9bc0fc52 (2a/2b/2c-mutant-*.log; applied by scripts/mutate.py)

Each log shows the mutant's `git diff`, the run, and then `git diff` after `git checkout --` restored the file
(empty, and `restored-clean`).

| Mutant | Result | Key line |
|---|---|---|
| a. The wait dropped: the asker calls `readRepository(directory)` itself | exit=1. TestOneCheckoutIsListedOnceHoweverManyAsk FAIL | `one checkout was listed 132 times, want once` |
| b. The `if run.failed { CompareAndDelete }` block deleted | exit=1. TestAFailedListingIsNeverKept FAIL | `the next asker after a failure got &{... all:true ...}, want a fresh listing` |
| c. `repositoryKey` returns `directory` as given | exit=1. TestOneCheckoutByThreeNamesIsListedOnce FAIL | `one checkout by four spellings was listed 4 times, want once` |

Mutant a also turned TestOneCheckoutByThreeNamesIsListedOnce red (4 listings). That is expected: with no wait, every
asker after the first lists again. Each mutant made its own named test go red.

## 3. Whole package and vet, at 9bc0fc52

- `go vet ./internal/buildcache` (31-vet.log): exit=0.
- `go test ./internal/buildcache -count=1 -timeout 30m` (30-package.log): **exit=1**, 93 s. Five top-level tests
  fail (list in 30-failing.txt):
  TestAHandedHouseFileTheCodeUnderTestCouldWriteIsRefused (7 subtests),
  TestAHandedHouseValuesMachineFilesWithoutDpkg, TestAPlaceValuesEveryPathItCoversFirst,
  TestAProductNamingItsMachineFailsItsBuild, TestTheHousesBoxAndMomentNeverValueAKey.
- The same five, run alone at 9bc0fc52 (32-failing-at-head.log) and at main 2a032b3c (33-failing-at-main.log:
  internal/buildcache checked out from 2a032b3c, with the new test file removed): **main fails the same five, the same
  way.** After normalizing temp paths, hashes and durations, the two logs are identical (34-main-vs-head-failures.diff,
  diff exit=0).
- The messages are `run as root, the test's own file is root's`, `run as root, every directory is writable`,
  `its directory ... can be written by this process`, and relative_test.go:51
  `a tool in the home directory: the build gave <nil>, want a refusal`. Inferred, not proven: these come from this
  box running as root, which the input tests notice, and not from this change. Main fails them identically here.

## 4. The real sweep: TestProduct_HelperOracle, main 2a032b3c against 9bc0fc52

Chosen: `stage1/cohere/lint/helpers/oracles TestProduct_HelperOracle`. It is a GoBuildIn of the cohere module, so its
keying asks Tracked under cohere and cohere/TypeScript. product-order.json gives it no products to wait for. Two test
binaries were built with `go test -c` (40-build-binaries.log): one at 9bc0fc52, and one with only
internal/buildcache/buildcache.go set back to 2a032b3c (the change's diff touches nothing else that the binary
links). Both ran from the same checkout's package directory, so both swept the same cohere/TypeScript work tree.
Driver: scripts/sweep.sh, which runs `strace -f -c -e trace=newfstatat,lstat`, then `strace -f -e trace=execve` on a
second run.

The first run (head-first, at 9bc0fc52) was a cache miss that built the oracle: 349 s, 344,070 newfstatat. It includes
go build, so it compares with nothing. The warm runs (all cache hits, all exit=0, all PASS):

| Run | newfstatat calls (errors) | lstat | wall under strace -c | test's own time | git execs |
|---|---|---|---|---|---|
| main-warm-1 | 269,345 (2,166) | 0 | 8.08 s | 8.04 s | 9 |
| main-warm-2 | 269,345 (2,166) | 0 | 8.70 s | 8.67 s | 9 |
| head-warm-1 | 310,706 (2,166) | 0 | 11.35 s | 11.31 s | 9 |
| head-warm-2 | 310,706 (2,166) | 0 | 10.69 s | 10.66 s | 9 |

Unstraced wall, three each (wall-unstraced.log): main 0.785, 0.882, 0.771 s; head 0.843, 0.880, 0.937 s.

What git ran (`*.git-commands.txt`): exactly the same at main and at head, in every run. That is
`ls-files -s -v -z` and `diff-files --name-only -z` once each in /home/user/adamic, /home/user/adamic/cohere and
/home/user/adamic/cohere/TypeScript, plus 3 `git status --porcelain=2 -uno`.

### Observed

- In this product test, at main, cohere/TypeScript is already listed **once**. The race this change fixes (several
  askers in one process missing together) does not happen in this test on this box, so this test cannot show the
  saving.
- The head does **41,361 more newfstatat** than main, deterministically (the same counts in both runs), and is slower
  under strace (about 10.7 to 11.3 s against 8.1 to 8.7 s). Unstraced, the gap is within noise to slightly slower
  (head's mean 0.887 s against main's 0.813 s, three runs each).
- Where the extra stats come from (`strace -f -e trace=newfstatat`, raw in `*.paths.raw.gz`, per-path difference in
  sweep/path-diff.txt, net +45,241 in that run): `AT_SYMLINK_NOFOLLOW` stats of the ancestors of every directory asked.
  `/home`, `/home/user`, `/home/user/adamic` and `/home/user/adamic/cohere` each run about 6,860 times at head against
  about 50 at main. `cohere/internal` (+4,141), `cohere/internal/lint` (+3,080), `cohere/TypeScript` (+1,164) and so on
  appear too.

### Inferred

- The extra stats are `repositoryKey`'s `filepath.EvalSymlinks`, which lstats every component of the path. It runs on
  every `repositoryAt` call, a cache hit included, before the map is looked at. Tracked calls repositoryAt for each
  ancestor of each name it is asked about, about 6,860 calls in this test. At main a hit was a bare `sync.Map` Load.
  One way out would be to look up the directory as given first, and resolve only on a miss (or keep the resolved key
  per spelling). That is a suggestion only. Nothing here changes it.
- The second commit (1f6fc757) does not touch repositoryKey, so this cost should remain there. It was not measured.
- A product test that does race on one checkout (Builder's trace on #npd568j saw three listings in one test) should
  show the saving. None was found and measured here.

## 5. Second commit, on Kirk's instruction: 1f6fc7575ec69c14937574de09a51b1c8f2709fc

`buildcache: an asker waiting on a repository's reading wakes to its answer, and a reading that panics frees its
waiters (#npd568j)`, with `Co-Authored-By: Ahra <ahra@ahra.ai>`. It was pushed to `compiler/repository-once`
(9bc0fc52..1f6fc757) with a plain push.

- The reading moved into `readRepositoryRun`, whose free (`close(run.done)`) is deferred. On a panic it marks the run
  failed and panicked, drops it from repositories (before the free, so a re-asker finds it gone), frees the waiters,
  and lets the panic go on.
- Beyond the review's ask: a waiter that wakes to a panicked run asks again instead of taking its nil files, which
  Tracked would read as everything tracked, the same silent wrong answer as the must. Mutant `take-panicked` below
  pins it.
- `waitingOnRepository` is a test hook, called by an asker just before it parks on another's reading.
- TestEveryWaiterWakesToTheReadingsAnswer: the listing seam blocks until the hook has counted 15 parked askers (out
  of 16), then answers. Every asker must get non-nil files holding the listing's one file, the same pointer for all,
  and exactly one listing.
- TestAReadingThatPanicsFreesItsWaiters: the first listing waits until one asker is parked, then panics. Both askers
  recover. Exactly one must have panicked with the listing's own value, and the other must get the second listing's
  answer. A later asker gets the same kept answer, with exactly 2 readings. Every wait is bounded by 30 s.

Runs at 1f6fc757, in this checkout (the go.work and cohere submodule are the ones above):

- The five repository tests, `-count=1 -v -race` (50-fix-tests.log): exit=0, all PASS.
- `gofmt -l internal/buildcache` printed nothing, and `go vet ./internal/buildcache` exit=0 (60-fix-gofmt-vet.log).
- `go test ./internal/buildcache -count=1 -timeout 30m` (61-fix-package.log): exit=1. The failing set is exactly
  30-failing.txt's (the same five root failures as at 9bc0fc52 and main; `diff` empty). Nothing new failed.

Mutants on 1f6fc757 (scripts/mutate2.py), each alone, with `git diff` empty after restore (`restored-clean` in each log):

| Mutant | Log | Result |
|---|---|---|
| early-free (the review's must): `close(run.done)` moved before `readRepository` | 5x-fix-mutant-early-free.log (`-race`) | exit=1. TestEveryWaiterWakesToTheReadingsAnswer: `asker 0 of 16 got <nil>`. TestAReadingThatPanicsFreesItsWaiters: `got <nil>`. The race detector also flags TestOneCheckoutIsListedOnceHoweverManyAsk |
| early-free without `-race` | 5x-fix-mutant-early-free-norace.log | exit=1. The three old tests PASS. TestEveryWaiterWakesToTheReadingsAnswer FAILs by its own assertion (`asker 0 of 16 got <nil>, want the listing's one tracked file, a.txt`), so the new test catches it without the race detector |
| no-defer (the review's should): the free runs only on a normal return | 5x-fix-mutant-no-defer.log | exit=1. TestAReadingThatPanicsFreesItsWaiters FAILs after 30.00 s: `an asker of a reading that panicked never got an answer: its waiters were never freed`. All others PASS |
| take-panicked (the extra): a waiter returns a panicked run's files | 5x-fix-mutant-take-panicked.log | exit=1. TestAReadingThatPanicsFreesItsWaiters: `the asker waiting on a reading that panicked got <nil>`. All others PASS |

## Not covered

- No Loom verdict (`ahra loom check`) and no products-first whole suite. Only internal/buildcache and the one
  product test ran.
- The sweep compared main against 9bc0fc52, not 1f6fc757.
- No product test was found here that races on cohere/TypeScript, so the change's intended saving was not observed.
