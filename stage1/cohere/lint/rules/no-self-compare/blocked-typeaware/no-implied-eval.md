# @typescript-eslint/no-implied-eval

Stopped before registering a partial rule.

Required Go call: `type_checking.IsBuiltinSymbolLike(ctx.Program, ctx.TypeChecker, t, "FunctionConstructor")`, `cohere/internal/lint/rules/typescript/no_implied_eval.go:280`.

Shared builtin-symbol/base-type predicate is not available on RuleContext.

Reproducer:

```tsx
new Function("return 1");
```

Shared helpers and checker lifecycle were not edited. No upstream parity or mutant pass is claimed for this rule in this continuation.
