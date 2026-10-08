# react-hooks/static-components

Stopped before registering a partial rule.

Required Go call: `high_level_intermediate_representation.ForFunction(ctx, functionNode)`, `cohere/internal/lint/rules/react/static_components.go:120`.

Shared source-to-HIR/SSA/capture lowering is absent from the lint RuleContext.

Reproducer:

```tsx
function App() { function Child() { return <div />; } return <Child />; }
```

Shared helpers and checker lifecycle were not edited. No upstream parity or mutant pass is claimed for this rule in this continuation.
