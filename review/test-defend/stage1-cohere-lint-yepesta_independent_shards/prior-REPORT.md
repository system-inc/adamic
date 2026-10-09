u113 audited at ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2; all 24 named functions exist.
Family grouping gives five rows: comparison, two witnesses, setup aliases, oracle product.
Four production mutants survived the bounded matrix with demonstrated behavior changes.
Both witnesses failed under weakened checks; setup caught the missing oracle, the product check did not.
Vacuity is unknown because the empty-answer probe could not finish C emission within budget.

```json
[
  {
    "test": "TestMutantsReactJsxNoCommentTextnodes family",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/yepesta_independent_shards_test.go",
    "seconds": 6.574,
    "oracle": "Node runs the same copied, already-mutated TypeScript as the compiled products; full byte equality checks compiler fidelity. Go cohere disagreement is logged but not asserted by the shards. Union construction expectations are self-written.",
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
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMutantsReactJsxNoCommentTextnodes family",
      "TestMutantsReactJsxNoCommentTextnodesMutantKilled",
      "TestMutantsReactJsxNoCommentTextnodesPlantedFailure",
      "TestMutantsReactJsxNoCommentTextnodesSetup family",
      "TestProduct_YepestaOracle"
    ],
    "evidence": "M1,M2,M3: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run '^(TestMutantsReactJsxNoCommentTextnodes.*|TestProduct_YepestaOracle)$'; package PASS and all 24 test pass events. M4-warm-lowered.log and five M4-isolated-*.log likewise pass. No proven assertion failure under these four production mutants.",
    "members": [
      "TestMutantsReactJsxNoCommentTextnodesUnion",
      "TestMutantsReactJsxNoCommentTextnodes_000",
      "TestMutantsReactJsxNoCommentTextnodes_001",
      "TestMutantsReactJsxNoCommentTextnodes_002",
      "TestMutantsReactJsxNoCommentTextnodes_003",
      "TestMutantsReactJsxNoCommentTextnodes_004",
      "TestMutantsReactJsxNoCommentTextnodes_005",
      "TestMutantsReactJsxNoCommentTextnodes_006",
      "TestMutantsReactJsxNoCommentTextnodes_007",
      "TestMutantsReactJsxNoCommentTextnodes_008",
      "TestMutantsReactJsxNoCommentTextnodes_009",
      "TestMutantsReactJsxNoCommentTextnodes_010",
      "TestMutantsReactJsxNoCommentTextnodes_011",
      "TestMutantsReactJsxNoCommentTextnodes_012",
      "TestMutantsReactJsxNoCommentTextnodes_013",
      "TestMutantsReactJsxNoCommentTextnodes_014",
      "TestMutantsReactJsxNoCommentTextnodes_015"
    ]
  },
  {
    "test": "TestMutantsReactJsxNoCommentTextnodesMutantKilled",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/yepesta_independent_shards_test.go",
    "seconds": 1.8,
    "oracle": "Go cohere runs outside Adamic. This witness requires any byte disagreement, not a specified missing JSX finding; healthy subprocess exit is checked separately.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W2: yepesta_independent_shards_test.go:183: JSX textnode mutant survived on Node",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMutantsReactJsxNoCommentTextnodes family",
      "TestMutantsReactJsxNoCommentTextnodesMutantKilled",
      "TestMutantsReactJsxNoCommentTextnodesPlantedFailure",
      "TestMutantsReactJsxNoCommentTextnodesSetup family",
      "TestProduct_YepestaOracle"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run '^TestMutantsReactJsxNoCommentTextnodesMutantKilled$' > W2.log 2>&1; yepesta_independent_shards_test.go:183: JSX textnode mutant survived on Node"
  },
  {
    "test": "TestMutantsReactJsxNoCommentTextnodesPlantedFailure",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/yepesta_independent_shards_test.go",
    "seconds": 1.651,
    "oracle": "Self-written inserted output and expected catch count 1. The owner shard is logged, not asserted against its assignment. Go cohere supplies the output being copied.",
    "oracle_kind": [
      "self",
      "external-run"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: yepesta_independent_shards_test.go:227: planted disagreement caught by 0 shards",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMutantsReactJsxNoCommentTextnodes family",
      "TestMutantsReactJsxNoCommentTextnodesMutantKilled",
      "TestMutantsReactJsxNoCommentTextnodesPlantedFailure",
      "TestMutantsReactJsxNoCommentTextnodesSetup family",
      "TestProduct_YepestaOracle"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run '^TestMutantsReactJsxNoCommentTextnodesPlantedFailure$' > W1.log 2>&1; yepesta_independent_shards_test.go:227: planted disagreement caught by 0 shards"
  },
  {
    "test": "TestMutantsReactJsxNoCommentTextnodesSetup family",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/yepesta_independent_shards_test.go",
    "seconds": 1.423,
    "oracle": "Self: shared product preparation must succeed. No explicit validation of nonempty rows or bundle paths: S2 persisted empty fields and all four setup aliases passed.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S1: yepesta_independent_shards_test.go:273: shared preparation failed",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMutantsReactJsxNoCommentTextnodes family",
      "TestMutantsReactJsxNoCommentTextnodesMutantKilled",
      "TestMutantsReactJsxNoCommentTextnodesPlantedFailure",
      "TestMutantsReactJsxNoCommentTextnodesSetup family",
      "TestProduct_YepestaOracle"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run '^TestProduct_YepestaOracle$|^TestMutantsReactJsxNoCommentTextnodes(OracleSetup|LoweredSetup|NativeSetup|_Setup)$' > S1.log 2>&1; yepesta_independent_shards_test.go:273: shared preparation failed",
    "members": [
      "TestMutantsReactJsxNoCommentTextnodesOracleSetup",
      "TestMutantsReactJsxNoCommentTextnodesLoweredSetup",
      "TestMutantsReactJsxNoCommentTextnodesNativeSetup",
      "TestMutantsReactJsxNoCommentTextnodes_Setup"
    ]
  },
  {
    "test": "TestProduct_YepestaOracle",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/yepesta_products_test.go",
    "seconds": 1.428,
    "oracle": "Self: build callback success only. It does not assert executable existence: S1 returned successfully without building it, oracle_exists=false in construction-witnesses.json.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMutantsReactJsxNoCommentTextnodes family",
      "TestMutantsReactJsxNoCommentTextnodesMutantKilled",
      "TestMutantsReactJsxNoCommentTextnodesPlantedFailure",
      "TestMutantsReactJsxNoCommentTextnodesSetup family",
      "TestProduct_YepestaOracle"
    ],
    "evidence": "S1 construction run: --- PASS: TestProduct_YepestaOracle (1.70s); construction-witnesses.json proves oracle_exists=false."
  }
]
```

