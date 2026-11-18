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
// anonymous or mapped type by its symbol (anonymous, mapped). The controls: a real Map in a Box<T> five
// fields down fits, the same with a FakeMap stops at the real check, and a predicate over a non-generic
// self-referential node, tsc's shape behind most of the 35 sites, fits.
// Mutant: the gate ignoring typeNesting's refusal runs every growing shape to containerFitPairCap, so each
// reddens on its stop ("a fit walk past 4096 distinct pairs") and on the pair count.
func TestContainerFitBoundsGrowingGenerics(t *testing.T) {
	const ownDeep = "interface Own<T> {\n\treadonly next: Own<Own<T>> | undefined;\n\treadonly m: Map<string, T>;\n}\n" +
		"interface Deep<T> {\n\treadonly next: Deep<Deep<T>> | undefined;\n\treadonly m: ReadonlyMap<string, T>;\n}\n"
	for _, probe := range []struct{ name, declarations, value, target, want string }{
		{"reference", ownDeep, "Own<number>", "Deep<number>", containerFitDepthStop},
		{"tuple", "interface Own<T> {\n\treadonly next: Own<[T, T]> | undefined;\n\treadonly m: Map<string, T>;\n}\n" +
			"interface Deep<T> {\n\treadonly next: Deep<[T, T]> | undefined;\n\treadonly m: ReadonlyMap<string, T>;\n}\n", "Own<number>", "Deep<number>", containerFitDepthStop},
		{"anonymous", "type Own<T> = { readonly next: Own<{ readonly x: T }> | undefined; readonly m: Map<string, T> };\n" +
			"type Deep<T> = { readonly next: Deep<{ readonly x: T }> | undefined; readonly m: ReadonlyMap<string, T> };\n", "Own<number>", "Deep<number>", containerFitDepthStop},
		{"mapped", "type Own<T> = { readonly [K in keyof T]: Own<{ readonly v: T[K] }> } & { readonly m: Map<string, T> };\n" +
			"type Deep<T> = { readonly [K in keyof T]: Deep<{ readonly v: T[K] }> } & { readonly m: ReadonlyMap<string, T> };\n", "Own<{ readonly a: number }>", "Deep<{ readonly a: number }>", containerFitDepthStop},
		{"box_real_map", "interface Box<T> {\n\treadonly v: T;\n}\n", "{ a: { b: { c: { d: { e: Box<Map<string, number>> } } } } }", "{ readonly a: { readonly b: { readonly c: { readonly d: { readonly e: Box<ReadonlyMap<string, number>> } } } } }", ""},
		{"box_fake_map", containerFitBoundFakeMap + "interface Box<T> {\n\treadonly v: T;\n}\n", "{ a: { b: { c: { d: { e: Box<FakeMap> } } } } }", "{ readonly a: { readonly b: { readonly c: { readonly d: { readonly e: Box<ReadonlyMap<string, number>> } } } } }", containerFitStop},
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
