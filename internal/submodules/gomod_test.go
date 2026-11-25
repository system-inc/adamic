package submodules

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// goModHead is the module and go lines every case's go.mod opens with.
const goModHead = "module example.com/main\n\ngo 1.25\n"

// goModToday is the replace block of Adamic's go.mod before cohere's move, trimmed to the modules its readers ask
// for, the shim's parser replaced twice as main's go.mod replaces it.
const goModToday = "module github.com/system-inc/adamic\n\ngo 1.25\n\n" +
	"// Stage 0 reads the checker's proven types in process, through the shims cohere generates.\n" +
	"replace (\n" +
	"\tgithub.com/microsoft/TypeScript/tsc => ./cohere/TypeScript/tsc\n" +
	"\tgithub.com/system-inc/cohere => ./cohere\n" +
	"\tgithub.com/microsoft/TypeScript/tsc/shim/parser => ./cohere/TypeScript-shim/parser\n" +
	"\tgithub.com/microsoft/TypeScript/tsc/shim/checker => ./cohere/TypeScript-shim/checker\n" +
	"\tgithub.com/microsoft/TypeScript/tsc/shim/parser => ./cohere/TypeScript-shim/parser\n" +
	"\tgithub.com/microsoft/TypeScript/tsc/shim/vfs/osvfs => ./cohere/TypeScript-shim/vfs/osvfs\n" +
	")\n\nrequire (\n\tgithub.com/system-inc/cohere v0.0.0\n\tgithub.com/microsoft/TypeScript/tsc/shim/parser v0.0.0\n)\n"

// goModMoved is the same go.mod after the move: TypeScript under libraries/typescript and the shims under
// libraries/typescript_shim.
var goModMoved = strings.NewReplacer("./cohere/TypeScript/", "./cohere/libraries/typescript/", "./cohere/TypeScript-shim/", "./cohere/libraries/typescript_shim/").Replace(goModToday)

// GoModReplacements reads every shape of replace directive go accepts to the directives go reads, in its order, as
// internal/buildcache's goModReplacements reads them (its test holds that reader to go mod edit -json on these same
// shapes). Mutants: a block's entries ignored ("a block of several"); a comment line read as a directive ("comments
// everywhere"); an old-side version read as the replacement ("versions on either side"); a quoted => taken for the
// arrow ("an old path that is a quoted =>"); every duplicate kept where go keeps the last ("a later replacement of
// the same module wins"); CRLF's \r kept in a path ("CRLF line endings").
func TestGoModReplacementsReadEveryShapeGoAccepts(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		content string
		want    []Replacement
	}{
		{"no replace at all", goModHead + "\nrequire example.com/x v1.0.0\n", nil},
		{"one line", goModHead + "\nreplace example.com/a => ../a\n", []Replacement{{ModulePath: "example.com/a", ReplacementPath: "../a"}}},
		{"a block of several", goModHead + "\nreplace (\n\texample.com/a => ../a\n\texample.com/b => ./b\n\texample.com/c => /absolute/c\n)\n", []Replacement{
			{ModulePath: "example.com/a", ReplacementPath: "../a"},
			{ModulePath: "example.com/b", ReplacementPath: "./b"},
			{ModulePath: "example.com/c", ReplacementPath: "/absolute/c"},
		}},
		{"versions on either side", goModHead + "\nreplace example.com/a v1.0.0 => ../a\nreplace example.com/b => example.com/fork v1.2.3\nreplace example.com/c v1.2.0 => example.com/c-fork v1.3.0\n", []Replacement{
			{ModulePath: "example.com/a", ModuleVersion: "v1.0.0", ReplacementPath: "../a"},
			{ModulePath: "example.com/b", ReplacementPath: "example.com/fork", ReplacementVersion: "v1.2.3"},
			{ModulePath: "example.com/c", ModuleVersion: "v1.2.0", ReplacementPath: "example.com/c-fork", ReplacementVersion: "v1.3.0"},
		}},
		{"quoted paths", goModHead + "\nreplace \"example.com/a\" => \"../a b\"\nreplace example.com/b => \"../b\\x20c\"\n", []Replacement{
			{ModulePath: "example.com/a", ReplacementPath: "../a b"},
			{ModulePath: "example.com/b", ReplacementPath: "../b c"},
		}},
		{"an old path that is a quoted =>", goModHead + "\nreplace \"=>\" => ../arrow\n", []Replacement{{ModulePath: "=>", ReplacementPath: "../arrow"}}},
		{"comments everywhere", goModHead + "\n// replace example.com/decoy => ../decoy\nreplace example.com/a => ../a // at the end\n\nreplace (\n\t// example.com/decoy => ../decoy\n\texample.com/b => \"../b//c\" // a // inside a quoted path is the path's\n)\n", []Replacement{
			{ModulePath: "example.com/a", ReplacementPath: "../a"},
			{ModulePath: "example.com/b", ReplacementPath: "../b//c"},
		}},
		{"spread among other blocks", goModHead + "\nrequire (\n\texample.com/x v1.0.0 // indirect\n)\n\nreplace example.com/a => ../a\n\nexclude (\n\texample.com/x v0.9.0\n)\n\nretract [v1.0.0, v1.0.5]\n\nreplace (\n\texample.com/b => ../b\n)\n", []Replacement{
			{ModulePath: "example.com/a", ReplacementPath: "../a"},
			{ModulePath: "example.com/b", ReplacementPath: "../b"},
		}},
		{"CRLF line endings", strings.ReplaceAll(goModHead+"\nreplace example.com/a => ../a\n\nreplace ( // a comment\n\texample.com/b => ./b // trailing\n)\n", "\n", "\r\n"), []Replacement{
			{ModulePath: "example.com/a", ReplacementPath: "../a"},
			{ModulePath: "example.com/b", ReplacementPath: "./b"},
		}},
		{"empty blocks and replace( with no space", goModHead + "\nreplace ()\n\nreplace (\n)\n\nreplace(\n\texample.com/a => ../a\n)\n", []Replacement{{ModulePath: "example.com/a", ReplacementPath: "../a"}}},
		{"replace in other names", "module example.com/replace\n\ngo 1.25\n\nrequire replace.example.com/x v1.0.0\n\nreplace replace.example.com/x => ../replace\n", []Replacement{{ModulePath: "replace.example.com/x", ReplacementPath: "../replace"}}},
		{"a later replacement of the same module wins", goModHead + "\nreplace example.com/m => ../m1\nreplace example.com/b => ../b\nreplace example.com/m v1.0.0 => ../m-one\nreplace \"example.com/m\" => ../m2\n", []Replacement{
			{ModulePath: "example.com/b", ReplacementPath: "../b"},
			{ModulePath: "example.com/m", ModuleVersion: "v1.0.0", ReplacementPath: "../m-one"},
			{ModulePath: "example.com/m", ReplacementPath: "../m2"},
		}},
		{"no newline at the end", goModHead + "\nreplace example.com/a => ../a", []Replacement{{ModulePath: "example.com/a", ReplacementPath: "../a"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := GoModReplacements([]byte(test.content))
			if err != nil {
				t.Fatalf("GoModReplacements: %v\n%s", err, test.content)
			}
			if !slices.Equal(got, test.want) {
				t.Fatalf("GoModReplacements reads %+v, want %+v, of\n%s", got, test.want, test.content)
			}
		})
	}
}

