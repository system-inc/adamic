# Slot 02 batch 12

Three .a helpers preserve private Go composition through explicit dependencies.

`resolve(theme, candidate, present, keys, options, values, deps)` calls the actual key-resolution dependency once with unchanged arguments and key-slice identity. Missing keys stop immediately. The bitwise union of caller and stored options selects inline values, including empty strings; otherwise the variable-reference dependency owns the returned value and presence. A missing map entry projects Go's zero value. Theme handles identify stable state supplied to externally owned dependencies.

`variableDeclaration(builder, node, declaration, deps)` handles nil before inspecting a typed declaration projection. Annotation evaluation comes first. Destructuring evaluates the initializer before binding; ordinary declarations read their name, then evaluate an existing initializer and write the name. Numeric node -1 projects nil. Actual Go AsVariableDeclaration requires a variable-declaration node; arbitrary wrong-kind nodes are outside the private helper contract. Callbacks receive original builder/node identities and own expression and binding behavior.

`declineListeners(context, ruleName, system, deps)` returns a fresh concrete DeclineListener with a SourceFile callback. The registry subscribes that callback by the named SourceFile kind. Each instance marks itself reported before description/report callbacks, suppressing repeat and reentrant calls. Message id is designSystemUnavailable. Context and system handles must refer to stable snapshots of Go's value arguments, preserving contained source-file/error/system pointer identity; a mutable worker-global context handle is insufficient. Keep the concrete class type at return/storage boundaries so Adamic retains method prototype origin. No shared registration or finding API is changed.

The standalone oracle reuses unchanged actual consumer-source captures from batch6 and batch10 and reparses them with pinned Go cohere. Theme resolution calls real Go resolveKey/variableReference under dependency-only tracing. CFG entry instrumentation retains the original method and uses observable expression/binding dependencies that mutate the original builder's current block. Two fresh decline listeners per source test independent state and bounded callback reentrancy. Description values come from the actual Go helper. Expected output is removed from Adamic fixture inputs before comparisons.

Run with /workspace/adamic-tools/env.sh sourced:

```
go test ./stage1/cohere/lint/helpers -run '^TestSlot02Batch12$' -count=1 -v -timeout=20m > /tmp/slot02-batch12.log 2>&1
```

RULES.md names the consumers. readiness.json removes only this slot's retained helpers. Whole-rule findings, fixes, suggestions and final dependency wiring are outside this helper unit.
