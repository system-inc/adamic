Built: slice 2 re-cut as one commit on current main, preserving callback lowering, Weak checks, Error semantics and self-comparison.
Commits: base d79302e50c022c997c00f7c68ddb7475f9ee65fc; source ba754087868e26dae54500672c00b2c35b2b0412; delivery SHA is in the final response and git log.
Commands and outputs: focused fixtures, guards, reader checks and changed-package vet pass; counts add 46 owned rows and change no existing row.
Mutants: all 17 original semantic mutants, four extended probes and one structural-only inline check were caught; every catcher is below.
Not covered: full package suites, the full gate, WASI and macOS. No PR or Loom submission was made.

No textual conflicts occurred. units.go auto-merged main's ae1d1c70 clang jobserver support with the source's inline-helper placement. Main's counts work first added digests (39b99b53, 9810993a), then 1f07ce2d moved measurements into keyed products and made counts.md counts-only again; ac8ffaac and 9f8ce955 improved shard preparation. This fetched tip therefore has no digest column. Conservative choice: preserve current main's table and product API, rather than restore an obsolete schema. The two Weak registrations now return main's preparation tuple. No production lowering or runtime code needed manual adaptation.

Main's table was taken before exactly one Linux -update-counts run, clang 20.1.8. counts-attribution.json lists all 46 additions: 35 callback-widening, three miscompile-2a, five Weak and three self-compare rows. No existing row changed or disappeared. Runtime +/- lines exactly match the source (runtime-comparison.json): adamic.h +2, exceptions.c +15/-2, object.c +2; no differences.

Guards: SortedMap, exact runtime witnesses, literal native agreement registration and parallel-test checks pass. No new compiler-consuming Go package or runtime feature set was added, so neither dependency nor runtime-set declarations needed additions. Three serial reader key-discovery leaves gained their own Not parallel explanation, retaining the source's serial preparation. Native field-layout and placeless split-build guards pass. Reader products retain relative source/export names and build through main's buildcache. TestCallTargetReaders ran with -timeout=90s; vet covered all seven changed Go packages.

Every command used the sourced setup environment, GOPROXY=https://proxy.golang.org|direct, GOMAXPROCS=4, and four-CPU affinity. nproc=5, cgroup cpu.max=400000 100000. Setup completed in 496.716 seconds; all timing lines are in setup.log. A first guard attempt before the nested submodule checkout completed could not load tsc/go.mod and ran no test. The later actual guard run passed. A transient exec transport disconnect occurred during counts; the original bounded process completed successfully and was not repeated.

The source's test splits are retained. Every completed affected leaf remains below 60 seconds; individual timings are in *-seconds.json. Runtime archives were prepared one product group at a time. Test output was always written to logs. Worker runs used focused selectors, not entire package suites. Setup compiled all repository packages. No code was copied from cohere.

## Commands

