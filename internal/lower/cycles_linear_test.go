package lower

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func cycleIndexTypes(t *testing.T, source string) (*cycleFinder, map[string]*checker.Type) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	entry := program.Files()[0]
	checked, release := program.Checker(context.Background(), entry)
	t.Cleanup(release)
	l := &lowering{program: program, checker: checked, result: &ir.Program{}, this: -1, functionIndex: -1}
	types := map[string]*checker.Type{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindVariableDeclaration && node.Name() != nil {
			location := node.Name()
			if node.AsVariableDeclaration().Initializer != nil {
				location = node.AsVariableDeclaration().Initializer
			}
			types[node.Name().Text()] = checked.GetTypeAtLocation(location)
		}
		return node.ForEachChild(visit)
	}
	entry.AsNode().ForEachChild(visit)
	return &cycleFinder{l: l, where: map[*checker.Type]*ast.Node{}}, types
}

// A structural view hides the only back edge. The property posting list must
// retain its actual allocation shape, even when the view itself has no back slot.
func TestCycleShapeIndexKeepsHiddenEdge(t *testing.T) {
	t.Parallel()
	f, types := cycleIndexTypes(t, `declare const view: {visible:number}; declare const actual: {visible:number; hidden:{target:true}}; declare const target:{target:true};`)
	f.shapes = []*checker.Type{types["actual"]}
	if !f.reaches(types["view"], cycleNode{proven: types["target"]}) {
		t.Fatal("shape index missed a hidden strong edge")
	}
}

func TestCycleShapeIndexOptionalProperties(t *testing.T) {
	t.Parallel()
	f, types := cycleIndexTypes(t, `declare const view:{visible?:number}; declare const actual:{hidden:{target:true}}; declare const target:{target:true};`)
	f.shapes = []*checker.Type{types["actual"]}
	candidates := f.shapeCandidates(types["view"])
	if len(candidates) != 1 || candidates[0] != types["actual"] {
		t.Fatal("optional property discarded a possible shape")
	}
}

func TestCycleShapeIndexInvalidatesReachability(t *testing.T) {
	t.Parallel()
	f, types := cycleIndexTypes(t, `declare const view:{visible:number}; declare const actual:{visible:number;hidden:{target:true}}; declare const target:{target:true};`)
	if f.reaches(types["view"], cycleNode{proven: types["target"]}) {
		t.Fatal("unrelated types reached before a shape existed")
	}
	f.shapes = append(f.shapes, types["actual"])
	if !f.reaches(types["view"], cycleNode{proven: types["target"]}) {
		t.Fatal("cached negative answer hid a newly discovered shape")
	}
}

func TestCycleLiteralLayoutsKeepAccessorsDistinct(t *testing.T) {
	t.Parallel()
	f, types := cycleIndexTypes(t, `const stored={value:1};const also={value:2};const accessor={get value():number{return 1;}};`)
	if f.literalType(types["stored"]) != f.literalType(types["also"]) {
		t.Fatal("identical strong layouts were not folded")
	}
	if f.literalType(types["stored"]) == f.literalType(types["accessor"]) {
		t.Fatal("accessor folded into a stored strong slot")
	}
}

// Repeated allocation shapes share one indexed layout. Count the candidates,
// rather than asserting a machine-dependent timing ratio.
func TestCycleShapeIndexRepeatedLayouts(t *testing.T) {
	t.Parallel()
	source := "declare const view:{value:number};\n"
	for i := 0; i < 1000; i++ {
		source += fmt.Sprintf("const item%d={value:%d};\n", i, i)
	}
	f, types := cycleIndexTypes(t, source)
	for i := 0; i < 1000; i++ {
		f.shapes = append(f.shapes, types[fmt.Sprint("item", i)])
	}
	if candidates := f.shapeCandidates(types["view"]); len(candidates) != 1 {
		t.Fatalf("1000 identical layouts left %d candidates", len(candidates))
	}
}

func TestCycleLiteralLayoutsKeepTargetsDistinct(t *testing.T) {
	t.Parallel()
	f, types := cycleIndexTypes(t, `const number={value:1};const text={value:"held"};`)
	if f.literalType(types["number"]) == f.literalType(types["text"]) {
		t.Fatal("different strong slot types were folded")
	}
}
func TestCycleWriteIndexKeepsUnknown(t *testing.T) {
	t.Parallel()
	f, types := cycleIndexTypes(t, `declare const holder:{edge:number};`)
	f.writes = []fresh.Write{{Kind: fresh.WriteField, Name: "other"}, {Kind: fresh.WriteUnknown}, {Kind: fresh.WriteField, Name: "edge"}}
	if got := f.cycleUnproven(fresh.WriteField, types["holder"], "edge"); got != &f.writes[1] {
		t.Fatal("unknown write disappeared from the slot index")
	}
}
func TestCycleWriteIndexKeepsFirstWrite(t *testing.T) {
	t.Parallel()
	f, types := cycleIndexTypes(t, `declare const holder:{edge:number};`)
	f.writes = []fresh.Write{{Kind: fresh.WriteField, Name: "other"}, {Kind: fresh.WriteField, Name: "edge"}, {Kind: fresh.WriteField, Name: "edge"}, {Kind: fresh.WriteUnknown}}
	if got := f.cycleUnproven(fresh.WriteField, types["holder"], "edge"); got != &f.writes[1] {
		t.Fatal("slot index changed the first unproven write")
	}
}
