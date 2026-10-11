package buildcache

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// A repository of its own under the test's directory, so keys hash files the test controls.
func repository(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	write(t, directory, "source/a.c", "int a;\n")
	write(t, directory, "source/b.c", "int b;\n")
	write(t, directory, "flags.txt", "-O2\n")
	return directory
}

func write(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func key(t *testing.T, root string, inputs Inputs) string {
	t.Helper()
	value, err := Key(root, inputs)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

var base = Inputs{Name: "port", Files: []string{"source", "flags.txt"}, Flags: []string{"-DADAMIC"}, Toolchain: []string{"clang 20.1.8"}}

// Every input that can change a product changes its key: drop any of them from the hash and its subtest fails.
func TestEveryInputChangesTheKey(t *testing.T) {
	t.Parallel()
	changed := map[string]func(t *testing.T, root string) Inputs{
		"name":         func(t *testing.T, root string) Inputs { in := base; in.Name = "port2"; return in },
		"flag":         func(t *testing.T, root string) Inputs { in := base; in.Flags = []string{"-DADAMIC_TSGO"}; return in },
		"flag order":   func(t *testing.T, root string) Inputs { in := base; in.Flags = []string{"-DADAMIC", "-O0"}; return in },
		"toolchain":    func(t *testing.T, root string) Inputs { in := base; in.Toolchain = []string{"clang 21.0.0"}; return in },
		"file list":    func(t *testing.T, root string) Inputs { in := base; in.Files = []string{"source"}; return in },
		"file content": func(t *testing.T, root string) Inputs { write(t, root, "source/a.c", "int a2;\n"); return base },
		"added file":   func(t *testing.T, root string) Inputs { write(t, root, "source/c.c", ""); return base },
		"renamed file": func(t *testing.T, root string) Inputs {
			if err := os.Rename(filepath.Join(root, "source/b.c"), filepath.Join(root, "source/b2.c")); err != nil {
				t.Fatal(err)
			}
			return base
		},
		"exec bit": func(t *testing.T, root string) Inputs {
			if err := os.Chmod(filepath.Join(root, "flags.txt"), 0o755); err != nil {
				t.Fatal(err)
			}
			return base
		},
		"empty directory": func(t *testing.T, root string) Inputs {
			if err := os.Mkdir(filepath.Join(root, "source", "empty"), 0o755); err != nil {
				t.Fatal(err)
			}
			return base
		},
	}
	for name, change := range changed {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := repository(t)
			before := key(t, root, base)
			if after := key(t, root, change(t, root)); after == before {
				t.Fatalf("changing the %s left the key %s", name, before)
			}
		})
	}
	t.Run("nothing changed", func(t *testing.T) {
		t.Parallel()
		root := repository(t)
		before := key(t, root, base)
		later := time.Now().Add(time.Hour)
		if err := os.Chtimes(filepath.Join(root, "source/a.c"), later, later); err != nil {
			t.Fatal(err)
		}
		if after := key(t, root, base); after != before {
			t.Fatalf("a touched file with the same content moved the key: %s, then %s", before, after)
		}
	})
	t.Run("outside the repository", func(t *testing.T) {
		t.Parallel()
		root := repository(t)
		for _, name := range []string{"/etc/hosts", "../escape"} {
			in := base
			in.Files = []string{name}
			if _, err := Key(root, in); err == nil {
				t.Fatalf("input %s was accepted", name)
			}
		}
	})
}

// No product relocating holds is keyed as a v1 product was: Workshop's store held v1 trees' products under the very keys
// landable-6 asked for, with other bytes (Oct 11). A change to what a product holds for the same inputs is a new
// keyVersion, and this test names each one, so moving it is a decision. Mutant: keyVersion back to v1.
func TestAKeyNamesHowProductsAreHeld(t *testing.T) {
	t.Parallel()
	root := repository(t)
	if keyVersion != "v2" {
		t.Fatalf("keyVersion is %q: name the new version here, and why its products' bytes differ, in keyVersion's comment", keyVersion)
	}
	old, err := keyAt("v1", root, base)
	if err != nil {
		t.Fatal(err)
	}
	if now := key(t, root, base); now == old {
		t.Fatalf("the same inputs key %s under v1 and under %s", now, keyVersion)
	}
}

