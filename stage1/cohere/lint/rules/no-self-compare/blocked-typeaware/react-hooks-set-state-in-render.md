# react-hooks/set-state-in-render

Stopped before registering a partial rule.

Required Go call: `high_level_intermediate_representation.ForFunction(ctx, functionNode)`, `cohere/internal/lint/rules/react/set_state_in_render.go:183`.

Shared source-to-HIR/SSA/capture lowering is absent from the lint RuleContext.

Reproducer:

```tsx
function App() { const [x, setX] = useState(0); setX(1); return <div>{x}</div>; }
```

Shared helpers and checker lifecycle were not edited. No upstream parity or mutant pass is claimed for this rule in this continuation.
