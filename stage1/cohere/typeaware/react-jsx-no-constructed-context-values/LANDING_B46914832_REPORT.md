Rebased twelve implemented wave-20 analyses onto the current lint area integration, with current main included.
Validated source 0a8eed884329ef0965de8f52699df17f2c140a86; evidence commit follows on codex/typeaware-wave-20.
Checks: 1,132 controls/870 findings, complete fixes and suggestions, both corpora, sanitizers, released handles, checker, registry, Node and vet PASS.
Mutants: sixteen semantic, six facts, five registry retentions, one retained source handle on three questions, three compiled listeners and three JSON declarations caught.
Not covered: shared typed registration, own emitted-JavaScript comparison, invalid raw fragment options, all possible source shapes, or the full gate and its seventeen mandatory input checks.

Base origin/area/stage1-lint b46914832d70e00847d82d5d221ab7bb24040c53
includes current main c7991b900362796aefd111474e65eb5398e91953, the requested
50a5f105 integration and finding harness 41eb6eab2. All 35 own commits replayed
cleanly from b84a9d931. The area update migrates legacy lint rules onto the
registry and adds shared context helpers; it changes no compiler, bridge or
private type-aware source. No shared files were edited during this landing.
A second fetch confirmed the bases unchanged before collecting evidence.
Only codex/typeaware-wave-20 is pushed, using an exact lease against old tip
1eab77c2732190a2eb39539de78c007aa0587bb6. Main and area are never push targets.

Setup passed in 64s: Go 1.27.1 ready 0s; clang 20.1.8 ready 1s;
Node 24.19.0 ready 1s; submodules 1s; cache 64s. nproc 5, four-core quota.
Environment /workspace/adamic-tools/env.sh. The cloud runtime skill verified
current ready observations and enforced restricted networking. No secrets were
printed. The twelve fixed test logs were inspected before archiving: only gate
statuses, public fixture paths, timings and mutant outcomes, no credentials,
environment dumps, fetch or push logs.

All private native runners and checker archives were rebuilt against this base.
Controls/findings: first three 48/48, Nexus 121/160, core 682/502, JSX undefined
53/29, JSX fragments 57/34 and constructed context 171/97. Core covers 662 valid
upstream cases and 20 extra cases; the previously documented independently
panicking invalid upstream TSX case remains excluded. Comparisons hold complete
findings, fixes and suggestions. Both frozen corpora match normally and under
ASan/UBSan/LSan: 287 repository roots/18,485 output bytes and 77 TypeScript
compiler roots/5,241 bytes. Both have zero findings; controls provide the
positive cases. The pinned TypeScript v6.0.3 source was provided to the gate.

First-three gate PASS 306.511s; checker PASS .855s; registry PASS .203s;
filtered Node oracle PASS 1.005s; vet output empty. Ordinary worker Node caches
were used, as allowed by CLAUDE.md; this is not an uncached integration gate.
The filter includes closures, method closures, generic functions, regions,
maps/text, regexp-cycle closures and all five proven-type fixtures (assertions,
class guards, guards, satisfies and upcasts). The one-byte oracle mutant passes
its required failure proof. Named kind declarations match 204 bytes normally
and sanitized. No selected test skipped.

Observed whole-process native / Go seconds, while gates overlap:

| Analysis suite | Repository | Compiler |
| --- | --- | --- |
| First three | 1.275885 / 0.553346 | 7.199830 / 1.925110 |
| Nexus | 0.469722 / 0.268967 | 3.143718 / 0.620156 |
| Core | 0.418188 / 0.331214 | 2.636361 / 0.632155 |
| JSX undefined | 0.701168 / 0.617849 | 4.395540 / 1.495202 |
| JSX fragments | 1.275495 / 1.030854 | 3.424101 / 0.837672 |
| Constructed context | 0.472188 / 0.321852 | 2.538738 / 0.717322 |

Native remains slower. These are observations, not an isolated speed benchmark;
this landing makes no performance change.

All sixteen semantic mutants compile, exit 0 with empty stderr, and fail only
independent Go byte comparisons: floating promises, implied eval, void operand,
process output, race timeout, blocking stream, promise rejection, regex suggestion
span, rest parameters, component name, fragment pragma object, construction
kind, missing memo list (case-161), render escape (case-140), helper escape
(case-136), and factory identity (case-146). Six fact mutants alter
accessed-property, callback-parameters, declaration-chain, resolved-callee,
module-records and binding-origin; comparison catches all. Five registry
retention mutants exit 0 instead of required released-handle panic 70.
Constructed MemoView released queries panic 70 on binding-origin, type-shape
and resolved-callee; one source mutant omits release and exits 0 for all three,
caught by the panic requirement. This is one mutant on three questions.
Three compiling listener mutants and three rule.json mutants replace an
expected kind with Identifier and are caught by independent Go declarations.
Full logs, 12594 nonbinary output/fixture files and source hashes are
archived in validation/landing-b46914832 beside this report.

