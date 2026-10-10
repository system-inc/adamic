package buildcache

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// A rig is a traced Workshop in miniature (#vt46geg): a git checkout of its own as the repository, a build cache, and
// a trace directory the products build under. The trace itself is written by the test, as strace -f would print it,
// so each test says exactly what each build read.
type rig struct {
	root, cache, trace string
	pid                int
	lines              []string
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
	original := recipeFor
	recipeFor = func(string, journalEntry) recipe {
		return recipe{entries: []readEntry{{Kind: "content", Path: "go.mod"}}}
	}
	t.Cleanup(func() { recipeFor = original })
	return &rig{root: root, cache: cache, trace: trace, pid: os.Getpid()}
}

func gitIn(t *testing.T, directory string, arguments ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", directory, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "init.defaultBranch=main"}, arguments...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", arguments, err, output)
	}
}

// forgetTracked drops what git ls-files said, after a test changes what a repository tracks.
func forgetTracked() {
	repositories.Range(func(key, _ any) bool { repositories.Delete(key); return true })
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
	settlement, err := Settle(r.trace, strings.NewReader(strings.Join(r.lines, "\n")+"\n"))
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

// reading looks a product up as a runner does: from the read sets in the cache, never building.
func reading(t *testing.T, inputs Inputs) (string, error) {
	t.Helper()
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
	if want := []string{"content go.mod", "exists include/missing.h", "listing source", "content source/a.c"}; !slices.Equal(kinds, want) {
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

// A build that read something no key can name is refused, named, and neither placed, recorded nor published; the
// same build with only nameable reads is settled and published, so the refusal is what stops it.
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
			t.Setenv("ADAMIC_BUILD_STORE", server.URL)
			t.Setenv("ADAMIC_BUILD_STORE_WRITE", server.URL+"/public")
			token(t)
			pending := Product(t, port, building("built"))
			build := r.built(t)["port"]
			r.window(build.Build, func() {
				r.open(r.pid, "source/a.c")
				plant.read(r)
			})
			settlement := r.settle(t)
			if plant.reason == "" {
				if len(settlement.Settled) != 1 || len(store.writes) == 0 || strings.Contains(settlement.Settled[0], "publish failed") {
					t.Fatalf("a nameable build: settled %q, refused %q, published %q", settlement.Settled, settlement.Refused, store.writes)
				}
				return
			}
			if len(settlement.Refused) != 1 || !strings.Contains(settlement.Refused[0], plant.reason) || len(settlement.Settled) != 0 {
				t.Fatalf("settled %q, refused %q; want a refusal naming %q", settlement.Settled, settlement.Refused, plant.reason)
			}
			if len(store.writes) != 0 {
				t.Fatalf("a refused build was published: %q", store.writes)
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
	if want := []string{"go.mod", "go.sum", "include/x.h", "source/a.c", "source/b.c"}; !slices.Equal(paths, want) {
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
	r.window(built["header"].Build, func() { r.open(r.pid, "include/x.h") })
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

// Away from a trace a product is never built for others: a miss is refused, named, unless ADAMIC_BUILD_STORE=off asks
// for a build for this machine alone, which a runner's read mode never finds.
// Not parallel: newRig.
func TestAnUntracedMachineBuildsOnlyForItself(t *testing.T) {
	newRig(t)
	t.Setenv("ADAMIC_BUILD_TRACE", "")
	t.Setenv("ADAMIC_BUILD_STORE", "")
	never := func(string) error { t.Fatal("an untraced machine built a product"); return nil }
	if _, err := Get(port, never); !errors.Is(err, ErrUntraced) || !strings.Contains(err.Error(), "ADAMIC_BUILD_STORE=off") {
		t.Fatalf("an untraced miss read as %v, want ErrUntraced", err)
	}
	t.Setenv("ADAMIC_BUILD_STORE", "off")
	local := Product(t, port, building("local"))
	if !strings.Contains(local, string(filepath.Separator)+"local"+string(filepath.Separator)) {
		t.Fatalf("a local build is at %s", local)
	}
	if again := Product(t, port, never); again != local {
		t.Fatalf("a local hit found %s, want %s", again, local)
	}
	if _, err := reading(t, port); !errors.Is(err, ErrNotBuilt) || !strings.Contains(err.Error(), "no traced build") {
		t.Fatalf("read mode found a local product: %v", err)
	}
}

// The recipe is the binary's linked packages: this test binary's own package by its files and listing, its module's
// go.mod and go.sum; a binary that names no package can't be keyed.
// Not parallel: go list runs in the repository.
func TestTheRecipeIsTheBinarysLinkedPackages(t *testing.T) {
	root := repositoryRootForTest(t)
	found := recipeOf(root, journalEntry{Package: "github.com/system-inc/adamic/internal/buildcache.test"})
	if found.reason != "" {
		t.Fatal(found.reason)
	}
	for _, want := range []readEntry{{Kind: "content", Path: "internal/buildcache/reads.go"}, {Kind: "content", Path: "internal/buildcache/precise_test.go"},
		{Kind: "listing", Path: "internal/buildcache"}, {Kind: "content", Path: "go.mod"}} {
		if !slices.Contains(found.entries, want) {
			t.Errorf("the recipe lacks %+v", want)
		}
	}
	for _, unknown := range []string{"", "command-line-arguments"} {
		if refused := recipeOf(root, journalEntry{Package: unknown}); refused.reason == "" {
			t.Errorf("a binary built from %q has a recipe", unknown)
		}
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
	if value, _ := entryValue(r.root, r.cache, readEntry{Kind: "link", Path: "lib"}, 0); value != "link source" {
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
	if got, want := r.entries(t, build.NameKey), []string{"content go.mod", "exists include/x.h", "content notes.txt", "content source/a.c"}; !slices.Equal(got, want) {
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
	if got, want := r.entries(t, build.NameKey), []string{"content go.mod", "content source/a.c"}; !slices.Equal(got, want) {
		t.Fatalf("the read set is %q, want %q", got, want)
	}
	r.retrace(t, "sibling")
	r.lines = nil
	other := Inputs{Name: "sibling's"}
	Product(t, other, building("built"))
	build = r.built(t)["sibling's"]
	r.add(r.pid, `clone(child_stack=NULL, flags=SIGCHLD) = %d`, sibling)
	r.write(sibling, "/tmp/adamic-test-sibling/out")
	r.window(build.Build, func() { r.open(r.pid, "/tmp/adamic-test-sibling/out") })
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
	inCheckout, err := evaluate(r.root, r.cache, nameKey, set, 0)
	if err != nil {
		t.Fatal(err)
	}
	object, err := exec.Command("git", "-C", r.root, "rev-parse", "HEAD:source/a.c").Output()
	if err != nil || inCheckout.values[0] != "file false "+strings.TrimSpace(string(object)) {
		t.Fatalf("source/a.c values %q, git names it %q (%v)", inCheckout.values[0], object, err)
	}
	unpacked := t.TempDir()
	for _, name := range []string{"source/a.c", "source/b.c", "go.mod"} {
		content, _ := os.ReadFile(filepath.Join(r.root, name))
		write(t, unpacked, name, string(content))
	}
	if inSource, err := evaluate(unpacked, r.cache, nameKey, set, 0); err != nil || inSource.key != inCheckout.key {
		t.Fatalf("a checkout keyed %s, its unpacked source %s (%v)", inCheckout.key, inSource.key, err)
	}
	// Changed after git was asked, with git's answer still held: hashed, so the key moves.
	write(t, r.root, "source/a.c", "int a2;\n")
	if after, _ := evaluate(r.root, r.cache, nameKey, set, 0); after.key == inCheckout.key {
		t.Fatal("a file changed after git was asked kept its old id")
	}
}
