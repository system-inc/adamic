Unit u097: all three requested names exist at a467d1a1571c43e0f01fc4efcb43890280e4a1ac.
Constructor: sacred within the bounded matrix; witness: witness; growth: setup-check.
Three production mutants: two caught, one survivor with changed refusal location.
Median binary seconds: 0.458, 0.201, 0.009; nproc 5 (cgroup quota 4).
Whole package cooked at 90.823 seconds; package and repository uniqueness remain unknown.

[
  {
    "test": "TestPrinterConstructorGap",
    "package": "stage1/cohere/graphql/printer",
    "file": "stage1/cohere/graphql/printer/gaps_test.go",
    "seconds": 0.458,
    "oracle": "Node executes the fixture and checks exit 0 and ready\\n; Lower must return self-written Refused.What substring. Refused.Where is not checked: M03 passes with a different refusal location.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M01",
      "M02"
    ],
    "unique_kills": [
      "M01",
      "M02"
    ],
    "last_proven_fail": "M02: gaps_test.go:34: gap changed; update GAPS.md and remove workaround: <nil>",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestPrinterConstructorGap",
      "TestPrinterWhitespacePlantedDisagreement",
      "TestPrinterWhitespaceShardGrowth"
    ],
    "evidence": "ADAMIC_MUTANT=M02 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/graphql/printer/ -run '^(TestPrinterConstructorGap|TestPrinterWhitespacePlantedDisagreement|TestPrinterWhitespaceShardGrowth)$' > M02.log 2>&1; gaps_test.go:34: gap changed; update GAPS.md and remove workaround: <nil>"
  },
  {
    "test": "TestPrinterWhitespacePlantedDisagreement",
    "package": "stage1/cohere/graphql/printer",
    "file": "stage1/cohere/graphql/printer/gaps_test.go",
    "seconds": 0.201,
    "oracle": "Self-written expected owner, planted case ID and exactly one disagreement. Node transports synthetic answers; it is not a semantic printer oracle.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: gaps_test.go:195: 0 shards caught planted disagreement, want 1",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestPrinterConstructorGap",
      "TestPrinterWhitespacePlantedDisagreement",
      "TestPrinterWhitespaceShardGrowth"
    ],
    "evidence": "timeout 120 go test -overlay /workspace/u097-tmp/verify/W01/overlay.json -json -count=1 -timeout 90s ./stage1/cohere/graphql/printer/ -run '^(TestPrinterConstructorGap|TestPrinterWhitespacePlantedDisagreement|TestPrinterWhitespaceShardGrowth)$' > W01.log 2>&1; gaps_test.go:195: 0 shards caught planted disagreement, want 1",
    "witness_kills": [
      "W01"
    ]
  },
  {
    "test": "TestPrinterWhitespaceShardGrowth",
    "package": "stage1/cohere/graphql/printer",
    "file": "stage1/cohere/graphql/printer/gaps_test.go",
    "seconds": 0.009,
    "oracle": "Self-written fixed shard count, complete union with unchanged cases, and unchanged owner after growth.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S01: gaps_test.go:205: growth changed shard count",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [
      "P02"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestPrinterConstructorGap",
      "TestPrinterWhitespacePlantedDisagreement",
      "TestPrinterWhitespaceShardGrowth"
    ],
    "evidence": "timeout 120 go test -overlay /workspace/u097-tmp/verify/S01/overlay.json -json -count=1 -timeout 90s ./stage1/cohere/graphql/printer/ -run '^(TestPrinterConstructorGap|TestPrinterWhitespacePlantedDisagreement|TestPrinterWhitespaceShardGrowth)$' > S01.log 2>&1; gaps_test.go:205: growth changed shard count",
    "setup_kills": [
      "S01",
      "S02"
    ]
  }
]

| ID | Origin/main file:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/lower/class.go:741 | last = statement.End() -> last = 0 (change constant) | TestPrinterConstructorGap |
| M02 | internal/lower/class.go:765 | node.Pos() >= l.unsetUntil -> node.Pos() < l.unsetUntil (flip condition) | TestPrinterConstructorGap |
| M03 | internal/lower/class.go:770 | drop return nil from ordinary-field exception (drop statement) | [] |
| W01 | stage1/cohere/graphql/printer/shards_test.go:256 | printerShardDisagreement returns nil at entry (weaken comparison) | TestPrinterWhitespacePlantedDisagreement |
| S01 | stage1/cohere/graphql/printer/gaps_test.go:154 | whitespaceShards allocation bound adds one (off-by-one construction bound) | TestPrinterWhitespaceShardGrowth |
| S02 | stage1/cohere/graphql/printer/gaps_test.go:158 | drop entire case-assignment loop (drop construction loop) | TestPrinterWhitespaceShardGrowth, TestPrinterWhitespacePlantedDisagreement |
| P02 | stage1/cohere/graphql/printer/gaps_test.go:153 | whitespaceShards returns nil at entry (empty construction probe) | TestPrinterWhitespaceShardGrowth, TestPrinterWhitespacePlantedDisagreement |
| P01 | internal/lower/lower.go:20 | Lower returns nil, nil at entry (empty-answer probe) | TestPrinterConstructorGap |

