package oracle

import (
	"os"
	"strings"
	"testing"
)

func init() {
	for _, name := range []string{"79", "193", "variants", "required", "propagation"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/optional_after_call_" + name + ".a", true, false})
	}
	// Computed optional indexing is an existing lowering gap, separate from the narrowing check.
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/optional_after_call_gaps/computed.a", false, false})
}

// TestOptionalAfterCallCounts updates this unit and its existing narrowing controls. The area-views base
// has unrelated fixture refusals, so its table-wide count refresh cannot finish.
func TestOptionalAfterCallCounts(t *testing.T) {
	recorded, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(recorded)
	paths := []string{
		"internal/load/testdata/0.1/compile/09_tree.ts",
	}
	for _, name := range []string{"79", "193", "variants", "required", "propagation"} {
		paths = append(paths, "internal/oracle/testdata/optional_after_call_"+name+".a")
	}
	// Required reads now create catchable errors and add exception successors,
	// which can change counts and reuse for the existing narrowing controls.
	for _, name := range []string{
		"narrowed_reads", "narrowed_writes", "narrowed_methods", "narrowed_fields",
		"narrowed_numbers", "narrowed_compared", "reuse_narrowed", "regexp_null_narrowed",
		"047cb0d_n_element_plain", "047cb0d_n_element", "047cb0d_n_element_method",
		"047cb0d_n_param", "047cb0d_n_numparam", "047cb0d_n_coalesce",
		"047cb0d_n_typeof", "047cb0d_n_template", "047cb0d_n_optional",
		"047cb0d_n_arrayindex", "047cb0d_n_paren", "047cb0d_n_conditional",
		"e4eec87_f1_field_narrowed", "e4eec87_f1_class_narrowed",
		"e4eec87_f1_alias_narrowed", "e4eec87_f1_field_present",
	} {
		paths = append(paths, "internal/oracle/testdata/"+name+".a")
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			row := counted(t, path, false, nil, false, false)
			prefix := "| " + path + " |"
			previous := ""
			for _, line := range strings.Split(text, "\n") {
				if strings.HasPrefix(line, prefix) {
					previous = line
					break
				}
			}
			// Older rows omit the two graph counters. Keep their spelling when
			// the measured graph counters are zero, avoiding unrelated churn.
			if len(strings.Split(previous, "|")) == 9 && strings.HasSuffix(row, " | 0 | 0 |") {
				row = strings.TrimSuffix(row, " | 0 | 0 |") + " |"
			}
			if *updateCounts {
				if previous == "" {
					at := strings.Index(text, "\n## ")
					if at < 0 {
						text = strings.TrimRight(text, "\n") + "\n" + row + "\n"
					} else {
						text = text[:at] + "\n" + row + text[at:]
					}
				} else {
					text = strings.Replace(text, previous, row, 1)
				}
			} else if previous != row {
				t.Errorf("recorded: %s\nmeasured: %s", previous, row)
			}
		})
	}
	if *updateCounts && !t.Failed() {
		if err := os.WriteFile(countsPath, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
