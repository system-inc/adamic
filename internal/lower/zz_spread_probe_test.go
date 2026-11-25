package lower

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// Scratch probe for #492fbpd's spread lead, never for landing: logs what each program does on Node (the
// source itself), the JavaScript backend and native under the sanitizers.
func TestZZSpreadProbe(t *testing.T) {
	const absent = "interface Box { items: number[] | undefined; counter: Counter | undefined }\n" +
		"class Counter { count(...xs: number[]): number { return xs.length; } }\n" +
		"const box: Box = { items: [1], counter: new Counter() };\n" +
		"const more: number[] = [2, 3];\n" +
		"function clear(): void { box.items = undefined; box.counter = undefined; }\n"
	for _, probe := range []struct{ name, source string }{
		{"absent push", absent + "if (box.items !== undefined) {\n    clear();\n    console.log(String((box.items?.push)(...more)));\n}\n"},
		{"absent concat", absent + "if (box.items !== undefined) {\n    clear();\n    console.log(String((box.items?.concat)(...more)));\n}\n"},
		{"absent unshift", absent + "if (box.items !== undefined) {\n    clear();\n    console.log(String((box.items?.unshift)(...more)));\n}\n"},
		{"absent class method", absent + "if (box.counter !== undefined) {\n    clear();\n    console.log(String((box.counter?.count)(...more)));\n}\n"},
		{"present push", absent + "console.log(String((box.items?.push)(...more)));\nconsole.log(String(box.items?.length));\n"},
		{"present concat", absent + "console.log(String((box.items?.concat)(...more)));\n"},
		{"present class method", absent + "console.log(String((box.counter?.count)(...more)));\n"},
		{"plain optional push", absent + "console.log(String(box.items?.push(...more)));\nclear();\nconsole.log(String(box.items?.push(...more)));\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			directory := t.TempDir()
			runner, err := filepath.Abs("../../oracle/node.mjs")
			if err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(directory, "source.ts")
			if err := os.WriteFile(source, []byte(probe.source), 0o644); err != nil {
				t.Fatal(err)
			}
			run := func(label string, command *exec.Cmd) {
				command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
				var out, errs bytes.Buffer
				command.Stdout, command.Stderr = &out, &errs
				code := 0
				if err := command.Run(); err != nil {
					var exit *exec.ExitError
					if !errors.As(err, &exit) {
						t.Logf("PROBE %s %s: run error %v", probe.name, label, err)
						return
					}
					code = exit.ExitCode()
				}
				stderr := errs.String()
				if len(stderr) > 600 {
					stderr = stderr[:600]
				}
				t.Logf("PROBE %s %s: exit %d stdout %q stderr %q", probe.name, label, code, out.String(), stderr)
			}
			run("node", exec.Command("node", "--disable-warning=ExperimentalWarning", "--experimental-strip-types", source))
			program, err := lowerSourceAs(t, probe.source, ".ts")
			if err != nil {
				t.Logf("PROBE %s lower: %v", probe.name, err)
				return
			}
			t.Logf("PROBE %s lower: lowers", probe.name)
			generated := filepath.Join(directory, "generated.mjs")
			if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
				t.Fatal(err)
			}
			run("javascript", exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated))
			binary := filepath.Join(directory, "native")
			if err := native.Build(native.C(program), binary, native_runtime.Options{Sanitize: true}); err != nil {
				t.Logf("PROBE %s native build: %v", probe.name, err)
				return
			}
			run("native", exec.Command(binary))
		})
	}
}
