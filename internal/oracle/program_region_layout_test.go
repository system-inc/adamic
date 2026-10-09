package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestProgramRegionFieldOnlyAdoptionMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/program_region/optional.a"))
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(programRegionLowered(t, path, true))
	size := regexp.MustCompile(`adamic_object_size\((\w+)->shape->count\)`)
	if len(size.FindAllString(code, -1)) != 1 {
		t.Fatal("one adopted optional object required")
	}
	mutant := size.ReplaceAllString(code, "sizeof *$1 + $1->shape->count * sizeof $1->slots[0]")
	binary := filepath.Join(t.TempDir(), "field-only")
	if err := native.Build(mutant, binary, native.Options{ProgramRegion: true, Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{leakSanitizer()}, binary)
	if result.exitCode == 0 || !strings.Contains(string(result.stderr), "AddressSanitizer: heap-buffer-overflow") {
		t.Fatalf("field-only size mutant escaped: %d %s", result.exitCode, result.stderr)
	}
	if disagreement(onNode(t, path), result) == "" {
		t.Fatal("mutant agreed with Node")
	}
	t.Log("field-only adoption mutant caught by ASan heap-buffer-overflow")
}
