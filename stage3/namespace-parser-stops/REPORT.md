Built the parser Debug nested-method receiver and namespace class fixes; callable merge follows separately.
Base f0cc334cb4ef4352a4bce61eb6022514fd2f1b67; receiver bab96720; class implementation is this report's commit on codex/notyet-namespace-reads.
Receiver oracle PASS 2.792s, full lower PASS 224.643s; class oracle and scoped lower PASS; exact ff2d95cc probes reproduced.
Wrong receiver caught by Node stdout; omitted class registration caught by exit-70 ReferenceError in both backends; early calls pinned.
Not covered yet: callable merges, full parser execution or a new full stage-3 census.

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

## Setup

bash cloud/setup.sh succeeded: Go 0.095s, Node 0.128s, markdown 0.276s, submodules 0.364s, clang 1.083s, Go build 175.107s, warm cache 175.399s, total 175.607s. nproc=5, cgroup quota=4 CPUs; Go 1.27.1, Node 24.19.0, clang 20.1.8. Every build/test shell sourced /workspace/adamic-tools/env.sh. [Evidence](evidence/) contains raw compressed test/setup logs. No main or area branch is pushed.
