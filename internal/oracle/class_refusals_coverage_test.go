package oracle

import (
	"path/filepath"
	"testing"
)

// These controls use the shared source, backend, release, sanitizer and leak oracle.
func init() {
	for _, path := range []string{
		"internal/oracle/testdata/class_refusals_safe_super_no_override.a",
		"internal/oracle/testdata/class_refusals_safe_super_ordered.a",
		"internal/oracle/testdata/class_refusals_safe_super_ordered_getter.a",
		"internal/oracle/testdata/class_refusals_safe_super_inherited_reset.a",
		"internal/oracle/testdata/class_refusals_safe_iterator_base.a",
		"internal/oracle/testdata/class_refusals_safe_iterator_unchanged.a",
		"internal/oracle/testdata/class_refusals_safe_iterator_closed_parameter.a",
		"internal/oracle/testdata/class_refusals_safe_iterator_factory.a",
		"internal/oracle/testdata/class_refusals_safe_keys_plain.a",
		"internal/oracle/testdata/class_refusals_safe_keys_copy.a",
		"internal/oracle/testdata/class_refusals_safe_keys_class.a",
		"internal/oracle/testdata/class_refusals_safe_private_instance.a",
		"internal/oracle/testdata/class_refusals_safe_private_static.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}

// Node records the behavior independently of the refusal analysis in internal/lower.
func TestClassRefusalsWitnessesOnNode(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, output string }{
		{"super_parentheses", "score NaN\n"},
		{"super_optional", "score NaN\n"},
		{"super_arrow_call", "score NaN\n"},
		{"super_getter_override", "score NaN\n"},
		{"super_nested", "score NaN\n"},
		{"super_generic", "score NaN\n"},
		{"super_interface", "score NaN\n"},
		{"super_union", "score 10\n"},
		{"iterator_for_of", "0\nscaled:return\n"},
		{"iterator_spread", "0,100,200\n"},
		{"iterator_from", "0,100,200\n"},
		{"iterator_destructure", "scaled:return\n0\n"},
		{"iterator_alias", "0,100,200\n"},
		{"iterator_arrow", "0,100,200\n"},
		{"iterator_getter", "0,100,200\n"},
		{"iterator_generic", "0,100,200\n"},
		{"iterator_interface", "0,100,200\n"},
		{"iterator_union", "0,100,200\n"},
		{"iterator_optional", "0,100,200\n"},
		{"iterator_return_alias", "0,100,200\n"},
		{"keys_interface", "value\n"},
		{"keys_arrow", "value\n"},
		{"keys_generic", "value\n"},
		{"keys_getter", "value\n"},
		{"keys_union", "value\n"},
		{"keys_optional", "value\n"},
		{"keys_nested", "value\n"},
		{"private_arrow", "s1\n"},
		{"private_getter", "s1\n"},
		{"private_generic", "s1\n"},
		{"private_union", "s1\n"},
		{"private_nested", "s1\n"},
		{"private_write", "t2\n"},
		{"keys_optional_call", "value\n"},
		{"keys_nested_arrow", "value\n"},
		{"private_optional_holder", "s1\n"},
		{"private_interface", "s1\n"},
		{"iterator_optional_chain", "0,100,200\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/class_refusals_refused", probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			observed := onNode(t, path)
			if observed.exitCode != 0 || string(observed.stdout) != probe.output || len(observed.stderr) != 0 {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", observed.exitCode, observed.stdout, observed.stderr)
			}
		})
	}
}
