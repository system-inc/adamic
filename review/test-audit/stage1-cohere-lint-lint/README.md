u110: all 35 supplied test names exist at 1f34d0d300301faebc94d397adee1a090923d2c7.
Eleven grouped rows: six witness, one bounded sacred, one untrue, three cannot-judge.
Warm tools verified; nproc 5; npm ci 0.467 seconds.
The node-table family passes its own empty-entry probe; M3 survives the production matrix.
Sources restored after green bounded controls; evidence is under this directory.

```json
[
  {
    "test": "TestLegacyMutants",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/lint_test.go",
    "seconds": 7.695,
    "oracle": "Go cohere findings; negative mutant-agreement assertion",
    "oracle_kind": "external-run",
    "kills": [
      "W1"
    ],
    "unique_kills": [],
    "last_proven_fail": "W1: lint_test.go:551: overlap winner misreported mutant survived on Node",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 6,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLegacyMutants",
      "TestDecorationOptionMutant",
      "TestCountGuardMutant",
      "TestThroughput",
      "TestNodeTableIsLinkOnlyUnionAndPlantedFailure",
      "TestOwnedWitnessesAssignmentStable",
      "TestOwnedWitnessesPlantedDisagreement"
    ],
    "evidence": "ADAMIC_U110_WITNESS=W1; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^(TestLegacyMutants|TestDecorationOptionMutant|TestCountGuardMutant|TestThroughput|TestNodeTableIsLinkOnlyUnionAndPlantedFailure|TestOwnedWitnessesAssignmentStable|TestOwnedWitnessesPlantedDisagreement)$ > W1.log 2>&1; lint_test.go:551: overlap winner misreported mutant survived on Node"
  },
  {
    "test": "TestDecorationOptionMutant",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/lint_test.go",
    "seconds": 8.143,
    "oracle": "Go cohere findings; negative decoration-mutant agreement assertion",
    "oracle_kind": "external-run",
    "kills": [
      "W2"
    ],
    "unique_kills": [],
    "last_proven_fail": "W2: lint_test.go:572: decoration range mutant survived on Node",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 6,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLegacyMutants",
      "TestDecorationOptionMutant",
      "TestCountGuardMutant",
      "TestThroughput",
      "TestNodeTableIsLinkOnlyUnionAndPlantedFailure",
      "TestOwnedWitnessesAssignmentStable",
      "TestOwnedWitnessesPlantedDisagreement"
    ],
    "evidence": "ADAMIC_U110_WITNESS=W2; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^(TestLegacyMutants|TestDecorationOptionMutant|TestCountGuardMutant|TestThroughput|TestNodeTableIsLinkOnlyUnionAndPlantedFailure|TestOwnedWitnessesAssignmentStable|TestOwnedWitnessesPlantedDisagreement)$ > W2.log 2>&1; lint_test.go:572: decoration range mutant survived on Node"
  },
  {
    "test": "TestCountGuardMutant",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/lint_test.go",
    "seconds": 12.711,
    "oracle": "Go cohere ordinary output and count; negative count-only agreement assertion",
    "oracle_kind": "external-run",
    "kills": [
      "W3"
    ],
    "unique_kills": [],
    "last_proven_fail": "W3: lint_test.go:600: Node count mutant survived",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 6,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLegacyMutants",
      "TestDecorationOptionMutant",
      "TestCountGuardMutant",
      "TestThroughput",
      "TestNodeTableIsLinkOnlyUnionAndPlantedFailure",
      "TestOwnedWitnessesAssignmentStable",
      "TestOwnedWitnessesPlantedDisagreement"
    ],
    "evidence": "ADAMIC_U110_WITNESS=W3; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^(TestLegacyMutants|TestDecorationOptionMutant|TestCountGuardMutant|TestThroughput|TestNodeTableIsLinkOnlyUnionAndPlantedFailure|TestOwnedWitnessesAssignmentStable|TestOwnedWitnessesPlantedDisagreement)$ > W3.log 2>&1; lint_test.go:600: Node count mutant survived"
  },
  {
    "test": "TestThroughput",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/lint_test.go",
    "seconds": 0.023,
    "oracle": "Go cohere count, when benchmark enabled; all session timing runs skipped",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestThroughput",
      "TestNodeTableIsLinkOnly",
      "TestNodeTableIsLinkOnlyFamily",
      "TestOwnedWitnessesAssignmentStable",
      "TestOwnedWitnesses"
    ],
    "evidence": "No proven failure; see timing and bounded matrix logs."
  },
  {
    "test": "TestMutants",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/lint_test.go",
    "seconds": null,
    "oracle": "Go cohere findings; negative registered-rule mutant agreement assertions",
    "oracle_kind": "external-run",
    "kills": [
      "W4"
    ],
    "unique_kills": [],
    "last_proven_fail": "W4: lint_test.go:745: nexus-abbreviated-identifier whole-word finding suppressed mutant survived on Node",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 6,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLegacyMutants",
      "TestDecorationOptionMutant",
      "TestCountGuardMutant",
      "TestThroughput",
      "TestNodeTableIsLinkOnlyUnionAndPlantedFailure",
      "TestOwnedWitnessesAssignmentStable",
      "TestOwnedWitnessesPlantedDisagreement",
      "TestMutants"
    ],
    "evidence": "ADAMIC_U110_WITNESS=W4; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestMutants$/^nexus\\-abbreviated\\-identifier_whole\\-word_finding_suppressed$ > W4-Mutants-bounded.log 2>&1; lint_test.go:745: nexus-abbreviated-identifier whole-word finding suppressed mutant survived on Node"
  },
  {
    "test": "TestNodeTableIsLinkOnly",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/node_table_split_test.go",
    "seconds": 1.475,
    "oracle": "Self-authored corpus census and hash-shard union",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestThroughput",
      "TestNodeTableIsLinkOnly",
      "TestNodeTableIsLinkOnlyFamily",
      "TestOwnedWitnessesAssignmentStable",
      "TestOwnedWitnesses"
    ],
    "evidence": "No proven failure; see timing and bounded matrix logs."
  },
  {
    "test": "TestNodeTableIsLinkOnlyFamily",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/node_table_split_test.go",
    "seconds": 3.146,
    "oracle": "Same native port with plain versus junk-row option, no outside authority",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestThroughput",
      "TestNodeTableIsLinkOnly",
      "TestNodeTableIsLinkOnlyFamily",
      "TestOwnedWitnessesAssignmentStable",
      "TestOwnedWitnesses"
    ],
    "evidence": "No proven failure; see timing and bounded matrix logs."
  },
  {
    "test": "TestNodeTableIsLinkOnlyUnionAndPlantedFailure",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/node_table_split_test.go",
    "seconds": 0.232,
    "oracle": "Self-authored synthetic union corruption and disagreement",
    "oracle_kind": "self",
    "kills": [
      "W5"
    ],
    "unique_kills": [],
    "last_proven_fail": "W5: node_table_split_test.go:383: planted failure survived",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 6,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLegacyMutants",
      "TestDecorationOptionMutant",
      "TestCountGuardMutant",
      "TestThroughput",
      "TestNodeTableIsLinkOnlyUnionAndPlantedFailure",
      "TestOwnedWitnessesAssignmentStable",
      "TestOwnedWitnessesPlantedDisagreement"
    ],
    "evidence": "ADAMIC_U110_WITNESS=W5; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^(TestLegacyMutants|TestDecorationOptionMutant|TestCountGuardMutant|TestThroughput|TestNodeTableIsLinkOnlyUnionAndPlantedFailure|TestOwnedWitnessesAssignmentStable|TestOwnedWitnessesPlantedDisagreement)$ > W5.log 2>&1; node_table_split_test.go:383: planted failure survived"
  },
  {
    "test": "TestOwnedWitnessesAssignmentStable",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/owned_witness_units_test.go",
    "seconds": 0.041,
    "oracle": "Self-authored stable hash assignment and empty corpus assertion",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestThroughput",
      "TestNodeTableIsLinkOnly",
      "TestNodeTableIsLinkOnlyFamily",
      "TestOwnedWitnessesAssignmentStable",
      "TestOwnedWitnesses"
    ],
    "evidence": "No proven failure; see timing and bounded matrix logs."
  },
  {
    "test": "TestOwnedWitnesses",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/owned_witness_units_test.go",
    "seconds": 7.231,
    "oracle": "Go cohere findings against Node, emitted JavaScript and native; self census checks",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M2"
    ],
    "unique_kills": [
      "M1",
      "M2"
    ],
    "last_proven_fail": "M2: owned_witness_units_test.go:509: shard-010 Node: case 8 line 50: port \"/tmp/adamic-gate/TestOwnedWitnesses_010734240652/067/nexus-consistency-require-type-suffix-0.ts:8:7\", Go \"/tmp/adamic-gate/TestOwnedWitnesses_010734240652/067/nexus-consistency-require-type-suffix-0.ts:1:11\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestThroughput",
      "TestNodeTableIsLinkOnly",
      "TestNodeTableIsLinkOnlyFamily",
      "TestOwnedWitnessesAssignmentStable",
      "TestOwnedWitnesses"
    ],
    "evidence": "selector=M2; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^(TestThroughput|TestNodeTableIsLinkOnly|TestNodeTableIsLinkOnly_[0-9]{3}|TestOwnedWitnessesAssignmentStable|TestOwnedWitnesses(_Setup|Union|_[0-9]{3}))$ > M2.log 2>&1; owned_witness_units_test.go:509: shard-010 Node: case 8 line 50: port \"/tmp/adamic-gate/TestOwnedWitnesses_010734240652/067/nexus-consistency-require-type-suffix-0.ts:8:7\", Go \"/tmp/adamic-gate/TestOwnedWitnesses_010734240652/067/nexus-consistency-require-type-suffix-0.ts:1:11\""
  },
  {
    "test": "TestOwnedWitnessesPlantedDisagreement",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/owned_witness_units_test.go",
    "seconds": 0.216,
    "oracle": "Self-authored synthetic native disagreement and exactly-one-shard failure",
    "oracle_kind": "self",
    "kills": [
      "W6"
    ],
    "unique_kills": [],
    "last_proven_fail": "W6: owned_witness_units_test.go:448: planted disagreement survived",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 6,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLegacyMutants",
      "TestDecorationOptionMutant",
      "TestCountGuardMutant",
      "TestThroughput",
      "TestNodeTableIsLinkOnlyUnionAndPlantedFailure",
      "TestOwnedWitnessesAssignmentStable",
      "TestOwnedWitnessesPlantedDisagreement"
    ],
    "evidence": "ADAMIC_U110_WITNESS=W6; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^(TestLegacyMutants|TestDecorationOptionMutant|TestCountGuardMutant|TestThroughput|TestNodeTableIsLinkOnlyUnionAndPlantedFailure|TestOwnedWitnessesAssignmentStable|TestOwnedWitnessesPlantedDisagreement)$ > W6.log 2>&1; owned_witness_units_test.go:448: planted disagreement survived"
  }
]
```

