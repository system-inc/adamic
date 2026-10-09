Unit u150: all 16 requested names exist at a7448d73cd17f16362b6cbc5c5c111080da64e43.
Nine grouped rows: eight file shards form one family; Union has an independent planted-disagreement assertion.
Clean whole-package run hit the outer 120-second limit; all nine selected rows passed three narrowed clean runs.
Four production mutants, three port-entry probes, one setup probe, three witness checks and one construction break were run.
Evidence branch: test-audit/stage1-cohere-yaml-compose; directory: review/test-audit/stage1-cohere-yaml-compose/.

```json
[
  {
    "test": "TestFormatterMatchesGo",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/format_test.go",
    "seconds": 46.414,
    "oracle": "Go cohere formatter exact bytes; lowered Node and emitted JavaScript; original Prettier has 42 input-specific binary/minification exceptions plus count check.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M2",
      "M3",
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: format_test.go:137: native ASan/UBSan/LSan: byte 23905: got \"ry proof, so a package one of them refused is never on the registry.\\\\n    needs:\\\\n      [\\\\n        release,\\\\n        swift-on-intel,\\\\n        windows,\\\\n        \", want \"ry proof, so a package one of them refused is never on the registry.\\\\n    needs: [release, swift-on-intel, windows, clean-machine]\\\\n    if: ${{ inputs.publish }\"",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestComposeMatchGo",
      "TestCSTMatchesGo",
      "TestFileDriver family"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P3"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestFormatterMatchesGo",
      "TestComposeMatchGo",
      "TestCSTMatchesGo",
      "TestFileDriver family"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestFormatterMatchesGo)$' > M4-0-matrix.log; format_test.go:137: native ASan/UBSan/LSan: byte 23905: got \"ry proof, so a package one of them refused is never on the registry.\\\\n    needs:\\\\n      [\\\\n        release,\\\\n        swift-on-intel,\\\\n        windows,\\\\n        \", want \"ry proof, so a package one of them refused is never on the registry.\\\\n    needs: [release, swift-on-intel, windows, clean-machine]\\\\n    if: ${{ inputs.publish }\""
  },
  {
    "test": "TestComposeMatchGo",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/compose_test.go",
    "seconds": 23.844,
    "oracle": "Go cohere composition output compared byte-for-byte with native, Node, emitted JavaScript and original yaml library.",
    "oracle_kind": "external-run",
    "kills": [
      "M2"
    ],
    "unique_kills": [],
    "last_proven_fail": "M2: compose_test.go:148: native ASan/UBSan/LSan: byte 9: got \"case 0\\n1.1|0|0|0|00210021=007400610067003a00790061006d006c002e006f00720067002c00320030003\", want \"case 0\\n1.2|0|0|0|00210021=007400610067003a00790061006d006c002e006f00720067002c00320030003\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestFormatterMatchesGo"
    ],
    "mutants_in_matrix": 2,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 46.414,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestFormatterMatchesGo",
      "TestComposeMatchGo",
      "TestCSTMatchesGo",
      "TestFileDriver family"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestComposeMatchGo)$' > M2-1-matrix.log; compose_test.go:148: native ASan/UBSan/LSan: byte 9: got \"case 0\\n1.1|0|0|0|00210021=007400610067003a00790061006d006c002e006f00720067002c00320030003\", want \"case 0\\n1.2|0|0|0|00210021=007400610067003a00790061006d006c002e006f00720067002c00320030003\""
  },
  {
    "test": "TestCSTMatchesGo",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/cst_test.go",
    "seconds": 9.378,
    "oracle": "Go cohere CST serialization compared byte-for-byte with native, Node, emitted JavaScript and original yaml library.",
    "oracle_kind": "external-run",
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1: cst_test.go:78: native ASan/UBSan/LSan: byte 180487: got \"[],newline|30285|14|000a||[]|-|-|-|[]|[],]|[]|0,]|0,]|0,]|0,]|0,]|-|[]|[]\\nlines 1,76,78,168,268,366,467,519,521,618,716,816,883,885,988,1087,1189,1292,1351,1353\", want \"[],newline|30285|14|000a||[]|-|-|-|[]|[],]|[]|0,]|0,]|0,]|0,]|0,]|-|[]|[]\\nlines 0,76,78,168,268,366,467,519,521,618,716,816,883,885,988,1087,1189,1292,1351,1353\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestFileDriver family"
    ],
    "mutants_in_matrix": 1,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": 6.431,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestFormatterMatchesGo",
      "TestComposeMatchGo",
      "TestCSTMatchesGo",
      "TestFileDriver family"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestCSTMatchesGo)$' > M1-2-matrix.log; cst_test.go:78: native ASan/UBSan/LSan: byte 180487: got \"[],newline|30285|14|000a||[]|-|-|-|[]|[],]|[]|0,]|0,]|0,]|0,]|0,]|-|[]|[]\\nlines 1,76,78,168,268,366,467,519,521,618,716,816,883,885,988,1087,1189,1292,1351,1353\", want \"[],newline|30285|14|000a||[]|-|-|-|[]|[],]|[]|0,]|0,]|0,]|0,]|0,]|-|[]|[]\\nlines 0,76,78,168,268,366,467,519,521,618,716,816,883,885,988,1087,1189,1292,1351,1353\""
  },
  {
    "test": "TestComposeMutants",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/compose_test.go",
    "seconds": 30.555,
    "oracle": "Go cohere output and planted port mutants; judged by weakening only the agreement comparison.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: compose_test.go:209: native missed mutant",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestComposeMutants"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestComposeMutants)$' > W1-3-matrix.log; compose_test.go:209: native missed mutant",
    "witness_kills": [
      "W1"
    ]
  },
  {
    "test": "TestCSTMutants",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/cst_test.go",
    "seconds": 8.705,
    "oracle": "Go cohere CST output and planted port mutants; judged by weakening only the agreement comparison.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W2: cst_test.go:143: native missed mutant",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCSTMutants"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestCSTMutants)$' > W2-4-matrix.log; cst_test.go:143: native missed mutant",
    "witness_kills": [
      "W2"
    ]
  },
  {
    "test": "TestFileDriverUnion",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/file_driver_shards_test.go",
    "seconds": 0.582,
    "oracle": "Native control versus Go cohere, then planted stdout disagreements plus shard enumeration assertions; judged by weakening the planted comparison.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W3: file_driver_shards_test.go:201: planted disagreement caught by []",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestFileDriverUnion"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestFileDriverUnion)$' > W3-5-matrix.log; file_driver_shards_test.go:201: planted disagreement caught by []",
    "witness_kills": [
      "W3"
    ]
  },
  {
    "test": "TestFileDriver family",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/file_driver_shards_test.go",
    "seconds": 6.431,
    "oracle": "Go cohere file-mode formatter output compared byte-for-byte against native executable, emitted JavaScript and Node for each shard input.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M3",
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: file_driver_shards_test.go:252: native file 0: byte 23413: got \"ery proof, so a package one of them refused is never on the registry.\\n    needs:\\n      [\\n        release,\\n        swift-on-intel,\\n        windows,\\n        clean\", want \"ery proof, so a package one of them refused is never on the registry.\\n    needs: [release, swift-on-intel, windows, clean-machine]\\n    if: ${{ inputs.publish }}\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestFormatterMatchesGo"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P3"
    ],
    "subsumer_seconds": 46.414,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestFormatterMatchesGo",
      "TestComposeMatchGo",
      "TestCSTMatchesGo",
      "TestFileDriver family"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestFileDriver_000|TestFileDriver_001|TestFileDriver_002|TestFileDriver_003|TestFileDriver_004|TestFileDriver_005|TestFileDriver_006|TestFileDriver_007)$' > M4-6-matrix.log; file_driver_shards_test.go:252: native file 0: byte 23413: got \"ery proof, so a package one of them refused is never on the registry.\\n    needs:\\n      [\\n        release,\\n        swift-on-intel,\\n        windows,\\n        clean\", want \"ery proof, so a package one of them refused is never on the registry.\\n    needs: [release, swift-on-intel, windows, clean-machine]\\n    if: ${{ inputs.publish }}\"",
    "members": [
      "TestFileDriver_000",
      "TestFileDriver_001",
      "TestFileDriver_002",
      "TestFileDriver_003",
      "TestFileDriver_004",
      "TestFileDriver_005",
      "TestFileDriver_006",
      "TestFileDriver_007"
    ],
    "vacuous_subcases": [
      "empty file: unchanged Go oracle ok\\t\\n; P3 native exit 0, empty stdout and stderr (empty-file-probe.json)"
    ]
  },
  {
    "test": "TestBundledParserDifference",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/format_test.go",
    "seconds": 0.102,
    "oracle": "Original yaml and Prettier actually run; output checked against a self-written six-line snapshot of known library differences. No Adamic port is executed.",
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
    "matrix_rows": [],
    "evidence": "Three clean runs passed; no admissible Adamic mutation reaches this external-library-only row."
  },
  {
    "test": "TestFileDriver_Setup",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/format_test.go",
    "seconds": 0.526,
    "oracle": "Suite construction builds and publishes native/Node/Go products; construction break S1 is judged, not production behavior. Parent TestMain preflight child does the real work.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S1: file_driver_shards_test.go:138: build yaml-file-driver-native: open /tmp/u150/cache/S1/2b644a6a5e774b64bd9c0828637f0ad8943a4dd0480099186ae26eaf305aa0ca/format.c: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [
      "P4"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestFileDriver_Setup"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestFileDriver_Setup)$' > S1-8-matrix.log; file_driver_shards_test.go:138: build yaml-file-driver-native: open /tmp/u150/cache/S1/2b644a6a5e774b64bd9c0828637f0ad8943a4dd0480099186ae26eaf305aa0ca/format.c: no such file or directory",
    "construction_kills": [
      "S1"
    ]
  }
]
```

