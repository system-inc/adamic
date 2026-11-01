package lower

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

func TestStaticGetterProbe(t *testing.T) {
	prelude := `interface Positioned { readonly kind: number; readonly pos: number; }
interface Clocked extends Positioned { readonly kind: 1; }
class Clock { static kind = 1; static get pos(): number { return 3; } }
`
	for _, probe := range []struct{ name, source string }{
		{"cast", prelude + `function show(value: Positioned): string { if (value.kind === 1) { const c = value as Clocked; return String(c.pos); } return "?"; }
console.log(show(Clock));
`},
		{"cast-direct", prelude + `const c = Clock as Positioned as Clocked;
console.log(String(c.pos));
`},
		{"predicate", prelude + `function isClocked(value: Positioned): value is Clocked { return value.kind === 1; }
function show(value: Positioned): string { if (isClocked(value)) { return String(value.pos); } return "?"; }
console.log(show(Clock));
`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, err := lowerTypeScriptAssertionSource(t, probe.source)
			if err != nil {
				t.Logf("PROBE %s: refused: %v", probe.name, err)
				return
			}
			_, dispatched := accessorReachReads(program, "pos")
			viewed, _ := accessorReachReads(program, "pos")
			directory := t.TempDir()
			generated := filepath.Join(directory, "generated.mjs")
			os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644)
			runner, _ := filepath.Abs("../../oracle/node.mjs")
			source := filepath.Join(directory, "main.ts")
			os.WriteFile(source, []byte(probe.source), 0o644)
			outputs := map[string]string{}
			for name, command := range map[string]*exec.Cmd{
				"node":       exec.Command("node", "--disable-warning=ExperimentalWarning", runner, source),
				"javascript": exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated),
			} {
				var out, errs bytes.Buffer
				command.Stdout, command.Stderr = &out, &errs
				command.Run()
				outputs[name] = out.String() + "|" + errs.String()
			}
			binary := filepath.Join(directory, "native")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				outputs["native"] = "build: " + err.Error()
			} else {
				command := exec.Command(binary)
				var out, errs bytes.Buffer
				command.Stdout, command.Stderr = &out, &errs
				command.Run()
				outputs["native"] = out.String() + "|" + errs.String()
			}
			t.Logf("PROBE %s: lowered, checked pos reads %d, dispatcher calls %d, outputs %q", probe.name, viewed, dispatched, outputs)
		})
	}
}
