# Defense of taste representation rows

Source base: 619e7a4cf33741cc04bc78dc4c0c8ba0e31d73fb, fetched origin/main this session. All three target names and their subsumers exist among 276 current top-level tests. The current whole package, including newly added tests, was used for every mutant matrix.

CODE UNDER TEST: Adamic's Go lowering production code, specifically property representation refusals in property, buffer checks in typedArrayUnsupported, and property-read contract attachment in readObjectField and readinessStatement. ORACLE: handwritten NotYet/Refused class and diagnostic fragment expectations for the taste and typed-array gaps rows, and handwritten transitive IR object/boolean contract expectations for the view-object row. These target rows use self oracles; no Node, Go cohere, test, or harness was mutated. The eraser is not executed by the metadata row.

## Findings

All three rows are defended. D1 changes the boxed-union refusal's diagnostic constant at object.go:525. It preserves NotYet but loses the precise reason. Only TestTasteRepresentationLimitsStayExplicit fails, at taste_representation_test.go:28. Its enum subsumer asserts error class and never reaches that boxed-union branch. This defense is about diagnostic specificity, not an admission or output check.

D2 changes the resizable ArrayBuffer argument bound from >1 to >2 at typed_arrays.go:49. A valid two-argument resizable-buffer construction is now diagnosed as generic ArrayBuffer. Only TestTypedArrayGaps fails, at typed_arrays_test.go:45. TestTypedArraysLower exercises supported typed arrays, not unsupported resizable buffers. This defense likewise guards a distinct diagnostic boundary; the mutated code still refuses the input.

D3 drops the initial property-read ViewContract assignment at object.go:586. All enabled rows pass. readinessStatement restores the contract using ViewTypeID, so this is an equivalent candidate at final-IR level, not demonstrated unguarded behavior. This failed attempt led to D4, which changes the final attachment at readiness.go:293 to contract ID zero. Only TestViewObjectContractsAreAvailableToEraser fails, at view_objects_test.go:23: object read has no semantic contract for the eraser. Its subsumer builds a recursive contract graph directly and does not lower property reads or run readiness. D4 changes actual lowered IR metadata. No fixture names, target-specific selectors or test edits were inserted in production.

Each D1, D2 and D4 completed matrix has 273 passing top-level tests, the single target failure, and two top-level skips. D3 has 274 passes and two skips. TestOriginalCycleLedger and TestOptionalWideningCensus skipped. TestMixedUnionContractGraph also has an explicitly unimplemented array subcase that skipped. Skips are unknown, not passes. Exact passed/failed/skipped names are in matrix.json and rows.json. Uniqueness is established across all enabled current package rows, not across the repository or skipped tests.

## Coverage

Six clean per-test coverage runs used -coverpkg=./internal/lower and -run '^Name$'. All passed. Target-minus-subsumer exclusive Go blocks: taste 792, typed gaps 237, view objects 1694. coverage-differences.json contains the complete block lists. D1's object.go:525 refusal block, D2's typed_arrays.go:49 resizable-buffer block and D4's readiness.go:293 final metadata attachment block are each exclusive relative to their named subsumer. Coverage exclusivity is only the lead; whole-package mutant observations establish the catches. Raw profiles, function inventories and the coverage driver are saved.

## Validation and replay

Every standalone diff applies to the stated origin/main base and passes go vet ./internal/lower using its production-only Go overlay. There are no switches in the standalone diffs. Apply the desired diff to the base and use its own ADAMIC_BUILD_CACHE_DIR, then run timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . with output redirected to a log. The observed overlay commands are recorded in rows.json. Overlays replace one production source file each; tests and oracle inputs are unchanged. Native products have separate compiler-mutant caches. Production source in the working checkout was never changed.

## Baseline, tools and timings

Warm /workspace/adamic-tools/env.sh worked; setup skipped. nproc=5. npm ci in stage3/api completed in about 1 second before baseline. Whole-package clean baseline cooked at 90.096 binary seconds, with no test failure before timeout. It completed 272 top-level rows; the four unfinished rows, TestOptionalWideningAllowed, TestPredicateUseRegions, TestSuppressionDirectivesAreRefused and TestUnprovenPredicateReturnsAreRefused, all passed in a 3.772-second recovery run. Combined clean observations cover every enabled row.

Clean coverage binary seconds: taste 0.762; enum subsumer 2.654; typed gaps 1.746; typed-array subsumer 2.315; view objects 0.105; recursive subsumer 0.112. Whole-package mutant binary seconds: D1 67.292, D2 61.748, D3 63.220, D4 59.781. Each remained below 90 seconds, so no mutant matrix needed narrowing. Total logged test-binary time is about 353 seconds; elapsed work about 11 minutes. These times include any builds performed inside tests; independent native-product build times were not extracted. No three-run timing medians were requested for this defense.

## Brief feedback and limits

1. The requested 15 GB threshold cannot be met on /tmp, whose total capacity is 8.8 GB. Before cleanup /tmp had 6.3 GB free and /workspace 6.3 GB. Identified earlier /tmp/defend-native-math and /tmp/defend-enum-* scratch/cache directories were removed. Recheck showed 6.4 GB and 6.3 GB respectively. These are separate filesystems, so removing /tmp caches cannot free /workspace. No disk-full failure occurred, and repo/tools were retained.
2. The full-package baseline exceeded 90 seconds, but four-row recovery completed clean coverage of enabled rows. The matrices subsequently fit in 60 to 67 seconds. Whether a package cooks depends on cache warmth and compilation load, so a baseline timeout should not permanently force a narrower matrix.
3. The requested remote defense branch already existed and contained evidence for other rows in this audit unit. This session is stored in session-619e7a4c-20261009T1718 under the requested evidence directory, and the remote history is merged without overwriting its evidence or force pushing. Naming branches only by the broad audit unit creates collisions between distinct defender assignments.
4. The audit's typed-array failure line was 55; current origin/main places the assertion at line 45. All mutant production locations and coverage profiles use the actual starting commit. Tests should be identified by current go test -list, as the brief correctly directs.
5. The initial metadata assignment is repaired later. Coverage identified it as a lead, but the first mutant did not change the final answer. No survivor is called unguarded without a changed output witness. The second attempt targets the final metadata construction.
6. The taste row's enum-prototype failure text says want NotYet while its assertion correctly checks Refused. The name's explicit-boundary promise is checked. The metadata row verifies availability of transitive contracts rather than running the eraser, which matches its name but must not be described as an eraser execution test. No row remains undefended.

No test was deleted, rewritten or weakened. No pull request was opened. Repo-wide uniqueness, the skipped corpus-dependent rows, and eraser execution were not covered.
