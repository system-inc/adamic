// This file is not built here. gitignore_test.go lays it into cohere's own package,
// cohere/internal/gitignore, with `go test -overlay`, so that it can call cohere's matcher and cohere's
// test helpers (generateTree, buildCheckIgnoreFixture, checkIgnoreExpectations, matchLines) exactly as
// cohere's tests do, without a byte of the submodule changing.
//
// It has two jobs, chosen by the request ADAMIC_PORT_REQUEST names:
//
//   - generate: lay out every tree the port is asked about on disk and describe it, with every query, as
//     JSON: cohere's semantic tree, its refusals, generated trees from cohere's own generator, real
//     trees, and git's t0008 fixture with the answers git's script states; and every glob case: cohere's
//     documented rules and git's t3070 corpus with the answers its script states.
//   - answer: decide every query with Go cohere, and match every glob, writing the lines the port's
//     main.ts writes.

package gitignore

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// adamicRequest is what gitignore_test.go asks.
type adamicRequest struct {
	Mode      string   `json:"mode"`
	Scratch   string   `json:"scratch"`
	Seed      int64    `json:"seed"`
	Generated int      `json:"generated"`
	GitSource string   `json:"gitSource"`
	RealTrees []string `json:"realTrees"`
	Largest   bool     `json:"largest"`
	Cases     string   `json:"cases"`
	Output    string   `json:"output"`
}

// adamicCases is every tree and glob the port is asked about.
type adamicCases struct {
	Trees    []adamicTree     `json:"trees"`
	Patterns []adamicPatterns `json:"patterns"`
	Globs    []adamicGlob     `json:"globs"`
}

// adamicPatterns is a list of lines for CompilePatterns, and the paths to decide by it. git has no word
// for a list that is not a file, so these are Go cohere's and the port's alone.
type adamicPatterns struct {
	Name    string        `json:"name"`
	Lines   []string      `json:"lines"`
	Queries []adamicQuery `json:"queries"`
}

// adamicTree is one working tree: where it is on disk, what of it the matcher reads, and the paths to
// decide. Git says how git's answers are had: "run" to run git check-ignore on the tree, "script" when
// Expected holds the answers git's own test script states, and "" when git has nothing to say (a tree
// cohere refuses).
type adamicTree struct {
	Name     string        `json:"name"`
	Disk     string        `json:"disk"`
	Root     string        `json:"root"`
	Entries  []adamicEntry `json:"entries"`
	Queries  []adamicQuery `json:"queries"`
	Git      string        `json:"git"`
	Expected []string      `json:"expected,omitempty"`
}

type adamicEntry struct {
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	Contents string `json:"contents,omitempty"`
	Target   string `json:"target,omitempty"`
}

type adamicQuery struct {
	Path        string `json:"path"`
	IsDirectory bool   `json:"isDirectory"`
}

// adamicGlob is one glob case. Expected is git's answer, "0" or "1", when git's corpus states one.
type adamicGlob struct {
	Pattern  string `json:"pattern"`
	Text     string `json:"text"`
	Path     bool   `json:"path"`
	Expected string `json:"expected,omitempty"`
}

// adamicRoot is the root the port's trees show in messages; Go cohere's own, a directory on disk, is
// written as it in the answers, so the two can be compared byte for byte.
const adamicRoot = "/repository"

// Not parallel: writes the fixed output paths supplied by the environment request.
func TestAdamicPortCases(t *testing.T) {
	requestPath := os.Getenv("ADAMIC_PORT_REQUEST")
	if requestPath == "" {
		t.Skip("ADAMIC_PORT_REQUEST names no request")
	}
	contents, err := os.ReadFile(requestPath)
	if err != nil {
		t.Fatal(err)
	}
	var request adamicRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		t.Fatal(err)
	}
	switch request.Mode {
	case "generate":
		cases := adamicGenerate(t, request)
		encoded, err := json.Marshal(cases)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(request.Output, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
	case "answer":
		contents, err := os.ReadFile(request.Cases)
		if err != nil {
			t.Fatal(err)
		}
		var cases adamicCases
		if err := json.Unmarshal(contents, &cases); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(request.Output, []byte(adamicAnswer(cases)), 0o644); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatalf("no mode %q", request.Mode)
	}
}

