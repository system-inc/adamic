u116: all 11 live tests audited as four rows at ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2.
Clean baseline passed in 59.762s; nproc=5; no baseline skips.
Verdicts: one sacred family, two setup checks, one witness.
Six primary mutants, one supplemental mutant and seven entry probes were caught; no survivors.
Evidence: test-audit/stage1-cohere-lint-inventory, review/test-audit/stage1-cohere-lint-inventory/.

```json
[
  {
    "test": "TestInventoryEngine family",
    "package": "stage1/cohere/lint/inventory",
    "file": "stage1/cohere/lint/inventory/engine_shards_test.go",
    "seconds": 2.598,
    "oracle": "Self-written dependency, ranking, string, path and denominator assertions. TestRegisteredCorpusControl executes Go cohere no-debugger, then checks self-written counts, not diagnostic identity; a different one-diagnostic failure could satisfy this count-only control. No outside authority was copied or checked.",
    "oracle_kind": [
      "self",
      "external-run"
    ],
    "kills": [
      "M1",
      "M2",
      "M3",
      "M4",
      "M5",
      "M7"
    ],
    "unique_kills": [
      "M1",
      "M2",
      "M3",
      "M4",
      "M5",
      "M7"
    ],
    "last_proven_fail": "M7: engine_shards_test.go:289: case TestTransitiveSiblingAndMethodDependencies: exit status 1",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 6,
    "probe_kills": [
      "P1",
      "P2",
      "P3",
      "P4",
      "P5",
      "P6",
      "P7"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestInventoryEngine family",
      "TestInventoryEngineUnion",
      "TestInventoryEngineShardAssignment",
      "TestInventoryEngineShardMutant"
    ],
    "members": [
      "TestInventoryEngine_000",
      "TestInventoryEngine_001",
      "TestInventoryEngine_002",
      "TestInventoryEngine_003",
      "TestInventoryEngine_004",
      "TestInventoryEngine_005",
      "TestInventoryEngine_006",
      "TestInventoryEngine_007"
    ],
    "supplemental_kills": [
      "M6"
    ],
    "vacuous_subcases": [
      "TestInventoryEngine_002 (empty shard)",
      "TestInventoryEngine_007 (empty shard)",
      "TestInventoryEngine_005 (empty shard)",
      "TestInventoryEngine_006 (empty shard)"
    ],
    "entry_probes": {
      "trace": "P1",
      "rankings": "P2",
      "countText": "P3",
      "hasSelector": "P4",
      "familyPassed": "P5",
      "measured": "P6",
      "measureCorpus": "P7"
    },
    "evidence": "ADAMIC_MUTANT=M7 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/inventory/ -run . > M7.log 2>&1; engine_shards_test.go:289: case TestTransitiveSiblingAndMethodDependencies: exit status 1"
  },
  {
    "test": "TestInventoryEngineUnion",
    "package": "stage1/cohere/lint/inventory",
    "file": "stage1/cohere/lint/inventory/main_test.go",
    "seconds": 2.5869999999999997,
    "oracle": "Self-written eight-entry top-level table and exact live shard-union checks.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S2: main_test.go:10: enumerated 7 top-level shards, want 8",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 6,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": false,
    "matrix_rows": [
      "TestInventoryEngine family",
      "TestInventoryEngineUnion",
      "TestInventoryEngineShardAssignment",
      "TestInventoryEngineShardMutant"
    ],
    "evidence": " timeout 120 go test -json -count=1 -timeout 90s -overlay=/tmp/u116/evidence/S2.overlay.json ./stage1/cohere/lint/inventory/ -run . > S2.log 2>&1; main_test.go:10: enumerated 7 top-level shards, want 8",
    "construction_kills": [
      "S2"
    ]
  },
  {
    "test": "TestInventoryEngineShardAssignment",
    "package": "stage1/cohere/lint/inventory",
    "file": "stage1/cohere/lint/inventory/engine_shards_test.go",
    "seconds": 0.002,
    "oracle": "Self-written stable hash assignment, growth, empty/repeated-corpus rejection checks.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S1: engine_shards_test.go:162: repeated case TestExistingA",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 6,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": false,
    "matrix_rows": [
      "TestInventoryEngine family",
      "TestInventoryEngineUnion",
      "TestInventoryEngineShardAssignment",
      "TestInventoryEngineShardMutant"
    ],
    "evidence": " timeout 120 go test -json -count=1 -timeout 90s -overlay=/tmp/u116/evidence/S1.overlay.json ./stage1/cohere/lint/inventory/ -run . > S1.log 2>&1; engine_shards_test.go:162: repeated case TestExistingA",
    "construction_kills": [
      "S1"
    ]
  },
  {
    "test": "TestInventoryEngineShardMutant",
    "package": "stage1/cohere/lint/inventory",
    "file": "stage1/cohere/lint/inventory/engine_shards_test.go",
    "seconds": 2.57,
    "oracle": "Self-written expected singleton failure at TestUnknownFrequencyIsNotZero, with matching failure text and shard. W1 masks the subprocess error, so the expected catch disappears.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: engine_shards_test.go:238 on origin/main (overlay line 239): unknown-frequency mutant caught by [], want exactly shard-003",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 6,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": false,
    "matrix_rows": [
      "TestInventoryEngine family",
      "TestInventoryEngineUnion",
      "TestInventoryEngineShardAssignment",
      "TestInventoryEngineShardMutant"
    ],
    "evidence": " timeout 120 go test -json -count=1 -timeout 90s -overlay=/tmp/u116/evidence/W1.overlay.json ./stage1/cohere/lint/inventory/ -run . > W1.log 2>&1; engine_shards_test.go:239: unknown-frequency mutant caught by [], want exactly shard-003; overlay line 239 maps to origin/main line 238",
    "weakened_check_kills": [
      "W1"
    ]
  }
]
```

