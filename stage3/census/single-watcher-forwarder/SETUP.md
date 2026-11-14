# Setup command record

All source edits and proof outputs are in /workspace/scratch. The shared source
checkout /workspace/adamic was only read; git fetch and worktree registration
necessarily updated its shared Git object/ref metadata. No Loom command ran.
No missing-credential error occurred.

Initial setup commands (shell exits):

- `git worktree add --detach /workspace/scratch/single-watcher-forwarder 10f8f6b4d0f7ce9d56f2558752eab3f095b6d9b2`: 128, `fatal: invalid reference: 10f8f6b4d0f7ce9d56f2558752eab3f095b6d9b2`.
- `git fetch origin 10f8f6b4d0f7ce9d56f2558752eab3f095b6d9b2 typescript/scout-514-1280`: 0.
- Same `git worktree add --detach` retry: 0.
- `git fetch origin refs/heads/typescript/scout-514-1280:refs/remotes/origin/typescript/scout-514-1280`: 0.
- `git submodule update --init --recursive --depth 1`: 0.
- `git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/scratch/single-watcher-typescript`: 0; exact HEAD 050880ce59e30b356b686bd3144efe24f875ebc8.
- `npm ci` in the scratch worktree's stage3/api: 0.
- `ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/single-watcher-typescript stage3/apply.sh /workspace/scratch/single-watcher-before`: 0; base-apply.log.gz.
- `go version`: 0, go1.27.1 linux/amd64.
- `go build -o /workspace/scratch/single-watcher-evidence/adamic ./cmd/adamic`: 0; repeated with a full machine-readable receipt in commands.jsonl.
- `git ls-remote --heads origin typescript/single-watcher-forwarder-1 typescript/single-watcher-forwarder-evidence`: 0; neither remote branch existed.
- `git worktree add -b typescript/single-watcher-forwarder-evidence /workspace/scratch/single-watcher-evidence-branch 10f8f6b4d0f7ce9d56f2558752eab3f095b6d9b2`: 0.
- `git switch -c typescript/single-watcher-forwarder-1`: 0.
- `git commit -m 'Evidence checkpoint: single watcher Road C setup and proof drivers'`: 0, 0dc8f5e7b.
- `git push origin HEAD:refs/heads/typescript/single-watcher-forwarder-evidence`: 0, created branch without force.

The initial npm installation used npm's default home cache, and baseline apply
used its default temporary/cache locations. That is a setup deviation from the
all-writes-in-scratch request. Subsequent commands source single-watcher-env.sh,
which explicitly directs temporary files, npm, Go, stage3 and build caches to
scratch. npm ci was repeated with that environment. No shared source file was
edited. This deviation is disclosed rather than claiming an entirely
scratch-only session.

commands.jsonl records subsequent proof commands, all arguments (including all
79 type roots), selected environment, exit and duration. Failure attempts are
retained: the initial runtime VM omitted FileSystemEntryKind; runtime-fixed
supplies its actual File=0/Directory=1 values and passes. A shard-plan generation
invocation without the required artifact argument exited 2; the sanction
regeneration was byte-identical, so no input hashes moved and the unchanged
three generated files pass their freshness check. An initial caller-search
Python literal accidentally encoded word boundaries as backspaces (exit 1);
the corrected raw-string search and AST count are the recorded inventories.
