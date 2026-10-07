# Type-aware wave 07

Branch: `codex/typeaware-wave-07`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Claimed rules, positions 19, 20 and 21 after excluding the 26 existing ports
from VOLUME_REPORT.md's combined compiler and repository checker-dependent
ranking (`validation-volume/compiler-all.counts` and `repository-all.counts`):

| Position | Rule | Combined findings |
| --- | --- | ---: |
| 19 | no-undef-init | 15 |
| 20 | @typescript-eslint/prefer-for-of | 13 |
| 21 | @typescript-eslint/consistent-indexed-object-style | 12 |

All origin heads were fetched before selection. No matching rule port filenames
or claims were found under stage1 on those heads. No rule was skipped.
New Adamic sources will use `.a`.

## Continuation claim

The original three ports were tested and pushed through 6264bab2 before this
continuation. On October 7, 2026, all 325 origin refs were fetched. Selection
combines the 197 checker-dependent rows linked by VOLUME_REPORT.md, descending
combined volume with lexical ties. It excludes the 26 existing ports identified
by the production oracle suites, ports on origin/main (ef3d907e) and
origin/codex/tsgo-c-library (5afbdb83), and every rule named in Markdown claims
on any origin branch (96 ranked names). Counts and inventory membership are not
implementations. The following are the first three eligible names:

1. nexus/correctness-no-process-exit-after-output (0 compiler, 0 repository).
2. nexus/correctness-no-uncleared-race-timeout (0 compiler, 0 repository).
3. nexus/correctness-require-blocking-standard-streams (0 compiler, 0 repository).

No matching native implementation or claim for these three was found on the
specified origin branches. This update is pushed before implementation. New
sources use .a; shared registration generator and test harness remain untouched.

Continuation status: blocked before implementation. The first rule needs the
exact resolved call-signature declaration, which existing checker questions do
not expose. A new question requires the shared Program.Inspect dispatcher to
register it; Ahra's correction prohibits shared-file changes and says to stop
on another blocker. No shared files changed in this continuation, and none of
the three new rules is represented as complete. See
../wave07_continuation/BLOCKED.md. No additional rules are claimed.

## Continuation implementation status

All three continuation rules now have native .a implementations and pass private-overlay
Go byte comparisons, per-rule mutants, raw-fact mutants, released-handle checks and
ASan/UBSan/LSan gates. See ../wave07_continuation/REPORT.md. Production dispatcher
registration remains pending under the shared-file restriction; no further rules
are claimed. Earlier blocked status above records the initial capability diagnosis.

## Second continuation claim

All prior claimed ports and their private-overlay validation are pushed through
158557b6; production registration of the first continuation remains pending.
A fresh fetch on October 7, 2026 inspected 356 origin refs and 33 distinct
Markdown claim blobs, covering 132 ranked names. Excluding the original 26
production ports on origin/codex/tsgo-c-library (including abbreviated native
filenames and oracle registry aliases), ports on origin/main, and claims on all
origin heads, the first eligible rules in the combined 197-row volume ranking are:

1. prefer-promise-reject-errors (0 compiler, 0 repository).
2. prefer-regex-literals (0 compiler, 0 repository).
3. prefer-rest-params (0 compiler, 0 repository).

This claim is pushed before implementation. New native files use .a and remain
in the worker-owned wave07_next directory; shared harness and generators are
not edited.

## Second continuation implementation status

The second trio is implemented and validated. prefer-rest-params and
prefer-promise-reject-errors use the normal archive; prefer-regex-literals uses a
private overlay pending one shared dispatcher case. Full byte streams including
nonempty regex suggestions, per-rule mutants, a raw-fact mutant, released-handle
checks and ASan/UBSan/LSan gates pass. See ../wave07_next/REPORT.md and evidence.
No additional claims are taken after this trio.

## Third continuation claim

Prior claimed ports and evidence are pushed through 87ebb321. The fresh all-head
fetch inspected 389 origin refs and 33 distinct Markdown claim blobs, naming
142 ranked rules. Using the combined 197-row ranking and the original oracle
port inventory, excluding ports on origin/main and origin/codex/tsgo-c-library
and claims on every origin head, the first three eligible rules are:

1. react-hooks/set-state-in-effect (0 compiler, 0 repository).
2. react-hooks/set-state-in-render (0 compiler, 0 repository).
3. react-hooks/static-components (0 compiler, 0 repository).

This claim is pushed before implementation. New Adamic files will be .a and
shared harness, dispatcher and generator files remain untouched.

## Third continuation status

Blocked before rule implementation: the shared native parser cannot parse JSX.
All three positive production Go controls report; the native parser exits 70 at
the self-closing JSX slash in each. Ordinary TypeScript parses successfully.
See ../wave07_react/REPORT.md and its reproducible probe evidence. Ahra's
shared-file restriction prevents extending the parser in this unit. All three
claims remain incomplete; no additional rules are claimed.

## Third continuation dependency recheck

Upstream JSX support is now available on origin/codex/stage1-jsx-lint at
a8a62d62ca49db7415e14c3887dd305022b17309. An isolated build accepts all three
previously failing controls (exit 0, jsx 1), plus the ordinary control (jsx 0).
The wave branch's shared parser remains unchanged pending integration by its
owner. No React rule is implemented yet; these three claims remain incomplete.
The updated report records exact dependency evidence. No more rules claimed.

## Landing refresh

Rebased onto origin/main e8ba3d5d after an earlier rebase to e011f8f6. All nine
completed ports were rebuilt and re-greened against production Go, including
sanitizers, per-rule mutants, bridge-fact mutants and released-handle checks.
See ../wave07_react/LANDING_REPORT.md for fresh timings and exact gate evidence.
The three React claims remain unported pending shared JSX parser integration;
the four private-overlay bridge routes also remain pending. No new claims.

## Numeric listener declarations

All nine completed ports now declare numeric listener kinds, verified against
the pinned production Go registration maps (358 identical bytes, sanitizer
checks and nine numeric-key mutants). See ../wave07_react/SPEED_REPORT.md.
The shared parser still exposes string kinds and the shared numeric driver is
pending, so rule-body migration remains blocked. React claims and four bridge
registration routes remain pending. No new rules are claimed.
