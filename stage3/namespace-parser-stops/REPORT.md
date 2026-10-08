Built all three parser namespace probes: nested-method receiver, namespace class, and qualified callable merge.
Feature codex/notyet-namespace-reads: receiver bab96720, class e444c07b, callable implementation is this report's commit; origin/main merged before each push.
go test: post-main uncached Node/native/JavaScript oracle PASS 28.556s, full lower PASS 183.871s, counts PASS 198.865s; go vet ./... and gofmt pass.
Runtime mutants: wrong receiver, omitted registration, attached calls redirected to root, caught in both backends; source mutant disabling intrinsic guard fails both refusal pins.
Not covered: whole parser execution, a new stage-3 census, namespace object escape/reflection, namespace-owned receiver support, mutable exports or class merges.

## Reproduction

All three exact minimal programs were read from parser proof ff2d95cc and preserved unchanged as .a files beside this report. [before.json](before.json) records source Node and native build commands, exit codes, stdout and exact located diagnostics. Node exits 0 for all three. On f0cc334c, native refuses the nested receiver at 3:5, the class at 3:5, and the callable merge at 4:5. These are parser discovery rows 11/12, 13 and 14 respectively. No discovery stubs or original-source adaptations were applied here.

## Unit 1: nested method receiver

The this in native-namespace-object-receiver.a belongs to the returned object's value method. The previous containsThis walk falsely attributed it to attachFlowNodeDebugInfoWorker. namespaceOwnThis follows arrows but stops at ordinary nested functions, methods and classes. Each construct retains its existing receiver lowering. The same lexical boundary applies to ESM namespace-qualified function checks. Namespace-owned this retains the existing named receiver refusal; its inherited arrow is separately pinned.

The exact probe now prints receiver declaration loaded on Node, native and generated JavaScript. [namespace_method_receiver.a](../../internal/oracle/testdata/namespace_method_receiver.a) executes the returned method: its object's flags are 7, then 11 after mutation, while Debug.flags is 99. All three executions print `7:99` then `11`. Leak checks pass.

TestParserNamespaceReceiverMutant replaces the actual method's flags property read in IR with the namespace's 99. Both backends exit 0 with clean sanitizer output but differ from Node stdout. This mutant tests actual receiver behavior, beyond loading an unused declaration.

Commands wrote directly to logs:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestParserNamespaceReceiver' -count=1 -timeout 10m -v
go test ./internal/lower -run 'TestNamespace|TestModuleNamespace' -count=1 -timeout 10m
go test ./internal/lower -run 'TestNamespaceReceiver' -count=1 -timeout 10m
go test ./internal/lower -count=1 -timeout 30m
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
go vet ./internal/lower ./internal/oracle
```

Oracle passes 2.792s, namespace/module tests pass 12.047s, receiver-scope tests pass 3.087s, full lower passes 224.643s, counts pass 260.044s. The full repository gate was not run under the worker filtered-oracle allowance.

Two counts rows were added: exact declaration-only receiver probe 0/0/0/0/0/0; executed method 7/7/9/16/5/0 (allocation/free/retain/release/peak/live). Existing rows are unchanged.

## Unit 2: namespace class

The unchanged native-namespace-class.a probe now prints `class declaration loaded` on Node and both backends. Namespace classes are admitted into the existing scoped class registration pass, without changing name lookup or instance layouts. needsStatics now ensures namespace classes have constructor readiness storage even if they have no static members. The declaration's public binding becomes ready after its existing ordered static initialization path. Construction checks it before evaluating arguments.

[namespace_class_registration.a](../../internal/oracle/testdata/namespace_class_registration.a) executes constructors and instance methods, reading and updating namespace-private state. All three executions print `0:1:false` followed by `construct;construct;`. Leak checks pass. The actual registration mutant removes the public DebugTypeMapper Declare from Main. Both backends stop at the constructor read with exit 70 and `ReferenceError: Cannot access 'DebugTypeMapper' before initialization`; an arbitrary backend failure does not satisfy this pin.

TestNamespaceClassEarlyConstructionStaysLoud preserves the helper-mediated early construction refusal. Reopening, class/namespace merging and unsupported class shapes retain their existing loud limits. This unit does not relax the conservative namespace initialization preflight on this integrated base.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestParserNamespaceClass' -count=1 -timeout 10m -v
go test ./internal/lower -run 'TestNamespace|TestModuleNamespace|TestStatic' -count=1 -timeout 10m
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
```

Final class oracle passes 7.500s, scoped lowering passes 7.178s, counts pass 122.666s, vet and whitespace checks pass. New counts are declaration-only class 1/1/1/4/1/0 and executed class 8/8/5/18/7/0; existing counts are unchanged. Raw final oracle and counts logs are in evidence. The first runtime fixture used a mixed number/string `+` that this base does not lower; its final template-literal output avoids that independent limitation. The direct early-construction attempt was rejected by TypeScript itself; the final helper pin reaches lowering instead.

