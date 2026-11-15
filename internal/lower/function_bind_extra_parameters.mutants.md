# Mutants for a bound function that declares more parameters than its bound type names (#04f5s0s)

Each one edit to internal/lower/function_bind.go, run on a house box, then put back. Measured at -7 (dd0e2e7e), Loom job J-ar0w3ws.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| The reach's narrowing removed: every function of the bound type counted, as -2 did | `if known {` | `if known && false {` | TestFunctionBindExtraParameters' known_reach alone, whose unary reaches only one: it stops, "bind of a function that declares 2 parameters where the type it's bound as names 1", on two, held as Unary too. TestFunctionBind and every other row stay green |
| The count dropped | `if candidate := l.result.Functions[target]; declaresBeyond(candidate, len(call.Arguments)) {` | `if candidate := l.result.Functions[target]; false && declaresBeyond(candidate, len(call.Arguments)) {` | TestFunctionBindExtraParameters' narrow_type, narrow_default and the four property fixtures, each lowering where it should stop: narrow_default prints "seen at -1" twice where Node prints the index on sanitized native and the JavaScript backend, and narrow_type, whose line fell past the job's print limit, is what J-xet5sem measured on main, "seen at undefined" twice; and the property fixtures lower and agree with Node, the over-refusals the reach leaves. own_type, narrow_direct and known_reach stay green |

The witness Lower asked for first, `const none: ((x: number) => number) | undefined = undefined; none?.bind(undefined)`, a reach known with no function, stops sooner: functionBind refuses an optional-chain bind whose target isn't a property access ("bind through an optional chain", function_bind.go), and a property access projects through its literal, unknown today. So mReach's witness is known_reach, a reach known with one function.

absent_property, present_property, absent_fixed_property and null_property are known over-refusals: function_bind.a's honest optional-property binds, moved out one to a fixture, each stopping because its field read's reach is unknown and every function of the bound type counts. #x909e6q lifts them.
