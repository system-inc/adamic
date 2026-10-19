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
