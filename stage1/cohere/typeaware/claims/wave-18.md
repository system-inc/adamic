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
