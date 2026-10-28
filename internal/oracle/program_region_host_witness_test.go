package oracle

import (
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func programRegionPending(t *testing.T, path, feature string) {
	t.Helper()
	source, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.LowerWithOptions(context.Background(), source, lower.Options{ProgramRegion: true})
	if err == nil {
		t.Fatalf("pending feature %s is now available; activate %s", feature, path)
	}
	if !strings.Contains(err.Error(), feature) {
		t.Fatalf("pending %s failed elsewhere: %v", feature, err)
	}
	// A pending feature is a pass that proves the boundary: the program is refused for this feature and nothing
	// else, and the test fails the day the feature lowers, so the fixture is activated rather than skipped.
	t.Logf("registered pending %s: %v", feature, err)
}
func TestProgramRegionHost14AgreesWithNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../../stage3/fixtures/host/14_getCurrentDirectory.a")
	if err != nil {
		t.Fatal(err)
	}
	counts := checkProgramRegion(t, programRegionLowered(t, path, true), path)
	if counts[5] != 4 {
		t.Fatalf("host members %d, want 4", counts[5])
	}
}

func TestProgramRegionHost14DefaultAgreesWithNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../../stage3/fixtures/host/14_getCurrentDirectory.a")
	if err != nil {
		t.Fatal(err)
	}
	checkProgramRegionMode(t, programRegionLowered(t, path, false), path, false)
}

const programHost14Path = "stage3/fixtures/host/14_getCurrentDirectory.a"

func programHost14WASM(t *testing.T, counted bool) string {
	t.Helper()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path := filepath.Join(repository, programHost14Path)
	code := native.C(programRegionLowered(t, path, false))
	binary := filepath.Join(t.TempDir(), "host.wasm")
	options := native.Options{Target: "wasm32-wasi", Count: counted}
	var err error
	if counted {
		archive := filepath.Join(programRegionRuntime(t, options, native.Features(code)), "runtime.a")
		err = native.BuildWithRuntime(code, binary, options, archive)
	} else {
		err = native.Build(code, binary, options)
	}
	if err != nil {
		t.Fatal(err)
	}
	return binary
}
func TestProduct_ProgramHost14WASICounted(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	programRegionRuntime(t, native.Options{Target: "wasm32-wasi", Count: true}, []string{"ADAMIC_NODE_HOST"})
}
func TestProgramRegionHost14DefaultWASIAgreesWithNode(t *testing.T) {
	t.Parallel()
	binary := programHost14WASM(t, false)
	path, err := filepath.Abs(filepath.Join(repository, programHost14Path))
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	actual := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "wasi.mjs"), binary)
	if diff := disagreement(expected, actual); diff != "" {
		t.Fatalf("WASI: %s: %s", diff, actual.stderr)
	}
	counted := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "wasi.mjs"), programHost14WASM(t, true))
	if counted.exitCode != expected.exitCode || string(counted.stdout) != string(expected.stdout) {
		t.Fatalf("counted WASI differs: %+v", counted)
	}
	if report := leakcheck.Unbalanced(leakRun(counted)); report != "" {
		t.Fatal(report)
	}
	t.Logf("WASI agrees with Node; leak check clean: %s", counted.stderr)
}
func TestProgramRegionHost14DefaultShimPending(t *testing.T) {
	t.Parallel()
	binary := programHost14WASM(t, false)
	if !shimFileFixtures[programHost14Path] {
		t.Fatal("host file access must be listed")
	}
	inventory := execute(t, "node", "-e", `const fs=require('node:fs');const m=new WebAssembly.Module(fs.readFileSync(process.argv[1]));console.log(JSON.stringify(WebAssembly.Module.imports(m).filter(i=>i.module==='wasi_snapshot_preview1').map(i=>i.name).sort()));`, binary)
	var imports []string
	if inventory.exitCode != 0 {
		t.Fatalf("imports: %s", inventory.stderr)
	}
	if err := json.Unmarshal(inventory.stdout, &imports); err != nil {
		t.Fatal(err)
	}
	const want = "args_get, args_sizes_get, clock_time_get, environ_get, environ_sizes_get, fd_close, fd_fdstat_get, fd_prestat_dir_name, fd_prestat_get, fd_readdir, fd_seek, fd_write, path_create_directory, path_filestat_get, path_open, path_readlink, path_remove_directory, path_unlink_file, proc_exit, random_get"
	if actual := strings.Join(imports, ", "); actual != want {
		t.Fatalf("imports %s, want %s", actual, want)
	}
	actual := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "wasi-shim.mjs"), binary)
	if actual.exitCode != 1 || !strings.Contains(string(actual.stderr), "AdamicUnsupportedWASI: Unsupported WASI import: random_get") {
		t.Fatalf("pending shim failed elsewhere or became runnable: %+v", actual)
	}
	// The file-free shim deliberately supplies neither random_get nor host files, so it must fail on exactly that.
	t.Log("pending shim random_get for mkdtemp and host filesystem access; the file-free shim deliberately supplies neither")
}
func programHost14ProgramWASIPending(t *testing.T) {
	t.Helper()
	err := native.ValidateOptions(native.Options{ProgramRegion: true, Target: "wasm32-wasi"})
	const reason = "native: Program regions require a one-shot native CLI"
	if err == nil || err.Error() != reason {
		t.Fatalf("pending Program WASI boundary changed: %v", err)
	}
	// Program regions are refused off the native CLI, and this proves the refusal holds with its exact reason.
	t.Log(reason)
}
func TestProgramRegionHost14ProgramWASIPending(t *testing.T) {
	t.Parallel()
	programHost14ProgramWASIPending(t)
}
func TestProgramRegionHost14ProgramShimPending(t *testing.T) {
	t.Parallel()
	programHost14ProgramWASIPending(t)
}
