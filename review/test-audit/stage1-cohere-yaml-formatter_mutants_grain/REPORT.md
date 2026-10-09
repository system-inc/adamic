u151: starting origin/main a7448d73cd17f16362b6cbc5c5c111080da64e43; nproc 5.
All 26 named Tests exist; family grouping yields four rows.
Verdicts: two setup-check, one witness, one untrue witness family.
Three empty-entry probes: two product rows and the six-witness family pass their own probes.
Evidence pushed under review/test-audit/stage1-cohere-yaml-formatter_mutants_grain/.

```json
[
  {
    "test": "TestProduct_YAMLFormatterMutantsOracle",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/formatter_mutants_grain_test.go",
    "seconds": 1.008,
    "oracle": "Go cohere generates the expected-output artifact, but this row checks self-written corpus counts, wrapper enumeration, union and required artifact availability, without comparing formatter answers.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S1: formatter_mutants_grain_test.go:130: open /tmp/u151/cache/S1/f5e6f0a5be2887bad0408ac8ed7e255631d937ca100a8237f84b54632f07bb6b/cases.txt: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_YAMLFormatterMutantsOracle",
      "TestProduct_YAMLFormatterMutant family",
      "TestFormatterMutants family",
      "TestFormatterMutantsPlantedFailure"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u151/evidence/S1.overlay.json -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestProduct_YAMLFormatterMutantsOracle|TestProduct_YAMLFormatterMutant000Sources|TestProduct_YAMLFormatterMutant000Lowered|TestProduct_YAMLFormatterMutant000Native|TestProduct_YAMLFormatterMutant001Sources|TestProduct_YAMLFormatterMutant001Lowered|TestProduct_YAMLFormatterMutant001Native|TestProduct_YAMLFormatterMutant002Sources|TestProduct_YAMLFormatterMutant002Lowered|TestProduct_YAMLFormatterMutant002Native|TestProduct_YAMLFormatterMutant003Sources|TestProduct_YAMLFormatterMutant003Lowered|TestProduct_YAMLFormatterMutant003Native|TestProduct_YAMLFormatterMutant004Sources|TestProduct_YAMLFormatterMutant004Lowered|TestProduct_YAMLFormatterMutant004Native|TestProduct_YAMLFormatterMutant005Sources|TestProduct_YAMLFormatterMutant005Lowered|TestProduct_YAMLFormatterMutant005Native|TestFormatterMutants_000|TestFormatterMutants_001|TestFormatterMutants_002|TestFormatterMutants_003|TestFormatterMutants_004|TestFormatterMutants_005|TestFormatterMutantsPlantedFailure)$' > S1.log 2>&1; formatter_mutants_grain_test.go:130: open /tmp/u151/cache/S1/f5e6f0a5be2887bad0408ac8ed7e255631d937ca100a8237f84b54632f07bb6b/cases.txt: no such file or directory",
    "raw_failing_line": "formatter_mutants_grain_test.go:130: open /tmp/u151/cache/S1/f5e6f0a5be2887bad0408ac8ed7e255631d937ca100a8237f84b54632f07bb6b/cases.txt: no such file or directory",
    "members": [
      "TestProduct_YAMLFormatterMutantsOracle"
    ],
    "timing_samples": [
      1.118,
      0.957,
      1.008
    ],
    "construction_kills": [
      "S1"
    ],
    "weakened_check": null
  },
  {
    "test": "TestProduct_YAMLFormatterMutant family",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/formatter_mutants_grain_test.go",
    "seconds": 8.407,
    "oracle": "Self-written successful product construction at source, lower and native levels, for six mutant recipes. Successful preparation is the whole assertion; the returned products are not checked by these wrappers.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S2: formatter_mutants_grain_test.go:450: mutant preparation failed",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_YAMLFormatterMutantsOracle",
      "TestProduct_YAMLFormatterMutant family",
      "TestFormatterMutants family",
      "TestFormatterMutantsPlantedFailure"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u151/evidence/S2.overlay.json -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestProduct_YAMLFormatterMutantsOracle|TestProduct_YAMLFormatterMutant000Sources|TestProduct_YAMLFormatterMutant000Lowered|TestProduct_YAMLFormatterMutant000Native|TestProduct_YAMLFormatterMutant001Sources|TestProduct_YAMLFormatterMutant001Lowered|TestProduct_YAMLFormatterMutant001Native|TestProduct_YAMLFormatterMutant002Sources|TestProduct_YAMLFormatterMutant002Lowered|TestProduct_YAMLFormatterMutant002Native|TestProduct_YAMLFormatterMutant003Sources|TestProduct_YAMLFormatterMutant003Lowered|TestProduct_YAMLFormatterMutant003Native|TestProduct_YAMLFormatterMutant004Sources|TestProduct_YAMLFormatterMutant004Lowered|TestProduct_YAMLFormatterMutant004Native|TestProduct_YAMLFormatterMutant005Sources|TestProduct_YAMLFormatterMutant005Lowered|TestProduct_YAMLFormatterMutant005Native|TestFormatterMutants_000|TestFormatterMutants_001|TestFormatterMutants_002|TestFormatterMutants_003|TestFormatterMutants_004|TestFormatterMutants_005|TestFormatterMutantsPlantedFailure)$' > S2.log 2>&1; formatter_mutants_grain_test.go:450: mutant preparation failed",
    "raw_failing_line": "formatter_mutants_grain_test.go:450: mutant preparation failed",
    "members": [
      "TestProduct_YAMLFormatterMutant000Sources",
      "TestProduct_YAMLFormatterMutant000Lowered",
      "TestProduct_YAMLFormatterMutant000Native",
      "TestProduct_YAMLFormatterMutant001Sources",
      "TestProduct_YAMLFormatterMutant001Lowered",
      "TestProduct_YAMLFormatterMutant001Native",
      "TestProduct_YAMLFormatterMutant002Sources",
      "TestProduct_YAMLFormatterMutant002Lowered",
      "TestProduct_YAMLFormatterMutant002Native",
      "TestProduct_YAMLFormatterMutant003Sources",
      "TestProduct_YAMLFormatterMutant003Lowered",
      "TestProduct_YAMLFormatterMutant003Native",
      "TestProduct_YAMLFormatterMutant004Sources",
      "TestProduct_YAMLFormatterMutant004Lowered",
      "TestProduct_YAMLFormatterMutant004Native",
      "TestProduct_YAMLFormatterMutant005Sources",
      "TestProduct_YAMLFormatterMutant005Lowered",
      "TestProduct_YAMLFormatterMutant005Native"
    ],
    "timing_samples": [
      8.077,
      8.407,
      8.797
    ],
    "construction_kills": [
      "S2"
    ],
    "weakened_check": null
  },
  {
    "test": "TestFormatterMutants family",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/formatter_mutants_grain_test.go",
    "seconds": 13.303,
    "oracle": "Go cohere formatting supplies executed expected bytes. Native and source Node run six altered ports; each must differ somewhere. This aggregate inequality oracle accepts any wrong output and does not require the intended changed case.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_YAMLFormatterMutantsOracle",
      "TestProduct_YAMLFormatterMutant family",
      "TestFormatterMutants family",
      "TestFormatterMutantsPlantedFailure"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u151/evidence/W1.overlay.json -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestProduct_YAMLFormatterMutantsOracle|TestProduct_YAMLFormatterMutant000Sources|TestProduct_YAMLFormatterMutant000Lowered|TestProduct_YAMLFormatterMutant000Native|TestProduct_YAMLFormatterMutant001Sources|TestProduct_YAMLFormatterMutant001Lowered|TestProduct_YAMLFormatterMutant001Native|TestProduct_YAMLFormatterMutant002Sources|TestProduct_YAMLFormatterMutant002Lowered|TestProduct_YAMLFormatterMutant002Native|TestProduct_YAMLFormatterMutant003Sources|TestProduct_YAMLFormatterMutant003Lowered|TestProduct_YAMLFormatterMutant003Native|TestProduct_YAMLFormatterMutant004Sources|TestProduct_YAMLFormatterMutant004Lowered|TestProduct_YAMLFormatterMutant004Native|TestProduct_YAMLFormatterMutant005Sources|TestProduct_YAMLFormatterMutant005Lowered|TestProduct_YAMLFormatterMutant005Native|TestFormatterMutants_000|TestFormatterMutants_001|TestFormatterMutants_002|TestFormatterMutants_003|TestFormatterMutants_004|TestFormatterMutants_005|TestFormatterMutantsPlantedFailure)$' > W1.log 2>&1; W1: all six witness members passed with formatterMutantSurvived returning nil; survivor detection disabled",
    "raw_failing_line": "W1: all six witness members passed with formatterMutantSurvived returning nil; survivor detection disabled",
    "members": [
      "TestFormatterMutants_000",
      "TestFormatterMutants_001",
      "TestFormatterMutants_002",
      "TestFormatterMutants_003",
      "TestFormatterMutants_004",
      "TestFormatterMutants_005"
    ],
    "timing_samples": [
      13.211,
      13.303,
      13.423
    ],
    "construction_kills": [],
    "weakened_check": "W1"
  },
  {
    "test": "TestFormatterMutantsPlantedFailure",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/formatter_mutants_grain_test.go",
    "seconds": 0.007,
    "oracle": "Self-written synthetic oracle/wrong bytes require exactly shard 3 to reject the planted survivor. No external formatter executes in this row.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: formatter_mutants_grain_test.go:494: planted survivor caught by [], want [3]",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_YAMLFormatterMutantsOracle",
      "TestProduct_YAMLFormatterMutant family",
      "TestFormatterMutants family",
      "TestFormatterMutantsPlantedFailure"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u151/evidence/W1.overlay.json -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestProduct_YAMLFormatterMutantsOracle|TestProduct_YAMLFormatterMutant000Sources|TestProduct_YAMLFormatterMutant000Lowered|TestProduct_YAMLFormatterMutant000Native|TestProduct_YAMLFormatterMutant001Sources|TestProduct_YAMLFormatterMutant001Lowered|TestProduct_YAMLFormatterMutant001Native|TestProduct_YAMLFormatterMutant002Sources|TestProduct_YAMLFormatterMutant002Lowered|TestProduct_YAMLFormatterMutant002Native|TestProduct_YAMLFormatterMutant003Sources|TestProduct_YAMLFormatterMutant003Lowered|TestProduct_YAMLFormatterMutant003Native|TestProduct_YAMLFormatterMutant004Sources|TestProduct_YAMLFormatterMutant004Lowered|TestProduct_YAMLFormatterMutant004Native|TestProduct_YAMLFormatterMutant005Sources|TestProduct_YAMLFormatterMutant005Lowered|TestProduct_YAMLFormatterMutant005Native|TestFormatterMutants_000|TestFormatterMutants_001|TestFormatterMutants_002|TestFormatterMutants_003|TestFormatterMutants_004|TestFormatterMutants_005|TestFormatterMutantsPlantedFailure)$' > W1.log 2>&1; formatter_mutants_grain_test.go:494: planted survivor caught by [], want [3]",
    "raw_failing_line": "formatter_mutants_grain_test.go:491: planted survivor caught by [], want [3]",
    "members": [
      "TestFormatterMutantsPlantedFailure"
    ],
    "timing_samples": [
      0.008,
      0.007,
      0.007
    ],
    "construction_kills": [],
    "weakened_check": "W1"
  }
]
```

