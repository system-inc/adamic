Task #0p2kkth. Implementation `0dfe22fb540f7a7649a1811d62f9b1c9e2a74016` on `compiler/lower-void-widening-split`, with sole parent `6c69f66f55e754889f8fa88e17f93f887980b808`. This commit must land with chg_tsr07c1d. Evidence is on `evidence/compiler/lower-void-widening-split`.

Changes since ad029e0c: compare source and target recursively through array/tuple elements, corresponding target properties, unions on either side, constrained type parameters, and same-reference generic arguments (including Promise). A type-pair seen-set terminates recursive types. A 64-pair depth bound fails closed with NotYet; depth-limit.ts.txt and its stderr show a 70-level relation stopping there. Diagnostics name the container path. The comparison is used at every existing relation site.

Async return sites use the declared function result after awaiting it. Promise<void> forwarding therefore stays legal, while Promise<unknown> returning an awaited void call is refused. async_plain.a and async_coverage_discard.a are byte-identical to the requested split base. The regexp replace callback keeps its explicit undefined result. The previous void-union implicit return support and the late visitor position immediately before optional widening remain.

Implementation locations: internal/lower/void_widening.go:12 (site discovery), :37 (async returns), :95 (diagnostic), :149 (bounded structural comparison); internal/lower/refusals.go:234 (visitor); internal/lower/functions.go:334 (implicit void-union exit); internal/oracle/void_widening_test.go:13 and :23 (legal registration and directory test).

All 27 refused source programs finish on Node; each has a first-line a-check header and its complete repository-relative diagnostic pinned. Three earlier diagnostics now stop sooner at their containing initializer: shorthand, v05, and v06. These are the pinned diagnostics:

