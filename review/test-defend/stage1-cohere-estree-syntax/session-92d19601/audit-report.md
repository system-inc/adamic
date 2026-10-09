# u089 audit

Start: origin/main 3a1c8b5792fb66503290d2d494ce96302bc4386d. Branch test-audit/stage1-cohere-estree-syntax. nproc 5.

24 of the 25 requested names exist. TestThreePortMutants_SetupRequired vanished. The twelve ThreePort numbered shards form one family, leaving 13 current rows. ThreePort rows moved to three_port_mutants_deadline_shards_test.go; LossyInputControl moved to lossy_input_control_products_test.go.

Code under test: the .ts ESTree port entered by main.run and pipeline.answer. Conservative inventory and observed Node coverage are stored as function-inventory.txt and reached-functions.json. Production mutations affect input validation, referenceError, written and main.run. Go cohere and upstream Node libraries are oracles and were never mutated.

All current rows/families passed three original-source timing runs with split native builds and library/throughput opt-ins enabled. No scoped row skipped. Whole package baseline timed out at 90.020s without an earlier failure; the grammar baseline separately passed in 39.271s before split builds were enabled.

## Rows

```json

[
  {
    "test": "TestSyntaxGrammar",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/syntax_test.go",
    "seconds": 23.798,
    "oracle": "Go cohere canonical bytes compared with source Node, sanitized native and emitted JavaScript.",
    "oracle_kind": "external-run",
    "kills": [
      "M2",
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: syntax_test.go:32: native: line 137: Go \"stripped type T=ReturnType<<U>(x:U)=>number>; const b=foo<<U>(x:U)=>number>(()=>1);\", port \"stripped type T=ReturnType<<U>(x:U)=>number>; const b=foo<<U>(x:U)=>number>(()=>1)\"",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestSyntaxRefusals",
      "TestThroughput",
      "TestCookedSurrogates"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1",
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestSyntaxGrammar",
      "TestSyntaxRefusals",
      "TestThroughput",
      "TestCookedSurrogates",
      "TestLossyInputRefusal"
    ],
    "evidence": "selector /tmp/u089/mutant=M4; ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=4 ADAMIC_ESTREE_LIBRARY=/tmp/u089/library ADAMIC_ESTREE_BENCHMARK=1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^TestSyntaxGrammar$'; syntax_test.go:32: native: line 137: Go \"stripped type T=ReturnType<<U>(x:U)=>number>; const b=foo<<U>(x:U)=>number>(()=>1);\", port \"stripped type T=ReturnType<<U>(x:U)=>number>; const b=foo<<U>(x:U)=>number>(()=>1)\""
  },
  {
    "test": "TestSyntaxRefusals",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/syntax_test.go",
    "seconds": 22.348,
    "oracle": "Go cohere error-status count, then handwritten refusal requirements for all port variants: empty stdout and an ESTree parser diagnostic within a deadline. It does not require each exact diagnostic cause.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M2"
    ],
    "unique_kills": [],
    "last_proven_fail": "M2: syntax_test.go:54: [node --disable-warning=ExperimentalWarning /workspace/adamic/oracle/node.mjs /workspace/adamic/stage1/cohere/estree/main.ts /tmp/adamic-gate/TestSyntaxRefusals2722443132/001/0001.ts]: timeout=false exit=<nil> stdout=459 stderr=",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestSyntaxGrammar"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1",
      "P2"
    ],
    "subsumer_seconds": 23.798,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestSyntaxGrammar",
      "TestSyntaxRefusals",
      "TestThroughput",
      "TestCookedSurrogates",
      "TestLossyInputRefusal"
    ],
    "evidence": "selector /tmp/u089/mutant=M2; ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=4 ADAMIC_ESTREE_LIBRARY=/tmp/u089/library ADAMIC_ESTREE_BENCHMARK=1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^TestSyntaxRefusals$'; syntax_test.go:54: [node --disable-warning=ExperimentalWarning /workspace/adamic/oracle/node.mjs /workspace/adamic/stage1/cohere/estree/main.ts /tmp/adamic-gate/TestSyntaxRefusals2722443132/001/0001.ts]: timeout=false exit=<nil> stdout=459 stderr="
  },
  {
    "test": "TestSyntaxLibraries",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/syntax_test.go",
    "seconds": 1.064,
    "oracle": "Go cohere raw/postprocessed JSON compared with pinned typescript-estree and Prettier on Node. It never executes the Adamic port.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestSyntaxGrammar",
      "TestSyntaxRefusals",
      "TestThroughput",
      "TestCookedSurrogates",
      "TestLossyInputRefusal"
    ],
    "evidence": "Three clean timing runs passed; no authorized production target: this row executes Go cohere/upstream libraries and never the Adamic port. Mutating those would mutate the oracle."
  },
  {
    "test": "TestTypeMemberLibraryGap",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/syntax_test.go",
    "seconds": 0.925,
    "oracle": "Go cohere must execute successfully, but its returned AST is discarded; pinned typescript-estree must print the self-written refused label. It never executes the Adamic port.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestSyntaxGrammar",
      "TestSyntaxRefusals",
      "TestThroughput",
      "TestCookedSurrogates",
      "TestLossyInputRefusal"
    ],
    "evidence": "Three clean timing runs passed; no authorized production target: this row executes Go cohere/upstream libraries and never the Adamic port. Mutating those would mutate the oracle."
  },
  {
    "test": "TestThreePortMutants_Setup",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/three_port_mutants_deadline_shards_test.go",
    "seconds": 9.769,
    "oracle": "Self check that shared product preparation produces a non-nil prepared object.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W2: shared product build failed; origin three_port_mutants_deadline_shards_test.go:340 (overlay reported line 331)",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestThreePortMutants_Setup"
    ],
    "evidence": "go test -overlay=setup-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^TestThreePortMutants_Setup$'; shared product build failed (origin line 340, overlay line 331)"
  },
  {
    "test": "TestThreePortMutants family",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/three_port_mutants_deadline_shards_test.go",
    "seconds": 11.711,
    "oracle": "Go cohere canonical records; built-in port mutants must differ from Go on witnesses while source and native agree. Weakened-check proof uses the existing synthetic proof branch with selector -1.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: three_port_mutants_deadline_shards_test.go:452: case 071: source Node mutant survived",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestThreePortMutants_000",
      "TestThreePortMutants_001",
      "TestThreePortMutants_002",
      "TestThreePortMutants_003",
      "TestThreePortMutants_004",
      "TestThreePortMutants_005",
      "TestThreePortMutants_006",
      "TestThreePortMutants_007",
      "TestThreePortMutants_008",
      "TestThreePortMutants_009",
      "TestThreePortMutants_010",
      "TestThreePortMutants_011",
      "TestThreePortMutants_ShardProof"
    ],
    "evidence": "ADAMIC_THREE_PORT_PROOF=-1 go test -overlay=witness-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestThreePortMutants_[0-9]{3}|TestThreePortMutants_ShardProof)$'; three_port_mutants_deadline_shards_test.go:452: case 071: source Node mutant survived",
    "members": [
      "TestThreePortMutants_000",
      "TestThreePortMutants_001",
      "TestThreePortMutants_002",
      "TestThreePortMutants_003",
      "TestThreePortMutants_004",
      "TestThreePortMutants_005",
      "TestThreePortMutants_006",
      "TestThreePortMutants_007",
      "TestThreePortMutants_008",
      "TestThreePortMutants_009",
      "TestThreePortMutants_010",
      "TestThreePortMutants_011"
    ]
  },
  {
    "test": "TestThreePortMutants_ShardProof",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/three_port_mutants_deadline_shards_test.go",
    "seconds": 0.372,
    "oracle": "Synthetic oracle/mutant records; each planted survivor must fail exactly its owning shard with the expected message.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: three_port_mutants_deadline_shards_test.go:492: unexpected failure in TestThreePortMutants_002: === RUN   TestThreePortMutants_002",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestThreePortMutants_000",
      "TestThreePortMutants_001",
      "TestThreePortMutants_002",
      "TestThreePortMutants_003",
      "TestThreePortMutants_004",
      "TestThreePortMutants_005",
      "TestThreePortMutants_006",
      "TestThreePortMutants_007",
      "TestThreePortMutants_008",
      "TestThreePortMutants_009",
      "TestThreePortMutants_010",
      "TestThreePortMutants_011",
      "TestThreePortMutants_ShardProof"
    ],
    "evidence": "ADAMIC_THREE_PORT_PROOF=-1 go test -overlay=witness-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestThreePortMutants_[0-9]{3}|TestThreePortMutants_ShardProof)$'; three_port_mutants_deadline_shards_test.go:492: unexpected failure in TestThreePortMutants_002: === RUN   TestThreePortMutants_002"
  },
  {
    "test": "TestThroughput",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/throughput_test.go",
    "seconds": 27.646,
    "oracle": "Go cohere complete canonical bytes checked on every rotated Go, source-Node and native invocation. Throughput values have no pass/fail speed threshold.",
    "oracle_kind": "external-run",
    "kills": [
      "M3",
      "M4"
    ],
    "unique_kills": [
      "M3"
    ],
    "last_proven_fail": "M4: throughput_test.go:45: source Node: line 121: Go \"stripped type X = import('x').A<T>; type Y = typeof import('x', { with: { type: 'json' } }).A; type Z = typeof this.a;\", port \"stripped type X = import('x').A<T>; type Y = typeof import('x', { with: { type: 'json' } }).A; type Z = typeof this.a\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1",
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestSyntaxGrammar",
      "TestSyntaxRefusals",
      "TestThroughput",
      "TestCookedSurrogates",
      "TestLossyInputRefusal"
    ],
    "evidence": "selector /tmp/u089/mutant=M4; ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=4 ADAMIC_ESTREE_LIBRARY=/tmp/u089/library ADAMIC_ESTREE_BENCHMARK=1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^TestThroughput$'; throughput_test.go:45: source Node: line 121: Go \"stripped type X = import('x').A<T>; type Y = typeof import('x', { with: { type: 'json' } }).A; type Z = typeof this.a;\", port \"stripped type X = import('x').A<T>; type Y = typeof import('x', { with: { type: 'json' } }).A; type Z = typeof this.a\""
  },
  {
    "test": "TestCookedSurrogates",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/utf8_test.go",
    "seconds": 21.753,
    "oracle": "Go cohere canonical bytes compared with source Node, sanitized native and emitted JavaScript.",
    "oracle_kind": "external-run",
    "kills": [
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: utf8_test.go:22: Node: line 13: Go \"stripped '\\\\u005cud800a\\\\u005cudc00';\", port \"stripped '\\\\u005cud800a\\\\u005cudc00'\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestSyntaxGrammar"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1",
      "P2"
    ],
    "subsumer_seconds": 23.798,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestSyntaxGrammar",
      "TestSyntaxRefusals",
      "TestThroughput",
      "TestCookedSurrogates",
      "TestLossyInputRefusal"
    ],
    "evidence": "selector /tmp/u089/mutant=M4; ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=4 ADAMIC_ESTREE_LIBRARY=/tmp/u089/library ADAMIC_ESTREE_BENCHMARK=1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^TestCookedSurrogates$'; utf8_test.go:22: Node: line 13: Go \"stripped '\\\\u005cud800a\\\\u005cudc00';\", port \"stripped '\\\\u005cud800a\\\\u005cudc00'\""
  },
  {
    "test": "TestCookedSurrogateMutant",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/utf8_test.go",
    "seconds": 21.931,
    "oracle": "Go cohere output; the built-in surrogate serialization mutant must disagree on source Node and sanitized native.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: utf8_test.go:36: Node mutant survived",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCookedSurrogateMutant"
    ],
    "evidence": "ADAMIC_THREE_PORT_PROOF=-1 go test -overlay=witness-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^TestCookedSurrogateMutant$'; utf8_test.go:36: Node mutant survived"
  },
  {
    "test": "TestCookedSurrogateLibraryGap",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/utf8_test.go",
    "seconds": 0.948,
    "oracle": "Go cohere output must contain the expected replacement sequence; upstream Node parse must produce the expected UTF-16 JSON string. Both gap labels are handwritten and the port is never executed.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestSyntaxGrammar",
      "TestSyntaxRefusals",
      "TestThroughput",
      "TestCookedSurrogates",
      "TestLossyInputRefusal"
    ],
    "evidence": "Three clean timing runs passed; no authorized production target: this row executes Go cohere/upstream libraries and never the Adamic port. Mutating those would mutate the oracle."
  },
  {
    "test": "TestLossyInputRefusal",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/utf8_test.go",
    "seconds": 21.721,
    "oracle": "Self-written requirement to refuse both malformed input and a literal replacement character, with empty stdout and the expected cannot recover original UTF-8 bytes message. Node runs the port, not an independent oracle.",
    "oracle_kind": "self",
    "kills": [
      "M1"
    ],
    "unique_kills": [
      "M1"
    ],
    "last_proven_fail": "M1: utf8_test.go:69: [node --disable-warning=ExperimentalWarning /workspace/adamic/oracle/node.mjs /workspace/adamic/stage1/cohere/estree/main.ts /tmp/adamic-gate/TestLossyInputRefusal205561031/002/input.ts]: timeout=false exit=<nil> stdout=381 stderr=",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1",
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestSyntaxGrammar",
      "TestSyntaxRefusals",
      "TestThroughput",
      "TestCookedSurrogates",
      "TestLossyInputRefusal"
    ],
    "evidence": "selector /tmp/u089/mutant=M1; ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=4 ADAMIC_ESTREE_LIBRARY=/tmp/u089/library ADAMIC_ESTREE_BENCHMARK=1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^TestLossyInputRefusal$'; utf8_test.go:69: [node --disable-warning=ExperimentalWarning /workspace/adamic/oracle/node.mjs /workspace/adamic/stage1/cohere/estree/main.ts /tmp/adamic-gate/TestLossyInputRefusal205561031/002/input.ts]: timeout=false exit=<nil> stdout=381 stderr="
  },
  {
    "test": "TestLossyInputControl",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/lossy_input_control_products_test.go",
    "seconds": 5.415,
    "oracle": "Go cohere output; the built-in disabled-input-guard mutant must disagree on source Node and sanitized native.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: lossy_input_control_products_test.go:60: Node lossy-input mutant survived",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestLossyInputControl"
    ],
    "evidence": "ADAMIC_THREE_PORT_PROOF=-1 go test -overlay=witness-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^TestLossyInputControl$'; lossy_input_control_products_test.go:60: Node lossy-input mutant survived"
  }
]

```

