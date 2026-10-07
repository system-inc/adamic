Rebased all wave-20 work onto main 39638d9e2 and preserved named shared harness 41eb6eab2.
Validated source 5dc1ba68f12cd06b52db286b3286cfa4b57e0c6d; fragment source commit is now 4e4d58063; evidence commit follows.
Checks: all eleven implemented analyses PASS, 961 controls/773 findings, both corpora, sanitizers, released handles, checker, Node, registry and vet.
Mutants: eleven semantic, six checker-fact, five released-registry, three compiling named-listener and three JSON mutations caught.
Not covered: constructed-context-values source analysis, shared typed registration, own emitted JavaScript and the full repository gate.

Main advanced during verification of the prior 995a89bef push. A detached local
base combined origin/main 39638d9e278d38bb5aeae887f46d55a70e47aaad with
41eb6eab2b6de45ede0a40250765be295ee25fbd. Own work was rebased onto that
base using the earlier combined dependency base 192e69077 as the exclusion.
Three inherited shared commits were dropped because their patches were already
upstream; no conflicts and no shared source edits. Main and the named harness
are ancestors of the validated source. The branch is pushed to its own name
with an exact lease against prior remote tip 995a89bef945babce99cfb945518065f93898049.
No push targets main or an area branch.

Every implemented source analysis has an independent Go comparison after this
rebase: the original three TypeScript-eslint rules, three Nexus rules, three
core rules, JSX undefined components and JSX fragments. Each compares complete
messages, ranges, fixes and suggestions and executes native sanitizers. Frozen
corpora are the branch's 287 repository and 77 TypeScript compiler roots, with
18,485 and 5,241 complete matching output bytes. Positive controls supply the
finding coverage because the frozen corpora have zero findings.

JSX listeners declare named ast.Kind values. Entry methods consume the node
they are handed, with driver selection outside the rule and no entry-kind
string dispatch or entry refetch. Metadata checks do not claim a semantic port
for jsx-no-constructed-context-values. That analysis remains unfinished; it
is not marked parked under the HIR exception. The three older React compiler
claims remain parked. No new claims were added.

Shared typed registration remains blocked: RuleContext lacks the checker
program/lease, configuration and root manifest. The two JSX analyses use the
existing private typed context, not a placeholder shared rule. The named
harness consolidation did not add checker access. See ANALYSIS_REPORT.md for
fragment semantics, foreign declaration controls and the deliberate Go no-fix
behavior. The earlier shared JSX tree gate passed 54 source trees/45,527 bytes
on 41eb6eab2; it was not repeated after this main rebase because no shared JSX
source changed. Full repository/shared lint gates and own emitted JavaScript
were not run.

Scratch space was reclaimed by removing 2,636,861,000 bytes of explicitly named
obsolete generated binaries and checker archives. Sources, logs, fixtures,
manifests and current builds were retained. Toolchain reused from the recorded
setup: Go 1.27.1, clang 20.1.8, Node 24.19.0; setup Go 0s, clang 1s, Node 1s,
submodules 1s, cache 121s, total 121s; nproc 5, quota four cores. No fresh setup
timing is asserted. The cloud-environment runtime skill verified current
network readiness after the environment resumed; no credential values read.

Commands source /workspace/adamic-tools/env.sh and write every output to files:

```sh
ADAMIC_WAVE20_ARTIFACTS=/workspace/wave20-validation/landing-396-first ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave20-typescript go test ./stage1/cohere/typeaware -run '^TestWave20AgreementAndMutants$' -count=1 -v -timeout 30m > /tmp/wave20-396-first.log 2>&1
python3 stage1/cohere/typeaware/no-process-exit-after-output/verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/396-second --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/next/restored > /tmp/wave20-396-second.log 2>&1
python3 stage1/cohere/typeaware/prefer-promise-reject-errors/verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/396-third --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/third/valid-cases > /tmp/wave20-396-third.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-no-undef/verify.py --artifacts /workspace/wave20-validation/396-undef --compiler /workspace/wave20-validation/396-third/adamic --checker /workspace/wave20-validation/396-third/checker.a --sanitized-checker /workspace/wave20-validation/396-third/checker-asan.a > /tmp/wave20-396-undef.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-fragments/verify.py --artifacts /workspace/wave20-validation/396-fragments --compiler /workspace/wave20-validation/396-third/adamic --checker /workspace/wave20-validation/396-third/checker.a --sanitized-checker /workspace/wave20-validation/396-third/checker-asan.a > /tmp/wave20-396-fragments.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-fragments/listener_verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/396-listeners --compiler /workspace/wave20-validation/396-third/adamic --checker /workspace/wave20-validation/396-third/checker.a > /tmp/wave20-396-listeners.log 2>&1
go test ./bridge/tsgo/checker -count=1 > /tmp/wave20-396-checker.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|maps_and_text)\.a$' -count=1 -timeout 10m > /tmp/wave20-396-node.log 2>&1
go test ./stage1/cohere/lint/registry -count=1 > /tmp/wave20-396-registry.log 2>&1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > /tmp/wave20-396-vet.log 2>&1
```

All eleven semantic mutants compile, exit 0 with empty stderr, then fail Go
finding bytes: floating promise, implied eval, void operand, process output,
race timeout, blocking stream, promise reject, regex constructor, arguments
rest conversion, undefined component name and fragment pragma object. Six
raw-fact mutants alter accessed properties, callback parameters, declaration
chains, resolved callees, module records or binding origins; independent
finding bytes catch them. Five registry-retention mutants exit 0 where the
released-handle oracle requires panic 70. Three named-listener mutations
compile and differ from independently parsed Go listener bytes; three JSON
metadata mutations are checked separately. They are not semantic checks of the
unfinished constructed-context-values analysis.

The private fragment gate supplies the two decoded modes directly. Invalid raw
option decoding is not exposed by this private runner and is not covered; its
shared configuration adapter remains part of the pending registration work.

Final outputs: first gate PASS 263.477s, 48 controls/48 findings and 13,004
matching control bytes; Nexus PASS 121 cases/160 findings; core PASS 682
cases/502 findings; JSX undefined components PASS 53 cases/29 findings;
fragments PASS 57 cases/34 findings. All ordinary and sanitizer corpora pass.
Checker PASS .706s; external Node oracle PASS .607s; registry PASS .105s;
vet output empty. The named JSX declarations match 204 production bytes.

Observed whole-process native/Go seconds, with overlapping landing gates:
first repository .894963/.362192, compiler 9.520126/2.297141; Nexus repository
.365277/.265006, compiler 3.178514/.522275; core repository .365837/.217216,
compiler 2.675512/.723081; undefined components repository .651393/.423930,
compiler 3.060235/1.138764; fragments repository .681540/.416235, compiler
3.398590/1.237236. These are observations, not isolated speed benchmarks.

Exact guard outputs follow; every test's complete stdout/stderr is archived.

```text
first: wave20_test.go:119: floating: exit 0, empty stderr, byte oracle catches byte 66
first: wave20_test.go:119: eval: exit 0, empty stderr, byte oracle catches byte 6453
first: wave20_test.go:119: void: exit 0, empty stderr, byte oracle catches byte 7728
first: wave20_test.go:134: accessed-property question mutant: exit 0, empty stderr, byte oracle catches byte 466
first: wave20_test.go:134: callback-parameters question mutant: exit 0, empty stderr, byte oracle catches byte 9854
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
