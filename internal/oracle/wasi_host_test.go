package oracle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func buildWASI(test *testing.T, source, binary string) string {
	test.Helper()
	if err := native.Build(source, binary, native.Options{Target: "wasm32-wasi"}); err != nil {
		var refused *native.TargetRefused
		if errors.As(err, &refused) {
			return refused.Error()
		}
		test.Fatal(err)
	}
	return ""
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

// Token pasting bypasses the target-aware generated-C check, so these calls
// exercise the runtime refusal that a caller without target information sees.
func wasiHostRuntimeRefusal(test *testing.T, index int) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	probes := []struct{ prefix, suffix, arguments, reason string }{
		{"adamic_fs_file_", "mkdtemp", "NULL", "fs.mkdtempSync requires temporary directory creation"},
		{"adamic_node_", "pid", "", "process.pid requires process identifiers"},
		{"adamic_node_", "platform", "", "process.platform requires a Node host platform"},
		{"adamic_node_", "memory_usage", "", "process.memoryUsage requires allocator observations"},
		{"adamic_node_", "argv", "", "process.argv requires executable paths"},
		{"adamic_fs_file_", "open", "&path, &flag, 0600", "file creation permission bits are not supported"},
		{"adamic_fs_file_", "mkdir", "&path, false, 0700", "directory creation permission bits are not supported"},
		{
			"adamic_fs_file_",
			"utimes",
			"&path, 1000.5, 1000.5",
			"fs.utimesSync timestamp precision or range is unavailable",
		},
	}
	for _, probe := range probes[index : index+1] {
		test.Run(probe.suffix, func(test *testing.T) {
			source := fmt.Sprintf(
				"#include \"adamic.h\"\n#define MEMBER %s ## %s\nint main(void) { adamic_string path = ADAMIC_STRING(\"unused\"), flag = ADAMIC_STRING(\"w\"); adamic_start(0, NULL); (void)MEMBER(%s); return 0; }\n",
				probe.prefix,
				probe.suffix,
				probe.arguments,
			)
			if probe.suffix == "mkdtemp" {
				// The runtime exports temporary-directory operations only in its host feature.
				source = "#define ADAMIC_NODE_HOST 1\n" + source
			}
			binary := filepath.Join(test.TempDir(), "runtime-refusal.wasm")
			if reason := buildWASI(test, source, binary); reason != "" {
				test.Fatalf(
					"%s: expected runtime refusal %q, unexpectedly refused at build: %s",
					probe.suffix,
					probe.reason,
					reason,
				)
			}
			actual := execute(
				test,
				"node",
				"--disable-warning=ExperimentalWarning",
				filepath.Join(repository, "oracle/wasi.mjs"),
				binary,
			)
			if actual.exitCode != 70 ||
				len(actual.stdout) != 0 ||
				string(actual.stderr) != "adamic: panic: wasm32-wasi: "+probe.reason+"\n" {
				test.Fatalf("%s: want runtime refusal %q, got %+v", probe.suffix, probe.reason, actual)
			}
		})
	}
}

func wasiFileAgreesWithNode(test *testing.T, index int) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/wasi.mjs"))
	if err != nil {
		test.Fatal(err)
	}
	for _, fixture := range fsFileFixtures[index : index+1] {
		test.Run(fixture, func(test *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_fs_file_"+fixture+".a"))
			if err != nil {
				test.Fatal(err)
			}
			program, err := lowered(test, path)
			if err != nil {
				test.Fatal(err)
			}
			shared := sharedDirectory(test)
			binary := filepath.Join(shared, "program.wasm")
			if reason := buildWASI(test, native.C(program), binary); reason != "" {
				assertWASIRefusal(test, fixture, reason, wasiFileRefusedFixtures)
				return
			}
			nodeRun, wasmRun := fsFilePrepare(test, shared, "node"), fsFilePrepare(test, shared, "wasi")
			expected := onNodeWith(test, nodeRun, path)
			actual := executeInput(
				test,
				wasmRun,
				nil,
				"node",
				append([]string{"--disable-warning=ExperimentalWarning", runner, binary}, wasmRun.arguments...)...,
			)
			if assertWASIRefusal(test, fixture, wasiRefusalReason(actual), wasiFileRefusedFixtures) {
				return
			}
			if difference := disagreement(expected, actual); difference != "" {
				test.Errorf(
					"%s\nnode: exit %d, stdout %q, stderr %q\nwasi: exit %d, stdout %q, stderr %q",
					difference,
					expected.exitCode,
					expected.stdout,
					expected.stderr,
					actual.exitCode,
					actual.stdout,
					actual.stderr,
				)
			}
			if difference := filesDiffer(
				fsFileSnapshot(test, nodeRun.arguments[0]),
				fsFileSnapshot(test, wasmRun.arguments[0]),
			); difference != "" {
				test.Errorf("filesystem effects: %s", difference)
			}
		})
	}
}

