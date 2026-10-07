# Type-aware wave 05

Branch: `codex/typeaware-wave-05`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Claimed rules, positions 13, 14 and 15 after excluding the 26 existing ports
from the combined checker-dependent ranking recorded by VOLUME_REPORT.md:

| Position | Rule | Compiler | Repository | Total |
| --- | --- | ---: | ---: | ---: |
| 13 | @typescript-eslint/no-unnecessary-type-parameters | 30 | 1 | 31 |
| 14 | no-useless-assignment | 17 | 12 | 29 |
| 15 | @typescript-eslint/restrict-template-expressions | 23 | 0 | 23 |

Selection uses `validation-volume/compiler-all.counts` and
`validation-volume/repository-all.counts`, descending combined count with lexical
ties. The 16 volume-suite and 10 coverage-suite ports are excluded before
indexing. All origin heads were fetched and searched for matching port filenames
and claim contents before this commit. No matching port or claim was found;
configuration references are not implementations. No rules were skipped.

## Continuation 1

After completing and pushing the original three rules through c5403d46,
fetched all origin heads on October 7, 2026. Combined VOLUME_REPORT.md's
197 checker-dependent count rows, sorted by descending combined count with
lexical ties, and excluded ports on origin/main and origin/codex/tsgo-c-library
and names reserved in every origin branch's typeaware claims.

The next three unclaimed rules are reserved here before implementation:

- nexus/correctness-no-process-exit-after-output (0 compiler, 0 repository)
- nexus/correctness-no-uncleared-race-timeout (0 compiler, 0 repository)
- nexus/correctness-require-blocking-standard-streams (0 compiler, 0 repository)

These names have no implementation on main or the bridge branch. Inventory,
configuration and skip-types records are not ports. Positive controls are
required because the recorded corpus counts are zero. Implementation stays in
this wave's own directories, without editing the shared generator or harness.
