# Type-aware wave 25

Branch: codex/typeaware-wave-25
Base: origin/codex/tsgo-c-library, 0d540f413625f016f20fea39761c7b184f335de6.

Selection combines validation-volume/compiler-all.counts and
validation-volume/repository-all.counts, descending total then lexical rule name,
excluding the 26 rules documented in README.md and COVERAGE_REPORT.md.
The method-signature-style baseline port is absent from the 197 checker registry
rows and therefore does not remove a row from this ranking.

| Remaining position | Rule | Status |
| --- | --- | --- |
| 73 | accessor-pairs | Skipped: already ported on origin/codex/stage1-lint-batch4 at stage1/cohere/lint/grouped_accessor_pairs.ts |
| 74 | base/correctness-require-graphql-nullable-parity | Claimed by wave 25 |
| 75 | base/correctness-require-matching-inject-type | Claimed by wave 25 |

All three have zero findings on both recorded populations. No replacement is
selected for the skipped rule. All origin branch tips were fetched and distinct
stage1/cohere source and claim blobs checked before this claim.

## Next three, October 7 continuation

Checked origin/codex/tsgo-c-library at 5afbdb83da2ed7ad9815657cd3f6ececd5294bf6
and origin/main at ef3d907ecdc4c771b016f7d9c52372def057a340.
Fetched all 320 origin refs; inspected the 30 distinct claim blobs.
These are the first three by descending combined volume, lexical ties,
excluding ports on those two branches and every rule named in any origin claim.

- nexus/correctness-no-collection-misuse (combined volume 0)
- nexus/correctness-no-discarded-outcome (combined volume 0)
- nexus/correctness-no-discarded-pure-result (combined volume 0)

Claim commit b9a19759 was pushed before implementation.
These three are completed in the continuation report; no further rules are claimed.

## Third batch, October 7

The previous three ports, tests and evidence were pushed in 8bd28ec4.
Fetched all 324 origin refs and inspected textual claim records on every branch.
The first three remaining checker-dependent rules by combined recorded volume,
lexical ties, excluding ports on origin/main and origin/codex/tsgo-c-library,
are claimed here:

- nexus/correctness-no-process-exit-after-output (combined volume 0)
- nexus/correctness-no-uncleared-race-timeout (combined volume 0)
- nexus/correctness-require-blocking-standard-streams (combined volume 0)

The preceding listener assertion, leaked-number-render and mock-on-module-
namespace rules are reserved by other origin claims. The baseline branch tips
remain 5afbdb83da2ed7ad9815657cd3f6ececd5294bf6 and
ef3d907ecdc4c771b016f7d9c52372def057a340, respectively.
Claim commit b22f8858 was pushed successfully before implementation.
The three native .a ports, independent Go comparisons, mutants, sanitizer and
released-handle checks are complete in WAVE_25_THIRD_REPORT.md. No further
rules are claimed in this batch.

## Fourth batch, October 7

The preceding process/timeout ports and evidence were tested and pushed in
0df6cc4c and 0a14418e before this selection. Fetched all 348 origin refs and
inspected 77 distinct textual claim records. The baseline tips remain
5afbdb83da2ed7ad9815657cd3f6ececd5294bf6 (tsgo-c-library) and
ef3d907ecdc4c771b016f7d9c52372def057a340 (main).
Baseline TypeScript diagnostic names are normalized to their registry prefix.
The first three remaining checker-dependent rules by descending combined
volume and lexical ties are reserved here:

- no-throw-literal (combined volume 0)
- no-useless-backreference (combined volume 0)
- prefer-arrow-callback (combined volume 0)

Implementation begins only after this claim commit is pushed successfully.

Claim abdad124 was pushed before implementation. These three native .a ports
and their byte comparisons, mutants, sanitizers and released-handle checks are
complete in WAVE_25_FOURTH_REPORT.md. No additional rules are claimed.

## Fifth batch, October 7

The preceding three rules are complete, tested and pushed in 056799e9.
An explicit fetch of every origin head refreshed 394 refs; all 77 distinct
textual claim blobs were checked. The first three unported, unclaimed rules
in the combined-volume ranking are reserved here:

- react-hooks/unsupported-syntax (combined volume 0)
- react-hooks/use-memo (combined volume 0)
- react/boolean-prop-naming (combined volume 0)

Baseline main: e011f8f60899586d6373a5ccb07335ad82cfbf3c
Baseline tsgo-c-library: 5afbdb83da2ed7ad9815657cd3f6ececd5294bf6
Implementation begins only after this claim is pushed.

Fifth-batch status: unsupported-syntax and use-memo are ported and verified.
Boolean-prop-naming remains partial: arbitrary regex, cross-file annotations
and typed memo/forwardRef wrapper paths refuse explicitly. The tested partial
implementation and evidence are committed together; this claim remains reserved.
No additional rules are claimed. See WAVE_25_FIFTH_REPORT.md.

## Landing unit, October 7

Rebased all work onto origin/main at e8ba3d5d81de4d3773c723914fccd4c76248b965.
All five wave suites and the carried ten-rule bridge dependency are rebuilt
and green against independent Go bytes, mutants, sanitizers and released
handles. The report and evidence are in WAVE_25_LANDING_REPORT.md.
Boolean-prop-naming remains partial and reserved; no new rules are claimed.

## Third landing, October 7

Rebased onto origin/main f8013f0baac41ddc340d76f83bddde38536a8f07 and
revalidated all five wave suites, the inherited checker dependency, bridge,
released handles, sanitizers, uncached filtered Node oracle and vet.
Evidence: validation-wave-25-landing-third and WAVE_25_LANDING_REPORT.md.
Boolean-prop-naming remains partial and reserved. No additional claims.

## Regex instruction, October 7

Default boolean-prop-naming matching now uses a JS RegExp literal, with fresh
Go byte agreement, a regex mutant, sanitizer and released-handle validation.
Configured patterns explicitly refuse pending native dynamic RegExp lowering;
the new RegExp(pattern, 'u') compiler probe demonstrates that exact blocker.
Imported props and typed wrappers remain unfinished local paths. This claim
is partial, not parked under the IR/SSA/capture exception. No new rules claimed.
