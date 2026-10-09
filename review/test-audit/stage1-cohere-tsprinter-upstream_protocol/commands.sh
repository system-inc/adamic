#!/usr/bin/env bash
# Run from repository root. Dependency paths refer to this session's workspace.
source /workspace/adamic-tools/env.sh
export TMPDIR=/workspace/u142-tmp
export ADAMIC_TS_PRETTIER=/workspace/u097-prettier
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/u111-typescript
# No mutation replay commands: these rows never reach permitted Adamic code.
for name in TestTSCCorpusUpstreamDifferences TestStatementUpstreamDifferences; do
  for sample in 1 2 3; do
    timeout 120 go test -count=1 -timeout 90s ./stage1/cohere/tsprinter/ -run "^${name}$" > "review/test-audit/stage1-cohere-tsprinter-upstream_protocol/replay-${name}-${sample}.log" 2>&1
  done
done
