Repaired the assigned internal/oracle gate family for lowering-chain steps 21 and 24.
Code commit: 2b2b1095; verified merged tree: b60f08f9, with origin/main 76c59c81.
Uncached selected oracle tests passed in 3.485s; call-target guard and vet passed; lane checks passed.
Mutants caught: omitted registration, removed stack guard, altered WASI stdout/function name, removed symbol guard, bypassed generic return proof.
Not covered: the other workers' families, whole packages, the full gate, or unselected WASI fixtures.

Started from the explicitly requested 379dbfb3, rather than the later after-chain tip.
No other worker's unlanded branch was merged. Current main was merged into this delivery branch.

Every row below passed on b60f08f9, carrying code commit 2b2b1095. The evidence-only commit does not change the tested code.

| Test leaf | Result | Seconds | Verified SHA |
|---|---|---:|---|
| TestEnumInitializationUnknownPinned/unknown-before | PASS | 1.82 | b60f08f9 |
| TestEnumInitializationUnknownPinned/unknown-callback | PASS | 1.29 | b60f08f9 |
| TestEnumInitializationUnknownPinned/unknown-property | PASS | 1.49 | b60f08f9 |
| TestParserNamespaceClassRegistrationMutant | PASS | 2.21 | b60f08f9 |
| TestParserConstructionUnsetUse | PASS | 1.54 | b60f08f9 |
| TestParserStackTinyLimit | PASS, 64 and 128 KiB, release and sanitized | 0.38 | b60f08f9 |
| TestGenericBodyRelationsMaybeBind | PASS | 0.33 | b60f08f9 |
| TestCheckedJSONNextRefusals/json_dictionary_alias_write | PASS | 0.19 | b60f08f9 |
| TestWASIAgreesWithNode, weak_single_narrowed leaf | PASS | 0.34 | b60f08f9 |
| TestWASIAgreesWithNode, try_stack leaf | PASS | 0.35 | b60f08f9 |
| TestWASIRuledWeakStopMutant (new) | PASS, mutant caught | 1.07 | b60f08f9 |
| TestWASIRuledStackStopMutant (new) | PASS, mutant caught | 1.09 | b60f08f9 |
| TestParserStackGuardMutant | PASS, mutant caught | 0.58 | b60f08f9 |
| TestRuledBackendOutcomes | PASS | 1.62 | b60f08f9 |
| TestTerminalStackStop | PASS | 1.06 | b60f08f9 |

Source Node confirms exit 1 and stdout before for all three unknown-enum witnesses.
The compiler's uncaught readiness errors deliberately omit engine-specific stderr.
The registration mutant now runs inside a catch, after a positive control checked against source Node.
Omitting DebugTypeMapper registration prints exactly Cannot access 'DebugTypeMapper' before initialization,
exits 0 through the catch, and differs from Node's successful registration output. Both backends catch it;
both the control and the caught native mutant pass leak checks.

For factory-use, source Node exits 1 with TypeError after true. The separately pinned compiler-owned
unset-use diagnostic remains terminal with exit 70 and names the field, factory, location, and expected type.
At tiny native stack limits, exception_repeat is the first guarded function exhausted while constructing
parser input. The test now pins that precise name. Node cannot safely start under these limits; its
ordinary-stack overflow behavior is separately checked by the existing stack controls.

The original maybeBind .a still refuses explicit any. A temporary .a variant replacing only its
any[] constraint with unknown[] exposes the dependent OmitThisParameter return refusal. The test checks
both refusals, preserving the later check rather than accepting any earlier refusal.

The WASI comparison applies the existing exact backend ruling table. No ruling was added or widened.
V8 keeps the weak target alive; native reference counting frees it. Native stack stops name depth;
the JavaScript terminal guard keeps V8's unnamed message. Existing source-Node ruling controls also pass.

The alias-write failure was a real compiler panic in placeholder discovery, before its intended refusal.
A property of an any-valued receiver has no declared symbol. A minimal guard in internal/lower/placeholder.go
avoids querying its nonexistent symbol type and allows the original checked-JSON alias refusal to fire.
Scope assumption: the brief's instruction to fix real bugs authorizes this necessary seven-line lowering
repair for the assigned oracle witness. No prohibited lowering/native orchestration files were edited.

