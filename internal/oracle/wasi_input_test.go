package oracle

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Match oracle/wasi.mjs root access, plus libc's cwd alias and explicit absolute
// scratch/dev aliases. A dot preopen alone shadows these absolute paths.
func onV8Input(t *testing.T, how inputRun, binary string) run {
	t.Helper()
	arguments := []string{"--disable-warning=ExperimentalWarning", "--input-type=module", "-e", `
import { readFileSync } from 'node:fs';
import { WASI } from 'node:wasi';
const binary = process.argv[1];
const preopens = { '/': '/' };
if (process.argv[2] === 'cwd') Object.assign(preopens, { '.': process.cwd(), [process.argv[3]]: process.argv[3], '/dev': '/dev' });
const wasi = new WASI({version:'preview1', args:[binary, ...process.argv.slice(4)], env:process.env, preopens, returnOnExit:true});
const module = await WebAssembly.compile(readFileSync(binary));
const instance = await WebAssembly.instantiate(module, wasi.getImportObject());
process.exitCode = wasi.start(instance);`, binary, "root", filepath.Dir(binary)}
	if how.directory != "" {
		arguments[len(arguments)-2] = "cwd"
	}
	return executeInput(t, how, nil, "node", append(arguments, how.arguments...)...)
}

// Exact named observations, never a blanket permission to disagree. Walk's two
// empty-path errors are a runtime bug pending codex/wasi-empty-path.
func expectedEngineBehavior(t *testing.T, path string, expected run, how inputRun, wasmtime bool) (run, bool) {
	t.Helper()
	name := filepath.Base(path)
	limited := false
	replace := func(old, changed string) {
		t.Helper()
		if strings.Count(string(expected.stdout), old) != 1 {
			t.Fatalf("%s witness missing exact control %q", name, old)
		}
		expected.stdout = []byte(strings.Replace(string(expected.stdout), old, changed, 1))
	}
	if name == "walk.a" {
		entries, err := os.ReadDir(how.directory)
		if err != nil {
			t.Fatal(err)
		}
		names := make([]string, len(entries))
		for i, entry := range entries {
			names[i] = entry.Name()
		}
		sort.Strings(names)
		replace("cannot read directory : no such directory\n", "listed "+strings.Join(names, ",")+"\n")
		replace("cannot read status of : no such file\n", "directory false\n")
		t.Log("KNOWN RUNTIME BUG walk.a: empty directory/status paths resolve to cwd; pending codex/wasi-empty-path")
		limited = true
		if wasmtime {
			replace("<dir>: 4 names, bad� name closed locked unlisted\n  bad� name: not looked up\n  closed: directory, true\n    cannot read directory <dir>/closed: permission denied\n  locked: directory, true\n    <dir>/locked: 0 names, \n  unlisted: directory, true\n    cannot read directory <dir>/unlisted: permission denied\n", "cannot read directory <dir>: failed\n")
			t.Log("HOST LIMIT walk.a: wasmtime rejects directory entries with non-UTF-8 filenames")
		}
	}
	if !wasmtime {
		return expected, limited
	}
	switch name {
	case "write_stdout_order.a":
		expected = run{stdout: []byte("first\nthird, after cannot write /dev/stdout: permission denied\n")}
		limited = true
	case "write_stderr_order.a":
		expected = run{stdout: []byte("first\nthird\n")}
		limited = true
	case "prompt_then_read.a":
		expected = run{stdout: []byte("ready\ncannot read /dev/stdin: permission denied\n")}
		limited = true
	case "arguments.a":
		expected = run{exitCode: 1, stderr: []byte("Error: failed to convert \"a\\xFFb\" to utf-8\n")}
		limited = true
	}
	if limited && name != "walk.a" {
		t.Log("HOST LIMIT", name, "exact pinned wasmtime stdout/stderr/exit asserted")
	}
	return expected, limited
}

func runWASIInputFixture(t *testing.T, compiler, fixture string, arguments []string, unreadable, writes, wasmtime bool) bool {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, fixture))
	if err != nil {
		t.Fatal(err)
	}
	shared := sharedDirectory(t)
	binary := buildEngineFixture(t, compiler, path, shared)
	how := inputRun{directory: filepath.Dir(path), arguments: append([]string{}, arguments...)}
	if os.Geteuid() == 0 {
		how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
	}
	if unreadable {
		file := filepath.Join(shared, "unreadable.txt")
		if err := os.WriteFile(file, []byte("secret\n"), 0); err != nil {
			t.Fatal(err)
		}
		how.arguments = append(how.arguments, file)
	}
	if writes {
		how.arguments = append([]string{writable(t, shared, "written")}, how.arguments...)
	}
	expected := onNodeWith(t, how, path)
	var files map[string]string
	if writes {
		files = snapshot(t, how.arguments[0])
		if err := os.RemoveAll(how.arguments[0]); err != nil {
			t.Fatal(err)
		}
		writable(t, shared, "written")
	}
	expected, limited := expectedEngineBehavior(t, fixture, expected, how, wasmtime)
	var actual run
	if wasmtime {
		actual = onWasmtime(t, how, binary)
	} else {
		actual = onV8Input(t, how, binary)
	}
	if writes {
		if difference := filesDiffer(files, snapshot(t, how.arguments[0])); difference != "" {
			t.Error(difference)
		}
	}
	if difference := disagreement(expected, actual); difference != "" {
		t.Errorf("%s: %s\nexpected exit=%d stdout=%q stderr=%q\nactual exit=%d stdout=%q stderr=%q", fixture, difference, expected.exitCode, expected.stdout, expected.stderr, actual.exitCode, actual.stdout, actual.stderr)
	}
	return limited
}

