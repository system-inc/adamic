Rebased all twelve implemented wave-20 analyses onto current main and the integrated lint area branch.
Validated source 3dcf3c918db0695bd00551d7a130024245877b16; landing evidence commit follows on codex/typeaware-wave-20.
Checks: 1,132 controls/870 findings, complete fixes/suggestions, both corpora, sanitizers, released handles, checker, Node, registry and vet PASS.
Mutants: sixteen semantic, six checker facts, five registry retentions, one retained source handle on three questions, three compiled listeners and three JSON declarations caught.
Not covered: shared typed registration, own emitted-JavaScript comparison, invalid raw fragment options, all source shapes or the full gate and its seventeen newly mandatory input checks.

The required rebase is clean. Base origin/area/stage1-lint is
b84a9d9314b65d3d0261ee017e233287b4f071da and includes current main
c7991b900362796aefd111474e65eb5398e91953, the requested 50a5f105 integration
and finding harness 41eb6eab2. All 32 own commits replayed from the prior area
base d65a8f931 without conflicts. Constructed-context implementation is now
be4c8457f. A second fetch immediately before collecting evidence confirmed both
bases unchanged. Only codex/typeaware-wave-20 is pushed, with an exact lease
against old own tip aeb074ff234a77a1f2eb0c54e7c5522e855290ed. Main and area
are never push targets. No compiler, shared harness or registration source was
changed during this landing. Integrated Linux leak helpers were retained.

Setup passed: Go 1.27.1 ready 0s, clang 20.1.8 ready 0s, Node 24.19.0 ready
0s, submodules 0s, cache 176s, total 176s. nproc is 5, with a four-core quota.
Environment: /workspace/adamic-tools/env.sh. Completed own scratch executables
and archives totaling 7,255,877,855 bytes were removed before fresh builds;
source fixtures, logs and manifests were retained. Cloud environment network
policy and runtime readiness were checked using the runtime skill; no secrets
were printed.

All twelve private native analyses were rebuilt against the rebased compiler.
Suite counts are first three 48 controls/48 findings, Nexus 121/160, core
682/502, JSX undefined 53/29, JSX fragments 57/34 and constructed context
171/97. Core controls include 662 valid upstream cases and 20 extra cases;
the previously documented independently panicking invalid upstream TSX case
remains excluded. Wire comparisons include findings, automatic edits and all
suggestions. Both frozen corpora pass normally and with ASan/UBSan/LSan:
287 repository roots/18,485 complete output bytes and 77 TypeScript compiler
roots/5,241 bytes. Both have zero findings; controls supply the positive cases.
The pinned TypeScript v6.0.3 source was supplied, never silently omitted.

First-three gate PASS 303.352s; checker PASS 1.053s; registry PASS .105s;
filtered external Node oracle PASS 36.725s; vet output empty. Node comparisons
include closures, method closures, generic functions, regions, maps/text,
regexp-cycle closures and all five newly integrated proven-type fixtures:
assertions, class guards, guards, satisfies and upcasts. The one-byte oracle
mutant also passes its required failure proof. The Go subtest filter admits the
regexp-cycle fixture in addition to the requested names; its result is included.
Named listener comparisons match 204 bytes normally and sanitized.

Observed whole-process native / Go seconds, with overlapping gates:

| Analysis suite | Repository | Compiler |
| --- | --- | --- |
| First three | 1.058890 / .446553 | 9.888622 / 1.992721 |
| Nexus | 0.566097 / 0.270491 | 2.644905 / 0.627569 |
| Core | 0.366446 / 0.268576 | 2.633637 / 0.617149 |
| JSX undefined | 0.638817 / 0.452971 | 3.596612 / 0.846827 |
| JSX fragments | 0.745813 / 0.489590 | 3.647568 / 1.147739 |
| Constructed context | 0.518971 / 0.316027 | 2.690430 / 0.830626 |

These measurements do not isolate speed improvements. Native remains slower;
this landing changes no algorithm or compiler performance behavior.

Every semantic mutant compiles, exits 0 with empty stderr and is caught only by
independent Go finding bytes: floating promises, implied eval, void operand,
process output, race timeout, blocking stream, promise rejection, regex
suggestion span, rest parameters, component name, fragment pragma object,
construction kind, missing memo list (case-161), render escape (case-140),
helper escape (case-136) and factory identity (case-146). Six fact mutants
alter accessed-property, callback-parameters, declaration-chain, resolved-callee,
module-records and binding-origin; the same comparison catches all six.
Five registry-retention mutants exit 0 instead of the required released-handle
panic 70. Constructed MemoView released probes panic 70 on binding-origin,
type-shape and resolved-callee. One source mutant omits release and exits 0 for
all three questions; the panic requirement catches it. This is one mutant on
three questions. Three compiled listener mutants and three rule.json mutants
replace an expected kind with Identifier; Go declarations catch each.
Exact outputs, source hashes and 12594 nonbinary artifact files are archived
under validation/landing-c7991b900 beside this report.

