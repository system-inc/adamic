# L2 stopping proof

| Proof item | Result |
| --- | --- |
| Base | origin/main f0271c6a7b2e92900eb7046d8246cfb89f52cfb6 |
| Upstream | TypeScript 6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8 |
| Setup and pinned parser install | Exit 0; retained logs |
| Full baseline apply, --build-product | Exit 0 |
| Full amended apply | Not run; no implementation constructed |
| Ordinary adamic types, all 79 baseline roots | Exit 0, zero diagnostics |
| Ordinary adamic types, all 79 proposal roots | Nonzero, seven diagnostics |
| Source-scope check | All 746 sources compared; only the exact two core.ts initializer substitutions |
| Adapter 30 on both fully applied trees | Identical preexisting addRange refusal (exit 1); zero source changes |
| Adapter 30 at its post-apply checkpoint, with/without initializers | Both exit 0, identical stdout/stderr, zero source changes |
| Paired replay of every adapter | Not run after checker blocker |
| Node witnesses | Ten of ten return values equal, including all requested edge classes |
| Whole-lane differential real-call count | Not started; no count or 0-difference claim |
| Isolated stock checker | Baseline zero diagnostics, proposal seven |
| Isolated six-assertion scope extension | Zero diagnostics; measurement only, not implemented |
| Ordinary builds and emitted JavaScript proof | Not run after checker blocker |
| Public declaration byte comparison | Not run after checker blocker |
| Full core.ts census before/after | Not run after checker blocker |
| Original-owner mutant | Not run; no accepted candidate exists |
| LEDGER.md | Unchanged; no JavaScript-changing implementation submitted |

The first confirmation runner stopped at 30's baseline refusal because it
incorrectly required successful full-tree replay. replay30.py records the
completed paired comparison and the successful post-30 checkpoint pair.

The scratch-only census overlay build was stopped after finding the checker
blocker, before a census binary or result was used. Measurement preparation was
stopped after the completed baseline apply. No lane test was started. No task
process remains running at closeout. All retained Go source is named .go.txt.

Reproduction, after source /workspace/adamic-tools/env.sh and bash cloud/setup.sh:

```sh
npm ci --prefix stage3/api
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
stage3/apply.sh --build-product /workspace/scratch/levenshtein/before
# Copy all src, replace only the two initializers, and run both sides:
# adamic types receives all 79 sorted src/compiler/**/*.ts roots together.
# confirm.py.txt in the archive records exact commands and the actual run cleanup;
# it is a read-only record containing run-specific PIDs, not a reproduction command.
node stage3/adapt/99-correctness-core-levenshtein/type-probe.cjs
node stage3/adapt/99-correctness-core-levenshtein/assertion-probe.cjs
node stage3/adapt/99-correctness-core-levenshtein/witness.cjs
```

The scripts retain the actual scratch paths. Extract evidence.tar.gz under
/workspace/scratch/levenshtein when reproducing the standalone owner probes.
Full logs, source manifests, both complete core.ts files, owner sources,
checker options, ordinary binary hash, and both adapter-30 outcomes are packed
with internal SHA256SUMS and an adjacent archive SHA-256 file.
