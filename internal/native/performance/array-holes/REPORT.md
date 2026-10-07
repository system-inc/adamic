Built: numeric length constructors, own holes, indexed access, length writes, callback methods, searches and joins; unported consumers are explicitly refused.
Commits: 1b86df6e (inventory), 41603d77 (probe), f7e1008a (operations), 91c5687b (proof integration), 85764567 (nullable boundary); all pushed on codex/library-array-holes.
Commands and outputs: scanner prints 4 on both backends; ten accepted fixtures and 31 refusal fixtures pass; final lower/fresh/IR gates and vet pass; Array test262 has zero disagreements.
Mutants: forEach, map, join, length bound and hole count fail Node comparison; admission, exception edge, mutation effect, freshness and nullable-diagnostic mutants fail their respective checks.
Uncovered: mixed-program dense-access performance guarantee, unported operations, arbitrary any[] element use, maximum-string/resource limits, actual macOS execution, and eight inherited test262 clang crashes.

## Base, setup and publication

Started from origin/library/merge-p2-trial at
c762b555e7c0b39b105e2d9208732a5a311b2e6f, not origin/main. The claims inventory
was the first push, before implementation. This resolves the prompt's conflict
between its claims-first requirement and its later probe-first ordering.
The unchanged scanner source came from
8d1b1141289fcd7340231b6950eacf6aa7f3a91f,
stage3/drivers/scanner/probes/array-length-constructor.a.
No pull request was opened.

Ran bash cloud/setup.sh, then sourced /workspace/adamic-tools/env.sh.
Setup succeeded; no module-download workaround was needed to recover a failed
setup. After integration's steering, Go commands used
GOPROXY='https://proxy.golang.org|direct'. Setup timing lines:

| Step | Elapsed seconds |
| --- | ---: |
| Go 1.27.1 ready | 0.058 |
| Node 24.19.0 ready | 0.060 |
| clang 20.1.8 ready | 0.444 |
| Markdown dependencies | 1.031 |
| Submodules | 9.617 |
| Go build | 179.929 |
| Cache | 180.047 |
| Done | 180.085 |

nproc: 5; CPU quota: 4. npm ci --prefix stage3/api succeeded: three packages
in two seconds. Logs are in evidence/.

## Representation and scope

Present numeric properties live in the existing ordered map. Missing entries
are holes; present undefined is a real entry. The sparse array's capacity field
counts holes, so filling an absent indexed slot decrements that count. Numeric
properties that are not Array indices do not grow length. Shrinking deletes
indexed properties and releases references, while keeping non-index properties.
Construction at 4294967295 is constant space. RangeError has a distinct nominal
class, with name RangeError and message Invalid array length; an ordinary Error
with that message does not pass instanceof RangeError. Throws propagate through
ordinary calls as well as local catches.

Dense arrays retain their existing element storage. Measured sizeof(adamic_array)
was 56 before and 64 after, with a 16-byte heap header in both; the allocator's
64-byte class still holds both. Dense-only programs emit their original indexed
access and assignment code. A compilation containing holes or a length write
uses dense/sparse accessors and conservatively treats every array as potentially
holey. This includes aliases, fields, parameters, returns and callbacks.
It can refuse unrelated dense consumers and adds a runtime representation branch
for dense accesses in such a mixed compilation. A universal no-slowdown guarantee
for those accesses has not been established.

The complete operation inventory is
[the claims file](../../../lower/library_array_holes_claims.md). Supported operations include
map, forEach, filter, some, every, reduce with an initializer, indexOf,
lastIndexOf, includes, scalar join/toString/String, and keys() loops. Sort,
reflection, JSON, spread, direct iteration and other unported consumers are named
NotYet, rather than treating holes as present undefined. Nullable element
representations and nullable aliases are refused. Bare any[] constructors may
observe length; contextual or explicit element types are required for element
operations. This does not prove the unannotated any[] arrays in the complete
TypeScript levenshtein implementation admissible.

## Node and proof checks

The ten accepted sources run against independent Node 24.19.0 source execution,
sanitized native, release native, and generated JavaScript. Native leak checks
are enabled; detect_leaks=1 is Linux-only in the existing harness. The new C file
defines _DARWIN_C_SOURCE immediately after _POSIX_C_SOURCE. No existing macOS
fix was removed, but this worker ran Linux.

Operation fixtures cover all holes, start/middle/end holes and partly filled
arrays. They distinguish explicit undefined from absence, map output presence,
callback indices, resize mutation, owned strings/objects, boolean slots, maximum
length, invalid lengths, and exception propagation. Thirty-one rejected sources
also run independently on Node; thirty have named lowerer refusals and the
console array argument has the existing string-only type refusal.

