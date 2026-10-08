# nexus/performance-no-independent-await-in-loop

Stopped on area `9b7547976bea1b00f8b04508b9102d03511f65b3`.

Required Go call: `checker.SkipAlias(symbol, analysis.ctx.TypeChecker)` at `cohere/internal/lint/rules/nexus/performance_no_independent_await_in_loop.go:479`.

Alias resolution, resolved value-symbol identity and declaration bodies are not checker questions. The rule follows called functions to decide whether an await is ordered.

Reproducer: `blocked.ts.txt`. This is an analysis input, not a successful parity witness. No descriptor is registered and no upstream parity or mutant result is claimed. Shared checker and helper files remain untouched.
