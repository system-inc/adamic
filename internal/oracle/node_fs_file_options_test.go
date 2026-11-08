package oracle

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	fileSystemFileFixtures = append(fileSystemFileFixtures, "write_positions")
}

func TestNodeFileSystemWritePositionsWasi(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/node_fs_file_write_positions.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	shared := sharedDirectory(test)
	truthRun := fileSystemFilePrepare(test, shared, "node")
	truth := onNodeWith(test, truthRun, path)
	binary := filepath.Join(shared, "program.wasm")
	if buildError := native.Build(native.C(program), binary, native.Options{Target: "wasm32-wasi"}); buildError != nil {
		test.Fatal(buildError)
	}
	how := fileSystemFilePrepare(test, shared, "wasi")
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/wasi.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	got := executeInput(
		test,
		how,
		nil,
		"node",
		"--disable-warning=ExperimentalWarning",
		runner,
		binary,
		how.arguments[0],
	)
	if difference := disagreement(truth, got); difference != "" {
		test.Fatalf("%s: Node %q, WASI %q, stderr %q", difference, truth.stdout, got.stdout, got.stderr)
	}
	if difference := filesDiffer(
		fileSystemFileSnapshot(test, truthRun.arguments[0]),
		fileSystemFileSnapshot(test, how.arguments[0]),
	); difference != "" {
		test.Fatal(difference)
	}
}

// Both mutations remain well-typed, finish normally, and pass sanitizer and
// leak checks. Only the source's output on Node detects their wrong semantics.
func TestNodeFileSystemConsumedOptionsMutants(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct {
		name, path, operation, helper string
		input                         bool
	}{
		{
			"stat options",
			"node_options_const_capture.a",
			"stat",
			`static adamic_object *host_options_mutant(const adamic_string *path,bool throws) { (void)throws;return adamic_fs_file_stat(path,false); }`,
			false,
		},
		{
			"empty write position",
			"node_fs_file_write_positions.a",
			"write",
			`static double host_options_mutant(double fd,const adamic_string *value,double position) { return adamic_fs_file_write(fd,value,position<0?0:position); }`,
			true,
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", probe.path))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			if !probe.input {
				// Keep both original and mutant successful: the source catches the
				// required missing-file exception, while the mutant returns instead.
				path = filepath.Join(test.TempDir(), "main.a")
				source := `import {statSync} from 'node:fs'; const options={throwIfNoEntry:true} as const; try {statSync('missing-host-options-mutant-entry',options); console.log('returned');} catch(error) {if(error instanceof Error)console.log('threw');}`
				if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
					test.Fatal(writeError)
				}
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			code := native.C(program)
			if !probe.input {
				truth := onNode(test, path)
				binary := filepath.Join(test.TempDir(), "original")
				if buildError := native.Build(code, binary, native.Options{Sanitize: true}); buildError != nil {
					test.Fatal(buildError)
				}
				got := executeWith(
					test,
					[]string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"},
					binary,
				)
				if difference := disagreement(truth, got); difference != "" {
					test.Fatal(difference)
				}
				if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
					if difference := disagreement(truth, onWasi(test, code)); difference != "" {
						test.Fatal(difference)
					}
				}
			}
			mutated := strings.ReplaceAll(code, "adamic_fs_file_"+probe.operation+"(", "host_options_mutant(")
			if code == mutated {
				test.Fatal("mutant changed nothing")
			}
			mutated = strings.Replace(mutated, "#include \"adamic.h\"", "#include \"adamic.h\"\n"+probe.helper, 1)
			shared := sharedDirectory(test)
			binary := filepath.Join(shared, "mutant")
			if buildError := native.Build(mutated, binary, native.Options{Sanitize: true}); buildError != nil {
				test.Fatal(buildError)
			}
			environment := []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}
			var got, truth run
			if probe.input {
				how := fileSystemFilePrepare(test, shared, "mutant-input")
				got = executeInput(test, how, environment, binary, how.arguments...)
				truth = onNodeWith(test, fileSystemFilePrepare(test, shared, "node"), path)
			} else {
				got = executeWith(test, environment, binary)
				truth = onNode(test, path)
			}
			if got.exitCode != 0 || len(got.stderr) != 0 || truth.exitCode != 0 || len(truth.stderr) != 0 {
				test.Fatalf("comparison preconditions failed: mutant %+v, Node %+v", got, truth)
			}
			if bytes.Equal(got.stdout, truth.stdout) {
				test.Fatal("Node did not catch mutant")
			}
			test.Logf(
				"Node alone caught mutation: Node %q, mutant %q; exit 0, clean ASan/UBSan and leaks",
				truth.stdout,
				got.stdout,
			)
		})
	}
}

func TestNodeFileSystemRmUndefinedNodeMutant(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"recursive", "force", "maxRetries", "retryDelay"} {
		test.Run(name, func(test *testing.T) {
			path := filepath.Join(test.TempDir(), "main.a")
			source := "import {rmSync} from 'node:fs'; try {rmSync('missing-host-rm-option-entry',{" +
				name +
				":undefined});} catch(error) {if(error instanceof Error)console.log(error.message);}"
			if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
				test.Fatal(writeError)
			}
			_, loweredError := lowered(test, path)
			var missing *lower.NotYet
			if !errors.As(loweredError, &missing) || !strings.Contains(missing.What, "present as undefined") {
				test.Fatalf("want option validation refusal, got %v", loweredError)
			}
			truth := onNode(test, path)
			if !strings.Contains(string(truth.stdout), "Received undefined") {
				test.Fatalf("Node didn't reject the present option: %+v", truth)
			}
			// Model the old lowering mutant: a present undefined becomes the default.
			fallback := "false"
			if name == "maxRetries" {
				fallback = "0"
			}
			if name == "retryDelay" {
				fallback = "100"
			}
			mutantPath := filepath.Join(test.TempDir(), "main.a")
			if writeError := os.WriteFile(
				mutantPath,
				[]byte(strings.Replace(source, name+":undefined", name+":"+fallback, 1)),
				0600,
			); writeError != nil {
				test.Fatal(writeError)
			}
			program, loweredError := lowered(test, mutantPath)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			binary := filepath.Join(test.TempDir(), "mutant")
			if buildError := native.Build(
				native.C(program),
				binary,
				native.Options{Sanitize: true},
			); buildError != nil {
				test.Fatal(buildError)
			}
			got := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if got.exitCode != 0 || len(got.stderr) != 0 || truth.exitCode != 0 || len(truth.stderr) != 0 {
				test.Fatalf("mutant failed outside Node comparison: mutant %+v, Node %+v", got, truth)
			}
			if bytes.Equal(got.stdout, truth.stdout) {
				test.Fatal("Node didn't catch collapsed undefined option")
			}
			test.Logf(
				"Node alone caught default substitution: Node %q, mutant %q; exit 0, clean sanitizers and leaks",
				truth.stdout,
				got.stdout,
			)
		})
	}
}
