# Mutants for a bound function that declares more parameters than its bound type names (#04f5s0s)

Each one edit to internal/lower/function_bind.go, run on a house box, then put back. To be measured.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| The reach's narrowing removed: every function of the bound type counted, as -2 did | `if known {` | `if known && false {` | to be measured: TestFunctionBindExtraParameters' known_reach, whose unary reaches only one, stopping on two, held as Unary too, which declares 2 parameters where Unary names 1 |
| The count dropped | `if candidate := l.result.Functions[target]; declaresBeyond(candidate, len(call.Arguments)) {` | `if candidate := l.result.Functions[target]; false && declaresBeyond(candidate, len(call.Arguments)) {` | to be measured: TestFunctionBindExtraParameters' narrow_type and narrow_default lowering and differing from Node, as J-xet5sem measured on main |

The witness Lower asked for first, `const none: ((x: number) => number) | undefined = undefined; none?.bind(undefined)`, a reach known with no function, stops sooner: functionBind refuses an optional-chain bind whose target isn't a property access ("bind through an optional chain", function_bind.go), and a property access projects through its literal, unknown today. So mReach's witness is known_reach, a reach known with one function.

absent_property is a known over-refusal: main's honest `absentProperty.f?.bind(otherReceiver)`, moved out of function_bind.a, stops because its reach is unknown and every () => number function counts. The absent-field reach follow-up lifts it.
