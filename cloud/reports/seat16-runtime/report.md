Built: the ordered seat 16 runtime train on runtime/seat16-runtime, based on area/runtime 14504c7a.
Commits: thirteen merge commits listed below; cc69f085 adds the MapClear classification caught by the integration gate.
Commands and outputs: all builds pass; final uncached oracle ok 182.479s, flow ok 137.504s, counts ok 69.610s; vet and format empty.
Mutants: shared final-release guard, slab use after free, two fresh-cycle guards, documented count, two refusal inputs, uncaught separator, and missing MapClear classification were caught.
Not covered: the full repository test gate, macOS leaks, a separate WASI execution lane, or the queued batch 8 reference-slot optimization.

# Seat 16 runtime train, October 7, 2026

Base: origin/area/runtime `14504c7a4bf90595311eceb9d259e1a5a28f1a67`, fetched at the start of this train. Reference: origin/cloud/integrate-16 `c31cbeb`.
No rebase, history rewriting, PR, main push or area push. Every incoming SHA is an ancestor of the resulting branch.

## Ordered merges and resolutions

None of the thirteen requested SHAs was an ancestor of the fetched area tip. Therefore none was an ancestry no-op, including release-fast-path and slab-checks. The release fast-path implementation was already present; its merge preserved that implementation and added missing evidence and checks. The concurrency-containing ref was separate from this area's ancestry. Non-trivial choices are also explained in each merge commit.

| Incoming branch or coverage | Pinned SHA | Merge commit | Status and resolution |
| --- | --- | --- | --- |
| runtime/release-fast-path | `b9f22f0` | `6c12495e` | Merged; heap.c comment conflict: keep area implementation, import evidence and shared-release mutant. |
| release-fast-path coverage | `c0cc4d7` | `f6b13505` | Merged; heap.c comment conflict: keep area; counts: keep area table. |
| runtime/slab-checks | `220a915` | `b61bf78c` | Merged; clean Git merge; adapt newer map_hash_test.go from private slabs to exported Slabs, retain all area native options. |
| runtime/handover-throw | `be68139` | `2c322697` | Merged; oracle registry union; counts: keep area. |
| handover-throw coverage | `ac48654` | `61915120` | Merged; counts: keep area; fixture registry merges cleanly. |
| runtime/aside-throw | `3756e2f` | `db2aa3fe` | Merged; oracle registry union; counts: keep area. |
| runtime/memory-guard-fixtures | `0e3881f` | `58799a29` | Merged; oracle registry union including area constructor-capture guard; counts: keep area. |
| codex/memory-worked-examples | `bca0d5a` | `4cf15d6f` | Merged; oracle registry union; counts: keep area; remeasure worked-example rows and explanations, preserve historical C dates and new string-view policy. |
| runtime/fresh-probes | `2403dd3` | `92de9f11` | Merged; clean merge, preserve area proof. |
| codex/fresh-clobbered-loads | `dddb655` | `6b0456cd` | Merged; clean merge, tests and documentation only. |
| runtime/error-spread | `e73fd72` | `c19fa5ad` | Merged; keep NoReuse private/accessor rule plus incoming Error and view guards; oracle registry union. |
| runtime/uncaught-name | `bc75558` | `688d96cf` | Merged; oracle registry union; counts: keep area. |
| coverage/error-names | `69d3c87` | `616bc86f` | Merged; oracle registry union; flow: keep sixteen NotYet exclusions and area long-normalization exclusion; counts: keep area. |

The current area string views, lint release fixes, stdout edges, WASI hooks and borrow-chain coverage remain. The map-hash test needed the incoming exported `Options.Slabs` spelling; its malloc/slab cases remain. Incoming historical performance and coverage reports were preserved as historical evidence, not relabeled as measurements on this tree.

## Integration failure and correction

The final coverage merge exposed a previously unchecked `Map.clear()` in release_fast_graph.a. `TestEveryMutationIsInItsRange` reported `scratch$11` mutating at instruction 14, order 18, outside range [6, 11). MapClear returns void and was missing from `writes` in internal/flow/infer.go, so inference dropped its mutation effects. Commit `cc69f085` adds MapClear to that existing classification, without excluding the fixture or changing the check. An isolated overlay with that entry passed 771 mutations, zero declined and zero invalid ranges. The missing entry is the caught mutant. Range inference has no native-emission consumer today.

