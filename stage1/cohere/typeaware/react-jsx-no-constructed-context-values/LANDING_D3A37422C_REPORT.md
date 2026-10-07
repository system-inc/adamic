Rebased twelve implemented wave-20 analyses onto the typeof-null main landing and current lint area integration.
Validated source 4c9aa73b463f2774e8cd0fea5047c22f6f68979c; evidence commit follows on codex/typeaware-wave-20.
Checks: 1,132 controls/870 findings, complete fixes and suggestions, both corpora, sanitizers, released handles, checker, registry, Node and vet PASS.
Mutants: sixteen semantic, six facts, five registry retentions, one retained source handle on three questions, six listener declarations and eight typeof observations caught.
Not covered: shared typed registration and field-5 transport, own emitted-JavaScript comparison, invalid raw fragment options, all source shapes, or the full gate and its seventeen mandatory input checks.

Base origin/area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898
includes current main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06, b46914832's
legacy registry migration, 50a5f105 and finding harness 41eb6eab2. All 36 own
commits replayed cleanly from b46914832. Main changes native typeof null and
slot-presence classification, constructor/string classification and lowering;
new fixtures and their mutants were included in this landing's filtered Node
oracle. No owned rule source or shared file was edited. A second fetch confirmed
the bases unchanged during validation. Only codex/typeaware-wave-20 is pushed,
with an exact lease against old tip c4a1baa4a25144dc3b762d6d37aecb8f7d786cc2.
Main and area are never push targets.

Setup passed in 186s: Go 1.27.1 ready 1s, clang 20.1.8 ready 1s,
Node 24.19.0 ready 1s, submodules 1s, cache 186s. nproc 5 with a four-core
quota. Environment /workspace/adamic-tools/env.sh. Removed 72 completed own
ELF/archives totaling 2,730,842,816 bytes before fresh builds, preserving fixtures,
logs and manifests. Cloud runtime skill verified current readiness and enforced
restricted networking. No secret values were read or printed. The twelve fixed
test logs were inspected before archiving; only statuses, timings and public
fixture outputs, no fetch/push/setup logs or environment dumps.

Every private native runner and checker archive was rebuilt on the new runtime.
Controls/findings: first three 48/48, Nexus 121/160, core 682/502, JSX undefined
53/29, JSX fragments 57/34, constructed context 171/97. Core covers 662 valid
upstream cases and 20 extra cases; the documented independently panicking
invalid upstream TSX case remains excluded. All output bytes include findings,
automatic edits and suggestions. Both frozen corpora match normally and with
ASan/UBSan/LSan: 287 repository roots/18,485 output bytes and 77 TypeScript
compiler roots/5,241 bytes. Both have zero findings; controls provide positive
cases. The pinned TypeScript v6.0.3 source was supplied.

First-three gate PASS 309.130s; checker PASS .890s; registry PASS .157s;
filtered Node oracle PASS 35.409s; vet output empty. Worker caches were allowed,
not an uncached integration gate. The Node filter covers 18 fixtures: closures,
method closures, generic functions, regions, maps/text, regexp-cycle closures,
five proven-type witnesses and all seven new typeof witnesses. The one-byte
oracle mutant passed its failure proof. No selected test skipped.

Eight additional runtime mutant observations passed: one constructor, one static
string, five restored-null observations (literal/variable/call, comparison,
switch, slots and roll fixtures), and one missing-slot-presence mutation. Each
compiles, exits 0 with empty stderr, and differs from Node only in stdout.
These are unmodified upstream compiler tests, additional to this unit's rule
mutants. Their complete public witness outputs are in the fixed Node log archive.
Named syntax-kind declarations also match 204 bytes normally and sanitized.

Observed whole-process native / Go seconds, while gates overlap:

| Analysis suite | Repository | Compiler |
| --- | --- | --- |
| First three | 1.566233 / 0.648815 | 10.270824 / 3.650437 |
| Nexus | 0.419535 / 0.274274 | 2.384355 / 0.530871 |
| Core | 0.418185 / 0.317190 | 2.696448 / 0.619441 |
| JSX undefined | 0.731955 / 0.657637 | 3.461960 / 0.941937 |
| JSX fragments | 0.597457 / 0.904862 | 3.276388 / 1.092321 |
| Constructed context | 0.474833 / 0.318113 | 2.399972 / 0.581514 |

Native is slower on the compiler corpus in all six suites. These are
observations from overlapping runs, not isolated speed benchmarks; this unit
changes no performance behavior.

All sixteen rule semantic mutants compile, exit 0 with empty stderr and fail
only independent Go finding bytes: floating promises, implied eval, void operand,
process output, race timeout, blocking stream, promise rejection, regex suggestion
span, rest parameters, component name, fragment pragma object, construction
kind, missing memo list (case-161), render escape (case-140), helper escape
(case-136), and factory identity (case-146). Six fact mutants alter
accessed-property, callback-parameters, declaration-chain, resolved-callee,
module-records and binding-origin; byte comparison catches all. Five registry
retention mutants exit 0 instead of required released-handle panic 70.
Constructed MemoView released queries panic 70 on binding-origin, type-shape
and resolved-callee; one source mutant omits release and exits 0 on all three,
caught by the panic requirement. This is one mutant on three questions.
Three compiling listener mutants and three rule.json mutations substitute
Identifier for an expected kind; independent Go declarations catch all six.
Full logs, 12594 nonbinary output/fixture files and source hashes are in
validation/landing-d3a37422c beside this report.

