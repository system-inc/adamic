Built @typescript-eslint/default-param-last as .a modules with directory registration, an unchanged upstream Go adapter and owned witnesses.
Commits: pushed continuation claim 0a9a5ae6; implementation 411d5137.
Checks: 336 supported compiler/stage1/upstream/witness rows, 12,401,212 canonical bytes identical on Go, source Node, emitted JavaScript and sanitized native; final suite PASS 258.271s.
Mutant: typescript-default-param-last-wrong-answer compiles/runs cleanly and is caught only by Go comparison on all three Adamic executions.
Not covered: default integration, pinned CLI self-lint of .a, arbitrary parser recovery and the full repository gate.

Native / Node / Go findings per second, best of three mixed-corpus rounds:
81.82 / 112.65 / 499.51. No parser gap occurs in this rule's captured cases or named corpora.
See [the continuation report](../structure-tailwind-no-physical-direction/REPORT.md)
for complete commands, setup timing, raw evidence, each mutation, scopes, limits
and scratch-only reproduction. All six authored Adamic modules in this batch
use .a; no shared sources or generated registries are committed.
