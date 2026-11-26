package lower

import "testing"

// A scratch probe for #nxb3xh1 head 2, never landed: which source programs reach checkRangeUnionReads' stop.
func TestStopWhereProbeRangeUnions(t *testing.T) {
	t.Parallel()
	prefix := rangeTextRange + rangeNodeArray
	for _, probe := range []struct{ name, source string }{
		{"keys in a function", "export function count(range: TextRange): number { return Object.keys(range).length; }\nconsole.log(String(count({ pos: 1, end: 2 })));\n"},
		{"entries in a function", "export function count(range: TextRange): number { return Object.entries(range).length; }\nconsole.log(String(count({ pos: 1, end: 2 })));\n"},
		{"hasOwn in a function", "export function owns(range: TextRange): boolean { return Object.hasOwn(range, 'pos'); }\nconsole.log(String(owns({ pos: 1, end: 2 })));\n"},
		{"for-in in a function", "export function count(range: TextRange): number { let n = 0; for (const key in range) { n += key.length; } return n; }\nconsole.log(String(count({ pos: 1, end: 2 })));\n"},
		{"assign in a function", "export function copy(range: TextRange): number { const made = Object.assign({ pos: 0, end: 0 }, range); return made.end; }\nconsole.log(String(copy({ pos: 1, end: 2 })));\n"},
		{"keys in main", "const range: TextRange = { pos: 1, end: 2 };\nconsole.log(String(Object.keys(range).length));\n"},
		{"for-in in main", "const range: TextRange = { pos: 1, end: 2 };\nlet n = 0;\nfor (const key in range) { n += key.length; }\nconsole.log(String(n));\n"},
		{"hasOwn in main", "const range: TextRange = { pos: 1, end: 2 };\nconsole.log(String(Object.hasOwn(range, 'pos')));\n"},
	} {
		program, err := lowerFirstLinkTS(t, prefix+probe.source, true)
		if err != nil {
			t.Logf("probe %s: traps: error %v", probe.name, err)
		} else {
			t.Logf("probe %s: traps: rows %+v abandoned %+v", probe.name, firstLinkLocalSites(program), program.TrapJudgmentsAbandoned)
		}
		_, err = lowerFirstLinkTS(t, prefix+probe.source, false)
		t.Logf("probe %s: no traps: %s", probe.name, firstLinkFromFile(err, "main.ts"))
	}
}
