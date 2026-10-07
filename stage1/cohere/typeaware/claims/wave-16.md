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
