# Type-aware lint wave 14

Branch: `codex/typeaware-wave-14`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Claimed rules, positions 40, 41 and 42 after excluding the 26 existing ports:

1. `@typescript-eslint/no-array-delete`
2. `@typescript-eslint/no-base-to-string`
3. `@typescript-eslint/no-extraneous-class`

Selection uses the combined compiler and repository checker-dependent counts
linked by VOLUME_REPORT.md in validation-volume, descending total with lexical
rule-name ties. Each selected rule has compiler 1, repository 0, total 1.
All fetched origin branch trees were checked for matching claims and rule-named
port files before this claim. No selected rule was found claimed or ported.
No rules were skipped.

## Continuation on October 7, 2026

After pushing c043d14a, fetched all 320 origin refs without recursing submodules.
The first three remaining checker-dependent rules in VOLUME_REPORT.md's combined
compiler and repository counts, descending volume with lexical ties, are:

1. `nexus/correctness-no-global-listener-target-assertion`
2. `nexus/correctness-no-leaked-number-render`
3. `nexus/correctness-no-mock-on-module-namespace`

Each has zero compiler and repository findings. Excluded the original 26 ports
on origin/main and origin/codex/tsgo-c-library and every rule named in Markdown
under stage1/cohere/typeaware/claims on every fetched origin branch, including
incomplete reservations. The earlier collection and discarded-result candidates
are already claimed by other continuations. These three have no matching native
port on either base branch and no matching claim. This update must be pushed
before implementation. Positive controls and comparison-only mutants are required.

Continuation status after Ahra's correction: the listener assertion and module
namespace rules are implemented and validated. The leaked-number-render rule
remains unimplemented because the native parser cannot parse JSX child nodes.
A Go-positive .a witness, parsed as virtual TSX by the independent Go harness,
reports bytes 49:54; the native parser refuses with panic 70 at byte 54.
This is a partial continuation, not three completed ports. Implementation stops
at this blocker, with no further rules claimed and no shared harness or generator
edits. The six existing checker-dispatch lines were added before the correction.
See ../WAVE_14_NEXT_REPORT.md for commands, artifacts and limits.

## Leaked-render judgments completed

On the next resume, the leaked-render rule's native judgments and JSX-context
visitor were implemented in no_leaked_number_render.a and integrated into the
owned wave_14_next.a driver. Twenty-two isolated controls supply only synthetic
JSX parent contexts to both native and the unchanged production Go listener.
They match on 18 complete findings and 7704 bytes, including empty fixes and
suggestions; attributes and ordinary source stay silent. Sanitizers pass and
an exit-0, empty-stderr judgment mutant differs at byte 51.
Real JSX source remains blocked by the shared native parser, independently of
.a module support in the shared lint harness. No shared parser, generator or
harness file was changed. This is not a claim of end-to-end JSX parity.
No additional rules are claimed while that original completion gate is blocked.
See ../WAVE_14_RENDER_REPORT.md for the evidence and exact scope.

## Third batch, October 7, 2026

The preceding six rule implementations and their available comparisons are pushed
at 5bd31ee5. Real JSX remains a shared-parser gap, documented in
WAVE_14_RENDER_REPORT.md. The latest instruction permits moving on after pushing
the available implementation and reporting that gap.

Fetched all 335 origin refs. After excluding the 26 ports on origin/main and
origin/codex/tsgo-c-library and every reservation under typeaware/claims on every
origin branch, the first three remaining rules in the combined by-volume ranking
are claimed here:

1. `no-invalid-regexp`
2. `no-label-var`
3. `no-misleading-character-class`

All three have zero compiler and repository findings. Neither base branch has
rule-named native files for them. There are 58 remaining unclaimed ranking rows
before this reservation. This claim is pushed before implementation. New native
files use .a; shared parser, registration generator and harness remain untouched.

Third-batch status: native scoped-label judgments, constructor selection and
flag validation, and regex-literal class judgments are implemented. Fifty-one
supported controls produce 49 complete byte-identical findings under all three
sanitizers, with a mutant per rule plus scope-meaning and Unicode-quote mutants.
Both frozen corpora also agree. This remains partial: undefined labels are a
shared parser gap; string-pattern validation, constructor reference/source
mapping and lone-surrogate flags are explicit refused dependency paths. These
are demonstrated by independent Go-positive witnesses. No further batch is
claimed. See ../WAVE_14_THIRD_REPORT.md.