| ID | Starting file:line | Change | Failed rows |
|---|---|---|---|
| M1 | stage1/cohere/lint/rules/react-jsx-no-comment-textnodes/rule.ts:5 | context.kind(index) !== 'JsxText' -> context.kind(index) === 'JsxText' | [] |
| M2 | stage1/cohere/lint/rules/react-jsx-no-comment-textnodes/rule.ts:13 | character === '\n' -> character === '\r' | [] |
| M3 | stage1/cohere/lint/rules/react-jsx-no-comment-textnodes/rule.ts:9 | context.source.slice(node.pos, node.end) -> context.source.slice(node.pos, node.end - 1) | [] |
| M4 | stage1/cohere/lint/lint.ts:104 | this.findings.sort(compareFindings); -> [drop statement] | [] |
| W1 | stage1/cohere/lint/lint_test.go:158 | replace comparison body with return "" | TestMutantsReactJsxNoCommentTextnodesPlantedFailure |
| W2 | stage1/cohere/lint/yepesta_independent_shards_test.go:182 | condition -> true, treating every answer as equal | TestMutantsReactJsxNoCommentTextnodesMutantKilled |
| S1 | stage1/cohere/lint/yepesta_independent_shards_test.go:266 | drop oracle build call and return nil | TestMutantsReactJsxNoCommentTextnodesSetup family |
| S2 | stage1/cohere/lint/yepesta_independent_shards_test.go:321 | marshal yepestaBundle{} | [] |

W1/W2 are witness edits and S1/S2 are construction edits. They are excluded from production kills and uniqueness. P1 is an unvalidated probe, not a mutant.

Survivors:

- M1: Node stub on `// comment`: one finding before, zero after.
- M2: Node stub on `a\n// comment`: one finding before, zero after.
- M3: Node stub on `//`: one finding before, zero after.
- M4: actual Node port on the 87-case persisted manifest: first finding moved from column 1 to column 46 after dropping the sort.
- S2 construction survivor: persisted bundle has empty Oracle/Port/JS/Binary and null Rows/Witnesses; four setup aliases passed.

Witness commands, exact output, and source-copy probes are in survivor-witnesses.json, survivor_witness.py, M*-witness-*.log and construction-witnesses.json.

Brief ambiguities and costs:

