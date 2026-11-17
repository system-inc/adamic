package cloud

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// sandboxed-job.sh keeps the kit's own files outside the tree it lays, so a walker that reads every file under $PWD never
// meets them (internal/acheck read the cache's runtime.a as programs on CLI's J-0pmqbfg; Adamic's rule on #qee4r6g). The
// build cache and the TypeScript copy go to LOOM_JOB_SCRATCH where the runner names one and to TMPDIR where it doesn't,
// and a scratch inside the tree is refused before the command.
//
// Mutants: ADAMIC_BUILD_CACHE_DIR exported under $PWD again (both outside cases red on the cache); the TypeScript copy
// under $PWD again (both red on the copy); the inside-the-tree refusal taken out (the inside case sources clean).
func TestSandboxedJobKeepsTheKitOutsideTheTree(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs("sandboxed-job.sh")
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
			sourced := `. "$SANDBOXED_JOB_SCRIPT" || { echo "sourced=failed"; exit 0; }
echo "sourced=ok"
sandboxed_typescript_source > /dev/null || exit 1
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
			kept = append(kept, "SANDBOXED_JOB_SCRIPT="+script, "TMPDIR="+filepath.Join(outside, "tmp"), "ADAMIC_TYPESCRIPT_SOURCE="+source)
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

// sandboxed_products gives its go test no clock but the job's own ceiling: go's default 10 minutes killed internal/oracle's
// products before they settled on Lunatic's 8 cores (Library's J-1xphvcc, routed by Runtime on #jobs), so a slow box
// read as the head's red. The test sources the script in a temporary tree beside a stand-in go that writes down the
// arguments it was given, and holds that the products' go test asked `-timeout 0`.
//
// Mutant: the `-timeout 0` taken out of sandboxed_products' go test (the stand-in's arguments name no timeout, red here).
func TestSandboxedProductsGivesGoNoClockButTheJobs(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs("sandboxed-job.sh")
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
	sourced := `. "$SANDBOXED_JOB_SCRIPT" || exit 1
sandboxed_typescript_source() { :; }
sandboxed_products ./internal/oracle
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
		"SANDBOXED_JOB_SCRIPT="+script)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("sandboxed_products: %v\n%s", err, output)
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
	t.Fatalf("sandboxed_products' go test named no -timeout 0, so go's own 10 minutes would end a slow settle: %q", arguments)
}

// sandboxed_git leaves a tree that is already a repository as it is and returns 0, so an `&&` chain goes on to its command:
// the runner lays sandboxed trees with a .git since Nov 16, and the refusal it made then cost Emit's jobs a run each
// (#6zvxhjq). A tree with no .git still becomes one commit holding the tree as laid.
//
// Mutant: the old refusal put back (`[ ! -e .git ] || sandboxedJobFail ... || return 1`), red on the repository case.
func TestSandboxedGitLeavesARepositoryAsItIs(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs("sandboxed-job.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, repository := range []bool{true, false} {
		t.Run(map[bool]string{true: "a tree laid with its .git", false: "a tree laid with none"}[repository], func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			tree := filepath.Join(directory, "tree")
			if err := os.MkdirAll(tree, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(tree, "go.mod"), []byte("module example.invalid/tree\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			git := func(arguments ...string) string {
				command := toolchain.Command("git", append([]string{"-C", tree, "-c", "user.name=test", "-c", "user.email=test@loom.invalid", "-c", "commit.gpgsign=false"}, arguments...)...)
				if command.Err != nil {
					t.Fatal(command.Err)
				}
				output, err := command.CombinedOutput()
				if err != nil {
					t.Fatalf("git %v: %v\n%s", arguments, err, output)
				}
				return strings.TrimSpace(string(output))
			}
			head := ""
			if repository {
				git("init", "-q")
				git("add", "-A")
				git("commit", "-q", "-m", "the runner's own commit")
				head = git("rev-parse", "HEAD")
			}
			command := toolchain.Command("bash", "-c", `. "$SANDBOXED_JOB_SCRIPT" || exit 1
sandboxed_git && echo "sandboxed_git=ok"`)
			if command.Err != nil {
				t.Fatal(command.Err)
			}
			command.Dir = tree
			environment := command.Env
			if environment == nil {
				environment = os.Environ()
			}
			kept := make([]string, 0, len(environment)+2)
			for _, variable := range environment {
				name, _, _ := strings.Cut(variable, "=")
				if name != "TMPDIR" && name != "ADAMIC_BUILD_CACHE_DIR" && name != "LOOM_JOB_SCRATCH" {
					kept = append(kept, variable)
				}
			}
			command.Env = append(kept, "TMPDIR="+filepath.Join(directory, "tmp"), "SANDBOXED_JOB_SCRIPT="+script)
			output, err := command.CombinedOutput()
			said := string(output)
			if err != nil || !strings.Contains(said, "sandboxed_git=ok") {
				t.Fatalf("want sandboxed_git to return 0, got %v:\n%s", err, said)
			}
			if repository {
				if now := git("rev-parse", "HEAD"); now != head {
					t.Errorf("the tree's HEAD moved from %s to %s: its history wasn't left alone", head, now)
				}
				if !strings.Contains(said, "already a repository") {
					t.Errorf("want it said that the tree is already a repository, got:\n%s", said)
				}
				return
			}
			if count := git("rev-list", "--count", "HEAD"); count != "1" {
				t.Errorf("want one commit holding the tree as laid, got %s", count)
			}
		})
	}
}
