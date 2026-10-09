# u088 test quality audit

Started from origin/main 571e74cf555b9db994c5dec6c2f8dbee676e5111.
Branch test-audit/stage1-cohere-estree-scalars. Evidence only; production and harness source restored.

## Result

14 rows after grouping: 1 bounded sacred, 3 witness, 1 setup-check, 7 untrue, 2 cannot-judge.
All 42 isolated clean timing runs passed. Whole baseline and combined cold slice cooked at 90.033s and 90.046s without an assertion-failing test event before timeout. Restored combined scope passed in 15.972s, with no skips.

CODE UNDER TEST, named before mutations: the TypeScript ESTree port's parsing, conversion and canonical serialization; construction/process control for setup rows. Go cohere and npm implementations are the ORACLE, never mutation targets. The two scalar tests call only those oracles, not the port, so their port worthiness cannot be judged without violating the brief.

reached-functions-all.json records 271 source-port V8 function ranges (229 named) from 16 clean driver cases. This reach inventory was collected before mutants were frozen. It covers the production matrix's port entry; construction-functions.txt is a conservative separate declaration inventory. This does not claim native C coverage or dynamic coverage of every setup helper.

The native products for M01, M02, M03 and P01 each rebuilt in their own cache. Each matrix ran every parser shard and its union. Outside-slice kills are unknown. Unique kills are bounded candidates, not package-wide uniqueness. The witnesses and construction checks have their own runs, excluded from production kills. There are no subsumption findings.

M03 survived all live cases. On `type X = { ) };`, source Node stderr changed from diagnostic 1131 to 1132 at byte 11, while the matrix stayed green. Both runs exited with a parser panic and empty stdout. This demonstrates an unguarded diagnostic code, not an equivalent mutant.

P01 makes answer return its empty string at entry. CLI wrapping still prints one newline. All actual accepted/refused cases reject that empty answer; empty shards and the union have no port input to exercise. The family is not vacuous.

W01 disables firstDifference. The planted disagreement and mapped-constraint witness fail. TestSyntaxMutantsUnion still passes because its child fails even without the plant, providing the expected mutant-survived signal. W02 disables portStallControlResult. Its planted survivor detects this, while the actual deadline-control family remains green. W03 disables the acceptance comparison and both syntax control leaves remain green.

S01 removes the process-group setting and the deadline construction check fails. S02-S05 write the expected artifacts under `missing`; all targeted product wrappers pass despite absence of oracle, port.c, port or ready.json. Artifact directory evidence is in Sxx-artifacts.json. These are untrue for the tested construction contract, not a claim that compiler errors cannot fail the builders.

## Commands and validation

Full commands and logs are in results.json, matrix.json, mutation-status.json, harness-status.json and timings.json. Raw logs are ignored locally; lossless .log.gz copies are committed. Every standalone diff applies to the starting commit. All eight Go witness/construction diffs pass go vet ./stage1/cohere/estree/. All four port/probe diffs passed their native build inside the matrix, with no compiler-warning kill.

Harness switches compiled once. Normal dependencies were built first, then copied using hard links to each construction cache. Only the targeted product keys were invalidated; manifests list every invalidated key. Consequently construction changes executed and could not be hidden by an old artifact. The harness switch is evidence, not a standalone production mutation. Standalone W01/W02 remove entire comparison bodies; W02 also removes its unused fmt import.

Failure references in results.json are mapped to original source lines. Raw switched-source line numbers are preserved in logs, with harness-line-map.json and harness-switch.diff documenting the mapping.

## Cost

Warm setup skipped, 0s; nproc 5. stage3/api npm ci ran before baseline. A fresh pinned npm install took about 4s; npm ci verified its generated lock later, before the final clean run. Installation time was not otherwise separately instrumented. Versions: @typescript-eslint/typescript-estree 8.65.0, typescript 6.0.3, prettier 3.9.6.

