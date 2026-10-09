# Lowering chain slices

Slice 1 is the independent c-portability member, built on current main
5e33a17b186a8a2218d27b69b21e2de5acc5b750. Chain head is
f5236b48ee11d680c5288e3074c2fec6ad26a09a; the chain/main
merge base is 88adf55532b89fb155d68eacbf490bea34ffda26.

**Partial plan. Only slice 1 is certified.** This census does not establish a
complete minimal dependency graph or certify the later slice order. An applicable
branch patch does not prove semantic independence. Stacked ancestors are removed from the identified step24, per-backend, search,
optional-presence and eep ranges. Optional-calls still includes lowering-a in
its checked range and remains pending proper extraction. Remaining assignments below are explicitly pending, not green slices.
No later branch is created or stacked on this delivery.

## Member census

Merge parents and complete commit messages, including repair histories, are in
merge-history.txt and commit-history.txt. The table includes the earlier lowering-a
and lowering-b members actually present in the cut, rather than treating those
stack envelopes as new compiler members. Fix/verification branches are listed in
the repair ledger rather than counted a second time as feature members.

Checks use an isolated main index and exclude review, docs and counts. Branch-net evidence is in member-patch-checks.json; refined source-range evidence
is in own-net-patch-checks.json. The table uses the refined ranges. The counts exclusions
avoid confusing old table context with compiler dependencies. The tsgo cache check
is the direct chain commit; only internal/native/tsgo.go is its implementation.

| Member | Source SHA | Extracted-range applies to main | Depends on | Slice | Waiting tasks |
| --- | --- | --- | --- | --- | --- |
| optional-chain-after-call-main | `5028fa69` | yes | pending own-net and semantic audit | pending | task attribution pending |
| spread-shorthand | `e82c9646` | yes | pending own-net and semantic audit | pending | task attribution pending |
| optional-calls-main | `b6ef2cb4` | no | pending own-net and semantic audit | pending | task attribution pending |
| c-portability-main | `ebad12cf` | yes | none (verified) | 1 | #277egjv, #9y65q7e |
| generic-body-relations | `2cc109e5` | yes | pending own-net and semantic audit | pending | #js89dcw |
| generics-scout-main | `48391dcf` | yes | pending own-net and semantic audit | pending | task attribution pending |
| iteration-main | `518eca83` | yes | pending own-net and semantic audit | pending | task attribution pending |
| namespace-value | `26ccff9c` | no | pending own-net and semantic audit | pending | task attribution pending |
| checked-any | `57ea02f8` | no | pending own-net and semantic audit | pending | #mydv4kd |
| project-references-main | `e2aa3750` | yes | pending own-net and semantic audit | pending | task attribution pending |
| refusal-rulings-main | `8799f518` | yes | pending own-net and semantic audit | pending | task attribution pending |
| assignment-proofs-main | `a6f05066` | yes | pending own-net and semantic audit | pending | task attribution pending |
| tsgo.go-cache | `63f5bb07` | yes | pending own-net and semantic audit | pending | task attribution pending |
| miscompile-fxspptb-2b | `6de2edb5` | yes | pending own-net and semantic audit | pending | task attribution pending |
| placeholder-nonnull-main | `bafb9ef4` | no | pending own-net and semantic audit | pending | task attribution pending |
| step24-parser-main | `e503868c` | no | placeholder-nonnull-main; additional edges pending | pending | task attribution pending |
| miscompile-fxspptb-2a | `373f0055` | yes | pending own-net and semantic audit | pending | task attribution pending |
| callback-widening | `36472529` | yes | pending own-net and semantic audit | pending | #1xxqbrk |
| generators-main | `6f0ebd15` | yes | pending own-net and semantic audit | pending | #qk8rztp |
| exceptions-21-main | `205586a0` | no | pending own-net and semantic audit | pending | task attribution pending |
| inherit-guards | `16a0b626` | no | pending own-net and semantic audit | pending | #a898y40 |
| self-compare-main | `387c2826` | yes | pending own-net and semantic audit | pending | #7c6b4pq |
| feature-set-link-main | `e1efb527` | no | pending own-net and semantic audit | pending | #hmab710 |
| per-backend-stops | `7a151f7f` | yes | pending own-net and semantic audit | pending | #z1vjxxd |
| search-shrink | `c0c4e102` | no | pending own-net and semantic audit | pending | #b75jjs3 |
| optional-presence-next | `897d0e79` | no | pending own-net and semantic audit | pending | #r3chqza |
| eep-presence | `21ceb23c` | no | pending audit; optional-presence is its source base, not proof of necessity | pending | #eep1m5z |
| lint-features | `35353870` | no | pending fixture audit; RuntimeLibraryForSource/sourceFlags already exist on main | pending | #wj4pmt1, #hmab710 |

## Certified slice order

1. c-portability-main ebad12cf alone. Closes #277egjv and #9y65q7e after
   integration accepts the delivery. Its code commits 56ee1ddc, 230a996d and
   099a7aaf are represented once as one member commit. No other chain member or
   repair changes its portability implementation. Counts are regenerated on main.

The remaining slices and #63pvx2b attribution are not certified in this delivery.
Two to five members should be grouped only after proving the necessary symbol,
fixture, IR-field and repair edges. Merge order and shared files are not dependency
proofs. The placeholder contract dependency requires own-net validation; other source ancestry is not a technical dependency proof.

## Repair attribution ledger

These are observed repair commits, not newly applied commits on slice 1. A shared
repair must be split by owning hunk when its owners land separately.

