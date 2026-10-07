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
