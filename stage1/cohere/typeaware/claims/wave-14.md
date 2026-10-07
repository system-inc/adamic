# Type-aware lint wave 14

Branch: `codex/typeaware-wave-14`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Claimed rules, positions 40, 41 and 42 after excluding the 26 existing ports:

1. `@typescript-eslint/no-array-delete`
2. `@typescript-eslint/no-base-to-string`
3. `@typescript-eslint/no-extraneous-class`

Selection uses the combined compiler and repository checker-dependent counts
linked by VOLUME_REPORT.md in validation-volume, descending total with lexical
rule-name ties. Each selected rule has compiler 1, repository 0, total 1.
All fetched origin branch trees were checked for matching claims and rule-named
port files before this claim. No selected rule was found claimed or ported.
No rules were skipped.

## Continuation on October 7, 2026

After pushing c043d14a, fetched all 320 origin refs without recursing submodules.
The first three remaining checker-dependent rules in VOLUME_REPORT.md's combined
compiler and repository counts, descending volume with lexical ties, are:

1. `nexus/correctness-no-global-listener-target-assertion`
2. `nexus/correctness-no-leaked-number-render`
3. `nexus/correctness-no-mock-on-module-namespace`

Each has zero compiler and repository findings. Excluded the original 26 ports
on origin/main and origin/codex/tsgo-c-library and every rule named in Markdown
under stage1/cohere/typeaware/claims on every fetched origin branch, including
incomplete reservations. The earlier collection and discarded-result candidates
are already claimed by other continuations. These three have no matching native
port on either base branch and no matching claim. This update must be pushed
before implementation. Positive controls and comparison-only mutants are required.
