HISTORICAL: resolved after merging origin/area/stage1-lint. The production rule is restored and all 59 upstream cases match on the merged harness.

# Object constructor: captured script kind blocker

Source: `cohere/internal/lint/rules/core/no_object_constructor_test.go:55` (filename at line 11); the firing cases include JSX followed by `Object()` under `repository/source/ObjectConstructor.ts`.
The unchanged unified `TestRulesAgree` calls the Go oracle with that captured `.ts` path. The oracle fails before findings: TypeScript expects `>` instead of parsing the JSX tag. `jsx-in-ts.ts.txt` is the smallest reproducer; interpreting the same bytes as TSX succeeds.
No checker question is missing. This requires the shared capture/manifest to preserve the upstream case's script kind. Changing the rule's test prefix would drop upstream cases, and changing its adapter cannot affect parsing performed before rule dispatch.
The unfinished port is retained below `port/`, outside the registry's rule directories. Its witnesses and mutant passed, but full upstream parity did not complete. It is not registered or claimed green.
