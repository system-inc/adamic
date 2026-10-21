package buildcache

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// goModHead is the module and go lines every case's go.mod opens with, so go accepts the file.
const goModHead = "module example.com/main\n\ngo 1.25\n"

// The local replacements goModReplacements reads from a go.mod are go's own: for every shape of replace directive go
// accepts, the New paths go mod edit -json reports with no version, in its order. Loom's runners refuse go mod edit,
// so the parser answers in its place, and go itself is the oracle, run on each file. Two shapes go refuses are left
// out because no go.mod holds them: a backquoted path (go reads only "-quoted strings) and a leading byte-order mark.
// Mutants, each failing a named case: a versioned replacement counted ("versioned replacements left out"); a block's
// entries ignored ("a block of several"); a comment line read as a directive ("comments everywhere"); a quoted path's
// quotes stripped without unquoting its escapes ("quoted escapes"); a line split at its first "=>" ("an old path
// quoting =>"); a "//" inside a quoted path taken for a comment ("comments everywhere"); an unquoted path read past
// the "//" that ends it ("a comment ending an unquoted path"); a line matched by the prefix "replace" ("replace in
// other names"); "replace(" unread as a block ("a block with no space before its paren"); CRLF's "\r" kept in a path
// ("CRLF line endings"); a one-line "replace ()" read as a directive ("empty blocks"); a quoted token taken for the
// punctuation it spells, "=>" for the arrow or ")" for a block's end ("an old path that is a quoted =>", "quoted
// parens as a block's old paths"); every replacement kept where go keeps only the last of those replacing one module
// at one version, in the last one's place ("the same module replaced twice", "a later replacement of the same module
// wins", "an identical duplicate inside a block and across one").
func TestGoModReplacementsAreGosForEveryShapeOfReplace(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, content string }{
		{"no replace at all", goModHead + "\nrequire example.com/x v1.0.0\n"},
		{"one line", goModHead + "\nreplace example.com/a => ../a\n"},
		{"a block of several", goModHead + "\nreplace (\n\texample.com/a => ../a\n\texample.com/b => ./b\n\texample.com/c => /absolute/c\n)\n"},
		{"versioned replacements left out", goModHead + "\nreplace example.com/a => example.com/fork v1.2.3\n\nreplace (\n\texample.com/b => ../b\n\texample.com/c => example.com/c-fork v0.0.0-20240101000000-abcdefabcdef\n\texample.com/d => ../d\n)\n"},
		{"an old-side version", goModHead + "\nreplace example.com/a v1.0.0 => ../a\n\nreplace (\n\texample.com/b v1.2.0 => example.com/fork v1.3.0\n\texample.com/c v2.0.0+incompatible => ./c\n)\n"},
		{"quoted paths with a space", goModHead + "\nreplace \"example.com/a\" => \"../a b\"\n\nreplace (\n\t\"example.com/c\" v1.0.0 => \"./c d\"\n)\n"},
		{"quoted escapes", goModHead + "\nreplace example.com/a => \"../a\\x20b\\u00e9\"\nreplace example.com/b => \"../b\\\"q\"\n"},
		{"an old path quoting =>", goModHead + "\nreplace \"example.com/a => ../decoy\" => ../real\n"},
		{"an old path that is a quoted =>", goModHead + "\nreplace \"=>\" => ../arrow\n"},
		{"quoted parens as a block's old paths", goModHead + "\nreplace (\n\t\")\" => ../close\n\t\"(\" => ../open\n\texample.com/a => ../a\n)\n"},
		{"quoted replace and // as old paths", goModHead + "\nreplace \"replace\" => ../word\nreplace \"//\" => \"../c//d\"\n"},
		{"a quoted new path holding =>", goModHead + "\nreplace example.com/a => \"../a => b\"\n"},
		{"comments everywhere", goModHead + "\n// replace example.com/decoy => ../decoy\nreplace example.com/a => ../a // at the end of a line\n\nreplace (\n\t// on its own line, inside a block\n\t// example.com/decoy => ../decoy\n\texample.com/b => \"../b//c\" // a // inside a quoted path is the path's\n\texample.com/c => \"../c // d\"\n)\n"},
		{"a comment ending an unquoted path", goModHead + "\nreplace example.com/a => ../a//not part of the path\nreplace example.com/b => ./b// neither\n"},
		{"absolute and ./ and bare dots", goModHead + "\nreplace (\n\texample.com/a => /absolute/a\n\texample.com/b => ./b\n\texample.com/c => .\n\texample.com/d => ..\n\texample.com/e => C:/windows/e\n\texample.com/f => ../ünïcode@v1\n)\n"},
		{"spread among other blocks", goModHead + "\ntoolchain go1.25.1\n\ngodebug default=go1.21\n\nrequire (\n\texample.com/x v1.0.0 // indirect\n\texample.com/y v1.1.0\n)\n\nreplace example.com/a => ../a\n\nexclude (\n\texample.com/x v0.9.0\n)\n\nreplace (\n\texample.com/b => ../b\n)\n\nretract [v1.0.0, v1.0.5] // a range\n\nretract (\n\tv0.1.0 // a mistake\n)\n\ntool example.com/x/cmd/tool\n\nreplace example.com/c => example.com/c-fork v1.0.0\n\nignore ./node_modules\n\nreplace example.com/d => ./d\n"},
		{"CRLF line endings", strings.ReplaceAll(goModHead+"\nreplace example.com/a => ../a\n\nreplace ( // a comment\n\texample.com/b => \"../b\"\n\texample.com/c => ./c // trailing\n)\n", "\n", "\r\n")},
		{"a block whose parens carry comments", goModHead + "\nreplace ( // opens the block\n\texample.com/a => ../a\n) // closes it\n"},
		{"tabs and spaces", goModHead + "\nreplace\texample.com/a\t=>\t../a\nreplace   example.com/b   v1.0.0   =>   ../b   \n\nreplace (\n    example.com/c => ../c\n\t \texample.com/d\t => \t./d\t\n)\n"},
		{"empty blocks", goModHead + "\nreplace ()\n\nreplace (\n)\n\nreplace (\n\n\t// nothing but a comment\n\n)\n"},
		{"a block with no space before its paren", goModHead + "\nreplace(\n\texample.com/a => ../a\n)\n"},
		{"blank lines inside a block", goModHead + "\nreplace (\n\n\texample.com/a => ../a\n\n\n\texample.com/b => ../b\n\n)\n"},
		{"replace in other names", "module example.com/replace\n\ngo 1.25\n\nrequire replace.example.com/x v1.0.0\n\nrequire (\n\treplace.example.com/y v1.0.0\n)\n\nreplace replace.example.com/x => ../replace\n"},
		{"the same module replaced twice", goModHead + "\nreplace example.com/a => ../a\nreplace example.com/a v1.0.0 => ../a1\nreplace example.com/b => ../b\nreplace example.com/b => ../b\n"},
		{"a later replacement of the same module wins", goModHead + "\nreplace example.com/m => ../m1\nreplace example.com/b => ../b\nreplace \"example.com/m\" => ../m2\nreplace example.com/q => ../q\nreplace example.com/q => example.com/fork v1.0.0\nreplace example.com/r => example.com/fork v1.0.0\nreplace (\n\texample.com/r => ../r\n)\n"},
		{"an identical duplicate inside a block and across one", goModHead + "\nreplace (\n\texample.com/a => ../a\n\texample.com/b => ../b\n\texample.com/a => ../a\n)\n\nreplace example.com/b => ../b\n\nreplace (\n\texample.com/c v1.0.0 => ../c1\n\texample.com/c v1.1.0 => ../c2\n\texample.com/c => ../c\n)\n"},
		{"no newline at the end", goModHead + "\nreplace example.com/a => ../a"},
		{"no newline after a closing paren", goModHead + "\nreplace (\n\texample.com/a => ../a\n)"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			write(t, directory, "go.mod", test.content)
			want := goModEditReplacements(t, filepath.Join(directory, "go.mod"))
			got, err := goModReplacements([]byte(test.content))
			if err != nil {
				t.Fatalf("goModReplacements: %v\n%s", err, test.content)
			}
			if !slices.Equal(got, want) {
				t.Fatalf("goModReplacements says %q, go says %q, of\n%s", got, want, test.content)
			}
		})
	}
}

