# prefer-regex-literals

Stopped before registering a partial rule.

Required Go call: `reference.NewTracker(ctx.SourceFile, ctx.TypeChecker, nil)`, `cohere/internal/lint/rules/core/prefer_regex_literals.go:196`.

Shared reference tracker is absent; the old private tracker is not reused.

Reproducer:

```tsx
const R = RegExp; new R("x");
```

Shared helpers and checker lifecycle were not edited. No upstream parity or mutant pass is claimed for this rule in this continuation.
