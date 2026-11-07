package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
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

// methodSlotMembers keeps the members of a union slot a literal's method is assignable to, and when it's
// assignable to none, every member with a call signature, so the method is judged against all of them, never
// none (#wjqtypf, Lower's must on b3e90772). tsc takes a function into a union of function types only through
// a member, so no program is known to reach that fallback, and this hands methodSlotMembers a function type
// assignable to neither member of `((x: number) => W) | ((x: boolean) => W)` directly: `(x: string) => W`.
// Mutant (n): the fallback dropped, under which the list is empty and the method is judged against nothing.
func TestMethodSlotMembersFallBackToEveryCallable(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.a")
	source := "interface W { t: number }\nconst w: W = { t: 2 };\nconst own = (x: string): W => w;\nconst slot: ((x: number) => W) | ((x: boolean) => W) = (x: number): W => w;\nconsole.log(String(own('a').t) + typeof slot);\n"
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
	var own, slot *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindVariableDeclaration {
			switch node.Name().Text() {
			case "own":
				own = node.Name()
			case "slot":
				slot = node.Name()
			}
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	if own == nil || slot == nil {
		t.Fatalf("probe shape: own %v, slot %v", own, slot)
	}
	l := &lowering{program: program, checker: checked, result: &ir.Program{}, this: -1, functionIndex: -1}
	ownType, slotType := checked.GetTypeAtLocation(own), checked.GetTypeAtLocation(slot)
	if slotType.Flags()&checker.TypeFlagsUnion == 0 || len(slotType.Types()) != 2 {
		t.Fatalf("slot %s isn't a union of two", checked.TypeToString(slotType))
	}
	for _, member := range slotType.Types() {
		if checked.IsTypeAssignableTo(ownType, member) {
			t.Fatalf("%s is assignable to %s, so the probe doesn't reach the fallback", checked.TypeToString(ownType), checked.TypeToString(member))
		}
	}
	if members := l.methodSlotMembers(ownType, slotType); len(members) != 2 {
		t.Fatalf("got %d members, want both of %s", len(members), checked.TypeToString(slotType))
	}
}
