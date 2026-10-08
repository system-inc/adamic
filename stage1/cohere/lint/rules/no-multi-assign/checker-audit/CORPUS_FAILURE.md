# Required corpus check failed

`TestCompilerAndStage1Agree` failed after 885.25 seconds on the area-based branch. Its sanitized native scanner was killed by the existing ten-minute command deadline in `stage1/cohere/lint/lint_test.go:75`. The cgroup memory counters reported zero OOM events and zero OOM kills. No finding mismatch was observed: execution did not finish.

The check covered 872 files. `evidence/corpus-manifest.txt` preserves their order and `evidence/corpus-input-hashes.json` preserves their bytes. The branch only adds Markdown notes, raw `.ts.txt`/`.tsx.txt` examples and evidence; none is selected by this corpus's `.ts`/`.a` filter. No implementation was changed.

Reproduce from this branch with the pinned compiler checkout:

```bash
source /workspace/adamic-tools/env.sh
export GOMAXPROCS=4 GOFLAGS=-buildvcs=false
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-12/typescript
go test -json -count=1 -timeout=60m -run '^TestCompilerAndStage1Agree$' ./stage1/cohere/lint > /tmp/wave12-corpus-reproducer.jsonl 2>&1
```

TypeScript input pin: `050880ce59e30b356b686bd3144efe24f875ebc8`. Area input: `c4bdc23fa86d55cf7e579989201c11258f4d3a62`. Source paths in the manifest use this workspace's roots; adjust those roots when reproducing elsewhere.

The harness deadline, corpus selection and assertions were left unchanged. This is a failed required check, not a passing comparison and not a missing-input skip. The root cause of the native runtime cost has not been isolated.