The shared RuleContext and Settings were read whole after rebase. Context now
has a syntax root and extra syntax helpers, but still lacks a checker program/
lease, checker configuration and root manifest. Shared typed registration of
the three JSX analyses remains blocked by that API, within the restriction to
owned rule directories. Native source analyses and private typed runners are
green; no fake shared implementation was added. These are AST/checker rules,
not parked HIR rules. The three older React compiler claims remain parked for
HIR/SSA/capture analysis. Named ast.Kind declarations and handed-node entries
are retained; no entry refetch or entry-kind string dispatch was introduced.

The full repository gate was not run. Its seventeen mandatory external-input
checks are not claimed green and are not counted as skipped passes. Unrelated
parser packages were not selected; their separate npm inputs remain absent:
ADAMIC_CSS_LIBRARY, ADAMIC_GRAPHQL_LIBRARY, ADAMIC_VALUES_LIBRARY,
ADAMIC_SELECTOR_LIBRARY, ADAMIC_MEDIA_QUERY_LIBRARY,
ADAMIC_CSS_PRINTER_LIBRARY and ADAMIC_JSON_PRETTIER. Whole TypeScript parser
and scanner parity packages were not run either. No check was relaxed, deleted,
or bypassed. This report claims the exact filtered commands below are green.

Fresh selection inspected 650 origin refs and
33 distinct Markdown claim blobs. All 197 inventory names
are covered by 172 claimed names and 25 baseline checker-dependent ports.
No unclaimed rule remains. Exact name boundaries admit colon punctuation while
keeping core and namespaced rules distinct. No new claim was made.

Rebase command: `git rebase --onto origin/area/stage1-lint b84a9d9314b65d3d0261ee017e233287b4f071da`.
Own push command: `git push --force-with-lease=refs/heads/codex/typeaware-wave-20:1eab77c2732190a2eb39539de78c007aa0587bb6 origin HEAD:refs/heads/codex/typeaware-wave-20`.
All test output went to files. Exact commands:

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /tmp/wave20-b469-setup.log 2>&1
ADAMIC_WAVE20_ARTIFACTS=/workspace/wave20-validation/b469-first ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave20-typescript go test ./stage1/cohere/typeaware -run '^TestWave20AgreementAndMutants$' -count=1 -v -timeout 30m > /tmp/wave20-b469-first.log 2>&1
python3 stage1/cohere/typeaware/no-process-exit-after-output/verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/b469-second --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/next/restored > /tmp/wave20-b469-second.log 2>&1
python3 stage1/cohere/typeaware/prefer-promise-reject-errors/verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/b469-third --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/third/valid-cases > /tmp/wave20-b469-third.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-fragments/verify.py --artifacts /workspace/wave20-validation/b469-fragments --compiler /workspace/wave20-validation/b469-third/adamic --checker /workspace/wave20-validation/b469-third/checker.a --sanitized-checker /workspace/wave20-validation/b469-third/checker-asan.a > /tmp/wave20-b469-fragments.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-no-undef/verify.py --artifacts /workspace/wave20-validation/b469-undef --compiler /workspace/wave20-validation/b469-third/adamic --checker /workspace/wave20-validation/b469-third/checker.a --sanitized-checker /workspace/wave20-validation/b469-third/checker-asan.a > /tmp/wave20-b469-undef.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-no-constructed-context-values/verify.py --artifacts /workspace/wave20-validation/b469-constructed --compiler /workspace/wave20-validation/b469-third/adamic --checker /workspace/wave20-validation/b469-third/checker.a --sanitized-checker /workspace/wave20-validation/b469-third/checker-asan.a > /tmp/wave20-b469-constructed.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-no-constructed-context-values/mutants.py --artifacts /workspace/wave20-validation/b469-memo-mutants --controls /workspace/wave20-validation/b469-constructed --compiler /workspace/wave20-validation/b469-third/adamic --checker /workspace/wave20-validation/b469-third/checker.a > /tmp/wave20-b469-memo-mutants.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-fragments/listener_verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/b469-listeners --compiler /workspace/wave20-validation/b469-third/adamic --checker /workspace/wave20-validation/b469-third/checker.a > /tmp/wave20-b469-listeners.log 2>&1
go test ./bridge/tsgo/checker -count=1 > /tmp/wave20-b469-checker.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|maps_and_text|proven_assertions|proven_class_guards|proven_guards|proven_satisfies|proven_upcasts)\.a$' -count=1 -v -timeout 10m > /tmp/wave20-b469-node.log 2>&1
go test ./stage1/cohere/lint/registry -count=1 > /tmp/wave20-b469-registry.log 2>&1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > /tmp/wave20-b469-vet.log 2>&1
```
