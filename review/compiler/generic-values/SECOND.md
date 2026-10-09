Built immutable generic declaration aliases toward step 16, including direct calls at multiple types and concrete escapes.
Commit: this commit, after 000991f1 on compiler/generic-values.
Checks: alias/default oracles pass 1.024s; alias leaf 0.50s, escape mutant 0.24s, early-read leaf 0.62s; nested lowering passes 0.378s; reader guard passes 39.364s; counts refresh passes 115.877s.
Mutants: a planted higher-rank escape is refused with a slot fix; removing alias readiness changes early/4 to 4/4 and is rejected by Node stdout in both backends.
Uncovered: nested generic callable escapes retain explicit NotYet; higher-rank slot binder diagnostics and the final admission/census proof follow in the next unit.

A const alias with no type annotation binds to the original generic declaration, including chains of aliases. Register aliases before hoisted bodies are lowered. Each direct call selects its own checker-resolved instance; passing, returning or storing the alias uses contextual value instantiation. Mutable aliases do not get this exception, and an explicitly annotated polymorphic slot is not treated as an alias.

The alias binding still has readiness-only storage. It never stores an erased generic entry. Reads check that storage before evaluating call arguments or an escaping instance, preserving the temporal dead zone and lexical capture behavior. The planted source prints 4 x, exercises number/string concrete escapes, passing to a concrete parameter and a nested declaration alias. The early-read control compares unchanged source Node with both backends and sanitizers/leaks.

Commands, bounded and logged:
- go test ./internal/oracle -run '^TestGenericValue' -timeout 90s -v
- go test ./internal/lower -run '^TestNestedFunction' -timeout 90s
- go test ./internal/ir -run '^TestCallTargetReaders$' -timeout 90s
- go test ./internal/oracle -run '^TestCountsAreRecorded$' -timeout 210s -args -update-counts
- go test -overlay=/tmp/generic-values-alias-mutant.json ./internal/oracle -run '^TestGenericValueAliasEarlyRead$' -timeout 90s -v (expected failure)

The first push succeeded. Integration lane checks for it passed gofmt/tools/parallel/.a checks in 22.5s but skipped vet after their internal limit; explicit go vet over lower, ir, javascript, native and oracle completed successfully. The final report will record the second lane check and push result.
