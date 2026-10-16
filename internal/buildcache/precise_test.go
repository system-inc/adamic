package buildcache

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math/rand/v2"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

// A rig is a traced Workshop in miniature (#vt46geg): a git checkout of its own as the repository, a build cache, and
// a trace directory the products build under. The trace itself is written by the test, as strace -f would print it,
// so each test says exactly what each build read.
type rig struct {
	root, cache, trace string
	pid                int
	lines              []string
	// before is the tree's state when the run began; empty, the tree is taken as unchanged.
	before string
}

// Not parallel: the rig changes the working directory, repositoryRoot and the build environment.
func newRig(t *testing.T) *rig {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"go.mod": "module github.com/system-inc/adamic\n\ngo 1.27\n", "source/a.c": "int a;\n", "source/b.c": "int b;\n",
		"include/x.h": "#define X 1\n", "notes.txt": "a note\n", "go.sum": "",
	} {
		write(t, root, name, content)
	}
	gitIn(t, root, "init", "-q")
	gitIn(t, root, "add", ".")
	gitIn(t, root, "commit", "-q", "-m", "tree")
	inRepository(t, root)
	cache, _ := cached(t)
	cache, _ = filepath.EvalSymlinks(cache)
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", cache)
	trace := filepath.Join(t.TempDir(), fmt.Sprintf("trace-%s", strings.ReplaceAll(t.Name(), "/", "-")))
	if err := os.MkdirAll(trace, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_BUILD_TRACE", trace)
	original, originalLean := recipeFor, leanRecipeFor
	recipeFor = func(string, journalEntry) recipe {
		return recipe{entries: []readEntry{{Kind: "content", Path: "go.mod"}}}
	}
	leanRecipeFor = func(string, string, journalEntry) recipe {
		return recipe{entries: []readEntry{{Kind: "content", Path: "go.mod"}}}
	}
	t.Cleanup(func() { recipeFor, leanRecipeFor = original, originalLean })
	return &rig{root: root, cache: cache, trace: trace, pid: os.Getpid()}
}

// linkDeclaredTools links each tool the declared tools directory (ADAMIC_TOOLS) holds into bin, a tools directory a
// test makes to stand a fake tool in, but for a name bin already holds: pointed at that directory, toolchain.Path still
// finds every other tool where it found it before, and a sandbox (ADAMIC_SANDBOX=1) still finds them at all. A name in
// both bin and go/bin is bin's, as Path finds it.
func linkDeclaredTools(t *testing.T, bin string) {
	t.Helper()
	declared := os.Getenv("ADAMIC_TOOLS")
	if declared == "" {
		return
	}
	for _, directory := range []string{filepath.Join(declared, "bin"), filepath.Join(declared, "go", "bin")} {
		entries, _ := os.ReadDir(directory)
		for _, entry := range entries {
			link := filepath.Join(bin, entry.Name())
			if _, err := os.Lstat(link); err == nil {
				continue
			}
			if err := os.Symlink(filepath.Join(directory, entry.Name()), link); err != nil {
				t.Fatal(err)
			}
		}
	}
}

// gitIn runs git in directory, never starting git's background auto-gc or maintenance: a detached one still writing
// .git/objects when the test ends fails t.TempDir's cleanup (TestVouchingIsALookupAtTheCorpusSize, whose commit of a
// corpus starts one: `.git/objects: directory not empty`).
func gitIn(t *testing.T, directory string, arguments ...string) {
	t.Helper()
	command := toolchain.GitCommand(append([]string{"-C", directory, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "init.defaultBranch=main", "-c", "gc.auto=0", "-c", "maintenance.auto=false"}, arguments...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", arguments, err, output)
	}
}

// forgetTracked drops what git ls-files said, after a test changes what a repository tracks.
func forgetTracked() {
	repositories.Range(func(key, _ any) bool { repositories.Delete(key); return true })
}

// forgetHeldForLife drops what a process whose tree can't change under it holds for its life (readTree, readProducts),
// as a new process starts with nothing: a runner's in read mode, or a traced one, which every rig's trace makes this
// test process.
func forgetHeldForLife() {
	readTree.Clear()
	readProducts.Clear()
}

func (r *rig) add(pid int, format string, arguments ...any) {
	r.lines = append(r.lines, fmt.Sprintf("%d ", pid)+fmt.Sprintf(format, arguments...))
}

func (r *rig) path(name string) string {
	if filepath.IsAbs(name) {
		return name
	}
	return filepath.Join(r.root, name)
}

func (r *rig) open(pid int, name string) {
	r.add(pid, `openat(AT_FDCWD, %q, O_RDONLY|O_CLOEXEC) = 3<%s>`, r.path(name), r.path(name))
}

func (r *rig) write(pid int, name string) {
	r.add(pid, `openat(AT_FDCWD, %q, O_WRONLY|O_CREAT|O_TRUNC|O_CLOEXEC, 0644) = 3<%s>`, r.path(name), r.path(name))
}

func (r *rig) missing(pid int, name string) {
	r.add(pid, `newfstatat(AT_FDCWD, %q, 0x7ffd5e1c, 0) = -1 ENOENT (No such file or directory)`, r.path(name))
}

func (r *rig) list(pid int, name string) {
	r.add(pid, `getdents64(3<%s>, 0x55d0 /* 4 entries */, 32768) = 112`, r.path(name))
}

func (r *rig) marker(pid int, event, id string) {
	r.add(pid, `access("/adamic-trace/%s/%s", F_OK) = -1 ENOENT (No such file or directory)`, event, id)
}

// built is every build the run journaled, by product name.
func (r *rig) built(t *testing.T) map[string]journalEntry {
	t.Helper()
	entries, err := readJournal(r.trace)
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]journalEntry{}
	for _, entry := range entries {
		if entry.Event == "built" {
			found[entry.Name] = entry
		}
	}
	return found
}

// window wraps reads in a build's markers.
func (r *rig) window(id string, reads func()) {
	r.marker(r.pid, "build-begin", id)
	reads()
	r.marker(r.pid, "build-end", id)
}

func (r *rig) settle(t *testing.T) Settlement {
	t.Helper()
	t.Chdir(r.root)
	before := r.before
	if before == "" {
		var err error
		if before, err = TreeState(); err != nil {
			t.Fatal(err)
		}
	}
	settlement, err := Settle(r.trace, strings.NewReader(strings.Join(r.lines, "\n")+"\n"), before)
	if err != nil {
		t.Fatal(err)
	}
	r.lines = nil
	return settlement
}

func building(content string) func(directory string) error {
	return func(directory string) error {
		return os.WriteFile(filepath.Join(directory, "product"), []byte(content), 0o644)
	}
}

// reading looks a product up as a runner does: from the read sets in the cache, never building, in a process of its
// own, holding nothing an earlier lookup read (a test changes the tree between them, which a runner's never does).
func reading(t *testing.T, inputs Inputs) (string, error) {
	t.Helper()
	forgetHeldForLife()
	t.Setenv("ADAMIC_BUILD_CACHE", "read")
	t.Setenv("ADAMIC_BUILD_TRACE", "")
	defer t.Setenv("ADAMIC_BUILD_CACHE", "")
	return Get(inputs, func(string) error { t.Fatal("read mode called the build"); return nil })
}

var port = Inputs{Name: "port", Files: []string{"source", "include"}, Flags: []string{"-DADAMIC"}, Toolchain: []string{"clang 20.1.8"}}

// A traced build is keyed by what it read: a change to a file it read misses, a change to one it didn't hits, and the
// declared Files are held against the reads, not used as the key.
// Not parallel: newRig.
func TestATracedBuildIsKeyedByWhatItRead(t *testing.T) {
	r := newRig(t)
	pending := Product(t, port, building("built"))
	if !strings.Contains(pending, filepath.Join("pending", filepath.Base(r.trace))) {
		t.Fatalf("a traced build placed its product at %s before it was settled", pending)
	}
	build := r.built(t)["port"]
	r.window(build.Build, func() {
		r.open(r.pid, "source/a.c")
		r.list(r.pid, "source")
		r.missing(r.pid, "include/missing.h")
	})
	settlement := r.settle(t)
	if len(settlement.Settled) != 1 || len(settlement.Refused) != 0 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	if !strings.Contains(settlement.Settled[0], "declared Files: 0 reads undeclared, 1 declared unread") {
		t.Fatalf("the drift isn't reported: %s", settlement.Settled[0])
	}
	if _, err := os.Stat(pending); err == nil {
		t.Fatal("the pending product stayed after it was settled")
	}
	file, err := loadReads(r.cache, build.NameKey)
	if err != nil || len(file.Sets) != 1 {
		t.Fatalf("read sets: %+v, %v", file, err)
	}
	var kinds []string
	for _, entry := range file.Sets[0].Entries {
		kinds = append(kinds, entry.Kind+" "+entry.Path)
	}
	if want := []string{"settings binary", "content go.mod", "exists include/missing.h", "listing source", "content source/a.c"}; !slices.Equal(kinds, want) {
		t.Fatalf("the read set is %q, want %q", kinds, want)
	}
	product, err := reading(t, port)
	if err != nil || product != filepath.Join(r.cache, file.Sets[0].Key) {
		t.Fatalf("a runner found %q, %v; want %s", product, err, file.Sets[0].Key)
	}
	if content, _ := os.ReadFile(filepath.Join(product, "product")); string(content) != "built" {
		t.Fatalf("the product reads %q", content)
	}
	// Outside the read set: a tracked file nobody read, one the build listed but never opened, an untracked file where
	// the build looked and missed. Each is a hit.
	for _, change := range []func(){
		func() { write(t, r.root, "notes.txt", "another note\n") },
		func() { write(t, r.root, "source/b.c", "int b2;\n") },
		func() { write(t, r.root, "include/x.h", "#define X 2\n") },
		func() { write(t, r.root, "include/missing.h", "untracked\n") },
	} {
		change()
		if found, err := reading(t, port); err != nil || found != product {
			t.Fatalf("a change outside the read set missed: %q, %v", found, err)
		}
	}
	// Inside it: the content of a file it read, a name in a directory it listed, a tracked file where it missed.
	for _, change := range []struct {
		name, named string
		change      func()
	}{
		{"content", "content source/a.c", func() { write(t, r.root, "source/a.c", "int a2;\n") }},
		{"listing", "listing source", func() { write(t, r.root, "source/c.c", ""); gitIn(t, r.root, "add", "source/c.c") }},
		{"absent", "exists include/missing.h", func() {
			write(t, r.root, "include/missing.h", "now tracked\n")
			gitIn(t, r.root, "add", "include/missing.h")
		}},
	} {
		gitIn(t, r.root, "reset", "-q", "--hard")
		gitIn(t, r.root, "clean", "-q", "-f", "-d")
		forgetTracked()
		if found, err := reading(t, port); err != nil || found != product {
			t.Fatalf("%s: the tree before the change missed: %q, %v", change.name, found, err)
		}
		change.change()
		forgetTracked()
		_, err := reading(t, port)
		if !errors.Is(err, ErrNotBuilt) || !strings.Contains(err.Error(), change.named) {
			t.Fatalf("%s: a change in the read set read as %v; want a miss naming %q", change.name, err, change.named)
		}
	}
}

// A build that read something no key can name is refused, named, and neither placed nor recorded; the same build with
// only nameable reads is settled and recorded, so the refusal is what stops it. Neither writes to the store.
// Not parallel: newRig.
func TestAReadNoKeyCanNameIsRefused(t *testing.T) {
	home, _ := os.UserHomeDir()
	elsewhere := filepath.Join(home, ".adamic-not-keyable")
	for name, plant := range map[string]struct {
		reason string
		read   func(r *rig)
	}{
		"untracked":          {"untracked: scratch.txt", func(r *rig) { write(t, r.root, "scratch.txt", "x"); r.open(r.pid, "scratch.txt") }},
		".git":               {"git: .git/HEAD", func(r *rig) { r.open(r.pid, ".git/HEAD") }},
		"home":               {"home: " + elsewhere, func(r *rig) { r.open(r.pid, elsewhere) }},
		"a tree write":       {"tree write: source/a.c", func(r *rig) { r.write(r.pid, "source/a.c") }},
		"temp not made":      {"temp not made by this run", func(r *rig) { r.open(r.pid, "/tmp/adamic-left-by-someone") }},
		"another place":      {"other: /opt/elsewhere/lib.h", func(r *rig) { r.open(r.pid, "/opt/elsewhere/lib.h") }},
		"an unknown product": {"unknown product", func(r *rig) { r.open(r.pid, filepath.Join(r.cache, strings.Repeat("ab", 32), "main.ts")) }},
		"a local product":    {"local product", func(r *rig) { r.open(r.pid, filepath.Join(r.cache, "local", strings.Repeat("cd", 32), "x")) }},
		"nothing":            {"", func(r *rig) {}},
	} {
		t.Run(name, func(t *testing.T) {
			r := newRig(t)
			store := &fakeStore{objects: map[string][]byte{}}
			server := httptest.NewServer(store)
			t.Cleanup(server.Close)
			t.Setenv("ADAMIC_BUILD_STORE", "traced="+server.URL)
			pending := Product(t, port, building("built"))
			build := r.built(t)["port"]
			r.window(build.Build, func() {
				r.open(r.pid, "source/a.c")
				plant.read(r)
			})
			settlement := r.settle(t)
			if plant.reason == "" {
				if file, _ := loadReads(r.cache, build.NameKey); len(settlement.Settled) != 1 || len(file.Sets) != 1 || len(store.writes) != 0 {
					t.Fatalf("a nameable build: settled %q, refused %q, read sets %+v, wrote %q", settlement.Settled, settlement.Refused, file.Sets, store.writes)
				}
				return
			}
			if len(settlement.Refused) != 1 || !strings.Contains(settlement.Refused[0], plant.reason) || len(settlement.Settled) != 0 {
				t.Fatalf("settled %q, refused %q; want a refusal naming %q", settlement.Settled, settlement.Refused, plant.reason)
			}
			if len(store.writes) != 0 {
				t.Fatalf("a refused build wrote to the store: %q", store.writes)
			}
			if _, err := os.Stat(pending); err == nil {
				t.Fatal("a refused product stayed in the cache")
			}
			if file, _ := loadReads(r.cache, build.NameKey); len(file.Sets) != 0 {
				t.Fatalf("a refused build recorded a read set: %+v", file.Sets)
			}
			entries, _ := os.ReadDir(r.cache)
			for _, entry := range entries {
				if isKey(entry.Name()) {
					t.Fatalf("a refused build placed %s", entry.Name())
				}
			}
		})
	}
}

