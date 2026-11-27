package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// graphSharedAllocationBalanced lowers source as main.<extension>, and holds it to Node and to the counted
// build's balance; graph says whether it must have selected a graph region.
func graphSharedAllocationBalanced(t *testing.T, extension, source string, graph bool) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main."+extension)
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if graph && len(program.GraphTypes) == 0 {
		t.Fatal("the cycle selected no graph region")
	}
	t.Logf("%d graph types", len(program.GraphTypes))
	truth := onNode(t, path)
	got, _ := natively(t, program)
	if difference := disagreement(truth, got); difference != "" {
		t.Fatalf("native: %s; Node %q; native %q; stderr %q", difference, truth.stdout, got.stdout, got.stderr)
	}
	counted := filepath.Join(t.TempDir(), "counted")
	if err := native.Build(native.C(program), counted, native_runtime.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	ran := execute(t, counted)
	if report := leakcheck.Unbalanced(leakcheck.Run{Stdout: ran.stdout, Stderr: ran.stderr, ExitCode: ran.exitCode}); report != "" {
		t.Fatal(report)
	}
}

// Selection's sweeps relate two views by the finder's own relation, never the checker's (#nxb3xh1 head 6,
// mayShareAllocation): a known cycle still selects its graph region and frees it whole, a doubly linked list
// (graph_regions_list.a) and a tree whose children name their parent through an array, whose element type an
// array candidate holds.
func TestGraphSharedAllocationKeepsKnownCycles(t *testing.T) {
	t.Parallel()
	list, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/graph_regions_list.a"))
	if err != nil {
		t.Fatal(err)
	}
	tree := "interface Parent { readonly name: string; kids: Kid[] }\ninterface Kid { readonly label: string; parent: Parent | undefined }\nfunction run(): void {\n\tconst root: Parent = { name: `root${1}`, kids: [] };\n\tfor (let index = 0; index < 3; index++) {\n\t\tconst kid: Kid = { label: `kid${index}`, parent: undefined };\n\t\tkid.parent = root;\n\t\troot.kids.push(kid);\n\t}\n\tlet named = 0;\n\tfor (const kid of root.kids) {\n\t\tnamed += kid.parent === root ? 1 : 0;\n\t}\n\tconsole.log(`${root.kids.length} ${named}`);\n}\nrun();\n"
	for _, row := range []struct{ name, source string }{
		{"a doubly linked list", string(list)},
		{"a tree whose children name their parent", tree},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			graphSharedAllocationBalanced(t, "a", row.source, true)
		})
	}
}

// TestGraphViewDepthEdges' narrowed field with arms that share a name (Memory's condition A on #nxb3xh1):
// Y2 holds Y1's a, typed apart, so the checker relates neither way and viewAlsoTable links Y1 to Y2, while the
// finder's own relation for selection, which reads names alone, would relate them. viewAlsoTable keeps the
// checker's relation, so the edge stays and the cycle is found; widened to mayShareAllocation it would skip the
// pair as one allocation that reachLinks' shapes road, which walks by IsTypeAssignableTo, never walks, and the
// plain program's cycle would leak by the counted build (the mutant in internal/lower's
// shared_allocation.mutants.md). The rows are .ts, where the cast stays checked.
func TestGraphViewDepthSharedNames(t *testing.T) {
	t.Parallel()
	types := "interface Cell { held: Y1 | undefined }\ninterface Y1 { a: number[] }\ninterface Y2 { a: string; b: Cell | undefined; c: number }\n"
	anchor := "interface G { g: G | undefined }\nconst anchor: G = { g: undefined };\nanchor.g = anchor;\n"
	narrowed := "interface Base { readonly kind: number; readonly f: Y1 | Y2 }\ninterface Narrow extends Base { readonly kind: 1; readonly f: Y1 }\nfunction make(cell: Cell): Base {\n\tconst y2: Y2 = { a: `y${2}`, b: cell, c: 7 };\n\treturn { kind: 1, f: y2 };\n}\n"
	run := "function run(): void {\n\tconst cell: Cell = { held: undefined };\n\tconst base = make(cell);\n\tif (base.kind === 1) {\n\t\tconst n = base as Narrow;\n\t\tconst y1: Y1 = n.f;\n\t\tcell.held = y1;\n\t}\n\tconsole.log(`${cell.held === undefined ? \"none\" : \"held\"}`);\n}\nrun();\n"
	for _, row := range []struct{ name, source string }{
		{"a narrowed field sharing a name, in a graph program", types + anchor + narrowed + run},
		{"a narrowed field sharing a name, in a plain program", types + narrowed + run},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			graphSharedAllocationBalanced(t, "ts", row.source, false)
		})
	}
}
