TestProduct_profile_controls_lowered: not defended after three production attempts.
The previous untrue verdict is disproved: D1 and D2 each fail this row.
Both kills are shared by TestProduct_corpus_lowered and all six other lowering-product rows.

Starting origin/main: bfe0553300773c0b37db2c10df97adeb909705f8.

CODE UNDER TEST: production Go loading, lowering and TSGo C emission invoked by the volume_suite.ts build recipe, specifically load.Load, lower.Lower, lowering.tsgo and native.TSGoC. Production function coverage is in functions.txt; per-test profiles preserve every reached block. The audit described suite construction as the code under test. This defense leaves that harness intact and mutates the actual compiler called by it.
ORACLE: self. The test requires the builder to load, lower, emit and write C without an error. It discards the returned path and never runs or compares the emitted program. No external authority decides its answer.

Coverage and semantic comparison: fresh-cache target and TestProduct_corpus_lowered runs each pass. Their production coverage has 4,560 shared nonzero blocks and zero exclusive blocks in either direction. Both tests call volumeProfileControlsLowered(typeAwareProductHarness(t), "") and load the same volume_suite.ts. This is an identical lowering recipe, not a Node/native executor pair; the twin exception does not apply.

Baseline and matrix boundary: the clean whole package reached its 90-second test timeout, with no preceding completed test failures. The seven current TestProduct_*_lowered rows pass together in 47.515 seconds. All seven were included in each mutation matrix, including the newly added TestProduct_corpus_lowered. Other tests reaching the compiler and semantic consumers of these products are outside this bounded matrix; their kills remain unknown. callers.txt records direct builder callers, and list.log records the complete current test census.

Attempts, frozen before their results:
D1 internal/lower/tsgo.go:27 flips the TSGo-enabled guard. All seven fail; target failure says Adamic refuses an unlinked typescript-go library call at volume_suite.ts:22:17. Binary 5.568 seconds.
D2 internal/native/tsgo.go:40 flips the unique-body count condition. All seven fail; target failure says native: missing unique tsgoProgram library body. Binary 78.076 seconds.
D3 internal/native/tsgo.go:70 drops the entire source-body replacement statement. All seven pass, binary 43.555 seconds. The emitted volume.c really changes: clean output has three return adamic_tsgo_* bridge calls; mutant output has zero and retains the tsgoProgram panic placeholder. D3-output-witness.json preserves hashes and the actual C definition before/after. This is an observed emission change, not an equivalent candidate. Execution of that changed C was not measured.
Each diff applies to the recorded origin/main and go vet passed for its mutated Go package. No tests, harnesses, or oracles were changed. Each mutant used its own fresh ADAMIC_BUILD_CACHE_DIR. Source was restored after every attempt.

Name/assertion finding: the name accurately identifies a lowering-product construction check. Its success assertion does not establish C content or executable semantics, illustrated by D3. The historical "profile controls" name does not mean this row executes the profile control cases; it builds volume_suite.ts.

Unclear or costly parts:
1. Audit code under test was construction, and its mutations were in _test.go builders. The defense bans those changes, so this report explicitly distinguishes the reachable production compiler from the untouched construction harness.
2. The audit predates the exact duplicate TestProduct_corpus_lowered. Current-source coverage and the matrix include it.
3. Whole-package baseline exceeded the 90-second binary budget, requiring a green bounded matrix. This prevents package-wide uniqueness claims, but observing another catcher is enough to defeat a proposed unique kill.
4. Compiler coverage and fresh product caches force real multi-megabyte C emission. Failed shared-product callbacks may be retried by another parallel test, which increased D2 cost.
5. Verdict instructions allow correcting an untrue row to subsumed, but the final defense enum lacks subsumed. rows.json uses defense="not defended" and observed_verdict="subsumed" to preserve both meanings.

Timing and omissions: warm toolchain setup skipped; nproc=5. npm ci in stage3/api reported 681 ms. Whole baseline 90.031 seconds. Coverage binaries 35.784 and 35.211 seconds. Bounded baseline 47.515 seconds. Mutation binary total 127.199 seconds. run-metadata.json includes each mutation's overall vet/build/run wall time. No package-wide replay, native execution, compiler tests in other packages, or test edits were performed. No tests should be deleted on this bounded evidence alone.