| Program | Diagnostic |
| --- | --- |
| any_return.a | `internal/oracle/testdata/void_widening_refused/any_return.a:3:26: Adamic 0.1 refuses a void-returning function at the initializer site seen as () => any; declare the function's real return type (adamic/no-void-widening)` |
| assignment.a | `internal/oracle/testdata/void_widening_refused/assignment.a:4:9: Adamic 0.1 refuses a void-returning function at the assignment site seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| async_return_unknown.a | `internal/oracle/testdata/void_widening_refused/async_return_unknown.a:3:56: Adamic 0.1 refuses a void value at a return site whose function returns unknown; declare the function's real return type (adamic/no-void-widening)` |
| container_array.a | `internal/oracle/testdata/void_widening_refused/container_array.a:4:42: Adamic 0.1 refuses a void-returning function at the initializer site through [] seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| container_interface.a | `internal/oracle/testdata/void_widening_refused/container_interface.a:5:37: Adamic 0.1 refuses a void-returning function at the initializer site through <1> seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| container_nested.a | `internal/oracle/testdata/void_widening_refused/container_nested.a:4:77: Adamic 0.1 refuses a void-returning function at the initializer site through .list[].run seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| container_object.a | `internal/oracle/testdata/void_widening_refused/container_object.a:4:47: Adamic 0.1 refuses a void-returning function at the initializer site through .run seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| container_promise.a | `internal/oracle/testdata/void_widening_refused/container_promise.a:4:38: Adamic 0.1 refuses a void-returning function at the initializer site through <1> seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| container_recursive.a | `internal/oracle/testdata/void_widening_refused/container_recursive.a:6:20: Adamic 0.1 refuses a void-returning function at the initializer site through .run seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| container_tuple.a | `internal/oracle/testdata/void_widening_refused/container_tuple.a:4:48: Adamic 0.1 refuses a void-returning function at the initializer site through [0] seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| container_union_source.a | `internal/oracle/testdata/void_widening_refused/container_union_source.a:5:47: Adamic 0.1 refuses a void-returning function at the initializer site through .run seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| container_union_target.a | `internal/oracle/testdata/void_widening_refused/container_union_target.a:4:77: Adamic 0.1 refuses a void-returning function at the initializer site through .run seen as () => number; declare the function's real return type (adamic/no-void-widening)` |
| empty_object_return.a | `internal/oracle/testdata/void_widening_refused/empty_object_return.a:3:42: Adamic 0.1 refuses a void-returning function at the initializer site seen as (() => void) \| (() => {}); declare the function's real return type (adamic/no-void-widening)` |
| generic_union.a | `internal/oracle/testdata/void_widening_refused/generic_union.a:6:26: Adamic 0.1 refuses a void-returning function at the argument site seen as (() => void) \| (() => number); declare the function's real return type (adamic/no-void-widening)` |
| rest_argument.a | `internal/oracle/testdata/void_widening_refused/rest_argument.a:8:6: Adamic 0.1 refuses a void-returning function at the argument site seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| return_unknown.a | `internal/oracle/testdata/void_widening_refused/return_unknown.a:4:9: Adamic 0.1 refuses a void value at a return site whose function returns unknown; declare the function's real return type (adamic/no-void-widening)` |
| shorthand.a | `internal/oracle/testdata/void_widening_refused/shorthand.a:3:48: Adamic 0.1 refuses a void-returning function at the initializer site through .make seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| v01_widen_unknown.a | `internal/oracle/testdata/void_widening_refused/v01_widen_unknown.a:4:30: Adamic 0.1 refuses a void-returning function at the initializer site seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| v01_widen_unknown.ts | `internal/oracle/testdata/void_widening_refused/v01_widen_unknown.ts:4:30: Adamic 0.1 refuses a void-returning function at the initializer site seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| v02_widen_return.a | `internal/oracle/testdata/void_widening_refused/v02_widen_return.a:4:9: Adamic 0.1 refuses a void-returning function at the return site seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| v03_widen_explicit_generic.a | `internal/oracle/testdata/void_widening_refused/v03_widen_explicit_generic.a:7:30: Adamic 0.1 refuses a void-returning function at the argument site seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| v04_widen_union.a | `internal/oracle/testdata/void_widening_refused/v04_widen_union.a:4:47: Adamic 0.1 refuses a void-returning function at the initializer site seen as (() => void) \| (() => number); declare the function's real return type (adamic/no-void-widening)` |
| v05_widen_object_field.a | `internal/oracle/testdata/void_widening_refused/v05_widen_object_field.a:7:20: Adamic 0.1 refuses a void-returning function at the initializer site through .run seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| v06_widen_array.a | `internal/oracle/testdata/void_widening_refused/v06_widen_array.a:4:34: Adamic 0.1 refuses a void-returning function at the initializer site through [] seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| v10_widen_param.a | `internal/oracle/testdata/void_widening_refused/v10_widen_param.a:7:18: Adamic 0.1 refuses a void-returning function at the argument site seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| w01_void_to_union_return.a | `internal/oracle/testdata/void_widening_refused/w01_void_to_union_return.a:4:35: Adamic 0.1 refuses a void-returning function at the initializer site seen as () => void \| number; declare the function's real return type (adamic/no-void-widening)` |
| w03_union_void_returns_void_call.a | `internal/oracle/testdata/void_widening_refused/w03_union_void_returns_void_call.a:5:9: Adamic 0.1 refuses a void value at a return site whose function returns void \| number; declare the function's real return type (adamic/no-void-widening)` |

Source Node, generated JavaScript, and sanitized native agree on the accepted fixture with exit 0, empty stderr, and this stdout:

```text
done
number
4
none
discarded
```

Both restored async forwarding fixtures and regexp_replace/callback.a also pass the focused oracle. All 27 refusal subtests pass. Full focused output is in logs/focused.log.

Mutants use Go overlays, leaving repository source unchanged. The nil-refusal mutant fails all 27 refusal subtests. The union-admitting mutant fails v04's refusal subtest; its native and generated JavaScript builds succeed and both print undefined, while source Node prints number (all exits 0, no stderr). The mutant without structural recursion fails container_array.a's refusal subtest; sanitized native build succeeds, then execution exits 1 with ASan SEGV in adamic_union_typeof, while source Node exits 0 and prints number. Its generated JavaScript prints number. Raw output, overlays, and mutation sources are retained.

