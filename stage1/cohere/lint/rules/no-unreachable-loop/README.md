# no-unreachable-loop

Directory-local registration follows docs/lint-registration.md. The rule adapts the parsed AST into a flat NodeIndex arena, uses IndexRoots and Build, and reports a loop when its reachable control-flow block has no iterating back edge. Ignore defaults and options retain upstream meaning. Each graph is disposed after its root is checked.

The pinned upstream TestNoUnreachableLoop corpus contributes 1,633 unique source/options cases (3,276 recordings), counted by testdata/upstream_count.py. The registry compares findings and bytes with the unchanged Go rule on Node, emitted JavaScript and sanitized native. The mutant reverses reachable iteration classification and is checked on Node and native. testdata/upstream_count.json records the pin and counts.
