package typeaware

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"testing"
)

// Inventory copied from TestVolumeProfileAgreementAndMutants at 7a10c877.
// Read the actual execution tables, rather than maintaining a second purported
// enumeration. Optional corpus checks remain owned even when their env is unset.
func TestVolumeProfilePartition(t *testing.T) {
	t.Parallel()
	expected := []struct{ check, owner string }{
		{"controls", "TestVolumeProfileControls"},
		{"controls-asan", "TestVolumeProfileControls"},
		{"binding-slot", "TestVolumeProfileMutants"},
		{"scope-containment", "TestVolumeProfileMutants"},
		{"first-binding", "TestVolumeProfileMutants"},
		{"index-kind", "TestVolumeProfileOverlays"},
		{"root-kind", "TestVolumeProfileOverlays"},
		{"index-end", "TestVolumeProfileOverlays"},
		{"repository", "TestVolumeProfileCorpora"},
		{"repository-asan", "TestVolumeProfileCorpora"},
		{"compiler", "TestVolumeProfileCorpora"},
		{"compiler-asan", "TestVolumeProfileCorpora"},
	}
	want := map[string]string{}
	for _, row := range expected {
		if _, duplicate := want[row.check]; duplicate {
			t.Fatalf("duplicate inventory check %s", row.check)
		}
		want[row.check] = row.owner
	}
	owners := map[string][]string{}
	for _, slice := range []struct{ file, test string }{
		{"volume_profile_Controls_test.go", "TestVolumeProfileControls"},
		{"volume_profile_Mutants_test.go", "TestVolumeProfileMutants"},
		{"volume_profile_Overlays_test.go", "TestVolumeProfileOverlays"},
		{"volume_profile_Corpora_test.go", "TestVolumeProfileCorpora"},
	} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Clean(slice.file), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		var function *ast.FuncDecl
		for _, declaration := range tree.Decls {
			if f, ok := declaration.(*ast.FuncDecl); ok && f.Name.Name == slice.test {
				function = f
			}
		}
		if function == nil {
			t.Fatalf("missing owner %s", slice.test)
		}
		parallel, compare, mutantCheck := false, false, false
		ast.Inspect(function.Body, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
					if receiver, ok := selector.X.(*ast.Ident); ok {
						if receiver.Name == "t" && selector.Sel.Name == "Parallel" {
							parallel = true
						}
						if receiver.Name == "h" && selector.Sel.Name == "compare" {
							compare = true
						}
						if receiver.Name == "bytes" && (selector.Sel.Name == "Equal" || selector.Sel.Name == "Contains") {
							mutantCheck = true
						}
					}
				}
			}
			// Each execution case is a struct literal whose first field is its name.
			row, ok := node.(*ast.CompositeLit)
			if !ok || len(row.Elts) == 0 {
				return true
			}
			name, ok := row.Elts[0].(*ast.BasicLit)
			if !ok || name.Kind != token.STRING {
				return true
			}
			key, err := strconv.Unquote(name.Value)
			if err != nil {
				t.Fatal(err)
			}
			// Exclude string arrays used for flags and native-global generation.
			if row.Type != nil {
				return true
			}
			owners[key] = append(owners[key], slice.test)
			if slice.test == "TestVolumeProfileCorpora" {
				owners[key+"-asan"] = append(owners[key+"-asan"], slice.test)
			}
			return true
		})
		if !parallel {
			t.Errorf("%s has no parallel shard", slice.test)
		}
		if slice.test == "TestVolumeProfileControls" || slice.test == "TestVolumeProfileCorpora" {
			if !compare {
				t.Errorf("%s lost byte comparison", slice.test)
			}
		} else if !mutantCheck {
			t.Errorf("%s lost mutant rejection", slice.test)
		}
	}
	for check, owner := range want {
		actual := owners[check]
		if len(actual) != 1 || actual[0] != owner {
			t.Errorf("%s: owners %v, want exactly %s", check, actual, owner)
		}
	}
	for check, actual := range owners {
		if _, exists := want[check]; !exists {
			t.Errorf("unexpected check %s owned by %v", check, actual)
		}
	}
	legacy, err := parser.ParseFile(token.NewFileSet(), "profile_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range legacy.Decls {
		if f, ok := declaration.(*ast.FuncDecl); ok && f.Name.Name == "TestVolumeProfileAgreementAndMutants" {
			t.Fatal("legacy profile test doubles the partition")
		}
	}
	for _, row := range expected {
		t.Log(fmt.Sprintf("%s -> %s", row.check, row.owner))
	}
}
