not green

Pushed branch: cloud/integrate-stack-q4
Pushed SHA: dc135a74f70de5016f3025d876bc96920dc3fb0a
Exact first-observed q3 base: b91dc0ebf8a7c776b20d07fbdee7fe1fdb84b88e

Merge sources, in order:
- area/stage1-format: 196e821a98d699158bd7394b8953161dcb81ff63
- stage3-progress/2026-10-07: eb4143953295676dd7f7045810f30ee89da91ffc
- devtools/census-main: 01cfcc238f282bc0da1e9992f1f7860a3deaa31b

Conflict resolutions:
- stage1/cohere/markdownblocks/width_test.go: retained base fatal dependency preflight and incoming Node width mutants, with sanitized-native width mutant comparisons retained too.
- internal/skipcensus/log.go: incoming provision explanations retained.
- internal/skipcensus/census_test.go: all proofs retained; historical declarations append only absent (file, ID) pairs. Historical mismatched-tree assertion still fails because the merged inventory classifies all 25 historical skips.
- internal/skipcensus/testdata/skips.json: union of both inventories, base metadata for shared sites; no reclassification to hide drift.
- stage3/progress.json merged cleanly; all four October 9 milestones remain false. counts.md had no conflict.

Cohere submodule: 7945d102a6c18dd36adf9114a758ce646e8b2359
Node: v24.19.0; Prettier: 3.9.6; yaml-unist-parser: 3.2.0; markdown-width pins: emoji-regex 10.6.0, get-east-asian-width 1.6.0, narrow-emojis 0.0.3.
gofmt -l cmd internal: empty output, exit 0. go vet ./...: exit 0.
Disk TMPDIR: /workspace/adamic-gate-tmp, mode 1777. ADAMIC_GATE_UNCACHED=1 ADAMIC_TEST_WASI=1 ADAMIC_ORACLE_WASI=1.

JSON terminal-event counts (nested test events included; no-test-files packages counted separately):
| Run | Test pass | Test fail | Test skip | Package pass | Package fail | Package skip |
|---|---:|---:|---:|---:|---:|---:|
| order-check.jsonl | 1613 | 0 | 39 | 3 | 0 | 0 |
| q4-whole.jsonl | 185 | 0 | 0 | 6 | 0 | 14 |
| q4-whole-restart.jsonl | 8371 | 35 | 72 | 40 | 10 | 24 |
| q4-rerun-cmd-adamic-gate.jsonl | 48 | 2 | 0 | 0 | 1 | 0 |
| q4-rerun-internal-fuzz.jsonl | 30 | 2 | 0 | 0 | 1 | 0 |
| q4-rerun-stage1-cohere-estree.jsonl | 6 | 0 | 2 | 0 | 0 | 0 |

Command timings and exits:
```json
{
  "order": {
    "exit": 0,
    "wall_seconds": 130.022,
    "command": [
      "go",
      "test",
      "-count=1",
      "-json",
      "./internal/ir",
      "./cmd/adamic",
      "./internal/flow"
    ],
    "log": "/workspace/order-check.jsonl"
  },
  "gofmt": {
    "exit": 0,
    "wall_seconds": 0.362,
    "command": [
      "gofmt",
      "-l",
      "cmd",
      "internal"
    ],
    "log": "/workspace/q4-gofmt.log"
  },
  "vet": {
    "exit": 0,
    "wall_seconds": 0.429,
    "command": [
      "go",
      "vet",
      "./..."
    ],
    "log": "/workspace/q4-vet.log"
  },
  "whole": {
    "exit": 1,
    "wall_seconds": 6590.723,
    "command": [
      "go",
      "test",
      "-count=1",
      "-timeout",
      "60m",
      "-json",
      "./stage1/cohere/markdownblocks",
      "./stage1/cohere/lint",
      "./stage1/cohere/css",
      "./stage1/cohere/typeaware",
      "./internal/unicodeproperties",
      "./internal/oracle",
      "./stage1/cohere/json",
      "./stage1/cohere/markdowninline",
      "./stage1/typescript/parser",
      "./internal/native",
      "./stage1/cohere/cssnumbers",
      "./stage1/cohere/cssstrings",
      "./stage1/cohere/lint/helpers/comments",
      "./internal/flow",
      "./bridge/tsgo",
      "./stage1/cohere/lint/helpers",
      "./stage1/typescript/scanner",
      "./cmd/adamic-test262",
      "./internal/fresh",
      "./stage1/cohere/graphql",
      "./internal/lower",
      "./stage1/cohere/formatfiles",
      "./stage1/cohere/gitignore",
      "./stage1/cohere/lint/inventory",
      "./stage1/cohere/selector",
      "./stage1/cohere/values",
      "./stage1/cohere/suppression",
      "./internal/fuzz",
      "./cloud",
      "./stage3/fixtures",
      "./stage1/cohere/mediaquery",
      "./internal/regexp",
      "./cmd/adamic",
      "./cmd/adamic-meter",
      "./internal/load",
      "./internal/ir",
      "./cmd/adamic-gate",
      "./cmd/adamic-stage1-progress",
      "./stage1/cohere/lint/registry",
      "./bridge/tsgo/checker",
      "./bench",
      "./bench/regex",
      "./bridge/tsgo/archive",
      "./bridge/tsgo/cost",
      "./bridge/tsgo/oracle",
      "./bridge/tsgo/spec",
      "./cmd/adamic-fuzz",
      "./cmd/adamic-reduce",
      "./cmd/lint-registry",
      "./internal/javascript",
      "./stage1/cohere/markdownblocks/tools/generate_classes",
      "./stage1/cohere/markdownblocks/tools/generate_entities",
      "./stage1/cohere/markdownblocks/tools/generate_width",
      "./stage3/census/latent/tool",
      "./stage3/census/tool",
      "./stage3/drivers/parser/probe",
      "./cloud/reports/decode-ascii",
      "./cloud/reports/release-lto",
      "./cmd/adamic-lint-check",
      "./cmd/adamic-metamorphic",
      "./cmd/adamic-refusals",
      "./internal/boundedrun",
      "./internal/boundedrun/testfixture",
      "./internal/leakcheck",
      "./internal/metamorphic",
      "./internal/nodepin",
      "./internal/refusalprobe",
      "./internal/skipcensus",
      "./internal/skipcensus/cmd",
      "./stage1/cohere/estree",
      "./stage1/cohere/graphql/printer",
      "./stage1/cohere/tsprinter",
      "./stage1/cohere/yaml",
      "./verify/coverage/runtimelibrary"
    ],
    "log": "/workspace/q4-whole-restart.jsonl"
  },
  "rerun:./cmd/adamic-gate": {
    "exit": 1,
    "wall_seconds": 45.997,
    "command": [
      "go",
      "test",
      "-count=1",
      "-timeout",
      "60m",
      "-json",
      "./cmd/adamic-gate"
    ],
    "log": "/workspace/q4-rerun-cmd-adamic-gate.jsonl"
  },
  "rerun:./internal/fuzz": {
    "exit": 1,
    "wall_seconds": 50.105,
    "command": [
      "go",
      "test",
      "-count=1",
      "-timeout",
      "60m",
      "-json",
      "./internal/fuzz"
    ],
    "log": "/workspace/q4-rerun-internal-fuzz.jsonl"
  }
}
```
Overall corrected-run wall time: pending seconds.

