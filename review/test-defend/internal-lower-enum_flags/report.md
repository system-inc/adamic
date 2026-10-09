Seven rows defended by unique production catches; five remain cannot-judge under the seven final-mutant cap.
Baseline passed; every requested row still exists; 36 additional current top-level rows were included.
Source and tests restored unchanged; evidence prepared for test-defend/internal-lower-enum_flags.

Starting origin/main: 76c59c81e8617cea1892a01841494895927a712c

```json
[
  {
    "test": "TestFlagEnumsOpen",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestNumericEnumLiteralPromises"
    ],
    "defense": "defended",
    "unique_mutant": "D7 internal/lower/assignments.go:14",
    "attempts": [
      {
        "mutant": "D7",
        "file_line": "internal/lower/assignments.go:14",
        "change": "change option: flag-shaped enum compound left shift becomes right shift",
        "rows_failed": [
          "TestFlagEnumsOpen"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/enum-defend/cache/D7 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > D7.log 2>&1; enum_flags_test.go:72: JavaScript backend stdout = \"0\\n\", source Node = \"2\\n\"",
    "code_under_test": "enumRefusal, enumAssignable, assignment, enumConstant and enumFields",
    "oracle": "Node source versus lowered JavaScript output for positive cases; self-written Refused type for member promises",
    "coverage_exclusive_blocks": 1225,
    "finding": "Assignment coverage exclusive to literal-promises subsumer; flag-shaped compound shift update is a semantic distinction. The unused reason column still is not asserted.",
    "rows_passed_file": "D7-passed-rows.json"
  },
  {
    "test": "TestFlagEnumsDomain",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestFlagEnumInlineIteration"
    ],
    "defense": "cannot-judge",
    "unique_mutant": null,
    "attempts": [],
    "evidence": "go test -json -count=1 -timeout 90s -coverpkg=github.com/system-inc/adamic/internal/lower -coverprofile=TestFlagEnumsDomain.cover ./internal/lower/ -run ^TestFlagEnumsDomain$; 498 exclusive blocks versus TestFlagEnumInlineIteration. No aimed mutant allocated within the seven final-mutant cap; no three-attempt negative verdict claimed.",
    "code_under_test": "flagDomainSeen, assignment/combine, enumFields and container lowering",
    "oracle": "Node source versus lowered JavaScript output",
    "coverage_exclusive_blocks": 498,
    "finding": "Runtime bitwise/container effects are checked. The row does not assert the internal flagDomain classification; TestEnumFlagProofsWithoutObservableLoweringEffect does. Its domain name should not be taken as a direct proof-algorithm assertion."
  },
  {
    "test": "TestEnumNeverDefault",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestNumericEnumNeverProof"
    ],
    "defense": "cannot-judge",
    "unique_mutant": null,
    "attempts": [],
    "evidence": "go test -json -count=1 -timeout 90s -coverpkg=github.com/system-inc/adamic/internal/lower -coverprofile=TestEnumNeverDefault.cover ./internal/lower/ -run ^TestEnumNeverDefault$; 178 exclusive blocks versus TestNumericEnumNeverProof. No aimed mutant allocated within the seven final-mutant cap; no three-attempt negative verdict claimed.",
    "code_under_test": "enumDefaultUnreachable, enumNeverCheck and switch lowering",
    "oracle": "Node source versus lowered JavaScript output",
    "coverage_exclusive_blocks": 178,
    "finding": "Calls only declared Red/Green/Blue cases. It neither executes an out-of-enum default nor asserts an inserted never guard, despite the default-focused name."
  },
  {
    "test": "TestFlagEnumLiteralSpellings",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestFlagEnumInlineIteration"
    ],
    "defense": "defended",
    "unique_mutant": "D3 internal/lower/enums.go:160",
    "attempts": [
      {
        "mutant": "D3",
        "file_line": "internal/lower/enums.go:160",
        "change": "change constant by one for hexadecimal numeric enum initializers",
        "rows_failed": [
          "TestFlagEnumLiteralSpellings"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/enum-defend/cache/D3-source-spelling timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > D3.log 2>&1; enum_flags_test.go:108: JavaScript backend stdout = \"1/1/1073741824/1073741825/High\\n\", source Node = \"0/1/1073741824/1073741825/High\\n\"",
    "code_under_test": "enumConstant and enumFields",
    "oracle": "Node source versus lowered JavaScript output",
    "coverage_exclusive_blocks": 59,
    "finding": "Shared enumConstant lines with subsumer, but hexadecimal direct initializers distinguish the input. Output values and reverse names now checked.",
    "rows_passed_file": "D3-passed-rows.json"
  },
  {
    "test": "TestFlagEnumMemberAliases",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestFlagEnumInlineIteration"
    ],
    "defense": "defended",
    "unique_mutant": "D4 internal/lower/enums.go:160",
    "attempts": [
      {
        "mutant": "D4",
        "file_line": "internal/lower/enums.go:160",
        "change": "change constant by one for a qualified same-enum member alias",
        "rows_failed": [
          "TestFlagEnumMemberAliases"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/enum-defend/cache/D4 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > D4.log 2>&1; enum_flags_test.go:113: JavaScript backend stdout = \"1/3/3/Alias/Qualified\\n0,1,2,None,A,B,Alias,Qualified\\n\", source Node = \"1/2/3/Alias/Qualified\\n0,1,2,None,A,B,Alias,Qualified\\n\"",
    "code_under_test": "enumConstant and enumFields",
    "oracle": "Node source versus lowered JavaScript output",
    "coverage_exclusive_blocks": 232,
    "finding": "Qualified same-enum aliases distinguish the input from foreign aliases and ordinary bitwise members. Alias values, reverse names and keys are checked.",
    "rows_passed_file": "D4-passed-rows.json"
  },
  {
    "test": "TestEnumNameEnumeration",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestFlagEnumInlineIteration"
    ],
    "defense": "defended",
    "unique_mutant": "D5 internal/lower/library_for_in.go:71",
    "attempts": [
      {
        "mutant": "D5",
        "file_line": "internal/lower/library_for_in.go:71",
        "change": "flip enum plain-enumerable origin condition",
        "rows_failed": [
          "TestEnumNameEnumeration"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/enum-defend/cache/D5 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > D5.log 2>&1; enum_flags_test.go:118: Lower refused acceptance row: /tmp/adamic-gate/TestEnumNameEnumeration1852416373/001/main.a:1:72: stage 0 can't lower for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) yet",
    "code_under_test": "plainEnumerableObject and forIn; enumFields",
    "oracle": "Node source versus lowered JavaScript output",
    "coverage_exclusive_blocks": 71,
    "finding": "Enum for-in origin differs from subsumer for-of iteration. Enumerated names now observed.",
    "rows_passed_file": "D5-passed-rows.json"
  },
  {
    "test": "TestFlagEnumInlineIteration",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestEnumNameEnumeration"
    ],
    "defense": "cannot-judge",
    "unique_mutant": null,
    "attempts": [],
    "evidence": "go test -json -count=1 -timeout 90s -coverpkg=github.com/system-inc/adamic/internal/lower -coverprofile=TestFlagEnumInlineIteration.cover ./internal/lower/ -run ^TestFlagEnumInlineIteration$; 606 exclusive blocks versus TestEnumNameEnumeration. No aimed mutant allocated within the seven final-mutant cap; no three-attempt negative verdict claimed.",
    "code_under_test": "forOf, libraryArrayForOf, flagDomainSeen and contextual object lowering",
    "oracle": "Node source versus lowered JavaScript output",
    "coverage_exclusive_blocks": 606,
    "finding": "The inline loop observes 0 then 3 and a reverse name. It does not directly assert a flag-domain classification; its observable iteration promise is checked."
  },
  {
    "test": "TestFlagEnumAliasBoundaries",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestNumericEnumLiteralPromises"
    ],
    "defense": "cannot-judge",
    "unique_mutant": null,
    "attempts": [],
    "evidence": "go test -json -count=1 -timeout 90s -coverpkg=github.com/system-inc/adamic/internal/lower -coverprofile=TestFlagEnumAliasBoundaries.cover ./internal/lower/ -run ^TestFlagEnumAliasBoundaries$; 985 exclusive blocks versus TestNumericEnumLiteralPromises. No aimed mutant allocated within the seven final-mutant cap; no three-attempt negative verdict claimed.",
    "code_under_test": "enumAssignable, flagMemberWidened, castProof and forOf",
    "oracle": "Node agreement for positive alias/iterable cases; self-written Refused type for two negative cases",
    "coverage_exclusive_blocks": 985,
    "finding": "Positive boundary outputs now checked. Negative member-slot and Mutable<T> cases accept any Refused, so an unrelated refusal can satisfy them."
  },
  {
    "test": "TestEnumInitializationGraphMemo",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestEnumNamespaceSharedCycle"
    ],
    "defense": "defended",
    "unique_mutant": "D1 internal/lower/namespaces_call_graph.go:29",
    "attempts": [
      {
        "mutant": "D1",
        "file_line": "internal/lower/namespaces_call_graph.go:29",
        "change": "flip condition: reuse active nodes and zero/multiple-target caches, but recompute the first completed single-target cache",
        "rows_failed": [
          "TestEnumInitializationGraphMemo"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/enum-defend/cache/D1 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > D1.log 2>&1; enum_initialization_reach_test.go:58: enum memo revisited bodies: 14",
    "code_under_test": "namespaceCallGraph.reach/discover cache and transitive propagation",
    "oracle": "Self-written reach count and exact body expansion counts",
    "coverage_exclusive_blocks": 1,
    "finding": "One exclusive block at namespaces_call_graph.go:135-136. Completed single-target cache queries differ on shared cache-hit lines. Counts genuinely assert work, not elapsed time.",
    "rows_passed_file": "D1-passed-rows.json"
  },
  {
    "test": "TestEnumNamespaceSharedCycle",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestEnumInitializationGraphMemo"
    ],
    "defense": "defended",
    "unique_mutant": "D2 internal/lower/namespaces_call_graph.go:140",
    "attempts": [
      {
        "mutant": "D2",
        "file_line": "internal/lower/namespaces_call_graph.go:140",
        "change": "drop statement for nonroot members of a two-node component",
        "rows_failed": [
          "TestEnumNamespaceSharedCycle"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/enum-defend/cache/D2 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > D2.log 2>&1; enum_initialization_reach_test.go:83: one shared component must reach both enum and namespace: map[]",
    "code_under_test": "namespaceCallGraph.reach/discover Tarjan components and reach sharing",
    "oracle": "Self-written two declaration identities and exact expansion count",
    "coverage_exclusive_blocks": 12,
    "finding": "Exclusive active-edge handling at 42-46 and namespace read at 103-104. Two-node component differs from three-node namespace cycle.",
    "rows_passed_file": "D2-passed-rows.json"
  },
  {
    "test": "TestStringEnumsStayClosed",
    "package": "internal/lower",
    "prior_verdict": "untrue",
    "subsumed_by": [],
    "defense": "defended",
    "unique_mutant": "D6 internal/lower/cast_proof.go:105",
    "attempts": [
      {
        "mutant": "D6",
        "file_line": "internal/lower/cast_proof.go:105",
        "change": "change diagnostic repair constant to empty for primitive-string to string-enum casts",
        "rows_failed": [
          "TestStringEnumsStayClosed"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/enum-defend/cache/D6 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > D6.log 2>&1; enums_open_test.go:55: want closed string enum refusal, got /tmp/adamic-gate/TestStringEnumsStayClosed366403871/001/main.a:1:70: Adamic 0.1 refuses a cast the runtime can't check;",
    "code_under_test": "castProof, enumAssignable and refuseWidening",
    "oracle": "Self-written Refused type and enum-members/no-unchecked-cast diagnostic alternatives",
    "coverage_exclusive_blocks": 0,
    "finding": "No exclusive block. Primitive-to-string-enum inputs distinguish shared cast proof. D6 proves the diagnostic assertion can fail, not that rejection can be bypassed.",
    "rows_passed_file": "D6-passed-rows.json"
  },
  {
    "test": "TestNumericEnumLiteralPromises",
    "package": "internal/lower",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestFlagEnumAliasBoundaries"
    ],
    "defense": "cannot-judge",
    "unique_mutant": null,
    "attempts": [],
    "evidence": "go test -json -count=1 -timeout 90s -coverpkg=github.com/system-inc/adamic/internal/lower -coverprofile=TestNumericEnumLiteralPromises.cover ./internal/lower/ -run ^TestNumericEnumLiteralPromises$; 39 exclusive blocks versus TestFlagEnumAliasBoundaries. No aimed mutant allocated within the seven final-mutant cap; no three-attempt negative verdict claimed.",
    "code_under_test": "enumAssignable, enumRefusal, enumStoredType and enumMemberOrigin",
    "oracle": "Self-written Refused type",
    "coverage_exclusive_blocks": 39,
    "finding": "The named promises are refused, but the row does not require an enum-literal diagnostic. An unrelated Refused can satisfy it. Coverage includes indexed stored-type and open-tag differences."
  }
]
```