| ID | Origin file:line | Change | Observed failed rows |
|---|---|---|---|
| M1 | stage1/cohere/yaml/cstParser.ts:799 | this.lineStarts.push(0); -> this.lineStarts.push(1); | TestCSTMatchesGo, TestFormatterMatchesGo, TestFileDriver family |
| M2 | stage1/cohere/yaml/directives.ts:79 | version = '1.2'; -> version = '1.1'; | TestComposeMatchGo, TestFormatterMatchesGo |
| M3 | stage1/cohere/yaml/format.ts:36 | result.text = (bom ? '\ufeff' : '') + printer.format(root); -> result.text = (bom ? '' : '') + printer.format(root); | TestFormatterMatchesGo, TestFileDriver family |
| M4 | stage1/cohere/yaml/layout.ts:202 | 80 - this.column -> 40 - this.column | TestFormatterMatchesGo, TestFileDriver family |
| S1 | stage1/cohere/yaml/file_driver_shards_test.go:130 | filepath.Join(directory, "format.c") -> filepath.Join(directory, "missing.c") | TestFileDriver_Setup |
| W1 | stage1/cohere/yaml/compose_test.go:208 | Weaken equality to always agree within the witness; assertions unchanged | TestComposeMutants |
| W2 | stage1/cohere/yaml/cst_test.go:142 | Weaken equality to always agree within the witness; assertions unchanged | TestCSTMutants |
| W3 | stage1/cohere/yaml/file_driver_shards_test.go:195 | Always agree on planted output; native precondition comparison unchanged | TestFileDriverUnion |