| Mutant | Observation and catcher |
|---|---|
| Omit actual namespace class registration in IR | Both backends print the exact caught ReferenceError message; Node stdout comparison rejects it. |
| Remove emitted stack guards | Compiles and runs; ASan stack-overflow, exit 1; TestParserStackGuardMutant catches it. |
| Change actual weak fixture string before to before! | Real WASI artifact exits 70 and prints before!; exact ruled native stdout comparison rejects it. |
| Rename actual depth function to wrongDepth | Real WASI artifact's terminal message names wrongDepth; exact ruled native stderr comparison rejects it. |
| Remove new nil-symbol guard | TestCheckedJSONNextRefusals/json_dictionary_alias_write fails with the original checker nil dereference, go test exit 1. |
| Bypass genericBodyAllows in the return refusal | TestGenericBodyRelationsMaybeBind fails its dependent-return assertion with the later dynamic-this NotYet, go test exit 1. |

The two transient lowering mutants were restored after independent runs. Their patches and full logs are
in this directory. No mutant was caught solely by a compiler warning. Existing exact ruling controls also
reject changes to each side's exit, stdout and stderr. There are no new registered fixtures or counts rows;
counts.md did not need regeneration.

Commands used the environment file /workspace/adamic-tools/env.sh. Long-running shell commands carried
an outer timeout 90s; each go test also carried -timeout 90s. Test output was written to files and then read.

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout 90s bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
timeout 90s bash cloud/setup.sh --wasi-sdk
```

The first setup reached Go 0.055s, Node 0.071s, clang 0.410s, Markdown dependencies 1.032s,
and submodules 21.382s, then exited 124 during cache warming. The WASI setup reached its SDK at
21.994s, then also exited 124 during cache warming. Two initial focused test builds hit the same shell
limit while compiling dependencies. Bounded retries warmed those dependencies and the third baseline
completed, exposing the recorded failures. The final setup completed: Node ready 0.028s, Go ready 0.029s, submodules ready 0.073s,
Markdown dependencies ready 0.078s, clang ready 0.189s, WASI SDK ready 0.205s, Go build ready
42.654s, test binaries deferred 42.804s, build cache warm 42.806s, done 42.839s.
Its exact lines are recorded in setup-complete.log.
The first fetch did not populate origin/compiler/after-chain under the checkout's narrow fetch refspec;
an explicit git fetch origin compiler/after-chain resolved it, and the branch was reset to the requested
379dbfb3 before any edits. nproc is 5; cpu.max is 400000 100000, a four-CPU quota.

The baseline and focused commands selected these tests only:

```sh
go test ./internal/oracle -run 'TestEnumInitializationUnknownPinned|TestParserNamespaceClassRegistrationMutant|TestParserConstructionUnsetUse|TestParserStackTinyLimit|TestGenericBodyRelationsMaybeBind|TestCheckedJSONNextRefusals/json_dictionary_alias_write' -count=1 -v -timeout 90s
go test ./internal/oracle -run 'TestEnumInitializationUnknownPinned|TestParserNamespaceClassRegistrationMutant|TestParserConstructionUnsetUse|TestParserStackTinyLimit' -count=1 -v -timeout 90s
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run 'TestWASIAgreesWithNode/shard-[0-9]+/internal/oracle/testdata/(4ddd17f_weak_single_narrowed.a|d96d304_try_stack.a)$' -count=1 -v -timeout 90s
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run 'TestWASIRuled.*Mutant|TestParserNamespaceClassRegistrationMutant|TestGenericBodyRelationsMaybeBind|TestParserStackGuardMutant' -count=1 -v -timeout 90s
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestParserNamespaceClassRegistrationMutant -count=1 -v -timeout 90s
```

The final focused command was run before and after merging main, with output in final-focus.log and
merged-focus.log respectively. Both passed, in 3.011s and 3.485s:

```sh
ADAMIC_GATE_UNCACHED=1 ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run 'TestEnumInitializationUnknownPinned|TestParserNamespaceClassRegistrationMutant|TestParserConstructionUnsetUse|TestParserStackTinyLimit|TestGenericBodyRelationsMaybeBind|TestCheckedJSONNextRefusals/json_dictionary_alias_write|TestWASIAgreesWithNode/shard-[0-9]+/internal/oracle/testdata/(4ddd17f_weak_single_narrowed.a|d96d304_try_stack.a)$|TestWASIRuled.*Mutant|TestParserStackGuardMutant|TestRuledBackendOutcomes|TestTerminalStackStop' -count=1 -v -timeout 90s
go test ./internal/ir -run TestCallTargetReaders -count=1 -v -timeout 90s
go vet ./internal/oracle ./internal/lower
git diff --check
git fetch -q origin main devtools/fast-gate cloud/merge-tree
git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

TestCallTargetReaders passed in 26.300s. Vet and diff --check exited 0 without diagnostics.
Lane output: lane checks 6.8 s: gofmt and tools on 275 Go files, t.Parallel on 22 test packages; vet 22 packages.
The two lowering source mutants ran the corresponding single-test commands above and each exited 1.
No whole package test run or full gate was run. No cohere code was copied.
