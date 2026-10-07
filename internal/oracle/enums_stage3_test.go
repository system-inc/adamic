package oracle

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, name := range []string{"01_token_range.a", "02_node_range.a", "03_jsdoc_range.a", "04_parse_tree_mask.a", "05_local_export_flags.a", "09_symbol_exclusion_mask.a", "10_string_enum.a", "12_diagnostic_reverse_lookup.a", "13_regex_map_keys.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"stage3/fixtures/enums/" + name, true, false})
	}
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/enums_names.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{
		"stage3/fixtures/enums/07_regex_array_map.a", true, true,
	})
}

// Keep the original upstream contracts intact. These boundaries require source adaptation,
// rather than turning off the checker or treating mutable enum containers as covariant.
func TestStage3EnumBoundaries(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, rule string
		checker    bool
	}{
		{"06_set_node_flags.a", "invariant-mutable", false},
		{"08_debug_format_enum.a", "TS2532", true},
		{"11_format_syntax_kind.a", "TS2532", true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/enums", probe.name))
			if err != nil {
				t.Fatal(err)
			}
			program, loadErr := load.Load([]string{path})
			err = loadErr
			if err == nil {
				_, err = lower.Lower(context.Background(), program)
			}
			var refused *lower.Refused
			var checked *load.CheckError
			if (probe.checker && !errors.As(err, &checked)) || (!probe.checker && !errors.As(err, &refused)) || err == nil || !strings.Contains(err.Error(), probe.rule) {
				t.Fatalf("want boundary %s, got %v", probe.rule, err)
			}
		})
	}
}

func TestStage3EnumSparseArrayBoundary(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/enums/07_regex_array_map.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := run{stderr: []byte("adamic: panic: index 1 is outside an array of length 0\n"), exitCode: 70}
	native, _ := natively(t, program)
	for name, result := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, result); difference != "" {
			t.Errorf("%s: %s; %+v", name, difference, result)
		}
	}
	if node := onNode(t, path); node.exitCode != 0 {
		t.Fatalf("Node must grow its sparse array: %+v", node)
	}
}

// Integer reverse keys stay sorted, while forward member names retain declaration order.
// This mutant finishes normally; only the independent source observation catches its ordering.
func TestEnumNameEnumerationMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/enums_names.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for _, statement := range program.Main {
		declaration, ok := statement.(ir.Declare)
		if !ok || program.Locals[declaration.Local].Name != "Names" {
			continue
		}
		literal, ok := declaration.Value.(ir.ObjectLiteral)
		if !ok {
			continue
		}
		first, second := -1, -1
		for index, field := range literal.Fields {
			if field.Name == "Zero" {
				first = index
			}
			if field.Name == "One" {
				second = index
			}
		}
		if first >= 0 && second >= 0 {
			literal.Fields[first], literal.Fields[second] = literal.Fields[second], literal.Fields[first]
			changed = true
		}
	}
	if !changed {
		t.Fatal("mutant changed no enum fields")
	}
	mutant, _ := natively(t, program)
	if mutant.exitCode != 0 || len(mutant.stderr) != 0 {
		t.Fatalf("mutant must finish cleanly: %+v", mutant)
	}
	if difference := disagreement(onNode(t, path), mutant); difference != "stdout differs" {
		t.Fatalf("want Node stdout catch, got %q", difference)
	}
}
