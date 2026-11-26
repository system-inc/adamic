package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Probe for #hqamjaw, never for main: with entries_provenance.go's nullable-spread stop deleted, does each
// enumeration of a literal whose spread may be absent agree with Node on every backend?
func TestHqaProbe(t *testing.T) {
	programs := map[string]string{
		"keys.ts": `function f(present: boolean): { a: number } | undefined { return present ? { a: 2 } : undefined; }
for (const present of [false, true]) {
	const m = f(present);
	const o = { ...m, x: 1 };
	console.log(Object.keys(o).join(",") + " " + Object.entries(o).map((e) => e[0] + "=" + String(e[1])).join(",") + " " + Object.values(o).map((v) => String(v)).join(","));
}
`,
		"for_in.ts": `function f(present: boolean): { a: number } | undefined { return present ? { a: 2 } : undefined; }
for (const present of [false, true]) {
	const o = { ...f(present), x: 1 };
	const seen: string[] = [];
	for (const k in o) { seen.push(k); }
	console.log(seen.join(","));
}
`,
		"stringify.ts": `function f(present: boolean): { a: number; s: string } | undefined { return present ? { a: 2, s: "t" } : undefined; }
for (const present of [false, true]) {
	const o = { ...f(present), x: 1 };
	console.log(JSON.stringify(o) ?? "none");
}
`,
		"later_write.ts": `interface Wide { a?: number; x: number }
function f(present: boolean): { a: number } | undefined { return present ? { a: 2 } : undefined; }
const o: Wide = { ...f(false), x: 1 };
o.a = 5;
console.log(Object.keys(o).join(",") + " " + Object.values(o).map((v) => String(v)).join(","));
`,
		"nested_module.ts": `const base: { level: number } | undefined = undefined as { level: number } | undefined;
const table = { ...base, name: "n" };
const other = { one: 1, two: 2 };
console.log(Object.keys(table).join(",") + "|" + Object.keys(other).join(","));
`,
		"presence.ts": `function f(present: boolean): { a: number } | undefined { return present ? { a: 2 } : undefined; }
for (const present of [false, true]) {
	const o = { ...f(present), x: 1 };
	console.log(String("a" in o) + " " + String(Object.hasOwn(o, "a")) + " " + String(o.hasOwnProperty("a")));
}
`,
		"respread.ts": `function f(present: boolean): { a: number } | undefined { return present ? { a: 2 } : undefined; }
for (const present of [false, true]) {
	const o = { ...f(present), x: 1 };
	const copy = { ...o };
	console.log(Object.keys(copy).join(",") + " " + String("a" in copy));
}
`,
		"assign.ts": `function f(present: boolean): { a: number } | undefined { return present ? { a: 2 } : undefined; }
for (const present of [false, true]) {
	const o = { ...f(present), x: 1 };
	const target: { a?: number; x?: number } = {};
	Object.assign(target, o);
	console.log(Object.keys(target).join(",") + " " + String("a" in target));
}
`,
		"in_alone.ts": `function f(present: boolean): { a: number } | undefined { return present ? { a: 2 } : undefined; }
for (const present of [false, true]) {
	const o = { ...f(present), x: 1 };
	console.log(String("a" in o));
}
`,
		"own_property.ts": `function f(present: boolean): { a: number } | undefined { return present ? { a: 2 } : undefined; }
for (const present of [false, true]) {
	const o = { ...f(present), x: 1 };
	console.log(String(o.hasOwnProperty("a")));
}
`,
		"in_module.ts": `const base: { level: number } | undefined = undefined as { level: number } | undefined;
const table = { ...base, name: "n" };
console.log(String("level" in table) + " " + String(table.hasOwnProperty("level")));
`,
	}
	directory := t.TempDir()
	for _, name := range []string{"keys.ts", "for_in.ts", "stringify.ts", "later_write.ts", "nested_module.ts", "presence.ts", "respread.ts", "assign.ts", "in_alone.ts", "own_property.ts", "in_module.ts"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(directory, name)
			if err := os.WriteFile(path, []byte(programs[name]), 0o644); err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			program, err := lowered(t, path)
			if err != nil {
				fmt.Printf("HQA %s node=%q lowering: %v\n", name, node.stdout, err)
				return
			}
			sanitized, _ := nativelyUncached(t, program)
			release := releasedUncached(t, program)
			script := onJavaScriptBackend(t, program)
			for index, got := range []run{sanitized, release, script} {
				backend := []string{"sanitized", "release", "javascript"}[index]
				if difference := disagreement(node, got); difference != "" {
					fmt.Printf("HQA %s %s DISAGREES: %s node=%q got exit=%d stdout=%q stderr=%q\n", name, backend, difference, node.stdout, got.exitCode, got.stdout, got.stderr)
				} else {
					fmt.Printf("HQA %s %s agrees %q\n", name, backend, got.stdout)
				}
			}
		})
	}
}
