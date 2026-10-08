# react-hooks/error-boundaries

The port follows pinned Go cohere's ErrorBoundaries rule: report JSX created in a caught try block inside a top-level component or hook eligible for React compilation. Catch/finally bodies, nested compilation units, rest/invalid ref parameters, non-node returns, and Unicode hook/component naming keep Go's exact behavior. The shared rules/react helper package owns the function/name contracts; the rule owns its listener, try-clause identity test, message, witness, oracle adapter, and mutant.

69 unique upstream source/file/options cases pass Go/source Node/emitted JavaScript/sanitized-native findings comparisons. The owned witness reports in Go and every port backend. The try-block mutant compiles and is caught by output comparison on Node, emitted JavaScript, and native. No fixes or suggestions are attached, matching Go.

Registration follows `docs/lint-registration.md`; no shared registry or harness files are changed.
