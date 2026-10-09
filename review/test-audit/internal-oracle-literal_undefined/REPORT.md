u062: requested witness exists at starting origin/main 09769cb5ddc8.
Verdict: witness, proven by weakened stdout comparison.
Clean median: 0.369 binary seconds, nproc 5.
W01 and P01 fail both subcases at the intended comparison assertion.
Evidence: test-audit/internal-oracle-literal_undefined; harness restored.

```json
[
  {
    "test": "TestLiteralUndefinedOracleCatchesMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/literal_undefined_test.go",
    "seconds": 0.369,
    "oracle": "Actual Node runs original .a source; disagreement compares stdout bytes. Self-written required classification stdout differs, with clean native exit/stderr preconditions. The row expects that classification rather than precise differing bytes.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "W01"
    ],
    "unique_kills": [],
    "last_proven_fail": "W01 literal_undefined_test.go:55: got \"\", want stdout differs",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLiteralUndefinedOracleCatchesMutant"
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT=W01 ADAMIC_BUILD_CACHE_DIR=/tmp/u062/cache/W01 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^TestLiteralUndefinedOracleCatchesMutant$' > W01.log 2>&1; literal_undefined_test.go:55: got \"\", want stdout differs",
    "mutation_kind": "witness-check",
    "production_mutants": 0
  }
]
```

| ID | Kind | Origin location | Change | Failed rows |
|---|---|---|---|---|
| W01 | witness-check | internal/oracle/oracle_test.go:720 | Drop stdout case and its return | TestLiteralUndefinedOracleCatchesMutant, both fixture subcases |
| P01 | empty-answer probe | internal/oracle/oracle_test.go:716 | Return empty string at entry | TestLiteralUndefinedOracleCatchesMutant, both fixture subcases |

Survivors: none in the bounded witness/probe runs.

# u062 literal undefined witness audit

Starting origin/main: 09769cb5ddc8067d2fa052a8c760d813894c8919. The requested name exists in go test -list . ./internal/oracle/ and remains in internal/oracle/literal_undefined_test.go. No move, vanished row, family grouping or helper classification applies. Its two fixture subcases remain one top-level row.

CODE UNDER TEST: the agreement check disagreement in internal/oracle/oracle_test.go, as witnessed by TestLiteralUndefinedOracleCatchesMutant. The compiler and C runtime prepare the witness's built-in production bug but are not mutated in this audit. ORACLE: Node running original .a fixture source, and the self-written stdout differs classification. The test body, Node runner and fixtures are unchanged. The only harness changes are the explicitly permitted witnessed-check weakening and its empty-answer probe.

functions.json lists all oracle-package helpers on the uncached row path, plus TestMain. Anonymous callbacks are the subtests, bounded cleanup, gateOnce.Do, sourceIdentity.visit and cachedNode execution callback. Package fixture-registration init functions are setup, not calls from the row. Preparation crosses into load.Load, lower.Lower, native.C and native.Build; no files in those packages were changed. The fixed plan and function inventory were saved before weakened-check runs. callers.txt records other uses of disagreement.

W01 removes the stdout switch case and its return, while preserving exit/stderr comparisons. It fails both fixture subcases only at literal_undefined_test.go:55: got "", want stdout differs. P01 returns empty at disagreement entry and likewise fails both subcases. Its standalone diff replaces the whole comparison body with return "" so go vet sees no unreachable statements; its switched form returns conditionally at entry. The control and restored original row pass. Every matrix trial uses ADAMIC_GATE_UNCACHED=1; logs report node hits=0 misses=2. The original C harnesses still compile and finish cleanly under sanitizers: a failed native precondition would appear at line 52, not the observed comparison assertion at line 55.

This row is a witness, so the brief's production-mutant quota/uniqueness/subsumption rules do not decide its verdict. One weakening of its guarded stdout comparison decides witness. W01 is labeled witness-check everywhere; P01 is only a probe. No survivor, production kill, package-unique kill or repo-wide uniqueness is claimed. The test's built-in generated-C mutants are not new standalone origin mutations from this session.

Both standalone diffs were applied separately to exact starting source and passed go vet ./internal/oracle/. validation.json has commands, statuses and times. switch.patch preserves the single compiled selector; all harness edits were restored before the evidence commit. No compiler/native-product cache changes are needed for a comparison edit, but each switch run still has its own ADAMIC_BUILD_CACHE_DIR. Node observation caching is explicitly bypassed.

Baseline and bounded scope: full package baseline reached the test binary's 90-second timeout with no preceding Test fail event. The witness had not completed in that run. Its three separate uncached clean runs passed, so the matrix was narrowed to this unit's one witness row. Other disagreement callers were found and listed but not audited. Their outcomes and uniqueness outside the bounded row are unknown. All standalone package vet checks and the restored row passed.

Brief ambiguities and costs:
* The supplied commit/file example names 8de93800f4. Required fresh origin/main was 09769cb5ddc8067d2fa052a8c760d813894c8919; every location uses that starting commit.
* The request says never mutate the harness, then explicitly allows weakening the check for witnesses. The witness exception applies here; production mutations would measure its preconditions rather than its guarded comparison.
* The general request for roughly three production mutants conflicts with witness judgment. Only one relevant stdout-check weakening was needed; no verdict rests on production mutants.
* A complete package run does not fit 90 seconds. It consumed 92.041 wall seconds, then the audit narrowed to the requested witness. Other callers remain unknown.
* Cached Node results would weaken the claim that Node actually ran this session. All row trials bypassed result caching, and additional direct Node probes saved original fixture stdout.
* The parent's PASS line is 0.00 seconds because its fixture subtests run in parallel. Timing therefore uses the permitted test binary ok line: 0.369, 0.380, 0.343 seconds.
* The row asserts the disagreement category, not exact divergent bytes. This is appropriate to its witness role, but it does not itself validate the unmutated native product. No claim about that product's correctness is made here.
* Mandatory npm ci is not a dependency of these two direct fixtures but was run. No node_modules references were found in oracle test files. Warm setup was skipped.
* The kills field has no separate witness-mutation schema. This report records W01 there with mutation_kind witness-check and production_mutants 0; unique_kills is empty. Empty-answer P01 appears only in probe_kills.
* Seven out-of-unit rows skipped in the default full baseline. They are listed in baseline-skips.json; no requested row skipped. Those other opt-ins were not enabled or audited.

No other packages' tests, full-package mutation replay, production compiler mutations or repo-wide replay were run. No main push or pull request was made.

Timing summary:
```json
{
  "origin_commit": "09769cb5ddc8067d2fa052a8c760d813894c8919",
  "nproc": 5,
  "setup_seconds": 0,
  "npm_seconds": 0.4784249309996085,
  "scope_seconds": 6.480558112001745,
  "baseline_seconds": 92.04136136899979,
  "baseline_binary_seconds": 90.099,
  "clean_binary_trials": [
    0.369,
    0.38,
    0.343
  ],
  "clean_binary_median": 0.369,
  "clean_timing_wall_seconds": 7.268475207001757,
  "switch_build_seconds": 6.6871489210025175,
  "matrix_wall_seconds": 7.690392622000218,
  "standalone_validation_seconds": 0.7368001830000139,
  "restored_wall_seconds": 2.606543498000974,
  "baseline_skips": [
    "TestEntriesAcceptance",
    "TestStage3FixtureHook",
    "TestWASIAgreesWithNode",
    "TestWASIEmission",
    "TestWASIOracleCatchesMutants",
    "TestWASIRunnerCatchesMutants",
    "TestWASIShardPlantedFixture"
  ]
}
```
