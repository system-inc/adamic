package cloud

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// sealed-job.sh keeps the kit's own files outside the tree it lays, so a walker that reads every file under $PWD never
// meets them (internal/acheck read the cache's runtime.a as programs on CLI's J-0pmqbfg; Adamic's rule on #qee4r6g). The
// build cache and the TypeScript copy go to LOOM_JOB_SCRATCH where the runner names one and to TMPDIR where it doesn't,
// and a scratch inside the tree is refused before the command.
//
// Mutants: ADAMIC_BUILD_CACHE_DIR exported under $PWD again (both outside cases red on the cache); the TypeScript copy
// under $PWD again (both red on the copy); the inside-the-tree refusal taken out (the inside case sources clean).
func TestSealedJobKeepsTheKitOutsideTheTree(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs("sealed-job.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name    string
		scratch func(tree, outside string) string
		refused string
	}{
		{"the runner names no scratch", func(tree, outside string) string { return "" }, ""},
		{"the runner names a scratch beside the tree", func(tree, outside string) string { return filepath.Join(outside, "scratch") }, ""},
		{"the runner names a scratch inside the tree", func(tree, outside string) string { return filepath.Join(tree, "scratch") }, "is inside the tree"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			tree, outside, source := filepath.Join(directory, "tree"), filepath.Join(directory, "outside"), filepath.Join(directory, "typescript")
			for _, made := range []string{tree, outside, source} {
				if err := os.Mkdir(made, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(source, "pin"), []byte("x\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			scratch := test.scratch(tree, outside)
			sourced := `. "$SEALED_JOB_SCRIPT" || { echo "sourced=failed"; exit 0; }
echo "sourced=ok"
sealed_typescript_source > /dev/null || exit 1
printf 'cache=%s\ncopy=%s\n' "$ADAMIC_BUILD_CACHE_DIR" "$ADAMIC_TYPESCRIPT_SOURCE"
`
			command := toolchain.Command("bash", "-c", sourced)
			if command.Err != nil {
				t.Fatal(command.Err)
			}
			command.Dir = tree
			environment := command.Env
			if environment == nil {
				environment = os.Environ()
			}
			kept := make([]string, 0, len(environment))
			for _, variable := range environment {
				name, _, _ := strings.Cut(variable, "=")
				if name != "LOOM_JOB_SCRATCH" && name != "TMPDIR" && name != "ADAMIC_BUILD_CACHE_DIR" && name != "ADAMIC_TYPESCRIPT_SOURCE" {
					kept = append(kept, variable)
				}
			}
			kept = append(kept, "SEALED_JOB_SCRIPT="+script, "TMPDIR="+filepath.Join(outside, "tmp"), "ADAMIC_TYPESCRIPT_SOURCE="+source)
			if scratch != "" {
				kept = append(kept, "LOOM_JOB_SCRATCH="+scratch)
			}
			command.Env = kept
			output, err := command.CombinedOutput()
			said := string(output)
			if err != nil {
				t.Fatalf("bash: %v\n%s", err, said)
			}
			if test.refused != "" {
				if !strings.Contains(said, "sourced=failed") || !strings.Contains(said, test.refused) {
					t.Fatalf("want the source refused saying %q, got:\n%s", test.refused, said)
				}
				return
			}
			if !strings.Contains(said, "sourced=ok") {
				t.Fatalf("want the source to pass, got:\n%s", said)
			}
			realTree, err := filepath.EvalSymlinks(tree)
			if err != nil {
				t.Fatal(err)
			}
			for _, key := range []string{"cache", "copy"} {
				value := ""
				for _, line := range strings.Split(said, "\n") {
					if found, ok := strings.CutPrefix(line, key+"="); ok {
						value = found
					}
				}
				if value == "" {
					t.Fatalf("want a %s line, got:\n%s", key, said)
				}
				resolved, err := filepath.EvalSymlinks(value)
				if err != nil {
					t.Fatalf("the %s %s wasn't made: %v", key, value, err)
				}
				if resolved == realTree || strings.HasPrefix(resolved, realTree+string(filepath.Separator)) {
					t.Errorf("the %s %s is inside the tree %s", key, value, tree)
				}
				if scratch != "" && !strings.HasPrefix(value, scratch+string(filepath.Separator)) {
					t.Errorf("the %s %s isn't under the runner's scratch %s", key, value, scratch)
				}
			}
		})
	}
}
