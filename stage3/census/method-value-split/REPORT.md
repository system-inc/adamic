# Method-value split: blocked before measurement

Counts per class and top 20 methods: **not measured**. `sites.json` is an empty array because no scan ran; it does not indicate zero sites.

## SourceFileLike.getPositionOfLineAndCharacter

Readers in `src/` and presence in `tests/baselines/reference/api/typescript.d.ts`: **not measured**. The adapted tree was not produced.

## Build failure and stopping point

Fetched `origin/main` and confirmed commit `1046dcfdbbdca449996d5045dbff7e052e680a1c`. Used a fresh detached worktree at `/workspace/scratch/method-value-split`; the shared checkout was not edited.

Following `stage3/README.md`, ran from that worktree:

```sh
stage3/apply.sh /workspace/scratch/method-value-adapted > /workspace/scratch/method-value-apply.log 2>&1
```

Exit status: **1**. The cached product was absent:

```text
ProductMissing: product 45628f1f2d52124cc2cb80f3dd695a05e4797d95b1a2ce39e7a534c83e6bfda2 missing from /home/agent/.cache/adamic-stage3/products
```

The automatic local-build fallback then failed with this exact final error:

```text
RuntimeError: set ADAMIC_TYPESCRIPT_SOURCE to the gate inputs' TypeScript v6.0.3 checkout at 050880ce59e30b356b686bd3144efe24f875ebc8
```

`ADAMIC_TYPESCRIPT_SOURCE` was unset. Stopped here as requested when the adapted tree does not build. No adaptation or compiler changes were made. No scan or stage 0 census was run, so scanner.ts:469:12 was not independently cross-checked.

## What the scan cannot see

No scan results exist in this report. The requested compiler-API method-read scan would be a **static approximation of stage 0’s unbound-method rule**, not a reproduction of its refusals. No conclusions about the presence/optional-call/escape split can be drawn from this blocked run.
