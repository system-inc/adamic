package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

// A probe for #70kdt66, never landing: a record read the checker narrowed present carries records.go's
// ir.Defined, which throws a TypeError when a call it doesn't see deleted the key. Under ?. Node reads
// undefined there. Each program is run on Node, lowered, built native (sanitized, with the leak check
// when it finishes) and run on the JavaScript backend, and one line says what each printed and its exit,
// so the question is whether a try around the read catches the throw (a silent branch Node never
// takes) or the stop stays uncatchable (exit 70 every time).
func TestDefinedOptionalProbe(t *testing.T) {
	t.Parallel()
	const setup = "const r: Record<string, number[]> = {};\n" +
		"function clear(): void { delete r[\"k\"]; }\n" +
		"r[\"k\"] = [];\n" +
		"clear();\n"
	for _, probe := range []struct{ name, body string }{
		{"p1_bare", "console.log(String(r[\"k\"]?.includes(1)));\n" +
			"console.log(\"after\");\n"},
		{"p2_try_includes", "try { console.log(String(r[\"k\"]?.includes(1))); } catch (e) { console.log(\"caught\"); }\n" +
			"console.log(\"after\");\n"},
		{"p3_try_push", "try { const n = r[\"k\"]?.push(1); console.log(String(n)); } catch (e) { console.log(\"caught\"); }\n" +
			"console.log(\"after\");\n"},
		{"p4_async_try", "async function later(): Promise<void> { clear(); }\n" +
			"async function probe(): Promise<void> {\n" +
			"  r[\"k\"] = [];\n" +
			"  await later();\n" +
			"  try { console.log(String(r[\"k\"]?.includes(1))); } catch (e) { console.log(\"caught\"); }\n" +
			"  console.log(\"after\");\n" +
			"}\n" +
			"await probe();\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), probe.name+".a")
			if err := os.WriteFile(path, []byte(setup+probe.body), 0600); err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			program, err := lowered(t, path)
			if err != nil {
				t.Logf("PROBE %s node=%q|%d native=refused js=refused", probe.name, truth.stdout, truth.exitCode)
				t.Logf("PROBE-DETAIL %s refusal=%q", probe.name, err.Error())
				return
			}
			native, binary := natively(t, program)
			backend := onJavaScriptBackend(t, program)
			t.Logf("PROBE %s node=%q|%d native=%q|%d js=%q|%d", probe.name, truth.stdout, truth.exitCode, native.stdout, native.exitCode, backend.stdout, backend.exitCode)
			report := "not asked"
			if native.exitCode == 0 {
				report = leaks(t, program, binary)
			}
			t.Logf("PROBE-DETAIL %s node-stderr=%q native-stderr=%q js-stderr=%q leaks=%q", probe.name, truth.stderr, native.stderr, backend.stderr, report)
		})
	}
}
