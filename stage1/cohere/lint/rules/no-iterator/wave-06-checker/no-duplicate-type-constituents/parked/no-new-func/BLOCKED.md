# no-new-func

Stopped on area `9b7547976bea1b00f8b04508b9102d03511f65b3`.

Required Go call: `ctx.TypeChecker.GetSymbolAtLocation(identifier)` at `cohere/internal/lint/rules/core/no_new_native_nonconstructor.go:103`.

resolvesToAGlobal needs symbol.Declarations[0] and that source file's IsDeclarationFile at lines 104-108. symbol-origin instead exposes ValueDeclaration filename; type-origin describes the type symbol, not the resolved value symbol. No shared resolvesToAGlobal helper exists.

Reproducer: `blocked.ts.txt`. This is an analysis input, not a successful parity witness. No descriptor is registered and no upstream parity or mutant result is claimed. Shared checker and helper files remain untouched.
