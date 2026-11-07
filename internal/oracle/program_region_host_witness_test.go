package oracle

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native/runtimeproducts"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func programRegionPending(test *testing.T, path, feature string) {
	test.Helper()
	source, err := load.Load([]string{path})
	if err != nil {
		test.Fatal(err)
	}
	_, err = lower.LowerWithOptions(context.Background(), source, lower.Options{ProgramRegion: true})
	if err == nil {
		test.Fatalf("pending feature %s is now available; activate %s", feature, path)
	}
	if !strings.Contains(err.Error(), feature) {
		test.Fatalf("pending %s failed elsewhere: %v", feature, err)
	}
	// A pending feature is a pass that proves the boundary: the program is refused for this feature and nothing
	// else, and the test fails the day the feature lowers, so the fixture is activated rather than skipped.
	test.Logf("registered pending %s: %v", feature, err)
}
func TestProgramRegionHost14AgreesWithNode(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs("../../stage3/fixtures/host/14_getCurrentDirectory.a")
	if err != nil {
		test.Fatal(err)
	}
	counts := checkProgramRegion(test, programRegionLowered(test, path, true), path)
	if counts[5] != 4 {
		test.Fatalf("host members %d, want 4", counts[5])
	}
}

func TestProgramRegionHost14DefaultAgreesWithNode(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs("../../stage3/fixtures/host/14_getCurrentDirectory.a")
	if err != nil {
		test.Fatal(err)
	}
	checkProgramRegionMode(test, programRegionLowered(test, path, false), path, false)
}

const programHost14Path = "stage3/fixtures/host/14_getCurrentDirectory.a"

func programHost14WASM(test *testing.T, counted bool) string {
	test.Helper()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path := filepath.Join(repository, programHost14Path)
	code := native.C(programRegionLowered(test, path, false))
	binary := filepath.Join(test.TempDir(), "host.wasm")
	options := native.Options{Target: "wasm32-wasi", Count: counted}
	var err error
	if counted {
		archive := filepath.Join(runtimeproducts.ProgramArchive(test, options, native.Features(code)), "runtime.a")
		err = native.BuildWithRuntime(code, binary, options, archive)
	} else {
		err = native.Build(code, binary, options)
	}
	reason := ""
	if err != nil {
		var refusal *native.TargetRefused
		if !errors.As(err, &refusal) {
			test.Fatal(err)
		}
		reason = refusal.Error()
	}
	if assertWASIRefusal(test, programHost14Path, reason, wasiRefusedFixtures) {
		return ""
	}
	return binary
}
func TestProduct_ProgramHost14WASICounted(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	runtimeproducts.ProgramArchive(
		test,
		native.Options{Target: "wasm32-wasi", Count: true},
		[]string{"ADAMIC_NODE_HOST"},
	)
}
func TestProgramRegionHost14DefaultWASIAgreesWithNode(test *testing.T) {
	test.Parallel()
	binary := programHost14WASM(test, false)
	if binary == "" {
		return
	}
	path, err := filepath.Abs(filepath.Join(repository, programHost14Path))
	if err != nil {
		test.Fatal(err)
	}
	expected := onNode(test, path)
	actual := execute(
		test,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle", "wasi.mjs"),
		binary,
	)
	if diff := disagreement(expected, actual); diff != "" {
		test.Fatalf("WASI: %s: %s", diff, actual.stderr)
	}
	counted := execute(
		test,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle", "wasi.mjs"),
		programHost14WASM(test, true),
	)
	if counted.exitCode != expected.exitCode || string(counted.stdout) != string(expected.stdout) {
		test.Fatalf("counted WASI differs: %+v", counted)
	}
	if report := leakcheck.Unbalanced(leakRun(counted)); report != "" {
		test.Fatal(report)
	}
	test.Logf("WASI agrees with Node; leak check clean: %s", counted.stderr)
}
func TestProgramRegionHost14DefaultShimPending(test *testing.T) {
	test.Parallel()
	binary := programHost14WASM(test, false)
	if binary == "" {
		return
	}
	if !shimFileFixtures[programHost14Path] {
		test.Fatal("host file access must be listed")
	}
	inventory := execute(
		test,
		"node",
		"-e",
		`const fs=require('node:fs');const m=new WebAssembly.Module(fs.readFileSync(process.argv[1]));console.log(JSON.stringify(WebAssembly.Module.imports(m).filter(i=>i.module==='wasi_snapshot_preview1').map(i=>i.name).sort()));`,
		binary,
	)
	var imports []string
	if inventory.exitCode != 0 {
		test.Fatalf("imports: %s", inventory.stderr)
	}
	if err := json.Unmarshal(inventory.stdout, &imports); err != nil {
		test.Fatal(err)
	}
	const want = "args_get, args_sizes_get, clock_time_get, environ_get, environ_sizes_get, fd_close, fd_fdstat_get, fd_prestat_dir_name, fd_prestat_get, fd_readdir, fd_seek, fd_write, path_create_directory, path_filestat_get, path_open, path_readlink, path_remove_directory, path_unlink_file, proc_exit, random_get"
	if actual := strings.Join(imports, ", "); actual != want {
		test.Fatalf("imports %s, want %s", actual, want)
	}
	actual := execute(
		test,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle", "wasi-shim.mjs"),
		binary,
	)
	if actual.exitCode != 1 ||
		!strings.Contains(string(actual.stderr), "AdamicUnsupportedWASI: Unsupported WASI import: random_get") {
		test.Fatalf("pending shim failed elsewhere or became runnable: %+v", actual)
	}
	// The file-free shim deliberately supplies neither random_get nor host files, so it must fail on exactly that.
	test.Log(
		"pending shim random_get for mkdtemp and host filesystem access; the file-free shim deliberately supplies neither",
	)
}
func programHost14ProgramWASIPending(test *testing.T) {
	test.Helper()
	err := native.ValidateOptions(native.Options{ProgramRegion: true, Target: "wasm32-wasi"})
	const reason = "native: Program regions require a one-shot native CLI"
	if err == nil || err.Error() != reason {
		test.Fatalf("pending Program WASI boundary changed: %v", err)
	}
	// Program regions are refused off the native CLI, and this proves the refusal holds with its exact reason.
	test.Log(reason)
}
func TestProgramRegionHost14ProgramWASIPending(test *testing.T) {
	test.Parallel()
	programHost14ProgramWASIPending(test)
}
func TestProgramRegionHost14ProgramShimPending(test *testing.T) {
	test.Parallel()
	programHost14ProgramWASIPending(test)
}
