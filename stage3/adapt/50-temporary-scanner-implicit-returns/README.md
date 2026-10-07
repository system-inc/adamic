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

## Completed validation

The README-only announcement was pushed in 586caf4 before adapt.cjs was written.
The adapter now adds exactly two explicit undefined returns in scanner.ts and
changes no other source file. Reapplication reports zero files and zero returns
changed. Removing only getShebang's added return in scratch is caught by
scanner.ts:966:43 TS7030. Neither strict checker option was disabled.

The unfiltered stage 3 baseline passes: 106,367 tests, zero failures, zero baseline
differences, 347.707 seconds total. The scanner's 509,014 token lines on a fixed
pre-edit compiler corpus match Node's control byte for byte (27,879,197 bytes).
Native checking loses the two scanner TS7030 findings and still stops on 2,652
other diagnostics. Full logs, ordered gate results and the stopping point are in
../../drivers/scanner/BLOCKERS.md and its evidence directory.
