# Step 06b: Program runtime core

This unit carries the runtime core from `d23a4fdc` onto `63e0056f`.
Compiler membership and lowering belong to contract 06a. No type classification,
cycle census, lowering, emission or graph-region operations are included here.

`Options.ProgramRegion` or `ADAMIC_PROGRAM_REGION=1` enables
`ADAMIC_PROGRAM_REGION`. `ValidateOptions` refuses non-native targets and request
exports with `native: Program regions require a one-shot native CLI`.

`adamic_program_adopt_owned(value, size)` takes a fresh, unaliased counted
allocation and returns its replacement address. Callers must use that returned
address before publishing any aliases, interior-cell addresses, Weak handles or
canonical caches. Adoption preserves child references without retaining or
releasing them. Moving the storage is one logical allocation, not a second
allocation. The caller supplies the complete allocation size; objects use
`adamic_object_size(count)`, which includes values, insertion ranks, readiness
bytes and representation bytes. Environment owners are rebased to the moved
`adamic_environment`, following this stack's cell-owner representation.

The two-word prefix holds an unused region pointer and the intrusive member-list
link. It is the only graph-region machinery carried. Program membership uses
slab bit 28; bit 29 is reserved for graphs, bit 30 marks statement-region values,
and bit 31 marks shared values. With the switch enabled, chunk numbers plus one
stay strictly below bit 28. Without it, the existing bit-29 bound remains.

Members have zero header references. Inline and slow retain/release paths do
nothing to their ownership; counting builds still record logical calls.
`adamic_program_region_end()` first forgets member and environment-cell Weak
handles, then releases all counted children while every member remains alive,
then frees all member storage. Members count as region values rather than
ordinary frees. Calling end again on an empty region is harmless. `adamic_heap_end`
ends the Program before destroying allocator chunks.

The member registry is main-thread only. Compiler contract 06a must exclude
adoption in task bodies, as well as publication to tasks. `adamic_share` refuses
members, including environment-cell owner redirects, with
`Program region members cannot cross into parallel work`. This unit does not
implement moving or sharing a Program region, and does not establish compiler
membership or a lifetime for long-lived services.

The C harness starts with counted objects, an environment and a Map, adopts them,
and holds a real counted string and object as children. It checks object-layout
tails, cyclic member links, canonical closure cleanup, no-op member ownership,
Weak invalidation, explicit end and process-exit end. Its payload observations
are compared with Node running `oracle.a`; ownership checks use native assertions,
ASan, UBSan and `internal/leakcheck`. The expected counted line has 11 allocations,
7 ordinary frees and 4 region values. Mutants release a member early, omit member
or cell Weak invalidation, and copy an object without its insertion-order tail.

## Validation on Linux

Toolchain: Go 1.27.1, clang 20.1.8, Node v24.19.0, x86-64 Linux;
`nproc` reports 5 and cgroup CPU quota is 4 cores. The first setup filled the disk
while warming the Go cache. I stopped that setup, ran `go clean -cache`, and
reran `bash cloud/setup.sh` successfully. Its cumulative timing lines were: Go
0.039s, Node 0.062s, Markdown dependencies 0.132s, submodules 0.164s, clang
0.415s, build 666.796s, test binaries deferred 668.911s, warm cache 668.963s,
and done 669.803s. Environment: `/workspace/adamic-tools/env.sh`.
`npm ci --prefix stage3/api --ignore-scripts` installed the locked host types
before the final counts check.

All test output was redirected to log files. These commands ran:

```sh
go build ./... > /tmp/step06-core-build.log 2>&1
go vet ./... > /tmp/step06-core-vet-final.log 2>&1
gofmt -l cmd internal > /tmp/step06-core-gofmt.log
go test ./internal/native -run '^TestProgramRegion' -count=1 -timeout 20m -v > /tmp/step06-core-final-focused.log 2>&1
go test ./internal/native -run 'Test(ProgramRegion|RuntimeReleasePaths|ReleaseSharedValueAndUnsafeMutant|ClosureConvention|RuntimeStatics)' -count=1 -timeout 20m -v > /tmp/step06-core-runtime-checks.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/.*(closures|weak|cycles_)' -count=1 -timeout 20m -v > /tmp/step06-core-oracle-selected.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 20m > /tmp/step06-core-counts-final.log 2>&1
```

Build and vet passed. The focused runtime suite passed (252.217s), including
release paths, closure calling conventions, runtime-static auditing, all twelve
ThreadSanitizer protection mutants, parallel probes and signal/exit probes.
The final Program tests passed (3.144s), including
malloc and slabs under ASan/UBSan, the release build, the common leak helper,
Node output comparisons, sharing refusals, fresh-adoption refusals and the
wasm32-wasi refusal. Explicit end and process-exit end both reported:

```text
adamic: counts: allocations 11 frees 7 retains 14 releases 22 peak 11 regions 4
```

| Mutant | Check that caught it |
| --- | --- |
| Release destroys a member | ASan heap-use-after-free |
| End skips member Weak invalidation | `adamic_weak_target(weak) == NULL` assertion |
| End skips interior-cell Weak invalidation | `adamic_weak_target(cell_weak) == NULL` assertion |
| Adoption omits the insertion-order tail | ASan heap-buffer-overflow |

The selected oracle ran 16 fixtures: 14 passed and the two
`reuse_weak_after_reuse.a` and `reuse_weak_during_spread.a` failed to compile.
The final counts check failed in 36 fixtures, all with the same compiler/runtime
layout mismatch: generated `adamic_object_present(object, cache.index)` refers
to a nonexistent `adamic_slot_cache.index`; the runtime has `packed` instead.
The cache definition and the relevant lines in `emit_objects.go` and `reuse.go`
are unchanged from `63e0056f`. This establishes the unchanged mismatch in the
base sources; I did not run a separate complete gate on the base commit.
No counts table was regenerated, and a complete unchanged-counts claim cannot
be made because the test stops before comparing its table when fixtures fail.

Repository-wide formatting lists `internal/flow/corpus_units_test.go` and
`internal/fresh/corpus_units_test.go`, both unchanged from the base. All changed
Go files are formatted. I also started the whole native package, then stopped
that run while it was processing the unrelated exhaustive decoder corpus;
it has no complete result. The complete repository gate was not run.
