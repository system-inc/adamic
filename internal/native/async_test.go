package native_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Inline programs guard compiler integration without depending on piece 6's corpus.
// Node reads the source independently of both backends.
func TestAsyncCompilerSuspensionsMatchNode(t *testing.T) {
	t.Parallel()
	compilerConcurrencyMatchesNode(t, "async.a", `async function decorate(text: string): Promise<string> {
const first = await Promise.resolve(text + '-one');
await Promise.resolve();
const second = await Promise.resolve(first + '-two');
console.log(second);
return second + '-three';
}
const result = await decorate('dynamic-' + 'input');
console.log(result);
`, "dynamic-input-one-two\ndynamic-input-one-two-three\n")
}

func TestParallelCompilerSharedReferencesMatchNode(t *testing.T) {
	t.Parallel()
	compilerConcurrencyMatchesNode(t, "parallel.a", `import { parallelMap } from 'adamic';
const lookup: ReadonlyMap<string, string> = new Map([['a', 'value-' + 'a'], ['b', 'value-' + 'b']]);
const pick = (item: string, index: number): string => (lookup.get(item) ?? 'missing') + ':' + index;
function summarize(item: string, index: number): string { return pick(item, index); }
const items: readonly string[] = ['a' + '', 'b' + ''];
console.log(parallelMap(items, summarize).join('|'));
`, "value-a:0|value-b:1\n")
}

func compilerConcurrencyMatchesNode(t *testing.T, name, source, expected string) {
	t.Helper()
	directory := t.TempDir()
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	loader, err := filepath.Abs(filepath.Join("..", "..", "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	run := func(environment []string, name string, arguments ...string) leakcheck.Run {
		command, release := boundedrun.Command(10*time.Second, name, arguments...)
		defer release()
		command.Env = append(os.Environ(), environment...)
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		if command.ProcessState == nil {
			t.Fatalf("running %s: %v", name, err)
		}
		return leakcheck.Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: command.ProcessState.ExitCode()}
	}
	witness := run(nil, "node", "--disable-warning=ExperimentalWarning", loader, path)
	if witness.ExitCode != 0 || string(witness.Stdout) != expected || len(witness.Stderr) != 0 {
		t.Fatalf("source Node: %+v", witness)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	backend := filepath.Join(directory, "backend.mjs")
	if err := os.WriteFile(backend, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	got := run(nil, "node", "--disable-warning=ExperimentalWarning", loader, backend)
	if got.ExitCode != witness.ExitCode || !bytes.Equal(got.Stdout, witness.Stdout) || !bytes.Equal(got.Stderr, witness.Stderr) {
		t.Fatalf("JavaScript %+v differs from Node %+v", got, witness)
	}
	code := native.C(program)
	binary := filepath.Join(directory, "sanitized")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got = run([]string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if got.ExitCode != witness.ExitCode || !bytes.Equal(got.Stdout, witness.Stdout) || !bytes.Equal(got.Stderr, witness.Stderr) {
		t.Fatalf("ASan/UBSan native %+v differs from Node %+v", got, witness)
	}
	report, err := leakcheck.Check(leakcheck.Program{C: code, Sanitized: binary, Counted: filepath.Join(directory, "counted"), Execute: run})
	if err != nil || report != "" {
		t.Fatalf("leak check: %v %s", err, report)
	}
	t.Log("source Node, JavaScript, ASan/UBSan native and leak check agree")
}
