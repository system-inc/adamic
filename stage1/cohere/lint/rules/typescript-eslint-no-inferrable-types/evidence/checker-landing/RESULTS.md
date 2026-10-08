# Completion results

{
  "top_level": {
    "pass": 34,
    "fail": 0,
    "skip": 1
  },
  "subtests": {
    "pass": 80,
    "fail": 0,
    "skip": 0
  },
  "failed": [],
  "skipped": [
    "TestCheckerBridgeRefusalPending"
  ],
  "execution": {
    "start_load": [
      0.10986328125,
      1.61767578125,
      2.4541015625
    ],
    "nproc": 5,
    "exit": 0,
    "wall_seconds": 2430.1381732689997,
    "end_load": [
      2.16162109375,
      4.19384765625,
      5.71240234375
    ]
  },
  "sampled_load1": {
    "min": 0.10986328125,
    "median": 6.292236328125,
    "max": 11.1865234375
  }
}

Commands: source /workspace/adamic-tools/env.sh; GOMAXPROCS=4 GOFLAGS=-buildvcs=false GOPROXY="https://proxy.golang.org|direct" ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-14-typescript ADAMIC_LINT_BENCH=1 ADAMIC_LINT_PROFILE_DIR=<fresh directory> ADAMIC_LINT_PROFILE_SNAPSHOTS=<same directory> go test -json -count=1 -timeout=60m ./stage1/cohere/lint. Tests ran concurrently on nproc=5 (CPU quota=4); sampled load and real wall time are above. Compiler source pin 050880ce59e30b356b686bd3144efe24f875ebc8; cohere pin 7945d102a6c18dd36adf9114a758ce646e8b2359.

Outcomes below are raw full-gate outcomes; targeted repaired-mutant results are retained separately. No failed shared checks or pending skip were relaxed.
