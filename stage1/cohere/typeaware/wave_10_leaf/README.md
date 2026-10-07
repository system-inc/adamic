Added live handed-node native adapters for symbol-description and structure/react-hook-no-any-type, plus require-await reporting.
The branch is rebased and oracle-green on origin/main b8fb957a; no new rules are claimed.
The two live rules pass full Go-byte controls/corpora, sanitizers, released-handle checks and native mutants; source lint reports 0.
Require-await body filters and 32 reporting controls pass their narrower Go authorities; the incomplete full listener still refuses candidates.
Full require-await contextual/generic/heritage contract decisions and shared-registry integration remain unimplemented.

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
