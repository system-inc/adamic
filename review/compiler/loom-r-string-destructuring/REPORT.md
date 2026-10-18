Rebased string object-destructuring lowering onto exact main dbdf977f5b, preserving the source commit's own change.
Source: 26077f9635a133aa87a2d7a51d7b1c765f907d55; base: dbdf977f5b8b0979d414ebe6f20e7aadd312e309; delivery: one commit on compiler/loom-r-string-destructuring.
Validation: focused lowering, Node/backend/sanitizer fixture, flow corpus, IR readers, guards, counts regeneration, reader closure and vet; exact commands and results are in this directory's JSON and raw .log.txt files.
Mutants: wrong member and skipped empty-pattern source caught by output comparisons; missing fix caught by refusal assertion; allocation count 35 caught against measured 34 twice.
Not covered: full packages/gate, ahra/Loom, full native scanner, WASI, a new TypeScript census/admission sweep, or unavailable optional library/cycle-ledger inputs.

Applied with git cherry-pick --no-commit of the source commit, after a clean git merge-tree --write-tree. There were no conflicted hunks and no conflict resolutions. No main..source diff was used. own-change-equivalence.json compares every added/removed line in all five non-counts code/fixture files with the source commit's own diff: all identical. The source commit's review evidence is preserved unchanged, with new execution evidence here.

The source is saved once, with its actual IR representation. Plain object bindings of the static string length property use the existing UTF-16 StringLength read, including aliases and string parameters. Empty patterns still evaluate their source. Other members, computed names, defaults, rest and nested names remain explicit refusals. Conservative assumption: this rebase preserves the source change's static-length-only scope; it does not broaden string-member support.

The inherited pinned-050880ce census found 324 patterns, exactly one object-binding string source, and only the member length. The original requested scanner reduction was scanner.ts:778; the pinned tree locates it at scanner.ts:1332. These are archived observations, not a census rerun on this base. The original admission sweep is likewise archived, not claimed as new evidence.

The positive fixture is internal/oracle/testdata/string_destructuring.a. It covers scanIdentifierParts(), source call counts, duplicate aliases, an empty binding, empty/ASCII/BMP/astral/lone-surrogate strings, and a destructured string parameter. The refusal test TestStringDestructuringMemberDiagnostic checks charCodeAt, main.a:1:9 and a concrete fix. No refused top-level oracle fixture is added; the refusal source is inline in the lower test. The archived unsupported-member.a is under review, outside the oracle/flow fixture census.

Counts were not hand-merged: the auto-applied table was replaced by the exact base table, then regenerated once, in its entirety, on Linux with house clang 20.1.8:

    GOMAXPROCS=4 timeout 240 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 180s -parallel 4 -v -args -update-counts

All 1,059 measured leaves passed; package 123.762s. Exactly one row was added, and no existing row changed or disappeared:

    | internal/oracle/testdata/string_destructuring.a | 34 | 34 | 18 | 46 | 6 | 0 |

Both the source commit's own runtime diff and this change's runtime diff against the exact base are empty. Runtime changed lines: none. No runtime feature set, runtime C line, IR map, or new compiler-consuming package was introduced. No runtime clearance is needed for an altered line.

Focused core commands all passed (GOMAXPROCS=4, count=1, verbose logs, timeout 90s unless stated above):

- internal/lower: -run '^(TestStringDestructuring.*|Test.*Destructur.*)$'. Package 0.434s; new refusal leaf 0.14s.
- internal/oracle: -run '^TestNativeAgreesWithNode_097$/internal/oracle/testdata/string_destructuring[.]a$', ADAMIC_GATE_UNCACHED=1. Fixture leaf 9.34s, package 9.396s. Both backends held to source Node, with native ASan/UBSan, release and counted/leak checks. The earlier source-report selector also actually ran the fixture, 24.54s; the explicit selector was verified rather than assuming a no-test success.
- internal/ir: -run '^TestCallTargetReaders$' -timeout 90s. Test 41.35s; all named readers passed.
- internal/flow: -run '^(TestFlowCorpusRemainder|TestFlowCorpusUnitsCoverEveryProgram|TestFlowCorpusSetupIsShared)$'. Package 7.044s; new fixture remainder leaf 0.22s, all four flow analyses.
- Guards across internal/ir, internal/lower, cmd/adamic-gate and internal/oracle: TestEveryMapAProgramHoldsIsSorted, TestNativeAgreementRowsNameRegisteredPrograms, TestEveryTestIsParallelOrSaysWhy, TestEveryDuplicatedProductDeclarationAgrees and TestRuntimeSetsAreEachUsed. Passed; 1,330 product declarations checked.
- timeout-bounded go vet ./internal/lower ./internal/oracle: exit 0.

