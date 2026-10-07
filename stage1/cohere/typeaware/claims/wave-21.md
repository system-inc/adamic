# Type-aware wave 21

Base: 0d540f413625f016f20fea39761c7b184f335de6.

Selected from VOLUME_REPORT.md's linked all-family count tables, sorted by
combined compiler and repository volume descending, with lexical ties, excluding
the 26 rules already ported on the base:

| Remaining position | Rule | Combined findings |
| --- | --- | ---: |
| 61 | @typescript-eslint/no-misused-promises | 0 |
| 62 | @typescript-eslint/no-misused-spread | 0 |
| 63 | @typescript-eslint/no-mixed-enums | 0 |

These rules are reserved for codex/typeaware-wave-21. No claim files or named
ports were found on fetched origin branches before this claim.

## Final status

`@typescript-eslint/no-mixed-enums` is implemented and verified in 63746b1e.
`@typescript-eslint/no-misused-promises` and `@typescript-eslint/no-misused-spread`
are unimplemented. Their reservations are released for reassignment.
This wave is incomplete; see ../WAVE_21_REPORT.md.

## Continuation claim

After fetching all origin heads on October 7, 2026, the next three rules never
named in any origin claim file and not ported on main (ef3d907e) or the bridge
branch (5afbdb83) are reserved for this branch:

| Original remaining position | Rule | Combined findings |
| --- | --- | ---: |
| 91 | nexus/correctness-no-collection-misuse | 0 |
| 92 | nexus/correctness-no-discarded-outcome | 0 |
| 93 | nexus/correctness-no-discarded-pure-result | 0 |

Previously named rules, including explicitly released reservations, were skipped.
This claim is pushed before implementation.

Continuation completion: all three continuation rules are implemented and
verified in c565946d. See ../WAVE_21_NEXT_REPORT.md and
../validation-wave-21-next/ for byte comparisons, mutants, sanitizer results
and timings. No further claims were taken after Ahra's correction.

## Released-reservation-aware continuation

Fetched all 325 origin refs after the instruction to include explicit releases.
Wave 22 now actively reserves no-misused-promises, no-misused-spread and
concurrency-no-lost-update, so those released reservations are skipped.
The first three available rules in the same combined-volume ranking are:

| Original remaining position | Rule | Combined findings |
| --- | --- | ---: |
| 97 | nexus/correctness-no-process-exit-after-output | 0 |
| 98 | nexus/correctness-no-uncleared-race-timeout | 0 |
| 99 | nexus/correctness-require-blocking-standard-streams | 0 |

No matching native implementation on origin/main or origin/codex/tsgo-c-library
and no active origin claim was found. These three are reserved for this branch.
This claim update is pushed before writing implementation code.

Released-reservation-aware continuation status: all three new reservations
remain unimplemented. Work stopped at missing whole-program source/module
resolution facts under Ahra's restriction on shared-file edits. See
../WAVE_21_RELEASED_REPORT.md and ../validation-wave-21-released/ for the live
checker control, explicit refusals, probe mutants and Go production tests.
No further rules were claimed.

Released-reservation-aware continuation completion: all three rules are now
implemented and validated. See ../WAVE_21_PROCESS_REPORT.md and
../validation-wave-21-process/. The previous missing-program-facts blocker was
resolved with isolated raw bridge questions and three dispatch arms.

## Fourth batch claim

All previously reserved rules are implemented, tested and pushed through
880f6f6f. Fetched every origin head, audited 347 origin refs and 33 claim
Markdown files, and checked native sources on main ef3d907e and bridge 5afbdb83.
Explicitly released promises, spread and lost-update reservations now have
active wave-22 claims and were skipped. The first three available rules in
combined descending volume, with lexical ties, are reserved here:

| Full ranking position | Rule | Combined findings |
| --- | --- | ---: |
| 146 | no-obj-calls | 0 |
| 147 | no-object-constructor | 0 |
| 148 | no-promise-executor-return | 0 |

None is ported on either base branch or named in any fetched origin claim.
This claim update is pushed before implementation.

