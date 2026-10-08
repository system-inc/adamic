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

## Current-main landing refresh

Rebased all owned wave commits onto origin/main f8013f0b. All nine completed
rule gates, numeric listener-map checks, sanitizers, mutants, bridge/decoder
guards, 30 Node fixtures and 12 iterator refusals pass. A disk-full failure
was recovered using obsolete owned scratch binaries; its log and successful
retry are preserved. See ../wave07_react/CURRENT_LANDING_REPORT.md. Three React
claims and four bridge routes remain pending; no new rules are claimed.

## React analysis parking

Under Ahra's October 7 parking instruction, react-hooks/set-state-in-effect,
react-hooks/set-state-in-render and react-hooks/static-components are parked.
Their blocker is missing native high-level IR, single-assignment/value-flow and
closure/capture analyses; cohere's analysis modules are being ported on #dnv6f2c.
JSX support is landing on area/stage1-lint. The pushed probes and dependency
evidence remain; these are not completed implementations. They count as finished
only for the landing-first work-in-progress cap.

## Fourth continuation claim

The owned branch is landing-ready at 3028b011f on current main f8013f0b. A fresh
all-head fetch inspected 529 origin refs and 33 distinct claim blobs. Selection
uses the 197-row VOLUME_REPORT combined compiler/repository ranking and excludes
the original production ports, native ports on main/base and all origin claims.
The first eligible name needing capture/escape stability analysis,
react/jsx-no-constructed-context-values, is skipped until #dnv6f2c. The next three
without that analysis, all zero-volume on the frozen corpora, are:

1. react/jsx-fragments.
2. react/jsx-no-undef.
3. react/no-adjacent-inline-elements.

Only counts/inventory mentions were found for these names on main/base, with no
native implementations. No origin claim matches any of the three. This update
is pushed before rule implementation. New modules use .a, numeric listener kinds
in each rule.json and handlers taking the handed node. Shared parser, generator,
harness and dispatcher files remain untouched. JSX parser integration is still
pending; isolated published-parser validation may establish rule behavior first.

## Fourth continuation implementation status

All three owned handlers are implemented in wave07_jsx as .a modules with numeric
rule.json listeners and handed-node dispatch. On rebased main c01907a7 the
ordinary/sanitized positive controls, four option combinations, both frozen
corpora, three message mutants, three dispatch mutants and released-handle
checks pass. Both corpora also match with the unchanged main parser. Its
positive JSX blocker is GreaterThanToken versus SlashToken at byte 97; shared
JSX parser/driver integration is pending. The isolated published parser is
a8a62d62ca49db7415e14c3887dd305022b17309. No shared files were changed.

An owned, licensed library snapshot resolves virtual declaration-source paths
without a shared bridge edit; all 114 sources pass an external-byte sanitizer
probe and a valid data mutant. All prior nine ports were re-greened after the
rebase, including their bridge/decoder guards, filtered Node oracles and vet.
See ../wave07_jsx/REPORT.md and its evidence. The three earlier analysis-heavy
React claims stay parked with the named IR/SSA/capture blocker. No more rules
are claimed in this update.

## Named-kind landing and exhausted audit

The three JSX rule descriptors now declare the exact ast.Kind names without
the Kind prefix and set node: true. Their full comparison, sanitizer, live/
released-handle and three named-dispatch-mutant gate passes again. The shared
ab70f38d4 context has no checker-program handle, so standalone type-aware
handlers are still not registered there; positive JSX on main remains blocked.

A refreshed audit inspects 579 origin refs and 33 distinct claim blobs, including
main/base descriptors. It has no remaining unported, unclaimed ranking entries.
The previously two available names were claimed by wave 18 during validation.
No new wave-07 claim was made. Main advanced to b8fb957a; the branch is rebased
and its owned oracle includes the inherited-static-field-read fixture. See the
named landing report and fresh evidence for actual green checks and source SHA.

## Integrated parser landing refresh

Rebased all wave commits onto area/stage1-lint 7481e032, containing main
39638d9e. All twelve implemented rules, 17 landing steps, named JSX listener
checks, legacy listener checks, sanitizers, valid mutants and released handles
pass again. Actual integrated parser sources now pass 131 positive JSX controls;
no private parser copy is needed. Three hook claims remain parked solely on
native IR/SSA/capture analysis. Four private bridge routes and shared checker
context access remain pending. See ../wave07_jsx/AREA_LANDING_REPORT.md.
No new rules are claimed by this update.

## Required-input landing refresh

Rebased onto area/stage1-lint d65a8f93, containing current main 39638d9e.
All twelve implemented ports pass again, together with 18 owned landing steps
and the supplied-input external correctness suite: 16 root tests, 100 passing
events, zero skips and zero failures. Exact pinned inputs and no-skip receipts
are archived. React hook claims stay parked on IR/SSA/capture analysis; four
production bridge routes and shared checker context remain pending. See
../wave07_jsx/REQUIRED_INPUTS_REPORT.md. No additional claim is made.

Latest landing refresh: rebased onto origin/area/stage1-lint b84a9d931 and origin/main
c7991b900; all 19 owned landing steps and 100 required external test events pass
with zero skips or failures. Twelve existing ports rerun; three hook claims remain
parked for native IR, SSA and capture analysis. No additional rules claimed.
See wave07_jsx/LATEST_LANDING_REPORT.md and its pinned evidence.

Generated-driver refresh: rebased onto area/stage1-lint b46914832, containing
main c7991b900; all 19 landing steps, JSX/listener/library proofs and 100 external
passing events re-green with zero skips or failures. No additional claims.
The three hook claims remain parked for native IR, SSA and capture analysis.
Fresh source/evidence pins and blockers: wave07_jsx/DRIVER_LANDING_REPORT.md.

Typeof refresh: rebased onto area/stage1-lint d3a37422c containing main b6b1538b0.
All 19 landing steps, JSX/listener/library proofs, 43 Node fixtures and eight
incoming typeof mutant executions pass. Required external receipt: 100 passing
events, zero skips or failures. No additional claims. Three hooks remain parked
for native high-level IR, SSA and capture analysis. See TYPEOF_LANDING_REPORT.md.

## October 8 shared-checker merge landing

Merged the current lint area, without rebasing or deleting either side's checks.
Six existing ports now have owned unified descriptors using RuleContext.checker.
Four fully match every captured row plus witnesses: no-undef-init, prefer-for-of,
prefer-rest-params and jsx-no-undef. Indexed-object-style is 119/121, blocked by
empty index-signature recovery; promise-reject-errors is 123/124, blocked by a
malformed TSX capture rejected by the shared Go adapter. All six owned witnesses
and mutants pass. All 81 registered rule mutants pass. The six remaining existing
ports are pending shared checker routing or external declaration view, and the
three hook claims remain parked. Exact consumers and current package receipts:
wave07_jsx/CHECKER_BLOCKERS.md and wave07_jsx/CHECKER_LANDING_REPORT.md.
No new ownership is claimed.
