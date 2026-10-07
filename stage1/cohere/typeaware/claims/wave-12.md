# Type-aware wave 12

Branch: `codex/typeaware-wave-12`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Claims, positions 34, 35 and 36 after excluding the 26 existing ports:

| Position | Rule | Compiler | Repository | Total |
| --- | --- | ---: | ---: | ---: |
| 34 | @typescript-eslint/no-redeclare | 2 | 0 | 2 |
| 35 | nexus/correctness-no-test-on-global-regex | 2 | 0 | 2 |
| 36 | nexus/correctness-no-write-only-collection | 1 | 1 | 2 |

Selection sums VOLUME_REPORT.md's linked validation-volume/compiler-all.counts
and repository-all.counts, sorts by descending total then lexical rule name,
and excludes the sixteen volume-suite and ten coverage-suite ports.
Fetched every origin branch and checked stage1 claim files and matching port
filenames before this commit. No competing claim or port was found.
