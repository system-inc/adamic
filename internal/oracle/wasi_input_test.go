package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Match oracle/wasi.mjs: one root preopen and the working directory the
// cwd-dependent runtime adapters enter before resolving relative paths.
func onV8Input(t *testing.T, how inputRun, binary string) run {
	t.Helper()
	arguments := []string{"--disable-warning=ExperimentalWarning", "--input-type=module", "-e", `
import { readFileSync } from 'node:fs';
import { WASI } from 'node:wasi';
const binary = process.argv[1];
const wasi = new WASI({version:'preview1', args:[binary, ...process.argv.slice(2)], env:{...process.env, ADAMIC_WASI_CWD:process.cwd()}, preopens:{'/':'/'}, returnOnExit:true});
const module = await WebAssembly.compile(readFileSync(binary));
const instance = await WebAssembly.instantiate(module, wasi.getImportObject());
process.exitCode = wasi.start(instance);`, binary}
	return executeInput(t, how, nil, "node", append(arguments, how.arguments...)...)
}

// An engine the input fixtures run on. V8 answers to Node unchanged; an engine with named host
// limits rewrites Node's output into exactly what it must print instead.
type wasmEngine struct {
	run    func(t *testing.T, how inputRun, binary string) run
	limits func(t *testing.T, path string, expected run) (run, bool)
}

var v8Engine = wasmEngine{run: onV8Input}

// buildEngineFixture builds the fixture's wasm32 command in process, as TestWASIAgreesWithNode does, from the runtime
// products built ahead. It once built cmd/adamic and ran adamic build, and on Loom's uncached boxes every shard
// building the whole compiler at once ran each past its 60 s budget.
func buildEngineFixture(t *testing.T, path, directory string) string {
	t.Helper()
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "program.wasm")
	buildWASI(t, native.C(program), binary)
	return binary
}

func runWASIInputFixture(t *testing.T, fixture string, arguments []string, unreadable, writes bool, engine wasmEngine) bool {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, fixture))
	if err != nil {
		t.Fatal(err)
	}
	shared := sharedDirectory(t)
	binary := buildEngineFixture(t, path, shared)
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
	wasiRuntimeRefusal(t, actual)
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

// inputFixturePaths is every input fixture's path, in the order inputFixtures runs them.
func inputFixturePaths() []string {
	paths := make([]string, len(inputFixtures))
	for index, fixture := range inputFixtures {
		paths[index] = fixture.path
	}
	return paths
}

// TestWASIInputAgreesWithNode runs every input fixture's wasm32 build on V8, as a family (families_test.go):
// TestWASIInputAgreesWithNode_000 and on each run the fixtures their shard owns, since each is an adamic build of
// the whole runtime when the gate runs uncached, and all eight in one test ran past Loom's 60 s budget. Within a
// shard the fixtures run in order, and each shard logs its own counts.
func runWASIInputShard(t *testing.T, shard int) {
	t.Helper()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	rows := familyCases(t, "TestWASIInputAgreesWithNode", shard, inputFixturePaths())
	if len(rows) == 0 {
		return
	}
	planted, probing := familyProbe("TestWASIInputAgreesWithNode")
	if err := native.ValidateOptions(native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	pass, limits, fail, skip := 0, 0, 0, 0
	defer func() {
		t.Logf("V8 input shard=%d fixtures=%d pass=%d limitations=%d fail=%d skip=%d", shard, pass+limits+fail+skip, pass, limits, fail, skip)
	}()
	for _, index := range rows {
		fixture := inputFixtures[index]
		engine := v8Engine
		if probing && fixture.path == planted {
			// TestFamiliesCatchAPlantedCase's mutant: one byte more on what V8 printed.
			engine.run = func(t *testing.T, how inputRun, binary string) run {
				actual := onV8Input(t, how, binary)
				actual.stdout = append(actual.stdout, '!')
				return actual
			}
		}
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
			limited = runWASIInputFixture(t, fixture.path, fixture.arguments, fixture.unreadable, fixture.writes, engine)
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
	// Generated main returns the host's exit status after its ordinary cleanup.
	const mainReturn = "return adamic_host_exit_status();"
	index := strings.LastIndex(original, mainReturn)
	if index < 0 {
		t.Fatal("missing main return")
	}
	if disagreement(expected, probe(original[:index]+strings.Replace(original[index:], mainReturn, "return 23;", 1))) != "exit codes differ" {
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
