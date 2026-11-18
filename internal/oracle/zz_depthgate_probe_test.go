package oracle

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// A probe for #rt5evf5 head 2, never for landing: what the identity-bound walk answers on each shape.
func TestDepthGateProbe(t *testing.T) {
	const fakeMap = "const backing = new Map<string, number>([['a', 1]]);\nclass FakeMap implements ReadonlyMap<string, number> {\n\treadonly size = 1;\n\tget(key: string): number | undefined { return backing.get(key); }\n\thas(key: string): boolean { return key === 'a'; }\n\tforEach(): void {}\n\tkeys() { return backing.keys(); }\n\tvalues() { return backing.values(); }\n\tentries() { return backing.entries(); }\n\t[Symbol.iterator]() { return backing.entries(); }\n}\n"
	only := os.Getenv("DEPTHGATE_ONLY")
	for _, probe := range []struct{ name, source string }{
		{"generic_real", "interface Box<T> {\n\treadonly v: T;\n}\nconst real = new Map<string, number>([['a', 1]]);\nconst box: Box<Map<string, number>> = { v: real };\nconst v: { a: { b: { c: { d: { e: Box<Map<string, number>> } } } } } = { a: { b: { c: { d: { e: box } } } } };\nconst w: { readonly a: { readonly b: { readonly c: { readonly d: { readonly e: Box<ReadonlyMap<string, number>> } } } } } = v;\nconsole.log(String(w.a.b.c.d.e.v.has('a')));\n"},
		{"generic_fake", fakeMap + "interface Box<T> {\n\treadonly v: T;\n}\nconst box: Box<FakeMap> = { v: new FakeMap() };\nconst v: { a: { b: { c: { d: { e: Box<FakeMap> } } } } } = { a: { b: { c: { d: { e: box } } } } };\nconst w: { readonly a: { readonly b: { readonly c: { readonly d: { readonly e: Box<ReadonlyMap<string, number>> } } } } } = v;\nconsole.log(String(w.a.b.c.d.e.v.has('a')));\n"},
		{"deep_reference", "interface Deep<T> {\n\treadonly next: Deep<Deep<T>> | undefined;\n\treadonly m: ReadonlyMap<string, T>;\n}\ninterface Own<T> {\n\treadonly next: Own<Own<T>> | undefined;\n\treadonly m: Map<string, T>;\n}\nconst o: Own<number> = { next: undefined, m: new Map<string, number>([['a', 1]]) };\nconst d: Deep<number> = o;\nconsole.log(String(d.m.has('a')));\n"},
		{"deep_anonymous", "type Deep<T> = { readonly next: Deep<{ readonly x: T }> | undefined; readonly m: ReadonlyMap<string, T> };\ntype Own<T> = { readonly next: Own<{ readonly x: T }> | undefined; readonly m: Map<string, T> };\nconst o: Own<number> = { next: undefined, m: new Map<string, number>([['a', 1]]) };\nconst d: Deep<number> = o;\nconsole.log(String(d.m.has('a')));\n"},
		{"deep_tuple", "interface Deep<T> {\n\treadonly next: Deep<[T, T]> | undefined;\n\treadonly m: ReadonlyMap<string, T>;\n}\ninterface Own<T> {\n\treadonly next: Own<[T, T]> | undefined;\n\treadonly m: Map<string, T>;\n}\nconst o: Own<number> = { next: undefined, m: new Map<string, number>([['a', 1]]) };\nconst d: Deep<number> = o;\nconsole.log(String(d.m.has('a')));\n"},
	} {
		if only != "" && only != probe.name {
			continue
		}
		for _, extension := range []string{".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				path := filepath.Join(t.TempDir(), probe.name+extension)
				if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
					t.Fatal(err)
				}
				truth := onNode(t, path)
				started := time.Now()
				program, err := lowered(t, path)
				t.Logf("PROBE %s lowering %s: %v", probe.name, time.Since(started), err)
				if err != nil {
					return
				}
				sanitized, _ := nativelyUncached(t, program)
				for backend, got := range map[string]run{"sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
					t.Logf("PROBE %s %s: %q", probe.name, backend, disagreement(truth, got))
				}
			})
		}
	}
}
