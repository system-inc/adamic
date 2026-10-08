Built ESM namespace qualification through live export bindings, with readiness checks at undecided reads.
Implementation 7940545ad54af008da299b88e7ae3a1768351c53; prerequisites c41c0e06 and 9a1f14c5, current area b68b2fe1, main efe9f404 included.
Validation: full lower PASS 116.129s, uncached namespace/import-cycle oracle PASS 79.151s, counts PASS 85.873s; five hash-verified replays.
Mutants caught: two removed readiness checks and frozen export reads in both backends; wrong-unit replay assertion; existing scoped namespace mutants rerun.
Not covered: all 557/51 roots, full tsc execution, ESM namespace object representation, or the later off-main TypeScript namespace expansion.

## Sites covered

The three sampled Debug signatures are census echoes on the requested combined base: none reproduces with complete project bindings. This does not claim their enclosing functions compile successfully. Both sampled imported performance signatures reproduce before the fix and disappear afterward.

| Exact diagnostic | Selected unit | Before | After / remaining finding |
| --- | --- | --- | --- |
| core.ts:108:5, reading Debug | zipWith, core.ts:106:1 | Does not reproduce | Arrays of V; generic function as a value at debug.ts:226:74; rolled-back result binding |
| binder.ts:582:9, reading Debug | nested bindSourceFile, binder.ts:571:5 | Does not reproduce | Set of __String; structural method with statics; Path; void/undefined call; undefined-returning function |
| core.ts:878:13, reading Debug | relativeComplement, core.ts:871:1 | Does not reproduce | Arrays of T at argument/result uses; rolled-back result binding |
| binder.ts:503:5, reading performance | bindSourceFile, binder.ts:502:1 | Reproduces | No findings in the selected replay unit |
| tracing.ts:191:9, reading performance | writeEvent, tracing.ts:187:5 | Reproduces | Four call-to-PropertyAccessExpression stops at tracing.ts:192:9, 193:21, 194:19, 195:9 (fs/JSON operations) |

The two performance units also expose the same root at binder.ts:505:5 and 506:5, and tracing.ts:196:9 and 197:9. All six observed performance diagnostics disappear. These are six covered diagnostic sites in two selected units, not a new census total or proof of whole-function/backend success for tsc.

The priority totals 557 Debug and 51 performance were supplied by the user. We did not remeasure those buckets. The committed 03:52 entry census used for exact source positions has 511 unique lowering reading-Debug sites and 52 reading-performance sites; its compiler-directory census has 509 and 52. These are distinct snapshots/scopes, not inferred deltas.

`performanceCore.ts:72:16` refers to a host-global performance object. It is deliberately not one of the two imported performance samples. `Debug` is a TypeScript namespace declaration; the imported `performance` is instead an ESM namespace assembled by `_namespaces/ts.performance.ts` and re-exported from `_namespaces/ts.ts`.

## Reproduction and evidence

The complete original adapted entry project is at `/tmp/namespace-init-sys-typescript`, pinned to upstream TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. All 81 entry-project source hashes match `stage3/meter/runs/20261008T035244Z.latent-full/tsc/source-manifest.json`. The replay keeps the entire project, declarations and ancestor bindings. It selects the exact smallest eligible unit, using the existing guarded census overlay; no source reduction, binding deletion or production-loader bypass was introduced. The census remains a no-output observation on a checker-rejected program.

[Before runs](evidence/before/runs.json), [after runs](evidence/after/runs.json), complete compressed JSON findings, exact refusal text and command logs are committed under evidence. Replay exit 0 means the requested exact signature reproduced; exit 1 with replay signature did not reproduce means it did not. All five final executions exit 1 for that signature assertion. The audit separately pins the selected unit and attempted status, so absence cannot pass through a different function or a skipped unit.

