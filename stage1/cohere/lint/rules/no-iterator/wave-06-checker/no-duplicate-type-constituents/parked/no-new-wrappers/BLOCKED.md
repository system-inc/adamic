# no-new-wrappers

Stopped on area `9b7547976bea1b00f8b04508b9102d03511f65b3`.

Required Go call: `ctx.TypeChecker.GetSymbolAtLocation(identifier)` at `cohere/internal/lint/rules/core/no_new_native_nonconstructor.go:103`.

The same missing shared resolvesToAGlobal declaration predicate is called from no_new_wrappers.go:118. The current symbol-origin answer is not that predicate.

Reproducer: `blocked.ts.txt`. This is an analysis input, not a successful parity witness. No descriptor is registered and no upstream parity or mutant result is claimed. Shared checker and helper files remain untouched.
