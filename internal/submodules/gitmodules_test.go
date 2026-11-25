package submodules

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// adamicGitmodules is Adamic's own .gitmodules as main carries it.
const adamicGitmodules = "[submodule \"cohere\"]\n\tpath = cohere\n\turl = https://github.com/system-inc/cohere.git\n\tshallow = true\n"

// cohereTodayGitmodules is cohere's .gitmodules before the move: TypeScript at the top of the tree.
const cohereTodayGitmodules = "[submodule \"TypeScript\"]\n\tpath = TypeScript\n\turl = https://github.com/system-inc/TypeScript.git\n\tbranch = adamic\n"

// cohereMovedGitmodules is cohere's .gitmodules after the move, as its libraries branch writes it: the section keeps
// its name while its path moves, and go-types joins it under libraries/go, named by its path.
const cohereMovedGitmodules = "[submodule \"TypeScript\"]\n\tpath = libraries/typescript\n\turl = https://github.com/system-inc/TypeScript.git\n\tbranch = adamic\n" +
	"[submodule \"libraries/go/go-types\"]\n\tpath = libraries/go/go-types\n\turl = https://github.com/system-inc/go-types.git\n\tbranch = go1.27.4\n"

// typeScriptGitmodules is typescript-go's own .gitmodules, the third depth: Microsoft's TypeScript for its tests.
const typeScriptGitmodules = "[submodule \"_submodules/TypeScript\"]\n\tpath = _submodules/TypeScript\n\turl = https://github.com/microsoft/TypeScript.git\n"

// writeTree writes each file of files under root, slash-separated paths to contents, making their directories. An
// entry ending in / is an empty directory: a submodule that isn't checked out.
func writeTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		target := filepath.Join(root, filepath.FromSlash(name))
		if strings.HasSuffix(name, "/") {
			if err := os.MkdirAll(target, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// Parse reads the .gitmodules grammar git writes and people edit, each case's sections by name, path and url in file
// order: quoted names and their escapes, both comment marks wherever git allows them, whitespace, case, CRLF and a
// byte-order mark, other sections and keys skipped, and a later key standing over an earlier one. Mutants: a name's
// \" left escaped ("a quoted name with escapes"); # or ; inside a quoted value taken for a comment ("comments
// everywhere git allows them"); a key compared case-sensitively ("case as git reads it"); CRLF's \r kept ("CRLF and a
// byte-order mark"); the first path kept over a later one ("a later key stands").
func TestGitmodulesParseReadsGitsGrammar(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		content string
		want    []Submodule
	}{
		{"adamic's own", adamicGitmodules, []Submodule{{Name: "cohere", Path: "cohere", URL: "https://github.com/system-inc/cohere.git"}}},
		{"cohere moved", cohereMovedGitmodules, []Submodule{
			{Name: "TypeScript", Path: "libraries/typescript", URL: "https://github.com/system-inc/TypeScript.git"},
			{Name: "libraries/go/go-types", Path: "libraries/go/go-types", URL: "https://github.com/system-inc/go-types.git"},
		}},
		{"a quoted name with escapes", "[submodule \"a \\\"quoted\\\" \\\\ na\\me\"]\n\tpath = a\n", []Submodule{{Name: "a \"quoted\" \\ name", Path: "a"}}},
		{"comments everywhere git allows them", "# a comment\n; another\n[submodule \"a\"] # after the header\n\t# inside the section\n\tpath = a # after a value\n\turl = \"https://example.com/a;b#c\" ; after a quoted value\n", []Submodule{{Name: "a", Path: "a", URL: "https://example.com/a;b#c"}}},
		{"whitespace", "  [submodule   \"a b\"]  \n path=a\n\t\turl\t =\t  x  y \t\n[submodule \"c\"]\n\tpath = \"  spaced path  \"\n", []Submodule{{Name: "a b", Path: "a", URL: "x  y"}, {Name: "c", Path: "  spaced path  "}}},
		{"case as git reads it", "[SubModule \"Mixed\"]\n\tPATH = p\n\tUrl = u\n", []Submodule{{Name: "Mixed", Path: "p", URL: "u"}}},
		{"CRLF and a byte-order mark", "\xef\xbb\xbf[submodule \"a\"]\r\n\tpath = a\r\n\turl = u\r\n", []Submodule{{Name: "a", Path: "a", URL: "u"}}},
		{"other sections and keys skipped", "[core]\n\tbare = false\n[submodule \"a\"]\n\tpath = a\n\tshallow\n\tupdate = none\n\tfetchRecurseSubmodules = false\n[alias]\n\tpath = elsewhere\n", []Submodule{{Name: "a", Path: "a"}}},
		{"a later key stands", "[submodule \"a\"]\n\tpath = first\n\turl = first\n\tpath = second\n\turl = second\n", []Submodule{{Name: "a", Path: "second", URL: "second"}}},
		{"a path cleaned", "[submodule \"a\"]\n\tpath = ./libraries//typescript/\n", []Submodule{{Name: "a", Path: "libraries/typescript"}}},
		{"value escapes", "[submodule \"a\"]\n\tpath = a\\tb\n\turl = \"q\\\"\\\\\"\n", []Submodule{{Name: "a", Path: "a\tb", URL: "q\"\\"}}},
		{"no submodules", "# nothing\n", nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := Parse([]byte(test.content))
			if err != nil {
				t.Fatalf("Parse: %v\n%s", err, test.content)
			}
			for index := range test.want {
				test.want[index].Declared = FileName
			}
			if !slices.Equal(got, test.want) {
				t.Fatalf("Parse reads %+v, want %+v, of\n%s", got, test.want, test.content)
			}
		})
	}
}

