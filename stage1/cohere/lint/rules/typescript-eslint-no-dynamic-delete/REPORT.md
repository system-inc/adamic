Built: @typescript-eslint/no-dynamic-delete in owned .a modules.
Commits: pre-code claim e9a72247; implementation 1d8da3ab; shared owned evidence commit follows.
Commands and outputs: four-way fixtures/corpus PASS; native/Node/Go 1.62/2.35/9.70 findings per second.
Mutant: accepting unary plus compiled and exited cleanly, then failed comparison with Go on all three runtimes.
Not covered: full repository gate or ordinary production registration without the shared .a compatibility patch.

Complete reproduction, observations, logs and limitations are in
[the next-three report](../typescript-eslint-no-duplicate-enum-values/REPORT.md).
The complete capture has 42 upstream cases plus the owned witness, 25 findings,
8,616 identical fixture bytes and 12,340,345 identical corpus bytes on Go, Node
source, emitted JavaScript and sanitized native. All 77 compiler and 139 stage1
sources were compared. The natural timing corpus has 2 findings in 216 files.
