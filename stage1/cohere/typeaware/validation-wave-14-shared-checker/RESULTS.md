# Completion results

{
  "top_level": {
    "pass": 25,
    "fail": 9,
    "skip": 1
  },
  "subtests": {
    "pass": 84,
    "fail": 2,
    "skip": 0
  },
  "failed": [
    "TestJsxLintReleaseAndThroughput",
    "TestJsxLintTrees",
    "TestRulesAgree",
    "TestCompilerAndStage1Agree",
    "TestNodeTableIsLinkOnly",
    "TestShardsAgree",
    "TestMutants/wave14-no_extraneous_class",
    "TestMutants/wave14-no_mock_on_module_namespace",
    "TestMutants",
    "TestProfileSnapshotsAgree",
    "TestOwnedWitnesses"
  ],
  "skipped": [
    "TestCheckerBridgeRefusalPending"
  ],
  "execution": {
    "start_load": [
      0.8134765625,
      1.59228515625,
      2.40576171875
    ],
    "nproc": 5,
    "exit": 1,
    "wall_seconds": 2340.142571421,
    "end_load": [
      2.3759765625,
      4.609375,
      5.92529296875
    ]
  },
  "sampled_load1": {
    "min": 0.8134765625,
    "median": 6.40283203125,
    "max": 11.1865234375
  }
}

Commands: source /workspace/adamic-tools/env.sh; GOMAXPROCS=4 GOFLAGS=-buildvcs=false GOPROXY="https://proxy.golang.org|direct" ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-14-typescript ADAMIC_LINT_BENCH=1 ADAMIC_LINT_PROFILE_DIR=<fresh directory> ADAMIC_LINT_PROFILE_SNAPSHOTS=<same directory> go test -json -count=1 -timeout=60m ./stage1/cohere/lint. Tests ran concurrently on nproc=5 (CPU quota=4); sampled load and real wall time are above. Compiler source pin 050880ce59e30b356b686bd3144efe24f875ebc8; cohere pin 7945d102a6c18dd36adf9114a758ce646e8b2359.

Outcomes below are raw full-gate outcomes; targeted repaired-mutant results are retained separately. No failed shared checks or pending skip were relaxed.
