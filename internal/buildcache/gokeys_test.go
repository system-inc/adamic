package buildcache

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// goKeysPackage is a package of the rig's own module that go builds with no cgo: a main, a file its build constraints
// leave out, and an embedded directory, and a package of its own it imports, which embeds nothing.
var goKeysPackage = map[string]string{
	"app/main.go":    "package main\n\nimport (\n\t_ \"embed\"\n\n\t\"github.com/system-inc/adamic/lib\"\n)\n\n//go:embed data\nvar data string\n\nfunc main() { println(data, lib.Name) }\n",
	"lib/lib.go":     "package lib\n\nconst Name = \"lib\"\n",
	"app/ignored.go": "//go:build ignore\n\npackage main\n",
	"app/data/a.txt": "a\n",
}

// goKeysRig is a rig whose tree holds goKeysPackage, committed.
func goKeysRig(t *testing.T) *rig {
	t.Helper()
	r := newRig(t)
	for name, content := range goKeysPackage {
		write(t, r.root, name, content)
	}
	gitIn(t, r.root, "add", ".")
	gitIn(t, r.root, "commit", "-q", "-m", "app")
	forgetTracked()
	return r
}

// goCommandFor is the release's go command as Settle knows it (traceFacts.goCommand).
func goCommandFor(t *testing.T) string {
	t.Helper()
	output, err := toolchain.Command("go", "env", "GOROOT").Output()
	if err != nil {
		t.Fatal(err)
	}
	return resolved(filepath.Join(strings.TrimSpace(string(output)), "bin", "go"))
}

// goBuild writes a traced go build into the rig's trace: inside build id's window, the test's process reads what owner
// names, and starts go, which reads what goes names and starts a C compiler that reads what compiler names.
func (r *rig) goBuild(t *testing.T, id string, owner, goes, compiler []string) {
	t.Helper()
	command, child, grandchild := goCommandFor(t), r.pid+1, r.pid+2
	r.window(id, func() {
		for _, name := range owner {
			r.open(r.pid, name)
		}
		r.add(r.pid, `clone(child_stack=NULL, flags=CLONE_VM|CLONE_VFORK|SIGCHLD) = %d`, child)
		r.add(child, `execve(%q, ["go", "build"], 0x7ffd5e1c /* 30 vars */) = 0`, command)
		for _, name := range goes {
			r.open(child, name)
		}
		if len(compiler) > 0 {
			r.add(child, `clone(child_stack=NULL, flags=CLONE_VM|CLONE_VFORK|SIGCHLD) = %d`, grandchild)
			r.add(grandchild, `execve("/usr/lib/gcc/x86_64-linux-gnu/13/cc1", ["cc1"], 0x7ffd5e1c /* 30 vars */) = 0`)
			for _, name := range compiler {
				r.open(grandchild, name)
			}
			r.add(grandchild, "+++ exited with 0 +++")
		}
		r.add(child, "+++ exited with 0 +++")
	})
}

// settledReport is a settlement's one settled product's report, failing the test unless exactly one settled.
func settledReport(t *testing.T, settlement Settlement) Report {
	t.Helper()
	var settled []Report
	for _, report := range settlement.Reports {
		if report.Outcome == ReportSettled {
			settled = append(settled, report)
		}
	}
	if len(settled) != 1 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	return settled[0]
}

