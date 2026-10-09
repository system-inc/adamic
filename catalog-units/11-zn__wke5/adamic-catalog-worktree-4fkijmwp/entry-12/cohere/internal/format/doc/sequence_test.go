package doc

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestASequenceIsTheConcatOfItsParts builds each random spec twice, once with every array a Concat and once
// with every array a *Sequence, and asks every walk the same question of both: printing, cleaning, mapping,
// stripping and the break predicates must not be able to tell them apart (#fyw36kf).
func TestASequenceIsTheConcatOfItsParts(t *testing.T) {
	t.Parallel()
	random := rand.New(rand.NewSource(20261005))
	withArrays := 0
	for index := range 3000 {
		generator := &specGenerator{random: random}
		spec := generator.node()
		options := Options{PrintWidth: 8 + random.Intn(40), TabWidth: 2}
		concats := func() Doc { return buildSpecAs(spec, false) }
		sequences := func() Doc { return buildSpecAs(spec, true) }
		if showDoc(concats()) != showDoc(sequences()) {
			t.Fatalf("case %d: the two builds show differently", index)
		}
		if holdsConcat(spec) {
			withArrays++
		}

		questions := []struct {
			name string
			ask  func(Doc) string
		}{
			{"Print", func(document Doc) string { return Print(document, options) }},
			{"CleanDoc", func(document Doc) string { return showDoc(CleanDoc(document)) }},
			{"MapDoc", func(document Doc) string { return showDoc(MapDoc(document, cleanDocFn)) }},
			{"StripTrailingHardline", func(document Doc) string { return showDoc(StripTrailingHardline(document)) }},
			{"RemoveLines", func(document Doc) string { return showDoc(RemoveLines(document)) }},
			{"WillBreak", func(document Doc) string { return boolText(WillBreak(document)) }},
			{"CanBreak", func(document Doc) string { return boolText(CanBreak(document)) }},
			{"IsEmptyDoc", func(document Doc) string { return boolText(IsEmptyDoc(document)) }},
		}
		for _, question := range questions {
			if got, want := question.ask(sequences()), question.ask(concats()); got != want {
				t.Fatalf("case %d: %s of the sequences is\n%s\nand of the concats\n%s", index, question.name, got, want)
			}
		}
	}
	// The comparison can only fail if the specs hold arrays to build as sequences.
	if withArrays < 1000 {
		t.Fatalf("only %d of 3000 specs held an array", withArrays)
	}
}

// holdsConcat is whether a spec builds at least one array of its own, which buildSpecAs can make a
// *Sequence.
func holdsConcat(node *spec) bool {
	if node.Kind == "concat" {
		return true
	}
	for _, child := range node.Children {
		if holdsConcat(child) {
			return true
		}
	}
	return false
}

func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

// TestOnlyPartsAsksWhetherADocIsAnArray walks every Go file under internal/format and fails on a type
// assertion or type switch case naming Concat or Sequence anywhere but Parts. A walk that asks for one and
// not the other would treat a *Sequence as some other doc on exactly the path the corpus may not reach,
// so every walk asks Parts, which knows both.
func TestOnlyPartsAsksWhetherADocIsAnArray(t *testing.T) {
	t.Parallel()

	planted := `package javascript

func f(d doc.Doc) {
	_, _ = d.(doc.Concat)
	switch d.(type) {
	case *doc.Sequence:
	}
}
`
	if found := arrayTypeChecks(t, "planted.go", planted); len(found) != 2 {
		t.Fatalf("the check found %d of the 2 planted checks: %v", len(found), found)
	}

	// The test runs in internal/format/doc, so its parent is internal/format.
	var found []string
	err := filepath.WalkDir("..", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		found = append(found, arrayTypeChecks(t, path, string(source))...)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(found) > 0 {
		t.Fatalf("ask doc.Parts instead of naming Concat or Sequence:\n%s", strings.Join(found, "\n"))
	}
}

// arrayTypeChecks is the positions in source where a type assertion or type switch case names Concat or
// Sequence, outside the function Parts.
func arrayTypeChecks(t *testing.T, path string, source string) []string {
	t.Helper()
	files := token.NewFileSet()
	file, err := parser.ParseFile(files, path, source, 0)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	var found []string
	report := func(expression ast.Expr) {
		if namesArrayType(expression) {
			found = append(found, files.Position(expression.Pos()).String())
		}
	}
	for _, declaration := range file.Decls {
		if function, isFunction := declaration.(*ast.FuncDecl); isFunction && function.Recv == nil && function.Name.Name == "Parts" && file.Name.Name == "doc" {
			continue
		}
		ast.Inspect(declaration, func(node ast.Node) bool {
			switch typed := node.(type) {
			case *ast.TypeAssertExpr:
				if typed.Type != nil {
					report(typed.Type)
				}
			case *ast.TypeSwitchStmt:
				for _, statement := range typed.Body.List {
					for _, listed := range statement.(*ast.CaseClause).List {
						report(listed)
					}
				}
			}
			return true
		})
	}
	return found
}

// namesArrayType is whether expression is Concat or Sequence, bare or as doc.Concat, behind a pointer or not.
func namesArrayType(expression ast.Expr) bool {
	if star, isStar := expression.(*ast.StarExpr); isStar {
		expression = star.X
	}
	var name string
	switch typed := expression.(type) {
	case *ast.Ident:
		name = typed.Name
	case *ast.SelectorExpr:
		if packageName, isIdent := typed.X.(*ast.Ident); isIdent && packageName.Name == "doc" {
			name = typed.Sel.Name
		}
	}
	return name == "Concat" || name == "Sequence"
}