| ID | Origin file:line | Change | Failed rows |
|---|---|---|---|
| W1 | stage1/cohere/yaml/formatter_mutants_grain_test.go:245 | weakened comparison: func formatterMutantSurvived(actual, expected []byte) error { 	return nil } | TestFormatterMutantsPlantedFailure |
| S1 | stage1/cohere/yaml/formatter_mutants_grain_test.go:125 | construction artifact option change: os.WriteFile(filepath.Join(directory, "missing-cases.txt"), data, 0644) | TestFormatterMutants family, TestProduct_YAMLFormatterMutantsOracle |
| S2 | stage1/cohere/yaml/formatter_mutants_grain_test.go:342 | construction artifact option change: os.WriteFile(filepath.Join(directory, file+".missing"), source, 0644) | TestFormatterMutants family, TestProduct_YAMLFormatterMutant family |
| P1 | stage1/cohere/yaml/formatter_mutants_grain_test.go:245 | empty entry probe: return nil at entry | TestFormatterMutantsPlantedFailure |
| P2 | stage1/cohere/yaml/formatter_mutants_grain_test.go:86 | empty entry probe: return "" at entry | TestFormatterMutants family |
| P3 | stage1/cohere/yaml/formatter_mutants_grain_test.go:297 | empty entry probe: return "", "" at entry | TestFormatterMutants family |

