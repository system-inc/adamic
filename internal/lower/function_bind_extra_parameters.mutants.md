# Mutants for a bound function that declares more parameters than its bound type names (#04f5s0s)

One edit to internal/lower/function_bind.go, run on a house box, then put back. To be measured.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| declaresBeyond's call dropped from checkBoundArguments | `} else if found == nil && declaresBeyond(candidate, len(call.Arguments)) {` | `} else if false && found == nil && declaresBeyond(candidate, len(call.Arguments)) {` | to be measured: internal/oracle's TestFunctionBindExtraParameters, narrow_type and narrow_default lowering and differing from Node, as J-xet5sem measured on main |
