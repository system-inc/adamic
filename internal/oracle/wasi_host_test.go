package oracle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func buildWASI(t *testing.T, source, binary string) {
	t.Helper()
	if err := native.Build(source, binary, native.Options{Target: "wasm32-wasi"}); err != nil {
		var refused *native.TargetRefused
		if errors.As(err, &refused) {
			t.Skipf("target refusal: %v", refused)
		}
		t.Fatal(err)
	}
}

// Only these exact, owned runtime refusals are target gaps. A build failure,
// ordinary panic, trap, or unknown mismatch must still fail the oracle.
func wasiRuntimeRefusal(t *testing.T, actual run) {
	t.Helper()
	if actual.exitCode != 70 {
		return
	}
	for _, reason := range []string{
		"file creation permission bits are not supported",
		"directory creation permission bits are not supported",
		"fs.utimesSync timestamp precision or range is unavailable",
	} {
		if string(actual.stderr) == "adamic: panic: wasm32-wasi: "+reason+"\n" {
			t.Skipf("target refusal: %s", reason)
		}
	}
}

// The input fixtures hold the host members and their filesystem effects. Run
// them with the same directory, arguments and uid as the native input leg.
func TestWASIInputAgreesWithNode(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/wasi.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range inputFixtures {
		t.Run(fixture.path, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			shared := sharedDirectory(t)
			binary := filepath.Join(shared, "program.wasm")
			buildWASI(t, native.C(program), binary)
			how := inputRun{directory: filepath.Dir(path), arguments: fixture.arguments}
			if os.Geteuid() == 0 {
				how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
			}
			if fixture.unreadable {
				unreadable := filepath.Join(shared, "unreadable.txt")
				if err := os.WriteFile(unreadable, []byte("secret\n"), 0000); err != nil {
					t.Fatal(err)
				}
				how.arguments = append(append([]string{}, how.arguments...), unreadable)
			}
			prepared := func(name string) inputRun {
				if !fixture.writes {
					return how
				}
				return inputRun{directory: how.directory, arguments: append([]string{writable(t, shared, name)}, how.arguments...), credential: how.credential}
			}
			nodeRun, wasmRun := prepared("node"), prepared("wasi")
			expected := onNodeWith(t, nodeRun, path)
			actual := executeInput(t, wasmRun, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, binary}, wasmRun.arguments...)...)
			wasiRuntimeRefusal(t, actual)
			if difference := disagreement(expected, actual); difference != "" {
				t.Errorf("%s\nnode: exit %d, stdout %q, stderr %q\nwasi: exit %d, stdout %q, stderr %q", difference, expected.exitCode, expected.stdout, expected.stderr, actual.exitCode, actual.stdout, actual.stderr)
			}
			if fixture.writes {
				if difference := filesDiffer(snapshot(t, nodeRun.arguments[0]), snapshot(t, wasmRun.arguments[0])); difference != "" {
					t.Errorf("filesystem effects: %s", difference)
				}
			}
		})
	}
}

// Token pasting bypasses the target-aware generated-C check, so these calls
// exercise the runtime refusal that a caller without target information sees.
func TestWASIHostRuntimeRefusals(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	for _, probe := range []struct{ prefix, suffix, arguments, reason string }{
		{"adamic_fs_file_", "mkdtemp", "NULL", "fs.mkdtempSync requires temporary directory creation"},
		{"adamic_node_", "pid", "", "process.pid requires process identifiers"},
		{"adamic_node_", "platform", "", "process.platform requires a Node host platform"},
		{"adamic_node_", "columns", "", "process.stdout.columns requires terminal size"},
		{"adamic_node_", "memory_usage", "", "process.memoryUsage requires allocator observations"},
		{"adamic_node_", "argv", "", "process.argv requires executable paths"},
		{"adamic_fs_file_", "open", "&path, &flag, 0600", "file creation permission bits are not supported"},
		{"adamic_fs_file_", "mkdir", "&path, false, 0700", "directory creation permission bits are not supported"},
		{"adamic_fs_file_", "utimes", "&path, 1000.5, 1000.5", "fs.utimesSync timestamp precision or range is unavailable"},
	} {
		t.Run(probe.suffix, func(t *testing.T) {
			source := fmt.Sprintf("#include \"adamic.h\"\n#define MEMBER %s ## %s\nint main(void) { adamic_string path = ADAMIC_STRING(\"unused\"), flag = ADAMIC_STRING(\"w\"); adamic_start(0, NULL); (void)MEMBER(%s); return 0; }\n", probe.prefix, probe.suffix, probe.arguments)
			binary := filepath.Join(t.TempDir(), "runtime-refusal.wasm")
			buildWASI(t, source, binary)
			actual := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), binary)
			if actual.exitCode != 70 || len(actual.stdout) != 0 || string(actual.stderr) != "adamic: panic: wasm32-wasi: "+probe.reason+"\n" {
				t.Fatalf("runtime refusal: %+v", actual)
			}
		})
	}
}

func TestWASIFileAgreesWithNode(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/wasi.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fsFileFixtures {
		t.Run(fixture, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_fs_file_"+fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			shared := sharedDirectory(t)
			binary := filepath.Join(shared, "program.wasm")
			buildWASI(t, native.C(program), binary)
			nodeRun, wasmRun := fsFilePrepare(t, shared, "node"), fsFilePrepare(t, shared, "wasi")
			expected := onNodeWith(t, nodeRun, path)
			actual := executeInput(t, wasmRun, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, binary}, wasmRun.arguments...)...)
			wasiRuntimeRefusal(t, actual)
			if difference := disagreement(expected, actual); difference != "" {
				t.Errorf("%s\nnode: exit %d, stdout %q, stderr %q\nwasi: exit %d, stdout %q, stderr %q", difference, expected.exitCode, expected.stdout, expected.stderr, actual.exitCode, actual.stdout, actual.stderr)
			}
			if difference := filesDiffer(fsFileSnapshot(t, nodeRun.arguments[0]), fsFileSnapshot(t, wasmRun.arguments[0])); difference != "" {
				t.Errorf("filesystem effects: %s", difference)
			}
		})
	}
}
