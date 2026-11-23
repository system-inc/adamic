package lower

import "testing"

// Temporary probe (#1atvfp5, Language's sibling downcast across two modules): never landed.
func TestEnumerationReachSiblingProbe(t *testing.T) {
	const held = "export type Shared = { b: number };\nconst literal = { b: 2, get a(): number { return 1; } };\nexport const shared: Shared = literal;\n"
	for _, row := range []struct{ name, main string }{
		{"values", "import { type Shared, shared } from './held.ts';\ntype Wanted = { b: number; c: number };\nconst wanted = shared as Wanted;\nfunction read(view: Wanted): number[] { return Object.values(view); }\nconsole.log(read(wanted).join(','));\n"},
		{"entries", "import { type Shared, shared } from './held.ts';\ntype Wanted = { b: number; c: number };\nconst wanted = shared as Wanted;\nfor (const [key, value] of Object.entries(wanted)) console.log(key + '=' + String(value));\n"},
		{"json", "import { encodeJson } from 'adamic';\nimport { type Shared, shared } from './held.ts';\ntype Wanted = { b: number; c: number };\nconst wanted = shared as Wanted;\nconsole.log(encodeJson<Wanted>(wanted));\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			path, program, err := enumerationReachLower(t, ".ts", held, row.main)
			t.Logf("sibling probe %s lowered: err=%v", row.name, err)
			if err != nil {
				return
			}
			code, out, errs := brandedNative(t, program)
			t.Logf("sibling probe %s native: exit=%d stdout=%q stderr=%q", row.name, code, out, errs)
			agreeEntry(t, path, program, 0, nil, nil)
			t.Logf("sibling probe %s: JavaScript agreed with Node", row.name)
		})
	}
}
