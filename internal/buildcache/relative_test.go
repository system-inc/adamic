package buildcache

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A product built here that names this machine fails its build, naming the file and the place (#tqrqx60); one that
// names the tree, the cache and other products by Relative's names builds, and Absolute reads them back.
// Not parallel: cached calls t.Setenv for ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG and ADAMIC_BUILD_CACHE.
func TestAProductNamingItsMachineFailsItsBuild(t *testing.T) {
	cache, _ := cached(t)
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	typescript := filepath.Join(t.TempDir(), "typescript")
	if err = os.Mkdir(typescript, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_TYPESCRIPT_SOURCE", typescript)
	inputs := func(name string) Inputs { in := thisPackage; in.Name = "relocatable " + name; return in }
	writes := func(content func(directory string) string) func(directory string) error {
		return func(directory string) error {
			return os.WriteFile(filepath.Join(directory, "manifest.json"), []byte(content(directory)), 0o644)
		}
	}
	for _, refused := range []struct {
		name, place string
		content     func(directory string) string
	}{
		{"a tree file", "the repository", func(string) string { return `{"go.mod":"` + filepath.Join(root, "go.mod") + `"}` }},
		{"another product", "the build cache", func(string) string { return filepath.Join(cache, strings.Repeat("ab", 32), "main.ts") }},
		{"its own directory", "the build cache", func(directory string) string { return "overlay " + filepath.Join(directory, "main.go") }},
		{"a tool in the home directory", "the home directory", func(string) string { return "clang=" + filepath.Join(home, "adamic-tools/llvm/bin/clang") }},
		{"a gate input", "ADAMIC_TYPESCRIPT_SOURCE", func(string) string { return filepath.Join(typescript, "src/compiler/checker.ts") + "\tchecker" }},
	} {
		_, err := Get(inputs(refused.name), writes(refused.content))
		if err == nil || !strings.Contains(err.Error(), "manifest.json names") || !strings.Contains(err.Error(), refused.place) {
			t.Errorf("%s: the build gave %v, want a refusal naming manifest.json and %s", refused.name, err, refused.place)
		}
	}
	entries, _ := os.ReadDir(cache)
	for _, entry := range entries {
		if entry.IsDir() {
			t.Fatalf("a refused product left %s in the cache", entry.Name())
		}
	}

	// The same names, written relative to the roots a reader supplies, build; Absolute reads each back here.
	other := filepath.Join(cache, strings.Repeat("ab", 32), "main.ts")
	var own string
	product := Product(t, inputs("relative"), writes(func(directory string) string {
		own = Relative(filepath.Join(directory, "main.go"))
		return Relative(filepath.Join(root, "go.mod")) + "\n" + Relative(other) + "\n" + own + "\n" + Relative(filepath.Join(typescript, "src/compiler/checker.ts"))
	}))
	content, err := os.ReadFile(filepath.Join(product, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(content), "\n")
	key := filepath.Base(product)
	if want := []string{"<repository>/go.mod", "<build cache>/" + strings.Repeat("ab", 32) + "/main.ts", "<build cache>/" + key + "/main.go", "<ADAMIC_TYPESCRIPT_SOURCE>/src/compiler/checker.ts"}; strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Fatalf("Relative wrote %q, want %q", lines, want)
	}
	for index, want := range []string{filepath.Join(root, "go.mod"), other, filepath.Join(product, "main.go"), filepath.Join(typescript, "src/compiler/checker.ts")} {
		if got := Absolute(lines[index]); got != want {
			t.Errorf("Absolute(%q) is %q, want %q", lines[index], got, want)
		}
	}

	// A directory that isn't a product is read as it is, even when its text says <repository> (a package whose comments
	// name it), and nothing is written beside it.
	if directory, err := Resolved(filepath.Join(root, "internal/buildcache")); err != nil || directory != filepath.Join(root, "internal/buildcache") {
		t.Fatalf("this package resolved to %s, %v", directory, err)
	}
	if copies, _ := filepath.Glob(filepath.Join(root, "internal", ".buildcache.resolved-*")); len(copies) != 0 {
		t.Fatalf("a copy was written beside a package: %v", copies)
	}

	// A binary's paths are debug information, and a macOS .dSYM bundle's too: neither is refused.
	Product(t, inputs("binary"), func(directory string) error {
		if err := os.MkdirAll(filepath.Join(directory, "port.dSYM"), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, "port.dSYM", "port.yml"), []byte(root), 0o644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "port"), []byte("\x7fELF\x00"+root), 0o755)
	})

	// Uncached, a product lives in a temporary directory of its own: another names it <build cache>/<key> and finds it.
	t.Setenv("ADAMIC_BUILD_CACHE", "off")
	first := Product(t, inputs("uncached"), writes(func(string) string { return "first" }))
	second := Product(t, inputs("names the uncached one"), writes(func(string) string { return Relative(filepath.Join(first, "manifest.json")) }))
	named, err := os.ReadFile(filepath.Join(second, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(named), "<build cache>/") || Absolute(string(named)) != filepath.Join(first, "manifest.json") {
		t.Fatalf("the uncached product is named %q, read back as %q, want %s", named, Absolute(string(named)), filepath.Join(first, "manifest.json"))
	}
	if _, err := Get(inputs("names the uncached one raw"), writes(func(string) string { return filepath.Join(first, "manifest.json") })); err == nil {
		t.Fatal("an uncached product naming another's temporary directory was accepted")
	}
}

// One product, built at one path and read at another (#tqrqx60): Workshop builds a tree at A into its cache, a runner
// holds the same tree at B and Workshop's cache at another place, and A and Workshop's cache are gone. The runner hits
// the product by key and reads the tree file its manifest names. A manifest holding A's absolute path, as products did
// before relocatable, is put in the runner's cache directly (no build here would accept it) and fails to read.
func TestAProductReadsTheSameUnderAnotherPath(t *testing.T) {
	t.Parallel()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	workshop, runner := t.TempDir(), t.TempDir()
	tree, cache := filepath.Join(workshop, "work", "adamic"), filepath.Join(workshop, "products")
	write(t, tree, "go.mod", "module github.com/system-inc/adamic\n")
	write(t, tree, "data/answer.txt", "forty-two\n")
	child := func(tree, cache, mode string) (string, error) {
		command := exec.Command(executable, "-test.run=^TestRelocationChild$", "-test.count=1")
		command.Dir = tree
		command.Env = append(os.Environ(), "ADAMIC_BUILDCACHE_RELOCATION="+mode, "ADAMIC_BUILD_CACHE_DIR="+cache, "ADAMIC_BUILD_CACHE=",
			"ADAMIC_BUILD_STORE=off", "ADAMIC_BUILD_LOG="+filepath.Join(cache, "builds.log"))
		output, err := command.CombinedOutput()
		return string(output), err
	}
	output, err := child(tree, cache, "build")
	if err != nil || !strings.Contains(output, "read forty-two") || !strings.Contains(output, "port resolved") {
		t.Fatalf("building at %s: %v\n%s", tree, err, output)
	}
	if output, err = child(tree, cache, "build absolute"); err == nil || !strings.Contains(output, "holds a path of the machine that built it") {
		t.Fatalf("a product naming %s built: %v\n%s", tree, err, output)
	}

	moved, movedCache := filepath.Join(runner, "srv", "7", "adamic"), filepath.Join(runner, "cache")
	if err = os.MkdirAll(filepath.Dir(moved), 0o755); err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(tree, moved); err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(cache, movedCache); err != nil {
		t.Fatal(err)
	}
	if output, err = child(moved, movedCache, "read"); err != nil || !strings.Contains(output, "read forty-two") || !strings.Contains(output, "port resolved") {
		t.Fatalf("reading at %s: %v\n%s", moved, err, output)
	}
	if log, _ := os.ReadFile(filepath.Join(movedCache, "builds.log")); !strings.Contains(string(log), " hit ") || strings.Count(string(log), " miss ") != 1 {
		t.Fatalf("the runner rebuilt instead of hitting Workshop's product:\n%s", log)
	}

	// Before relocatable: the same product holding Workshop's path, under the key the runner asks for.
	entries, err := os.ReadDir(movedCache)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() && len(entry.Name()) == 64 {
			if err = os.WriteFile(filepath.Join(movedCache, entry.Name(), "manifest.txt"), []byte(filepath.Join(tree, "data/answer.txt")), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if output, err = child(moved, movedCache, "read"); err == nil || !strings.Contains(output, "no such file") {
		t.Fatalf("a product holding Workshop's path read on the runner: %v\n%s", err, output)
	}
}

// The child of TestAProductReadsTheSameUnderAnotherPath: a process whose repository is the tree it runs in.
func TestRelocationChild(t *testing.T) {
	mode := os.Getenv("ADAMIC_BUILDCACHE_RELOCATION")
	if mode == "" {
		t.Skip("run by TestAProductReadsTheSameUnderAnotherPath")
	}
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	inputs := Inputs{Name: "relocation", Files: []string{"data"}}
	if mode == "build absolute" {
		inputs.Name += " absolute"
	}
	product := Product(t, inputs, func(directory string) error {
		if mode == "read" {
			return fmt.Errorf("the runner built %s instead of finding it", inputs.Name)
		}
		answer := filepath.Join(root, "data/answer.txt")
		if mode == "build" {
			answer = Relative(answer)
		}
		if err := os.WriteFile(filepath.Join(directory, "manifest.txt"), []byte(answer), 0o644); err != nil {
			return err
		}
		// A mutant port whose import reaches back into the tree, written as the build's own steps need it, then made
		// the product's.
		if err := os.MkdirAll(filepath.Join(directory, "port"), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, "port/main.ts"), []byte("import answer from '"+filepath.Join(root, "data/answer.txt")+"';\n"), 0o644); err != nil {
			return err
		}
		if mode == "build" {
			return RelativeFiles(directory)
		}
		return nil
	})
	manifest, err := os.ReadFile(filepath.Join(product, "manifest.txt"))
	if err != nil {
		t.Fatal(err)
	}
	answer, err := os.ReadFile(Absolute(string(manifest)))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("read %s", answer)
	stored, err := os.ReadFile(filepath.Join(product, "port/main.ts"))
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := Resolved(product)
	if err != nil {
		t.Fatal(err)
	}
	port, err := os.ReadFile(filepath.Join(resolved, "port/main.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if want := "import answer from '" + filepath.Join(root, "data/answer.txt") + "';\n"; string(port) != want || resolved == product || !strings.Contains(string(stored), "'<repository>/data/answer.txt'") {
		t.Fatalf("the product's port reads %q, resolved here %q in %s, want %q", stored, port, resolved, want)
	}
	fmt.Println("port resolved")
}
