# nexus/correctness-no-uncleared-race-timeout

Stopped before registering a partial rule.

Required Go call: `type_checking.IsSymbolFromDefaultLibrary(ctx.Program, ctx.TypeChecker.GetSymbolAtLocation(constructor))`, `cohere/internal/lint/rules/nexus/correctness_no_uncleared_race_timeout.go:173`.

Shared default-library symbol predicate is absent.

Reproducer:

```tsx
Promise.race([work(), new Promise((_resolve, reject) => setTimeout(reject, 1))]);
```

Shared helpers and checker lifecycle were not edited. No upstream parity or mutant pass is claimed for this rule in this continuation.