Fourth batch implementation and validation: all three native rule classes are
ported with complete finding/fix/suggestion agreement on supported controls and
both frozen corpora, per-rule mutants and released/sanitizer checks. The object-
constructor rule remains blocked on two JSX inputs and one keyword-label input
in the shared parser; its rule and suggestion logic are complete. See
../WAVE_21_CORE_REPORT.md and ../validation-wave-21-core/ for exact refusals.
These reservations are retained, not released. No subsequent batch was claimed.


## Fifth batch claim

Previous native rule implementations and validation are pushed through 6af00212.
The three object-constructor inputs blocked by the shared parser are explicitly
recorded in WAVE_21_CORE_REPORT.md; all remaining rule logic is ported.
Fetched all 389 origin refs and inspected 33 Markdown claim documents and
native source ports on main e011f8f6 and bridge 5afbdb83. Explicit releases
were inspected: exhaustive-deps has active wave-12 and wave-27 claims;
core prefer-promise-reject-errors has active continuation claims;
promises, spread and lost-update have active wave-22 claims.
The first three available entries in the combined descending-volume ranking,
with lexical ties, are reserved here before writing code:

| Full ranking position | Rule | Combined findings |
| --- | --- | ---: |
| 168 | react-hooks/set-state-in-effect | 0 |
| 169 | react-hooks/set-state-in-render | 0 |
| 170 | react-hooks/static-components | 0 |

None is ported on either base branch or named in any fetched origin claim.
This claim update is pushed before implementation.


Fifth batch status: all three reservations remain unimplemented, blocked on
missing native React HIR lowering/SSA and graph transforms. Static-components
also encounters native JSX parser refusals. Independent Go positives, native
parser/sanitizer prerequisites and guard mutants are recorded in
../WAVE_21_REACT_REPORT.md and ../validation-wave-21-react/. No native rule
agreement, qualifying native rule mutants or completed ports are claimed.
Reservations are retained; no subsequent rules were claimed.


Fifth batch native-core continuation: all three validator cores, native post-
dominance and the isolated raw alias/symbol-name question are now implemented.
Prepared-HIR comparisons pass on 91 controls and both frozen corpora, with
native core mutants, released handles and sanitizers. Source-to-HIR lowering,
compilation-unit/memo annotations and production harness integration remain
unimplemented. A two-creator phi also produces two different byte outputs from
unchanged Go on identical input. See ../WAVE_21_REACT_CORE_REPORT.md and
../validation-wave-21-react-core/. This is partial core coverage, not completed
native source ports. No further reservation was taken.


Numeric listener continuation: all thirteen owned modules now export pinned numeric syntaxKinds declarations, checked independently against production Go registrations in native and ASAN builds with a qualifying wrong-kind mutant. The shared parser still exposes only string kinds, so numeric dispatch and handed-node integration remain blocked outside this unit. React source ports are still partial; reservations are retained. No new batch claimed. See ../WAVE_21_LISTENER_REPORT.md.


Rule JSON continuation: wave21_rules/<module>/rule.json now declares the Go listener kinds for all thirteen owned modules, checked against unmodified production Go maps alongside native numeric exports under normal and ASAN builds. This is inactive metadata, not source-handler registration. Shared numeric ParseNode/handed-node integration and native React HIR lowering remain missing. The harness branch now handles .a and suggestions. React reservations remain retained; no new claims. See ../WAVE_21_RULE_JSON_REPORT.md.


Current-main landing continuation: rebased onto f8013f0ba with all 24 patches unchanged. Seven owned wave-21 suites, five inherited bridge suites, expanded external Node oracle and checker/vet passed again, including all rule/listener/adjacency mutants and guards. React source ports remain partial and their reservations are retained. No new claim. See ../WAVE_21_F801_LANDING_REPORT.md and ../validation-wave-21-f801/.


## Parked React analysis reservations

Per Ahra's explicit parking instruction, these three retained reservations are
now PARKED and count as finished for the landing-first cap:

- react-hooks/set-state-in-effect: native source-to-high-level-IR lowering, single-assignment graph transforms, memo erasure/inlining and capture/context translation are missing.
- react-hooks/set-state-in-render: native source-to-high-level-IR lowering, single-assignment graphs, capture/context translation and memo/compilation-unit annotations are missing.
- react-hooks/static-components: native source-to-high-level-IR lowering, single-assignment phi data and compilation-unit classification are missing; JSX parsing also remains a shared dependency.

