Completed native require-await contract decisions beside symbol-description and structure/react-hook-no-any-type.
The branch is rebased onto origin/area/stage1-lint d65a8f931 and all nine completed ports are oracle-green.
Three live rules match full Go finding/fix/suggestion bytes on controls and compiler77/repository287; require-await also matches 87 upstream sources.
Native mutants, released-handle rejection, ASan/UBSan, raw-question tests, source lint and the focused uncached Node oracle pass.
Shared registry/context integration and emitted-JavaScript comparison of the live lint runner remain outside this unit; the three React claims stay PARKED.

Earlier observations below are historical; the latest completion section supersedes their require-await blockers.

The three earlier React claims are PARKED under the user's explicit instruction.
The first six rules remain completed; their full landing observations are in
../wave_10_react/README.md. No additional rule is claimed beyond these three.

Each directory has rule.json listener metadata using canonical typescript-go
ast.Kind names and "node": true, as specified by the registry at harness commit
ab70f38d4. Symbol and hook listeners declare CallExpression; require-await declares
FunctionDeclaration, FunctionExpression, ArrowFunction, MethodDeclaration,
GetAccessor, SetAccessor and Constructor. The generated driver hands each listener
(node, index), with optional parent when requested. Numeric kind minting is not
planned and is not a blocker. All new supplied-fact kind fields now use canonical names too. The symbol/hook
Rule classes receive the handed ParseNode, and the owned live runner loads one
checker program and parses each file once.

These are partial listener declarations, not complete Discover registrations.
The old unrecognized entry field is removed. The registry additionally requires
named factories/classes, hooks, provenance, owned oracle adapters, mutants and
witnesses. Named create/Rule/visit entries are now present for symbol and hook;
full Discover packaging and the shared RuleContext checker attachment remain
unwritten. The owned live Context carries the checker handle directly. No shared registry or
harness is changed. reportNode and reportRange are present in the referenced
RuleContext; its source does not expose the live checker facts these rules need.

Symbol-description implements all verdict logic on its supplied call facts,
including parentheses-normalized callee syntax and the FIRST declaration's
source-file flag, matching the production helper rather than its broader comment.
Hook-result checking implements the ASCII uppercase/digit hook-name predicate,
raw callee versus normalized React receiver, result-type Any flag and exact policy
wording. Its blinded-rule names are supplied raw registry metadata; the test oracle
--catalog mode reads the live registry's ResolvesReactValueTypes flags without
making a lint decision. The linked catalog currently names three rules. No list
is frozen into the native rule.

require-await implements its modifier/generator/body guards and named-kind preorder
body walk, including nested function/class barriers, concise bodies, for-await,
and the correct await-using composite mask. Reporting candidates explicitly panic
with NotYet and exit 70. Contextual promise demand, declared generic signature
replay and heritage member contract decisions are NOT ported. Function-head
ranges and removeAsync suggestions now have the separate native reporting module
and reporting-only source controls described below. The current owned interface models only the body facts, not a
finished function adapter. This is incomplete work, not a complete rule hidden
behind a harness exception.

The remaining shared-harness work is attaching checker/program data to its
RuleContext and mapping these diagnostics to reportNode/reportRange. The two
owned live adapters now construct their facts directly from native parse nodes
and the existing raw checker bridge. The
handed-node contract is now specified on origin/lint-rules/harness, though it is
not on this landing base. The symbol/hook live checker adapters use existing runtime-context origin and
raw-shape questions; no new bridge question is registered. Existing bridge modes supply many relevant raw
type/origin facts; absence of a lint-verdict question is not a blocker.
No rule verdict is delegated to Go by the native modules.

Validation (source the setup environment, redirect output to a log):

```sh
python3 stage1/cohere/typeaware/wave_10_leaf/validate.py \
  --stage0 /workspace/wave-10-f801-original/adamic \
  --artifacts /workspace/wave-10-leaf-final \
  > /tmp/wave-10-leaf-final-validation.log 2>&1
```

The reusable harness materializes TypeScript oracle inputs from controls.json
and builds the unchanged production Go registry rules as overlays. New Adamic
sources are all .a; generated .ts files are TypeScript oracle inputs, not Adamic
modules. Native controls receive hand-authored raw fact projections corresponding
to those inputs. They do not parse the inputs or query a live bridge. Therefore
agreement here proves isolated verdict behavior, not adapter correctness.

