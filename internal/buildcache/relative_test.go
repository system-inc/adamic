package buildcache

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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
		{"its own directory", "a product directory", func(directory string) string { return "overlay " + filepath.Join(directory, "main.go") }},
		{"a tool in the home directory", "the home directory", func(string) string { return "clang=" + filepath.Join(home, "adamic-tools/llvm/bin/clang") }},
		{"a gate input", "ADAMIC_TYPESCRIPT_SOURCE", func(string) string { return filepath.Join(typescript, "src/compiler/checker.ts") + "\tchecker" }},
	} {
		_, err := Get(inputs(refused.name), writes(refused.content))
		if err == nil || !strings.Contains(err.Error(), "manifest.json names") || !strings.Contains(err.Error(), refused.place) {
			t.Errorf("%s: the build gave %v, want a refusal naming manifest.json and %s", refused.name, err, refused.place)
		}
	}
	// A product built for this machine alone lands under local/, so nothing a build made may stay anywhere in the cache.
	filepath.WalkDir(cache, func(path string, entry fs.DirEntry, err error) error {
		if err == nil && entry.IsDir() && path != cache && entry.Name() != "local" {
			t.Fatalf("a refused product left %s in the cache", path)
		}
		return nil
	})

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
	// A product names itself by its name key, the same wherever its directory is (relative.go).
	nameKey := NameKey(root, inputs("relative"))
	if want := []string{"<repository>/go.mod", "<build cache>/" + strings.Repeat("ab", 32) + "/main.ts", "<build cache>/" + nameKey + "/main.go", "<ADAMIC_TYPESCRIPT_SOURCE>/src/compiler/checker.ts"}; strings.Join(lines, "\n") != strings.Join(want, "\n") {
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

	// Uncached, a product lives in a temporary directory of its own: another names it <build cache>/<name key> and finds it.
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
	if err != nil || !strings.Contains(output, "read forty-two") || !strings.Contains(output, "port resolved") || !strings.Contains(output, "dependent found") {
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
	if output, err = child(moved, movedCache, "read"); err != nil || !strings.Contains(output, "read forty-two") || !strings.Contains(output, "port resolved") || !strings.Contains(output, "dependent found") {
		t.Fatalf("reading at %s: %v\n%s", moved, err, output)
	}
	// Workshop's two builds, the product and the one keyed by its resolved copy, and only hits on the runner.
	if log, _ := os.ReadFile(filepath.Join(movedCache, "builds.log")); !strings.Contains(string(log), " hit ") || strings.Count(string(log), " miss ") != 2 {
		t.Fatalf("the runner rebuilt instead of hitting Workshop's product:\n%s", log)
	}

	// Before relocatable: the same product holding Workshop's path, under the key the runner asks for (built untraced,
	// it lives under local/).
	planted := 0
	for _, directory := range []string{movedCache, filepath.Join(movedCache, "local")} {
		entries, err := os.ReadDir(directory)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() && len(entry.Name()) == 64 {
				if err = os.WriteFile(filepath.Join(directory, entry.Name(), "manifest.txt"), []byte(filepath.Join(tree, "data/answer.txt")), 0o644); err != nil {
					t.Fatal(err)
				}
				planted++
			}
		}
	}
	if planted == 0 {
		t.Fatal("no product in the runner's cache to plant Workshop's path in")
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

	// A product keyed by a file of the resolved copy, as estree's misc port is keyed by its main.ts: the copy's name holds
	// a hash of this machine's roots, and the key names the product's file instead, so the runner hits Workshop's
	// (#sgemgmn).
	Product(t, Inputs{Name: "relocation dependent", Files: []string{"data"}, Flags: []string{"main=" + filepath.Join(resolved, "port/main.ts")}}, func(directory string) error {
		if mode == "read" {
			return fmt.Errorf("the runner built the product keyed by %s instead of finding it", filepath.Join(resolved, "port/main.ts"))
		}
		return os.WriteFile(filepath.Join(directory, "built"), []byte("built\n"), 0o644)
	})
	fmt.Println("dependent found")
}

// A key names a product by its key wherever this process holds it (#sgemgmn): in the cache, in the copy Resolved made
// of it beside the cache's, whose name holds a hash of this machine's roots, and built uncached in a temporary
// directory of its own, which main's uncached gate publishes under the key it computes.
// Not parallel: cached calls t.Setenv for ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG and ADAMIC_BUILD_CACHE.
func TestAKeyNamesAProductByItsKeyWhereverItIs(t *testing.T) {
	cached(t)
	t.Setenv("TMPDIR", t.TempDir())
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	build := func(directory string) error {
		return os.WriteFile(filepath.Join(directory, "main.ts"), []byte("import '"+Relative(filepath.Join(root, "go.mod"))+"';\n"), 0o644)
	}
	for _, mode := range []string{"", "off"} {
		t.Setenv("ADAMIC_BUILD_CACHE", mode)
		inputs := thisPackage
		inputs.Name = "named by its key, cache " + mode
		product := Product(t, inputs, build)
		key, err := Key(root, inputs)
		if err != nil {
			t.Fatal(err)
		}
		copied, err := Resolved(product)
		if err != nil {
			t.Fatal(err)
		}
		if copied == product {
			t.Fatalf("cache %q: %s, which names <repository>, resolved to itself", mode, product)
		}
		want := "main=<build cache>/" + key + "/main.ts"
		for _, held := range []string{product, copied} {
			if got := portable(root, "main="+filepath.Join(held, "main.ts")); got != want {
				t.Errorf("cache %q: a key names %s as %q, want %q", mode, filepath.Join(held, "main.ts"), got, want)
			}
		}
	}
}

// forgetHeld drops what this process holds by name key, as a process that never got these products starts.
func forgetHeld(t *testing.T) {
	t.Helper()
	clear := func() {
		productDirectories.Range(func(key, _ any) bool { productDirectories.Delete(key); return true })
		productKeys.Range(func(key, _ any) bool { productKeys.Delete(key); return true })
		resolvedDirectories.Range(func(key, _ any) bool { resolvedDirectories.Delete(key); return true })
	}
	clear()
	t.Cleanup(clear)
}

// One product names itself by its name key, and the name reads back wherever the reader's cache holds it (#tqrqx60 with
// #vt46geg): while its traced build is pending, once settled under its read set's key (what Loom's runner unpacks, the
// key's directory and the name key's read sets, read with no network), and built for this machine alone under local/.
// Each read is from a process holding none of them. A name whose read set no longer values to a cached product reads
// as <cache>/<name key>, which holds nothing.
// Not parallel: the rig changes the working directory, repositoryRoot and the build environment.
func TestANameKeyReadsBackWhereverItsProductLives(t *testing.T) {
	r := newRig(t)
	forgetHeld(t)
	var own string
	pending := Product(t, port, func(directory string) error {
		own = Relative(filepath.Join(directory, "product"))
		return os.WriteFile(filepath.Join(directory, "product"), []byte(own), 0o644)
	})
	nameKey := NameKey(r.root, port)
	if own != "<build cache>/"+nameKey+"/product" {
		t.Fatalf("a building product named itself %q, want its name key %s", own, nameKey)
	}
	forgetHeld(t)
	if got := Absolute(own); got != filepath.Join(pending, "product") {
		t.Fatalf("pending, the name read back as %s, want %s", got, filepath.Join(pending, "product"))
	}
	// Under the trace, the directory a name reads back to, and a resolved copy of it, are journaled as that product's,
	// so a build reading there reads the product (trace.go).
	copied, err := Resolved(pending)
	if err != nil || copied == pending {
		t.Fatalf("the pending product resolved to %s, %v; want a copy", copied, err)
	}
	entries, err := readJournal(r.trace)
	if err != nil {
		t.Fatal(err)
	}
	for _, directory := range []string{pending, copied} {
		if !slices.ContainsFunc(entries, func(entry journalEntry) bool {
			return entry.Event == "found" && entry.NameKey == nameKey && entry.Directory == directory
		}) {
			t.Fatalf("%s isn't journaled as found under %s", directory, nameKey[:12])
		}
	}

	build := r.built(t)["port"]
	r.window(build.Build, func() { r.open(r.pid, "source/a.c") })
	if settlement := r.settle(t); len(settlement.Settled) != 1 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	file, err := loadReads(r.cache, nameKey)
	if err != nil || len(file.Sets) != 1 {
		t.Fatalf("read sets: %+v, %v", file, err)
	}
	settled := filepath.Join(r.cache, file.Sets[0].Key)
	t.Setenv("ADAMIC_BUILD_TRACE", "")
	forgetHeld(t)
	if got := Absolute(own); got != filepath.Join(settled, "product") {
		t.Fatalf("settled, the name read back as %s, want %s", got, filepath.Join(settled, "product"))
	}
	if content, _ := os.ReadFile(filepath.Join(settled, "product")); string(content) != own {
		t.Fatalf("the settled product holds %q, want its own name %q, never its key", content, own)
	}

	// A read changed: no set values to a cached product, so the name finds nothing rather than a stale product.
	write(t, r.root, "source/a.c", "int a2;\n")
	forgetTracked()
	forgetHeld(t)
	if got := Absolute(own); got != filepath.Join(r.cache, nameKey, "product") {
		t.Fatalf("after a read changed, the name read back as %s, want the empty %s", got, filepath.Join(r.cache, nameKey, "product"))
	}

	// Built for this machine alone: found by the inputs its name key points to, keyed on the tree at hand.
	local := Inputs{Name: "local port", Files: []string{"source"}}
	product := Product(t, local, func(directory string) error {
		return os.WriteFile(filepath.Join(directory, "product"), []byte(Relative(filepath.Join(r.root, "go.mod"))), 0o644)
	})
	if filepath.Dir(product) != filepath.Join(r.cache, "local") {
		t.Fatalf("an untraced product is at %s, want under local/", product)
	}
	forgetHeld(t)
	localName := "<build cache>/" + NameKey(r.root, local) + "/product"
	if got := Absolute(localName); got != filepath.Join(product, "product") {
		t.Fatalf("local, the name read back as %s, want %s", got, filepath.Join(product, "product"))
	}
	forgetHeld(t)
	if copied, err := Resolved(product); err != nil || copied == product {
		t.Fatalf("a local product another process didn't build resolved to %s, %v; want a copy", copied, err)
	}
	write(t, r.root, "source/b.c", "int b2;\n")
	forgetTracked()
	forgetHeld(t)
	if got := Absolute(localName); got == filepath.Join(product, "product") {
		t.Fatalf("after a declared file changed, the name still read back as the old local product %s", got)
	}
}

// A traced build is scanned like any other: one naming the machine that built it is refused, and nothing is placed.
// Not parallel: the rig changes the working directory, repositoryRoot and the build environment.
func TestATracedProductNamingItsMachineIsRefused(t *testing.T) {
	r := newRig(t)
	forgetHeld(t)
	_, err := Get(port, func(directory string) error {
		return os.WriteFile(filepath.Join(directory, "product"), []byte(filepath.Join(r.root, "go.mod")), 0o644)
	})
	if err == nil || !strings.Contains(err.Error(), "holds a path of the machine that built it") {
		t.Fatalf("a traced product naming the repository built: %v", err)
	}
	// filled leaves only its lock file: no product directory, no scratch one.
	entries, _ := os.ReadDir(pendingDirectory(r.cache))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lock") {
			t.Fatalf("a refused traced product left %s pending", entry.Name())
		}
	}
	if len(r.built(t)) != 0 {
		t.Fatalf("a refused traced product was journaled as built: %v", r.built(t))
	}
}
