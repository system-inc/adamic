package oracle

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// The decoder once made its objects and arrays in the runtime with no graph plan. In a graph program, a
// decoded value that held a container could be closed into a cycle outside every region: on main
// 72f23939, `b.kids.push(a)` over a decoded Tree left 6 of 51 values unfreed by the counted build,
// with LeakSanitizer reporting nothing (#bvsjdte). 2a stopped such a decode; now each schema node for an
// object, array or tuple carries its allocation identity, the emitter writes whether the cycle proof
// placed that type in a graph region, and the decoder adopts what it makes for a flagged node
// (#3wv7dw5). Each case is judged by the counted build, exact where LeakSanitizer misses cycles this
// small.
//
// Mutant: write every node's flag false in the emitter (library_json_decode.go), and the decoded cycle
// leaks by the counted build, as the nested schema's does.
const jsonDecodeGraphTree = `import { decodeJson } from 'adamic';
interface Tree {
	name: string;
	kids: Tree[];
}
`

// jsonDecodeGraphNode is one node of an emitted decoder schema: its kind, its expected type's text,
// and its graph flag, the initializer's last member (adamic_decode_node in json_decode.h).
var jsonDecodeGraphNode = regexp.MustCompile(`\{"(\w+)", -?\d+, "((?:[^"\\]|\\.)*)", [^{}"]*, "(?:[^"\\]|\\.)*", (true|false)\}`)

func TestJSONDecodeInAGraphProgram(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, source string
		// flags is each object or array node's graph flag in the emitted schema, by its expected type.
		flags map[string]bool
	}{
		// The leak 2a stopped: a decoded Tree given a cycle, now held by a graph region.
		{"decoded value on a cycle", jsonDecodeGraphTree + `const result = decodeJson<Tree>('{"name":"a","kids":[{"name":"b","kids":[]}]}');
if (result.kind !== 'Error') {
	const a = result.value;
	const b = a.kids[0];
	if (b !== undefined) {
		b.kids.push(a);
		console.log(` + "`${a.name} ${b.name} ${b.kids.length}`" + `);
	}
}
`, map[string]bool{"Tree": true, "Tree[]": true}},
		// A literal Tree cycle seeds the program; the decode never joins it, and its Tree is adopted all
		// the same, since a seed anywhere makes every decoded Tree a value the program could close a
		// cycle through.
		{"seeded elsewhere", jsonDecodeGraphTree + `const x: Tree = { name: 'x' + '', kids: [] };
const y: Tree = { name: 'y' + '', kids: [x] };
x.kids.push(y);
const result = decodeJson<Tree>('{"name":"a","kids":[]}');
console.log(result.kind === 'Error' ? result.message : ` + "`${result.value.name} ${x.kids.length}`" + `);
`, map[string]bool{"Tree": true, "Tree[]": true}},
		// The same decode with no cycle anywhere seeds nothing, so the program isn't a graph program and
		// no node is flagged.
		{"no graph program", jsonDecodeGraphTree + `const result = decodeJson<Tree>('{"name":"a","kids":[{"name":"b","kids":[]}]}');
if (result.kind !== 'Error') {
	const b = result.value.kids[0];
	console.log(` + "`${result.value.name} ${b === undefined ? '-' : b.name}`" + `);
}
`, map[string]bool{"Tree": false, "Tree[]": false}},
		// A decoded value that holds no container is no graph type, so a graph program decodes it
		// unflagged.
		{"flat decode in a graph program", jsonDecodeGraphTree + `interface Count {
	readonly name: string;
	readonly count: number;
}
const x: Tree = { name: 'x' + '', kids: [] };
const y: Tree = { name: 'y' + '', kids: [x] };
x.kids.push(y);
const result = decodeJson<Count>('{"name":"a","count":2}');
console.log(result.kind === 'Error' ? result.message : ` + "`${result.value.name} ${result.value.count} ${x.kids.length}`" + `);
`, map[string]bool{"Count": false}},
		// A graph type's array inside a type that isn't one: Forest and its string[] are made as before,
		// and only the Trees and their arrays join a region, the cycle closed among them freed with it.
		{"nested schema", jsonDecodeGraphTree + `interface Forest {
	title: string;
	trees: Tree[];
	tags: string[];
}
const result = decodeJson<Forest>('{"title":"f","trees":[{"name":"a","kids":[{"name":"b","kids":[]}]}],"tags":["x"]}');
if (result.kind !== 'Error') {
	const forest = result.value;
	const a = forest.trees[0];
	const b = a === undefined ? undefined : a.kids[0];
	if (a !== undefined && b !== undefined) {
		b.kids.push(a);
		console.log(` + "`${forest.title} ${a.name} ${b.name} ${b.kids.length} ${forest.tags.length}`" + `);
	}
}
`, map[string]bool{"Forest": false, "Tree[]": true, "Tree": true, "string[]": false}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.a")
			if err := os.WriteFile(path, []byte(c.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			code := native.C(program)
			flags := map[string]bool{}
			for _, node := range jsonDecodeGraphNode.FindAllStringSubmatch(code, -1) {
				if node[1] == "object" || node[1] == "array" {
					flags[node[2]] = node[3] == "true"
				}
			}
			for expected, want := range c.flags {
				got, found := flags[expected]
				if !found || got != want {
					t.Errorf("the schema's %s node: graph %t (found %t), want %t; every flag %v", expected, got, found, want, flags)
				}
			}
			if len(flags) != len(c.flags) {
				t.Errorf("the schema's flags %v, want exactly %v", flags, c.flags)
			}
			truth := onNode(t, path)
			sanitized, _ := natively(t, program)
			for backend, got := range map[string]run{"sanitized": sanitized, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Errorf("%s: %s; Node %q; got %q; stderr %q", backend, difference, truth.stdout, got.stdout, got.stderr)
				}
			}
			counted := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(code, counted, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			ran := execute(t, counted)
			t.Logf("%s: %s", c.name, leakcheck.CountsLine.Find(ran.stderr))
			if report := leakcheck.Unbalanced(leakcheck.Run{Stdout: ran.stdout, Stderr: ran.stderr, ExitCode: ran.exitCode}); report != "" {
				t.Fatal(report)
			}
		})
	}
}
