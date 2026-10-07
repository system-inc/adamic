# Type-aware wave 26

Base: 0d540f413625f016f20fea39761c7b184f335de6.

The VOLUME_REPORT.md linked all-family compiler and repository count tables are
sorted by combined volume descending, with lexical ties, excluding the 26 rules
already ported on the base.

| Remaining position | Rule | Combined findings |
| --- | --- | ---: |
| 76 | base/correctness-require-matching-operation-context | 0 |
| 77 | base/correctness-require-matching-provider-return | 0 |
| 78 | base/correctness-require-optional-relation | 0 |

Reserved for codex/typeaware-wave-26. No claim or named port for these rules was
found across fetched origin branches before this claim. Zero corpus findings
require positive generated controls and independent byte-oracle mutants.

## Next batch, October 7

Fetched all 321 origin remote refs. Excluded the 26 ports on
origin/codex/tsgo-c-library (5afbdb83da2ed7ad9815657cd3f6ececd5294bf6),
ports on origin/main (ef3d907ecdc4c771b016f7d9c52372def057a340), and
93 ranked rule names claimed in Markdown claim files on any origin branch.
These are the first three remaining by combined volume descending and lexical ties:

| Rule | Combined findings |
| --- | ---: |
| nexus/correctness-no-global-listener-target-assertion | 0 |
| nexus/correctness-no-leaked-number-render | 0 |
| nexus/correctness-no-mock-on-module-namespace | 0 |

Reserved for this same branch before implementation. The previous three remain
claimed and completed. Positive controls will exercise these zero-volume rules.

## Third batch, October 7

After completing and pushing both prior batches (a21c17fb), fetched all
330 origin refs. Base remains 5afbdb83 and main ef3d907e. Excluding ranked
base/main ports and 108 rule names mentioned in origin Markdown claims,
the first three available by combined volume and lexical ties are:

| Rule | Combined findings |
| --- | ---: |
| no-eval | 0 |
| no-extend-native | 0 |
| no-func-assign | 0 |

Reserved on codex/typeaware-wave-26 before implementation. No further
rules are reserved in this batch.

## Fourth batch, October 7

All nine previous claims are implemented, compared, sanitized and pushed at
a385f1e4. Fetched all 341 origin refs. Base is 5afbdb83 and main ef3d907e.
Excluding existing ports and 117 ranked names mentioned in origin Markdown
claims leaves these first three available, each with zero combined findings:

- no-new-func
- no-new-native-nonconstructor
- no-new-wrappers

Reserved on codex/typeaware-wave-26 before implementation.

## Fifth batch, October 7

All twelve previous claims are implemented, tested and pushed at 135bbbc7.
Fetched 356 origin refs. Base remains 5afbdb83 and main ef3d907e. Excluding
base/main ports and 132 ranked names mentioned in origin Markdown claims leaves
40 available rules. The first three by combined volume and lexical ties are:

- prefer-promise-reject-errors
- prefer-regex-literals
- prefer-rest-params

Each has zero combined corpus findings. Reserved on codex/typeaware-wave-26
before implementation.

## Sixth batch, October 7

All fifteen preceding claims are complete and pushed at 8180b758.
Fetched 389 origin refs. Base is 5afbdb83; main advanced to e011f8f6.
Excluding base/main ports and 142 ranked names mentioned in origin claims leaves
30 available rules. The first three by combined volume and lexical ties are:

- react-hooks/set-state-in-effect
- react-hooks/set-state-in-render
- react-hooks/static-components

Each has zero combined corpus findings. Reserved on codex/typeaware-wave-26
before implementation.

### Sixth batch prerequisite

These three remain claimed and are not implemented or validated. The current
bridge has AST and checker-fact questions but no HIR/SSA graph question. All three
production rules consume Cohere's HIR, not merely source syntax:

- set-state-in-effect needs capture translation, ref-derived data/control flow
  and the manual-memoization-erased graph.
