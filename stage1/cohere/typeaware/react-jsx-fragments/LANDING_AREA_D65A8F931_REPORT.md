Rebased wave-20 onto integrated lint d65a8f931, including main 39638d9e2 and harness 41eb6eab2.
Validated source 26f1d659fa51e7d4e5351512e8c6256b8c58662a; evidence commit follows on the same branch.
Checks: eleven analyses, 961 controls/773 findings, both corpora, sanitizer, released handles, checker, Node, registry and vet PASS.
Mutants: eleven semantic, six checker-fact, five released-registry, three compiling listener and three JSON mutations caught.
Not covered: constructed-context-values source, shared typed registration, own emitted JavaScript, invalid raw fragment options and full repository gate.

Fetched all origin heads and rebased onto origin/area/stage1-lint at
 d65a8f931c98655936ae04c6899f38f14862b73e, which includes the requested
50a5f105 integration. Main 39638d9e278d38bb5aeae887f46d55a70e47aaad is an
ancestor. The correct exclusion base was the previous combined dependency base
964a1de2aea5668f7edfbf775b7072549b259c0d. All 29 own commits replayed without
conflicts. Two initial attempts used broader exclusion bases, encountered
inherited shared-file conflicts and were aborted without resolutions. No
shared source was edited; integrated Linux leak helpers were retained.
Only codex/typeaware-wave-20 is pushed, with an exact lease against prior tip
0617fcad2a630ab9bf5123012ccb9e89af1635a7. No additional claims were made.

Fresh setup succeeded: Go 1.27.1 ready 0s; clang 20.1.8 ready 0s;
Node 24.19.0 ready 0s; submodules 0s; cache 153s; total 153s.
Environment /workspace/adamic-tools/env.sh; nproc 5, four-core quota.
Removed 1,171,670,253 bytes of obsolete generated build artifacts before gates
and 911,523,077 bytes of completed gate executables/archives afterward.
Sources, fixtures, logs and manifests were retained.

The first three-rule test PASS 265.327s; Nexus 121 cases/160 findings;
core 682 cases/502 findings; undefined JSX 53 cases/29 findings;
fragments 57 cases/34 findings. Both frozen corpora match: 287 repository roots
and 77 compiler roots, 18,485 and 5,241 complete output bytes, respectively.
They have zero findings; positive controls provide semantic coverage.
Every comparison includes findings, fixes and suggestions, normal and sanitized.
Checker PASS 1.172s; filtered external Node oracle PASS 21.020s;
registry PASS .072s; vet output empty. Named listener comparisons match 204 bytes.

Observed whole-process native/Go seconds (overlapping gates, not isolated benchmarks):

| Analysis suite | Repository | Compiler |
| --- | --- | --- |
| First three | 1.093988 / .468825 | 10.025984 / 1.805656 |
| Nexus | .369276 / .221291 | 2.727536 / .522960 |
| Core | .365828 / .215880 | 2.385083 / .570072 |
| JSX undefined | .573640 / .335164 | 3.102191 / .783918 |
| JSX fragments | .485756 / .429463 | 2.801996 / .621870 |

Semantic mutants alter floating promises, implied eval, void operand, process
output, race timeout, blocking stream, promise rejection, regex suggestion span,
rest parameters, component names and fragment pragma object. Each compiles,
exits 0 with empty stderr and fails only the independent finding-byte comparison.
Six fact mutants alter accessed-property, callback-parameters, declaration-chain,
resolved-callee, module-records and binding-origin; finding comparisons catch all.
Five registry-retention mutants exit 0 instead of required released-handle panic
70. Three compiled listener mutations and three JSON manifest mutations replace
an expected kind with Identifier and are caught by independent Go declarations.
Exact guard outputs and full stdout/stderr are archived beside this report.
These metadata checks do not claim a semantic port of constructed-context-values.

The shared RuleContext and settings were read again after rebase. They still
lack checker program/lease, checker configuration and a root manifest. The two
JSX kernels use their existing private typed context; shared typed registration
remains blocked. Constructed-context-values source remains unfinished and is
not falsely parked under the HIR exception. The three older React compiler
claims remain parked for their documented HIR/SSA/capture blockers. Current JSX
metadata declares named ast.Kind listeners, consuming handed nodes; no shared
registration or test harness was changed. Earlier full JSX tree checks were not
repeated; this landing runs the gates below, not the entire repository suite.

Commands (all output to logs, with the environment sourced):

