# Mutants for a bound function that declares more parameters than its bound type names (#04f5s0s)

Each one edit to internal/lower/function_bind.go, run on a house box, then put back. To be measured.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| The reach's narrowing removed: every function of the bound type counted, as -2 did | `if known {` | `if known && false {` | to be measured: internal/oracle's TestFunctionBind, function_bind.a:46's honest `absentProperty.f?.bind(otherReceiver)` stopping again, as J-r4xdn3h measured at -2 |
| The count dropped | `if candidate := l.result.Functions[target]; declaresBeyond(candidate, len(call.Arguments)) {` | `if candidate := l.result.Functions[target]; false && declaresBeyond(candidate, len(call.Arguments)) {` | to be measured: TestFunctionBindExtraParameters' narrow_type and narrow_default lowering and differing from Node, as J-xet5sem measured on main |
