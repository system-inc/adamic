package sourcename_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

// tsDecisionsAllowed are the `.ts` decisions that are right to leave alone, by file (relative to the
// module) and the literal they name, each with why. A pinned house file or a framework's own file
// convention names one file or one tool's rule, and neither can be an Adamic `.a`. An entry marked
// pending is a decision its owner converts under that task; it leaves this list when it does.
var tsDecisionsAllowed = map[string]string{
	// A table keyed by extension, which its one reader looks up with the name sourcename.TreatedAs gives.
	"internal/edit/write.go: .ts":              "TypeScriptParsable reads its list with sourcename.TreatedAs(fileName)",
	"internal/format/prettier/handles.go: .ts": "parserFor reads its table with sourcename.TreatedAs(fileName)",

	// A tool's disk walk with no program to ask, so it cannot tell Adamic source from a static library.
	"command/formatter_comparison/main.go: .ts":                        "the comparison tool's corpus walk, which has no program",
	"internal/lint/rules/tailwind/tools/generate_variant/main.go: .ts": "the variant generator's walk of a Tailwind consumer, which has no program",

	// One file, by its house name: a pinned file is that file, never an Adamic one.
	"internal/lint/rules/base/consistency_no_hand_built_declared_error.go: /BaseError.ts":        "Base's own BaseError.ts",
	"internal/lint/rules/base/consistency_no_hand_built_declared_error.go: /CreateBaseErrors.ts": "Base's own CreateBaseErrors.ts",

	// A framework's own file convention: Next reads route.ts and its pageExtensions, which can't load .a.
	"internal/lint/rules/structure/file_context.go: /route.ts": "Next's route handler file convention",
	"internal/lint/ecmascript/nextjs/contract.go: .ts":         "Next's default pageExtensions",

	// Not a file's extension: what markdown's fenced code names.
	"internal/format/native/native.go: .ts":                "markdown fence parsers, keyed by the fence's language",
	"internal/format/markdown/languages_generated.go: .ts": "linguist's language table, generated, for markdown fences",
}

// tsDecision is one place cohere decides something on a name's `.ts` ending without asking
// sourcename.TreatedAs.
type tsDecision struct {
	file    string
	line    int
	literal string
	how     string
}

func (decision tsDecision) key() string { return decision.file + ": " + decision.literal }

// endsLikeTypeScript reports whether a literal is a `.ts` ending a decision could hinge on: `.ts`
// itself or a name or suffix ending in it, but not a declaration-file ending, since an Adamic file is
// never a declaration.
func endsLikeTypeScript(literal string) bool {
	if !strings.HasSuffix(literal, ".ts") {
		return false
	}
	for _, declaration := range []string{".d.ts", ".d.cts", ".d.mts"} {
		if strings.HasSuffix(literal, declaration) {
			return false
		}
	}
	return true
}

// stringLiteral is a basic string literal's value.
func stringLiteral(expression ast.Expr) (string, bool) {
	literal, isLiteral := expression.(*ast.BasicLit)
	if !isLiteral || literal.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(literal.Value)
	return value, err == nil
}

// callsSelector reports whether an expression is a call of package.function, by the names written.
func callsSelector(expression ast.Expr, packageName string, functions ...string) (*ast.CallExpr, bool) {
	call, isCall := expression.(*ast.CallExpr)
	if !isCall {
		return nil, false
	}
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector {
		return nil, false
	}
	identifier, isIdentifier := selector.X.(*ast.Ident)
	return call, isIdentifier && identifier.Name == packageName && slices.Contains(functions, selector.Sel.Name)
}

// treated reports whether an expression reads a name through sourcename.TreatedAs anywhere inside it
// (filepath.Ext(sourcename.TreatedAs(name)) included), or names the Adamic extension.
func treated(expression ast.Expr) bool {
	found := false
	ast.Inspect(expression, func(node ast.Node) bool {
		if call, isCall := node.(ast.Expr); isCall {
			if _, isTreated := callsSelector(call, "sourcename", "TreatedAs"); isTreated {
				found = true
			}
		}
		return !found
	})
	if found {
		return true
	}
	if selector, isSelector := expression.(*ast.SelectorExpr); isSelector && selector.Sel.Name == "AdamicExtension" {
		return true
	}
	value, isLiteral := stringLiteral(expression)
	return isLiteral && value == ".a"
}

