Temporary: comes out when codex/fallthrough-and-implicit-returns lands

Plan published before source edits on October 7, 2026.

Keep noFallthroughCasesInSwitch on. Review scanner.ts's intentional fallthrough
cases at original lines 444, 651, 845, 1553, 1563, 1686, 2133, 2765, 2838, 2907,
3300, 3480 and 3861. Replace each falling path with the same continuation and
an explicit exit. Preserve token positions, flags, return values, side effects,
lexical bindings and the target of every break/continue. Do not insert a bare
break into a path which originally ran the following case.

Census reason: TS7029. Restrict edits to scanner.ts, use the stock TypeScript
6.0.3 AST, document declined cases, and require idempotence. The closure contains
additional TS7029 cases outside scanner.ts; these remain separately tracked.

Validation required before calling this complete: the full stage 3 baseline
oracle, unchanged scanner token bytes on Node, a semantic fallthrough mutant
caught by token comparison or an upstream baseline, and idempotence. This is a
README-only plan. No fallthrough source edit has been made or validated here.
Results and any stopping point are recorded in stage3/drivers/scanner/BLOCKERS.md.

## Stopping point

This plan remains deferred. adapt.cjs is an explicit no-change placeholder so
apply.sh can enumerate the published plan without failing on a missing script.
It logs status deferred and the reason. No fallthrough rewrite was made, so no
fallthrough semantic mutant or completed repair is claimed. All 13 scanner-local
TS7029 findings remain in the ordered scanner blocker ledger.
