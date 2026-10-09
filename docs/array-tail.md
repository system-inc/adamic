# Dense Array long tail

Roadmap step 30 (#py0tfat). Array methods use the existing dense representation. Holes, sparse writes, inherited indexed properties, descriptors and property-presence operations remain refused. This unit changes library lowering, without adding a new runtime ABI or target-specific libc calls.

## Default sort

`sort()` and `sort(undefined)` admit dense number, boolean and string arrays. The comparator follows V8 SortCompareDefault: convert primitive elements to strings and compare UTF-16 code units. Sorting mutates the original array and returns that same reference. A supplied comparator retains the existing behavior. Optional comparator function values remain NotYet until the compiler represents their nullable callable value. Undefined elements and object coercion remain NotYet for the default comparator; their ordering is not approximated.

The TypeScript 6.0.3 source at 050880ce59e30b356b686bd3144efe24f875ebc8 calls default sort in src/compiler/sys.ts:1873,1874; core.ts:705; watchUtilities.ts:156,157; checker.ts:5564. Its comparator-based calls continue through the existing stable V8-derived runtime sort.

## reduceRight

`reduceRight(callback, initial)` snapshots the length after evaluating its arguments, visits indices in descending order, and passes accumulator, element, index and receiver. Callback growth is excluded by the snapshot; callback shrinkage skips indices no longer present in the dense array. Reference accumulators use ordinary IR ownership and closure calls.

Reduction without an initial value remains NotYet: the empty-array case requires a catchable TypeError descriptor, and assuming a nonempty input would miscompile a later empty call. Arbitrary array-like receivers, detached methods and heterogeneous representations remain refused. A callback returning never or a function returning undefined can encounter existing compiler representation limits. No predicate truthiness implementation is duplicated here.

## Verification contract

The sort and reduceRight fixtures are compared with source Node 24.19.0 on native, JavaScript and WASI. Reversing the default comparator and skipping reduceRight's index zero each produce clean executions caught only by Node stdout comparison. Refusal tests retain sparse-array, missing-initial, intrinsic-override and default-comparator representation boundaries. New Linux allocation rows are recorded in internal/oracle/counts.md; V8 sources are credited in THIRD_PARTY_NOTICES.md.

The supplied predicate-era census depends on codex/library-array-refused 465d6bd08a74f5a43bde0e2c62e528478081c248, which is not in this unit's fetched area base ca016bab19c040c44748924cde12976afbf68978. Scratch Go overlays measure that pending dependency independently; neither its implementation nor generated reports are added to this branch. Full refusal groupings and measurement output belong outside the checkout.
