package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/spread_positions.a",
		"stage3/fixtures/objects/03_resolution_cache_spreads.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

// The mutations change valid IR and must finish cleanly: Node's stdout alone catches them.
func TestSpreadPositionsMutants(t *testing.T) {
	t.Parallel()
	for _, mutation := range []string{"overwrite", "key_order", "snapshot"} {
		t.Run(mutation, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/spread_positions.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			for index, function := range program.Functions {
				if mutation == "snapshot" && function.Name == "spreadSnapshot" {
					program.Functions[index].Body[0] = ir.Return{Value: ir.Read{Local: function.Parameters[0], Of: ir.Object}}
					changed = true
					break
				}
				if function.Name != "spreadMerge" {
					continue
				}
				result := function.Body[0].(ir.Return)
				literal := result.Value.(ir.ObjectLiteral)
				if mutation == "key_order" {
					if len(literal.Fields) < 2 {
						continue
					}
					literal.Fields[0], literal.Fields[len(literal.Fields)-1] = literal.Fields[len(literal.Fields)-1], literal.Fields[0]
					changed = true
				} else {
					for fieldIndex, field := range literal.Fields {
						if field.Name != "a" || field.Value.Type() != ir.Number {
							continue
						}
						field.Value = ir.Property{Object: ir.Read{Local: function.Parameters[1], Of: ir.Object}, Name: "a", Of: ir.Number}
						literal.Fields[fieldIndex] = field
						changed = true
					}
				}
				result.Value = literal
				program.Functions[index].Body[0] = result
				if changed {
					break
				}
			}
			if !changed {
				t.Fatal("mutant target absent")
			}
			native, binary := natively(t, program)
			if native.exitCode != 0 || len(native.stderr) != 0 {
				t.Fatalf("mutant must finish cleanly: %d %q", native.exitCode, native.stderr)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatalf("mutant must be leak clean: %s", report)
			}
			if difference := disagreement(onNode(t, path), native); difference != "stdout differs" {
				t.Fatalf("want stdout differs, got %q", difference)
			}
			t.Logf("%s mutant compiled, finished and leaked nothing; Node caught stdout differs", mutation)
		})
	}
}

func TestSpreadHiddenFieldsStayRefused(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"spread_hidden_field.a", "spread_hidden_spread.a"} {
		_, err := lowered(t, filepath.Join(repository, "internal/oracle/testdata/spread_refused", path))
		if err == nil || !strings.Contains(err.Error(), "adamic/single-spread") {
			t.Fatalf("want soundness refusal, got %v", err)
		}
	}
}

// The runtime stores C field names, so a NUL key cannot be confused with its prefix.
func TestSpreadNulNamesStayNotYet(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"spread_nul_source.a", "spread_nul_field.a"} {
		_, err := lowered(t, filepath.Join(repository, "internal/oracle/testdata/spread_refused", name))
		if err == nil || !strings.Contains(err.Error(), "NUL") || !strings.Contains(err.Error(), "adamic/spread-not-first") {
			t.Errorf("%s: want NUL lowering limit, got %v", name, err)
		}
	}
}

func TestSpreadOpaqueRepresentationChangesStayNotYet(t *testing.T) {
	t.Parallel()
	_, err := lowered(t, filepath.Join(repository, "internal/oracle/testdata/spread_refused/spread_opaque_representation.a"))
	if err == nil || !strings.Contains(err.Error(), "slot ownership") || !strings.Contains(err.Error(), "adamic/spread-not-first") {
		t.Fatalf("want slot-ownership lowering limit, got %v", err)
	}
}
