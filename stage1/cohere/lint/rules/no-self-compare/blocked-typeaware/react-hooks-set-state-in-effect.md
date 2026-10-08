# react-hooks/set-state-in-effect

Stopped before registering a partial rule.

Required Go call: `high_level_intermediate_representation.ForFunctionWithoutManualMemoization`, `cohere/internal/lint/rules/react/set_state_in_effect.go:267`.

Shared source-to-HIR/SSA/capture lowering is absent from the lint RuleContext.

Reproducer:

```tsx
function App() { const [x, setX] = useState(0); useEffect(() => { setX(1); }, []); return <div>{x}</div>; }
```

Shared helpers and checker lifecycle were not edited. No upstream parity or mutant pass is claimed for this rule in this continuation.
