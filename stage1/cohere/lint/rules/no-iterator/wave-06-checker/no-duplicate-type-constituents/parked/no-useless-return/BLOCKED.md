# no-useless-return

Stopped on area `9b7547976bea1b00f8b04508b9102d03511f65b3`.

Required Go call: `control_flow_graph.Build` at `cohere/internal/lint/rules/core/no_useless_return.go:125`.

The shared control-flow graph builder/hooks are absent. GetPromisedTypeOfPromise at line 381 is also unanswered.

Reproducer: `blocked.ts.txt`. This is an analysis input, not a successful parity witness. No descriptor is registered and no upstream parity or mutant result is claimed. Shared checker and helper files remain untouched.