```sh
bash cloud/setup.sh > /tmp/wave20-area-setup.log 2>&1
ADAMIC_WAVE20_ARTIFACTS=/workspace/wave20-validation/area-first ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave20-typescript go test ./stage1/cohere/typeaware -run '^TestWave20AgreementAndMutants$' -count=1 -v -timeout 30m > /tmp/wave20-area-first.log 2>&1
python3 stage1/cohere/typeaware/no-process-exit-after-output/verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/area-second --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/next/restored > /tmp/wave20-area-second.log 2>&1
python3 stage1/cohere/typeaware/prefer-promise-reject-errors/verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/area-third --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/third/valid-cases > /tmp/wave20-area-third.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-no-undef/verify.py --artifacts /workspace/wave20-validation/area-undef --compiler /workspace/wave20-validation/area-third/adamic --checker /workspace/wave20-validation/area-third/checker.a --sanitized-checker /workspace/wave20-validation/area-third/checker-asan.a > /tmp/wave20-area-undef.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-fragments/verify.py --artifacts /workspace/wave20-validation/area-fragments --compiler /workspace/wave20-validation/area-third/adamic --checker /workspace/wave20-validation/area-third/checker.a --sanitized-checker /workspace/wave20-validation/area-third/checker-asan.a > /tmp/wave20-area-fragments.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-fragments/listener_verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/area-listeners --compiler /workspace/wave20-validation/area-third/adamic --checker /workspace/wave20-validation/area-third/checker.a > /tmp/wave20-area-listeners.log 2>&1
go test ./bridge/tsgo/checker -count=1 > /tmp/wave20-area-checker.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|maps_and_text)\.a$' -count=1 -timeout 10m > /tmp/wave20-area-node.log 2>&1
go test ./stage1/cohere/lint/registry -count=1 > /tmp/wave20-area-registry.log 2>&1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > /tmp/wave20-area-vet.log 2>&1
```

Exact mutation guard outputs:

```text
first: wave20_test.go:119: floating: exit 0, empty stderr, byte oracle catches byte 59
first: wave20_test.go:119: eval: exit 0, empty stderr, byte oracle catches byte 6411
first: wave20_test.go:119: void: exit 0, empty stderr, byte oracle catches byte 7658
first: wave20_test.go:134: accessed-property question mutant: exit 0, empty stderr, byte oracle catches byte 459
first: wave20_test.go:134: callback-parameters question mutant: exit 0, empty stderr, byte oracle catches byte 9763
first: wave20_test.go:187: released-registry mutant: exit 0, required panic catches it
second: KILLED {"mutant": "output", "case": "case-1055269108", "byte": 121, "exit": 0, "stderr_bytes": 0}
second: KILLED {"mutant": "timer", "case": "case-1049730460", "byte": 119, "exit": 0, "stderr_bytes": 0}
second: KILLED {"mutant": "blocking", "case": "case-1105167505", "byte": 1942, "exit": 0, "stderr_bytes": 0}
second: KILLED {"mutant": "declaration-chain", "case": "case-1049730460", "byte": 119, "exit": 0, "stderr_bytes": 0}
second: KILLED {"mutant": "resolved-callee", "case": "case-1148031436", "byte": 121, "exit": 0, "stderr_bytes": 0}
second: KILLED {"mutant": "module-records", "case": "case-1170353498", "byte": 1415, "exit": 0, "stderr_bytes": 0}
second: KILLED released-registry declaration-chain required panic catches exit 0
second: KILLED released-registry resolved-callee required panic catches exit 0
second: KILLED released-registry module-records required panic catches exit 0
third: KILLED {"mutant": "promise", "case": "case-1001187216", "byte": 77, "exit": 0, "stderr_bytes": 0}
third: KILLED {"mutant": "regex-suggestion-span", "case": "case-102540695", "byte": 615, "exit": 0, "stderr_bytes": 0}
third: KILLED {"mutant": "rest", "case": "case-1124234137", "byte": 106, "exit": 0, "stderr_bytes": 0}
third: KILLED {"mutant": "binding-origin", "case": "case-1006578395", "byte": 77, "exit": 0, "stderr_bytes": 0}
third: KILLED released-registry binding-origin required panic catches exit 0
undef: KILLED component-name mutant: compiled, exit 0, empty stderr, Go bytes differ
fragments: KILLED pragma-object mutant: compiled, exit 0, empty stderr, Go bytes differ
listeners: KILLED {'rule': 'fragments', 'byte': 20, 'exit': 0, 'stderr_bytes': 0}
listeners: KILLED {'rule': 'constructed', 'byte': 105, 'exit': 0, 'stderr_bytes': 0}
listeners: KILLED {'rule': 'undef', 'byte': 164, 'exit': 0, 'stderr_bytes': 0}
listeners: KILLED manifest react/jsx-fragments first kind replaced with Identifier
listeners: KILLED manifest react/jsx-no-constructed-context-values first kind replaced with Identifier
listeners: KILLED manifest react/jsx-no-undef first kind replaced with Identifier
```
