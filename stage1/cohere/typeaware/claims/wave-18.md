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


The premature continuation reservation from 340ca63c is withdrawn following
Ahra's correction. Only the original three rules above remain claimed.

The original Next visitor is now complete and byte-tested against the native
JSX slice from origin/codex/stage1-jsx-lint. WAVE_18_TITLE_REPORT.md records
its mutant, corpus/sanitizer agreement and remaining shared parser integration.

## Continuation after completing the original visitor

Original completion 1b6147b0 is pushed; its JSX parser integration dependency
is recorded in WAVE_18_TITLE_REPORT.md. The next three rules claimed here are:

1. no-class-assign
2. no-const-assign
3. no-constant-binary-expression

These are the first remaining rules in the combined by-volume ranking after
checking all 328 fetched origin refs, all claim Markdown, and ports on
origin/main (ef3d907e) and origin/codex/tsgo-c-library (5afbdb83). There are
105 claimed and 25 already-ported ranked rules. Each selected rule has zero
compiler and repository findings. The previously withdrawn Nexus candidates
are now claimed by other workers and are skipped. This claim is pushed before
implementation. No shared harness or registration generator will be edited.

Continuation implementation: the three core visitors are complete in 1cba7ee7.
WAVE_18_CORE_REPORT.md records byte agreement, enabled relational options,
mutants, sanitizers, released handles, corpus coverage and observed timings.

## Third batch after pushed core completion

Previous ports and evidence are pushed through c78d9bf9. The next three
rules reserved here, before any implementation, are:

1. no-new-func
2. no-new-native-nonconstructor
3. no-new-wrappers

Selection checks 341 fetched origin refs, 33 distinct claim
Markdown blobs, 117 claimed rules, and 25 ranked ports on
origin/main (ef3d907e) and origin/codex/tsgo-c-library
(5afbdb83). These are the first remaining names
in the combined by-volume ranking; all three have zero corpus findings.

Third-batch implementation: all three constructor rules are complete in
a63e3ffc. WAVE_18_CONSTRUCTOR_REPORT.md records complete byte comparisons,
no-library and ambient controls, mutants, sanitizers and quiet Go/native timing.

## Fourth batch after pushed constructor completion

Previous ports and evidence are pushed through c6d9d07f. The next rules
reserved before implementation are:

1. prefer-promise-reject-errors
2. prefer-regex-literals
3. prefer-rest-params

The audit checks 356 origin refs, 132 claimed names and
25 ranked baseline ports. Baselines remain main ef3d907e
and tsgo-c-library 5afbdb83. These are the first
remaining names in the combined volume ranking, each with zero corpus findings.

Fourth-batch implementation: all three preference rules are complete in e062ed93.
WAVE_18_PREFERENCE_REPORT.md records full findings/fixes/suggestions agreement,
419 controls, both corpora, options, mutants, sanitizers, released handles and timings.
The only shared edit is one bridge question registration line. No additional rules
are reserved in this turn.

## Fifth batch after pushed preference completion

Previous ports and evidence are pushed through eb1251ab. Reserved before code:

1. react-hooks/set-state-in-effect
2. react-hooks/set-state-in-render
3. react-hooks/static-components

The audit checks 389 origin refs, 142 claimed names and 25 ranked baseline ports.
Baselines are main e011f8f6 and tsgo-c-library 5afbdb83.
These are the first remaining names in the combined by-volume ranking.

Fifth-batch status: all three remain unported and reserved.
WAVE_18_REACT_BLOCKER_REPORT.md records the missing native React HIR/SSA and
control-dominance dependency. Existing Go oracle tests pass; native agreement,
mutants, sanitizers and timings are not claimed. No further rules were claimed.

Fifth-batch continuation: native reporting portions are implemented in 57614fd8.
wave_18_react_partial/REPORT.md records four byte-identical production diagnostics,
sanitizers and six reporting/refusal mutants. All three source-analysis ports
remain unfinished behind native HIR/SSA and JSX dependencies. No new claims.

Landing-first update: rebased onto origin/main e8ba3d5d; tested code 72cb3b4f.
All five existing wave-18 oracle suites, checker tests, vet, filtered uncached Node
checks and React reporting/refusal checks pass again. WAVE_18_LANDING_REPORT.md
records fresh evidence. React source analysis remains unfinished; no new claims.

