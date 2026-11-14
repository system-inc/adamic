# Canonical Path result amendment: checkpoint

Implementation: 61a2277e9fdd4108ed89fb5c9487cee7b959d47c, one commit on
e531d51d66c45368afa7a9598bc6baa8602b4b41. Authorization: #xn6fj0p.
Scout: 13b63e39bb0b304c83d4ee78ed5dd0ed9f8e3e1e, road (b).

The implementation branch is pushed. The thirteen-owner contract passes,
including LF/CRLF, refusal cases, unrelated edits and byte/mtime-preserving
no-write replay. Stock TypeScript 6.0.3 has zero diagnostics on the main baseline's
79 recursive compiler roots. The baseline emits 604 JavaScript source files.

This is an incomplete checkpoint. Full census, restoring mutant, amendment
stock/types, all-file byte comparison, declaration comparison, and whole lane
are pending. No completed acceptance or lowering claim is made here.
The full unfiltered lane is running on the implementation commit.

All repository source edits and build trees are in the isolated scratch worktree.
Initial stage3/api npm ci and baseline apply used their default cache location;
subsequent commands explicitly redirect npm, stage3, Go and temporary caches to
scratch using env.sh. This is a setup deviation from the all-writes-in-scratch
instruction; the shared checkout's source files were not changed.

No Loom submission was made.

## Second checkpoint

Fresh amendment apply passes. Stock TypeScript 6.0.3: 0 diagnostics on all 79
roots for both trees (0 new). All 604 emitted JavaScript files have identical
SHA-256 values, with no added or removed output file. Upstream baseline `hereby
dts` passes. Amendment census, mutant, declarations, types and whole lane are
still pending; this checkpoint is not acceptance.
