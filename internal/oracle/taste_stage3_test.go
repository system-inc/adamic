package oracle

// Preserve the real source fixtures. Unsupported entries remain in the manifest
// and observation report; these entries must match source Node in both backends.
func init() {
	for _, name := range []string{
		"01_diagnostic_code.a", "03_diagnostic_file.a", "04_jsx_runtime.a",
		"05_this_type.a", "07_compact.a", "09_literal_cache.a",
		"10_global_import_meta.a", "12_symbol_links.a",
		"13_void_callback.a", "14_relative_complement.a", "15_barrel.a", "16_scan_exclamation.a", "17_binder_flow.a",
		"18_named_export.a", "19_deprecated_flags.a", "20_jsdoc_terminate.a",
		"21_truthy_loops.a", "22_assignment_once.a", "23_labels_finally.a", "24_proportional_conditions.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"stage3/fixtures/taste/" + name, true, false})
	}
}