// adamicAnswer is Go cohere's answer to every case, in the lines main.ts writes.
func adamicAnswer(cases adamicCases) string {
	var output strings.Builder
	for _, tree := range cases.Trees {
		fmt.Fprintf(&output, "tree %s\n", tree.Name)
		root, _ := filepath.Abs(tree.Disk)
		shown := func(err error) string {
			message := strings.ReplaceAll(err.Error(), root, adamicRoot)
			if strings.Contains(err.Error(), ErrNestedRepository.Error()) && isNested(err) {
				message += " (nested repository)"
			}
			return message
		}
		matcher, err := New(tree.Disk)
		if err != nil {
			fmt.Fprintf(&output, "error %s\n", shown(err))
			continue
		}
		for _, query := range tree.Queries {
			ignored, source, err := matcher.IgnoredPath(query.Path, query.IsDirectory)
			if err != nil {
				fmt.Fprintf(&output, "error %s\n", shown(err))
				continue
			}
			fmt.Fprintf(&output, "%s %s\t%s\n", adamicBit(ignored), adamicSource(source), query.Path)
		}

		// The walk: each path through its own directory's matcher, entered from the nearest one entered
		// before, as main.ts's scopeFor does.
		fmt.Fprintf(&output, "walk %s\n", tree.Name)
		scopes := map[string]*Matcher{"": matcher}
		for _, query := range tree.Queries {
			directory := path.Dir(query.Path)
			if directory == "." {
				directory = ""
			}
			ancestor := directory
			entered, found := scopes[ancestor]
			for !found {
				ancestor = path.Dir(ancestor)
				if ancestor == "." {
					ancestor = ""
				}
				entered, found = scopes[ancestor]
			}
			scope := entered
			if ancestor != directory {
				scope, err = entered.Enter(directory)
				if err != nil {
					fmt.Fprintf(&output, "error %s\n", shown(err))
					continue
				}
				scopes[directory] = scope
			}
			ignored, source := scope.Ignored(query.Path, query.IsDirectory)
			excluded, excludedBy := scope.Excluded()
			fmt.Fprintf(&output, "%s %s\t%s\t%s %s\n", adamicBit(ignored), adamicSource(source), query.Path, adamicBit(excluded), adamicSource(excludedBy))
		}
	}
	for _, list := range cases.Patterns {
		fmt.Fprintf(&output, "patterns %s\n", list.Name)
		patterns, err := CompilePatterns(list.Lines, list.Name)
		if err != nil {
			fmt.Fprintf(&output, "error %s\n", err)
			continue
		}
		for _, query := range list.Queries {
			ignored, source := patterns.Ignored(query.Path, query.IsDirectory)
			fmt.Fprintf(&output, "%s %s\t%s\n", adamicBit(ignored), adamicSource(source), query.Path)
		}
	}
	for index, glob := range cases.Globs {
		compiled := compileGlob(glob.Pattern, glob.Path)
		fmt.Fprintf(&output, "glob %d %s\n", index, adamicBit(compiled.matches(glob.Text, glob.Path)))
	}
	return output.String()
}

func isNested(err error) bool {
	for err != nil {
		if err == ErrNestedRepository {
			return true
		}
		unwrapper, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = unwrapper.Unwrap()
	}
	return false
}