- set-state-in-render needs unconditional blocks and nested setter-call tracking.
- static-components needs SSA phis and instruction-order dynamic-value tracking.

Cohere's implementation is an internal Go package under
`cohere/internal/lint/ecmascript/high_level_intermediate_representation`. The
Adamic module cannot directly import it under Go's internal-package restriction.
An isolated raw-graph implementation and serializer, or a native equivalent,
are still required; returning Go lint verdicts would not establish an Adamic port.
This is not the shared .a-loading or suggestion-serialization gap. No shared
registration, generator, harness or compiler files were changed for this batch.
No new rule comparison, mutant, sanitizer or timing result is claimed.

The previous fifteen completed ports and their evidence remain at 8180b758.
Selection evidence for this batch is in wave-26-sixth-selection.json alongside
this claim. Work stopped here under Ahra's instruction to report other blockers.

### Origin recheck, October 7

Fetched all heads again: 417 origin refs, 197 ranked names, 25 baseline/main
ranked ports, 148 ranked names mentioned in claim files and 24 available names.
No additional rules were claimed because the sixth batch is unfinished. The
current audit is saved in wave-26-refresh.json.

Origin wave 06 now contains reporting-only partial implementations of the same
three rules and independently documents the missing HIR/SSA/capture/memoization
and control-dominance substrate. Other origin claim files overlap too. Our claim
e758e57f was committed at 02:30:28 UTC; wave 06 claim 3c3f5f4b at 02:30:37 UTC.
These overlapping reservations are recorded without asserting exclusive ownership
or duplicating another worker's partial implementation.

A scratch Go file importing Cohere's internal HIR was compiled using
`go build -mod=readonly`. It exited 1: the current Adamic module has no module
providing that package. Exact output is in wave-26-hir-import.log. This probe
confirms that direct import is unavailable with the current module configuration;
it does not prove that an isolated raw-graph implementation is impossible.
`go list` of that scratch file exited 0, which did not validate compilation.
No module configuration or shared files were changed. nproc remains 5.

No new native source analysis, corpus comparison, sanitizer result, mutant or
lint timing was obtained. All fifteen previously completed ports remain pushed
at 8180b758. This recheck leaves the same three claims unfinished.

### Sixth batch reporting progress

The branch remains rebased onto origin/main e8ba3d5d and the previous fifteen
ports remain green at f4bd9c34. Reporting portions of these three claims are
now implemented in wave_26_react_reporting/, with native, source Node, emitted
JavaScript, sanitizer and diagnostic-only mutant agreement. Their source
analyses are still unfinished and explicitly refuse with NotYet. This is not
a completion claim. See wave_26_react_reporting/REPORT.md for the precise
reporting-only scope and missing HIR engineering work. No further rules claimed.

### Sixth batch native source analyses