## Production mutants

| ID | Origin location | Change | Failed bounded rows |

|---|---|---|---|

| M1 | stage1/cohere/estree/pipeline.ts:11 | `if(text.includes('\ufffd'))` -> `if(false)` | TestLossyInputRefusal |

| M2 | stage1/cohere/estree/sourceReferences.ts:29 | `comment.end > firstToken` -> `comment.end < firstToken` | TestSyntaxGrammar, TestSyntaxRefusals |

| M3 | stage1/cohere/estree/protocol.ts:20 | `unit <= 126` -> `unit <= 125` | TestThroughput |

| M4 | stage1/cohere/estree/main.ts:8 | `.slice(0, -1)` -> `.slice(0, -2)` | TestSyntaxGrammar, TestThroughput, TestCookedSurrogates |

## Witness and setup changes

W1: estree_test.go:163, firstDifference returns an empty difference. Existing synthetic family control (-1) passes normally, then family members 002, 006 and 011 fail under W1. The shard proof, cooked surrogate mutant and lossy input control also fail. W1 is a harness weakening, never a production kill.

W2: three_port_mutants_deadline_shards_test.go:290, drop the complete shared once.Do construction statement. Setup now fails with shared product build failed. Its original caller line is 340; overlay deletion shifts the reported line to 331. Both harness diffs passed go vet with their overlays.

