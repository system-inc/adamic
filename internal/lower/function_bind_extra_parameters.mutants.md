# Mutants for a bound function that declares more parameters than its bound type names (#04f5s0s)

One edit to internal/lower/function_bind.go, run on a house box, then put back. To be measured.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| declaresBeyond's call dropped from checkBoundArguments | `candidate.RestElement != 0 \|\| declaresBeyond(candidate, len(call.Arguments)) {` | `candidate.RestElement != 0 {` | to be measured: internal/oracle's TestFunctionBindExtraParameters, narrow_type and narrow_default lowering and differing from Node, as J-xet5sem measured on main |