// Products are cached under a test-owned directory, keyed in this repository (Get finds it from the working
// directory), with the census log captured.
func cached(t *testing.T) (string, string) {
	t.Helper()
	cache := t.TempDir()
	log := filepath.Join(t.TempDir(), "builds.log")
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", cache)
	t.Setenv("ADAMIC_BUILD_LOG", log)
	t.Setenv("ADAMIC_BUILD_CACHE", "")
	// The real store is never reached from a test; the store's own tests serve one.
	t.Setenv("ADAMIC_BUILD_STORE", "off")
	t.Setenv("ADAMIC_BUILD_STORE_TOKEN", filepath.Join(t.TempDir(), "no-token"))
	return cache, log
}

var thisPackage = Inputs{Name: "buildcache test", Files: []string{"internal/buildcache/buildcache.go"}}

// Not parallel: cached calls t.Setenv for ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG and ADAMIC_BUILD_CACHE.
func TestABuildRunsOncePerKey(t *testing.T) {
	_, log := cached(t)
	var builds int
	build := func(directory string) error {
		builds++
		return os.WriteFile(filepath.Join(directory, "product"), []byte("built"), 0o644)
	}
	first := Product(t, thisPackage, build)
	second := Product(t, thisPackage, build)
	if builds != 1 || first != second {
		t.Fatalf("%d builds, products %s and %s", builds, first, second)
	}
	if content, err := os.ReadFile(filepath.Join(second, "product")); err != nil || string(content) != "built" {
		t.Fatalf("the product reads %q, %v", content, err)
	}
	lines, _ := os.ReadFile(log)
	if got := strings.Fields(strings.ReplaceAll(string(lines), "\n", " ")); len(got) != 10 || got[3] != "miss" || got[8] != "hit" || got[1] != "buildcache_test" {
		t.Fatalf("census lines: %q", lines)
	}
}

// Not parallel: cached calls t.Setenv for ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG and ADAMIC_BUILD_CACHE.
func TestAFailedBuildPublishesNothing(t *testing.T) {
	cache, _ := cached(t)
	failure := errors.New("clang failed")
	if _, err := Get(thisPackage, func(directory string) error {
		os.WriteFile(filepath.Join(directory, "half"), nil, 0o644)
		return failure
	}); !errors.Is(err, failure) {
		t.Fatalf("the build's error was lost: %v", err)
	}
	for _, directory := range []string{cache, filepath.Join(cache, "local")} {
		entries, _ := os.ReadDir(directory)
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() != "local" {
				t.Fatalf("a failed build left %s in %s", entry.Name(), directory)
			}
		}
	}
	var built bool
	Product(t, thisPackage, func(string) error { built = true; return nil })
	if !built {
		t.Fatal("the next call reused a failed build")
	}
}

// Not parallel: cached calls t.Setenv for ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG and ADAMIC_BUILD_CACHE.
func TestUncachedModeBuildsEveryTime(t *testing.T) {
	cached(t)
	t.Setenv("ADAMIC_BUILD_CACHE", "off")
	var builds int
	build := func(string) error { builds++; return nil }
	first := Product(t, thisPackage, build)
	second := Product(t, thisPackage, build)
	if builds != 2 || first == second {
		t.Fatalf("%d builds, products %s and %s", builds, first, second)
	}
}

