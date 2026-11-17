package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// p1 of #hny64km (Memory's): graphFlows' collect now stops at a demanded local some node writes with a
// value it doesn't follow (internal/lower/graph_flow.go's graphLocalWrites), and a for-of binding over
// a list of nodes isn't one. Its elements were planned where they were stored, and a fresh literal the
// body assigns it is a source as any assignment's value is, so the loop still lowers, the literal is
// planned into the cycle it's stored in, and the counted build frees everything. Held to Node and the
// counted build, over a local array and over an array literal, whose elements are the binding's
// sources too. A counted build runs here, not in internal/lower, which builds no binaries.
func TestGraphFlowReboundBinding(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ name, iterable string }{
		{"over a local array", "cells"},
		{"over an array literal", "[a, b]"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			source := `interface Cell { name: string; next: Cell | undefined }
const a: Cell = { name: 'a' + '', next: undefined };
const b: Cell = { name: 'b' + '', next: a };
a.next = b; // a cycle: the program is a graph program
const cells: Cell[] = [a, b];
for (let cell of ` + c.iterable + `) {
	cell = { name: cell.name + '!', next: a };
	b.next = cell;
}
let names = '';
let at: Cell | undefined = a;
for (let i = 0; i < 4 && at !== undefined; i++) {
	names += at.name + ' ';
	at = at.next;
}
console.log(names + String(cells.length));
`
			path := filepath.Join(t.TempDir(), "main.a")
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if len(program.GraphTypes) == 0 {
				t.Fatal("no graph types selected, so the loop measures nothing of graph regions")
			}
			truth := onNode(t, path)
			got, _ := natively(t, program)
			if difference := disagreement(truth, got); difference != "" {
				t.Fatalf("native: %s; Node %q; native %q; stderr %q", difference, truth.stdout, got.stdout, got.stderr)
			}
			counted := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(native.C(program), counted, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			ran := execute(t, counted)
			if report := leakcheck.Unbalanced(leakcheck.Run{Stdout: ran.stdout, Stderr: ran.stderr, ExitCode: ran.exitCode}); report != "" {
				t.Fatal(report)
			}
		})
	}
}
