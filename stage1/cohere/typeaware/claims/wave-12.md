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

All have zero compiler/repository volume. Production implementations use AST, scope and checker declarations, not React HIR Lower/Construct/SSA. This reservation is pushed before implementation. Native JSX and numeric supplied-node integration are dependencies to verify against the shared landing; rule.json will declare numeric kinds. No new shared files will be edited.
