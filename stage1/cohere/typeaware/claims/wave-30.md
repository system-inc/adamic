# Type-aware lint wave 30

Base: origin/codex/tsgo-c-library, 0d540f413625f016f20fea39761c7b184f335de6.

Assigned positions in the combined compiler plus repository ranking linked by
VOLUME_REPORT.md, descending count with lexical ties, excluding the 26 existing
ports before counting positions:

| Position | Rule | Combined findings |
| --- | --- | ---: |
| 88 | nexus/concurrency-no-lost-update | 0 |
| 89 | nexus/consistency-no-iso-string-date-cut | 0 |
| 90 | nexus/correctness-no-callback-in-parse-try | 0 |

The volume counter has 197 entries. Twenty-five of the existing ports occur in
that counter; method-signature-style is not checker-dependent in the registry
and does not occur there. This leaves 172 ranked candidates.

Fetched all origin heads and checked stage1 source paths and claim documents
before this claim. No conflicting port or claim was found for these rules.
No implementation was written before this claim commit.

Completion status: the ISO-date-cut and callback-in-parse-try rules are implemented.
The lost-update rule is unimplemented and its claim is released for reassignment.
See ../WAVE_30_REPORT.md for evidence and limits.

## Continuation claim

Fetched every origin head after pushing 333b5ccd. Selected the first three in
VOLUME_REPORT.md's combined compiler/repository checker ranking that are neither
ported on origin/codex/tsgo-c-library or origin/main nor named in any origin
branch's stage1/cohere/typeaware/claims files. The ranking has 197 checker rules;
25 of the original 26 ports are checker-dependent, and claim files mention 92
ranked rules. All higher-ranked remaining candidates are claimed.

- nexus/correctness-no-collection-misuse (combined volume 0)
- nexus/correctness-no-discarded-outcome (combined volume 0)
- nexus/correctness-no-discarded-pure-result (combined volume 0)

These three are claimed for this continuation. No implementation precedes this
claim commit and push. The earlier released lost-update claim remains released.

Continuation completion: all three rules above are implemented in a74cf501,
with independent Go byte agreement, per-rule mutants, sanitizer runs and released
handles verified. Evidence is in ../WAVE_30_NEXT_REPORT.md and validation-wave-30-next.
No further rules were claimed after Ahra's correction.

## Third batch claim

Fetched all origin heads after verifying b64a21b3 was pushed. The first three
remaining rules in VOLUME_REPORT.md's combined checker ranking, excluding ports
on origin/main and origin/codex/tsgo-c-library and claims on every origin branch:

- nexus/correctness-no-process-exit-after-output (combined volume 0)
- nexus/correctness-no-uncleared-race-timeout (combined volume 0)
- nexus/correctness-require-blocking-standard-streams (combined volume 0)

The ranking contains 197 checker rules, 25 base ports and 96 distinct claim mentions at
this selection. No implementation precedes this claim commit and push.

Third batch completion: all three rules are implemented and tested. The timer
port remains in 80e56233; process-exit and blocking-streams are completed in
182d776e. Independent Go byte agreement, per-rule comparison mutants, sanitizers,
released handles and timings are documented in ../WAVE_30_PROCESS_REPORT.md.
No shared harness or registration generator was edited. No subsequent batch
was claimed before these implementations and their evidence were pushed.

## Fourth batch claim

All prior retained claims are complete and pushed in 182d776e and 2ccc0f30.
Fetched every origin head afterward. The first three available rules in the
197-rule combined ranking, excluding 25 ports on main/c-library and 139 distinct
rule mentions in claim files on 359 origin refs, are:

- react-hooks/preserve-manual-memoization (combined volume 0)
- react-hooks/purity (combined volume 0)
- react-hooks/refs (combined volume 0)

These rules are claimed before implementation. The precise branch/blob snapshot
is validation-wave-30-fourth/selection.json. Released mentions in other claim
files are conservatively excluded, as requested by the any-branch claim check.

Fourth batch status: retained and unported. A private native/Go probe confirms
that the current native parser rejects valid JSX or parses it as a type assertion.
The native React HIR/SSA and memoization pipeline are also absent. No shared files
were edited to work around these blockers, and no later batch was claimed.
See ../WAVE_30_FOURTH_REPORT.md for exact controls, results and missing coverage.

Fourth batch progress in 1f2e7bbc: native ref-lattice operations, purity-property
propagation and manual-memo scope verdicts now match production Go helpers,
sanitized native and emitted JavaScript. Each component has a caught decision
mutant. These are partial components; the three complete analyses remain
unported. JSX support exists on codex/stage1-jsx-lint but is not integrated here;
native React HIR/SSA and the reactive memoization pipeline are still required.
The refreshed wave-04 branch has a later overlapping claim, recorded in
../WAVE_30_REACT_COMPONENTS_REPORT.md. No subsequent batch was claimed.

Landing-first update: this unit's only pushed branch was cleanly rebased onto
origin/main e8ba3d5d. All implemented wave-30 gates, inherited corpus comparisons,
sanitisers and released handles re-pass; the complete React analyses remain
partial and no new rules were claimed. See ../WAVE_30_LANDING_REPORT.md and
validation-wave-30-landing/rebase.json for current commits and failure proofs.

Numeric listener update: dc40511f declares pinned numeric SyntaxKind lists for all
eight implemented behavior ports. All nine existing oracle gates and eight new
listener-comparison mutants pass on unchanged origin/main e8ba3d5d. Numeric
ParseNode/handed-node driver APIs and complete React analysis prerequisites remain
blocked shared work; no new rules are claimed. See ../WAVE_30_LISTENERS_REPORT.md.
