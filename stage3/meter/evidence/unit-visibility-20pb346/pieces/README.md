# Bounded compiler-directory comparisons

These comparisons replace the requested whole compiler pass, at the user's
request after repeated environment kills. No new whole pass is launched.

Both binaries load the same complete compiler program on integration base
`d4c6b07aca1a7c55785226563c2ad79d552c1adf`. The before binary uses main's
unchanged latent overlay; the after binary uses this branch's cache overlay.
Both use `LATENT_FULL=1`, `LATENT_ASSERT_NO_OUTPUT=1`, `GOMEMLIMIT=8GiB`, and
exact newline-separated `LATENT_ONLY_FILE` membership. The source argument and
absolute file paths are identical. Runs are sequential and capped at 1,500
seconds each. Raw output, unit timings, checksums, source hashes, binary hashes,
wall times, and exact selections are saved per group.

Each side's evidence is committed immediately after it exits. Once the second
side completes, the runner asserts exact byte identity before committing the
pair verdict. Later runs also push each result immediately. A nonzero exit or
mismatch stops the runner. `coverage.json` lists all 79 files and disjoint groups.

These are comparisons between matching selections. Separate groups cannot be
concatenated to claim whole-pass identity: deferred generic `instantiated`
witnesses have pass scope, so grouping changes which callers contribute them.
This is the same finding that requires latent.jsonl to remain buffered.

The existing atom-piece support is speculative-only and bypasses the full-mode
private-field cache. Checker.ts alone cannot meet the time bound as a full-mode
file group. The choice of full-mode unit selection or speculative checker pieces
has been raised explicitly; no speculative result is presented as cache proof.
