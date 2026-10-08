# @typescript-eslint/no-meaningless-void-operator

Stopped before registering a partial rule.

Required Go call: `type_checking.IsThenableType(ctx.TypeChecker, argument, argumentType)`, `cohere/internal/lint/rules/typescript/no_meaningless_void_operator.go:191`.

Shared thenable predicate is not available on RuleContext.

Reproducer:

```tsx
declare const p: Promise<void>; void p;
```

Shared helpers and checker lifecycle were not edited. No upstream parity or mutant pass is claimed for this rule in this continuation.
