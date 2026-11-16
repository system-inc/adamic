package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

// Probe for #v5b6pzp, never for landing: Lower's reads of a promise admitted through an object literal's field
// or a union slot, each typed PromiseLike<number> and held as an object. It logs what each does and never fails.
func TestPromiseViewAdmissionProbe(t *testing.T) {
	t.Parallel()
	const literal = "const n: { p: PromiseLike<number> } = { p: Promise.resolve(7) };\n"
	const union = "const u: number | PromiseLike<number> = Promise.resolve(7);\n"
	const show = "function show(v: PromiseLike<number>): string {\n\treturn typeof v;\n}\n"
	for _, probe := range []struct{ name, source string }{
		{"literal_into_view", literal + "const q: PromiseLike<number> = n.p;\nconsole.log(typeof q);\n"},
		{"literal_into_argument", show + literal + "console.log(show(n.p));\n"},
		{"literal_awaited", literal + "async function main(): Promise<void> {\n\tconsole.log(String(await n.p));\n}\nvoid main();\n"},
		{"literal_then", literal + "n.p.then((v) => console.log(String(v)));\n"},
		{"union_into_view", union + "if (typeof u !== 'number') {\n\tconst w: PromiseLike<number> = u;\n\tconsole.log(typeof w);\n}\n"},
		{"union_into_argument", show + union + "if (typeof u !== 'number') {\n\tconsole.log(show(u));\n}\n"},
		{"union_awaited", union + "async function main(): Promise<void> {\n\tif (typeof u !== 'number') {\n\t\tconsole.log(String(await u));\n\t}\n}\nvoid main();\n"},
		{"union_then", union + "if (typeof u !== 'number') {\n\tu.then((v) => console.log(String(v)));\n}\n"},
		{"literal_then_member_typeof", literal + "console.log(typeof n.p.then);\n"},
		{"literal_then_in", literal + "console.log(String('then' in n.p));\n"},
		{"literal_then_value", literal + "const later = n.p.then((v: number): number => v + 1);\nasync function main(): Promise<void> {\n\tconsole.log(String(await later));\n}\nvoid main();\n"},
		{"union_then_member_typeof", union + "if (typeof u !== 'number') {\n\tconsole.log(typeof u.then);\n}\n"},
		{"union_then_in", union + "if (typeof u !== 'number') {\n\tconsole.log(String('then' in u));\n}\n"},
		{"union_then_value", union + "async function main(): Promise<void> {\n\tif (typeof u !== 'number') {\n\t\tconst later = u.then((v: number): number => v + 1);\n\t\tconsole.log(String(await later));\n\t}\n}\nvoid main();\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				path := filepath.Join(t.TempDir(), probe.name+extension)
				if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
					t.Fatal(err)
				}
				truth := onNode(t, path)
				program, err := lowered(t, path)
				if err != nil {
					t.Logf("PROBE %s: node exit %d stdout %q | STOPS: %v", probe.name+extension, truth.exitCode, truth.stdout, err)
					return
				}
				sanitized, _ := nativelyUncached(t, program)
				for backend, got := range map[string]run{"sanitized": sanitized, "release": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
					verdict := "agrees"
					if difference := disagreement(truth, got); difference != "" {
						verdict = "DISAGREES: " + difference
					}
					t.Logf("PROBE %s: node exit %d stdout %q | %s %s, stderr %q", probe.name+extension, truth.exitCode, truth.stdout, backend, verdict, got.stderr)
				}
			})
		}
	}
}
