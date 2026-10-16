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
	fsFileFixtures = append(fsFileFixtures, "write_positions")
}

func TestNodeFSWritePositionsWASI(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_fs_file_write_positions.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	shared := sharedDirectory(t)
	truthRun := fsFilePrepare(t, shared, "node")
	truth := onNodeWith(t, truthRun, path)
	binary := filepath.Join(shared, "program.wasm")
	if err := native.Build(native.C(program), binary, native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	how := fsFilePrepare(t, shared, "wasi")
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/wasi.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	got := executeInput(t, how, nil, "node", "--disable-warning=ExperimentalWarning", runner, binary, how.arguments[0])
	if difference := disagreement(truth, got); difference != "" {
		t.Fatalf("%s: Node %q, WASI %q, stderr %q", difference, truth.stdout, got.stdout, got.stderr)
	}
	if difference := filesDiffer(fsFileSnapshot(t, truthRun.arguments[0]), fsFileSnapshot(t, how.arguments[0])); difference != "" {
		t.Fatal(difference)
	}
}

// Both mutations remain well-typed, finish normally, and pass sanitizer and
// leak checks. Only the source's output on Node detects their wrong semantics.
func TestNodeFSConsumedOptionsMutants(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, path, operation, helper string
		input                         bool
	}{
		{"stat options", "node_options_const_capture.a", "stat", `static adamic_object *host_options_mutant(const adamic_string *path,bool throws) { (void)throws;return adamic_fs_file_stat(path,false); }`, false},
		{"empty write position", "node_fs_file_write_positions.a", "write", `static double host_options_mutant(double fd,const adamic_string *value,double position) { return adamic_fs_file_write(fd,value,position<0?0:position); }`, true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", probe.path))
			if err != nil {
				t.Fatal(err)
			}
			if !probe.input {
				// Keep both original and mutant successful: the source catches the
				// required missing-file exception, while the mutant returns instead.
				path = filepath.Join(t.TempDir(), "main.a")
				source := `import {statSync} from 'node:fs'; const options={throwIfNoEntry:true} as const; try {statSync('missing-host-options-mutant-entry',options); console.log('returned');} catch(error) {if(error instanceof Error)console.log('threw');}`
				if err := os.WriteFile(path, []byte(source), 0600); err != nil {
					t.Fatal(err)
				}
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			code := native.C(program)
			if !probe.input {
				truth := onNode(t, path)
				binary := filepath.Join(t.TempDir(), "original")
				if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
					t.Fatal(err)
				}
				got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
				if difference := disagreement(truth, got); difference != "" {
					t.Fatal(difference)
				}
				if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
					if difference := disagreement(truth, onWASI(t, code)); difference != "" {
						t.Fatal(difference)
					}
				}
			}
			mutated := strings.ReplaceAll(code, "adamic_fs_file_"+probe.operation+"(", "host_options_mutant(")
			if code == mutated {
				t.Fatal("mutant changed nothing")
			}
			mutated = strings.Replace(mutated, "#include \"adamic.h\"", "#include \"adamic.h\"\n"+probe.helper, 1)
			shared := sharedDirectory(t)
			binary := filepath.Join(shared, "mutant")
			if err := native.Build(mutated, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			environment := []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}
			var got, truth run
			if probe.input {
				how := fsFilePrepare(t, shared, "mutant-input")
				got = executeInput(t, how, environment, binary, how.arguments...)
				truth = onNodeWith(t, fsFilePrepare(t, shared, "node"), path)
			} else {
				got = executeWith(t, environment, binary)
				truth = onNode(t, path)
			}
			if got.exitCode != 0 || len(got.stderr) != 0 || truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("comparison preconditions failed: mutant %+v, Node %+v", got, truth)
			}
			if bytes.Equal(got.stdout, truth.stdout) {
				t.Fatal("Node did not catch mutant")
			}
			t.Logf("Node alone caught mutation: Node %q, mutant %q; exit 0, clean ASan/UBSan and leaks", truth.stdout, got.stdout)
		})
	}
}

func TestNodeFSRmUndefinedNodeMutant(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"recursive", "force", "maxRetries", "retryDelay"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "main.a")
			source := "import {rmSync} from 'node:fs'; try {rmSync('missing-host-rm-option-entry',{" + name + ":undefined});} catch(error) {if(error instanceof Error)console.log(error.message);}"
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := lowered(t, path)
			var missing *lower.NotYet
			if !errors.As(err, &missing) || !strings.Contains(missing.What, "present as undefined") {
				t.Fatalf("want option validation refusal, got %v", err)
			}
			truth := onNode(t, path)
			if !strings.Contains(string(truth.stdout), "Received undefined") {
				t.Fatalf("Node didn't reject the present option: %+v", truth)
			}
			// Model the old lowering mutant: a present undefined becomes the default.
			fallback := "false"
			if name == "maxRetries" {
				fallback = "0"
			}
			if name == "retryDelay" {
				fallback = "100"
			}
			mutantPath := filepath.Join(t.TempDir(), "main.a")
			if err := os.WriteFile(mutantPath, []byte(strings.Replace(source, name+":undefined", name+":"+fallback, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, mutantPath)
			if err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if got.exitCode != 0 || len(got.stderr) != 0 || truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("mutant failed outside Node comparison: mutant %+v, Node %+v", got, truth)
			}
			if bytes.Equal(got.stdout, truth.stdout) {
				t.Fatal("Node didn't catch collapsed undefined option")
			}
			t.Logf("Node alone caught default substitution: Node %q, mutant %q; exit 0, clean sanitizers and leaks", truth.stdout, got.stdout)
		})
	}
}
