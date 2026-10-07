# Mutants actually run

All fourteen mutants compile with strict C flags, exit 0, and have empty stderr under ASan/UBSan. Each is killed solely by the source-on-Node stdout comparison. Mutations are confined to each test's lowered IR/generated C; protected runtime sources are never edited. Tests also insist the mutation target exists.

| Mutant | Change | Fixture | Catch |
|---|---|---|---|
| boxed_slot | Add 1 to the first argument of each generated radix formatter call | library_number_immediate.a | Node stdout differs; sanitizer clean |
| undefined_digits | Treat absent digits as present only in generated toExponential calls | library_number_optional_formats.a | Node stdout differs; sanitizer clean |
| global_predicate | Replace generated isfinite calls with isnan | library_number_global_predicates.a | Node stdout differs; sanitizer clean |
| metadata_arity | Change generated binary64 2 to 3 | library_math_metadata.a | Node stdout differs; sanitizer clean |
| metadata_type | Change the lowered metadata string object to function | library_math_metadata.a | Node stdout differs; sanitizer clean |
| metadata_name | Change the lowered metadata name max to min | library_math_metadata.a | Node stdout differs; sanitizer clean |
| function_own | Invert generated string equality for the intrinsic own-property helper | library_math_function_own.a | Node stdout differs; sanitizer clean |
| clz32 | Replace generated clz32 calls with fround | library_math_number_math.a | Node stdout differs; sanitizer clean |
| fround | Replace generated fround calls with sign | library_math_number_math.a | Node stdout differs; sanitizer clean |
| imul | Replace generated imul calls with pow | library_math_number_math.a | Node stdout differs; sanitizer clean |
| constants | Change LN10 by one binary64 unit | library_math_number_math.a | Node stdout differs; sanitizer clean |
| conversion | Replace Number string conversion with parseFloat | library_math_number_convert.a | Node stdout differs; sanitizer clean |
| own_property | Invert Number constructor/prototype own-property results | library_math_number_convert.a | Node stdout differs; sanitizer clean |
| prototype | Add 1 to generated Number prototype toFixed receivers | library_math_number_prototype.a | Node stdout differs; sanitizer clean |

Commands and the complete successful transcript are in [report.md](report.md) and [unit-gate.log](unit-gate.log).

Verification corrections: the first global predicate fixture supplied non-number arguments that real tsc rejects, then used an unsupported mixed array/static-predicate union. It was corrected to use explicit Number conversion and a supported union parameter, with static predicates on the converted number. The first undefined-argument mutant also changed toPrecision and triggered a RangeError; it was narrowed to toExponential, whose zero-digit output differs while it still exits 0. These failed verification attempts are not counted as successful mutants.