Survivors: None among the four production mutants in the bounded matrix.

The brief and measurement limits

The brief says eight rows, but its extra-assertion rule gives nine: TestFileDriverUnion checks a planted disagreement that the eight shards do not. It remains a separate witness. All names stayed in the listed files. The supplied historical hash was not current origin/main; the audit uses the fetched commit named above.

TestMain runs file-driver preparation in a child process before the parent test alarm starts. Its 38.76-second cold setup plus the package run exceeded the outer 120-second limit without an earlier failing test. This is a cooked baseline, not a red baseline. Matrices were narrowed to the selected rows that import each changed module, as recorded in plan.json. Unselected rows and all other packages remain unknown. Local subsumption rests on only four production mutants, not a deletion recommendation.

Timings use the Go test binary package pass.Elapsed line, not subprocess wall time. File-family timing runs all eight members as one row. The setup parent normally returns immediately because TestMain already prepared products, so the package line includes its child work. Raw command wall times are separate. No row skipped with the exact yaml@2.9.0 and prettier@3.9.6 libraries enabled.

TestBundledParserDifference executes only the external original libraries and checks their known six-line difference snapshot. An Adamic production mutant cannot reach it; mutating those libraries would mutate the oracle. Its verdict is cannot-judge and its probe status is null. The formatter oracle separately checks exact expected outputs and accepts 42 individually identified original-library differences, then checks the total; it does not accept arbitrary same-count failures.

Every switched matrix command sets ADAMIC_MUTANT to its ID, writes that ID to /tmp/u150/selector, enables ADAMIC_YAML_LIBRARY=/tmp/u150/library, and selects ADAMIC_BUILD_CACHE_DIR=/tmp/u150/cache/switched, except S1 and P4 use fresh ID-specific caches. Exact environment and command construction are in mutate.py. The function inventory is a conservative transitive module/function superset, not a dynamic proof of every call or every anonymous callback. Four production mutants were selected from that inventory before running any mutant. This follows the port-specific four-mutant cap rather than the general three-per-row aim. Runtime file selection allows one native build per product; M0 checks all four functional rows before the matrix. Compiler source was not changed, so the shared switched cache holds exactly the same executable across selections. Each standalone native validation has its own cache.

P1-P3 suppress the actual port entry stdout, rather than a preparation helper. P4 returns a nil setup product and may abort the child; its row was run alone. Probe failures are separate from production kills. The witness checks weaken only the comparison, leaving assertions and planted inputs intact. Setup S1 writes the generated C under the wrong name, leaving the later read intact. The first standalone empty-entry guard used if(false), which TypeScript rejected because unreachable-body union narrowing does not apply. The saved probes instead delete the entry body and successfully rebuild. All standalone diffs contain no runtime selector and apply independently to the starting commit.

The empty-file subcase was independently replayed and passes P3, as recorded in empty-file-probe.json; the family still rejects P3 on nonempty files. Other positive subcases after an early fatal probe assertion remain unknown. No claims are made about dynamic call coverage, package-wide uniqueness, or repository-wide uniqueness. No test or oracle was edited for production mutants. The construction and witness edits are the explicit exceptions in the brief.

Tool setup: warm env, no cloud/setup.sh run; nproc=5. npm ci and exact optional-library installation succeeded; their individual wall times were not instrumented. Whole-package backstop was 120 seconds. Three clean measurement rounds total 425.9 shell wall seconds. Matrix commands total 523.84 shell wall seconds, including compilation and external runners. Per-run logs and build timings are in matrix-timings.json and standalone-validation.json. Final restoration is recorded in restored-baseline.log.
Standalone build/vet total: 113.59 seconds.
