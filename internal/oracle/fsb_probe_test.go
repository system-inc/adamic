package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

const fsbLets = "let first: number | string = 1;\nlet second: number | string = 'two';\n"

var fsbProbes = []struct{ name, source string }{
	{"union_swap", fsbLets + "[first, second] = [second, first];\nconsole.log(`${first} ${second}`);\n"},
	{"number_swap", "let first = 1;\nlet second = 2;\n[first, second] = [second, first];\nconsole.log(`${first} ${second}`);\n"},
	{"rotation", fsbLets + "let third: number | string = 3;\n[first, second, third] = [third, first, second];\nconsole.log(`${first} ${second} ${third}`);\n"},
	{"box_swap", "interface Box {\n\tlabel: string;\n}\nfunction show(value: string | Box): string {\n\treturn typeof value === 'string' ? value : value.label;\n}\nlet first: string | Box = { label: 'boxed' };\nlet second: string | Box = 'plain';\n[first, second] = [second, first];\nconsole.log(`${show(first)} ${show(second)}`);\n"},
	{"tuple_local", fsbLets + "const tuple: [number | string, number | string] = [second, first];\n[first, second] = tuple;\nconsole.log(`${first} ${second}`);\n"},
	{"nested", fsbLets + "[[first], second] = [[second], first];\nconsole.log(`${first} ${second}`);\n"},
	{"defaulted", fsbLets + "[first = 0, second] = [second, first];\nconsole.log(`${first} ${second}`);\n"},
	{"rest", fsbLets + "let third: number | string = 3;\nlet rest: (number | string)[] = [];\n[first, ...rest] = [second, third, first];\nconsole.log(`${first} ${rest.join(',')}`);\n"},
	{"object_twin", fsbLets + "({ x: first, y: second } = { x: second, y: first });\nconsole.log(`${first} ${second}`);\n"},
	{"declaration", fsbLets + "const [p, q]: [number | string, number | string] = [second, first];\nconsole.log(`${p} ${q}`);\n"},
	{"parameter", fsbLets + "function show([a, b]: [number | string, number | string]): string {\n\treturn `${a} ${b}`;\n}\nconsole.log(show([second, first]));\n"},
}

func TestFsbProbe(t *testing.T) {
	directory := t.TempDir()
	for _, probe := range fsbProbes {
		t.Run(probe.name, func(t *testing.T) {
			path := filepath.Join(directory, probe.name+".a")
			if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			t.Logf("PROBE %s node exit=%d stdout=%q stderr=%q", probe.name, node.exitCode, node.stdout, node.stderr)
			program, err := lowered(t, path)
			if err != nil {
				t.Logf("PROBE %s lowered: %v", probe.name, err)
				return
			}
			released := releasedUncached(t, program)
			t.Logf("PROBE %s native exit=%d stdout=%q stderr=%q", probe.name, released.exitCode, released.stdout, released.stderr)
			sanitized, _ := nativelyUncached(t, program)
			t.Logf("PROBE %s sanitized exit=%d stdout=%q stderr=%.300q", probe.name, sanitized.exitCode, sanitized.stdout, sanitized.stderr)
			script := onJavaScriptBackend(t, program)
			t.Logf("PROBE %s javascript exit=%d stdout=%q stderr=%.300q", probe.name, script.exitCode, script.stdout, script.stderr)
		})
	}
}
