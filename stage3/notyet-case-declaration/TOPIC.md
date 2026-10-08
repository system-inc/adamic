Topic-only continuation from main 6998ebc24ae353193cb1495d3d51308131a4b5c7, on codex/notyet-case-declaration-topic.
Carried only seven own non-merge commits; new morning witnesses require no additional production lowering changes.
Morning CSV e8c283b5 has 69 matching root signatures: 62 declarations, 6 optional-type cases, 1 undefined case.
Twelve switch fixtures pass Node source, JavaScript, release and sanitized native; all 15 semantic mutants are caught.
No selected switch kind needs checked views; no runtime C edits; whole compiler units and other stops remain unproven.

The morning reasons were inspected largest first. They are the same three kinds
already implemented. Declaration bindings remain scoped to the whole switch,
with readiness captured alongside each binding. Optional numeric cases retain
their absence tag. Literal undefined is encoded in the scrutinee representation;
dynamic expressions and other representation mismatches retain their old stops.

No other worker's commits were cherry-picked and no merge was made. The branch
started at the newest origin/main resolved when this continuation began. Its
seven carried commits map as follows:

| Original | Topic-only | Change |
| --- | --- | --- |
| a66178c6 | 10f82dbe | Binder reductions |
| e101af4f | 13008980 | Switch-wide scope and readiness |
| b8cbb9e5 | e638eda1 | Optional enum reductions |
| b7e2a9a6 | c5fe5063 | Optional case representation |
| a49f0526 | fac3bad3 | Undefined reduction |
| ccf20ba4 | eeedb686 | Undefined constant |
| 94c13783 | 642585c9 | Historical verification report |

In particular, the old area/compiler merge, the census replay commit, and the
old main merge are not carried. REPORT.md is the historical report, with its
original pins and results; this file records the fresh branch's verification.

The new declaration fixture reduces newly present checker.ts:6328:25 (grouped
labels with a declaration), checker.ts:39939:29 (a declaration in default after
conditional fallthrough), and typeSerializer.ts:499:17 (an early return before
two declarations). The optional fixture covers the morning six sites' actual
scrutinee sources: optional node.kind, an optional enum field, an array index,
and a call returning an optional enum. It includes zero, another present value,
and absence for each source. The source-only test separately records exact Node
output. They register from switch_case_morning_test.go, not oracle_test.go.

The morning raw CSV is pinned in morning-coverage.json, with each matching site
listed. The 69-site number counts matching root-reason signatures, not compiler
units proved to compile. No fresh whole-corpus measurement or retired-root
claim is made. Earlier designed cast refusals and unrelated initializer/helper
stops remain. None of these three switch kinds requires checked views, so no
kind was skipped on that ground.

Commands ran with /workspace/adamic-tools/env.sh sourced. Output went to logs.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestSwitchCase(SourceReductions|OptionalSources|UndefinedSource)|TestNativeAgreesWithNode/internal/oracle/testdata/switch_case_' -count=1 -v -timeout 10m
python3 internal/oracle/testdata/run-switch-case-mutants.py
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestSwitchCaseMorningSources|TestNativeAgreesWithNode/internal/oracle/testdata/switch_case_morning_' -count=1 -v -timeout 10m
python3 internal/oracle/testdata/run-switch-case-mutants.py morning-initializer-twice morning-optional-tag
go test ./internal/ir ./internal/lower -run 'Closure|Capture|Switch|Ready' -count=1 -timeout 10m
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestSwitchCase(SourceReductions|OptionalSources|UndefinedSource|MorningSources)|TestNativeAgreesWithNode/internal/oracle/testdata/(switch_case_|enums\.a|parameter_properties\.a)' -count=1 -v -timeout 10m
node --check oracle/node.mjs
```

The initial ten-fixture run passed in 12.424s; the new two-fixture run passed in
0.569s. Related IR tests passed in 0.015s and lowering tests in 2.506s.

| Mutant | Fixture | Catcher |
| --- | --- | --- |
| native-read | switch_case_read_tdz | Node exit comparison |
| native-write | switch_case_write_tdz | Node exit comparison |
| javascript-read | switch_case_read_tdz | Node exit comparison |
| javascript-write | switch_case_write_tdz | Node exit comparison |
| early-ready | switch_case_capture_tdz | Node exit comparison |
| entry-ready | switch_case_reentry_tdz | Node exit comparison |
| never-ready | switch_case_scope | Node exit comparison |
| initializer-twice | switch_case_scope | Node stdout comparison |
| oracle-transform | switch_case_scope | Node source/backend comparison |
| optional-tag | switch_case_optional_mapping | Node stdout comparison |
| optional-presence | switch_case_optional_mapping | Node stdout comparison |
| undefined-string | switch_case_undefined | Node stdout comparison |
| undefined-pair | switch_case_undefined | Node stdout comparison |
| morning-initializer-twice | switch_case_morning_declarations | Node stdout comparison |
| morning-optional-tag | switch_case_morning_optional | Node stdout comparison |

All mutants were run independently, required a semantic oracle comparison
failure, and restored in finally blocks. Build/refusal failures do not count.
No full package run or full repository gate was run in this continuation.
The oracle tests exercise the native and JavaScript packages; the additional
regressions cover the changed IR and lowering helpers.

Logs: /tmp/notyet-case-topic-{fixtures,mutants,counts,morning-fixtures,
morning-mutants,regressions,final-counts,final-fixtures,node-syntax}.log and
/tmp/adamic-switch-case-mutants/*.log. The first counts refresh passed in
19.898s and corrected the carried optional_mapping fixture's allocation count
for main's implementation (9 allocations/frees rather than 8). No other old
fixture count changed.

Final counts refresh passed in 20.483s and adds the two morning rows. The final
uncached fixture run, including existing enum/parameter-property transform
fallbacks, passed in 2.363s; native observations had 35 misses and Node had 33
misses. Node syntax checking passed. git diff --check passed. The branch has
zero merge commits above main; all changes in that range are this unit's own
commits. The branch is pushed once after these checks, then work stops.