// typescript-go's osvfs stats its own executable with every letter's case swapped as it starts, in every process that
// links it (Workshop, Oct 10: all 19 product-test packages were refused for it): that lookup is the file system's case
// sensitivity, keyed as each machine answers it, never refused as a path outside every place. A path in another case
// that no program of the run was is still the path it names. Mutants: caseProbe always false; record not filling ran
// (the program the build made and ran is refused); caseProbe not reading the journal's executables; caseSensitivity a
// constant.
// Not parallel: newRig.
func TestTheFileSystemsCaseProbeIsKeyedNotRefused(t *testing.T) {
	r := newRig(t)
	// A process journals itself once, into the one trace it runs under; this rig's trace is this process's now.
	journalProcess = sync.Once{}
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	directory, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	program := filepath.Join(directory, "oracle")
	r.window(build.Build, func() {
		r.open(r.pid, "source/a.c")
		r.missing(r.pid, swapCase(executable))
		r.write(r.pid, program)
		r.add(r.pid, `execve(%q, [%q], 0x7ffd5e1c /* 3 vars */) = 0`, program, program)
		r.missing(r.pid, swapCase(program))
	})
	settlement := r.settle(t)
	if len(settlement.Settled) != 1 || len(settlement.Refused) != 0 {
		t.Fatalf("a build whose programs asked the file system's case: settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	file, err := loadReads(r.cache, build.NameKey)
	if err != nil || len(file.Sets) != 1 || !slices.ContainsFunc(file.Sets[0].Entries, func(entry readEntry) bool { return entry.Kind == "filesystem" && entry.Path == "case" }) {
		t.Fatalf("the read set doesn't name the file system's case: %+v, %v", file.Sets, err)
	}
	if found, err := reading(t, port); err != nil || found != filepath.Join(r.cache, file.Sets[0].Key) {
		t.Fatalf("a runner on the same machine found %q, %v", found, err)
	}

	r = newRig(t)
	Product(t, port, building("built"))
	build = r.built(t)["port"]
	elsewhere := strings.ToUpper("/opt/elsewhere/lib.h")
	r.window(build.Build, func() { r.missing(r.pid, elsewhere) })
	if settlement := r.settle(t); len(settlement.Settled) != 1 {
		t.Fatalf("a path in another case that nothing ran: settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	if file, _ := loadReads(r.cache, build.NameKey); len(file.Sets) != 1 || !slices.Contains(sortedEntries(file.Sets[0].Entries), readEntry{Kind: "absolute", Path: elsewhere}) ||
		slices.ContainsFunc(file.Sets[0].Entries, func(entry readEntry) bool { return entry.Kind == "filesystem" }) {
		t.Fatalf("a path in another case that nothing ran isn't the path it names: %+v", file.Sets)
	}

	// The value is the file system's own answer: a file made here, asked for in the other case.
	probe := filepath.Join(directory, "Probe")
	write(t, directory, "Probe", "")
	want := "sensitive"
	if _, err := os.Stat(swapCase(probe)); err == nil {
		want = "insensitive"
	}
	if got := caseSensitivity(); got != want {
		t.Fatalf("the file system's case reads %q; a file made here says %q", got, want)
	}
	for _, entry := range file.Sets[0].Entries {
		if entry.Kind == "filesystem" && entry.Value != shortValue(want) {
			t.Fatalf("the read set recorded the file system's case as %s, not %q (%s)", entry.Value, want, shortValue(want))
		}
	}
}

// Every go command reads and counts into Go's telemetry in the home directory as it runs, and GOENV=off doesn't stop it
// (Workshop, Oct 10: every product was refused for ~/.config/go/telemetry/local/weekends). It is go's bookkeeping, as
// its build cache is: no entry, never a refusal. Content elsewhere in the home directory is still refused. Mutants: the
// telemetry directory not classed; the walk down to it not classed (its lookups keyed as home entries).
// Not parallel: newRig.
func TestGosTelemetryIsGosOwnBookkeeping(t *testing.T) {
	configuration, err := os.UserConfigDir()
	if err != nil {
		t.Skip("no configuration directory here:", err)
	}
	telemetry := filepath.Join(configuration, "go", "telemetry")
	r := newRig(t)
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	r.window(build.Build, func() {
		r.open(r.pid, "source/a.c")
		r.missing(r.pid, filepath.Join(configuration, "go"))
		r.open(r.pid, filepath.Join(telemetry, "local", "weekends"))
		r.write(r.pid, filepath.Join(telemetry, "local", "go@go1.27.1-go1.27.1-linux-amd64-2026-10-11.v1.count"))
	})
	settlement := r.settle(t)
	if len(settlement.Settled) != 1 || len(settlement.Refused) != 0 {
		t.Fatalf("a build whose go read its telemetry: settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	file, err := loadReads(r.cache, build.NameKey)
	if err != nil || len(file.Sets) != 1 {
		t.Fatalf("read sets: %+v, %v", file, err)
	}
	for _, entry := range file.Sets[0].Entries {
		if entry.Kind == "home" {
			t.Fatalf("go's telemetry keyed the product: %+v", entry)
		}
	}
	r = newRig(t)
	Product(t, port, building("built"))
	build = r.built(t)["port"]
	beside := filepath.Join(configuration, "go", "env")
	r.window(build.Build, func() { r.open(r.pid, beside) })
	if settlement := r.settle(t); len(settlement.Refused) != 1 || !strings.Contains(settlement.Refused[0], beside) {
		t.Fatalf("a configuration file beside the telemetry: settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
}

// Three reads Workshop's native products made that no class held (Oct 11, TestProduct_ExpressionsNative000 under
// cmd/traced): a lookup of a path outside every place (the module resolver's /adamic-prelude/package.json, clang's
// /opt/rh), keyed by whether and as what it exists; clang run through a link to the tools directory (~/.adamic-tools),
// keyed as the tool it resolves to; and go's os.RemoveAll opening the temporary directory without O_DIRECTORY, a
// directory with no content to read. Reading what a path outside every place holds is still refused. Mutants: looked
// paths outside every place refused; links not resolved; a temporary directory's open read as content.
// Not parallel: newRig.
func TestWhatABuildOnlyLooksForOrRunsThroughALinkIsKeyed(t *testing.T) {
	tools, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	write(t, tools, "bin/clang", "#!/bin/sh\necho clang version 20.1.8\n")
	if err := os.Chmod(filepath.Join(tools, "bin", "clang"), 0o755); err != nil {
		t.Fatal(err)
	}
	// The rig's git, and the key's, are toolchain's, found where they were before this tools directory took over.
	linkDeclaredTools(t, filepath.Join(tools, "bin"))
	t.Setenv("ADAMIC_TOOLS", tools)
	scratch, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	dotted := filepath.Join(scratch, ".adamic-tools")
	if err := os.Symlink(tools, dotted); err != nil {
		t.Fatal(err)
	}
	r := newRig(t)
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	r.window(build.Build, func() {
		r.open(r.pid, "source/a.c")
		r.missing(r.pid, "/adamic-prelude/package.json")
		r.missing(r.pid, "/opt/rh")
		r.add(r.pid, `execve(%q, [%q], 0x7ffd5e1c /* 3 vars */) = 0`, filepath.Join(dotted, "bin", "clang"), "clang")
		r.add(r.pid, `openat(AT_FDCWD, %q, O_RDONLY|O_CLOEXEC) = 8<%s>`, scratch, scratch)
	})
	settlement := r.settle(t)
	if len(settlement.Settled) != 1 || len(settlement.Refused) != 0 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	file, err := loadReads(r.cache, build.NameKey)
	if err != nil || len(file.Sets) != 1 {
		t.Fatalf("read sets: %+v, %v", file, err)
	}
	for _, want := range []readEntry{{Kind: "absolute", Path: "/adamic-prelude/package.json"}, {Kind: "absolute", Path: "/opt/rh"}, {Kind: "tool", Path: "bin/clang"}} {
		if !slices.Contains(sortedEntries(file.Sets[0].Entries), want) {
			t.Fatalf("the read set doesn't hold %+v: %+v", want, file.Sets[0].Entries)
		}
	}
	if found, err := reading(t, port); err != nil || found != filepath.Join(r.cache, file.Sets[0].Key) {
		t.Fatalf("a runner on the same machine found %q, %v", found, err)
	}

	r = newRig(t)
	Product(t, port, building("built"))
	build = r.built(t)["port"]
	r.window(build.Build, func() { r.open(r.pid, "/opt/elsewhere/lib.h") })
	if settlement := r.settle(t); len(settlement.Refused) != 1 || !strings.Contains(settlement.Refused[0], "other: /opt/elsewhere/lib.h") {
		t.Fatalf("content outside every place: settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
}

// The native runtime cache holds each compiled runtime under the address of everything its compile read
// (internal/native's runtimeKey, system headers included), as Go's build cache does, so a native build reading a
// runtime's headers and archive there is keyed by nothing more (Workshop, Oct 11: typescript-parser-native was refused
// for ~/.cache/adamic/runtime/<key>/adamic.h). Content elsewhere in the user cache directory is still refused.
// Mutants: the runtime cache not classed; the walk down to it not classed.
// Not parallel: newRig.
func TestTheNativeRuntimeCacheIsAddressedByWhatItsCompileRead(t *testing.T) {
	user, err := os.UserCacheDir()
	if err != nil {
		t.Skip("no user cache directory here:", err)
	}
	runtimes := filepath.Join(user, "adamic", "runtime")
	runtime := filepath.Join(runtimes, strings.Repeat("ab", 32))
	r := newRig(t)
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	r.window(build.Build, func() {
		r.open(r.pid, "source/a.c")
		r.missing(r.pid, filepath.Join(user, "adamic"))
		r.open(r.pid, filepath.Join(runtime, "adamic.h"))
		r.open(r.pid, filepath.Join(runtime, "runtime.a"))
	})
	settlement := r.settle(t)
	if len(settlement.Settled) != 1 || len(settlement.Refused) != 0 {
		t.Fatalf("a native build reading its runtime: settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	file, err := loadReads(r.cache, build.NameKey)
	if err != nil || len(file.Sets) != 1 {
		t.Fatalf("read sets: %+v, %v", file, err)
	}
	for _, entry := range file.Sets[0].Entries {
		if entry.Kind == "home" || entry.Kind == "absolute" {
			t.Fatalf("the runtime cache keyed the product: %+v", entry)
		}
	}
	r = newRig(t)
	Product(t, port, building("built"))
	build = r.built(t)["port"]
	beside := filepath.Join(user, "adamic", "elsewhere", "x.h")
	r.window(build.Build, func() { r.open(r.pid, beside) })
	if settlement := r.settle(t); len(settlement.Refused) != 1 || !strings.Contains(settlement.Refused[0], beside) {
		t.Fatalf("content beside the runtime cache: settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
}

// A build whose markers aren't both in the trace is a lost trace: refused, and so is every build of a trace that
// holds a process it never placed under a parent.
// Not parallel: newRig.
func TestALostTraceRefuses(t *testing.T) {
	r := newRig(t)
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	r.marker(r.pid, "build-begin", build.Build)
	r.open(r.pid, "source/a.c")
	if settlement := r.settle(t); len(settlement.Refused) != 1 || !strings.Contains(settlement.Refused[0], "lost trace") {
		t.Fatalf("a build that never ended: settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	r = newRig(t)
	Product(t, port, building("built"))
	build = r.built(t)["port"]
	r.window(build.Build, func() {
		r.open(r.pid, "source/a.c")
		r.open(r.pid+77, "include/x.h")
	})
	if settlement := r.settle(t); len(settlement.Refused) != 1 || !strings.Contains(settlement.Refused[0], "never placed under a parent") {
		t.Fatalf("a process with no parent: settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
}

// What a key reads is the key's: a thread between key-begin and key-end, and the processes it starts, can read .git
// (git ls-files) without the build being refused, and their reads never enter the read set. A child process the build
// starts is the build's, even when its calls print before its parent's fork returns, and so is a thread's.
// Not parallel: newRig.
func TestReadsAreTheBuildsOrTheKeys(t *testing.T) {
	r := newRig(t)
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	keyer, git, compiler, thread, server := r.pid+1, r.pid+2, r.pid+3, r.pid+4, r.pid+5
	r.add(r.pid, `clone(child_stack=0x7f, flags=CLONE_VM|CLONE_FS|CLONE_FILES|CLONE_SIGHAND|CLONE_THREAD|CLONE_SYSVSEM) = %d`, keyer)
	r.marker(keyer, "key-begin", "-")
	r.open(keyer, ".git/index")
	r.open(keyer, "notes.txt")
	r.add(keyer, `clone3({flags=CLONE_VM|CLONE_VFORK, exit_signal=SIGCHLD, stack=0x7f, stack_size=0x9000}, 88) = %d`, git)
	r.open(git, ".git/HEAD")
	r.add(git, "+++ exited with 0 +++")
	r.marker(keyer, "key-end", "-")
	r.window(build.Build, func() {
		r.add(r.pid, `clone3({flags=CLONE_VM|CLONE_VFORK, exit_signal=SIGCHLD, stack=0x7f, stack_size=0x9000}, 88 <unfinished ...>`)
		r.open(compiler, "include/x.h")
		r.add(r.pid, `<... clone3 resumed>) = %d`, compiler)
		r.add(compiler, `openat(AT_FDCWD<%s>, "source/a.c", O_RDONLY) = 3<%s>`, r.root, r.path("source/a.c"))
		r.add(compiler, "+++ exited with 0 +++")
		r.add(r.pid, `clone(child_stack=0x7f, flags=CLONE_VM|CLONE_THREAD) = %d`, thread)
		r.open(thread, "source/b.c")
		// A process the build started that is still running when the build ends (a server it talks to).
		r.add(r.pid, `clone(child_stack=NULL, flags=SIGCHLD) = %d`, server)
		r.open(server, "go.sum")
	})
	settlement := r.settle(t)
	if len(settlement.Settled) != 1 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	file, _ := loadReads(r.cache, build.NameKey)
	var paths []string
	for _, entry := range file.Sets[0].Entries {
		paths = append(paths, entry.Path)
	}
	if want := []string{"binary", "go.mod", "go.sum", "include/x.h", "source/a.c", "source/b.c"}; !slices.Equal(paths, want) {
		t.Fatalf("the read set holds %q, want %q (the child's and the thread's reads, never the key's)", paths, want)
	}
}

// A product that reads another is keyed by that product's key on the same tree: a change the other product read
// misses both, though the reader never read the changed file itself.
// Not parallel: newRig.
func TestAProductReadingAnotherIsKeyedByItsKey(t *testing.T) {
	r := newRig(t)
	header := Inputs{Name: "header", Files: []string{"include"}}
	headerProduct := Product(t, header, building("header"))
	Product(t, port, func(directory string) error {
		content, err := os.ReadFile(filepath.Join(headerProduct, "product"))
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "product"), append(content, " and port"...), 0o644)
	})
	built := r.built(t)
	// header is built in port's own process, which writes header's product and then reads it: a product is an input
	// whoever wrote it.
	r.window(built["header"].Build, func() {
		r.open(r.pid, "include/x.h")
		r.write(r.pid, filepath.Join(headerProduct, "product"))
	})
	r.window(built["port"].Build, func() {
		r.open(r.pid, filepath.Join(headerProduct, "product"))
		r.open(r.pid, "source/a.c")
	})
	if settlement := r.settle(t); len(settlement.Settled) != 2 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	file, _ := loadReads(r.cache, built["port"].NameKey)
	if !slices.ContainsFunc(file.Sets[0].Entries, func(entry readEntry) bool { return entry.Kind == "product" && entry.Path == built["header"].NameKey }) {
		t.Fatalf("port's read set doesn't name header's product: %+v", file.Sets[0].Entries)
	}
	if _, err := reading(t, port); err != nil {
		t.Fatalf("port before the change: %v", err)
	}
	write(t, r.root, "include/x.h", "#define X 2\n")
	_, err := reading(t, port)
	if !errors.Is(err, ErrNotBuilt) || !strings.Contains(err.Error(), "reads product "+built["header"].NameKey[:12]) {
		t.Fatalf("a change header read left port %v", err)
	}
	// header rebuilt on the new tree has a new key, so port, built from the old header, still misses.
	r.retrace(t, "header")
	Product(t, header, building("header 2"))
	r.window(r.built(t)["header"].Build, func() { r.open(r.pid, "include/x.h") })
	if settlement := r.settle(t); len(settlement.Settled) != 1 {
		t.Fatalf("header's rebuild: settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	if _, err := reading(t, header); err != nil {
		t.Fatalf("header after its rebuild: %v", err)
	}
	if _, err = reading(t, port); !errors.Is(err, ErrNotBuilt) || !strings.Contains(err.Error(), "product "+built["header"].NameKey) {
		t.Fatalf("port, built from the old header, read as %v", err)
	}
}

// Away from a trace a product is built as before, for this machine alone: keyed by its declared Files and never found by a runner's read mode. Only Workshop's tree builder, ADAMIC_BUILD_STORE=traced, refuses to build
// outside a trace.
// Not parallel: newRig.
func TestAnUntracedMachineBuildsOnlyForItself(t *testing.T) {
	newRig(t)
	t.Setenv("ADAMIC_BUILD_TRACE", "")
	for _, setting := range []string{"", "off", "http://127.0.0.1:1"} {
		t.Setenv("ADAMIC_BUILD_STORE", setting)
		t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
		local := Product(t, port, building("local"))
		if !strings.Contains(local, string(filepath.Separator)+"local"+string(filepath.Separator)) {
			t.Fatalf("ADAMIC_BUILD_STORE=%q: a local build is at %s", setting, local)
		}
		never := func(string) error { t.Fatal("a local product was built twice"); return nil }
		if again := Product(t, port, never); again != local {
			t.Fatalf("a local hit found %s, want %s", again, local)
		}
		if _, err := reading(t, port); !errors.Is(err, ErrNotBuilt) || !strings.Contains(err.Error(), "no traced build") {
			t.Fatalf("read mode found a local product: %v", err)
		}
	}
	t.Setenv("ADAMIC_BUILD_STORE", "traced=http://127.0.0.1:1")
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
	if _, err := Get(port, func(string) error { t.Fatal("the tree builder built untraced"); return nil }); !errors.Is(err, ErrUntraced) {
		t.Fatalf("the tree builder outside a trace: %v, want ErrUntraced", err)
	}
}

// The recipe is the binary's linked packages: this test binary's own package by its files, test files included, and
// listing, a dependency by the files compiled into the binary and never its test files (#c2chyab), its module's
// go.mod and go.sum; a binary that names no package can't be keyed. Mutant: every package's test files taken.
// Not parallel: go list runs in the repository.
func TestTheRecipeIsTheBinarysLinkedPackages(t *testing.T) {
	root := repositoryRootForTest(t)
	found := recipeOf(root, journalEntry{Package: "github.com/system-inc/adamic/internal/buildcache.test"})
	if found.reason != "" {
		t.Fatal(found.reason)
	}
	for _, want := range []readEntry{{Kind: "content", Path: "internal/buildcache/reads.go"}, {Kind: "content", Path: "internal/buildcache/precise_test.go"},
		{Kind: "listing", Path: "internal/buildcache"}, {Kind: "content", Path: "go.mod"}, {Kind: "content", Path: "internal/toolchain/toolchain.go"}} {
		if !slices.Contains(found.entries, want) {
			t.Errorf("the recipe lacks %+v", want)
		}
	}
	if test := (readEntry{Kind: "content", Path: "internal/toolchain/toolchain_test.go"}); slices.Contains(found.entries, test) {
		t.Errorf("the recipe holds a dependency's test file, %s, which the binary never compiled", test.Path)
	}
	for _, unknown := range []string{"", "command-line-arguments"} {
		if refused := recipeOf(root, journalEntry{Package: unknown}); refused.reason == "" {
			t.Errorf("a binary built from %q has a recipe", unknown)
		}
	}
}

// The recipe is the same wherever Go's build cache sits (#pqq9z5q): go list -test names the test main it generated by
// its absolute place in that cache, which is go's, made from the test files the recipe keys, never a tree path. Joined
// to the package's directory it was one (internal/buildcache/<GOCACHE>/…-d, valued absent), so every product a
// builder settled keyed its Go cache's location: 1,271 of 1,271 in Server's comparison against Workshop. Mutant: the
// absolute name joined to the package's directory again.
// Not parallel: go list runs in the repository, under GOCACHE set through t.Setenv.
func TestTheRecipeIsTheSameWhereverGosBuildCacheSits(t *testing.T) {
	root := repositoryRootForTest(t)
	var recipes [][]readEntry
	caches := []string{t.TempDir(), filepath.Join(t.TempDir(), "elsewhere", "gocache")}
	for _, cache := range caches {
		t.Setenv("GOCACHE", cache)
		found := recipeOf(root, journalEntry{Package: "github.com/system-inc/adamic/internal/buildcache.test"})
		if found.reason != "" {
			t.Fatal(found.reason)
		}
		for _, entry := range found.entries {
			for _, place := range caches {
				for _, spelling := range variants(place) {
					if strings.Contains(entry.Path, strings.TrimPrefix(filepath.ToSlash(spelling), "/")) {
						t.Errorf("the recipe under GOCACHE=%s names a Go cache's place: %+v", cache, entry)
					}
				}
			}
		}
		recipes = append(recipes, sortedEntries(found.entries))
	}
	if !slices.Equal(recipes[0], recipes[1]) {
		t.Fatalf("one binary's recipe under two GOCACHE places: %d and %d entries, differing", len(recipes[0]), len(recipes[1]))
	}
}

// retrace starts another traced run in the same rig: a new trace directory, as cmd/traced makes for each run.
func (r *rig) retrace(t *testing.T, name string) {
	t.Helper()
	r.trace = filepath.Join(filepath.Dir(r.trace), filepath.Base(r.trace)+"-"+name)
	if err := os.MkdirAll(r.trace, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_BUILD_TRACE", r.trace)
	t.Setenv("ADAMIC_BUILD_CACHE", "")
}

func (r *rig) entries(t *testing.T, nameKey string) []string {
	t.Helper()
	file, err := loadReads(r.cache, nameKey)
	if err != nil || len(file.Sets) == 0 {
		t.Fatalf("no read set for %s: %v", nameKey[:12], err)
	}
	var found []string
	for _, entry := range file.Sets[0].Entries {
		found = append(found, entry.Kind+" "+entry.Path)
	}
	return found
}

// A read through a symbolic link reads the link (where it points) and what it resolves to, by the descriptor's
// resolved path; a symbolic link to a directory keys as a link, never as absent; a link out of the tree is classed as
// where it lands.
// Not parallel: newRig.
func TestAReadThroughASymbolicLinkKeysTheLinkAndItsTarget(t *testing.T) {
	r := newRig(t)
	os.Symlink("source/a.c", filepath.Join(r.root, "alias.c"))
	os.Symlink("source", filepath.Join(r.root, "lib"))
	gitIn(t, r.root, "add", "alias.c", "lib")
	gitIn(t, r.root, "commit", "-q", "-m", "links")
	forgetTracked()
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	r.window(build.Build, func() {
		r.add(r.pid, `openat(AT_FDCWD, %q, O_RDONLY|O_CLOEXEC) = 3<%s>`, r.path("alias.c"), r.path("source/a.c"))
		r.add(r.pid, `openat(AT_FDCWD, %q, O_RDONLY|O_DIRECTORY|O_CLOEXEC) = 3<%s>`, r.path("lib"), r.path("source"))
		r.list(r.pid, "source")
		r.add(r.pid, `newfstatat(AT_FDCWD, %q, {st_mode=S_IFREG|0644, st_size=7, ...}, 0) = 0`, r.path("lib/b.c"))
	})
	if settlement := r.settle(t); len(settlement.Settled) != 1 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	got := r.entries(t, build.NameKey)
	for _, want := range []string{"link alias.c", "link lib", "content source/a.c", "exists source/b.c", "listing source"} {
		if !slices.Contains(got, want) {
			t.Fatalf("the read set %q lacks %q", got, want)
		}
	}
	if value, _ := entryValue(r.root, r.cache, readEntry{Kind: "link", Path: "lib"}, 0, nil); value != "link source" {
		t.Fatalf("a symbolic link to a directory values %q", value)
	}
	write(t, r.root, "source/a.c", "int a2;\n")
	if _, err := reading(t, port); !errors.Is(err, ErrNotBuilt) || !strings.Contains(err.Error(), "content source/a.c") {
		t.Fatalf("a changed link target read as %v", err)
	}
	gitIn(t, r.root, "checkout", "-q", "--", "source/a.c")
	os.Remove(filepath.Join(r.root, "alias.c"))
	os.Symlink("source/b.c", filepath.Join(r.root, "alias.c"))
	if _, err := reading(t, port); !errors.Is(err, ErrNotBuilt) || !strings.Contains(err.Error(), "link alias.c") {
		t.Fatalf("a repointed link read as %v", err)
	}

	// A link out of the tree lands where no key can name.
	r.retrace(t, "outward")
	outside := filepath.Join(t.TempDir(), "elsewhere.h")
	write(t, filepath.Dir(outside), "elsewhere.h", "#define Y 1\n")
	os.Symlink(outside, filepath.Join(r.root, "outward.h"))
	gitIn(t, r.root, "add", "outward.h")
	gitIn(t, r.root, "commit", "-q", "-m", "outward")
	forgetTracked()
	other := Inputs{Name: "outward"}
	Product(t, other, building("built"))
	build = r.built(t)["outward"]
	resolvedOutside, _ := filepath.EvalSymlinks(outside)
	r.window(build.Build, func() {
		r.add(r.pid, `openat(AT_FDCWD, %q, O_RDONLY) = 3<%s>`, r.path("outward.h"), resolvedOutside)
	})
	if settlement := r.settle(t); len(settlement.Refused) != 1 || !strings.Contains(settlement.Refused[0], "temp not made by this run") {
		t.Fatalf("a link out of the tree: settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
}

// Each process has its own working directory: a child that changes its own (chdir, or fchdir to a descriptor) reads
// relative paths from there, and its parent's stays where it was.
// Not parallel: newRig.
func TestEachProcessReadsFromItsOwnWorkingDirectory(t *testing.T) {
	r := newRig(t)
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	child := r.pid + 1
	r.window(build.Build, func() {
		r.add(r.pid, `clone(child_stack=NULL, flags=CLONE_CHILD_CLEARTID|SIGCHLD) = %d`, child)
		r.add(child, `chdir(%q) = 0`, r.path("source"))
		r.add(child, `openat(AT_FDCWD, "a.c", O_RDONLY) = 3<%s>`, r.path("source/a.c"))
		r.add(child, `fchdir(4<%s>) = 0`, r.path("include"))
		r.add(child, `newfstatat(AT_FDCWD, "x.h", {st_mode=S_IFREG|0644, st_size=12, ...}, 0) = 0`)
		r.add(child, "+++ exited with 0 +++")
		r.add(r.pid, `openat(AT_FDCWD, "notes.txt", O_RDONLY) = 3<%s>`, r.path("notes.txt"))
	})
	if settlement := r.settle(t); len(settlement.Settled) != 1 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	if got, want := r.entries(t, build.NameKey), []string{"settings binary", "content go.mod", "exists include/x.h", "content notes.txt", "content source/a.c"}; !slices.Equal(got, want) {
		t.Fatalf("the read set is %q, want %q", got, want)
	}
}

// A build whose reads depend on what it read (a selector naming which file to read) is keyed by the selector: a
// change to it misses, the rebuild records its own reads first, and the earlier set still finds the earlier product
// when the tree goes back.
// Not parallel: newRig.
func TestAReadThatDependsOnAnotherRebuildsTracedAndKeepsBothSets(t *testing.T) {
	r := newRig(t)
	write(t, r.root, "selector.txt", "a\n")
	gitIn(t, r.root, "add", "selector.txt")
	gitIn(t, r.root, "commit", "-q", "-m", "selector")
	forgetTracked()
	reads := func(chosen string) {
		Product(t, port, building("built from "+chosen))
		build := r.built(t)["port"]
		r.window(build.Build, func() {
			r.open(r.pid, "selector.txt")
			r.open(r.pid, "source/"+chosen)
		})
		if settlement := r.settle(t); len(settlement.Settled) != 1 {
			t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
		}
	}
	reads("a.c")
	first, err := reading(t, port)
	if err != nil {
		t.Fatal(err)
	}
	write(t, r.root, "selector.txt", "b\n")
	if _, err := reading(t, port); !errors.Is(err, ErrNotBuilt) || !strings.Contains(err.Error(), "content selector.txt") {
		t.Fatalf("a changed selector read as %v", err)
	}
	r.retrace(t, "b")
	reads("b.c")
	second, err := reading(t, port)
	if err != nil || second == first {
		t.Fatalf("after the traced rebuild: %q, %v (the first was %q)", second, err, first)
	}
	if content, _ := os.ReadFile(filepath.Join(second, "product")); string(content) != "built from b.c" {
		t.Fatalf("the rebuilt product reads %q", content)
	}
	write(t, r.root, "selector.txt", "a\n")
	if found, err := reading(t, port); err != nil || found != first {
		t.Fatalf("back on the first tree: %q, %v; want %q", found, err, first)
	}
}

// What the build's own processes wrote before reading it is output, not input: it never joins the read set, in the
// home directory or in temp. A temp file another process wrote, outside the build's, is refused.
// Not parallel: newRig.
func TestWhatTheBuildMadeIsNeverRead(t *testing.T) {
	r := newRig(t)
	home, _ := os.UserHomeDir()
	made := filepath.Join(home, ".adamic-made-by-the-build", "out.o")
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	tool, sibling := r.pid+1, r.pid+2
	r.add(r.pid, `mkdirat(AT_FDCWD, "/tmp/adamic-test-dir", 0700) = 0`)
	r.add(r.pid, `clone(child_stack=NULL, flags=SIGCHLD) = %d`, sibling)
	r.window(build.Build, func() {
		r.add(r.pid, `clone(child_stack=NULL, flags=SIGCHLD) = %d`, tool)
		r.write(tool, made)
		r.open(tool, made)
		r.open(tool, "/tmp/adamic-test-dir/input")
		r.add(tool, "+++ exited with 0 +++")
		r.open(r.pid, made)
		r.open(r.pid, "source/a.c")
	})
	if settlement := r.settle(t); len(settlement.Settled) != 1 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	if got, want := r.entries(t, build.NameKey), []string{"settings binary", "content go.mod", "content source/a.c"}; !slices.Equal(got, want) {
		t.Fatalf("the read set is %q, want %q", got, want)
	}
	r.retrace(t, "sibling")
	r.lines = nil
	other := Inputs{Name: "sibling's"}
	Product(t, other, building("built"))
	build = r.built(t)["sibling's"]
	// go test (the root) starts two test binaries; the one building reads what the other wrote.
	test := r.pid + 3
	r.add(r.pid, `clone(child_stack=NULL, flags=SIGCHLD) = %d`, sibling)
	r.write(sibling, "/tmp/adamic-test-sibling/out")
	r.add(sibling, "+++ exited with 0 +++")
	r.add(r.pid, `clone(child_stack=NULL, flags=SIGCHLD) = %d`, test)
	r.marker(test, "build-begin", build.Build)
	r.open(test, "/tmp/adamic-test-sibling/out")
	r.marker(test, "build-end", build.Build)
	if settlement := r.settle(t); len(settlement.Refused) != 1 || !strings.Contains(settlement.Refused[0], "temp not made by this run") {
		t.Fatalf("a temp file a sibling made: settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
}

// A file the build read that git doesn't track is refused even when it is gone by the time the build settles (an
// installed node_modules, a generated file a test removed): it was there when it was read, so it is never absent.
// Not parallel: newRig.
func TestAnUntrackedReadIsRefusedEvenWhenItIsGone(t *testing.T) {
	r := newRig(t)
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	r.window(build.Build, func() {
		r.open(r.pid, "source/a.c")
		r.open(r.pid, "node_modules/left-pad/index.js")
		r.missing(r.pid, "node_modules/right-pad/index.js")
	})
	settlement := r.settle(t)
	if len(settlement.Refused) != 1 || !strings.Contains(settlement.Refused[0], "untracked: node_modules/left-pad/index.js") || strings.Contains(settlement.Refused[0], "right-pad") {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
}

// Two settlements of one product on one tree never lose each other's reads: each set recorded while the other still
// values as recorded is their union, so a change to a file either build read misses, under any interleaving.
// Not parallel: newRig.
func TestSettlementsOfOneProductKeepEveryRead(t *testing.T) {
	r := newRig(t)
	nameKey := NameKey(r.root, port)
	names := []string{"source/a.c", "source/b.c", "include/x.h", "notes.txt", "go.sum", "go.mod"}
	done := make(chan error, len(names))
	for _, name := range names {
		go func() {
			_, _, _, err := recordUnion(r.root, r.cache, nameKey, "port", readSet{Entries: []readEntry{{Kind: "content", Path: name}}}, func(key string) string {
				os.MkdirAll(filepath.Join(r.cache, key), 0o755)
				return filepath.Join(r.cache, key)
			})
			done <- err
		}()
	}
	for range names {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
	file, _ := loadReads(r.cache, nameKey)
	if len(file.Sets) != 1 || len(file.Sets[0].Entries) != len(names) {
		t.Fatalf("after %d settlements on one tree: %+v", len(names), file.Sets)
	}
	for _, name := range names[:4] {
		before, _ := os.ReadFile(filepath.Join(r.root, name))
		write(t, r.root, name, "changed\n")
		if _, err := reading(t, port); !errors.Is(err, ErrNotBuilt) || !strings.Contains(err.Error(), "content "+name) {
			t.Fatalf("a change to %s read as %v", name, err)
		}
		write(t, r.root, name, string(before))
	}
}

// A checkout values a clean file by the id git's index holds, without reading it, and an unpacked source with no git
// hashes the same file to the same id; a file changed after git was asked is hashed, never valued by the stale id.
// Not parallel: newRig.
func TestAFileIsValuedByItsGitObject(t *testing.T) {
	r := newRig(t)
	set := readSet{Entries: []readEntry{{Kind: "content", Path: "source/a.c"}, {Kind: "content", Path: "go.mod"}, {Kind: "listing", Path: "source"}}}
	nameKey := NameKey(r.root, port)
	// Old enough to be clean by git's racy margin and ours.
	for _, name := range []string{"source/a.c", "go.mod"} {
		os.Chtimes(filepath.Join(r.root, name), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour))
	}
	gitIn(t, r.root, "update-index", "--refresh", "-q")
	forgetTracked()
	inCheckout, err := evaluate(r.root, r.cache, nameKey, set, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	object, err := toolchain.GitCommand("-C", r.root, "rev-parse", "HEAD:source/a.c").Output()
	if err != nil || inCheckout.values[0] != "file false "+strings.TrimSpace(string(object)) {
		t.Fatalf("source/a.c values %q, git names it %q (%v)", inCheckout.values[0], object, err)
	}
	unpacked := t.TempDir()
	for _, name := range []string{"source/a.c", "source/b.c", "go.mod"} {
		content, _ := os.ReadFile(filepath.Join(r.root, name))
		write(t, unpacked, name, string(content))
	}
	if inSource, err := evaluate(unpacked, r.cache, nameKey, set, 0, nil); err != nil || inSource.key != inCheckout.key {
		t.Fatalf("a checkout keyed %s, its unpacked source %s (%v)", inCheckout.key, inSource.key, err)
	}
	// Changed after git was asked, with git's answer still held: hashed, so the key moves. Valued by a new process: the
	// rig's trace makes this one hold what it valued for its life, and a traced run whose tree changed is refused whole.
	write(t, r.root, "source/a.c", "int a2;\n")
	forgetHeldForLife()
	if after, _ := evaluate(r.root, r.cache, nameKey, set, 0, nil); after.key == inCheckout.key {
		t.Fatal("a file changed after git was asked kept its old id")
	}
}

// A product another test binary built (a sibling under go test, ended before this build) is an input, by its key:
// what a sibling wrote is never this build's output, and a product is a product whoever wrote it (review 1).
// Not parallel: newRig.
func TestASiblingsProductIsAnInput(t *testing.T) {
	r := newRig(t)
	header := Inputs{Name: "header", Files: []string{"include"}}
	headerProduct := Product(t, header, building("header v1"))
	Product(t, port, building("port built from header v1"))
	built := r.built(t)
	goTest, first, second := r.pid, r.pid+1, r.pid+2
	r.add(goTest, `clone(child_stack=NULL, flags=SIGCHLD) = %d`, first)
	r.marker(first, "build-begin", built["header"].Build)
	r.open(first, "include/x.h")
	r.write(first, filepath.Join(filepath.Dir(headerProduct), ".building-x", "product"))
	r.add(first, `rename(%q, %q) = 0`, filepath.Join(filepath.Dir(headerProduct), ".building-x"), headerProduct)
	r.marker(first, "build-end", built["header"].Build)
	r.add(first, "+++ exited with 0 +++")
	r.add(goTest, `clone(child_stack=NULL, flags=SIGCHLD) = %d`, second)
	r.marker(second, "build-begin", built["port"].Build)
	r.open(second, filepath.Join(headerProduct, "product"))
	r.open(second, "source/a.c")
	r.marker(second, "build-end", built["port"].Build)
	r.add(second, "+++ exited with 0 +++")
	if settlement := r.settle(t); len(settlement.Settled) != 2 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	if got := r.entries(t, built["port"].NameKey); !slices.Contains(got, "product "+built["header"].NameKey) {
		t.Fatalf("port's read set %q doesn't name header's product", got)
	}
	if _, err := reading(t, port); err != nil {
		t.Fatal(err)
	}
	write(t, r.root, "include/x.h", "#define X 2\n")
	if _, err := reading(t, port); !errors.Is(err, ErrNotBuilt) {
		t.Fatalf("header's input changed and port read as %v", err)
	}
}

// Two products built at once in one test process each hold the other's reads in their windows (the superset Kirk
// accepted), but a read of a product while its build hadn't ended is that build writing its own output, never an
// input: both settle, and neither names the other (Home's dry run, Oct 11: bridge-test-api and tsgo-asan.a were each
// refused for reading the other, a cycle no key can value). A product read after its build ended stays an input,
// which TestAProductReadingAnotherIsKeyedByItsKey holds. Mutants: the rule dropped (the cycle refuses both); every
// product this trace built dropped (TestAProductReadingAnotherIsKeyedByItsKey reds).
// Not parallel: newRig.
func TestSiblingsBuildingAtOnceAreNotEachOthersInputs(t *testing.T) {
	r := newRig(t)
	header := Inputs{Name: "header", Files: []string{"include"}}
	headerProduct := Product(t, header, building("header"))
	portProduct := Product(t, port, building("port"))
	built := r.built(t)
	r.marker(r.pid, "build-begin", built["header"].Build)
	r.marker(r.pid, "build-begin", built["port"].Build)
	r.open(r.pid, "include/x.h")
	r.write(r.pid, filepath.Join(headerProduct, "product"))
	r.open(r.pid, filepath.Join(headerProduct, "product"))
	r.open(r.pid, "source/a.c")
	r.write(r.pid, filepath.Join(portProduct, "product"))
	r.open(r.pid, filepath.Join(portProduct, "product"))
	r.marker(r.pid, "build-end", built["header"].Build)
	r.marker(r.pid, "build-end", built["port"].Build)
	if settlement := r.settle(t); len(settlement.Settled) != 2 || len(settlement.Refused) != 0 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	for _, pair := range [][2]string{{"header", "port"}, {"port", "header"}} {
		if got := r.entries(t, built[pair[0]].NameKey); slices.Contains(got, "product "+built[pair[1]].NameKey) {
			t.Fatalf("%s's read set names %s, a sibling still being built when it was read: %q", pair[0], pair[1], got)
		}
	}
}

// A product many others read is valued once per pass: 12 layers of 3 products, each reading the 3 below, is 3^12
// evaluations without a memo (Home's dry run, Oct 11: nine traces settled for over two hours, ten deep) and 36 with
// one. Two products reading each other fail at once, named as a cycle. Mutants: productFor without its memo (the graph
// runs past its deadline); the cycle never marked (it runs to depth 16 down every branch).
// Not parallel: newRig.
func TestAProductIsValuedOncePerPass(t *testing.T) {
	r := newRig(t)
	nameKeyOf := func(layer, index int) string {
		return fmt.Sprintf("%064x", layer*16+index+1)
	}
	for layer := range 12 {
		for index := range 3 {
			set := readSet{Entries: []readEntry{{Kind: "content", Path: "go.mod"}}}
			if layer > 0 {
				for below := range 3 {
					set.Entries = append(set.Entries, readEntry{Kind: "product", Path: nameKeyOf(layer-1, below)})
				}
			}
			set.Entries = sortedEntries(set.Entries)
			evaluated, err := evaluate(r.root, r.cache, nameKeyOf(layer, index), set, 0, nil)
			if err != nil {
				t.Fatal(err)
			}
			set.Key = evaluated.key
			if err := recordReads(r.cache, nameKeyOf(layer, index), fmt.Sprintf("layer %d product %d", layer, index), set); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(r.cache, evaluated.key), 0o755); err != nil {
				t.Fatal(err)
			}
		}
	}
	started := time.Now()
	top, err := productFor(r.root, r.cache, nameKeyOf(11, 0), 0, nil)
	if err != nil || top.directory == "" {
		t.Fatalf("the top of the graph: %+v, %v", top, err)
	}
	if took := time.Since(started); took > 5*time.Second {
		t.Fatalf("valuing 36 products took %v", took)
	}
	cycle := func(from, to string) {
		t.Helper()
		set := readSet{Entries: sortedEntries([]readEntry{{Kind: "content", Path: "go.mod"}, {Kind: "product", Path: to}})}
		if err := recordReads(r.cache, from, "cycle "+from[:4], set); err != nil {
			t.Fatal(err)
		}
	}
	first, second := strings.Repeat("e", 64), strings.Repeat("f", 64)
	cycle(first, second)
	cycle(second, first)
	looked, err := productFor(r.root, r.cache, first, 0, nil)
	if err != nil || looked.directory != "" || !strings.Contains(looked.why, "cycle") {
		t.Fatalf("two products reading each other: %+v, %v", looked, err)
	}
}

// A write into the tree by any process of the run (go generate before go test, a test rewriting a golden file)
// refuses every build: what they read isn't the tree that would be keyed (review 1).
// Not parallel: newRig.
func TestATreeWriteAnywhereInTheRunRefusesIt(t *testing.T) {
	r := newRig(t)
	Product(t, port, building("built from generated"))
	build := r.built(t)["port"]
	shell, generate, test := r.pid, r.pid+1, r.pid+2
	r.add(shell, `clone(child_stack=NULL, flags=SIGCHLD) = %d`, generate)
	r.open(generate, "include/x.h")
	r.write(generate, "source/a.c")
	r.add(generate, "+++ exited with 0 +++")
	r.add(shell, `clone(child_stack=NULL, flags=SIGCHLD) = %d`, test)
	r.marker(test, "build-begin", build.Build)
	r.open(test, "source/a.c")
	r.marker(test, "build-end", build.Build)
	if settlement := r.settle(t); len(settlement.Refused) != 1 || !strings.Contains(settlement.Refused[0], "tree write: source/a.c") {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
}

// A read set is valued when the run settles, so a tree that changed during the run (a commit, an edit) refuses every
// build rather than key what was built under what the tree is now (review 2).
// Not parallel: newRig.
func TestATreeThatChangedDuringTheRunRefusesIt(t *testing.T) {
	for name, run := range map[string]struct{ before, during func(r *rig) }{
		"a commit": {during: func(r *rig) {
			write(t, r.root, "source/a.c", "int a_changed_mid_run;\n")
			gitIn(t, r.root, "commit", "-qam", "next")
		}},
		"an edit":           {during: func(r *rig) { write(t, r.root, "notes.txt", "edited mid-run\n") }},
		"an untracked file": {during: func(r *rig) { write(t, r.root, "left.txt", "a leftover\n") }},
		// Status says a file is modified, never what it holds: one dirty before the run and edited again during it.
		"a dirty file edited again": {
			before: func(r *rig) { write(t, r.root, "source/a.c", "int dirty_before;\n") },
			during: func(r *rig) { write(t, r.root, "source/a.c", "int edited_mid_run;\n") },
		},
		"an untracked file edited again": {
			before: func(r *rig) { write(t, r.root, "left.txt", "a leftover\n") },
			during: func(r *rig) { write(t, r.root, "left.txt", "another leftover\n") },
		},
		"a submodule's dirty file edited again": {
			before: func(r *rig) {
				module := t.TempDir()
				write(t, module, "inner.c", "int inner;\n")
				gitIn(t, module, "init", "-q")
				gitIn(t, module, "add", ".")
				gitIn(t, module, "commit", "-q", "-m", "module")
				gitIn(t, r.root, "-c", "protocol.file.allow=always", "submodule", "add", "-q", module, "module")
				gitIn(t, r.root, "commit", "-q", "-m", "module")
				write(t, r.root, "module/inner.c", "int dirty_before;\n")
			},
			during: func(r *rig) { write(t, r.root, "module/inner.c", "int edited_mid_run;\n") },
		},
		"a package upgraded": {
			before: func(r *rig) {
				dpkg := t.TempDir()
				write(t, dpkg, "status", "Package: libc6-dev\nArchitecture: amd64\nVersion: 2.39-0ubuntu8\n")
				original := dpkgDirectory
				dpkgDirectory = dpkg
				t.Cleanup(func() { dpkgDirectory = original })
			},
			during: func(r *rig) {
				write(t, dpkgDirectory, "status", "Package: libc6-dev\nArchitecture: amd64\nVersion: 2.39-0ubuntu9\n")
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			r := newRig(t)
			if run.before != nil {
				run.before(r)
				forgetTracked()
			}
			before, err := TreeState()
			if err != nil {
				t.Fatal(err)
			}
			r.before = before
			Product(t, port, building("built"))
			r.window(r.built(t)["port"].Build, func() { r.open(r.pid, "source/a.c") })
			run.during(r)
			forgetTracked()
			if settlement := r.settle(t); len(settlement.Refused) != 1 || !strings.Contains(settlement.Refused[0], "the tree changed during the run") {
				t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
			}
		})
	}
}

// What a build reads of the machine is keyed (review 3): a file a package installed by the package's version, a link,
// a directory and a file no package owns by themselves; an upgrade or a changed file misses.
// Not parallel: newRig, and it points the machine's places and dpkg's database at the test.
func TestWhatABuildReadsOfTheMachineIsKeyed(t *testing.T) {
	r := newRig(t)
	machine, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	system := filepath.Join(machine, "usr")
	dpkg := filepath.Join(machine, "dpkg")
	ld, stdio, crt := filepath.Join(system, "bin/ld"), filepath.Join(system, "include/stdio.h"), filepath.Join(system, "lib/gcc/crtbegin.o")
	installed := map[string]string{ld: "the linker", stdio: "int printf(const char *, ...);"}
	for path, content := range installed {
		write(t, filepath.Dir(path), filepath.Base(path), content)
	}
	write(t, system, "lib/gcc/crtbegin.o", "an object no package owns")
	os.Symlink(ld, filepath.Join(system, "cc"))
	sums := func(path string) string {
		return fmt.Sprintf("%x  %s\n", md5.Sum([]byte(installed[path])), strings.TrimPrefix(path, "/"))
	}
	write(t, dpkg, "info/binutils:amd64.list", system+"/bin\n"+ld+"\n")
	write(t, dpkg, "info/binutils:amd64.md5sums", sums(ld))
	write(t, dpkg, "info/libc6-dev:amd64.list", stdio+"\n")
	write(t, dpkg, "info/libc6-dev:amd64.md5sums", sums(stdio))
	status := func(libc string) {
		write(t, dpkg, "status", "Package: binutils\nArchitecture: amd64\nVersion: 2.42-4\n\nPackage: libc6-dev\nArchitecture: amd64\nVersion: "+libc+"\n")
		packageIndexes.Delete(dpkg)
	}
	status("2.39-0ubuntu8")
	originalPlaces, originalDpkg := systemPlaces, dpkgDirectory
	systemPlaces, dpkgDirectory = []string{system}, dpkg
	t.Cleanup(func() { systemPlaces, dpkgDirectory = originalPlaces, originalDpkg; packageIndexes.Delete(dpkg) })
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	child := r.pid + 1
	r.window(build.Build, func() {
		r.open(r.pid, "source/a.c")
		r.add(r.pid, `clone(child_stack=NULL, flags=SIGCHLD) = %d`, child)
		r.add(child, `execve(%q, ["ld"], 0x7ffd /* 30 vars */) = 0`, ld)
		r.open(child, crt)
		r.list(child, filepath.Join(system, "lib/gcc"))
		r.open(child, stdio)
		r.add(child, `newfstatat(AT_FDCWD, %q, {st_mode=S_IFLNK|0777, st_size=7, ...}, AT_SYMLINK_NOFOLLOW) = 0`, filepath.Join(system, "cc"))
		r.add(child, "+++ exited with 0 +++")
	})
	if settlement := r.settle(t); len(settlement.Settled) != 1 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	got := r.entries(t, build.NameKey)
	for _, want := range []string{"system " + ld, "system " + stdio, "system " + crt, "system " + filepath.Join(system, "lib/gcc"), "system " + filepath.Join(system, "cc")} {
		if !slices.Contains(got, want) {
			t.Fatalf("the read set %q lacks %q", got, want)
		}
	}
	if value := systemValue(stdio); value != "package libc6-dev:amd64 version 2.39-0ubuntu8" {
		t.Fatalf("a file dpkg installed values %q", value)
	}
	if _, err := reading(t, port); err != nil {
		t.Fatal(err)
	}
	status("2.39-0ubuntu9")
	if _, err := reading(t, port); !errors.Is(err, ErrNotBuilt) || !strings.Contains(err.Error(), "system "+stdio) {
		t.Fatalf("libc6-dev upgraded and port read as %v", err)
	}
	status("2.39-0ubuntu8")
	// Edited in place, the package's version unchanged: dpkg's md5 no longer matches, so it is keyed by its content.
	write(t, system, "include/stdio.h", "int printf(const char *, ...); /* patched */")
	if _, err := reading(t, port); !errors.Is(err, ErrNotBuilt) || !strings.Contains(err.Error(), "system "+stdio) {
		t.Fatalf("an installed file edited in place, port read as %v", err)
	}
	write(t, system, "include/stdio.h", installed[stdio])
	if _, err := reading(t, port); err != nil {
		t.Fatalf("the file restored: %v", err)
	}
	write(t, system, "lib/gcc/crtbegin.o", "another object")
	if _, err := reading(t, port); !errors.Is(err, ErrNotBuilt) || !strings.Contains(err.Error(), "system "+crt) {
		t.Fatalf("an unowned file changed and port read as %v", err)
	}
}

// A file git stops comparing with its index (assume-unchanged, skip-worktree) is hashed, never valued by its index
// id, which may be stale (review 4).
// Not parallel: newRig.
func TestAFileGitStopsComparingIsHashed(t *testing.T) {
	for _, flag := range []string{"--assume-unchanged", "--skip-worktree"} {
		t.Run(flag, func(t *testing.T) {
			r := newRig(t)
			set := readSet{Entries: []readEntry{{Kind: "content", Path: "source/a.c"}}}
			nameKey := NameKey(r.root, port)
			old := time.Now().Add(-time.Hour)
			os.Chtimes(filepath.Join(r.root, "source/a.c"), old, old)
			gitIn(t, r.root, "update-index", "--refresh", "-q")
			forgetTracked()
			before, _ := evaluate(r.root, r.cache, nameKey, set, 0, nil)
			gitIn(t, r.root, "update-index", flag, "source/a.c")
			write(t, r.root, "source/a.c", "int Z;\n")
			os.Chtimes(filepath.Join(r.root, "source/a.c"), old, old)
			forgetTracked()
			// Valued by a new process, as after the edit above: the rig's trace makes this one hold what it valued.
			forgetHeldForLife()
			if after, _ := evaluate(r.root, r.cache, nameKey, set, 0, nil); after.key == before.key {
				t.Fatalf("source/a.c edited under %s kept its key", flag)
			}
		})
	}
}

// The binary's build settings (tags, ldflags, -race, GOEXPERIMENT) are in the recipe: a binary built otherwise finds
// none of the products this one built (review 5).
// Not parallel: newRig, and it sets the settings the key is valued with.
func TestTheBinarysBuildSettingsKeyItsProducts(t *testing.T) {
	r := newRig(t)
	Product(t, port, building("built"))
	r.window(r.built(t)["port"].Build, func() { r.open(r.pid, "source/a.c") })
	if settlement := r.settle(t); len(settlement.Settled) != 1 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	if _, err := reading(t, port); err != nil {
		t.Fatal(err)
	}
	settingsOverride = binarySettings() + "\n-race=true"
	t.Cleanup(func() { settingsOverride = "" })
	if _, err := reading(t, port); !errors.Is(err, ErrNotBuilt) || !strings.Contains(err.Error(), "settings binary") {
		t.Fatalf("a binary built with -race read as %v", err)
	}
}

// A read set file of another version (v2 had a package kind, gone since) is no read set: never valued, never trusted.
// Not parallel: newRig.
func TestAReadSetFileOfAnotherVersionIsIgnored(t *testing.T) {
	r := newRig(t)
	nameKey := NameKey(r.root, port)
	write(t, r.cache, nameKey+".reads", `{"version": "buildcache reads v2", "name": "port", "sets": [{"key": "`+strings.Repeat("a", 64)+`", "entries": [{"kind": "package", "path": "libc6-dev:amd64"}]}]}`)
	if file, err := loadReads(r.cache, nameKey); err != nil || len(file.Sets) != 0 {
		t.Fatalf("a v2 read set file loaded as %+v, %v", file, err)
	}
}

// A machine file's hash is kept across processes in the build cache, by path, inode, modification time and size: a
// later process reads it back without hashing, and a file changed in place is hashed again.
// Not parallel: newRig.
func TestAMachineFilesHashIsKeptAcrossProcesses(t *testing.T) {
	r := newRig(t)
	path := filepath.Join(t.TempDir(), "crtbegin.o")
	write(t, filepath.Dir(path), "crtbegin.o", "an object")
	first := systemValue(path)
	entries, _ := os.ReadDir(filepath.Join(r.cache, "machine"))
	if len(entries) != 1 {
		t.Fatalf("the machine cache holds %d entries", len(entries))
	}
	// Another process: nothing in memory; what the cache holds is what it reads.
	kept := filepath.Join(r.cache, "machine", entries[0].Name())
	os.WriteFile(kept, []byte("kept-by-an-earlier-process"), 0o644)
	systemContents.Range(func(key, _ any) bool { systemContents.Delete(key); return true })
	if got := systemValue(path); got != "file false kept-by-an-earlier-process" {
		t.Fatalf("another process valued it %q, after %q", got, first)
	}
	write(t, filepath.Dir(path), "crtbegin.o", "another object, longer")
	if got := systemValue(path); got == first || strings.Contains(got, "kept-by") {
		t.Fatalf("a file changed in place valued %q", got)
	}
}

// The report Loom's tree builder reads (ADAMIC_TRACE_REPORT, one JSON line each): a settled product by its name, name
// key, key and how many reads its set names; a refused one by its reason and no key; and a run that can't be keyed by
// a lost line first, then each of its products refused. Mutants: a settled report without its key, a refusal never
// reported, the lost line dropped, and a field renamed off the wire (nameKey).
// Not parallel: newRig.
func TestTheReportNamesEachProductsOutcome(t *testing.T) {
	lines := func(t *testing.T, reports []Report) []map[string]any {
		t.Helper()
		file := filepath.Join(t.TempDir(), "report.jsonl")
		if err := WriteReport(file, reports); err != nil {
			t.Fatal(err)
		}
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var found []map[string]any
		for _, line := range strings.Split(strings.TrimSuffix(string(content), "\n"), "\n") {
			var fields map[string]any
			if err := json.Unmarshal([]byte(line), &fields); err != nil {
				t.Fatalf("a report line isn't JSON: %q: %v", line, err)
			}
			found = append(found, fields)
		}
		return found
	}
	r := newRig(t)
	header := Inputs{Name: "header", Files: []string{"include"}}
	Product(t, header, building("header"))
	Product(t, port, building("port"))
	built := r.built(t)
	r.window(built["header"].Build, func() { r.open(r.pid, "include/x.h") })
	write(t, r.root, "scratch.txt", "x")
	r.window(built["port"].Build, func() {
		r.open(r.pid, "source/a.c")
		r.open(r.pid, "scratch.txt")
	})
	settlement := r.settle(t)
	got := lines(t, settlement.Reports)
	if len(got) != 2 {
		t.Fatalf("the report is %v; want header settled and port refused", got)
	}
	file, err := loadReads(r.cache, built["header"].NameKey)
	if err != nil || len(file.Sets) != 1 {
		t.Fatalf("header's read sets: %+v, %v", file, err)
	}
	settled := map[string]any{"name": "header", "nameKey": built["header"].NameKey, "key": file.Sets[0].Key, "outcome": "settled",
		"reads": float64(len(file.Sets[0].Entries)), "window": BuildSuperset}
	if fmt.Sprint(got[0]) != fmt.Sprint(settled) {
		t.Fatalf("the settled line is %v; want %v", got[0], settled)
	}
	reason, _ := got[1]["reason"].(string)
	if got[1]["name"] != "port" || got[1]["nameKey"] != built["port"].NameKey || got[1]["outcome"] != "refused" ||
		!strings.Contains(reason, "untracked: scratch.txt") || got[1]["key"] != nil {
		t.Fatalf("the refused line is %v", got[1])
	}
	r = newRig(t)
	Product(t, port, building("built in a tree it wrote"))
	build := r.built(t)["port"]
	r.write(r.pid, "include/x.h")
	r.window(build.Build, func() { r.open(r.pid, "source/a.c") })
	got = lines(t, r.settle(t).Reports)
	if len(got) != 2 {
		t.Fatalf("the report of a run that can't be keyed is %v; want the lost line, then port refused", got)
	}
	reason, _ = got[0]["reason"].(string)
	if len(got[0]) != 2 || got[0]["outcome"] != "lost" || !strings.Contains(reason, "tree write: include/x.h") {
		t.Fatalf("the lost line is %v", got[0])
	}
	if got[1]["name"] != "port" || got[1]["outcome"] != "refused" || got[1]["key"] != nil {
		t.Fatalf("a product of a lost run reads %v", got[1])
	}
}

// strace writes a path's bytes outside printable ASCII as octal escapes, in a quoted argument and in a path it decodes
// from a descriptor alike (strace 6.8 on Home: openat(AT_FDCWD, "w\303\251.txt", O_RDONLY) = 3</tmp/w\303\251.txt>),
// and git ls-files -z names the same file by its bytes. Every spelling is read back to its bytes, so a tracked file with
// such a name is keyed, never refused as untracked (Home's dry run, Oct 11: internal/oracle/testdata/walking/é.txt and
// reading/\357\277\275.txt refused 29 products). Mutants: the opened path, the listed directory, or an at-call's directory
// left escaped, or the directory fchdir moves to.
// Not parallel: newRig.
func TestAPathStraceEscapesIsReadAsItsBytes(t *testing.T) {
	r := newRig(t)
	write(t, r.root, "source/é.c", "int e;\n")
	write(t, r.root, "dé/x.txt", "x\n")
	write(t, r.root, "dé/\uFFFD.txt", "replacement\n")
	gitIn(t, r.root, "add", ".")
	gitIn(t, r.root, "commit", "-q", "-m", "names outside ASCII")
	forgetTracked()
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	escaped := func(name string) string {
		var spelled strings.Builder
		for _, b := range []byte(r.path(name)) {
			if b < 0x20 || b >= 0x7f || b == '"' || b == '<' || b == '>' || b == '\\' {
				fmt.Fprintf(&spelled, `\%o`, b)
				continue
			}
			spelled.WriteByte(b)
		}
		return spelled.String()
	}
	r.window(build.Build, func() {
		r.add(r.pid, `openat(AT_FDCWD, "%s", O_RDONLY|O_CLOEXEC) = 3<%s>`, escaped("source/é.c"), escaped("source/é.c"))
		r.add(r.pid, `openat(AT_FDCWD, "%s", O_RDONLY|O_DIRECTORY|O_CLOEXEC) = 4<%s>`, escaped("dé"), escaped("dé"))
		r.add(r.pid, `getdents64(4<%s>, 0x55d0 /* 4 entries */, 32768) = 112`, escaped("dé"))
		r.add(r.pid, `openat(4<%s>, "x.txt", O_RDONLY|O_CLOEXEC) = 5<%s>`, escaped("dé"), escaped("dé/x.txt"))
		r.add(r.pid, `openat(4<%s>, "\357\277\275.txt", O_RDONLY|O_CLOEXEC) = 6<%s>`, escaped("dé"), escaped("dé/\uFFFD.txt"))
		r.add(r.pid, `fchdir(4<%s>) = 0`, escaped("dé"))
		r.add(r.pid, `newfstatat(AT_FDCWD, "y.txt", 0x7ffd5e1c, 0) = -1 ENOENT (No such file or directory)`)
	})
	if settlement := r.settle(t); len(settlement.Settled) != 1 || len(settlement.Refused) != 0 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	got := r.entries(t, build.NameKey)
	for _, want := range []string{"content source/é.c", "listing dé", "content dé/x.txt", "content dé/\uFFFD.txt", "exists dé/y.txt"} {
		if !slices.Contains(got, want) {
			t.Fatalf("the read set %q doesn't name %q", got, want)
		}
	}
}

// A machine file reached through a link is keyed as itself, never as what its descriptor decodes to: WSL makes
// /etc/resolv.conf a link to /mnt/wsl/resolv.conf, and strace decodes the open's descriptor to the target, a place no
// key can name (Workshop, Oct 12: every traced store lookup's DNS read refused its product, 866 of 866 in the smoke's
// tree). Mutant: the decoded target recorded for a machine file too, refused as other.
// Not parallel: newRig.
func TestAMachineFileThroughALinkIsKeyedAsItself(t *testing.T) {
	if _, err := os.Stat("/etc/resolv.conf"); err != nil {
		t.Skip("no /etc/resolv.conf here")
	}
	r := newRig(t)
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	r.window(build.Build, func() {
		r.add(r.pid, `openat(AT_FDCWD, "/etc/resolv.conf", O_RDONLY|O_CLOEXEC) = 3</mnt/wsl/resolv.conf>`)
	})
	if settlement := r.settle(t); len(settlement.Settled) != 1 || len(settlement.Refused) != 0 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	if got := r.entries(t, build.NameKey); !slices.Contains(got, "system /etc/resolv.conf") || slices.ContainsFunc(got, func(entry string) bool { return strings.Contains(entry, "/mnt/wsl") }) {
		t.Fatalf("the read set %q doesn't key /etc/resolv.conf as itself alone", got)
	}
}

// A build that opens one of the Windows GPU driver's libraries (WSL's mount) refuses its product: the driver is its
// box's Windows version, which no key names (#6eexfyv). Only a look at it is keyed, as a system path. Mutant: the
// library opened keyed as a system file.
// Not parallel: newRig.
func TestABuildThatOpensWSLsDriverRefuses(t *testing.T) {
	r := newRig(t)
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	r.window(build.Build, func() {
		r.add(r.pid, `openat(AT_FDCWD, "/usr/lib/wsl/lib/libcuda.so.1", O_RDONLY|O_CLOEXEC) = 3`)
	})
	if settlement := r.settle(t); len(settlement.Refused) != 1 || !strings.Contains(settlement.Refused[0], "the Windows GPU driver: /usr/lib/wsl/lib/libcuda.so.1") {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
}

// A build that reads the machine's time zone refuses its product: /etc/localtime is whatever zone the box was set to,
// which no key names (#cbw5jtf, Server against the house, Oct 12). A look at it is keyed as a system path. Mutant: the
// zone file read keyed as a system file.
// Not parallel: newRig.
func TestABuildThatReadsTheMachinesTimeZoneRefuses(t *testing.T) {
	r := newRig(t)
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	r.window(build.Build, func() {
		r.add(r.pid, `openat(AT_FDCWD, "/etc/localtime", O_RDONLY|O_CLOEXEC) = 3`)
	})
	if settlement := r.settle(t); len(settlement.Refused) != 1 || !strings.Contains(settlement.Refused[0], "the machine's time zone: /etc/localtime") {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
}

// What the loader looked for and missed in its own directories (a soname its cache doesn't name, tried in each) is keyed
// path by path, so a library later installed there moves the key, which lets the loader's cache leave those
// directories out (#6eexfyv). A machine directory a build only stat'd or opened to search is keyed by whether and as
// what it exists, never by its names, which are whatever the machine installed; one it listed (getdents) is keyed by
// its names (#vdha67a). Mutants: the loader's misses dropped; a stat'd directory keyed by its names; a listed one keyed
// by existence.
// Not parallel: newRig.
func TestTheLoadersMissesAndALookedIntoMachineDirectoryAreKeyedHonestly(t *testing.T) {
	r := newRig(t)
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	r.window(build.Build, func() {
		r.add(r.pid, `openat(AT_FDCWD, "/lib/x86_64-linux-gnu/libadamic-absent.so.1", O_RDONLY|O_CLOEXEC) = -1 ENOENT (No such file or directory)`)
		r.add(r.pid, `newfstatat(AT_FDCWD, "/usr/bin", {st_mode=S_IFDIR|0755, st_size=4096, ...}, 0) = 0`)
		r.add(r.pid, `openat(AT_FDCWD, "/usr/lib", O_RDONLY|O_NONBLOCK|O_CLOEXEC|O_DIRECTORY) = 3</usr/lib>`)
		r.add(r.pid, `openat(AT_FDCWD, "/usr/share", O_RDONLY|O_NONBLOCK|O_CLOEXEC|O_DIRECTORY) = 4</usr/share>`)
		r.add(r.pid, `getdents64(4</usr/share>, 0x55d0 /* 40 entries */, 32768) = 1280`)
	})
	if settlement := r.settle(t); len(settlement.Settled) != 1 || len(settlement.Refused) != 0 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	got := r.entries(t, build.NameKey)
	for _, want := range []string{"system /lib/x86_64-linux-gnu/libadamic-absent.so.1", "system directory /usr/bin", "system directory /usr/lib", "system /usr/share"} {
		if !slices.Contains(got, want) {
			t.Fatalf("the read set %q doesn't name %q", got, want)
		}
	}
	for _, unwanted := range []string{"system /usr/bin", "system /usr/lib", "system directory /usr/share"} {
		if slices.Contains(got, unwanted) {
			t.Fatalf("the read set %q names %q", got, unwanted)
		}
	}
}

// Tracked looks for each directory's .git as keying, as it lists the repository: a traced build that asks Tracked of a
// directory (tracked.PathsUnder's walk of the tree) never reads .git, which no key can name (home-5's dry run, Oct 11,
// where tracked.Git's own look refused cssnumbers and cssstrings). Mutant: the look outside keying.
// Not parallel: newRig, and statGit.
func TestTrackedLooksForGitAsKeying(t *testing.T) {
	r := newRig(t)
	write(t, r.root, "source/deeper/c.c", "int c;\n")
	gitIn(t, r.root, "add", ".")
	gitIn(t, r.root, "commit", "-q", "-m", "deeper")
	forgetTracked()
	var looks, unkeyed []string
	original := statGit
	statGit = func(name string) (os.FileInfo, error) {
		looks = append(looks, name)
		if keyingNow.Load() == 0 {
			unkeyed = append(unkeyed, name)
		}
		return original(name)
	}
	t.Cleanup(func() { statGit = original })
	if !Tracked(r.root, filepath.Join(r.root, "source", "deeper", "c.c"), false) {
		t.Fatal("a committed file isn't tracked")
	}
	if len(looks) < 3 || len(unkeyed) != 0 {
		t.Fatalf("looked for .git at %q; outside keying at %q", looks, unkeyed)
	}
}

// A family of products that read one another and the same directories is valued with each tree entry read once per
// Settle, not once per pass: 8 products, each reading the ones built before it and listing source, read the tree's
// few entries a handful of times, where a pass each read them 8 times over (home-5's dry run, Oct 11: css's PostCSS
// answers settled for over 20 minutes). The held values go when Settle ends. Mutant: Settle holds nothing.
// Not parallel: newRig.
func TestAFamilySettlesReadingTheTreeOnce(t *testing.T) {
	r := newRig(t)
	const members = 8
	var products []string
	for index := range members {
		products = append(products, Product(t, Inputs{Name: fmt.Sprintf("member%d", index), Files: []string{"source"}}, building(fmt.Sprint(index))))
	}
	built := r.built(t)
	for index := range members {
		child := r.pid + 1 + index
		r.add(r.pid, `clone(child_stack=NULL, flags=SIGCHLD) = %d`, child)
		r.marker(child, "build-begin", built[fmt.Sprintf("member%d", index)].Build)
		r.open(child, "source/a.c")
		r.list(child, "source")
		for before := range index {
			r.open(child, filepath.Join(products[before], "product"))
		}
		r.marker(child, "build-end", built[fmt.Sprintf("member%d", index)].Build)
		r.add(child, "+++ exited with 0 +++")
	}
	read := treeValuesRead.Load()
	if settlement := r.settle(t); len(settlement.Settled) != members {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	if read = treeValuesRead.Load() - read; read > 12 {
		t.Fatalf("settling %d products read %d tree values; the tree's few entries should each be read once", members, read)
	}
	if settledTree.Load() != nil {
		t.Fatal("the tree's values are still held after Settle")
	}
}

// The census log (ADAMIC_BUILD_LOG) is buildcache's own bookkeeping, as the trace's journal is: a build that reads or
// writes it is keyed without it, wherever it lives (Planner's shape B proof on Cloud, Oct 11: a build tree's log under
// the home directory refused two estree products). Mutant: the log classed by where it is (here, a temporary file the
// run didn't make).
// Not parallel: newRig.
func TestTheCensusLogIsNoInput(t *testing.T) {
	r := newRig(t)
	log := os.Getenv("ADAMIC_BUILD_LOG")
	write(t, filepath.Dir(log), filepath.Base(log), "build earlier 0123456789ab hit 0.01\n")
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	r.window(build.Build, func() {
		r.open(r.pid, "source/a.c")
		r.open(r.pid, log)
	})
	if settlement := r.settle(t); len(settlement.Settled) != 1 || len(settlement.Refused) != 0 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	for _, entry := range r.entries(t, build.NameKey) {
		if strings.Contains(entry, filepath.Base(log)) {
			t.Fatalf("the read set names the census log: %s", entry)
		}
	}
}

// declaredDriftByScan is declaredDrift as it was, meeting every entry for every read: what it says is the rule.
func declaredDriftByScan(declared, reads []string) *drift {
	if len(declared) == 0 {
		return nil
	}
	covered := func(name, entry string) bool {
		return name == entry || strings.HasPrefix(name, strings.TrimSuffix(entry, "/")+"/")
	}
	result := &drift{}
	read := map[string]bool{}
	for _, name := range reads {
		declaredHere := false
		for _, entry := range declared {
			if covered(name, entry) {
				declaredHere, read[entry] = true, true
			}
		}
		if !declaredHere && name != "." {
			result.Undeclared++
			result.UndeclaredPaths = append(result.UndeclaredPaths, name)
		}
	}
	for _, entry := range declared {
		if !read[entry] && !unkeyed(entry) {
			result.Unread++
			result.UnreadPaths = append(result.UnreadPaths, entry)
		}
	}
	sort.Strings(result.UndeclaredPaths)
	sort.Strings(result.UnreadPaths)
	result.UndeclaredPaths, result.UnreadPaths = result.UndeclaredPaths[:min(20, len(result.UndeclaredPaths))], result.UnreadPaths[:min(20, len(result.UnreadPaths))]
	return result
}

// The declared drift says what it said when it met every entry for every read, on the entries and reads that bend the
// rule (a directory with and without its slash, a name that only starts like an entry, the root, absolute paths), and
// does it at tsprinter's size in a moment: 76,958 reads against 75,696 entries settled for over 20 minutes on home-6's
// dry run (Oct 11). Mutants: an entry with a slash covering itself; an entry covering a name it only starts like.
func TestTheDeclaredDriftIsTheScansAtAnySize(t *testing.T) {
	t.Parallel()
	cases := []struct{ declared, reads []string }{
		{[]string{"a", "a/", "a/b", "ab", ".", "", "/", "/abs", "c/d/"}, []string{"a", "a/b", "a/b/c", "ab", "ab/x", "abc", ".", "c", "c/d", "c/d/e", "/abs", "/abs/x", "/other", "x/y"}},
		{[]string{"internal/x/"}, []string{"internal/x", "internal/x/y.go", "internal/xy/z.go"}},
		{[]string{"go.mod", "go.sum"}, []string{"go.mod"}},
	}
	var declared, reads []string
	for index := range 300 {
		declared = append(declared, fmt.Sprintf("d%d/f%d.go", index%17, index), fmt.Sprintf("d%d/", index%23))
		reads = append(reads, fmt.Sprintf("d%d/f%d.go", index%19, index), fmt.Sprintf("d%d/sub/f%d.go", index%29, index), fmt.Sprintf("e%d", index))
	}
	cases = append(cases, struct{ declared, reads []string }{declared, reads})
	for _, each := range cases {
		if got, want := declaredDrift(each.declared, each.reads), declaredDriftByScan(each.declared, each.reads); !reflect.DeepEqual(got, want) {
			t.Fatalf("declared %q, reads %q:\ngot  %+v\nwant %+v", each.declared, each.reads, got, want)
		}
	}
	declared, reads = nil, nil
	for index := range 76_000 {
		declared = append(declared, fmt.Sprintf("corpus/%d/%d.ts", index%400, index))
		reads = append(reads, fmt.Sprintf("corpus/%d/%d.ts", index%400, index+index%2))
	}
	started := time.Now()
	declaredDrift(declared, reads)
	if took := time.Since(started); took > 3*time.Second {
		t.Fatalf("76,000 reads against 76,000 entries took %v", took)
	}
}

// The drift's sample of undeclared and unread paths is the first 20 of all of them, sorted, however the reads arrive:
// they come from a map, so a sample cut before it was sorted named other paths each run (Loom, Oct 11). Mutant: the
// sample cut before sorting.
func TestTheDriftSamplesTheSameTwentyPathsInAnyOrder(t *testing.T) {
	t.Parallel()
	var declared, reads []string
	for index := range 50 {
		declared = append(declared, fmt.Sprintf("declared/%02d.go", index))
		reads = append(reads, fmt.Sprintf("undeclared/%02d.go", index))
	}
	var wantUndeclared, wantUnread []string
	for index := range 20 {
		wantUndeclared = append(wantUndeclared, fmt.Sprintf("undeclared/%02d.go", index))
		wantUnread = append(wantUnread, fmt.Sprintf("declared/%02d.go", index))
	}
	for permutation := range 8 {
		shuffled := func(paths []string) []string {
			paths = slices.Clone(paths)
			rand.New(rand.NewPCG(uint64(permutation), 1)).Shuffle(len(paths), func(i, j int) { paths[i], paths[j] = paths[j], paths[i] })
			return paths
		}
		got := declaredDrift(shuffled(declared), shuffled(reads))
		if got.Undeclared != 50 || got.Unread != 50 || !slices.Equal(got.UndeclaredPaths, wantUndeclared) || !slices.Equal(got.UnreadPaths, wantUnread) {
			t.Fatalf("permutation %d: %+v", permutation, got)
		}
	}
}

// The settlement holds the command's own exit as the trace records it, for cmd/traced to read when strace's own status
// is strace's, not the command's. A child's exit is never the command's. Mutant: any process's exit taken.
// Not parallel: newRig.
func TestTheSettlementHoldsTheCommandsExit(t *testing.T) {
	for _, each := range []struct {
		lines  []string
		exited bool
		status int
	}{
		{[]string{"command +++ exited with 0 +++"}, true, 0},
		{[]string{"command +++ exited with 2 +++"}, true, 2},
		{[]string{"child +++ exited with 7 +++"}, false, -1},
		{[]string{"command +++ killed by SIGKILL +++"}, false, -1},
		{nil, false, -1},
	} {
		r := newRig(t)
		child := r.pid + 1
		r.open(r.pid, "source/a.c")
		r.add(r.pid, `clone(child_stack=NULL, flags=SIGCHLD) = %d`, child)
		r.open(child, "source/b.c")
		for _, line := range each.lines {
			r.lines = append(r.lines, strings.NewReplacer("command", fmt.Sprint(r.pid), "child", fmt.Sprint(child)).Replace(line))
		}
		settlement := r.settle(t)
		if settlement.CommandExited != each.exited || each.exited && settlement.CommandStatus != each.status {
			t.Fatalf("%q: exited %v with %d, want %v with %d", each.lines, settlement.CommandExited, settlement.CommandStatus, each.exited, each.status)
		}
	}
}

// A reader is keyed by the bytes it read of another product, not by the key they're under: when a second settlement of
// the other product records the union of both read sets, its key moves and its bytes don't, and the reader still finds
// its product (home-7's dry run, Oct 11: 6 products went stale this way). New bytes still miss it, which
// TestAProductReadingAnotherIsKeyedByItsKey holds. Mutant: a product read valued by its key.
// Not parallel: newRig.
func TestAReaderOutlivesItsInputsKeyMoving(t *testing.T) {
	r := newRig(t)
	header := Inputs{Name: "header", Files: []string{"include"}}
	headerProduct := Product(t, header, building("header"))
	Product(t, port, func(directory string) error {
		content, err := os.ReadFile(filepath.Join(headerProduct, "product"))
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "product"), append(content, " and port"...), 0o644)
	})
	built := r.built(t)
	r.window(built["header"].Build, func() {
		r.open(r.pid, "include/x.h")
		r.write(r.pid, filepath.Join(headerProduct, "product"))
	})
	r.window(built["port"].Build, func() {
		r.open(r.pid, filepath.Join(headerProduct, "product"))
		r.open(r.pid, "source/a.c")
	})
	if settlement := r.settle(t); len(settlement.Settled) != 2 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	before, err := reading(t, header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reading(t, port); err != nil {
		t.Fatalf("port before header's key moved: %v", err)
	}
	// Another settlement of header that also read notes.txt: the union of the two sets, under a new key, the same bytes.
	set := readSet{Entries: sortedEntries([]readEntry{{Kind: "content", Path: "include/x.h"}, {Kind: "content", Path: "notes.txt"}, {Kind: "content", Path: "go.mod"}, {Kind: "settings", Path: "binary"}})}
	_, evaluated, _, err := recordUnion(r.root, r.cache, built["header"].NameKey, "header", set, func(key string) string {
		moved := filepath.Join(r.cache, key)
		if err := os.MkdirAll(moved, 0o755); err != nil {
			t.Fatal(err)
		}
		content, _ := os.ReadFile(filepath.Join(before, "product"))
		write(t, moved, "product", string(content))
		return moved
	})
	if err != nil {
		t.Fatal(err)
	}
	if after, err := reading(t, header); err != nil || after == before || after != filepath.Join(r.cache, evaluated.key) {
		t.Fatalf("header's key didn't move: %s, then %s (%v)", before, after, err)
	}
	if _, err := reading(t, port); err != nil {
		t.Fatalf("port, whose header's key moved and bytes didn't, read as %v", err)
	}
}

// A product's digest is its bytes wherever it sits: the same files in another directory digest the same, and a
// file's content, its executable bit or a link's target each move it. A later process reads it from beside the
// product, and one reached through a link is its bytes too. Mutants: the executable bit left out; the digest naming its
// directory.
func TestAProductsDigestIsItsBytes(t *testing.T) {
	t.Parallel()
	product := func(executable bool, content, target string) string {
		directory := filepath.Join(t.TempDir(), strings.Repeat("ab", 32))
		write(t, directory, "bin/tool", content)
		mode := os.FileMode(0o644)
		if executable {
			mode = 0o755
		}
		if err := os.Chmod(filepath.Join(directory, "bin/tool"), mode); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(directory, "current")); err != nil {
			t.Fatal(err)
		}
		return directory
	}
	digest := func(directory string) string {
		t.Helper()
		value, err := productDigest(directory)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	base := digest(product(true, "tool", "bin/tool"))
	if again := digest(product(true, "tool", "bin/tool")); again != base {
		t.Fatalf("the same files elsewhere digest %s, then %s", base, again)
	}
	for name, other := range map[string]string{"content": product(true, "tool 2", "bin/tool"), "executable bit": product(false, "tool", "bin/tool"), "link target": product(true, "tool", "bin")} {
		if digest(other) == base {
			t.Errorf("a product differing in its %s digests the same", name)
		}
	}
	kept := product(true, "tool", "bin/tool")
	digest(kept)
	productDigests.Delete(kept)
	if err := os.WriteFile(kept+".digest", []byte(strings.Repeat("c", 64)), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := digest(kept); got != strings.Repeat("c", 64) {
		t.Fatalf("a later process hashed again rather than read %s.digest: %s", filepath.Base(kept), got)
	}
	// A product reached through a link, as a runner's unit cache links one from the box's shared products, digests as
	// its bytes. Mutant: the link walked as itself.
	linked := filepath.Join(t.TempDir(), strings.Repeat("ab", 32))
	if err := os.Symlink(product(true, "tool", "bin/tool"), linked); err != nil {
		t.Fatal(err)
	}
	if got := digest(linked); got != base {
		t.Fatalf("a product reached through a link digests %s, its bytes %s", got, base)
	}
}

// A run that finds a product already built reports it, found, under the key it was found by: a test's products are
// the ones it settled and the ones it found (Planner's step 5 proof on Home, Oct 11: a tree that counted only
// settlements lost every product its tests found). One this run built itself is its settlement's line, never found
// too. Mutant: found products left out of the report.
// Not parallel: newRig.
func TestTheReportNamesWhatTheRunFound(t *testing.T) {
	r := newRig(t)
	header := Inputs{Name: "header", Files: []string{"include"}}
	Product(t, header, building("header"))
	r.window(r.built(t)["header"].Build, func() { r.open(r.pid, "include/x.h") })
	if settlement := r.settle(t); len(settlement.Settled) != 1 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	file, err := loadReads(r.cache, NameKey(r.root, header))
	if err != nil || len(file.Sets) == 0 {
		t.Fatalf("header's read sets: %+v, %v", file, err)
	}
	r.retrace(t, "found")
	Product(t, header, building("never built: found"))
	Product(t, port, building("port"))
	r.window(r.built(t)["port"].Build, func() { r.open(r.pid, "source/a.c") })
	reports := r.settle(t).Reports
	var found, settled []Report
	for _, report := range reports {
		switch report.Outcome {
		case ReportFound:
			found = append(found, report)
		case ReportSettled:
			settled = append(settled, report)
		}
	}
	if len(found) != 1 || found[0].Name != "header" || found[0].Key != file.Sets[0].Key || len(settled) != 1 || settled[0].Name != "port" {
		t.Fatalf("the report is %+v; want header found under %s and port settled", reports, file.Sets[0].Key[:12])
	}
}

// A run that finds a product whole reports, found, each product the set that found it reads, through theirs in turn: a
// runner values that set, each upstream's set in turn, and needs each one's read set and product (#mmpt615, 2e95d29e's
// void: tsprinter's mutant corpora were found whole, their Go oracle was never named, so no tree shipped it). Here port
// reads header, which reads base. Found in the cache or fetched from the store, the same. Mutants: the upstreams valued
// and never journaled; only a set's own upstreams named, never theirs; a fetch journaling its key's upstreams as none.
// Not parallel: newRig.
func TestARunThatFindsAProductNamesWhatItsSetReads(t *testing.T) {
	r := newRig(t)
	base := Inputs{Name: "base", Files: []string{"source"}}
	header := Inputs{Name: "header", Files: []string{"include"}}
	reading := func(upstream string) func(string) error {
		return func(directory string) error {
			content, err := os.ReadFile(filepath.Join(upstream, "product"))
			if err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(directory, "product"), append(content, " and more"...), 0o644)
		}
	}
	baseProduct := Product(t, base, building("base"))
	headerProduct := Product(t, header, reading(baseProduct))
	built := r.built(t)
	r.window(built["base"].Build, func() {
		r.open(r.pid, "source/b.c")
		r.write(r.pid, filepath.Join(baseProduct, "product"))
	})
	r.window(built["header"].Build, func() {
		r.open(r.pid, filepath.Join(baseProduct, "product"))
		r.open(r.pid, "include/x.h")
		r.write(r.pid, filepath.Join(headerProduct, "product"))
	})
	if settlement := r.settle(t); len(settlement.Settled) != 2 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	// port is built by a run of its own that finds header, so its set reads header alone, never base beside it.
	r.retrace(t, "port")
	headerFound := Product(t, header, building("never built: found"))
	Product(t, port, reading(headerFound))
	built["port"] = r.built(t)["port"]
	r.window(built["port"].Build, func() {
		r.open(r.pid, filepath.Join(headerFound, "product"))
		r.open(r.pid, "source/a.c")
	})
	if settlement := r.settle(t); len(settlement.Settled) != 1 {
		t.Fatalf("port: settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	if entries := r.entries(t, built["port"].NameKey); slices.Contains(entries, "product "+built["base"].NameKey) || !slices.Contains(entries, "product "+built["header"].NameKey) {
		t.Fatalf("port's set reads %q; want header's product and never base's", entries)
	}
	keys := map[string]string{}
	for _, name := range []string{"base", "header"} {
		file, err := loadReads(r.cache, built[name].NameKey)
		if err != nil || len(file.Sets) == 0 {
			t.Fatalf("%s's read sets: %+v, %v", name, file, err)
		}
		keys[name] = file.Sets[0].Key
	}
	foundBy := func(trace string, product func()) map[string]string {
		t.Helper()
		r.retrace(t, trace)
		product()
		found := map[string]string{}
		for _, report := range r.settle(t).Reports {
			if report.Outcome == ReportFound {
				found[report.NameKey] = report.Key
			}
		}
		return found
	}
	found := foundBy("found", func() { Product(t, port, building("never built: found")) })
	if len(found) != 3 || found[built["header"].NameKey] != keys["header"] || found[built["base"].NameKey] != keys["base"] || found[built["port"].NameKey] == "" {
		t.Fatalf("the run found %v; want port, the header its set reads under %s, and the base header's reads under %s", found, keys["header"][:12], keys["base"][:12])
	}
	// Fetched from the store under the key its set gives, the same: that set's upstreams are named.
	portKey := found[built["port"].NameKey]
	if err := os.Rename(filepath.Join(r.cache, portKey), filepath.Join(t.TempDir(), portKey)); err != nil {
		t.Fatal(err)
	}
	store := &fakeStore{objects: map[string][]byte{}}
	server := httptest.NewServer(store)
	t.Cleanup(server.Close)
	t.Setenv("ADAMIC_BUILD_STORE", server.URL)
	t.Setenv("ADAMIC_BUILD_AUDIT", "0")
	store.put(t, portKey, map[string]string{"product": "base and more and more"})
	forgetHeldForLife()
	t.Cleanup(forgetHeldForLife)
	found = foundBy("fetched", func() { Product(t, port, building("never built: fetched")) })
	if len(found) != 3 || found[built["header"].NameKey] != keys["header"] || found[built["base"].NameKey] != keys["base"] || found[built["port"].NameKey] != portKey || len(store.reads) == 0 {
		t.Fatalf("the run fetched port and found %v, after %d store reads; want port, header and base", found, len(store.reads))
	}
}

// A build that looks for a product in the cache and doesn't find it read nothing: a pending product of this trace that
// isn't there, or a key with no product, is no entry and no refusal. Reading what is there of a product the run
// never built is still refused. Mutant: a missed lookup classed by its path.
// Not parallel: newRig.
func TestALookupInTheCacheThatMissedIsNoRead(t *testing.T) {
	r := newRig(t)
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	r.window(build.Build, func() {
		r.open(r.pid, "source/a.c")
		r.missing(r.pid, filepath.Join(r.cache, "pending", filepath.Base(r.trace), strings.Repeat("c2", 32)))
		r.missing(r.pid, filepath.Join(r.cache, strings.Repeat("d3", 32), "product"))
	})
	if settlement := r.settle(t); len(settlement.Settled) != 1 || len(settlement.Refused) != 0 {
		t.Fatalf("lookups that missed: settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	r.retrace(t, "read")
	Product(t, Inputs{Name: "reader", Files: []string{"source"}}, building("reader"))
	reader := r.built(t)["reader"]
	r.window(reader.Build, func() { r.open(r.pid, filepath.Join(r.cache, strings.Repeat("e4", 32), "product")) })
	if settlement := r.settle(t); len(settlement.Refused) != 1 || !strings.Contains(settlement.Refused[0], "unknown product") {
		t.Fatalf("a read of a product nobody built: settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
}

// The directories a module's go.mod and go.work replace modules with, as the recipe keys their go.mod and go.sum: a
// replaced module selects versions through its go.mod whether or not the binary links a package of it (#c2chyab).
// A versioned replacement names no directory. Mutant: the workspace's replacements left out.
func TestLocalReplacementsAreTheMainModulesAndTheWorkspaces(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write(t, root, "go.mod", "module example.com/main\n\ngo 1.22\n\nreplace example.com/shim => ./shim\n\nreplace example.com/far => example.com/other v1.0.0\n")
	write(t, root, "shim/go.mod", "module example.com/shim\n\ngo 1.22\n")
	write(t, root, "go.work", "go 1.22\n\nuse .\n\nreplace example.com/tool => ./tool\n")
	directories, err := localReplacements(root, root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{filepath.Join(root, "shim"), filepath.Join(root, "tool")}
	if !slices.Equal(directories, want) {
		t.Fatalf("replacements %q, want %q", directories, want)
	}
}

// A miss counts every differing entry by kind, the ones past the eighth it names included (Judge, #m5z9d6q): Loom's
// judge voids a miss as the box's only when every differing entry is a system file, so a tree file past the eighth
// must still show in the counts. Here eighteen system files differ, a tree file's content past the eighth, and a
// product last; one system file is unchanged.
func TestAMissCountsEveryDifferingEntryByKind(t *testing.T) {
	t.Parallel()
	var set readSet
	var values []string
	for index := range 21 {
		entry := readEntry{Kind: "system", Path: fmt.Sprintf("/usr/include/h%02d.h", index), Value: shortValue("recorded")}
		switch index {
		case 15:
			entry.Kind, entry.Path = "content", "source/a.c"
		case 20:
			entry.Kind, entry.Path = "product", "0123abcd"
		}
		set.Entries = append(set.Entries, entry)
		values = append(values, "now")
	}
	values[3] = "recorded"
	got := changed(set, evaluation{key: "k", values: values})
	want := "20 of 21 entries it read differ from its newest traced build (by kind: content 1, product 1, system 18): " +
		"system /usr/include/h00.h, system /usr/include/h01.h, system /usr/include/h02.h, system /usr/include/h04.h, " +
		"system /usr/include/h05.h, system /usr/include/h06.h, system /usr/include/h07.h, system /usr/include/h08.h, and 12 more"
	if got != want {
		t.Fatalf("the miss reads\n%s\nwant\n%s", got, want)
	}
	// Eight or fewer are all named, with their counts still ahead of them.
	got = changed(readSet{Entries: set.Entries[14:17]}, evaluation{key: "k", values: values[14:17]})
	if want := "3 of 3 entries it read differ from its newest traced build (by kind: content 1, system 2): system /usr/include/h14.h, content source/a.c, system /usr/include/h16.h"; got != want {
		t.Fatalf("the miss reads\n%s\nwant\n%s", got, want)
	}
}

// A dependency's ignored files stay in a recipe, since a build-constraint edit can compile one in, but not its ignored
// test files: another platform's _test.go is compiled into no binary but its own package's tests, which still key it.
// Mutants, each caught: a dependency's ignored test files kept; its other ignored files dropped; the package under
// test's ignored test files dropped.
// Not parallel: go list runs in the module.
func TestARecipeKeepsADependencysIgnoredFilesButNotItsIgnoredTests(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	other := "windows"
	if runtime.GOOS == "windows" {
		other = "linux"
	}
	for name, content := range map[string]string{
		"go.mod":                    "module example.com/ignored\n\ngo 1.22\n",
		"a/a.go":                    "package a\n\nfunc A() int { return 1 }\n",
		"a/a_" + other + ".go":      "package a\n\nfunc onOther() {}\n",
		"a/a_" + other + "_test.go": "package a\n",
		"a/a_test.go":               "package a\n",
		"cmd/b/main.go":             "package main\n\nimport \"example.com/ignored/a\"\n\nfunc main() { _ = a.A() }\n",
	} {
		write(t, root, name, content)
	}
	gitIn(t, root, "init", "-q")
	gitIn(t, root, "add", ".")
	gitIn(t, root, "commit", "-q", "-m", "module")
	forgetTracked()
	t.Setenv("GOWORK", "off")
	binary := recipeOf(root, journalEntry{Package: "example.com/ignored/cmd/b"})
	if binary.reason != "" {
		t.Fatal(binary.reason)
	}
	has := func(found recipe, path string) bool {
		return slices.Contains(found.entries, readEntry{Kind: "content", Path: path})
	}
	if !has(binary, "a/a_"+other+".go") || has(binary, "a/a_"+other+"_test.go") || has(binary, "a/a_test.go") {
		t.Fatalf("cmd/b's recipe %+v: a's ignored file kept, its test files, ignored or not, left out", binary.entries)
	}
	tests := recipeOf(root, journalEntry{Package: "example.com/ignored/a.test"})
	if tests.reason != "" || !has(tests, "a/a_"+other+"_test.go") || !has(tests, "a/a_test.go") {
		t.Fatalf("a's own tests' recipe %+v (%s) lacks its test files", tests.entries, tests.reason)
	}
}

// TypeScript's JavaScript sys asks whether the file system is case-sensitive by looking for its own __filename with
// every letter's case swapped: a file node read, not a program it ran (#4gp2ytz, Compiler's find in stage3/api's
// typescript.js). That lookup is the file system's case, never an absolute path naming the checkout, so one tree at two
// checkout paths keys alike. Mutant that fails it: caseProbe asking only of programs run, so the swapped path is keyed
// absolute.
// Not parallel: newRig.
func TestACaseProbeOfAFileTheRunReadIsTheFileSystemsCase(t *testing.T) {
	r := newRig(t)
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	read := filepath.Join(r.root, "source/a.c")
	r.window(build.Build, func() {
		r.open(r.pid, "source/a.c")
		r.missing(r.pid, swapCase(read))
	})
	if settlement := r.settle(t); len(settlement.Settled) != 1 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	entries := r.entries(t, build.NameKey)
	if !slices.Contains(entries, "filesystem case") || slices.ContainsFunc(entries, func(entry string) bool { return strings.HasPrefix(entry, "absolute ") }) {
		t.Fatalf("entries %q: want the file system's case, and no absolute path naming the checkout", entries)
	}
}

// leanRun settles one traced run of a build of port, lean or not, in a rig of its own: before is a file its test reads
// before the build (a shard's own case), and the build reads source/a.c in its window, and notes.txt in a process it
// starts. It returns the run's report for port and the set's entries.
func leanRun(t *testing.T, lean bool, before string) (Report, []string) {
	t.Helper()
	r := newRig(t)
	if lean {
		LeanProduct(t, port, building("built"))
	} else {
		Product(t, port, building("built"))
	}
	build := r.built(t)["port"]
	child := r.pid + 1
	r.open(r.pid, before)
	r.window(build.Build, func() {
		r.open(r.pid, "source/a.c")
		r.add(r.pid, `clone(child_stack=NULL, flags=CLONE_CHILD_CLEARTID|SIGCHLD) = %d`, child)
		r.open(child, "notes.txt")
		r.add(child, "+++ exited with 0 +++")
	})
	settlement := r.settle(t)
	if len(settlement.Settled) != 1 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	for _, report := range settlement.Reports {
		if report.Outcome == ReportSettled {
			return report, r.entries(t, build.NameKey)
		}
	}
	t.Fatalf("no settled report: %+v", settlement.Reports)
	return Report{}, nil
}

// A lean build's read set is its own (#mmpt615, Node's ruling): what its process and the processes it starts read in
// its window, and every file its declared Files name, read or not, never what its test read before it. So two tests
// that read different cases before asking for one lean product settle it under one key, where a build that isn't lean
// keeps the superset, its test's earlier reads included, and settles two. The report names how each was taken.
// Mutants that each fail it: lean ignored; the declared files left out; the window's subprocesses dropped.
// Not parallel: newRig.
func TestALeanBuildIsKeyedByItsWindowAndItsDeclaredFilesAlone(t *testing.T) {
	first, entries := leanRun(t, true, "go.sum")
	if first.Window != BuildWindowed {
		t.Fatalf("a lean build's reads were taken as %q", first.Window)
	}
	for _, want := range []string{"content source/a.c", "content source/b.c", "listing source", "content include/x.h", "content notes.txt"} {
		if !slices.Contains(entries, want) {
			t.Errorf("the lean set lacks %q: %q", want, entries)
		}
	}
	if slices.Contains(entries, "content go.sum") {
		t.Errorf("the lean set holds what its test read before it: %q", entries)
	}
	if second, _ := leanRun(t, true, "include/x.h"); second.Key != first.Key {
		t.Errorf("two tests reading different files before one lean build settled it under %s and %s", first.Key[:12], second.Key[:12])
	}
	one, oneEntries := leanRun(t, false, "go.sum")
	two, _ := leanRun(t, false, "notes.txt")
	if one.Window != BuildSuperset || !slices.Contains(oneEntries, "content go.sum") || one.Key == two.Key {
		t.Errorf("a build that isn't lean: taken as %q, holds go.sum %v, keys %s and %s; want the superset, its test's reads in it, two keys",
			one.Window, slices.Contains(oneEntries, "content go.sum"), one.Key[:12], two.Key[:12])
	}
}

// Only the packages buildcache reviewed build lean (leanCallers): a lean build is keyed without what its caller read
// before it, so a caller that handed it bytes of its own would key it without them. A caller is any file that imports
// buildcache, under any name, a dot import too, and calls LeanProduct or LeanGet, buildcache's own subpackages included
// (Mentat's S3: a text match let an aliased import through). buildcache's own files are the reviewer, never a caller,
// and cohere is a module of its own, which can't import adamic's internal packages. Mutants: a caller added to the tree
// and not to leanCallers; one importing buildcache under another name.
func TestOnlyReviewedPackagesBuildLean(t *testing.T) {
	t.Parallel()
	root := repositoryRootForTest(t)
	callers := leanCallersIn(t, root)
	for _, caller := range callers {
		if !slices.Contains(leanCallers, caller) {
			t.Errorf("%s builds lean, and isn't one of buildcache's reviewed lean callers %v", caller, leanCallers)
		}
	}
	if !slices.Contains(callers, "internal/native/runtimeproducts") {
		t.Fatalf("lean callers found: %v; want runtimeproducts among them", callers)
	}
	aliased := t.TempDir()
	write(t, aliased, "caller/aliased.go", "package caller\n\nimport cache \"github.com/system-inc/adamic/internal/buildcache\"\n\nvar _, _ = cache.LeanGet(cache.Inputs{}, nil)\n")
	write(t, aliased, "dotted/dotted.go", "package dotted\n\nimport . \"github.com/system-inc/adamic/internal/buildcache\"\n\nvar _, _ = LeanGet(Inputs{}, nil)\n")
	write(t, aliased, "plain/plain.go", "package plain\n\nimport \"github.com/system-inc/adamic/internal/buildcache\"\n\nvar _, _ = buildcache.Get(buildcache.Inputs{}, nil)\n")
	if found := leanCallersIn(t, aliased); !slices.Equal(found, []string{"caller", "dotted"}) {
		t.Errorf("lean callers by an alias and a dot import: %v; want caller and dotted", found)
	}
}

// leanCallersIn is every package under root, by its directory relative to root, a file of which calls buildcache's
// LeanProduct or LeanGet.
func leanCallersIn(t *testing.T, root string) []string {
	t.Helper()
	const importPath = "github.com/system-inc/adamic/internal/buildcache"
	found := map[string]bool{}
	filepath.WalkDir(root, func(name string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "node_modules" || name == filepath.Join(root, "cohere")) {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || filepath.Dir(name) == filepath.Join(root, "internal", "buildcache") {
			return nil
		}
		content, err := os.ReadFile(name)
		if err != nil || !strings.Contains(string(content), importPath) {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), name, content, parser.SkipObjectResolution)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			return nil
		}
		local := ""
		for _, imported := range file.Imports {
			if strings.Trim(imported.Path.Value, "`\"") != importPath {
				continue
			}
			local = "buildcache"
			if imported.Name != nil {
				local = imported.Name.Name
			}
		}
		lean := func(called string) bool { return called == "LeanProduct" || called == "LeanGet" }
		ast.Inspect(file, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.SelectorExpr:
				if receiver, ok := node.X.(*ast.Ident); ok && local != "" && local != "." && receiver.Name == local && lean(node.Sel.Name) {
					found[filepath.Dir(name)] = true
				}
			case *ast.Ident:
				if local == "." && lean(node.Name) {
					found[filepath.Dir(name)] = true
				}
			}
			return true
		})
		return nil
	})
	var callers []string
	for directory := range found {
		relative, _ := filepath.Rel(root, directory)
		callers = append(callers, filepath.ToSlash(relative))
	}
	sort.Strings(callers)
	return callers
}

// leanSettle settles one lean build of inputs in a rig of its own, with trace writing the run around the build's id. It returns the settlement, the settled report (empty
// when none settled) and the settled set's entries.
func leanSettle(t *testing.T, inputs Inputs, trace func(r *rig, id string)) (Settlement, Report, []string) {
	t.Helper()
	r := newRig(t)
	product(t, inputs, building("built"), &leanBuild{})
	build := r.built(t)[inputs.Name]
	trace(r, build.Build)
	settlement := r.settle(t)
	for _, report := range settlement.Reports {
		if report.Outcome == ReportSettled {
			return settlement, report, r.entries(t, build.NameKey)
		}
	}
	return settlement, Report{}, nil
}

// A file a lean build's test wrote before its window is never the build's own output (Mentat's M1): the reads behind
// those bytes are the test's, outside the window, so the file reads by content, and a temporary file no key can name
// refuses the build. One the build's processes wrote in its window is theirs, whatever was written there before. Mutants
// that each fail it: the window's writes counted since the run began; a writer kept once a process for the whole run.
// Not parallel: newRig.
func TestAWriteBeforeALeanBuildsWindowIsAnInputOfIt(t *testing.T) {
	const generated = "/tmp/adamic-written-before-the-window/gen.h"
	before := func(r *rig, id string) {
		r.open(r.pid, "notes.txt")
		r.write(r.pid, generated)
		r.window(id, func() {
			r.open(r.pid, "source/a.c")
			r.open(r.pid, generated)
		})
	}
	if settlement, _, _ := leanSettle(t, port, before); len(settlement.Refused) != 1 || !strings.Contains(settlement.Refused[0], "temp not made by this run") {
		t.Errorf("a lean build reading what its test wrote before its window: settled %q, refused %q; want it refused as temp not made by this run",
			settlement.Settled, settlement.Refused)
	}
	for name, trace := range map[string]func(r *rig, id string){
		"written in the window": func(r *rig, id string) {
			r.window(id, func() {
				r.write(r.pid, generated)
				r.open(r.pid, generated)
			})
		},
		"written before and again in the window": func(r *rig, id string) {
			r.write(r.pid, generated)
			r.window(id, func() {
				r.write(r.pid, generated)
				r.open(r.pid, generated)
			})
		},
	} {
		if settlement, report, _ := leanSettle(t, port, trace); report.Window != BuildWindowed {
			t.Errorf("%s: settled %q, refused %q; want it the build's own output, settled windowed", name, settlement.Settled, settlement.Refused)
		}
	}
}

// A build-begin from a thread the trace hasn't placed yet opens its window at the marker, for the process the build's
// id names, so what that process's placed threads read before the thread is placed is the build's (Mentat's M2). A
// marker whose id names a process the trace doesn't know yet opens the window once its thread is placed, late, and a
// lean build whose window opened late keys by its superset, as its report says (S1). One opened for the id's process
// whose thread the trace then placed in another process is lost. Mutants that each fail it: no opening at the marker; a
// late window read as placed; the marker's process never checked against the opening's.
// Not parallel: newRig.
func TestALeanBuildsWindowOpensAtItsMarker(t *testing.T) {
	thread := func(r *rig) int { return r.pid + 7 }
	settlement, report, entries := leanSettle(t, port, func(r *rig, id string) {
		r.open(r.pid, "go.sum")
		r.marker(thread(r), "build-begin", id)
		r.open(r.pid, "notes.txt")
		r.add(r.pid, `clone(child_stack=0x7f, flags=CLONE_VM|CLONE_THREAD) = %d`, thread(r))
		r.open(thread(r), "source/a.c")
		r.marker(thread(r), "build-end", id)
	})
	if report.Window != BuildWindowed || !slices.Contains(entries, "content notes.txt") || slices.Contains(entries, "content go.sum") {
		t.Errorf("a marker on a thread placed after it: settled %q, refused %q, taken as %q, entries %q; want it windowed, holding notes.txt read before its thread was placed, never go.sum",
			settlement.Settled, settlement.Refused, report.Window, entries)
	}
	settlement, report, entries = leanSettle(t, port, func(r *rig, id string) {
		command := r.pid + 100
		r.open(command, "go.sum")
		r.marker(r.pid, "build-begin", id)
		r.add(command, `clone(child_stack=NULL, flags=CLONE_CHILD_CLEARTID|SIGCHLD) = %d`, r.pid)
		r.open(r.pid, "source/a.c")
		r.marker(r.pid, "build-end", id)
	})
	if report.Window != BuildLeanSuperset || !slices.Contains(entries, "content source/a.c") {
		t.Errorf("a marker of a process the trace placed after it: settled %q, refused %q, taken as %q, entries %q; want the lean superset",
			settlement.Settled, settlement.Refused, report.Window, entries)
	}
	other := func(r *rig) int { return r.pid + 50 }
	settlement, report, _ = leanSettle(t, port, func(r *rig, id string) {
		r.open(r.pid, "go.sum")
		r.add(r.pid, `clone(child_stack=NULL, flags=CLONE_CHILD_CLEARTID|SIGCHLD) = %d`, other(r))
		r.marker(thread(r), "build-begin", id)
		r.add(other(r), `clone(child_stack=0x7f, flags=CLONE_VM|CLONE_THREAD) = %d`, thread(r))
		r.open(thread(r), "source/a.c")
		r.marker(thread(r), "build-end", id)
	})
	if report.Window != BuildLeanSuperset {
		t.Errorf("a window opened for the id's process, its thread placed in another while it's open: settled %q, refused %q, taken as %q; want the lean superset",
			settlement.Settled, settlement.Refused, report.Window)
	}
	settlement, _, _ = leanSettle(t, port, func(r *rig, id string) {
		r.open(r.pid, "go.sum")
		r.add(r.pid, `clone(child_stack=NULL, flags=CLONE_CHILD_CLEARTID|SIGCHLD) = %d`, other(r))
		r.marker(thread(r), "build-begin", id)
		r.marker(r.pid, "build-end", id)
		r.add(other(r), `clone(child_stack=0x7f, flags=CLONE_VM|CLONE_THREAD) = %d`, thread(r))
	})
	if len(settlement.Refused) != 1 || !strings.Contains(settlement.Refused[0], "the trace placed its marker in another") {
		t.Errorf("a window opened for the id's process, ended, its thread then placed in another: settled %q, refused %q; want it lost", settlement.Settled, settlement.Refused)
	}
}

// A lean build's declared path it can't read whole refuses it, named, never dropped (Mentat's S2, Node's must): the
// build may have taken its bytes before its window, and the key would lack them. A file or a directory the tree doesn't
// carry, and a directory that can't be read, each refuse; a declared path that isn't there reads as absent and settles.
// Mutants that each fail it: an untracked file dropped; an untracked directory skipped; a walk's error ignored; the
// declared directory's own error read as absent.
// Not parallel: newRig.
func TestALeanBuildsDeclaredPathItCantReadWholeRefusesIt(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root reads a directory whatever its mode")
	}
	seal := func(t *testing.T, directory string) {
		if err := os.Chmod(directory, 0); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Chmod(directory, 0o755) })
	}
	withFiles := func(files ...string) Inputs {
		inputs := port
		inputs.Files = files
		return inputs
	}
	for name, test := range map[string]struct {
		inputs  Inputs
		make    func(t *testing.T, r *rig)
		refusal string
	}{
		"an untracked declared file": {withFiles("source", "stray.h"), func(t *testing.T, r *rig) { write(t, r.root, "stray.h", "#define S 1\n") },
			"stray.h is a file the tree doesn't carry"},
		"an untracked file in a declared directory": {port, func(t *testing.T, r *rig) { write(t, r.root, "source/extra.c", "int e;\n") },
			"extra.c is a file the tree doesn't carry"},
		"an untracked directory in a declared directory": {port, func(t *testing.T, r *rig) { write(t, r.root, "source/generated/g.c", "int g;\n") },
			"generated is a directory the tree doesn't carry"},
		"a declared directory it can't read": {port, func(t *testing.T, r *rig) { seal(t, filepath.Join(r.root, "include")) },
			"declared include can't be read whole"},
		"a directory it can't read in a declared directory": {port, func(t *testing.T, r *rig) {
			write(t, r.root, "source/deep/d.c", "int d;\n")
			gitIn(t, r.root, "add", ".")
			gitIn(t, r.root, "commit", "-q", "-m", "deep")
			forgetTracked()
			seal(t, filepath.Join(r.root, "source", "deep"))
		}, "declared source can't be read whole"},
		"a declared file that isn't there": {withFiles("source", "gone.h"), nil, ""},
	} {
		settlement, report, entries := leanSettle(t, test.inputs, func(r *rig, id string) {
			if test.make != nil {
				test.make(t, r)
			}
			r.window(id, func() { r.open(r.pid, "source/a.c") })
		})
		switch {
		case test.refusal == "" && (report.Window != BuildWindowed || !slices.Contains(entries, "exists gone.h")):
			t.Errorf("%s: settled %q, refused %q, entries %q; want it settled, the path read as absent", name, settlement.Settled, settlement.Refused, entries)
		case test.refusal != "" && (len(settlement.Refused) != 1 || !strings.Contains(settlement.Refused[0], test.refusal)):
			t.Errorf("%s: settled %q, refused %q; want it refused naming %q", name, settlement.Settled, settlement.Refused, test.refusal)
		}
	}
}

// A windowed lean build of buildcache's own is keyed by the code that makes it (leanRecipeFor, #earlycut, Node's approved
// design), never by the recipe of the binary that settled it. A build that isn't lean, a lean build whose window opened
// late, and a lean caller's windowed build keep the binary's whole recipe: a caller's code is keyed lean only once its
// census holds, and native's archives reach the jobserver until it's handed in (#yzwm4rj, Node's reading 1). Mutants
// that each fail it: a windowed lean build given the whole recipe; a late window given the lean recipe; a caller's
// windowed build given the lean recipe; Report.Lean set for the whole recipe.
// Not parallel: newRig, and it replaces recipeFor and leanRecipeFor.
func TestALeanBuildIsKeyedByTheCodeThatMakesIt(t *testing.T) {
	caller := modulePath + "/internal/native/runtimeproducts"
	run := func(lean *leanBuild, late bool) (Report, []string, []string) {
		t.Helper()
		r := newRig(t)
		var asked []string
		recipeFor = func(string, journalEntry) recipe {
			asked = append(asked, "whole")
			return recipe{entries: []readEntry{{Kind: "content", Path: "go.mod"}}}
		}
		leanRecipeFor = func(_, called string, _ journalEntry) recipe {
			asked = append(asked, "lean "+called)
			return recipe{entries: []readEntry{{Kind: "content", Path: "notes.txt"}}}
		}
		product(t, port, building("built"), lean)
		build := r.built(t)["port"]
		if late {
			command := r.pid + 100
			r.open(command, "go.sum")
			r.marker(r.pid, "build-begin", build.Build)
			r.add(command, `clone(child_stack=NULL, flags=CLONE_CHILD_CLEARTID|SIGCHLD) = %d`, r.pid)
			r.open(r.pid, "source/a.c")
			r.marker(r.pid, "build-end", build.Build)
		} else {
			r.window(build.Build, func() { r.open(r.pid, "source/a.c") })
		}
		settlement := r.settle(t)
		for _, report := range settlement.Reports {
			if report.Outcome == ReportSettled {
				return report, r.entries(t, build.NameKey), asked
			}
		}
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
		return Report{}, nil, nil
	}
	report, entries, asked := run(&leanBuild{}, false)
	if report.Window != BuildWindowed || !report.Lean || !slices.Contains(entries, "content notes.txt") || slices.Contains(entries, "content go.mod") || !slices.Equal(asked, []string{"lean "}) {
		t.Errorf("a windowed lean build of buildcache's own: taken as %q, entries %q, recipes asked %q; want the lean recipe alone", report.Window, entries, asked)
	}
	report, entries, asked = run(&leanBuild{caller: caller}, false)
	if report.Window != BuildWindowed || report.Lean || !slices.Contains(entries, "content go.mod") || slices.Contains(entries, "content notes.txt") || !slices.Equal(asked, []string{"whole"}) {
		t.Errorf("a lean caller's windowed build: taken as %q, entries %q, recipes asked %q; want its window and the binary's whole recipe", report.Window, entries, asked)
	}
	for name, lean := range map[string]*leanBuild{"a build that isn't lean": nil, "a lean build whose window opened late": {}} {
		_, entries, asked := run(lean, lean != nil)
		if !slices.Contains(entries, "content go.mod") || slices.Contains(entries, "content notes.txt") || !slices.Equal(asked, []string{"whole"}) {
			t.Errorf("%s: entries %q, recipes asked %q; want the binary's whole recipe", name, entries, asked)
		}
	}
}

// A lean build's recipe is the code that makes it, from the tree (leanRecipeOf): buildcache and every package of the tree
// it links, each by every non-test file, and a lean caller's own package, never what the caller links. So a runtime
// archive keys runtimeproducts, internal/toolchain and buildcache's gobuild.go (goProduct's closure and
// withOutsideHeaders, which run a go build) and no file of the compiler, where the settling binary's whole recipe keys
// the compiler's. buildcache's own lean build keys no caller and buildcache whole too, and a process naming no package
// is refused, as recipeOf refuses it. Mutants that each fail it: the caller listed with what it links; buildcache keyed
// by its make files alone; the caller left out; the refusal dropped.
func TestALeanRecipeIsTheCodeThatMakesTheBuild(t *testing.T) {
	t.Parallel()
	root := repositoryRootForTest(t)
	caller := modulePath + "/internal/native/runtimeproducts"
	paths := func(code recipe) []string {
		t.Helper()
		if code.reason != "" {
			t.Fatalf("no recipe: %s", code.reason)
		}
		var listed []string
		for _, entry := range code.entries {
			listed = append(listed, entry.Kind+" "+entry.Path)
		}
		sort.Strings(listed)
		return listed
	}
	settling := journalEntry{Package: caller + ".test"}
	lean, own := paths(leanRecipeOf(root, caller, settling)), paths(leanRecipeOf(root, "", journalEntry{Package: buildcachePackage + ".test"}))
	whole := paths(recipeOf(root, settling))
	if refused := leanRecipeOf(root, caller, journalEntry{Executable: "/tmp/go-build1/b001/exe/main"}); refused.reason == "" {
		t.Errorf("a process naming no package has a lean recipe: %d entries", len(refused.entries))
	}
	has := func(listed []string, prefix string) bool {
		return slices.ContainsFunc(listed, func(entry string) bool { return strings.HasPrefix(entry, prefix) })
	}
	for _, want := range []string{"content internal/buildcache/make_go.go", "content internal/buildcache/gobuild.go", "content internal/buildcache/settle.go", "content internal/toolchain/", "content internal/native/runtimeproducts/runtimeproducts.go", "listing internal/native/runtimeproducts", "content go.mod"} {
		if !has(lean, want) {
			t.Errorf("a runtime archive's lean recipe lacks %q: %q", want, lean)
		}
	}
	for _, unwanted := range []string{"content internal/native/library.go", "content internal/native/emit_expressions.go", "content internal/ir/", "content internal/flow/", "content bridge/", "content internal/buildcache/precise_test.go"} {
		if has(lean, unwanted) {
			t.Errorf("a runtime archive's lean recipe holds %q", unwanted)
		}
	}
	if has(own, "content internal/native/runtimeproducts/") || !has(own, "content internal/buildcache/gobuild.go") {
		t.Errorf("buildcache's own lean recipe: %q", own)
	}
	if !has(whole, "content internal/native/emit_expressions.go") {
		t.Errorf("the settling binary's whole recipe lacks the compiler: %q", whole)
	}
}

// A lean build is buildcache's own when buildcache asks for it, a reviewed caller's when one of leanCallers does, and no
// lean build when any other package does: its code is unreviewed, so it keeps the whole recipe. The caller is read from
// the runtime's spelling of a function. Mutant: an unreviewed caller's build made lean.
func TestALeanBuildIsBuildcachesOrAReviewedCallers(t *testing.T) {
	t.Parallel()
	for symbol, want := range map[string]string{
		"github.com/system-inc/adamic/internal/native/runtimeproducts.Build.func1": modulePath + "/internal/native/runtimeproducts",
		"github.com/system-inc/adamic/internal/buildcache.goProduct":               buildcachePackage,
		"github.com/system-inc/adamic/internal/buildcache.(*goRequest).build":      buildcachePackage,
		"main.main": "main",
	} {
		if got := functionPackage(symbol); got != want {
			t.Errorf("%s is %q's, want %q", symbol, got, want)
		}
	}
	if lean := leanBuildOf(buildcachePackage); lean == nil || lean.caller != "" {
		t.Errorf("buildcache's own build: %+v", lean)
	}
	if lean := leanBuildOf(modulePath + "/internal/native/runtimeproducts"); lean == nil || lean.caller != modulePath+"/internal/native/runtimeproducts" {
		t.Errorf("a reviewed caller's build: %+v", lean)
	}
	if lean := leanBuildOf(modulePath + "/stage1/cohere/lint"); lean != nil {
		t.Errorf("an unreviewed caller's build is lean: %+v", lean)
	}
}

// A settled product whose name key ADAMIC_BUILD_AUDIT_KEYS lists is rebuilt on its local hit under the process's own
// temporary directory and compared, and the audit line Loom reads says clean or mismatch, keeping the rebuilt copy on a
// mismatch (#earlycut's sampled key check). A name key it doesn't list is never rebuilt, and a list named that can't be
// read fails the lookup. Mutants that each fail it: the audit skipped on a hit; the rebuild placed beside the product; an
// unlisted name key audited; an unreadable list noted and passed over.
// Not parallel: newRig.
func TestAListedProductIsRebuiltAndComparedOnItsLocalHit(t *testing.T) {
	r := newRig(t)
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	r.window(build.Build, func() { r.open(r.pid, "source/a.c") })
	if settlement := r.settle(t); len(settlement.Settled) != 1 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	t.Setenv("ADAMIC_BUILD_TRACE", "")
	temporary, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TMPDIR", temporary)
	look := func(listed string, content string) string {
		t.Helper()
		forgetHeldForLife()
		keys := filepath.Join(t.TempDir(), "keys")
		if err := os.WriteFile(keys, []byte(listed+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Setenv("ADAMIC_BUILD_AUDIT_KEYS", keys)
		log := filepath.Join(t.TempDir(), "builds.log")
		t.Setenv("ADAMIC_BUILD_LOG", log)
		var rebuilt []string
		if _, err := Get(port, func(directory string) error {
			rebuilt = append(rebuilt, directory)
			return building(content)(directory)
		}); err != nil {
			t.Fatal(err)
		}
		for _, directory := range rebuilt {
			if !strings.HasPrefix(directory, temporary+string(filepath.Separator)) {
				t.Errorf("rebuilt at %s, beside the product, not under the process's temporary directory", directory)
			}
		}
		lines, _ := os.ReadFile(log)
		return string(lines)
	}
	if lines := look(build.NameKey, "built"); !regexp.MustCompile(`(?m)^audit port [0-9a-f]{64} clean [0-9a-f]{64}$`).MatchString(lines) {
		t.Errorf("a listed product rebuilt alike: census %q, want a clean audit line", lines)
	}
	lines := look(build.NameKey, "built differently")
	match := regexp.MustCompile(`(?m)^audit port [0-9a-f]{64} mismatch ([0-9a-f]{64}) ([0-9a-f]{64}) (.+)$`).FindStringSubmatch(lines)
	if match == nil || match[1] == match[2] {
		t.Fatalf("a listed product rebuilt otherwise: census %q, want a mismatch audit line", lines)
	}
	if kept, err := os.ReadFile(filepath.Join(match[3], "product")); err != nil || string(kept) != "built differently" {
		t.Errorf("the rebuilt copy isn't kept at %s: %q, %v", match[3], kept, err)
	}
	if lines := look(strings.Repeat("0", 64), "built differently"); strings.Contains(lines, "audit ") {
		t.Errorf("an unlisted product was audited: census %q", lines)
	}
	// A list named that can't be read fails the lookup, naming it, never an audit of nothing that reads clean.
	forgetHeldForLife()
	missing := filepath.Join(t.TempDir(), "no-such-list")
	t.Setenv("ADAMIC_BUILD_AUDIT_KEYS", missing)
	if _, err := Get(port, building("built")); err == nil || !strings.Contains(err.Error(), missing) {
		t.Errorf("a lookup under an unreadable audit list: %v, want an error naming %s", err, missing)
	}
}

// Under ADAMIC_BUILD_LEAN=off, Loom's switch while a lean mismatch is open, no build is lean: buildcache's own and a
// reviewed caller's alike take the whole recipe. Mutant: the switch unread.
// Not parallel: t.Setenv.
func TestNoBuildIsLeanWhileLoomTurnsLeanOff(t *testing.T) {
	for _, caller := range []string{buildcachePackage, modulePath + "/internal/native/runtimeproducts"} {
		if leanBuildOf(caller) == nil {
			t.Fatalf("%s's build isn't lean with lean on", caller)
		}
	}
	t.Setenv("ADAMIC_BUILD_LEAN", "off")
	for _, caller := range []string{buildcachePackage, modulePath + "/internal/native/runtimeproducts"} {
		if lean := leanBuildOf(caller); lean != nil {
			t.Errorf("%s's build is lean under ADAMIC_BUILD_LEAN=off: %+v", caller, lean)
		}
	}
}
