package lower

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Mutant run and restored: removing the rule 800 check returns the callable-contract
// Refused here. The module scan visits post before the ordinary rule walk reaches new.
func TestMergedHostMemberUsesNodeNotYet(t *testing.T) {
	t.Parallel()
	source := `import * as inspector from "node:inspector";
const session = new inspector.Session();
session.post("Profiler.enable", () => {});`
	_, err := lowerSourceAs(t, source, ".ts")
	var stop *NotYet
	want := "main.ts:3:1: stage 0 can't lower node:inspector.Session.post yet"
	if !errors.As(err, &stop) || !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("want %s, got %v", want, err)
	}
}

func TestMergedHostProgramCallableKeepsContract(t *testing.T) {
	t.Parallel()
	source := `class Session {}
interface Session {
 post(method: string): void;
 post(method: number): void;
}
function use(session: Session): void {
 session.post("Profiler.enable");
}`
	_, err := lowerSourceAs(t, source, ".ts")
	var stop *Refused
	want := "main.ts:7:2: Adamic 0.1 refuses a checked view with a callable contract; prove the callable body rather than asserting its signature"
	if !errors.As(err, &stop) || !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("want %s, got %v", want, err)
	}
}

// No registered bundled member reaches mergedFieldRead's contract path without an
// augmentation: EventEmitter inherits its methods and Buffer has no merged class.
// Augment the real, implemented Hash.update member to exercise the nil answer from
// rule 800. Its callable contract remains refused, exactly as before #nv1wemp.
// Mutant run and restored: returning rule 800's answer unconditionally bypasses
// that contract and makes this pin red. The ordinary host call still agrees with Node.
func TestMergedHostImplementedMemberKeepsContract(t *testing.T) {
	t.Parallel()
	source := `import { createHash } from "node:crypto";
declare module "node:crypto" {
 interface Hash { update(data: string): this; }
}
console.log(createHash("sha256").update("a").digest("hex"));`
	path := writeAgreementSource(t, "main.ts", source)
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	l := lowering{program: program, checker: checked, result: &ir.Program{}}
	var read *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "update" {
			read = node
		}
		return node.ForEachChild(visit)
	}
	file.AsNode().ForEachChild(visit)
	if read == nil {
		t.Fatal("missing Hash.update read")
	}
	field := checked.GetSymbolAtLocation(read.Name())
	if field == nil || initializedClassMember(field) || !nodeLibraryMemberImplemented(field) {
		t.Fatal("Hash.update must be implemented and still need the merged-member proof")
	}
	if err := l.nodeLibraryRefusal(read); err != nil {
		t.Fatalf("implemented host member: %v", err)
	}
	var stop *Refused
	err = l.mergedFieldRead(read)
	want := "main.ts:5:13: Adamic 0.1 refuses a checked view with a callable contract; prove the callable body rather than asserting its signature"
	if !errors.As(err, &stop) || filepath.Base(stop.Where) != "main.ts:5:13" || !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("want %s, got %v", want, err)
	}
	_, err = Lower(context.Background(), program)
	if !errors.As(err, &stop) || !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("whole lowering: want %s, got %v", want, err)
	}
	lowersAndAgreesWithNodeAs(t, `import { createHash } from "node:crypto";
console.log(createHash("sha256").update("a").digest("hex"));`, ".ts")
}