The worked-example docs retain dated historical C excerpts. Current measured list counts are 7/7/6/16/5/0 (six next-field reads are lent). Current string-example counts are 14/14/2/17/5/0 (two ASCII character results are immortal). Their prose and the slice-sharing description were updated while keeping the area's newer string-view policy.

## Counts

The complete area table was retained through all merges; no incoming historical row replaced it. Final TestCountsAreRecorded regenerated all rows on the integrated tree with an 8 MiB stack. Comparison against the saved area table: **430 original rows retained, all six numbers unchanged; 26 new rows; 456 total rows. No existing row rises or falls.** Thus the requested list of rising rows is empty. count-diff.json lists every added row and asserts that no original path disappeared. Panicking new fixtures intentionally record counts at the panic, as the table's header specifies; they do not finish cleanup.

## Verification commands

Every verification shell sourced /workspace/adamic-tools/env.sh, with TMPDIR=/tmp/adamic-gate (mode 1777). Setup log: Go 0s, clang 1s, Node 1s, submodules 1s, cache warm/done 133s; nproc=5, cgroup CPU quota=4, memory=17.6 GB. Tools: Go 1.27.1, clang 20.1.8, Node 24.19.0.

After each merge:

```sh
go build ./... > NN-build.log 2>&1
go test -count=1 -timeout 30m -skip '^TestCountsAreRecorded$' <touched packages> > NN-packages.log 2>&1
```

Only the recorded-count check was deferred during merges, pending one complete final regeneration. The full fresh package needed no skip. Package selection: native for 01; oracle for 02; native+oracle for 03; oracle for 04, 05, 06; native+oracle for 07; oracle for 08, 09; fresh for 10; lower+oracle for 11; native+oracle for 12; flow+oracle for 13. Gate 13 initially failed only the MapClear range check above; its oracle passed. The corrected full flow package and final uncached oracle were rerun.

```sh
gofmt -l cmd internal > final-format.log 2>&1
go vet ./... > final-vet.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts > final-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -parallel 4 -timeout 30m ./internal/oracle > final-oracle.log 2>&1
go test -count=1 -timeout 30m ./internal/flow > final-flow.log 2>&1
```

The first uncached oracle passed in 285.097s. A concurrent post-fix run hit the existing three-second wall deadline in the CPU-limit signal probe (all three children returned timeout status 124). Its log is retained. The isolated signal probe passed in 28.252s, with all three CPU-limit runs terminating with SIGXCPU. The complete final uncached run uses -parallel 4 to bound CPU contention; no assertion or timeout was changed. The corrected full flow package passed in 137.504s. The complete final uncached oracle passed in 182.479s, exit 0. Final count regeneration passed in 69.610s, exit 0; vet and gofmt logs are empty. The final run includes every oracle test, with no -run or -skip filter. Uncached oracle means source Node, generated JavaScript Node, sanitized malloc native, release native, sanitized slabs, leak checks for finishing programs, input/permission/output probes and recorded counts are all enabled. Native package gates also run the decode and normalization sweeps.

## Mutants actually run

All isolated Go overlays are retained in evidence. They change build inputs without changing repository files. No mutant was stopped by clang's warnings.

| Mutant | Observed catcher |
| --- | --- |
| Release a shared value when decrement leaves one reference (`> 1` instead of `!= 0`) | TestReleaseSharedValueAndUnsafeMutant: compiles, ASan heap-use-after-free on dynamic shared text; control agrees with Node and leak-checks clean. |
| Retain or read length after freeing a runtime string | TestFreedValuesAreCaughtWithSlabs: malloc heap-use-after-free and slab use-after-poison; untouched controls clean. |
| Drop callee.clobbers from deferred-write reach judgment | clobber_link_loop accepted; refusal test fails; LeakSanitizer reports 42,480 bytes in 800 allocations. |
| Drop before[o] from deferred-write reach judgment | escaped_before accepted; refusal test fails; LeakSanitizer reports 42,180 bytes in 800 allocations. |
| Change list's documented allocations from 7 to 8 in the document read by the check | TestMemoryExampleCountsMatchDocumentation/list.a fails with measured row 7. |
| Substitute accepted Weak-parent tree for the refused strong-parent input | TestMemoryExamplesRefused/tree fails: want cycle refusal with fix, got nil. |
| Substitute accepted function-declaration closure program for refused self-capture input | TestMemoryExamplesRefused/closure fails: want cycle refusal with fix, got nil. |
| Format a separator whenever message is nonempty, even when name is empty | Actual oracle stderr differs: Node `adamic: panic: changed 3`, native `adamic: panic: : changed 3`; both exit 70. |
| Omit MapClear from writes | Actual Node mutation trace fails outside [6,11), as above; restored entry passes. |

