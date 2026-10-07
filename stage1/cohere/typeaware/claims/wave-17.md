# Type-aware wave 17

Branch: `codex/typeaware-wave-17`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Positions 49, 50 and 51 after excluding the base branch's 26 ports from the
combined compiler plus repository counts linked by VOLUME_REPORT.md, sorted by
volume descending and rule name ascending:

- 49: `@next/next/no-async-client-component` (0 findings)
- 50: `@next/next/no-duplicate-head` (0 findings)
- 51: `@next/next/no-script-component-in-head` (0 findings)

All origin branch tips were fetched and their stage1 source and claims inspected
before this claim. None of these three rules was already ported or claimed.
The nextjs syntax claim on origin/codex/stage1-nextjs-lint names other rules.

The shared parser's JSX boundary affects the two Head rules. Any unsupported
input must be reported explicitly rather than counted as silent agreement.

## Validated status

The non-JSX async rule is implemented in dd5481bc. Both Head rules remain
unimplemented, and async components containing JSX are also unsupported by the
shared native parser. Positive Go witnesses and explicit native parser failures
are recorded in [the wave report](../WAVE_17_REPORT.md). This is a partial unit,
not three completed ports. No rule was skipped for an existing claim or port.

## Continuation claim

Fetched all origin heads on October 7, 2026: 320 remote refs and 30 unique
claim files. Selection uses the combined compiler plus repository counts linked
by VOLUME_REPORT.md, descending volume and ascending rule name. Existing ports
on origin/codex/tsgo-c-library (5afbdb83) and origin/main (ef3d907e), and exact
rule names claimed on any origin branch, are excluded.

The first three eligible rules are:

- `nexus/correctness-no-collection-misuse` (global rank 116, 0 findings)
- `nexus/correctness-no-discarded-outcome` (global rank 117, 0 findings)
- `nexus/correctness-no-discarded-pure-result` (global rank 118, 0 findings)

All earlier rules are already ported or claimed. None of these three is ported
on either base branch or named in any origin claim file. This update is pushed
before implementation, on the existing codex/typeaware-wave-17 branch.

## Continuation validated status

The three Nexus ports are implemented in 0be2374d. All three agree with the
unchanged production Go rules on the frozen compiler and repository manifests,
and 53 control files, normally and under sanitizers. Per-rule mutants, the new
awaited-ancestry fact mutant, and released-handle/registry checks pass.

The additional old request-refusal harness cannot construct its unknown-question
mutant because it hardcodes the former bridge error-return line. The runtime
refusal itself still passes. Per Ahra's correction, that shared harness was not
edited, and no further rules were claimed. The failure and all evidence are in
[the continuation report](../WAVE_17_NEXT_REPORT.md). The original Next JSX
limitations recorded above remain unchanged.

## Original JSX rules continued

Both Head rule decisions are now implemented in their own `.a` files. The
40-fixture raw-AST projection test matches complete production Go diagnostics
(21 findings, 10,044 bytes), normally and under sanitizers, and catches a mutant
for each rule. This validates native rule decisions, not the shared parser.
Production JSX parsing still refuses with panic 70. The own-directory projection
runner is validation only. See [the Head report](../WAVE_17_HEADS_REPORT.md).

## Third batch claim

The prior claimed implementations and explicit shared-parser limitations are
pushed in 8f3d8d05. A fresh October 7 fetch inspected 335 origin refs, 33 unique
Markdown claim blobs, and native stage1 source on main and the library branch.
The first eligible checker-dependent rules in the combined by-volume ranking are:

- `no-global-assign` (global rank 137, 0 findings)
- `no-implicit-globals` (global rank 138, 0 findings)
- `no-implied-eval` (global rank 139, 0 findings)

None is already ported on either requested base or claimed on any origin branch.
The complete scan is preserved in validation-wave-17-third/selection.json.
This claim is pushed before implementation, on the same branch.

## Third batch validated status

All three global-rule ports are implemented in a7306edb. Default and option
controls, both frozen corpora, ASan/UBSan, per-rule and bridge-fact mutants, and
both new questions' released-handle controls pass. The native/Go timings and
complete evidence are in [the third report](../WAVE_17_THIRD_REPORT.md). No shared
harness, parser or registration generator was changed. Original JSX parsing and
the existing shared unknown-request mutant anchor remain integration gaps.

## Fourth batch claim

