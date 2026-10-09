package typeaware

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
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
	for _, slice := range []struct {
		file, test string
		shards     int
	}{
		{"volume_profile_Controls_test.go", "TestVolumeProfileControls", testVolumeProfileControlsShards},
		{"volume_profile_Mutants_test.go", "TestVolumeProfileMutants", testVolumeProfileMutantsShards},
		{"volume_profile_Overlays_test.go", "TestVolumeProfileOverlays", testVolumeProfileOverlaysShards},
		{"volume_profile_Corpora_test.go", "TestVolumeProfileCorpora", testVolumeProfileCorporaShards},
	} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Clean(slice.file), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		tops := map[string]bool{}
		union := false
		for _, declaration := range tree.Decls {
			f, ok := declaration.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if f.Name.Name == slice.test+"Union" {
				union = true
			}
			if strings.HasPrefix(f.Name.Name, slice.test+"_") && f.Name.Name != slice.test+"_Setup" {
				tops[f.Name.Name] = true
			}
		}
		if !union {
			t.Errorf("%s: missing top-level union", slice.test)
		}
		for i := 0; i < slice.shards; i++ {
			if !tops[fmt.Sprintf("%s_%03d", slice.test, i)] {
				t.Errorf("missing top-level shard %s_%03d", slice.test, i)
			}
		}
		if len(tops) != slice.shards {
			t.Errorf("%s: %d top-level shards, want %d", slice.test, len(tops), slice.shards)
		}
		ast.Inspect(tree, func(node ast.Node) bool {
			table, ok := node.(*ast.CompositeLit)
			if !ok {
				return true
			}
			array, ok := table.Type.(*ast.ArrayType)
			if !ok {
				return true
			}
			switch array.Elt.(type) {
			case *ast.StructType, *ast.Ident:
			default:
				return true
			}
			for _, element := range table.Elts {
				row, ok := element.(*ast.CompositeLit)
				if !ok || len(row.Elts) == 0 {
					continue
				}
				name, ok := row.Elts[0].(*ast.BasicLit)
				if !ok || name.Kind != token.STRING {
					continue
				}
				key, err := strconv.Unquote(name.Value)
				if err != nil {
					t.Fatal(err)
				}
				if strings.HasPrefix(key, "typeaware ") {
					continue
				}
				owners[key] = append(owners[key], slice.test)
				if slice.test == "TestVolumeProfileCorpora" {
					owners[key+"-asan"] = append(owners[key+"-asan"], slice.test)
				}
			}
			return true
		})
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
