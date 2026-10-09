Unit u144 starts at 60f24a6ef736a6016edeca63df075678186bba8d.
All 35 requested declarations exist, grouped into 14 rows; none moved, vanished or skipped.
Whole-package and combined-slice baselines timed out; all grouped baselines and three-run timings passed.
Three eligible construction mutants plus one supplemental replacement: five setup-check rows, one untrue row on one applicable mutant, eight cannot-judge rows.
All 35 declarations passed their own empty-answer construction probes; nproc=5; uniqueness is bounded.

```json
[
  {
    "test": "TestProduct buildProduct family",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/build_products_test.go",
    "members": [
      "TestProduct_stage0",
      "TestProduct_checker",
      "TestProduct_checker_asan"
    ],
    "seconds": 6.468,
    "timing_samples": [
      7.087,
      6.468,
      2.457
    ],
    "timing_logs": [
      "review/test-audit/stage1-cohere-typeaware-build_products/baseline-group-01.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-01-2.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-01-3.log"
    ],
    "nproc": 5,
    "oracle": "Self-written successful construction and artifact-read/existence checks; returned products discarded. Build commands do not compare native findings with an outside oracle.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "unit_mutants_in_matrix": 4,
    "probe_kills": [],
    "own_probes": [
      "P01",
      "P02"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=P01 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_stage0)$' > P01.log 2>&1; TestProduct_stage0 PASS with empty entry answer\nADAMIC_MUTANT=P02 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_checker|TestProduct_checker_asan)$' > P02.log 2>&1; TestProduct_checker, TestProduct_checker_asan PASS with empty entry answer",
    "reason": "No applicable executed ordinary mutation of this row's own builder was proved; empty probes cannot establish a verdict. Production port behavior was not mutated.",
    "construction_only": true,
    "supplemental_kills": []
  },
  {
    "test": "TestProduct lowered family",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/build_products_test.go",
    "members": [
      "TestProduct_six_lowered",
      "TestProduct_typeaware_lowered",
      "TestProduct_facts_cost_lowered",
      "TestProduct_query_cost_lowered",
      "TestProduct_volume_lowered"
    ],
    "seconds": 4.51,
    "timing_samples": [
      37.27,
      4.51,
      3.349
    ],
    "timing_logs": [
      "review/test-audit/stage1-cohere-typeaware-build_products/baseline-group-02.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-02-2.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-02-3.log"
    ],
    "nproc": 5,
    "oracle": "Self-written successful construction and artifact-read/existence checks; returned products discarded. Build commands do not compare native findings with an outside oracle.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "unit_mutants_in_matrix": 4,
    "probe_kills": [],
    "own_probes": [
      "P03"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=P03 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_six_lowered|TestProduct_typeaware_lowered|TestProduct_facts_cost_lowered|TestProduct_query_cost_lowered|TestProduct_volume_lowered)$' > P03.log 2>&1; TestProduct_six_lowered, TestProduct_typeaware_lowered, TestProduct_query_cost_lowered, TestProduct_volume_lowered, TestProduct_facts_cost_lowered PASS with empty entry answer",
    "reason": "No applicable executed ordinary mutation of this row's own builder was proved; empty probes cannot establish a verdict. Production port behavior was not mutated.",
    "construction_only": true,
    "supplemental_kills": []
  },
  {
    "test": "TestProduct native family",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/build_products_test.go",
    "members": [
      "TestProduct_suite",
      "TestProduct_suite_asan",
      "TestProduct_typeaware_native",
      "TestProduct_typeaware_native_asan",
      "TestProduct_facts_cost_native",
      "TestProduct_query_cost_native",
      "TestProduct_volume_native"
    ],
    "seconds": 13.289,
    "timing_samples": [
      54.809,
      13.289,
      11.989
    ],
    "timing_logs": [
      "review/test-audit/stage1-cohere-typeaware-build_products/baseline-group-03.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-03-2.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-03-3.log"
    ],
    "nproc": 5,
    "oracle": "Self-written successful construction and artifact-read/existence checks; returned products discarded. Build commands do not compare native findings with an outside oracle.",
    "oracle_kind": "self",
    "kills": [
      "M2"
    ],
    "unique_kills": [
      "M2"
    ],
    "last_proven_fail": "M2: build_products_test.go:56: open /home/agent/.cache/adamic-build/e79e584ae0277617405c9712f2bad2fc5e702d9eb56ec5030331531ce8633eea/native.missing: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "unit_mutants_in_matrix": 4,
    "probe_kills": [],
    "own_probes": [
      "P04"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_facts_cost_native",
      "TestProduct_profile_missing_binding",
      "TestProduct_query_cost_native",
      "TestProduct_suite",
      "TestProduct_suite_asan",
      "TestProduct_typeaware_native",
      "TestProduct_typeaware_native_asan",
      "TestProduct_volume_native"
    ],
    "evidence": "ADAMIC_MUTANT=M2 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_suite|TestProduct_suite_asan|TestProduct_typeaware_native|TestProduct_typeaware_native_asan|TestProduct_facts_cost_native|TestProduct_query_cost_native|TestProduct_volume_native|TestProduct_profile_missing_binding)$' > M2.log 2>&1; build_products_test.go:56: open /home/agent/.cache/adamic-build/e79e584ae0277617405c9712f2bad2fc5e702d9eb56ec5030331531ce8633eea/native.missing: no such file or directory\nADAMIC_MUTANT=P04 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_suite|TestProduct_suite_asan|TestProduct_typeaware_native|TestProduct_typeaware_native_asan|TestProduct_facts_cost_native|TestProduct_query_cost_native|TestProduct_volume_native)$' > P04.log 2>&1; TestProduct_suite, TestProduct_facts_cost_native, TestProduct_volume_native, TestProduct_query_cost_native, TestProduct_typeaware_native, TestProduct_typeaware_native_asan, TestProduct_suite_asan PASS with empty entry answer",
    "reason": "",
    "construction_only": true,
    "supplemental_kills": []
  },
  {
    "test": "TestProduct sixBuildProduct family",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/build_products_test.go",
    "members": [
      "TestProduct_six_oracle",
      "TestProduct_typeaware_oracle",
      "TestProduct_facts_cost_go",
      "TestProduct_volume_oracle",
      "TestProduct_type_symbol_oracle"
    ],
    "seconds": 3.169,
    "timing_samples": [
      3.169,
      3.067,
      4.004
    ],
    "timing_logs": [
      "review/test-audit/stage1-cohere-typeaware-build_products/final-timing-01-1.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/final-timing-01-2.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/final-timing-01-3.log"
    ],
    "nproc": 5,
    "oracle": "Self-written successful construction and artifact-read/existence checks; returned products discarded. Build commands do not compare native findings with an outside oracle.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "M1: build_products_test.go:43: missing product /home/agent/.cache/adamic-build/5019b6c5cbbef43ab7811ec0858918983a6180d70f4957bd6de599bd94901b8a/oracle.missing: stat /home/agent/.cache/adamic-build/5019b6c5cbbef43ab7811ec0858918983a6180d70f4957bd6de599bd94901b8a/oracle.missing: no such file or directory",
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "unit_mutants_in_matrix": 4,
    "probe_kills": [],
    "own_probes": [
      "P05"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_facts_cost_go",
      "TestProduct_six_oracle",
      "TestProduct_type_symbol_archive",
      "TestProduct_type_symbol_native",
      "TestProduct_type_symbol_oracle",
      "TestProduct_type_symbol_truth",
      "TestProduct_typeaware_oracle",
      "TestProduct_volume_oracle"
    ],
    "evidence": "ADAMIC_MUTANT=M1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_six_oracle|TestProduct_typeaware_oracle|TestProduct_facts_cost_go|TestProduct_volume_oracle|TestProduct_type_symbol_archive|TestProduct_type_symbol_oracle|TestProduct_type_symbol_native|TestProduct_type_symbol_truth)$' > M1.log 2>&1; build_products_test.go:43: missing product /home/agent/.cache/adamic-build/5019b6c5cbbef43ab7811ec0858918983a6180d70f4957bd6de599bd94901b8a/oracle.missing: stat /home/agent/.cache/adamic-build/5019b6c5cbbef43ab7811ec0858918983a6180d70f4957bd6de599bd94901b8a/oracle.missing: no such file or directory\nADAMIC_MUTANT=P05 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_six_oracle|TestProduct_typeaware_oracle|TestProduct_facts_cost_go|TestProduct_volume_oracle|TestProduct_type_symbol_archive|TestProduct_type_symbol_oracle)$' > P05.log 2>&1; TestProduct_six_oracle, TestProduct_volume_oracle, TestProduct_typeaware_oracle, TestProduct_type_symbol_oracle, TestProduct_facts_cost_go, TestProduct_type_symbol_archive PASS with empty entry answer",
    "reason": "No applicable executed ordinary mutation of this row's own builder was proved; empty probes cannot establish a verdict. Production port behavior was not mutated.",
    "construction_only": true,
    "supplemental_kills": [
      "M1"
    ]
  },
  {
    "test": "TestProduct_type_symbol_archive",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/build_products_test.go",
    "members": [
      "TestProduct_type_symbol_archive"
    ],
    "seconds": 1.393,
    "timing_samples": [
      1.409,
      1.2810000000000001,
      1.393
    ],
    "timing_logs": [
      "review/test-audit/stage1-cohere-typeaware-build_products/final-timing-02-1.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/final-timing-02-2.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/final-timing-02-3.log"
    ],
    "nproc": 5,
    "oracle": "Self-written successful construction and artifact-read/existence checks; returned products discarded. Build commands do not compare native findings with an outside oracle.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "M1: volume_type_symbol_test.go:192: missing product /home/agent/.cache/adamic-build/d4ebfed40a4115a901cac2c4d93e651302ee30b9eac4f228491adea41bca48cb/type-symbol-checker.a.missing: stat /home/agent/.cache/adamic-build/d4ebfed40a4115a901cac2c4d93e651302ee30b9eac4f228491adea41bca48cb/type-symbol-checker.a.missing: no such file or directory",
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "unit_mutants_in_matrix": 4,
    "probe_kills": [],
    "own_probes": [
      "P05"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_facts_cost_go",
      "TestProduct_six_oracle",
      "TestProduct_type_symbol_archive",
      "TestProduct_type_symbol_native",
      "TestProduct_type_symbol_oracle",
      "TestProduct_type_symbol_truth",
      "TestProduct_typeaware_oracle",
      "TestProduct_volume_oracle"
    ],
    "evidence": "ADAMIC_MUTANT=M1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_six_oracle|TestProduct_typeaware_oracle|TestProduct_facts_cost_go|TestProduct_volume_oracle|TestProduct_type_symbol_archive|TestProduct_type_symbol_oracle|TestProduct_type_symbol_native|TestProduct_type_symbol_truth)$' > M1.log 2>&1; volume_type_symbol_test.go:192: missing product /home/agent/.cache/adamic-build/d4ebfed40a4115a901cac2c4d93e651302ee30b9eac4f228491adea41bca48cb/type-symbol-checker.a.missing: stat /home/agent/.cache/adamic-build/d4ebfed40a4115a901cac2c4d93e651302ee30b9eac4f228491adea41bca48cb/type-symbol-checker.a.missing: no such file or directory\nADAMIC_MUTANT=P05 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_six_oracle|TestProduct_typeaware_oracle|TestProduct_facts_cost_go|TestProduct_volume_oracle|TestProduct_type_symbol_archive|TestProduct_type_symbol_oracle)$' > P05.log 2>&1; TestProduct_six_oracle, TestProduct_volume_oracle, TestProduct_typeaware_oracle, TestProduct_type_symbol_oracle, TestProduct_facts_cost_go, TestProduct_type_symbol_archive PASS with empty entry answer",
    "reason": "No applicable executed ordinary mutation of this row's own builder was proved; empty probes cannot establish a verdict. Production port behavior was not mutated.",
    "construction_only": true,
    "supplemental_kills": [
      "M1"
    ]
  },
  {
    "test": "TestProduct_profile_controls_lowered",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/build_products_test.go",
    "members": [
      "TestProduct_profile_controls_lowered"
    ],
    "seconds": 0.79,
    "timing_samples": [
      33.495,
      0.759,
      0.79
    ],
    "timing_logs": [
      "review/test-audit/stage1-cohere-typeaware-build_products/baseline-group-05.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-05-2.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-05-3.log"
    ],
    "nproc": 5,
    "oracle": "Self-written successful construction and artifact-read/existence checks; returned products discarded. Build commands do not compare native findings with an outside oracle.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "unit_mutants_in_matrix": 4,
    "probe_kills": [],
    "own_probes": [
      "P06"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_corpus_native",
      "TestProduct_profile_controls_lowered",
      "TestProduct_profile_controls_native"
    ],
    "evidence": "ADAMIC_MUTANT=P06 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_profile_controls_lowered)$' > P06.log 2>&1; TestProduct_profile_controls_lowered PASS with empty entry answer",
    "reason": "One ordinary construction mutant changed its returned path to a nonexistent C file; the row passed. No failure proved by this fixed menu.",
    "construction_only": true,
    "supplemental_kills": []
  },
  {
    "test": "TestProduct_profile_controls_native",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/build_products_test.go",
    "members": [
      "TestProduct_profile_controls_native"
    ],
    "seconds": 3.44,
    "timing_samples": [
      3.145,
      4.025,
      3.44
    ],
    "timing_logs": [
      "review/test-audit/stage1-cohere-typeaware-build_products/final-timing-03-1.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/final-timing-03-2.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/final-timing-03-3.log"
    ],
    "nproc": 5,
    "oracle": "Self-written successful construction and artifact-read/existence checks; returned products discarded. Build commands do not compare native findings with an outside oracle.",
    "oracle_kind": "self",
    "kills": [
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3: volume_profile_Controls_test.go:264: open /tmp/u144/cache/M3-cold/f3a6e9a8c3c3f7486a3578c08abf86c13bf20d9fb693d09986a17f9218dba4dc/missing.c: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "unit_mutants_in_matrix": 4,
    "probe_kills": [],
    "own_probes": [
      "P07"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_corpus_native",
      "TestProduct_profile_controls_lowered",
      "TestProduct_profile_controls_native"
    ],
    "evidence": "ADAMIC_MUTANT=M3 ADAMIC_BUILD_CACHE_DIR=/tmp/u144/cache/M3-cold timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^TestProduct_profile_controls_native$' > M3-cold.log 2>&1; volume_profile_Controls_test.go:264: open /tmp/u144/cache/M3-cold/f3a6e9a8c3c3f7486a3578c08abf86c13bf20d9fb693d09986a17f9218dba4dc/missing.c: no such file or directory\nADAMIC_MUTANT=P07 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_profile_controls_native)$' > P07.log 2>&1; TestProduct_profile_controls_native PASS with empty entry answer",
    "reason": "",
    "construction_only": true,
    "supplemental_kills": []
  },
  {
    "test": "TestProduct_profile_missing_binding",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/build_products_test.go",
    "members": [
      "TestProduct_profile_missing_binding"
    ],
    "seconds": 2.632,
    "timing_samples": [
      2.632,
      2.416,
      3.5
    ],
    "timing_logs": [
      "review/test-audit/stage1-cohere-typeaware-build_products/final-timing-04-1.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/final-timing-04-2.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/final-timing-04-3.log"
    ],
    "nproc": 5,
    "oracle": "Self-written successful construction and artifact-read/existence checks; returned products discarded. Build commands do not compare native findings with an outside oracle.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "unit_mutants_in_matrix": 4,
    "probe_kills": [],
    "own_probes": [
      "P08"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_facts_cost_native",
      "TestProduct_profile_missing_binding",
      "TestProduct_query_cost_native",
      "TestProduct_suite",
      "TestProduct_suite_asan",
      "TestProduct_typeaware_native",
      "TestProduct_typeaware_native_asan",
      "TestProduct_volume_native"
    ],
    "evidence": "ADAMIC_MUTANT=P08 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_profile_missing_binding)$' > P08.log 2>&1; TestProduct_profile_missing_binding PASS with empty entry answer",
    "reason": "No applicable executed ordinary mutation of this row's own builder was proved; empty probes cannot establish a verdict. Production port behavior was not mutated.",
    "construction_only": true,
    "supplemental_kills": []
  },
  {
    "test": "TestProduct profile_mutant family",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/build_products_test.go",
    "members": [
      "TestProduct_profile_mutant_binding_slot",
      "TestProduct_profile_mutant_scope_containment",
      "TestProduct_profile_mutant_first_binding"
    ],
    "seconds": 4.369,
    "timing_samples": [
      66.407,
      4.194,
      4.369
    ],
    "timing_logs": [
      "review/test-audit/stage1-cohere-typeaware-build_products/baseline-group-07.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-07-2.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-07-3.log"
    ],
    "nproc": 5,
    "oracle": "Self-written successful construction and artifact-read/existence checks; returned products discarded. Build commands do not compare native findings with an outside oracle.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "unit_mutants_in_matrix": 4,
    "probe_kills": [],
    "own_probes": [
      "P09"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=P09 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_profile_mutant_binding_slot|TestProduct_profile_mutant_scope_containment|TestProduct_profile_mutant_first_binding)$' > P09.log 2>&1; TestProduct_profile_mutant_first_binding, TestProduct_profile_mutant_binding_slot, TestProduct_profile_mutant_scope_containment PASS with empty entry answer",
    "reason": "No applicable executed ordinary mutation of this row's own builder was proved; empty probes cannot establish a verdict. Production port behavior was not mutated.",
    "construction_only": true,
    "supplemental_kills": []
  },
  {
    "test": "TestProduct corpusGo family",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/build_products_test.go",
    "members": [
      "TestProduct_corpus_stage0",
      "TestProduct_corpus_checker",
      "TestProduct_corpus_checker_asan",
      "TestProduct_corpus_oracle"
    ],
    "seconds": 2.917,
    "timing_samples": [
      28.994,
      2.917,
      2.833
    ],
    "timing_logs": [
      "review/test-audit/stage1-cohere-typeaware-build_products/baseline-group-08.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-08-2.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-08-3.log"
    ],
    "nproc": 5,
    "oracle": "Self-written successful construction and artifact-read/existence checks; returned products discarded. Build commands do not compare native findings with an outside oracle.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "unit_mutants_in_matrix": 4,
    "probe_kills": [],
    "own_probes": [
      "P10"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=P10 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_corpus_stage0|TestProduct_corpus_checker|TestProduct_corpus_checker_asan|TestProduct_corpus_oracle)$' > P10.log 2>&1; TestProduct_corpus_stage0, TestProduct_corpus_oracle, TestProduct_corpus_checker_asan, TestProduct_corpus_checker PASS with empty entry answer",
    "reason": "No applicable executed ordinary mutation of this row's own builder was proved; empty probes cannot establish a verdict. Production port behavior was not mutated.",
    "construction_only": true,
    "supplemental_kills": []
  },
  {
    "test": "TestProduct_corpus_native",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/build_products_test.go",
    "members": [
      "TestProduct_corpus_native"
    ],
    "seconds": 4.565,
    "timing_samples": [
      21.209,
      4.153,
      4.5649999999999995
    ],
    "timing_logs": [
      "review/test-audit/stage1-cohere-typeaware-build_products/baseline-group-09.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-09-2.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-09-3.log"
    ],
    "nproc": 5,
    "oracle": "Self-written successful construction and artifact-read/existence checks; returned products discarded. Build commands do not compare native findings with an outside oracle.",
    "oracle_kind": "self",
    "kills": [
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3: volume_profile_Corpora_test.go:169: open /home/agent/.cache/adamic-build/f3a6e9a8c3c3f7486a3578c08abf86c13bf20d9fb693d09986a17f9218dba4dc/missing.c: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "unit_mutants_in_matrix": 4,
    "probe_kills": [],
    "own_probes": [
      "P11"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_corpus_native",
      "TestProduct_profile_controls_lowered",
      "TestProduct_profile_controls_native"
    ],
    "evidence": "ADAMIC_MUTANT=M3 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_profile_controls_lowered|TestProduct_profile_controls_native|TestProduct_corpus_native)$' > M3.log 2>&1; volume_profile_Corpora_test.go:169: open /home/agent/.cache/adamic-build/f3a6e9a8c3c3f7486a3578c08abf86c13bf20d9fb693d09986a17f9218dba4dc/missing.c: no such file or directory\nADAMIC_MUTANT=P11 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_corpus_native)$' > P11.log 2>&1; TestProduct_corpus_native PASS with empty entry answer",
    "reason": "",
    "construction_only": true,
    "supplemental_kills": []
  },
  {
    "test": "TestProduct_type_symbol_native",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/build_products_test.go",
    "members": [
      "TestProduct_type_symbol_native"
    ],
    "seconds": 3.492,
    "timing_samples": [
      60.713,
      2.993,
      3.492
    ],
    "timing_logs": [
      "review/test-audit/stage1-cohere-typeaware-build_products/baseline-group-10.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-10-2.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-10-3.log"
    ],
    "nproc": 5,
    "oracle": "Self-written successful construction and artifact-read/existence checks; returned products discarded. Build commands do not compare native findings with an outside oracle.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "M1: volume_type_symbol_test.go:192: missing product /home/agent/.cache/adamic-build/d4ebfed40a4115a901cac2c4d93e651302ee30b9eac4f228491adea41bca48cb/type-symbol-checker.a.missing: stat /home/agent/.cache/adamic-build/d4ebfed40a4115a901cac2c4d93e651302ee30b9eac4f228491adea41bca48cb/type-symbol-checker.a.missing: no such file or directory",
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "unit_mutants_in_matrix": 4,
    "probe_kills": [],
    "own_probes": [
      "P12"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_facts_cost_go",
      "TestProduct_six_oracle",
      "TestProduct_type_symbol_archive",
      "TestProduct_type_symbol_native",
      "TestProduct_type_symbol_oracle",
      "TestProduct_type_symbol_truth",
      "TestProduct_typeaware_oracle",
      "TestProduct_volume_oracle"
    ],
    "evidence": "ADAMIC_MUTANT=M1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_six_oracle|TestProduct_typeaware_oracle|TestProduct_facts_cost_go|TestProduct_volume_oracle|TestProduct_type_symbol_archive|TestProduct_type_symbol_oracle|TestProduct_type_symbol_native|TestProduct_type_symbol_truth)$' > M1.log 2>&1; volume_type_symbol_test.go:192: missing product /home/agent/.cache/adamic-build/d4ebfed40a4115a901cac2c4d93e651302ee30b9eac4f228491adea41bca48cb/type-symbol-checker.a.missing: stat /home/agent/.cache/adamic-build/d4ebfed40a4115a901cac2c4d93e651302ee30b9eac4f228491adea41bca48cb/type-symbol-checker.a.missing: no such file or directory\nADAMIC_MUTANT=P12 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_type_symbol_native)$' > P12.log 2>&1; TestProduct_type_symbol_native PASS with empty entry answer",
    "reason": "No applicable executed ordinary mutation of this row's own builder was proved; empty probes cannot establish a verdict. Production port behavior was not mutated.",
    "construction_only": true,
    "supplemental_kills": [
      "M1"
    ]
  },
  {
    "test": "TestProduct_type_symbol_fixtures",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/build_products_test.go",
    "members": [
      "TestProduct_type_symbol_fixtures"
    ],
    "seconds": 0.015,
    "timing_samples": [
      0.034,
      0.014,
      0.015
    ],
    "timing_logs": [
      "review/test-audit/stage1-cohere-typeaware-build_products/baseline-group-11.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-11-2.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-11-3.log"
    ],
    "nproc": 5,
    "oracle": "Self-written successful construction and artifact-read/existence checks; returned products discarded. Build commands do not compare native findings with an outside oracle.",
    "oracle_kind": "self",
    "kills": [
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: volume_type_symbol_test.go:247: open /home/agent/.cache/adamic-build/646951ff62aa11e5d2a3ff5db8a93d0444d22b5b9839d5f4a88001acad5a37d3/missing.manifest: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "unit_mutants_in_matrix": 4,
    "probe_kills": [],
    "own_probes": [
      "P13"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_type_symbol_fixtures",
      "TestProduct_type_symbol_truth"
    ],
    "evidence": "ADAMIC_MUTANT=M4 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_type_symbol_fixtures|TestProduct_type_symbol_truth)$' > M4.log 2>&1; volume_type_symbol_test.go:247: open /home/agent/.cache/adamic-build/646951ff62aa11e5d2a3ff5db8a93d0444d22b5b9839d5f4a88001acad5a37d3/missing.manifest: no such file or directory\nADAMIC_MUTANT=P13 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_type_symbol_fixtures)$' > P13.log 2>&1; TestProduct_type_symbol_fixtures PASS with empty entry answer",
    "reason": "",
    "construction_only": true,
    "supplemental_kills": []
  },
  {
    "test": "TestProduct_type_symbol_truth",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/build_products_test.go",
    "members": [
      "TestProduct_type_symbol_truth"
    ],
    "seconds": 1.429,
    "timing_samples": [
      2.213,
      1.228,
      1.429
    ],
    "timing_logs": [
      "review/test-audit/stage1-cohere-typeaware-build_products/baseline-group-12.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-12-2.log",
      "review/test-audit/stage1-cohere-typeaware-build_products/timing-12-3.log"
    ],
    "nproc": 5,
    "oracle": "Self-written successful construction and artifact-read/existence checks; outputs discarded. Go/clang execution builds artifacts and supplies no semantic expected answer. Type-symbol truth collects Go output without comparison.",
    "oracle_kind": "self",
    "kills": [
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: volume_type_symbol_test.go:247: open /home/agent/.cache/adamic-build/646951ff62aa11e5d2a3ff5db8a93d0444d22b5b9839d5f4a88001acad5a37d3/missing.manifest: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 2,
    "unit_mutants_in_matrix": 4,
    "probe_kills": [],
    "own_probes": [
      "P14"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_facts_cost_go",
      "TestProduct_six_oracle",
      "TestProduct_type_symbol_archive",
      "TestProduct_type_symbol_fixtures",
      "TestProduct_type_symbol_native",
      "TestProduct_type_symbol_oracle",
      "TestProduct_type_symbol_truth",
      "TestProduct_typeaware_oracle",
      "TestProduct_volume_oracle"
    ],
    "evidence": "ADAMIC_MUTANT=M1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_six_oracle|TestProduct_typeaware_oracle|TestProduct_facts_cost_go|TestProduct_volume_oracle|TestProduct_type_symbol_archive|TestProduct_type_symbol_oracle|TestProduct_type_symbol_native|TestProduct_type_symbol_truth)$' > M1.log 2>&1; volume_type_symbol_test.go:62: missing product /home/agent/.cache/adamic-build/85cade732e53b5081e2f3ed2298c3a0582868e6114428c1d32e1e46a9de113d3/type-symbol-oracle.missing: stat /home/agent/.cache/adamic-build/85cade732e53b5081e2f3ed2298c3a0582868e6114428c1d32e1e46a9de113d3/type-symbol-oracle.missing: no such file or directory\nADAMIC_MUTANT=M4 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_type_symbol_fixtures|TestProduct_type_symbol_truth)$' > M4.log 2>&1; volume_type_symbol_test.go:247: open /home/agent/.cache/adamic-build/646951ff62aa11e5d2a3ff5db8a93d0444d22b5b9839d5f4a88001acad5a37d3/missing.manifest: no such file or directory\nADAMIC_MUTANT=P14 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestProduct_type_symbol_truth)$' > P14.log 2>&1; TestProduct_type_symbol_truth PASS with empty entry answer",
    "reason": "",
    "construction_only": true,
    "supplemental_kills": [
      "M1"
    ]
  }
]
```

