# Reproduction commands

Run from the appropriate pinned worktree. Every long command is supervised by run.py or its detached launch.py wrapper; these record the child's own exit code.

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/overload-typescript
bash cloud/setup.sh > setup.log.txt 2>&1
npm ci --prefix stage3/api > npm.log.txt 2>&1
bash stage3/apply.sh /workspace/overload-adapted > apply.log.txt 2>&1
```

Only the parent applies. Both compilers use this completed tree. For each compiler, use its own worktree and overlay paths:

```sh
python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/overload-SIDE-overlay > overlay.log.txt 2>&1
python3 stage3/meter/entry_overlay.py /workspace/overload-SIDE-overlay /workspace/overload-SIDE-entry > entry.log.txt 2>&1
go build -buildvcs=false -overlay=/workspace/overload-SIDE-entry/overlay.json -o /workspace/overload-SIDE-census ./stage3/census/latent/tool > build.log.txt 2>&1
```

The exact four measurement commands follow. They ran through detached supervisors, with LATENT_ASSERT_NO_OUTPUT=1 and a 30-minute inactivity cutoff.

```sh
env -u LATENT_SPECULATIVE LATENT_FULL=1 LATENT_ONLY_FILE=/workspace/overload-adapted/src/compiler/debug.ts LATENT_ASSERT_NO_OUTPUT=1 GOMEMLIMIT=5GiB LATENT_TIMES=/workspace/overload-census/stage3/census/overload-callback-8fabe4c2/debug-parent-times.jsonl LATENT_PROGRESS_FILE=/workspace/overload-census/stage3/census/overload-callback-8fabe4c2/debug-parent-progress.json /workspace/overload-parent-census /workspace/overload-adapted/src/tsc/tsc.ts /workspace/overload-census/stage3/census/overload-callback-8fabe4c2/debug-parent.jsonl > debug-parent.log.txt 2>&1
env -u LATENT_SPECULATIVE LATENT_FULL=1 LATENT_ONLY_FILE=/workspace/overload-adapted/src/compiler/debug.ts LATENT_ASSERT_NO_OUTPUT=1 GOMEMLIMIT=5GiB LATENT_TIMES=/workspace/overload-census/stage3/census/overload-callback-8fabe4c2/debug-head-times.jsonl LATENT_PROGRESS_FILE=/workspace/overload-census/stage3/census/overload-callback-8fabe4c2/debug-head-progress.json /workspace/overload-head-census /workspace/overload-adapted/src/tsc/tsc.ts /workspace/overload-census/stage3/census/overload-callback-8fabe4c2/debug-head.jsonl > debug-head.log.txt 2>&1
env -u LATENT_FULL -u LATENT_SPECULATIVE -u LATENT_ONLY_FILE -u LATENT_PROGRESS_FILE LATENT_ASSERT_NO_OUTPUT=1 GOMEMLIMIT=5GiB LATENT_TIMES=/workspace/overload-census/stage3/census/overload-callback-8fabe4c2/entry-parent-times.jsonl /workspace/overload-parent-census /workspace/overload-adapted/src/tsc/tsc.ts /workspace/overload-census/stage3/census/overload-callback-8fabe4c2/entry-parent.jsonl > entry-parent.log.txt 2>&1
env -u LATENT_FULL -u LATENT_SPECULATIVE -u LATENT_ONLY_FILE -u LATENT_PROGRESS_FILE LATENT_ASSERT_NO_OUTPUT=1 GOMEMLIMIT=5GiB LATENT_TIMES=/workspace/overload-census/stage3/census/overload-callback-8fabe4c2/entry-head-times.jsonl /workspace/overload-head-census /workspace/overload-adapted/src/tsc/tsc.ts /workspace/overload-census/stage3/census/overload-callback-8fabe4c2/entry-head.jsonl > entry-head.log.txt 2>&1
```

After both sides complete, pair with the parent's pinned stock TypeScript:

```sh
python3 stage3/census/pair/pair.py PARENT.jsonl.gz HEAD.jsonl.gz /workspace/overload-adapted --typescript stage3/api/node_modules/typescript --json paired.json --report paired.md > pair.log.txt 2>&1
```

The first preparation transport was interrupted; the retained status files distinguish the failed early overlay, recovered continuing build/apply exits, and successful setup retry. The actual tool commands and phases are preserved in the neighboring status and launch JSON files.
