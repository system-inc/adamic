# Slot 02 batch 11

Three separate .a helpers preserve private Go composition through explicit dependencies.

`ingest(nodes, roots, path, skipped, deps)` reads a numeric CSS node arena. atRule and container are projections of actual Go Kind and IsContainer, not guesses based on children. It traverses ordered roots, recursively visits ordinary containers and unknown container at-rules, and dispatches imports, themes, utilities and custom variants. Import resolution precedes loading. Dependency errors use numeric identity handles, with -1 for nil, and are returned unchanged immediately. Config and plugin directives append their exact name, params and current path. A special at-rule does not also traverse its body; its dependency owns that work. The dependency callbacks can capture collector state, including the same skipped list. Filesystem access, parsing, child mutation and externally owned operations are not reimplemented.

`nestedFunction(builder, node, memberHeader)` delegates once with opaque original identities, even for a nil node or an unreachable current block. memberHeader owns syntax and parameter/decorator behavior.

`variableDeclarationList(builder, node, isDeclarationList, hasDeclarations, declarations, visit)` handles nil before reading any list, preserves nil/empty declaration storage and invokes visit in order with original identities. A nonnil caller must supply the actual declaration-list tag; the helper refuses an invalid private-helper precondition. Go's AsVariableDeclarationList is a type assertion and panics for a wrong kind, so this is not a total arbitrary-AST predicate. The nil guard and typed adapter are distinct. Numeric node -1 means nil; declaration storage and ordered handles project Go's actual NodeList. The externally owned variableDeclaration operation remains a callback.

The oracle loads the unchanged actual runtime consumer-source captures from batch6 and batch10, verifies all six Tailwind and four CFG consumers against the frozen ledger, parses them again in pinned Go cohere and uses actual CSS nodes and graph helper entries. Captures are reused, not represented as a fresh upstream package run. See evidence/capture-manifest.json and those batches' original capture logs for provenance. Additional real parser controls contain multiple declarations; a factory-created declaration list has nil Declarations. Go dependency overlays preserve original method bodies and replace only the externally owned calls. Callback traces include identity, order, mutable builder/block effects and direct error identity. Go CSS parser errors are outside this helper's input contract and are excluded from replay, not accepted as clean lint results.

Run from the repository root with /workspace/adamic-tools/env.sh sourced:

```
go test ./stage1/cohere/lint/helpers -run '^TestSlot02Batch11$' -count=1 -v -timeout=20m > /tmp/slot02-batch11.log 2>&1
```

RULES.md names every consumer. readiness.json subtracts this slot's thirty-three retained helpers only. Whole-rule findings and dependency wiring are outside this unit.
