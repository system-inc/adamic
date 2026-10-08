# prefer-rest-params

Stopped before registration. Required call: `ctx.TypeChecker.GetSymbolAtLocation(identifier)` at `cohere/internal/lint/rules/core/prefer_rest_params.go:114`, followed by the nil-symbol distinction and `len(symbol.Declarations) == 0` at line 118.

The area bridge has no node-symbol presence/declaration-count question. Its `symbol-origin` answer at `bridge/tsgo/checker/facts.go:206` returns only the value declaration's file and returns the same empty string for an unresolved name and a declaration-free implicit arguments symbol. `declarations` accepts only class/interface declarations. `binding-origin` exists on the old wave-20 branch but is not on the area base; it was not imported into a new port. The provisional listener and descriptor were removed before testing or committing.

Reproducer pair:

```ts
arguments; // unresolved, no finding
function f() { arguments; } // resolved implicit symbol, finding
```

The original symbol identity/declaration list is also needed to reject a parameter, import, catch binding, or hoisted var named arguments. No syntactic scope approximation or private checker was added. No new upstream parity or mutant pass is claimed.