// Parse refuses what no .gitmodules holds rather than guessing at it, each case naming the words its refusal carries.
// Mutants: a path climbing out with .. kept ("a path that climbs out"); a section with no path kept ("a section with
// no path").
func TestGitmodulesParseRefusesWhatNoGitmodulesHolds(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, content, says string }{
		{"a continued value", "[submodule \"a\"]\n\tpath = a\\\n\tb\n", "continued"},
		{"the deprecated header", "[submodule.a]\n\tpath = a\n", "deprecated"},
		{"an unquoted submodule header", "[submodule]\n\tpath = a\n", "no quoted name"},
		{"an unended quoted name", "[submodule \"a]\n\tpath = a\n", "doesn't end on its line"},
		{"text after a header", "[submodule \"a\"] path = a\n", "isn't a comment"},
		{"an absolute path", "[submodule \"a\"]\n\tpath = /absolute/a\n", "absolute"},
		{"a path that climbs out", "[submodule \"a\"]\n\tpath = libraries/../../a\n", "climbs out"},
		{"the repository itself", "[submodule \"a\"]\n\tpath = ./\n", "the repository itself"},
		{"a path with no value", "[submodule \"a\"]\n\tpath\n", "names no value"},
		{"a section with no path", "[submodule \"a\"]\n\turl = u\n", "names no path"},
		{"an unended quoted value", "[submodule \"a\"]\n\tpath = \"a\n", "doesn't end on its line"},
		{"an escape git refuses", "[submodule \"a\"]\n\tpath = a\\qb\n", "escape git refuses"},
		{"a key before any section", "path = a\n", "before any section"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := Parse([]byte(test.content))
			if err == nil {
				t.Fatalf("Parse reads %+v, want a refusal saying %q, of\n%s", got, test.says, test.content)
			}
			if !strings.Contains(err.Error(), test.says) {
				t.Fatalf("Parse refuses with %q, want it to say %q", err, test.says)
			}
		})
	}
}

// Before the move, cohere's TypeScript is found by its name at cohere/TypeScript, through Adamic's .gitmodules and
// then cohere's, in a tree with no .git, as a Loom runner's source is. TypeScript's own directory is empty, as a
// submodule not checked out is, and holds no .gitmodules. Mutant: Find matching a submodule's url where it should
// match its name, red here since no url is "TypeScript".
func TestSubmodulePathFindsTypeScriptBeforeTheMove(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		".gitmodules":        adamicGitmodules,
		"cohere/.gitmodules": cohereTodayGitmodules,
		"cohere/TypeScript/": "",
	})
	if _, err := os.Stat(filepath.Join(root, ".git")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("the tree holds a .git (%v), so it proves nothing about a runner's source", err)
	}
	for name, want := range map[string]string{"cohere": "cohere", "TypeScript": "cohere/TypeScript"} {
		got, err := Path(root, name)
		if err != nil {
			t.Fatalf("Path(%q): %v", name, err)
		}
		if got != want {
			t.Errorf("Path(%q) is %q, want %q", name, got, want)
		}
	}
}

