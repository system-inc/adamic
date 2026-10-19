package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Use the same cwd, arguments, credentials, and filesystem inputs as the source.
func TestNodeFSDirectoryWASIAgreement(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	for _, name := range []string{"node_path_posix", "node_path_basename", "node_path_relative", "node_fs_directory_entries", "node_fs_directory_union", "node_fs_directory_realpath", "node_fs_directory_system", "node_fs_directory_stat_options", "node_fs_directory_permissions", "node_fs_directory_symlink"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			shared := sharedDirectory(t)
			prepare := func(label string) inputRun {
				how := inputRun{directory: filepath.Dir(path)}
				if name == "node_fs_directory_symlink" {
					how.arguments = []string{writable(t, shared, label)}
				}
				if os.Geteuid() == 0 {
					how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
				}
				return how
			}
			truth := onNodeWith(t, prepare("node"), path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("Node failed: %+v", truth)
			}
			binary := filepath.Join(shared, "program.wasm")
			if err := native.Build(native.C(program), binary, native.Options{Target: "wasm32-wasi"}); err != nil {
				t.Fatal(err)
			}
			runner, err := filepath.Abs(filepath.Join(repository, "oracle/wasi.mjs"))
			if err != nil {
				t.Fatal(err)
			}
			how := prepare("wasi")
			args := append([]string{"--disable-warning=ExperimentalWarning", runner, binary}, how.arguments...)
			got := executeInput(t, how, nil, "node", args...)
			if difference := disagreement(truth, got); difference != "" {
				t.Fatalf("%s: Node stdout %q stderr %q exit %d; WASI stdout %q stderr %q exit %d", difference, truth.stdout, truth.stderr, truth.exitCode, got.stdout, got.stderr, got.exitCode)
			}
		})
	}
}

func TestNodeFSDirectoryWASICWDMutant(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_path_posix.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	shared := sharedDirectory(t)
	binary := filepath.Join(shared, "program.wasm")
	if err := native.Build(native.C(program), binary, native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	runner, err := os.ReadFile(filepath.Join(repository, "oracle/wasi.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	mutated := strings.Replace(string(runner), "ADAMIC_WASI_CWD: process.cwd()", "ADAMIC_WASI_CWD: '/'", 1)
	if mutated == string(runner) {
		t.Fatal("cwd mutant changed nothing")
	}
	mutant := filepath.Join(shared, "runner.mjs")
	if err := os.WriteFile(mutant, []byte(mutated), 0644); err != nil {
		t.Fatal(err)
	}
	how := inputRun{directory: filepath.Dir(path)}
	truth := onNodeWith(t, how, path)
	got := executeInput(t, how, nil, "node", "--disable-warning=ExperimentalWarning", mutant, binary)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant failed outside Node comparison: Node exit %d stderr %q; WASI exit %d stderr %q", truth.exitCode, truth.stderr, got.exitCode, got.stderr)
	}
	if disagreement(truth, got) != "stdout differs" {
		t.Fatal("Node didn't catch wrong initial cwd")
	}
	t.Log("root cwd mutant caught only by Node stdout; clean successful WASI run")
}