// tsDecisionsIn finds every `.ts` decision in a file that does not ask sourcename.TreatedAs.
func tsDecisionsIn(fileSet *token.FileSet, file *ast.File, relative string) []tsDecision {
	var found []tsDecision
	add := func(node ast.Node, literal string, how string) {
		found = append(found, tsDecision{file: relative, line: fileSet.Position(node.Pos()).Line, literal: literal, how: how})
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.CallExpr:
			if call, isSuffixCall := callsSelector(typed, "strings", "HasSuffix", "TrimSuffix", "CutSuffix"); isSuffixCall && len(call.Args) == 2 {
				if literal, isLiteral := stringLiteral(call.Args[1]); isLiteral && endsLikeTypeScript(literal) && !treated(call.Args[0]) {
					add(call, literal, "a suffix check on a name that is not sourcename.TreatedAs")
				}
			}
			if call, isKind := callsSelector(typed, "core", "GetScriptKindFromFileName"); isKind && len(call.Args) == 1 && !treated(call.Args[0]) {
				add(call, "GetScriptKindFromFileName", "a script kind read from a name that is not sourcename.TreatedAs, which is Unknown for `.a`")
			}
			if call, isRegexp := callsSelector(typed, "regexp", "MustCompile", "Compile"); isRegexp && len(call.Args) == 1 {
				if literal, isLiteral := stringLiteral(call.Args[0]); isLiteral && strings.Contains(literal, `\.ts`) {
					add(call, literal, "a regexp on a `.ts` ending; check sourcename.TreatedAs(name) with strings.HasSuffix instead")
				}
			}
		case *ast.BinaryExpr:
			if typed.Op != token.EQL && typed.Op != token.NEQ {
				return true
			}
			for _, pair := range [][2]ast.Expr{{typed.X, typed.Y}, {typed.Y, typed.X}} {
				if literal, isLiteral := stringLiteral(pair[0]); isLiteral && literal == ".ts" && !treated(pair[1]) {
					add(typed, literal, "an extension compared with `.ts` that is not sourcename.TreatedAs's")
				}
			}
		case *ast.SwitchStmt:
			if typed.Tag == nil || treated(typed.Tag) {
				return true
			}
			for _, statement := range typed.Body.List {
				clause, isClause := statement.(*ast.CaseClause)
				if !isClause {
					continue
				}
				for _, expression := range clause.List {
					if literal, isLiteral := stringLiteral(expression); isLiteral && literal == ".ts" {
						add(clause, literal, "a switch case on `.ts` over a value that is not sourcename.TreatedAs's")
					}
				}
			}
		case *ast.CompositeLit:
			namesTypeScript, namesAdamic := false, false
			for _, element := range typed.Elts {
				value := element
				if keyValue, isKeyValue := element.(*ast.KeyValueExpr); isKeyValue {
					value = keyValue.Key
				}
				if literal, isLiteral := stringLiteral(value); isLiteral && literal == ".ts" {
					namesTypeScript = true
				}
				if treated(value) {
					namesAdamic = true
				}
			}
			if namesTypeScript && !namesAdamic {
				add(typed, ".ts", "an extension list or map that names `.ts` and not `.a`")
			}
		}
		return true
	})
	return found
}

// tsDecisions loads every package of the module and finds every `.ts` decision outside this package.
func tsDecisions(t *testing.T) []tsDecision {
	t.Helper()
	moduleRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	config := &packages.Config{Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedSyntax, Dir: moduleRoot}
	loaded, err := packages.Load(config, "./...")
	if err != nil {
		t.Fatalf("loading the module: %v", err)
	}
	if count := packages.PrintErrors(loaded); count > 0 {
		t.Fatalf("%d errors loading the module, so the walk would read a partial tree", count)
	}
	var found []tsDecision
	for _, loadedPackage := range loaded {
		if strings.HasSuffix(loadedPackage.PkgPath, "/internal/types/sourcename") {
			continue
		}
		for index, file := range loadedPackage.Syntax {
			relative, err := filepath.Rel(moduleRoot, loadedPackage.CompiledGoFiles[index])
			if err != nil {
				t.Fatal(err)
			}
			found = append(found, tsDecisionsIn(loadedPackage.Fset, file, filepath.ToSlash(relative))...)
		}
	}
	sort.Slice(found, func(left, right int) bool {
		return found[left].file < found[right].file || found[left].file == found[right].file && found[left].line < found[right].line
	})
	return found
}

// Every decision cohere makes on a name's `.ts` ending asks sourcename.TreatedAs, so an Adamic `.a` file
// is decided as the same file named `.ts` would be (#6mhafvb). One that doesn't fails here by file and
// line, unless it is on the allowed list with its reason. And every allowed entry still exists, so the
// list can't keep a reason for a decision that has gone.
func TestEveryTypeScriptExtensionDecisionTreatsAdamicAsTypeScript(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for _, decision := range tsDecisions(t) {
		seen[decision.key()] = true
		if _, isAllowed := tsDecisionsAllowed[decision.key()]; !isAllowed {
			t.Errorf("%s:%d: %s (%q): ask sourcename.TreatedAs, or add %q to tsDecisionsAllowed with why it can't be an Adamic file",
				decision.file, decision.line, decision.how, decision.literal, decision.key())
		}
	}
	for key := range tsDecisionsAllowed {
		if !seen[key] {
			t.Errorf("tsDecisionsAllowed names %q, which no longer exists: remove it", key)
		}
	}
}

// The guard sees each kind of decision it exists for, and passes each once it asks TreatedAs: the control
// that the module-wide pass above is a walk that could have failed.
func TestTheTypeScriptExtensionGuardSeesEachKindOfDecision(t *testing.T) {
	t.Parallel()
	source := `package probe
func probe(name, extension string) {
	_ = strings.HasSuffix(name, ".test.ts")
	_ = strings.HasSuffix(sourcename.TreatedAs(name), ".test.ts")
	_ = extension == ".ts"
	_ = filepath.Ext(sourcename.TreatedAs(name)) == ".ts"
	_ = []string{".ts", ".tsx"}
	_ = []string{".ts", ".tsx", ".a"}
	_ = map[string]bool{".ts": true}
	_ = regexp.MustCompile(` + "`" + `\.ts$` + "`" + `)
	_ = strings.HasSuffix(name, ".d.ts")
	switch filepath.Ext(name) {
	case ".ts", ".mts":
	}
	switch filepath.Ext(sourcename.TreatedAs(name)) {
	case ".ts":
	}
	_ = core.GetScriptKindFromFileName(name)
	_ = core.GetScriptKindFromFileName(sourcename.TreatedAs(name))
}`
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "probe.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, decision := range tsDecisionsIn(fileSet, file, "probe.go") {
		lines = append(lines, fmt.Sprintf("%d %s", decision.line, decision.literal))
	}
	want := []string{"3 .test.ts", "5 .ts", "7 .ts", "9 .ts", "10 \\.ts$", "13 .ts", "18 GetScriptKindFromFileName"}
	if !slices.Equal(lines, want) {
		t.Fatalf("the guard found %v, want %v", lines, want)
	}
}
