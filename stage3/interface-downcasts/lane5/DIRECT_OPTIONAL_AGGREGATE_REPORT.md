Built: original ranks 21 and 32 certified, 14 fixtures, and a callable traversal binding-pattern panic fixed.
Commits: implementation 7175993b699064b613b0d39f9d971e223d152d44; continued from 164dac2c24a2302f6e55d543d04c7e5f8e10798e.
Checks: Node, native release, ASan/UBSan, JavaScript, leak checks, focused guards, vet, original-span verification and lane counts pass; whole-table counts remain blocked.
Mutants: native/JavaScript arity, result and parameter checks; deferred payload registration; implementation-name traversal; two original-declaration mutations all caught.
Uncovered: 20/2818 conservative pairs and 910/11063 candidate reads certified; 2798 pairs and 10153 reads remain. Static totals remain 4/308 pairs and 34/1503 reads.

Reporting date: October 12, as requested. Execution environment date: October 8.

Continued the pushed branch codex/views-callables at its full resolved SHA above, after reading CLAUDE.md and BOXING_REPORT.md, AGGREGATE_REPORT.md, MIXED_AGGREGATE_REPORT.md, ARRAY_CALLABLE_REPORT.md and the integration checkpoint. The requested ranks 6, 7, 22 and 26 were already certified in those reports.

This batch certifies NodeFactory.createVariableStatement (rank 21, 60 candidate reads) and NodeFactory.createParameterDeclaration (rank 32, 45 candidate reads). Complete original method declarations and original direct reads are retained; neighboring helper bodies and aggregate carriers are reduced. Optional arguments, explicit undefined, array/object alternatives and boxed aggregate reads are exercised. Each family has good, optional-values, wrong-value, wrong-arity, wrong-result, wrong-members and wrong-parameter-payload fixtures. Good executions match Node across release native, sanitized native and JavaScript; leak checks pass. Invalid callable signatures refuse at the member read; wrong payload values refuse at the later reached name.value or modifiers[0]!.value read.

Observation: original rank 17 createVariableDeclaration and rank 18 createVariableDeclarationList source reads are binding elements in parser.ts, not direct property reads. The retained fixtures match Node output 9 but lowering refuses unsupported destructuring representation conversion. These two pairs (136 candidate reads) are not certified. The verifier now recognizes original binding spans as well as property spans. No method-detachment or callback calling-convention refusal was relaxed. The Union callable exception and entry-live-mutation.a were left for the integrator.

The required whole-table counts run exposed a lane-owned panic: same-name producer traversal called Node.Text on a variable binding pattern. A reduced Node-backed destructured-sibling.a reproduces it. The fix compares only implementation property names. Conservative assumption: computed implementation names remain opaque and cannot prove the producer safe. The fixture now reaches the preserved rest parameter outside a nongeneric named function refusal; restoring the unsafe traversal reproduces the AST panic. No edits were made to emit.go, lower.go, native.go or oracle_test.go.

Independent truth uses TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8, source file hashes and exact declaration/read spans. direct-optional-original-witnesses.json records the two certified pairs; optional-aggregate-original-witnesses.json records the two blocked bindings. Candidate read totals describe the conservative inventory, not measured production reachability.

Commands (all test output written directly to logs):

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/lane5-continued-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'Test.*View.*Callable|TestCheckedViewCallable|TestPrepareViewCallableRead|TestDefaultTaggedInterface' -skip '^TestCheckedViewCallableCounts$' -count=1 -timeout 10m > /tmp/lane5-final-oracle.log 2>&1
go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript -run 'Test.*(Callable|CallTarget|LazyView|ArrayContract)' -count=1 -timeout 5m > /tmp/lane5-final-guards.log 2>&1
go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript > /tmp/lane5-final-vet.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-continued-lane-counts-update.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m > /tmp/lane5-final-lane-counts.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-final-all-counts.log 2>&1
node stage3/interface-downcasts/lane5/verify-aggregate-witnesses.cjs /workspace/scratch/lane5-ts-pin 21,32 direct-optional-original-witnesses.json
node stage3/interface-downcasts/lane5/verify-aggregate-witnesses.cjs /workspace/scratch/lane5-ts-pin 17,18 optional-aggregate-original-witnesses.json
node stage3/interface-downcasts/lane5/verify-optional-aggregate-fixtures.cjs /workspace/scratch/lane5-ts-pin
node stage3/interface-downcasts/lane5/verify-optional-aggregate-fixtures.cjs /workspace/scratch/lane5-ts-pin optional-aggregate-original-witnesses.json
python3 stage3/interface-downcasts/lane5/run-optional-aggregate-mutants.py > /tmp/lane5-optional-mutants-restored.log 2>&1
python3 stage3/interface-downcasts/lane5/run-optional-witness-mutants.py /workspace/scratch/lane5-ts-pin > /tmp/lane5-final-witness-mutants.log 2>&1
git diff --check
```

Setup finished: Node .068s, Go .070s, clang .427s, markdown .984s, submodules 249.619s, build 446.941s, cache warm 447.035s, done 447.059s. nproc: 5; cgroup quota: four CPUs. Initial test attempted before submodules arrived failed for missing cohere/TypeScript/tsc/go.mod; setup completed without copying submodule code. Tools: Node 24.19.0, Go 1.27.1, clang 20.1.8. Final callable oracle passed in 82.269s; focused guards passed (ir 17.001s, lower 1.727s, native 8.505s, JavaScript 1.381s); vet passed. Original verification found two declarations/105 reads and two binding declarations/136 reads; fixture verification passed 14 and 2 respectively.

Mutant evidence (all mutations restored):

| Mutant | Catcher |
| --- | --- |
| native-arity | Both wrong-arity fixtures executed 9 instead of refusing the member read. |
| javascript-arity | Both wrong-arity fixtures executed 9 instead of refusing the member read. |
| native-result | Both wrong-result fixtures reached invalid aggregate reads; sanitized execution reported SEGV. |
| javascript-result | Both wrong-result fixtures failed later at node.value instead of the callable member read. |
| native-parameters | Wrong-members fixtures reached a later union panic or executed 5 rather than refusing the signature. |
| javascript-parameters | Both wrong-members fixtures executed rather than refusing the signature. |
| payload-reads | Both wrong-parameter-payload fixtures reached AddressSanitizer SEGV instead of the checked descendant read. |
| implementation-names | Destructured-sibling refusal regression caught the restored Node.Text BindingPattern panic. |
| parameter-declaration original contract | Removing undefined from modifiers was rejected by the exact original-declaration verifier. |
| variable-statement original contract | Removing the array declarationList alternative was rejected by the exact original-declaration verifier. |

Final lane counts verification passed in 17.793s. counts.md adds exactly 14 measured fixture rows; no prior rows changed. The required whole-table updater failed before this batch's runtime changes on unrelated predicate guards, nominal subclass writes, process.exit as a value, overload/undefined handling and native invalid frees. Installing the repository-pinned @types/node with npm ci --prefix stage3/api removed the missing ambient-types failure. The retry exposed the lane-owned traversal panic fixed here; the final retry completed without that panic and failed in 39.415s, including pre-existing node_fs_directory_entries entry.name checking. Its full output is retained separately. Existing failures outside the lane prevent claiming whole-table counts or full-gate success. No whole package test or full gate was run.

Remaining higher-ranked predicate, generic, overload, rest, intrinsic and optional-Boolean callable boundaries remain conservative refusals. Durable raw evidence is in logs/optional-aggregates. No PR opened, no other branch pushed.