## Probe

P1: pipeline.answer at origin pipeline.ts:10 returns an empty string, which the CLI logs as a newline. P2: main.run at origin main.ts:3 returns immediately and produces no output. P2 is the primary CLI-entry vacuity probe; P1 separately checks the semantic answer entry. Both are separate from production mutants. Only the five ordinary port rows are probed; witness/setup/library-only rows have vacuous=null.

## Survivors

None in the five-row bounded matrix.

## Brief ambiguities and costs

The brief says 14 rows and lists 25 historical functions. Grouping its twelve shards explains 14; the vanished SetupRequired reduces the current count to 13. No replacement test was silently substituted.

The historical file reference is 8de93800f4; the required fresh origin/main is 3a1c8b57. All diff locations use the actual starting commit.

The three library-only rows compare or document Go/upstream behavior without executing the port. There is no authorized port mutation target for them. They are cannot-judge, not untrue, and their clean runs alone prove no port gate.

The family is a witness: production mutations can break its preparation or anchors without proving its comparison. Its verdict comes only from weakening that comparison through a Go overlay. The synthetic proof branch was used for the family weakening to exercise its checker without unnecessary native rebuilds; its cost was still measured by three ordinary full runs.

The 90-second package budget cannot fit the package or the five ordinary port rows together. The matrix runs each of those five individually and is bounded. Other port agreement tests and product/corpus/shard rows outside that set were not replayed; package or repository uniqueness is unknown.

