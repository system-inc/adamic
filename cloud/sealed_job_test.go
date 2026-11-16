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

// sealed_products gives its go test no clock but the job's own ceiling: go's default 10 minutes killed internal/oracle's
// products before they settled on Lunatic's 8 cores (Library's J-1xphvcc, routed by Runtime on #jobs), so a slow box
// read as the head's red. The test sources the script in a temporary tree beside a stand-in go that writes down the
// arguments it was given, and holds that the products' go test asked `-timeout 0`.
//
// Mutant: the `-timeout 0` taken out of sealed_products' go test (the stand-in's arguments name no timeout, red here).
func TestSealedProductsGivesGoNoClockButTheJobs(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs("sealed-job.sh")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	tree, bin := filepath.Join(directory, "tree"), filepath.Join(directory, "bin")
	for _, made := range []string{filepath.Join(tree, "cloud", "fast-gate"), filepath.Join(tree, "internal", "oracle"), bin} {
		if err := os.MkdirAll(made, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(tree, "cloud", "fast-gate", "compiler-dependencies.json"), []byte(`{"packages":{}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	said := filepath.Join(directory, "go-arguments")
	stand := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$GO_ARGUMENTS\"\n"
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(stand), 0o755); err != nil {
		t.Fatal(err)
	}
	// The TypeScript copy is the image's road, not this test's: it stands in as done.
	sourced := `. "$SEALED_JOB_SCRIPT" || exit 1
sealed_typescript_source() { :; }
sealed_products ./internal/oracle
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
	path := os.Getenv("PATH")
	kept := make([]string, 0, len(environment)+4)
	for _, variable := range environment {
		name, value, _ := strings.Cut(variable, "=")
		switch name {
		case "PATH":
			path = value
		case "TMPDIR", "ADAMIC_BUILD_CACHE_DIR", "LOOM_JOB_SCRATCH":
		default:
			kept = append(kept, variable)
		}
	}
	command.Env = append(kept, "PATH="+bin+":"+path, "TMPDIR="+filepath.Join(directory, "tmp"), "GO_ARGUMENTS="+said,
		"SEALED_JOB_SCRIPT="+script)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("sealed_products: %v\n%s", err, output)
	}
	data, err := os.ReadFile(said)
	if err != nil {
		t.Fatalf("the stand-in go was never run: %v", err)
	}
	arguments := strings.Split(strings.TrimSpace(string(data)), "\n")
	for index, argument := range arguments {
		if argument == "-timeout" && index+1 < len(arguments) && arguments[index+1] == "0" {
			return
		}
		if argument == "-timeout=0" {
			return
		}
	}
	t.Fatalf("sealed_products' go test named no -timeout 0, so go's own 10 minutes would end a slow settle: %q", arguments)
}