Symbol: 13 controls, 4 findings, 2036 identical diagnostic bytes. Hook: 16 controls,
7 findings, 5806 identical bytes, including all message/span/fix/suggestion fields
(the latter counts are zero). Require-await: 14 context-free controls, 798 bytes
of Boolean candidate results derived from production Go's actual findings. Its
ranges/messages/suggestions are not compared. Each module passes ASan/UBSan with
empty stderr. Three source mutants compile and exit 0 with empty stderr; only
comparison catches them. Candidate refusal is separately checked at exit 70.

An initial isolated timing observation over the original artifact paths was:
symbol core native 1.784ms versus full Go 33.759ms; hook core 1.915ms versus full
Go 29.239ms. These are NOT native-lint speed comparisons: native receives facts,
whereas Go loads/parses/checks source. No full native pipeline timing is claimed.
Setup/environment is unchanged from the last landing: total 100s, nproc 5.

No native compiler77/repository287 comparison for these three, live handle release
check, full upstream matrix, emitted-JavaScript comparison or full repository gate
is claimed. This paragraph records the earlier isolated-only validation. Current live
symbol/hook coverage is below; require-await remains incomplete and no new
reservation is authorized.
Evidence contains raw stdout/stderr, parser numbers, source lint and mutant logs.

## Current landing on c01907a7

Main advances again to c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06.
The only changed path under compiler/bridge/stage1/config scope is an added
internal/oracle/stage3_hook_test.go; production compiler, bridge, stage1,
submodule and configuration sources are unchanged. Rebase succeeds cleanly,
yielding source tip 39310513. Main is fetched again after the gates and remains
c01907a7. The six completed ports are green on this base; the new three remain
partial for the reasons above.

Rerun the exact earlier six-rule gates, with the same frozen manifests and
artifact settings and -count=1: original three PASS 92.098s; timeout PASS
50.990s; final process/blocking PASS 163.061s. All controls and compiler77/
repository287 findings/fixes/suggestions agree with Go, normal and sanitized.
All six clean-exit native mutants and both retaining-registry mutants are caught;
released queries exit 70. The final gate reuses only the process compiler/bridge
archives whose production sources are unchanged. No new process bootstrap or
full repository gate is claimed. Logs are in evidence/landing-c019.

The new isolated gate reruns with --artifacts /workspace/wave-10-leaf-c019:
symbol 2023 bytes, hook 5790 bytes, require-await filter 784 bytes; normal and
sanitized results agree with their respective Go authority. Their clean-exit
mutants are caught at bytes 59, 63 and 558. Different artifact-path lengths
account for changed serialized byte positions. Candidate reporting still refuses
at exit 70; full live-node integration and reporting are not validated.

New complete-process timing observations for the previously completed process
and blocking rules below use three alternating runs and compare each stdout.
They include load/parse/checker/serialization; worker load is uncontrolled.
They are separate from the incomparable isolated-core timing above.

| Rule and corpus | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| process-controls | 3.622435 | 0.079991 | 45.286x |
| process-compiler | 1.689670 | 0.292615 | 5.774x |
| process-repository | 0.279804 | 0.172372 | 1.623x |
| blocking-controls | 5.667838 | 0.077326 | 73.298x |
| blocking-compiler | 2.148552 | 0.307177 | 6.995x |
| blocking-repository | 0.366139 | 0.131993 | 2.774x |

## Canonical listener contract correction

The user confirmed canonical kind names and the ab70f38d4 handed-node contract.
The three listener declarations now use that contract. Earlier numeric-API
blocker descriptions are superseded; adapter implementation and require-await
reporting remain incomplete. Main remains c01907a7 after fetching origin, so
this metadata/documentation correction does not change the previously green
runtime sources or require another rebase. No additional rule is claimed.

Validation uses the exact registry.go from ab70f38d4 with a scratch probe:

```sh
go run /tmp/wave-10-kind-contract/registry.go /tmp/wave-10-kind-contract/probe.go \
  stage1/cohere/typeaware/wave_10_leaf/symbol_description/rule.json \
  stage1/cohere/typeaware/wave_10_leaf/react_hook_no_any_type/rule.json \
  stage1/cohere/typeaware/wave_10_leaf/require_await/rule.json \
  > /tmp/wave-10-kind-contract-validation.log 2>&1
```

