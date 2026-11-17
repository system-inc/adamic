# Overload arity and value uses — partial findings

Baseline: `49f1d643064a9231f7f343d8bbd02ba718866e6a`. Tasks #5fws6q8 and #ev14ch9. Report only; no compiler changes.

Code evidence: `internal/lower/expression.go:2135` defines `overloadSetValueStop`, which refuses “an overloaded function as a value”. `functionValue` calls it at line 2161; `nestedReference` calls it in `internal/lower/nested_functions.go:240`. This applies independently of the road that accepted the declaration.

`censusOverload` uses the overload parameter list, treating missing parameters as undefined (`internal/lower/census_small.go:202`). Body proof repeats that premise (`internal/lower/census_overload_proof.go:78`). `overloadResultCheck` omits a call check when that declaration-level body proof succeeds (`internal/lower/overload_result_check.go:147`). This is not a proof using each call’s actual arguments.

Native compiler build and the program matrix are in progress. No runtime conclusion yet.
