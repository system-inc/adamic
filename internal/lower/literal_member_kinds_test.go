package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// refuseWidening judges an object literal's members of the kinds judgedLiteralMembers lists, and any other
// kind is not yet wherever a slot names its property (#wjqtypf): a kind the language adds later must fail
// closed, not lower unjudged. Every kind an object literal can hold today is listed, so no program reaches
// that clause, and this hands refuseWidening a member of an unlisted kind directly: a class's property
// declaration `p: W = w`, set as a member of a literal written into `{ p: W }`. Mutant: the fail-closed
// clause at refuseWidening's top dropped, under which the member falls to the default case, which judges
// expressions only, and lowers.
func TestUnlistedLiteralMemberKindFailsClosed(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.a")
	source := "interface W { t: number }\nconst w: W = { t: 2 };\nclass K { p: W = w; }\nconst v: { p: W } = { p: w };\nconsole.log(String(v.p.t + new K().p.t));\n"
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	t.Cleanup(release)
	var member, literal *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if member == nil && node.Kind == ast.KindPropertyDeclaration {
			member = node
		}
		if literal == nil && node.Kind == ast.KindObjectLiteralExpression && node.Parent.Kind == ast.KindVariableDeclaration && node.Parent.Name().Text() == "v" {
			literal = node
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	if member == nil || literal == nil {
		t.Fatalf("probe shape: member %v, literal %v", member, literal)
	}
	if judgedLiteralMembers[member.Kind] {
		t.Fatalf("%v is judged, so it can't stand in for an unlisted kind", member.Kind)
	}
	member.Parent = literal
	l := &lowering{program: program, checker: checked, result: &ir.Program{}, this: -1, functionIndex: -1}
	err = l.refuseWidening(member)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "in an object literal written where a slot names its property") {
		t.Fatalf("got %v, want the unjudged literal member's NotYet", err)
	}
}
