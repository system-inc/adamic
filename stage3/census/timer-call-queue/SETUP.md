# Setup receipts

The initial shared checkout lacked the requested adaptation and origin/main. A fresh bare clone of https://github.com/system-inc/adamic.git supplied main f869de8c543afc26075c9f49538041b5273dcff5. A fresh detached worktree was created at that exact tip, /tmp/adamic-timer-queue-work. All branch work is there; evidence is a separate worktree at /workspace/timer-call-queue-scratch/evidence.

Recursive submodule initialization succeeded at the commits in setup.json. No credentials were missing. During setup, scratch Git storage was moved from the small /tmp filesystem to /workspace/timer-call-queue-scratch/repo.git. The active nested transfer was stopped, its incomplete scratch copy removed, absolute cohere worktree configuration repaired and git submodule update --init --recursive --depth 1 completed. The original /tmp Git path is a symlink. Git connectivity verification passed. The interrupted transfer's errors were “clone died of signal 15” and “cannot chdir to ../../../../../adamic-timer-queue-work/cohere: No such file or directory”; these were our scratch relocation, not credentials.

Pristine microsoft/TypeScript was cloned at v6.0.3. git describe --tags --exact-match HEAD confirmed v6.0.3, and HEAD is exactly 050880ce59e30b356b686bd3144efe24f875ebc8. Every apply/lane command sets ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-timer-queue-typescript. npm ci in the detached worktree's stage3/api exited 0.

The managed toolchain is sourced from /workspace/adamic-tools/env.sh. It names /tmp/adamic-gate as TMPDIR; that initially missing scratch directory was created before the successful census/CLI builds. Overlay Go sources and binaries are outside stage3. No Go source was added under stage3.

The full main apply exited 0. The unmodified stage 0 CLI and existing census tool were built at main's compiler code. The census uses make_overlay.py's measurement-only scratch overlay and asserts that no usable IR/output program is returned. Binary hashes are in setup.json. Full explicit commands are in the retained scripts and the lane/oracle execution receipts.
