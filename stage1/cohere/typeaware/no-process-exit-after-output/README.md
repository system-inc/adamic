These directories implement `nexus/correctness-no-process-exit-after-output`,
`nexus/correctness-no-uncleared-race-timeout`, and
`nexus/correctness-require-blocking-standard-streams` as native Adamic `.a` modules.
`rule.a` in each directory exports its rule; `main.a` runs it independently.
`suite.a` runs all three. Existing shared harness and registration generator files
are untouched. Integration into that harness remains the shared worker's work.

The output rule constructs its own control-flow graph and walks output chains
through loops, branches, catches, and finally blocks. It follows one level of
resolved callees, including module helpers, and observes the pinned checker's
`never` return bit. Blocking-stream analysis shares the graph and walks reachable
function bodies, callback arguments, and native import closure calculations.
Timeout analysis checks the real Promise and timer declarations, follows const
initializers, and checks whether a timeout handle is retained or read.

Three raw checker questions are defined in isolated files on both sides:
`declaration-chain`, `resolved-callee`, and `module-records`. The Go side returns
symbol ancestry, resolved signatures, and compiler file/import records. It does
not call the production lint rules or return their verdicts. The independent Go
oracle calls the unmodified production cohere rules and imports no bridge code.

The production rules expose no fixes or suggestions; the comparison includes the
complete finding fields and their empty fixes and suggestions, not just counts.
The committed compressed fixture tape contains all 121 input projects from their
upstream Go tests, including the 14 helper-call projects with a custom setup hook.
The original upstream Go tests also ran with their original expected findings.

Reproduce with the configured toolchain:

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/no-process-exit-after-output/unpack.py /tmp/wave20-cases
python3 stage1/cohere/typeaware/no-process-exit-after-output/verify.py \
  --repository "$PWD" --artifacts /tmp/wave20-checks \
  --compiler /path/to/TypeScript --cases /tmp/wave20-cases \
  > /tmp/wave20-checks.log 2>&1
```

Use TypeScript v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`, for the frozen compiler manifest.
The gate requires at least 121 projects, positive controls for every rule,
normal and ASan/UBSan/LSan byte equality, three rule and three answer mutants,
the existing frozen repository and compiler corpora, released-handle rejection,
a registry-retention mutant, and three alternating Go/native timing rounds.
