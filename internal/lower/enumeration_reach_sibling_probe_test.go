package lower

import "testing"

// Temporary probe (#1atvfp5, Language's sibling downcast across two modules): never landed.
func TestEnumerationReachSiblingProbe(t *testing.T) {
	const held = "export type Shared = { b: number };\nconst getterLiteral = { b: 2, d: 4, get a(): number { return 1; } };\nexport const getterShared: Shared = getterLiteral;\n"
	for _, row := range []struct{ name, main string }{
		{"skip values", "import { type Shared, shared } from './held.ts';\ntype Wanted = { b: number; c: number };\nconst wanted = shared as Wanted;\nfunction read(view: Wanted): number[] { return Object.values(view); }\nconsole.log(read(wanted).join(','));\n"},
		{"skip entries", "import { type Shared, shared } from './held.ts';\ntype Wanted = { b: number; c: number };\nconst wanted = shared as Wanted;\nfor (const [key, value] of Object.entries(wanted)) console.log(key + '=' + String(value));\n"},
		{"unnamed getter values", "import { type Shared, getterShared } from './held.ts';\ntype Other = { b: number; d: 4 };\nconst other = getterShared as Other;\nfunction read(view: Other): number[] { return Object.values(view); }\nconsole.log(read(other).join(','));\n"},
		{"unnamed getter keys", "import { type Shared, getterShared } from './held.ts';\ntype Other = { b: number; d: 4 };\nconst other = getterShared as Other;\nfunction read(view: Other): string[] { return Object.keys(view); }\nconsole.log(read(other).join(','));\n"},
		{"getter slot values", "import { type Shared, getterShared } from './held.ts';\ntype Narrow = { a: 1; b: number };\nconst narrow = getterShared as Narrow;\nfunction read(view: Narrow): number[] { return Object.values(view); }\nconsole.log(read(narrow).join(','));\n"},
		{"getter slot entries", "import { type Shared, getterShared } from './held.ts';\ntype Narrow = { a: 1; b: number };\nconst narrow = getterShared as Narrow;\nfor (const [key, value] of Object.entries(narrow)) console.log(key + '=' + String(value));\n"},
		{"getter slot field read", "import { type Shared, getterShared } from './held.ts';\ntype Narrow = { a: 1; b: number };\nconst narrow = getterShared as Narrow;\nconsole.log(String(narrow.a) + ',' + String(narrow.b));\n"},
		{"skip json", "import { encodeJson } from 'adamic';\nimport { type Shared, shared } from './held.ts';\ntype Wanted = { b: number; c: number };\nconst wanted = shared as Wanted;\nconsole.log(encodeJson<Wanted>(wanted));\n"},
	} {
		if len(row.name) > 4 && row.name[:5] == "skip " {
			continue
		}
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