The shared RuleContext and Settings still lack a checker program/lease,
checker configuration and root manifest. Registering the three JSX analyses
cannot be completed within owned rule directories against that API. The full
source analyses and private typed runners are green; no zero-finding shared
placeholder was added. These AST/checker analyses are not parked HIR rules.
The three older React compiler claims remain parked for HIR/SSA/capture analysis.
Named ast.Kind metadata and handed-node entry points are retained, with no
entry refetch or entry-kind string dispatch inside each JSX analysis.

The full repository gate was not run. Its seventeen newly mandatory correctness
checks are not claimed green or counted as skipped passes. The separate external
npm parser inputs are absent: ADAMIC_CSS_LIBRARY, ADAMIC_GRAPHQL_LIBRARY,
ADAMIC_VALUES_LIBRARY, ADAMIC_SELECTOR_LIBRARY, ADAMIC_MEDIA_QUERY_LIBRARY,
ADAMIC_CSS_PRINTER_LIBRARY and ADAMIC_JSON_PRETTIER. Those unrelated packages
were not selected in this unit's filtered validation. No check was relaxed,
deleted or bypassed. The selected checks have no skipped cases. The parser and
scanner whole compiler-parity packages were not run by this unit either;
TypeScript source is present for the corpus comparisons actually run above.
The landing claim is green for this branch's twelve rule oracles and the exact
commands below, not for the entire repository gate.

Fresh all-origin selection inspected 637 refs and
33 distinct Markdown claim blobs: 197 inventory rules,
172 claimed names and 25 baseline checker-dependent ports. No unclaimed rule
remains. Exact boundaries admit colon punctuation without confusing core and
namespaced names. No additional claim was made. Full selection evidence is
compressed beside the test logs.

The rebase command was `git rebase --onto origin/area/stage1-lint d65a8f931c98655936ae04c6899f38f14862b73e`; its log confirms success. The own-branch push uses `git push --force-with-lease=refs/heads/codex/typeaware-wave-20:aeb074ff234a77a1f2eb0c54e7c5522e855290ed origin HEAD:refs/heads/codex/typeaware-wave-20`.

Commands, output always written to files:

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /tmp/wave20-c799-setup.log 2>&1
ADAMIC_WAVE20_ARTIFACTS=/workspace/wave20-validation/c799-first ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave20-typescript go test ./stage1/cohere/typeaware -run '^TestWave20AgreementAndMutants$' -count=1 -v -timeout 30m > /tmp/wave20-c799-first.log 2>&1
python3 stage1/cohere/typeaware/no-process-exit-after-output/verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/c799-second --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/next/restored > /tmp/wave20-c799-second.log 2>&1
python3 stage1/cohere/typeaware/prefer-promise-reject-errors/verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/c799-third --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/third/valid-cases > /tmp/wave20-c799-third.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-fragments/verify.py --artifacts /workspace/wave20-validation/c799-fragments --compiler /workspace/wave20-validation/c799-third/adamic --checker /workspace/wave20-validation/c799-third/checker.a --sanitized-checker /workspace/wave20-validation/c799-third/checker-asan.a > /tmp/wave20-c799-fragments.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-no-undef/verify.py --artifacts /workspace/wave20-validation/c799-undef --compiler /workspace/wave20-validation/c799-third/adamic --checker /workspace/wave20-validation/c799-third/checker.a --sanitized-checker /workspace/wave20-validation/c799-third/checker-asan.a > /tmp/wave20-c799-undef.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-no-constructed-context-values/verify.py --artifacts /workspace/wave20-validation/c799-constructed --compiler /workspace/wave20-validation/c799-third/adamic --checker /workspace/wave20-validation/c799-third/checker.a --sanitized-checker /workspace/wave20-validation/c799-third/checker-asan.a > /tmp/wave20-c799-constructed.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-no-constructed-context-values/mutants.py --artifacts /workspace/wave20-validation/c799-memo-mutants --controls /workspace/wave20-validation/c799-constructed --compiler /workspace/wave20-validation/c799-third/adamic --checker /workspace/wave20-validation/c799-third/checker.a > /tmp/wave20-c799-memo-mutants.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-fragments/listener_verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/c799-listeners --compiler /workspace/wave20-validation/c799-third/adamic --checker /workspace/wave20-validation/c799-third/checker.a > /tmp/wave20-c799-listeners.log 2>&1
go test ./bridge/tsgo/checker -count=1 > /tmp/wave20-c799-checker.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|maps_and_text|proven_assertions|proven_class_guards|proven_guards|proven_satisfies|proven_upcasts)\.a$' -count=1 -v -timeout 10m > /tmp/wave20-c799-node.log 2>&1
go test ./stage1/cohere/lint/registry -count=1 > /tmp/wave20-c799-registry.log 2>&1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > /tmp/wave20-c799-vet.log 2>&1
```
