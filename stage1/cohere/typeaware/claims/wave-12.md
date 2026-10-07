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

## Continuation after completing the first three

The first three are implemented and pushed in `6f852293`, with evidence in
`1a1be27f`. After fetching all origin heads, these are the first three entries
in the combined VOLUME_REPORT ranking neither ported on main/the bridge branch
nor claimed under the claims directory on any origin branch:

- `nexus/correctness-no-process-exit-after-output` (0 compiler, 0 repository).
- `nexus/correctness-no-uncleared-race-timeout` (0 compiler, 0 repository).
- `nexus/correctness-require-blocking-standard-streams` (0 compiler, 0 repository).

The audit inspected 324 origin refs, with 98 ranking names claimed and 25
matching port names on main/the bridge branch; 74 entries remained. Exact
module filename checks across all origin branches found no competing ports.
This claim update is pushed before implementation. Existing shared harness and
registration generator files will not be edited.

## Third batch after completing six rules

The second batch is implemented in `75fb78e8`, with validation in `97aab772`.
After fetching all origin heads, the next three unclaimed ranking entries are:

- `no-new-func` (0 compiler, 0 repository).
- `no-new-native-nonconstructor` (0 compiler, 0 repository).
- `no-new-wrappers` (0 compiler, 0 repository).

The audit inspected 341 origin refs, 118 claimed ranking names and 25 checker
ports on main/the bridge branch. There were 54 remaining entries. Matching
module filename checks across all origin refs found no existing ports of these
three. The earlier remaining Nexus entries are now claimed and were skipped.
This claim update is pushed before implementation; shared files stay untouched.

## Fourth batch after completing nine rules

The third batch is pushed in `32f968f3`, with evidence in `ba3b7db2` and
compressed reference inputs in `5da49595`. After fetching all origin heads,
the next three available checker-dependent ranking entries are:

- `prefer-regex-literals` (0 compiler, 0 repository).
- `prefer-rest-params` (0 compiler, 0 repository).
- `react-hooks/exhaustive-deps` (0 compiler, 0 repository).

The audit inspected 356 origin refs, 133 claimed ranking names and 25 checker
ports on main/the bridge branch, leaving 39 entries. Matching module filename
checks across all origin refs found no competing ports. This claim update is
pushed before implementation. Shared harness and generator files stay untouched.

## Fifth batch after completing twelve rules

Fourth batch implementation 974dc275 and evidence acb1ff7c are pushed.
After fetching all origin heads, the next three available checker-dependent
entries in the combined VOLUME_REPORT ranking are:

- `react-hooks/set-state-in-effect` (0 compiler, 0 repository).
- `react-hooks/set-state-in-render` (0 compiler, 0 repository).
- `react-hooks/static-components` (0 compiler, 0 repository).

Audit: 389 origin refs, 142 claimed ranking names, 25 checker ports on
main/the bridge branch, 30 remaining entries. Exact matching module filenames
across origin refs found no competing ports. This claim is pushed before code.
Shared harness and generator files remain untouched.

## Parked React HIR batch

Status: PARKED under Ahra's explicit continuation instruction.
- react-hooks/set-state-in-effect: missing native source HIR Lower/Construct/SSA, captures, memo erasure/inlining and control dependence.
- react-hooks/set-state-in-render: missing native source HIR Lower/Construct/SSA, captures and compilation-unit/memo scope preparation.
- react-hooks/static-components: missing native source HIR/SSA and JSX lowering.

These are not completed source ports. Existing evidence is in wave12_fifth/REPORT.md. The parked claims count as finished only for the landing-first cap. Analysis modules are being ported on #dnv6f2c; JSX support is landing on area/stage1-lint. This worker does not push those integration branches.

## Sixth batch after parking the HIR claims

Fresh all-origin audit: 529 refs and 33 distinct Markdown claim blobs. Ranking comes from both VOLUME_REPORT linked count files. Base ports were checked by registry names in native source, because suite implementation filenames differ from rule names. Eighteen candidates remained; exact native module filename checks on all origin refs found no existing modules for them. The first three which do not consume the parked HIR analysis are:
- react/jsx-fragments
- react/jsx-no-constructed-context-values
- react/jsx-no-undef

All have zero compiler/repository volume. Production implementations use AST, scope and checker declarations, not React HIR Lower/Construct/SSA. This reservation is pushed before implementation. Native JSX parsing and supplied-node source integration are dependencies to verify against the shared landing; rule.json declares named ast.Kind listeners. No new shared files will be edited.

Sixth-batch status: ACTIVE, INCOMPLETE, blocked on shared checker integration and incomplete source discovery/callbacks. Native JSX parsing is available after the area rebase. Native message rendering is implemented and validated separately in wave12_sixth/REPORT.md. This is not source finding parity, and these three claims are not the parked HIR batch.

Landing refresh on fetched main c01907a70: all twelve completed default-options ports re-green; exhaustive-deps uses the shared table RegExp literal. The three sixth-batch JSX claims remain ACTIVE and INCOMPLETE because native JSX source parsing is absent on this main; numeric callbacks are no longer required by the corrected contract. Announced harness ab70f38d4 is not yet on fetched main. No additional claims. Evidence and fresh timings: ../wave12_fifth/landing/C019_REPORT.md.

Named-kind correction: all twelve completed listener exports and three active rule.json lists now use pinned ast.Kind names; active descriptors set node: true. A numeric adapter is not a blocker. Current source blocker is native JSX parsing; source callbacks/verdicts remain incomplete. All twelve oracles re-green on c01907a70, and reporting/metadata checks pass. No additional claims. Evidence and fresh timings: ../wave12_fifth/landing/NAMED_REPORT.md.

Landing refresh on fetched main b8fb957aa: all twelve owned full oracles, sanitizers and mutants re-green after retaining inherited-static-field support. Harness 41eb6eab2 remains outside fetched main; native JSX parsing still blocks the three active source claims. Named kinds/node: true remain correct. No new claims. Evidence and fresh timings: ../wave12_fifth/landing/B8FB_REPORT.md.

Landing refresh on origin/area/stage1-lint 7481e0324 (including current main 39638d9e): JSX parsing is now available. The fourth batch includes the seventeen previously excluded JSX controls and has removed its obsolete refusal guard. The active three rules remain partial: shared RuleContext supplies no checker/file-program handle, source Node has no tsgoProgram export (probe exits 70), and emitted-JavaScript compilation refuses an unlinked checker call (probe exits 1). No shared integration files or additional claims are changed. See ../wave12_fifth/landing/AREA_REPORT.md for current gates, mutants, timings and exact limits.
