# ecmascript/control_flow_graph

The Oct 8 07:57 ruling uses concrete private-constructor indices until generic enclosing-type-parameter instantiation is supported. BlockIndex is imported as CfgBlockIndex from ../../../arena/arena_index.a; NodeIndex is CfgNodeIndex in that same home. Each has one readonly slot, and only its own static push constructs it and publishes its freshly created record. Absence is undefined; slot zero is valid. BlockArena owns the ordered Block array; successors and predecessors contain indices, preserving Go cfg.go:336 insertion order, duplicate edges and nil-edge meaning. Every read checks bounds and the arena lifetime. graph.arena.dispose frees the owning array as one unit. AST parent/child relationships likewise use a separate NodeIndex arena. No reference cycles or runtime regex compilation are required. A graph and its path-analysis state must be disposed together; borrowed block records must not escape that scope.

[SYMBOLS.md](SYMBOLS.md) maps all 79 triage symbols, four exposed query/ownership helpers and the predecessor adapter to individual files. arena.a, ast.a, builder.a, path_analysis.a and support.a hold representation types and supported primitive adapters. finalize_predecessors.a populates predecessor lists in Go source-block order. main.a is the test driver, not the rule entry point.

normalize_big_int_literal.a is reused byte-for-byte from origin/codex/lint-helpers-from-lint-wave1-12, wave12/control_flow_normalize_bigint_literal.a. The earlier partial CFG ports supplied scalar forwarding contracts; the graph walkers, roots and path analysis are integrated from pinned Go cohere 7945d102a6c18dd36adf9114a758ce646e8b2359 with the arena representation. Go slice/chunk storage is replaced by supported arrays without changing observed event or edge order.

The capture overlay wraps real Build, IndexRoots and AnalyzePaths calls from every upstream array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks test, plus the upstream root/path-analysis tests and targeted Go controls. It records actual AST facts and hook effects. The driver replays opaque event payload IDs, comparing event placement, Emit block/slot, hook order, ordered graph structure, root ownership and path queries; it does not port the other consuming rule bodies. Direct Go observations also cover scalar helpers and labels. Deterministic compressed fixtures are regenerated from Go in the test suite.

Tests compare the complete fixture bytes on source Node, emitted JavaScript and ASan/UBSan native. Each semantic mutant must compile, run successfully and differ on Node; its first discriminating Go capture is then checked on sanitized native. The independent off-by-one index mutant must stop with the bounds diagnostic on both runtimes. Compiler checks reject outside construction, bare numbers and cross-graph classes. Reads reject handles from another arena on Node and native. indexAtSlot only looks up an already minted live handle; it never constructs or casts one from a number. Raw -1 values appear only in the unchanged Go wire protocol and ordinary numeric algorithm counters, never in arena references.

Run with /workspace/adamic-tools/env.sh sourced:

```
go test ./stage1/cohere/arena ./stage1/cohere/lint/helpers/... -count=1 -v -timeout=3h
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/tailwind-typescript ADAMIC_LINT_BENCH=1 go test ./stage1/cohere/lint -count=1 -v -timeout=3h
```

The proof rule is [no-unreachable-loop](../../rules/no-unreachable-loop/README.md). Completing this helper dependency removes the frozen helper blockers for array-callback-return and react-hooks/rules-of-hooks; their rule ports still require their own proof. consistent-return retains its rule-local helpers and property.Name dependency. No current CFG compiler gap remains; docs/memory.md is owned by hir-01.

The index class and record are created together because the supported cycle checker requires publication of fresh values. Jump and try stacks use supported array-copy updates to preserve the same order without mutable cycle-capable stack writes. The generic collapse must replace the concrete classes and every call site together, with no compatibility shims.

[BOXING.md](BOXING.md) records measured allocation and reference-count costs; the brand remains intact.
