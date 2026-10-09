Unit u086, starting commit 12e77e8972a2e606cab6db05d84f428246a85339.
15 requested rows present; JSX agreement expands to one family of 33 listed tests.
All requested clean baselines pass; whole package exceeds its 90 second budget.
Bounded verdicts: 4 sacred, 3 subsumed, 3 witness, 4 setup-check, 1 cannot-judge.
Seven production mutants caught; four entry probes caught; no production survivors.

```json
[
  {
    "test": "TestPostfixValueGap",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/gaps_test.go",
    "seconds": 0.098,
    "oracle": "Node runs fixture and verifies 7,1; self-written NotYet substring. M05 proves article wording, not postfix implementation.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M05"
    ],
    "unique_kills": [
      "M05"
    ],
    "last_proven_fail": "M05: gaps_test.go:33: postfix value gap changed: /workspace/adamic/stage1/cohere/estree/gaps/postfixValue.ts:3:23: stage 0 can't lower an PostfixUnaryExpression yet",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestInterfaceDefaultGap",
      "TestInterfaceTypeMethodGap",
      "TestMethodReplacementGap",
      "TestPostfixValueGap"
    ],
    "evidence": "ADAMIC_MUTANT=M05 ADAMIC_BUILD_CACHE_DIR=/tmp/u086/cache/M05 ADAMIC_ESTREE_LIBRARY=/tmp/u086/library timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestPostfixValueGap|TestMethodReplacementGap|TestInterfaceDefaultGap|TestInterfaceTypeMethodGap)$' > M05.log; gaps_test.go:33: postfix value gap changed: /workspace/adamic/stage1/cohere/estree/gaps/postfixValue.ts:3:23: stage 0 can't lower an PostfixUnaryExpression yet",
    "members": [
      "TestPostfixValueGap"
    ]
  },
  {
    "test": "TestMethodReplacementGap",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/gaps_test.go",
    "seconds": 0.114,
    "oracle": "Node runs replacement/original fixture; self-written Refused substring requires arrow guidance. M07 proves diagnostic guidance.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M07"
    ],
    "unique_kills": [
      "M07"
    ],
    "last_proven_fail": "M07: gaps_test.go:53: method replacement gap changed: /workspace/adamic/stage1/cohere/estree/gaps/methodReplacement.ts:11:1: Adamic 0.1 refuses a method read as a value (next would lose its object, and this with it)",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestInterfaceDefaultGap",
      "TestInterfaceTypeMethodGap",
      "TestMethodReplacementGap",
      "TestPostfixValueGap"
    ],
    "evidence": "ADAMIC_MUTANT=M07 ADAMIC_BUILD_CACHE_DIR=/tmp/u086/cache/M07 ADAMIC_ESTREE_LIBRARY=/tmp/u086/library timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestPostfixValueGap|TestMethodReplacementGap|TestInterfaceDefaultGap|TestInterfaceTypeMethodGap)$' > M07.log; gaps_test.go:53: method replacement gap changed: /workspace/adamic/stage1/cohere/estree/gaps/methodReplacement.ts:11:1: Adamic 0.1 refuses a method read as a value (next would lose its object, and this with it)",
    "members": [
      "TestMethodReplacementGap"
    ]
  },
  {
    "test": "TestRawInputGap",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/gaps_test.go",
    "seconds": 1.001,
    "oracle": "Node source runtime is the expected reader/UTF-8 output; native and emitted JS must match byte-for-byte. Go cohere must distinguish malformed bytes.",
    "oracle_kind": "external-run",
    "kills": [
      "M03"
    ],
    "unique_kills": [],
    "last_proven_fail": "M03: gaps_test.go:86: reader changed: \"8:9://\ufffd\\nx;\\n\" != \"8:8://\ufffd\\nx;\\n\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestJSXAgreement family"
    ],
    "mutants_in_matrix": 1,
    "probe_kills": [
      "PUtf8"
    ],
    "subsumer_seconds": 3.998,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestJSXAgreement family",
      "TestRawInputGap"
    ],
    "evidence": "ADAMIC_MUTANT=M03 ADAMIC_BUILD_CACHE_DIR=/tmp/u086/cache/switch ADAMIC_ESTREE_LIBRARY=/tmp/u086/library timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestRawInputGap|TestJSXAgreement(_[0-9]{3})?)$' (selector file /tmp/u086/selector contains M03) > M03.log; gaps_test.go:86: reader changed: \"8:9://\ufffd\\nx;\\n\" != \"8:8://\ufffd\\nx;\\n\"",
    "members": [
      "TestRawInputGap"
    ],
    "subsumption_basis_mutants": 1
  },
  {
    "test": "TestParserRecoveryGap",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/gaps_test.go",
    "seconds": 15.06,
    "oracle": "Self-written requirement: source Node and native parser must each exceed 1s. Only timeout checked; any unrelated hang could pass.",
    "oracle_kind": "self",
    "kills": [
      "M02"
    ],
    "unique_kills": [
      "M02"
    ],
    "last_proven_fail": "M02: gaps_test.go:113: expected bounded external timeout; got exit status 70",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [
      "PParser"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestParserRecoveryGap"
    ],
    "evidence": "ADAMIC_MUTANT=M02 ADAMIC_BUILD_CACHE_DIR=/tmp/u086/cache/switch ADAMIC_ESTREE_LIBRARY=/tmp/u086/library timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestParserRecoveryGap)$' (selector file /tmp/u086/selector contains M02) > M02.log; gaps_test.go:113: expected bounded external timeout; got exit status 70",
    "members": [
      "TestParserRecoveryGap"
    ]
  },
  {
    "test": "TestInterfaceDefaultGap",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/gaps_test.go",
    "seconds": 0.107,
    "oracle": "Node actually prints 5; exact self-written NotYet path:10:12 and prototype-erasure label.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M06"
    ],
    "unique_kills": [],
    "last_proven_fail": "M06: gaps_test.go:135: recorded lowering gap changed: /workspace/adamic/stage1/cohere/estree/gaps/interfaceDefault.ts:10:12:0: stage 0 can't lower a class method through a view that erases its prototype origin yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestInterfaceTypeMethodGap"
    ],
    "mutants_in_matrix": 3,
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": 0.416,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestInterfaceDefaultGap",
      "TestInterfaceTypeMethodGap",
      "TestMethodReplacementGap",
      "TestPostfixValueGap"
    ],
    "evidence": "ADAMIC_MUTANT=M06 ADAMIC_BUILD_CACHE_DIR=/tmp/u086/cache/M06 ADAMIC_ESTREE_LIBRARY=/tmp/u086/library timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestPostfixValueGap|TestMethodReplacementGap|TestInterfaceDefaultGap|TestInterfaceTypeMethodGap)$' > M06.log; gaps_test.go:135: recorded lowering gap changed: /workspace/adamic/stage1/cohere/estree/gaps/interfaceDefault.ts:10:12:0: stage 0 can't lower a class method through a view that erases its prototype origin yet",
    "members": [
      "TestInterfaceDefaultGap"
    ],
    "subsumption_basis_mutants": 1
  },
  {
    "test": "TestInterfaceTypeMethodGap",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/interface_type_test.go",
    "seconds": 0.416,
    "oracle": "Node actually prints 1; exact self-written NotYet path:10:12 and label; default-free callback control must lower and run as 1.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M06"
    ],
    "unique_kills": [],
    "last_proven_fail": "M06: interface_type_test.go:33: recorded lowering gap changed: /workspace/adamic/stage1/cohere/estree/gaps/interfaceTypeMethod.ts:10:12:0: stage 0 can't lower a class method through a view that erases its prototype origin yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestInterfaceDefaultGap"
    ],
    "mutants_in_matrix": 3,
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": 0.107,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestInterfaceDefaultGap",
      "TestInterfaceTypeMethodGap",
      "TestMethodReplacementGap",
      "TestPostfixValueGap"
    ],
    "evidence": "ADAMIC_MUTANT=M06 ADAMIC_BUILD_CACHE_DIR=/tmp/u086/cache/M06 ADAMIC_ESTREE_LIBRARY=/tmp/u086/library timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestPostfixValueGap|TestMethodReplacementGap|TestInterfaceDefaultGap|TestInterfaceTypeMethodGap)$' > M06.log; interface_type_test.go:33: recorded lowering gap changed: /workspace/adamic/stage1/cohere/estree/gaps/interfaceTypeMethod.ts:10:12:0: stage 0 can't lower a class method through a view that erases its prototype origin yet",
    "members": [
      "TestInterfaceTypeMethodGap"
    ],
    "subsumption_basis_mutants": 1
  },
  {
    "test": "TestJSXAgreement family",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/jsx_agreement_product_shards_test.go",
    "seconds": 3.998,
    "oracle": "Go cohere ESTree canonical bytes compared with source Node, sanitized native and emitted JS port output.",
    "oracle_kind": "external-run",
    "kills": [
      "M01",
      "M03",
      "M04"
    ],
    "unique_kills": [
      "M01",
      "M04"
    ],
    "last_proven_fail": "M04: jsx_agreement_product_shards_test.go:79: stage1/cohere/estree/jsx_test.go:jsx:source-9ba96f7b7696a8abb1f81450cf8157abfb23f6ab2fec32fe0c44048141058ebc Node: line 6: Go \"2 JSXElement 0 24 0 0 0 0 24 0\", port \"3 JSXElement 0 24 0 0 0 0 24 0\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "probe_kills": [
      "PMain"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestJSXAgreement family",
      "TestRawInputGap"
    ],
    "evidence": "ADAMIC_MUTANT=M04 ADAMIC_BUILD_CACHE_DIR=/tmp/u086/cache/switch ADAMIC_ESTREE_LIBRARY=/tmp/u086/library timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestJSXAgreement(_[0-9]{3})?)$' (selector file /tmp/u086/selector contains M04) > M04.log; jsx_agreement_product_shards_test.go:79: stage1/cohere/estree/jsx_test.go:jsx:source-9ba96f7b7696a8abb1f81450cf8157abfb23f6ab2fec32fe0c44048141058ebc Node: line 6: Go \"2 JSXElement 0 24 0 0 0 0 24 0\", port \"3 JSXElement 0 24 0 0 0 0 24 0\"",
    "members": [
      "TestJSXAgreement",
      "TestJSXAgreement_000",
      "TestJSXAgreement_001",
      "TestJSXAgreement_002",
      "TestJSXAgreement_003",
      "TestJSXAgreement_004",
      "TestJSXAgreement_005",
      "TestJSXAgreement_006",
      "TestJSXAgreement_007",
      "TestJSXAgreement_008",
      "TestJSXAgreement_009",
      "TestJSXAgreement_010",
      "TestJSXAgreement_011",
      "TestJSXAgreement_012",
      "TestJSXAgreement_013",
      "TestJSXAgreement_014",
      "TestJSXAgreement_015",
      "TestJSXAgreement_016",
      "TestJSXAgreement_017",
      "TestJSXAgreement_018",
      "TestJSXAgreement_019",
      "TestJSXAgreement_020",
      "TestJSXAgreement_021",
      "TestJSXAgreement_022",
      "TestJSXAgreement_023",
      "TestJSXAgreement_024",
      "TestJSXAgreement_025",
      "TestJSXAgreement_026",
      "TestJSXAgreement_027",
      "TestJSXAgreement_028",
      "TestJSXAgreement_029",
      "TestJSXAgreement_030",
      "TestJSXAgreement_031"
    ],
    "vacuous_subcases": [
      "TestJSXAgreement_016",
      "TestJSXAgreement_004",
      "TestJSXAgreement_005",
      "TestJSXAgreement_007",
      "TestJSXAgreement_006",
      "TestJSXAgreement_015",
      "TestJSXAgreement_013",
      "TestJSXAgreement_014",
      "TestJSXAgreement_031",
      "TestJSXAgreement_030",
      "TestJSXAgreement",
      "TestJSXAgreement_028",
      "TestJSXAgreement_026",
      "TestJSXAgreement_002",
      "TestJSXAgreement_003",
      "TestJSXAgreement_020",
      "TestJSXAgreement_022",
      "TestJSXAgreement_021",
      "TestJSXAgreement_010",
      "TestJSXAgreement_011",
      "TestJSXAgreement_001",
      "TestJSXAgreement_000",
      "TestJSXAgreement_018",
      "TestJSXAgreement_019",
      "TestJSXAgreement_009",
      "TestJSXAgreement_017"
    ],
    "vacuous_subcases_note": "Union and empty assigned shards pass; all nine positive cases belong to seven failing nonempty shards."
  },
  {
    "test": "TestJSXAgreementPlantedDisagreement",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/jsx_agreement_product_shards_test.go",
    "seconds": 0.209,
    "oracle": "Self synthetic agree/disagree values; exactly one actual top-level shard must reject the planted disagreement.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: jsx_agreement_product_shards_test.go:119: planted disagreement caught by 0 shards, want one",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestJSXAgreementPlantedDisagreement",
      "TestJSXMutant",
      "TestJSXMutantPlantedSurvivor"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u086/cache/W01 ADAMIC_ESTREE_LIBRARY=/tmp/u086/library timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestJSXAgreementPlantedDisagreement|TestJSXMutant|TestJSXMutantPlantedSurvivor)$' > W01.log; jsx_agreement_product_shards_test.go:119: planted disagreement caught by 0 shards, want one",
    "members": [
      "TestJSXAgreementPlantedDisagreement"
    ],
    "construction_or_witness_kills": [
      "W01"
    ]
  },
  {
    "test": "TestJSXOriginalLibraries",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/jsx_test.go",
    "seconds": 0.917,
    "oracle": "Go cohere compared with installed @typescript-eslint/typescript-estree 8.65.0, TypeScript 6.0.3 and Prettier 3.9.6 through Node; exact known differences are self-written. This row runs no Adamic product.",
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
    "evidence": "Clean baseline and three measurements; row does not invoke Adamic, so oracle mutations forbidden > baseline-small.log",
    "members": [
      "TestJSXOriginalLibraries"
    ]
  },
  {
    "test": "TestJSXMutant",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/jsx_test.go",
    "seconds": 1.423,
    "oracle": "Go cohere canonical bytes; built-in port mutant must disagree with Node/native/emitted JS. Only W01 weakening decides witness verdict.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: jsx_test.go:126: stage1/cohere/estree/jsx_test.go:jsx-mutant:source-711191cc7aae173eac0a328f14c50d5982847677cc17e1e8879acbce4cf26fff Node: mutant survived",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestJSXAgreementPlantedDisagreement",
      "TestJSXMutant",
      "TestJSXMutantPlantedSurvivor"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u086/cache/W01 ADAMIC_ESTREE_LIBRARY=/tmp/u086/library timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestJSXAgreementPlantedDisagreement|TestJSXMutant|TestJSXMutantPlantedSurvivor)$' > W01.log; jsx_test.go:126: stage1/cohere/estree/jsx_test.go:jsx-mutant:source-711191cc7aae173eac0a328f14c50d5982847677cc17e1e8879acbce4cf26fff Node: mutant survived",
    "members": [
      "TestJSXMutant"
    ],
    "construction_or_witness_kills": [
      "W01"
    ]
  },
  {
    "test": "TestJSXMutantPlantedSurvivor",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/jsx_test.go",
    "seconds": 0.016,
    "oracle": "Self synthetic agreement/disagreement; exactly one planned survivor must fail its leaf.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: jsx_test.go:134: planted failure must be caught only by TestJSXMutantPlantedSurvivor/shard-010: exit=exit status 1",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestJSXAgreementPlantedDisagreement",
      "TestJSXMutant",
      "TestJSXMutantPlantedSurvivor"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u086/cache/W01 ADAMIC_ESTREE_LIBRARY=/tmp/u086/library timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestJSXAgreementPlantedDisagreement|TestJSXMutant|TestJSXMutantPlantedSurvivor)$' > W01.log; jsx_test.go:134: planted failure must be caught only by TestJSXMutantPlantedSurvivor/shard-010: exit=exit status 1",
    "members": [
      "TestJSXMutantPlantedSurvivor"
    ],
    "construction_or_witness_kills": [
      "W01"
    ]
  },
  {
    "test": "TestMiscShardUnionRejectsInvalidEnumeration",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/misc_shards_test.go",
    "seconds": 0.007,
    "oracle": "Self invalid-enumeration rejection cases.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "H01: misc_shards_test.go:222: invalid enumeration accepted: {ids:[] count:2}",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMiscShardUnionRejectsInvalidEnumeration"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u086/cache/H01 ADAMIC_ESTREE_LIBRARY=/tmp/u086/library timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestMiscShardUnionRejectsInvalidEnumeration)$' > H01.log; misc_shards_test.go:222: invalid enumeration accepted: {ids:[] count:2}",
    "members": [
      "TestMiscShardUnionRejectsInvalidEnumeration"
    ],
    "construction_or_witness_kills": [
      "H01"
    ]
  },
  {
    "test": "TestMiscShardGrowthKeepsAssignments",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/misc_shards_test.go",
    "seconds": 0.008,
    "oracle": "Self shard assignment equality before/after insertion and reordering.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "H02: misc_shards_test.go:418: case stage1/cohere/estree/exports_test.go:export:0 moved from shard-002 to shard-001",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestMiscShardGrowthKeepsAssignments"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u086/cache/H02 ADAMIC_ESTREE_LIBRARY=/tmp/u086/library timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestMiscShardGrowthKeepsAssignments)$' > H02.log; misc_shards_test.go:418: case stage1/cohere/estree/exports_test.go:export:0 moved from shard-002 to shard-001",
    "members": [
      "TestMiscShardGrowthKeepsAssignments"
    ],
    "construction_or_witness_kills": [
      "H02"
    ]
  },
  {
    "test": "TestRecoveryCacheInputKeys",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/recovery_cache_inputs_test.go",
    "seconds": 0.598,
    "oracle": "Self cache key equal for copied unchanged source, different for changed mutable source.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "H03: recovery_cache_inputs_test.go:32: changed mutant source did not change the product key",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestRecoveryCacheInputKeys"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u086/cache/H03 ADAMIC_ESTREE_LIBRARY=/tmp/u086/library timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestRecoveryCacheInputKeys)$' > H03.log; recovery_cache_inputs_test.go:32: changed mutant source did not change the product key",
    "members": [
      "TestRecoveryCacheInputKeys"
    ],
    "construction_or_witness_kills": [
      "H03"
    ]
  },
  {
    "test": "TestRecoveryNativeRecipe",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/recovery_cache_test.go",
    "seconds": 0.056,
    "oracle": "Independent product bytes must agree; native fixtures must trigger Clang undefined/address/leak sanitizer diagnostics.",
    "oracle_kind": [
      "self",
      "external-run"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "H04: recovery_shards_test.go:38: setup wall=5.931330s builds=5.920626s without-builds=0.010704s",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestRecoveryNativeRecipe"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u086/cache/H04 ADAMIC_ESTREE_LIBRARY=/tmp/u086/library timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestRecoveryNativeRecipe)$' > H04.log; recovery_shards_test.go:38: setup wall=5.931330s builds=5.920626s without-builds=0.010704s",
    "members": [
      "TestRecoveryNativeRecipe"
    ],
    "construction_or_witness_kills": [
      "H04"
    ]
  }
]
```

