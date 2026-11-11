# Paired view guard roads census

Task 3zex15c. In progress. Report only.

Head branch checked remotely at start and still points to 769b0e534aa9106cec787a670d0243cd3734c9c0. Parent 16429a304. Both compiler worktrees use their own source SHA and pinned cohere 5360753e19e4417675a90e92a1823bea66b13918, TypeScript-go d92d9bfee114c80be2c375d72edae966176e3a4f. Shared adapted tree uses only parent's apply.sh, TypeScript 6.0.3 source 050880ce59e30b356b686bd3144efe24f875ebc8.

The previous census method at 287b32b3:stage3/census/view-pairs-734aef5f/REPORT.md was read. Its bounded, per-unit pointer-deduplicated traversal is historical evidence, not a prediction of this guard's new paths. These runs use the unchanged meter latent overlay and entry overlay, with LATENT_ASSERT_NO_OUTPUT=1. Full-mode checks use LATENT_FULL=1 and LATENT_ONLY_FILE for each affected adapted file. Raw and distinct comparisons use stage3/census/pair.

run.py records the actual child's wait4 exit status, elapsed monotonic seconds, and peak RSS in KiB. Each child runs in a new process group, which is cleaned on completion. Logs are files. No production source changes are included.

Apply completed with child exit 0 in 1310.184 seconds, peak RSS 2473852 KiB. source-manifest.json hashes every adapted src file. Both compiler builds and all 20 pairing tests passed. First-error measurements are pending.
