package text_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// goWhitespaceMarker opens the comment that keeps a Go-whitespace call: the call's own line ends with
// it, or the line above the call holds it, and the rest of the comment says why Go's set is the right
// one there, for instance a path or a key from cohere's own settings, which JavaScript never trims.
const goWhitespaceMarker = "Go whitespace:"

// goWhitespaceFunctions are the standard library's whitespace functions, by import path. Each uses
// unicode.IsSpace's set, which is not JavaScript's: see text.IsWhitespace.
var goWhitespaceFunctions = map[string][]string{
	"strings": {"TrimSpace", "Fields"},
	"bytes":   {"TrimSpace", "Fields"},
	"unicode": {"IsSpace"},
}

// goWhitespaceUnswept pins how many unmarked Go-whitespace calls a file still holds, at its exact count,
// so nothing new joins them. A file leaves this list when its count reaches zero, and the count only goes
// down: a new call in a pinned file fails just as it would anywhere else.
//
// lint's 139 sites were swept under #z4nssqs. Each moved to text.TrimWhitespace, text.WhitespaceFields
// or text.IsWhitespace, because it reads JavaScript or answers as a JavaScript tool would, or kept Go's
// set with a marker saying why. These remain.
var goWhitespaceUnswept = map[string]int{
	// lint: a gap that is only trivia, or what TypeScript's JSDoc parser kept, which is the scanner's set
	// (it also skips U+0085 and U+200B), neither Go's nor JavaScript's (#0pbc8mv).
	"internal/lint/ecmascript/comments/comments.go":       1,
	"internal/lint/ecmascript/dotnotation/dotnotation.go": 1,
	"internal/lint/rules/typescript/no_deprecated.go":     1,
	"internal/lint/rules/typescript/no_invalid_this.go":   1,
}

// goWhitespaceSite is one unmarked reference to a Go whitespace function.
type goWhitespaceSite struct {
	file     string
	line     int
	function string
}

// goWhitespaceSitesIn finds every unmarked reference to a Go whitespace function in a file, called or
// passed as a value (`strings.TrimFunc(value, unicode.IsSpace)`), through whatever name the file imports
// the package under.
func goWhitespaceSitesIn(fileSet *token.FileSet, file *ast.File, relative string) []goWhitespaceSite {
	localNames := map[string]string{}
	for _, specification := range file.Imports {
		path, err := strconv.Unquote(specification.Path.Value)
		if err != nil {
			continue
		}
		if _, watched := goWhitespaceFunctions[path]; !watched {
			continue
		}
		local := path
		if specification.Name != nil {
			local = specification.Name.Name
		}
		localNames[local] = path
	}
	if len(localNames) == 0 {
		return nil
	}

	markedLines := map[int]bool{}
	for _, group := range file.Comments {
		for _, comment := range group.List {
			body := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(comment.Text, "//"), "/*")) // Go whitespace: a Go comment's text
			if strings.HasPrefix(body, goWhitespaceMarker) {
				// The comment's own line keeps a trailing call, and the line below its group keeps the call
				// there.
				markedLines[fileSet.Position(comment.Pos()).Line] = true
				markedLines[fileSet.Position(group.End()).Line+1] = true
			}
		}
	}

	var found []goWhitespaceSite
	ast.Inspect(file, func(node ast.Node) bool {
		selector, isSelector := node.(*ast.SelectorExpr)
		if !isSelector {
			return true
		}
		identifier, isIdentifier := selector.X.(*ast.Ident)
		if !isIdentifier || identifier.Obj != nil {
			// A local variable named `strings` is not the package.
			return true
		}
		path, isWatched := localNames[identifier.Name]
		if !isWatched || !slices.Contains(goWhitespaceFunctions[path], selector.Sel.Name) {
			return true
		}
		line := fileSet.Position(selector.Pos()).Line
		if !markedLines[line] {
			found = append(found, goWhitespaceSite{file: relative, line: line, function: path + "." + selector.Sel.Name})
		}
		return true
	})
	return found
}

