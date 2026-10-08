# no-unreachable-loop

Directory-local registration follows docs/lint-registration.md. The rule adapts the parsed AST into a flat private-constructor CfgNodeIndex arena, uses IndexRoots and Build, and reports a loop when its reachable control-flow block has no iterating back edge. Ignore defaults and options retain upstream meaning. Each graph is disposed after its root is checked; the AST arena is disposed after root traversal. Index classes are imported from the shared arena_index.a home.

The pinned upstream TestNoUnreachableLoop corpus contributes 1,633 unique source/options cases (3,276 recordings), counted by testdata/upstream_count.py. The registry compares findings and bytes with the unchanged Go rule on Node, emitted JavaScript and sanitized native. The mutant reverses reachable iteration classification and is checked on Node and native. testdata/upstream_count.json records the pin and counts.

The area syntax-mutant harness uses a native canary. native_mutant_test.go additionally proves this rule's own mutant on Node and sanitized native, with identical mutated outputs. It calls t.Parallel and isolates all copies. Run go test ./stage1/cohere/lint/rules/no-unreachable-loop -count=1 -timeout=3h.