| ID | origin file:line | Change | Failed grouped rows |
|---|---|---|---|
| M1 supplemental | stage1/cohere/typeaware/six_builds_test.go:323 | `path := filepath.Join(dir, product.File)` to `path := filepath.Join(dir, product.File + ".missing")` |  |
| M2 | stage1/cohere/typeaware/typeaware_test.go:245 | `data, err := os.ReadFile(binary)` to `data, err := os.ReadFile(binary + ".missing")` | TestProduct native family |
| M3 | stage1/cohere/typeaware/volume_profile_Controls_test.go:201 | `return filepath.Join(directory, "volume.c")` to `return filepath.Join(directory, "missing.c")` | TestProduct_profile_controls_native, TestProduct_corpus_native |
| M4 | stage1/cohere/typeaware/volume_type_symbol_test.go:233 | `manifest := filepath.Join(directory, "controls.manifest")` to `manifest := filepath.Join(directory, "missing.manifest")` | TestProduct_type_symbol_fixtures, TestProduct_type_symbol_truth |

There are no wholly surviving ordinary mutants in the bounded matrix. M3 survived TestProduct_profile_controls_lowered and the initial warm TestProduct_profile_controls_native run. Its returned missing.c did not exist, while the same product directory held volume.c, 3,719,831 bytes. CorpusNative failed reading missing.c. ControlsNative failed on that path in the isolated-cache replay, at 44.621 test-binary seconds. The lowered declaration still passed. See survivor-witness.json and M3-cold.log.

