package testpolicy

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
)

// NotAMachinePathMarker opens the comment that excuses string literals naming a home directory: at the end
// of a literal's own line, or directly above a statement or declaration, which excuses every line of it. It
// must give the reason after the colon: "// Not a machine path: the input whose home prefix is trimmed".
const NotAMachinePathMarker = "Not a machine path:"

// HomePath is one string literal in a test that names a home directory.
type HomePath struct {
	Position token.Position
	Value    string
}

var (
	homeDirectory         = regexp.MustCompile(`(^|[^A-Za-z0-9_])/(Users|home)/[^/\s"']+/`)
	reasonAfterPathMarker = regexp.MustCompile(`(?i)` + regexp.QuoteMeta(NotAMachinePathMarker) + `\s*\S`)
)

// HomePaths is every string literal in one file's source that names a home directory, /Users/<name>/ or
// /home/<name>/, and is not excused by a NotAMachinePathMarker comment with a reason.
//
// A test that reads a path on one developer's machine skips everywhere else, so its coverage exists on that
// machine alone and nobody can tell (#sycrdr6). Code outside the repository is asked for by name through
// internal/corpus, whose variables say where it is. A literal that only looks like a home path, such as the
// input to a function that trims one, says so where a reader will see it.
func HomePaths(fileName string, source []byte) ([]HomePath, error) {
	files := token.NewFileSet()
	file, err := parser.ParseFile(files, fileName, source, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	// A marker excuses its own line, for a comment at a literal's end, and the whole of the node that starts on
	// the line below it, for a comment above a statement or declaration.
	excused := map[int]bool{}
	excusedAbove := map[int]bool{}
	for _, group := range file.Comments {
		if !reasonAfterPathMarker.MatchString(group.Text()) {
			continue
		}
		excused[files.Position(group.Pos()).Line] = true
		excusedAbove[files.Position(group.End()).Line+1] = true
	}
	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil || !excusedAbove[files.Position(node.Pos()).Line] {
			return true
		}
		for line := files.Position(node.Pos()).Line; line <= files.Position(node.End()).Line; line++ {
			excused[line] = true
		}
		return true
	})
	var found []HomePath
	ast.Inspect(file, func(node ast.Node) bool {
		literal, isLiteral := node.(*ast.BasicLit)
		if !isLiteral || literal.Kind != token.STRING {
			return true
		}
		value, err := strconv.Unquote(literal.Value)
		if err != nil || !homeDirectory.MatchString(value) {
			return true
		}
		position := files.Position(literal.Pos())
		if !excused[position.Line] {
			found = append(found, HomePath{Position: position, Value: value})
		}
		return true
	})
	return found, nil
}
