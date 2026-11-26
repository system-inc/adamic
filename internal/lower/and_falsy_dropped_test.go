package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"strings"
	"testing"
)

// An && whose consumer drops a falsy result (andFalsyDropped, Adamic's ruling of Nov 19 on #snypn8c, rule 3) hands
// no slot its left, so only its right is judged: tsc's compareSignaturesRelated under || (boolean_or, checker.ts:21987;
// double_and_or, :22020; number_and_assign through &=, :22066), a ! (element_flags, :32547, and not_and, with a
// conditional's test) and an if condition (if_condition). A left that can only be truthy never reaches the result
// (truthy_literal, falsyArms). Each lowers, and the oracle holds each to Node (internal/oracle/and_falsy_dropped_test.go).
// Mutants in internal/oracle/testdata/and_falsy_dropped/mutants.md: M1, the helper answering no, refuses boolean_or
// and double_and_or in freshOrWidened's && case (M1b) and the number and flag lefts in refuseWidening's site walk
// (M1a); M4, truthy literals kept, refuses truthy_literal.
func TestAndFalsyDroppedLowers(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"boolean_or", "double_and_or", "number_and_assign", "element_flags", "if_condition", "not_and", "truthy_literal"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile("../oracle/testdata/and_falsy_dropped/" + name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := lowerSourceAs(t, string(source), ".a"); err != nil {
				t.Fatalf("want it lowered, got %v", err)
			}
		})
	}
}

// A falsy && result some slot sees before any consumer drops it is judged as it is, refused in today's words at the
// number left whose 0 or NaN reaches the slot: stored into a Ternary slot first, then used under || (stored_first), on
// an ||'s right, which hands the result on whole (or_right), under ??, which keeps falsy values (nullish), and a
// number left returned directly (number_stored). Mutants: M2, an ||'s right counted as dropping, lowers or_right; M3,
// ?? counted, lowers nullish (mutants.md).
func TestAndFalsyDroppedStopsAreRefused(t *testing.T) {
	t.Parallel()
	for name, want := range map[string]string{
		"stored_first":  "main.a:9:26: ",
		"or_right":      "main.a:8:35: ",
		"nullish":       "main.a:8:18: ",
		"number_stored": "main.a:7:9: ",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile("../oracle/testdata/and_falsy_dropped/" + name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			want += "Adamic 0.1 refuses an unproven value assigned to a numeric literal or enum member slot Ternary.False; compare with this literal and return that constant, or widen the slot to the whole numeric enum or number (adamic/enum-literal)"
			if _, err := lowerSourceAs(t, string(source), ".a"); err == nil || !strings.HasSuffix(err.Error(), want) {
				t.Fatalf("want %q, got %v", want, err)
			}
		})
	}
}

// Every walk that judges an && left, by its falsy arms (falsyArms) or as the left of an && (andOfLeft), asks
// andFalsyDropped, the one helper the consumer's drop lives in, and the walks are exactly the ones named here, so a
// third walk judging an && left is named here and asks the helper too, never a copy of the drop. Mutant M5: the list
// short by one, which this reds.
func TestAndFalsyDroppedWalks(t *testing.T) {
	t.Parallel()
	walks := []string{"freshOrWidened", "refuseWidening"}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	var judging []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, entry.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if !isFunction || function.Body == nil {
				continue
			}
			calls := map[string]bool{}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, isCall := node.(*ast.CallExpr)
				if !isCall {
					return true
				}
				switch callee := call.Fun.(type) {
				case *ast.Ident:
					calls[callee.Name] = true
				case *ast.SelectorExpr:
					calls[callee.Sel.Name] = true
				}
				return true
			})
			if !calls["falsyArms"] && !calls["andOfLeft"] {
				continue
			}
			judging = append(judging, function.Name.Name)
			if !calls["andFalsyDropped"] {
				t.Errorf("%s (%s) judges an && left but never asks andFalsyDropped, so a falsy result its consumer drops is still judged against a slot", function.Name.Name, entry.Name())
			}
		}
	}
	slices.Sort(judging)
	if !slices.Equal(judging, walks) {
		t.Errorf("the walks judging an && left are %v, want %v: name a new walk here, and have it ask andFalsyDropped", judging, walks)
	}
}
