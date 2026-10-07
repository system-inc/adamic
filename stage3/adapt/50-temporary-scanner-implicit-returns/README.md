Temporary: comes out when codex/fallthrough-and-implicit-returns lands

Plan published before source edits on October 7, 2026.

Keep noImplicitReturns on. Add explicit `return undefined;` at the end of
getShebang (src/compiler/scanner.ts:966) and scanIdentifier
(src/compiler/scanner.ts:2425), whose original falloff already returns undefined.
Preserve every original statement and the inferred optional result. Match the
functions with stock TypeScript 6.0.3's AST, require exactly one of each, and make
the adaptation idempotent. Only scanner.ts is edited.

Census reason: TS7030. These are the two scanner-local observations from the
adaptation-10-only scratch build. The closure also contains other TS7030 sites;
this adaptation does not claim to repair them.

Validation required before calling this complete: stage3/apply.sh, the stage 3
baseline oracle with no test filter, unchanged scanner token bytes on Node,
a scratch removal of one added return caught by TS7030, and an idempotence run.
This README-only commit announces the edit; it does not claim it was made or
validated. Results are recorded in stage3/drivers/scanner/BLOCKERS.md.
