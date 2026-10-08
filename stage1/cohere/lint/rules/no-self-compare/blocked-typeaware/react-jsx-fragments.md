# react/jsx-fragments

Stopped before registering a partial rule.

Required Go call: `jsx.ElementParts(opening)`, `cohere/internal/lint/rules/react/jsx_fragments.go:207`.

Shared JSX element accessor is absent.

Reproducer:

```tsx
import React from "react"; const x = <React.Fragment />;
```

Shared helpers and checker lifecycle were not edited. No upstream parity or mutant pass is claimed for this rule in this continuation.

Additionally, area has no `binding-origin` or equivalent node-symbol declaration list; the old private bridge extension was not imported.
