5249bfbc973800b377b0c39bd2335ad4e7fe7a0a is the tested compiler tip on codex/namespaces-tsc; the final evidence-only pushed tip is returned in delivery.
Built local const enums, proved namespace receivers, namespace class readiness, qualified callable merging and identical overload contracts.
All three exact front24 probes match Node in both backends; nine of ten shapes lower; the original twelve stay 6 Compiles, 3 NotYet, 3 Refused.
Fourteen compiler/IR mutants and eight independent source mutants caught; the census audit also catches range and attribution mutants.
Ten of the thirteen Debug NotYet sites remain; whole Debug/parser execution is not established and representation/postcondition boundaries are preserved.

Main was merged only into codex/namespaces-tsc, with merge commit 3d3fd32c,
including origin/main f4efdd2369311d1420aa53fdf5c1a55bdda811d4. The final fetch
and merge report Already up to date. No push or merge into main or area/ branches
was performed. Every group was pushed separately; no force push or PR.

## Pushed groups and counts

| Group | Original Debug sites | Commit | Movement |
| --- | ---: | --- | --- |
| Object/cache observation | 4 | 8bb1382c | 0; explicit container boundary kept |
| Unknown assertions | 2 | f18f5ea2 | 0; postcondition boundaries kept |
| Local const enums | 2 | 6a80eb29 | 2 original Debug sites cleared |
| Receiver scopes | 0 | cda5f0f1 | front24 receiver probe 0 to 1 |
| Namespace classes | 0 | 3e18dad6 | front24 accepted probes 1 to 2 |
| Callable namespaces | 1 | df66a0c7 | 1 original Debug site cleared; front24 accepted probes 2 to 3 |
| Generic function value | 1 | aa581553 | 0; polymorphic value boundary kept |
| Nullable generic | 1 | 7bc31a70 | 0; absence-tag boundary kept |
| Overload contracts | 1 | 1df9157b | real site retained; normalized shapes 6 to 9 |
| Never parameter | 1 | 5249bfbc | 0; bottom ABI boundary kept |

The original twelve-slice accepted count remains six for all these groups.
The added probes and normalized shapes measure narrower declaration behavior;
they are not substitutions for real-body fixtures. `groups.json` holds full
commit ids. `../progress-matrix.json` contains two actual refreshed twelve-row
build matrices at df66a0c7 and 5249bfbc, each matching Node for all successes.
Fixture 05 still stops on string truthiness after callable merging, 01 on boolean
operator semantics, and 10 on container escape. 02 and 03 retain main's cast
refusals; 06 retains its generic-method refusal. 09 already compiled under main's
open numeric-enum policy before this unit; its source/refusal policy was not
changed here. The earlier exact 06/09 judgments remain in ../DECISIONS.md.

## Exact new parser probes

`front24-results.json` records byte-identical sources from parser-proof 5d777de3
in the three native-*.a files, compiled by the final compiler with sanitizers
and leak detection, plus emitted JavaScript. All exit 0 with empty stderr:

| Probe | Node/native/JavaScript stdout |
| --- | --- |
| native-namespace-object-receiver.a | receiver declaration loaded |
| native-namespace-class.a | class declaration loaded |
| native-callable-namespace.a | callable namespace loaded |

These exact probes declare behavior rather than exercise all of it. Expanded
fixtures exercise live receiver state (3 then 7), detached receiver differences
(3:0 on Node, named production refusal), ordered static/field/constructor effects,
instance identity and instanceof, class TDZ, pending-namespace construction,
callable properties, identity and receivers, and identity reads before Debug
exists. The production receiver proof visits every runtime module; an imported
escape cannot inherit a single-file receiver proof. Object methods bind their
own receiver; arrows inherit the enclosing one. Receiver identity, writes,
computed keys, optional reads and unknown value edges remain named boundaries.

Callable merging is a qualified fixed-property subset, with canonical symbol
identity comparisons and readiness checks in source order. It does not create
an escaping/reflected callable object. Such observations and namespace/class
merging stay NotYet. Live declared receiver properties must match their export
storage type exactly. Namespace classes use existing symbol registration and
class machinery; even classes without statics have ordered readiness storage.

## Final unchanged-source census

The audited, measurement-only binary ran over the same 79 adapted compiler files
from TypeScript 6.0.3 commit 050880ce59e30b356b686bd3144efe24f875ebc8.
SHA-256 for every source matches the previous unit before and after measurement.
It took 186.050 seconds. `debug-meter.json`, evidence/source-hashes.json and
complete evidence/latent.jsonl.gz retain the observations. Recursive extraction
includes dependency findings and deduplicates kind, location, reason and text.

The original 55 reading Debug sites remain zero. The thirteen other Debug
NotYet sites reduce to ten: two local const enum declarations (879:9, 893:9)
and the function/namespace merge (137:5) clear. Remaining groups are:

| Sites | Cause | Judgment and narrow sound acceptance |
| --- | --- | --- |
| 169:49, 170:22, 189:56, 190:14 | Computed cache access/replacement observes the namespace object | Current refusal is necessary without a real canonical live container and mutable callable exports; a snapshot is wrong. |
| 213:28 | assert's unknown parameter | Tagged unknown truthiness and a proved postcondition are required; the boolean-only proof is narrower than JavaScript truthiness. |
| 365:29 | empty generic type assertion's unknown parameter | Production Refused is sound and necessary: an empty body cannot prove arbitrary T. A concrete verified narrowing or explicitly non-narrowing source adaptation is required. |
| 251:45 | Generic self function value | Needs a polymorphic callable ABI, or proved opaque identity-only metadata use. One concrete instantiation cannot silently stand in for every instantiation. |
| 255:37 | T plus null and undefined | Needs distinct absence tags and concrete T storage; collapsing null into undefined is unsound. |
| 281:5 | Differing overloaded assertion contracts | Identical concrete contracts now lower; actual predicates and varying parameter/result contracts require proof/checks, never indiscriminate signature removal. |
| 273:33 | never parameter | Safe NotYet, but too strict for a helper whose every use is proved dead. Narrow acceptance: omit the proved unreachable helper or specify a bottom-argument ABI that evaluates divergence before any call. |

