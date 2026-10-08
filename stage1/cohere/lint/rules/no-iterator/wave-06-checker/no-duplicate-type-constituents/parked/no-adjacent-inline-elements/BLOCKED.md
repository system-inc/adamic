# react/no-adjacent-inline-elements

Stopped on area `9b7547976bea1b00f8b04508b9102d03511f65b3`.

Required Go call: `ctx.TypeChecker.GetSymbolAtLocation(identifier)` at `cohere/internal/lint/rules/react/checked_requires_onchange_or_readonly.go:352`.

The shared isPragmaCreateElementCall/bindsToPragmaImport helper is absent; its bare-call arm needs every resolved symbol declaration and import/binding initializer syntax at lines 356-357. The namespaced/JSX arm alone would be incomplete.

Reproducer: `blocked.tsx.txt`. This is an analysis input, not a successful parity witness. No descriptor is registered and no upstream parity or mutant result is claimed. Shared checker and helper files remain untouched.
