# prefer-promise-reject-errors

Stopped before registering a partial rule.

Required Go call: `property.AccessedName(callee, property.Static)`, `cohere/internal/lint/rules/core/prefer_promise_reject_errors.go:247`.

Shared static member-name evaluation is absent; RuleContext.property only returns the name node and does not evaluate computed keys.

Reproducer:

```tsx
Promise[`re${"ject"}`]("bad");
```

Shared helpers and checker lifecycle were not edited. No upstream parity or mutant pass is claimed for this rule in this continuation.