Linux x86_64, clang 20.1.8, Go 1.27.1, Node 24.19.0. Formatting and vet logs are empty. Required gates passed:

```sh
gofmt -l cmd internal
go vet ./internal/lower ./internal/oracle
go test -count=1 -timeout 30m ./internal/lower ./internal/fresh ./internal/flow
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle
```

The first uncached run started before counts regeneration finished, cached the missing new record, and failed 46 top-level tests for that single cause. oracle-before-counts.log preserves it. The final run started after the record existed and passed. No code changed between these two runs.

Only internal/oracle/testdata/counts/runtime/internal/oracle/testdata/void_widening_legal.a.counts was added. Its single row is allocations/frees/retains/releases/peak/regions = 8/8/7/18/6/0. No existing runtime or predicate file moved. counts.md is absent.

The sweep built the exact requested base and the implementation commit, then ran adamic c with both compilers on 3709 unique programs, excluding .d.ts, cohere/, and node_modules. This includes all 3611 current tracked programs and all 3699 programs in the prior manifest. The requested base lacks 98 historical programs; those were read from the preserved 141fec77 checkout instead of restored to this branch. The 10 additions are the new probes. All source roots and full observations appear in admission.jsonl. No review file was written. Timeouts: 0. First stops became this refusal in 27 programs:

- `internal/oracle/testdata/void_widening_refused/any_return.a`
- `internal/oracle/testdata/void_widening_refused/assignment.a`
- `internal/oracle/testdata/void_widening_refused/async_return_unknown.a`
- `internal/oracle/testdata/void_widening_refused/container_array.a`
- `internal/oracle/testdata/void_widening_refused/container_interface.a`
- `internal/oracle/testdata/void_widening_refused/container_nested.a`
- `internal/oracle/testdata/void_widening_refused/container_object.a`
- `internal/oracle/testdata/void_widening_refused/container_promise.a`
- `internal/oracle/testdata/void_widening_refused/container_recursive.a`
- `internal/oracle/testdata/void_widening_refused/container_tuple.a`
- `internal/oracle/testdata/void_widening_refused/container_union_source.a`
- `internal/oracle/testdata/void_widening_refused/container_union_target.a`
- `internal/oracle/testdata/void_widening_refused/empty_object_return.a`
- `internal/oracle/testdata/void_widening_refused/generic_union.a`
- `internal/oracle/testdata/void_widening_refused/rest_argument.a`
- `internal/oracle/testdata/void_widening_refused/return_unknown.a`
- `internal/oracle/testdata/void_widening_refused/shorthand.a`
- `internal/oracle/testdata/void_widening_refused/v01_widen_unknown.a`
- `internal/oracle/testdata/void_widening_refused/v01_widen_unknown.ts`
- `internal/oracle/testdata/void_widening_refused/v02_widen_return.a`
- `internal/oracle/testdata/void_widening_refused/v03_widen_explicit_generic.a`
- `internal/oracle/testdata/void_widening_refused/v04_widen_union.a`
- `internal/oracle/testdata/void_widening_refused/v05_widen_object_field.a`
- `internal/oracle/testdata/void_widening_refused/v06_widen_array.a`
- `internal/oracle/testdata/void_widening_refused/v10_widen_param.a`
- `internal/oracle/testdata/void_widening_refused/w01_void_to_union_return.a`
- `internal/oracle/testdata/void_widening_refused/w03_union_void_returns_void_call.a`

Exit pairs (base, head): `{'(0, 0)': 2201, '(1, 1)': 1477, '(2, 2)': 9, '(0, 1)': 22}`. There are 9 existing compiler panics on both builds, with 9 matching after normalizing build roots and addresses. Other exit-status changes outside the intended refusals: 0. Full stacks and current/original corpus subtotals are recorded in admission-other-movement.json. The head build records the delivered SHA with vcs.modified=false; the base uses -buildvcs=false because its separate worktree links the pinned cohere dependency. Build provenance and binary hashes are retained.

