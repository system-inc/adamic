# Replay bootstrap certificate

These are fresh local native/Node certificates for the shared framing and unit 2
closure. receipt.json records hashes of every Go checkpoint in the full census;
the Go producer is rerun by TestCheckpointReplayOracle, rather than committing
large generated source/dump trees. It compares Go with both runtimes after fresh
arena decode, dump and re-encode, and corrupts an instruction anchor.

Reproduce from the repository root after cloud setup:

```bash
export GOPROXY='https://proxy.golang.org|direct'
HIR_REPLAY_CENSUS_EXPORT=/tmp/hir-replay-census go test ./stage1/cohere/high_level_intermediate_representation -run '^TestCheckpointReplayOracle$' -count=1 -v
```

The export directory contains checkpoint-manifest.tsv and each Go-produced
checkpoint, source, hir-v1 dump and checker snapshot. Overlay adapter files are
compiled with lintoracle beside the unchanged Go package.

construction-clone-gaps.txt includes full construction/clone certificates and
focused mutants. The prior complete 79-mutant matrix is retained in
../unit2-step28/; it is not claimed as freshly rerun here. static-components.txt
certifies upstream and owned cases and its creation mutant. Counts use the same
12-function corpus as earlier measurements; all 4,164 allocations are freed.