// The shim's directory and TypeScript's are read from a go.mod on disk by their module paths, before cohere's move
// and after it, the shim's parser found once though the go.mod replaces it twice. Mutants: duplicates kept (the
// parser refused as replaced twice); the replacement read from the wrong side of => (the module path returned).
func TestReplacementDirectoryFindsTheShimInBothLayouts(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, content string
		want          map[string]string
	}{
		{"before the move", goModToday, map[string]string{
			"github.com/microsoft/TypeScript/tsc":                "./cohere/TypeScript/tsc",
			"github.com/microsoft/TypeScript/tsc/shim/parser":    "./cohere/TypeScript-shim/parser",
			"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs": "./cohere/TypeScript-shim/vfs/osvfs",
			"github.com/system-inc/cohere":                       "./cohere",
		}},
		{"after the move", goModMoved, map[string]string{
			"github.com/microsoft/TypeScript/tsc":                "./cohere/libraries/typescript/tsc",
			"github.com/microsoft/TypeScript/tsc/shim/parser":    "./cohere/libraries/typescript_shim/parser",
			"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs": "./cohere/libraries/typescript_shim/vfs/osvfs",
			"github.com/system-inc/cohere":                       "./cohere",
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			name := filepath.Join(t.TempDir(), "go.mod")
			if err := os.WriteFile(name, []byte(test.content), 0o644); err != nil {
				t.Fatal(err)
			}
			replacements, err := ReadGoModReplacements(name)
			if err != nil {
				t.Fatal(err)
			}
			for module, want := range test.want {
				got, err := ReplacementDirectory(replacements, module)
				if err != nil {
					t.Fatalf("ReplacementDirectory(%s): %v", module, err)
				}
				if got != want {
					t.Errorf("ReplacementDirectory(%s) is %q, want %q", module, got, want)
				}
			}
		})
	}
}

// A lookup by module path refuses loudly: a module no directive replaces, the refusal naming what the go.mod does
// replace; one replaced at two old versions; one replaced by another module at a version, which is no directory; and
// a go.mod whose replace go doesn't read, naming its line and its file. Mutants: a module path matched by prefix (the
// shim's parent found by its parser); the first of two versions kept (no refusal).
func TestReplacementDirectoryRefusesWhatNamesNoOneDirectory(t *testing.T) {
	t.Parallel()
	replacements, err := GoModReplacements([]byte(goModToday + "\nreplace example.com/twice => ../twice\nreplace example.com/twice v1.0.0 => ../twice-one\nreplace example.com/fork => example.com/forked v1.2.3\n"))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct{ module, says string }{
		{"github.com/microsoft/TypeScript/tsc/shim", "no replace directive replaces github.com/microsoft/TypeScript/tsc/shim; it replaces github.com/microsoft/TypeScript/tsc, github.com/system-inc/cohere, github.com/microsoft/TypeScript/tsc/shim/checker"},
		{"example.com/twice", "example.com/twice is replaced 2 times, by version"},
		{"example.com/fork", "replaced by the module example.com/forked v1.2.3, not a directory"},
	} {
		got, err := ReplacementDirectory(replacements, test.module)
		if err == nil {
			t.Fatalf("ReplacementDirectory(%s) is %q, want a refusal", test.module, got)
		}
		if !strings.Contains(err.Error(), test.says) {
			t.Errorf("ReplacementDirectory(%s) refuses with %q, which doesn't say %q", test.module, err, test.says)
		}
	}
	if _, err := ReplacementDirectory(nil, "example.com/a"); err == nil || !strings.Contains(err.Error(), "the go.mod has none") {
		t.Errorf("a lookup in a go.mod with no replace says %v, want that it has none", err)
	}
	name := filepath.Join(t.TempDir(), "go.mod")
	if err := os.WriteFile(name, []byte(goModHead+"\nreplace example.com/a => ../a extra words\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := ReadGoModReplacements(name); err == nil || !strings.Contains(err.Error(), name+": line 5: a replace go doesn't read") {
		t.Errorf("an unreadable replace reads %+v, %v, want a refusal naming the file and line 5", got, err)
	}
}
