package native_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// tsc's CompilerOptions in miniature: a declared fixed object (configFile) and a declared dictionary
// (paths) over an index signature whose union also holds scalars and two arrays. Each is the only member
// of its runtime class in the slot's union, so its read is a heap test, not typeof.
const recordMemberHeapOptions = `interface Source { text: string; diagnostics: string[] }
interface Options { configFile?: Source; paths?: { [index: string]: string[] }; rootDirs?: string[]; [option: string]: string | number | boolean | string[] | number[] | { [index: string]: string[] } | Source | undefined }
`

const recordMemberHeapSource = recordMemberHeapOptions + `function describe(options: Options): string {
  const file = options.configFile;
  const paths = options.paths;
  return (file ? file.text + ":" + file.diagnostics.length : "none") + " " + (paths ? Object.keys(paths).join(",") : "nopaths");
}
const o: Options = { configFile: { text: "tsconfig", diagnostics: ["a", "b"] }, paths: { "@x/*": ["src/*"] } };
console.log(describe(o));
console.log(describe({}));
`

// Both reads go through a checked narrow, one to each heap class, and print what Node prints on the
// JavaScript backend and natively under the sanitizers, letting go of everything (leakcheck: the counted
// build and leaks on macOS, LeakSanitizer on Linux).
func TestRecordMemberHeapReadsMatchNode(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(recordMemberHeapSource), 0600); err != nil {
		test.Fatal(err)
	}
	truth, err := exec.Command("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", path).Output()
	if err != nil || string(truth) != "tsconfig:2 @x/*\nnone nopaths\n" {
		test.Fatalf("source Node: %v, %q", err, truth)
	}
	program := recordMemberHeapLowered(test, path)
	arms := map[ir.Type]bool{}
	for _, function := range program.Functions {
		if function.Name != "record_declared_member" {
			continue
		}
		if narrow, ok := function.Body[len(function.Body)-1].(ir.Return).Value.(ir.Narrow); ok && narrow.Checked {
			arms[narrow.To] = true
		}
	}
	if !arms[ir.Object] || !arms[ir.Record] {
		test.Fatalf("checked arms %v, want a fixed object's and a dictionary's", arms)
	}
	js := filepath.Join(test.TempDir(), "main.mjs")
	if err := os.WriteFile(js, []byte(javascript.JavaScript(program)), 0600); err != nil {
		test.Fatal(err)
	}
	if got, err := exec.Command(
		"node",
		"--disable-warning=ExperimentalWarning",
		"../../oracle/node.mjs",
		js,
	).Output(); err != nil || !bytes.Equal(got, truth) {
		test.Fatalf("JavaScript: %v, %q; Node %q", err, got, truth)
	}
	code := native.C(program)
	stdout, stderr, exit, binary := recordMemberHeapRun(test, code)
	if exit != 0 || stdout != string(truth) || stderr != "" {
		test.Fatalf("native: exit %d, stdout %q, stderr %q; Node %q", exit, stdout, stderr, truth)
	}
	if report := leakcheck.Report(test, code, binary); report != "" {
		test.Fatal(report)
	}
}

// A computed name in a record literal can put the dictionary under configFile and the fixed object under
// paths (recordComputedEntry leaves a declared object kind to its read). Node reads either and says so;
// native stops at the read, since only the shape test tells a fixed object from a dictionary. Each lie's
// mutant drops its arm's shape test from the C, and the read then lets the other class through.
func TestRecordMemberHeapReadsStopOnTheOtherClass(test *testing.T) {
	test.Parallel()
	for _, lie := range []struct {
		name, write, read, declared, shape string
	}{
		{
			"a dictionary under configFile",
			`const lie: { [index: string]: string[] } = { "@x/*": ["src/*"] };`,
			"configFile",
			"Source",
			" || adamic_record_is((const adamic_object *)",
		},
		{
			"a fixed object under paths",
			`const lie: Source = { text: "t", diagnostics: [] };`,
			"paths",
			"{ [index: string]: string[]; }",
			" || !adamic_record_is((const adamic_object *)",
		},
	} {
		test.Run(lie.name, func(test *testing.T) {
			test.Parallel()
			source := recordMemberHeapOptions + `const keys: string[] = ["` + lie.read + `"];
const key = keys[0] ?? "other";
` + lie.write + `
const options: Options = { [key]: lie };
const read = options.` + lie.read + `;
console.log(read ? "present" : "absent");
`
			want := "adamic: panic: field read failed: options." +
				lie.read +
				" is not a " +
				lie.declared +
				"; expected " +
				lie.declared +
				", found another member of the index signature's union, which a computed-key or index-signature write may have stored\n"
			code := recordMemberHeapC(test, source)
			if stdout, stderr, exit, _ := recordMemberHeapRun(
				test,
				code,
			); exit != 70 || stdout != "" || stderr != want {
				test.Fatalf("native: exit %d, stdout %q, stderr %q; want exit 70 and %q", exit, stdout, stderr, want)
			}
			if !strings.Contains(code, lie.shape) {
				test.Fatalf("no shape test %q in the C", lie.shape)
			}
			mutant := strings.Replace(code, lie.shape, strings.Replace(lie.shape, "|| ", "|| 0 * ", 1), 1)
			if stdout, _, exit, _ := recordMemberHeapRun(test, mutant); exit == 70 || stdout != "present\n" {
				test.Fatalf(
					"the mutant without the shape test: exit %d, stdout %q; want it to read the lie as Node does",
					exit,
					stdout,
				)
			}
		})
	}
}

// The lies are written by a computed key, which .a refuses at compile time when it can land under a
// declared property (#npqr3sx), so the program is .ts, where the read's run-time check is the stop.
func recordMemberHeapC(test *testing.T, source string) string {
	test.Helper()
	path := filepath.Join(test.TempDir(), "main.ts")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		test.Fatal(err)
	}
	return native.C(recordMemberHeapLowered(test, path))
}

func recordMemberHeapLowered(test *testing.T, path string) *ir.Program {
	test.Helper()
	checked, err := load.Load([]string{path})
	if err != nil {
		test.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		test.Fatal(err)
	}
	return program
}

// recordMemberHeapRun builds code under the sanitizers and runs it, returning the binary too.
func recordMemberHeapRun(test *testing.T, code string) (string, string, int, string) {
	test.Helper()
	binary := filepath.Join(test.TempDir(), "native")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	command := exec.Command(binary)
	command.Env = append(os.Environ(), "UBSAN_OPTIONS=halt_on_error=1")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	exit := 0
	if failed, ok := err.(*exec.ExitError); ok {
		exit = failed.ExitCode()
	} else if err != nil {
		test.Fatal(err)
	}
	return stdout.String(), stderr.String(), exit, binary
}
