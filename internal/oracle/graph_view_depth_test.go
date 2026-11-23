package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// A checked view tests at most its top, so a field the target narrows hands out what the source's field holds
// (#9489gt0): Base.f: Y1 | Y2 seen as Narrow.f: Y1, through interfaceCast's kind test, gives a Y2 as a Y1, and a
// Y1 slot it's stored in holds a Y2 that holds the slot's own holder. The cycle finder took two types for one
// allocation only when they relate, so the cycle went unseen and leaked 2 by the counted build, in a graph program
// and in a plain one. Each view's unrelated arms beneath its top are edges reachLinks and graphLinks follow
// (viewAlsoTable, Lower's ruling), so each row is one graph region freed whole: held to Node, and balanced by the
// counted build. The rows are .ts, where the cast stays checked; .a refuses the downcast.
//
// An element read through a view (Narrow.items: readonly Y1[] over Base's readonly (Y1 | Y2)[]) stops first
// today, as a checked view's array read through an index, so that row pins the stop by its words. viewAlsoTable
// pairs elements already; when the stop lifts the row goes red here, and becomes a counted-build row.
//
// Mutant: viewAlsoTable walking no view, and every field row leaks by the counted build.
func TestGraphViewDepthEdges(t *testing.T) {
	t.Parallel()
	types := "interface Cell { held: Y1 | undefined }\ninterface Y1 { a: number[] }\ninterface Y2 { b: Cell | undefined; c: number }\n"
	anchor := "interface G { g: G | undefined }\nconst anchor: G = { g: undefined };\nanchor.g = anchor;\n"
	narrowed := "interface Base { readonly kind: number; readonly f: Y1 | Y2 }\ninterface Narrow extends Base { readonly kind: 1; readonly f: Y1 }\nfunction make(cell: Cell): Base {\n\tconst y2: Y2 = { b: cell, c: 7 };\n\treturn { kind: 1, f: y2 };\n}\n"
	run := func(read string) string {
		return "function run(): void {\n\tconst cell: Cell = { held: undefined };\n\tconst base = make(cell);\n\tif (base.kind === 1) {\n\t\tconst n = base as Narrow;\n\t\tconst y1: Y1 = " + read + ";\n\t\tcell.held = y1;\n\t}\n\tconsole.log(`${cell.held === undefined ? \"none\" : \"held\"}`);\n}\nrun();\n"
	}
	twoDeep := "interface Mid { readonly g: Y1 | Y2 }\ninterface MidNarrow { readonly g: Y1 }\ninterface Base { readonly kind: number; readonly f: Mid }\ninterface Narrow extends Base { readonly kind: 1; readonly f: MidNarrow }\nfunction make(cell: Cell): Base {\n\tconst y2: Y2 = { b: cell, c: 7 };\n\tconst mid: Mid = { g: y2 };\n\treturn { kind: 1, f: mid };\n}\n"
	element := "interface Base { readonly kind: number; readonly items: readonly (Y1 | Y2)[] }\ninterface Narrow extends Base { readonly kind: 1; readonly items: readonly Y1[] }\nfunction make(cell: Cell): Base {\n\tconst y2: Y2 = { b: cell, c: 7 };\n\treturn { kind: 1, items: [y2] };\n}\n"
	cases := []struct{ name, source, stop string }{
		{"a narrowed field, in a graph program", types + anchor + narrowed + run("n.f"), ""},
		{"a narrowed field, in a plain program", types + narrowed + run("n.f"), ""},
		{"a field narrowed two deep", types + twoDeep + run("n.f.g"), ""},
		{"a narrowed element, stopped at the read", types + element + "function run(): void {\n\tconst cell: Cell = { held: undefined };\n\tconst base = make(cell);\n\tif (base.kind === 1) {\n\t\tconst n = base as Narrow;\n\t\tconst item = n.items[0];\n\t\tif (item !== undefined) {\n\t\t\tcell.held = item;\n\t\t}\n\t}\n\tconsole.log(`${cell.held === undefined ? \"none\" : \"held\"}`);\n}\nrun();\n", "a checked view's array read through index (not checked yet)"},
		{"the control, the Y2 slot seen whole", "interface Cell { held: Y1 | Y2 | undefined }\ninterface Y1 { a: number[] }\ninterface Y2 { b: Cell | undefined; c: number }\nfunction run(): void {\n\tconst cell: Cell = { held: undefined };\n\tconst y2: Y2 = { b: cell, c: 7 };\n\tcell.held = y2;\n\tconsole.log(`${cell.held === undefined ? \"none\" : \"held\"}`);\n}\nrun();\n", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, []byte(c.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if c.stop != "" {
				if err == nil || !strings.Contains(err.Error(), c.stop) {
					t.Fatalf("want the stop %q, got %v", c.stop, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
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
		})
	}
}