Exit 0: all three listener subsets pass canonical ast.Kind validation. The actual
registry Render function produces their named kind buckets and visit(node, index)
calls using explicitly supplied scratch factory/class placeholders. Negative
controls reject numeric kinds, an unknown name, duplicate kinds and node: false.
This checks listener metadata and generated dispatch, not full Discover or module
compilation. Evidence preserves the probe and its output. The existing native
sources are unchanged; no new runtime, sanitizer or corpus run is claimed.

## Live adapters and reporting continuation on c01907a7

The initial fetch confirmed main remained c01907a7; the final fetch advanced it
to b8fb957a, requiring the landing rerun below. Runtime changes are confined to the owned
leaf modules. No shared harness/generator/bridge registration or protected
compiler file is edited. The earlier six completed rules' runtime sources are
unchanged from their recorded green landing. Environment/setup is unchanged:
last setup total 100s; nproc again reports 5. Setup was not rerun this time.

The owned live driver dispatches only CallExpression to the selected rule. It
fetches each dispatch node once and hands that node to visit(node, index).
The listeners inspect callee/receiver syntax, not the given node's kind to decide
relevance. Symbol uses the raw first declaration's declaration-file flag; hook
uses the raw call-result type flags and the live catalog's raw rule names.
No bridge verdict or Go lint predicate supplies the native findings.

Hook names now use the RegExp literal /^use[A-Z0-9]/u. The regex branch's table
has no row for these three rules: the pinned Go hook predicate is byte comparisons
implementing that documented pattern, not a Go regexp. The new digit-removal
RegExp mutant is killed by production-Go comparison. require-await's semicolon
continuation predicate also uses a RegExp literal. None of these rules has a
user-supplied regex option.

Run the final owned gates after sourcing /workspace/adamic-tools/env.sh:

```sh
python3 stage1/cohere/typeaware/wave_10_leaf/validate.py \
  --stage0 /workspace/wave-10-f801-original/adamic \
  --artifacts /workspace/wave-10-leaf-named \
  > /tmp/wave-10-leaf-named-final.log 2>&1
python3 stage1/cohere/typeaware/wave_10_leaf/live_validate.py \
  --stage0 /workspace/wave-10-f801-original/adamic \
  --checker /workspace/wave-10-f801-original/checker.a \
  --checker-asan /workspace/wave-10-f801-original/checker-asan.a \
  --isolated /workspace/wave-10-leaf-named \
  --artifacts /workspace/wave-10-leaf-live-validation \
  --compiler-config /workspace/wave-10-typescript/src/compiler/tsconfig.json \
  --compiler-manifest /workspace/wave-10-compiler.manifest \
  --repository-manifest /workspace/wave-10-repository.manifest \
  > /tmp/wave-10-leaf-live-final.log 2>&1
python3 stage1/cohere/typeaware/wave_10_leaf/require_await/validate_reporting.py \
  --stage0 /workspace/wave-10-f801-original/adamic \
  --checker /workspace/wave-10-f801-original/checker.a \
  --checker-asan /workspace/wave-10-f801-original/checker-asan.a \
  --oracle /workspace/wave-10-leaf-named/await-oracle \
  --artifacts /workspace/wave-10-await-report-validation \
  > /tmp/wave-10-await-report-final.log 2>&1
```

All three exit 0. Live symbol controls: 13 files, 4 findings, 2036 bytes; hook:
16 files, 7 findings, 5806 bytes. Each live rule matches Go on compiler77
(5318 bytes) and frozen repository287 (18485 bytes), including complete finding,
fix and suggestion fields. Corpus findings are zero; positive controls prevent
vacuous corpus agreement. Normal and ASan/UBSan runs agree with empty native
stderr. Released-handle queries reject at exit 70 under both builds, with exact
stderr: adamic: panic: invalid or released checker handle\n.

Symbol's first-declaration mutant and hook's result-type mutant each compile,
exit 0 with empty stderr, and differ only in byte comparison at 60/64. The
isolated await-using filter mutant is caught at 568, and the hook digit-pattern
mutant at 807. The prior candidate refusal remains exit 70.

