package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A probe for #afs4xvb, not for landing.
const zzEnum = "enum E { A, B }\nfunction stray(value: unknown): string { return \"seen \" + String(value); }\n"

func TestZzProbeEnumNeverBetween(t *testing.T) {
	line := func(r run) string {
		stderr := strings.ReplaceAll(strings.TrimSpace(string(r.stderr)), "\n", " ")
		if len(stderr) > 60 {
			stderr = stderr[:60]
		}
		return fmt.Sprintf("%d %q %q", r.exitCode, strings.TrimSpace(string(r.stdout)), stderr)
	}
	short := func(err error) string {
		text := err.Error()
		if index := strings.LastIndex(text, "/"); index >= 0 && index < 120 {
			text = text[index+1:]
		}
		if len(text) > 130 {
			text = text[:130]
		}
		return text
	}
	for _, probe := range []struct{ name, source string }{
		{"path", zzEnum + "interface Holder { e: E }\nfunction flip(holder: Holder): void { holder.e = E.A; }\nfunction describe(holder: Holder): string {\n\tif (holder.e === E.A) return 'a';\n\tflip(holder);\n\tif (holder.e === E.B) return 'b';\n\t{ const seen: unknown = holder.e; return stray(seen); }\n}\nconsole.log(describe({ e: E.B }));\n"},
		{"deep", zzEnum + "interface Holder { e: E }\ninterface Outer { holder: Holder }\nfunction flip(outer: Outer): void { outer.holder.e = E.A; }\nfunction describe(outer: Outer): string {\n\tif (outer.holder.e === E.A) return 'a';\n\tflip(outer);\n\tif (outer.holder.e === E.B) return 'b';\n\t{ const seen: unknown = outer.holder.e; return stray(seen); }\n}\nconsole.log(describe({ holder: { e: E.B } }));\n"},
		{"let", zzEnum + "function describe(start: E): string {\n\tlet e: E = start;\n\tconst flip = (): void => { e = E.A; };\n\tif (e === E.A) return 'a';\n\tflip();\n\tif (e === E.B) return 'b';\n\t{ const seen: unknown = e; return stray(seen); }\n}\nconsole.log(describe(E.B));\n"},
		{"switch_implicit", zzEnum + "interface Holder { e: E }\nfunction flip(holder: Holder): void { holder.e = E.A; }\nfunction describe(holder: Holder): string {\n\tif (holder.e === E.A) return 'a';\n\tflip(holder);\n\tswitch (holder.e) {\n\t\tcase E.B: return 'b';\n\t}\n\treturn 'after';\n}\nconsole.log(describe({ e: E.B }));\n"},
		{"switch_default", zzEnum + "interface Holder { e: E }\nfunction flip(holder: Holder): void { holder.e = E.A; }\nfunction describe(holder: Holder): string {\n\tif (holder.e === E.A) return 'a';\n\tflip(holder);\n\tswitch (holder.e) {\n\t\tcase E.B: return 'b';\n\t\tdefault: return stray(holder.e);\n\t}\n}\nconsole.log(describe({ e: E.B }));\n"},
		{"nothing_between", zzEnum + "interface Holder { e: E }\nfunction describe(holder: Holder): string {\n\tif (holder.e === E.A) return 'a';\n\tif (holder.e === E.B) return 'b';\n\t{ const seen: unknown = holder.e; return stray(seen); }\n}\nfunction other(): number { return 7; }\nconsole.log(describe({ e: E.B }) + ' ' + describe({ e: other() }));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			path := filepath.Join(t.TempDir(), probe.name+extension)
			if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			program, err := lowered(t, path)
			if err != nil {
				t.Logf("PROBE %s%s | node %s | stops %s", probe.name, extension, line(node), short(err))
				continue
			}
			native, _ := natively(t, program)
			t.Logf("PROBE %s%s | node %s | native %s | js %s", probe.name, extension, line(node), line(native), line(onJavaScriptBackend(t, program)))
		}
	}
	paths, err := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata/enums_*never*.a"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		if _, err := lowered(t, path); err != nil {
			t.Logf("PROBE %s stops %s", filepath.Base(path), short(err))
		} else {
			t.Logf("PROBE %s lowers", filepath.Base(path))
		}
	}
}