// goWhitespaceSites walks the ported code, lint and format, and finds every unmarked site. Test files
// are left out: a test trimming its own expected output never reaches a verdict a user sees.
func goWhitespaceSites(t *testing.T) []goWhitespaceSite {
	t.Helper()
	moduleRoot, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	var found []goWhitespaceSite
	files := 0
	for _, directory := range []string{"internal/lint", "internal/format"} {
		walkError := filepath.WalkDir(filepath.Join(moduleRoot, directory), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			fileSet := token.NewFileSet()
			file, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
			if err != nil {
				return err
			}
			files++
			relative, err := filepath.Rel(moduleRoot, path)
			if err != nil {
				return err
			}
			found = append(found, goWhitespaceSitesIn(fileSet, file, filepath.ToSlash(relative))...)
			return nil
		})
		if walkError != nil {
			t.Fatalf("walking %s: %v", directory, walkError)
		}
	}
	// A walk that read nothing would pass everything, so it has to have read the tree.
	if files < 1000 {
		t.Fatalf("the walk read %d files under internal/lint and internal/format, which is not the tree", files)
	}
	sort.Slice(found, func(left, right int) bool {
		return found[left].file < found[right].file || found[left].file == found[right].file && found[left].line < found[right].line
	})
	return found
}

// Ported code reads whitespace as JavaScript does (#z4nssqs). A Go whitespace call in lint or format
// fails here by file and line unless it carries a `// Go whitespace:` comment saying why Go's set is
// right there, or its file is pinned in goWhitespaceUnswept at its exact count. A pin above the count
// fails too, so the list can only shrink.
func TestPortedCodeReadsWhitespaceAsJavaScriptDoes(t *testing.T) {
	t.Parallel()
	byFile := map[string][]goWhitespaceSite{}
	for _, site := range goWhitespaceSites(t) {
		byFile[site.file] = append(byFile[site.file], site)
	}
	for file, sites := range byFile {
		pinned := goWhitespaceUnswept[file]
		if len(sites) == pinned {
			continue
		}
		if len(sites) < pinned {
			t.Errorf("%s holds %d unmarked Go-whitespace calls, below its pin of %d: lower the pin in goWhitespaceUnswept", file, len(sites), pinned)
			continue
		}
		for _, site := range sites {
			t.Errorf("%s:%d: %s uses Go's whitespace, which is not JavaScript's: use text.TrimWhitespace, text.WhitespaceFields or text.IsWhitespace, or say why Go's set is right with a `// %s` comment",
				site.file, site.line, site.function, goWhitespaceMarker)
		}
	}
	for file, pinned := range goWhitespaceUnswept {
		if len(byFile[file]) == 0 {
			t.Errorf("goWhitespaceUnswept pins %s at %d, and it holds none: remove it", file, pinned)
		}
	}
}

// The guard sees each kind of site it exists for, through an aliased import and as a value, and keeps
// each marked one: the control that the tree-wide pass above is a walk that could have failed.
func TestTheWhitespaceGuardSeesEachKindOfSite(t *testing.T) {
	t.Parallel()
	source := `package probe

import (
	"bytes"
	"strings"
	text "unicode"
)

func probe(value string, raw []byte) {
	_ = strings.TrimSpace(value)
	_ = strings.Fields(value)
	_ = bytes.TrimSpace(raw)
	_ = strings.TrimFunc(value, text.IsSpace)
	_ = strings.TrimSpace(value) // Go whitespace: a settings key
	// Go whitespace: a path from the command line, which
	// no JavaScript tool reads.
	_ = strings.Fields(value)
	_ = strings.TrimLeft(value, " ")
	strings := []string{}
	_ = strings
}
`
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "probe.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, site := range goWhitespaceSitesIn(fileSet, file, "probe.go") {
		lines = append(lines, fmt.Sprintf("%d %s", site.line, site.function))
	}
	want := []string{"10 strings.TrimSpace", "11 strings.Fields", "12 bytes.TrimSpace", "13 unicode.IsSpace"}
	if !slices.Equal(lines, want) {
		t.Fatalf("the guard found %v, want %v", lines, want)
	}
}
