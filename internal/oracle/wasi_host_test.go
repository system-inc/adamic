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
func wasiRefusalReason(actual run) string {
	if actual.exitCode != 70 {
		return ""
	}
	for _, reason := range []string{
		"file creation permission bits are not supported",
		"directory creation permission bits are not supported",
		"fs.utimesSync timestamp precision or range is unavailable",
	} {
		if string(actual.stderr) == "adamic: panic: wasm32-wasi: "+reason+"\n" {
			return reason
		}
	}
	return ""
}

func wasiRuntimeRefusal(t *testing.T, actual run) {
	t.Helper()
	if reason := wasiRefusalReason(actual); reason != "" {
		t.Skipf("target refusal: %s", reason)
	}
}

// The input fixtures hold the host members and their filesystem effects. Run
// them with the same directory, arguments and uid as the native input leg.
func wasiInputAgreesWithNode(t *testing.T, index int) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/wasi.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range inputFixtures[index : index+1] {
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
func wasiHostRuntimeRefusal(t *testing.T, index int) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	probes := []struct{ prefix, suffix, arguments, reason string }{
		{"adamic_fs_file_", "mkdtemp", "NULL", "fs.mkdtempSync requires temporary directory creation"},
		{"adamic_node_", "pid", "", "process.pid requires process identifiers"},
		{"adamic_node_", "platform", "", "process.platform requires a Node host platform"},
		{"adamic_node_", "columns", "", "process.stdout.columns requires terminal size"},
		{"adamic_node_", "memory_usage", "", "process.memoryUsage requires allocator observations"},
		{"adamic_node_", "argv", "", "process.argv requires executable paths"},
		{"adamic_fs_file_", "open", "&path, &flag, 0600", "file creation permission bits are not supported"},
		{"adamic_fs_file_", "mkdir", "&path, false, 0700", "directory creation permission bits are not supported"},
		{"adamic_fs_file_", "utimes", "&path, 1000.5, 1000.5", "fs.utimesSync timestamp precision or range is unavailable"},
	}
	for _, probe := range probes[index : index+1] {
		t.Run(probe.suffix, func(t *testing.T) {
			source := fmt.Sprintf("#include \"adamic.h\"\n#define MEMBER %s ## %s\nint main(void) { adamic_string path = ADAMIC_STRING(\"unused\"), flag = ADAMIC_STRING(\"w\"); adamic_start(0, NULL); (void)MEMBER(%s); return 0; }\n", probe.prefix, probe.suffix, probe.arguments)
			if probe.suffix == "mkdtemp" {
				// The runtime exports temporary-directory operations only in its host feature.
				source = "#define ADAMIC_NODE_HOST 1\n" + source
			}
			binary := filepath.Join(t.TempDir(), "runtime-refusal.wasm")
			buildWASI(t, source, binary)
			actual := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), binary)
			if actual.exitCode != 70 || len(actual.stdout) != 0 || string(actual.stderr) != "adamic: panic: wasm32-wasi: "+probe.reason+"\n" {
				t.Fatalf("runtime refusal: %+v", actual)
			}
		})
	}
}

