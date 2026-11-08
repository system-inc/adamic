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
func onV8Input(test *testing.T, how inputRun, binary string) run {
	test.Helper()
	arguments := []string{"--disable-warning=ExperimentalWarning", "--input-type=module", "-e", `
import { readFileSync } from 'node:fs';
import { WASI } from 'node:wasi';
const binary = process.argv[1];
const wasi = new WASI({version:'preview1', args:[binary, ...process.argv.slice(2)], env:{...process.env, ADAMIC_WASI_CWD:process.cwd()}, preopens:{'/':'/'}, returnOnExit:true});
const module = await WebAssembly.compile(readFileSync(binary));
const instance = await WebAssembly.instantiate(module, wasi.getImportObject());
process.exitCode = wasi.start(instance);`, binary}
	return executeInput(test, how, nil, "node", append(arguments, how.arguments...)...)
}

// An engine the input fixtures run on. V8 answers to Node unchanged; an engine with named host
// limits rewrites Node's output into exactly what it must print instead.
type wasmEngine struct {
	run    func(test *testing.T, how inputRun, binary string) run
	limits func(test *testing.T, path string, expected run) (run, bool)
}

var v8Engine = wasmEngine{run: onV8Input}

// buildEngineFixture builds the fixture's wasm32 command in process, as TestWASIAgreesWithNode does, from the runtime
// products built ahead. It once built cmd/adamic and ran adamic build, and on Loom's uncached boxes every shard
// building the whole compiler at once ran each past its 60 s budget.
func buildEngineFixture(test *testing.T, path, directory string) (string, string) {
	test.Helper()
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	binary := filepath.Join(directory, "program.wasm")
	return binary, buildWasi(test, native.C(program), binary)
}

func runWasiInputFixture(
	test *testing.T,
	fixture string,
	arguments []string,
	unreadable, writes, plainNames bool,
	engine wasmEngine,
) (bool, bool) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, fixture))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	shared := sharedDirectory(test)
	binary, reason := buildEngineFixture(test, path, shared)
	if reason != "" {
		assertWasiRefusal(test, fixture, reason, wasiInputRefusedFixtures)
		return false, true
	}
	how := inputRun{directory: filepath.Dir(path), arguments: append([]string{}, arguments...)}
	if os.Geteuid() == 0 {
		how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
	}
	if unreadable {
		file := filepath.Join(shared, "unreadable.txt")
		if writeError := os.WriteFile(file, []byte("secret\n"), 0); writeError != nil {
			test.Fatal(writeError)
		}
		how.arguments = append(how.arguments, file)
	}
	if writes {
		how.arguments = append([]string{writableDirectory(test, shared, "written", plainNames)}, how.arguments...)
	}
	expected := onNodeWith(test, how, path)
	var files map[string]string
	if writes {
		files = snapshot(test, how.arguments[0])
		if removeError := os.RemoveAll(how.arguments[0]); removeError != nil {
			test.Fatal(removeError)
		}
		writableDirectory(test, shared, "written", plainNames)
	}
	limited := false
	if engine.limits != nil {
		expected, limited = engine.limits(test, fixture, expected)
	}
	actual := engine.run(test, how, binary)
	if assertWasiRefusal(test, fixture, wasiRefusalReason(actual), wasiInputRefusedFixtures) {
		return false, true
	}
	if writes {
		if difference := filesDiffer(files, snapshot(test, how.arguments[0])); difference != "" {
			test.Error(difference)
		}
	}
	if difference := disagreement(expected, actual); difference != "" {
		test.Errorf(
			"%s: %s\nexpected exit=%d stdout=%q stderr=%q\nactual exit=%d stdout=%q stderr=%q",
			fixture,
			difference,
			expected.exitCode,
			expected.stdout,
			expected.stderr,
			actual.exitCode,
			actual.stdout,
			actual.stderr,
		)
	}
	return limited, false
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
func runWASIInputShard(test *testing.T, shard int) {
	test.Helper()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	rows := familyCases(test, "TestWASIInputAgreesWithNode", shard, inputFixturePaths())
	if len(rows) == 0 {
		return
	}
	planted, probing := familyProbe("TestWASIInputAgreesWithNode")
	if validateOptionsError := native.ValidateOptions(
		native.Options{Target: "wasm32-wasi"},
	); validateOptionsError != nil {
		test.Fatal(validateOptionsError)
	}
	pass, limits, refused, fail, skip := 0, 0, 0, 0, 0
	defer func() {
		test.Logf(
			"V8 input shard=%d fixtures=%d pass=%d limitations=%d refused=%d fail=%d skip=%d",
			shard,
			pass+limits+refused+fail+skip,
			pass,
			limits,
			refused,
			fail,
			skip,
		)
	}()
	for _, index := range rows {
		fixture := inputFixtures[index]
		engine := v8Engine
		if probing && fixture.path == planted {
			// TestFamiliesCatchAPlantedCase's mutant: one byte more on what V8 printed.
			engine.run = func(test *testing.T, how inputRun, binary string) run {
				actual := onV8Input(test, how, binary)
				actual.stdout = append(actual.stdout, '!')
				return actual
			}
		}
		test.Run(fixture.path, func(test *testing.T) {
			limited, wasRefused := false, false
			defer func() {
				if test.Failed() {
					fail++
					test.Log("FAIL", fixture.path)
				} else if test.Skipped() {
					skip++
					test.Log("SKIP", fixture.path)
				} else if wasRefused {
					refused++
				} else if limited {
					limits++
					test.Log("LIMIT", fixture.path)
				} else {
					pass++
					test.Log("PASS", fixture.path)
				}
			}()
			limited, wasRefused = runWasiInputFixture(
				test,
				fixture.path,
				fixture.arguments,
				fixture.unreadable,
				fixture.writes,
				fixture.plainNames,
				engine,
			)
		})
	}
}