The brief and costs need these qualifications:

- The supplied file reference was from 8de93800f4, but fetching origin/main produced 60f24a6e. All 35 names still exist in the same declaration file. The authoritative list and starting SHA are saved.
- These are build-product declarations, not native agreement tests. Treating their missing-binding or mutant names as witnesses would audit the wrong check. They never execute the planted native failures here. The real disagreement checks are elsewhere, outside this unit.
- The applicable code under test is suite construction. The four edits use the explicitly permitted setup-check construction exception. No port source, Go cohere oracle, test body, or expected finding was changed. Successful construction is a self-written contract, even when a Go or clang child produces the artifact. This audit proves no semantic port correctness.
- Families have nested builders. Common buildProduct, lowered, native, sixBuildProduct, corpusGo and profile-mutant recipes are grouped. TypeSymbolArchive has an additional unique-overlay precondition and stays separate; MissingBinding has source-replacement preconditions and stays separate from ControlsNative. Refined groups were retimed three times on their own. Initial superseded groups and logs remain as evidence.
- The complete package timed out in unrelated sequential TestInspectRequestRefusals before paused product rows ran. The exact 35-row slice also exceeded 90 seconds because native products built concurrently. Neither timeout was counted as a mutant kill or a red assertion baseline. Twelve baseline groups completed green; final refined row timings also passed. Outside the reached row lists, kills remain unknown.
- Warm installed tools were not warm Node dependencies. npm ci ran before baseline. Cache-backed builders created substantial cold lowering and clang work despite the warm environment. The largest ordinary mutant binary run was M2 at 76.419 seconds; the cold M3 replay took 44.621 seconds. No mutant run timed out or panicked.
- Cached products can mask a construction mutation. The ControlsLowered returned path is consumed inside ControlsNative's cache-miss callback. A warm cached binary bypassed that read. An isolated cache exposed the failure. The final matrix retains both observations; no cached survivor is mistaken for unchanged behavior.
- Empty answers here are construction entry probes, not empty native findings. The tests discard returned paths, C strings, fixtures and collected truth. All assigned probe runs passed, so vacuous=true refers precisely to that construction contract. No semantic positive/negative subcases ran and there are no vacuous_subcases claims.
- Three eligible ordinary mutations do not honestly cover every builder. M1 originally changed the validator and was excluded. The replacement constructs a wrong path with the validator intact, but was chosen after the superseded run, so it is supplemental and supports no verdict or uniqueness claim. Eight rows remain cannot-judge rather than being declared redundant or untrue. ControlsLowered is untrue on the one applicable wrong-return mutation in this menu, not proof that every conceivable constructor failure would pass it. Its build error checks could still catch other breaks, which were not mutated in this session.
- Setup-check verdicts establish that a construction can fail. They do not earn sacred, slow-worthy or production subsumption claims. M2's unique_kills value is unique among the reached bounded rows only. Central replay must settle all broader uniqueness, including other product families outside this unit.
- The switch was only a scratch implementation of the frozen edits and separate probes. All standalone mutants and probes compile under go vet and apply against the starting source. Standalone probe diffs use a non-nil caller guard to avoid unreachable-code vet errors; the observed selector probes perform the same early return. The superseded-switch.txt and superseded-M1-validator.log files are excluded evidence. Fourteen probes are recorded separately and do not support the setup-check verdicts.
- Tool transport disconnected briefly during a read and recovered. A scratch formatting command initially lacked gofmt on PATH; sourcing the already validated tool environment fixed it before any matrix execution. These events are not test failures.

