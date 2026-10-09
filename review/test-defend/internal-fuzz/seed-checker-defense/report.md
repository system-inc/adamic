# Seed and checker defense

Main: b8bcadb2c493173855f19d7e5c508b34f5eeb5b6. Audit branch: test-audit/internal-fuzz, audited main 7b18d0576930caca4e22ce2eef92fcf563af52d0. All file:line references in plans and rows are against starting main. The existing defense branch addressed different rows; its evidence is preserved outside this subdirectory.

## Code under test and oracle

Code under test: internal/fuzz's source generator, entered through Generate, GenerateWithout and GenerateFeatures, then generator.program, expression and statement helpers, including the seed-driven ownership and override scenes. All nine mutations are in generate.go. No checker, loader, lowering implementation, test or harness was mutated.

Oracle: self for all three requested rows. OneSeed compares generated source for repeated seeds 1..20 and demands different source for seeds 1 and 2. GeneratedPrograms uses handwritten expectations that the TypeScript checker accepts generated sources and Lower returns no error. RegexPrograms uses handwritten checker-acceptance expectations for the same generator's seeds 1..30. These rows do not run Node. The checker and lowerer are the acceptance checks, not mutated code. GeneratedPrograms discards the returned IR.

## Baseline and coverage

Warm tools worked; setup skipped. npm ci in stage3/api succeeded in 412 ms. nproc=5. Clean package baseline passed in 13.330 test-binary seconds. All 18 current top-level rows were included, with no skips. The audit had 15; the added rows are TestReduceRejectsDifferentRefusal, TestExactSignatureRejectsSuffix and TestBytesSharedRejects63Bytes.

coverage.py records exact commands: separate -count=1 -run '^Name$' -coverprofile files with -coverpkg=./internal/fuzz, and the rest of the package excluding OneSeed for its comparison. Profiles and logs are saved. Coverage describes reached Go blocks, not an external oracle's internals.

OneSeed has no exclusive covered block versus the rest. Its semantic distinction is reproducibility and seed variation, rather than source validity. GeneratedPrograms has five exclusive blocks versus RegexPrograms: holder fallback at 495, constant-string comparison at 952, string-array slice at 1086, string-array map at 1088..1093, and override spread at overrides.go:51. These blocks are also reached by OverridesShapesAndLower. The broader 40-seed UndefinedNumbersShapes corpus additionally checks the same full Generate sources for seeds 31..40, which caught every G mutant.

RegexPrograms reaches regex and other enabled-feature blocks not covered by its audit subsumer OverridesShapesAndLower. Its full Generate seeds 1..30 are, however, an exact deterministic subset of GeneratedPrograms' seeds 1..60. GeneratedPrograms performs the same checker acceptance check. Shared-line input semantics explain the overlap without treating exclusive coverage as a verdict.

## Attempts and observations

The initial seven mutations were frozen in plan.json before running them. The follow-up G2 and G3 were frozen in plan-2.json before their runs. Each mutation was independently applied to clean source, vetted with go vet ./internal/fuzz, run against all 18 current tests with a fresh ADAMIC_BUILD_CACHE_DIR, and restored. Every standalone diff also passes git apply --check after restoration; diff-validation.json records this. No runtime selector remains in production source.

OneSeed: S1 fixes Generate's seed argument to 1. It fails OneSeed at fuzz_test.go:23, proving the prior untrue finding is overturned, and also fails OctoberFeaturesAppear, OwnershipShapes and UndefinedNumbersShapes. S2 zeros the scene-cadence seed while leaving PCG entropy unchanged: only ownership and override vocabulary rows fail. S3 swaps the two PCG constructor arguments: the whole package passes, including reproducibility. These latter attempts do not show an incorrect output; they show that changing seed plumbing can preserve the sampled reproducibility contract. No unique catch was established after three aimed attempts.

GeneratedPrograms: G1 changes the slice method constant to sliceMissing, G2 drops the guard that chooses a string variable to avoid comparing disjoint literals, and G3 changes map's receiver type from NumberArray to StringArray while preserving its Number item declaration. They all fail GeneratedPrograms and UndefinedNumbersShapes. G1 and G2 also fail OverridesShapesAndLower; G3 does not. RegexPrograms passes all three. The old named subsumer therefore does not cover these defenses, but the package still has another catcher. The subsumption hint here rests on these three attempts.

RegexPrograms: R1 drops the finder declaration, R2 changes exec to execMissing, and R3 changes the replacement receiver expression from String to Number. Each fails RegexPrograms, GeneratedPrograms, OwnershipShapes and UndefinedNumbersShapes. OverridesShapesAndLower passes all three. The old named subsumer is not a cover for these regex-specific breaks; GeneratedPrograms is the observed cover.

matrix.json contains every passing and failing top-level row, exact commands, diagnostic lines, and wall times. rows.json contains the requested per-row deliverable and current observed subsumers. The audit's prior subsumers were RegexPrograms for GeneratedPrograms and OverridesShapesAndLower for RegexPrograms; they are preserved here explicitly rather than misrepresented as current mutant catchers.

## Cooked run and limitations

S1's whole-package run hit the test binary's 90-second timeout (95.123 seconds including compilation). TestShrinkKeepsOnlyWhatFails searches successive generated programs for a delete marker. The fixed seed's program never supplies it, so that search never finishes. A timeout output event was tagged with another parallel test; it is not a production panic in that test. Raw S1.log is retained. S1-recovery reruns with only the stalled shrinker skipped, its own fresh cache, and finishes all other 17 rows in 12.754 wall seconds. It observes the same four failures. S1 results are bounded with the shrinker's completion unknown; no package-unique claim relies on it. All eight other whole-package runs complete within budget. There were no ordinary aborting production panics.

The final restored run checks the three requested rows. Production and tests are unchanged. Three finite attempts do not prove that a future unique break cannot exist, and this report recommends no deletion.

## Name versus assertions

OneSeed's assertions do check the stated sampled determinism contract. They do not establish equality for all seeds or pairwise distinction for all different seeds; no broader guarantee is inferred. GeneratedPrograms does check checker and lowering acceptance as named, but does not validate the returned IR. RegexPrograms does check checker acceptance, but it uses the full generator rather than a regex-only corpus and does not separately assert that each seed contains regex syntax. No performance promise or threshold applies to these rows.

## Brief issues and costs

The final JSON enum omits the prose verdict 'subsumed' for a previously untrue row that now fails alongside another catcher. rows.json uses 'not defended' for absence of a unique catch, while this report explicitly records that OneSeed is now demonstrably true and its S1 catch is shared.

/tmp is an 8.8 GB filesystem, so the requested 15 GB free threshold cannot be met there. Only the earlier named /tmp/deletion-unicode scratch directory was removed; /workspace had 18 GB free. Neither repository nor tools were removed. No disk-related baseline failure occurred.

The automatic permission reviewer timed out before one baseline command executed. A permitted retry succeeded. No task action remains blocked. The preexisting defense branch required fetching and preserving its unrelated evidence before adding this session.

Setup was skipped. Baseline binary time was 13.330 seconds; npm ci was 0.412 seconds. Nine mutant commands plus S1 recovery used 235.346 wall seconds in total, including compilation. Per-run wall timings are in matrix.json and recovery metadata. Coverage compilation/running was additional; coverage logs report binary times. Fresh native caches were used for every mutant. The only unfinished test behavior is S1's stalled shrinker. No other packages were replayed.