Representative commands are in evidence/commands.md; logs record exits and concrete catches. These prove the checks listed, not every incoming guard independently. Existing full oracle built-in mutants also execute during the package gate; the report does not invent separate manual mutation claims for each imported throw or view guard.

## Build configurations

Counts are heap/reference event counts, not instruction measurements. Their release counted builds use this exact flag vector, sanitizers off, no LTO or CPU override:

```sh
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -DADAMIC_COUNT -O2 -I <runtime-library-directory> -o <binary> <generated-main.c> <runtime-link-inputs> -lm
```

Uncounted release omits -DADAMIC_COUNT. Sanitized builds replace -O2 with -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all. The sanitized slab lane additionally sets -DADAMIC_SLABS; sanitized ordinary builds allocate through malloc. The code uses the existing RuntimeLinkFlags platform-specific link inputs; no added LTO, sanitizer or floating-point contraction flag is hidden in the numbers.

## Limits

The complete uncached **oracle** and all touched packages are the requested worker gate; the full repository test suite was not run. No macOS or separate WASI execution claim is made. This train does not include the queued destruction reference-slot optimization, new parser/scanner edits, character-read changes or compiler write-path changes. Logs are retained compressed as evidence/*.log.gz (the plain scratch logs are ignored by the repository); the count comparison is retained under evidence; scratch reproductions remain at /workspace/scratch/seat16-runtime.

## Package results after each merge

All thirteen go build ./... logs are empty and succeeded. Package outputs follow; gate 13 retains its caught flow failure, with the corrected result above.

```text
01: ok  	github.com/system-inc/adamic/internal/native	267.171s
02: ok  	github.com/system-inc/adamic/internal/oracle	126.855s
03: ok  	github.com/system-inc/adamic/internal/native	300.718s
ok  	github.com/system-inc/adamic/internal/oracle	190.842s
04: ok  	github.com/system-inc/adamic/internal/oracle	149.058s
05: ok  	github.com/system-inc/adamic/internal/oracle	147.697s
06: ok  	github.com/system-inc/adamic/internal/oracle	156.785s
07: ok  	github.com/system-inc/adamic/internal/native	294.344s
ok  	github.com/system-inc/adamic/internal/oracle	181.935s
08: ok  	github.com/system-inc/adamic/internal/oracle	151.832s
09: ok  	github.com/system-inc/adamic/internal/oracle	43.177s
10: ok  	github.com/system-inc/adamic/internal/fresh	13.516s
11: ok  	github.com/system-inc/adamic/internal/lower	18.105s
ok  	github.com/system-inc/adamic/internal/oracle	175.972s
12: ok  	github.com/system-inc/adamic/internal/native	316.667s
ok  	github.com/system-inc/adamic/internal/oracle	221.026s
13: --- FAIL: TestEveryMutationIsInItsRange (163.77s)
    --- FAIL: TestEveryMutationIsInItsRange/programs (0.02s)
        --- FAIL: TestEveryMutationIsInItsRange/programs/../oracle/testdata/release_fast_graph.a (0.55s)
            ranges_test.go:63: function 14 (removed): scratch$11 was mutated at instruction 14 (order 18), outside its range [6, 11)
    ranges_test.go:78: 107355 mutations checked, 0 on a range the pass left unset, 0 invalid ranges
FAIL
FAIL	github.com/system-inc/adamic/internal/flow	163.801s
ok  	github.com/system-inc/adamic/internal/oracle	352.151s
FAIL
```