Native validator cores, numeric and rule.json listener declarations, raw type-name
questions and the completed prepared-HIR byte/mutant/sanitizer proofs are pushed
through 3c46337e5 on main f8013f0ba. The source pipeline is not completed.
Cohere's analysis modules are being ported to Adamic on #dnv6f2c; JSX support is
landing on area/stage1-lint. These claims are parked, not released.


## Sixth batch claim, without high-level analysis

Main f8013f0ba is an ancestor of this branch. The owned branch's seven rule
oracles and inherited bridge/Node/checker checks are green and pushed through
3c46337e5; the three earlier React analysis claims are explicitly parked above.
Fetched all 528 origin refs and inspected all 33 distinct Markdown claim blobs.
Earlier entries with no claim mention are already ported on the bridge base.
Explicitly released reservations remain available unless another active claim
reserves them; promises/spread/lost-update and exhaustive-deps have continuation
reservations and are skipped.

These first three available entries of the 197-rule combined by-volume ranking
do not use cohere's high-level IR, single-assignment or capture modules:

| Full ranking position | Rule | Combined findings |
| --- | --- | ---: |
| 180 | react/jsx-fragments | 0 |
| 181 | react/jsx-no-constructed-context-values | 0 |
| 182 | react/jsx-no-undef | 0 |

No native source implementation was found on origin/main or
origin/codex/tsgo-c-library, and none is mentioned in any origin claim document.
These three are reserved for codex/typeaware-wave-21. JSX syntax support is a
shared dependency, separate from the parked high-level analysis modules.
This claim update is pushed before writing any rule implementation.

Sixth batch native-core status: all three numeric handed-node rule cores are
implemented in wave21_jsx, including all six context-value messages and its memo
stability/escape analysis. Prepared-syntax comparisons pass for 216 controls in
three modes and both frozen corpora, with a real native mutant per rule,
sanitizers, external source/emitted Node and released-handle checks. Source
integration remains BLOCKED on shared JSX parsing and shared node-role/live-fact
adaptation; raw syntax/checker metadata is supplied by an explicit
Go test provider. These are partial native cores, not completed source ports.
Reservations remain retained, not released or automatically parked. See
../wave21_jsx/README.md and ../wave21_jsx/validation/. No new batch was claimed.

## Final available rule claim

All owned changes are landing-ready and pushed through
12b844c8eeeac8dbd6ce13bba8aada590c236d78, rebased onto current
main 39638d9e278d38bb5aeae887f46d55a70e47aaad. The seven earlier suites
rebuild and pass in 739.445 s, and all JSX byte-oracle, native mutant, sanitizer,
released-handle and targeted Node checks are green. The existing source gaps
remain explicitly reported; no shared registration generator or harness changed.

Fetched all 603 origin refs and inspected all 33 distinct Markdown claim blobs.
The 197-rule combined ranking has only one remaining available entry, rather
than three: **react/style-prop-object**, full ranking position 192, zero compiler
and repository volume. No native port exists on main or the bridge base, and
no origin claim mentions this name. All other entries are ported on the base or
actively reserved. Explicit releases were checked: promises, spread and
lost-update now have wave-22 claims; exhaustive-deps has wave-12/27 claims; the
released core promise-rejection attempt has other active continuation claims.

react/style-prop-object is reserved for this branch. It does not require
high-level IR, single-assignment or capture analysis. JSX parsing and shared
checker/node integration remain dependencies; the createElement and identifier
logic will also be ported. This claim is pushed before writing its implementation.
No additional available rule is omitted to manufacture a three-rule batch.

Final rule native-core status: style-prop-object is implemented in its owned
wave21_jsx/style-prop-object directory, including JSX, createElement, all pragma
bindings, first declarations and shorthand. All 82 controls in three profiles,
source/emitted Node, full raw-AST frozen corpora, sanitizers, a byte-only native
rule mutant and real released-handle checks pass. Native JSX parsing refuses
the source witness at byte 5, and shared node/checker adaptation and production
registration are not connected. This is a partial source port with its complete
prepared-AST core pushed; the reservation is retained, not released. See the
owned README and validation logs for exact times and coverage. No further rule
is available in the audited ranking, and no further claim is taken.
