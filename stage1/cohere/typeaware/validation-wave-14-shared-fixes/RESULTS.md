# Completion results

{
  "top_level": {
    "pass": 33,
    "fail": 2,
    "skip": 1
  },
  "subtests": {
    "pass": 91,
    "fail": 0,
    "skip": 0
  },
  "failed": [
    "TestCompilerAndStage1Agree",
    "TestProfileSnapshotsAgree"
  ],
  "skipped": [
    "TestCheckerBridgeRefusalPending"
  ],
  "execution": {
    "start_load": [
      0.0,
      0.0537109375,
      1.451171875
    ],
    "nproc": 5,
    "exit": 1,
    "wall_seconds": 1965.0701829120007,
    "end_load": [
      1.51806640625,
      2.0224609375,
      2.42138671875
    ]
  },
  "sampled_load1": {
    "min": 0.0,
    "median": 2.19580078125,
    "max": 5.02490234375
  }
}

Commands: source /workspace/adamic-tools/env.sh; GOMAXPROCS=4 GOFLAGS=-buildvcs=false GOPROXY="https://proxy.golang.org|direct" ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-14-typescript ADAMIC_LINT_BENCH=1 ADAMIC_LINT_PROFILE_DIR=<fresh directory> ADAMIC_LINT_PROFILE_SNAPSHOTS=<same directory> go test -json -count=1 -timeout=60m ./stage1/cohere/lint. Tests ran concurrently on nproc=5 (CPU quota=4); sampled load and real wall time are above. Compiler source pin 050880ce59e30b356b686bd3144efe24f875ebc8; cohere pin 7945d102a6c18dd36adf9114a758ce646e8b2359.

Outcomes below are raw full-gate outcomes; targeted repaired-mutant results are retained separately. No failed shared checks or pending skip were relaxed.