## Surrogate flags completed on resume

Fetched origin and removed the lone-surrogate flags refusal in the owned native
rule. Three additional .a controls bring the supported comparison to 54 inputs,
60 findings and 19,352 identical bytes in normal and sanitizer builds. Six
compiled comparison-only mutants and the released-handle mutant are caught.
Both frozen corpora still agree. The rule suite passes in 99.844 s.
Undefined labels, native pattern validation and constructor tracking/source
mapping remain explicit Go-positive boundaries; real JSX remains blocked in the
preceding batch. No further rules are claimed while these incomplete paths
remain. See ../WAVE_14_SURROGATE_REPORT.md for the current commands, evidence and
native-versus-Go timing.

## Constructor class judgments and tracking completed on resume

The owned class rule now ports constructor checking, native global reference
tracking and constant evaluation, with byte-based Go-compatible cooked source
mapping in three new .a helpers. The previous constructor and alias refusal
witnesses now compare normally. Sixty-two constructor/reference controls match
on 58 findings and 24,379 bytes under normal and sanitizer builds. Eighty-three
upstream default constructor cases also match on 89 findings and 38,351 bytes;
one legacy-octal case is excluded by the independent strict-module Go parser.
The original 54 controls and both frozen corpora still agree. Thirteen compiled
comparison-only mutants and the released-handle mutant are caught; the final
owned suite passes in 174.233 s. No shared generator/harness/parser or bridge
code changed. Native pattern validation, undefined labels and real JSX remain
incomplete, so no further batch is claimed. See ../WAVE_14_CONSTRUCTOR_REPORT.md.

## Landing cap, October 7, 2026

Rebased the existing branch onto origin/main e8ba3d5d, yielding 75caf8c1 before
the landing adjustments. Two owned rules now construct nominal Repair values
required by current main. All four owned oracle suites pass again, with nineteen
semantic comparison mutants, released-handle checks and sanitizer corpus checks.
The complete bridge packages and filtered uncached Node oracle also pass.
No new rules are claimed. Native pattern validation, undefined labels and real
JSX remain incomplete. See ../WAVE_14_LANDING_REPORT.md for current evidence.

## Numeric listener declarations, October 7, 2026

All nine owned rule classes now declare numeric syntaxKinds arrays matching
production Go's actual listener maps. The owned verifier catches a wrong-kind
declaration mutant for each rule. All four native oracle suites pass again with
nineteen semantic comparison mutants, sanitizer corpus checks and released-handle
checks. Current shared ParseNode has only kind: string, so numeric handed-node
dispatch remains blocked on the shared parser/driver work; no shared file changed.
Native pattern validation, undefined labels and real JSX remain incomplete.
No additional rules are claimed. See ../WAVE_14_LISTENERS_REPORT.md.

## Native pattern slice and refreshed landing gate

regexp_pattern_validation.a now supplies native ASCII literal/flat-class
validation and exact Go rewrite errors. The former '[' refusal reports normally;
broader grammar such as '(' still explicitly refuses. The 115 new controls match
36 findings under sanitizers, with a comparison-only range mutant. A further
741-constructor ASCII/range matrix matches 168 findings byte for byte.
After main advanced, this branch rebased onto f8013f0b at tested head 852560ba.
All owned suites pass again in 505.910 s, along with bridge, filtered uncached
Node and vet. Remaining pattern grammar, shared numeric dispatch, undefined
labels and real JSX are incomplete. No new rules are claimed.
See ../WAVE_14_PATTERN_REPORT.md for current commands, mutants and timing.

## Effective regex contract dependency

The branch remains based on current main f8013f0b. No new rules are claimed.
no-invalid-regexp is blocked under the new no-custom-matcher requirement on
shared runtime RegExp construction and recoverable Go-compatible validation
errors. A minimal new RegExp(pattern, 'u') native probe refuses nonconstant
patterns; Node accepts the same program. Existing pattern validation is partial
and is not claimed complete or compliant with the new contract. No further
custom pattern validation is added. The exact compiler refusal, Go/Node error
difference and declaration mutant checks are retained in
../WAVE_14_REGEX_CONTRACT_REPORT.md and validation-wave-14-regex-contract.
The leaked-render gap is JSX parsing, not IR, single assignment or captures,
so its claim is not falsely marked parked on those analyses.

