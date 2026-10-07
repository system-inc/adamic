package oracle

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/accessors.a",
		"internal/oracle/testdata/accessors_order.a",
		"internal/oracle/testdata/accessors_hierarchy.a",
		"internal/oracle/testdata/accessors_ownership.a",
		"internal/oracle/testdata/accessors_throw.a",
		"internal/oracle/testdata/accessors_analyses.a",
		"internal/oracle/testdata/accessors_coverage_nominal_result.a",
		"internal/oracle/testdata/accessors_coverage_borrow_safe.a",
		"internal/oracle/testdata/accessors_coverage_setter_only_override.a",
		"internal/oracle/testdata/accessors_coverage_operators.a",
		"internal/oracle/testdata/accessors_coverage_scalars.a",
		"internal/oracle/testdata/accessors_coverage_weak.a",
		"internal/oracle/testdata/accessors_coverage_string_snapshot.a",
		"internal/oracle/testdata/accessors_coverage_loop.a",
		"internal/oracle/testdata/accessors_coverage_union_setter.a",
		"internal/oracle/testdata/accessors_coverage_collections.a",
		"internal/oracle/testdata/accessors_coverage_functions.a",
		"internal/oracle/testdata/accessors_coverage_constructor.a",
		"internal/oracle/testdata/accessors_coverage_override_types.a",
		"internal/oracle/testdata/accessors_coverage_generic_inherited.a",
		"internal/oracle/testdata/accessors_coverage_union_narrowed.a",
		"internal/oracle/testdata/accessors_coverage_setter_throw.a",
		"internal/oracle/testdata/accessors_coverage_operand_throw.a",
		"internal/oracle/testdata/accessors_coverage_assignment_order.a",
		"internal/oracle/testdata/accessors_coverage_objects.a",
		"internal/oracle/testdata/accessors_coverage_optional_references.a",
	} {
		// Main has no optional boolean slot and requires matching getter/setter
		// representations. Keep these original probes as explicit NotYet checks.
		lowers := path != "internal/oracle/testdata/accessors_coverage_scalars.a" && path != "internal/oracle/testdata/accessors_coverage_union_setter.a"
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, lowers, false})
	}
}