Survivors: no production mutants were planted. W1 survives the six execution-witness members, while the planted-survivor test fails; this is a demonstrated weakened-check finding, not equivalent production behavior. P2/P3 surviving product wrappers are empty-answer findings, not mutant survivors.

The brief calls this 21 rows but lists 26 Test functions. All 26 exist at current origin/main a7448d73cd17f16362b6cbc5c5c111080da64e43, in the named file; none moved or vanished. Reading the bodies and applying the family rule produces four rows. Eighteen product wrappers share one checker over mutant index and build level; six execution wrappers share one checker over mutant index. The oracle product and the planted survivor have different assertions and remain separate. Complete members are recorded in members.json.

The requested whole package has 72 tests. Its TestMain builds file-driver products in a separate setup process before m.Run. The outer 120-second timeout terminated the whole-package run after that preflight and later tests; the main binary had not yet emitted its own 90-second timeout. No semantic failure was recorded. This is an over-budget incomplete baseline, not a red assertion baseline. The separate full 26-test slice passed in 82.762 seconds. It overlapped the last part of the first baseline, which can affect cold-build timing, but no source edits existed during either baseline. Mutations began only after the slice was green. Other package rows are unknown. Every verdict and cost here is for the requested slice; no package uniqueness is claimed.

The package tools were warm, but optional YAML/Prettier libraries were not assumed warm. API npm ci ran first. yaml@2.9.0 and prettier@3.9.6 were installed through npm ci in a dedicated directory. Those libraries apply to other package tests, not this slice: these witnesses execute Go cohere and Node and have no library opt-in. The accepted slice did not skip. The whole run was terminated before its complete skip inventory could be obtained. No extra out-of-slice run was made just to obtain that inventory.

