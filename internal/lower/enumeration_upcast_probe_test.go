package lower

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// Temporary probe (#1atvfp5, Lower's upcast roads): never lands. Each row's literal sits in held.ts so rule
// 700 never decides it, and the row logs whether it lowers, what native prints, and whether JavaScript
// agrees with Node.
func TestEnumerationUpcastProbe(t *testing.T) {
	const held = "export const o = { b: 1, get a(): number { return 2; } } as { b: number };\nexport const o3 = ({ get a(): number { return 2; }, b: 1 }) as {};\n"
	for _, row := range []struct{ name, main string }{
		{"u1 keys", "import { o } from './held';\nconsole.log(Object.keys(o).join(','));\n"},
		{"u1 values", "import { o } from './held';\nfunction read(v: { b: number }): number[] { return Object.values(v); }\nconsole.log(read(o).join(','));\n"},
		{"u2 entries", "import { o } from './held';\nfunction show(p: { b: number }): void { for (const [k, v] of Object.entries(p)) console.log(k + '=' + String(v)); }\nshow(o);\n"},
		{"u2 spread", "import { o } from './held';\nfunction copy(p: { b: number }): string { const c = { ...p }; return Object.keys(c).join(','); }\nconsole.log(copy(o));\n"},
		{"u3 for-in", "import { o3 } from './held';\nfor (const k in o3) console.log(k);\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			directory := t.TempDir()
			if err := os.WriteFile(filepath.Join(directory, "held.ts"), []byte(held), 0o644); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(directory, "main.ts")
			if err := os.WriteFile(path, []byte(row.main), 0o644); err != nil {
				t.Fatal(err)
			}
			checked, err := load.Load([]string{path})
			if err != nil {
				t.Logf("upcast probe %s: load: %v", row.name, err)
				return
			}
			program, err := Lower(context.Background(), checked)
			t.Logf("upcast probe %s lowered: err=%v", row.name, err)
			if err != nil {
				return
			}
			code, out, errs := brandedNative(t, program)
			t.Logf("upcast probe %s native: exit=%d stdout=%q stderr=%q", row.name, code, out, errs)
			agreeEntry(t, path, program, 0, nil, nil)
			t.Logf("upcast probe %s: JavaScript agreed with Node", row.name)
		})
	}
}
