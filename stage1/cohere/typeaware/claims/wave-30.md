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

Landing refresh: wave-30 was cleanly rebased onto origin/main f8013f0b after main
advanced. All ten wave-30 gates, bridge checks and fifteen filtered Node oracle
fixtures pass again, including twenty-one comparison mutants. No new claims or
implementation edits were made; shared numeric-dispatch and full React analysis
prerequisites remain blocked. See ../WAVE_30_RELAND_REPORT.md and its rebase map.

## Parked React analysis claims

As directed by Ahra, these three claims are now PARKED and count as finished
for the landing-first cap. Their native components and oracle evidence are pushed.

- react-hooks/preserve-manual-memoization: native React high-level IR, single-assignment, reactive scopes and memoization/capture analysis are missing.
- react-hooks/purity: native React high-level IR, single-assignment and render/capture analysis are missing.
- react-hooks/refs: native React high-level IR, single-assignment and capture analysis are missing.

Analysis modules are being ported on #dnv6f2c; JSX support is landing on
area/stage1-lint. No integration branch is pushed by this unit. The partial
components remain available for integration; this status supersedes the earlier
requirement to block subsequent claims on complete React analyses.

## Fifth batch claim

On current main f8013f0b, wave-30's pushed branch is rebased and its ten gates
are green. Refreshed all 529 origin refs and distinguished actual claim documents
from inventory JSON stored under claims/. The first three ranked rules unported
on main/c-library and unclaimed on any origin branch, without a React HIR/SSA
analysis dependency, are:

- react/jsx-fragments
- react/jsx-no-constructed-context-values
- react/jsx-no-undef

These three are claimed before implementation. Snapshot evidence is
../validation-wave-30-fifth/selection.json. No new code precedes this claim push.

Fifth batch progress in 1d023cc6c: three native components, partial rule.json kind
declarations and independent production-Go component tests are implemented. Each
component has a caught comparison mutant and sanitized/JavaScript agreement.
All three full Go TSX controls report while the native parser refuses them. Full
source-rule findings, fixes/suggestions, corpora and timings remain blocked by
shared native JSX AST and numeric handed-node/registration interfaces. These are
partial components, not completed rule ports. See ../WAVE_30_FIFTH_REPORT.md.

Landing and regex refresh: wave-30 was rebased cleanly onto origin/main c01907a70,
with all twelve gates and twenty-four comparison mutants green again. A source
and shared-table audit finds no Go regex matcher for these rules, so no regex
replacement is introduced. Parked analysis claims and partial JSX statuses remain
unchanged; no new rules claimed. See ../WAVE_30_REGEX_LANDING_REPORT.md.

Unicode message progress: bba112af7 extends the constructed-context message
component to scalar Unicode names, using a reproducible JavaScript RegExp literal
from pinned Go Unicode 17.0.0 printability data. Expanded production-Go byte
comparisons and a successful-exit astral escape mutant pass, alongside native
sanitizers and emitted JavaScript. Unpaired surrogates explicitly refuse. Main
remains c01907a70; the announced ab70f38d4 shared harness is not on main. Full
JSX rule statuses and blockers remain unchanged. No new rules claimed.
See ../WAVE_30_UNICODE_REPORT.md and validation-wave-30-unicode/.

Named-kind correction: the three partial JSX rule.json declarations now contain
only named kinds, matching the announced registry's ast.Kind spellings. Numeric
metadata is removed. Their owned check compares names against pinned Go and
rejects a changed name for each descriptor. All component comparisons and four
behavior mutants re-pass. The remaining shared prerequisite is native JSX
extraction and handed-node integration, not a numeric parser API. Earlier numeric
API comments are superseded by Ahra's correction. These descriptors remain
partial component metadata, not registered full rules; no new claim is made.
See ../WAVE_30_KIND_NAMES_REPORT.md.

Landing-first refresh: rebased all 29 own commits cleanly onto origin/main
b8fb957aa, accepting the inherited-static-field compiler fix. All twelve wave-30
gates pass again, including corpus byte agreement, sanitizers, released handles,
25 output-comparison mutants and three metadata name mutants. The newly landed
compiler behavior also passes the uncached Node oracle. No new claims or shared
file edits. Full JSX statuses and parked React analyses remain unchanged.
See ../WAVE_30_STATIC_LANDING_REPORT.md and its 29-commit rebase mapping.

Area landing refresh: rebased onto integrated area/stage1-lint, then included
its runtime-profile advancement at d65a8f931. All twelve own gates pass again,
with the full shared package green on the first area snapshot and focused shared
parity plus runtime/Node checks green on the final snapshot. Native JSX extraction
now succeeds for all six old prerequisite controls, and the owned probes require
it instead of treating it as a failure. Old JSX-parser blocker statements are
superseded. The three JSX rules remain partial: registered RuleContext has no
checker program handle or checker-node mapping for their symbol-resolution facts.
Parser.path does provide source identity. Existing three React analysis claims
remain PARKED on native HIR/SSA/capture analysis. No new rules claimed and no
shared files edited. See ../WAVE_30_AREA_LANDING_REPORT.md for exact proof scope.

Required-input follow-up: main/area remain unchanged and the unit branch is
already rebased and pushed. The formerly skipped shared lint compiler/stage1
comparison now runs with ADAMIC_TYPESCRIPT_SOURCE set to pinned TypeScript 6.0.3.
It passes on 440 files and 21021585 identical Go/Node/emitted-JavaScript/sanitized
native bytes. The clean-running emitted-JavaScript mismatch mutant is caught.
No implementation, harness or skip condition changes, and no new rules claimed.
Other packages' required external checks are not represented as run. See
../WAVE_30_REQUIRED_INPUT_REPORT.md and validation-wave-30-required-input/.

Landing-first predicate refresh: main advanced to c7991b900 and the lint area
includes it at b84a9d931. Rebased all 34 own commits cleanly, retained the new
proven-predicate/relation lowering and record runtime, and reran all twelve own
gates plus the required pinned-input compiler/stage1 comparison. All pass, along
with five uncached Node fixtures, inherited lowering/refusal and record-runtime
mutants, and vet. Shared checker-context gap and partial/PARKED statuses remain
unchanged. No new claims or implementation edits. See
../WAVE_30_PREDICATE_LANDING_REPORT.md and its exact rebase/validation evidence.


Unified landing disposition: all fourteen retained claims are PARKED; the eight
standalone ports lack the registered checker context, and the six partial React/JSX
ports lack checker or HIR/SSA/capture integration. Exact reproducer commands and
TSX sources are in ../parked/wave-30.md. This branch is a parked archive, not
a unified landing candidate. Step 1 is skipped under Ahra's all-blocked exception.
