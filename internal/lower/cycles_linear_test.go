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

func cycleIndexTypes(test *testing.T, source string) (*cycleFinder, map[string]*checker.Type) {
	test.Helper()
	path := filepath.Join(test.TempDir(), "main.a")
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	entry := program.Files()[0]
	checked, release := program.Checker(context.Background(), entry)
	test.Cleanup(release)
	lowering := &lowering{program: program, checker: checked, result: &ir.Program{}, this: -1, functionIndex: -1}
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
	return &cycleFinder{lowering: lowering, where: map[*checker.Type]*ast.Node{}}, types
}

// A structural view hides the only back edge. The property posting list must
// retain its actual allocation shape, even when the view itself has no back slot.
func TestCycleShapeIndexKeepsHiddenEdge(test *testing.T) {
	test.Parallel()
	finder, types := cycleIndexTypes(
		test,
		`declare const view: {visible:number}; declare const actual: {visible:number; hidden:{target:true}}; declare const target:{target:true};`,
	)
	finder.shapes = []*checker.Type{types["actual"]}
	if !finder.reaches(types["view"], cycleNode{proven: types["target"]}) {
		test.Fatal("shape index missed a hidden strong edge")
	}
}

func TestCycleShapeIndexOptionalProperties(test *testing.T) {
	test.Parallel()
	finder, types := cycleIndexTypes(
		test,
		`declare const view:{visible?:number}; declare const actual:{hidden:{target:true}}; declare const target:{target:true};`,
	)
	finder.shapes = []*checker.Type{types["actual"]}
	candidates := finder.shapeCandidates(types["view"])
	if len(candidates) != 1 || candidates[0] != types["actual"] {
		test.Fatal("optional property discarded a possible shape")
	}
}

func TestCycleShapeIndexInvalidatesReachability(test *testing.T) {
	test.Parallel()
	finder, types := cycleIndexTypes(
		test,
		`declare const view:{visible:number}; declare const actual:{visible:number;hidden:{target:true}}; declare const target:{target:true};`,
	)
	if finder.reaches(types["view"], cycleNode{proven: types["target"]}) {
		test.Fatal("unrelated types reached before a shape existed")
	}
	finder.shapes = append(finder.shapes, types["actual"])
	if !finder.reaches(types["view"], cycleNode{proven: types["target"]}) {
		test.Fatal("cached negative answer hid a newly discovered shape")
	}
}

func TestCycleLiteralLayoutsKeepAccessorsDistinct(test *testing.T) {
	test.Parallel()
	finder, types := cycleIndexTypes(
		test,
		`const stored={value:1};const also={value:2};const accessor={get value():number{return 1;}};`,
	)
	if finder.literalType(types["stored"]) != finder.literalType(types["also"]) {
		test.Fatal("identical strong layouts were not folded")
	}
	if finder.literalType(types["stored"]) == finder.literalType(types["accessor"]) {
		test.Fatal("accessor folded into a stored strong slot")
	}
}

// Repeated allocation shapes share one indexed layout. Count the candidates,
// rather than asserting a machine-dependent timing ratio.
func TestCycleShapeIndexRepeatedLayouts(test *testing.T) {
	test.Parallel()
	source := "declare const view:{value:number};\n"
	for index := 0; index < 1000; index++ {
		source += fmt.Sprintf("const item%d={value:%d};\n", index, index)
	}
	finder, types := cycleIndexTypes(test, source)
	for index := 0; index < 1000; index++ {
		finder.shapes = append(finder.shapes, types[fmt.Sprint("item", index)])
	}
	if candidates := finder.shapeCandidates(types["view"]); len(candidates) != 1 {
		test.Fatalf("1000 identical layouts left %d candidates", len(candidates))
	}
}

func TestCycleLiteralLayoutsKeepTargetsDistinct(test *testing.T) {
	test.Parallel()
	finder, types := cycleIndexTypes(test, `const number={value:1};const text={value:"held"};`)
	if finder.literalType(types["number"]) == finder.literalType(types["text"]) {
		test.Fatal("different strong slot types were folded")
	}
}
func TestCycleWriteIndexKeepsUnknown(test *testing.T) {
	test.Parallel()
	finder, types := cycleIndexTypes(test, `declare const holder:{edge:number};`)
	finder.writes = []fresh.Write{
		{Kind: fresh.WriteField, Name: "other"},
		{Kind: fresh.WriteUnknown},
		{Kind: fresh.WriteField, Name: "edge"},
	}
	if got := finder.cycleUnproven(fresh.WriteField, types["holder"], "edge"); got != &finder.writes[1] {
		test.Fatal("unknown write disappeared from the slot index")
	}
}
func TestCycleWriteIndexKeepsFirstWrite(test *testing.T) {
	test.Parallel()
	finder, types := cycleIndexTypes(test, `declare const holder:{edge:number};`)
	finder.writes = []fresh.Write{
		{Kind: fresh.WriteField, Name: "other"},
		{Kind: fresh.WriteField, Name: "edge"},
		{Kind: fresh.WriteField, Name: "edge"},
		{Kind: fresh.WriteUnknown},
	}
	if got := finder.cycleUnproven(fresh.WriteField, types["holder"], "edge"); got != &finder.writes[1] {
		test.Fatal("slot index changed the first unproven write")
	}
}
