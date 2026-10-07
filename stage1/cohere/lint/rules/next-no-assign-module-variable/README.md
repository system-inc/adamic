# @next/next/no-assign-module-variable

The listener checks every VariableStatement's declaration list, reports once on
the whole statement for a plain identifier named module, and preserves Go's
silence on destructuring, parameters, imports and loop declaration lists. There
is no fix or option surface. Source and messages are Adamic .a files.

The upstream oracle is cohere's unchanged next.NoAssignModuleVariable. The
module_name_ignored mutant changes only the identifier match; it compiles and
runs cleanly, then loses the positive finding on all three port executions.

Validation, captured cases, compiler/stage1 comparisons, throughput and limits
are in the continuation report under ../../claims/wave1-15-next-report.md.
Default directory registration requires the scratch compatibility proposal
owned by ../structure-tailwind-no-physical-direction; it is not installed.
