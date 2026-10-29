package oracle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestNodeProcessWASIAgreement(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	for _, one := range []struct{ name, refusal string }{
		{"host", "process.platform requires a Node host platform"},
		{"performance", ""}, {"performance_core", ""}, {"cleared_marks", ""}, {"missing_mark", ""}, {"tmpdir", ""},
		{"os_namespace", "process.platform requires a Node host platform"},
		{"terminal", "process.stdout.columns requires terminal size"},
		{"blocking", "process.stdout._handle requires terminal flag control"},
	} {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_process_"+one.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			got := onWASI(t, native.C(program))
			if one.name == "host" {
				if got.exitCode != 70 || string(got.stderr) != wasiHostPlatformPanic {
					t.Fatalf("want main's process.platform refusal: %+v", got)
				}
				return
			}
			if one.refusal != "" {
				if got.exitCode != 70 || !strings.Contains(string(got.stderr), "wasm32-wasi: "+one.refusal) {
					t.Fatalf("want explicit refusal %q: %+v", one.refusal, got)
				}
				t.Log(one.refusal)
				return
			}
			truth := onNode(t, path)
			if difference := disagreement(truth, got); difference != "" {
				t.Fatalf("%s: Node %+v WASI %+v", difference, truth, got)
			}
		})
	}
}

func TestNodeProcessCachedDirectoryWASI(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_process_cached_cwd.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "program.wasm")
	if err := native.Build(native.C(program), binary, native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	driver, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_process_cached_cwd.py"))
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/wasi.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	sourceRunner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	truth := execute(t, "python3", driver, "node", "--disable-warning=ExperimentalWarning", sourceRunner, path)
	got := execute(t, "python3", driver, "node", "--disable-warning=ExperimentalWarning", runner, binary)
	if difference := disagreement(truth, got); difference != "" {
		t.Fatalf("%s: Node %+v WASI %+v", difference, truth, got)
	}
}

func TestNodeProcessFreshRemovedCwdWASIRefusal(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/node_process_cached_cwd.a"))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Replace(string(source), "const first = process.cwd();", "process.chdir('.');", 1)
	text = strings.Replace(text, "process.cwd() === first", "process.cwd().length > 0", 1)
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "program")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	wasm := filepath.Join(t.TempDir(), "program.wasm")
	if err := native.Build(native.C(program), wasm, native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	absolute := func(relative string) string {
		path, err := filepath.Abs(filepath.Join(repository, relative))
		if err != nil {
			t.Fatal(err)
		}
		return path
	}
	driver := absolute("internal/oracle/testdata/node_process_cached_cwd.py")
	observe := func(command ...string) run {
		result := execute(t, "python3", append([]string{driver}, command...)...)
		if result.exitCode != 0 || len(result.stderr) != 0 {
			t.Fatalf("barrier driver failed: %+v", result)
		}
		var output struct {
			Stdout, Stderr []byte
			ExitCode       int
		}
		if err := json.Unmarshal(result.stdout, &output); err != nil {
			t.Fatal(err)
		}
		return run{stdout: output.Stdout, stderr: output.Stderr, exitCode: output.ExitCode}
	}
	truth := observe("node", "--disable-warning=ExperimentalWarning", absolute("oracle/node.mjs"), path)
	if truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "uv_cwd") {
		t.Fatalf("want Node's fresh-cwd error: %+v", truth)
	}
	if difference := disagreement(truth, observe(binary)); difference != "" {
		t.Fatal("native: " + difference)
	}
	got := observe("node", "--disable-warning=ExperimentalWarning", absolute("oracle/wasi.mjs"), wasm)
	reason := "wasm32-wasi: process.cwd requires directory handles when saved cwd is unavailable"
	if got.exitCode != 70 || !strings.Contains(string(got.stderr), reason) {
		t.Fatalf("want explicit capability refusal: %+v", got)
	}
	t.Log(reason)
}
