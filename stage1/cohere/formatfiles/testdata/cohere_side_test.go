// This file is not built here. formatfiles_test.go lays it into cohere's own package,
// cohere/internal/format/formatfiles, with `go test -overlay`, so that it can build trees with cohere's
// own test helpers (writeTree and settingsTree, from enumerate_test.go) and walk them with cohere's
// Enumerate, without a byte of the submodule changing.
//
// It lays out every tree the port is asked about on disk, and writes the cases file main.ts reads and
// Go cohere's answers, in the words main.ts prints. The trees are cohere's own: every tree its tests in
// this package build with writeTree or settingsTree, read out of those tests' Go with go/ast and built
// with the same helpers, except those whose settings name ignorePatterns globs, which the port doesn't
// carry. Then generated trees, from a seeded generator whose names and contents start at every edge of
// every class the walk decides on.

package formatfiles

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/cohere/internal/format/formatoptions"
)

// adamicRequest is what formatfiles_test.go asks.
type adamicRequest struct {
	Scratch   string `json:"scratch"`
	Seed      int64  `json:"seed"`
	Generated int    `json:"generated"`
	Cases     string `json:"cases"`
	Answers   string `json:"answers"`
}

// adamicHandled are the extensions the walk is told the formatter takes, lowercased as Enumerate's
// DeclinedExtensions are.
var adamicHandled = []string{".ts", ".js", ".json", ".md", ".\u03c3"}

// adamicHandles is the handles predicate both sides use.
func adamicHandles(path string) bool {
	return slices.Contains(adamicHandled, strings.ToLower(filepath.Ext(path)))
}

// adamicTree is one tree to walk: its name and root, and the paths below it to ask about.
type adamicTree struct {
	name      string
	root      string
	paths     []string
	walkBelow string
}

// adamicCohereTrees builds every tree cohere's tests in this package build with writeTree or
// settingsTree from literals, including settings trees with lint globs.
func adamicCohereTrees(t *testing.T) ([]adamicTree, int) {
	t.Helper()
	files := token.NewFileSet()
	file, err := parser.ParseFile(files, "enumerate_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	unquote := func(expression ast.Expr) (string, bool) {
		literal, isLiteral := expression.(*ast.BasicLit)
		if !isLiteral || literal.Kind != token.STRING {
			return "", false
		}
		text, err := strconv.Unquote(literal.Value)
		return text, err == nil
	}
	contentsOf := func(expression ast.Expr) (map[string]string, bool) {
		literal, isLiteral := expression.(*ast.CompositeLit)
		if !isLiteral {
			return nil, false
		}
		contents := map[string]string{}
		for _, element := range literal.Elts {
			pair, isPair := element.(*ast.KeyValueExpr)
			if !isPair {
				return nil, false
			}
			key, keyIsString := unquote(pair.Key)
			value, valueIsString := unquote(pair.Value)
			if !keyIsString || !valueIsString {
				return nil, false
			}
			contents[key] = value
		}
		return contents, true
	}
	var trees []adamicTree
	leftOut := 0
	ast.Inspect(file, func(node ast.Node) bool {
		call, isCall := node.(*ast.CallExpr)
		if !isCall {
			return true
		}
		name, isName := call.Fun.(*ast.Ident)
		if !isName {
			return true
		}
		var root string
		var contents map[string]string
		switch {
		case name.Name == "writeTree" && len(call.Args) == 2:
			found, isFound := contentsOf(call.Args[1])
			if !isFound {
				return true
			}
			contents = found
			root = writeTree(t, contents)
		case name.Name == "settingsTree" && len(call.Args) == 4:
			house, houseIsString := unquote(call.Args[1])
			patterns, patternsAreString := unquote(call.Args[2])
			found, isFound := contentsOf(call.Args[3])
			if !houseIsString || !patternsAreString || !isFound {
				return true
			}
			contents = found
			root = settingsTree(t, house, patterns, contents)
		default:
			return true
		}
		var paths []string
		for path := range contents {
			paths = append(paths, path)
		}
		slices.Sort(paths)
		trees = append(trees, adamicTree{name: fmt.Sprintf("cohere %d", len(trees)), root: root, paths: paths})
		return true
	})

	trees = append(trees,
		adamicTree{name: "settings relative to a parent", root: settingsTree(t, `[]`, `["source/generated/**", "**/*.code.js"]`, map[string]string{
			"source/a.ts": "", "source/generated/schema.ts": "", "source/bundle.code.js": "", "source/odd.code.js/keep.ts": "",
		}), walkBelow: "source"},
		adamicTree{name: "nested repository ignores stay outside", root: settingsTree(t, `["pnpm-lock.yaml"]`, `["projects/**"]`, map[string]string{
			"projects/listed/.git/HEAD": "", "projects/listed/b.ts": "", "projects/listed/pnpm-lock.yaml": "",
		}), walkBelow: "projects/listed"},
	)
	return trees, leftOut
}