// A product go builds from packages with no cgo is keyed by what its key declares go reads (Inputs.Go) and by every read
// of its build but go's own processes', so a cold Go cache, whose go reads sources, runs compilers and looks around, and a
// warm one, whose go reads almost nothing, settle it under one key (#mmpt615). Whether it's such a product is a fact of
// its inputs (no package of its go list -deps holds cgo), never of what ran: the cold run here even starts a C compiler
// under go. The test's own reads in the window are the product's still. Mutants that fail it: go's reads kept (notes.txt
// keys the cold run), and a build kept whole whenever a C compiler ran under go (the cold run keys include/x.h's
// compiler read and the warm doesn't).
// Not parallel: newRig.
func TestAProductGoBuiltWithNoCgoIsKeyedTheSameFromAColdGoCacheAndAWarmOne(t *testing.T) {
	r := goKeysRig(t)
	inputs, err := GoInputs("app", "./app", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"listing app", "listing app/data", "listing lib", "content lib/lib.go", "content app/main.go", "content app/ignored.go", "content app/data/a.txt", "exists vendor", goToolchainEntry} {
		if !slices.Contains(inputs.Go, want) {
			t.Fatalf("its declared go reads lack %q: %q", want, inputs.Go)
		}
	}
	settle := func(goes, compiler []string) (string, []string) {
		t.Helper()
		Product(t, inputs, building("built"))
		r.goBuild(t, r.built(t)[inputs.Name].Build, []string{"source/a.c"}, goes, compiler)
		r.add(r.pid, "+++ exited with 0 +++")
		settled := settledReport(t, r.settle(t))
		return settled.Key, r.entries(t, settled.NameKey)
	}
	cold, entries := settle([]string{"app/main.go", "notes.txt"}, []string{"include/x.h"})
	for _, want := range []string{"content source/a.c", "content app/main.go", "listing app/data", "go toolchain"} {
		if !slices.Contains(entries, want) {
			t.Fatalf("the cold run's set lacks %q: %q", want, entries)
		}
	}
	for _, unwanted := range []string{"content notes.txt", "content include/x.h"} {
		if slices.Contains(entries, unwanted) {
			t.Fatalf("the cold run's set keys go's read %q: %q", unwanted, entries)
		}
	}
	r.retrace(t, "warm")
	warmCache := t.TempDir()
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", warmCache)
	r.cache = warmCache
	forgetHeldForLife()
	if warm, _ := settle(nil, nil); warm != cold {
		t.Fatalf("a warm Go cache keyed the product %s, a cold one %s", warm, cold)
	}
}

