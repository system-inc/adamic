# nexus/correctness-require-blocking-standard-streams

Stopped before registering a partial rule.

Required Go call: `control_flow_graph.Build(root, control_flow_graph.Hooks[event]{...})`, `cohere/internal/lint/rules/nexus/correctness_require_blocking_standard_streams.go:840`.

Shared CFG construction is absent.

Reproducer:

```tsx
#!/usr/bin/env node
process.exit(0);
```

Shared helpers and checker lifecycle were not edited. No upstream parity or mutant pass is claimed for this rule in this continuation.