M03 survivor witness: control.log reports nestedInner.ts:5:29; M03.log reports nestedConstructor.ts:6:9, with the same Refused.What. All three assigned rows pass M03. This is changed, unguarded refusal-location behavior within this matrix, not an equivalent candidate. Other package rows may catch it.

W01 is the weakened-comparison witness check, S01 and S02 are construction faults, and P01/P02 are probes. They are separate from production kills. S02/P02 break the witness's construction precondition; those failures do not establish witness strength. Only W01 establishes its verdict. Probe P02 is judged only for the construction entry called by growth. The disagreement witness has no applicable production-entry empty probe, so vacuity is null. No families were grouped: the three rows check different things.

Clarifications, costs, and limits:

- The advertised stage1 port location is misleading for this slice. ConstructorGap calls Go lower.Lower directly and inspects a refusal. Its code under test is the compiler, not a .a printer. The other two rows test the suite's comparison and construction. No port, fixture, Node implementation, or oracle source was mutated.
- The constructor row pins a known compiler gap rather than demonstrating semantic correctness. Both production kills remove the expected refusal even though Node still executes the valid fixture. Its self-written message oracle accepts a different refusal with the same wording under M03. It also does not check Node stderr.
- The full clean package hit its 90-second binary timeout, with no preceding assertion failure observed. The narrowed clean control and three standalone timing runs per assigned row passed. Kills and unique_kills are bounded to exactly the listed matrix_rows. No package uniqueness or repository uniqueness is claimed. potential-lowering-callers.txt lists excluded callers for replay.
- TestPrinterThroughput skipped in the whole-package baseline behind ADAMIC_GRAPHQL_PRINTER_BENCH. It is outside this slice. None of the three assigned rows skipped. Prettier was enabled for baseline with ADAMIC_GRAPHQL_PRETTIER pointing at the installed external dependencies.
- Warm tools passed the env.sh check; setup was skipped. npm ci ran in stage3/api before baseline (387 ms npm-reported). The fresh optional oracle directory was installed before baseline with pinned Prettier 3.9.6 and GraphQL 17.0.2 (1 second npm-reported). It used npm install to create dependencies and a lock, rather than npm ci in that newly created directory.
- The source reach inventory contains 254 positive-coverage lowering functions, saved before mutants were chosen. Three menu mutants cover lastFieldAssignment and useOfThis; this meets approximately three per production row, rather than treating the witness and construction rows as production tests. It is not exhaustive path coverage.
- Every standalone diff passed git apply --check --cached against the starting commit and an isolated go vet overlay. M03 retains the empty if because its condition still uses field, so the deletion compiles. P01 drops the full Lower body and now-unused imports; W01/P02 replace full helper bodies to avoid unreachable-code vet failures. selector.diff captures the compiled scratch switch, which was restored after runs.
- Source-relative failure lines in overlay logs move when a body or loop is deleted. W01 failure is origin gaps_test.go:195; S01 is :205; S02 union failure is origin :209 (log :205), and its witness-precondition failure is origin :195 (log :191); P02 count failure is origin :205 (log :197). Mutation locations in the table are all against the starting origin/main.
- No native rebuild is required for the assigned compiler-refusal row because it never emits or executes a native product. Each production run nevertheless used its own ADAMIC_BUILD_CACHE_DIR. The switched Go source was compiled once through Go's cache; later modes reuse it. Command wall minus binary elapsed is only an upper bound on build/driver overhead, not a measured native rebuild time.
- /usr/bin/time was absent during final restoration verification. That command exited 127; verification was rerun with Python's monotonic timer and saved to restored.json/restored.log.

Timings: setup 0 s; dependency commands report 0.387 s and 1 s. Three-run timing commands total 49.563 s. Switched control and production/probe matrix commands total 23.582 s, binary elapsed total 0.950 s. Comparison/construction/probe runs total 25.658 s. Seven isolated vet builds total 10.639 s. Full baseline binary elapsed 90.823 s. Coverage and final restoration logs are also included. Entire work was about 15 minutes including inspection and evidence preparation.
Not covered: full-package mutant matrix, repo-wide replay, all lowering paths, native port correctness and throughput. All source scratch changes restored. Evidence contains replayable diffs, source reach inventory, fixed plans, drivers, timings, JSON matrices, raw logs and results.
