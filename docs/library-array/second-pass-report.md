Built: dense Array overloads, reductions, receiver calls, string iteration, default sorting, and optional shrinking searches in both backends.
Commits: claim 91574e9; implementation e0c2d8a; constructor safety fix 3c312dd; main integration 8fc9808.
Commands and outputs: validity runner Array 191 → 242 passing, zero disagreements; Linux package, uncached Array oracle, counts and vet checks pass.
Mutants: 34 executions caught by Node comparisons or explicit refusal checks; restored after each execution.
Uncovered: language foundations, first-class intrinsic aliases, remaining generic receivers, and other libraries; full repository functional gate was not run.

## Measurements

The unchanged validity runner was built from a separate checkout of
origin/codex/test262-ts-validity af12899. The compiler is this branch, incorporating
origin/main 50045bd and Array first pass 34da3d8. Test262 is pinned at
c8c798898646638cd0c24879f8e0374e847e7d74; stock tsc is 6.0.3; adaptation is enabled.
Only the refused column is work owed. The 495 figure predates the first Array pass;
this actual baseline has 437 refusals.

| Survey | Pass | Disagreement | Refused | Crashed | Skipped |
|---|---:|---:|---:|---:|---:|
| Before | 191 | 0 | 437 | 0 | 610 |
| After | 242 | 0 | 386 | 0 | 610 |

Both surveys contain 1844 not-typescript outcomes and 3082 total inputs. All 191
baseline passes remain passing. The 51 new passes are archived in
second-pass-new-passes.json; complete runner tables are second-pass-before.json
and second-pass-after.json. First-blocker classifications and exact reasons for
all 437 original refusals are in second-pass-final-refusals.json. They include
51 newly compiling inputs and 386 remaining refusals; later blockers can remain
masked. The first claim commit precedes compiler and fixture edits.

An intermediate survey exposed four disagreements after admitting constructors:
indexed writes past a dense prefix, including reassignment to an empty array.
These now refuse explicitly rather than manufacturing holes or huge arrays.
second-pass-investigation.json preserves that observation. A separate optional
numeric constructor probe revealed that a numeric union could select the length
overload at runtime; ambiguous single-argument constructors now refuse.

## Changes

Dense `Array(...)`, `new Array(...)`, and `Array.of(...)` preserve argument order;
nonzero/dynamic numeric lengths and numeric unions refuse. Ordinary IR implements
reduce/reduceRight, including omitted seeds, saved length, missing-index skipping,
retained callback arguments, and exact empty-array TypeErrors. Default sorting
supports dense nonoptional number, boolean and string elements using the existing
UTF-16 lexical comparator. Object coercion and optional-element partitioning
remain refused.

Immediate `.call` supports proven dense array receivers, nullish errors and
supported primitive receivers. Mutating methods on primitive strings reproduce
V8's immutable character/length errors, including UTF-16 indexes. Escaping boxes,
unproven array-like shapes, and object coercion remain refused. `Array.from(string)`
iterates Unicode code points without a mapper; Array.prototype.length is observed
only when its intrinsic identity and absence of mutation are proven.

V8 13.6.233.17 ports are named in THIRD_PARTY_NOTICES.md: elements.cc,
array-reduce.tq, array-reduce-right.tq, array-from.tq, builtins-array.cc,
array-shift.tq and array-unshift.tq. Existing IR/runtime machinery holds these
algorithms without a new runtime allocation model.

## Shrinking searches and correction

The requested string findLastIndex fixture and findLast twin exercise the inserted
check. Observed behavior is an Adamic panic, exit 70, rather than a RangeError.
Raw Node visits the removed indexes with undefined; therefore these two fixtures
use the existing checked-oracle policy and additionally record raw Node's complete
output. Node findLastIndex prints `3 dd`, `2 undefined`, `1 undefined`, `0 dd`, `-1`;
Adamic prints `3 dd` then stops with the documented search-shrink message. Skipping
removed indexes is distinguishable from both behaviors and is caught in each
backend for each twin.

For element types that can hold undefined, all four find methods now pass the
missing element as undefined and continue. ArrayVisit carries an explicit proven
AllowsUndefined bit: Leaf and Leaf | undefined otherwise share a pointer
representation. Numeric absence uses the existing optional-number encoding;
reference absence uses NULL. Existing ownership rules retain callback arguments.
The exact user number and Leaf fixtures print `1`, `2`, `hole`, `-1` in Node,
native and JavaScript. An additional fixture covers optional number and Leaf
find/findLast/findLastIndex and optional heap strings. The nonoptional stop remains.

The integration-owned environment line in library_array_mutant_test.go is untouched.
Shared dispatch changes are limited to the necessary constructor/search hooks and
IR proof metadata; lower.go, emit.go, native.go and oracle_test.go are untouched.

## Mutants and catchers

Every native output mutant below compiled and finished without sanitizer or leak
reports, and was caught by stdout comparison with Node. All mutations were restored.

