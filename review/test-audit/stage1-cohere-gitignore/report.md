Audited 33 listed tests as 6 rows at 3774cd5ec6b2d8a8c06ac3277e4c7b58c3b58334; nproc=5.
Clean full package cooked at 90 seconds; all bounded clean groups passed.
Verdicts: 1 sacred, 1 subsumed, 1 untrue in the sampled matrix, 2 witnesses, 1 setup-check.
Four production mutants were caught; all three applicable empty-answer probes were caught.
Evidence is on test-audit/stage1-cohere-gitignore under review/test-audit/stage1-cohere-gitignore/.

```json
[
  {
    "test": "TestThePortAnswersAsGoCohereAndGitDo comparison family",
    "package": "stage1/cohere/gitignore",
    "file": "stage1/cohere/gitignore/answers_shards_test.go; stage1/cohere/gitignore/gitignore_test.go",
    "seconds": 36.73,
    "oracle": "Go cohere and git check-ignore actually run; Git v2.51.0 t0008/t3070 scripted expectations corroborate answers. Coverage and nonzero counts are self checks. The extended clean baseline checked the script values against the port and Go cohere.",
    "oracle_kind": [
      "external-run",
      "external-authority",
      "self"
    ],
    "kills": [
      "M1",
      "M2",
      "M3",
      "M4"
    ],
    "unique_kills": [
      "M2",
      "M3"
    ],
    "last_proven_fail": "M4: answers_shards_test.go:488: native and Go cohere differ: line 2: \"1 ::\\tabc\", Go cohere \"0 ::\\tabc\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestThePortAnswersAsGoCohereAndGitDo comparison family",
      "TestPathAnswersAsGosPathPackage",
      "TestEachGapStandsWhereGapsMdSaysItDoes",
      "TestPortAnswerShardPlantedDisagreement",
      "TestPortAnswerShardAssignmentSurvivesCorpusGrowth"
    ],
    "evidence": "git apply review/test-audit/stage1-cohere-gitignore/diffs/M4.diff; COHERE_GIT_SOURCE=/tmp/u094/git-source ADAMIC_BUILD_CACHE_DIR=/tmp/u094/cache/M4 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/gitignore/ -run '^TestThePortAnswersAsGoCohereAndGitDo(_Setup|Union|_00[0-9]|_01[0-5])$' > review/test-audit/stage1-cohere-gitignore/M4-answers.log 2>&1; answers_shards_test.go:488: native and Go cohere differ: line 2: \"1 ::\\tabc\", Go cohere \"0 ::\\tabc\"",
    "members": [
      "TestThePortAnswersAsGoCohereAndGitDo_Setup",
      "TestThePortAnswersAsGoCohereAndGitDo_000",
      "TestThePortAnswersAsGoCohereAndGitDo_001",
      "TestThePortAnswersAsGoCohereAndGitDo_002",
      "TestThePortAnswersAsGoCohereAndGitDo_003",
      "TestThePortAnswersAsGoCohereAndGitDo_004",
      "TestThePortAnswersAsGoCohereAndGitDo_005",
      "TestThePortAnswersAsGoCohereAndGitDo_006",
      "TestThePortAnswersAsGoCohereAndGitDo_007",
      "TestThePortAnswersAsGoCohereAndGitDo_008",
      "TestThePortAnswersAsGoCohereAndGitDo_009",
      "TestThePortAnswersAsGoCohereAndGitDo_010",
      "TestThePortAnswersAsGoCohereAndGitDo_011",
      "TestThePortAnswersAsGoCohereAndGitDo_012",
      "TestThePortAnswersAsGoCohereAndGitDo_013",
      "TestThePortAnswersAsGoCohereAndGitDo_014",
      "TestThePortAnswersAsGoCohereAndGitDo_015",
      "TestThePortAnswersAsGoCohereAndGitDoUnion"
    ],
    "unknown": [
      "M1: comparison shard _006 timed out; _002 fallback was interrupted. No package-wide uniqueness claim."
    ]
  },
  {
    "test": "TestThePortAnswersAsGoCohereAndGitDo mutant witness family",
    "package": "stage1/cohere/gitignore",
    "file": "stage1/cohere/gitignore/answers_shards_test.go",
    "seconds": 42.18,
    "oracle": "Built-in wrong ports must disagree with executed Go cohere, and with git where applicable. W1 disables the byte comparator; witness expectations then fail.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: answers_shards_test.go:529: natively mutant \"R2 the size limit one byte lower\" survives",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestThePortAnswersAsGoCohereAndGitDo mutant witness family"
    ],
    "evidence": "COHERE_GIT_SOURCE=/tmp/u094/git-source ADAMIC_BUILD_CACHE_DIR=/tmp/u094/cache/W1 ADAMIC_GITIGNORE_LARGEST=1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/gitignore/ -run '^TestThePortAnswersAsGoCohereAndGitDo_026$' > review/test-audit/stage1-cohere-gitignore/W1-largest.log 2>&1; answers_shards_test.go:529: natively mutant \"R2 the size limit one byte lower\" survives",
    "members": [
      "TestThePortAnswersAsGoCohereAndGitDo_016",
      "TestThePortAnswersAsGoCohereAndGitDo_017",
      "TestThePortAnswersAsGoCohereAndGitDo_018",
      "TestThePortAnswersAsGoCohereAndGitDo_019",
      "TestThePortAnswersAsGoCohereAndGitDo_020",
      "TestThePortAnswersAsGoCohereAndGitDo_021",
      "TestThePortAnswersAsGoCohereAndGitDo_022",
      "TestThePortAnswersAsGoCohereAndGitDo_023",
      "TestThePortAnswersAsGoCohereAndGitDo_024",
      "TestThePortAnswersAsGoCohereAndGitDo_025",
      "TestThePortAnswersAsGoCohereAndGitDo_026"
    ],
    "witness_kills": [
      "W1"
    ],
    "timing_configuration": "Median of the three standard-corpus family runs, each skipping _026. The opt-in boundary member is timed separately.",
    "boundary_seconds": 27.698,
    "unknown": [],
    "witness_members_proven": 11,
    "narrowed_evidence": "W1 shard _017: answers_shards_test.go:529: natively mutant \"entering from below the root forgetting the files above\" survives"
  },
  {
    "test": "TestPathAnswersAsGosPathPackage",
    "package": "stage1/cohere/gitignore",
    "file": "stage1/cohere/gitignore/path_test.go",
    "seconds": 2.461,
    "oracle": "Go standard-library path.Clean, Base and Dir are called to construct expected bytes; Node also runs the port. Built-in wrong-port subcases were checked separately with W1.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: path_test.go:80: native and Go's path package differ: line 1: \"\\t.\\t.\", Go cohere \".\\t.\\t.\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestThePortAnswersAsGoCohereAndGitDo comparison family"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": 36.73,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestThePortAnswersAsGoCohereAndGitDo comparison family",
      "TestPathAnswersAsGosPathPackage",
      "TestEachGapStandsWhereGapsMdSaysItDoes",
      "TestPortAnswerShardPlantedDisagreement",
      "TestPortAnswerShardAssignmentSurvivesCorpusGrowth"
    ],
    "evidence": "COHERE_GIT_SOURCE=/tmp/u094/git-source ADAMIC_BUILD_CACHE_DIR=/tmp/u094/cache/M4 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/gitignore/ -run '^TestPathAnswersAsGosPathPackage$' > review/test-audit/stage1-cohere-gitignore/M4-path.log 2>&1; path_test.go:80: native and Go's path package differ: line 1: \"\\t.\\t.\", Go cohere \".\\t.\\t.\"",
    "subsumption_basis": "2 caught production mutants; a hint, not a deletion recommendation",
    "witness_kills": [
      "W1"
    ]
  },
  {
    "test": "TestEachGapStandsWhereGapsMdSaysItDoes",
    "package": "stage1/cohere/gitignore",
    "file": "stage1/cohere/gitignore/gaps_test.go",
    "seconds": 1.198,
    "oracle": "Node runs every fixture and must match handwritten GAPS.md stdout; native stdout must match the same value. All ten gaps are closed.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P3"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestThePortAnswersAsGoCohereAndGitDo comparison family",
      "TestPathAnswersAsGosPathPackage",
      "TestEachGapStandsWhereGapsMdSaysItDoes",
      "TestPortAnswerShardPlantedDisagreement",
      "TestPortAnswerShardAssignmentSurvivesCorpusGrowth"
    ],
    "evidence": "COHERE_GIT_SOURCE=/tmp/u094/git-source ADAMIC_BUILD_CACHE_DIR=/tmp/u094/cache/M4 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/gitignore/ -run '^TestEachGapStandsWhereGapsMdSaysItDoes$' > review/test-audit/stage1-cohere-gitignore/M4-gaps.log 2>&1; ok github.com/system-inc/adamic/stage1/cohere/gitignore 1.921s. P3: gaps_test.go:73: natively: exit 0, stdout \"\", stderr \"\"; Node prints \"1\\n\"",
    "limitation": "Only M4 is a reached compiler mutation. Coverage proves expression.go:508 executed. It survived this row; P3 empty-program probe fails. Untrue is limited to this frozen sample."
  },
  {
    "test": "TestPortAnswerShardPlantedDisagreement",
    "package": "stage1/cohere/gitignore",
    "file": "stage1/cohere/gitignore/answers_shards_test.go",
    "seconds": 0.011,
    "oracle": "Handwritten requirement: exactly one shard detects a planted disagreement, and it is portBucket(patterns/planted).",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: answers_shards_test.go:568: planted disagreement caught 0 times in shard -1",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": false,
    "matrix_rows": [
      "TestPortAnswerShardPlantedDisagreement"
    ],
    "evidence": "COHERE_GIT_SOURCE=/tmp/u094/git-source ADAMIC_BUILD_CACHE_DIR=/tmp/u094/cache/W1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/gitignore/ -run '^TestPortAnswerShardPlantedDisagreement$' > review/test-audit/stage1-cohere-gitignore/W1-planted.log 2>&1; answers_shards_test.go:568: planted disagreement caught 0 times in shard -1",
    "witness_kills": [
      "W1"
    ]
  },
  {
    "test": "TestPortAnswerShardAssignmentSurvivesCorpusGrowth",
    "package": "stage1/cohere/gitignore",
    "file": "stage1/cohere/gitignore/answers_shards_test.go",
    "seconds": 0.01,
    "oracle": "Handwritten construction invariant: adding a real-tree query preserves every existing query owner.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S1: answers_shards_test.go:673: adding a repository file moved \"tree/real repository/a.go/false\" from shard-000 to shard-001",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": false,
    "matrix_rows": [
      "TestPortAnswerShardAssignmentSurvivesCorpusGrowth"
    ],
    "evidence": "COHERE_GIT_SOURCE=/tmp/u094/git-source ADAMIC_BUILD_CACHE_DIR=/tmp/u094/cache/S1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/gitignore/ -run '^TestPortAnswerShardAssignmentSurvivesCorpusGrowth$' > review/test-audit/stage1-cohere-gitignore/S1-growth.log 2>&1; answers_shards_test.go:673: adding a repository file moved \"tree/real repository/a.go/false\" from shard-000 to shard-001",
    "construction_kills": [
      "S1"
    ]
  }
]
```