// Each thing a declared go read stands for moves a go-built product's key, as the reads of go's own processes did: an
// embedded file's byte, a file added where a pattern embeds, an edit to a file the build constraints leave out, a file
// added to a package's directory, a vendor directory at the module's root, and a replaced module's go.mod. A lookup on
// the changed tree misses, and the tree put back finds it again. Mutants that fail it: Inputs.Go without its listings
// (the added files), its ignored files, its existence checks or its replaced go.mod, each found on a changed tree.
// Not parallel: newRig.
func TestEachDeclaredGoReadMovesAGoBuiltProductsKey(t *testing.T) {
	r := goKeysRig(t)
	write(t, r.root, "go.mod", "module github.com/system-inc/adamic\n\ngo 1.27\n\nreplace example.com/shim => ./shim\n")
	write(t, r.root, "shim/go.mod", "module example.com/shim\n\ngo 1.27\n")
	gitIn(t, r.root, "add", ".")
	gitIn(t, r.root, "commit", "-q", "-m", "shim")
	forgetTracked()
	inputs, err := GoInputs("app", "./app", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(inputs.Go, "content shim/go.mod") {
		t.Fatalf("its declared go reads lack the replaced module's go.mod: %q", inputs.Go)
	}
	Product(t, inputs, building("built"))
	r.goBuild(t, r.built(t)[inputs.Name].Build, nil, nil, nil)
	r.add(r.pid, "+++ exited with 0 +++")
	settledReport(t, r.settle(t))
	if _, err := reading(t, inputs); err != nil {
		t.Fatalf("an unchanged tree: %v", err)
	}
	for _, change := range []struct {
		name   string
		change func()
	}{
		{"an embedded file's byte", func() { write(t, r.root, "app/data/a.txt", "b\n") }},
		{"a file added where a pattern embeds", func() { write(t, r.root, "app/data/b.txt", "b\n"); gitIn(t, r.root, "add", ".") }},
		{"an edit to a file its build constraints leave out", func() { write(t, r.root, "app/ignored.go", "package main\n") }},
		{"a file added to a package's directory", func() { write(t, r.root, "lib/more.go", "package lib\n"); gitIn(t, r.root, "add", ".") }},
		{"a vendor directory at the module's root", func() { write(t, r.root, "vendor/modules.txt", "\n"); gitIn(t, r.root, "add", ".") }},
		{"a replaced module's go.mod", func() { write(t, r.root, "shim/go.mod", "module example.com/shim\n\ngo 1.26\n") }},
	} {
		change.change()
		forgetTracked()
		if _, err := reading(t, inputs); !errors.Is(err, ErrNotBuilt) {
			t.Errorf("%s left the product found: %v", change.name, err)
		}
		gitIn(t, r.root, "reset", "-q", "--hard")
		gitIn(t, r.root, "clean", "-q", "-fdx", "--", "app", "lib", "vendor", "shim")
		forgetTracked()
		if _, err := reading(t, inputs); err != nil {
			t.Fatalf("the tree put back after %s: %v", change.name, err)
		}
	}
}

// A build whose packages hold cgo keeps its whole trace, go's processes and the C they run included, and keys as it did:
// its inputs declare no go reads (Node's ruling on #mmpt615, until a C side keyed by what each compile includes, rather
// than every file of every search directory, which differs on every box). Mutant that fails it: Inputs.Go set for a cgo
// build, which drops go's read of the header.
// Not parallel: newRig.
func TestAGoBuildWithCgoKeepsItsWholeTrace(t *testing.T) {
	r := goKeysRig(t)
	enabled, err := toolchain.Command("go", "env", "CGO_ENABLED").Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(enabled)) != "1" {
		t.Skip("cgo is off under toolchain's go here, so no package lists cgo files")
	}
	write(t, r.root, "withc/main.go", "package main\n\n// #include \"../include/x.h\"\nimport \"C\"\n\nfunc main() {}\n")
	gitIn(t, r.root, "add", ".")
	gitIn(t, r.root, "commit", "-q", "-m", "withc")
	forgetTracked()
	inputs, err := GoInputs("withc", "./withc", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if inputs.Go != nil {
		t.Fatalf("a cgo build declared go reads: %q", inputs.Go)
	}
	Product(t, inputs, building("built"))
	r.goBuild(t, r.built(t)[inputs.Name].Build, nil, []string{"withc/main.go"}, []string{"include/x.h"})
	r.add(r.pid, "+++ exited with 0 +++")
	if entries := r.entries(t, settledReport(t, r.settle(t)).NameKey); !slices.Contains(entries, "content include/x.h") {
		t.Fatalf("a cgo build's set left out its compiler's read: %q", entries)
	}
}

// The Go release a go-built product names is valued by its bytes: a byte of its standard library moves it, the same
// release installed elsewhere values the same, and a box hashes it once, keeping the value under a stamp of the release's
// programs and VERSION. Mutants that fail it: src left out (the library byte), or the release's path hashed (the other
// install).
func TestTheGoToolchainIsValuedByItsBytes(t *testing.T) {
	t.Parallel()
	release := func(library string) (string, string) {
		root := t.TempDir()
		tools := filepath.Join(root, "pkg", "tool", "linux_amd64")
		write(t, root, "bin/go", "go\n")
		for _, name := range goToolchainPrograms {
			write(t, tools, name, name+"\n")
		}
		write(t, root, "VERSION", "go1.27.1\n")
		write(t, root, "src/fmt/print.go", library)
		write(t, root, "pkg/include/textflag.h", "#define NOSPLIT 4\n")
		return root, tools
	}
	cache := t.TempDir()
	one, oneTools := release("package fmt\n")
	value := goToolchainValueAt(one, oneTools, cache)
	if !strings.HasPrefix(value, "go toolchain ") {
		t.Fatalf("a release valued %q", value)
	}
	other, otherTools := release("package fmt\n")
	if elsewhere := goToolchainValueAt(other, otherTools, t.TempDir()); elsewhere != value {
		t.Fatalf("one release installed elsewhere valued %q, here %q", elsewhere, value)
	}
	changed, changedTools := release("package fmt // changed\n")
	if moved := goToolchainValueAt(changed, changedTools, t.TempDir()); moved == value {
		t.Fatal("a byte of the standard library left the release's value as it was")
	}
	if kept, err := filepath.Glob(filepath.Join(cache, "machine", "go-toolchain-*")); err != nil || len(kept) != 1 {
		t.Fatalf("the value kept beside the build cache: %q, %v", kept, err)
	}
	if again := goToolchainValueAt(one, oneTools, cache); again != value {
		t.Fatalf("the kept value read back %q, hashed %q", again, value)
	}
	if err := os.WriteFile(filepath.Join(oneTools, "compile"), []byte("another compile\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if rehashed := goToolchainValueAt(one, oneTools, cache); rehashed == value {
		t.Fatal("a program of the release changed and its kept value stood")
	}
}

// Whether a build holds cgo is read from go's listing of the build itself, its overlay included, so a main the overlay
// synthesizes is classed by what it holds: one that imports C declares no go reads and keeps its whole trace, one that
// doesn't declares them. A build go can't list has no key at all, so none is declared from a listing go never gave
// (Node's ask on #mmpt615: unknown means traced, never declared). Mutants that fail it: the listing asked without the
// overlay (go lists the package on disk, which holds no cgo), and a listing error taken as no cgo.
// Not parallel: newRig.
func TestAGoBuildIsClassedByWhatItsOverlaySynthesizes(t *testing.T) {
	r := goKeysRig(t)
	enabled, err := toolchain.Command("go", "env", "CGO_ENABLED").Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(enabled)) != "1" {
		t.Skip("cgo is off under toolchain's go here, so no package lists cgo files")
	}
	write(t, r.root, "syn/doc.go", "package main\n")
	write(t, r.root, "overlays/cgo.go", "package main\n\nimport \"C\"\n\nfunc main() {}\n")
	write(t, r.root, "overlays/plain.go", "package main\n\nfunc main() {}\n")
	for name, replacement := range map[string]string{"cgo": "overlays/cgo.go", "plain": "overlays/plain.go"} {
		write(t, r.root, "overlays/"+name+".json", `{"Replace":{"`+filepath.Join(r.root, "syn", "main.go")+`":"`+filepath.Join(r.root, replacement)+`"}}`)
	}
	gitIn(t, r.root, "add", ".")
	gitIn(t, r.root, "commit", "-q", "-m", "syn")
	forgetTracked()
	synthesized := func(name string) (Inputs, error) {
		return GoInputs("syn-"+name, "./syn", []string{"-overlay=" + filepath.Join(r.root, "overlays", name+".json")}, nil)
	}
	if inputs, err := synthesized("cgo"); err != nil || inputs.Go != nil {
		t.Fatalf("a main its overlay synthesizes with cgo declared go reads %q (%v)", inputs.Go, err)
	}
	if inputs, err := synthesized("plain"); err != nil || !slices.Contains(inputs.Go, "content overlays/plain.go") {
		t.Fatalf("a main its overlay synthesizes without cgo declared %q (%v)", inputs.Go, err)
	}
	write(t, r.root, "broken/main.go", "package main\n\nimport \"github.com/system-inc/adamic/nowhere\"\n")
	gitIn(t, r.root, "add", ".")
	gitIn(t, r.root, "commit", "-q", "-m", "broken")
	forgetTracked()
	if inputs, err := GoInputs("broken", "./broken", nil, nil); err == nil {
		t.Fatalf("a build go can't list keyed, declaring %q", inputs.Go)
	}
}
