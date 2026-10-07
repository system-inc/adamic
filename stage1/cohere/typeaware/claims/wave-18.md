# Type-aware wave 18

Branch: codex/typeaware-wave-18
Base: origin/codex/tsgo-c-library, 0d540f413625f016f20fea39761c7b184f335de6

Claimed rules, positions 52, 53 and 54 among the unported checker-dependent
rules in VOLUME_REPORT.md's validation-volume counts:

1. @next/next/no-title-in-document-head
2. @typescript-eslint/await-thenable
3. @typescript-eslint/class-literal-property-style

Selection combines compiler-all.counts and repository-all.counts, excludes the
six original, ten volume and ten coverage ports, and sorts by descending combined
count then lexical rule name. All three selected rules have zero corpus findings.
The inventory ranking in validation-coverage is a different population and is
not used for this wave's positions.

Fetched every origin branch before claiming. No matching claim or named native
port file was found under stage1/cohere on those refs. No rule was skipped.

Implementation status: await-thenable and class-literal-property-style default
ports are complete in adea6fbe. The Next rule remains blocked on absent native
JSX parsing; WAVE_18_REPORT.md records a Go-positive native-refusal measurement.

## Next three rules

Following the October 7 continuation request, this branch additionally claims:

1. nexus/correctness-no-collection-misuse
2. nexus/correctness-no-discarded-outcome
3. nexus/correctness-no-discarded-pure-result

These are the first three remaining in VOLUME_REPORT.md's combined by-volume
ranking after excluding ports on fetched origin/codex/tsgo-c-library (5afbdb83)
and origin/main (ef3d907e), plus every claim on all 320 origin refs. Ninety
ranked rules are claimed; 25 ranked rules are already ported (the 26th base
port, method-signature-style, is outside this checker-dependent ranking).
All three new selections have zero compiler and repository findings. No
matching existing port or claim was found. This update is pushed before code.
