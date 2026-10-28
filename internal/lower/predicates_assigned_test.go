package lower

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

// A lowering collects every assigned symbol of the program once, and assignedSymbol answers as a fresh lowering does:
// two files in one program, a name assigned and a name only read in each, every name asked in order and then reversed
// (#pv935yw). Mutants: the walk reading the first file alone (the second file's assignment is missed), a kept answer
// read as false, or every name answered alike, each go red here and only here. TestPredicateCallbackContracts/reassigned
// holds the refusal itself, in one file, so none of the three reaches it.
func TestAssignedSymbolAnswersAsAFreshLowering(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	first, second := filepath.Join(directory, "main.a"), filepath.Join(directory, "other.a")
	for path, source := range map[string]string{
		first:  "let assignedHere = 1; assignedHere = 2; const readHere = 3; console.log((assignedHere + readHere).toString());\n",
		second: "let assignedThere = 1; assignedThere += 1; const readThere = 4; console.log((assignedThere + readThere).toString());\n",
	} {
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	program, err := load.Load([]string{first, second})
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Files()) < 2 {
		t.Fatalf("the program holds %d files, want both roots", len(program.Files()))
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	want := map[string]bool{"assignedHere": true, "readHere": false, "assignedThere": true, "readThere": false}
	var names []*ast.Node
	for _, file := range program.Files() {
		var visit ast.Visitor
		visit = func(n *ast.Node) bool {
			if n.Kind == ast.KindVariableDeclaration {
				if _, ok := want[n.Name().Text()]; ok {
					names = append(names, n.Name())
				}
			}
			n.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
	}
	if len(names) != len(want) {
		t.Fatalf("found %d of the %d names", len(names), len(want))
	}
	fresh := func() *lowering { return &lowering{program: program, checker: checked} }
	for _, name := range names {
		if got := fresh().assignedSymbol(fresh().symbol(name)); got != want[name.Text()] {
			t.Fatalf("%s alone: assigned=%t, want %t", name.Text(), got, want[name.Text()])
		}
	}
	kept := fresh()
	reversed := slices.Clone(names)
	slices.Reverse(reversed)
	for _, name := range append(slices.Clone(names), reversed...) {
		if got := kept.assignedSymbol(kept.symbol(name)); got != want[name.Text()] {
			t.Fatalf("%s from the memo: assigned=%t, want %t", name.Text(), got, want[name.Text()])
		}
	}
}