Boundary fixtures are in testdata/array-holes-boundaries, so the generic flow
tracer does not stringify billions of absent slots. That tracer uses map/join
to print every tracked array; its own resource limit can otherwise throw inside
the boundary program. Boundary sources remain registered in the ordinary oracle,
including sanitizer and leak checks. Rejected sources are in
testdata/array-holes-refused, outside the accepted-source flow globs.

| Command, with output redirected to a log | Result |
| --- | --- |
| go test ./internal/lower ./internal/oracle -run TestArrayHoles -count=1 -timeout 10m | pass; lower 0.306s, oracle 2.803s |
| go test ./internal/lower ./internal/fresh ./internal/ir -count=1 -timeout 10m | pass; 39.331s / 47.668s / 14.081s |
| go test ./internal/flow -run TestEveryFunctionIsInSingleAssignment -count=1 -timeout 10m | pass; 30.340s; SSA over all accepted root fixtures |
| go test ./internal/flow -run 'TestEveryMutationIsInItsRange/programs/../oracle/testdata/library_array_holes' -count=1 -timeout 10m | pass; 0.302s |
| go test ./internal/flow -run 'TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/library_array_holes' -count=1 -timeout 10m | pass; 0.320s |
| go test ./internal/flow -run 'TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/library_array_holes' -count=1 -timeout 10m | pass; 0.330s |
| Catch-specific range/liveness reruns after adding throwing callees | pass; 0.109s / 0.100s |
| go test ./internal/oracle -run 'TestArrayHoles\|TestCountsAreRecorded' -count=1 -timeout 10m -args -update-counts | pass; 41.827s |
| go vet ./... | pass, empty output |
| git diff --check | pass |

The initial broad command was:

go test ./internal/lower ./internal/native ./internal/javascript ./internal/ir ./internal/flow ./internal/fresh ./internal/oracle -count=1 -timeout 20m

Native passed in 325.260s, lower in 87.463s, and IR in 16.912s; JavaScript has
no standalone package tests. That run exposed new-node freshness/exception/effect
integration and refused-fixture layout issues, which were corrected and checked
above. The complete broad gate is not reported green. Its remaining unrelated
failures were a three-minute long-regex deadline under load, process pipe tests
expecting raw Node output loss, environment-dependent allocation rows, and the
existing process_bad_code flow path. Both process-pipe failures and the
process_bad_code flow failure reproduced on the pinned base; their logs are
included. The regex deadline was not separately reproduced on the base.

Only the Array fixture allocation rows were retained in counts.md. The measured
process_observations and node_fs_directory_system rows depend on this machine's
process/directory contents and were restored to the base values. The update run
passed before this restoration; a whole-machine count sweep is therefore not
claimed to match the committed table on this worker.

## Array test262 before and after

Same test262 commit: c8c798898646638cd0c24879f8e0374e847e7d74.
Command on both checkouts:

go run ./cmd/adamic-test262 -test262 /tmp/array-holes-test262 -adapt -jobs 4 built-ins/Array

| Checkout | Pass | Fail / disagreements | Refused | Crashed | Skipped | Total |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| c762b55 before | 125 | 0 | 2339 | 8 | 610 | 3082 |
| Final implementation after | 126 | 0 | 2338 | 8 | 610 | 3082 |

Both complete per-directory tables and diagnostics are in evidence/. All eight
crashes are the same pre-existing clang failures: indexOf 15.4.4.14-5-{10,11,31,32}
and lastIndexOf 15.4.4.15-5-{10,11,31,32}. These are not counted as agreements.

## Mutants actually run and restored

| Mutation | Check that failed |
| --- | --- |
| forEach visits missing slots as undefined | Node stdout differs in callbacks fixture; clean compilation and execution |
| map creates present undefined for missing input slots | Node stdout differs; mapped forEach count becomes five instead of zero/one |
| join's hole text becomes undefined | Node stdout differs for empty slots |
| Valid bound changes from >4294967295 to >=4294967295 | Node boundary/RangeError output differs at the largest valid length |
| Sparse write omits hole-count decrement | Node indexed read/output differs; filled slots disappear behind the all-holes shortcut |
| Disable conservative hole admission guard | TestArrayHolesAliasRefusals admits slice, and fails |
| Remove constructor/length-write exception successor | TestArrayHolesExceptionEdges fails |
| Remove length write from mutation-effect classification | Node trace observes catch fixture mutation outside the inferred range |
| Model constructor as unknown in freshness proof | TestEveryWriteIsRecordedAndKnown reports unknown nodes |
| Remove explicit nullable checks | Nullable diagnostic test fails because the constructor falls through to the separate unsupported-slot refusal; this is a diagnostic kill, not a runtime disagreement |

