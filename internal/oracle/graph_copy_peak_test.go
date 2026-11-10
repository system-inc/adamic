package oracle

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// A planned copy of a long-lived graph's array is a region of one (#5nrs3q2): dropped, it frees at
// its last release. Merged into the graph at adoption, as main did, each copy a loop dropped was held
// until the graph died: a.next.slice() peaked at 18 for 10 turns and 1008 for 1000 (main 0816e185).
// Stored into the graph, the copy joins it, and everything still frees at the end.
//
// Mutant: adamic_graph_adopt_owned joins its member children again (graph_regions.c), and the loop's
// peak grows with its count.
const graphCopyNode = `interface Node { name: string; next: Node[]; kids: Node[][] }
const a: Node = { name: 'a' + String(1), next: [], kids: [] };
const b: Node = { name: 'b' + String(2), next: [a], kids: [] };
a.next.push(b);
`

var graphCopyPeak = regexp.MustCompile(`peak (\d+)`)

func graphCopyRun(t *testing.T, source string) (peak int) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
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
	match := graphCopyPeak.FindSubmatch(ran.stderr)
	if match == nil {
		t.Fatalf("no counts line: %s", ran.stderr)
	}
	peak, err = strconv.Atoi(string(match[1]))
	if err != nil {
		t.Fatal(err)
	}
	return peak
}

func TestGraphCopyDroppedInALoop(t *testing.T) {
	t.Parallel()
	loop := func(turns int) string {
		return graphCopyNode + `let seen = 0;
for (let i = 0; i < ` + strconv.Itoa(turns) + `; i++) {
  const copy = a.next.slice();
  seen += copy.length;
}
console.log(` + "`${a.name} ${seen}`" + `);
`
	}
	few, many := graphCopyRun(t, loop(10)), graphCopyRun(t, loop(1000))
	if many != few {
		t.Fatalf("peak %d at 10 turns, %d at 1000: a dropped copy outlives its turn", few, many)
	}
}

func TestGraphCopyStored(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		// Stored into the graph, the copy joins it; popped back out, it's the statement's.
		"stored and popped": graphCopyNode + `a.kids.push(a.next.slice());
const popped = a.kids.pop();
console.log(` + "`${a.name} ${popped === undefined ? -1 : popped.length}`" + `);
`,
		// Language's pin (#5nrs3q2): a dropped copy takes a member of another graph, then that member
		// stores the copy, closing a cycle through it.
		"pushed into, then closed": graphCopyNode + `const c: Node = { name: 'c' + String(3), next: [], kids: [] };
c.next.push(c);
const copy = a.next.slice();
copy.push(c);
c.kids.push(copy);
console.log(` + "`${a.name} ${c.kids.length} ${copy.length}`" + `);
`,
		// The graph's last outside hold is a copy's counted edge.
		"outlived by its copy": `interface Node { name: string; next: Node[] }
function take(): Node[] {
  const a: Node = { name: 'a' + String(1), next: [] };
  const b: Node = { name: 'b' + String(2), next: [a] };
  a.next.push(b);
  return a.next.slice();
}
const kept = take();
const first = kept[0];
console.log(first === undefined ? '-' : first.name + ' ' + String(first.next.length));
`,
		// Sorting a copy writes it back through take, which merges it into the graph.
		"sorted": graphCopyNode + `const copy = [b, a].slice();
copy.sort((x, y) => (x.name < y.name ? -1 : 1));
const first = copy[0];
console.log(first === undefined ? '-' : first.name);
`,
	}
	for name, source := range cases {
		t.Run(strings.ReplaceAll(name, " ", "_"), func(t *testing.T) {
			t.Parallel()
			graphCopyRun(t, source)
		})
	}
}
