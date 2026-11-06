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
private-field cache. Oversized files therefore use a symmetric proof-only unit
selector in both scratch overlays. No shipped census template or compiler file
is changed for selection. `build-bounded.py` records that transformation and
`validate-selector.py` verifies unfiltered byte identity and exact selected-unit
coverage. `unit-plan.json` partitions the remaining 9,905 candidate identities
into 257 initial bins. The original six-file and fourteen-file groups cover the other
809 candidates. The complete inventory has 10,714 unique candidate identities.

`run-units.py` runs two independent pairs concurrently, with before then after
inside each pair, `GOMEMLIMIT=4GiB` on both sides, and a 1,200-second wall cap.
These wall times are resource bounds, not a performance benchmark. Completed
results are atomically saved, committed and plainly pushed under a publication
lock. Restarts validate saved binary/plan/output hashes and resume remaining
work. No speculative result is presented as cache proof.

After 41 unit pairs completed, Git's missing credential helper stopped a plain
push. The latest matching result remained committed and was then pushed using
the standard GitHub CLI credential helper. One concurrent after run was
interrupted by that controlled stop; its partial output and negative exit status
are preserved under `transport-interruption`, not counted as a completed proof.
No compiler test failure is claimed as pre-existing.

Observed checker pieces include substantial repeated file setup. The remaining
216 initial bins are therefore combined within each file, at most three bins
per new selection, into 96 bounded pairs. `execution-plan.json` keeps all 41
completed original selections unchanged and partitions the remaining candidate
identities exactly once. Both sides retain the 1,200-second cap. The final
verifier checks both plan hashes, the 41 original pairs, the 96 combined pairs,
and both file groups: 139 pairs covering all 79 files and 10,714 candidates.
The unpaired original checker-032 baseline remains saved; the new combined
selection remeasures both sides with matching witness scope.
