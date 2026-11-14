package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNeverReadExploreEnumRoads(t *testing.T) {
	for _, probe := range []struct{ name, source string }{
		{"enum_let", "enum E { A, B }\nfunction describe(start: E): string {\n\tlet e: E = start;\n\tconst flip = (): void => { e = E.A; };\n\tif (e === E.A) return 'a';\n\tflip();\n\tif (e === E.B) return 'b';\n\t{ const seen: unknown = e; return 'seen ' + String(seen); }\n}\nconsole.log(describe(E.B));\n"},
		{"enum_path", "enum E { A, B }\ninterface Holder { e: E }\nfunction flip(holder: Holder): void { holder.e = E.A; }\nfunction describe(holder: Holder): string {\n\tif (holder.e === E.A) return 'a';\n\tflip(holder);\n\tif (holder.e === E.B) return 'b';\n\t{ const seen: unknown = holder.e; return 'seen ' + String(seen); }\n}\nconsole.log(describe({ e: E.B }));\n"},
	} {
		path := filepath.Join(t.TempDir(), probe.name+".a")
		if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Logf("EXPLORE %s: %.150v", probe.name, err)
			continue
		}
		node, backend := onNode(t, path), onJavaScriptBackend(t, program)
		native, _ := natively(t, program)
		line := func(r run) string { return fmt.Sprintf("%d %q %.90q", r.exitCode, r.stdout, r.stderr) }
		t.Logf("EXPLORE %s: node %s | js %s | native %s", probe.name, line(node), line(backend), line(native))
	}
}