// Not parallel: deterministic counts and one runtime compilation at a time.
func TestWASIInputAgreesWithNode(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	compiler := filepath.Join(sharedDirectory(t), "adamic")
	if output, err := bounded(t, "go", "build", "-o", compiler, filepath.Join(repository, "cmd", "adamic")).CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, output)
	}
	pass, limits, fail, skip := 0, 0, 0, 0
	defer func() {
		t.Logf("V8 input fixtures=%d pass=%d limitations=%d fail=%d skip=%d", pass+limits+fail+skip, pass, limits, fail, skip)
	}()
	for _, fixture := range inputFixtures {
		t.Run(fixture.path, func(t *testing.T) {
			limited := false
			defer func() {
				if t.Failed() {
					fail++
					t.Log("FAIL", fixture.path)
				} else if t.Skipped() {
					skip++
					t.Log("SKIP", fixture.path)
				} else if limited {
					limits++
					t.Log("LIMIT", fixture.path)
				} else {
					pass++
					t.Log("PASS", fixture.path)
				}
			}()
			limited = runWASIInputFixture(t, compiler, fixture.path, fixture.arguments, fixture.unreadable, fixture.writes, false)
		})
	}
}

func TestWasmtimeCheckedWitness(t *testing.T) {
	t.Parallel()
	requireWasmtime(t)
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/writes_past_end.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	raw := onNodeWith(t, inputRun{}, path)
	checked := onJavaScriptBackend(t, program)
	if disagreement(raw, checked) != "exit codes differ" {
		t.Fatal("checked control must differ from raw source exit")
	}
	binary := filepath.Join(sharedDirectory(t), "checked.wasm")
	if err := native.Build(native.C(program), binary, native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(checked, onWasmtime(t, inputRun{}, binary)); difference != "" {
		t.Fatal(difference)
	}
	t.Logf("raw exit=%d checked exit=%d; wasmtime agrees with checked backend", raw.exitCode, checked.exitCode)
}

// The named exceptions still compare all three fields. Each independent mutant
// must fail that exact comparison; changed host behavior cannot become a pass.
func TestWASINamedBehaviorCatchesMutants(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"write_stdout_order.a", "write_stderr_order.a", "prompt_then_read.a", "arguments.a"} {
		t.Run(name, func(t *testing.T) {
			expected, limited := expectedEngineBehavior(t, name, run{}, inputRun{}, true)
			if !limited {
				t.Fatal("missing named assertion")
			}
			stdout := expected
			stdout.stdout = append(append([]byte{}, expected.stdout...), '!')
			stderr := expected
			stderr.stderr = append(append([]byte{}, expected.stderr...), '!')
			exit := expected
			exit.exitCode++
			for _, mutant := range []run{stdout, stderr, exit} {
				if disagreement(expected, mutant) == "" {
					t.Fatal("named behavior mutant escaped")
				}
			}
		})
	}
}

func TestWASIInputRunnerCatchesMutants(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, err := filepath.Abs(filepath.Join(repository, "dedication/dedication.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	original := native.C(program)
	expected := onNodeWith(t, inputRun{}, path)
	probe := func(source string) run {
		shared := sharedDirectory(t)
		binary := filepath.Join(shared, "probe.wasm")
		if err := native.Build(source, binary, native.Options{Target: "wasm32-wasi"}); err != nil {
			t.Fatal(err)
		}
		return onV8Input(t, inputRun{directory: filepath.Dir(path)}, binary)
	}
	if difference := disagreement(expected, probe(original)); difference != "" {
		t.Fatal("control", difference)
	}
	program.Strings[0] += "!"
	if disagreement(expected, probe(native.C(program))) != "stdout differs" {
		t.Fatal("stdout mutant escaped")
	}
	index := strings.LastIndex(original, "return 0;")
	if index < 0 {
		t.Fatal("missing main return")
	}
	if disagreement(expected, probe(original[:index]+strings.Replace(original[index:], "return 0;", "return 23;", 1))) != "exit codes differ" {
		t.Fatal("exit mutant escaped")
	}
	if disagreement(expected, probe("#include <stdio.h>\n"+original[:index]+"fputs(\"mutant\\n\",stderr); "+original[index:])) != "stderr differs" {
		t.Fatal("stderr mutant escaped")
	}
	t.Log("V8 input runner caught independently compiled stdout, stderr and exit mutants")
}

func TestWASIWalkBehaviorCatchesMutants(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/walk.a"))
	if err != nil {
		t.Fatal(err)
	}
	shared := sharedDirectory(t)
	how := inputRun{directory: filepath.Dir(path), arguments: []string{writable(t, shared, "written")}}
	if os.Geteuid() == 0 {
		how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
	}
	source := onNodeWith(t, how, path)
	for _, engine := range []bool{false, true} {
		expected, limited := expectedEngineBehavior(t, path, source, how, engine)
		if !limited {
			t.Fatal("missing walk assertion")
		}
		// Fixing either pending runtime bug must change the pinned result and be noticed.
		for _, pair := range [][2]string{{"directory false\n", "cannot read status of : no such file\n"}, {"listed ", "fixed listing "}} {
			changed := expected
			changed.stdout = []byte(strings.Replace(string(expected.stdout), pair[0], pair[1], 1))
			if disagreement(expected, changed) != "stdout differs" {
				t.Fatal("walk runtime mutant escaped")
			}
		}
		if engine {
			changed := expected
			changed.stdout = []byte(strings.Replace(string(expected.stdout), "cannot read directory <dir>: failed\n", "host directory now works\n", 1))
			if disagreement(expected, changed) != "stdout differs" {
				t.Fatal("walk host mutant escaped")
			}
		}
	}
	t.Log("walk empty-listing, empty-status and invalid-UTF-8 host assertion mutants caught")
}
