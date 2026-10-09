package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/e4eec87_u02_optional_absent.a",
		"internal/oracle/testdata/literal_optional_shapes.a",
		"internal/oracle/testdata/optional_field_write.a",
		"internal/oracle/testdata/optional_field_presence.a",
		"internal/oracle/testdata/optional_field_construction.a",
		"internal/oracle/testdata/optional_field_unknown.a",
		"internal/oracle/testdata/optional_field_alias.a",
		"internal/oracle/testdata/optional_field_alias_variants.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

// The old required-field lookup still compiles, but panics instead of returning typed undefined.
func TestLiteralOptionalOracleCatchesMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/e4eec87_u02_optional_absent.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	// Exercise the runtime lookup on a valid unreserved layout. Contextual source
	// literals now reserve these slots, so required lookup alone is harmless there.
	for index, statement := range program.Main {
		declaration, ok := statement.(ir.Declare)
		if !ok {
			continue
		}
		literal, ok := declaration.Value.(ir.ObjectLiteral)
		if !ok {
			continue
		}
		literal.Missing = nil
		declaration.Value = literal
		program.Main[index] = declaration
	}
	source := native.C(program)
	baseline := filepath.Join(t.TempDir(), "unreserved")
	if err := native.Build(source, baseline, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(onNode(t, path), execute(t, baseline)); difference != "" {
		t.Fatal(difference)
	}
	mutant := strings.ReplaceAll(source, "adamic_object_optional_field(", "adamic_object_field(")
	if mutant == source {
		t.Fatal("mutant target absent")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := execute(t, binary)
	if result.exitCode != 70 || !strings.Contains(string(result.stderr), "a field the checker proved is there is missing") {
		t.Fatalf("want restored missing-field panic: exit %d, stderr %q", result.exitCode, result.stderr)
	}
	if disagreement(onNode(t, path), result) == "" {
		t.Fatal("Node did not catch mutant")
	}
	t.Log("Node caught restored required-field lookup: missing-field panic, exit 70")
}
