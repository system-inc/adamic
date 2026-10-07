# Type-aware wave 16

Base: origin/codex/tsgo-c-library, 0d540f413625f016f20fea39761c7b184f335de6.
Branch: codex/typeaware-wave-16.

Claimed rules, positions after excluding the base's 26 ports from the combined
compiler and repository by-volume ranking linked by VOLUME_REPORT.md, descending
combined count with lexical ties:

| Position | Rule | Compiler | Repository |
| --- | --- | ---: | ---: |
| 46 | no-import-assign | 1 | 0 |
| 47 | prefer-exponentiation-operator | 0 | 1 |
| 48 | use-isnan | 0 | 1 |

All origin heads were fetched before selection. Searches of origin branches'
Adamic sources and claim files found no existing port or claim for these names.
No rules were skipped. Selection uses validation-volume/compiler-all.counts and
validation-volume/repository-all.counts and excludes the sixteen oracle_volume
subjects and ten validation-coverage/selection.json subjects. No lint code was
written before this claim commit.

## Follow-up selection

Fresh all-head fetch on October 7, 2026: bridge tip
5afbdb83da2ed7ad9815657cd3f6ececd5294bf6; main tip
ef3d907ecdc4c771b016f7d9c52372def057a340.
The first three remaining checker-dependent rules in descending combined volume,
with lexical ties, after excluding ports on those two heads and claims on every
origin head are:

- nexus/correctness-no-global-listener-target-assertion
- nexus/correctness-no-leaked-number-render
- nexus/correctness-no-mock-on-module-namespace

Each has compiler 0 and repository 0. The scan covered all 197 count-table rules
and 32 Markdown claim files across origin heads; 79 candidates remained.
This update is pushed before implementation begins.

## Third set, after the follow-up validation push

The previously claimed follow-up implementations and final overlay validation
were pushed in 828d4a57. Shared dispatch integration remains documented, with
its patch supplied rather than editing shared source.

A fresh all-head fetch selected the first three remaining candidates:

- nexus/security-no-interpolated-shell-command
- nexus/security-no-interpolated-sql-string
- no-alert

All three have compiler 0 / repository 0. Selection covered 197 ranked rules,
33 Markdown claim documents on all origin heads, and ports on main
(ef3d907ecdc4c771b016f7d9c52372def057a340) and the bridge branch
(5afbdb83da2ed7ad9815657cd3f6ececd5294bf6). 70 candidates remained after
exclusions. Additional searches of both heads' entire stage1/cohere trees found
only inventory/count mentions for these three, not ports. No rules were skipped.
This claim update is pushed before implementation.

## Fourth set

All prior ports and validation are pushed at b5c30a08. A fresh all-head fetch
on October 7, 2026 found 55 remaining candidates among 197 ranked rules after
excluding ports on main and the bridge branch and 117 names claimed in 33
Markdown documents across all origin heads. The first three are:

- no-new-func
- no-new-native-nonconstructor
- no-new-wrappers

Each has compiler 0 and repository 0. Searches of both integration heads found
only inventory and count mentions, with no implementation. This claim is pushed
before any implementation of these rules.

## Fifth set

The prior twelve ports and validation are pushed at 363fde32. A fresh all-head
fetch found 46 remaining candidates among 197 ranked rules, after excluding
integration-head ports and 126 names in 33 origin claim documents. The next
three, each with compiler 0 and repository 0, are:

- no-throw-literal
- no-useless-backreference
- prefer-arrow-callback

Whole stage1/cohere searches on main and the bridge found inventory/count
mentions only. This claim is pushed before implementation.

## Sixth set

Prior fifteen ports and validation are pushed at a278e5e7. Fresh all-head fetch
found 30 remaining rules among 197 ranked names, excluding integration-head
ports and 142 claimed names in 33 origin Markdown claim documents. The first
three, each with compiler 0 and repository 0, are:

- react-hooks/set-state-in-effect
- react-hooks/set-state-in-render
- react-hooks/static-components

Main is e011f8f60899586d6373a5ccb07335ad82cfbf3c; bridge is
5afbdb83da2ed7ad9815657cd3f6ececd5294bf6. Whole stage1/cohere searches found
only inventory/count references on those heads. This claim is pushed before code.

Sixth-set status: blocked before native implementation by unavailable native
React HIR lowering, SSA, capture mapping and graph analyses. These names remain
claimed, not ported. See WAVE16_SIXTH_REPORT.md for exact dependencies and the
production Go reference test results. No seventh set is claimed.

Prepared-HIR cores for these same three names now exist on wave 21 at bc6750a6,
with source-to-HIR/SSA integration still absent. These source claims remain
incomplete, not completed ports. Independent reference runs also reproduce
non-deterministic static-components creation-site message bytes; see
WAVE16_REACT_REFERENCE_REPORT.md. No duplicate core implementation or new
reservation was added in this continuation.

## React claims parked under the analysis-module instruction

The following retained claims are now **parked**, and count as finished for
landing-first scheduling under Ahra's explicit instruction. They remain partial
source work, not completed native source ports:

| Rule | Parked blocker |
| --- | --- |
| react-hooks/set-state-in-effect | Native source-to-HIR lowering, single-assignment construction, capture translation, memo erasure/inlining and control dominators |
| react-hooks/set-state-in-render | Native source-to-HIR lowering, single-assignment construction, capture translation and exact compilation-unit/memo gates |
| react-hooks/static-components | Native source-to-HIR lowering, single-assignment phis and source/capture correspondence; independently observed Go creator-message nondeterminism |

Cohere's analysis modules are being ported on #dnv6f2c; JSX support is landing
on area/stage1-lint. Existing prerequisite and independent Go reference evidence
is pushed, with native prepared-HIR validators available on wave 21. No Go
lowering/verdict shortcut or empty native finding implementation was added here.
The completed fifteen ports were rebased onto main f8013f0ba and re-greened
before push 2bd4efb74: oracle 425.666s, bridge 66.168s. Main is still f8013f0ba.
This status update changes no native implementation or harness.