All source file positions refer to the starting origin/main commit. Family labels in rows.json represent listed shard members, not additional top-level functions. Production matrix rows: TestThroughput (skipped), TestNodeTableIsLinkOnly (corpus setup), TestNodeTableIsLinkOnly_000 through _007 as one family, TestOwnedWitnessesAssignmentStable, and TestOwnedWitnesses_Setup/Union/_000 through _015 as one family. Witness matrix rows: LegacyMutants, DecorationOptionMutant, CountGuardMutant, Throughput (skipped), NodeTableIsLinkOnlyUnionAndPlantedFailure, OwnedWitnessesAssignmentStable, OwnedWitnessesPlantedDisagreement. TestMutants is an additional isolated, bounded witness run.

| ID | File:line | Change | Failed rows |
|---|---|---|---|
| M1 | stage1/cohere/lint/lint.ts:116 | `rules.visit(index, parent);` to `(drop statement)` | TestOwnedWitnesses |
| M2 | stage1/cohere/lint/lint.ts:21 | `return left.start - right.start;` to `return right.start - left.start;` | TestOwnedWitnesses |
| M3 | stage1/cohere/lint/lint.ts:77 | `if(this.junkRows) {` to `if(!this.junkRows) {` |  |
| E1 | stage1/cohere/lint/main.ts:11 | `function run(row: string, countOnly: boolean): number {` to `function run(row: string, countOnly: boolean): number { return 0;` | TestOwnedWitnesses |
| W1 | stage1/cohere/lint/lint_test.go:550 | `if bytes.Equal(side.run.output, want) {` to `if true {` | TestLegacyMutants |
| W2 | stage1/cohere/lint/lint_test.go:571 | `if bytes.Equal(side.run.output, want) {` to `if true {` | TestDecorationOptionMutant |
| W3 | stage1/cohere/lint/lint_test.go:599 | `if bytes.Equal(side.count.output, count) {` to `if true {` | TestCountGuardMutant |
| W4 | stage1/cohere/lint/lint_test.go:744 | `if bytes.Equal(side.run.output, want) {` to `if true {` | TestMutants |
| W5 | stage1/cohere/lint/node_table_split_test.go:293 | `if diff := difference(got, want); diff != "" {` to `if diff := ""; diff != "" {` | TestNodeTableIsLinkOnlyUnionAndPlantedFailure |
| W6 | stage1/cohere/lint/owned_witness_units_test.go:327 | `if diff := difference(got, want); diff != "" {` to `if diff := ""; diff != "" {` | TestOwnedWitnessesPlantedDisagreement |