The three claimed React rules now have standalone native .a source validators
in wave_26_react_reporting/*/rule.a and their own suite entry points. An isolated
raw HIR question provides graph and type facts; all lint decisions are native.
129 controls and 69 findings, plus 77 compiler and 287 repository files per rule,
match production Go normally and under sanitizers. Original-source Node agrees.
Three analysis mutants, raw SSA and released-handle mutants are caught. The
previous fifteen ports re-green on main e8ba3d5d. See
wave_26_react_reporting/SOURCE_REPORT.md for commands, complete mutation logs,
timings and limits. Checker-linked JavaScript emission still refuses in the
shared CLI; shared registration remains with its owner. No new claims taken.

### Current-main landing recheck

Main advanced to f8013f0b during the final refresh. The branch was rebased
without conflicts and all eighteen rule suites, their corpus/sanitizer bytes
and mutants, raw HIR, source Node, bridge and filtered Node regressions passed
again. landing_validation preserves this run; SOURCE_REPORT.md has updated
native/Go timings. Checker-linked JavaScript emission and shared registration
remain the documented integration gaps. No additional rules claimed.

## Next batch after landing, October 7

Previous eighteen ports pushed as fe21e4234b39cfb98a0fab643e45c42f02b5e9bf,
rebased onto origin/main f8013f0baac41ddc340d76f83bddde38536a8f07 and
re-greened against their independent production Go oracles.

Claim the first three available by combined volume and lexical tie order:

- `react/button-has-type` (0 compiler, 0 repository)
- `react/checked-requires-onchange-or-readonly` (0 compiler, 0 repository)
- `react/display-name` (0 compiler, 0 repository)

Fresh origin audit: 417 refs, 25 ranked rules ported on main or the bridge
baseline, 148 claimed names, 24 available. Full ref and claim evidence is
wave-26-next-refresh.json. Claim pushed before implementation.

Next batch complete natively: button-has-type, checked-requires-onchange-or-readonly
and display-name. Source commit c06def68, rebased onto c01907a7. All twenty-one
ports re-greened on that main base. New controls: 345 files, 163 findings and
58,921 exact Go/native/sanitized bytes; both frozen corpora match. Source Node,
three rule mutants, both question mutants, released-handle mutants and the full
bridge gate passed. Shared emitted-JavaScript checker linking and registration
remain integration gaps. Full evidence and timing observations are in
../wave_26_react_attributes/README.md. No further rules claimed.

## No further unclaimed rules, October 7

The tested twenty-one-rule tip eff34e791ef1ff25ea4ea32a1533a79d3bfd3f77
is pushed to this branch and contains current origin/main c01907a7. The
existing current-main oracle, sanitizer, released-handle and mutant evidence
is recorded in ../wave_26_react_attributes/README.md. Main has not advanced
since that validation. No implementation or test harness changed in this audit.

The checkout fetch configuration updates only main. Explicitly fetched
`+refs/heads/*:refs/remotes/origin/*` before scanning all 541 origin refs.
The original 197-rule by-volume population has 25 baseline/main ports and
172 names in origin Markdown claims; their union contains all 197 rules.
There are zero unclaimed candidates. No additional rules were claimed.
The complete ref SHAs, claim blobs, names and ranking are preserved in
wave-26-exhausted-audit.json.gz. Work stops at the requested exhaustion gate.

## Landing re-green on b8fb957a, October 7

All twenty-one ports were rebased onto main b8fb957a and their complete
oracles, rule/question mutants, both corpora, sanitizers and released-handle
checks passed again. See ../wave_26_react_attributes/LANDING_B8_REPORT.md
for the exact command ledger, mutant observations and native/Go costs.
The final all-branch fetch scanned 554 origin refs and again found zero
unclaimed ranked rules. No new claim was made.

## Area landing re-green, October 7

Rebased onto area/stage1-lint 7481e032, containing current main 39638d9e
and merged harness 41eb6eab2. All twenty-one rule oracles, mutants, both
corpora, sanitizers and released-handle checks passed again, along with
registry and merged harness checks. Exact commands, outputs and costs are
in ../wave_26_react_attributes/LANDING_AREA_REPORT.md. A final explicit
all-head fetch with pruning found 433 current origin refs and zero
unclaimed ranked rules. No further rules were claimed.

## Current-main area landing re-green, October 7

Rebased onto area b84a9d93 containing current main c7991b90. All twenty-one
ports were re-greened after main advanced during the earlier run. Twenty-five
current-base gate commands passed, including new lowering/record/oracle checks,
with zero selected test skips. See ../wave_26_react_attributes/LANDING_CURRENT_REPORT.md
for commands, outputs, every mutant, the earlier disk retry and native/Go costs.
The final pruned audit scanned 463 current origin refs and found no unclaimed
ranked rules. No additional claim was made.

Current-main landing refresh: rebased onto area d3a37422, containing main b6b1538b0. All 27 owned commands passed, including decoded-options and new typeof oracle/mutants. Twenty-one ports remain complete under their private runners. Final all-head audit: 505 origin refs, zero available rules. Evidence: ../wave_26_react_attributes/LANDING_D3_REPORT.md. No new claim.
