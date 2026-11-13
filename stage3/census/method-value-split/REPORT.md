# Method-value split: blocked before measurement (second attempt)

Counts per class and top 20 methods: **not measured**. `sites.json` is an empty array because no scan ran; it does not indicate zero sites.

## SourceFileLike.getPositionOfLineAndCharacter

Readers in `src/` and presence in `tests/baselines/reference/api/typescript.d.ts`: **not measured**. The adapted tree was not produced.

## Build failure and stopping point

Base: fetched `origin/main` at `1046dcfdbbdca449996d5045dbff7e052e680a1c`. Work was performed in the fresh detached scratch worktree `/workspace/scratch/method-value-split`, subsequently branched for reports. No shared-checkout source edits were made.

Cloned the public `https://github.com/microsoft/TypeScript.git` into `/workspace/scratch/method-value-typescript` and checked out the requested commit detached. Both `git rev-parse HEAD` and `git rev-parse v6.0.3` returned:

```text
050880ce59e30b356b686bd3144efe24f875ebc8
```

Reran from the Adamic scratch worktree, following `stage3/README.md`:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/method-value-typescript stage3/apply.sh /workspace/scratch/method-value-adapted-2 > /workspace/scratch/method-value-apply-2.log 2>&1
```

Exit status: **1**. The automatic local-build fallback passed source verification, then failed in `adaptation_inputs()` → `pinned_packages(stage / 'api')` → `installed_packages(project)`. Exact final error:

```text
RuntimeError: /workspace/scratch/method-value-split/stage3/api/node_modules is not an install of /workspace/scratch/method-value-split/stage3/api/package-lock.json ([Errno 2] No such file or directory: '/workspace/scratch/method-value-split/stage3/api/node_modules/.package-lock.json'); stage3 never runs npm, so the tree must provide them
```

This requests the Adamic worktree's `stage3/api/node_modules`, not the upstream TypeScript checkout's `node_modules`. No `npm ci` was run. Stopped here as explicitly requested if apply fails again. No scan or stage 0 census was run; scanner.ts:469:12 was not independently cross-checked. No adaptation or compiler changes were made.

## What the scan cannot see

No scan results exist in this report. The requested compiler-API method-read scan would be a **static approximation of stage 0’s unbound-method rule**, not a reproduction of its refusals. No conclusions about the presence/optional-call/escape split can be drawn from this blocked run.
