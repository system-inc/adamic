Optional indexing for roadmap step 18

Built numeric optional indexing for arrays, strings and typed arrays, guarded Map get, required-field indexing continuations, and successive explicit optional guards.
Delivery commit: the commit containing this report, on codex/optional-indexing, based on area fixture tip 4885cec50290686df487b62aac47c85d871ed40c.
Focused Node/backend oracle, lowering pins and counts refresh pass; command results are preserved in logs/.
Eleven independent compiler mutants are caught; the range-check mutant produces an ASan heap-buffer-overflow.
Sparse arrays, open records/MapLike, structural Map receivers and general ordinary continuations remain outside this implementation.

The base is saved in an owned expression local before testing null and undefined. The index or Map key occurs only in the present arm. Existing ArrayIndex, StringIndex and MapGet storage and bounds handling are reused. No production backend file changed. Required-field continuations preserve the property's existing layout metadata and guard its original receiver before either the field or index is evaluated. Parentheses continue to terminate a chain.

The source fixtures reduce TypeScript 6.0.3 checker.ts:7440 and builder.ts:1852 expressions, retain the unchanged isSymbolWithComputedName body from checker.ts:33489 and getTokenSourceMapRange body from factory/emitNode.ts:148, and retain checker.ts:36469's parameter indexing expression with scalar dependency reductions. Fixtures also cover absent and null receivers, missing elements/keys, UTF-16 string indexing, falsy stored values, side effects, reassignments of the selected base, throwing indices and successive optional guards. Typed-array variants exercise Uint8Array, Int32Array and Float64Array. Original-source observations, file hashes, stdout, stderr and exits are in node.json; node.cjs uses Node's stripTypeScriptTypes. All nine original-source observations exit 0 with empty stderr.

Validation commands (all output went directly to log files):

- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_' -count=1`: PASS, 1.072s; five fixtures, JavaScript and native stdout/stderr/exit comparisons, ASan, UBSan and leak checks.
- `go test ./internal/lower -run 'TestOptionalIndexing|TestWhatStageZeroCannotLower' -count=1`: PASS, 0.761s; existing gap pins and conservative boundary pins.
- `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts`: PASS, 54.206s.
- `python3 review/optional-indexing/run-mutants.py`: eleven independent mutations, each restored before the next, each test fails for its intended reason. mutants.json records commands, exits and catchers; logs/ preserves the output. Native emit_slots.go is changed only temporarily by the range mutant and restored.

Mutants and catchers:

| Mutation | Catcher |
| --- | --- |
| Evaluate an absent array's index | Node stdout counter mismatch |
| Omit array range check | ASan heap-buffer-overflow |
| Evaluate an absent continuation's index | Node exit mismatch; the index touches an absent declaration |
| Omit Map null guard | JavaScript backend exit mismatch |
| Reselect the base instead of reading its saved local | Node stdout mismatch |
| Omit numeric-index null guard | JavaScript backend mismatch |
| Evaluate absent Map key | Node stdout counter mismatch |
| Omit typed-array indexing continuation | UBSan null dereference |
| Omit typed-array ordinary-property boundary | Exact lowering NotYet pin |
| Omit structural Map receiver boundary | Exact lowering NotYet pin |
| Admit a string numeric key without numeric representation | Exact lowering NotYet pin |

The Map safety probe in map-shape.a is an actual object implementing ReadonlyMap with a user-defined get. Node prints 7. An initial candidate lowered it without an error; that observation alone did not establish native correctness. The final code conservatively scans compatible object literals and constructed receivers and retains NotYet when storage or method identity differs from the library Map. This may reject a genuine Map lookup in a program also containing a compatible structural receiver; dispatch through that view remains a separate compiler implementation problem.

Counts refresh adds these five rows (locals, kept, expressions, statements, functions, omitted): array 43/43/61/104/8/0; string 14/14/48/65/4/0; typed-array 32/32/23/62/6/0; Map 35/35/35/65/8/0; chain 52/52/70/107/8/0. The generator also moves logical_and_reference_maybe.a to registry order with its existing 8/8/11/22/4/0 unchanged, and removes the stale unregistered taste/17_binder_flow.a row. No other existing row's values changed.

