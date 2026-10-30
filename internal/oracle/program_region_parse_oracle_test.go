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

func programParseOracle(t *testing.T, enabled bool) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("Linux instrumented parse witness")
	}
	args := programParseArguments(t)
	absolute, err := filepath.Abs("testdata/program_region/parse.a")
	if err != nil {
		t.Fatal(err)
	}
	truth := execute(t, "node", append([]string{"--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", absolute}, args...)...)
	if truth.exitCode != 0 {
		t.Fatalf("source Node: %d %s", truth.exitCode, truth.stderr)
	}
	js := filepath.Join(programParseLower(t, enabled), "program.mjs")
	actual := execute(t, "node", append([]string{"--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", js}, args...)...)
	if difference := disagreement(truth, actual); difference != "" {
		t.Fatalf("JavaScript: %s", difference)
	}
	binary := filepath.Join(programParseBuildKind(t, enabled, true), "parse")
	actual = executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary, args...)
	actual.stdout = graphRSS.ReplaceAll(actual.stdout, nil)
	if difference := disagreement(truth, actual); difference != "" {
		t.Fatalf("sanitized parse: %s", difference)
	}
}
func TestProgramRegionParseOnAgreesWithNode(t *testing.T) { t.Parallel(); programParseOracle(t, true) }
func TestProgramRegionParseOffAgreesWithNode(t *testing.T) {
	t.Parallel()
	programParseOracle(t, false)
}

func TestProgramRegionParseWASIImports(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	const path = "internal/oracle/testdata/program_region/parse.a"
	const want = "args_get, args_sizes_get, clock_time_get, environ_get, environ_sizes_get, fd_close, fd_fdstat_get, fd_filestat_get, fd_prestat_dir_name, fd_prestat_get, fd_read, fd_seek, fd_write, path_filestat_get, path_open, proc_exit"
	if !shimFileFixtures[path] {
		t.Fatal("file-reading parser must be excluded from the file-free shim")
	}
	code, err := os.ReadFile(filepath.Join(programParseLower(t, false), "code.c"))
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "parse.wasm")
	if err := native.Build(string(code), binary, native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	inventory := execute(t, "node", "-e", `const fs=require('node:fs');const m=new WebAssembly.Module(fs.readFileSync(process.argv[1]));console.log(JSON.stringify(WebAssembly.Module.imports(m).filter(i=>i.module==='wasi_snapshot_preview1').map(i=>i.name).sort()));`, binary)
	var imports []string
	if inventory.exitCode != 0 {
		t.Fatalf("imports: %s", inventory.stderr)
	}
	if err := json.Unmarshal(inventory.stdout, &imports); err != nil {
		t.Fatal(err)
	}
	if actual := strings.Join(imports, ", "); actual != want {
		t.Fatalf("parser imports %s, want %s", actual, want)
	}
	t.Log(want)
}