func wasiFileAgreesWithNode(t *testing.T, index int) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/wasi.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fsFileFixtures[index : index+1] {
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

func TestWASIInputAgreesWithNode_000(t *testing.T) { wasiInputAgreesWithNode(t, 0) }

func TestWASIInputAgreesWithNode_001(t *testing.T) { wasiInputAgreesWithNode(t, 1) }

func TestWASIInputAgreesWithNode_002(t *testing.T) { wasiInputAgreesWithNode(t, 2) }

func TestWASIInputAgreesWithNode_003(t *testing.T) { wasiInputAgreesWithNode(t, 3) }

func TestWASIInputAgreesWithNode_004(t *testing.T) { wasiInputAgreesWithNode(t, 4) }

func TestWASIInputAgreesWithNode_005(t *testing.T) { wasiInputAgreesWithNode(t, 5) }

func TestWASIInputAgreesWithNode_006(t *testing.T) { wasiInputAgreesWithNode(t, 6) }

func TestWASIInputAgreesWithNode_007(t *testing.T) { wasiInputAgreesWithNode(t, 7) }

func TestWASIInputAgreesWithNode_008(t *testing.T) { wasiInputAgreesWithNode(t, 8) }

func TestWASIInputAgreesWithNode_009(t *testing.T) { wasiInputAgreesWithNode(t, 9) }

func TestWASIInputAgreesWithNode_010(t *testing.T) { wasiInputAgreesWithNode(t, 10) }

func TestWASIInputAgreesWithNode_011(t *testing.T) { wasiInputAgreesWithNode(t, 11) }

func TestWASIInputAgreesWithNode_012(t *testing.T) { wasiInputAgreesWithNode(t, 12) }

func TestWASIInputAgreesWithNode_013(t *testing.T) { wasiInputAgreesWithNode(t, 13) }

func TestWASIInputAgreesWithNode_014(t *testing.T) { wasiInputAgreesWithNode(t, 14) }

func TestWASIInputAgreesWithNode_015(t *testing.T) { wasiInputAgreesWithNode(t, 15) }

func TestWASIInputAgreesWithNode_016(t *testing.T) { wasiInputAgreesWithNode(t, 16) }

func TestWASIInputAgreesWithNode_017(t *testing.T) { wasiInputAgreesWithNode(t, 17) }

func TestWASIInputAgreesWithNode_018(t *testing.T) { wasiInputAgreesWithNode(t, 18) }

func TestWASIInputAgreesWithNode_019(t *testing.T) { wasiInputAgreesWithNode(t, 19) }

func TestWASIInputAgreesWithNode_020(t *testing.T) { wasiInputAgreesWithNode(t, 20) }

func TestWASIHostRuntimeRefusal_000(t *testing.T) { wasiHostRuntimeRefusal(t, 0) }

func TestWASIHostRuntimeRefusal_001(t *testing.T) { wasiHostRuntimeRefusal(t, 1) }

func TestWASIHostRuntimeRefusal_002(t *testing.T) { wasiHostRuntimeRefusal(t, 2) }

func TestWASIHostRuntimeRefusal_003(t *testing.T) { wasiHostRuntimeRefusal(t, 3) }

func TestWASIHostRuntimeRefusal_004(t *testing.T) { wasiHostRuntimeRefusal(t, 4) }

func TestWASIHostRuntimeRefusal_005(t *testing.T) { wasiHostRuntimeRefusal(t, 5) }

func TestWASIHostRuntimeRefusal_006(t *testing.T) { wasiHostRuntimeRefusal(t, 6) }

func TestWASIHostRuntimeRefusal_007(t *testing.T) { wasiHostRuntimeRefusal(t, 7) }

func TestWASIHostRuntimeRefusal_008(t *testing.T) { wasiHostRuntimeRefusal(t, 8) }

func TestWASIFileAgreesWithNode_000(t *testing.T) { wasiFileAgreesWithNode(t, 0) }

func TestWASIFileAgreesWithNode_001(t *testing.T) { wasiFileAgreesWithNode(t, 1) }

func TestWASIFileAgreesWithNode_002(t *testing.T) { wasiFileAgreesWithNode(t, 2) }

func TestWASIFileAgreesWithNode_003(t *testing.T) { wasiFileAgreesWithNode(t, 3) }

func TestWASIFileAgreesWithNode_004(t *testing.T) { wasiFileAgreesWithNode(t, 4) }

func TestWASIHostShardCoverage(t *testing.T) {
	if len(inputFixtures) != 21 || len(fsFileFixtures) != 6 {
		t.Fatalf("host fixture registry changed: input=%d files=%d; regenerate host shards", len(inputFixtures), len(fsFileFixtures))
	}
}

func TestWASIFileAgreesWithNode_005(t *testing.T) { wasiFileAgreesWithNode(t, 5) }

func TestWASIHostRefusalClassification(t *testing.T) {
	reasons := []string{"file creation permission bits are not supported", "directory creation permission bits are not supported", "fs.utimesSync timestamp precision or range is unavailable"}
	for _, reason := range reasons {
		exact := "adamic: panic: wasm32-wasi: " + reason + "\n"
		if wasiRefusalReason(run{exitCode: 70, stderr: []byte(exact)}) != reason {
			t.Fatal("owned refusal was lost")
		}
		for _, bad := range []run{{exitCode: 0, stderr: []byte(exact)}, {exitCode: 23, stderr: []byte(exact)}, {exitCode: 70, stderr: []byte(exact + "unexpected\n")}, {exitCode: 70, stderr: []byte(reason)}} {
			if wasiRefusalReason(bad) != "" {
				t.Fatalf("ordinary failure was hidden as a refusal: %+v", bad)
			}
		}
	}
	if wasiRefusalReason(run{exitCode: 70, stderr: []byte("adamic: panic: something else\n")}) != "" {
		t.Fatal("ordinary panic was hidden")
	}
}
