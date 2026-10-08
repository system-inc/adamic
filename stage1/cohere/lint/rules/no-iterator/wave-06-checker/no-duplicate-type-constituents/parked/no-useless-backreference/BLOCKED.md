# no-useless-backreference

Stopped on area `9b7547976bea1b00f8b04508b9102d03511f65b3`.

Required Go call: `reference.NewTracker(ctx.SourceFile, ctx.TypeChecker, nil)` at `cohere/internal/lint/rules/core/no_useless_backreference.go:100`.

The shared reference tracker and ConstantStringIn (line 118) are absent. A private tracker or regex syntax scanner would violate the requested shared-helper contract.

Reproducer: `blocked.ts.txt`. This is an analysis input, not a successful parity witness. No descriptor is registered and no upstream parity or mutant result is claimed. Shared checker and helper files remain untouched.
