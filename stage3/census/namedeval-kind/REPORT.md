# namedEvaluation kind census — #0kkpw2e

Initial measurement checkpoint. Report only: no adaptation, ledger or internal/ edit.

B = 932a6344af63dc7c8728fffac05584c176d0475f (merge base of dfff19bb and fetched origin/main 932a6344af63dc7c8728fffac05584c176d0475f). S1 = dfff19bb0ab5722136f7c19ea01505a1ab0946df. S2 = 12047c681cae67e2e8b83d2c2b57d969237bdd3f, stacked on S1.

All three apply keys match. One adapted TypeScript tree is produced by B and shared, unchanged, across all measurements. Its 79 compiler source hashes are in inputs.json.

Each compiler has its own worktree, own freshly generated latent overlay, own freshly generated scan-all overlay, and separate freshly built binaries. Only unchanged Go dependency caches are shared. No probe or overlay is reused across revisions. Build commands and child exit codes are retained beside this report.

Requested canonical run matrix: each revision runs full single-file namedEvaluation.ts and esDecorators.ts with LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, exact LATENT_ONLY_FILE and CommonJSBundle; then scan-all rooted at namedEvaluation.ts. Results and answers are pending.

Delta matching will use kind, phase, reason and message with positions stripped. Full original words, source positions and unit attribution will remain in the raw and paired records. A later wall reached in an already-blocked unit will be graded reached rather than new.