The five requested semantic mutants failed through Node comparison, not a
compiler/sanitizer crash. Full outputs are in evidence/mutant-*.log.gz.

## Dense hot loop

hot-loop.a performs 80 million dense indexed reads. Both binaries and independent
Node produce 680000000. The before and final emitted C files are byte-for-byte
identical; SHA-256 dc2e49f6ce7e145e8c1f10f2ff67cc9eeb73812bdccb6767e230fbeceae35270.
measure.py rebuilds both revisions and alternates execution order, discards two
warmup pairs and records fifteen measured pairs.

| Run | Before median seconds | After median seconds | After / before |
| --- | ---: | ---: | ---: |
| During heavy package checks | 0.241485 | 0.250030 | 1.03538 |
| After heavy native/oracle checks | 0.079186 | 0.081361 | 1.02746 |
| Final standalone reproducible run | 0.077952 | 0.077743 | 0.99732 |

All samples are included. Timing variability prevents a universal performance
claim; the final measurement found no dense-only regression, and identical
emitted C demonstrates the unchanged hot access path. Dense arrays in programs
that also contain holes still pay a representation branch. Sparse join currently
materializes string parts across length; gigantic join/map/reduce resource and
maximum-string errors were not exhaustively covered. Only construction, length
and indexed read/write were exercised at the maximum Array length.

Reproduce after sourcing the pinned toolchain and exporting GOPROXY:

python3 internal/native/performance/array-holes/measure.py --before /tmp/array-holes-base --after /workspace/adamic

## Every shared file touched

- internal/lower/assignments.go: length-write hook.
- internal/lower/exceptions.go: throwing expression propagation and admission hook.
- internal/lower/expression.go: constructor hook.
- internal/lower/object.go: constructor and RangeError field handling.
- internal/native/element_borrow.go: sparse indexed borrowing.
- internal/native/emit_arrays.go: callback/reduction presence checks.
- internal/native/emit_expressions.go: constructors, map, join, length and RangeError identity hooks.
- internal/native/emit_slots.go: slot presence lookup.
- internal/native/emit_statements.go: indexed store hook.
- internal/native/library_array.go: search dispatch.
- internal/native/runtime/adamic.h: sparse field and runtime declarations.
- internal/native/runtime/array.c: initialize sparse pointer.
- internal/native/runtime/heap.c: sparse map ownership destruction.
- internal/javascript/javascript.go: real constructors, ordinary index writes, hole-preserving map and length assignment.
- internal/flow/build.go: two throwing-expression exception successors.
- internal/flow/infer.go: one-line length-write mutation classification.
- internal/fresh/library_language.go: new Array-node proof hook.
- internal/oracle/counts.md: only Array fixture allocation rows.

New implementation is in lower/library_array_holes.go, native/library_array_holes.go,
native/runtime/array_holes.c, ir/array_holes.go, javascript/library_array_holes.go
and fresh/array_holes.go, with dedicated tests, claims, fixtures and evidence.
No changes to internal/native/emit.go, internal/lower/lower.go,
internal/native/native.go or internal/oracle/oracle_test.go; no one-line hooks
were needed in those four files.

## p2b merge and wasmtime milestone

Merged origin/library/merge-p2b 047e857207bee9aa62dbd22204445a0ab59d4319
with both parent histories retained in 14b3d1bc2e053387f3b694676e038c25ce34ef6f.
Conflict resolution retained all ten Array allocation rows, took integration's
unrelated host counts, and used integration's stage1 corpus/parser changes.
This assumes integration's stage1 revisions supersede the older base copies.
Host guards were imported without parallel edits. The merge imports an
integration change to internal/native/native.go; this unit made no edits there.

Added TestArrayHolesWasmtime in the existing unit test file
internal/oracle/array_holes_test.go. It builds each real WASI command and compares
stdout, stderr and exit code with independently executed Node source. It does
not accept target refusals, traps or skipped builds as agreements.

ADAMIC_ORACLE_WASI=1 ADAMIC_WASMTIME=/workspace/adamic-tools/wasmtime/wasmtime
go test ./internal/oracle -run TestArrayHolesWasmtime -count=1 -v -timeout 10m
passed all ten fixtures in 6.413s. Engine: wasmtime 49.0.2
(3c8a3e79a, 2026-10-02). The guessed v38.0.0 URL returned 404; the GitHub latest
release API identified v49.0.2, which was installed and pinned for this run.