The prior 127-site tsc pattern inventory is preserved and broken out by code shape, with source examples, in TSC-PATTERNS.md and tsc-patterns.json. It is a stock TypeScript 6.0.3 checker-API inventory at upstream 050880ce59e30b356b686bd3144efe24f875ebc8, not a new structural-site count or a pristine tsc acceptance claim. The same two checker diagnostics remain in the raw inventory. No rule was narrowed.

Known limitation retained from the previous report: a runtime-built string discarded through a void view leaks 223 bytes in three allocations on both old main and ad029e0c. That separate ownership defect and the later value-preserving ABI fix are outside this refusal change; prior evidence/compiler/lower-void-widening retains the witness and sanitizer logs. The supplied legal fixture uses number and literal-string forms.

Not run: the whole-repository ./... gate, a native build of tsc, or a new structural pattern inventory of all tsc sources. The requested package gates, full uncached oracle, original manifest, current tracked corpus, and all three mutants were run. No main push, force push, or review/ write was performed.

Implementation diff:

```text
0dfe22fb never cut alone: lands with chg_tsr07c1d (the counts split, 6c69f66f), because its counts are the split's per-fixture files.
 internal/lower/functions.go                        |   7 +
 internal/lower/refusals.go                         |   4 +
 internal/lower/void_widening.go                    | 229 +++++++++++++++++++++
 .../oracle/testdata/void_widening_legal.a.counts   |   1 +
 internal/oracle/testdata/regexp_replace/callback.a |   2 +-
 internal/oracle/testdata/void_widening_legal.a     |  38 ++++
 .../testdata/void_widening_refused/any_return.a    |   4 +
 .../void_widening_refused/any_return.a.refused     |   1 +
 .../testdata/void_widening_refused/assignment.a    |   5 +
 .../void_widening_refused/assignment.a.refused     |   1 +
 .../void_widening_refused/async_return_unknown.a   |   4 +
 .../async_return_unknown.a.refused                 |   1 +
 .../void_widening_refused/container_array.a        |   6 +
 .../container_array.a.refused                      |   1 +
 .../void_widening_refused/container_interface.a    |   6 +
 .../container_interface.a.refused                  |   1 +
 .../void_widening_refused/container_nested.a       |   6 +
 .../container_nested.a.refused                     |   1 +
 .../void_widening_refused/container_object.a       |   5 +
 .../container_object.a.refused                     |   1 +
 .../void_widening_refused/container_promise.a      |   6 +
 .../container_promise.a.refused                    |   1 +
 .../void_widening_refused/container_recursive.a    |   7 +
 .../container_recursive.a.refused                  |   1 +
 .../void_widening_refused/container_tuple.a        |   5 +
 .../container_tuple.a.refused                      |   1 +
 .../void_widening_refused/container_union_source.a |   6 +
 .../container_union_source.a.refused               |   1 +
 .../void_widening_refused/container_union_target.a |   5 +
 .../container_union_target.a.refused               |   1 +
 .../void_widening_refused/empty_object_return.a    |   4 +
 .../empty_object_return.a.refused                  |   1 +
 .../testdata/void_widening_refused/generic_union.a |   6 +
 .../void_widening_refused/generic_union.a.refused  |   1 +
 .../testdata/void_widening_refused/rest_argument.a |   8 +
 .../void_widening_refused/rest_argument.a.refused  |   1 +
 .../void_widening_refused/return_unknown.a         |   6 +
 .../void_widening_refused/return_unknown.a.refused |   1 +
 .../testdata/void_widening_refused/shorthand.a     |   4 +
 .../void_widening_refused/shorthand.a.refused      |   1 +
 .../void_widening_refused/v01_widen_unknown.a      |   6 +
 .../v01_widen_unknown.a.refused                    |   1 +
 .../void_widening_refused/v01_widen_unknown.ts     |   6 +
 .../v01_widen_unknown.ts.refused                   |   1 +
 .../void_widening_refused/v02_widen_return.a       |   7 +
 .../v02_widen_return.a.refused                     |   1 +
 .../v03_widen_explicit_generic.a                   |   8 +
 .../v03_widen_explicit_generic.a.refused           |   1 +
 .../void_widening_refused/v04_widen_union.a        |   6 +
 .../v04_widen_union.a.refused                      |   1 +
 .../void_widening_refused/v05_widen_object_field.a |   8 +
 .../v05_widen_object_field.a.refused               |   1 +
 .../void_widening_refused/v06_widen_array.a        |   7 +
 .../v06_widen_array.a.refused                      |   1 +
 .../void_widening_refused/v10_widen_param.a        |   7 +
 .../v10_widen_param.a.refused                      |   1 +
 .../w01_void_to_union_return.a                     |   6 +
 .../w01_void_to_union_return.a.refused             |   1 +
 .../w03_union_void_returns_void_call.a             |   8 +
 .../w03_union_void_returns_void_call.a.refused     |   1 +
 internal/oracle/void_widening_test.go              |  72 +++++++
 61 files changed, 541 insertions(+), 1 deletion(-)
```

