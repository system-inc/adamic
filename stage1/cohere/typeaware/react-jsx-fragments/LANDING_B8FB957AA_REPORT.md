Rebased wave 20 onto main b8fb957aa and preserved named shared harness ab70f38d4.
Validated source: bda6cfb41; the evidence commit follows on codex/typeaware-wave-20 only.
Checks: all ten implemented analyses, sanitizers, released handles, checker, Node, registry and vet PASS.
Mutants: ten semantic, six checker-fact, five released-registry, three compiling named-listener and three JSON mutations caught.
Not covered: full repository gate, two unfinished JSX source analyses, shared checker-context registration and three parked HIR/SSA analyses.

Main advanced from c01907a7 to b8fb957aa839a9e8cb0b54279dd9864fa317bd30.
The initial generic rebase attempted to replay the entire named shared harness
history and conflicted in shared lint files. It was aborted. The 24 own commits
were rebased using --onto origin/main 192e69077, then the prior combined harness
base was merged to retain ab70f38d4 without editing shared files. Main and the
named harness are both ancestors. The rewritten branch is published only with
an exact lease against its previously pushed 40b4de8a738de7014e95f9a2df52f5ff818ac9e8.
No push targets main or any area branch.

The latest user correction specifies named ast.Kind listeners. The three current
JSX manifests and Adamic declarations now use JsxElement, JsxFragment,
JsxOpeningElement and JsxSelfClosingElement as appropriate. Independent Go
source parsing supplies the expected names. Native and sanitizer declaration
outputs match 204 bytes. Each of three mutations substitutes Identifier,
compiles, exits zero with empty stderr, and differs only at the comparison.
The three JSON mutations are caught separately. Earlier numeric-interface
blockers are superseded. These manifests remain private metadata, not complete
shared-registry descriptors, and no placeholder handlers were registered.

The JSX undefined-component private analysis passes 53 controls with 29 findings,
including both allowGlobals modes, astral text and CommonJS controls. Complete
messages, byte ranges, fixes and suggestions match Go under ordinary and
ASan/UBSan/LSan runs. Both frozen corpora match 18,485 repository bytes and 5,241
compiler bytes. The RegExp component-name mutant compiles and exits zero with
empty stderr, then fails finding bytes; the released binding-origin query
panics 70. Observed median whole-process native/Go seconds are repository
0.483496/0.268755 and compiler 2.336136/0.520752. Other landing gates overlapped
these runs, so these are observations, not isolated speed benchmarks.

Shared registration remains pending: RuleContext exposes neither the typed
checker program/lease nor its configuration and root manifest. The private
analysis uses the existing Rules checker context. JSX fragment and constructed
context analyses remain unfinished, rather than parked HIR/SSA rules. No new
rules were claimed. The three earlier React compiler rules remain parked on
native HIR/SSA/capture analysis.

Toolchain reused from the recorded setup: Go 1.27.1, clang 20.1.8, Node 24.19.0;
setup Go 0s, clang 1s, Node 1s, submodules 1s, cache 121s, total 121s; nproc 5,
quota four cores. No fresh setup timing is asserted. Obsolete generated checker
archives from explicitly named own scratch directories were removed to reclaim
1,470,004,440 bytes; sources, logs, fixtures and current builds were retained.

Commands source /workspace/adamic-tools/env.sh. Every test output goes to a
file. Exact invocations for this landing:

```sh
ADAMIC_WAVE20_ARTIFACTS=/workspace/wave20-validation/landing-b8-first ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave20-typescript go test ./stage1/cohere/typeaware -run '^TestWave20AgreementAndMutants$' -count=1 -v -timeout 30m > /tmp/wave20-b8-first.log 2>&1
python3 stage1/cohere/typeaware/no-process-exit-after-output/verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/b8-second --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/next/restored > /tmp/wave20-b8-second.log 2>&1
python3 stage1/cohere/typeaware/prefer-promise-reject-errors/verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/b8-third --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/third/valid-cases > /tmp/wave20-b8-third.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-no-undef/verify.py --artifacts /workspace/wave20-validation/b8-undef --compiler /workspace/wave20-validation/b8-adamic --checker /workspace/wave20-validation/b8-third/checker.a --sanitized-checker /workspace/wave20-validation/b8-third/checker-asan.a > /tmp/wave20-b8-undef.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-fragments/listener_verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/b8-jsx-listeners-final --compiler /workspace/wave20-validation/b8-adamic --checker /workspace/wave20-validation/harness-third/checker.a > /tmp/wave20-b8-listeners-final.log 2>&1
go test ./bridge/tsgo/checker -count=1 > /tmp/wave20-b8-checker.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|maps_and_text)\.a$' -count=1 -timeout 10m > /tmp/wave20-b8-node.log 2>&1
go test ./stage1/cohere/lint/registry -count=1 > /tmp/wave20-b8-registry.log 2>&1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > /tmp/wave20-b8-vet.log 2>&1
```

An initial metadata probe used sort() without a comparator and was correctly
refused by Adamic. Declarations were placed in canonical order and the final
probe needs no sort. This failed build is retained, not counted as a mutant.
The initial editing script stopped before updating its independent Go adapter;
that adapter was corrected before the successful metadata gate.

Final landing outputs: first PASS 205.432s, 48 controls/48 findings/12,987
identical control bytes; Nexus PASS 121 cases/160 findings; core PASS 682
cases/502 findings; JSX undefined-component PASS 53 cases/29 findings. All
ordinary and sanitizer corpora passed. Checker PASS 0.774s; representative
Node oracle PASS 2.502s; shared registry PASS 0.094s; vet output empty.
The complete repository gate and the complete shared lint gate were not run.

Nexus native/Go observed medians: repository .520449/.270126 seconds,
compiler 2.676683/.628113. Core medians: repository .365187/.217248,
compiler 2.275668/.518517. The earlier nine semantic mutants and six checker
fact mutants all compiled, exited zero, and were caught by comparison bytes;
five released-registry mutations exited zero where panic 70 was required.
The tenth semantic mutant is the JSX component RegExp mutation described above.

Evidence is in validation/landing-b8fb957aa, excluding generated executables
and checker archives. Metadata mutant names denote their claimed analyses;
they are declaration checks, not semantic mutants for the two unfinished rules.
