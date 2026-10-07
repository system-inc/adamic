# Type-aware wave 04

Branch: `codex/typeaware-wave-04`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Claimed rules, positions 10, 11 and 12 after excluding the 26 base ports
from the combined compiler/repository ranking referenced by VOLUME_REPORT.md:

| Position | Rule | Combined findings |
| --- | --- | ---: |
| 10 | nexus/consistency-require-constant-casing | 84 |
| 11 | nexus/consistency-require-matching-return-type | 74 |
| 12 | @typescript-eslint/strict-void-return | 56 |

Counts come from validation-volume/compiler-all.counts and repository-all.counts,
sorted by descending combined count with lexical ties. Both oracle suites define
the excluded base ports. All origin heads were fetched and inspected before
this claim. Matches in inventories and count logs are not implementations.
No existing port or claim was found for these three rules. No rules were skipped.

New Adamic source files use `.a`. Checker extensions, if needed, have separate
files per question, with only one-line registration changes in shared files.

## Continuation claim

Original three rules completed and pushed through `67b15e35`.
Fetched all origin heads (325 remote references) on October 7, 2026.
Excluded the 25 checker-dependent base ports (the 26th is not in the checker
ranking) and every rule named in claim Markdown on any origin branch (98).
Combined counts descend, with lexical ties; all positive-volume candidates
are ported or claimed. The first three available entries are:

- nexus/correctness-no-process-exit-after-output (0)
- nexus/correctness-no-uncleared-race-timeout (0)
- nexus/correctness-require-blocking-standard-streams (0)

These rules are claimed for the continuation. This claim is committed and
pushed before any implementation. New code stays in this worker's directories;
the shared registration generator and existing harness will not be edited.

Continuation status: unfinished. Two raw checker questions and matching `.a`
decoders are prepared and compile, but public bridge dispatch cannot reach them.
The required shared registration edits have not been authorized under Ahra's
correction; no shared files were edited. See
`../wave_04_next/BLOCKER_REPORT.md` for the positive probe and exact missing lines.
No continuation rule parity, mutant, sanitizer or runtime result is claimed.

Continuation implementations completed in `99c9e8d3`. All three pass complete
native/Go comparisons over the frozen compiler and repository manifests,
positive controls, ASan/UBSan/leaks, per-rule span mutants and released-handle
checks. See [the continuation report](../wave_04_next/REPORT.md).

## Second continuation claim

The first continuation is complete and pushed through `8673be2b`.
After a fresh fetch of all 358 origin references, the combined volume ranking
contains 197 checker-dependent rules. The 25 ranked baseline ports and 139
rules named in remote claim files are excluded. The first three remaining are:

- `react-hooks/preserve-manual-memoization` (0)
- `react-hooks/purity` (0)
- `react-hooks/refs` (0)

These three are claimed before implementation. Their zero inventory counts are
selection evidence only. Native code will live in `wave_04_react/`; shared harness
and registration files remain outside this worker's changes.

Second continuation status: reporting portions are compiled and tested in
`758b9c60`. The three React analyses remain unfinished and claimed, blocked by
missing native React HIR and the shared parser's JSX support. No further rules
are claimed. See [the precise blocker report](../wave_04_react/REPORT.md).

## Parked React analyses

Status: **parked**, as authorized by Ahra's React parking instruction.
The partial native implementations and all available oracle evidence are pushed
on `codex/typeaware-wave-04` through `5e6e0b59`, based on main `f8013f0b`.
These claims count as finished for the landing-first cap, while remaining
reserved here and explicitly incomplete as full source lint ports.

- `react-hooks/preserve-manual-memoization`: blocked on native high-level IR,
  single-assignment lowering and the reactive memoization pipeline.
- `react-hooks/purity`: blocked on native high-level IR, phi values and closure
  capture analysis.
- `react-hooks/refs`: blocked on native high-level IR, single-assignment/phi
  values, capture propagation and source-to-ref transfer analysis. Its lattice,
  environment and finding-predicate kernels already match production Go.

Shared prerequisites: cohere analysis module ports on `#dnv6f2c` and JSX support
landing on `area/stage1-lint`. Numeric `listenerKinds` and `rule.json` `kinds`
are present for all three. Full source execution remains explicitly refused;
no full React corpus agreement or full-rule mutant is claimed. See
[the current report](../wave_04_react/REPORT.md).

## Third continuation claim

After parking the three HIR-dependent React rules, fetched all origin heads:
529 remote refs, 197 ranked checker-dependent rules, 154 rules named in remote
claims and 25 baseline/main ports. Eighteen entries remain unclaimed. The first
three by descending combined count and lexical ties are claimed here:

- `react/jsx-fragments` (0)
- `react/jsx-no-constructed-context-values` (0)
- `react/jsx-no-undef` (0)

These rules use AST and checker symbol/declaration questions, not React native
HIR, single-assignment or capture analysis. The claim is pushed before new
implementation. Code will stay in `wave_04_jsx/`, with separate rule directories,
numeric `rule.json` listeners and node-local judgment entry points. Shared
parser, driver, registration generator and harness remain outside this unit.
Selection provenance is preserved in `wave-04-jsx-selection.json`.

Third continuation status: partial numeric-node kernels and diagnostic reporters
are implemented and independently checked. Full source execution remains refused:
shared numeric JSX node integration is unavailable, and rule-local binding,
component and stability/escape decisions remain unfinished. These claims remain
reserved here. See [the JSX report](../wave_04_jsx/REPORT.md). No additional rules
are claimed.

Named harness update: adopted ab70f38d4 atop current main c01907a7. JSX syntax parsing now succeeds. The three JSX claims remain partial and reserved: shared Descriptor.kinds rejects numeric kinds, and own binding/attribute/component/stability source judgments remain unfinished. HIR React claims remain parked for native HIR, single-assignment and capture analysis. All six completed-rule oracles and partial helper checks are green again; no new claims. See ../wave_04_jsx/HARNESS_AB70.md.

Constructed-context progress: value-attribute extraction is now ported, with production-Go AST controls and sanitized comparison-only mutants. Current JSX claims remain partial and reserved; remaining shared numeric registry compatibility and own source judgments are listed in ../wave_04_jsx/ATTRIBUTES.md. No new claims.

Named-kind correction: all twelve owned listener manifests and module declarations now use upstream ast.Kind names per the latest instruction. Three JSX manifests deserialize successfully in the actual shared Descriptor; the earlier numeric compatibility blocker is resolved and superseded. Full JSX descriptors/adapters and source binding/component/stability judgments remain unfinished owned work; claims stay reserved. See ../wave_04_jsx/NAMED_KINDS.md. No additional claims.

Landing update: rebased onto b8fb957aa with merge topology preserved, adopted harness 41eb6eab2, and re-greened all six completed rules plus partial JSX/parked React helpers and named metadata. Earlier flattening-rebase blocker is resolved without shared edits. JSX claims remain reserved and unfinished. No new claims. Evidence: ../wave_04_jsx/LANDING_B8.md.

Integration landing update: rebased onto area/stage1-lint 7481e0324, containing required harness integration 50a5f105 and main 39638d9e2. Six completed-rule oracles and partial JSX/parked React/listener checks are green again. Current JSX claims remain reserved and unfinished; no new claims. See ../wave_04_jsx/LANDING_AREA.md.
