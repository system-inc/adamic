Fixed production-code plan, written before mutant runs. Lower is the only tested entry. P01 is a probe, not a mutant. Twelve mutants chosen across six production files. The whole clean package took 51.080 seconds, so the 20-minute budget limits matrix size below the 20-mutant aim.

M01 internal/lower/census_overload_proof.go:102 change option: !unchanged -> false
M02 internal/lower/census_overload_proof.go:43 return early: return l.censusReturnProof(implementation, overload, nil, nil) -> return false
M03 internal/lower/census_overload_proof.go:27 flip condition: !l.censusHasUndefined(produced) || l.censusHasUndefined(promised) -> !l.censusHasUndefined(produced) && l.censusHasUndefined(promised)
M04 internal/lower/census_small.go:134 flip condition: l.classAssignable(from, to) && l.widened(from, to, map[[2]*checker.Type]bool{}) == nil -> l.classAssignable(from, to) || l.widened(from, to, map[[2]*checker.Type]bool{}) == nil
M05 internal/lower/census_small.go:226 swap two arguments: !l.censusRelated(given, takes) -> !l.censusRelated(takes, given)
M06 internal/lower/census_small.go:145 change constant: ordinal := 0 -> ordinal := 1
M07 internal/lower/census_predicate_marker.go:65 flip condition: safe = safe && accepted -> safe = safe || accepted
M08 internal/lower/predicates.go:472 flip condition: return found && safe -> return found || safe
M09 internal/lower/predicates.go:47 change constant: "there is no body proving this parameter" -> ""
M10 internal/lower/element_access_fields.go:76 flip condition: len(names) == 0 -> len(names) != 0
M11 internal/lower/census_small.go:380 flip condition: left.Type() != ir.Boolean || right.Type() != ir.Boolean -> left.Type() == ir.Boolean || right.Type() == ir.Boolean
M12 internal/lower/invariance.go:188 flip condition: !fresh && !l.checker.IsReadonlySymbol(viewed) && (!l.enumAssignable(target, source) || !l.checker.IsTypeAssignableTo(target, source)) -> fresh && !l.checker.IsReadonlySymbol(viewed) && (!l.enumAssignable(target, source) || !l.checker.IsTypeAssignableTo(target, source))
P01 internal/lower/lower.go:20 empty-answer probe: func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) { -> func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {
	return nil, nil

Validation repairs, no new mutant semantics: M01 standalone condition uses unchanged && false, keeping the local read while disabling the same guard. P01 standalone replaces the complete Lower body and removes its now-unused fmt/filepath imports, avoiding unreachable-code vet errors. The switched variants already executed the identical false guard and empty entry answer.
