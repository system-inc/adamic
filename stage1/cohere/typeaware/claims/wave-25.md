# Type-aware wave 25

Branch: codex/typeaware-wave-25
Base: origin/codex/tsgo-c-library, 0d540f413625f016f20fea39761c7b184f335de6.

Selection combines validation-volume/compiler-all.counts and
validation-volume/repository-all.counts, descending total then lexical rule name,
excluding the 26 rules documented in README.md and COVERAGE_REPORT.md.
The method-signature-style baseline port is absent from the 197 checker registry
rows and therefore does not remove a row from this ranking.

| Remaining position | Rule | Status |
| --- | --- | --- |
| 73 | accessor-pairs | Skipped: already ported on origin/codex/stage1-lint-batch4 at stage1/cohere/lint/grouped_accessor_pairs.ts |
| 74 | base/correctness-require-graphql-nullable-parity | Claimed by wave 25 |
| 75 | base/correctness-require-matching-inject-type | Claimed by wave 25 |

All three have zero findings on both recorded populations. No replacement is
selected for the skipped rule. All origin branch tips were fetched and distinct
stage1/cohere source and claim blobs checked before this claim.

## Next three, October 7 continuation

Checked origin/codex/tsgo-c-library at 5afbdb83da2ed7ad9815657cd3f6ececd5294bf6
and origin/main at ef3d907ecdc4c771b016f7d9c52372def057a340.
Fetched all 320 origin refs; inspected the 30 distinct claim blobs.
These are the first three by descending combined volume, lexical ties,
excluding ports on those two branches and every rule named in any origin claim.

- nexus/correctness-no-collection-misuse (combined volume 0)
- nexus/correctness-no-discarded-outcome (combined volume 0)
- nexus/correctness-no-discarded-pure-result (combined volume 0)

Claim commit b9a19759 was pushed before implementation.
These three are completed in the continuation report; no further rules are claimed.
