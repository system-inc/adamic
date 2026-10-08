# react/jsx-no-undef

Stopped on area `9b7547976bea1b00f8b04508b9102d03511f65b3`.

Required Go call: `ctx.TypeChecker.GetSymbolAtLocation(reference)` at `cohere/internal/lint/rules/react/jsx_no_undef.go:101`.

The rule needs resolved symbol presence even without ValueDeclaration and all declaration filenames for jsxNoUndefDeclaredInFile at lines 128-132. symbol-origin collapses missing symbols and symbols lacking ValueDeclaration to the same empty filename.

Reproducer: `blocked.tsx.txt`. This is an analysis input, not a successful parity witness. No descriptor is registered and no upstream parity or mutant result is claimed. Shared checker and helper files remain untouched.