| ID | File:line at starting commit | Change | Failed rows |
|---|---|---|---|
| M01 (production) | stage1/cohere/estree/jsxConvert.ts:20 | 'JSXIdentifier' → 'JSXName' | TestJSXAgreement family |
| M02 (production) | stage1/typescript/parser/parser.ts:736 | const members: number[] = [];         while(this.kind() !== 'CloseBraceToken') → const members: number[] = [];         while(this.kind() === 'CloseBraceToken') | TestParserRecoveryGap |
| M03 (production) | internal/native/runtime/utf8.c:15 | return (double)text->length; → return (double)text->length + 1; | TestRawInputGap, TestJSXAgreement family |
| M04 (production) | stage1/cohere/estree/protocol.ts:66 | new DumpFrame(property.value.node, level + 1, -1) → new DumpFrame(property.value.node, level + 2, -1) | TestJSXAgreement family |
| M05 (production) | internal/lower/diagnostics.go:42 | return "a " + name → return "an " + name | TestPostfixValueGap |
| M06 (production) | internal/lower/diagnostics.go:33 | Where: l.program.Where(node), What: what → Where: l.program.Where(node) + ":0", What: what | TestInterfaceDefaultGap, TestInterfaceTypeMethodGap |
| M07 (production) | internal/lower/diagnostics.go:29 | return fmt.Sprintf("%s: Adamic 0.1 refuses %s; %s", r.Where, r.What, r.Fix) → return fmt.Sprintf("%s: Adamic 0.1 refuses %s", r.Where, r.What) | TestMethodReplacementGap |
| H01 (setup) | stage1/cohere/estree/misc_shards_test.go:32 | return nil, fmt.Errorf("%d cases cannot enumerate %d shards", len(ids), count) → return nil, nil | TestMiscShardUnionRejectsInvalidEnumeration |
| H02 (setup) | stage1/cohere/estree/misc_shards_test.go:41 | shard := miscShard(id, count) → shard := i % count | TestMiscShardGrowthKeepsAssignments |
| H03 (setup) | stage1/cohere/estree/recovery_cache_test.go:101 | sha256.Sum256([]byte(normalized)) → sha256.Sum256([]byte(normalized[:0])) | TestRecoveryCacheInputKeys |
| H04 (setup) | stage1/cohere/estree/recovery_cache_test.go:162 | args := native.Flags(native.Options{Sanitize: true}) → args := native.Flags(native.Options{Sanitize: false}) | TestRecoveryNativeRecipe |
| W01 (witness) | stage1/cohere/estree/misc_shards_test.go:114 | diff := firstDifference(want, got) → diff := "" | TestJSXAgreementPlantedDisagreement, TestJSXMutant, TestJSXMutantPlantedSurvivor |

