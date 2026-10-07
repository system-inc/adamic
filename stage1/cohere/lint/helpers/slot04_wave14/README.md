# Numeric utility helpers

`mod_float.a` exports IEEE remainder, with JavaScript `%` semantics. `format_number.a` preserves the pinned Go helper's fixed shortest-decimal text, positive 1e21 cutoff, NaN, negative infinity and signed zero. This deliberately preserves Go's `-0` and `-Inf` spellings. `multiple_of.a` checks parse success, nonnegativity, zero remainder and identical original spelling.

isMultipleOf takes the exact separately owned Go-equivalent strconv.ParseFloat callback, returning `{ok,value}`. JavaScript Number acceptance is not substituted for that parser. The oracle supplies actual Go parse results; integration must provide that dependency. No regex matcher is implemented.

The comparator observes modulo through the formatter, including its high-positive cutoff. Full lint findings and exhaustive floating-point bit patterns are not claimed. REPORT.md contains commands, consumer rules, coverage, mutants and the landing rebase evidence.