func TestWASIHostRuntimeRefusal_000(test *testing.T) {
	test.Parallel()
	wasiHostRuntimeRefusal(test, 0)
}

func TestWASIHostRuntimeRefusal_001(test *testing.T) {
	test.Parallel()
	wasiHostRuntimeRefusal(test, 1)
}

func TestWASIHostRuntimeRefusal_002(test *testing.T) {
	test.Parallel()
	wasiHostRuntimeRefusal(test, 2)
}

func TestWASIHostRuntimeRefusal_003(test *testing.T) {
	test.Parallel()
	wasiHostRuntimeRefusal(test, 3)
}

func TestWASIHostRuntimeRefusal_004(test *testing.T) {
	test.Parallel()
	wasiHostRuntimeRefusal(test, 4)
}

func TestWASIHostRuntimeRefusal_005(test *testing.T) {
	test.Parallel()
	wasiHostRuntimeRefusal(test, 5)
}

func TestWASIHostRuntimeRefusal_006(test *testing.T) {
	test.Parallel()
	wasiHostRuntimeRefusal(test, 6)
}

func TestWASIHostRuntimeRefusal_007(test *testing.T) {
	test.Parallel()
	wasiHostRuntimeRefusal(test, 7)
}

func TestWASIFileAgreesWithNode_000(test *testing.T) {
	test.Parallel()
	wasiFileAgreesWithNode(test, 0)
}

func TestWASIFileAgreesWithNode_001(test *testing.T) {
	test.Parallel()
	wasiFileAgreesWithNode(test, 1)
}

func TestWASIFileAgreesWithNode_002(test *testing.T) {
	test.Parallel()
	wasiFileAgreesWithNode(test, 2)
}

func TestWASIFileAgreesWithNode_003(test *testing.T) {
	test.Parallel()
	wasiFileAgreesWithNode(test, 3)
}

func TestWASIFileAgreesWithNode_004(test *testing.T) {
	test.Parallel()
	wasiFileAgreesWithNode(test, 4)
}

func TestWASIHostShardCoverage(test *testing.T) {
	test.Parallel()
	if len(inputFixtures) != 22 || len(fsFileFixtures) != 7 {
		test.Fatalf(
			"host fixture registry changed: input=%d files=%d; regenerate host shards",
			len(inputFixtures),
			len(fsFileFixtures),
		)
	}
}

func TestWASIFileAgreesWithNode_005(test *testing.T) {
	test.Parallel()
	wasiFileAgreesWithNode(test, 5)
}

func TestWASIFileAgreesWithNode_006(test *testing.T) {
	test.Parallel()
	wasiFileAgreesWithNode(test, 6)
}

func TestWASIHostRefusalClassification(test *testing.T) {
	test.Parallel()
	reasons := []string{
		"file creation permission bits are not supported",
		"directory creation permission bits are not supported",
		"fs.utimesSync timestamp precision or range is unavailable",
	}
	for _, reason := range reasons {
		exact := "adamic: panic: wasm32-wasi: " + reason + "\n"
		if wasiRefusalReason(run{exitCode: 70, stderr: []byte(exact)}) != reason {
			test.Fatal("owned refusal was lost")
		}
		for _, bad := range []run{
			{exitCode: 0, stderr: []byte(exact)},
			{exitCode: 23, stderr: []byte(exact)},
			{exitCode: 70, stderr: []byte(exact + "unexpected\n")},
			{exitCode: 70, stderr: []byte(reason)},
		} {
			if wasiRefusalReason(bad) != "" {
				test.Fatalf("ordinary failure was hidden as a refusal: %+v", bad)
			}
		}
	}
	if wasiRefusalReason(run{exitCode: 70, stderr: []byte("adamic: panic: something else\n")}) != "" {
		test.Fatal("ordinary panic was hidden")
	}
}