The prior batch and evidence are pushed at 101aff2f. A fresh October 7 fetch
inspected 348 origin refs, 33 unique Markdown claim blobs, and native stage1
source on origin/main and origin/codex/tsgo-c-library. The first eligible
checker-dependent entries in the combined by-volume ranking are:

- `no-throw-literal` (global rank 152, 0 findings)
- `no-useless-backreference` (global rank 153, 0 findings)
- `prefer-arrow-callback` (global rank 154, 0 findings)

Earlier entries were skipped because they are ported or claimed elsewhere.
These three are neither ported on the requested bases nor named in an origin
claim. The complete scan is in validation-wave-17-fourth/selection.json.
This claim is pushed before implementation, on codex/typeaware-wave-17.

## Fourth batch validated status

All three ports are implemented in b2a0bc72. The final 335 controls, exact fix
serialization, both option flags, frozen compiler and repository corpora,
ASan/UBSan, four comparison-only mutants and two released-handle probes pass.
No new checker question or shared-file edit was needed. Complete evidence and
native versus Go timings are in [the fourth report](../WAVE_17_FOURTH_REPORT.md).
The previously documented shared unknown-request mutant and JSX parser gaps
remain; no rule in this batch is blocked by them.

## Fifth batch claim

The fourth batch is pushed at ca3ed859. A fresh October 7 fetch inspected
389 origin refs and 33 unique Markdown claim blobs. The first eligible entries
in the combined checker-dependent by-volume ranking are:

- `react-hooks/unsupported-syntax` (global rank 171, 0 findings)
- `react-hooks/use-memo` (global rank 172, 0 findings)
- `react/boolean-prop-naming` (global rank 173, 0 findings)

Earlier entries were skipped as ported or claimed. These three are neither
ported on origin/main or origin/codex/tsgo-c-library nor claimed on any fetched
origin branch. The complete audit is validation-wave-17-fifth/selection.json.
This claim is pushed before implementation on codex/typeaware-wave-17.

## Fifth batch implemented and blocked status

Hook decisions are implemented in 34369377, with final interpolation correction
4f9f26dc. All hook message controls and the implemented default-pattern PropTypes
portion match Go under sanitizers, with a mutant per rule. Ordinary native
parsing matches 44 positive/negative controls and both frozen corpora.
JSX decisions are additionally validated through independent raw AST fixtures.
Production JSX still exits with parser panic 70. BooleanPropNaming is incomplete:
its configurable native regexp matcher is blocked by stage0's RegExp constructor
refusal, and complete component/props integration remains unwired. No complete
BooleanPropNaming parity is claimed and no further rules are claimed. Exact
commands, limits and timings are in [the fifth report](../WAVE_17_FIFTH_REPORT.md).

## Landing cap status

The sole published branch from this session was rebased onto main e8ba3d5d;
validated source tip 7138b515. All seven wave-17 oracles and the inherited
inventory oracle passed again, including sanitizer, fact, handle and comparison
mutants. No new rules were claimed. Full evidence and historical SHA mapping
are in [the landing report](../WAVE_17_LANDING_REPORT.md). Production JSX and
complete configurable BooleanPropNaming remain the documented unfinished scope;
current main explicitly refuses RegExp with a nonconstant pattern.

## Refreshed landing cap status

All existing work is rebased onto main f8013f0b at validated source 6ffc1aff.
All eight selected oracles passed again in 860.616s, including sanitizer and
mutation checks; no source repair or new claim. The 15 numeric listener
declarations pass verification. Exact evidence and remaining limits are in
[the refreshed landing report](../WAVE_17_LANDING2_REPORT.md).

## Parked React integration status

Per the latest instruction, the already pushed JSX-dependent scope is parked
and counts as finished for the landing-first work-in-progress cap. This status
is not a claim of complete end-to-end parity.

- `@next/next/no-async-client-component`: ordinary-parser port is green and
  pushed; JSX-bearing components are parked on native JSX parsing.
- `@next/next/no-duplicate-head`: native decisions and mutants are green and
  pushed; production integration is parked on native JSX parsing.
- `@next/next/no-script-component-in-head`: native decisions and mutants are
  green and pushed; production integration is parked on native JSX parsing.
- `react-hooks/unsupported-syntax`: ordinary native controls are green and
  pushed; JSX-bearing React analysis is parked on native JSX parsing.
- `react-hooks/use-memo`: tested ordinary-parser decisions are green and pushed;
  JSX-bearing React analysis is parked on native JSX parsing. No complete
  high-level IR or capture-analysis integration is claimed.
