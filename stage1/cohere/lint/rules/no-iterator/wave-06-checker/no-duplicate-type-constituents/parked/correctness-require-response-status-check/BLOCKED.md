# nexus/correctness-require-response-status-check

Stopped on area `9b7547976bea1b00f8b04508b9102d03511f65b3`.

Required Go call: `ctx.TypeChecker.GetSymbolAtLocation(name)` at `cohere/internal/lint/rules/nexus/correctness_require_response_status_check.go:278`.

Global-fetch classification needs all resolved symbol declarations and their ambient/global containers (lines 290-313), not ValueDeclaration filename. The shared CFG builder is also absent.

Reproducer: `blocked.ts.txt`. This is an analysis input, not a successful parity witness. No descriptor is registered and no upstream parity or mutant result is claimed. Shared checker and helper files remain untouched.