| Id | Origin/main file:line | Change | Failed rows |
|---|---|---|---|
| M1 | stage1/cohere/gitignore/path.ts:51 | `const rooted = path.startsWith('/');` becomes `const rooted = !path.startsWith('/');` | TestThePortAnswersAsGoCohereAndGitDo comparison family, TestPathAnswersAsGosPathPackage |
| M2 | stage1/cohere/gitignore/glob.ts:85 | `const crossesNothing = path && character === 0x2f;` becomes `const crossesNothing = path && character !== 0x2f;` | TestThePortAnswersAsGoCohereAndGitDo comparison family |
| M3 | stage1/cohere/gitignore/gitignore.ts:428 | `if (candidate.directoryOnly && !isDirectory) {` becomes `if (candidate.directoryOnly && isDirectory) {` | TestThePortAnswersAsGoCohereAndGitDo comparison family |
| M4 | internal/lower/expression.go:508 | `return ir.BooleanConstant{Value: node.Kind == ast.KindTrueKeyword}, nil` becomes `return ir.BooleanConstant{Value: node.Kind != ast.KindTrueKeyword}, nil` | TestThePortAnswersAsGoCohereAndGitDo comparison family, TestPathAnswersAsGosPathPackage |

Survivors: none among M1 through M4 in the bounded matrix. No claim about unsampled behavior.

