Built @next/next/no-assign-module-variable as .a modules with directory registration, an unchanged upstream Go adapter and owned witnesses.
Commits: pushed continuation claim 0a9a5ae6; implementation b83fd28c.
Checks: 232 supported compiler/stage1/upstream/witness rows, 12,356,510 canonical bytes identical on Go, source Node, emitted JavaScript and sanitized native; final suite PASS 258.271s.
Mutant: next-no-assign-module-variable-wrong-answer compiles/runs cleanly and is caught only by Go comparison on all three Adamic executions.
Not covered: default integration, pinned CLI self-lint of .a, arbitrary parser recovery and the full repository gate.

Native / Node / Go findings per second, best of three mixed-corpus rounds:
7.06 / 9.88 / 42.07. No parser gap occurs in this rule's captured cases or named corpora.
See [the continuation report](../structure-tailwind-no-physical-direction/REPORT.md)
for complete commands, setup timing, raw evidence, each mutation, scopes, limits
and scratch-only reproduction. All six authored Adamic modules in this batch
use .a; no shared sources or generated registries are committed.
