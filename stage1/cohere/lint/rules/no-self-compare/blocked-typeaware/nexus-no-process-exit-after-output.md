# nexus/correctness-no-process-exit-after-output

Stopped before registering a partial rule.

Required Go call: `control_flow_graph.Build(root.node, control_flow_graph.Hooks[correctnessNoProcessExitAfterOutputEvent]{...})`, `cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output.go:511`.

Shared parser-node CFG construction is absent; the old private Graph copy is not reused.

Reproducer:

```tsx
console.log("x"); process.exit(0);
```

Shared helpers and checker lifecycle were not edited. No upstream parity or mutant pass is claimed for this rule in this continuation.
