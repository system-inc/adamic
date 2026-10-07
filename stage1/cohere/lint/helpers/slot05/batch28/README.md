# Tailwind bare value dispatch, transforms and opacity

One public helper per .a file. bareValuePredicate(kind, predicates) selects the supplied positive, opacity, strict, spacing or font-stretch function. StrictPositiveInteger and GridRepeat share strict. Fraction, empty and unknown kinds have no predicate. Go nil is represented by {present: false, predicate: <placeholder>}: callers must check present before invoking predicate. The placeholder has no semantics when present is false. A tagged presence object is necessary because stage 0 explicitly refuses a function-or-null return. The original refused attempt is preserved in evidence/initial.log.gz.

bareValueTransform(kind, value) returns --spacing(value) for SpacingMultiplier, repeat(value, minmax(0, 1fr)) for GridRepeat, and the untouched input for every other kind. It does not validate value or add a caller's suffix.

isValidOpacityValue(value, multipleOf) delegates exactly once with the original value and divisor 0.25. The dependency owns Go float parsing, nonnegative modulo and canonical number printing. No numeric parser, regex matcher or matching engine is implemented here.

The owned driver receives unchanged Go dependency answers for each exact input. For predicate dispatch it observes selected function identity and the invoked boolean; the Go side uses reflect to identify the actual returned function, rather than reimplementing the dispatch. Opacity's mutated divisors use the actual Go isMultipleOf answers for 0.5 and 1, so mutant mismatches are semantic output differences with successful execution. These helpers are available under explicit dependency contracts; no whole rule is claimed ported. REPORT.md records scope and validation.
