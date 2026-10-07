# Type-aware wave 27

Branch: codex/typeaware-wave-27.
Base: origin/codex/tsgo-c-library, 0d540f413625f016f20fea39761c7b184f335de6.

Claimed rules, positions 79 through 81 after excluding existing ports:

- base/correctness-require-orm-column-nullable-parity
- base/correctness-require-serializable-nullable-parity
- base/correctness-require-verify-array-parity

Ranking reconstructed from VOLUME_REPORT.md's linked compiler-all.counts and
repository-all.counts, descending combined count with lexical ties. All three
have zero compiler and repository findings. The 26 existing ports exclude 25
entries from this 197-rule checker ranking; method-signature-style is outside
that checker-only table. No matching port file or claim was found in stage1
on any fetched origin branch before this claim. No rules were skipped.

## Second batch

The initial three ports, their independent byte oracle, normally exiting mutants,
corpus/sanitizer checks and evidence were pushed in
5da2444b15f307146e3353fe840c66f54bb62340 before selecting this batch.

After fetching all origin heads without recursion, the first available entries
in the same 197-rule combined by-volume ranking are:

- nexus/correctness-no-process-exit-after-output
- nexus/correctness-no-uncleared-race-timeout
- nexus/correctness-require-blocking-standard-streams

All have combined count zero. Checked ports on origin/codex/tsgo-c-library and
origin/main and Markdown claim records on all 325 fetched origin refs, including
unqualified names for these candidates. No matches were found for these three.
This claim update is pushed before implementing the second batch.

## Third batch

The first six ports were tested and pushed at
c0cd687567bcacbce9d6be2383ee3dd17ed3ca31 before this selection.
After fetching all origin heads without recursion, checked the 197-rule
combined by-volume ranking against ports on origin/main and
origin/codex/tsgo-c-library and Markdown claims on all 356 origin refs.
The first three remaining rules, each with combined count zero, are:

- prefer-regex-literals
- prefer-rest-params
- react-hooks/exhaustive-deps

Claim matching includes full, unqualified and underscore spellings. This update
is pushed before writing the third batch's implementation.

## Fourth batch

The first nine ports were tested and pushed at
24d4f417aa8800a6fc22f70fc8eaa8263f5d9a60 before this selection.
After fetching all origin heads without recursion, checked the 197-rule
combined by-volume ranking against ports on origin/main and
origin/codex/tsgo-c-library and Markdown claims on all 389 origin refs.
The first three remaining rules, each with combined count zero, are:

- react-hooks/set-state-in-effect
- react-hooks/set-state-in-render
- react-hooks/static-components

Claim matching includes full, unqualified and underscore spellings. This update
is pushed before writing the fourth batch's implementation.

Fourth-batch status: parked under Ahra's analysis-dependent React exception.
These claims count as finished only for the landing-first work-in-progress cap;
they are not implemented ports and retain their existing independent probes.

- react-hooks/set-state-in-effect: parked on native high-level IR, single-assignment
  value propagation and translation across captured closure contexts.
- react-hooks/set-state-in-render: parked on native high-level IR, single-assignment,
  post-dominance and nested setter propagation through function contexts.
- react-hooks/static-components: parked on native high-level IR and single-assignment
  phi/taint propagation; positive JSX input is also refused by the current parser.

Cohere's analysis modules are being ported on #dnv6f2c. JSX support is landing
through area/stage1-lint; this unit never pushes to that branch. Shared parser,
analysis, generator and test harness files remain untouched. The native JSX
probe exits 70 while Go reports one finding; see wave_27_fourth/PARKED.md and
the historical fourth-batch and landing reports. Earlier pending/stop language
in those frozen reports describes their earlier status and is superseded here.

## Fifth batch

The nine completed ports are rebased onto current main f8013f0b and green at
0bed2f946. The three fourth-batch React claims were explicitly parked and pushed
at 76e178e86 under Ahra's native-analysis exception before this selection.

Fetched all 529 origin refs and audited the 197-rule combined by-volume ranking
against ports on main/the bridge base and 33 Markdown claim records across all
origin branches. Remaining React-family rules are excluded under the parking
instruction. A remaining capture/escape-analysis rule is also excluded. The
first three eligible entries, each with combined volume zero, are:

- require-await
- symbol-description
- valid-typeof

None is ported on main/base or named in any origin claim record. These rules
use syntax and ordinary checker type/symbol facts rather than native HIR, SSA
or capture analysis. This claim update is pushed before implementation.