These rows are witnesses and construction checks, so the normal production-mutant matrix is inappropriate. Changing a working printer is not evidence that a disagreement witness can detect a disabled comparison. I used only the brief's allowed construction and weakened-check edits. Their kills do not count as production kills, and mutants_in_matrix is zero. The native programs compiled by these tests already include six built-in port mutants, but those are the tests' inputs, not production audit mutants. The optional three-mutants-per-row target therefore does not apply. Standalone diffs, Go vet outputs, exact commands and failures are retained for all six edits and probes.

The witness guard has inverted polarity: it returns an error when the deliberate wrong output equals the oracle, and nil when the wrong output differs. W1 makes it return nil unconditionally, removing detection of a surviving mutant. All six execution witnesses still pass. By the brief's witness rule, their family is untrue under this demonstrated weakening. This does not say their ports are correct or that their normal comparisons are useless. It shows they do not themselves reject removal of this guard. The separate planted-survivor witness does reject it, with caught [] instead of [3]. That row is witness, not sacred; the six other witness outputs cannot establish its production uniqueness.

The first overlay vet command placed -overlay before the Go subcommand and exited 2 without validating source. Its command and exit remain in runs.json. I corrected the order, reused the completed timing runs, and every final overlay validation passed. This command assembly error cost about a minute; it contributes no test evidence.

The execution witnesses compare an aggregate output stream with the aggregate oracle. Any difference is accepted, without requiring the intended mutant's affected case or checking that all unaffected cases match. Their command wrapper does require successful execution and rejects unexpected stderr. The expected bytes come from executed Go cohere; no external-authority label is used. The planted-survivor row uses synthetic bytes, so its oracle is self rather than external-run.

P2 initially failed vet because removing its preparation body left encoding/json unused. The standalone probe also removes that now-unused import. Its final compile check passed; the earlier unused-import failure is not a test kill.

The empty-return probes reveal a separate construction weakness. P2 returns an empty oracle product and its own top-level wrapper passes because it ignores the return. P3 returns two empty product paths and all 18 product wrappers pass for the same reason. These rows are vacuous under their own entries' probes even though construction defects S1 and S2 can make them fail elsewhere. P1 returns nil from the guard, so the six execution witnesses pass and are vacuous under this check-entry probe; the planted survivor fails and is not vacuous. These are probes of the actual check/construction entries, not claims about an empty native formatter's behavior. No production formatter probe was run because production formatting is not the object judged for witness rows.

S1 renames the required cases artifact, preserving Go cohere and its expected output. S2 renames copied source artifacts, so lowering cannot find the expected main.ts; the Go oracle remains untouched. Both use isolated build caches so stale successful construction cannot mask the break. Source-level product wrappers still pass S2, while lower/native wrappers fail, which is retained in the raw matrix. The family verdict rests on those observed failures, not on every member failing. Overlay edits shorten functions and shift diagnostics; reported lines are mapped back to origin with matching unchanged lines, and raw lines remain available.

Evidence was added only after tests because corpus enumeration discovers repository files. The tests' corpus pins, counts and build keys are recorded before the audit Markdown was added. Replay after publication may include additional review files. Build helper durations below are logged misses, may overlap, and are not clang-only times. No full transitive port/runtime coverage was measured, no unrelated packages were run, and no repository-wide uniqueness was attempted.

Setup skipped: warm env.sh worked. API npm ci: 334 ms; library npm ci: 359 ms. Whole baseline hit outer timeout 120 s; slice baseline: 82.762 s. Each of four rows ran alone three times. Recorded subsequent command wall total: 217.185 s. Total elapsed approximately eleven minutes, including reading and analysis. Detailed build product times are in builds.json.

bounded-baseline: binary 82.762 s; sum of logged product misses 277.42 s.

S1: binary 60.582 s; sum of logged product misses 179.61 s.

S2: binary 7.797 s; sum of logged product misses 13.76 s.

Not covered: the other 46 package tests, a completed whole-package baseline/skip inventory, production formatter mutation coverage, full transitive function coverage, repository-wide uniqueness. All standalone diffs apply to the starting source and their Go overlays passed go vet. Production source was left unchanged.
