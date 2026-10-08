# prefer-arrow-callback

Stopped on area `9b7547976bea1b00f8b04508b9102d03511f65b3`.

Required Go call: `ctx.TypeChecker.GetSymbolAtLocation(name)` at `cohere/internal/lint/rules/core/prefer_arrow_callback.go:331`.

Self-reference detection compares resolved value-symbol identity at line 343. Neither value-symbol identity nor declared/implicit arguments declaration count (line 245) is available.

Reproducer: `blocked.ts.txt`. This is an analysis input, not a successful parity witness. No descriptor is registered and no upstream parity or mutant result is claimed. Shared checker and helper files remain untouched.