The preserved ad029e0c inventory contains 127 direct relation sites in 11 files, from TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. This is the prior pattern inventory, not a new count of structural relations or a claim that tsc compiles under Adamic. The checker loaded 78 inputs and reported the two diagnostics in typescript-impact-ad029.json. The refusal rule was not narrowed.

| Code shape | Sites |
| --- | ---: |
| `void callback passed to forEach` | 57 |
| `void callback passed to forEachChild` | 15 |
| `void callback passed to forEachType` | 4 |
| `void fallback assigned to an any-returning timer field` | 1 |
| `void return into void \| WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram>` | 13 |
| `void return into void \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram>` | 5 |
| `void return into number \| void \| Symbol` | 32 |

void callback passed to forEach (57 sites):

```ts
// src/compiler/utilities.ts:8579
forEach(objectAllocatorPatchers, fn => fn(objectAllocator));
// src/compiler/factory/nodeFactory.ts:1164
forEach(nodeFactoryPatchers, fn => fn(factory));
```

void callback passed to forEachChild (15 sites):

```ts
// src/compiler/parser.ts:1313
forEachChild(node, addWorkItem, addWorkItem); // By using a stack above and `unshift` here, we emulate a depth-first preorder traversal
// src/compiler/parser.ts:10112
forEachChild(node, visitNode as (node: Node) => void, visitArray as (nodes: NodeArray<Node>) => void);
```

void callback passed to forEachType (4 sites):

```ts
// src/compiler/checker.ts:14832
forEachType(getLowerBoundOfKeyType(constraintType), addMemberForKeyType);
// src/compiler/checker.ts:14838
forEachType(propNameType, t => addMemberForKeyTypeWorker(keyType, t));
```

void fallback assigned to an any-returning timer field (1 site):

```ts
// src/compiler/watch.ts:679
setTimeout: maybeBind(system, system.setTimeout) || noop,
// src/compiler/core.ts:1794, the fallback declaration used at the sole site
export function noop(_?: unknown): void {}
```

This category has one relation site. The second example line is its fallback definition, not a second counted site.

void return into void | WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> (13 sites):

```ts
// src/compiler/executeCommandLine.ts:579
return sys.exit(ExitStatus.DiagnosticsPresent_OutputsSkipped);
// src/compiler/executeCommandLine.ts:584
return sys.exit(ExitStatus.Success);
```

void return into void | SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> (5 sites):

```ts
// src/compiler/executeCommandLine.ts:833
return sys.exit(ExitStatus.DiagnosticsPresent_OutputsSkipped);
// src/compiler/executeCommandLine.ts:839
return sys.exit(ExitStatus.Success);
```

void return into number | void | Symbol (32 sites):

```ts
// src/compiler/binder.ts:2871
return checkContextualIdentifier(node as Identifier);
// src/compiler/binder.ts:2882
return checkPrivateIdentifier(node as PrivateIdentifier);
```

All 127 records, including source/target types and grouped source lines, are in tsc-patterns.json. Callback patterns are early-exit traversal APIs that accept return types wider than void. The command-line patterns forward sys.exit() from watch/solution-builder union-returning functions. The binder pattern forwards void helpers from a number/void/Symbol dispatcher. The timer pattern supplies noop where the optional setTimeout field returns any.