M3 survivor: reversing the junk-row guard changed a direct port-source probe from `nodes=3 findings=1` to `nodes=6 findings=1`, while every executed production-matrix row passed or skipped. This is changed internal table behavior, rather than an equivalent candidate. It exposes a gap in checking that the junk-row option actually performs the perturbation. Ordinary findings intentionally remain invariant to unattached rows. Exact commands and outputs are in survivor.json and survivor-before.log/survivor-M3.log.

The code under test and oracle were named before planting: native lint port Linter, RuleContext, settings/comment helpers, registered rules and main.ts entry; Go cohere is the agreement oracle. The node-table family instead compares the same native program with plain and junk-row arguments, a self oracle. Negative comparison witnesses were judged solely by W1 through W6, not production-mutant precondition failures. Production plan is the fixed drop-statement, swap-arguments and condition-flip menu, plus E1. Function-candidates.txt lists 717 static port-function candidates collected before planting. It is not a dynamic coverage proof: exact function reachability across this large port was not measured, and some candidates may be unreached. Mutated functions were selected from that list.

No supplied test moved or vanished. The brief says eleven rows but its general grouping rule could also put the node-table setup inside its shard family. To retain eleven, corpus-only TestNodeTableIsLinkOnly remains separate from the native eight-shard family. The owned Setup, Union and sixteen shards are one family; the planted-disagreement witness remains separate. The node-table union/planted-failure test is a witness row because it deliberately checks a comparison failure as well as corrupt unions. Its W5 proves the disagreement part, not every census assertion.

