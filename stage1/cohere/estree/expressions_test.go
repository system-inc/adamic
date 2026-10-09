package estree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func recoveredExpressions() []string {
	return []string{"const C = @d class {};", "const C = @d @e(1) class Named extends Base { static x=1; };", "(@d class { m(){return 1;} });", "function f(){ await 'x'; await 1; await true; await this; }", "function f(){ yield 'x'; yield 1; yield true; }", "await\nx; yield\nx;", "await(x); yield(x);"}
}
func TestRecoveredExpressions(t *testing.T) {
	t.Parallel()
	list := manifest(t, recoveredExpressions())
	want := execute(t, "", goOracle(t), "--manifest", list)
	main, _ := filepath.Abs("main.ts")
	checkPort(t, main, []string{"--manifest", list}, want, false, false)
	t.Logf("%d decorated class / await / yield cases, %d bytes match Go", len(recoveredExpressions()), len(want))
}
func TestRecoveredExpressionMutant(t *testing.T) {
	t.Parallel()
	list := manifest(t, recoveredExpressions())
	want := execute(t, "", goOracle(t), "--manifest", list)
	path := mutantPort(t, "sourceLookahead.ts", "(scanner.flags & 1) === 0 &&\n        (token", "(scanner.flags & 1) >= 0 &&\n        (token")
	checkPort(t, path, []string{"--manifest", list}, want, true, false)
}

func TestUnattachedDecorator(t *testing.T) {
	t.Parallel()
	main, _ := filepath.Abs("main.ts")
	binary, script := build(t, main, true)
	for index, body := range []string{"@dec\nawait 1", "@dec\nx"} {
		t.Run(fmt.Sprintf("%04d", index), func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "input.ts")
			os.WriteFile(path, []byte(body), 0644)
			checkRefusalModes(t, main, binary, script, path, "ESTree unattached decorator")

		})
	}
}

func TestUnattachedDecoratorControl(t *testing.T) {
	t.Parallel()
	list := manifest(t, []string{"@dec\nawait 1", "@dec\nx"})
	statuses := string(execute(t, "", goOracle(t), "--audit", list, t.TempDir()))
	if strings.Count(statuses, `"status":"error"`) != 2 {
		t.Fatalf("Go decorator refusals changed: %s", statuses)
	}
	main := mutantPort(t, "pipeline.ts", `parser.nodes[id]?.kind === 'Decorator'`, `parser.nodes[id]?.kind === 'UnusedDecoratorControl'`)
	checkAcceptanceControl(t, main, list, 2)
}