Required test counts and output:
q4-whole-restart.jsonl: github.com/system-inc/adamic/stage1/cohere/lint/TestCompilerAndStage1Agree pass 175.14 seconds
```text
=== RUN   TestCompilerAndStage1Agree
    lint_test.go:528: compiler and stage1: 551 files
    lint_test.go:137: clang build: 60.924s; split=false jobs=0 flags=[-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -pthread -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all]
    lint_test.go:529: Go, Node, emitted JavaScript, native identical: 22885255 bytes
--- PASS: TestCompilerAndStage1Agree (175.14s)
```
q4-whole-restart.jsonl: github.com/system-inc/adamic/stage1/cohere/tsprinter/TestStatementsAgainstGoAndPrettier pass 177.29 seconds
```text
=== RUN   TestStatementsAgainstGoAndPrettier
    statements_test.go:83: === RUN   TestAdamicStatementCorpus
            adamic_statements_test.go:141: 816 files, 0 parse failures, 56759 supported statement/program fragments, 13 complete files
        --- PASS: TestAdamicStatementCorpus (3.04s)
        PASS
        ok  	github.com/system-inc/cohere/internal/format/javascript	3.054s
        
    statements_test.go:141: 56759 statement/program fragments byte-identical
--- PASS: TestStatementsAgainstGoAndPrettier (177.29s)
```

Every gate cache line:
q4-whole-restart.jsonl: gate cache: native hits=0 misses=2882
q4-whole-restart.jsonl: gate cache: node hits=0 misses=2605
q4-whole-restart.jsonl: gate cache: probe hits=0 misses=28

Every skipped test, with its own logged reason:

order-check.jsonl
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_coverage_unions.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_typeof.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_throw.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_plain.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_nested.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_three.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_coverage_values.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_coverage_typeof.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_coverage_reject_nested.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_coverage_reject_empty.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_coverage_reject_eager.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_coverage_parameters.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_coverage_discard.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_typeof.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_three.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_throw.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_plain.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_coverage_values.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_nested.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_coverage_reject_empty.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_coverage_typeof.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_coverage_reject_nested.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_coverage_unions.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_coverage_reject_eager.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_coverage_discard.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_coverage_parameters.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_typeof.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_throw.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_three.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_plain.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_coverage_unions.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_nested.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_coverage_values.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_coverage_typeof.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_coverage_reject_nested.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_coverage_reject_empty.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_coverage_reject_eager.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_coverage_discard.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_coverage_parameters.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle

q4-whole-restart.jsonl
- github.com/system-inc/adamic/stage1/cohere/css/TestCSSThroughput: css_test.go:405: set ADAMIC_CSS_BENCH=1 for throughput
- github.com/system-inc/adamic/internal/oracle/TestGCCAgreesWithNode: gcc_lane_test.go:21: opt in with ADAMIC_GCC_LANE=1
- github.com/system-inc/adamic/internal/oracle/TestNativeReleaseFlagsAgreeWithNode: release_flags_test.go:18: set ADAMIC_RELEASE_LANE=1 for the full shipping-build lane
- github.com/system-inc/adamic/internal/oracle/TestNativeExistingReleaseAgreesWithNode: release_flags_test.go:28: set ADAMIC_RELEASE_MEASURE=1 for the timing control
- github.com/system-inc/adamic/internal/oracle/TestReleaseAgreesWithNode: release_test.go:16: set ADAMIC_ORACLE_RELEASE=1 for the shipped release oracle
- github.com/system-inc/adamic/internal/oracle/TestReleaseOracleCatchesOneByte: release_test.go:59: set ADAMIC_ORACLE_RELEASE=1
- github.com/system-inc/adamic/internal/oracle/TestNativeMallocAgreesWithNode: slabs_test.go:30: set ADAMIC_SLAB_MEASURE=1 for the timing control
- github.com/system-inc/adamic/stage1/cohere/lint/TestJsxLintReleaseAndThroughput: jsx_integration_test.go:74: set ADAMIC_LINT_BENCH=1 for JSX throughput
- github.com/system-inc/adamic/stage1/cohere/json/TestProfileSnapshotsAgree: performance_test.go:16: set ADAMIC_JSON_PROFILE_BINARIES to profile snapshot binaries
- github.com/system-inc/adamic/stage1/cohere/lint/TestThroughput: lint_test.go:666: set ADAMIC_LINT_BENCH=1
- github.com/system-inc/adamic/stage1/cohere/css/TestCSSPrinterThroughput: print_test.go:204: set ADAMIC_CSS_PRINTER_BENCH=1
- github.com/system-inc/adamic/stage1/cohere/css/TestCSSProfileArtifacts: profile_test.go:19: set ADAMIC_CSS_PROFILE_DIR
- github.com/system-inc/adamic/stage1/cohere/css/TestCSSProfileSnapshotsAgree: profile_test.go:134: set ADAMIC_CSS_PROFILE_SNAPSHOTS
- github.com/system-inc/adamic/internal/oracle/TestStage3FixtureHook: stage3_hook_test.go:15: called only by the Stage 3 fixture runner
- github.com/system-inc/adamic/internal/oracle/TestLongArgumentsLeaveTheStackItsLimit/a_1_MiB_stack_and_400_KB_of_arguments: stack_test.go:48: Linux refuses arguments over a quarter of the stack, so this can't arise there
- github.com/system-inc/adamic/internal/oracle/TestLongArgumentsLeaveTheStackItsLimit/a_2_MiB_stack_and_900_KB_of_arguments: stack_test.go:48: Linux refuses arguments over a quarter of the stack, so this can't arise there
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_typeof.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_plain.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_three.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_nested.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_throw.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_coverage_typeof.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_coverage_unions.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_coverage_values.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_coverage_reject_eager.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_coverage_reject_empty.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_coverage_reject_nested.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_coverage_parameters.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/async_coverage_discard.a: trace_test.go:36: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_throw.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_three.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_typeof.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_coverage_values.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_plain.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_nested.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_coverage_unions.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_coverage_typeof.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_coverage_reject_nested.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_coverage_reject_empty.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_coverage_parameters.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_coverage_reject_eager.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestLivenessHoldsOnEveryPath/programs/../oracle/testdata/async_coverage_discard.a: trace_test.go:390: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_coverage_unions.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_typeof.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_throw.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_nested.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_three.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_plain.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_coverage_values.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_coverage_reject_nested.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_coverage_typeof.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_coverage_reject_empty.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_coverage_discard.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_coverage_reject_eager.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/flow/TestEveryMutationIsInItsRange/programs/../oracle/testdata/async_coverage_parameters.a: ranges_test.go:28: suspension-state tracing is NotYet; these programs use no synchronous lifetime proofs and have a separate async Node oracle
- github.com/system-inc/adamic/internal/native/TestNormalizeLongMeasurements: normalize_benchmark_test.go:19: set ADAMIC_NORMALIZE_BENCH_LOG to run the long-string measurements
- github.com/system-inc/adamic/stage1/typescript/scanner/TestProfileArtifacts: profile_test.go:19: set ADAMIC_SCANNER_PROFILE_DIR to a scratch directory
- github.com/system-inc/adamic/stage1/typescript/scanner/TestProfileSnapshotsAgree: profile_test.go:99: set ADAMIC_SCANNER_PROFILE_SNAPSHOTS to the artifact directories
- github.com/system-inc/adamic/stage1/typescript/parser/TestPerformance: performance_test.go:21: set ADAMIC_PARSER_BENCH=1 for best-of-five whole compiler parsing
- github.com/system-inc/adamic/stage1/typescript/parser/TestWholePerformance: performance_test.go:21: set ADAMIC_PARSER_BENCH=1 for best-of-five whole compiler parsing
- github.com/system-inc/adamic/stage1/typescript/scanner/TestPerformance: scanner_test.go:393: set ADAMIC_SCANNER_BENCH=1 for best-of-five throughput
- github.com/system-inc/adamic/cmd/adamic-test262/TestCompilerStartupMeasurement: performance_test.go:20: measurement only
- github.com/system-inc/adamic/internal/lower/TestOptionalWideningCensus: optional_widening_census_test.go:28: set OPTIONAL_WIDENING_CONFIG and OPTIONAL_WIDENING_OUTPUT to inventory a project
- github.com/system-inc/adamic/stage1/cohere/selector/TestSelectorThroughput: selector_test.go:388: set ADAMIC_SELECTOR_BENCH=1 to time parser throughput
- github.com/system-inc/adamic/stage3/fixtures/TestTransformedNodeRunnerGuardHook: runner_guard_test.go:13: subprocess hook
- github.com/system-inc/adamic/internal/metamorphic/TestTheLeakCheckReadsTheCounts: metamorphic_test.go:132: Not on macOS: Linux's leak check is LeakSanitizer's, run on the sanitized binary
- github.com/system-inc/adamic/internal/native/TestMeasureClangUnits: units_measure_test.go:17: set ADAMIC_CLANG_MEASURE to emitted C evidence directory
- github.com/system-inc/adamic/stage1/cohere/graphql/printer/TestPrinterThroughput: performance_test.go:17: set ADAMIC_GRAPHQL_PRINTER_BENCH=1 to measure verified throughput
- github.com/system-inc/adamic/internal/native/TestRecordBenchmark: record_test.go:359: set ADAMIC_RECORD_BENCH=1 for five-round Node comparisons
- github.com/system-inc/adamic/stage1/cohere/estree/TestRepositoryAgreement: corpus_test.go:18: set ADAMIC_ESTREE_CORPUS to a completed Go/Node corpus audit directory
- github.com/system-inc/adamic/stage1/cohere/estree/TestCorpusNativeRefusals: corpus_test.go:84: completed frozen corpus required
- github.com/system-inc/adamic/stage1/cohere/estree/TestThroughput: throughput_test.go:16: set ADAMIC_ESTREE_BENCHMARK=1 for full-output throughput

