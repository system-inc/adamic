package lower

import "testing"

// A probe, never landed: which record programs the record stop still holds under step 2.
func TestZZRecordStopProbe(t *testing.T) {
	t.Parallel()
	head := "interface Node { name: string; next: Node[]; r: Record<string, Node[]>; rs: Record<string, Node[]>[] }\nconst a: Node = { name: 'a' + '', next: [], r: {}, rs: [] };\nconst b: Node = { name: 'b' + '', next: [a], r: {}, rs: [] };\na.next.push(b);\n"
	cases := []struct{ name, source string }{
		{"record-stop", head + "a.r = { k: a.next.slice() };\nconst got = a.r['k'];\nconsole.log(`${a.name} ${got === undefined ? -1 : got.length}`);\n"},
		{"cycle-stop", head + "const record: Record<string, Node[]> = { k: a.next.slice() };\na.rs = [record];\nconst got = a.rs[0];\nconsole.log(`${a.name} ${got === undefined ? -1 : 'some'}`);\n"},
		{"unknown-field", "interface Node { name: string; next: Node[]; u: Record<string, unknown> }\nconst a: Node = { name: 'a' + '', next: [], u: {} };\nconst b: Node = { name: 'b' + '', next: [a], u: {} };\na.next.push(b);\na.u = { k: a };\nconsole.log(`${a.name} ${Object.keys(a.u).length}`);\n"},
		{"object-field", "interface Node { name: string; next: Node[]; u: Record<string, object> }\nconst a: Node = { name: 'a' + '', next: [], u: {} };\nconst b: Node = { name: 'b' + '', next: [a], u: {} };\na.next.push(b);\na.u = { k: a };\nconsole.log(`${a.name} ${Object.keys(a.u).length}`);\n"},
		{"unknown-array", "interface Node { name: string; next: Node[]; us: Record<string, unknown>[] }\nconst a: Node = { name: 'a' + '', next: [], us: [] };\nconst b: Node = { name: 'b' + '', next: [a], us: [] };\na.next.push(b);\nconst record: Record<string, unknown> = { k: a };\na.us = [record];\nconsole.log(`${a.name} ${a.us.length}`);\n"},
		{"unknown-push", "interface Node { name: string; next: Node[]; us: Record<string, unknown>[] }\nconst a: Node = { name: 'a' + '', next: [], us: [] };\nconst b: Node = { name: 'b' + '', next: [a], us: [] };\na.next.push(b);\na.us.push({ k: a });\nconsole.log(`${a.name} ${a.us.length}`);\n"},
		{"object-array-union", "interface Node { name: string; next: Node[]; u: Record<string, object | Node[]> }\nconst a: Node = { name: 'a' + '', next: [], u: {} };\nconst b: Node = { name: 'b' + '', next: [a], u: {} };\na.next.push(b);\na.u = { k: a.next.slice() };\nconsole.log(`${a.name} ${Object.keys(a.u).length}`);\n"},
		{"cell-unknown", "interface Node { name: string; next: Node[] }\nconst a: Node = { name: 'a' + '', next: [] };\nconst b: Node = { name: 'b' + '', next: [a] };\na.next.push(b);\nfunction run(): number {\n  let table: Record<string, unknown> = {};\n  const size = (): number => (table['size'] === undefined ? 0 : 1);\n  table = { size: size };\n  return size();\n}\nconsole.log(`${a.name} ${run()}`);\n"},
	}
	for _, c := range cases {
		_, err := lowerSource(t, c.source)
		if err != nil {
			t.Logf("PROBE %s: STOPS: %v", c.name, err)
		} else {
			t.Logf("PROBE %s: LOWERS", c.name)
		}
	}
}
