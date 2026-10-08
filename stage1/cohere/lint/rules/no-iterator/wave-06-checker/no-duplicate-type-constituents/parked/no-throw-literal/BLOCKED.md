# no-throw-literal

Stopped on area `9b7547976bea1b00f8b04508b9102d03511f65b3`.

Required Go call: `ctx.TypeChecker.GetSymbolAtLocation(identifier)` at `cohere/internal/lint/rules/core/no_undef_init.go:163`.

identifierIsShadowed needs every resolved symbol declaration and IsDeclarationFile at lines 167-169. An unresolved symbol and the global undefined can both have no declaration, unlike source shadows. No shared identifierIsShadowed helper is available.

Reproducer: `blocked.ts.txt`. This is an analysis input, not a successful parity witness. No descriptor is registered and no upstream parity or mutant result is claimed. Shared checker and helper files remain untouched.
