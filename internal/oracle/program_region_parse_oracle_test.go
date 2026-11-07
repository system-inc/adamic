package oracle

import (
	"encoding/json"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func programParseOracle(test *testing.T, enabled bool) {
	test.Helper()
	if runtime.GOOS != "linux" {
		test.Skip("Linux instrumented parse witness")
	}
	args := programParseArguments(test)
	absolute, err := filepath.Abs("testdata/program_region/parse.a")
	if err != nil {
		test.Fatal(err)
	}
	truth := execute(
		test,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", absolute}, args...)...,
	)
	if truth.exitCode != 0 {
		test.Fatalf("source Node: %d %s", truth.exitCode, truth.stderr)
	}
	js := filepath.Join(programParseLower(test, enabled), "program.mjs")
	actual := execute(
		test,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", js}, args...)...,
	)
	if difference := disagreement(truth, actual); difference != "" {
		test.Fatalf("JavaScript: %s", difference)
	}
	binary := filepath.Join(programParseBuildKind(test, enabled, true), "parse")
	actual = executeWith(
		test,
		[]string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"},
		binary,
		args...,
	)
	actual.stdout = graphRSS.ReplaceAll(actual.stdout, nil)
	if difference := disagreement(truth, actual); difference != "" {
		test.Fatalf("sanitized parse: %s", difference)
	}
}
func TestProgramRegionParseOnAgreesWithNode(test *testing.T) {
	test.Parallel()
	programParseOracle(test, true)
}
func TestProgramRegionParseOffAgreesWithNode(test *testing.T) {
	test.Parallel()
	programParseOracle(test, false)
}

func TestProgramRegionParseWASIImports(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	const path = "internal/oracle/testdata/program_region/parse.a"
	const want = "args_get, args_sizes_get, clock_time_get, environ_get, environ_sizes_get, fd_close, fd_fdstat_get, fd_filestat_get, fd_prestat_dir_name, fd_prestat_get, fd_read, fd_seek, fd_write, path_filestat_get, path_open, proc_exit"
	if !shimFileFixtures[path] {
		test.Fatal("file-reading parser must be excluded from the file-free shim")
	}
	code, err := os.ReadFile(filepath.Join(programParseLower(test, false), "code.c"))
	if err != nil {
		test.Fatal(err)
	}
	binary := filepath.Join(test.TempDir(), "parse.wasm")
	if err := native.Build(string(code), binary, native.Options{Target: "wasm32-wasi"}); err != nil {
		test.Fatal(err)
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
	if actual := strings.Join(imports, ", "); actual != want {
		test.Fatalf("parser imports %s, want %s", actual, want)
	}
	test.Log(want)
}
