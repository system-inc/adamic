# react/jsx-fragments

Stopped on area `9b7547976bea1b00f8b04508b9102d03511f65b3`.

Required Go call: `ctx.TypeChecker.GetSymbolAtLocation(identifier)` at `cohere/internal/lint/rules/react/jsx_fragments.go:290`.

Complete resolved symbol declaration nodes (including ImportSpecifier and BindingElement) are absent. scope-locals is an unresolved binder table, not the checker's symbol answer; type-origin reports a different symbol.

Reproducer: `blocked.tsx.txt`. This is an analysis input, not a successful parity witness. No descriptor is registered and no upstream parity or mutant result is claimed. Shared checker and helper files remain untouched.