Require-await reporting has 32 source fixtures and 20279 identical Go/native
bytes, normal and sanitized. They cover function/property/arrow head ranges,
export modifiers, method names including empty/computed/Unicode names, comments
following async, and semicolon insertion versus already terminated predecessors.
A reporting-range mutant compiles, exits 0 with empty stderr and is caught at
byte 64. These are reporting-only candidates with no contextual contracts:
they do NOT validate the incomplete full require-await verdict or corpus.

Source lint over all ten owned .a modules reports findings 0. The first live
release assertion expected the wrong error label; it was corrected to the actual
checker-handle diagnostic and the entire gate rerun. Source lint initially caught
two writes to a caller-owned scanner; scanner operations now live on Context.
The final reporting and live gates rerun after that refactor.

Three alternating complete-process observations per dataset, with identical
stdout checked on every run, give these medians. Worker load is uncontrolled;
these include load/parse/checker queries/serialization, not isolated fact verdicts.

| Rule and corpus | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| symbol-controls | 0.033449 | 0.025062 | 1.335x |
| symbol-compiler | 1.381533 | 0.282475 | 4.891x |
| symbol-repository | 0.233297 | 0.129675 | 1.799x |
| hook-controls | 0.017214 | 0.023670 | 0.727x |
| hook-compiler | 1.374404 | 0.284982 | 4.823x |
| hook-repository | 0.238282 | 0.131858 | 1.807x |

Remaining work is concrete: require-await must replay declared generic targets,
substitutions and contextual steps, then evaluate promise demands and heritage
member contracts before calling reporting.a. Existing bridge function-signatures
exposes call signatures, but no question exposes a resolved signature's Target()
and TypeParameters() needed for that declared-demand replay. That new raw-fact
question and its native consumer are not implemented. No complete require-await
port, full upstream matrix, emitted-JavaScript comparison or full repository gate
is claimed. The shared harness's RuleContext at ab70f38d4 has no checker/program
attachment; integration must connect this owned live context to it. No numeric
API is expected and no new rule is claimed.

## Current landing on b8fb957a

The final pre-push fetch advanced main to b8fb957aa839a9e8cb0b54279dd9864fa317bd30.
Rebase is clean and preserves main's inherited-static-field fix in
internal/native/emit_objects.go. No developer-tool allocator-check changes occur
in this two-commit advance; none is reverted. CLAUDE is unchanged. A fresh stage-0
compiler is built at /workspace/wave-10-b8-stage0 from rebased source tip 465969a5.
The process/blocking landing artifact's adamic is replaced with that fresh binary.
The reused process checker archives have unchanged Go bridge sources on this base.

Re-green the original six ports with the frozen compiler/repository manifests:

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_WAVE10_REPOSITORY_MANIFEST=/workspace/wave-10-repository.manifest
export ADAMIC_WAVE10_COMPILER_MANIFEST=/workspace/wave-10-compiler.manifest
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-10-typescript
ADAMIC_WAVE10_ARTIFACTS=/workspace/wave-10-f801-original \
  go test ./stage1/cohere/typeaware -run '^TestWave10AgreementAndMutants$' \
  -count=1 -timeout=10m -v > /tmp/wave-10-b8-original.log 2>&1
ADAMIC_WAVE10_NEXT_VALIDATE=1 ADAMIC_WAVE10_NEXT_ARTIFACTS=/workspace/wave-10-f801-timeout \
  go test ./stage1/cohere/typeaware/wave_10_next -run '^TestTimeoutAgreementAndMutants$' \
  -count=1 -timeout=10m -v > /tmp/wave-10-b8-timeout.log 2>&1
ADAMIC_WAVE10_LANDING_ARTIFACTS=/workspace/wave-10-f801-process \
ADAMIC_WAVE10_LANDING_FIXTURES=/workspace/wave-10-process-validation \
  go test ./stage1/cohere/typeaware/wave_10_next -run '^TestLandingNativeRulesAndMutants$' \
  -count=1 -timeout=10m -v > /tmp/wave-10-b8-landing.log 2>&1