Re-run against the same hash-matching adapted tree:

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/notyet-namespace-reads/replay.py /tmp/namespace-init-sys-typescript /tmp/namespace-reads-new-evidence
```

This uses the automatic `go run ./stage3/census/latent/replay` worker. `--worker PATH` reuses an overlay-built worker. `--expect-before` pins the original outcome using a before-fix worker. `LATENT_REPLAY_MUTANT_PARENT_SIBLING=1` deliberately chooses the wrong unit and must fail the selected-unit assertion. The runner validates the full source manifest before lowering.

## Implementation

The performance stop is qualification, not a pending declared namespace. The checker resolves an ESM namespace container to a SourceFile symbol; the old namespace recognizer only accepted ModuleDeclaration symbols. The ordinary object path then tried to read an unregistered `performance` local.

ESM qualifications now resolve to the export's checker symbol through namespace imports, named re-exports and star barrels. Top-level function declarations preserve their ESM instantiation-time hoisting and canonical detached identity. Mutable variable exports read the original module storage; they are never snapshots. Structural objects typed like a namespace retain their own object calls.

The existing import-cycle proof decides whether a provider finished before a module-body read. Proven initialized reads lower without readiness checks. Deferred or undecided variable reads retain `ir.Read.Checked`, including calls through a helper; Node's initialization error is matched by native and generated JavaScript. No unknown reach is silently marked initialized. There is no new call-graph walk.

ESM namespace object escape, reflection, computed access and receiver-observing functions remain named NotYet. Writes to ESM exports are checker errors. Narrowed boxed union members retain a representation NotYet on this base; ordinary local reads already have checked narrowing. The later TypeScript namespace expansion/reachability branch is absent from this area base and was not silently imported. The sampled Debug signatures are already absent with complete registration, so no TypeScript namespace initialization rule was weakened.

## Node fixtures and mutants

Six fixtures in `internal/oracle/testdata/module_namespace_reads/` are registered for the native/JavaScript/Node oracle. [node.json](node.json) records fresh independent source-Node runs.

| Entry | Node stdout | Exit | Purpose |
| --- | --- | ---: | --- |
| main.a | 0:1:2:2; true; 80:3:3 (three lines) | 0 | Barrel resolution, live writes, detached identity, helper call, structural-object isolation |
| early.a | empty | 70 | A helper reaches the export before its provider initializes |
| initialized.a | 7 | 0 | Same cycle entered with provider initialized before the helper |
| direct_early.a | empty | 70 | Direct qualified export read before initialization |
| direct_initialized.a | 9 | 0 | Same cycle entered in the proven safe order; emitted C has no value-read readiness check |
| hoisted.a | 4 | 0 | ESM function declaration can run before its provider body |

Both early entries print exactly `adamic: panic: ReferenceError: Cannot access 'value' before initialization` on stderr. Both backends match source Node, including exit and error text. Successful fixtures also pass leak checks.

Actual mutants executed:

1. In early.a, clear the real checked Read of value reached through the helper. Native and JavaScript finish with exit 0; Node's exit 70 catches both.
2. In direct_early.a, clear the checked direct namespace Read. Both backends likewise finish instead of stopping; the same Node pin catches both.
3. In main.a, replace the real mutable export reads in main with their initial zero. Both backends exit 0 with clean sanitizers, but stdout differs from Node. This proves live binding rather than snapshots.
4. Set the existing replay wrong-unit mutant. The audit exits 1 at its exact selected-unit assertion. This proves the sampled echo classification is not accepting another function's missing signature.
5. Existing wrong-scoped-function and wrong-scoped-constant namespace mutants were rerun; Node stdout catches both with clean native exits.

Mutants modify actual lowered IR or the actual replay selector. No build-warning or unrelated sanitizer failure counts as a catch.

## Validation, counts and integration

Started on codex/notyet-namespace-reads from the area tracking ref, then explicitly refreshed area/compiler because this checkout's default fetch refspec tracks only main. c41c0e06 fast-forwarded the initial base; 9a1f14c5 merged in 1427144f. Current area b68b2fe1 merged cleanly in 9c53434c. Final `git fetch origin` and `git merge --no-edit origin/main` reported Already up to date at efe9f404. No manual merge resolution, rebase or force push. No main or area branch was pushed; no PR opened.

All tests wrote directly to logs. The complete repository gate was not run under the filtered-oracle worker allowance. Exact final commands:

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/lower -count=1 -timeout 30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestModuleNamespace|TestNamespaceSemanticMutants|TestImportCycle|TestNativeAgreesWithNode/internal/oracle/testdata/(module_namespace_reads/|namespaces|import_cycles/|modules)' -count=1 -timeout 30m -v
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
go vet ./...
gofmt -l cmd internal
git diff --check
```

Full lower passed 116.129s. The uncached oracle passed 79.151s, native 66 misses/0 hits and Node 61 misses/0 hits. Counts passed 85.873s. Vet, formatting and whitespace checks produced no output. The final five-site replay audit passed after the union refusal pin; the before-fix audit confirmed the two reproductions; the selector mutant was caught. Raw evidence is committed.

Six counts rows were added; every existing row is unchanged. Counts are allocation/free/retain/release/peak/live:

| Entry | Counts |
| --- | --- |
| main.a | 13/13/4/19/7/0 |
| early.a | 0/0/0/0/0/0 |
| initialized.a | 2/2/0/2/2/0 |
| direct_early.a | 0/0/0/0/0/0 |
| direct_initialized.a | 2/2/0/2/2/0 |
| hoisted.a | 2/2/0/2/2/0 |

`bash cloud/setup.sh` succeeded. Timing lines: Node 0.098s, Go 0.119s, submodules 0.180s, markdown 0.237s, clang 0.566s, Go build 169.461s, warm cache 170.856s, total 171.377s. nproc=5, cgroup quota=4 CPUs; Node 24.19.0, Go 1.27.1, clang 20.1.8. Every build/test shell sourced `/workspace/adamic-tools/env.sh`.

The fresh apply.sh preparation completed but was not used for the exact replay: its newer adaptations changed source bytes. The preserved adapted tree above independently matched all original manifest hashes. The pre-existing untracked stage3/stricter-indexed-all directory was left untouched and excluded from commits.