| ID | Origin file:line | Change | Rows failed |
|---|---|---|---|
| M1 (production) | stage1/cohere/lint/inventory/testdata/engine.go:379 | `Direct: depth != 0` | TestInventoryEngine family, TestInventoryEngineShardMutant |
| M2 (production) | stage1/cohere/lint/inventory/testdata/engine.go:848 | `Count: len(names) + 1` | TestInventoryEngine family, TestInventoryEngineShardMutant |
| M3 (production) | stage1/cohere/lint/inventory/testdata/engine.go:745 | `*report.Count += 2` | TestInventoryEngine family, TestInventoryEngineShardMutant |
| M4 (production) | stage1/cohere/lint/inventory/testdata/engine.go:513 | `(?:disabled\(` | TestInventoryEngine family, TestInventoryEngineShardMutant |
| M5 (production) | stage1/cohere/lint/inventory/testdata/engine.go:701 | `if !strings.HasPrefix(line, "ok  \t"+rulesPrefix+family+"\t") {` | TestInventoryEngine family, TestInventoryEngineShardMutant |
| M6 (production) | stage1/cohere/lint/inventory/testdata/engine.go:736 | `+"; "+fmt.Sprint(value)` | TestInventoryEngine family, TestInventoryEngineShardMutant |
| M7 (production) | stage1/cohere/lint/inventory/testdata/engine.go:384 | `if false && next != nil { 				walk(next, depth+1)` | TestInventoryEngine family, TestInventoryEngineShardMutant |
| P1 (probe) | stage1/cohere/lint/inventory/testdata/engine.go:333 | `return nil, nil, false, false at entry (conditional probe)` | TestInventoryEngine family, TestInventoryEngineShardMutant |
| P2 (probe) | stage1/cohere/lint/inventory/testdata/engine.go:831 | `return nil at entry (conditional probe)` | TestInventoryEngine family, TestInventoryEngineShardMutant |
| P3 (probe) | stage1/cohere/lint/inventory/testdata/engine.go:863 | `return "" at entry (conditional probe)` | TestInventoryEngine family, TestInventoryEngineShardMutant |
| P4 (probe) | stage1/cohere/lint/inventory/testdata/engine.go:512 | `return false at entry (conditional probe)` | TestInventoryEngine family, TestInventoryEngineShardMutant |
| P5 (probe) | stage1/cohere/lint/inventory/testdata/engine.go:699 | `return false at entry (conditional probe)` | TestInventoryEngine family, TestInventoryEngineShardMutant |
| P6 (probe) | stage1/cohere/lint/inventory/testdata/engine.go:732 | `return at entry (conditional probe)` | TestInventoryEngine family, TestInventoryEngineShardMutant |
| P7 (probe) | stage1/cohere/lint/inventory/testdata/engine.go:770 | `return nil at entry (conditional probe)` | TestInventoryEngine family, TestInventoryEngineShardMutant |
| S1 (setup) | stage1/cohere/lint/inventory/engine_shards_test.go:105 | ` 		if !seen[name] {` | TestInventoryEngine family, TestInventoryEngineShardAssignment, TestInventoryEngineShardMutant, TestInventoryEngineUnion |
| S2 (setup) | stage1/cohere/lint/inventory/engine_shards_test.go:306 | `` | TestInventoryEngineUnion |
| W1 (witness) | stage1/cohere/lint/inventory/engine_shards_test.go:198 | `output, _ := command.CombinedOutput() 	return output, nil` | TestInventoryEngineShardMutant |

