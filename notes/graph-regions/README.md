# Graph region coverage

Base: origin/codex/graph-regions at 42e2a35. Ten new oracle programs agree on source Node, the JavaScript backend, native release, ASan/UBSan and the separate leak run. One additional program differs and is kept here, outside testdata.

## Case inventory

Paths in the existing column are under internal/oracle/testdata. A dash means the graph-specific combination was missing before this coverage branch. New names abbreviate graph_regions_coverage_NAME.a. This is a source-case inventory, not a claim of exhaustive machine branch coverage.

| Code case | Existing source evidence | Added evidence |
| --- | --- | --- |
| No unproven seed: fresh writes, ordinary counted allocation and statement arenas | fresh_parser.a, fresh_writes.a, fresh_calls.a, regions.a | Already covered |
| Unproven mutable object fields, post-creation parent and children writes through a helper | graph_regions_parse.a | Already covered |
| Strong next and previous fields without Weak | graph_regions_list.a | Already covered |
| Loop back edges, setting undefined, restoring a saved edge | graph_regions_flow.a | Already covered |
| Self store and twin cycles; same-member/same-region merge is a no-op | graph_regions_literals.a, graph_regions_regression_01.a | spread, arrays, maps |
| Returned graph root and dropping the last local anchor | graph_regions_anchor.a | throw, merge |
| Local outside reference survives dropping root | graph_regions_escape.a | boundary, merge |
| Cache containing graph values promoted even without a cache back edge; cache is sole owner after builder returns | graph_regions_cache.a, graph_regions_literals.a | maps |
| Symbol/declaration inverse through an array | graph_regions_symbols.a | Already covered |
| Object allocation's concrete, widened variable and contextual type identities | graph_regions_parse.a, graph_regions_literal_method.a | boundary, spread |
| Structural views, concrete generics, inherited/private fields and class graph flag | graph_regions_regression_10.a through graph_regions_regression_15.a | Already covered |
| Static constructor object, inherited static base and constructor interface | graph_regions_static.a, graph_regions_static_parent.a, graph_regions_static_interface.a | Already covered |
| Accessor closure capture and literal method closure capture | graph_regions_accessor.a, graph_regions_literal_method.a | Already covered |
| Union graph target versus undefined; container promotion examines union members | - | boundary |
| Readonly and proven strong paths in a selected ownership component | graph_regions_parse.a, graph_regions_symbols.a, fresh_refused/ctor_readonly.a | payload_loop |
| Readonly array/tuple, ReadonlyMap/ReadonlySet views and Set adapter next present/exhausted | Empty readonly array: graph_regions_regression_01.a | readonly (nonempty graph views and generic pass-through) |
| Arrays and tuples participating in cycles; Map entry pair objects | graph_regions_entries.a, graph_regions_regression_04.a | arrays, maps, iterator |
| Empty graph array and graph array literal with graph objects | graph_regions_parse.a | arrays, boundary |
| Array spread and runtime append of graph references | fresh_refused/spread_array.a | arrays |
| Array slice/concat populated-result adoption converts owned child counts to internal links | Concat: fresh_refused/concat.a | arrays (adds slice and uses results after removal) |
| Array map callback result transferred with graph_take | fresh_refused/map_callback.a | arrays |
| Array filter true/false branches, retained element transferred on true and released on false | - | arrays, boundary |
| Array.from owned callback result transferred into graph array | fresh_refused/array_from.a | arrays |
| New filled array adopted after population, fill existing graph array | Existing fill: fresh_refused/fill.a | arrays (adds new filled allocation) |
| Array indexed overwrite, push and undefined slot | graph_regions_parse.a | arrays, boundary |
| Returned splice result and pop acquire outside ownership before source slot is removed | - | arrays |
| Graph array sort temporary owners transferred back, old internal slots dropped | fresh_refused/call_comparator.a | arrays |
| Map empty/literal-pairs/from-Map construction and add_pairs graph holds | Empty construction: graph_regions_cache.a | maps |
| Map graph values with ordinary counted string keys; overwrite existing key, delete hit/miss and clear | Basic set: graph_regions_cache.a | maps |
| Map graph object keys and numeric values | graph_regions_regression_07.a | Already covered |
| Map graph keys and graph values, duplicate graph key disposal | - | maps |
| Map keys/values/entries snapshot arrays; entry tuples adopted with graph members | Entries: graph_regions_entries.a | maps |
| Map for-of iterator retains its graph, graph iterator allocation and entry binding | fresh_refused/map_iterator.a, fresh_refused/map_iterator_parenthesized.a | maps |
| Map keys/values/entries adapters: graph state, cell, closure, wrapper; next present and exhausted after collection local drops | Cycle construction in fresh_refused/ iterator probes | iterator |
| Set graph member construction, add duplicate, delete and clear, snapshot values | Construction/add: graph_regions_regression_08.a, graph_regions_regression_09.a | maps |
| Graph-to-graph field holds merge before publishing, drop skips internal release; graph-to-counted and counted-to-graph retain/release | Basic holds in parse/cache | payload_loop, boundary |
| NULL target hold/drop; plain counted programs bypass graph helpers | graph_regions_flow.a and ordinary oracle fixtures | boundary, merge |
| Borrowed parameters, owned local/global and returned reference boundary conventions | graph_regions_parse.a, graph_regions_regression_01.a, graph_regions_anchor.a | spread, boundary, throw |
| Named direct call and call through a function value carrying graph argument/result | Direct helpers: graph_regions_parse.a | spread |
| Graph capture cell, reference and scalar slots; owned capture initialization; arrow call and callback call | graph_regions_regression_03.a, graph_regions_regression_05.a | closure_loop |
| Indivisible named shared environment, disjoint captures propagated; direct sibling calls keep existing proof | graph_regions_closure.a, nested_mutual.a, nested_mixed.a, nested_captures.a | Already covered |
| Per-iteration graph capture cell replacement and escaping closures | - | closure_loop |
| Captured parameter/interior environment owner convention | nested_returned.a, nested_pattern_parameter.a, nested_three_levels.a | Already covered |
| Empty environment, hoisting, TDZ reads/writes and destructuring | nested_minimal.a, nested_hoisting.a, nested_tdz.a, nested_tdz_write.a, nested_destructured.a, nested_destructured_tdz.a | Already covered |
| One lone graph object needs no region record, self-link cannot add an internal count | graph_regions_regression_01.a | throw, boundary |
| Lone/lone join, established/lone join and established/established union with size swap; access through losing records | Parse/flow have joins; million has large joins | merge, arrays |
| Outside counts summed at merge and kept after two roots are dropped | Basic escaping graph: graph_regions_escape.a | merge |
| Overwritten pointer's old target retained until final region release; diagnostic reports unreachable members | graph_regions/million.a | payload_loop, merge |
| Strings and plain number/string arrays held by graph objects freed through iterative outside release | Dynamic strings: graph_regions_parse.a | payload_loop |
| Region free invalidates object Weak handle | No ordinary oracle fixture reads after expiry | weak_expiry.a here: outputs differ |
| Repeated graph construction and dropping each loop iteration; bounded peak | fresh_refused/ctor_readonly.a builds three graphs | payload_loop (20 iterations and retained orphan payloads) |
| Graph survives caught throw/finally, then last outside owner drops | Caught: fresh_refused/call_then_throws.a | throw (adds finally and explicit final drop) |
| Graph object spread must copy; cannot infer individual uniqueness from region count | - | spread |
| Graph array map and spread cannot reuse based on outside count; graph allocations excluded from statement arenas | Ordinary reuse and arenas: reuse.a, reuse_arrays.a, regions.a | arrays, spread, payload_loop |
| Counted-only mark report: lone/merged graph, outside roots, graph links, ordinary children ignored, overwritten members unmarked | graph_regions/million.a and branch counted tests | all ten counted teardown checks |
| Region reports absent from uncounted runs; count report with zero/nonzero records | Existing graph oracle | all ten CLI runs and counts |

