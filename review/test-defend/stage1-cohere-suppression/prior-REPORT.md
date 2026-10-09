u138 starts at d054e3578d4e53bef8cb8c9ab82d6ace69cdafeb; test list contains two top-level rows.
Clean whole-package baseline passed; no named row skipped.
Port row uniquely killed M1-M3; the gap row caught only M4, also caught by the port row.
Both intended-entry empty probes failed their rows; built-in witnesses failed under W1.
Sources restored; all seven standalone patches apply and compile through their appropriate builders.

```json
[
  {
    "package": "stage1/cohere/suppression",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "subsumed_by": [
      "TestThePortParsesAsGoCohereDoes"
    ],
    "subsumer_seconds": 23.626,
    "mutants_in_matrix": 4,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "test": "TestEachGapStandsWhereGapsMdSaysItDoes",
    "file": "stage1/cohere/suppression/gaps_test.go",
    "seconds": 0.362,
    "oracle": "Node actually runs both closed-gap programs and must match hand-written stdout values; native output is checked against those same values, including exit success and leak checks.",
    "kills": [
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: gaps_test.go:77: natively: exit 0, stdout \"false\\n\", stderr \"\"; Node prints \"true\\n\"",
    "verdict": "subsumed",
    "probe_kills": [
      "P2"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/suppression/ -run . > M4.log 2>&1; gaps_test.go:77: natively: exit 0, stdout \"false\\n\", stderr \"\"; Node prints \"true\\n\""
  },
  {
    "package": "stage1/cohere/suppression",
    "oracle_kind": "external-run",
    "subsumed_by": [],
    "subsumer_seconds": null,
    "mutants_in_matrix": 4,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "test": "TestThePortParsesAsGoCohereDoes",
    "file": "stage1/cohere/suppression/suppression_test.go",
    "seconds": 23.626,
    "oracle": "External Go cohere overlay computes complete answers. Source Node, emitted JavaScript and sanitized native output must agree byte for byte. M1-M3 cause healthy-output disagreements; M4 is caught by native exit 70. Built-in witness failures under M4 are excluded.",
    "kills": [
      "M1",
      "M2",
      "M3",
      "M4"
    ],
    "unique_kills": [
      "M1",
      "M2",
      "M3"
    ],
    "last_proven_fail": "M4 positive subcase: suppression_test.go:66: native: exit 70, stderr \"adamic: panic: undefined where the checker narrowed it away: a call since the narrowing put it back\\n\"",
    "verdict": "sacred",
    "probe_kills": [
      "P1"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/suppression/ -run . > M1.log 2>&1; suppression_test.go:69: Node and Go cohere differ: line 75: \"applied 0\", Go cohere \"applied 30\"",
    "witness_evidence": "W1: suppression_test.go:114: natively the mutant agrees with Go cohere: the comparison cannot see it",
    "witness_result": "positive subcase passes; all 12 builtin mutant subcases fail when firstDifference returns empty"
  }
]
```

| ID | Origin/main file:line | Change | Failed rows |
|---|---|---|---|
| M1 | stage1/cohere/suppression/suppression.ts:204 | candidate.applied++; -> [drop statement] | TestThePortParsesAsGoCohereDoes |
| M2 | stage1/cohere/suppression/directives.ts:27 | export const DirectiveOxlint = 'oxlint-disable'; -> export const DirectiveOxlint = 'oxlint-disabled'; | TestThePortParsesAsGoCohereDoes |
| M3 | stage1/cohere/suppression/scan.ts:122 | while (index < text.length && text[index] !== '\n') -> while (index < text.length && text[index] !== '\r') | TestThePortParsesAsGoCohereDoes |
| M4 | internal/native/emit_expressions.go:226 | == NULL -> != NULL for ir.IsUndefined | TestEachGapStandsWhereGapsMdSaysItDoes, TestThePortParsesAsGoCohereDoes |
| P1 | stage1/cohere/suppression/suppression.ts:248 | replace body with new Index([], [], buildLineIndex(sourceText)) | TestThePortParsesAsGoCohereDoes |
| P2 | internal/lower/lower.go:20 | replace body with return &ir.Program{}, nil; remove unused fmt/path imports | TestEachGapStandsWhereGapsMdSaysItDoes, TestThePortParsesAsGoCohereDoes |
| W1 | stage1/cohere/suppression/suppression_test.go:320 | replace body with return "" | TestThePortParsesAsGoCohereDoes |

P1/P2 are probes, not production mutants. P2 also fails the port comparison, but Lower is preparation for its port-under-test entry, so only P1 determines that row’s vacuity. W1 is the permitted witness edit and is excluded from production kill sets. All production mutants are caught; there are no survivors or equivalent candidates in this menu.

Brief ambiguities, limits, and costs:

