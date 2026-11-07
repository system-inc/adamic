# Reproduction and evidence

Source /workspace/adamic-tools/env.sh, run bash cloud/setup.sh and npm ci --prefix stage3/api, and export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source at 050880ce59e30b356b686bd3144efe24f875ebc8. The API parser is stock TypeScript 6.0.3.

The retained scripts use /workspace/adamic and /workspace/scratch/core-compare. prepare.sh records the baseline full apply; instrument.cjs changes only the scratch owner to compute both results while returning the original. count.py runs every lane unit through the unchanged shard.py. reconcile.py independently decodes the canonical archives. control.py runs the unchanged whole lane and compares exact identities, counts and baseline bytes. construct.cjs refuses to construct the adapter until all three gates pass.

post-measure.py runs contracts, actual-source Node witnesses, both ordinary compiler builds, owner-scoped JavaScript/declaration comparison, the after and mutant full censuses, and paired replay. close-proof.py then performs a fresh full apply of the complete main-plus-compare stack, checks all 746 source hashes against the independently proved tree, runs all 79 types roots and the full core census on that exact output, builds it with the pinned packages, and verifies its emitted bytes. Ordinary checks use the unmodified main Adamic binary. Only the full census binary has the reference measurement-only overlay; production Lower is disabled in that binary and both no-output assertions remain enabled.

The setup and early preparation corrections are retained: a slow full-history source clone was replaced by a shallow exact-commit fetch; the scratch types.py name was corrected to check-types.py; the observer was corrected to preserve CRLF. No call from a failed attempt or an excluded seed contributes to COUNT.json.

Canonical piece archives contain raw one-byte calls, decoded histograms, every unit report and observation, baseline diffs and SHA256SUMS. The evidence archive retains complete proof logs and scripts, the excluded seeds, raw censuses, exact emitted hunks, source manifests, owner witnesses, tool hashes and command results. Excluded JSON and packed timing attempts are separate archives. Root SHA256SUMS covers every retained file except itself.

## Compiler amendment

The count is retained byte-identically. COUNT-REUSE.json binds its parent commit and hashes. Debug.fail is reached on 0 of the 144,697,378 calls. Neither the lane nor its count was rerun.

Use /workspace/scratch/core-compare-amend for the current scripts. amendment-apply.py performs both complete applies from the original main base and the amended branch. amendment-proof.py reruns contracts, actual-source Node witnesses, both 79-root checks, both ordinary compiler builds, JavaScript scope and both public declarations, full core.ts censuses, the original-owner mutant and all 84 paired adapter replays. CHILD-EXITS.json records the actual child statuses; EXIT-CANARY.json proves child 7 propagates as wrapper 7.

The amendment archive contains the complete new logs, raw censuses, manifests, exact hunks, replay outcomes and command statuses. The initial proof and count archives remain historical evidence.