The added/touched behavior leaves are below 60s on this four-CPU cgroup: refusal 0.14s, oracle maximum 24.54s (explicit run 9.34s), flow 0.22s and counts regeneration 0.21s. These are test-leaf durations, not Go compilation command wall time. No over-budget leaf was rerun. The existing counts-product preparation/shard in the count-mutant command took 61.00s; its new fixture leaf took 0.12s, and that setup was not rerun.

The reverse production/test-import closure of lower and oracle contains 52 packages on this base, including the new internal/leakcheck reader. consumer-selection.json names the focused tests (or explicit compile-only selection for packages without a chosen behavior test). consumer-final-results.json records the completed closure and points to initial failures and their resolved prerequisite reruns. No whole package or full gate was run. Optional skips are enumerated in consumer-skips.json, with reasons retained in raw logs; they are not claimed as behavior passes.

Main's separate product ownership caused prerequisite failures in selected CLI/estree/lint readers. Those were resolved by running their declared owner tests first, without editing any product or another package's tests. CLI children also needed the existing default build-product cache explicitly inherited because that test changes XDG_CACHE_HOME. Initial failures remain in consumer-results.json/raw logs; final results retain the explanations and successful rerun logs. This was a test invocation prerequisite, not an inferred compiler fix.

All four source-report mutants were rerun on this tree:

| Mutation | Catcher |
|---|---|
| StringLength replaced with CharCodeAt at index zero | Both backend output comparisons reject 97 where Node prints 3; no sanitizer/compiler failure counted as a kill |
| Empty-pattern source evaluation removed | Both backend output comparisons reject source-call count 2 where Node prints 3 |
| Member refusal's fix removed | TestStringDestructuringMemberDiagnostic rejects the diagnostic lacking its fix |
| Counts family's recorded allocation count incremented from 34 to 35 | Both measurement-product and independent measurement comparisons reject 35 against 34 |

The first three use overlays regenerated from current production files; their sources are .go.txt and patches, never compilable Go in review. mutants.json and count-mutant.json contain invocations, results and catchers. Every invocation exited 1 for the intended reason, not build failure, warning, timeout or sanitizer error. Production files and the recorded counts were not mutated.

Toolchain setup: exported GOPROXY='https://proxy.golang.org|direct', ran timeout 600 bash cloud/setup.sh and sourced /workspace/adamic-tools/env.sh. Setup succeeded with Go 1.27.1, Node 24.19.0, clang 20.1.8; nproc=5, cpu.max=400000 100000 (four CPUs). setup.log.txt timing lines: node 0.431s; Go 0.449s; markdown dependency step 0.014s / ready 0.657s; clang 1.084s; submodules 1.411s; shared cache 2.927s; Go build 113.038s; test binaries deferred 113.160s; build cache warm 113.163s; done 113.285s. No setup workaround was needed. Cohere remained a referenced submodule; no code was copied from it.

Final audit passed: 52 reader packages checked, all selected runnable checks green after four prerequisite resolutions, and all four mutants caught. Five optional-input leaves skipped: cycle ledger and estree RecoveryLibraryGaps, TypeMemberLibraryGap, CookedSurrogateLibraryGap and PinnedNumericGaps. The initial consumer runner exits 1 because it correctly retains the four initial failures; final-audit.py validates each resolved rerun and every remaining result rather than overwriting those failures. prerequisite-commands.json records exact prerequisite/rerun commands. No test names silently matched nothing.

The staged whitespace check excluding literal review/**/*.patch context passes. Unified-diff context in the preserved mutant patches is kept intact. All fresh test logs use .log.txt and are tracked with the report.
