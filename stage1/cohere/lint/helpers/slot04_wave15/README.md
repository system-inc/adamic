# Integer and spacing helpers

One claimed helper per `.a` file. `positive_integer.a` exports isPositiveInteger: accepts zero, rejects empty input, non-ASCII digits and leading zeros, and delegates values longer than fifteen digits to roundTripsAsJavaScriptNumber. The single module-level `/[^0-9]/u` literal checks for nondigits, including line terminators and Unicode. No hand-rolled matcher is used. The Go helper is a byte loop, so the shared Go regex translation table has no row for it.

`round_trip.a` preserves Go float parse success, the 1e21 cutoff, and exact fixed-decimal reprinting. `spacing_multiplier.a` delegates to the verified isMultipleOf with divisor 0.25. Both use the existing slot04_wave14 numeric modules. Callers supply the exact Go-equivalent strconv.ParseFloat callback, returning `{ok,value}`; its implementation is an integration dependency. The oracle supplies actual Go parser results. Go's existing signed-zero behavior is preserved.

These helpers do not produce findings or offsets. No rule, registry, shared harness, or compiler files are changed. REPORT.md contains coverage, mutants, commands, consumers and limits.