## Landing refresh on main c01907a7

Rebased cleanly onto current main c01907a7 at tested source 3605a6fe.
All owned byte-oracle suites pass again in 426.726 s, with twenty semantic
mutants, nine declaration mutants, sanitizers and released-handle checks.
Bridge, filtered uncached Node and vet pass. The 741-constructor matrix agrees.
Shared regex table 071fb012 has no applicable owned-rule rows and confirms
the dynamic constructor gap. No new matcher, shared-file edits or new claims.
Remaining regex, JSX, label and numeric-dispatch boundaries are unchanged.
See ../WAVE_14_RESUME_REPORT.md and validation-wave-14-resume.

## Ranked pool exhausted and landing refreshed

All 197 ranking rows are now either original base ports or reserved under claims
on fetched origin branches: 584 refs, 33 distinct claim blobs, zero remaining.
No new rules claimed. Rebased onto current main b8fb957a at tested source
1c56e364; all owned byte suites PASS 461.469 s, with semantic mutants, sanitizer
corpus comparisons and released-handle checks. Bridge, filtered uncached Node
including inherited static-field reads, and vet pass. Named-kind clarification
is accepted; numeric API absence is no longer a selection blocker. Existing
regex, real JSX and undefined-label paths remain incomplete.
See ../WAVE_14_EXHAUSTED_REPORT.md and validation-wave-14-exhausted.

## Stage1-lint area integration and real JSX validation

Rebased onto fetched origin/area/stage1-lint 7481e032, including harness
41eb6eab2 and current main 39638d9e. At tested source c88028f3 all owned
Go byte suites PASS 478.251 s, plus bridge, filtered uncached Node, registry
and vet. Real JSX parsing is no longer the leaked-number-render blocker:
16 real TSX controls match 13 findings and 5,657 bytes under sanitizers,
with a compiled exit-0 comparison-only mutant. The continuation witness
also matches its one finding under sanitizers. Existing regex validation
and undefined-label gaps remain. No unclaimed ranked rules remain and no
new reservations are taken. See ../WAVE_14_AREA_REPORT.md and
validation-wave-14-area for exact inputs, complete streams and mutants.

## Area runtime refresh

Rebased onto area d65a8f93 at tested source 7a600bed3, including current
main 39638d9e. Shared runtime changes were retained without owned source
adjustments. All owned byte suites PASS 511.172 s with twenty-one semantic
mutants, nine declaration mutants, sanitizer corpora and released handles.
Bridge, filtered uncached Node including runtime last-index-of, registry
and vet pass. Real JSX remains green. Zero ranked rules are unclaimed;
no new reservations. Regex validation and undefined labels remain incomplete.
See ../WAVE_14_AREA_REFRESH_REPORT.md and validation-wave-14-area-refresh.

## Latest landing refresh, October 7

Rebased onto area b84a9d93 and current main c7991b90. Tested source a6e9ab6ba: owned oracle PASS 474.949s, bridge, uncached Node and five new proven fixtures, registry and vet pass. Twenty-one semantic byte mutants caught; ownership and sanitizer checks pass. No new claim; all-origin audit leaves no candidates. Remaining regex and undefined-label boundaries and full-stage exclusions are recorded in WAVE_14_LATEST_REPORT.md.

## Registry landing refresh, October 7

Rebased onto area b46914832, containing current main c7991b900. Source a89cd2eec: owned byte oracle PASS 480.336s, registry and vet pass, all semantic mutants and sanitizer/ownership checks pass. All-origin audit leaves no candidates. No new claim. Remaining boundaries and excluded gates are in WAVE_14_REGISTRY_REFRESH_REPORT.md.

## Typeof landing refresh, October 7

Rebased onto area d3a37422c and main b6b1538b0. Source 9fb3da437: owned oracle PASS 490.484s, uncached typeof Node checks PASS 12.241s, registry and vet pass. All semantic mutants, sanitizer and released-handle checks pass. No new claim; 658 origin refs leave zero candidates. Remaining gaps and excluded gates are in WAVE_14_TYPEOF_REFRESH_REPORT.md.
