# Setup commands and tool-reported exits

Working directory /workspace/adamic unless specified. These entries report command-tool outcomes; raw setup stdout was in the chat tool transcript. commands.jsonl records the later proof children directly, including cwd, arguments, actual exit and elapsed time. The full exploratory read-only command transcript remains in the chat; this file is not an exhaustive transcription of those inspection commands.

| Command | Exit |
| --- | ---: |
| git fetch origin main | 0 |
| GIT_TERMINAL_PROMPT=0 timeout 45 git ls-remote https://github.com/system-inc/adamic.git refs/heads/main | 0 |
| git worktree add --detach /workspace/scratch/sourcemap-splice-1 origin/main | 0 |
| git submodule update --init --recursive (private worktree) | 0 |
| git clone --no-checkout https://github.com/microsoft/TypeScript.git /workspace/scratch/sourcemap-typescript | 0 |
| npm ci (private worktree stage3/api) | 0 |
| git fetch origin e8a1529e6f6addd4ae8dc28ee07c217bcb7b0f0e (private worktree) | 0 |
| source /workspace/adamic-tools/env.sh; go build -o /workspace/scratch/sourcemap-adamic-main ./cmd/adamic | 1 (missing /tmp/adamic-gate) |
| mkdir -p /tmp/adamic-gate; source /workspace/adamic-tools/env.sh; go build -o /workspace/scratch/sourcemap-adamic-main ./cmd/adamic | 1 (nested submodule still initializing) |
| source /workspace/adamic-tools/env.sh; go build -o /workspace/scratch/sourcemap-adamic-main ./cmd/adamic (after submodules) | 0 |
| git worktree add -b typescript/sourcemap-splice-evidence /workspace/scratch/sourcemap-splice-evidence origin/main | 0 |
| git worktree add --detach /workspace/scratch/sourcemap-workers e8a1529e6f6addd4ae8dc28ee07c217bcb7b0f0e | 0 |
| rmdir cohere; ln -s /workspace/scratch/sourcemap-splice-1/cohere cohere (Workers scratch) | 0; subsequent stamped build exited 1, in journal |
| git checkout --detach 050880ce59e30b356b686bd3144efe24f875ebc8; git describe --tags --exact-match HEAD; node -p 'require("./package.json").version' (upstream clone) | 0 |
| kill -TERM 2843 | 1 (baseline had advanced; no such process) |
| kill -TERM 3230 | 0 (intentional baseline cancellation) |

Both compiler checkouts pin cohere 5360753e19e4417675a90e92a1823bea66b13918 and its nested TypeScript d92d9bfee114c80be2c375d72edae966176e3a4f. Workers reuses that identical checkout through a scratch symlink. Its VCS stamping failed, so the successful retry uses -buildvcs=false; compiler sources are unchanged.

Evidence publication: git add stage3/census/sourcemap-splice-evidence and the initial git commit exited 0. Raw logs use .log.txt because repository ignores exclude .log. The final local commit is amended to include them before its first push. Final verification git status --short; git diff --exit-code; git rev-parse HEAD in the amendment worktree exited 0, with no changes at f2546f1fcb7c76f8dd549094401246296b6998a7.