The whole-package clean baseline was cooked at 90.004 wall seconds, without any recorded failed test. A narrowed slice baseline was also cooked at 90.005, after passing LegacyMutants and node-table corpus setup. The next, seven-row bounded baseline passed in 20.317 wall / 18.021 binary seconds. Native families were separately established green after caching completed cold build stages. This was narrowing after over-budget runs, not auditing on a red baseline. Package-wide and repository-wide uniqueness remain unknown. Sacred and untrue are explicitly bounded conclusions over the production rows listed above. Three production mutants were used, below the suggested approximately-three-per-row target, because cold product preparation consumed the budget; no claim of broad mutation adequacy follows from this small set.

Cold binary preparation matters: initial node-table timing reached C emission 72.88 seconds and native build 9.01 seconds after 5.02 seconds of lowering, then the outer wall cutoff stopped completion. Initial owned-witness preparation logged 81.83 seconds of lowering/backends, then exceeded the wall budget. Warm median native-family times are therefore reported separately from the cold costs. Switched-source controls likewise first exceeded 90 seconds, then passed using the completed cached stages. No cooked run is a mutant kill. Every step was stopped by the 90-second wall controller; the overall session exceeded the approximate 20-minute budget and took about 28 minutes.

Full TestMutants timing exceeded 90 seconds, so no three-run median was produced. Repeating an unchanged cooked row twice more would add three minutes without giving a completed median. Its full weakened-check run with -failfast also exceeded 90 seconds because already queued parallel children kept running. W4-Mutants-bounded.log is the completed rerun of one failing leaf using the shared comparison: nexus-abbreviated-identifier_whole-word_finding_suppressed. W4 proves that shared comparison can be rejected; later full-row guards and native-canary behavior are unknown. W4-stop.json records that the attempted earlier stop found no remaining process after the wall cutoff.