Exact programs, production diagnostics, Node outputs and semantic mutants are
in each group's .a/.json/.results.json and linked reports:
OBJECT_OBSERVATION.md, UNKNOWN_ASSERTIONS.md, GENERIC_FUNCTION_VALUE.md,
NULLABLE_GENERIC.md, OVERLOADS.md and NEVER_PARAMETER.md. Source-only mutants
for retained boundaries do not claim native compiler support. Normalized
shapes now admit nine of ten; tracingEnabled still escapes. This is independent
of actual assertion bodies, arbitrary runtime objects and full parser behavior.

The production build of unchanged debug.ts was retried. It exits 1 in checking,
first at builder.ts:1246:69 (Path | undefined passed where string is required),
followed by other dependency diagnostics. Its complete log is preserved in
evidence/whole-debug-build.log.gz. No C/native whole-Debug or parser success is
claimed. Latent totals are 1,413 NotYet, 4,781 Refused, 3 panic findings and
3 skipped dependencies. They are measurement results on a checker-rejected
program, not runnable production output; unrelated main changes affect totals.

## Executed mutants

| Compiler or IR mutant | What caught it |
| --- | --- |
| Remove namespace observation guard | TestDebugNamespaceObservationBoundary loses the explicit container reason |
| Disable boolean assertion-parameter guard | Unknown-condition assertion boundary pin |
| Accept unproved empty assertion normal return | Empty generic assertion boundary pin |
| Change enum glyph to x | Node stdout differs in both backends |
| Remove deferred local-enum guard | TestDebugLocalConstEnums loses NotYet |
| Replace live namespace receiver reads with 99 | Node stdout differs in both backends |
| Restore the complete old recursive this refusal | TestNamespaceNestedReceiver fails on the exact front24 probe |
| Inspect only the entry module for receiver escapes | TestNamespaceReceiverImportedEscape loses its specific proof boundary |
| Remove no-static namespace class readiness | Node exits 70; both backends wrongly print unreachable and exit 0 |
| Remove namespace readiness before construction | Both backends give ReferenceError where Node gives TypeError |
| Invert canonical function identity | Node stdout differs in both backends |
| Erase identity-comparison readiness | Node exits 70; both backends wrongly print true and exit 0 |
| Skip differing overload contracts | TestDebugOverloadContracts loses NotYet and admits the invalid overload |
| Return 99 from an accepted overload implementation | Node stdout differs in native with clean sanitizers and JavaScript |

Eight source mutants are independently caught by Node: snapshot computed
namespace cache bindings; add an actual number test to the empty generic
assertion; change an enum glyph; bind a detached call to Debug; change generic
self identity; drop the null branch; correct the deliberately unsound overload
implementation; and force an admitted never-fixture branch to throw.
Each .results.json records the original and mutated stdout/exit. Compiler
mutations were restored. Guard mutants are diagnostic evidence, identified as
such, rather than alleged native support for rejected shapes.

A preliminary receiver-scope mutant changed only the predicate and survived:
it did not restore the old refusal branch. The corrected mutant restored the
whole prior branch and was caught. Broad count checks also exposed a new
signature guard calling Text on a destructuring pattern; checking identifier
kind fixed it, and all checks were rerun. Checker-invalid TS2450/TS2630 test cuts
were removed rather than bypassing checking. No failed build was counted as a
semantic mutant catch.

## Final verification and setup

| Command | Output |
| --- | --- |
| bash cloud/setup.sh | success; Go build 34.205s, cache warm 34.526s, total 34.572s |
| nproc | 5; cgroup cpu.max 400000 100000 |
| python3 stage3/namespaces/progress-matrix.py --label debug-groups-final --scratch /tmp/debug-groups-final-matrix | 6 Compiles, 3 NotYet, 3 Refused, 0 Checker |
| go test ./internal/lower -count=1 | pass, 22.033s |
| go test ./internal/oracle -run 'TestNativeAgreesWithNode\|TestNamespace' -count=1 -timeout 20m | pass, 146.121s |
| go test ./internal/oracle -run TestCountsAreRecorded -count=1 | pass, 33.581s; counts were refreshed after each new fixture group |
| go vet ./internal/lower ./internal/oracle | pass, empty output |
| python3 stage3/census/latent/audit.py /tmp/debug-groups-final-census | continuation, no output, production loader disabled, range/body and attribution checks pass; both audit mutants caught |
| final census with LATENT_ASSERT_NO_OUTPUT=1 | success, 186.050s, unchanged source hashes |

Toolchain environment: source /workspace/adamic-tools/env.sh; Go 1.27.1,
Node 24.19.0, clang 20.1.8. Setup timing lines and all test/mutant logs are under
evidence/. Test output was logged directly. No protected lower.go, native emit
files or oracle_test.go were edited; no cohere files were copied. Final evidence
commit changes only documentation and recorded observations after re-greening.
