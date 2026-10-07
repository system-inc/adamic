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
