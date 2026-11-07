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
	"time"
)

// A product built here that names this machine fails its build, naming the file and the place (#tqrqx60); one that
// names the tree, the cache and other products by Relative's names builds, and Absolute reads them back.
// Not parallel: cached calls test.Setenv for ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG and ADAMIC_BUILD_CACHE.
func TestAProductNamingItsMachineFailsItsBuild(test *testing.T) {
	cache, _ := cached(test)
	root, repositoryRootError := repositoryRoot()
	if repositoryRootError != nil {
		test.Fatal(repositoryRootError)
	}
	home, homeDirectoryError := os.UserHomeDir()
	if homeDirectoryError != nil {
		test.Fatal(homeDirectoryError)
	}
	typescript := filepath.Join(test.TempDir(), "typescript")
	if createDirectoryError := os.Mkdir(typescript, 0o755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	test.Setenv("ADAMIC_TYPESCRIPT_SOURCE", typescript)
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
		{
			"a tree file",
			"the repository",
			func(string) string { return `{"go.mod":"` + filepath.Join(root, "go.mod") + `"}` },
		},
		{
			"another product",
			"the build cache",
			func(string) string { return filepath.Join(cache, strings.Repeat("ab", 32), "main.ts") },
		},
		{
			"its own directory",
			"a product directory",
			func(directory string) string { return "overlay " + filepath.Join(directory, "main.go") },
		},
		{
			"a tool in the home directory",
			"the home directory",
			func(string) string { return "clang=" + filepath.Join(home, "adamic-tools/llvm/bin/clang") },
		},
		{
			"a gate input",
			"ADAMIC_TYPESCRIPT_SOURCE",
			func(string) string { return filepath.Join(typescript, "src/compiler/checker.ts") + "\tchecker" },
		},
	} {
		_, getError := Get(inputs(refused.name), writes(refused.content))
		if getError == nil ||
			!strings.Contains(getError.Error(), "manifest.json names") ||
			!strings.Contains(getError.Error(), refused.place) {
			test.Errorf(
				"%s: the build gave %v, want a refusal naming manifest.json and %s",
				refused.name,
				getError,
				refused.place,
			)
		}
	}
	// A gob is data a test reads back, not a binary's debug information: a path in one is refused as text's is, zero
	// bytes and all (landable-13b: estree's program.gob named its random build directory). Mutant: gob skipped as
	// binary.
	_, getError := Get(inputs("a gob"), func(directory string) error {
		return os.WriteFile(
			filepath.Join(directory, "program.gob"),
			append([]byte{0x1a, 0, 0xff}, "m"+filepath.Join(directory, "source", "main.ts")+":3:7"...),
			0o644,
		)
	})
	if getError == nil ||
		!strings.Contains(getError.Error(), "program.gob names") ||
		!strings.Contains(getError.Error(), "a product directory") {
		test.Errorf("a gob naming its own directory: the build gave %v, want a refusal naming program.gob", getError)
	}
	// A product built for this machine alone lands under local/, so nothing a build made may stay anywhere in the
	// cache.
	filepath.WalkDir(cache, func(path string, entry fs.DirEntry, pathError error) error {
		if pathError == nil && entry.IsDir() && path != cache && entry.Name() != "local" {
			test.Fatalf("a refused product left %s in the cache", path)
		}
		return nil
	})

	// The same names, written relative to the roots a reader supplies, build; Absolute reads each back here.
	other := filepath.Join(cache, strings.Repeat("ab", 32), "main.ts")
	var own string
	product := Product(test, inputs("relative"), writes(func(directory string) string {
		own = Relative(filepath.Join(directory, "main.go"))
		return Relative(filepath.Join(root, "go.mod")) +
			"\n" +
			Relative(other) +
			"\n" +
			own +
			"\n" +
			Relative(filepath.Join(typescript, "src/compiler/checker.ts"))
	}))
	content, readError := os.ReadFile(filepath.Join(product, "manifest.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	lines := strings.Split(string(content), "\n")
	// A product names itself by its name key, the same wherever its directory is (relative.go).
	nameKey := NameKey(root, inputs("relative"))
	if want := []string{
		"<repository>/go.mod",
		"<build cache>/" + strings.Repeat("ab", 32) + "/main.ts",
		"<build cache>/" + nameKey + "/main.go",
		"<ADAMIC_TYPESCRIPT_SOURCE>/src/compiler/checker.ts",
	}; strings.Join(lines, "\n") != strings.Join(want, "\n") {
		test.Fatalf("Relative wrote %q, want %q", lines, want)
	}
	for index, want := range []string{
		filepath.Join(root, "go.mod"),
		other,
		filepath.Join(product, "main.go"),
		filepath.Join(typescript, "src/compiler/checker.ts"),
	} {
		if got := Absolute(lines[index]); got != want {
			test.Errorf("Absolute(%q) is %q, want %q", lines[index], got, want)
		}
	}

	// A directory that isn't a product is read as it is, even when its text says <repository> (a package whose comments
	// name it), and nothing is written beside it.
	if directory, resolvedError := Resolved(
		filepath.Join(root, "internal/buildcache"),
	); resolvedError != nil || directory != filepath.Join(root, "internal/buildcache") {
		test.Fatalf("this package resolved to %s, %v", directory, resolvedError)
	}
	if copies, _ := filepath.Glob(filepath.Join(root, "internal", ".buildcache.resolved-*")); len(copies) != 0 {
		test.Fatalf("a copy was written beside a package: %v", copies)
	}

	// A binary's paths are debug information, and a macOS .dSYM bundle's too: neither is refused.
	Product(test, inputs("binary"), func(directory string) error {
		if createDirectoryError := os.MkdirAll(
			filepath.Join(directory, "port.dSYM"),
			0o755,
		); createDirectoryError != nil {
			return createDirectoryError
		}
		if writeError := os.WriteFile(
			filepath.Join(directory, "port.dSYM", "port.yml"),
			[]byte(root),
			0o644,
		); writeError != nil {
			return writeError
		}
		return os.WriteFile(filepath.Join(directory, "port"), []byte("\x7fELF\x00"+root), 0o755)
	})

	// Uncached, a product lives in a temporary directory of its own: another names it <build cache>/<name key> and
	// finds it.
	test.Setenv("ADAMIC_BUILD_CACHE", "off")
	first := Product(test, inputs("uncached"), writes(func(string) string { return "first" }))
	second := Product(
		test,
		inputs("names the uncached one"),
		writes(func(string) string { return Relative(filepath.Join(first, "manifest.json")) }),
	)
	named, readError := os.ReadFile(filepath.Join(second, "manifest.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	if !strings.HasPrefix(string(named), "<build cache>/") ||
		Absolute(string(named)) != filepath.Join(first, "manifest.json") {
		test.Fatalf(
			"the uncached product is named %q, read back as %q, want %s",
			named,
			Absolute(string(named)),
			filepath.Join(first, "manifest.json"),
		)
	}
	if _, getError := Get(
		inputs("names the uncached one raw"),
		writes(func(string) string { return filepath.Join(first, "manifest.json") }),
	); getError == nil {
		test.Fatal("an uncached product naming another's temporary directory was accepted")
	}
}

// One product, built at one path and read at another (#tqrqx60): Workshop builds a tree at A into its cache, a runner
// holds the same tree at B and Workshop's cache at another place, and A and Workshop's cache are gone. The runner hits
// the product by key and reads the tree file its manifest names. A manifest holding A's absolute path, as products did
// before relocatable, is put in the runner's cache directly (no build here would accept it) and fails to read.
func TestAProductReadsTheSameUnderAnotherPath(test *testing.T) {
	test.Parallel()
	executable, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	workshop, runner := test.TempDir(), test.TempDir()
	tree, cache := filepath.Join(workshop, "work", "adamic"), filepath.Join(workshop, "products")
	write(test, tree, "go.mod", "module github.com/system-inc/adamic\n")
	write(test, tree, "data/answer.txt", "forty-two\n")
	child := func(tree, cache, mode string) (string, error) {
		command := exec.Command(executable, "-test.run=^TestRelocationChild$", "-test.count=1")
		command.Dir = tree
		command.Env = append(
			os.Environ(),
			"ADAMIC_BUILDCACHE_RELOCATION="+mode,
			"ADAMIC_BUILD_CACHE_DIR="+cache,
			"ADAMIC_BUILD_CACHE=",
			"ADAMIC_BUILD_STORE=off",
			"ADAMIC_BUILD_LOG="+filepath.Join(cache, "builds.log"),
		)
		output, combinedOutputError := command.CombinedOutput()
		return string(output), combinedOutputError
	}
	output, childError := child(tree, cache, "build")
	if childError != nil ||
		!strings.Contains(output, "read forty-two") ||
		!strings.Contains(output, "port resolved") ||
		!strings.Contains(output, "inner read") ||
		!strings.Contains(output, "dependent found") {
		test.Fatalf("building at %s: %v\n%s", tree, childError, output)
	}
	if output, childError = child(
		tree,
		cache,
		"build absolute",
	); childError == nil || !strings.Contains(output, "holds a path of the machine that built it") {
		test.Fatalf("a product naming %s built: %v\n%s", tree, childError, output)
	}

	moved, movedCache := filepath.Join(runner, "srv", "7", "adamic"), filepath.Join(runner, "cache")
	if createDirectoryError := os.MkdirAll(filepath.Dir(moved), 0o755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	if renameError := os.Rename(tree, moved); renameError != nil {
		test.Fatal(renameError)
	}
	if renameError := os.Rename(cache, movedCache); renameError != nil {
		test.Fatal(renameError)
	}
	if output, childError = child(moved, movedCache, "read"); childError != nil ||
		!strings.Contains(output, "read forty-two") ||
		!strings.Contains(output, "port resolved") ||
		!strings.Contains(output, "inner read") ||
		!strings.Contains(output, "dependent found") {
		test.Fatalf("reading at %s: %v\n%s", moved, childError, output)
	}
	// Three misses, all at Workshop: the product, the inner one it names and the one keyed by its resolved copy.
	if log, _ := os.ReadFile(
		filepath.Join(movedCache, "builds.log"),
	); !strings.Contains(string(log), " hit ") || strings.Count(string(log), " miss ") != 3 {
		test.Fatalf("the runner rebuilt instead of hitting Workshop's product:\n%s", log)
	}

	// Before relocatable: the same product holding Workshop's path, under the key the runner asks for (built untraced,
	// it lives under local/).
	planted := 0
	for _, directory := range []string{movedCache, filepath.Join(movedCache, "local")} {
		entries, readDirectoryError := os.ReadDir(directory)
		if readDirectoryError != nil {
			test.Fatal(readDirectoryError)
		}
		for _, entry := range entries {
			if entry.IsDir() && len(entry.Name()) == 64 {
				if writeError := os.WriteFile(
					filepath.Join(directory, entry.Name(), "manifest.txt"),
					[]byte(filepath.Join(tree, "data/answer.txt")),
					0o644,
				); writeError != nil {
					test.Fatal(writeError)
				}
				planted++
			}
		}
	}
	if planted == 0 {
		test.Fatal("no product in the runner's cache to plant Workshop's path in")
	}
	if output, childError = child(
		moved,
		movedCache,
		"read",
	); childError == nil || !strings.Contains(output, "no such file") {
		test.Fatalf("a product holding Workshop's path read on the runner: %v\n%s", childError, output)
	}
}

// The child of TestAProductReadsTheSameUnderAnotherPath: a process whose repository is the tree it runs in.
// Not parallel: the child process TestAProductReadsTheSameUnderAnotherPath starts, alone in its process with its own
// tree and cache.
func TestRelocationChild(test *testing.T) {
	mode := os.Getenv("ADAMIC_BUILDCACHE_RELOCATION")
	if mode == "" {
		// The child half of TestAProductReadsTheSameUnderAnotherPath, which runs it with the mode set. Run on its own
		// it has nothing to check, so it returns rather than skipping: no skip for the census to account for.
		return
	}
	root, repositoryRootError := repositoryRoot()
	if repositoryRootError != nil {
		test.Fatal(repositoryRootError)
	}
	inputs := Inputs{Name: "relocation", Files: []string{"data"}}
	if mode == "build absolute" {
		inputs.Name += " absolute"
	}
	// An inner product the outer one names, keyed by a flag naming this checkout (as lowering a port is, by
	// repository=). The runner never gets it itself: it reads it only through the name in the outer product's bytes.
	inner := Inputs{Name: "relocation inner", Files: []string{"data"}, Flags: []string{"repository=" + root}}
	var innerDirectory string
	if mode != "read" {
		innerDirectory = Product(test, inner, func(directory string) error {
			return os.WriteFile(filepath.Join(directory, "inner.txt"), []byte("inner"), 0o644)
		})
	}
	product := Product(test, inputs, func(directory string) error {
		if mode == "read" {
			return fmt.Errorf("the runner built %s instead of finding it", inputs.Name)
		}
		answer := filepath.Join(root, "data/answer.txt")
		if mode == "build" {
			answer = Relative(answer)
		}
		if writeError := os.WriteFile(
			filepath.Join(directory, "manifest.txt"),
			[]byte(answer),
			0o644,
		); writeError != nil {
			return writeError
		}
		if writeError := os.WriteFile(
			filepath.Join(directory, "inner.name"),
			[]byte(Relative(filepath.Join(innerDirectory, "inner.txt"))),
			0o644,
		); writeError != nil {
			return writeError
		}
		// A mutant port whose import reaches back into the tree, written as the build's own steps need it, then made
		// the product's.
		if createDirectoryError := os.MkdirAll(filepath.Join(directory, "port"), 0o755); createDirectoryError != nil {
			return createDirectoryError
		}
		if writeError := os.WriteFile(
			filepath.Join(directory, "port/main.ts"),
			[]byte("import answer from '"+filepath.Join(root, "data/answer.txt")+"';\n"),
			0o644,
		); writeError != nil {
			return writeError
		}
		if mode == "build" {
			return RelativeFiles(directory)
		}
		return nil
	})
	manifest, readError := os.ReadFile(filepath.Join(product, "manifest.txt"))
	if readError != nil {
		test.Fatal(readError)
	}
	answer, readError := os.ReadFile(Absolute(string(manifest)))
	if readError != nil {
		test.Fatal(readError)
	}
	fmt.Printf("read %s", answer)
	stored, readError := os.ReadFile(filepath.Join(product, "port/main.ts"))
	if readError != nil {
		test.Fatal(readError)
	}
	resolved, resolvedError := Resolved(product)
	if resolvedError != nil {
		test.Fatal(resolvedError)
	}
	port, readError := os.ReadFile(filepath.Join(resolved, "port/main.ts"))
	if readError != nil {
		test.Fatal(readError)
	}
	if want := "import answer from '" + filepath.Join(root, "data/answer.txt") + "';\n"; string(port) != want ||
		resolved == product ||
		!strings.Contains(string(stored), "'<repository>/data/answer.txt'") {
		test.Fatalf("the product's port reads %q, resolved here %q in %s, want %q", stored, port, resolved, want)
	}
	fmt.Println("port resolved")
	name, readError := os.ReadFile(filepath.Join(product, "inner.name"))
	if readError != nil {
		test.Fatal(readError)
	}
	if content, readError := os.ReadFile(Absolute(string(name))); readError != nil || string(content) != "inner" {
		test.Fatalf("the inner product named %q read as %q, %v", name, content, readError)
	}
	fmt.Println("inner read")
	// A product keyed by a file of the resolved copy, as estree's misc port is keyed by its main.ts: the copy's name
	// holds a hash of this machine's roots, and the key names the product's file instead, so the runner hits Workshop's
	// (#sgemgmn).
	Product(test, Inputs{
		Name:  "relocation dependent",
		Files: []string{"data"},
		Flags: []string{"main=" + filepath.Join(resolved, "port/main.ts")},
	}, func(directory string) error {
		if mode == "read" {
			return fmt.Errorf(
				"the runner built the product keyed by %s instead of finding it",
				filepath.Join(resolved, "port/main.ts"),
			)
		}
		return os.WriteFile(filepath.Join(directory, "built"), []byte("built\n"), 0o644)
	})
	fmt.Println("dependent found")
}

// A key names a product by its name key wherever this process holds it (#sgemgmn, in walk-5b's one naming model): in
// the cache, in the copy Resolved made of it beside the cache's, whose name holds a hash of this machine's roots, and
// built uncached in a temporary directory of its own, which main's uncached gate settles under the key it computes.
// Not parallel: cached calls test.Setenv for ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG and ADAMIC_BUILD_CACHE.
func TestAKeyNamesAProductByItsKeyWhereverItIs(test *testing.T) {
	cached(test)
	test.Setenv("TMPDIR", test.TempDir())
	root, repositoryRootError := repositoryRoot()
	if repositoryRootError != nil {
		test.Fatal(repositoryRootError)
	}
	build := func(directory string) error {
		return os.WriteFile(
			filepath.Join(directory, "main.ts"),
			[]byte("import '"+Relative(filepath.Join(root, "go.mod"))+"';\n"),
			0o644,
		)
	}
	for _, mode := range []string{"", "off"} {
		test.Setenv("ADAMIC_BUILD_CACHE", mode)
		inputs := thisPackage
		inputs.Name = "named by its key, cache " + mode
		product := Product(test, inputs, build)
		key := NameKey(root, inputs)
		copied, resolvedError := Resolved(product)
		if resolvedError != nil {
			test.Fatal(resolvedError)
		}
		if copied == product {
			test.Fatalf("cache %q: %s, which names <repository>, resolved to itself", mode, product)
		}
		want := "main=<build cache>/" + key + "/main.ts"
		for _, held := range []string{product, copied} {
			if got := portable(root, "main="+filepath.Join(held, "main.ts")); got != want {
				test.Errorf("cache %q: a key names %s as %q, want %q", mode, filepath.Join(held, "main.ts"), got, want)
			}
		}
	}
}

// forgetHeld drops what this process holds by name key, as a process that never got these products starts.
func forgetHeld(test *testing.T) {
	test.Helper()
	clear := func() {
		productDirectories.Range(
			func(directory, nameKey any) bool { forgotten(nameKey.(string), directory.(string)); return true },
		)
		productKeys.Range(func(key, _ any) bool { productKeys.Delete(key); return true })
		resolvedDirectories.Range(func(key, _ any) bool { resolvedDirectories.Delete(key); return true })
	}
	clear()
	test.Cleanup(clear)
}

// One product names itself by its name key, and the name reads back wherever the reader's cache holds it (#tqrqx60 with
// #vt46geg): while its traced build is pending, once settled under its read set's key (what Loom's runner unpacks, the
// key's directory and the name key's read sets, read with no network), and built for this machine alone under local/.
// Each read is from a process holding none of them. A name whose read set no longer values to a cached product reads
// as <cache>/<name key>, which holds nothing.
// Not parallel: the rig changes the working directory, repositoryRoot and the build environment.
func TestANameKeyReadsBackWhereverItsProductLives(test *testing.T) {
	r := newRig(test)
	forgetHeld(test)
	var own string
	pending := Product(test, port, func(directory string) error {
		own = Relative(filepath.Join(directory, "product"))
		return os.WriteFile(filepath.Join(directory, "product"), []byte(own), 0o644)
	})
	nameKey := NameKey(r.root, port)
	if own != "<build cache>/"+nameKey+"/product" {
		test.Fatalf("a building product named itself %q, want its name key %s", own, nameKey)
	}
	forgetHeld(test)
	if got := Absolute(own); got != filepath.Join(pending, "product") {
		test.Fatalf("pending, the name read back as %s, want %s", got, filepath.Join(pending, "product"))
	}
	// Under the trace, the directory a name reads back to, and a resolved copy of it, are journaled as that product's,
	// so a build reading there reads the product (trace.go).
	copied, resolvedError := Resolved(pending)
	if resolvedError != nil || copied == pending {
		test.Fatalf("the pending product resolved to %s, %v; want a copy", copied, resolvedError)
	}
	entries, readJournalError := readJournal(r.trace)
	if readJournalError != nil {
		test.Fatal(readJournalError)
	}
	for _, directory := range []string{pending, copied} {
		if !slices.ContainsFunc(entries, func(entry journalEntry) bool {
			return entry.Event == "found" && entry.NameKey == nameKey && entry.Directory == directory
		}) {
			test.Fatalf("%s isn't journaled as found under %s", directory, nameKey[:12])
		}
	}

	build := r.built(test)["port"]
	r.window(build.Build, func() { r.open(r.pid, "source/a.c") })
	if settlement := r.settle(test); len(settlement.Settled) != 1 {
		test.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	file, loadReadsError := loadReads(r.cache, nameKey)
	if loadReadsError != nil || len(file.Sets) != 1 {
		test.Fatalf("read sets: %+v, %v", file, loadReadsError)
	}
	settled := filepath.Join(r.cache, file.Sets[0].Key)
	test.Setenv("ADAMIC_BUILD_TRACE", "")
	forgetHeld(test)
	if got := Absolute(own); got != filepath.Join(settled, "product") {
		test.Fatalf("settled, the name read back as %s, want %s", got, filepath.Join(settled, "product"))
	}
	if content, _ := os.ReadFile(filepath.Join(settled, "product")); string(content) != own {
		test.Fatalf("the settled product holds %q, want its own name %q, never its key", content, own)
	}

	// A read changed: no set values to a cached product, so the name finds nothing rather than a stale product.
	write(test, r.root, "source/a.c", "int a2;\n")
	forgetTracked()
	forgetHeld(test)
	if got := Absolute(own); got != filepath.Join(r.cache, nameKey, "product") {
		test.Fatalf(
			"after a read changed, the name read back as %s, want the empty %s",
			got,
			filepath.Join(r.cache, nameKey, "product"),
		)
	}

	// Built for this machine alone: found by the inputs its name key points to, keyed on the tree at hand.
	local := Inputs{Name: "local port", Files: []string{"source"}}
	product := Product(test, local, func(directory string) error {
		return os.WriteFile(
			filepath.Join(directory, "product"),
			[]byte(Relative(filepath.Join(r.root, "go.mod"))),
			0o644,
		)
	})
	if filepath.Dir(product) != filepath.Join(r.cache, "local") {
		test.Fatalf("an untraced product is at %s, want under local/", product)
	}
	forgetHeld(test)
	localName := "<build cache>/" + NameKey(r.root, local) + "/product"
	if got := Absolute(localName); got != filepath.Join(product, "product") {
		test.Fatalf("local, the name read back as %s, want %s", got, filepath.Join(product, "product"))
	}
	forgetHeld(test)
	if copied, resolvedError := Resolved(product); resolvedError != nil || copied == product {
		test.Fatalf(
			"a local product another process didn't build resolved to %s, %v; want a copy",
			copied,
			resolvedError,
		)
	}
	write(test, r.root, "source/b.c", "int b2;\n")
	forgetTracked()
	forgetHeld(test)
	if got := Absolute(localName); got == filepath.Join(product, "product") {
		test.Fatalf("after a declared file changed, the name still read back as the old local product %s", got)
	}
}

// A traced build is scanned like any other: one naming the machine that built it is refused, and nothing is placed.
// Not parallel: the rig changes the working directory, repositoryRoot and the build environment.
func TestATracedProductNamingItsMachineIsRefused(test *testing.T) {
	r := newRig(test)
	forgetHeld(test)
	_, getError := Get(port, func(directory string) error {
		return os.WriteFile(filepath.Join(directory, "product"), []byte(filepath.Join(r.root, "go.mod")), 0o644)
	})
	if getError == nil || !strings.Contains(getError.Error(), "holds a path of the machine that built it") {
		test.Fatalf("a traced product naming the repository built: %v", getError)
	}
	// filled leaves only its lock file: no product directory, no scratch one.
	entries, _ := os.ReadDir(pendingDirectory(r.cache))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lock") {
			test.Fatalf("a refused traced product left %s pending", entry.Name())
		}
	}
	if len(r.built(test)) != 0 {
		test.Fatalf("a refused traced product was journaled as built: %v", r.built(test))
	}
}

// A whole path a product carries from a tracked file among its own declared inputs is that input's text, not this
// machine's (#arw7837: json's pinned corpus carries the repository's reports, which name /tmp/adamic-gate/... runs), so
// it builds. The same path from an untracked file, or from a tracked file the product doesn't declare, is refused, as
// is a path an input names only a prefix of or a longer path under, and a build cache path whatever an input says.
// Not parallel: inRepository changes the working directory and repositoryRoot; cached and test.Setenv set the environment.
func TestAPathADeclaredInputCarriesPasses(test *testing.T) {
	root, resolveSymlinksError := filepath.EvalSymlinks(test.TempDir())
	if resolveSymlinksError != nil {
		test.Fatal(resolveSymlinksError)
	}
	home, resolveSymlinksError := filepath.EvalSymlinks(test.TempDir())
	if resolveSymlinksError != nil {
		test.Fatal(resolveSymlinksError)
	}
	test.Setenv("HOME", home)
	token := filepath.Join(home, "runs", "report-1", "cold")
	write(test, root, "go.mod", "module github.com/system-inc/adamic\n\ngo 1.27\n")
	write(test, root, "reports/run.json", `{"directory":"`+token+`"}`)
	write(test, root, "prefix/run.json", `{"directory":"`+filepath.Join(home, "runs")+`"}`)
	write(test, root, "longer/run.json", `{"directory":"`+filepath.Join(token, "more")+`"}`)
	write(test, root, "other/run.json", `{"directory":"`+token+`"}`)
	gitIn(test, root, "init", "-q")
	gitIn(test, root, "add", ".")
	gitIn(test, root, "commit", "-q", "-m", "tree")
	write(test, root, "untracked/run.json", `{"directory":"`+token+`"}`)
	inRepository(test, root)
	cache, _ := cached(test)
	forgetTracked()
	carrying := func(text string) func(directory string) error {
		return func(directory string) error {
			return os.WriteFile(filepath.Join(directory, "corpus"), []byte(text), 0o644)
		}
	}
	if _, getError := Get(
		Inputs{Name: "vouched", Files: []string{"reports"}},
		carrying("case "+token+"\n"),
	); getError != nil {
		test.Fatalf("a path a declared tracked input carries was refused: %v", getError)
	}
	for _, refused := range []struct {
		name, text string
		files      []string
	}{
		{"an untracked file", "case " + token + "\n", []string{"untracked"}},
		{"an undeclared tracked file (other/run.json carries it)", "case " + token + "\n", []string{"go.mod"}},
		{"a prefix in the input", "case " + token + "\n", []string{"prefix"}},
		{"a longer path in the input", "case " + token + "\n", []string{"longer"}},
		{"the build cache", "case " + filepath.Join(cache, "local", "product") + "\n", []string{"reports"}},
	} {
		if refused.name == "the build cache" {
			write(test, root, "reports/cache.json", filepath.Join(cache, "local", "product"))
			gitIn(test, root, "add", ".")
			gitIn(test, root, "commit", "-q", "-m", "cache")
			forgetTracked()
		}
		_, getError := Get(Inputs{Name: "refused " + refused.name, Files: refused.files}, carrying(refused.text))
		if getError == nil || !strings.Contains(getError.Error(), "holds a path of the machine that built it") {
			test.Errorf("%s: the build gave %v, want a refusal", refused.name, getError)
		}
	}
}

// Vouching is a lookup once the declared files are read: json's pinned corpus product, 819 declared tracked files and
// a token for every report path it carries, ran past go test's 10 minutes on landable-14b's tree (Oct 11) scanning
// every file for every token. Here 800 files and 8,000 tokens build with no token scanned for. Mutant: every token
// scanned for.
// Not parallel: inRepository changes the working directory and repositoryRoot; cached and test.Setenv set the environment.
func TestVouchingIsALookupAtTheCorpusSize(test *testing.T) {
	root, resolveSymlinksError := filepath.EvalSymlinks(test.TempDir())
	if resolveSymlinksError != nil {
		test.Fatal(resolveSymlinksError)
	}
	home, resolveSymlinksError := filepath.EvalSymlinks(test.TempDir())
	if resolveSymlinksError != nil {
		test.Fatal(resolveSymlinksError)
	}
	test.Setenv("HOME", home)
	write(test, root, "go.mod", "module github.com/system-inc/adamic\n\ngo 1.27\n")
	var carried strings.Builder
	for file := range 800 {
		var report strings.Builder
		for run := range 10 {
			token := filepath.Join(
				home,
				"runs",
				fmt.Sprintf("report-%d", file),
				fmt.Sprintf("run-%d", run),
				"result.json",
			)
			fmt.Fprintf(&report, `{"directory":"%s","padding":"%s"},`+"\n", token, strings.Repeat("x", 200))
			fmt.Fprintf(&carried, "case %s\n", token)
		}
		write(test, root, fmt.Sprintf("reports/%03d/run.json", file), report.String())
	}
	gitIn(test, root, "init", "-q")
	gitIn(test, root, "add", ".")
	gitIn(test, root, "commit", "-q", "-m", "tree")
	inRepository(test, root)
	cached(test)
	forgetTracked()
	scans := vouchingScans.Load()
	started := time.Now()
	_, getError := Get(Inputs{Name: "vouched at size", Files: []string{"reports"}}, func(directory string) error {
		return os.WriteFile(filepath.Join(directory, "corpus"), []byte(carried.String()), 0o644)
	})
	if getError != nil {
		test.Fatalf("8,000 vouched paths were refused: %v", getError)
	}
	if scanned := vouchingScans.Load() - scans; scanned != 0 {
		test.Fatalf("%d lookups of 8,000 tokens scanned every declared file", scanned)
	}
	if took := time.Since(started); took > 30*time.Second {
		test.Fatalf("8,000 tokens against 800 declared files took %v", took)
	}
}

// places resolves each product directory a process holds once, not once per key: a test that keys many products while
// holding many (bridge/tsgo's BridgeInputProofs, home-6's dry run, Oct 11) spent its 10 minutes resolving them again
// for every key, a stat per path element, each a stop under a trace. The fixed directories resolve once too (Cloud, Oct
// 17: tsprinter's NPM tests lstat'ed /home 32,600 times each). Mutants: product directories resolved on every call; the
// home directory resolved on every call.
// The cache and the output are directories that exist, since one that doesn't resolves again until it does
// (TestADirectoryResolvedBeforeItExistsResolvesOnceItDoes).
// Not parallel: it adds to productDirectories, which every key in the process reads, and sets the cache's variables.
func TestPlacesResolveAProductDirectoryOnce(test *testing.T) {
	test.Setenv("ADAMIC_BUILD_CACHE_DIR", test.TempDir())
	test.Setenv("ADAMIC_BUILD_OUTPUT_DIR", test.TempDir())
	parent := test.TempDir()
	var directories []string
	for index := range 50 {
		directory := filepath.Join(parent, fmt.Sprintf("product-%02d", index))
		if createDirectoryError := os.Mkdir(directory, 0o755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		directories = append(directories, directory)
		located(fmt.Sprintf("%064x", index), directory)
	}
	test.Cleanup(func() {
		for index, directory := range directories {
			forgotten(fmt.Sprintf("%064x", index), directory)
			directoryResolutions.Delete(directory)
		}
	})
	first := places("")
	before := resolutions.Load()
	second := places("")
	if again := resolutions.Load() - before; again != 0 {
		test.Fatalf(
			"places resolved %d paths the second time, with the repository, the cache, the home directory and 50 product directories it had resolved already",
			again,
		)
	}
	if len(first) != len(second) || len(first) < 50 {
		test.Fatalf("places gave %d, then %d", len(first), len(second))
	}
}

// A directory resolved before it exists resolves again once it does, so an output or a cache made after the first key
// is named by its real path, never left unresolved for the rest of the process; one that exists then resolves once.
// Mutant: the failed resolution memoized.
func TestADirectoryResolvedBeforeItExistsResolvesOnceItDoes(test *testing.T) {
	test.Parallel()
	parent := test.TempDir()
	target := filepath.Join(parent, "target")
	if createDirectoryError := os.Mkdir(target, 0o755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	link := filepath.Join(parent, "made-later")
	test.Cleanup(func() { rootResolutions.Delete(link) })
	if got := resolvedOnce(link); got != "" {
		test.Fatalf("a directory that isn't there resolved to %q", got)
	}
	if symlinkError := os.Symlink(target, link); symlinkError != nil {
		test.Fatal(symlinkError)
	}
	want, resolveSymlinksError := filepath.EvalSymlinks(target)
	if resolveSymlinksError != nil {
		test.Fatal(resolveSymlinksError)
	}
	if got := resolvedOnce(link); got != want {
		test.Fatalf("once it exists it resolves to %q, want %q", got, want)
	}
	if value, ok := rootResolutions.Load(link); !ok || value != want {
		test.Fatalf("the resolution that worked wasn't kept: %v %v", value, ok)
	}
}

// A product directory's resolution that failed, kept by its own rule, never stands for a root of the same path: the two
// are held apart (Tests' should on cb285d2e33). Mutant: one map for both.
func TestAProductDirectorysFailureNeverStandsForARoot(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "both")
	test.Cleanup(func() {
		directoryResolutions.Delete(path)
		rootResolutions.Delete(path)
	})
	if got := productResolved(path); got != "" {
		test.Fatalf("a directory that isn't there resolved to %q", got)
	}
	if createDirectoryError := os.Mkdir(path, 0o755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	want, resolveSymlinksError := filepath.EvalSymlinks(path)
	if resolveSymlinksError != nil {
		test.Fatal(resolveSymlinksError)
	}
	if got := resolvedOnce(path); got != want {
		test.Fatalf("the root resolves to %q, want %q: the product directory's failure stood for it", got, want)
	}
}

// Relative makes its places once, as keying, and again only when what they're made from changed: a product directory
// located or forgotten, or the environment (lint's profile lowering called it about 8,400 times, 2.3 million stats, 637
// s traced against 9.7 s, Home, Oct 12). Mutants: places made on every call; a located or forgotten directory, or a
// changed ADAMIC_ value, not making them again; the locations looked for outside keying.
// Not parallel: test.Setenv, statLocation and productDirectories.
func TestRelativeMakesItsPlacesOncePerChange(test *testing.T) {
	test.Setenv("ADAMIC_BUILD_TRACE", test.TempDir())
	first, second := test.TempDir(), test.TempDir()
	test.Setenv("ADAMIC_RELATIVE_PLACES", first)
	unkeyed := 0
	saved := statLocation
	statLocation = func(name string) (os.FileInfo, error) {
		if keyingNow.Load() == 0 {
			unkeyed++
		}
		return saved(name)
	}
	test.Cleanup(func() { statLocation = saved })
	if got := Relative(filepath.Join(first, "a")); got != "<ADAMIC_RELATIVE_PLACES>/a" {
		test.Fatalf("a located directory: %q", got)
	}
	before := resolutions.Load()
	for range 100 {
		Relative(filepath.Join(first, "a"))
	}
	if again := resolutions.Load() - before; again != 0 {
		test.Fatalf("100 calls with nothing changed resolved %d paths again", again)
	}
	test.Setenv("ADAMIC_RELATIVE_PLACES", second)
	if got := Relative(filepath.Join(second, "b")); got != "<ADAMIC_RELATIVE_PLACES>/b" {
		test.Fatalf("after the variable moved: %q", got)
	}
	product, nameKey := test.TempDir(), strings.Repeat("e", 64)
	located(nameKey, product)
	if got := Relative(filepath.Join(product, "c")); got != "<build cache>/"+nameKey+"/c" {
		test.Fatalf("a located product directory: %q", got)
	}
	forgotten(nameKey, product)
	if got := Relative(filepath.Join(product, "c")); got != filepath.Join(product, "c") {
		test.Fatalf("a forgotten product directory: %q", got)
	}
	// Absolute looks for locations itself, after a change only Relative hasn't seen.
	test.Setenv("ADAMIC_RELATIVE_PLACES", first)
	if got := Absolute("<ADAMIC_RELATIVE_PLACES>/d"); got != filepath.Join(first, "d") {
		test.Fatalf("read back after the variable moved back: %q", got)
	}
	if unkeyed != 0 {
		test.Fatalf("locations were looked for outside keying %d times", unkeyed)
	}
}

// A product reached through a link, as a runner's unit cache links one from the box's shared products, is read where it
// points: Resolved copies it with its names read back, and its digest is its bytes (mega's run 2, Oct 12: estree's
// mutant port came back unresolved, and node read its <repository> imports raw). Mutant: a walk that starts at the
// link.
// Not parallel: cached calls test.Setenv.
func TestAProductReachedThroughALinkIsReadWhereItPoints(test *testing.T) {
	cache, _ := cached(test)
	root, repositoryRootError := repositoryRoot()
	if repositoryRootError != nil {
		test.Fatal(repositoryRootError)
	}
	shared := filepath.Join(test.TempDir(), "loom-products-1", "blob", strings.Repeat("ab", 32))
	write(test, shared, "main.ts", "import '<repository>/go.mod';\n")
	link := filepath.Join(cache, strings.Repeat("ab", 32))
	if symlinkError := os.Symlink(shared, link); symlinkError != nil {
		test.Fatal(symlinkError)
	}
	test.Cleanup(func() { resolvedDirectories.Delete(link) })
	directory, resolvedError := Resolved(link)
	if resolvedError != nil {
		test.Fatal(resolvedError)
	}
	content, readError := os.ReadFile(filepath.Join(directory, "main.ts"))
	if readError != nil || directory == link || string(content) != "import '"+filepath.Join(root, "go.mod")+"';\n" {
		test.Fatalf("Resolved gave %s holding %q, %v; want a copy naming this checkout", directory, content, readError)
	}
	through, productDigestError := productDigest(link)
	if productDigestError != nil {
		test.Fatal(productDigestError)
	}
	if bytes, productDigestError := productDigest(shared); productDigestError != nil || through != bytes {
		test.Fatalf(
			"the product's digest through its link is %s, its bytes' %s (%v)",
			through,
			bytes,
			productDigestError,
		)
	}
}