// Read mode finds a product by its read sets and never builds: a product with none, or another name key, is missing,
// named with its name key, why, the cache it looked in and every input as this machine keys them.
// Not parallel: newRig.
func TestReadModeNamesAMissingProductAndNeverBuilds(t *testing.T) {
	r := newRig(t)
	_, log := cached(t)
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", r.cache)
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	r.window(build.Build, func() { r.open(r.pid, "source/a.c") })
	r.settle(t)
	if _, err := reading(t, port); err != nil {
		t.Fatalf("read mode missed a settled product: %v", err)
	}
	elsewhere := port
	elsewhere.Flags = []string{"-DOTHER"}
	_, err := reading(t, elsewhere)
	if !errors.Is(err, ErrNotBuilt) {
		t.Fatalf("a missing product read as %v, want ErrNotBuilt", err)
	}
	missing := NameKey(r.root, elsewhere)
	for _, want := range []string{"product port", missing[:12], "no traced build", r.cache, "TestProduct_", "flag -DOTHER", "file source"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("the error doesn't name %q: %v", want, err)
		}
	}
	entries, _ := os.ReadDir(r.cache)
	for _, entry := range entries {
		if strings.Contains(entry.Name(), missing[:12]) {
			t.Fatalf("read mode left %s in the cache for the missing product", entry.Name())
		}
	}
	if lines, _ := os.ReadFile(log); !strings.Contains(string(lines), " missing ") {
		t.Fatalf("the census has no missing line: %q", lines)
	}
}

