Built the .a hand-rolled-delay listener, with a reported top-level-await parser gap.
Commits: claim 69e2e1a; implementation 8e8acca; paired return-void implementation 865a747.
Checks: 47 upstream cases plus 212 compiler/stage1 files, 12,358,833 bytes identical on three backends against Go.
Mutant: wrong resolve name compiled and ran cleanly, then the Go comparison caught its missing finding on all three backends.
Not covered: one top-level-await fixture and ordinary .a registration; the complete unit is not merge-ready.

The complete unit report, commands, throughput and explicit blockers are in
[the paired report](../nexus-consistency-no-return-void/REPORT.md).
Native: 17.07 findings/s; Node: 25.16 findings/s; Go: 99.00 findings/s,
best of three on the mixed supported corpus with 21 findings.