Setup succeeded with GOPROXY=https://proxy.golang.org|direct. Timing lines: Go ready 0.026s; Node ready 0.028s; markdown ready 0.077s; submodules ready 0.078s; clang ready 0.173s; build ready 40.424s; warm cache 40.612s; total 40.641s. nproc=5, cgroup cpu.max=400000 100000. Tool versions: Go 1.27.1, Node 24.19.0, clang 20.1.8. Full setup output is preserved.

Ownership assumption: the obsolete optional-string-index gap row in lower_test.go belongs to this indexing unit. Only that row is removed from the shared test file. The scout's branch changes the separate optional-call row; its implementation files, fixtures and document were not edited or merged. The delivery contains this unit's changes on the requested area tip. No main or area branch was pushed or merged into.

Limits: sparse literals remain NotYet, so this unit does not provide a runtime representation of holes. The open-index-signature family is stopped at its existing language refusal rather than changing that ruling. The proposed open MapLike program in record-index-signature.a remains Refused with the existing named fix to use Map; Node's output is recorded but no accepted native compilation is claimed. Open Record remains NotYet. Fixed-record arbitrary keys require separate shape/presence handling. Existing nullable-string and mixed null/undefined representation limitations remain. String numeric keys such as `values?.["0"]` remain NotYet rather than being emitted with a numeric slot representation. General chains owned by the optional-call scout are not implemented here. No full package test or full gate was run; the shared gate follows the feature push. No cohere code was copied.


Same-input census comparison

Read the scout method and latent census README before measuring. A pristine compiler binary was built on the area fixture base before editing; the final candidate binary was rebuilt after all compiler mutants were restored. Both ran full mode with LATENT_ASSERT_NO_OUTPUT=1 against the same adapted TypeScript compiler directory. The run completed with exit 0 in both configurations. There are 79 measured compiler files and 79 compiler source hashes; checker diagnostics and file coverage are identical. This compiler-only scope differs from a wider compiler-and-driver manifest. measurement.json pins commands, binary hashes and raw hashes. census.json preserves each exact diagnostic root and its before attempts; before.jsonl.gz and after.jsonl.gz preserve the full raw runs.

| Exact reason | Before | After | Disappeared roots | Without a same-line replacement |
| --- | ---: | ---: | ---: | ---: |
| ?.[] on a value | 6 | 0 | 6 | 5 |
| a call through ?. (an optional call) | 136 | 108 | 28 | 0 |
| an optional chain longer than one step | 11 | 9 | 2 | 1 |
| optional chaining to .size on a value | 20 | 20 | 0 | 0 |
| ?. to a number, which would be number \| undefined | 3 | 3 | 0 | 0 |

The five index sites with no same-line replacement are checker.ts:9101, 33479, 33484, 33490 and factory/emitNode.ts:148. checker.ts:22931 now reaches an unproven Type-to-TypeVariable relation concerning optional constraint presence/type. The chain at checker.ts:53798 has no same-line replacement; checker.ts:36469 instead reaches the required-field storage boundary. All 28 changed optional-call roots still stop on the same line at the conservative structural-Map receiver guard (some also expose other refusals). They are diagnostic replacements, not 28 compiled Map sites. The reduced genuine-Map fixtures do compile and match Node.

These are exact diagnostic-root changes and same-line observations. Other boundaries in the containing attempted units may remain, and the full ledger omits final module ordering, ownership and backend passes. No whole-tsc compilation, hidden-byte reduction or complete-body retirement is claimed.

`python3 review/optional-indexing/summarize-census.py --audit` passes. `python3 review/optional-indexing/audit-mutants.py` independently changes a recorded family count and a compiler hash; both fail the audit, then the original artifacts are restored. audit-mutants.json and logs/ preserve both intended catchers. A manifest-scope assertion initially failed on the mistaken 82-file expectation; it was corrected to the actual compiler-only 79 files before any report was accepted.

Moved off the landing slice (Oct 8): record-index-signature.a refuses an index signature until records-maplike lands, and an .a file lands only when it checks on the tree it lands on. It is attached to task #p9v82wa and returns with the records slice.