func adamicBit(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func adamicSource(source Source) string {
	if source.IsZero() {
		return "::"
	}
	return source.String()
}

// adamicGenerate lays out and describes every case.
func adamicGenerate(t *testing.T, request adamicRequest) adamicCases {
	var cases adamicCases
	scratch := request.Scratch

	// cohere's TestTheMatcherFollowsGitignoreSemantics, its tree and its paths.
	semantics := filepath.Join(scratch, "semantics")
	writeTree(t, semantics, map[string]string{
		".gitignore": strings.Join([]string{
			"*.log", "!keep.log", "/anchored", "build/", "docs/generated", "\\#literal", "\\!bang", "trailing   ",
			"escaped\\ ", "**/deep/x", "vendor/", "!vendor/kept", "!kept-by-gitignore",
		}, "\n") + "\n",
		"sub/.gitignore":       "!*.log\nlocal\n/rooted\n",
		"sub/inner/.gitignore": "*.tmp\n",
		"build/":               "",
		"sub/build/":           "",
		"vendor/kept":          "",
		"docs/":                "",
		"sub/inner/":           "",
	})
	adamicGitInit(t, semantics, "# a comment\nper-repo\nkept-by-gitignore\n")
	semanticPaths := []adamicQuery{
		{"app.log", false}, {"keep.log", false}, {"sub/app.log", false}, {"anchored", false}, {"sub/anchored", false},
		{"build", true}, {"sub/build", true}, {"docs/generated", false}, {"sub/docs/generated", false},
		{"#literal", false}, {"!bang", false}, {"trailing", false}, {"escaped ", false}, {"deep/x", false},
		{"a/b/deep/x", false}, {"vendor/kept", false}, {"sub/local", false}, {"sub/inner/local", false},
		{"sub/rooted", false}, {"sub/inner/rooted", false}, {"sub/inner/a.tmp", false}, {"sub/a.tmp", false},
		{"per-repo", false}, {"sub/per-repo", false}, {"kept-by-gitignore", false},
	}
	adamicMakeQueries(t, semantics, semanticPaths)
	cases.Trees = append(cases.Trees, adamicDescribe(t, "semantics", semantics, semanticPaths, "run"))
	// The one path the cohere test asks as a file where a directory of the name exists: git is asked
	// about what is on disk, so this one is Go and the port only.
	cases.Trees = append(cases.Trees, adamicDescribe(t, "semantics, build as a file", semantics, []adamicQuery{{"build", false}}, ""))

	// cohere's TestAByteOrderMarkAndCarriageReturnsAreNotPatternText.
	byteOrderMark := filepath.Join(scratch, "byte-order-mark")
	writeTree(t, byteOrderMark, map[string]string{".gitignore": "\xEF\xBB\xBFfirst\r\nsecond\r\n", "first": "", "second": ""})
	adamicGitInit(t, byteOrderMark, "")
	cases.Trees = append(cases.Trees, adamicDescribe(t, "byte order mark", byteOrderMark, []adamicQuery{{"first", false}, {"second", false}}, "run"))

	// Bytes, not characters: `?` is one byte, so `??` matches `é`, and a set holds bytes.
	bytesTree := filepath.Join(scratch, "bytes")
	writeTree(t, bytesTree, map[string]string{".gitignore": "/??\n!/[é][é]\n*.ü\n", "é": "", "ab": "", "x": "", "名前.ü": "", "ü": ""})
	adamicGitInit(t, bytesTree, "")
	cases.Trees = append(cases.Trees, adamicDescribe(t, "bytes", bytesTree, []adamicQuery{{"é", false}, {"ab", false}, {"x", false}, {"名前.ü", false}, {"ü", false}}, "run"))

	// Only spaces end a pattern: a trailing tab is part of it, as gitignore(5) and cohere say.
	tabs := filepath.Join(scratch, "trailing-tabs")
	writeTree(t, tabs, map[string]string{".gitignore": "tabbed\t\nspaced  \nboth \t\n", "tabbed\t": "", "tabbed": "", "spaced": "", "both \t": "", "both": ""})
	adamicGitInit(t, tabs, "")
	cases.Trees = append(cases.Trees, adamicDescribe(t, "trailing tabs", tabs, []adamicQuery{{"tabbed\t", false}, {"tabbed", false}, {"spaced", false}, {"both \t", false}, {"both", false}}, "run"))

	// Paths as a caller may write them, before cleaning: IgnoredPath cleans each, as cohere does.
	unclean := filepath.Join(scratch, "unclean")
	writeTree(t, unclean, map[string]string{".gitignore": "*.log\n/rooted\nsub/*.tmp\n", "sub/.gitignore": "!keep.log\n", "sub/": ""})
	cases.Trees = append(cases.Trees, adamicDescribe(t, "unclean paths", unclean, []adamicQuery{
		{"./a.log", false}, {"sub//b.log", false}, {"sub/../c.log", false}, {"sub/./x.tmp", false}, {"./sub/keep.log", false},
		{"sub/../rooted", false}, {"sub//rooted", false}, {"./sub", true},
	}, ""))

	// An ignore file of exactly 100 MiB is read, as the Go reads one up to and including that size. It
	// makes every cases file over 100 MB, so it's asked only when the request says so.
	if request.Largest {
		largest := filepath.Join(scratch, "largest")
		contents := "big\n#" + strings.Repeat("x", 100<<20-len("big\n#")-1) + "\n"
		writeTree(t, largest, map[string]string{".gitignore": contents, "big": "", "small": ""})
		cases.Trees = append(cases.Trees, adamicDescribe(t, "an ignore file of exactly 100 MiB", largest, []adamicQuery{{"big", false}, {"small", false}}, ""))
	}

	// cohere's TestAnIgnoreFileThatCannotBeReadIsRefused: what cohere refuses, git has no word for.
	link := filepath.Join(scratch, "link")
	writeTree(t, link, map[string]string{"patterns": "x\n", "sub/": ""})
	if err := os.Symlink("../patterns", filepath.Join(link, "sub", ".gitignore")); err != nil {
		t.Fatal(err)
	}
	cases.Trees = append(cases.Trees, adamicDescribe(t, "a symbolic link in the tree", link, []adamicQuery{{"patterns", false}, {"sub/x", false}}, ""))
	nul := filepath.Join(scratch, "nul")
	writeTree(t, nul, map[string]string{".gitignore": "a\x00b\n"})
	cases.Trees = append(cases.Trees, adamicDescribe(t, "a NUL byte", nul, []adamicQuery{{"a", false}}, ""))
	notRegular := filepath.Join(scratch, "not-regular")
	writeTree(t, notRegular, map[string]string{"sub/.gitignore/": "", ".gitignore": "*.o\n"})
	cases.Trees = append(cases.Trees, adamicDescribe(t, "an ignore file that is a directory", notRegular, []adamicQuery{{"x.o", false}, {"sub/x.o", false}}, ""))

	// info/exclude lives outside the tree, so git follows a symbolic link to it, and cohere does too:
	// one to a file beside it, and one to nothing, which is no exclude file at all.
	linkedExclude := filepath.Join(scratch, "linked-exclude")
	writeTree(t, linkedExclude, map[string]string{"a.tmp": "", "b.tmp": "", "keep.tmp": ""})
	adamicGitInit(t, linkedExclude, "")
	writeTree(t, linkedExclude, map[string]string{".git/exclude-patterns": "*.tmp\n!keep.tmp\n"})
	if err := os.Remove(filepath.Join(linkedExclude, filepath.FromSlash(ExcludeFile))); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../exclude-patterns", filepath.Join(linkedExclude, filepath.FromSlash(ExcludeFile))); err != nil {
		t.Fatal(err)
	}
	cases.Trees = append(cases.Trees, adamicDescribe(t, "a linked exclude file", linkedExclude, []adamicQuery{{"a.tmp", false}, {"b.tmp", false}, {"keep.tmp", false}}, "run"))
	danglingExclude := filepath.Join(scratch, "dangling-exclude")
	writeTree(t, danglingExclude, map[string]string{"a.tmp": "", ".gitignore": "b.*\n"})
	adamicGitInit(t, danglingExclude, "")
	if err := os.Remove(filepath.Join(danglingExclude, filepath.FromSlash(ExcludeFile))); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../no-such-file", filepath.Join(danglingExclude, filepath.FromSlash(ExcludeFile))); err != nil {
		t.Fatal(err)
	}
	cases.Trees = append(cases.Trees, adamicDescribe(t, "an exclude file linked to nothing", danglingExclude, []adamicQuery{{"a.tmp", false}, {"b.tmp", false}}, "run"))

	nested := filepath.Join(scratch, "nested")
	writeTree(t, nested, map[string]string{"nested/.git/HEAD": "ref: refs/heads/main\n", ".gitignore": "*.o\n"})
	cases.Trees = append(cases.Trees, adamicDescribe(t, "a nested repository", nested, []adamicQuery{{"x.o", false}, {"nested/x.o", false}}, ""))

	// cohere's TestTheMatcherAgreesWithGitOnGeneratedTrees, with its generator and a fixed seed.
	random := rand.New(rand.NewSource(request.Seed))
	for index := 0; index < request.Generated; index++ {
		root := filepath.Join(scratch, fmt.Sprintf("generated-%d", index))
		exclude := generateTree(t, random, root)
		adamicGitInit(t, root, exclude)
		cases.Trees = append(cases.Trees, adamicDescribe(t, fmt.Sprintf("generated %d (seed %d)", index, request.Seed), root, adamicWalk(t, root), "run"))
	}

	// CompilePatterns, which cohere's tests do not ask about: each generated tree's root ignore file read
	// as a list, asked about every path of its tree, and the lines it refuses, with the quoting its
	// refusal uses.
	for _, generated := range cases.Trees {
		if !strings.HasPrefix(generated.Name, "generated ") {
			continue
		}
		for _, entry := range generated.Entries {
			if entry.Path == IgnoreFileName && entry.Kind == "File" {
				cases.Patterns = append(cases.Patterns, adamicPatterns{Name: "list of " + generated.Name, Lines: strings.Split(entry.Contents, "\n"), Queries: generated.Queries})
			}
		}
	}
	cases.Patterns = append(cases.Patterns,
		adamicPatterns{Name: "house ignore", Lines: []string{"# generated", "*.generated.ts", "!keep.generated.ts", "vendor/", "", "/rooted\r"},
			Queries: []adamicQuery{{"a/b.generated.ts", false}, {"keep.generated.ts", false}, {"vendor", true}, {"vendor", false}, {"rooted", false}, {"a/rooted", false},
				{"./a/b.generated.ts", false}, {"a//keep.generated.ts", false}, {"x/../keep.generated.ts", false}, {"vendor/", true}, {"./vendor", true}, {"a/../rooted", false}, {"a//rooted", false}}},
		adamicPatterns{Name: "a newline", Lines: []string{"fine", "not\tfine\n"}},
		adamicPatterns{Name: "a NUL byte", Lines: []string{"a\x00b"}},
		adamicPatterns{Name: "quoted", Lines: []string{"\"q\" \\ \a\b\f\r\v\x01\x1f\x7f é 🌍\n"}},
	)

	// cohere's TestTheMatcherAgreesWithGitOnRealTrees.
	for _, root := range request.RealTrees {
		cases.Trees = append(cases.Trees, adamicDescribe(t, "real "+filepath.Base(root), root, adamicWalk(t, root), "run"))
	}

	// cohere's TestGitCheckIgnoreCorpus: t0008's fixture, and the answers its script states.
	if request.GitSource != "" {
		cases.Trees = append(cases.Trees, adamicCheckIgnoreCorpus(t, request.GitSource, filepath.Join(scratch, "t0008")))
	}

	// cohere's TestGlobFollowsTheDocumentedRules, its cases with their answers.
	for _, documented := range []struct {
		pattern, text string
		path, want    bool
	}{
		{"foo", "foo", true, true}, {"foo", "fo", true, false}, {"f?o", "fxo", true, true}, {"f?o", "f/o", true, false},
		{"f?o", "f/o", false, true}, {"*.go", "main.go", true, true}, {"*.go", "a/main.go", true, false},
		{"*.go", "a/main.go", false, true}, {"a/*/c", "a/b/c", true, true}, {"a/*/c", "a/b/x/c", true, false},
		{"**/foo", "foo", true, true}, {"**/foo", "x/y/foo", true, true}, {"**/foo/bar", "x/foo/bar", true, true},
		{"abc/**", "abc/x", true, true}, {"abc/**", "abc/x/y", true, true}, {"abc/**", "abc", true, false},
		{"a/**/b", "a/b", true, true}, {"a/**/b", "a/x/b", true, true}, {"a/**/b", "a/x/y/b", true, true},
		{"a/**/b", "a/xb", true, false}, {"foo**/bar", "foo/bar", true, true}, {"foo**/bar", "foox/bar", true, true},
		{"foo**/bar", "foo/x/bar", true, false}, {"**foo", "x/foo", true, false}, {"[abc]", "b", true, true},
		{"[!abc]", "b", true, false}, {"[^abc]", "d", true, true}, {"[a-c]x", "bx", true, true}, {"[]]", "]", true, true},
		{"[a-]", "-", true, true}, {"[[:digit:]]", "7", true, true}, {"[[:digit:]]", "x", true, false},
		{"[[:nosuchclass:]]", "x", true, false}, {"[/]", "/", true, false}, {"[/]", "/", false, true},
		{"[abc", "[abc", true, false}, {`\*`, "*", true, true}, {`\*`, "x", true, false}, {`\a`, "a", true, true},
		{`foo\`, "foo", true, false}, {"*a*a*a*a*a*a*a*a*b", strings.Repeat("a", 200), true, false},
	} {
		cases.Globs = append(cases.Globs, adamicGlob{Pattern: documented.pattern, Text: documented.text, Path: documented.path, Expected: adamicBit(documented.want)})
	}

	// Every [:class:] fnmatch defines, against the bytes at each end of every class's ranges and just
	// past them, as a path and as a base name; and escaped range ends.
	classTexts := []string{"\x01", "\x08", "\t", "\n", "\v", "\r", "\x1f", " ", "!", "/", "0", "9", ":", "@", "A", "F", "G", "Z", "[", "`", "a", "f", "g", "z", "{", "~", "\x7f", "é"}
	for _, class := range []string{"alnum", "alpha", "blank", "cntrl", "digit", "graph", "lower", "print", "punct", "space", "upper", "xdigit"} {
		for _, text := range classTexts {
			for _, path := range []bool{true, false} {
				cases.Globs = append(cases.Globs, adamicGlob{Pattern: "[[:" + class + ":]]", Text: text, Path: path}, adamicGlob{Pattern: "[![:" + class + ":]]", Text: text, Path: path})
			}
		}
	}
	for _, pattern := range []string{`[a-\z]`, `[\a-c]`, `[a-\]]`, `[\]-a]`, `[!a-\z]`, `[\--\/]`} {
		for _, text := range []string{"a", "m", "z", "\\", "]", "^", "-", ".", "/", "{"} {
			cases.Globs = append(cases.Globs, adamicGlob{Pattern: pattern, Text: text, Path: false})
		}
	}

	// cohere's TestGitWildmatchCorpus: t3070's case-sensitive columns, with and without WM_PATHNAME.
	if request.GitSource != "" {
		script, err := os.ReadFile(filepath.Join(request.GitSource, "t", "t3070-wildmatch.sh"))
		if err != nil {
			t.Fatal(err)
		}
		for _, arguments := range matchLines(t, string(script)) {
			text, pattern := arguments[len(arguments)-2], arguments[len(arguments)-1]
			cases.Globs = append(cases.Globs,
				adamicGlob{Pattern: pattern, Text: text, Path: true, Expected: arguments[0]},
				adamicGlob{Pattern: pattern, Text: text, Path: false, Expected: arguments[2]})
		}
	}
	return cases
}

// adamicGitInit makes root a repository, as cohere's differential does, and writes its info/exclude
// when there is one to write. git init writes an exclude of comments of its own, which is kept, as it is
// what a clone holds.
func adamicGitInit(t *testing.T, root string, exclude string) {
	t.Helper()
	command := exec.Command("git", "init", "--quiet", root)
	command.Env = []string{"PATH=/usr/bin:/bin", "HOME=" + root, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null"}
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git init %s: %v\n%s", root, err, output)
	}
	if exclude != "" {
		writeTree(t, root, map[string]string{ExcludeFile: exclude})
	}
}

// adamicMakeQueries makes each queried path that is not there yet, a directory or an empty file, so
// git is asked about what the query says.
func adamicMakeQueries(t *testing.T, root string, queries []adamicQuery) {
	t.Helper()
	for _, query := range queries {
		full := filepath.Join(root, filepath.FromSlash(query.Path))
		if _, err := os.Lstat(full); err == nil {
			continue
		}
		if query.IsDirectory {
			writeTree(t, root, map[string]string{query.Path + "/": ""})
		} else {
			writeTree(t, root, map[string]string{query.Path: ""})
		}
	}
}

// adamicWalk is compareWithGit's walk: every path below root, not entering a nested repository or a
// directory the matcher excludes, whose own entries are still asked about.
func adamicWalk(t *testing.T, root string) []adamicQuery {
	t.Helper()
	matcher, err := New(root)
	if err != nil {
		t.Fatalf("%s: %v", root, err)
	}
	var queries []adamicQuery
	var walk func(scope *Matcher, directory string)
	walk = func(scope *Matcher, directory string) {
		entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(directory)))
		if err != nil {
			t.Fatalf("%s: %v", root, err)
		}
		for _, entry := range entries {
			name := entry.Name()
			relative := joinRelative(directory, name)
			if name == ".git" {
				continue
			}
			isDirectory := entry.Type()&fs.ModeType == fs.ModeDir
			if isDirectory {
				if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(relative), ".git")); err == nil {
					continue
				}
			}
			queries = append(queries, adamicQuery{relative, isDirectory})
			if isDirectory {
				if excluded, _ := scope.Excluded(); excluded {
					continue
				}
				child, err := scope.Enter(relative)
				if err != nil {
					t.Fatalf("%s: %v", root, err)
				}
				walk(child, relative)
			}
		}
	}
	walk(matcher, "")
	return queries
}

