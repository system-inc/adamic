package cloud

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

// sandboxed-job.sh takes the node's laid Go module cache only when loom-node's cache proxy laid this very commit whole
// (#nodecache): a lay file for the runner's repository and commit, naming the laid cache at its one path, whose goSums
// are the tree's go.sum and go.work.sum files exactly, each with its sha256. The lay is found by the name the cache
// gives it, the hex sha256 of the repository, a newline and the commit, in a lays/ of mode 0711, and a file at the old
// per-repository path is never read. Every other case keeps the image's GOMODCACHE and says why,
// a lay that doesn't parse included, and the preamble still returns 0. GOPROXY, GOSUMDB and GOFLAGS are the image's in
// every case. The test sources the script in a temporary tree with two go.sum files (one in a submodule) and a root
// go.work.sum, and decoys under the .git directories the walk skips, beside a temporary lays directory and laid cache.
//
// Mutants: the no-lay branch taking the laid cache (`image('no lay for this commit')` made `sys.exit(print('laid 214'))`,
// red on no lay); the sha256 check taken out (its `if hashlib.sha256(...) != named[path]:` made `if False:`, red on a
// go.sum that differs); the set check taken out (`found ^ set(named)` made `set()`, red on a go.sum the lay doesn't
// name); the walk finding go.sum alone (`name in ('go.sum', 'go.work.sum')` made `name == 'go.sum'`, red on the whole
// lay, whose go.work.sum the tree then seems to lack); the lay looked for at the old per-repository path
// (`os.path.join(lays, repository, commit + '.json')`, red on the whole lay and on the old path's).
func TestSandboxedJobTakesTheNodesLaidModulesOnlyWhole(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs("sandboxed-job.sh")
	if err != nil {
		t.Fatal(err)
	}
	const commit = "0123456789abcdef0123456789abcdef01234567"
	const goFlags = "-buildvcs=false -trimpath -mod=readonly"
	sums := map[string]string{"go.sum": "example.invalid/a v1.0.0 h1:a=\n", "cohere/go.sum": "example.invalid/b v1.0.0 h1:b=\n",
		"go.work.sum": "example.invalid/w v1.0.0 h1:w=\n"}
	for _, test := range []struct {
		name string
		// lay returns the lay file's bytes, or nil for no file; tree changes the tree after its go.sum files are written.
		lay  func(t *testing.T, laid string, goSums []map[string]string) []byte
		tree func(t *testing.T, tree string)
		laid bool
		why  string
	}{
		{"a lay for this commit, whole", nil, nil, true, ""},
		{"no lay for this commit", func(*testing.T, string, []map[string]string) []byte { return nil }, nil, false, "no lay for this commit"},
		{"a lay naming another commit", func(t *testing.T, laid string, goSums []map[string]string) []byte {
			return layOf(t, "adamic", "fedcba9876543210fedcba9876543210fedcba98", laid, goSums)
		}, nil, false, "the lay names another commit"},
		{"a lay naming another repository", func(t *testing.T, laid string, goSums []map[string]string) []byte {
			return layOf(t, "cohere", commit, laid, goSums)
		}, nil, false, "the lay names another repository"},
		{"a lay naming another cache", func(t *testing.T, laid string, goSums []map[string]string) []byte {
			return layOf(t, "adamic", commit, laid+"-elsewhere", goSums)
		}, nil, false, "the lay names the cache"},
		{"a go.sum that differs", nil, func(t *testing.T, tree string) {
			if err := os.WriteFile(filepath.Join(tree, "cohere", "go.sum"), []byte("example.invalid/b v1.0.1 h1:c=\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, false, "cohere/go.sum differs"},
		{"a go.work.sum that differs", nil, func(t *testing.T, tree string) {
			if err := os.WriteFile(filepath.Join(tree, "go.work.sum"), []byte("example.invalid/w v1.0.1 h1:x=\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, false, "go.work.sum differs"},
		{"a go.sum the lay doesn't name", nil, func(t *testing.T, tree string) {
			if err := os.MkdirAll(filepath.Join(tree, "tools"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(tree, "tools", "go.sum"), []byte("example.invalid/c v1.0.0 h1:c=\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, false, "tools/go.sum is in the tree alone"},
		{"a go.work.sum the lay doesn't name", nil, func(t *testing.T, tree string) {
			if err := os.WriteFile(filepath.Join(tree, "cohere", "go.work.sum"), []byte("example.invalid/v v1.0.0 h1:v=\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, false, "cohere/go.work.sum is in the tree alone"},
		{"a go.sum the tree lacks", nil, func(t *testing.T, tree string) {
			if err := os.Remove(filepath.Join(tree, "cohere", "go.sum")); err != nil {
				t.Fatal(err)
			}
		}, false, "cohere/go.sum is in the lay alone"},
		{"a lay that doesn't parse", func(*testing.T, string, []map[string]string) []byte { return []byte(`{"repository":"adamic",`) }, nil, false, "the lay doesn't parse"},
		// A file at the per-repository path the cache wrote before its name moved, lays/adamic/<commit>.json, and none at
		// the name: were the old path read, its why would be the stand-in's, never "no lay"
		{"a lay at the old per-repository path, which the cache no longer writes", func(*testing.T, string, []map[string]string) []byte { return nil }, func(t *testing.T, tree string) {
			old := filepath.Join(filepath.Dir(tree), "lays", "adamic")
			if err := os.Mkdir(old, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(old, commit+".json"), []byte(`{"repository":"adamic",`), 0o444); err != nil {
				t.Fatal(err)
			}
		}, false, "no lay for this commit"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			tree, lays, laid, image := filepath.Join(directory, "tree"), filepath.Join(directory, "lays"), filepath.Join(directory, "mod"), filepath.Join(directory, "image-mod")
			for _, made := range []string{filepath.Join(tree, "cohere"), filepath.Join(tree, ".git", "modules", "cohere"), filepath.Join(tree, "cohere", ".git"),
				laid, image} {
				if err := os.MkdirAll(made, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			// lays/ as the cache makes it, 0711: a job opens its own lay by name and lists nothing
			if err := os.Mkdir(lays, 0o711); err != nil {
				t.Fatal(err)
			}
			nameDigest := sha256.Sum256([]byte("adamic\n" + commit))
			layPath := filepath.Join(lays, hex.EncodeToString(nameDigest[:])+".json")
			goSums := []map[string]string{}
			for _, path := range []string{"cohere/go.sum", "go.sum", "go.work.sum"} {
				if err := os.WriteFile(filepath.Join(tree, path), []byte(sums[path]), 0o644); err != nil {
					t.Fatal(err)
				}
				digest := sha256.Sum256([]byte(sums[path]))
				goSums = append(goSums, map[string]string{"path": path, "sha256": hex.EncodeToString(digest[:])})
			}
			// Under the root's .git and a submodule's, where the walk never looks: go.sum files the lay doesn't name, so
			// reading either would refuse the lay.
			for _, decoy := range []string{filepath.Join(tree, ".git", "modules", "cohere", "go.sum"), filepath.Join(tree, "cohere", ".git", "go.sum")} {
				if err := os.WriteFile(decoy, []byte("not the tree's\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if test.tree != nil {
				test.tree(t, tree)
			}
			lay := layOf(t, "adamic", commit, laid, goSums)
			if test.lay != nil {
				lay = test.lay(t, laid, goSums)
			}
			if lay != nil {
				if err := os.WriteFile(layPath, lay, 0o444); err != nil {
					t.Fatal(err)
				}
			}
			command := toolchain.Command("bash", "-c", `. "$SANDBOXED_JOB_SCRIPT" || { echo "sourced=failed"; exit 0; }
echo "sourced=ok"
printf 'gomodcache=%s\ngoproxy=%s\ngosumdb=%s\ngoflags=%s\n' "$GOMODCACHE" "$GOPROXY" "$GOSUMDB" "$GOFLAGS"
`)
			if command.Err != nil {
				t.Fatal(command.Err)
			}
			command.Dir = tree
			environment := command.Env
			if environment == nil {
				environment = os.Environ()
			}
			kept := make([]string, 0, len(environment)+10)
			for _, variable := range environment {
				name, _, _ := strings.Cut(variable, "=")
				switch name {
				case "TMPDIR", "ADAMIC_BUILD_CACHE_DIR", "LOOM_JOB_SCRATCH", "LOOM_JOB_REPOSITORY", "LOOM_JOB_COMMIT", "SANDBOXED_JOB_LAYS",
					"SANDBOXED_JOB_LAID_GOMODCACHE", "GOMODCACHE", "GOPROXY", "GOSUMDB", "GOFLAGS":
				default:
					kept = append(kept, variable)
				}
			}
			command.Env = append(kept, "SANDBOXED_JOB_SCRIPT="+script, "TMPDIR="+filepath.Join(directory, "tmp"), "LOOM_JOB_REPOSITORY=adamic",
				"LOOM_JOB_COMMIT="+commit, "SANDBOXED_JOB_LAYS="+lays, "SANDBOXED_JOB_LAID_GOMODCACHE="+laid, "GOMODCACHE="+image,
				"GOPROXY=off", "GOSUMDB=off", "GOFLAGS="+goFlags)
			output, err := command.CombinedOutput()
			said := string(output)
			if err != nil || !strings.Contains(said, "sourced=ok") {
				t.Fatalf("want the source to pass, got %v:\n%s", err, said)
			}
			values := map[string]string{}
			for _, line := range strings.Split(said, "\n") {
				if name, value, ok := strings.Cut(line, "="); ok {
					values[name] = value
				}
			}
			for name, want := range map[string]string{"goproxy": "off", "gosumdb": "off", "goflags": goFlags} {
				if values[name] != want {
					t.Errorf("%s is %q, want the image's %q", name, values[name], want)
				}
			}
			if test.laid {
				if values["gomodcache"] != laid {
					t.Errorf("GOMODCACHE is %q, want the node's laid cache %s", values["gomodcache"], laid)
				}
				if want := "sandboxed-job: go modules: the node's laid cache for " + commit[:12] + " (214 modules)"; !strings.Contains(said, want) {
					t.Errorf("want %q said, got:\n%s", want, said)
				}
				return
			}
			if values["gomodcache"] != image {
				t.Errorf("GOMODCACHE is %q, want the image's %s", values["gomodcache"], image)
			}
			if want := "sandboxed-job: go modules: the image's cache (" + test.why; !strings.Contains(said, want) {
				t.Errorf("want %q said, got:\n%s", want, said)
			}
		})
	}
}

// layOf is a lay file as loom-node's cache proxy writes one.
func layOf(t *testing.T, repository, commit, gomodcache string, goSums []map[string]string) []byte {
	t.Helper()
	data, err := json.Marshal(map[string]any{"repository": repository, "commit": commit, "gomodcache": gomodcache, "goSums": goSums,
		"modules": 214, "laidAt": 1763337600000})
	if err != nil {
		t.Fatal(err)
	}
	return data
}