Survivors: none among M01 through M07. H01 through H04 are allowed construction breaks, W01 is the allowed witness weakening, and PMain/PParser/PUtf8/PLower are empty-answer probes. None contributes to production kills or uniqueness.

Scope, oracle and measurement limitations:

- The brief refers to commit 8de93800f4. The required fetch/start selected current origin/main 12e77e8972a2e606cab6db05d84f428246a85339. Every saved standalone diff and production location uses that actual starting commit.
- All 15 names exist in test-list.log. JSXAgreement and its planted-disagreement witness moved to jsx_agreement_product_shards_test.go. Agreement's union plus 32 wrappers form one family. The union alone does no case execution, so an exact-name timing would have measured only construction. The family timing runs all 33 members. No requested name vanished.
- Whole-package baseline timed out at 90.032 binary seconds during AcceptanceDiagnostics after AcceptanceGrammar passed in 52.98 seconds. No assertion failure preceded the timeout. The clean requested scope was then executed in bounded runs and passed, including installed optional original libraries. A restored single scoped run also passed. Kills outside each listed matrix set are unknown. Sacred and unique_kills mean uniqueness only within the bounded observed matrix, never proved package-wide or repository-wide uniqueness.
- callers.txt records preparation and checker callers. inventory.txt deliberately overapproximates TypeScript/parser functions and lowering functions; lower-coverage.out/lower-functions.txt measure the clean diagnostic/control path. The inventory is not an exact transitive TypeScript call graph. Unreached declarations remain in that source inventory. No whole-repository replay was performed.
- TestJSXOriginalLibraries compares Go cohere with upstream libraries and never invokes Adamic. Mutating either participant would violate the instruction to preserve the oracle. It is cannot-judge for this Adamic audit, not untrue. Optional libraries were installed and enabled using ADAMIC_ESTREE_LIBRARY=/tmp/u086/library; package and lock files are included. No requested row skipped. Whole-package rows after the timeout have unknown skip status.
- ParserRecoveryGap expects both source Node and native parser nontermination within its one-second subprocess deadlines. A different hang can satisfy that expectation. M02 changes the loop condition and produces exit status 70, so the catch proves the timeout check fires, not correctness of parser recovery. PParser returns the empty numeric answer and the row rejects completion.
- The three lowering mutants check diagnostic grammar, source location and suggested guidance. Their catches establish those exact assertions; they do not establish correct lowering of the refused features. Node fixture executions are independent behavior checks, while NotYet/Refused strings are self-written labels, not externally sourced authorities.
- RawInputGap's independent oracle is source Node runtime behavior. Its mutated code is native utf8Length, so PUtf8 probes that native entry. Mutating the shared rawInput.ts fixture would also alter this row's source Node expectation and would violate oracle isolation. Go cohere is used only as a malformed-input distinction control.
- The JSX family compares full canonical bytes. The empty-answer probe fails every nonempty shard. Its union and 25 empty shards still pass, listed in vacuous_subcases; they have no assigned positive inputs. The family is not vacuous. The lower probe aborts TypeMethodGap at its negative assertion, so the later positive control is not independently judged by that probe.
- JSXMutant and both planted proof tests are witnesses. Their verdicts come solely from W01 forcing miscCompare to report agreement. Their production-mutant preconditions do not count. The four construction rows are judged solely by H01 through H04, not by production mutations.
- Subsumption rests on one killed production mutant per row. RawInputGap catches only M03, also caught by JSX agreement. InterfaceDefaultGap and InterfaceTypeMethodGap both catch only M06 in their three-mutant matrix. Their medians are both reported; the faster row to retain is a defender decision. No deletion is proposed.
- The native rebuild cap limits the plan to four native product mutants, spread across JSX conversion, parser recovery, UTF-8 length and protocol traversal. Three additional direct lowering-diagnostic mutants require Go rebuilds only. This is seven mutants rather than the suggested three per row. Construction/witness changes and empty probes are not counted toward that production menu.
- Every native standalone mutant was built through the port/compiler with --sanitize and ADAMIC_NATIVE_SPLIT=1. The latter switches splitting and is not treated as a cache key. Every standalone Go mutant passed go vet ./internal/lower/. Harness changes passed go vet ./stage1/cohere/estree/. All standalone probes also built or vetted. git apply --check passed for every standalone diff after restoration.
- Instrumented TypeScript selects an existing alternative from /tmp/u086/selector at process start, allowing one source build per product. The C and Go alternatives use ADAMIC_MUTANT. M05, M06 and M07 each receive an independent ADAMIC_BUILD_CACHE_DIR; native port selector runs reuse compiled identical instrumented source. The full switch.diff is reproduction evidence, not a production change. All instrumentation was removed before the restored baseline.
- A native subprocess emits “adamic: panic” under M03. It does not panic the Go test binary: every selected top-level member produces a terminal pass/fail event. No Go test-binary panic forced standalone reruns.
- /usr/bin/time was absent; the first timing attempt exited 127. Python monotonic timing replaced it. The first failed attempt did not install stage3 dependencies; npm ci was then completed before the baseline test binary ran. Logs remain to show the mistake. Binary elapsed medians come from Go's own package pass lines, not shell time. Three separate -count=1 runs were used, with native caches warm. ParserRecoveryGap rebuilds through its uncached build helper, explaining its 15.060-second median.