// adamicDescribe is one tree as the port reads it: every `.gitignore` and `.git` entry in every
// directory below root, not entering `.git` itself, and info/exclude when `.git` is a directory.
func adamicDescribe(t *testing.T, name string, root string, queries []adamicQuery, git string) adamicTree {
	t.Helper()
	tree := adamicTree{Name: name, Disk: root, Root: adamicRoot, Queries: queries, Git: git}
	add := func(relative string) {
		full := filepath.Join(root, filepath.FromSlash(relative))
		info, err := os.Lstat(full)
		if err != nil {
			return
		}
		entry := adamicEntry{Path: relative}
		switch {
		case info.Mode()&fs.ModeSymlink != 0:
			target, err := os.Readlink(full)
			if err != nil {
				t.Fatal(err)
			}
			entry.Kind, entry.Target = "SymbolicLink", target
		case info.IsDir():
			entry.Kind = "Directory"
		case info.Mode().IsRegular():
			contents, err := os.ReadFile(full)
			if err != nil {
				t.Fatal(err)
			}
			entry.Kind, entry.Contents = "File", string(contents)
		default:
			entry.Kind = "Other"
		}
		tree.Entries = append(tree.Entries, entry)
	}
	var visit func(directory string)
	visit = func(directory string) {
		add(joinRelative(directory, IgnoreFileName))
		add(joinRelative(directory, ".git"))
		entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(directory)))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.Name() == ".git" || entry.Type()&fs.ModeType != fs.ModeDir {
				continue
			}
			visit(joinRelative(directory, entry.Name()))
		}
	}
	visit("")
	if info, err := os.Stat(filepath.Join(root, ".git")); err == nil && info.IsDir() {
		add(ExcludeFile)
		// A linked exclude file is read through its link, so what it points to is part of the tree too.
		if target, err := os.Readlink(filepath.Join(root, filepath.FromSlash(ExcludeFile))); err == nil && !filepath.IsAbs(target) {
			add(path.Clean(path.Join(path.Dir(ExcludeFile), filepath.ToSlash(target))))
		}
	}
	sort.Slice(tree.Entries, func(left, right int) bool { return tree.Entries[left].Path < tree.Entries[right].Path })
	return tree
}