| Chain commit | Owner or interaction | Evidence in changed source / commit message |
| --- | --- | --- |
| 9e1e73e0 | generics-scout + generic-body-relations | Scout's indexed-result/constraint expectations reflect the generic-body refusal. |
| d8e6f59c | exceptions-21 + miscompile-2a + optional calls + step24 | Runtime merge preserves heap thrown payload/pending flag, absent Error.message shape, NodeArray keys and required-read TypeError constructor. |
| 72a2497d | optional-calls / stage1 CSS gap | Parallel closed-gap check. |
| 2ca18b1b | checked-any and project-references | Declare compiler-consuming probe packages. |
| aea87960 | indexed witnesses; optional-calls; exceptions-21 | Ordinary element agreement, optional RegExp field/group storage, ruled WASI uncaught exit contract. |
| 91b0148f | project-references and checked-any fixture enumeration | Project selection only with references; ordinary mixed-file options and individual spec witnesses. |
| 1eaed07b | optional-calls / tsprinter lane | Remove stale serial-baseline entry. |
| 71972e18 | exceptions-21, namespace-value, readiness and spread-shorthand fixture owners | Catchable lexical/namespace readiness, MayThrow propagation, selector/YAML push expectations, generator outcome records. |
| a6cb5660 | optional-calls, spread-shorthand and exceptions/readiness test owners | Restore structural Map boundary, exact exit contract, parser/CSS/JSON gap checks and reader table. |
| 9679fbe4 | self-compare / split C header interaction | Keep static inline helpers defined in split headers. |
| f130c846 | search-shrink | Callback search count audit. |
| c775dc61 | readiness oracle guard | Source Node exit and guard-bypass mutant evidence. |
| 379dbfb3 | native records harness interaction | Initialize borrowed-object metadata. Needs owner attribution against the main implementation; not a lowering feature. |
| a233eec3 | namespace-value, miscompile-2a; spread-shorthand; self-compare | Fresh runtime corpus stores, closed push gaps, scanner self-compare guards, fixture directory owner. |
| 2b2b1095 | placeholder, namespace-value, generic-body, step24, per-backend, feature-link | Nil property-symbol guard; exact enum/namespace/stack/WASI expectations and maybeBind fixture. |
| 163e39f8, d26baa81, 84d03d57, c014d545, 897d0e79 | optional-presence-next | Dynamic reserving-copy descriptors, checked optional scalar/string writes and explicit metadata mask. |
| 4478575d | eep-presence | Tuple length and class property-presence containment. |
| 167aaf0c | optional/eep presence and merged reader/counts ledger | Retain main's approved reader, remove stale reader and regenerate combined counts. |
| 35353870 | lint-features / feature-set link | Export SourceFlags and use source-aware flags/runtime at emitted-C profiling sites. |

The e628f527 16-red request is represented by the after-chain repair families
379dbfb3, a233eec3 and 2b2b1095, then verification 7615fc36/50654a40.
Their source report retains the red-list validation. This delivery has not
completed a one-to-one red-to-member/repair-hunk attribution, and does not claim it.
Main merges and evidence-only commits remain in the complete history for review.

## Proposed extraction groups, pending the complete graph

This is a review work order. Only group 1 is a certified independently landing
slice; later groups require the unresolved dependency and repair-hunk audit.
No waiting task below is claimed closed by this delivery except the portability
work, which still awaits integration acceptance.

| Group | Members | Waiting tasks identified |
| --- | --- | --- |
| 1 | c-portability-main | #277egjv, #9y65q7e |
| 2 | generic-body-relations, generics-scout-main | #js89dcw |
| 3 | assignment-proofs-main, refusal-rulings-main | Attribution pending |
| 4 | optional-chain-after-call-main, spread-shorthand | Attribution pending |
| 5 | optional-calls-main | Step 18 storage contract #tvq1eqm |
| 6 | iteration-main, generators-main | #qk8rztp |
| 7 | namespace-value, miscompile-fxspptb-2a, exceptions-21-main | #c0ktyvy runtime reconciliation |
| 8 | project-references-main, checked-any | #mydv4kd |
| 9 | placeholder-nonnull-main, step24-parser-main | Step 24; task attribution pending |
| 10 | miscompile-fxspptb-2b, callback-widening | #1xxqbrk |
| 11 | self-compare-main, feature-set-link-main, tsgo.go-cache | #7c6b4pq, #hmab710 |
| 12 | per-backend-stops, search-shrink | #z1vjxxd, #b75jjs3 |
| 13 | optional-presence-next, eep-presence | #r3chqza, #eep1m5z |
| 14 | inherit-guards, lint-features | #a898y40, #hmab710, #wj4pmt1 |

Search's port explicitly preserves callback adaptation, establishing a
callback-widening -> search-shrink interaction that must be retained during
extraction. The exceptions merge explicitly preserves earlier optional-read
TypeError construction, namespace readiness and miscompile-2a's absent-message
layout. These repair interactions do not prove that the original source member
requires those earlier features when built alone on main.

## Admission delta and landing rule

Slice 1 changes only native C literal/static initialization and its tests/fixture.
internal/load, internal/lower, internal/ir, the checker gitlink and Go dependency
inputs are identical to the main base. Thus its frontend/lowering admission
function is unchanged for every input. The new fixture is covered separately by
Node and both backends; native C portability is the output property being changed.
The saved admission-delta.json records equality checks and their source revisions.

Integration lands this branch independently on main. Later work starts from main
containing accepted slices. The whole e271a596 cut may still land independently;
no later slice should reapply a member already accepted there.
