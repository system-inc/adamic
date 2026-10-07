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
