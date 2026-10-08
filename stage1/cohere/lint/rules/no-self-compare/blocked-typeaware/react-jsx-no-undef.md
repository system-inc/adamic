# react/jsx-no-undef

Stopped before registering a partial rule.

Required Go call: `jsx.ElementParts(node)`, `cohere/internal/lint/rules/react/jsx_no_undef.go:93`.

Shared JSX element accessor is absent.

Reproducer:

```tsx
const x = <Missing />;
```

Shared helpers and checker lifecycle were not edited. No upstream parity or mutant pass is claimed for this rule in this continuation.

Additionally, area has no `binding-origin` or equivalent node-symbol declaration list; the old private bridge extension was not imported.