The four native-mutant limit takes precedence over about three mutants per row. Four code-derived mutations were fixed before production results. The initially proposed diagnostic-condition flip was revised to the first input-validation guard before any production mutant ran.

The harness rebuilds a native port inside each ordinary row. A file selector avoids different switched source builds but does not stop this harness from invoking the compiler again. Per-row matrix elapsed times include rebuild and execution; those components were not separately instrumented. Standalone sanitized native rebuild times are separately recorded.

/usr/bin/time is unavailable. The initial dependency command failed before npm ran, and the package attempt afterward was not treated as a valid baseline. Dependencies were rerun with Bash timing before the valid baseline.

The standalone probe build was initially attempted before its diff had been generated; it failed to open the patch. The diff was then generated, applied and successfully compiled. No verdict rests on the failed command.

The original family/setup first run includes cold products; setup cost was 82.858s initially, then 9.394s and 9.769s. Medians reflect the required three runs rather than hiding the cold result. Timing runs enabled NODE_V8_COVERAGE to collect source reach. Native validation builds and witness runs overlapped some timing runs, so these include profiling and workspace contention rather than uncontended machine costs.

SyntaxRefusals checks a status count and generic parser refusal text, not the exact cause of every error. TypeMemberLibraryGap ignores the successful Go AST output. CookedSurrogateLibraryGap checks a specific substring and JSON value rather than whole AST agreement. Throughput enforces output correctness but no speed threshold.