- The heading says eight rows, the list says 24 functions, and the family rule groups these into five rows. The numbered shards plus their union form one family. The four setup functions have identical bodies calling yepestaFetch; grouping those aliases avoids treating their names as different coverage. All members are listed in rows.json. No names moved or vanished at this commit.
- Mutant names do not identify one kind of check here. The shard family asserts native/emitted-JavaScript agreement with Node running the same mutated port; it only logs disagreement with Go cohere. MutantKilled is an existential disagreement witness. PlantedFailure checks the byte comparator by counting one disagreeing shard. They require different audit edits despite sharing a prefix.
- The port-only mutation rule limits what this audit can conclude about compiler fidelity. Portable TypeScript errors are reproduced by Node and both compiled products. The four observed survivors prove altered lint behavior was not rejected by this bounded set, not that the compiler-agreement assertions cannot catch a backend defect. The untrue label is limited to this allowed mutation experiment; it is not a deletion recommendation.
- Setup labels imply separate oracle/lowered/native phases, but all four call the same full preparation. Cold setup exceeded the budget, even with warm tools. The complete package cooked without individual fail events; the named slice cooked during native preparation after publishing its lowered product, then passed in 26.918 s using that product. Warm family medians exclude cold building.
- A general return-at-entry insertion is not automatically type-correct in TypeScript: unreachable code lost its discriminated-union and optional-value narrowing. The replacement-body probe passed loading/lowering but cooked twice in C emission, so no empty-answer behavior was observed. Its diff is explicitly unvalidated, excluded from kills, and vacuous is null. The third narrowed attempt was cancelled rather than running five more cold preparations.
- M4 cooked at 90 seconds, then passed with cached lowering. Five isolated grouped-row runs also passed; no later row is inferred from the aborted first run. A successful binary can take more than 90 shell seconds after compilation, as M2 did. The brief explicitly defines the binary limit separately from its 120 second outer backstop; that interpretation is used here.
- The existing canary builder intentionally applies its own mutant.json transformation before compiling. All four audit diffs apply independently to the starting commit and successfully rebuild that package-native canary; the built-in transformation remains part of this unit, not an audit mutation. W/S diffs passed go vet. The uncompleted P1 diff is clearly separated from validated replay patches.
- Broad suite-content cache inputs rebuilt the Go oracle for harness-only edits. The audit coordinator also required a restart while M1 was running; its child completed with all 24 passing and the source was restored. M1 shell wall timing was lost, so only its recorded binary timing is claimed. Progress logs retain the failed resume and correction.
- source-functions.txt is a conservative static inventory of port and parser/scanner declarations, not a proof that every declaration ran. callers.txt lists other harness callers. The bounded matrix ran exactly these 24 functions, after grouping into five rows; it did not run every caller or establish package-wide/repo-wide uniqueness. No row is called sacred on that basis.

Timing and limits:

{
  "nproc": 5,
  "setup_skipped": true,
  "warm_tools_seconds": 0.045725176998530515,
  "npm_ci_seconds": 0.7061744540005748,
  "registry_seconds": 0.38813403400126845,
  "list_seconds": 8.863889277999988,
  "recorded_run_wall_seconds": 683.7741138940219,
  "recorded_binary_seconds": 658.074,
  "M1_shell_wall_unknown_after_coordinator_restart": true,
  "successful_product_miss_seconds": 311.25,
  "unrecorded_cancelled_probe": "P1-isolated-comparison cancelled after two emission timeouts; no assertion result",
  "generated_at_utc": "2026-10-09T12:47:06.668723+00:00"
}

Successful product misses (nested times, do not add to test wall again):

| Run | Lowered product seconds | Native product seconds |
|---|---:|---:|
| M1 | 71.71 | 4.79 |
| M2 | 72.54 | 3.53 |
| M3 | 70.03 | 5.39 |
| M4 | 70.58 | 12.68 |

Setup was skipped; npm ci and registry generation ran before baseline. No named row skipped. Mutated sources are restored. Patches M1-M4/W1-W2/S1-S2 apply against the starting commit; all completed products compiled and all Go scratch edits passed vet. The whole package, other packages, a dynamic complete call graph, and empty-answer behavior remain uncovered. P1’s unvalidated diff is retained only for diagnosis, not as a validated replay mutant.

Initial baselines add 90.382 + 90.049 + 26.918 = 207.349 binary seconds to the runner aggregate, for 865.423 logged binary seconds overall. Cancelled probe time is not included. Unit wall work was about 29 minutes.

Final restored-source named baseline passed in 7.896 binary seconds. Raw .log evidence is explicitly included despite the repository log ignore rule.
