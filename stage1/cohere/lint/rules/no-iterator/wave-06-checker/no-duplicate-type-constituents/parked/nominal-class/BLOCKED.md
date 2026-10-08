# adamic/nominal-class

Stopped on area `9b7547976bea1b00f8b04508b9102d03511f65b3`.

Required Go call: `flow.WalkerFor(ctx)` at `cohere/internal/lint/rules/adamic/nominal_class.go:68`.

The shared flow.Walker and flow.Pair/Site traversal are absent from RuleContext; the type graph also lacks ObjectFlags and symbol identity needed by nominalVerdict.

Reproducer: `blocked.ts.txt`. This is an analysis input, not a successful parity witness. No descriptor is registered and no upstream parity or mutant result is claimed. Shared checker and helper files remain untouched.