The TS port has no named main function: a top-level argument dispatcher calls main.run, which calls pipeline.answer. A semantic empty-string probe still lets console.log emit a newline. An additional zero-output main.run probe resolved that entry ambiguity and added one standalone rebuild plus five bounded runs, extending the approximate 30-minute budget.

The source function inventory is conservative. Node coverage records observed source functions on the clean inputs; it is not native coverage and does not establish exact reach for unexecuted branches.

The rebuilt native mutants also have independent before/after behavioral witnesses: M1 changes replacement-input refusal from exit 70 to exit 0; M2 changes an accepted trailing reference from exit 0 to exit 70; M3 and M4 change canonical bytes; P1 prints only a newline. These compare an untouched source-port copy with individually rebuilt native mutants and are behavior witnesses, not additional matrix kills.

No completed external corpus directory was supplied or generated. No scoped row required it; corpus rows outside this unit were not covered.

Subsumption and uniqueness are bounded hints over four production mutants, not deletion recommendations. No oracle, production Go compiler, or test assertion was modified for production kills.

## Timings

Warm setup skipped: 0s. npm ci: 0.472s. Pinned library installation: 3.089s. Native split enabled with four jobs on five CPUs.

Standalone sanitized rebuilds: M1=26.749s (exit 0), M2=21.858s (exit 0), M3=22.605s (exit 0), M4=23.8s (exit 0), P1=30.034s (exit 0), P2=28.775s (exit 0).

Original isolated invocation wall total: 657.392s.

Matrix invocation wall total: 744.113s.

Whole package: 90.020s, cooked. No tests in other packages and no central repo-wide replay were run.

## Evidence

Standalone production/probe diffs, permitted witness diffs, exact raw logs, timing triples, native build checks, source reach inventory, scratch switch, and executable scripts are kept beside this report. Production source is restored before commit.