func TestWasiInputRunnerCatchesMutants(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "dedication/dedication.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	original := native.C(program)
	expected := onNodeWith(test, inputRun{}, path)
	probe := func(source string) run {
		shared := sharedDirectory(test)
		binary := filepath.Join(shared, "probe.wasm")
		if buildError := native.Build(source, binary, native.Options{Target: "wasm32-wasi"}); buildError != nil {
			test.Fatal(buildError)
		}
		return onV8Input(test, inputRun{directory: filepath.Dir(path)}, binary)
	}
	if difference := disagreement(expected, probe(original)); difference != "" {
		test.Fatal("control", difference)
	}
	program.Strings[0] += "!"
	if disagreement(expected, probe(native.C(program))) != "stdout differs" {
		test.Fatal("stdout mutant escaped")
	}
	// Generated main returns the host's exit status after its ordinary cleanup.
	const mainReturn = "return adamic_host_exit_status();"
	index := strings.LastIndex(original, mainReturn)
	if index < 0 {
		test.Fatal("missing main return")
	}
	if disagreement(
		expected,
		probe(original[:index]+strings.Replace(original[index:], mainReturn, "return 23;", 1)),
	) != "exit codes differ" {
		test.Fatal("exit mutant escaped")
	}
	if disagreement(
		expected,
		probe("#include <stdio.h>\n"+original[:index]+"fputs(\"mutant\\n\",stderr); "+original[index:]),
	) != "stderr differs" {
		test.Fatal("stderr mutant escaped")
	}
	test.Log("V8 input runner caught independently compiled stdout, stderr and exit mutants")
}

func TestWasiWalkBehaviorCatchesMutants(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/walk.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	shared := sharedDirectory(test)
	how := inputRun{directory: filepath.Dir(path), arguments: []string{writable(test, shared, "written")}}
	if os.Geteuid() == 0 {
		how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
	}
	expected := onNodeWith(test, how, path)
	holdsNodeEmptyPathLines(test, expected)
	test.Log("walk empty-path mutants caught")
}

// Node's empty-path errors are held as Node prints them: losing either must be noticed.
func holdsNodeEmptyPathLines(test *testing.T, expected run) {
	test.Helper()
	for _, line := range []string{
		"cannot read directory : no such directory\n",
		"cannot read status of : no such file\n",
	} {
		if strings.Count(string(expected.stdout), line) != 1 {
			test.Fatalf("walk witness lost Node's line %q", line)
		}
		changed := expected
		changed.stdout = []byte(strings.Replace(string(expected.stdout), line, "", 1))
		if disagreement(expected, changed) != "stdout differs" {
			test.Fatal("walk empty-path mutant escaped")
		}
	}
}