func repositoryRootForTest(t *testing.T) string {
	t.Helper()
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// Not parallel: cached calls t.Setenv for ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG and ADAMIC_BUILD_CACHE.
func TestParallelCallersBuildOnce(t *testing.T) {
	cached(t)
	var builds atomic.Int32
	var group sync.WaitGroup
	products := make([]string, 8)
	for index := range products {
		group.Add(1)
		go func() {
			defer group.Done()
			directory, err := Get(thisPackage, func(string) error {
				builds.Add(1)
				time.Sleep(50 * time.Millisecond)
				return nil
			})
			if err != nil {
				t.Error(err)
			}
			products[index] = directory
		}()
	}
	group.Wait()
	if builds.Load() != 1 {
		t.Fatalf("%d builds for one key", builds.Load())
	}
	for _, product := range products {
		if product != products[0] {
			t.Fatalf("products differ: %v", products)
		}
	}
}

// A key names what changes a product, never which machine computed it (#t37sw0f): the checkout's path, the home
// directory, the build cache and Go's -p share are spelled the same on every machine, and every flag that changes the
// output still moves the key. Each pair below is one machine against another; drop a rule from portable and its pair
// fails.
// Not parallel: points the home directory and the build cache at each machine's through t.Setenv.
func TestAKeyNamesNoMachine(t *testing.T) {
	asMachine := func(home, root, value string) string {
		t.Setenv("HOME", home)
		t.Setenv("ADAMIC_BUILD_CACHE_DIR", filepath.Join("/srv", filepath.Base(home), "products"))
		return portable(root, value)
	}
	for _, same := range []struct{ name, workshop, runner string }{
		{"the checkout's path", "repository=/home/ahra/work/adamic", "repository=/srv/runner/7/adamic"},
		{"a path inside the checkout", "main=/home/ahra/work/adamic/stage1/cohere/lint", "main=/srv/runner/7/adamic/stage1/cohere/lint"},
		{"the home directory", "CC=/home/ahra/adamic-tools/llvm/bin/clang", "CC=/home/cloud/adamic-tools/llvm/bin/clang"},
		{"another product, by its path", "main=/srv/ahra/products/0123abcd/main.ts", "main=/srv/cloud/products/0123abcd/main.ts"},
		{"GOFLAGS' -p share, last", "GOFLAGS=-buildvcs=false -trimpath -p=12", "GOFLAGS=-buildvcs=false -trimpath"},
		{"GOFLAGS' -p share, first", "GOFLAGS=-p=12 -buildvcs=false -trimpath", "GOFLAGS=-p=3 -buildvcs=false -trimpath"},
		{"GOFLAGS' -p share, alone", "GOFLAGS=-p=12", "GOFLAGS="},
		{"go env -json's GOFLAGS", `{"GOFLAGS": "-p=12 -trimpath"}`, `{"GOFLAGS": "-trimpath"}`},
		{"a go build argument", "arguments -trimpath -p 12 -tags=fancy", "arguments -trimpath -tags=fancy"},
	} {
		workshop := asMachine("/home/ahra", "/home/ahra/work/adamic", same.workshop)
		runner := asMachine("/home/cloud", "/srv/runner/7/adamic", same.runner)
		if workshop != runner {
			t.Errorf("%s: Workshop keys %q, a runner %q", same.name, workshop, runner)
		}
	}
	for _, kept := range []struct{ name, value string }{
		{"a longer name beside the checkout", "/srv/work/adamic-bench/x"},
		{"a path ending in the home directory's name", "/chroot/home/ahra"},
		{"-parallel", "-parallel=4"},
		{"-pgo", "-pgo=off"},
		{"-p with a path", "-p=./tsconfig.json"},
	} {
		if got := asMachine("/home/ahra", "/srv/work/adamic", kept.value); got != kept.value {
			t.Errorf("%s: %q became %q", kept.name, kept.value, got)
		}
	}
	for _, moved := range []struct{ name, before, after string }{
		{"a tag", "GOFLAGS=-tags=plain -p=4", "GOFLAGS=-tags=fancy -p=4"},
		{"-trimpath", "GOFLAGS=-trimpath -p=4", "GOFLAGS=-p=4"},
		{"a C flag", "CGO_CFLAGS=-O1", "CGO_CFLAGS=-O2"},
		{"a path inside the tree", "main=/home/ahra/work/adamic/stage1/a", "main=/home/ahra/work/adamic/stage1/b"},
		{"a path outside the home and the tree", "CC=/usr/bin/clang", "CC=/opt/llvm/bin/clang"},
	} {
		if asMachine("/home/ahra", "/home/ahra/work/adamic", moved.before) == asMachine("/home/ahra", "/home/ahra/work/adamic", moved.after) {
			t.Errorf("changing %s left the value the same: %q and %q", moved.name, moved.before, moved.after)
		}
	}
	// A home of / would rewrite every absolute path: it is never a place.
	if got := asMachine("/", "/home/ahra/work/adamic", "/usr/bin/clang"); got != "/usr/bin/clang" {
		t.Errorf("with HOME=/ a path became %q", got)
	}

	// Key itself: the same files under two checkouts, the name and each flag naming its own checkout and -p share,
	// one key; a source byte moves it.
	keyAs := func(home string, root string, share int) string {
		t.Setenv("HOME", home)
		return key(t, root, Inputs{Name: "port " + filepath.Join(root, "source/a.c"), Files: []string{"source", "flags.txt"},
			Flags: []string{"source-root=" + filepath.Join(root, "source"), fmt.Sprintf("GOFLAGS=-trimpath -p=%d", share)}, Toolchain: []string{"clang 20.1.8"}})
	}
	one, two := repository(t), repository(t)
	// A submodule's .git file names its git directory by absolute path; an unpacked source has none.
	write(t, one, "source/.git", "gitdir: /home/ahra/work/adamic/.git/modules/source\n")
	if workshop, runner := keyAs("/home/ahra", one, 12), keyAs("/home/cloud", two, 3); workshop != runner {
		t.Fatalf("one tree at %s and %s keyed %s and %s", one, two, workshop, runner)
	}
	write(t, two, "source/a.c", "int a2;\n")
	if keyAs("/home/ahra", one, 12) == keyAs("/home/cloud", two, 3) {
		t.Fatal("a changed source byte left the key the same")
	}
}

// A key hashes what the tree carries (#smkk3et): a checkout's untracked files, under its own repository or a
// submodule's, never key a product, so a checkout and an unpacked source of it (tracked files only, no .git) key the
// same. A tracked byte still moves the key, and a file named in Files counts whether or not git tracks it.
func TestAKeyHashesWhatTheTreeCarries(t *testing.T) {
	t.Parallel()
	git := func(directory string, arguments ...string) {
		t.Helper()
		command := exec.Command("git", append([]string{"-C", directory, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "init.defaultBranch=main"}, arguments...)...)
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", arguments, err, output)
		}
	}
	carried := map[string]string{"source/a.c": "int a;\n", "source/b.c": "int b;\n", "flags.txt": "-O2\n", "module/inner.c": "int inner;\n"}
	checkout, unpacked := t.TempDir(), t.TempDir()
	for name, content := range carried {
		write(t, checkout, name, content)
		write(t, unpacked, name, content)
	}
	git(filepath.Join(checkout, "module"), "init", "-q")
	git(filepath.Join(checkout, "module"), "add", "inner.c")
	git(filepath.Join(checkout, "module"), "commit", "-q", "-m", "module")
	git(checkout, "init", "-q")
	git(checkout, "add", "source", "flags.txt", "module")
	for name, content := range map[string]string{"source/stray.txt": "a note\n", "source/build/out.o": "leftover", "module/stray.tmp": "x", "module/cache/entry": "y"} {
		write(t, checkout, name, content)
	}
	// A checkout whose submodule's files are linked in without its git directory: the repository above knows it only
	// as a gitlink, and every file under it counts.
	linked := t.TempDir()
	for name, content := range carried {
		write(t, linked, name, content)
	}
	git(filepath.Join(linked, "module"), "init", "-q")
	git(filepath.Join(linked, "module"), "add", "inner.c")
	git(filepath.Join(linked, "module"), "commit", "-q", "-m", "module")
	git(linked, "init", "-q")
	git(linked, "add", "source", "flags.txt", "module")
	if err := os.RemoveAll(filepath.Join(linked, "module", ".git")); err != nil {
		t.Fatal(err)
	}
	inputs := Inputs{Name: "port", Files: []string{"source", "flags.txt", "module"}}
	inSource := key(t, unpacked, inputs)
	if inCheckout := key(t, checkout, inputs); inCheckout != inSource {
		t.Fatalf("a checkout with untracked files keyed %s, its unpacked source %s", inCheckout, inSource)
	}
	if inLinked := key(t, linked, inputs); inLinked != inSource {
		t.Fatalf("a checkout with its submodule's files linked in keyed %s, its unpacked source %s", inLinked, inSource)
	}
	before := key(t, checkout, inputs)
	write(t, checkout, "module/inner.c", "int inner2;\n")
	if key(t, checkout, inputs) == before {
		t.Fatal("a tracked byte in a submodule left the key the same")
	}
	named := Inputs{Name: "port", Files: []string{"source/stray.txt"}}
	first := key(t, checkout, named)
	write(t, checkout, "source/stray.txt", "another note\n")
	if key(t, checkout, named) == first {
		t.Fatal("an untracked file named in Files isn't keyed")
	}
}

