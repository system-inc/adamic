# nexus/correctness-require-child-process-error-listener

Stopped on area `9b7547976bea1b00f8b04508b9102d03511f65b3`.

Required Go call: `ctx.TypeChecker.GetSymbolAtLocation(bound)` at `cohere/internal/lint/rules/nexus/correctness_require_child_process_error_listener.go:135`.

The question surface exposes neither resolved value-symbol identity nor its complete declaration nodes; symbol-origin only returns ValueDeclaration filename. GetShorthandAssignmentValueSymbol and GetExportSpecifierLocalTargetSymbol are also missing.

Reproducer: `blocked.ts.txt`. This is an analysis input, not a successful parity witness. No descriptor is registered and no upstream parity or mutant result is claimed. Shared checker and helper files remain untouched.
