package tsgo_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// These controls call the real C archive on all three legs: source on Node,
// emitted JavaScript on Node, and sanitized native Adamic. No checker mock.
func TestErrorsAsValues(t *testing.T) {
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	run := func(name string, command *exec.Cmd) ([]byte, []byte, error) {
		t.Helper()
		command.Dir = repository
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		for suffix, output := range map[string][]byte{"stdout": stdout.Bytes(), "stderr": stderr.Bytes()} {
			if err := os.WriteFile(filepath.Join(directory, name+"."+suffix), output, 0644); err != nil {
				t.Fatal(err)
			}
		}
		return stdout.Bytes(), stderr.Bytes(), err
	}
	must := func(name string, command *exec.Cmd) []byte {
		t.Helper()
		output, report, err := run(name, command)
		if err != nil {
			t.Fatalf("%s: %v\n%s\n%s", name, err, output, report)
		}
		return output
	}
	archive, compiler, addon := filepath.Join(directory, "tsgo.a"), filepath.Join(directory, "adamic"), filepath.Join(directory, "tsgo.node")
	must("archive", exec.Command("go", "build", "-buildmode=c-archive", "-o", archive, "./bridge/tsgo/archive"))
	must("compiler", exec.Command("go", "build", "-o", compiler, "./cmd/adamic"))
	flags := []string{"-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", "-fPIC", "-shared", "-Ibridge/tsgo", "bridge/tsgo/node/node.c", archive, "-lpthread", "-lm", "-o", addon}
	if runtime.GOOS == "darwin" {
		flags = append(flags, "-undefined", "dynamic_lookup")
	} else {
		flags = append(flags, "-ldl")
	}
	must("addon", exec.Command("clang", flags...))
	source := filepath.Join(repository, "bridge/tsgo/testdata/errors.a")
	config, file := filepath.Join(repository, "bridge/tsgo/testdata/tsconfig.json"), filepath.Join(repository, "bridge/tsgo/testdata/sample.ts")
	node := func(entry string) *exec.Cmd {
		command := exec.Command("node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), entry, config, file)
		command.Env = append(os.Environ(), "ADAMIC_TSGO_NODE="+addon)
		return command
	}
	nodeRun := func(name, entry string) []byte {
		t.Helper()
		output, report, err := run(name, node(entry))
		if err != nil || len(report) != 0 {
			t.Fatalf("%s: %v\n%s\n%s", name, err, output, report)
		}
		return output
	}
	truth := nodeRun("source", source)
	for _, required := range []string{
		"refused open missing: no file at ",
		"refused stale inspect: invalid or released checker handle\n",
		"refused stale query: invalid or released checker handle\n",
		"refused stale parts: invalid or released checker handle\n",
		"refused double release: invalid or released checker handle\n",
		"query : \"世界🌍\"\n", "released Ok\n", "caller continued\n",
	} {
		if !bytes.Contains(truth, []byte(required)) {
			t.Fatalf("source failed to exercise %q:\n%s", required, truth)
		}
	}
	if bytes.Contains(truth, []byte("unexpected")) {
		t.Fatalf("source control failed:\n%s", truth)
	}
	javascript := filepath.Join(directory, "errors.mjs")
	if err := os.WriteFile(javascript, must("emit-js", exec.Command(compiler, "js", source)), 0644); err != nil {
		t.Fatal(err)
	}
	if output := nodeRun("javascript", javascript); !bytes.Equal(output, truth) {
		t.Fatalf("JavaScript differs:\n%s\nwant:\n%s", output, truth)
	}
	binary := filepath.Join(directory, "errors")
	must("native-build", exec.Command(compiler, "build", source, "-o", binary, "--tsgo", archive, "--sanitize"))
	command := exec.Command(binary, config, file)
	if runtime.GOOS == "linux" {
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1")
	}
	output, report, err := run("native", command)
	if err != nil || !bytes.Equal(output, truth) || len(report) != 0 {
		t.Fatalf("native: %v\n%s\n%s\nwant:\n%s", err, output, report, truth)
	}
	t.Logf("real checker refusals: native and JavaScript agree with Node, %d bytes, caller continues", len(truth))

	original := filepath.Join(repository, "internal/native/runtime/tsgo.c")
	contents, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	mutations := []struct {
		name  string
		edits map[string]string
	}{
		{"panic", map[string]string{"tsgo_buffer_free(error);\n return failure(region, message);": "tsgo_buffer_free(error);\n adamic_panic(message->bytes, message->length);\n return failure(region, message);"}},
		{"false-ok", map[string]string{
			"static adamic_string error_kind = ADAMIC_STRING(\"Error\");": "/* mutant: Error kind removed */", "static const char *const error_names[] = {\"kind\", \"message\"};": "static const char *const error_names[] = {\"kind\", \"value\"};", "result->slots[0].reference = &error_kind;": "result->slots[0].reference = &ok_kind;"}},
	}
	for _, mutant := range mutations {
		t.Run(mutant.name, func(t *testing.T) {
			text := string(contents)
			for before, after := range mutant.edits {
				if strings.Count(text, before) != 1 {
					t.Fatalf("no unique mutant site: %s", before)
				}
				text = strings.Replace(text, before, after, 1)
			}
			replacement := filepath.Join(directory, mutant.name+".c")
			if err := os.WriteFile(replacement, []byte(text), 0644); err != nil {
				t.Fatal(err)
			}
			overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
			overlayPath := filepath.Join(directory, mutant.name+".json")
			if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
				t.Fatal(err)
			}
			mutatedCompiler := filepath.Join(directory, mutant.name+"-adamic")
			must(mutant.name+"-compiler", exec.Command("go", "build", "-overlay", overlayPath, "-o", mutatedCompiler, "./cmd/adamic"))
			mutatedBinary := filepath.Join(directory, mutant.name+"-native")
			must(mutant.name+"-build", exec.Command(mutatedCompiler, "build", source, "-o", mutatedBinary, "--tsgo", archive))
			got, stderr, err := run(mutant.name, exec.Command(mutatedBinary, config, file))
			if mutant.name == "panic" {
				if err == nil {
					t.Fatalf("panic mutant did not fail: %s", got)
				}
				exit, ok := err.(*exec.ExitError)
				if !ok || exit.ExitCode() != 70 || !bytes.Contains(stderr, []byte("adamic: panic: no file at")) {
					t.Fatalf("panic mutant failed for wrong reason: %v\n%s", err, stderr)
				}
			} else if err != nil || !bytes.Contains(got, []byte("unexpected Ok open missing")) {
				t.Fatalf("false Ok mutant failed for wrong reason: %v\n%s\n%s", err, got, stderr)
			}
			if bytes.Equal(got, truth) && err == nil {
				t.Fatal("mutant escaped Node comparison")
			}
			t.Log("compiled and executed; Node comparison catches " + mutant.name)
		})
	}
	// Refuse the unsupported target before resolving an archive or linking libc.
	refusal, report, err := run("wasi-refusal", exec.Command(compiler, "build", "--target", "wasm32-wasi", source, "-o", filepath.Join(directory, "errors.wasm"), "--tsgo", archive))
	if err == nil || !bytes.Contains(report, []byte("--tsgo is not supported for wasm32-wasi")) {
		t.Fatalf("WASI checker not refused: %v\n%s\n%s", err, refusal, report)
	}
}