## Cases without an Adamic source program

- Shared-region merge rejection: threads and the mark-shared operation have no Adamic language API. The branch has a direct C TestGraphRegionsRuntime shared-mode probe.
- Overflow of member/outside counts or region sizes and allocation failure: a short deterministic source cannot allocate or retain SIZE_MAX objects. No fault injection API exists.
- Invalid adoption (NULL, aliased or already adopted allocation), merging non-graph values, crossing unmerged regions, releasing without ownership, freeing with outside owners, unsupported diagnostic heap kinds: these are compiler/runtime invariant guards, unavailable to well-typed Adamic source. Valid language stores enforce the relevant invariant.
- Weak handles to interior environment cells: user code can reference captured values and closures, but cannot name runtime cells. Object Weak expiry is probed here.
- Watch-mode Program versions, cross-thread atomics and simultaneous merge: branch documentation explicitly says these are not built.
- Generic/block-scoped nested declarations, optional/default/rest nested parameters, dynamic this, first-class sibling/ancestor references and declaration rebinding: existing lower/nested_functions_test.go probes loud NotYet/checker errors. They cannot be passing oracle programs on this branch.

## Difference

weak_expiry.a, all exits 0 and stderr empty:

Node source and JavaScript backend:

```
child2
child2
```

Native release and sanitized native:

```
child2
expired
```

The source annotation Weak<Node> is erased on Node. internal/javascript/javascript.go:755 returns the value unchanged for ir.WeakOf, and ir.WeakTarget does the same at line 757. Native invalidates it at internal/native/runtime/graph_regions.c:328, adamic_weak_forget(each), when the last outside owner drops. This is documented existing Weak lifetime behavior, rather than evidence that native should retain the region. The ordinary source oracle cannot certify post-expiry Weak equality. Temporarily registering this notes program in the existing fixture list made TestNativeAgreesWithNode fail on stdout, then registration was removed.

## Mutation evidence