```

All pass: original 109.558s, timeout 81.747s, process/blocking 188.397s.
Finding/fix/suggestion comparisons, native sanitizer controls/corpora, all six
native mutants, two retaining-registry mutants and released queries remain green.
The three owned leaf scripts above rerun with --stage0 /workspace/wave-10-b8-stage0
and explicit --checker/--checker-asan paths under /workspace/wave-10-f801-process.
All exit 0 with the same comparison lengths and mutant byte positions as above;
32-case reporting and all release checks pass again.

An uncached focused external oracle also passes:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^(TestNativeAgreesWithNode|TestTheOracleCatchesOneByte)$/internal/oracle/testdata/regexp(_exec|_unicode)?[.]a$' \
  -count=1 -timeout=10m -v > /tmp/wave-10-b8-node-regexp.log 2>&1
```

PASS 1.516s: regexp.a, regexp_exec.a and regexp_unicode.a agree across source
Node, emitted JavaScript, native release and sanitized builds; the oracle's
one-byte mutation is caught. Cache counters show native misses=10 and node
misses=7, with zero hits. This is not an emitted-JavaScript comparison of the
new live lint runner and not a full repository gate.

Fresh alternating live-rule process timings on b8fb957a follow. These observations
run alongside the landing work; uncontrolled load prevents a performance-change
claim. Raw logs, stdout/stderr and timing samples are in evidence/landing-b8.

| Rule and corpus | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| symbol-controls | 0.033264 | 0.026076 | 1.276x |
| symbol-compiler | 1.390334 | 0.331947 | 4.188x |
| symbol-repository | 0.250274 | 0.142555 | 1.756x |
| hook-controls | 0.020389 | 0.034592 | 0.589x |
| hook-compiler | 1.514863 | 0.355884 | 4.257x |
| hook-repository | 0.250067 | 0.148570 | 1.683x |

## Completed require-await and landing on 39638d9e2

The rebase onto 39638d9e278d38bb5aeae887f46d55a70e47aaad is clean. Its advance from b8fb957a changes stage3 and landing documentation only; no compiler, bridge, stage1, toolchain or CLAUDE source changed. Stage 0 was nevertheless rebuilt as /workspace/wave-10-396-stage0. Setup reports go/clang/node/submodules ready in 0s each, build cache warm in 102s, total 102s; nproc=5 and cpu.max=400000 100000.

require_await_facts.go supplies raw function syntax, declared/resolved signatures, type-parameter identities, heritage roots, call signatures, index types and symbol property types. It contains no lint predicate. Its only shared edit is the require-await switch registration in facts.go. require_await_facts.a decodes these observations; contracts.a performs contextual walking, generic replay, stand-ins, union expansion, thenable callback checks and heritage exemption natively. The full Rule.visit takes the driver's node and index. The older supplied-facts-only visit deliberately still refuses candidates: body facts alone cannot decide a contract; this refusal does not run in the full listener.

All descriptors use canonical named kinds and node:true. No shared harness/generator or protected compiler file was edited. No new Adamic .ts file was added. No new regex matcher was needed; existing reporting and hook regexes remain JS RegExp literals. The pinned union constant is reused: TypeFlagsUnion=134217728, not standard TypeScript's 1048576. The independent Go comparison caught eight wrong findings from using the latter, and a retained native mutant proves this check.

Commands, with every test output captured to a file:

```sh
source /workspace/adamic-tools/env.sh
go test ./bridge/tsgo/checker -count=1 > /tmp/wave10-396-bridge.log 2>&1
go vet ./bridge/tsgo/checker > /tmp/wave10-396-vet.log 2>&1
/workspace/wave-10-source-gate /workspace/adamic/tsconfig.json /tmp/wave10-owned-sources.manifest > /tmp/wave10-396-source-lint.log 2>&1
python3 stage1/cohere/typeaware/wave_10_leaf/live_validate.py --stage0 /workspace/wave-10-396-stage0 --checker /workspace/wave-10-require-checker.a --checker-asan /workspace/wave-10-require-checker-asan.a --isolated /workspace/wave-10-require-isolated --artifacts /workspace/wave-10-require-396-live --compiler-config /workspace/wave-10-typescript/src/compiler/tsconfig.json --compiler-manifest /workspace/wave-10-compiler.manifest --repository-manifest /workspace/wave-10-repository.manifest > /tmp/wave10-require-396-live.log 2>&1
```

