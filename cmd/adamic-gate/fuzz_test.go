package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
	"testing"
)

// Keep the gate's selectable leaves and their disjoint union tied to the original seed sweeps.
func TestFuzzChildrenCoverOriginalSeeds(t *testing.T) {
	t.Chdir("../..")
	for _, test := range []struct {
		parent, file string
		first, last  int
		vocabulary   bool
	}{
		{"TestGeneratedProgramsCheckAndLower", "internal/fuzz/fuzz_test.go", 1, 60, false},
		{"TestRegexProgramsPassTheChecker", "internal/fuzz/fuzz_test.go", 1, 30, false},
		{"TestUndefinedNumbersShapes", "internal/fuzz/undefined_numbers_test.go", 1, 40, true},
		{"TestOverridesShapesAndLower", "internal/fuzz/overrides_test.go", 0, 49, true},
	} {
		t.Run(test.parent, func(t *testing.T) {
			names, err := children("github.com/system-inc/adamic/internal/fuzz", test.parent)
			if err != nil {
				t.Fatal(err)
			}
			var want []string
			if test.vocabulary {
				want = append(want, "vocabulary")
			}
			for first := test.first; first <= test.last; first += 5 {
				want = append(want, fmt.Sprintf("seeds-%03d-%03d", first, first+4))
			}
			if !reflect.DeepEqual(names, want) {
				t.Fatalf("gate leaves %v, want %v", names, want)
			}
			tree, err := parser.ParseFile(token.NewFileSet(), test.file, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			counts := map[int]int{}
			for _, declaration := range tree.Decls {
				function, ok := declaration.(*ast.FuncDecl)
				if !ok || function.Name.Name != test.parent {
					continue
				}
				ast.Inspect(function.Body, func(node ast.Node) bool {
					row, ok := node.(*ast.CompositeLit)
					if !ok || len(row.Elts) != 3 {
						return true
					}
					label, ok := row.Elts[0].(*ast.BasicLit)
					if !ok || label.Kind != token.STRING {
						return true
					}
					name, err := strconv.Unquote(label.Value)
					if err != nil || name == "vocabulary" {
						return true
					}
					bounds := [2]int{}
					for i := range bounds {
						literal, ok := row.Elts[i+1].(*ast.BasicLit)
						if !ok {
							t.Fatalf("%s has a nonliteral bound", name)
						}
						bounds[i], err = strconv.Atoi(literal.Value)
						if err != nil {
							t.Fatal(err)
						}
					}
					if name != fmt.Sprintf("seeds-%03d-%03d", bounds[0], bounds[1]) || bounds[1] < bounds[0] {
						t.Fatalf("%s disagrees with bounds %v", name, bounds)
					}
					for seed := bounds[0]; seed <= bounds[1]; seed++ {
						counts[seed]++
					}
					return false
				})
			}
			if len(counts) != test.last-test.first+1 {
				t.Fatalf("union has %d seeds, want %d", len(counts), test.last-test.first+1)
			}
			for seed := test.first; seed <= test.last; seed++ {
				if counts[seed] != 1 {
					t.Errorf("seed %d has %d owners, want one", seed, counts[seed])
				}
			}
		})
	}
}

// The expensive checkout is parent setup; neither assertion leaf depends on its sibling running.
func TestFuzzCheckoutChildren(t *testing.T) {
	t.Chdir("../..")
	for _, test := range []struct {
		parent string
		want   []string
	}{
		{"TestReduceKeepsTheSignature", []string{"signature", "reduction"}},
		{"TestFuzzerSharesRuntimeLibrary", []string{"library", "program"}},
	} {
		names, err := children("github.com/system-inc/adamic/internal/fuzz", test.parent)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(names, test.want) {
			t.Fatalf("%s leaves %v, want %v", test.parent, names, test.want)
		}
	}
}
