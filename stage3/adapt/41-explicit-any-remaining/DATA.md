# Compiler data comparison

compareDataObjects retains a shared caller T. The only nonrecursive compiler caller
compares currentOptions and newOptions, both CompilerOptions; recursive indexed
members preserve the corresponding domain. Two any tokens and one direct blocker
are removed, without a universal replacement alias or any cast. The body is unchanged.

Stock tsc has zero diagnostics. The entire utilities file emits byte-identical
JavaScript. The actual comparer runs on Node across equal/different settings,
nested arrays, ignored functions and its existing null behavior. Reversing the
actual scalar inequality fails those assertions.

Census any 21 -> 20. Restoring the original two any parameter annotations brings
it to21, exactly one extra any-value site. Refused stays5161, NotYet stays1468.
A discarded mutant restored only dst:any while src:T: it produced two TS7053
index diagnostics, so the census skipped its body. It is not claimed to catch
an any blocker. Its checker log is retained; the original-signature mutant is
the successful check. The JSON declaration-owner repair is measured separately.

Commands: stock tsc --noEmit; family-proof.cjs; data-proof.cjs; latent census
and census-report.py on control and restored original signature. Logs and complete
observations are in evidence/data. No boolean-condition logic was adapted.
