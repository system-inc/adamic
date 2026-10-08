package oracle

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Mutate a private translation unit, never the repository runtime. Every copied
// exported symbol gets its own name so the normal archive remains unchanged.
func TestFSFileAreaArrayHolesNodeMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_array_holes_callbacks.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/array_holes.c"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(contents)
	before := `hole_text = ADAMIC_STRING("")`
	if strings.Count(source, before) != 1 {
		t.Fatal("join mutation anchor moved")
	}
	source = strings.Replace(source, before, `hole_text = ADAMIC_STRING("undefined")`, 1)
	names := regexp.MustCompile(`\badamic_array_(holes(?:_\w+)?|is_range_error)\b`)
	rename := func(s string) string { return names.ReplaceAllString(s, "fs_area_mutant_$1") }
	source = rename(source) + "\n" + rename(native.C(program))
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"}, binary)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("mutant failed before Node comparison: %+v", actual)
	}
	if difference := disagreement(onNode(t, path), actual); difference != "stdout differs" {
		t.Fatalf("Node must catch stdout only, got %q", difference)
	}
	t.Log("hole-as-undefined join mutant compiles, exits cleanly and is caught only by Node stdout")
}
