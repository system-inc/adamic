package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/e4eec87_u02_optional_absent.a",
		"internal/oracle/testdata/literal_optional_shapes.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

// The old required-field lookup still compiles, but panics instead of returning typed undefined.
func TestLiteralOptionalOracleCatchesMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/e4eec87_u02_optional_absent.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	source := native.C(program)
	mutant := strings.ReplaceAll(source, "adamic_object_optional_field(", "adamic_object_field(")
	if mutant == source {
		test.Fatal("mutant target absent")
	}
	binary := filepath.Join(test.TempDir(), "mutant")
	if buildError := native.Build(mutant, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	result := execute(test, binary)
	if result.exitCode != 70 ||
		!strings.Contains(string(result.stderr), "a field the checker proved is there is missing") {
		test.Fatalf("want restored missing-field panic: exit %d, stderr %q", result.exitCode, result.stderr)
	}
	if disagreement(onNode(test, path), result) == "" {
		test.Fatal("Node did not catch mutant")
	}
	test.Log("Node caught restored required-field lookup: missing-field panic, exit 70")
}