Each row/family ran alone three times, count=1. Medians use the test binary's own ok line, not JSON elapsed events. Product reuse affects the medians; they are not cold-build medians. Native/compiler output caches were isolated for port mutations, but oracle reference results and unchanged runtime compilation could be reused.

| ID | Command wall seconds including build/run | Build miss lines |
|---|---:|---|
| M01 | 66.906 | stalls_test.go:81: build estree-misc-sanitized-emitted a5aaa048030c miss 56.24<br>stalls_test.go:89: build estree-misc-oracle-answers 4bcaf8a65eed miss 0.19 |
| M02 | 66.401 | stalls_test.go:81: build estree-misc-sanitized-emitted b050927616b9 miss 55.69<br>stalls_test.go:89: build estree-misc-oracle-answers 4bcaf8a65eed miss 0.11 |
| M03 | 71.322 | stalls_test.go:81: build estree-misc-sanitized-emitted bcf0cb63dc60 miss 57.81<br>stalls_test.go:89: build estree-misc-oracle-answers 4bcaf8a65eed miss 0.19 |
| P01 | 45.224 | stalls_test.go:81: build estree-misc-sanitized-emitted 3fa694d812aa miss 32.03<br>stalls_test.go:89: build estree-misc-oracle-answers 4bcaf8a65eed miss 0.23 |

The first clean stall-control run cost 61.922s; first isolated setup-family run cost 40.750s. Harness cache preparation commands cost 45.430s, 34.291s, 31.594s. Witness W02 cost 63.720s including a port rebuild. Go vet checks totaled about 5.948s. Final clean binary 15.972s. Per-product times are the instrumented build lines, which include dependency work when a builder nests other products. Pure compiler CPU and every individual clang invocation were not isolated.

## Brief problems and uncovered scope

- The opening says 15 rows, then gives 83 functions. Grouping plus branch-specific assertions gives 14 rows in this audit.
- The pinned location commit is old. TestProduct_SyntaxMutantsSetup vanished as a single name; three replacements in loom_family_products_test.go exist and were audited. The other 82 names remain. Thus the requested scope plus replacements has 85 functions. TestProduct_SyntaxMutantNative_000 exists outside the named slice and was not added; the native family is bounded to its two requested members.
- ScalarOriginalLibraries and PinnedNumericGaps test Go/npm relationships and never invoke Adamic. The instruction to mutate only the port prevents a worthiness experiment on them. The NaN oracle checks only a count, and the Node numeric expectations are handwritten.
- About three mutants per row, a 20-mutant cap and at most four native rebuild mutants cannot all be met here. Three production mutants were frozen from three reached functions. Witness and construction edits are separate.
- Production mutation matrix against a cold whole package cannot finish in 90s. Even the combined cold slice cooks; isolated clean families completed. No assertion-red baseline was audited.
- Source TypeScript has no environment getter; separate plain port mutations/rebuilds were used instead of a switched port. They remain replayable against origin/main.
- Witness leaves 001/002 assert the bad port accepts Go-refused input. They do not run a disagreement checker; weakening their acceptance assertion leaves them green. The union's false-positive signal was demonstrated separately.
- Builders are called without artifact validation. Missing-artifact construction mutants remain green. The evidence distinguishes that missing contract from compile-error preconditions.
- answer is the processing entry behind the CLI adapter. Its empty string prints a newline through main.ts. There is no separate void main function to return from at module entry.
- The V8 inventory describes clean source executions, not native machine-code coverage. Setup helpers have a conservative declaration inventory rather than complete dynamic reach coverage.
- Whole-package corpus/benchmark opt-ins were outside this slice and not enabled. The pinned library opt-in was enabled; no requested row skipped. Whole-package timeout occurred before optional rows completed, so their skip status is unknown.
- A location-mapping command initially omitted the warm env and could not find gofmt. It made no source changes; rerunning with env sourced corrected it. The first line-map capture occurred after restoration, so the exact switched source was reconstructed and formatted to recover correct original-line mappings.

No repository-wide replay, complete cold package run, other-package tests or off-slice uniqueness was covered. No main push or pull request.
