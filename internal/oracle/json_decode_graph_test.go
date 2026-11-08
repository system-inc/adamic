package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// The decoder makes its objects and arrays in the runtime with no graph plan. In a graph program, a
// decoded value that holds a container could be closed into a cycle outside every region: on main
// 72f23939, `b.kids.push(a)` over a decoded Tree left 6 of 51 values unfreed by the counted build,
// with LeakSanitizer reporting nothing (#bvsjdte). Until the decoder takes the plan (2b), such a
// decode is NotYet in a graph program, and only there: the boundary runs only when the program seeds
// a graph region (graph_types.go's graphTypes returns before it when nothing seeds).
//
// Mutant: drop the ir.JSONDecode case from graphAdoptionBoundary, and the first two cases lower;
// the first then leaks by the counted build (the probe's 45 of 51).
const jsonDecodeGraphTree = `import { decodeJson } from 'adamic';
interface Tree {
	name: string;
	kids: Tree[];
}
`

func TestJSONDecodeInAGraphProgram(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, source string
		stopped      bool
	}{
		// The leak itself: a decoded Tree given a cycle.
		{"decoded value on a cycle", jsonDecodeGraphTree + `const result = decodeJson<Tree>('{"name":"a","kids":[{"name":"b","kids":[]}]}');
if (result.kind !== 'Error') {
	const a = result.value;
	const b = a.kids[0];
	if (b !== undefined) {
		b.kids.push(a);
		console.log(` + "`${a.name} ${b.name} ${b.kids.length}`" + `);
	}
}
`, true},
		// A literal Tree cycle seeds the program; the decode never joins it, and still stops, since a
		// seed anywhere makes every decoded Tree a value the program could close a cycle through.
		{"seeded elsewhere", jsonDecodeGraphTree + `const x: Tree = { name: 'x' + '', kids: [] };
const y: Tree = { name: 'y' + '', kids: [x] };
x.kids.push(y);
const result = decodeJson<Tree>('{"name":"a","kids":[]}');
console.log(result.kind === 'Error' ? result.message : ` + "`${result.value.name} ${x.kids.length}`" + `);
`, true},
		// The same decode with no cycle anywhere seeds nothing, so the program isn't a graph program.
		{"no graph program", jsonDecodeGraphTree + `const result = decodeJson<Tree>('{"name":"a","kids":[{"name":"b","kids":[]}]}');
if (result.kind !== 'Error') {
	const b = result.value.kids[0];
	console.log(` + "`${result.value.name} ${b === undefined ? '-' : b.name}`" + `);
}
`, false},
		// A decoded value that holds no container holds nothing a cycle could return through, so a
		// graph program may decode it.
		{"flat decode in a graph program", jsonDecodeGraphTree + `const x: Tree = { name: 'x' + '', kids: [] };
const y: Tree = { name: 'y' + '', kids: [x] };
x.kids.push(y);
const result = decodeJson<{ readonly name: string; readonly count: number }>('{"name":"a","count":2}');
console.log(result.kind === 'Error' ? result.message : ` + "`${result.value.name} ${result.value.count} ${x.kids.length}`" + `);
`, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.a")
			if err := os.WriteFile(path, []byte(c.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if c.stopped {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "decodeJson<Tree> in a graph program") {
					t.Fatalf("got %v, want the decode stopped in a graph program", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			// What lowers runs as Node does and frees everything, judged by the counted build, which is
			// exact where LeakSanitizer misses cycles this small.
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