Toolchain refresh bash cloud/setup.sh --wasi-sdk passed in 31.027s;
WASI SDK ready at 0.269s, Go build 30.806s, cache 30.990s, nproc 5.
Go commands use GOPROXY=https://proxy.golang.org|direct. Node remains 24.19.0.

The operation/refusal rerun on the merged branch used:

go test ./internal/lower ./internal/oracle -run TestArrayHoles
-skip TestArrayHolesWasmtime -count=1 -v -timeout 10m

Lower passed in 0.812s; oracle passed in 36.192s. This includes the ten accepted
fixtures with sanitizers, native release, JavaScript, Linux LeakSanitizer and
31 independently executed Node refusal sources. The conservative alias and
nullable checks pass. No Array operation implementation changed in p2b.

## p2b Array test262 before and after

Both runs use test262 c8c798898646638cd0c24879f8e0374e847e7d74,
Node 24.19.0, and the same command:

go run ./cmd/adamic-test262 -test262 /tmp/array-holes-test262
-adapt -jobs 4 built-ins/Array

The before checkout is exactly 047e857 in /tmp/array-holes-p2b-base;
the after checkout contains merge 14b3d1bc and the unchanged Array implementation.

| Checkout | Pass | Fail / disagreements | Refused | Crashed | Skipped | Total |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 047e857 before | 125 | 0 | 2339 | 8 | 610 | 3082 |
| Merged Array implementation after | 126 | 0 | 2338 | 8 | 610 | 3082 |

Both complete per-directory tables are committed in evidence/p2b-test262-*.log.gz.
The eight crashes remain the inherited indexOf/lastIndexOf clang failures listed
above, not agreements. No runtime failure was reported as a disagreement.
Successful observations may use the existing content-addressed cache; these
worker runs do not claim the uncached integration gate.

## p2b requested mutants and uncached final gate

Added reproducible mutants.py. It changes one source at a time, writes each run
to its own log, requires a Node stdout disagreement, rejects compiler/sanitizer
failures as semantic kills, and restores the original bytes in finally.

| Requested mutant | Node check that killed it |
| --- | --- |
| (a) forEach visits missing slots as undefined | callback fixture stdout differs |
| (b) map makes missing slots present undefined | callback fixture mapped presence/count stdout differs |
| (c) join prints undefined for a hole | callback fixture joined text stdout differs |
| (d) upper RangeError bound is off by one | range fixture maximum-length stdout differs |
| (e) write does not decrement the hole count | callback fixture filled-slot observations stdout differ |

All five compiled and ran without a compiler or sanitizer failure. An initial
narrow (e) run used the scanner source, which prints only length and did not
kill this mutation. The callback fixture killed it on rerun; the reproducible
driver now selects that fixture. The initial driver log is preserved rather
than claiming every fixture distinguishes every mutation. Every source was
restored before the final checks.

Also added TestArrayHolesWasmtimeRunnerMutants to the unit's oracle test file.
Its real Wasm control prints the scanner's expected 4; changing output to 5
is caught as stdout differs, and returning 23 is caught as exit codes differ.
This proves the requested runner's comparison can fail independently of the
Array implementation mutants.

Final command, uncached:

ADAMIC_GATE_UNCACHED=1 ADAMIC_ORACLE_WASI=1
ADAMIC_WASMTIME=/workspace/adamic-tools/wasmtime/wasmtime
go test ./internal/lower ./internal/oracle -run TestArrayHoles
-count=1 -v -timeout 10m

Passed: lower 0.761s; oracle 12.181s. Ten accepted sources run sanitized native,
release native, generated JavaScript and wasmtime; 31 refusal sources run on
independent Node; Linux LeakSanitizer is enabled. There were no skips in the
wasmtime Array leg. go vet ./... and git diff --check pass with empty logs.
The complete repository gate was not rerun; this is the focused unit gate.

Additional shared files changed in this follow-up:
internal/oracle/array_holes_test.go (wasmtime and runner-mutant checks),
internal/lower/library_array_holes_claims.md (validation evidence), and this
report. Merge conflict resolutions touched internal/oracle/counts.md,
stage1/cohere/lint/inventory/testdata/engine.go, and
stage1/typescript/parser/testdata/oracle.go. Integration's changes, including
internal/native/native.go and its host runtime guards, are preserved by the
merge; the Array worker made no hand edits to those guard files or native.go.
The other three prohibited shared files remain untouched by this unit.