live_validate.py exits 0. Symbol controls: 13 sources, 4 findings, 2114 bytes; hook: 16 sources, 7 findings, 5902 bytes; require-await: 14 sources, 7 findings, 4866 bytes. Each rule matches compiler77 (5318 bytes) and repository287 (18485 bytes), with zero corpus findings, so the positive controls are essential. require-await also matches 87 mechanically extracted complete Go-test source programs, 59 findings and 40041 bytes. extract_upstream.go uses Go's AST and strconv.Unquote, including concatenated typed preludes. The retained JSON excludes three expected-head fragments and twelve Go-parse-invalid extracted literals. This is not every upstream test or option combination. All comparisons include full messages, ranges, fixes and suggestions, normal and ASan/UBSan. Released handles are rejected normally and sanitized, exit 70 and exact invalid-or-released panic text.

Clean-exit native mutants, caught only by the independent Go byte comparison: symbol declaration origin byte 66; hook result-any byte 70; await-using flags byte 2266; pinned union flag byte 34397; bypassing declared generic demand/self-inference byte 35762. The 32-case reporting gate matches 20215 bytes, normal/sanitized; shifting a reported range is caught at byte 62. Isolated body controls match 882 bytes; their await-using mutant differs at byte 628. Isolated symbol/hook mutants differ at 66/70; removing digits from the hook-name RegExp differs at 819. A Go overlay drops the new function-question arity guard: TestRequireAwaitRawQuestions fails with accepted-invalid-question, exit 1.

The earlier six ports are re-green on this base using the exact landing-b8 commands above with /tmp/wave10-396-* logs and fresh process archive/compiler: TestWave10AgreementAndMutants PASS 123.621s; TestTimeoutAgreementAndMutants PASS 60.369s; TestLandingNativeRulesAndMutants PASS 203.122s. Their native mutants are caught at bytes loop=548, redundant=4569, includes=8447, timeout=51, process=114, blocking=13904; both retaining-registry mutants exit 0 and are caught by requiring released-query panic 70. All normal/sanitized controls and both corpora pass.

Bridge tests PASS 0.212s; vet output empty; owned source lint findings 0. The uncached filtered RegExp Node/native/emitted-JavaScript oracle passes in 1.851s: three fixtures and the one-byte oracle mutant, native misses=10, Node misses=7, zero hits. This compares the RegExp fixtures, not the live lint runner's emitted JavaScript.

Three alternating whole-process timing samples give these medians under concurrent landing load; they are observations, not a speedup claim:

| Rule / corpus | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| symbol-controls | 0.046992 | 0.048466 | 0.970x |
| symbol-compiler | 1.887221 | 0.478514 | 3.944x |
| symbol-repository | 0.335965 | 0.217888 | 1.542x |
| hook-controls | 0.023686 | 0.036080 | 0.656x |
| hook-compiler | 1.728615 | 0.503694 | 3.432x |
| hook-repository | 0.403118 | 0.295351 | 1.365x |
| await-controls | 0.025282 | 0.031910 | 0.792x |
| await-compiler | 2.590656 | 0.486722 | 5.323x |
| await-repository | 0.356319 | 0.206339 | 1.727x |
| await-upstream | 0.050000 | 0.067608 | 0.740x |

Evidence is preserved in evidence/require-await-complete. Full repository go test ./... and the live runner's emitted-JavaScript oracle were not run. The shared harness at 41eb6eab2 still needs its checker attachment connected to this owned Context; shared wiring is reserved for its worker. Native high-level IR/SSA/capture/JSX analysis remains unavailable for the three PARKED React claims. The full require-await rule is now native-tested; no further rule was reserved before this completion was pushed.

## Post-completion availability audit

Completion and all landing evidence were pushed as d304ac858 to codex/typeaware-wave-10 before this audit; ls-remote confirmed that exact tip. A fresh explicit all-head fetch inspected 605 origin refs and 33 distinct Markdown claim blobs. The frozen combined VOLUME_REPORT ranking has 197 checker-dependent rules: 25 are existing ranked base/main ports, and all 172 others are mentioned in claims. No unclaimed ranked rule remains, so no new reservation is made.

The previously available final fifteen names were individually checked against claim lines: React candidates have explicit reservations on other waves; require-atomic-updates is reserved by wave-01/wave-08; valid-typeof is reserved by wave-05/wave-19/wave-24/wave-27. The audit preserves the exact main sha 39638d9e2 and port source paths in post-completion-availability.json. Shared registry wiring and the live lint runner's emitted-JavaScript comparison remain the exact integration limitations stated above.

