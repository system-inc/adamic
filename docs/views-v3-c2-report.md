# V3 c2 compatibility

The rehearsal and compiler/views-v3 source is b065fa5764b27478bb9b8acb1edef254731dd9a2. This separate tip merges cloud/land-stack-c2-c79c7572 at c79c75726375c937f0b18e18040e6769d916ed06. No c2 changes go back to the rehearsal.

The array case in share.c visits the sparse map when present. It visits dense elements only otherwise. In heap.c the same split belongs in static free_one, which releases the sparse map and then the ordinary properties and elements buffer. This tip does not add adamic_heap_free_children. Runtime owns that future graph-region integration.

The two array_holes.c mutable string headers are listed in runtime-statics.md with the requested classification, together with the runtime-file entry. The c2 audit also records the V1/V2/V3 contract files that arrived from the rehearsal. The incomplete adamic_array struct declaration is recorded explicitly because the storage scanner reports it; it creates no object.

The reference-array witness allocates length 4294967295, writes a freshly built string into slot 1, and captures the readonly alias in parallelMap. Node prints `4294967295:hole:ok:ok`. Its source reference constructor still stops with `NotYet holey Array of this element representation`. The test first asserts that refusal, then lowers a dense string-array control and replaces only that allocation with typed ir.ArrayHoles. It keeps every source element type. This is a runtime integration witness, not a new source-language admission. Both backends agree with Node; ASan/UBSan, release, ASan slabs and TSan worker variants run, with one thread and the default pool. Sanitized successful variants also run leak checks.

## Runtime changes to review

The earlier approved commits remain 04420be9a81f2f8c4f51b6e54001dd2c1ab711f8, ef762440f9def2672b8ecaecfeaae4e3fccbcfe9 and 3de13ddbb3698878864c3ae6bb0d6ecb8a7b4c2c. Their runtime hunks are listed in views-v3-checkpoint.md.

This c2 merge adds the sparse-map branch in share.c and keeps the approved sparse cleanup in heap.c's free_one. adamic.h, object.c and region.c conflict resolutions combine V2's tuple/readiness metadata with c2's dynamic object metadata and property order storage. The allocation size includes both metadata areas, and both heap and region producers initialize them. Array storage remains the approved V3 element_kind/sparse layout. No new array.c or array_holes.c behavior is added on c2 beyond the approved V3 changes.

The emitter resolves field metadata and presence using adamic_slot_index rather than the packed cache's removed index member, including reused object writes. This is c2 cache-layout compatibility. library_array_holes.go records reference storage as the shared physical reference kind for the runtime witness. The view_arrays.h include remains because generated array-read code needs its declarations.

## Validation

All commands source /workspace/adamic-tools/env.sh and use GOFLAGS=-buildvcs=false and GOMAXPROCS=2. Output is written to the named log files.

- Core: go build ./... and go vet ./internal/... pass at b065fa57; final-push-checks.log records both exits zero.
- Core: V3 fixtures and mutants pass; final-rebased-owned.log records the run. The a-check reports 26 files and zero failures in final-rebased-audit.log.
- Core: stage3/fixtures passes normally and with the platform guard lifted by a local Go overlay. final-census-stage3.log and final-census-stage3-linux.log record the results.
- Core: TestCountsAreRecorded -args -update-counts passes. Compared with 4fb447d5 there are 26 new V3 rows and no changed existing rows.
- Core: stage1/... -run 'Gap|Gaps|Probes' has 112 pass and 6 skip on both sides. No gap closes.
- Core: internal/ir, internal/javascript and internal/flow pass. internal/lower has the same TestNodeFSFileScratchOptionsBorrow failure as 4fb447d5. The full oracle with -timeout 30m has 3150 pass, 14 skip and the same inherited truthy-loops refusal failure and its parent. Against the base, exactly the 21 TestCheckedViewArrays subtests and their parent move from fail to pass; no other existing test changes result. The comparison JSON names every moved and added test. The initial default-timeout oracle runs expired; the 30m reruns completed.
- c2: go build ./... and go vet ./internal/... pass in /tmp/views-v3-c2-build-final.log and /tmp/views-v3-c2-vet-final.log.
- c2: go test ./internal/oracle -run 'TestCheckedViewArrays|TestCheckedViewArrayReadMutants|TestArray.*Mutant|TestArrayHoles|TestOriginalArrayWitnessNode' passes in /tmp/views-v3-c2-owned-final.log.
- c2: go test -timeout 30m ./internal/oracle -run TestCountsAreRecorded -args -update-counts passes in /tmp/views-v3-c2-counts-final.log. The new reference runtime witness has its counted row: 11 allocations, 11 frees, 14 retains, 19 releases, 9 peak live values and zero region values. No existing core row changes value; this is the only added row beyond the union of both parents.
- c2: go test ./internal/native -run TestRuntimeStaticsAreListed passes in /tmp/views-v3-c2-statics.log.
- c2: the gate a-check implementation reports 27 files and zero failures in /tmp/views-v3-c2-audit-final.log.
- c2: npm ci --prefix stage3/api succeeds. go test ./stage3/fixtures passes in /tmp/views-v3-c2-stage3-final.log; the local platform-guard overlay run passes in /tmp/views-v3-c2-stage3-linux-final.log. The guard is not changed in the commit.
- c2: the packed-slot reuse controls library_object_freeze, reuse, spread_undefined and call_targets_reuse pass against Node in both backends, including sanitized/slab/release native variants, in /tmp/views-v3-c2-reuse-controls.log.

The initial c2 stage3 run found three stale census entries for c2 admissions: taste/17_binder_flow.a, real/generic-optional-return/main.a and real/structural-method-statics/main.a. They are recorded as Compiles only after the Node/native comparison passed. The initial counts run found the last packed-cache index use in reused writes; the final run follows its correction.

## Mutants

TestArrayHolesSparseShareMutant removes the sparse branch from share.c. The reference witness catches the null dense-elements access under the sanitizer. TestArrayHolesSparseFreeMutant removes the sparse branch from free_one and catches the corresponding null dense-elements access at cleanup. Both controls agree with Node and finish without leaks.

A private Go overlay makes TestRuntimeStaticsAreListed read a copy of the table with the array_holes.c runtime-file and both string entries removed. It exits 1 and reports precisely the missing file, range_name and range_message. The positive inventory test passes. The mutation never changes the checked-in table. Its log is /tmp/views-v3-c2-statics-mutant.log.

The core's eight array-read omission cases, element kind, missing receiver, physical scalar/reference write layouts, sparse slot absence, numeric constructor absent slot and RangeError mutants also pass on c2. The core lowering boundary and constructor throw-edge mutants remain recorded in views-v3-checkpoint.md.

## Remaining language boundaries

This adds no callable, dictionary, intersection, erasure or cyclic-ownership admission. The broader V3 tuple/own-field/ranked-witness source certificates remain the inseparable pieces named in views-v3-checkpoint.md. The 21 reported array adapter failures are all repaired. Array membership in V2 untagged unions remains pinned awaiting its matcher, independently of ordinary array field reads. Mutable view writes remain refused until the declared source slot contract can be checked; inspecting current contents is insufficient.

Node allows a cast from 1[] to number[] followed by writing 2. Adamic's source type promise needs the declared 1 certificate checked at the write, in addition to physical storage. The proposed behavior is a checked source-contract stop. Tuple assertions also cannot make Node's array length fixed; their reader and owner certificate must preserve actual dynamic length and check selected positions. These questions and the inseparable original lane commits are listed in the core checkpoint for compiler's ruling.
