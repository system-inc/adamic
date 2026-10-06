# Concurrency compiler claims (#p286ycm)

Branch: codex/concurrency-compiler, from origin/main (5d4c801).
Contract: origin/codex/concurrency:docs/concurrency.md.

Create internal/lower/shareable.go and internal/lower/parallel.go for structural
Shareable inference, capture checks and conservative task effect proofs.
Create internal/native/parallel.go for the runtime ABI call.
Add small ParallelMap hooks following ReadTextFile in internal/ir/ir.go,
internal/lower/input.go, internal/javascript/javascript.go,
internal/fresh/fresh.go and internal/native/emit_expressions.go.
Add concurrency oracle fixtures, exact refusal checks and tests for these proofs.
Add a small native gate in internal/oracle/oracle_test.go naming codex/concurrency.
Claim any necessary embedded adamic declaration and Node oracle shim hooks after
inspection, before editing them. Reuse cycles.go's traversal; any changes there
will be reported to its owner explicitly.

Do not edit internal/native/runtime/, internal/native/emit_objects.go,
internal/native/emit.go, internal/lower/lower.go or internal/native/native.go.

Inspection additions: internal/load/prelude.d.ts (the public signature),
oracle/adamic.mjs (the Node sequential witness), and a three-line entry hook in
internal/lower/refusals.go so task async/effects diagnostics precede general ones.
Create internal/oracle/concurrency_test.go, internal/lower/parallel_test.go and
internal/native/parallel_compiler_test.go. Concurrency fixtures use their own
oracle list until runtime integration, so existing native counts remain intact.

Exception propagation requires adding ParallelMap beside ArrayMap in
internal/lower/exceptions.go's closure-call case. This is a one-token hook;
without it a throwing task's caller would not carry the exception onward.

## Compiler report

Built the public parallelMap declaration, dedicated IR node, sequential Node
witness and JavaScript helper, native ABI lowering, Shareable/capture/effect
proofs, exception propagation, and 13 accepted plus 28 exact refusal fixtures.
The native adapter retains the original task closure (preserving its identity),
immutable reference-valued globals used by its call graph, and runtime result
ownership metadata. No stand-in runtime or runtime source edits were added.

### Readonly evidence

The checker represents both readonly T[] and ReadonlyArray<T> with the
ReadonlyArray library symbol; ordinary T[] has Array. Readonly tuples use
TargetTupleType().IsReadonly(). Explicit readonly fields, Readonly<T> mapped
fields and as const object fields all report true through IsReadonlySymbol.
Readonly is only a slot property: a readonly slot containing Map remains mutable.
Direct checker tests cover these observations. Hidden mutable subclass fields
and Readonly<Map<...>> method wrappers have exact refusal witnesses too.

The implementation reuses cycleFinder.use, fields, made and its visited-type
map. Production cycles.go is unchanged; there is no second recursive type walker
and no ownership change to coordinate with @system_adamic_runtime. Its visited
guard was temporarily mutated for the required recursion experiment, then
restored byte-for-byte.

### Verification

Toolchain setup succeeded: Go ready 0s; clang, Node and submodules ready 1s;
build cache warm and done 98s. nproc reported 5. Environment file:
/workspace/adamic-tools/env.sh. Node v24.19.0, Go 1.27.1, clang 20.1.8.

Required lower/javascript tests and the full oracle passed. The full oracle
before the added large fixture took 157.785s. Logs are
/tmp/concurrency-lower-final.log and /tmp/concurrency-oracle-final.log.
Native compiler ABI unit checks passed (/tmp/concurrency-abi-final.log).
Concurrency fixtures compare source execution on Node with emitted JavaScript,
including item/index order and the first middle-item exception.

All six requested mutants compiled and were caught, then restored:

| Mutation | Witness | Failure |
| --- | --- | --- |
| Mutable array is Shareable | nested_array | Expected Refused, got nil |
| Captured let permitted | captured_let | Expected Refused, got nil |
| Call graph skipped | global_write | Expected summarize -> record refusal, got nil |
| Readonly field blesses mutable Map | map_field | Expected options.cache refusal, got nil |
| Recursive visited guard removed | recursive_tree | Go stack overflow, failing test |
| Node result order reversed | numbers | 40,21,92,13 versus 13,92,21,40 |

Logs: /tmp/concurrency-mutant-{mutable_array,captured_let,call_graph,map_field,
recursive,node_order}.log. The initial Map mutant exposed an unrelated generic
Map construction limitation; its fixture was corrected to explicit generic
arguments and rerun, producing the intended missing-refusal failure.

Native fixtures are gated by default with an explicit codex/concurrency reason:
the compiler branch has no runtime definition. ADAMIC_CONCURRENCY_RUNTIME=1
opts into actual native oracle checks after integrating the runtime.
A detached worktree merged runtime da8d3bc with this compiler branch without
changing either source branch. All concurrency fixtures passed native sanitizer,
release and leak checks with ADAMIC_THREADS=1 and 5. A further 2048-item fixture
passed with five threads, exceeding the runtime scheduling grain and exercising
shared global string roots plus fresh reference results. Integration logs:
/tmp/concurrency-integration-one.log, /tmp/concurrency-integration-five.log and
/tmp/concurrency-integration-large.log.

### Deliberate conservative limits

Async work and moving mutable values remain parts 3 and 2 respectively.
Unknown closure calls and class dispatch are refused. Helpers mutating a fresh
object through a parameter and some nested closures over mutable task-local
bindings can be refused because freshness is not propagated context-sensitively.
Readonly tuples are supported as element types; top-level tuple-valued items
are conservatively refused rather than adapted from their native object layout.
Concurrency fixtures have a separate oracle list, so counts.md native baseline
rows remain pending runtime integration. No full repository-wide test suite or
exhaustive TypeScript conformance run was performed.

## ABI revision

The runtime owner's subsequent decision replaces closure result metadata with
an explicit third parameter:

```c
adamic_array *adamic_parallel_map(adamic_array *items, adamic_closure *work, bool references);
```

The compiler emits this declaration and passes Result.IsReference() directly.
The adapter still preserves closure identity and shared global marking roots;
it no longer writes result_references. The ABI unit check covers both false
(number results) and true (string results), and rejects metadata writes.
Native fixture gating on codex/concurrency remains unchanged. The fetched
runtime branch still declared the previous two-argument signature during this
revision, so the earlier native integration results document the previous ABI;
updated native integration requires the runtime owner's three-argument update.
Required compiler packages and the full oracle are rerun for this revision.

Revision verification: lower/javascript packages passed; the full oracle passed
in 17.636s (/tmp/concurrency-oracle-v2.log); the ABI unit check passed in 0.007s
(/tmp/concurrency-abi-v2.log). All six mutants were rerun and caught by the same
witnesses, with all source restored (/tmp/concurrency-mutants-v2.log).
