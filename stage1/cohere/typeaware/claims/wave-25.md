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