TestThroughput skipped all three measurements because ADAMIC_LINT_BENCH was unset. Its 0.023 seconds is the skip-path package time, not benchmark execution cost. The available cohere/TypeScript checkout is d92d9bfee114c80be2c375d72edae966176e3a4f, rather than the benchmark's required 050880ce59e30b356b686bd3144efe24f875ebc8; it was not silently substituted. The throughput verdict is cannot-judge.

TestNodeTableIsLinkOnly executes corpus construction and union checks, without executing the native lint entry. TestOwnedWitnessesAssignmentStable checks Go test-helper planning and hashing. Meaningful mutants for those rows would edit the harness, which the brief forbids except for witnesses, or the Go oracle corpus. Neither was done; both are cannot-judge. Their own empty-entry probes were not performed. The brief's boolean vacuous schema cannot encode an unobserved relevant probe, so these rows and the skipped throughput row use null rather than a fabricated false. E1 passed all eight node-table members and failed owned witnesses. The node-table row is vacuous. Each witness failed when its own comparison was disabled, so its vacuous flag is false.

All ten standalone diffs apply independently to the starting commit. Each Go witness diff passed go vet with a source overlay; vet-results.json records commands and durations. go vet does not type-check .ts or .a source. The native switch controls compiled the production branches, including the empty-entry branch, but separate unconditional native rebuilds for each standalone port diff were not performed. A single Go vet passed the package containing the three port switches and E1; it is a Go compilation check, not extra TypeScript proof. No unused Go variable was introduced: W5/W6 change the comparison initializer instead of inserting an unconditional return.

No production mutant panicked. E1 produced a completed bounded matrix. No package-wide kill matrix was affordable after the clean package exceeded 90 seconds. Kills outside the production and witness sets are unknown. No other package tests were run. No external-authority values were claimed or checked. Node was the port execution side, while Go cohere supplied the lint oracle. Full trace logs, exact top-level/member pass/fail/skip lists, fixed plans, standalone diffs, switch reconstruction, vet overlays and survivor probes are retained.

Measured costs: npm ci 0.467 seconds; warm setup skipped. Initial whole-package baseline 90.004 seconds. Recorded runs after that total 1084.711 wall seconds, including 479.982 seconds for timing attempts, 44.142 for production plus E1, and 228.060 for witness runs including the cooked full W4. Logged successful product build misses total 446.410 seconds and are included in run costs. Per-product cold/rebuild timings are in build-times.json. These totals overlap by purpose and must not be added together. Source inspection, report preparation and tool-call latency are outside the recorded test totals.

Every production/test edit and the temporary audit_u110.ts helper was removed after the green final bounded control. Only review evidence is committed. No push to main and no pull request.