## Unit 3: qualified callable namespace

The unchanged native-callable-namespace.a prints `callable namespace loaded` on Node and both backends. One namespace may merge with a function in its same lexical scope. Direct calls, fixed qualified exports, typeof and canonical strict function identity comparisons use existing checker bindings and canonical closure values. No callable property container is emitted.

[namespace_callable_properties.a](../../internal/oracle/testdata/namespace_callable_properties.a) executes Debug.log, Debug.log.error, a detached attached-function call, identity comparisons, typeof, a fixed property, and a module-level callable merge. Node and both backends print:

```text
log:first
error:3:second
error:3:third
true:true:true:function:3
print:fourth
7:true
```

TestParserCallableNamespaceMutant redirects actual IR Call targets from the attached error function to the root log function, including its detached forwarder. Both mutated backends exit 0 with clean sanitizers and disagree with Node stdout. The mutant changes real call resolution, not just the fixture's expected text.

TestCallableNamespaceLimitsStayLoud pins object escape, reflection, export replacement, intrinsic collisions and early-call behavior. TestCallableNamespaceReceiverStaysLoud pins actual receiver use in a merged module-level function. Neither qualified nor detached receiver-dependent calls are silently lowered. The compiler rejects intrinsic export names rather than treating immutable JavaScript function metadata as an ordinary successful namespace assignment. Mutable exports, reopenings, class merges and the conservative initialization preflight remain unchanged. The final fixture places all runtime namespace declarations before execution to obey that preflight.

An additional source mutant makes callableNamespaceIntrinsic always return false. Both intrinsic and intrinsic_method refusal subtests fail with `got <nil>, want function intrinsic` (0.223s). They catch otherwise successful lowering, not a build failure. The source is restored byte-for-byte from its pre-mutation copy before final validation and commit. This mutant separately proves the new collision guard can fail.

Final validation commands wrote directly to files:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestParserNamespace|TestParserCallableNamespace|TestModuleNamespace|TestNamespaceSemanticMutants' -count=1 -timeout 10m -v
go test ./internal/lower -run 'TestNamespace|TestModuleNamespace|TestCallableNamespace' -count=1 -timeout 10m
go test ./internal/lower -count=1 -timeout 30m
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
go vet ./internal/lower ./internal/oracle
go vet ./...
gofmt -l cmd internal
git diff --check
git fetch origin
git merge --no-edit origin/main
git push origin HEAD:codex/notyet-namespace-reads
```

The final combined oracle passes 34.151s, uncached: native 28 misses/0 hits, Node 38 misses/0 hits. It rechecks all three units, prior ESM namespace readiness/live exports, and prior scoped-function/constant mutants. Scoped lowering passes 2.779s, full lowering package passes 183.871s, counts pass 198.865s. Repository-wide vet and formatting checks pass. New counts: callable declaration probe 0/0/0/0/0/0, executed callable properties 13/13/3/20/5/0; all existing rows remain unchanged. Raw package, counts, vet, formatting and main-merge evidence is preserved beside the report. The full repository gate and a new full parser run were not performed.

After restoring the intrinsic mutant, scoped namespace tests pass 30.105s. The final fetch advanced main to ffe6efc1bd30de2539c1e1416230ca650105a0c7. Merge 2d706cb3 incorporates that main into this feature branch without rebase or force. Its only changed files are documentation/velocity/landings.csv and stage1/cohere/json/progress_test.go; namespace outcomes and counts do not change. The combined uncached Node/native/JavaScript oracle and repository-wide vet are rerun after that merge; post-merge oracle passes 28.556s with native 28 misses/0 hits and Node 38 misses/0 hits; post-merge vet exits 0. Their raw logs are recorded in evidence.

| Exact parser probe | Before on f0cc334c | Unit moving it | After |
| --- | --- | --- | --- |
| native-namespace-object-receiver.a, debug.ts:429/526 | Refused, this attributed to namespace function | bab96720, 1 probe | Node/native/JavaScript match |
| native-namespace-class.a, debug.ts:739 | NotYet, namespace class member | e444c07b, 1 probe | Node/native/JavaScript match |
| native-callable-namespace.a, debug.ts:50 | NotYet, runtime merge | This callable commit, 1 probe | Node/native/JavaScript match |

These are probe outcomes, not an assertion that entire Debug or the parser now runs. Three probes move, with one separate feature-branch push per finished unit. No main or area branch was pushed or merged into.

## Setup and environment

bash cloud/setup.sh succeeded: Go 0.095s, Node 0.128s, markdown 0.276s, submodules 0.364s, clang 1.083s, Go build 175.107s, warm cache 175.399s, total 175.607s. nproc=5, cgroup quota=4 CPUs; Go 1.27.1, Node 24.19.0, clang 20.1.8. Every build/test shell sourced /workspace/adamic-tools/env.sh. [Evidence](evidence/) contains raw compressed test/setup logs. No main or area branch is pushed.