- `react/boolean-prop-naming`: default-pattern PropTypes decisions and message
  controls are green and pushed; complete component/props integration is
  unwired and parked. Additional blocker: nonconstant RegExp patterns are
  refused by stage 0, so configured-pattern parity remains incomplete.

Native cohere analysis module integration is tracked by #dnv6f2c; JSX support is
landing through area/stage1-lint. This branch does not modify those shared
modules or push to area/stage1-lint. Prior byte-oracle, sanitizer and mutant
results are in WAVE_17_LANDING2_REPORT.md. The numeric listener declarations are
already pushed, but the existing implementations still need conversion to the
shared loaded-node interface.

A fresh scan inspected 529 origin refs and 33 unique Markdown claim blobs.
Main remains f8013f0b, validated source and evidence remain e3677e31 before this
status update. `require-atomic-updates` remains unclaimed but requires the native
control-flow and escape analysis and is ineligible for an analysis-independent
batch. Other rules are not claimed by this status update.

## Regex requirement update

The already ported BooleanPropNaming default now uses its exact pinned Go pattern
as one JS RegExp literal, replacing the handwritten matcher. The fifth-batch
byte oracle, sanitizer and mutation checks pass again; a literal-class mutant
is independently caught. Configurable patterns and full React integration remain
parked. Evidence: [regex update](../validation-wave-17-regex/README.md).
No new rules are claimed by this update.

## Current-main landing and shared regex row

Existing work is rebased onto main c01907a7, rebased tip f36356f1. All eight
selected oracles passed again (909.575s). The remaining message interpolation
matcher now uses shared regex row react/boolean_prop_naming.go:621; the final
fifth-batch oracle passes in 171.661s and a literal mutant is caught at byte
1134. No new claim; parked scopes remain explicit.
Evidence: [third landing report](../WAVE_17_LANDING3_REPORT.md).

## Sixth batch claim (withdrawn)

The existing work and parked scopes are pushed at fdbab035, on current main
c01907a7. A fresh origin fetch confirms both tips unchanged. The remaining
React candidates are evaluated individually rather than categorically excluded:
these production implementations need AST/component/binding helpers, not the
native high-level IR, single-assignment or capture-analysis modules parked by
#dnv6f2c. The earlier non-React-only stopping interpretation was too broad.

The next three unported/unclaimed entries are:

- `react/no-danger-with-children` (global rank 185, 0 findings)
- `react/no-multi-comp` (global rank 186, 0 findings)
- `react/no-namespace` (global rank 187, 0 findings)

The initial fetched snapshot reported no competing claim or source hit for
these three. Claim 14713630 was pushed before implementation. A later explicit
all-heads refresh revealed wave 11 already reserving the same three; this
reservation is withdrawn as described below. New handlers take their handed numeric-kind node
and declare numeric kinds in rule.json. Shared harness ab70f38d4 is visible on
origin/lint-rules/harness but is not yet on main; any production-integration gap
will be named separately from native decision validation. No shared registration,
parser, harness or compiler file will be edited for this batch.


### Sixth batch withdrawal after refreshed origin audit

The explicit all-heads fetch now has 583 origin refs and 33 distinct Markdown
claim blobs. Wave 11 reserves all three in 667e6a47a4dcc9047044b0316eb1525ba2a5a48d
(commit time 05:34:09 UTC), earlier than our 14713630 claim (05:43:56 UTC).
Commit times do not establish push visibility order, but this is a duplicate
reservation, so all three are skipped here in favor of wave 11. No duplicate
implementation is committed or offered as a production port. The experimental
code is preserved in /workspace/wave-17-sixth/withdrawn-implementation.patch.

The native experiments passed on 309 controls and both frozen corpora, including
sanitizers, each comparison-only mutant and released-handle checks. Their input
is an independent projected numeric AST, not the production native parser;
full production integration was not established. Detailed evidence and timing
limits are in ../WAVE_17_SIXTH_WITHDRAWAL_REPORT.md.

After excluding main/library ports and every origin Markdown reservation,
including the competing wave-11 claim, no unclaimed rule remains among the
197 checker-dependent entries. No further claim is made. Existing wave-17 work
remains based on current main c01907a7 and has its previously reported green
landing checks. Shared model ab70f38d4 is still not on main. No main, area,
shared harness, parser or generator file was changed or pushed.