| Mutant | Catching fixture/check |
|---|---|
| Constructor argument order | library_array_dense_construction.a |
| reduceRight stride | library_array_reductions.a |
| Reduction saved-length boundary | library_array_reductions.a |
| Empty reduction TypeError message | library_array_reductions.a |
| Nullish receiver TypeError message | library_array_receiver_calls.a |
| Primitive every result | library_array_receiver_calls.a |
| String-from iteration order | library_array_from_string.a |
| String mutation readonly message | library_array_string_mutations.a |
| Generic forward boundary | library_array_search.a |
| Generic backwards negative sign | library_array_search.a |
| Array.isArray classification | library_array_metadata.a |
| copyWithin undefined end | library_array_copy_within.a |
| Iterator index | library_array_iterators.a |
| Join separator | library_array_join.a |
| Method metadata | library_array_metadata.a |
| with replacement | library_array_with.a |
| flatMap order | library_array_flat_map.a |
| Spliced copy | library_array_spliced.a |
| Flat order | library_array_flat.a |
| copyWithin write | library_array_copy_within.a |
| Search direction | library_array_search.a |
| Copy reversal | library_array_copy.a |
| Default sort ordering | library_array_copy.a |
| Reverse callback order | library_array_find_last.a |
| Skip missing index, native and JS (4 executions) | String findLastIndex fixture and findLast twin; checked prefix/panic and raw Node comparisons |
| Stop regardless of optional type, native and JS (4 executions) | Exact optional number and Leaf fixtures; Node comparison catches exit 70 after `1`, `2` |
| Remove constructor dense-prefix guard | Three indexed-write refusal probes unexpectedly compile |
| Remove numeric-union overload guard | Optional numeric constructor unexpectedly compiles; mixed numeric constructor loses the specific refusal |

This is 24 ordinary output mutants, eight shrinking-search backend executions,
and two source guard mutations: 34 executions. The guard mutations are caught
before C by meaningful refusal tests. See second-pass-logs/ for output evidence.
The optional search tests regenerate original emitted artifacts to prove restoration.

## Linux verification and toolchain

`bash cloud/setup.sh` succeeded. Timing lines: Go ready 0s; clang ready 0s;
Node ready 0s; submodules ready 0s; build cache warm 84s; done 84s.
Go 1.27.1, clang 20.1.8, Node 24.19.0. `nproc` prints 5;
cgroup cpu.max is 400000 100000; memory was reported as 17.6 GB.
Every build/test shell sources /workspace/adamic-tools/env.sh.

Tests write to files. Commands and results:

```
go test ./internal/ir ./internal/lower ./internal/flow ./internal/native ./internal/javascript ./cmd/adamic-test262 -count=1 -timeout 15m
# pass: lower 23.034s, flow 106.690s, native 144.185s, runner 70.981s; ir/JS have no standalone tests
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestArray(Optional)?ShrinkingSearch.*|TestArrayFamilyMutants|TestNativeAgreesWithNode/internal/oracle/testdata/(library_array_|maybe_collections|searches|adversarial_order|closures_throw)' -count=1 -v -timeout 15m
# pass 18.653s; native 0 cache hits/82 misses; Node 0 cache hits/78 misses
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
# pass 35.104s
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m
# pass 23.616s
go test ./internal/lower -count=1
# after numeric-union fix: pass 7.999s; dedicated restored refusal check passes
go vet ./...
# pass
git diff --check
# pass
```

The pinned cohere CLI does not recognize .a despite the repository's sourceExtensions
configuration. Its named-file check fails with "not a TypeScript or JavaScript
file". Identical scratch .ts mirrors, never committed, pass types-only checking
(12 files, including the prelude). The full mirror lint reports 66 findings,
including deliberately exercised Array constructors/default sort, exact supplied
fixture syntax, callback parameters, and prelude style. It is not claimed green;
no broad lint exceptions or checker edits were introduced. Type, Node, sanitizer,
count and package gates above are the Linux evidence. The full repository functional
suite was not run.

## Remaining work and handoff

The claim's original ownership is preserved in the ledger so scope cannot be
silently rewritten. Newly exposed first blockers are separately classified.
Sparse presence, indexed extension, descriptors, generic/empty/heterogeneous
slots, coercion, function metadata/identity, constructor aliases in instanceof,
and other language features need compiler representations outside this slice.
The original claim gives every language family a one-line reproducer for
@system_adamic; second-pass-handoff.md adds the newly exposed families.

The remaining directly owned Array blockers are first-class intrinsic reads/aliases
(14 inputs), map.call on an unproven array-like receiver (one), and shift.call on
a function receiver (one). These were not implemented: representing an intrinsic
as a user closure changes identity, metadata and generic call behavior; treating
an unproven object or function as a dense array changes presence, property and
length semantics. They continue to refuse explicitly. Other-library constructors
and Object.freeze are left to their owners. This second pass does not complete
every TypeScript-valid Array test.
