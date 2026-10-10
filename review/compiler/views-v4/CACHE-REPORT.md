Built step (1): weak callable-view adapter interning on main's runtime.
Commit: see the compiler/views-v4 commit carrying this report.
Tests: C cache 7.687s; oracle 19.954s; lower 0.917s, native 7.178s, JavaScript 1.011s.
Mutants: skip interning, dying retain, skip recheck, release before removal, canonical factory, strong leak root all caught.
Pending: runtime conditions 3 and 4; adapter execution and language-level identity are later steps.

The global intrusive weak table uses a pthread mutex. Factories run outside the lock; insertion rechecks and releases losing candidates after unlocking. Lookup refuses count-zero adapters. Destruction removes the exact entry before releasing its underlying. Adapter factories must return fresh noncanonical counted closures; re-viewing unwraps the underlying. Hidden pointer links prevent the nonowning table from concealing leaks from LeakSanitizer.

Main's heap reference counts remain single-threaded. The dying-adapter test injects a reentrant lookup into the actual count-zero destruction window, repeats freeing/re-viewing 10,000 times, and catches the resurrecting mutant with ASan heap-use-after-free. The allocation test re-views 100,000 times: two allocations total, steady retains; bypassing interning gives 100,002 allocations. Factory reentrancy proves make runs outside the mutex and the winner is rechecked without requiring the unlanded pool. Positive fixtures pass ASan and LeakSanitizer; the intentional abandoned-adapter test requires a leak report. The canonical-factory negative exits before destroying the invalid combination and is not a leak-clean assertion.

C fixtures compile the actual closure, heap, weak and count runtime units; unrelated destructor stubs abort if reached. Full-runtime oracle and release-path regressions also passed. No area/runtime code was imported.

Pending tests carry the exact reason: awaits runtime's slice with the thread-safe heap, pool and region adoption (area/runtime). Each names its condition. Condition 3 covers the static audit documents/listing and pool-worker ThreadSanitizer re-view fixture; the table is already global and mutex-protected, but shared heap counts await that slice. Condition 4 covers graph/Program adoption refusing adapters and an adapter over a real Program-region underlying, runtime's slice 6c #c2zbg7a. Whichever lane lands second enables condition 3.

Commands (output in adjacent logs):
- GOPROXY='https://proxy.golang.org|direct' timeout 90 bash cloud/setup.sh: 40.604s; markdown 0.073s, clang 0.167s, go build 40.415s, cache warm 40.571s; nproc 5, CPU quota 4.
- timeout 90 go test ./internal/native -run '^TestViewAdapter' -count=1 -v -timeout 90s
- ADAMIC_GATE_UNCACHED=1 timeout 90 go test ./internal/oracle -run '^Test(ViewAdapterBaselineCounts|V4Direct|V4Escape)' -count=1 -v -timeout 90s
- timeout 90 go test ./internal/lower ./internal/native ./internal/javascript -run '^Test(ViewCallable.*|ViewCallables.*|PrepareViewCallableRead|CallableNamespace.*|RuntimeReleasePaths|ClosureConventionRuntimeFeaturesIgnoreLiterals)$' -count=1 -v -timeout 90s

New C leaves: Intern 2.93s, Bounded 4.24s, Recheck 3.86s, Dying 4.89s, Order 4.22s, Canonical 6.18s, WeakLeakRoot 3.43s. New count leaves: Closures 0.23s, Methods 0.18s, ReceiverRest 4.65s. Four runtime prerequisite leaves skip; the four existing compiler adapter-execution/identity leaves remain pending. No new oracle fixture rows; counts.md is unchanged.

Recorded counts before -> after (allocations, frees, retains, releases, peak, regions):
- closures.a: 58 | 58 | 48 | 84 | 26 | 0 -> identical.
- method_closures.a: 20 | 20 | 14 | 30 | 8 | 0 -> identical.
- closure_convention_receiver_rest.a: 11 | 11 | 12 | 23 | 7 | 0 -> identical.

WASI retains its single-thread lock-free build path; WASI was not tested in this unit. Lane checks and go vet outputs are saved alongside this report after committing.

Lane checks passed in 8.2s: gofmt/tools on 102 Go files, t.Parallel on 8 test packages, vet on 8 packages. Separate go vet ./internal/native ./internal/oracle passed with no output. Sanitizer-log trailing whitespace was stripped for git diff hygiene; diagnostic text is unchanged.
