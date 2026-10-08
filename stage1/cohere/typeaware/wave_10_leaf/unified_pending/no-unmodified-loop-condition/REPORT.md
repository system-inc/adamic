# Shared parser blocker

The shared checker port is complete and its owned witness is byte-identical across Go, Node, emitted JavaScript and sanitized native. Full upstream replay fails in `TestNoUnmodifiedLoopConditionUnnamedFunctionDeclarationHasNoNameToReach`, `cohere/internal/lint/rules/core/no_unmodified_loop_condition_test.go:241`, on the unnamed declaration at line 252.

Reproducer: `var foo = 0; while (foo) { } export default function () { ++foo; }`. Native scanner exits 70: `parser slice unsupported primary CloseParenToken at 52`. Go accepts this valid TypeScript declaration. The listener, oracle adapter, witness and mutant remain here pending shared parser support; this directory is deliberately outside the active registry. The mutant has not been certified on this unified adapter.

Evidence: `../../evidence/unified-landing/loop-parity-failure.jsonl.gz`.