The matrix ran the whole package for each final mutant. Each D*-passed-rows.json lists exactly the observed passing top-level rows. matrix.json lists failures and skips. No test, oracle, or preparation helper was mutated. The diagnostic mutant D6 alters a production repair string, not its expected test text. All final diffs apply against the starting commit and passed go vet ./internal/lower/.

Coverage: each requested row and its named subsumer ran independently with -coverpkg on internal/lower. The untrue string row was compared with a whole-package run excluding it. Profiles, per-row exclusive block positions and coverage-comparisons.json preserve the comparison. Source line spans in Go coverage blocks are leads, not proof of exclusive semantics.

Brief ambiguities and time costs:
- The three-honest-attempt negative rule conflicts with a seven-mutant ceiling for twelve rows. Final matrices cost roughly fifty seconds, so seven meaningful final mutants were prioritized. Five unallocated rows remain cannot-judge; no deletion is justified by that result.
- The prior audit was based on 8171b317, rather than the brief's older file reference. Main is now later and 36 rows were added with none removed. Most positive enum assertions now run Node comparisons. The previous acceptance-only limitations are not current evidence.
- Initial wide D1 cache disabling expanded diamond paths exponentially. It was interrupted; its partial log supports no verdict. Narrowed D1 disables only the first completed single-target cache entry and completes normally.
- Initial D3 selected normalized AST numeric text, so hexadecimal spelling was invisible and the package passed. Corrected sourceExpression reads source spelling. The inactive diff/log are retained and excluded from final defenses.
- D6 establishes unique diagnostic sensitivity. It leaves rejection enabled and does not prove a missing soundness barrier. The broad StringEnumsStayClosed name covers only the two cast cases actually asserted.
- Default package runs skip TestOriginalCycleLedger, TestOptionalWideningCensus and one TestMixedUnionContractGraph subcase. Their opt-in scenarios remain unknown. Defended means unique among the observed full default-package run; those skips and repo-wide uniqueness are not established.

For every cannot-judge row, rows.json records its name/assertion limits. Domain and inline iteration observe runtime output, not classification internals; never-default does not execute the default; negative alias/literal rows check Refused type without requiring the particular promised reason. Graph performance names are backed by exact expansion counts, a valid work assertion.

Timing: {"setup_seconds": 0, "npm_seconds": 0.538092886999948, "baseline_wall_seconds": 40.92384540800413, "baseline_binary_seconds": 38.863, "coverage_wall_seconds": 83.98269420199358, "final_matrix_wall_seconds": 342.41540663199703, "inactive_spelling_attempt_seconds": 47.54738754600112, "note": "Wide cache attempt was interrupted and excluded. Every final matrix has its own build-cache directory; D3 corrected spelling uses D3-source-spelling."}. No other package tests ran. The broad aborted attempt and inactive selector consumed extra time but neither is used as a proven catch.