- The package has two actual top-level functions; neither is a subprocess helper. The gap function has two table subcases and the port function has one positive agreement subcase plus twelve built-in mutant witnesses. These stay two rows because the brief defines a top-level function as a row. No names vanished or moved relative to the starting registry.
- The port function is a mixed positive/witness row, which the verdict vocabulary does not directly represent. Calling the entire function a witness would discard its real Go-cohere comparison. Its sacred verdict rests on positive-subcase failures under M1-M3. W1 separately proves its twelve witness subcases, while that row’s positive subcase passes with the comparator disabled.
- “For stage1 ports mutate only the port” does not cover the gap function’s real code under test: it compiles small programs to exercise Adamic’s compiler. Changing those fixture programs would change the oracle input rather than break the compiler. The fourth frozen mutation therefore changes native undefined-comparison emission. Its Go package passed vet, a gap product linked, and it used a dedicated ADAMIC_BUILD_CACHE_DIR. No Go cohere or Node implementation was mutated.
- The general three-mutants-per-row target conflicts with the four-rebuild exception for a stage1 port without a selector. This audit uses four production mutants total: three across port accounting, directive grammar, and comment scanning; one generic compiler comparison flip. The gap subsumption finding rests on exactly one caught mutant, not a broad overlap demonstration.
- The gap row is far cheaper: 0.362 seconds versus the port row’s 23.626 seconds. The brief’s current subsumption rule imposes no speed restriction. It therefore labels the gap row subsumed and records the slower subsumer’s median. M4 is observed differently too: the gap produces a healthy wrong answer, while the port rejects a native runtime panic. This kill-set inclusion is a one-mutant hint, not grounds to delete the gap row.
- Native exit 70 is an execution failure, not a Go test-binary panic. The M4 package matrix ran both top-level rows, but the runner conservatively reran each alone because the captured stderr contained “panic:”. Both failures were observed again. Built-in witness failures caused by that runtime failure are excluded from evidence that a witness works.
- “Empty value” is not directly nil for the typed Index API. P1 returns a valid Index with empty directives/references and retains the source’s line index. P2 returns an empty IR Program so the emitter can build a healthy no-output executable instead of panicking on nil. Both eliminate their entries’ substantive answer and produce concrete comparison failures. No verdict or uniqueness rests on probes.
- The first standalone CLI link commands failed because the audit did not create /tmp/u138 before requesting output files there. This was an audit setup error, not a source mutant failure. The package matrices independently compiled and executed native products. After creating the directory, all three port standalone builds passed; the initial failed logs are retained and excluded from valid rebuild timings.
- Warm toolchain verification succeeded and setup was skipped. npm ci ran in stage3/api. The package tests do not load ESLint’s node_modules: eslint_whitespace.cjs is a documented manual probe, not part of either test’s oracle. No external-authority or ESLint-run claim is made from that file.
- The port function declaration inventory was saved before mutation results. A broad compiler declaration inventory was expanded after the port mutations had begun, before M4; it is a conservative list rather than an exact pre-run reach proof. Restored-source Go coverage and Node V8 coverage subsequently record the reached compiler and port functions. No mutation was selected or changed from that later coverage. Anonymous callbacks and module initialization are included in those observed inventories.
- The whole package fit the budget, so every production mutant ran both rows, and bounded is false. Neither other packages nor a repo-wide uniqueness replay ran. Standalone patches permit that central replay. The original 2,500 generated cases and fixed corpus sources remain enabled; no corpus size was reduced for timing or mutation runs.

Timing:

```json
{
  "nproc": 5,
  "setup_skipped": true,
  "tools_seconds": 0.02976606999800424,
  "npm_ci_seconds": 0.628910082999937,
  "test_list_wall_seconds": 5.581958622999082,
  "baseline_binary_seconds": 23.979,
  "baseline_wall_seconds": 26.273780945000908,
  "medians": {
    "TestEachGapStandsWhereGapsMdSaysItDoes": 0.362,
    "TestThePortParsesAsGoCohereDoes": 23.626
  },
  "standalone_native_rebuild_wall_seconds": {
    "M1": 1.7843902509994223,
    "M2": 1.3939972440020938,
    "M3": 1.4678164230026596,
    "M4": 6.425311250000959
  },
  "runner_wall_seconds": 311.0778573259995,
  "post_validation_and_coverage_wall_seconds": 29.246260924002854,
  "logged_runner_binary_seconds": 241.43800000000002,
  "final_clean_binary_seconds": 21.28,
  "generated_utc": "2026-10-09T13:04:44.340972+00:00"
}
```

Native CLI rebuild timings include Go command startup, checking, lowering, C emission and sanitized linking; phase-only times are not separately claimed. The package matrices also rebuild its twelve built-in witness products. Each median is from three -count=1 isolated runs using the test binary’s own ok line. Coverage and extra validation timings are excluded from those medians.

No production survivors. No cooked step. No tests skipped. Sources restored and final whole-package baseline passed. Uncovered: other packages, other generated seeds, larger real-source corpus, exhaustive compiler defect classes, and repo-wide uniqueness. Go compiler function coverage and V8 function counts demonstrate reach, not exhaustive branch coverage.
