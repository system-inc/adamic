# @typescript-eslint/no-floating-promises

Stopped before registering a partial rule.

Required Go call: `type_checking.IsPromiseLike(ctx.Program, ctx.TypeChecker, typePart)`, `cohere/internal/lint/rules/typescript/no_floating_promises.go:372`.

Shared promise/default-library predicate is not ported to RuleContext. The graph facts exist, but duplicating the shared checking helper privately is forbidden.

Reproducer:

```tsx
declare function f(): Promise<void>; f();
```

Shared helpers and checker lifecycle were not edited. No upstream parity or mutant pass is claimed for this rule in this continuation.