| Check | Exit | Wall seconds | Command |
|---|---:|---:|---|
| parallel | 0 | 50.117 | `go test ./cmd/adamic-gate -run ^TestEveryTestIsParallelOrSaysWhy$ -count=1 -json -timeout=90s -parallel=4` |
| maps | 0 | 159.56 | `go test ./internal/ir -run ^TestEveryMapAProgramHoldsIsSorted$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-sets | 0 | 9.631 | `go test ./internal/oracle -run ^TestRuntimeSetsAreEachUsed$ -count=1 -json -timeout=90s -parallel=4` |
| agreement-registry | 0 | 13.559 | `go test ./internal/lower -run ^TestNativeAgreementRowsNameRegisteredPrograms$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Sanitized0 | 0 | 12.992 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesSanitized0$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Sanitized1 | 0 | 12.372 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesSanitized1$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Sanitized2 | 0 | 12.443 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesSanitized2$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Release0 | 0 | 9.294 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesRelease0$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Release1 | 0 | 10.476 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesRelease1$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Release2 | 0 | 10.077 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesRelease2$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Counted0 | 0 | 9.815 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesCounted0$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Counted1 | 0 | 9.707 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesCounted1$ -count=1 -json -timeout=90s -parallel=4` |
| runtime-Counted2 | 0 | 9.831 | `go test ./internal/oracle -run ^TestProduct_NativeRuntimesCounted2$ -count=1 -json -timeout=90s -parallel=4` |
| oracle | 0 | 15.326 | `go test ./internal/oracle -run ^Test(CallbackWidening\|WeakDifferential\|Miscompile2A\|SelfCompare) -count=1 -json -timeout=90s -parallel=4` |
| native | 0 | 14.61 | `go test ./internal/native -run ^Test(ErrorUndefinedMessageHasNoOwnProperty\|SplitSelfCompareAgreesWithNode)$ -count=1 -json -timeout=90s -parallel=4` |
| ir | 0 | 2.791 | `go test ./internal/ir -run ^Test(CallbackAdapterTargets\|ClosureTargetsBoundOnlyProvenValues)$ -count=1 -json -timeout=90s -parallel=4` |
| flow | 0 | 18.391 | `go test ./internal/flow -run ^TestCallbackWidening -count=1 -json -timeout=90s -parallel=4` |
| fresh | 0 | 7.66 | `go test ./internal/fresh -run ^TestCallbackWidening -count=1 -json -timeout=90s -parallel=4` |
| fixtures | 0 | 32.816 | `go test ./stage3/fixtures -run ^Test(FixturesSelfCompare\|FixtureDirectoriesHaveTopLevelTests) -count=1 -json -timeout=90s -parallel=4` |
| reader-key-Native | 0 | 4.391 | `go test ./internal/ir -run ^TestCallTargetExportInputsNative$ -count=1 -json -timeout=90s -parallel=4` |
| reader-key-Lower | 0 | 4.245 | `go test ./internal/ir -run ^TestCallTargetExportInputsLower$ -count=1 -json -timeout=90s -parallel=4` |
| reader-key-Fresh | 0 | 4.193 | `go test ./internal/ir -run ^TestCallTargetExportInputsFresh$ -count=1 -json -timeout=90s -parallel=4` |
| reader-key-Flow | 0 | 3.935 | `go test ./internal/ir -run ^TestCallTargetExportInputsFlow$ -count=1 -json -timeout=90s -parallel=4` |
| reader-archive-Native | 0 | 5.957 | `go test ./internal/ir -run ^TestProduct_CallTargetReaderArchivesNative$ -count=1 -json -timeout=90s -parallel=4` |
| reader-archive-Lower | 0 | 5.967 | `go test ./internal/ir -run ^TestProduct_CallTargetReaderArchivesLower$ -count=1 -json -timeout=90s -parallel=4` |
| reader-archive-Fresh | 0 | 5.922 | `go test ./internal/ir -run ^TestProduct_CallTargetReaderArchivesFresh$ -count=1 -json -timeout=90s -parallel=4` |
| reader-archive-Flow | 0 | 5.92 | `go test ./internal/ir -run ^TestProduct_CallTargetReaderArchivesFlow$ -count=1 -json -timeout=90s -parallel=4` |
| readers | 0 | 5.983 | `go test ./internal/ir -run ^TestCallTargetReaders -count=1 -json -timeout=90s -parallel=4` |
| native-guards | 0 | 3.784 | `go test ./internal/native -run ^Test(RuntimeFieldLayoutsAreIncluded\|SplitBuildsNameNoPlaceOfTheirInputs)$ -count=1 -json -timeout=90s -parallel=4` |
| vet | 0 | 21.855 | `go vet ./internal/lower ./internal/native ./internal/ir ./internal/flow ./internal/fresh ./internal/oracle ./stage3/fixtures` |
| counts | 0 | 53.129 | `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -v -timeout=600s -args -update-counts` |
| counts-table | 0 | 2.84 | `go test ./internal/oracle -run ^TestCountsAreRecordedTable$ -count=1 -json -timeout=90s -parallel=4` |

## Mutants

| Mutant | Caught | Seconds | Expected catcher |
|---|---|---:|---|
| error-spread | True | 17.953 | want path-bearing stop |
| maybe-setter | True | 17.041 | want path-bearing stop |
| long-name | True | 17.198 | want path-bearing stop |
| computed-name | True | 16.18 | Unhandled case in Node.Text |
| constructor-capture | True | 16.023 | want path-bearing stop |
| optional-error | True | 30.002 | member access within null pointer |
| error-own-message | True | 14.529 | native , Node  |
| method-views | True | 16.327 | both backends unexpectedly admit the pinned view NotYet, release: exit=0 stdout="t1 none, release: stdout differs, AddressSanitizer |
| union-liveness | True | 15.76 | release: exit codes differ, sanitized: exit codes differ, runtime error: member access within null pointer |
| weak-callback | True | 15.373 | both backends must stop at |
| adaptation-omitted | True | 16.679 | native |
| nullable-token-omitted | True | 15.269 | exit |
| unrepresented-parameter-admitted | True | 17.354 | expected unrepresented parameter refusal |
| adapter-target-omitted | True | 13.554 | Unknown:true |
| split-numeric-fold | True | 8.624 | native: <nil>, Node  |
| refusal-revert | True | 17.517 | --- FAIL: |
| weak-callback-isolated | True | 17.188 | both backends unexpectedly admit the pinned callback NotYet, AddressSanitizer: heap-buffer-overflow |
| weak-counter-registration-omitted | True | 7.677 | want the row of |
| static-inline-definition | True | 8.618 | inline function 'adamic_unit_adamic_same_number' is not defined |
| fixture-owner-omitted | True | 6.35374905599997 | has no top-level test |
| reader-allowlist-omitted | True | 8.235816606000071 | unapproved call-target read |

Mutant commands and logs are retained in this review directory. The inline-definition mutant is structural only, caught by clang and not credited as a semantic kill. Mutant Go sources use .go.txt. TestSelfCompareConstantMutant built and ran its NaN mutant successfully and observed the stdout disagreement with Node; the test passes only when that semantic mutant is killed.