Shared RuleContext/Settings are unchanged by this rebase. Context has a syntax
root and has(index, kind), but lacks a checker program/lease, checker config and
root manifest. Shared typed registration of the three JSX analyses remains
blocked inside the restriction to owned rule directories. All native analyses
and private runners are green; no fake shared implementation was added. These
are AST/checker rules, not parked HIR rules. The three earlier React compiler
claims remain parked for HIR/SSA/capture analysis. Named kinds and handed-node
entries are retained; no entry refetch or entry-kind string dispatch was added.

The private supported option flags construct upstream Go option structs for
promise rejection, redundant regex wrapping, JSX fragment mode and JSX globals.
No field-5 shared adapter is registered for these private analyses; shared
field-5 transport remains uncovered. The new options guard was not modified or
bypassed, and no options-dropping guard panic appeared in the selected gates.

The full repository gate was not run. Its seventeen mandatory external-input
checks are not claimed green and are not counted as skipped passes. Unrelated
parser packages were not selected; separate npm inputs remain absent:
ADAMIC_CSS_LIBRARY, ADAMIC_GRAPHQL_LIBRARY, ADAMIC_VALUES_LIBRARY,
ADAMIC_SELECTOR_LIBRARY, ADAMIC_MEDIA_QUERY_LIBRARY,
ADAMIC_CSS_PRINTER_LIBRARY and ADAMIC_JSON_PRETTIER. Whole TypeScript parser
and scanner parity packages were not run either. No test was relaxed, deleted,
or bypassed. This report claims only the exact filtered commands below are green.

Fresh audit: 665 origin refs, 33
distinct Markdown claim blobs, 197 inventory names, 172 claimed and 25 baseline
checker-dependent ports. No unclaimed rule remains. Exact boundaries admit
colon punctuation without confusing core and namespaced names. No new claim.

Rebase: `git rebase --onto origin/area/stage1-lint b46914832d70e00847d82d5d221ab7bb24040c53`.
Own push: `git push --force-with-lease=refs/heads/codex/typeaware-wave-20:c4a1baa4a25144dc3b762d6d37aecb8f7d786cc2 origin HEAD:refs/heads/codex/typeaware-wave-20`.
Every test output went to a log. Exact commands:

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /tmp/wave20-d3a-setup.log 2>&1
ADAMIC_WAVE20_ARTIFACTS=/workspace/wave20-validation/d3a-first ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave20-typescript go test ./stage1/cohere/typeaware -run '^TestWave20AgreementAndMutants$' -count=1 -v -timeout 30m > /tmp/wave20-d3a-first.log 2>&1
python3 stage1/cohere/typeaware/no-process-exit-after-output/verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/d3a-second --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/next/restored > /tmp/wave20-d3a-second.log 2>&1
python3 stage1/cohere/typeaware/prefer-promise-reject-errors/verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/d3a-third --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/third/valid-cases > /tmp/wave20-d3a-third.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-fragments/verify.py --artifacts /workspace/wave20-validation/d3a-fragments --compiler /workspace/wave20-validation/d3a-third/adamic --checker /workspace/wave20-validation/d3a-third/checker.a --sanitized-checker /workspace/wave20-validation/d3a-third/checker-asan.a > /tmp/wave20-d3a-fragments.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-no-undef/verify.py --artifacts /workspace/wave20-validation/d3a-undef --compiler /workspace/wave20-validation/d3a-third/adamic --checker /workspace/wave20-validation/d3a-third/checker.a --sanitized-checker /workspace/wave20-validation/d3a-third/checker-asan.a > /tmp/wave20-d3a-undef.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-no-constructed-context-values/verify.py --artifacts /workspace/wave20-validation/d3a-constructed --compiler /workspace/wave20-validation/d3a-third/adamic --checker /workspace/wave20-validation/d3a-third/checker.a --sanitized-checker /workspace/wave20-validation/d3a-third/checker-asan.a > /tmp/wave20-d3a-constructed.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-no-constructed-context-values/mutants.py --artifacts /workspace/wave20-validation/d3a-memo-mutants --controls /workspace/wave20-validation/d3a-constructed --compiler /workspace/wave20-validation/d3a-third/adamic --checker /workspace/wave20-validation/d3a-third/checker.a > /tmp/wave20-d3a-memo-mutants.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-fragments/listener_verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/d3a-listeners --compiler /workspace/wave20-validation/d3a-third/adamic --checker /workspace/wave20-validation/d3a-third/checker.a > /tmp/wave20-d3a-listeners.log 2>&1
go test ./bridge/tsgo/checker -count=1 > /tmp/wave20-d3a-checker.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestTypeOf(Constructor|StringLiteral|Null|NullSlotPresence)Mutant$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|maps_and_text|proven_assertions|proven_class_guards|proven_guards|proven_satisfies|proven_upcasts|typeof_dispatch|typeof_string_literal|typeof_null|typeof_null_compare|typeof_null_switch|typeof_null_slots|typeof_null_roll)\.a$' -count=1 -v -timeout 10m > /tmp/wave20-d3a-node.log 2>&1
go test ./stage1/cohere/lint/registry -count=1 > /tmp/wave20-d3a-registry.log 2>&1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > /tmp/wave20-d3a-vet.log 2>&1
```
