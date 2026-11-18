package lower

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// containerFitBoundFakeMap is a class that implements ReadonlyMap without being one, held as an object.
const containerFitBoundFakeMap = "const backing = new Map<string, number>([['a', 1]]);\n" +
	"class FakeMap implements ReadonlyMap<string, number> {\n" +
	"\treadonly size = 1;\n" +
	"\tget(key: string): number | undefined { return backing.get(key); }\n" +
	"\thas(key: string): boolean { return key === 'a'; }\n" +
	"\tforEach(): void {}\n" +
	"\tkeys() { return backing.keys(); }\n" +
	"\tvalues() { return backing.values(); }\n" +
	"\tentries() { return backing.entries(); }\n" +
	"\t[Symbol.iterator]() { return backing.entries(); }\n" +
	"}\n"

// Past containerFitDepth a generic pair that can reach a map keeps walking until either side is a generic
// growing deeper on the walk's path, where typeNesting refuses it and the walk fails closed
// (containerFitDepthStop). An end-to-end program can stop on another wall before the walk is reached (the
// expanding shapes below hang in records.go's sameRecordStorage on main, #fzx0j5z), so this hands
// containerWhereOtherGoes the two declared types directly and bounds it by the pairs it judges, never a
// clock. One growing shape per way typeNesting names a recursion identity: a type reference by its generic
// target (reference, and tuple, whose growth passes through a tuple argument), and an instantiated
// anonymous type by its symbol (anonymous). A mapped type is named by its symbol as an anonymous one is;
// grown through an intersection with a map field, its walk reads two pairs and fits on J-01p6kwe, so it
// isn't a growing shape here. The controls: a real Map in a Box<T> five
// fields down fits, the same with a FakeMap stops at the real check, and a predicate over a non-generic
// self-referential node, tsc's shape behind most of the 35 sites, fits. A generic that repeats without
// growing (repeating: a Box<T> holding a Box<T>, seen as a View<T> holding a View<T>) ends by the walk's
// visited set of pairs in progress, never by typeNesting, which counts growth, or by the cap.
// Mutants: the gate ignoring typeNesting's refusal runs a growing shape to containerFitPairCap or past any
// clock (on J-ssgz7y7 reference judged 16,393 pairs and stopped at "a fit walk past 4096 distinct pairs",
// where each reads 21 at the head, and tuple ran past Go's 20-minute test timeout), so each reddens; the
// gate stopping every generic pair that can reach a map, as before #rt5evf5 head 2, reddens box_real_map
// and box_fake_map, which then stops at the gate before its leaf is read; the visited set of pairs in
// progress removed (containerPairGoes' walk.busy check) runs repeating to the cap.
func TestContainerFitBoundsGrowingGenerics(t *testing.T) {
	const ownDeep = "interface Own<T> {\n\treadonly next: Own<Own<T>> | undefined;\n\treadonly m: Map<string, T>;\n}\n" +
		"interface Deep<T> {\n\treadonly next: Deep<Deep<T>> | undefined;\n\treadonly m: ReadonlyMap<string, T>;\n}\n"
	for _, probe := range []struct{ name, declarations, value, target, want string }{
		{"reference", ownDeep, "Own<number>", "Deep<number>", containerFitDepthStop},
		{"tuple", "interface Own<T> {\n\treadonly next: Own<[T, T]> | undefined;\n\treadonly m: Map<string, T>;\n}\n" +
			"interface Deep<T> {\n\treadonly next: Deep<[T, T]> | undefined;\n\treadonly m: ReadonlyMap<string, T>;\n}\n", "Own<number>", "Deep<number>", containerFitDepthStop},
		{"anonymous", "type Own<T> = { readonly next: Own<{ readonly x: T }> | undefined; readonly m: Map<string, T> };\n" +
			"type Deep<T> = { readonly next: Deep<{ readonly x: T }> | undefined; readonly m: ReadonlyMap<string, T> };\n", "Own<number>", "Deep<number>", containerFitDepthStop},
		{"box_real_map", "interface Box<T> {\n\treadonly v: T;\n}\n", "{ a: { b: { c: { d: { e: Box<Map<string, number>> } } } } }", "{ readonly a: { readonly b: { readonly c: { readonly d: { readonly e: Box<ReadonlyMap<string, number>> } } } } }", ""},
		{"box_fake_map", containerFitBoundFakeMap + "interface Box<T> {\n\treadonly v: T;\n}\n", "{ a: { b: { c: { d: { e: Box<FakeMap> } } } } }", "{ readonly a: { readonly b: { readonly c: { readonly d: { readonly e: Box<ReadonlyMap<string, number>> } } } } }", containerFitStop},
		{"repeating", "interface Box<T> {\n\treadonly self: Box<T>;\n\treadonly m: Map<string, T>;\n}\n" +
			"interface View<T> {\n\treadonly self: View<T>;\n\treadonly m: ReadonlyMap<string, T>;\n}\n", "Box<number>", "View<number>", ""},
		{"node_predicate", "interface Node {\n\treadonly kind: number;\n\treadonly parent: Node | undefined;\n\treadonly locals: ReadonlyMap<string, Node> | undefined;\n}\n" +
			"interface Leaf extends Node {\n\treadonly kind: 1;\n}\n", "(nodes: readonly Node[], test: (node: Node) => node is Leaf) => readonly Leaf[]", "(nodes: readonly Node[], test: (node: Node) => node is Leaf) => readonly Node[]", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			source := probe.declarations + "declare const value: " + probe.value + ";\ndeclare const target: " + probe.target + ";\n"
			program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
			if err != nil {
				t.Fatal(err)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			declared := map[string]*checker.Type{}
			for _, statement := range file.Statements.Nodes {
				if statement.Kind != ast.KindVariableStatement {
					continue
				}
				for _, declaration := range statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
					declared[declaration.Name().Text()] = checked.GetTypeAtLocation(declaration.Name())
				}
			}
			l := &lowering{program: program, checker: checked, result: &ir.Program{}}
			var judged atomic.Int64
			containerFitPairsJudged.Store(&judged)
			defer containerFitPairsJudged.Store(nil)
			got := l.containerWhereOtherGoes(declared["value"], declared["target"], false)
			t.Logf("%s: %d pairs judged, stop %q", probe.name, judged.Load(), got)
			if got != probe.want {
				t.Fatalf("got the stop %q, want %q", got, probe.want)
			}
			if judged.Load() > 1000 {
				t.Fatalf("%d pairs judged, past the bound of 1000 that a walk typeNesting ends stays under", judged.Load())
			}
		})
	}
}
