Unit u020: internal/fuzz at 7b18d0576930caca4e22ce2eef92fcf563af52d0.
15 top-level rows, no cross-top-level families, no skips; baseline passed in 18.779s.
15 menu mutants plus one supplemental comparison removal; counts: {'untrue': 1, 'subsumed': 5, 'sacred': 7, 'overlapping': 1, 'helper': 1}.
Empty-answer probes recorded separately; lowering vacuity is entry-specific.
Evidence: review/test-audit/internal-fuzz/ on test-audit/internal-fuzz.

[
  {
    "test": "TestOneSeedOneProgram",
    "package": "internal/fuzz",
    "file": "internal/fuzz/fuzz_test.go:15",
    "seconds": 0.022,
    "oracle": "Self-written same-seed equality and different-seed inequality of generated source.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 16,
    "eligible_mutants": 15,
    "probe_kills": [
      "PGenerate",
      "PWithout",
      "PFeatures",
      "PShareCuts"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestOneSeedOneProgram",
      "TestGeneratedProgramsCheckAndLower",
      "TestRegexProgramsPassTheChecker",
      "TestOctoberFeaturesAppear",
      "TestShrinkKeepsOnlyWhatFails",
      "TestSharedSliceCutsShare",
      "TestOwnershipShapes",
      "TestJudgeReadsThePanicLine",
      "TestOverridesShapesAndLower",
      "TestSourceTreeCutsAtItems",
      "TestReduceKeepsTheSignature",
      "TestExecuteCPUHelper",
      "TestExecuteCPULimit",
      "TestFuzzerSharesRuntimeLibrary",
      "TestUndefinedNumbersShapes"
    ],
    "evidence": "16 planned whole-package runs; see runs.json; No production mutant failure observed",
    "entry_probes": [
      "PGenerate"
    ],
    "evidence_details": {
      "command": "16 planned whole-package runs; see runs.json",
      "line": "No production mutant failure observed"
    }
  },
  {
    "test": "TestGeneratedProgramsCheckAndLower",
    "package": "internal/fuzz",
    "file": "internal/fuzz/fuzz_test.go:30",
    "seconds": 2.49,
    "oracle": "Self-written expectation that all 60 generated programs are accepted by load.Load and lower.Lower; checks errors only, never inspects returned IR.",
    "oracle_kind": "self",
    "kills": [
      "M10"
    ],
    "unique_kills": [],
    "last_proven_fail": "M10: fuzz_test.go:61: seed 1: the checker refused it: /tmp/adamic-gate/TestGeneratedProgramsCheckAndLowerseeds-001-005429757921/001/program.a:212:7: error TS2451: Cannot redeclare block-scoped variable 'maybeValue0'.",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestRegexProgramsPassTheChecker"
    ],
    "mutants_in_matrix": 16,
    "eligible_mutants": 15,
    "probe_kills": [
      "PGenerate",
      "PWithout",
      "PFeatures",
      "PShareCuts"
    ],
    "subsumer_seconds": 0.531,
    "vacuous": true,
    "bounded": false,
    "matrix_rows": [
      "TestOneSeedOneProgram",
      "TestGeneratedProgramsCheckAndLower",
      "TestRegexProgramsPassTheChecker",
      "TestOctoberFeaturesAppear",
      "TestShrinkKeepsOnlyWhatFails",
      "TestSharedSliceCutsShare",
      "TestOwnershipShapes",
      "TestJudgeReadsThePanicLine",
      "TestOverridesShapesAndLower",
      "TestSourceTreeCutsAtItems",
      "TestReduceKeepsTheSignature",
      "TestExecuteCPUHelper",
      "TestExecuteCPULimit",
      "TestFuzzerSharesRuntimeLibrary",
      "TestUndefinedNumbersShapes"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M10 ADAMIC_MUTANT=M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M10.log 2>&1; fuzz_test.go:61: seed 1: the checker refused it: /tmp/adamic-gate/TestGeneratedProgramsCheckAndLowerseeds-001-005429757921/001/program.a:212:7: error TS2451: Cannot redeclare block-scoped variable 'maybeValue0'.",
    "entry_probes": [
      "PLower"
    ],
    "vacuous_entry": "lower.Lower returns nil IR with nil error; vocabulary assertions remain active.",
    "evidence_details": {
      "command": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M10 ADAMIC_MUTANT=M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M10.log 2>&1",
      "line": "fuzz_test.go:61: seed 1: the checker refused it: /tmp/adamic-gate/TestGeneratedProgramsCheckAndLowerseeds-001-005429757921/001/program.a:212:7: error TS2451: Cannot redeclare block-scoped variable 'maybeValue0'.",
      "log": "review/test-audit/internal-fuzz/M10.log"
    }
  },
  {
    "test": "TestRegexProgramsPassTheChecker",
    "package": "internal/fuzz",
    "file": "internal/fuzz/fuzz_test.go:73",
    "seconds": 0.531,
    "oracle": "Self-written expectation of checker acceptance for 30 generated programs; uses the same full Generate source as the lowering row, not a regex-only corpus.",
    "oracle_kind": "self",
    "kills": [
      "M10"
    ],
    "unique_kills": [],
    "last_proven_fail": "M10: fuzz_test.go:97: seed 1: the checker refused it: /tmp/adamic-gate/TestRegexProgramsPassTheCheckerseeds-001-005945517187/001/program.a:212:7: error TS2451: Cannot redeclare block-scoped variable 'maybeValue0'.",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestOverridesShapesAndLower"
    ],
    "mutants_in_matrix": 16,
    "eligible_mutants": 15,
    "probe_kills": [
      "PGenerate",
      "PWithout",
      "PFeatures",
      "PShareCuts"
    ],
    "subsumer_seconds": 0.834,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestOneSeedOneProgram",
      "TestGeneratedProgramsCheckAndLower",
      "TestRegexProgramsPassTheChecker",
      "TestOctoberFeaturesAppear",
      "TestShrinkKeepsOnlyWhatFails",
      "TestSharedSliceCutsShare",
      "TestOwnershipShapes",
      "TestJudgeReadsThePanicLine",
      "TestOverridesShapesAndLower",
      "TestSourceTreeCutsAtItems",
      "TestReduceKeepsTheSignature",
      "TestExecuteCPUHelper",
      "TestExecuteCPULimit",
      "TestFuzzerSharesRuntimeLibrary",
      "TestUndefinedNumbersShapes"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M10 ADAMIC_MUTANT=M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M10.log 2>&1; fuzz_test.go:97: seed 1: the checker refused it: /tmp/adamic-gate/TestRegexProgramsPassTheCheckerseeds-001-005945517187/001/program.a:212:7: error TS2451: Cannot redeclare block-scoped variable 'maybeValue0'.",
    "entry_probes": [
      "PGenerate"
    ],
    "evidence_details": {
      "command": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M10 ADAMIC_MUTANT=M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M10.log 2>&1",
      "line": "fuzz_test.go:97: seed 1: the checker refused it: /tmp/adamic-gate/TestRegexProgramsPassTheCheckerseeds-001-005945517187/001/program.a:212:7: error TS2451: Cannot redeclare block-scoped variable 'maybeValue0'.",
      "log": "review/test-audit/internal-fuzz/M10.log"
    }
  },
  {
    "test": "TestOctoberFeaturesAppear",
    "package": "internal/fuzz",
    "file": "internal/fuzz/fuzz_test.go:106",
    "seconds": 0.091,
    "oracle": "Self-written substring vocabulary across 200 generated seeds; no generated program execution.",
    "oracle_kind": "self",
    "kills": [
      "M02"
    ],
    "unique_kills": [],
    "last_proven_fail": "M02: fuzz_test.go:120: 200 seeds never wrote extends",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestSharedSliceCutsShare"
    ],
    "mutants_in_matrix": 16,
    "eligible_mutants": 15,
    "probe_kills": [
      "PGenerate",
      "PWithout",
      "PFeatures",
      "PShareCuts"
    ],
    "subsumer_seconds": 0.019,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestOneSeedOneProgram",
      "TestGeneratedProgramsCheckAndLower",
      "TestRegexProgramsPassTheChecker",
      "TestOctoberFeaturesAppear",
      "TestShrinkKeepsOnlyWhatFails",
      "TestSharedSliceCutsShare",
      "TestOwnershipShapes",
      "TestJudgeReadsThePanicLine",
      "TestOverridesShapesAndLower",
      "TestSourceTreeCutsAtItems",
      "TestReduceKeepsTheSignature",
      "TestExecuteCPUHelper",
      "TestExecuteCPULimit",
      "TestFuzzerSharesRuntimeLibrary",
      "TestUndefinedNumbersShapes"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M02 ADAMIC_MUTANT=M02 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M02.log 2>&1; fuzz_test.go:120: 200 seeds never wrote extends",
    "entry_probes": [
      "PGenerate"
    ],
    "evidence_details": {
      "command": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M02 ADAMIC_MUTANT=M02 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M02.log 2>&1",
      "line": "fuzz_test.go:120: 200 seeds never wrote extends",
      "log": "review/test-audit/internal-fuzz/M02.log"
    }
  },
  {
    "test": "TestShrinkKeepsOnlyWhatFails",
    "package": "internal/fuzz",
    "file": "internal/fuzz/fuzz_test.go:127",
    "seconds": 0.013,
    "oracle": "Self-written table.delete marker predicate, three-line bound, and original-versus-copy source check. Tests shrinker, not compiler or agreement oracle.",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M15"
    ],
    "unique_kills": [
      "M03"
    ],
    "last_proven_fail": "M15: fuzz_test.go:148: shrunk to 540 lines after 1202 tries, want at most 3:",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 16,
    "eligible_mutants": 15,
    "probe_kills": [
      "PGenerate",
      "PWithout",
      "PFeatures",
      "PShrink",
      "PShareCuts"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestOneSeedOneProgram",
      "TestGeneratedProgramsCheckAndLower",
      "TestRegexProgramsPassTheChecker",
      "TestOctoberFeaturesAppear",
      "TestShrinkKeepsOnlyWhatFails",
      "TestSharedSliceCutsShare",
      "TestOwnershipShapes",
      "TestJudgeReadsThePanicLine",
      "TestOverridesShapesAndLower",
      "TestSourceTreeCutsAtItems",
      "TestReduceKeepsTheSignature",
      "TestExecuteCPUHelper",
      "TestExecuteCPULimit",
      "TestFuzzerSharesRuntimeLibrary",
      "TestUndefinedNumbersShapes"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M15 ADAMIC_MUTANT=M15 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M15.log 2>&1; fuzz_test.go:148: shrunk to 540 lines after 1202 tries, want at most 3:",
    "entry_probes": [
      "PShrink"
    ],
    "evidence_details": {
      "command": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M15 ADAMIC_MUTANT=M15 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M15.log 2>&1",
      "line": "fuzz_test.go:148: shrunk to 540 lines after 1202 tries, want at most 3:",
      "log": "review/test-audit/internal-fuzz/M15.log"
    }
  },
  {
    "test": "TestSharedSliceCutsShare",
    "package": "internal/fuzz",
    "file": "internal/fuzz/fuzz_test.go:157",
    "seconds": 0.019,
    "oracle": "Self-written sharing geometry and vocabulary. bytesShared copies an older runtime rule (64-byte minimum, quarter-owner ratio); current runtime uses header/storage ratio 8 and no minimum. It never checks real runtime sharing.",
    "oracle_kind": "self",
    "kills": [
      "M02"
    ],
    "unique_kills": [],
    "last_proven_fail": "M02: fuzz_test.go:196: vocabulary missing: probe false long string true sharing slice false",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestOctoberFeaturesAppear"
    ],
    "mutants_in_matrix": 16,
    "eligible_mutants": 15,
    "probe_kills": [
      "PGenerate",
      "PWithout",
      "PFeatures",
      "PShareCuts"
    ],
    "subsumer_seconds": 0.091,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestOneSeedOneProgram",
      "TestGeneratedProgramsCheckAndLower",
      "TestRegexProgramsPassTheChecker",
      "TestOctoberFeaturesAppear",
      "TestShrinkKeepsOnlyWhatFails",
      "TestSharedSliceCutsShare",
      "TestOwnershipShapes",
      "TestJudgeReadsThePanicLine",
      "TestOverridesShapesAndLower",
      "TestSourceTreeCutsAtItems",
      "TestReduceKeepsTheSignature",
      "TestExecuteCPUHelper",
      "TestExecuteCPULimit",
      "TestFuzzerSharesRuntimeLibrary",
      "TestUndefinedNumbersShapes"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M02 ADAMIC_MUTANT=M02 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M02.log 2>&1; fuzz_test.go:196: vocabulary missing: probe false long string true sharing slice false",
    "entry_probes": [
      "PGenerate",
      "PShareCuts"
    ],
    "evidence_details": {
      "command": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M02 ADAMIC_MUTANT=M02 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M02.log 2>&1",
      "line": "fuzz_test.go:196: vocabulary missing: probe false long string true sharing slice false",
      "log": "review/test-audit/internal-fuzz/M02.log"
    }
  },
  {
    "test": "TestOwnershipShapes",
    "package": "internal/fuzz",
    "file": "internal/fuzz/fuzz_test.go:207",
    "seconds": 0.842,
    "oracle": "Self-written vocabulary/cadence over eight seeds plus checker/lowering success; no runtime ownership observations and no returned IR assertion.",
    "oracle_kind": "self",
    "kills": [
      "M02",
      "M10",
      "M11"
    ],
    "unique_kills": [
      "M11"
    ],
    "last_proven_fail": "M11: fuzz_test.go:251: seeds 1 to 8 never generated new OwnMarked(",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 16,
    "eligible_mutants": 15,
    "probe_kills": [
      "PGenerate",
      "PWithout",
      "PFeatures",
      "PShareCuts"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": false,
    "matrix_rows": [
      "TestOneSeedOneProgram",
      "TestGeneratedProgramsCheckAndLower",
      "TestRegexProgramsPassTheChecker",
      "TestOctoberFeaturesAppear",
      "TestShrinkKeepsOnlyWhatFails",
      "TestSharedSliceCutsShare",
      "TestOwnershipShapes",
      "TestJudgeReadsThePanicLine",
      "TestOverridesShapesAndLower",
      "TestSourceTreeCutsAtItems",
      "TestReduceKeepsTheSignature",
      "TestExecuteCPUHelper",
      "TestExecuteCPULimit",
      "TestFuzzerSharesRuntimeLibrary",
      "TestUndefinedNumbersShapes"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M11 ADAMIC_MUTANT=M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M11.log 2>&1; fuzz_test.go:251: seeds 1 to 8 never generated new OwnMarked(",
    "entry_probes": [
      "PLower"
    ],
    "vacuous_entry": "lower.Lower returns nil IR with nil error; vocabulary assertions remain active.",
    "evidence_details": {
      "command": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M11 ADAMIC_MUTANT=M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M11.log 2>&1",
      "line": "fuzz_test.go:251: seeds 1 to 8 never generated new OwnMarked(",
      "log": "review/test-audit/internal-fuzz/M11.log"
    }
  },
  {
    "test": "TestJudgeReadsThePanicLine",
    "package": "internal/fuzz",
    "file": "internal/fuzz/judge_test.go:15",
    "seconds": 0.007,
    "oracle": "Self-written synthetic Run outcomes and expected Finding/Checked labels; does not run Node.",
    "oracle_kind": "self",
    "kills": [
      "M06",
      "M08"
    ],
    "unique_kills": [
      "M06"
    ],
    "last_proven_fail": "M08: judge_test.go:46: undefined put back after a narrowing: finding (javascript backend exit code differs, native exit code differs), want checked",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 16,
    "eligible_mutants": 15,
    "probe_kills": [
      "PJudge"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestOneSeedOneProgram",
      "TestGeneratedProgramsCheckAndLower",
      "TestRegexProgramsPassTheChecker",
      "TestOctoberFeaturesAppear",
      "TestShrinkKeepsOnlyWhatFails",
      "TestSharedSliceCutsShare",
      "TestOwnershipShapes",
      "TestJudgeReadsThePanicLine",
      "TestOverridesShapesAndLower",
      "TestSourceTreeCutsAtItems",
      "TestReduceKeepsTheSignature",
      "TestExecuteCPUHelper",
      "TestExecuteCPULimit",
      "TestFuzzerSharesRuntimeLibrary",
      "TestUndefinedNumbersShapes"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M08 ADAMIC_MUTANT=M08 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M08.log 2>&1; judge_test.go:46: undefined put back after a narrowing: finding (javascript backend exit code differs, native exit code differs), want checked",
    "entry_probes": [
      "PJudge"
    ],
    "evidence_details": {
      "command": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M08 ADAMIC_MUTANT=M08 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M08.log 2>&1",
      "line": "judge_test.go:46: undefined put back after a narrowing: finding (javascript backend exit code differs, native exit code differs), want checked",
      "log": "review/test-audit/internal-fuzz/M08.log"
    }
  },
  {
    "test": "TestOverridesShapesAndLower",
    "package": "internal/fuzz",
    "file": "internal/fuzz/overrides_test.go:15",
    "seconds": 0.834,
    "oracle": "Self-written effect/argument vocabulary across 50 seeds plus checker/lowering success; no native override execution and no returned IR assertion.",
    "oracle_kind": "self",
    "kills": [
      "M02",
      "M10",
      "M13"
    ],
    "unique_kills": [
      "M13"
    ],
    "last_proven_fail": "M13: overrides_test.go:49: 50 seeds never generated worker.run({ ...borrowed,",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 16,
    "eligible_mutants": 15,
    "probe_kills": [
      "PWithout",
      "PFeatures",
      "PShareCuts"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": false,
    "matrix_rows": [
      "TestOneSeedOneProgram",
      "TestGeneratedProgramsCheckAndLower",
      "TestRegexProgramsPassTheChecker",
      "TestOctoberFeaturesAppear",
      "TestShrinkKeepsOnlyWhatFails",
      "TestSharedSliceCutsShare",
      "TestOwnershipShapes",
      "TestJudgeReadsThePanicLine",
      "TestOverridesShapesAndLower",
      "TestSourceTreeCutsAtItems",
      "TestReduceKeepsTheSignature",
      "TestExecuteCPUHelper",
      "TestExecuteCPULimit",
      "TestFuzzerSharesRuntimeLibrary",
      "TestUndefinedNumbersShapes"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M13 ADAMIC_MUTANT=M13 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M13.log 2>&1; overrides_test.go:49: 50 seeds never generated worker.run({ ...borrowed,",
    "entry_probes": [
      "PLower"
    ],
    "vacuous_entry": "lower.Lower returns nil IR with nil error; vocabulary assertions remain active.",
    "evidence_details": {
      "command": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M13 ADAMIC_MUTANT=M13 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M13.log 2>&1",
      "line": "overrides_test.go:49: 50 seeds never generated worker.run({ ...borrowed,",
      "log": "review/test-audit/internal-fuzz/M13.log"
    }
  },
  {
    "test": "TestSourceTreeCutsAtItems",
    "package": "internal/fuzz",
    "file": "internal/fuzz/reduce_test.go:10",
    "seconds": 0.007,
    "oracle": "Self-written exact source strings for six parsed-list deletions; no external specification checked.",
    "oracle_kind": "self",
    "kills": [
      "M04",
      "M09"
    ],
    "unique_kills": [
      "M04"
    ],
    "last_proven_fail": "M09: reduce_test.go:37: without 1 [1, 3): \"f(a,);\\n[1, 2];\\n\", want \"f(a);\\n[1, 2];\\n\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 16,
    "eligible_mutants": 15,
    "probe_kills": [
      "PSourceCut"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestOneSeedOneProgram",
      "TestGeneratedProgramsCheckAndLower",
      "TestRegexProgramsPassTheChecker",
      "TestOctoberFeaturesAppear",
      "TestShrinkKeepsOnlyWhatFails",
      "TestSharedSliceCutsShare",
      "TestOwnershipShapes",
      "TestJudgeReadsThePanicLine",
      "TestOverridesShapesAndLower",
      "TestSourceTreeCutsAtItems",
      "TestReduceKeepsTheSignature",
      "TestExecuteCPUHelper",
      "TestExecuteCPULimit",
      "TestFuzzerSharesRuntimeLibrary",
      "TestUndefinedNumbersShapes"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M09 ADAMIC_MUTANT=M09 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M09.log 2>&1; reduce_test.go:37: without 1 [1, 3): \"f(a,);\\n[1, 2];\\n\", want \"f(a);\\n[1, 2];\\n\"",
    "entry_probes": [
      "PSourceCut"
    ],
    "diagnostic_kills": [
      "PSource: constructor preparation diagnostic, not a code-entry probe"
    ],
    "evidence_details": {
      "command": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M09 ADAMIC_MUTANT=M09 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M09.log 2>&1",
      "line": "reduce_test.go:37: without 1 [1, 3): \"f(a,);\\n[1, 2];\\n\", want \"f(a);\\n[1, 2];\\n\"",
      "log": "review/test-audit/internal-fuzz/M09.log"
    }
  },
  {
    "test": "TestReduceKeepsTheSignature",
    "package": "internal/fuzz",
    "file": "internal/fuzz/reduce_test.go:69",
    "seconds": 2.624,
    "oracle": "Self-written exact non-null refusal prefix and reduced minimum, observed against own compiler; tests reducer preservation, not a witness whose sole purpose is proving the agreement checker can fail. Supplemental weakened keeps comparison survives.",
    "oracle_kind": "self",
    "kills": [
      "M09",
      "M15"
    ],
    "unique_kills": [],
    "last_proven_fail": "M15: reduce_test.go:96: reduced to",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestShrinkKeepsOnlyWhatFails",
      "TestSourceTreeCutsAtItems"
    ],
    "mutants_in_matrix": 16,
    "eligible_mutants": 15,
    "probe_kills": [
      "PReduce",
      "PExecute",
      "PSourceCut",
      "PPrepare",
      "PLower"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestOneSeedOneProgram",
      "TestGeneratedProgramsCheckAndLower",
      "TestRegexProgramsPassTheChecker",
      "TestOctoberFeaturesAppear",
      "TestShrinkKeepsOnlyWhatFails",
      "TestSharedSliceCutsShare",
      "TestOwnershipShapes",
      "TestJudgeReadsThePanicLine",
      "TestOverridesShapesAndLower",
      "TestSourceTreeCutsAtItems",
      "TestReduceKeepsTheSignature",
      "TestExecuteCPUHelper",
      "TestExecuteCPULimit",
      "TestFuzzerSharesRuntimeLibrary",
      "TestUndefinedNumbersShapes"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M15 ADAMIC_MUTANT=M15 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M15.log 2>&1; reduce_test.go:96: reduced to",
    "entry_probes": [
      "PReduce"
    ],
    "diagnostic_kills": [
      "PSource: constructor preparation diagnostic, not a code-entry probe"
    ],
    "evidence_details": {
      "command": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M15 ADAMIC_MUTANT=M15 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M15.log 2>&1",
      "line": "reduce_test.go:96: reduced to",
      "log": "review/test-audit/internal-fuzz/M15.log"
    }
  },
  {
    "test": "TestExecuteCPUHelper",
    "package": "internal/fuzz",
    "file": "internal/fuzz/run_test.go:15",
    "seconds": 0.007,
    "oracle": "Only subprocess entry, returns immediately unless parent sets ADAMIC_FUZZ_CPU_HELPER; parent TestExecuteCPULimit.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "helper",
    "subsumed_by": [],
    "mutants_in_matrix": 16,
    "eligible_mutants": 15,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": false,
    "matrix_rows": [
      "TestOneSeedOneProgram",
      "TestGeneratedProgramsCheckAndLower",
      "TestRegexProgramsPassTheChecker",
      "TestOctoberFeaturesAppear",
      "TestShrinkKeepsOnlyWhatFails",
      "TestSharedSliceCutsShare",
      "TestOwnershipShapes",
      "TestJudgeReadsThePanicLine",
      "TestOverridesShapesAndLower",
      "TestSourceTreeCutsAtItems",
      "TestReduceKeepsTheSignature",
      "TestExecuteCPUHelper",
      "TestExecuteCPULimit",
      "TestFuzzerSharesRuntimeLibrary",
      "TestUndefinedNumbersShapes"
    ],
    "evidence": "16 planned whole-package runs; see runs.json; Parent: TestExecuteCPULimit",
    "entry_probes": [],
    "evidence_details": {
      "command": "16 planned whole-package runs; see runs.json",
      "line": "Parent: TestExecuteCPULimit"
    }
  },
  {
    "test": "TestExecuteCPULimit",
    "package": "internal/fuzz",
    "file": "internal/fuzz/run_test.go:36",
    "seconds": 4.391,
    "oracle": "Self-written TimedOut and exit expectations over Go CPU-helper and shell subprocesses; kernel CPU behavior exercised, no external copied authority. SIGXCPU check asserts timeout classification only.",
    "oracle_kind": "self",
    "kills": [
      "M07"
    ],
    "unique_kills": [
      "M07"
    ],
    "last_proven_fail": "M07: run_test.go:57: SIGXCPU death was not timed out: {Stdout:[] Stderr:[] ExitCode:-1 TimedOut:false}",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 16,
    "eligible_mutants": 15,
    "probe_kills": [
      "PExecute"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestOneSeedOneProgram",
      "TestGeneratedProgramsCheckAndLower",
      "TestRegexProgramsPassTheChecker",
      "TestOctoberFeaturesAppear",
      "TestShrinkKeepsOnlyWhatFails",
      "TestSharedSliceCutsShare",
      "TestOwnershipShapes",
      "TestJudgeReadsThePanicLine",
      "TestOverridesShapesAndLower",
      "TestSourceTreeCutsAtItems",
      "TestReduceKeepsTheSignature",
      "TestExecuteCPUHelper",
      "TestExecuteCPULimit",
      "TestFuzzerSharesRuntimeLibrary",
      "TestUndefinedNumbersShapes"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M07 ADAMIC_MUTANT=M07 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M07.log 2>&1; run_test.go:57: SIGXCPU death was not timed out: {Stdout:[] Stderr:[] ExitCode:-1 TimedOut:false}",
    "entry_probes": [
      "PExecute"
    ],
    "vacuous_subcases": [
      "saturated: PExecute returns Run{}; this positive subcase still passes"
    ],
    "evidence_details": {
      "command": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M07 ADAMIC_MUTANT=M07 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M07.log 2>&1",
      "line": "run_test.go:57: SIGXCPU death was not timed out: {Stdout:[] Stderr:[] ExitCode:-1 TimedOut:false}",
      "log": "review/test-audit/internal-fuzz/M07.log"
    }
  },
  {
    "test": "TestFuzzerSharesRuntimeLibrary",
    "package": "internal/fuzz",
    "file": "internal/fuzz/runtime_test.go:10",
    "seconds": 2.388,
    "oracle": "Source Node versus native and JavaScript backend via Try, with byte/exit comparison and sanitizers; also self-written runtime path equality and shared runtime stdout.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M08"
    ],
    "unique_kills": [],
    "last_proven_fail": "M08: runtime_test.go:32: finding: javascript backend stdout differs, native stdout differs",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestJudgeReadsThePanicLine"
    ],
    "mutants_in_matrix": 16,
    "eligible_mutants": 15,
    "probe_kills": [
      "PJudge",
      "PExecute",
      "PTry",
      "PPrepare",
      "PLower"
    ],
    "subsumer_seconds": 0.007,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestOneSeedOneProgram",
      "TestGeneratedProgramsCheckAndLower",
      "TestRegexProgramsPassTheChecker",
      "TestOctoberFeaturesAppear",
      "TestShrinkKeepsOnlyWhatFails",
      "TestSharedSliceCutsShare",
      "TestOwnershipShapes",
      "TestJudgeReadsThePanicLine",
      "TestOverridesShapesAndLower",
      "TestSourceTreeCutsAtItems",
      "TestReduceKeepsTheSignature",
      "TestExecuteCPUHelper",
      "TestExecuteCPULimit",
      "TestFuzzerSharesRuntimeLibrary",
      "TestUndefinedNumbersShapes"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M08 ADAMIC_MUTANT=M08 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M08.log 2>&1; runtime_test.go:32: finding: javascript backend stdout differs, native stdout differs",
    "entry_probes": [
      "PTry",
      "PPrepare"
    ],
    "evidence_details": {
      "command": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M08 ADAMIC_MUTANT=M08 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M08.log 2>&1",
      "line": "runtime_test.go:32: finding: javascript backend stdout differs, native stdout differs",
      "log": "review/test-audit/internal-fuzz/M08.log"
    }
  },
  {
    "test": "TestUndefinedNumbersShapes",
    "package": "internal/fuzz",
    "file": "internal/fuzz/undefined_numbers_test.go:21",
    "seconds": 1.613,
    "oracle": "Self-written source/shape vocabulary and opt-in expectations across 40 seeds plus checker/lowering success; no execution or returned IR assertion. Vocabulary derives its shape list from same generator, so shared omissions can agree.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02",
      "M10"
    ],
    "unique_kills": [
      "M01"
    ],
    "last_proven_fail": "M10: undefined_numbers_test.go:58: seed 21: the checker refused it: /tmp/adamic-gate/TestUndefinedNumbersShapesseeds-021-0251459133209/001/program.a:100:49: error TS2362: The left-hand side of an arithmetic operation must be of type 'any', 'number', 'bigint' or an enum type.",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 16,
    "eligible_mutants": 15,
    "probe_kills": [
      "PGenerate",
      "PWithout",
      "PFeatures",
      "PShareCuts"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": false,
    "matrix_rows": [
      "TestOneSeedOneProgram",
      "TestGeneratedProgramsCheckAndLower",
      "TestRegexProgramsPassTheChecker",
      "TestOctoberFeaturesAppear",
      "TestShrinkKeepsOnlyWhatFails",
      "TestSharedSliceCutsShare",
      "TestOwnershipShapes",
      "TestJudgeReadsThePanicLine",
      "TestOverridesShapesAndLower",
      "TestSourceTreeCutsAtItems",
      "TestReduceKeepsTheSignature",
      "TestExecuteCPUHelper",
      "TestExecuteCPULimit",
      "TestFuzzerSharesRuntimeLibrary",
      "TestUndefinedNumbersShapes"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M10 ADAMIC_MUTANT=M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M10.log 2>&1; undefined_numbers_test.go:58: seed 21: the checker refused it: /tmp/adamic-gate/TestUndefinedNumbersShapesseeds-021-0251459133209/001/program.a:100:49: error TS2362: The left-hand side of an arithmetic operation must be of type 'any', 'number', 'bigint' or an enum type.",
    "entry_probes": [
      "PLower"
    ],
    "vacuous_entry": "lower.Lower returns nil IR with nil error; vocabulary assertions remain active.",
    "evidence_details": {
      "command": "ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/M10 ADAMIC_MUTANT=M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run . > review/test-audit/internal-fuzz/M10.log 2>&1",
      "line": "undefined_numbers_test.go:58: seed 21: the checker refused it: /tmp/adamic-gate/TestUndefinedNumbersShapesseeds-021-0251459133209/001/program.a:100:49: error TS2362: The left-hand side of an arithmetic operation must be of type 'any', 'number', 'bigint' or an enum type.",
      "log": "review/test-audit/internal-fuzz/M10.log"
    }
  }
]

Mutants (origin/main lines):

| ID | Location | Menu/change | Failed rows |
|---|---|---|---|
| M01 | internal/fuzz/generate.go:74 | change constant: rand.NewPCG(seed, 0x61646d6963) -> rand.NewPCG(0, 0x61646d6963) | TestUndefinedNumbersShapes |
| M02 | internal/fuzz/generate.go:86 | flip condition: return !g.without[feature] -> return g.without[feature] | TestOctoberFeaturesAppear, TestSharedSliceCutsShare, TestOwnershipShapes, TestOverridesShapesAndLower, TestUndefinedNumbersShapes |
| M03 | internal/fuzz/shrink.go:22 | change constant: }, 1, 3000) -> }, 1, 0) | TestShrinkKeepsOnlyWhatFails |
| M04 | internal/fuzz/source.go:227 | change option: return []namedList{{node.ArgumentList(), true}} -> return []namedList{{node.ArgumentList(), false}} | TestSourceTreeCutsAtItems |
| M05 | internal/fuzz/reduce.go:58 | flip condition: if s.Exact { -> if !s.Exact { |  |
| M06 | internal/fuzz/run.go:200 | change constant: outcome.Native.ExitCode == 70 -> outcome.Native.ExitCode == 71 | TestJudgeReadsThePanicLine |
| M07 | internal/fuzz/run.go:324 | change constant: status.Signal() == syscall.SIGXCPU -> status.Signal() == syscall.SIGTERM | TestExecuteCPULimit |
| M08 | internal/fuzz/run.go:272 | flip condition: case !bytes.Equal(expected.Stdout, actual.Stdout): -> case bytes.Equal(expected.Stdout, actual.Stdout): | TestJudgeReadsThePanicLine, TestFuzzerSharesRuntimeLibrary |
| M09 | internal/fuzz/source.go:345 | drop statement: cut.start = items.ends[start-1] ->  | TestSourceTreeCutsAtItems, TestReduceKeepsTheSignature |
| M10 | internal/fuzz/generate.go:167 | drop statement: g.names++ ->  | TestGeneratedProgramsCheckAndLower, TestRegexProgramsPassTheChecker, TestOwnershipShapes, TestOverridesShapesAndLower, TestUndefinedNumbersShapes |
| M11 | internal/fuzz/ownership.go:16 | change constant: switch g.seed % 8 { -> switch g.seed % 7 { | TestOwnershipShapes |
| M12 | internal/fuzz/shrink.go:297 | change constant: return text(Number, "0") -> return text(Number, "1") |  |
| M13 | internal/fuzz/overrides.go:43 | change constant: switch (g.seed / 10) % 5 { -> switch (g.seed / 10) % 4 { | TestOverridesShapesAndLower |
| M14 | internal/fuzz/generate.go:1248 | off-by-one bound: return sliceBytes >= 64 && sliceBytes >= ownerBytes/4 -> return sliceBytes >= 63 && sliceBytes >= ownerBytes/4 |  |
| M15 | internal/fuzz/shrink.go:103 | drop statement: s.current = candidates[index] ->  | TestShrinkKeepsOnlyWhatFails, TestReduceKeepsTheSignature |
| M16 | internal/fuzz/reduce.go:272 | drop comparison, weakened-check witness candidate: && candidate.Has(signature) -> && true |  |

Survivor witnesses are in clean-witness.log and M05/M12/M14/M16-witness.log.


Survivors, with this-session behavior witnesses:

- M05: Exact=true, Text=abc, line abc-suffix matches false before and true after. Exact matching is unguarded by this matrix.
- M12: simplest(Number).String() returns 0 before and 1 after. This output choice is unguarded; both are valid numeric simplifications, so no claim that 1 is an incorrect literal.
- M14: bytesShared(63,128) returns false before and true after. The boundary is unguarded by this matrix; the copied rule is itself older than the runtime.
- M16 (supplemental): keeps accepts a different refusal signature false before and true after. Removing this comparison is not a whole-statement drop, so no verdict rests on it. The real reducer test passes despite the loosened check.

Brief ambiguities, mismatches, and costs:

- No “Your rows” section was supplied. Scope is all 15 names actually listed by go test at the starting commit. None moved or vanished relative to a supplied list because there was no such list to compare. Subtests are not separate top-level rows. No top-level wrappers differ only by inputs to one checker; vocabulary and lowering assertions keep the similarly shaped rows distinct.
- The warm env.sh worked, but required submodules were missing. npm ci succeeded (3 packages, 603ms). Required recursive submodule initialization completed within its 90-second cap; its exact elapsed duration was not recorded. Initial cold test discovery hit 90 seconds and was stopped. The cached retry succeeded; baseline package binary time was 18.779s. No setup script was needed.
- The regex-named checker row calls the same full Generate as the general checker/lower row, rather than restricting generation to regex features. Its observed one-mutant subsumption is evidence for this small matrix, not a deletion recommendation.
- The code under test includes the fuzzer's production verdict classifier. TestJudgeReadsThePanicLine supplies synthetic expected labels as its self oracle; those expectations were never edited. Shrink and Reduce test reducer products with predicates/refused programs, rather than being tests solely intended to prove an agreement harness can fail. None was classified as a witness or suite-construction setup check.
- “Own entry” is ambiguous for rows checking both generation and lowering. The four true vacuous values apply specifically to Lower: nil IR plus nil error passes the whole row. Their generator entry probes do fail. entry_probes and vacuous_entry in audit.json preserve that distinction. PSource is a constructor preparation diagnostic only; PSourceCut, not PSource, judges the cut operation. No production verdict rests on a probe.
- The positive saturated CPU subcase passes PExecute's empty Run, while infinite and SIGXCPU fail. The row is not vacuous; this subcase is recorded explicitly.
- Fuzzer sharing geometry does not directly exercise runtime sharing. The helper's 64-byte/quarter-owner rule differs from current string_share.c's storage/header ratio 8 and no minimum. Both source versions were read in this session. No outside specification value was checked and no external-authority oracle is claimed.
- Three mutants per 15 rows would exceed the 20-mutant cap. The fixed matrix used 15 menu mutations and one supplemental comparison removal. Every subsumption conclusion rests on one killed mutant; overlapping reduction rests on two. M16 survived and is excluded from verdicts.
- OneSeedOneProgram passed every production mutant, including random-seed and name-counter changes. “Untrue” means no failure observed under this finite honest try, not a universal proof that determinism can never fail. Its own empty generator probe does fail.
- Go parent PASS durations omit time spent in parallel descendants. Medians use each isolated invocation's ok package-binary line, with -count=1 on all three runs, not shell elapsed or parent duration. nproc=5.
- All raw test output is in files. Panicking probes were rerun individually for every row; no unexecuted row was counted as passing or failing. All production mutant runs completed within the binary budget. The only cooked step was cold discovery compilation. Clean coverage verification also passed: 10.737s, 84.9% package statement coverage.
- Before mutations, function-inventory.txt recorded all production functions conservatively, including unreached ones. reached-functions.txt and coverage.out then identify measured reachability. Neither file claims coverage of other packages' complete function sets.

Timing and scope:

- Three-run isolated timing commands: 130.801s total shell elapsed. Row medians are binary times in audit.json.
- Matrix and initial probes, including all individual panic reruns and additional entry probes: 521.157s shell elapsed. Lower entry probe: 21.717s shell elapsed.
- Standalone mutant apply-check and Go vet validations: 6.361s shell elapsed; all 16 passed. switch-vet.log and additional-probes-vet.log also passed.
- Builds inside Prepare and runtime-library compilation are included in run elapsed, not separately instrumented. No .a/.ts port was mutated or rebuilt. Every run had ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/<id>; the lowering probe used its own cache too.
- Not covered: repo-wide uniqueness, other packages, every possible mutant, runtime correctness of the generated vocabulary scenes, or an independent tsc/ECMAScript authority check. Production files and temporary witness tests were restored. No PR or main push.
