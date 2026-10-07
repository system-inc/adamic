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