// go.work.sum is never keyed (Workshop's proof, Oct 10): go writes it during a run, so one checkout had it and
// another didn't, and the products naming it keyed apart, though it changes no build's output.
func TestGoWorkSumKeysNothing(t *testing.T) {
	t.Parallel()
	with, without := repository(t), repository(t)
	write(t, with, "go.work.sum", "golang.org/x/text v0.42.0 h1:x\n")
	keyed := key(t, with, Inputs{Name: "port", Files: []string{"source", "go.work.sum"}})
	if bare := key(t, without, Inputs{Name: "port", Files: []string{"source"}}); keyed != bare {
		t.Fatalf("a go.work.sum keyed %s, none %s", keyed, bare)
	}
	write(t, with, "source/go.work.sum", "another\n")
	if key(t, with, Inputs{Name: "port", Files: []string{"source", "go.work.sum"}}) != keyed {
		t.Fatal("a go.work.sum under a named directory moved the key")
	}
	if strings.Contains(describe(with, Inputs{Name: "port", Files: []string{"go.work.sum"}}), "go.work.sum") {
		t.Fatal("the recorded inputs name a go.work.sum the key doesn't read")
	}
}

// One tool installed under two prefixes reports the same, and a note on standard error (go downloading the tree's
// release the first time) never enters its report (#t37sw0f).
func TestAToolsReportNamesTheToolNotItsPlace(t *testing.T) {
	t.Parallel()
	report := func(prefix, note string) string {
		bin := filepath.Join(prefix, "bin")
		if err := os.MkdirAll(bin, 0o755); err != nil {
			t.Fatal(err)
		}
		script := "#!/bin/sh\necho 'fakecc version 20.1.8'\necho \"resources: " + prefix + "/lib/fakecc/20\"\necho '" + note + "' >&2\n"
		if err := os.WriteFile(filepath.Join(bin, "fakecc"), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
		return Tool(filepath.Join(bin, "fakecc"), "--version")
	}
	workshop, runner := report(filepath.Join(t.TempDir(), "home/ahra/adamic-tools"), "fakecc: downloading 20.1.8"), report(filepath.Join(t.TempDir(), "opt/adamic-tools"), "")
	if workshop != runner || !strings.Contains(workshop, "fakecc --version: fakecc version 20.1.8") || !strings.Contains(workshop, "<fakecc>/lib/fakecc/20") {
		t.Fatalf("one tool under two prefixes reported\n%s\nand\n%s", workshop, runner)
	}
	// A tool that writes only to standard error is named by it.
	silent := filepath.Join(t.TempDir(), "quiet")
	os.WriteFile(silent, []byte("#!/bin/sh\necho 'quiet 1.0' >&2\n"), 0o755)
	if got := Tool(silent, "-v"); got != "quiet -v: quiet 1.0" {
		t.Fatalf("a tool reporting on standard error reported %q", got)
	}
}

// GOTOOLCHAIN, GOROOT and GOTOOLDIR name the release they select; a location is never valued (#t37sw0f).
func TestAGoSettingNamesTheReleaseNotThePolicy(t *testing.T) {
	t.Parallel()
	for _, same := range [][2]string{
		{goSetting("GOTOOLCHAIN", "auto", "go1.27.1"), goSetting("GOTOOLCHAIN", "local", "go1.27.1")},
		{goSetting("GOROOT", "/home/ahra/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.1.linux-amd64", "go1.27.1"), goSetting("GOROOT", "/usr/local/go", "go1.27.1")},
		{goSetting("GOMODCACHE", "/home/ahra/go/pkg/mod", "go1.27.1"), goSetting("GOMODCACHE", "/var/cache/go", "go1.27.1")},
		{goSetting("GOWORK", "", "go1.27.1"), goSetting("GOWORK", "off", "go1.27.1")},
	} {
		if same[0] != same[1] {
			t.Errorf("%q and %q differ", same[0], same[1])
		}
	}
	if goSetting("GOTOOLCHAIN", "auto", "go1.27.1") == goSetting("GOTOOLCHAIN", "auto", "go1.27.2") {
		t.Error("two releases keyed the same")
	}
	if goSetting("CGO_ENABLED", "1", "go1.27.1") == goSetting("CGO_ENABLED", "0", "go1.27.1") {
		t.Error("CGO_ENABLED left the setting the same")
	}
}

// The same clang installed under two homes is one tool, and another version is another: the report keeps what
// names the tool and drops InstalledDir, the one line that says only where it is (#79thccs).
func TestAToolsReportDoesntDependOnWhereItIsInstalled(t *testing.T) {
	t.Parallel()
	report := func(version, home string) string {
		return toolReport("clang version " + version + " (https://github.com/llvm/llvm-project 87f0227cb601)\n" +
			"Target: x86_64-unknown-linux-gnu\nThread model: posix\nInstalledDir: " + home + "/adamic-tools/llvm/bin\n")
	}
	if ahra, cloud := report("20.1.8", "/home/ahra"), report("20.1.8", "/root"); ahra != cloud {
		t.Fatalf("one clang under two homes reported\n%s\nand\n%s", ahra, cloud)
	}
	if report("20.1.8", "/home/ahra") == report("20.1.9", "/home/ahra") {
		t.Fatal("two clang versions reported the same")
	}
	if got := report("20.1.8", "/home/ahra"); strings.Contains(got, "InstalledDir") || !strings.Contains(got, "Target: x86_64-unknown-linux-gnu") || !strings.Contains(got, "Thread model: posix") {
		t.Fatalf("the report is %q", got)
	}
	if got := toolReport("go version go1.27.1 linux/amd64\n"); got != "go version go1.27.1 linux/amd64" {
		t.Fatalf("a one-line report became %q", got)
	}
}

// Not parallel: Tool writes the package-level tools map.
func TestToolNamesItselfOnce(t *testing.T) {
	if value := Tool("go", "version"); !strings.HasPrefix(value, "go version: go version go") {
		t.Fatalf("Tool reported %q", value)
	}
	if value := Tool("adamic-no-such-tool"); !strings.Contains(value, "adamic-no-such-tool: ") || !strings.Contains(value, "(") {
		t.Fatalf("a missing tool reported %q", value)
	}
}