Costs and limits: {
  "setup": "warm env validated; bootstrap skipped",
  "npm_ci_seconds": 0.796,
  "nproc": 5,
  "whole_baseline_test_seconds": 90.497,
  "product_slice_baseline_test_seconds": 90.054,
  "group_baseline_command_wall_seconds": 464.5758482420042,
  "timing_repeat_command_wall_seconds": 163.74228767400018,
  "mutation_and_probe_command_wall_seconds": 171.2692765479951,
  "isolated_M3_test_seconds": 44.621,
  "standalone_mutant_vet_seconds": 1.673296057002517,
  "standalone_probe_vet_seconds": 5.058816742995987,
  "builds": "Per-phase build timings are in build-phases.json. Command wall includes Go test compilation; row medians use the test binary package Elapsed line. Costs overlap parallel child phases and are not additive."
}

The session began around 12:58 UTC and evidence preparation finished around 13:28 UTC. No bootstrap was needed. Per-phase build durations are saved rather than inferred from Go command wall time. All requested names, timing runs, probe runs and four bounded construction matrices are covered. Production semantic mutants, unrelated package rows, repo-wide uniqueness, and eight ordinary constructor verdicts remain uncovered. There were no requested-row skips. Scratch source was restored; only evidence is committed. No main push or pull request.
