# no-css-tags

Rule logic is compared against unmodified Go cohere using Go-projected ASTs on Node, emitted JavaScript and ASan/UBSan native. Independent-parser integration is blocked by the missing stage1 JSX adapter; the rule explicitly refuses a tree that might have reinterpreted JSX.

See [the unit follow-up report](../no-unsafe-function-type/FOLLOWUP.md) for exact coverage, excluded fixtures, commands, timing, every mutant and remaining gaps.

## Unified harness landing

The current takeover supersedes the historical adapter and harness limits above. The descriptor now uses `node: true` and the rule entry is `.a`. See [the landing report](../../claims/wave1-06-land-report.md) for current evidence and limits.
