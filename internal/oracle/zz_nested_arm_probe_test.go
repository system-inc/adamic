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

// A probe for #9489gt0, never landed: a downcast view tests its top arm only, and a field the target narrows
// (Base.f: Y1 | Y2 seen as Narrow.f: Y1) can hold a Y2 under Y1's view. Each row prints whether it lowers, the
// stop's words if it doesn't, whether native agrees with Node, and the counted build's balance.
func TestNestedArmProbe(t *testing.T) {
	t.Parallel()
	types := "interface Cell { held: Y1 | undefined }\ninterface Y1 { a: number[] }\ninterface Y2 { b: Cell | undefined; c: number }\ninterface Base { readonly kind: number; readonly f: Y1 | Y2 }\ninterface Narrow extends Base { readonly kind: 1; readonly f: Y1 }\n"
	anchor := "interface G { g: G | undefined }\nconst anchor: G = { g: undefined };\nanchor.g = anchor;\n"
	lie := "function make(cell: Cell): Base {\n\tconst y2: Y2 = { b: cell, c: 7 };\n\treturn { kind: 1, f: y2 };\n}\n"
	store := "function run(): void {\n\tconst cell: Cell = { held: undefined };\n\tconst base = make(cell);\n\tif (base.kind === 1) {\n\t\tconst n = base as Narrow;\n\t\tconst y1: Y1 = n.f;\n\t\tcell.held = y1;\n\t}\n\tconsole.log(`${cell.held === undefined ? \"none\" : \"held\"}`);\n}\nrun();\n"
	read := "function run(): void {\n\tconst cell: Cell = { held: undefined };\n\tconst base = make(cell);\n\tif (base.kind === 1) {\n\t\tconst n = base as Narrow;\n\t\tconsole.log(`${n.f.a.length}`);\n\t}\n}\nrun();\n"
	class := "interface Cell { held: Y1 | undefined }\ninterface Y1 { a: number[] }\ninterface Y2 { b: Cell | undefined; c: number }\nclass Base {\n\tconstructor(public f: Y1 | Y2) {}\n\tsetF(value: Y1 | Y2): void {\n\t\tthis.f = value;\n\t}\n}\nclass Narrow extends Base {\n\tdeclare f: Y1;\n}\n"
	classStore := "function run(): void {\n\tconst cell: Cell = { held: undefined };\n\tconst n = new Narrow({ a: [1] });\n\tconst y2: Y2 = { b: cell, c: 7 };\n\tn.setF(y2);\n\tconst y1: Y1 = n.f;\n\tcell.held = y1;\n\tconsole.log(`${cell.held === undefined ? \"none\" : \"held\"}`);\n}\nrun();\n"
	cases := []struct{ name, source string }{
		{"ts: lie stored, graph program", types + anchor + lie + store},
		{"ts: lie stored, no anchor", types + lie + store},
		{"ts: lie read, graph program", types + anchor + lie + read},
		{"class: a parent method writes the narrowed field, stored, graph program", class + anchor + classStore},
		{"class: a parent method writes the narrowed field, stored, no anchor", class + classStore},
		{"ts class: a parent method writes the narrowed field, stored, graph program", class + anchor + classStore},
		{"ts: lie read, no anchor", types + lie + read},
		{"lie stored, graph program", types + anchor + lie + store},
		{"lie stored, no anchor", types + lie + store},
		{"lie read, graph program", types + anchor + lie + read},
		{"lie read, no anchor", types + lie + read},
		{"control: the Y2 slot seen whole", "interface Cell { held: Y1 | Y2 | undefined }\ninterface Y1 { a: number[] }\ninterface Y2 { b: Cell | undefined; c: number }\n" + anchor + "function run(): void {\n\tconst cell: Cell = { held: undefined };\n\tconst y2: Y2 = { b: cell, c: 7 };\n\tcell.held = y2;\n\tconsole.log(`${cell.held === undefined ? \"none\" : \"held\"}`);\n}\nrun();\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			name := "main.a"
			if strings.HasPrefix(c.name, "ts") {
				name = "main.ts"
			}
			path := filepath.Join(t.TempDir(), name)
			if err := os.WriteFile(path, []byte(c.source), 0600); err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			t.Logf("PROBE %s: node stdout %q exit %d stderr %.200q", c.name, truth.stdout, truth.exitCode, truth.stderr)
			program, err := lowered(t, path)
			if err != nil {
				t.Logf("PROBE %s: STOP %v", c.name, err)
				return
			}
			got, _ := natively(t, program)
			t.Logf("PROBE %s: native stdout %q exit %d stderr %.300q; disagreement %q", c.name, got.stdout, got.exitCode, got.stderr, disagreement(truth, got))
			counted := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(native.C(program), counted, native_runtime.Options{Count: true}); err != nil {
				t.Logf("PROBE %s: counted build failed %v", c.name, err)
				return
			}
			ran := execute(t, counted)
			report := leakcheck.Unbalanced(leakcheck.Run{Stdout: ran.stdout, Stderr: ran.stderr, ExitCode: ran.exitCode})
			t.Logf("PROBE %s: counted exit %d, unbalanced %q", c.name, ran.exitCode, report)
		})
	}
}