// adamicCheckIgnoreCorpus is TestGitCheckIgnoreCorpus's fixture and its expectations, as one tree whose
// answers are the ones t0008 states, skipping what that test skips.
func adamicCheckIgnoreCorpus(t *testing.T, source string, root string) adamicTree {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(source, "t", "t0008-ignores.sh"))
	if err != nil {
		t.Fatal(err)
	}
	script := strings.Split(string(contents), "\n")
	buildCheckIgnoreFixture(t, root, script)

	var queries []adamicQuery
	var expected []string
	seen := map[string]bool{}
	for _, expectation := range checkIgnoreExpectations(t, script) {
		switch {
		case strings.Contains(expectation.file, "global_excludes"):
			continue
		case expectation.indexMode && strings.Contains(expectation.path, "ignored-but-in-index"):
			continue
		}
		resolved := path.Clean(path.Join(expectation.directory, expectation.path))
		if resolved == "." || seen[resolved] {
			continue
		}
		seen[resolved] = true
		info, statError := os.Lstat(filepath.Join(root, filepath.FromSlash(resolved)))
		queries = append(queries, adamicQuery{resolved, statError == nil && info.IsDir()})
		if expectation.file == "" {
			expected = append(expected, fmt.Sprintf("0 ::\t%s", resolved))
			continue
		}
		ignored := !strings.HasPrefix(expectation.pattern, "!")
		expected = append(expected, fmt.Sprintf("%s %s:%s:%s\t%s", adamicBit(ignored), expectation.file, expectation.line, expectation.pattern, resolved))
	}
	tree := adamicDescribe(t, "t0008", root, queries, "script")
	tree.Expected = expected
	return tree
}