// After the move, the same name finds TypeScript at cohere/libraries/typescript, go-types beside it under
// libraries/go is found by its own name, and typescript-go's own submodule, a third depth, by its name, each path
// joined through every enclosing submodule's. Read reports each one's depth and the file that declares it. Mutants:
// Find matching a submodule's url where it should match its name, red on every lookup; a nested path left unjoined
// (TypeScript at libraries/typescript); a submodule's own .gitmodules unread (go-types absent).
func TestSubmodulePathFindsTypeScriptAfterTheMove(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		".gitmodules":        adamicGitmodules,
		"cohere/.gitmodules": cohereMovedGitmodules,
		"cohere/libraries/typescript/.gitmodules": typeScriptGitmodules,
		"cohere/libraries/go/go-types/":           "",
	})
	for name, want := range map[string]string{
		"TypeScript":             "cohere/libraries/typescript",
		"libraries/go/go-types":  "cohere/libraries/go/go-types",
		"_submodules/TypeScript": "cohere/libraries/typescript/_submodules/TypeScript",
	} {
		got, err := Path(root, name)
		if err != nil {
			t.Fatalf("Path(%q): %v", name, err)
		}
		if got != want {
			t.Errorf("Path(%q) is %q, want %q", name, got, want)
		}
	}
	submodules, err := Read(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []Submodule{
		{Name: "cohere", Path: "cohere", URL: "https://github.com/system-inc/cohere.git", Depth: 0, Declared: ".gitmodules"},
		{Name: "TypeScript", Path: "cohere/libraries/typescript", URL: "https://github.com/system-inc/TypeScript.git", Depth: 1, Declared: "cohere/.gitmodules"},
		{Name: "libraries/go/go-types", Path: "cohere/libraries/go/go-types", URL: "https://github.com/system-inc/go-types.git", Depth: 1, Declared: "cohere/.gitmodules"},
		{Name: "_submodules/TypeScript", Path: "cohere/libraries/typescript/_submodules/TypeScript", URL: "https://github.com/microsoft/TypeScript.git", Depth: 2, Declared: "cohere/libraries/typescript/.gitmodules"},
	}
	if !slices.Equal(submodules, want) {
		t.Fatalf("Read reads\n%+v\nwant\n%+v", submodules, want)
	}
}

// A name no .gitmodules declares is refused, and the refusal names every submodule found, with its path and the file
// that declares it, so the reader sees what moved. A name differing only in case is absent, since git compares
// subsection names case-sensitively. A root with no .gitmodules is refused as a missing file, never read as a tree
// with no submodules. Mutants: a name compared case-insensitively (typescript found); a missing root .gitmodules read
// as empty (no error).
func TestSubmodulePathRefusesAnAbsentName(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		".gitmodules":        adamicGitmodules,
		"cohere/.gitmodules": cohereTodayGitmodules,
	})
	for _, name := range []string{"typescript", "TypeScript-shim", "go-types"} {
		got, err := Path(root, name)
		if err == nil {
			t.Fatalf("Path(%q) is %q, want a refusal", name, got)
		}
		for _, says := range []string{`"` + name + `"`, `"cohere" at cohere (.gitmodules)`, `"TypeScript" at cohere/TypeScript (cohere/.gitmodules)`} {
			if !strings.Contains(err.Error(), says) {
				t.Errorf("Path(%q) refuses with %q, which doesn't say %s", name, err, says)
			}
		}
	}
	if _, err := Path(t.TempDir(), "TypeScript"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Path in a tree with no .gitmodules says %v, want a missing file", err)
	}
}

