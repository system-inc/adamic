u164: three named functions found in their listed files; none moved or vanished.
Base cf735d9fba9e38de6368575e5630e44375a86eaf; nproc 5; warm setup skipped.
Corrected whole-package baseline passed 17.604s; restored baseline passed 16.798s.
Verdicts: one helper, one witness, one setup-check; W3 is a proven survivor.
Evidence: test-audit/stage3-fixtures-runner_guard under review/test-audit/stage3-fixtures-runner_guard/.

```json
[
  {
    "test": "TestTransformedNodeRunnerGuardHook",
    "package": "stage3/fixtures",
    "file": "stage3/fixtures/runner_guard_test.go:10",
    "seconds": 0.01,
    "oracle": "Self-written output text shape assertions. Active helper reads the derived runner; it never executes Node. Default whole-package entry skips unless its parent supplies ADAMIC_RUNNER_GUARD_REPOSITORY.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "helper",
    "parent": "TestTransformedNodeRunnerGuard",
    "subsumed_by": [],
    "mutants_in_matrix": 6,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestFixtures family",
      "TestPrepareFixtureOracleHook",
      "TestFixturePaths",
      "TestFixtureShardManifest",
      "TestTransformedNodeRunnerGuardHook",
      "TestTransformedNodeRunnerGuard",
      "TestFixtureDirectoriesHaveTopLevelTests"
    ],
    "evidence": "ADAMIC_RUNNER_GUARD_REPOSITORY=/workspace/adamic timeout 120 go test -json -count=1 -timeout 90s ./stage3/fixtures/ -run '^TestTransformedNodeRunnerGuardHook$' > P1-hook.log 2>&1; runner_guard_test.go:19: open : no such file or directory",
    "timing_mode": "helper activated with repository environment"
  },
  {
    "test": "TestTransformedNodeRunnerGuard",
    "package": "stage3/fixtures",
    "file": "stage3/fixtures/runner_guard_test.go:27",
    "seconds": 0.054,
    "oracle": "Self-written nine-input accepted/rejected table. Valid inputs require successful child exit. Invalid inputs require failed child exit and the exact shape-guard diagnostic marker, so a later hook failure is rejected. No Node execution.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "witness_kills": [
      "W1",
      "W2"
    ],
    "last_proven_fail": "W2 runner_guard_test.go:62: unexpected runner must fail at the shape guard: {Stdout:--- FAIL: TestTransformedNodeRunnerGuardHook (0.00s)",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 6,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestFixtures family",
      "TestPrepareFixtureOracleHook",
      "TestFixturePaths",
      "TestFixtureShardManifest",
      "TestTransformedNodeRunnerGuardHook",
      "TestTransformedNodeRunnerGuard",
      "TestFixtureDirectoriesHaveTopLevelTests"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage3/fixtures/ -run . > W2.log 2>&1; runner_guard_test.go:62: unexpected runner must fail at the shape guard: {Stdout:--- FAIL: TestTransformedNodeRunnerGuardHook (0.00s)",
    "surviving_weakenings": [
      "W3"
    ]
  },
  {
    "test": "TestFixtureDirectoriesHaveTopLevelTests",
    "package": "stage3/fixtures",
    "file": "stage3/fixtures/units_test.go:19",
    "seconds": 0.013,
    "oracle": "Self-written suite-construction rule: every fixture directory has one direct top-level TestFixtures declaration naming its literal directory; actual declarations are parsed and compared to filesystem directories and status rows.",
    "oracle_kind": "self",
    "kills": [
      "S1",
      "S2",
      "S3"
    ],
    "unique_kills": [
      "S1",
      "S2",
      "S3"
    ],
    "last_proven_fail": "S3 units_test.go:51: TestFixturesTaste must call testFixtureDirectory directly",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 6,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": false,
    "matrix_rows": [
      "TestFixtures family",
      "TestPrepareFixtureOracleHook",
      "TestFixturePaths",
      "TestFixtureShardManifest",
      "TestTransformedNodeRunnerGuardHook",
      "TestTransformedNodeRunnerGuard",
      "TestFixtureDirectoriesHaveTopLevelTests"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage3/fixtures/ -run . > S3.log 2>&1; units_test.go:51: TestFixturesTaste must call testFixtureDirectory directly",
    "probe_note": "P2 empties the static unit declaration registry. This is not a callable-entry probe, so callable-entry vacuity is not assigned."
  }
]
```

| ID | File:line on origin/main | Change | Failed rows |
| --- | --- | --- | --- |
| W1 | stage3/fixtures/fixtures_test.go:133 | Disable the complete shape condition with impossible erasable+transformedCalls < 0 | TestTransformedNodeRunnerGuard |
| W2 | stage3/fixtures/fixtures_test.go:133 | Runtime URL count != 1 becomes > 1 | TestTransformedNodeRunnerGuard |
| W3 | stage3/fixtures/fixtures_test.go:133 | Total stripTypeScriptTypes prefix count != 1 becomes > 2 |  |
| S1 | stage3/fixtures/fixtures_test.go:148 | Drop TestFixturesAssertions sole call statement | TestFixtureDirectoriesHaveTopLevelTests |
| S2 | stage3/fixtures/fixtures_test.go:149 | Change TestFixturesCycles directory constant cycles to enums | TestFixtureDirectoriesHaveTopLevelTests |
| S3 | stage3/fixtures/fixtures_test.go:158 | Return early from TestFixturesTaste by replacing its sole call with return | TestFixtureDirectoriesHaveTopLevelTests |