W1 is a harness weakening, S1 a permitted construction break, and P1 through P3 are probes. Their diffs are separate and do not support production uniqueness or subsumption.

Brief issues and costs:

- The listed shard prefix combines answer comparisons, setup/union coverage, and wrong-port witnesses in one checker. Grouping all as one production row would hide witness behavior. We split comparisons and witnesses, and retain construction coverage with the comparison family.
- The full clean package and the M1 comparison group cooked at 90 seconds. M1 has a proven assertion kill in _014 and a timed-out _006; other kills outside the observed set remain unknown.
- The timeout did not kill M1's Node process. The saved process probe shows parent PID 1 and 95.8% CPU after 12 minutes. We terminated its process group. Matrix and weakened-check costs after M1 were inflated; the three standard clean timing runs preceded this orphan.
- W1's cold witness-family run also cooked, after nine witness shards had failed. The opt-in boundary was run separately, and the remaining witness shard failed on its separate rerun.
- The 100 MiB opt-in changes the whole corpus, not just its one witness. Standard family timing excludes this optional member; its three separate enabled timings are reported in boundary_seconds. A full family median with the enlarged corpus was not measured.
- The stage1 rebuild cap conflicts with three mutants per row. We used four frozen production mutants and separate entry probes. Native validation and cache miss/fetch times are retained in costs.json; multiple native products are required by the tests' own built-in mutants.
- The gap row has a distinct compiler entry. It reaches M4 but still passes. This is one reached compiler mutation, not evidence that all ten compiler features are unfailable. Its empty-program probe fails.
- P1/P2 are module entries rather than functions, so their empty answer is implemented by deleting module execution after the helper declarations. P3 returns an empty IR program from Lower at entry under a probe-only selector.
- Coverage lists every reached Go lowering function for the gap row. The TypeScript function inventory is a static call inventory, not instrumented branch coverage; Matcher.root and Matcher.directory are not called by the drivers.
- No repo-wide uniqueness run was attempted. Central replay has standalone no-switch M1 through M4 diffs, validated against the starting commit and built natively; M4 also passed go vet.

Setup/build/run timings:

- Warm env worked; setup skipped. Tool validation 0.07616508199862437 s; npm ci 0.6376268339990929 s; nproc 5.
- Standalone native mutant build commands: 31.317 s total, including Go command compilation. Individual timings are in matrix-runs.json.
- Three-run timing commands: 302.727 s wall total; production matrix commands: 383.417 s wall total; witness/probe commands: 374.171 s wall total. These phases include compilation and execution, so they are not pure compiler times.
- Not covered: full-package green completion, the timed-out M1 shard, a full enlarged-corpus family timing, all compiler gap behaviors beyond M4, other packages, or repo-wide uniqueness.

The separate W1 shard-017 proof took 9.889 s wall and failed with the disabled comparison. All 11 mutant witness members have observed weakened-check failures. The required Go cohere oracle runs were performed by this package harness; no other package was audited or used for repo-wide replay.
