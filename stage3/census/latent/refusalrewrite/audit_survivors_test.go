package refusalrewrite

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"strings"
	"testing"
)

func minimalRefusalSource(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("testdata/two-outer-returns.go.txt")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestFullOwnerGuardRequiresFunctionDeclaration(t *testing.T) {
	t.Parallel()
	output, err := Rewrite(minimalRefusalSource(t))
	if err != nil {
		t.Fatal(err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), "rewritten.go", output, 0)
	if err != nil {
		t.Fatal(err)
	}
	guards := 0
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != "latentRefuse" {
			continue
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			conditional, ok := node.(*ast.IfStmt)
			if !ok {
				return true
			}
			condition := printed(conditional.Cond)
			if !strings.Contains(condition, "latentFullEnabled()") {
				return true
			}
			guards++
			// Evaluate the actual emitted condition with Go's constant evaluator.
			// Full mode owns a function body only when both facts hold; a non-function
			// has no body to inspect, and disabled full mode must use the ordinary path.
			for _, probe := range []struct {
				full bool
				kind int
				want bool
			}{
				{false, 0, false}, {false, 1, false}, {true, 0, false}, {true, 1, true},
			} {
				expression := strings.NewReplacer("latentFullEnabled()", fmt.Sprint(probe.full), "node.Kind", fmt.Sprint(probe.kind), "ast.KindFunctionDeclaration", "1").Replace(condition)
				value, err := types.Eval(token.NewFileSet(), nil, token.NoPos, expression)
				if err != nil {
					t.Fatalf("evaluating emitted owner guard %q: %v", expression, err)
				}
				if value.Value == nil || value.Value.Kind() != constant.Bool {
					t.Fatalf("owner guard is not boolean: %q", expression)
				}
				if got := constant.BoolVal(value.Value); got != probe.want {
					t.Errorf("owner guard %q: full=%v function=%v got %v, want %v", condition, probe.full, probe.kind == 1, got, probe.want)
				}
			}
			return true
		})
	}
	if guards != 2 {
		t.Fatalf("expected guards for both refusal walkers, found %d", guards)
	}
}

func TestRewriteAcceptsTwoOuterRefusalReturns(t *testing.T) {
	t.Parallel()
	source := minimalRefusalSource(t)
	before, err := parser.ParseFile(token.NewFileSet(), "before.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	returns := 0
	ast.Inspect(before, func(node ast.Node) bool {
		if _, nested := node.(*ast.FuncLit); nested {
			return false
		}
		if _, returned := node.(*ast.ReturnStmt); returned {
			returns++
		}
		return true
	})
	if returns != 2 {
		t.Fatalf("minimal supported refusal input has %d outer returns, want 2", returns)
	}
	output, err := Rewrite(source)
	if err != nil {
		t.Fatalf("supported two-return refusal rejected: %v", err)
	}
	after, err := parser.ParseFile(token.NewFileSet(), "after.go", output, 0)
	if err != nil {
		t.Fatal(err)
	}
	var original, production *ast.FuncDecl
	latent := false
	for _, declaration := range before.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Name.Name == "refuse" {
			original = function
		}
	}
	for _, declaration := range after.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok {
			if function.Name.Name == "refuse" {
				production = function
			}
			if function.Name.Name == "latentRefuse" {
				latent = true
			}
		}
	}
	if original == nil || production == nil || printed(original) != printed(production) {
		t.Fatal("production refusal changed")
	}
	if !latent {
		t.Fatal("two-return input did not produce the latent refusal walker")
	}
}