M6 is supplemental and excluded from kills and uniqueness. Production witness failures in this table are broken preconditions and are excluded from kills and uniqueness. Only W1 decides the witness verdict. Setup edits and probes are separate evidence, not production kills. Uniqueness is package-local over four grouped rows; repository-wide replay remains central.

Survivors: none among the six primary production mutants or the supplemental M6. No equivalent-candidate claim was needed.

The brief caused the following friction and limits:

- This stage1 package tests a Go inventory generator, not an Adamic native port. Its implementation lives under testdata and is compiled as a virtual file inside cohere. A plain go vet of the outer package would miss it; each standalone engine diff was vetted through its real cohere overlay boundary. No native rebuild or compiler cache isolation was needed.
- The live list contains eight shard wrappers but only eight inner controls. Four shards are empty and pass every probe. Grouping them as one family avoids crediting empty wrappers as independent worthy tests; the empty members are recorded as vacuous subcases. The union has an extra top-level-table assertion, so it stays separate.
- There is no single semantic entry for the family. Seven entry probes were run, one per reached entry. Every probe fails the family; setup and witness vacuity remain null because those rows do not assert a production answer.
- The countText entry probe breaks the built-in witness literal anchor. A preliminary switch made that precondition fail even when the probe was off. Those preliminary logs are retained but not used for verdicts. The production matrix was replayed with the anchor restored, after a green switch-clean run. Only P3 retains the anchor-precondition failure, which is excluded from witness evidence.
- The preliminary path-format mutant replaced a path with its basename, which was not a literal member of the fixed menu. It was replaced before the accepted matrix with M6, changing the path separator constant from colon-space to semicolon-space. M6 was selected after the preliminary results, so it is supplemental and no verdict rests on either path-format version.
- The construction runner initially found two occurrences of the duplicate guard. It stopped before any construction edit; S1 was then restricted to the partition function by its exact indentation.
- Production mutants frequently also fail the built-in witness through an unexpected failing control. Those failures do not prove the witness comparison works. Returning nil errors in W1 independently makes its expected caught list empty and proves the witness can fail.
- The actual registered no-debugger control checks counts, not diagnostic identity. The rest of the engine expected answers are self-written. Executing a real external rule does not independently establish the complete inventory result.
- Several generator paths are not reached by this package: load, entries, inspectBranches, statusFor, captureTests, restoreCapture, collectFiles, markdown, either main, and outer main.go run/must. They were read but not mutated. This audit judges the present tests, not complete generator coverage.
- The whole-package cold baseline fit the 90s budget, so every accepted mutant ran the complete outer package. No narrowing, timeout, outer panic or unknown row outcome occurred. Native stage1 builds and other packages were not run.
- The request refers to warm setup but requires npm ci even for this Go-only package. The install succeeded; no additional node_modules directories are loaded by these tests. The README describes .a checker counts as unknown, while the live inner corpus control requires a completed measurement; this audit used observed current behavior rather than that prose.

Setup: warm env.sh, no cloud setup; npm ci reported 429ms. Baseline: 59.762s in the test binary, including the cold overlay build. Timed rows used three independent -count=1 package invocations each: medians family 2.598s, union 2.587s, assignment 0.002s, witness 2.570s. Recorded subsequent commands consumed 174.182s of wall time including builds, accepted and preliminary matrices, timings and vet. Per-run build durations are in builds.json; no native products were built. Total session duration is recorded when evidence is published.

Publishing completed about 10 minutes after the start of this unit. Source and test files were restored; only evidence is committed.