Prepared-HIR continuation: native React validator cores are implemented in
76e453e7. wave_18_react_partial/CORE_REPORT.md records full prepared-input
comparisons, sanitizer runs, four byte mutants and the independently reproduced
Go phi-message ambiguity. Native source integration remains unfinished. No new claims.

## React HIR claims parked

Per Ahra's parking instruction, react-hooks/set-state-in-effect,
react-hooks/set-state-in-render and react-hooks/static-components are parked.
Their native reporters and prepared-HIR validator cores are pushed and tested,
including sanitizers and byte-only mutants. Blockers: native high-level IR
Lower/Construct, SSA, nested-function capture translation, compilation-unit
gates and memo shadow analysis. CoHere's analysis modules are being ported on
#dnv6f2c; JSX support is landing on area/stage1-lint. They count as finished
for the landing cap and remain reserved, not available to another claimant.

## Sixth batch after landing-ready parking

Landing-ready tip 8cf6bf224 is pushed and contains main f8013f0b; all six
wave-18 oracle suites pass on that base, as WAVE_18_F801_LANDING_REPORT.md
records. Reserved before implementation:

1. react/jsx-fragments
2. react/jsx-no-constructed-context-values
3. react/jsx-no-undef

Selection audited 529 fetched origin refs and 33 distinct claim Markdown blobs.
The ranking combines compiler-all.counts and repository-all.counts, descending
total then lexical rule name, and excludes baseline ports and all claimed names.
The five higher-volume Adamic/Nexus candidates are already covered by the ten
ports in COVERAGE_REPORT.md and are skipped. These three selected rules each
have zero corpus findings, no native named port was found on an origin branch,
and none requires React HIR/SSA/capture analysis. Native syntax and raw checker
binding facts are their dependencies. This reservation is pushed before code.
New rules will declare numeric SyntaxKind listeners in their own rule.json
and consume the supplied node. Shared harness and parser files are untouched.

Sixth-batch eligibility correction before code: jsx-no-constructed-context-values
is parked because its separate stability module requires nested callback/callee
return and capture/escape analysis (anyEscapes, StaysHome and functionEvaluation).
It remains reserved, and its implementation has not started. The next eligible
ranked rule replaces it: react/no-adjacent-inline-elements. The active sixth
batch is jsx-fragments, jsx-no-undef and no-adjacent-inline-elements.
No-adjacent-inline-elements uses JSX sibling syntax and createElement import
resolution only; it needs no high-level IR/SSA/capture analysis. This update
is pushed before implementation too.

Sixth-batch implementation is complete in 937dc881: jsx-fragments,
jsx-no-undef and no-adjacent-inline-elements run on numeric supplied nodes.
wave_18_jsx/REPORT.md records complete default/options/corpus byte agreement,
sanitisers, one mutant per rule, numeric-query and stale-handle mutations,
released-handle refusal and native/Go timings. Native uses raw syntax/binding
facts from the C bridge; shared parser/driver files remain untouched.
The four capture/HIR-dependent claims above remain parked. No further batch
is reserved in this turn.

## Landing on main c01907a70

All existing wave-18 work rebased onto fetched main c01907a70 and re-green
at f619ea20dd1d3f4650532eb063effcc0ecd7b037. Six wave-18 suites, rebuilt JSX parity/sanitizers/mutants,
bridge, parked reporters, vet and filtered uncached Node pass. Four analysis
claims remain parked. No new reservation was made in this landing unit.
See wave_18_jsx/LANDING_C019_REPORT.md and validation-c019 for evidence.

## Final unclaimed pair

Reserved react/static-property-placement and react/style-prop-object.
Fetched 576 origin refs and inspected 33 distinct claim Markdown blobs.
These are the only two remaining unclaimed entries in the 197-rule ranking,
after baseline ports and all origin claims are excluded. Neither requires
HIR/SSA/capture analysis; both use syntax and declaration facts. No named
port exists on an origin branch. Main remains c01907a70 and this branch
is landing-ready at a37d18a51. There is no third unclaimed ranked rule.
New rule.json kinds use ast.Kind names as required by the registry.

Final pair complete at 5e8c9523e4969be8796e645523a63cf98d4aa546. Both rules match complete Go records over
73 controls, six option configurations and both corpora, normally and under
sanitizers. Per-rule byte-only mutants and new-question released-handle checks
pass. See wave_18_component_props/REPORT.md. No unclaimed ranked rule remains;
the four previously parked analysis claims retain their named blockers.
