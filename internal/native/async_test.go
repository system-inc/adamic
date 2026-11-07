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
func TestAsyncCompilerSuspensionsMatchNode(test *testing.T) {
	test.Parallel()
	compilerConcurrencyMatchesNode(test, "async.a", `async function decorate(text: string): Promise<string> {
const first = await Promise.resolve(text + '-one');
await Promise.resolve();
const second = await Promise.resolve(first + '-two');
console.log(second);
return second + '-three';
}
const result = await decorate('dynamic-' + 'input');
console.log(result);
let version = 1;
function change(): number { version = 2; return 7; }
console.log(String(version + await Promise.resolve(change())));
console.log(String(version));
`, "dynamic-input-one-two\ndynamic-input-one-two-three\n8\n2\n")
}

func TestParallelCompilerSharedReferencesMatchNode(test *testing.T) {
	test.Parallel()
	compilerConcurrencyMatchesNode(test, "parallel.a", `import { parallelMap } from 'adamic';
const lookup: ReadonlyMap<string, string> = new Map([['a', 'value-' + 'a'], ['b', 'value-' + 'b']]);
const pick = (item: string, index: number): string => (lookup.get(item) ?? 'missing') + ':' + index;
function summarize(item: string, index: number): string { return pick(item, index); }
const items: readonly string[] = ['a' + '', 'b' + ''];
console.log(parallelMap(items, summarize).join('|'));
`, "value-a:0|value-b:1\n")
}

func compilerConcurrencyMatchesNode(test *testing.T, name, source, expected string) {
	test.Helper()
	directory := test.TempDir()
	path := filepath.Join(directory, name)
	if writeError := os.WriteFile(path, []byte(source), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	loader, absolutePathError := filepath.Abs(filepath.Join("..", "..", "oracle", "node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	run := func(environment []string, name string, arguments ...string) leakcheck.Run {
		command, release := boundedrun.Command(10*time.Second, name, arguments...)
		defer release()
		command.Env = append(os.Environ(), environment...)
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		runError := command.Run()
		if command.ProcessState == nil {
			test.Fatalf("running %s: %v", name, runError)
		}
		return leakcheck.Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: command.ProcessState.ExitCode()}
	}
	witness := run(nil, "node", "--disable-warning=ExperimentalWarning", loader, path)
	if witness.ExitCode != 0 || string(witness.Stdout) != expected || len(witness.Stderr) != 0 {
		test.Fatalf("source Node: %+v", witness)
	}
	checked, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), checked)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	backend := filepath.Join(directory, "backend.mjs")
	if writeError := os.WriteFile(backend, []byte(javascript.JavaScript(program)), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	got := run(nil, "node", "--disable-warning=ExperimentalWarning", loader, backend)
	if got.ExitCode != witness.ExitCode ||
		!bytes.Equal(got.Stdout, witness.Stdout) ||
		!bytes.Equal(got.Stderr, witness.Stderr) {
		test.Fatalf("JavaScript %+v differs from Node %+v", got, witness)
	}
	code := native.C(program)
	binary := filepath.Join(directory, "sanitized")
	if buildError := native.Build(code, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	got = run([]string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if got.ExitCode != witness.ExitCode ||
		!bytes.Equal(got.Stdout, witness.Stdout) ||
		!bytes.Equal(got.Stderr, witness.Stderr) {
		test.Fatalf("ASan/UBSan native %+v differs from Node %+v", got, witness)
	}
	report, checkError := leakcheck.Check(
		leakcheck.Program{C: code, Sanitized: binary, Counted: filepath.Join(directory, "counted"), Execute: run},
	)
	if checkError != nil || report != "" {
		test.Fatalf("leak check: %v %s", checkError, report)
	}
	test.Log("source Node, JavaScript, ASan/UBSan native and leak check agree")
}

// Async entry must copy the ordinary callable ABI, including hidden actual
// counts and packed default/rest parameters, into the suspension frame.
func TestAsyncOrdinaryCallableArgumentsMatchNode(test *testing.T) {
	test.Parallel()
	compilerConcurrencyMatchesNode(
		test,
		"async_arguments.a",
		`async function describe(value: string = "fallback", ...rest: string[]): Promise<string> {
 const count = arguments.length;
 await Promise.resolve();
 return String(count) + ":" + value + ":" + rest.join("|");
}
console.log(await describe());
console.log(await describe("value", "extra"));
const run = async (value: number = 7): Promise<string> => {
 await Promise.resolve();
 return String(value);
};
console.log(await run());
console.log(await run(9));
const pack = async (first: string = "head", ...rest: string[]): Promise<string> => {
 await Promise.resolve();
 return first + ":" + rest.join("|");
};
console.log(await pack());
console.log(await pack("x", "y", "z"));
`,
		"0:fallback:\n2:value:extra\n7\n9\nhead:\nx:y|z\n",
	)
}
