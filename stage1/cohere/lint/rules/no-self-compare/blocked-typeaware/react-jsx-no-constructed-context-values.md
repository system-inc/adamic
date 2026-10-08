# react/jsx-no-constructed-context-values

Stopped before registering a partial rule.

Required Go call: `walk.ctx.TypeChecker.GetTypeAtLocation(node)`, `cohere/internal/lint/rules/react/jsx_no_constructed_context_values_stability.go:283`.

Foreign callee-body node queries cannot be addressed through Checker.ask, which binds all node indices to the current file/parser. askFile selects the current SourceFile, and the area questions do not provide resolved callee declarations, module text or a file/path/span query for a foreign expression. The old private MemoView/checker-file views are not reused.

Reproducer:

```tsx
// helper.ts
export function make() { return {}; }
// main.tsx
import { make } from "./helper";
function App() { const value = React.useMemo(() => make(), []); return <Ctx.Provider value={value} />; }
```

Shared helpers and checker lifecycle were not edited. No upstream parity or mutant pass is claimed for this rule in this continuation.