## Landing on the shared lint integration branch

The requested rebase onto origin/area/stage1-lint is clean. The fetched tip is d65a8f931c98655936ae04c6899f38f14862b73e, beyond the user-named 50a5f105 merge. It contains current main 39638d9e2. Rebased implementation is 1e4aab909; source tip before this evidence commit is 9683d8a05. Nothing in the nine completed rules changed. Incoming registry, suggestion model, JSX parser and allocator/string-runtime changes are retained. No shared or protected source file was edited during this landing unit. The Go bridge sources are unchanged by the integration base, so the checked normal/sanitized C archives remain valid; stage 0 was rebuilt as /workspace/wave-10-area-stage0.

Commands are the previous landing commands with the fresh stage0 and /tmp/wave10-area-* logs. live_validate.py uses --artifacts /workspace/wave-10-area-live; validate_reporting.py uses /workspace/wave-10-area-reporting. TestWave10AgreementAndMutants passes in 125.459s, TestTimeoutAgreementAndMutants in 122.547s and TestLandingNativeRulesAndMutants in 280.999s. Normal/sanitized controls and compiler77/repository287 match full findings, fixes and suggestions. Original-suite repository output now contains 2 findings and 19044 bytes, rather than the earlier 1/18863: these are changed source contents on the same frozen root list after integration, and both authorities agree. Compiler output remains 9 findings/8346 bytes. Timeout, process and blocking corpus outputs remain 5318/18485 bytes with zero findings.

All three latest live rules pass every control/corpus comparison, ASan/UBSan, released-handle panic 70, native decision mutants and alternating benchmark output equality. Symbol controls remain 2114 bytes, hook 5902, require-await 4866, and require-await's 87 upstream sources 40041. Compiler/repository per-rule outputs remain 5318/18485 bytes. Native mutants are caught only by Go comparison at bytes symbol=66, hook=70, await-using=2266, wrong-union-flag=34397 and generic-self-inference=35762. The original six mutants are caught at loop=548, redundant=4569, includes=8447, timeout=51, process=114 and blocking=13904. Both released-registry mutants are caught by requiring panic 70. The reporting gate matches 20215 bytes normally/sanitized and kills its range mutant at byte 62.

Bridge tests pass in 0.335s, bridge vet is empty, owned source lint reports 0. go run ./cmd/lint-registry succeeds and lists all 15 incoming registered rules; generated output stays untracked. The focused uncached RegExp Node/native/emitted-JavaScript gate passes in 14.795s, including its one-byte mutant; native misses=10, Node misses=7, zero hits. Setup reports Go, clang, Node and submodules ready in 0s each; cache warm and total 109s; nproc=5, cpu.max=400000 100000.

The exact previously parked JSX reproduction now succeeds: the freshly built shared native parser accepts const frame = <iframe /> and emits JsxSelfClosingElement. The former parser and invented numeric-kind blockers are closed; canonical kind names are supported. This parser observation does not certify the three parked React rule implementations, their finding bytes or mutants. Their PARKED status is retained under the user's parking instruction; no React port is claimed here.

The incoming shared RuleContext still has no checker program or source-path attachment. Connecting this owned live Context to it remains the concrete shared-registry integration gap, and no shared files are changed to conceal it. The live lint runner's emitted-JavaScript comparison and full repository go test ./... are not covered. Detailed logs and outputs are in evidence/landing-area.

| Rule / corpus | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| symbol-controls | 0.056745 | 0.037905 | 1.497x |
| symbol-compiler | 2.044455 | 0.464823 | 4.398x |
| symbol-repository | 0.272406 | 0.215180 | 1.266x |
| hook-controls | 0.023760 | 0.029407 | 0.808x |
| hook-compiler | 1.568316 | 0.420295 | 3.731x |
| hook-repository | 0.419808 | 0.239600 | 1.752x |
| await-controls | 0.050397 | 0.051979 | 0.970x |
| await-compiler | 2.388200 | 0.474994 | 5.028x |
| await-repository | 0.398284 | 0.211842 | 1.880x |
| await-upstream | 0.054693 | 0.066230 | 0.826x |

These are three alternating whole-process medians under concurrent gate load, not evidence of a performance change.