// Every go.mod the tree carries, its own and each checked-out submodule's (cohere's TypeScript shims replace one
// another), reads to go's local replacements: the parser meets the files the recipe reads, not only the shapes above.
// The tree's files are the ones Tracked names, so a runner's source, which has no .git, lists them from its manifest.
// Mutants: a block's entries ignored (the root go.mod replaces in a block); every replacement kept where go keeps the
// last of a module's (the root go.mod replaces the TypeScript shim's parser twice).
func TestGoModReplacementsAreGosForEveryTrackedGoMod(t *testing.T) {
	t.Parallel()
	root := repositoryRootForTest(t)
	var found []string
	err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if name != root && (entry.Name() == ".git" || entry.Name() == "node_modules" || !Tracked(root, name, true)) {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() == "go.mod" && Tracked(root, name, false) {
			found = append(found, name)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(found, filepath.Join(root, "go.mod")) {
		t.Fatalf("the tree's own go.mod isn't among the %d found: %q", len(found), found)
	}
	for _, name := range found {
		relative, _ := filepath.Rel(root, name)
		t.Run(relative, func(t *testing.T) {
			content, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			want := goModEditReplacements(t, name)
			got, err := goModReplacements(content)
			if err != nil {
				t.Fatalf("goModReplacements: %v", err)
			}
			if !slices.Equal(got, want) {
				t.Fatalf("goModReplacements says %q, go says %q", got, want)
			}
		})
	}
	t.Logf("%d go.mod files", len(found))
}

// goModEditReplacements is go's answer for the go.mod at name: each replacement's New path that names no version, in
// the order go mod edit -json lists them. A file go refuses fails the test, since no go.mod holds it.
// The tree's go.work.sum holds every checksum the workspace needs, so go never writes one beside the go.work: a product's
// key runs go list in the tree's root workspace (goRequest.keyInputs), and a unit's source is shared by every later unit
// of its tree on the box, so a sum go wrote there was one unit's write reaching the next (#2t2zgkf). It's checked as a
// runner's closed source would meet it: a copy of the root go.work naming the same modules by their absolute paths,
// beside a copy of the tree's go.work.sum, both read-only, listed by go itself over every package of the workspace and
// their tests, which must read and write nothing more. Mutant: any one line of the sums left out, red on the write.
func TestTheWorkspaceSumsAreWholeSoGoWritesNoneBesideTheWork(t *testing.T) {
	t.Parallel()
	root := repositoryRootForTest(t)
	work, sums := filepath.Join(t.TempDir(), "workspace"), ""
	if err := os.Mkdir(work, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(work, 0o755) })
	run := func(arguments ...string) []byte {
		t.Helper()
		command := toolchain.CommandIn(root, "go", arguments...)
		command.Env = append(command.Env, "GOWORK="+filepath.Join(work, "go.work"))
		output, err := command.Output()
		if err != nil {
			var exit *exec.ExitError
			if errors.As(err, &exit) {
				t.Fatalf("go %s: %v\n%s", strings.Join(arguments, " "), err, exit.Stderr)
			}
			t.Fatalf("go %s: %v", strings.Join(arguments, " "), err)
		}
		return output
	}
	for _, name := range []string{"go.work", "go.work.sum"} {
		content, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatalf("the tree's %s: %v", name, err)
		}
		if name == "go.work.sum" {
			sums = string(content)
		}
		if err := os.WriteFile(filepath.Join(work, name), content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var file struct {
		Use []struct{ DiskPath string }
	}
	if err := json.Unmarshal(run("work", "edit", "-json"), &file); err != nil {
		t.Fatal(err)
	}
	if len(file.Use) == 0 {
		t.Fatal("the root go.work uses no module")
	}
	edit := []string{"work", "edit"}
	for _, use := range file.Use {
		module := filepath.Join(root, filepath.FromSlash(use.DiskPath))
		if _, err := os.Stat(filepath.Join(module, "go.mod")); err != nil {
			t.Fatalf("the workspace's module %s isn't in this tree, as a checkout without its submodules: %v", use.DiskPath, err)
		}
		edit = append(edit, "-dropuse="+use.DiskPath, "-use="+module)
	}
	run(edit...)
	for _, name := range []string{"go.work", "go.work.sum"} {
		if err := os.Chmod(filepath.Join(work, name), 0o444); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chmod(work, 0o555); err != nil {
		t.Fatal(err)
	}
	run("list", "-deps", "-test", "-json", "all")
	after, err := os.ReadFile(filepath.Join(work, "go.work.sum"))
	if err != nil || string(after) != sums {
		t.Fatalf("go changed the workspace's sums (%v):\n%s", err, after)
	}
	if entries, err := os.ReadDir(work); err != nil || len(entries) != 2 {
		t.Fatalf("the workspace holds %d entries after go list, want its go.work and go.work.sum alone (%v)", len(entries), err)
	}
}

func goModEditReplacements(t *testing.T, name string) []string {
	t.Helper()
	output, err := toolchain.CommandIn(filepath.Dir(name), "go", "mod", "edit", "-json", name).Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			t.Fatalf("go mod edit -json refuses the go.mod: %v\n%s", err, exit.Stderr)
		}
		t.Fatalf("go mod edit -json: %v", err)
	}
	var file struct {
		Replace []struct {
			New struct{ Path, Version string }
		}
	}
	if err := json.Unmarshal(output, &file); err != nil {
		t.Fatalf("go mod edit -json: %v\n%s", err, output)
	}
	var paths []string
	for _, replacement := range file.Replace {
		if replacement.New.Version == "" {
			paths = append(paths, replacement.New.Path)
		}
	}
	return paths
}
