package oracle

import (
	"os"
	"path/filepath"
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

// An engine the input fixtures run on. V8 answers to Node unchanged; an engine with named host
// limits rewrites Node's output into exactly what it must print instead.
type wasmEngine struct {
	run    func(t *testing.T, how inputRun, binary string) run
	limits func(t *testing.T, path string, expected run) (run, bool)
}

var v8Engine = wasmEngine{run: onV8Input}

func buildEngineFixture(t *testing.T, compiler, path, directory string) string {
	t.Helper()
	binary := filepath.Join(directory, "program.wasm")
	command := bounded(t, compiler, "build", "--target", "wasm32-wasi", path, "-o", binary)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("adamic build: %v\n%s", err, output)
	}
	return binary
}

func runWASIInputFixture(t *testing.T, compiler, fixture string, arguments []string, unreadable, writes bool, engine wasmEngine) bool {
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
	limited := false
	if engine.limits != nil {
		expected, limited = engine.limits(t, fixture, expected)
	}
	actual := engine.run(t, how, binary)
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
			limited = runWASIInputFixture(t, compiler, fixture.path, fixture.arguments, fixture.unreadable, fixture.writes, v8Engine)
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
	expected := onNodeWith(t, how, path)
	holdsNodeEmptyPathLines(t, expected)
	t.Log("walk empty-path mutants caught")
}

// Node's empty-path errors are held as Node prints them: losing either must be noticed.
func holdsNodeEmptyPathLines(t *testing.T, expected run) {
	t.Helper()
	for _, line := range []string{"cannot read directory : no such directory\n", "cannot read status of : no such file\n"} {
		if strings.Count(string(expected.stdout), line) != 1 {
			t.Fatalf("walk witness lost Node's line %q", line)
		}
		changed := expected
		changed.stdout = []byte(strings.Replace(string(expected.stdout), line, "", 1))
		if disagreement(expected, changed) != "stdout differs" {
			t.Fatal("walk empty-path mutant escaped")
		}
	}
}