// The pieces generated trees are made of, each at an edge of a class the walk decides on.
var (
	// File names: extensions in every case, with Go's and JavaScript's lowercasing apart (U+0130, a
	// final \u03a3), letters that lowercase outside ASCII, extensions whose UTF-16 and UTF-8 orders differ
	// (U+FF21 and U+1F600), an extension of a dot alone and none, the house list's names and near
	// misses, names the ignore files below name, hidden names, and names that sort at the edges.
	adamicFileNames = []string{
		"a.ts", "B.TS", "c.Ts", "d.Js", "e.JSON", "f.md", "g.MD", "h.\u0130", "i.\u0391\u03a3", "j.\u03a3", "k.\u03c3",
		"l.\uff21", "m.\U0001f600", "n.\ue000", "o.", "noext", ".hidden", ".hidden.ts", "p.d.ts", "q.tsx", "r.TS ",
		"next-env.d.ts", "pnpm-lock.yaml", "x.sqlite", "x.SQLITE", "x.sqlite3", "x.db", "x.db-wal", "x.db-shm", "x.db-wa",
		"keep.log", "x.log", "top.ts", "built.ts", "#hash.ts", " space.ts", "\u00e9.ts", "e\u0301.ts", "Z.ts", "_.ts", "~.ts",
		".prettierignore", "package.json",
		// Names the output escapes, at the edges of what it escapes: DEL, U+001F, a tab, a newline, a
		// quote and a backslash, and U+0020 and U+007E just inside.
		"del\u007f.ts", "unit\u001f.ts", "tab\t.ts", "line\n.ts", "quote\".ts", "back\\slash.ts", "~tilde .ts",
	}
	// Directory names, the same way: the ones the ignore files prune, hidden ones, nested repositories'
	// homes, and names that sort at the edges.
	adamicDirectoryNames = []string{"src", "deep", "dist", "node_modules", "build", ".cache", "vendor", "projects", "Z", "\u00e9", "\U0001f600", "a b"}
	// Ignore file lines: plain, anchored, directory-only, negated, globbed, escaped, and lines naming the
	// pieces above in another case.
	adamicIgnoreLines = []string{
		"*.log", "!keep.log", "dist/", "/top.ts", "node_modules", "**/deep", "built.ts", "*.TS", "*.\u0130", "\\#hash.ts",
		"vendor/", "!vendor/", "build", "*.ts", "!a.ts", "\u00e9.ts", "/src/deep/", "*.sqlite", "# comment", "", "a b/",
	}
)