// A name twice at one depth is refused, naming each place it was found: twice in one file, and once in each of two
// submodules' files at one depth. Mutant: the first of two kept (no refusal in one file).
func TestSubmodulePathRefusesANameTwiceAtOneDepth(t *testing.T) {
	t.Parallel()
	oneFile := t.TempDir()
	writeTree(t, oneFile, map[string]string{
		".gitmodules":        adamicGitmodules,
		"cohere/.gitmodules": cohereTodayGitmodules + "[submodule \"TypeScript\"]\n\tpath = libraries/typescript\n\turl = https://github.com/system-inc/TypeScript.git\n",
	})
	twoFiles := t.TempDir()
	writeTree(t, twoFiles, map[string]string{
		".gitmodules":          adamicGitmodules + "[submodule \"other\"]\n\tpath = other\n\turl = https://example.com/other.git\n",
		"cohere/.gitmodules":   "[submodule \"shared\"]\n\tpath = shared\n\turl = https://example.com/shared.git\n",
		"other/.gitmodules":    "[submodule \"shared\"]\n\tpath = vendor/shared\n\turl = https://example.com/shared.git\n",
		"cohere/shared/":       "",
		"other/vendor/shared/": "",
	})
	for _, test := range []struct {
		root, name string
		says       []string
	}{
		{oneFile, "TypeScript", []string{`"TypeScript" appears 2 times`, "at cohere/TypeScript (cohere/.gitmodules, depth 1)", "at cohere/libraries/typescript (cohere/.gitmodules, depth 1)"}},
		{twoFiles, "shared", []string{`"shared" appears 2 times`, "at cohere/shared (cohere/.gitmodules, depth 1)", "at other/vendor/shared (other/.gitmodules, depth 1)"}},
	} {
		got, err := Path(test.root, test.name)
		if err == nil {
			t.Fatalf("Path(%q) is %q, want a refusal", test.name, got)
		}
		for _, says := range test.says {
			if !strings.Contains(err.Error(), says) {
				t.Errorf("Path(%q) refuses with %q, which doesn't say %s", test.name, err, says)
			}
		}
	}
}

// A name at two depths is refused too, since answering with either would be a guess: TypeScript in cohere's
// .gitmodules at depth 1 and again in TypeScript's own at depth 2. The refusal names both places, each with the file
// that declares it and its depth. Mutant: the shallowest depth answering (cohere/TypeScript returned, no refusal).
func TestSubmodulePathRefusesANameAtTwoDepths(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		".gitmodules":                   adamicGitmodules,
		"cohere/.gitmodules":            cohereTodayGitmodules,
		"cohere/TypeScript/.gitmodules": "[submodule \"TypeScript\"]\n\tpath = _submodules/TypeScript\n\turl = https://github.com/microsoft/TypeScript.git\n",
	})
	got, err := Path(root, "TypeScript")
	if err == nil {
		t.Fatalf("a name at two depths answers %q, want a refusal", got)
	}
	for _, says := range []string{`"TypeScript" appears 2 times`, "at cohere/TypeScript (cohere/.gitmodules, depth 1)", "at cohere/TypeScript/_submodules/TypeScript (cohere/TypeScript/.gitmodules, depth 2)"} {
		if !strings.Contains(err.Error(), says) {
			t.Errorf("Path(%q) refuses with %q, which doesn't say %s", "TypeScript", err, says)
		}
	}
}

// A nested .gitmodules is read with the same refusals as the root's, and its refusal names the file: a submodule's
// path that climbs out of its own repository would name a place outside it once joined. Mutant: the nested file's
// path joined unchecked (a path outside cohere returned).
func TestSubmodulesReadRefusesANestedPathThatClimbsOut(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		".gitmodules":        adamicGitmodules,
		"cohere/.gitmodules": "[submodule \"TypeScript\"]\n\tpath = ../TypeScript\n\turl = https://github.com/system-inc/TypeScript.git\n",
	})
	got, err := Read(root)
	if err == nil {
		t.Fatalf("Read reads %+v, want a refusal", got)
	}
	for _, says := range []string{filepath.Join(root, "cohere", ".gitmodules"), "climbs out"} {
		if !strings.Contains(err.Error(), says) {
			t.Errorf("Read refuses with %q, which doesn't say %s", err, says)
		}
	}
}