One line of internal/native/runtime/graph_regions.c, release_outside, was changed from outside_release(value) to outside_release(NULL). The payload_loop oracle failed only at the separate LeakSanitizer check: 16,260 bytes in 360 allocations. It compiled and produced normal output. The source was restored in a finally block. The restored ten-fixture behavior/count/free run passed. The first nine-fixture restored run had native cache hits 0/misses 36, Node hits 0/misses 18. The readonly view program was added on the final inventory pass and verified afterward. All ten counts rows have allocations equal to frees and no statement arena values. payload_loop reports 20 regions and 40 merges, peak 17 live values, and frees 424 of 424 allocations. Each iteration reports 3 graph members, 2 reachable and 1 unreachable before the final drop.

## Toolchain and commands

bash cloud/setup.sh was rerun after the initial cache warm overlapped the branch switch and failed with inconsistent source errors. The stable-checkout run passed: Go ready 0s; clang ready 0s; Node ready 0s; submodules ready 0s; build cache warm 81s; done 81s. nproc: 5; cpu.max: 400000 100000. The printed env file /workspace/adamic-tools/env.sh was sourced for every Go build/test shell. Go 1.27.1, clang 20.1.8, Node 24.19.0.

Initial probe runs found unsupported Array-as-function syntax and a long optional chain, plus console.log requiring a string and a statically narrowed global Weak. Those programs were adjusted to supported syntax before the successful runs. Two initial test filters selected no tests (Go separates test names at slashes); those runs are not counted as verification.

Successful focused commands:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/graph_regions_coverage' -count=1 -timeout 30m > /tmp/graph-coverage-oracle-all.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/graph-coverage-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/graph_regions_coverage|TestGraphRegionsCountsAndFree/internal/oracle/testdata/graph_regions_coverage' -count=1 -v -timeout 30m > /tmp/graph-coverage-restored.log 2>&1
gofmt -w internal/oracle/oracle_test.go
gofmt -l cmd internal > /tmp/graph-coverage-format.log 2>&1
go vet ./... > /tmp/graph-coverage-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/graph-coverage-gate.log 2>&1
```

Each of the ten testdata programs and notes/graph-regions/weak_expiry.a was separately built with:

```sh
go run ./cmd/adamic build "$file" -o "/tmp/graph-coverage-binaries/$name"
"/tmp/graph-coverage-binaries/$name"
node --disable-warning=ExperimentalWarning oracle/node.mjs "$file"
```

Outputs are recorded in the per-program .build.log, .native.log and .node.log under /tmp/graph-coverage-binaries. The oracle also emitted and executed each JavaScript backend and compiled sanitizer and release variants. Mutant and Weak differential oracle commands used the same TestNativeAgreesWithNode filter with graph_regions_coverage_payload_loop.a and notes/graph-regions/weak_expiry.a respectively, ADAMIC_GATE_UNCACHED=1, -count=1 and -timeout 30m. Their logs are /tmp/graph-coverage-mutant.log and /tmp/graph-coverage-weak-oracle.log.

## Final checks

The final ten-program behavior/count/free run plus the complete counts-table check passed in 59.690s, with zero native and Node observation cache hits. Counts regeneration with all ten fixtures passed in 44.796s. Only ten new rows are added; every previous counts row is unchanged. The tenth readonly program frees 20/20 allocations, with one region and nine merges. The final table and registry contain all ten programs.

The full repository gate was started before the tenth fixture was added. It was stopped with SIGTERM after about eight minutes (exit 143), along with its 26 descendants, after separately reproducing the branch's documented stage1 gap expectation failures. Its buffered log contains only the two bench no-test-file entries; no completed package result is claimed from that partial gate. The final focused run validates all ten programs after that registry change.

The additional baseline check actually run was:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/typescript/parser ./stage1/cohere/markdownblocks -run 'TestStrongAstParentGap|TestParserRepresentationProbes' -count=1 -v -timeout 10m > /tmp/graph-coverage-existing-gaps.log 2>&1
```

It exited 1: TestStrongAstParentGap expects a refusal but receives nil; markdownblocks gaps 2_state_arrow_cycle and 4_structural_ranges also now compile, and 1_recursive_state returns the newer first-class nested function reference NotYet diagnostic. These failures are already documented in the base branch's docs/memory.md. The stage1 tests and gap documents were not changed.

Final commands (after the readonly addition):

```sh
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/graph-coverage-counts-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/graph_regions_coverage|TestGraphRegionsCountsAndFree/internal/oracle/testdata/graph_regions_coverage|TestCountsAreRecorded' -count=1 -v -timeout 30m > /tmp/graph-coverage-final.log 2>&1
go run ./cmd/adamic build internal/oracle/testdata/graph_regions_coverage_readonly.a -o /tmp/graph-coverage-binaries/readonly-counted --count
/tmp/graph-coverage-binaries/readonly-counted
git diff --check
```

The extra counted CLI attempt initially placed --count before -o and printed usage (exit 2); the argument order above succeeded. The ordinary CLI build and source Node run for readonly also passed, printing 7 7 true, true, true on three lines. gofmt and vet had empty logs. No compiler/runtime production changes remain.