Survivor W3: a scratch source containing one recognized transform call, one unknown-mode call and one runtime URL is refused before W3 (helper exit 1 with the shape-guard diagnostic) and accepted after W3 (helper exit 0). The full original package, including the nine-case witness, passes under W3. This demonstrates an uncovered hybrid-call guard weakening; it is not an equivalent candidate. See W3-hybrid-before.log, W3-hybrid-after.log, and hybrid-runner/oracle/node.mjs.

Probes: P1 returns the empty path at transformedNodeRunner entry. Active helper fails with open : no such file or directory; its parent rejects all nine subcases. P2 removes all eleven static fixture unit declarations. The directory checker reports missing units and an empty directory union. Neither probe contributes a mutant kill.

Brief friction, interpretations and practical limits:

1. Starting origin/main is cf735d9fba9e38de6368575e5630e44375a86eaf, newer than the supplied 8de93800f4. All three names remain in their listed files. The complete discovery list is saved.
2. These are harness quality rows. The helper and witness do not run Node, despite their names. Their expected runner text, accepted shapes, exit categories and diagnostic marker are hand-written self oracles. Live Go subprocess execution of our own test binary is not an independent authority.
3. The helper intentionally skips by default. It belongs to TestTransformedNodeRunnerGuard, which activates it in child processes. I activated it against the real repository for all three standalone Good timings and its empty-entry probe. In every whole-package run, only this helper entry skipped. No substantive row was skipped.
4. The parent deliberately supplies broken runner shapes and expects the shape check to reject them. It was judged as a witness by weakening that check. W1 and W2 prove the witness can reject a disabled or loosened check. W3 shows this proof does not cover every guard component: the built-in table lacks a recognized-plus-unknown hybrid call.
5. W2 is caught even though the child still fails its later output-shape assertion. The parent requires the specific guard diagnostic, not just a nonzero exit. That stage requirement is demonstrated by runner_guard_test.go:62 in W2.log.
6. The directory row is a setup check. Its code under test is the eleven direct fixture-unit declarations, not Adamic lowering or its fixtures' recorded Node answers. S1 keeps a discoverable Test name but removes its work; S2 makes two declarations name enums and leaves cycles without a unit; S3 returns without running taste. All three construction failures are caught without changing the checker or any manifest.
7. A runtime mutation switch cannot represent these construction changes cleanly. The checker inspects the actual AST and requires one direct call statement. An inactive switch would already violate that condition. I replayed each standalone diff separately instead of adding a switch to the declarations.
8. The exact callable-entry empty-answer rule does not map to static declarations. P2 supplies an empty construction registry, and the checker fails, but its vacuous field remains null. No standard-library parser or checker assertion was changed to manufacture an empty answer.
9. W mutations use the explicit witness exception for harness edits; S mutations use the setup-check exception for construction edits. No compiler, production port, oracle/node.mjs, recorded behavior, or test assertion was changed. Witness catches are recorded separately from production kills. Setup kills are labeled construction mutations; their uniqueness is over the complete package matrix.
10. The package baseline fit the budget, so all six mutants ran against the complete package. The eleven wrappers that differ only in directory input were grouped into TestFixtures family for matrix counts. This matrix is not bounded to the three scoped rows, and no subsumption verdict is assigned to these special rows.
11. /usr/bin/time was absent. My first npm timing command therefore did not execute npm ci, although the initial package run passed. I corrected this before any mutation: npm ci succeeded in 0.503 seconds, then a replacement whole-package baseline passed. The initial baseline is retained as evidence but is not the qualifying baseline.
12. The scope says three rows, but only two have a check verdict: the helper has its parent instead. I used three guard weakenings and three declaration changes, approximately three mutants per applicable check. No native rebuild is required for these Go harness/construction mutations. Each diff compiled with go vet ./stage3/fixtures/.
13. Guard coverage concerns exact string spellings, not semantic JavaScript equivalence. The source runner is accepted through the known stripTypeScriptTypes spelling and runtime URL spelling. This audit demonstrates the existing string guard and the missing hybrid-input witness, not robustness against every formatting or JavaScript syntax variant.
14. No compiler correctness, fixture semantic coverage, runtime ownership, repo-wide tests or unrelated package audit was attempted. No production mutation was retained. The survivor input is synthetic text for the guard; it is never executed as JavaScript.

Timing: setup 0s, nproc 5, npm 0.503s. Nine isolated timing runs totaled 0.240 binary seconds and 20.260 command wall seconds. Six package matrix runs totaled 132.625 wall seconds. Per-run compile validation and all command durations are in runs.json. The first exploratory baseline reported oracle-hook preparation 8.39s. Unit work ran approximately 13:47 to 14:10 UTC.