Timing and uncovered work:

Warm tool setup was skipped (0 seconds); nproc is 5. stage3/api npm ci took 0.403 seconds; optional library installation and npm ci took 2.449 and 0.501 seconds. Native standalone mutant builds: validate-M01 43.21s, validate-M02 13.847s, validate-M03 10.388s, validate-M04 22.032s. The instrumented clean JSX product run took 52.888 shell seconds, parser 16.536, and raw-input 8.110, including construction/builds. Detailed wall command times and native build logs are recorded, rather than attributed entirely to test execution. The initial bounded clean command times were parser 17.064 seconds, JSX witness 40.533, and JSX family 46.529; the family binary was 44.912 seconds cold. Standalone probe build times are in probe-builds.json. The 45 median measurement commands total 150.465 shell seconds. Production matrix commands total 38.925 seconds; probe commands 25.426 seconds. Setup/witness command times are separately in mutant-commands.json.

Not covered: whole-package completion, kills outside the bounded matrix, repo-wide uniqueness, exhaustive mutant menus or an exact TypeScript transitive call graph. TestJSXOriginalLibraries has no authorized Adamic mutant. No requested row is omitted. No production source change is retained. No pull request is opened.

Restored scoped binary: 23.478 seconds, all 47 member tests passed and no skips.