// adamicGenerated lays out one tree under root and returns the paths below it to ask about.
func adamicGenerated(t *testing.T, random *rand.Rand, root string) []string {
	t.Helper()
	pick := func(choices []string) string { return choices[random.Intn(len(choices))] }
	var paths []string
	// A path whose directory can't be made, since a file or a link already stands where a directory of
	// it would go (a .git file, then .git/HEAD), is skipped: the tree already has something there.
	write := func(relative string, contents string) {
		path := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return
		}
		if _, err := os.Lstat(path); err == nil {
			return
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, relative)
	}
	link := func(relative string, target string) {
		path := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return
		}
		if _, err := os.Lstat(path); err == nil {
			return
		}
		if err := os.Symlink(target, path); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, relative)
	}
	ignoreFile := func() string {
		var lines []string
		for range 1 + random.Intn(4) {
			lines = append(lines, pick(adamicIgnoreLines))
		}
		return strings.Join(lines, "\n") + "\n"
	}

	// Directories, up to three deep.
	directories := []string{""}
	for range random.Intn(5) {
		parent := directories[random.Intn(len(directories))]
		if strings.Count(parent, "/") >= 2 {
			continue
		}
		directories = append(directories, filepath.Join(parent, pick(adamicDirectoryNames)))
	}
	for _, directory := range directories {
		if directory != "" {
			if err := os.MkdirAll(filepath.Join(root, directory), 0o755); err != nil {
				t.Fatal(err)
			}
			paths = append(paths, directory)
		}
		for range random.Intn(5) {
			name := pick(adamicFileNames)
			// What formatoptions.Resolve reads at the root is left out: the port is given its answer.
			if directory == "" && name == "package.json" {
				continue
			}
			if directory == "" && name == ".prettierignore" && random.Intn(10) != 0 {
				continue
			}
			write(filepath.Join(directory, name), "x\n")
		}
		if random.Intn(3) == 0 {
			write(filepath.Join(directory, ".gitignore"), ignoreFile())
		}
		// A nested repository, in every shape: a clone's .git directory, a submodule's .git file, a
		// .git that is a link to a directory, and one that is a link to nothing.
		if directory != "" && random.Intn(5) == 0 {
			switch random.Intn(4) {
			case 0:
				write(filepath.Join(directory, ".git", "HEAD"), "ref: refs/heads/main\n")
			case 1:
				write(filepath.Join(directory, ".git"), "gitdir: ../.git/modules/x\n")
			case 2:
				link(filepath.Join(directory, ".git"), "..")
			case 3:
				link(filepath.Join(directory, ".git"), "missing")
			}
		}
		// Symbolic links: to a file, to a directory, to nothing, out of the tree, and to another link.
		if random.Intn(3) == 0 {
			targets := []string{"a.ts", ".", "..", "missing", "/", "loop", "linked"}
			link(filepath.Join(directory, pick([]string{"linked", "loop", "to.ts", "to-dir"})), pick(targets))
		}
	}
	// The root's own repository, sometimes, with its exclude file; and, rarely, since each refuses the
	// whole walk, a .gitignore that is a link or a .prettierignore that is a link to nothing.
	switch random.Intn(20) {
	case 0:
		write(".git/info/exclude", ignoreFile())
	case 1, 4, 5:
		write(".git/HEAD", "ref: refs/heads/main\n")
	case 2:
		link(".gitignore", "a.ts")
	case 3:
		// A .prettierignore that is a link to nothing, which Enumerate refuses as Lstat finds it.
		link(".prettierignore", "missing")
	}
	return paths
}