q4-rerun-stage1-cohere-estree.jsonl
- github.com/system-inc/adamic/stage1/cohere/estree/TestRepositoryAgreement: corpus_test.go:18: set ADAMIC_ESTREE_CORPUS to a completed Go/Node corpus audit directory
- github.com/system-inc/adamic/stage1/cohere/estree/TestCorpusNativeRefusals: corpus_test.go:84: completed frozen corpus required

Skipped tests naming one of the 12 required variables: []

Every failure, first output lines and full output in q4-events-summary.json:

q4-whole-restart.jsonl: github.com/system-inc/adamic/stage1/cohere/typeaware/TestSixRuleAgreementAndMutants 1165.57 seconds
```text
    suite_test.go:374: want all 77 files, got 78
--- FAIL: TestSixRuleAgreementAndMutants (1165.57s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestWASIEmission/internal/oracle/testdata/error_spread.a 0.04 seconds
```text
    wasi_test.go:132: fixture does not lower
--- FAIL: TestWASIEmission/internal/oracle/testdata/error_spread.a (0.04s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestWASIEmission/internal/oracle/testdata/error_spread_view.a 0.06 seconds
```text
    wasi_test.go:132: fixture does not lower
--- FAIL: TestWASIEmission/internal/oracle/testdata/error_spread_view.a (0.06s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestWASIEmission/internal/oracle/testdata/coverage_error_construct.a 0.09 seconds
```text
    wasi_test.go:132: fixture does not lower
--- FAIL: TestWASIEmission/internal/oracle/testdata/coverage_error_construct.a (0.09s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestWASIEmission/internal/oracle/testdata/coverage_error_call.a 0.06 seconds
```text
    wasi_test.go:132: fixture does not lower
--- FAIL: TestWASIEmission/internal/oracle/testdata/coverage_error_call.a (0.06s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestWASIEmission/internal/oracle/testdata/coverage_error_mutated.a 0.04 seconds
```text
    wasi_test.go:132: fixture does not lower
--- FAIL: TestWASIEmission/internal/oracle/testdata/coverage_error_mutated.a (0.04s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestWASIEmission/internal/oracle/testdata/coverage_error_optional.a 0.05 seconds
```text
    wasi_test.go:132: fixture does not lower
--- FAIL: TestWASIEmission/internal/oracle/testdata/coverage_error_optional.a (0.05s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestWASIEmission/internal/oracle/testdata/coverage_error_view_return.a 0.03 seconds
```text
    wasi_test.go:132: fixture does not lower
--- FAIL: TestWASIEmission/internal/oracle/testdata/coverage_error_view_return.a (0.03s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestWASIEmission/internal/oracle/testdata/coverage_regexp_literal.a 0.04 seconds
```text
    wasi_test.go:132: fixture does not lower
--- FAIL: TestWASIEmission/internal/oracle/testdata/coverage_regexp_literal.a (0.04s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestWASIEmission/internal/oracle/testdata/coverage_regexp_new.a 0.07 seconds
```text
    wasi_test.go:132: fixture does not lower
--- FAIL: TestWASIEmission/internal/oracle/testdata/coverage_regexp_new.a (0.07s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestWASIEmission/internal/oracle/testdata/coverage_regexp_call.a 0.07 seconds
```text
    wasi_test.go:132: fixture does not lower
--- FAIL: TestWASIEmission/internal/oracle/testdata/coverage_regexp_call.a (0.07s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestWASIEmission/internal/oracle/testdata/coverage_regexp_view.a 0.03 seconds
```text
    wasi_test.go:132: fixture does not lower
--- FAIL: TestWASIEmission/internal/oracle/testdata/coverage_regexp_view.a (0.03s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestWASIEmission/internal/oracle/testdata/coverage_view_array.a 0.08 seconds
```text
    wasi_test.go:132: fixture does not lower
--- FAIL: TestWASIEmission/internal/oracle/testdata/coverage_view_array.a (0.08s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestWASIEmission/internal/oracle/testdata/coverage_view_number.a 0.04 seconds
```text
    wasi_test.go:132: fixture does not lower
--- FAIL: TestWASIEmission/internal/oracle/testdata/coverage_view_number.a (0.04s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestWASIEmission/internal/oracle/testdata/coverage_view_string.a 0.06 seconds
```text
    wasi_test.go:132: fixture does not lower
--- FAIL: TestWASIEmission/internal/oracle/testdata/coverage_view_string.a (0.06s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestWASIEmission/internal/oracle/testdata/coverage_view_boolean.a 0.05 seconds
```text
    wasi_test.go:132: fixture does not lower
--- FAIL: TestWASIEmission/internal/oracle/testdata/coverage_view_boolean.a (0.05s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestWASIEmission/internal/oracle/testdata/coverage_view_closure.a 0.06 seconds
```text
    wasi_test.go:132: fixture does not lower
--- FAIL: TestWASIEmission/internal/oracle/testdata/coverage_view_closure.a (0.06s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestWASIEmission/internal/oracle/testdata/reland_refused/generic_instance_key_json.a 0.08 seconds
```text
    wasi_test.go:132: fixture does not lower
--- FAIL: TestWASIEmission/internal/oracle/testdata/reland_refused/generic_instance_key_json.a (0.08s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestWASIEmission 416.92 seconds
```text
=== RUN   TestWASIEmission
--- FAIL: TestWASIEmission (416.92s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestSlabLaneCatchesRecycledRelease 0.18 seconds
```text
    slabs_test.go:119: mutant insertion points changed
--- FAIL: TestSlabLaneCatchesRecycledRelease (0.18s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle/TestReleaseBuildConfigurationAgrees 0.02 seconds
```text
    release_guard_test.go:42: release configurations differ
        shipping: options=native.Options{Compiler:"", Release:true, Target:"", Request:false, Split:false, Jobs:0, Sanitize:false, ThreadSanitize:false, Count:false, Coverage:false, cpu:"", Slabs:false, Malloc:false} flags=["-std=c11" "-Wall" "-Wextra" "-Werror" "-pedantic" "-Wno-unused-variable" "-Wno-unused-but-set-variable" "-Wno-unused-function" "-Wno-unused-parameter" "-Wno-self-assign" "-ffp-contract=off" "-fno-optimize-sibling-calls" "-pthread" "-O2" "-flto=thin"] runtime-key=25598b0473365e4b392a3b24cb655c069dd5eb66b066b92ae36724720ad0343e
        oracle: options=native.Options{Compiler:"", Release:false, Target:"", Request:false, Split:false, Jobs:0, Sanitize:false, ThreadSanitize:false, Count:false, Coverage:false, cpu:"", Slabs:false, Malloc:false} flags=["-std=c11" "-Wall" "-Wextra" "-Werror" "-pedantic" "-Wno-unused-variable" "-Wno-unused-but-set-variable" "-Wno-unused-function" "-Wno-unused-parameter" "-Wno-self-assign" "-ffp-contract=off" "-fno-optimize-sibling-calls" "-pthread" "-O2"] runtime-key=791cfbc99fd42b6cb146c5765aed3a29a078177960e5673f28441354618bfb4d
--- FAIL: TestReleaseBuildConfigurationAgrees (0.02s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/oracle 2744.396 seconds
```text
oracle: node v24.19.0 at /workspace/adamic-tools/bin/node
FAIL
gate cache: native hits=0 misses=2882
gate cache: node hits=0 misses=2605
gate cache: probe hits=0 misses=28
FAIL	github.com/system-inc/adamic/internal/oracle	2744.395s
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/stage1/cohere/lint 3600.136 seconds
```text
FAIL	github.com/system-inc/adamic/stage1/cohere/lint	3600.136s
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/stage1/cohere/typeaware 3600.064 seconds
```text
FAIL	github.com/system-inc/adamic/stage1/cohere/typeaware	3600.064s
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/native/TestRuntimeStaticsProtectionMutants/normalization_pairs 17.07 seconds
```text
    runtime_statics_parallel_test.go:69: race mutant was not caught (attempt 2): <nil>
        adamic: counts: allocations 16387 frees 16387 retains 0 releases 16387 peak 11 regions 0
        runtime statics clean
--- FAIL: TestRuntimeStaticsProtectionMutants/normalization_pairs (17.07s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/native/TestRuntimeStaticsProtectionMutants 175.42 seconds
```text
=== RUN   TestRuntimeStaticsProtectionMutants
--- FAIL: TestRuntimeStaticsProtectionMutants (175.42s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/cloud/TestRepositoryPassesCohereBaseline 44.81 seconds
```text
    cohere_gate_test.go:27: cohere baseline: exit status 1
        fatal: Needed a single revision
        cohere gate failed: Command '['git', 'rev-parse', '--verify', 'origin/main']' returned non-zero exit status 128.
        cohere gate coverage: 10526 repository files mirrored
          types: 1974 files in scope
          lint: 1974 files in scope
          format: 436 programs checked, 5571 files left out by the ignore patterns
          config-format: tsconfig.json and CohereSettings.json checked at their real paths
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/cloud 46.717 seconds
```text
FAIL
FAIL	github.com/system-inc/adamic/cloud	46.716s
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/fuzz/TestLivenessAndFieldShapes 55.82 seconds
```text
    liveness_fields_test.go:56: seed 2: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestLivenessAndFieldShapes282020665/001/program.a:9:3: Adamic 0.1 refuses cannot move items1[0]: element is not a fresh object literal with only scalar literal fields; return it through the results
    liveness_fields_test.go:56: seed 5: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestLivenessAndFieldShapes282020665/001/program.a:7:3: Adamic 0.1 refuses cannot move items1[0][]: nested array ownership is not proven; return it through the results
    liveness_fields_test.go:56: seed 10: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestLivenessAndFieldShapes282020665/001/program.a:7:3: Adamic 0.1 refuses cannot move items1[0][]: nested array ownership is not proven; return it through the results
    liveness_fields_test.go:56: seed 12: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestLivenessAndFieldShapes282020665/001/program.a:8:3: Adamic 0.1 refuses cannot move items1[0]: element is not a fresh object literal with only scalar literal fields; return it through the results
    liveness_fields_test.go:56: seed 15: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestLivenessAndFieldShapes282020665/001/program.a:8:3: Adamic 0.1 refuses cannot move items1[0]: element is not a fresh object literal with only scalar literal fields; return it through the results
    liveness_fields_test.go:56: seed 16: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestLivenessAndFieldShapes282020665/001/program.a:698:31: Adamic 0.1 refuses parallelMap items are not shareable: Bag39.tags is a mutable string[]; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)
    liveness_fields_test.go:56: seed 17: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestLivenessAndFieldShapes282020665/001/program.a:8:3: Adamic 0.1 refuses cannot move items1[0]: element is not a fresh object literal with only scalar literal fields; return it through the results
    liveness_fields_test.go:56: seed 19: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestLivenessAndFieldShapes282020665/001/program.a:9:3: Adamic 0.1 refuses cannot move items1[0]: element is not a fresh object literal with only scalar literal fields; return it through the results
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/fuzz/TestOperatorsShapes 55.83 seconds
```text
    operators_test.go:71: seed 10: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestOperatorsShapes2311601471/001/program.a:942:31: Adamic 0.1 refuses parallelMap items are not shareable: Bag36.tags is a mutable string[]; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)
    operators_test.go:71: seed 11: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestOperatorsShapes2311601471/001/program.a:860:43: Adamic 0.1 refuses parallelMap work writes state another task can reach: record25 writes the global 'total24'; keep work and everything it calls pure; write only task locals and fresh objects
    operators_test.go:71: seed 16: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestOperatorsShapes2311601471/001/program.a:826:31: Adamic 0.1 refuses parallelMap items are not shareable: Bag39.tags is a mutable string[]; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)
    operators_test.go:71: seed 25: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestOperatorsShapes2311601471/001/program.a:868:57: Adamic 0.1 refuses task capture 'step57' is not shareable: step57 is not an immutable binding; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)
    operators_test.go:71: seed 27: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestOperatorsShapes2311601471/001/program.a:802:57: Adamic 0.1 refuses task capture 'step24' is not shareable: step24 is not an immutable binding; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)
    operators_test.go:71: seed 32: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestOperatorsShapes2311601471/001/program.a:857:57: Adamic 0.1 refuses task capture 'step24' is not shareable: step24 is not an immutable binding; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)
    operators_test.go:71: seed 33: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestOperatorsShapes2311601471/001/program.a:955:43: Adamic 0.1 refuses parallelMap work writes state another task can reach: record42 writes the global 'total41'; keep work and everything it calls pure; write only task locals and fresh objects
    operators_test.go:71: seed 36: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestOperatorsShapes2311601471/001/program.a:884:43: Adamic 0.1 refuses parallelMap work writes state another task can reach: record34 writes the global 'total33'; keep work and everything it calls pure; write only task locals and fresh objects
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/native/TestRuntimeStaticsSignalAndExit/handler_buffer_mutant 15.13 seconds
```text
    runtime_statics_parallel_test.go:96: signal mutant not caught: signal: terminated
        signal worker line
        signal worker line
        signal worker line
        signal worker line
        signal worker line
        signal worker line
        signal worker line
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/native/TestRuntimeStaticsSignalAndExit 94.02 seconds
```text
=== RUN   TestRuntimeStaticsSignalAndExit
--- FAIL: TestRuntimeStaticsSignalAndExit (94.02s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/fuzz 134.198 seconds
```text
FAIL
FAIL	github.com/system-inc/adamic/internal/fuzz	134.198s
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/refusalprobe/TestCatalogCoverage 0 seconds
```text
    probe_test.go:15: refusal catalog is incomplete: helper parallelPreflight, helper typedArrayUnsupported, helper enumRefusal, helper refuseOptionalWidening
--- FAIL: TestCatalogCoverage (0.00s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/skipcensus/TestProvidedHistoricalPlainLog 0.02 seconds
```text
    census_test.go:338: mismatched tree must fail closed: <nil> measurement	github.com/system-inc/adamic/cmd/adamic-test262	TestCompilerStartupMeasurement	TestCompilerStartupMeasurement:c4a4cc95e992b9399ab2a21ae7e3a1caca9c71a5d9b0b7f5d48200073d19ce03	Opt-in timing or profile artifact comparison; does not replace correctness verification.
        not-applicable	github.com/system-inc/adamic/internal/metamorphic	TestTheLeakCheckReadsTheCounts	TestTheLeakCheckReadsTheCounts:c25bf1830cfbf9413673dfec69544a474d2a30e77d371825c69128882094f2e1	macOS-only counted-build leak witness; Linux uses LeakSanitizer on the sanitized binary.
        measurement	github.com/system-inc/adamic/internal/native	TestNormalizeLongMeasurements	TestNormalizeLongMeasurements:6dcd1081c12d5b6018fcc9cefc4b64e74675be953be75185fa2be02d217f45de	ADAMIC_NORMALIZE_BENCH_LOG names a writable JSONL output for opt-in long-string timing and RSS observations.
        measurement	github.com/system-inc/adamic/internal/native	TestRecordBenchmark	TestRecordBenchmark:06ee23e4bcdbf34555e793a31d7dc4a10bdb21fb5c9a110c068aa7683ad1f84d	Opt-in benchmark: set ADAMIC_RECORD_BENCH=1 for five-round Node timing comparisons; no timing threshold verifies correctness.
        opt-in-lane	github.com/system-inc/adamic/internal/oracle	TestGCCAgreesWithNode	TestGCCAgreesWithNode:8e1ac89c8ea9525bf996608391643f7c58b47070d50d2866cc1b34260c01c49f	Dedicated GCC verification lane opts in with ADAMIC_GCC_LANE=1. ADAMIC_LANE_CC selects the compiler (default gcc); ADAMIC_LANE_SANITIZE=1 enables sanitizers. Missing compiler or failed builds fail the opted-in lane rather than skipping it.
        opt-in-lane	github.com/system-inc/adamic/internal/oracle	TestNativeReleaseFlagsAgreeWithNode	TestNativeReleaseFlagsAgreeWithNode:acee7bcd3dc3537b672326f62074666026f5bf75dfbc5007fde603e48d7d3efd	Dedicated shipping-build verification lane opts in with ADAMIC_RELEASE_LANE=1. The lane traces production release options; missing tools or failed builds fail the opted-in lane rather than skipping it.
        measurement	github.com/system-inc/adamic/internal/oracle	TestNativeExistingReleaseAgreesWithNode	TestNativeExistingReleaseAgreesWithNode:657fcad842d7dc04a11e8f0a17b161826c0f0fd4c245d76eff93418ab613fe0d	Opt-in timing control: ADAMIC_RELEASE_MEASURE=1 compares the existing release variant on finishing fixtures; not a substitute for the shipping-build verification lane.
        measurement	github.com/system-inc/adamic/internal/oracle	TestNativeMallocAgreesWithNode	TestNativeMallocAgreesWithNode:4be973d40ef3554399a7b55adc4838f67420bcdfeaba943cfec966c3f7a02827	ADAMIC_SLAB_MEASURE=1 enables the existing malloc allocator timing control, separate from the slab correctness lane.
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/skipcensus/TestCensus 0.61 seconds
```text
    census_test.go:31: removed skip internal/native/wasm_test.go:TestWASI:84bb45dda60b9ac4a325def472053ec24083a3eca7b2cf2474e71763541bf80c
        removed skip internal/native/wasm_test.go:TestWASI:c21b865ade4f7ec33e3c43e989d4915bb320441b56260e70d1b404d8e6ee400c
        removed skip internal/native/wasm_test.go:TestWASI:e331d554fc45e12edd105c0ed30539b5b8161c822ca013eeb9b559b9a29e4822
        removed skip internal/native/wasm_test.go:TestWASI:f69cd3364c297e515a7579901e72251584219181e3d65f6db4443ff8a7833420
        removed skip internal/oracle/wasi_test.go:TestWASIEmission:01adf590dfb4d6d0d1ab0c887e7724905ac27af2f60ef5b900d00a0f2adc4dc9
        stale metadata stage1/cohere/lint/jsx_integration_test.go:TestJsxLintReleaseAndThroughput:0ac2731cc03aa88aa1db988e0910063d36067cbcb25e2a776be9408d8e4d1ebf
        stale metadata stage1/cohere/lint/lint_test.go:TestCompilerAndStage1Agree:1dbf140729f17ef866b29b574189ddf438e2a4dd6762d9bdcd8e5dae68bf6b7e
        stale metadata stage1/cohere/lint/lint_test.go:TestThroughput:0ac2731cc03aa88aa1db988e0910063d36067cbcb25e2a776be9408d8e4d1ebf
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/skipcensus 0.627 seconds
```text
FAIL
FAIL	github.com/system-inc/adamic/internal/skipcensus	0.627s
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/refusalprobe 14.001 seconds
```text
FAIL
FAIL	github.com/system-inc/adamic/internal/refusalprobe	14.000s
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/native/TestWASI 0.05 seconds
```text
    wasm_test.go:41: WASI toolchain missing or unusable (sysroot, linker, builtins): exit status 1
        wasm-ld: error: cannot open /workspace/adamic-tools/llvm/lib/clang/20/lib/wasm32-unknown-wasi/libclang_rt.builtins.a: No such file or directory
        clang: error: linker command failed with exit code 1 (use -v to see invocation)
--- FAIL: TestWASI (0.05s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/cmd/adamic-gate/TestWASICurrentSourceAudit 0.92 seconds
```text
    wasi_test.go:143: skip census table does not match tree: removed skip internal/native/wasm_test.go:TestWASI:84bb45dda60b9ac4a325def472053ec24083a3eca7b2cf2474e71763541bf80c
        removed skip internal/native/wasm_test.go:TestWASI:c21b865ade4f7ec33e3c43e989d4915bb320441b56260e70d1b404d8e6ee400c
        removed skip internal/native/wasm_test.go:TestWASI:e331d554fc45e12edd105c0ed30539b5b8161c822ca013eeb9b559b9a29e4822
        removed skip internal/native/wasm_test.go:TestWASI:f69cd3364c297e515a7579901e72251584219181e3d65f6db4443ff8a7833420
        removed skip internal/oracle/wasi_test.go:TestWASIEmission:01adf590dfb4d6d0d1ab0c887e7724905ac27af2f60ef5b900d00a0f2adc4dc9
        stale metadata stage1/cohere/lint/jsx_integration_test.go:TestJsxLintReleaseAndThroughput:0ac2731cc03aa88aa1db988e0910063d36067cbcb25e2a776be9408d8e4d1ebf
        stale metadata stage1/cohere/lint/lint_test.go:TestCompilerAndStage1Agree:1dbf140729f17ef866b29b574189ddf438e2a4dd6762d9bdcd8e5dae68bf6b7e
        stale metadata stage1/cohere/lint/lint_test.go:TestThroughput:0ac2731cc03aa88aa1db988e0910063d36067cbcb25e2a776be9408d8e4d1ebf
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/cmd/adamic-gate/TestCurrentCohereGateIsRequired 1.05 seconds
```text
    required_environment_test.go:178: skip census table does not match tree: removed skip internal/native/wasm_test.go:TestWASI:84bb45dda60b9ac4a325def472053ec24083a3eca7b2cf2474e71763541bf80c
        removed skip internal/native/wasm_test.go:TestWASI:c21b865ade4f7ec33e3c43e989d4915bb320441b56260e70d1b404d8e6ee400c
        removed skip internal/native/wasm_test.go:TestWASI:e331d554fc45e12edd105c0ed30539b5b8161c822ca013eeb9b559b9a29e4822
        removed skip internal/native/wasm_test.go:TestWASI:f69cd3364c297e515a7579901e72251584219181e3d65f6db4443ff8a7833420
        removed skip internal/oracle/wasi_test.go:TestWASIEmission:01adf590dfb4d6d0d1ab0c887e7724905ac27af2f60ef5b900d00a0f2adc4dc9
        stale metadata stage1/cohere/lint/jsx_integration_test.go:TestJsxLintReleaseAndThroughput:0ac2731cc03aa88aa1db988e0910063d36067cbcb25e2a776be9408d8e4d1ebf
        stale metadata stage1/cohere/lint/lint_test.go:TestCompilerAndStage1Agree:1dbf140729f17ef866b29b574189ddf438e2a4dd6762d9bdcd8e5dae68bf6b7e
        stale metadata stage1/cohere/lint/lint_test.go:TestThroughput:0ac2731cc03aa88aa1db988e0910063d36067cbcb25e2a776be9408d8e4d1ebf
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/cmd/adamic-gate 138.883 seconds
```text
FAIL
FAIL	github.com/system-inc/adamic/cmd/adamic-gate	138.883s
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/internal/native 1155.921 seconds
```text
FAIL
FAIL	github.com/system-inc/adamic/internal/native	1155.920s
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/stage1/cohere/estree/TestAcceptanceDiagnostics 54.43 seconds
```text
    acceptance_test.go:62: [node --disable-warning=ExperimentalWarning /workspace/adamic/oracle/node.mjs /workspace/adamic/stage1/cohere/estree/main.ts /workspace/adamic-gate-tmp/TestAcceptanceDiagnostics2906676902/001/0003.ts]: timeout=context deadline exceeded exit=signal: killed stdout=0 stderr=adamic: panic: ESTree parser: function or constructor type must be parenthesized in a union or intersection
--- FAIL: TestAcceptanceDiagnostics (54.43s)
```

q4-whole-restart.jsonl: github.com/system-inc/adamic/stage1/cohere/estree 2160.657 seconds
```text
FAIL
FAIL	github.com/system-inc/adamic/stage1/cohere/estree	2160.657s
```

q4-rerun-cmd-adamic-gate.jsonl: github.com/system-inc/adamic/cmd/adamic-gate/TestWASICurrentSourceAudit 0.28 seconds
```text
    wasi_test.go:143: skip census table does not match tree: removed skip internal/native/wasm_test.go:TestWASI:84bb45dda60b9ac4a325def472053ec24083a3eca7b2cf2474e71763541bf80c
        removed skip internal/native/wasm_test.go:TestWASI:c21b865ade4f7ec33e3c43e989d4915bb320441b56260e70d1b404d8e6ee400c
        removed skip internal/native/wasm_test.go:TestWASI:e331d554fc45e12edd105c0ed30539b5b8161c822ca013eeb9b559b9a29e4822
        removed skip internal/native/wasm_test.go:TestWASI:f69cd3364c297e515a7579901e72251584219181e3d65f6db4443ff8a7833420
        removed skip internal/oracle/wasi_test.go:TestWASIEmission:01adf590dfb4d6d0d1ab0c887e7724905ac27af2f60ef5b900d00a0f2adc4dc9
        stale metadata stage1/cohere/lint/jsx_integration_test.go:TestJsxLintReleaseAndThroughput:0ac2731cc03aa88aa1db988e0910063d36067cbcb25e2a776be9408d8e4d1ebf
        stale metadata stage1/cohere/lint/lint_test.go:TestCompilerAndStage1Agree:1dbf140729f17ef866b29b574189ddf438e2a4dd6762d9bdcd8e5dae68bf6b7e
        stale metadata stage1/cohere/lint/lint_test.go:TestThroughput:0ac2731cc03aa88aa1db988e0910063d36067cbcb25e2a776be9408d8e4d1ebf
```

q4-rerun-cmd-adamic-gate.jsonl: github.com/system-inc/adamic/cmd/adamic-gate/TestCurrentCohereGateIsRequired 0.3 seconds
```text
    required_environment_test.go:178: skip census table does not match tree: removed skip internal/native/wasm_test.go:TestWASI:84bb45dda60b9ac4a325def472053ec24083a3eca7b2cf2474e71763541bf80c
        removed skip internal/native/wasm_test.go:TestWASI:c21b865ade4f7ec33e3c43e989d4915bb320441b56260e70d1b404d8e6ee400c
        removed skip internal/native/wasm_test.go:TestWASI:e331d554fc45e12edd105c0ed30539b5b8161c822ca013eeb9b559b9a29e4822
        removed skip internal/native/wasm_test.go:TestWASI:f69cd3364c297e515a7579901e72251584219181e3d65f6db4443ff8a7833420
        removed skip internal/oracle/wasi_test.go:TestWASIEmission:01adf590dfb4d6d0d1ab0c887e7724905ac27af2f60ef5b900d00a0f2adc4dc9
        stale metadata stage1/cohere/lint/jsx_integration_test.go:TestJsxLintReleaseAndThroughput:0ac2731cc03aa88aa1db988e0910063d36067cbcb25e2a776be9408d8e4d1ebf
        stale metadata stage1/cohere/lint/lint_test.go:TestCompilerAndStage1Agree:1dbf140729f17ef866b29b574189ddf438e2a4dd6762d9bdcd8e5dae68bf6b7e
        stale metadata stage1/cohere/lint/lint_test.go:TestThroughput:0ac2731cc03aa88aa1db988e0910063d36067cbcb25e2a776be9408d8e4d1ebf
```

q4-rerun-cmd-adamic-gate.jsonl: github.com/system-inc/adamic/cmd/adamic-gate 45.756 seconds
```text
FAIL
FAIL	github.com/system-inc/adamic/cmd/adamic-gate	45.755s
```

q4-rerun-internal-fuzz.jsonl: github.com/system-inc/adamic/internal/fuzz/TestLivenessAndFieldShapes 15.13 seconds
```text
    liveness_fields_test.go:56: seed 2: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestLivenessAndFieldShapes3204622551/001/program.a:9:3: Adamic 0.1 refuses cannot move items1[0]: element is not a fresh object literal with only scalar literal fields; return it through the results
    liveness_fields_test.go:56: seed 5: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestLivenessAndFieldShapes3204622551/001/program.a:7:3: Adamic 0.1 refuses cannot move items1[0][]: nested array ownership is not proven; return it through the results
    liveness_fields_test.go:56: seed 10: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestLivenessAndFieldShapes3204622551/001/program.a:7:3: Adamic 0.1 refuses cannot move items1[0][]: nested array ownership is not proven; return it through the results
    liveness_fields_test.go:56: seed 12: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestLivenessAndFieldShapes3204622551/001/program.a:8:3: Adamic 0.1 refuses cannot move items1[0]: element is not a fresh object literal with only scalar literal fields; return it through the results
    liveness_fields_test.go:56: seed 15: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestLivenessAndFieldShapes3204622551/001/program.a:8:3: Adamic 0.1 refuses cannot move items1[0]: element is not a fresh object literal with only scalar literal fields; return it through the results
    liveness_fields_test.go:56: seed 16: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestLivenessAndFieldShapes3204622551/001/program.a:698:31: Adamic 0.1 refuses parallelMap items are not shareable: Bag39.tags is a mutable string[]; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)
    liveness_fields_test.go:56: seed 17: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestLivenessAndFieldShapes3204622551/001/program.a:8:3: Adamic 0.1 refuses cannot move items1[0]: element is not a fresh object literal with only scalar literal fields; return it through the results
    liveness_fields_test.go:56: seed 19: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestLivenessAndFieldShapes3204622551/001/program.a:9:3: Adamic 0.1 refuses cannot move items1[0]: element is not a fresh object literal with only scalar literal fields; return it through the results
```

q4-rerun-internal-fuzz.jsonl: github.com/system-inc/adamic/internal/fuzz/TestOperatorsShapes 15.45 seconds
```text
    operators_test.go:71: seed 10: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestOperatorsShapes1759839319/001/program.a:942:31: Adamic 0.1 refuses parallelMap items are not shareable: Bag36.tags is a mutable string[]; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)
    operators_test.go:71: seed 11: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestOperatorsShapes1759839319/001/program.a:860:43: Adamic 0.1 refuses parallelMap work writes state another task can reach: record25 writes the global 'total24'; keep work and everything it calls pure; write only task locals and fresh objects
    operators_test.go:71: seed 16: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestOperatorsShapes1759839319/001/program.a:826:31: Adamic 0.1 refuses parallelMap items are not shareable: Bag39.tags is a mutable string[]; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)
    operators_test.go:71: seed 25: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestOperatorsShapes1759839319/001/program.a:868:57: Adamic 0.1 refuses task capture 'step57' is not shareable: step57 is not an immutable binding; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)
    operators_test.go:71: seed 27: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestOperatorsShapes1759839319/001/program.a:802:57: Adamic 0.1 refuses task capture 'step24' is not shareable: step24 is not an immutable binding; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)
    operators_test.go:71: seed 32: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestOperatorsShapes1759839319/001/program.a:857:57: Adamic 0.1 refuses task capture 'step24' is not shareable: step24 is not an immutable binding; make the whole reachable value readonly and bind captures with const; moving mutable values into tasks belongs to concurrency part 2 (#p286ycm)
    operators_test.go:71: seed 33: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestOperatorsShapes1759839319/001/program.a:955:43: Adamic 0.1 refuses parallelMap work writes state another task can reach: record42 writes the global 'total41'; keep work and everything it calls pure; write only task locals and fresh objects
    operators_test.go:71: seed 36: stage 0 didn't lower it: /workspace/adamic-gate-tmp/TestOperatorsShapes1759839319/001/program.a:884:43: Adamic 0.1 refuses parallelMap work writes state another task can reach: record34 writes the global 'total33'; keep work and everything it calls pure; write only task locals and fresh objects
```

q4-rerun-internal-fuzz.jsonl: github.com/system-inc/adamic/internal/fuzz 47.301 seconds
```text
FAIL
FAIL	github.com/system-inc/adamic/internal/fuzz	47.300s
```

Order probe first start events matched ir, cmd/adamic, flow: True
go list ./... exact package set: 74 packages.
Order used:
```text
./stage1/cohere/markdownblocks
./stage1/cohere/lint
./stage1/cohere/css
./stage1/cohere/typeaware
./internal/unicodeproperties
./internal/oracle
./stage1/cohere/json
./stage1/cohere/markdowninline
./stage1/typescript/parser
./internal/native
./stage1/cohere/cssnumbers
./stage1/cohere/cssstrings
./stage1/cohere/lint/helpers/comments
./internal/flow
./bridge/tsgo
./stage1/cohere/lint/helpers
./stage1/typescript/scanner
./cmd/adamic-test262
./internal/fresh
./stage1/cohere/graphql
./internal/lower
./stage1/cohere/formatfiles
./stage1/cohere/gitignore
./stage1/cohere/lint/inventory
./stage1/cohere/selector
./stage1/cohere/values
./stage1/cohere/suppression
./internal/fuzz
./cloud
./stage3/fixtures
./stage1/cohere/mediaquery
./internal/regexp
./cmd/adamic
./cmd/adamic-meter
./internal/load
./internal/ir
./cmd/adamic-gate
./cmd/adamic-stage1-progress
./stage1/cohere/lint/registry
./bridge/tsgo/checker
./bench
./bench/regex
./bridge/tsgo/archive
./bridge/tsgo/cost
./bridge/tsgo/oracle
./bridge/tsgo/spec
./cmd/adamic-fuzz
./cmd/adamic-reduce
./cmd/lint-registry
./internal/javascript
./stage1/cohere/markdownblocks/tools/generate_classes
./stage1/cohere/markdownblocks/tools/generate_entities
./stage1/cohere/markdownblocks/tools/generate_width
./stage3/census/latent/tool
./stage3/census/tool
./stage3/drivers/parser/probe
./cloud/reports/decode-ascii
./cloud/reports/release-lto
./cmd/adamic-lint-check
./cmd/adamic-metamorphic
./cmd/adamic-refusals
./internal/boundedrun
./internal/boundedrun/testfixture
./internal/leakcheck
./internal/metamorphic
./internal/nodepin
./internal/refusalprobe
./internal/skipcensus
./internal/skipcensus/cmd
./stage1/cohere/estree
./stage1/cohere/graphql/printer
./stage1/cohere/tsprinter
./stage1/cohere/yaml
./verify/coverage/runtimelibrary
```
Added packages: ["./cloud/reports/decode-ascii", "./cloud/reports/release-lto", "./cmd/adamic-lint-check", "./cmd/adamic-metamorphic", "./cmd/adamic-refusals", "./internal/boundedrun", "./internal/boundedrun/testfixture", "./internal/leakcheck", "./internal/metamorphic", "./internal/nodepin", "./internal/refusalprobe", "./internal/skipcensus", "./internal/skipcensus/cmd", "./stage1/cohere/estree", "./stage1/cohere/graphql/printer", "./stage1/cohere/tsprinter", "./stage1/cohere/yaml", "./verify/coverage/runtimelibrary"]
Dropped packages: []

Official setup exports, sourced from /workspace/adamic-tools/env.sh (all 12 required values confirmed before tests):
```json
{
  "PATH": "/workspace/adamic-tools/bin:/workspace/adamic-tools/go/bin:/home/agent/.local/bin:/opt/codex/runtimes/codex-primary-runtime/dependencies/bin/override:/opt/codex/runtimes/codex-primary-runtime/dependencies/python/bin:/opt/codex/runtimes/codex-primary-runtime/dependencies/node/bin:/opt/codex/runtimes/codex-primary-runtime/dependencies/bin:/opt/codex/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/home/agent/.local/bin:/opt/codex/runtimes/codex-primary-runtime/dependencies/bin/fallback",
  "GOTOOLCHAIN": "auto",
  "GOPROXY": "https://proxy.golang.org|direct",
  "TMPDIR": "/tmp/adamic-gate",
  "ADAMIC_MARKDOWNWIDTH_DEPS": "/workspace/adamic-tools/markdown-width",
  "WASI_SYSROOT": "/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot",
  "ADAMIC_TYPESCRIPT_SOURCE": "/workspace/adamic-tools/gate-inputs/typescript",
  "ADAMIC_CSS_FIXTURES": "/workspace/adamic-tools/gate-inputs/css-fixtures",
  "ADAMIC_CSSNUMBERS_LIBRARY": "/workspace/adamic-tools/gate-inputs/css-printer",
  "ADAMIC_CSSSTRINGS_LIBRARY": "/workspace/adamic-tools/gate-inputs/css-printer",
  "ADAMIC_MARKDOWNINLINE_LIBRARY": "/workspace/adamic-tools/gate-inputs/css-printer/node_modules/prettier",
  "ADAMIC_CSS_LIBRARY": "/workspace/adamic-tools/gate-inputs/css",
  "ADAMIC_GRAPHQL_LIBRARY": "/workspace/adamic-tools/gate-inputs/graphql",
  "ADAMIC_MEDIA_QUERY_LIBRARY": "/workspace/adamic-tools/gate-inputs/media-query",
  "ADAMIC_SELECTOR_LIBRARY": "/workspace/adamic-tools/gate-inputs/selector",
  "ADAMIC_VALUES_LIBRARY": "/workspace/adamic-tools/gate-inputs/values",
  "ADAMIC_GRAPHQL_PRETTIER": "/workspace/adamic-tools/gate-inputs/css-printer",
  "ADAMIC_JSON_PRETTIER": "/workspace/adamic-tools/gate-inputs/json-prettier",
  "ADAMIC_CSS_PRINTER_LIBRARY": "/workspace/adamic-tools/gate-inputs/css-printer",
  "ADAMIC_ESTREE_LIBRARY": "/workspace/adamic-tools/gate-inputs/css-printer",
  "ADAMIC_TS_PRETTIER": "/workspace/adamic-tools/gate-inputs/css-printer",
  "ADAMIC_YAML_LIBRARY": "/workspace/adamic-tools/gate-inputs/css-printer",
  "ADAMIC_GITIGNORE_LARGEST": "/workspace/adamic-tools/gate-inputs/gitignore/.gitignore",
  "ADAMIC_CLANG_TSGO_ARCHIVE": "/workspace/adamic-tools/gate-inputs/checker/tsgo.a"
}
```
Additional inputs: ADAMIC_GATE_COHERE=1; ADAMIC_CYCLE_LEDGER_ROOT=$ADAMIC_TYPESCRIPT_SOURCE; ADAMIC_CYCLE_LEDGER_OUTPUT=/workspace/q4-cycle-ledger.json. Generated ignored compiler diagnostic map caused the typeaware 78-file assertion; git status --porcelain -- src/compiler was empty, git check-ignore identified .gitignore:23.

Setup timing lines and nproc:
```text
setup: go ready (0.085s)
setup: submodules ready (0.110s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.319s)
setup: node v24.19.0 installed (published SHA-256 verified); step-duration=2.376s
setup: node ready (2.492s)
setup: markdown dependencies installed (npm ci, integrity verified); step-duration=0.724s
setup: markdown dependencies ready (3.261s)
setup: stage3 API dependencies installed (npm ci --prefix stage3/api, integrity verified); step-duration=0.961s
setup: stage3 API dependencies checked (4.271s)
setup: gate input css-printer installed (npm ci, integrity verified); step-duration=1.895s
setup: gate input css installed (npm ci, integrity verified); step-duration=0.896s
setup: gate input graphql installed (npm ci, integrity verified); step-duration=1.698s
setup: gate input media-query installed (npm ci, integrity verified); step-duration=0.861s
setup: gate input selector installed (npm ci, integrity verified); step-duration=1.046s
setup: gate input values installed (npm ci, integrity verified); step-duration=0.884s
setup: gate input json-prettier installed (npm ci, integrity verified); step-duration=0.883s
setup: gate input shared Prettier paths shared prettier 3.9.6: CSS numbers, CSS strings and Markdown inline paths verified; step-duration=0.088s
setup: gate npm inputs ready (12.579s)
setup: gate input CSS fixtures installed (exact commit, sparse checkout, Git integrity and counts verified: {'.css': 157, '.scss': 90, '.less': 43}); step-duration=1.317s
setup: gate input TypeScript source installed (depth-one exact commit, git object integrity checked); step-duration=16.568s
setup: gate input 100 MiB gitignore generated (deterministic 100 MiB, SHA256 verified); step-duration=0.215s
setup: gate corpora ready (18.216s)
setup: module dependencies downloaded and verified (22 module directories); step-duration=31.610s
setup: module dependencies ready (31.775s)
setup: wasi sdk ready (/workspace/adamic-tools/wasi-sdk) (34.691s)
setup: gate input checker archive built (Go content-addressed inputs, archive bytes stamped); step-duration=189.841s
setup: go build ready (383.321s)
setup: test binaries deferred (use --warm-tests) (383.441s)
setup: build cache warm (383.446s)
setup: workspace sums restored to the commit (383.535s)
setup: build-flags commit=dc135a74f70de5016f3025d876bc96920dc3fb0a nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gate-inputs=true gate-archive=true load-before=0.24 0.05 0.02 1/151 2240 load-after=9.23 5.08 2.25 1/159 8066
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (383.593s)
setup: node v24.19.0
```

Stage 3 apply exit: 0; apply wall seconds: 199.573. Stage 3 oracle expected lane:
```json
{
  "tree": "/workspace/stage3-applied",
  "nproc": 5,
  "workers": 4,
  "runners": "all",
  "tests": null,
  "node": "v24.19.0",
  "phases": {
    "install": {
      "command": [
        "npm",
        "ci",
        "--no-audit",
        "--no-fund"
      ],
      "exit": 0,
      "seconds": 4.828
    },
    "build": {
      "command": [
        "npm",
        "run",
        "build"
      ],
      "exit": 0,
      "seconds": 4.428
    },
    "tests": {
      "command": [
        "npm",
        "test",
        "--",
        "--workers=4",
        "--lint=false",
        "--no-colors"
      ],
      "exit": 1,
      "seconds": 475.69
    }
  },
  "status": "fail",
  "counts": {
    "passing": 106366,
    "failing": 1,
    "pending": 0
  },
  "baseline_diffs": [
    "api/typescript.d.ts"
  ],
  "wall_seconds": 485.96
}
```
Only failing name: unittests:: Public APIs for typescript.d.ts should be acknowledged when they change

Census command: go run ./internal/skipcensus/cmd /workspace/q4-whole-restart.jsonl
Exit: pending
Output verbatim:
```text
pending
```

Gate-logs branch: gate-logs/dc135a74f70d/20261007T220149/whole
Interrupted-attempt gate-logs branch: gate-logs/dc135a74f70d/20261007T215717/whole
Publisher worktree commit: 37163b619ce52650da319f06554721c18b1a1d1b. Default signals verified as SigIgn=0000000000000000 (none).

Runner limitations: the first partial whole run used ./... because the runner misread workspace module names; it was stopped and replaced with the exact longest-first order on the same SHA. The Stage 3 supervisor was interrupted during that restart, and the complete oracle was rerun successfully to the expected Linux report. stage3/apply.sh changed patch-set.md; its generated copy was preserved outside the repository and the file restored to the pushed commit before later gates. Broad deadline matching also reran cmd/adamic-gate and internal/fuzz due passing deadline-mutant output; original assertion failures remain failures. All conflict Go files were read whole; a full file-by-file incoming formatter audit was not completed.

