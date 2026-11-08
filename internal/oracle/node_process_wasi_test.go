package oracle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestNodeProcessWasiAgreement(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	for _, name := range []string{
		"host",
		"performance",
		"performance_core",
		"cleared_marks",
		"missing_mark",
		"tmpdir",
		"os_namespace",
		"terminal",
		"blocking",
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/node_process_"+name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			got, reason := observeWasi(test, native.C(program))
			// The handle fixtures have their own exact, asserted runtime refusal; it
			// is not one of wasiRefusalReason's three filesystem refusals.
			if want := wasiProcessRefusedFixtures[name]; want != "" &&
				reason == "" &&
				got.exitCode == 70 &&
				len(got.stdout) == 0 &&
				string(got.stderr) == "adamic: panic: wasm32-wasi: "+want+"\n" {
				reason = want
			}
			if assertWasiRefusal(test, name, reason, wasiProcessRefusedFixtures) {
				return
			}
			truth := onNode(test, path)
			if difference := disagreement(truth, got); difference != "" {
				test.Fatalf("%s: Node %+v WASI %+v", difference, truth, got)
			}
		})
	}
}

func TestNodeProcessCachedDirectoryWasi(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/node_process_cached_cwd.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	binary := filepath.Join(test.TempDir(), "program.wasm")
	if buildError := native.Build(native.C(program), binary, native.Options{Target: "wasm32-wasi"}); buildError != nil {
		test.Fatal(buildError)
	}
	driver, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/node_process_cached_cwd.py"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/wasi.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	sourceRunner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := execute(test, "python3", driver, "node", "--disable-warning=ExperimentalWarning", sourceRunner, path)
	got := execute(test, "python3", driver, "node", "--disable-warning=ExperimentalWarning", runner, binary)
	if difference := disagreement(truth, got); difference != "" {
		test.Fatalf("%s: Node %+v WASI %+v", difference, truth, got)
	}
}

func TestNodeProcessFreshRemovedCwdWasiRefusal(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	source, readError := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/node_process_cached_cwd.a"))
	if readError != nil {
		test.Fatal(readError)
	}
	text := strings.Replace(string(source), "const first = process.cwd();", "process.chdir('.');", 1)
	text = strings.Replace(text, "process.cwd() === first", "process.cwd().length > 0", 1)
	path := filepath.Join(test.TempDir(), "main.a")
	if writeError := os.WriteFile(path, []byte(text), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	binary := filepath.Join(test.TempDir(), "program")
	if buildError := native.Build(native.C(program), binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	wasm := filepath.Join(test.TempDir(), "program.wasm")
	if buildError := native.Build(native.C(program), wasm, native.Options{Target: "wasm32-wasi"}); buildError != nil {
		test.Fatal(buildError)
	}
	absolute := func(relative string) string {
		path, absolutePathError := filepath.Abs(filepath.Join(repository, relative))
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		return path
	}
	driver := absolute("internal/oracle/testdata/node_process_cached_cwd.py")
	observe := func(command ...string) run {
		result := execute(test, "python3", append([]string{driver}, command...)...)
		if result.exitCode != 0 || len(result.stderr) != 0 {
			test.Fatalf("barrier driver failed: %+v", result)
		}
		var output struct {
			Stdout, Stderr []byte
			ExitCode       int
		}
		if decodeError := json.Unmarshal(result.stdout, &output); decodeError != nil {
			test.Fatal(decodeError)
		}
		return run{stdout: output.Stdout, stderr: output.Stderr, exitCode: output.ExitCode}
	}
	truth := observe("node", "--disable-warning=ExperimentalWarning", absolute("oracle/node.mjs"), path)
	if truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "uv_cwd") {
		test.Fatalf("want Node's fresh-cwd error: %+v", truth)
	}
	if difference := disagreement(truth, observe(binary)); difference != "" {
		test.Fatal("native: " + difference)
	}
	got := observe("node", "--disable-warning=ExperimentalWarning", absolute("oracle/wasi.mjs"), wasm)
	reason := "wasm32-wasi: process.cwd requires directory handles when saved cwd is unavailable"
	if got.exitCode != 70 || !strings.Contains(string(got.stderr), reason) {
		test.Fatalf("want explicit capability refusal: %+v", got)
	}
	test.Log(reason)
}