// TestAdamicPortCases lays out the trees and writes the cases file and Go cohere's answers.
func TestAdamicPortCases(t *testing.T) {
	requestPath := os.Getenv("ADAMIC_PORT_REQUEST")
	if requestPath == "" {
		t.Skip("run by stage1/cohere/formatfiles/formatfiles_test.go in the adamic repository")
	}
	contents, err := os.ReadFile(requestPath)
	if err != nil {
		t.Fatal(err)
	}
	var request adamicRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		t.Fatal(err)
	}

	// Every tree lives under the scratch directory the request names, which outlives this test, so the
	// port can walk it after: cohere's trees are copied there from the helpers' temporary directories,
	// links kept as links, and generated ones are laid out there.
	trees, leftOut := adamicCohereTrees(t)
	cohereCount := len(trees)
	if cohereCount < 8 {
		t.Fatalf("read %d trees out of cohere's tests, where there are more: the reading has stopped working", cohereCount)
	}
	for index := range trees {
		copied := filepath.Join(request.Scratch, fmt.Sprintf("cohere-%d", index))
		if output, err := exec.Command("cp", "-a", trees[index].root, copied).CombinedOutput(); err != nil {
			t.Fatalf("copying %s: %v\n%s", trees[index].root, err, output)
		}
		trees[index].root = filepath.Join(copied, trees[index].walkBelow)
	}
	repository, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, root := range []string{repository, filepath.Join(repository, "cohere"), filepath.Join(repository, "cohere/TypeScript")} {
		trees = append(trees, adamicTree{name: "real " + root, root: root})
	}
	random := rand.New(rand.NewSource(request.Seed))
	for number := range request.Generated {
		root := filepath.Join(request.Scratch, fmt.Sprintf("generated-%d", number))
		if err := os.Mkdir(root, 0o755); err != nil {
			t.Fatal(err)
		}
		paths := adamicGenerated(t, random, root)
		trees = append(trees, adamicTree{name: fmt.Sprintf("generated %d", number), root: root, paths: paths})
	}

	escape := strings.NewReplacer("\\", "\\\\", "\t", "\\t", "\n", "\\n", "\r", "\\r")
	var cases, answers strings.Builder
	record := func(fields ...string) {
		for index, field := range fields {
			if index > 0 {
				cases.WriteByte('\t')
			}
			cases.WriteString(escape.Replace(field))
		}
		cases.WriteByte('\n')
	}
	handled := append([]string{"handles"}, adamicHandled...)
	record(handled...)
	started := time.Now()
	walked := 0
	for _, tree := range trees {
		root, err := filepath.EvalSymlinks(tree.root)
		if err != nil {
			t.Fatal(err)
		}
		resolution, err := formatoptions.Resolve(root)
		if err != nil {
			t.Fatalf("%s: formatoptions.Resolve, whose answer the port is given, refuses the tree: %v", tree.name, err)
		}
		walked++
		record("tree", tree.name, root)
		record(append([]string{"patterns"}, resolution.IgnorePatterns...)...)
		fmt.Fprintf(&answers, "tree %s\n", tree.name)
		declared := "0"
		if resolution.HouseIgnoreDeclared {
			declared = "1"
		}
		record(append([]string{"house", declared, resolution.Source}, resolution.HouseIgnore...)...)

		enumeration, err := Enumerate(root, adamicHandles)
		if err != nil {
			fmt.Fprintf(&answers, "enumerate error %s\n", err.Error())
		} else {
			fmt.Fprintf(&answers, "walked %d\n", enumeration.Walked)
			for _, name := range slices.Sorted(adamicKeys(enumeration.IgnoredByLayer)) {
				fmt.Fprintf(&answers, "layer %s %d\n", adamicQuote(name), enumeration.IgnoredByLayer[name])
			}
			fmt.Fprintf(&answers, "unhandled %d\n", enumeration.Unhandled)
			fmt.Fprintf(&answers, "symbolic-links %d\n", enumeration.SymbolicLinks)
			for _, extension := range slices.Sorted(adamicKeys(enumeration.DeclinedExtensions)) {
				fmt.Fprintf(&answers, "declined %s %d\n", adamicQuote(extension), enumeration.DeclinedExtensions[extension])
			}
			for _, nested := range enumeration.NestedRepositories {
				fmt.Fprintf(&answers, "nested %s\n", adamicQuote(nested))
			}
			for _, directory := range enumeration.Directories {
				fmt.Fprintf(&answers, "directory %s\n", adamicQuote(directory))
			}
			for _, ignoreFile := range enumeration.IgnoreFiles {
				fmt.Fprintf(&answers, "ignore-file %s\n", adamicQuote(ignoreFile))
			}
			for _, file := range enumeration.Files {
				fmt.Fprintf(&answers, "file %s\n", adamicQuote(file))
			}
		}
		below, err := NestedRepositoriesBelow(root)
		if err != nil {
			fmt.Fprintf(&answers, "nested-below error %s\n", err.Error())
		} else {
			for _, nested := range below {
				fmt.Fprintf(&answers, "nested-below %s\n", adamicQuote(nested))
			}
		}

		// The questions: each path below the root and the root itself, as a directory that may hold a
		// repository, as a file that may lie in one, and as where settings may sit; and paths outside
		// the root, and a relative one.
		asked := []string{root, filepath.Dir(root), "relative/file.ts", "/"}
		for _, relative := range tree.paths {
			asked = append(asked, filepath.Join(root, relative))
		}
		for _, path := range asked {
			record("has", path)
			fmt.Fprintf(&answers, "has %s %s\n", adamicQuote(path), adamicBit(HasOwnRepository(path)))
			record("contains", path)
			fmt.Fprintf(&answers, "contains %s %s\n", adamicQuote(path), adamicQuote(NestedRepositoryContaining(root, path)))
			if strings.HasPrefix(path, root) {
				record("boundary", path, root)
				fmt.Fprintf(&answers, "boundary %s %s %s\n", adamicQuote(path), adamicQuote(root), adamicBit(repositoryBoundaryBetween(path, root)))
			}
		}
	}
	if err := os.WriteFile(request.Cases, []byte(cases.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(request.Answers, []byte(answers.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("formatter enumeration: %.6fs, %d files offered", time.Since(started).Seconds(), strings.Count(answers.String(), "\nfile "))
	t.Logf("%d trees built as cohere's tests build them (%d left out), %d generated, %d walked", cohereCount, leftOut, request.Generated, walked)
}

func adamicKeys(counts map[string]int) func(func(string) bool) {
	return func(yield func(string) bool) {
		for key := range counts {
			if !yield(key) {
				return
			}
		}
	}
}

func adamicBit(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

// adamicQuote is main.ts's quote.
func adamicQuote(text string) string {
	var builder strings.Builder
	builder.WriteByte('"')
	for _, character := range text {
		switch {
		case character == '\\':
			builder.WriteString("\\\\")
		case character == '"':
			builder.WriteString("\\\"")
		case character == '\n':
			builder.WriteString("\\n")
		case character == '\r':
			builder.WriteString("\\r")
		case character == '\t':
			builder.WriteString("\\t")
		case character < 0x20 || character == 0x7f:
			fmt.Fprintf(&builder, "\\u%04x", character)
		default:
			builder.WriteRune(character)
		}
	}
	builder.WriteByte('"')
	return builder.String()
}
