package worker

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Not parallel: native corpus and fixture builds share the CLI and are bounded integration work.
func TestWorkerEncodeJSON(t *testing.T) {
	if os.Getenv("ADAMIC_TEST_WORKER_JSON") != "1" {
		t.Skip("set ADAMIC_TEST_WORKER_JSON=1 for native/Node/Worker JSON encoding comparisons")
	}
	for _, tool := range []string{"go", "clang", "node"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("missing %s: %v", tool, err)
		}
	}
	if err := exec.Command("node", "--input-type=module", "-e", "import {stripTypeScriptTypes,registerHooks} from 'node:module'; if(typeof stripTypeScriptTypes !== 'function'||typeof registerHooks !== 'function') process.exit(1)").Run(); err != nil {
		t.Skipf("missing Node 24 type stripping and module hooks: %v", err)
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	run := func(name string, args ...string) []byte {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Dir = root
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("%s %v: %v\nstdout: %s\nstderr: %s", name, args, err, &stdout, &stderr)
		}
		return stdout.Bytes()
	}
	compiler := filepath.Join(scratch, "adamic")
	run("go", "build", "-o", compiler, "./cmd/adamic")
	fixtures, err := filepath.Glob(filepath.Join(root, "internal/oracle/testdata/json_encode_*.a"))
	if err != nil {
		t.Fatal(err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no encodeJson fixtures selected")
	}
	for _, fixture := range fixtures {
		t.Run(filepath.Base(fixture), func(t *testing.T) {
			directory := t.TempDir()
			source, err := os.ReadFile(fixture)
			if err != nil {
				t.Fatal(err)
			}
			entry := filepath.Join(directory, "entry.a")
			write(t, entry, string(source)+"\nimport type { HttpRequest, HttpResponse } from 'adamic/http';\nexport function handle(request: HttpRequest): HttpResponse { return {status:200,headers:[],body:request.body}; }\n")
			generated := filepath.Join(directory, "worker")
			if strings.HasSuffix(fixture, "_notyet.a") {
				// The merged encoder includes a refusal fixture alongside successes.
				for _, args := range [][]string{{"worker", entry, "--out", generated}, {"build", fixture, "-o", filepath.Join(directory, "native")}} {
					cmd := exec.Command(compiler, args...)
					cmd.Dir = root
					output, err := cmd.CombinedOutput()
					if err == nil || !bytes.Contains(output, []byte("encodeJson of a class instance is not yet supported; describe the data with an interface")) {
						t.Fatalf("expected class-instance NotYet for %v: %v\n%s", args, err, output)
					}
				}
				t.Log("Worker and native both retain the class-instance NotYet refusal")
				return
			}
			run(compiler, "worker", entry, "--out", generated)
			native := filepath.Join(directory, "native")
			run(compiler, "build", fixture, "-o", native)
			expected := run(native)
			oracle := run("node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", fixture)
			actual := run("node", "internal/worker/json_decode_run.mjs", filepath.Join(generated, "handler.mjs"))
			if !bytes.Equal(expected, oracle) || !bytes.Equal(expected, actual) {
				t.Fatalf("native:\n%s\nNode:\n%s\nWorker:\n%s", expected, oracle, actual)
			}
			t.Log("native, Node source and generated Worker output identical")
		})
	}
	corpus := filepath.Join(scratch, "corpus.json")
	witness := filepath.Join(scratch, "stringify.txt")
	run("node", "internal/worker/json_encode_corpus.mjs", corpus, witness)
	native := filepath.Join(scratch, "native")
	run(compiler, "build", "internal/worker/testdata/json_encode/native.a", "-o", native)
	expected := run(native, corpus)
	stringify, err := os.ReadFile(witness)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(expected, stringify) {
		t.Fatal("native encoding differs from Node JSON.stringify on declared-order values")
	}
	oracle := run("node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", "internal/worker/testdata/json_encode/native.a", corpus)
	if !bytes.Equal(expected, oracle) {
		t.Fatal("native corpus output differs from Node source")
	}
	observations := filepath.Join(scratch, "expected.txt")
	if err := os.WriteFile(observations, expected, 0644); err != nil {
		t.Fatal(err)
	}
	generated := filepath.Join(scratch, "worker")
	run(compiler, "worker", "internal/worker/testdata/json_encode/worker.a", "--out", generated)
	t.Log(string(run("node", "internal/worker/json_decode_run.mjs", filepath.Join(generated, "handler.mjs"), corpus, observations)))
	runtimePath := filepath.Join(generated, "adamic.mjs")
	original, err := os.ReadFile(runtimePath)
	if err != nil {
		t.Fatal(err)
	}
	mutants := []struct{ name, before, after string }{
		{"omit-falsy-optional", "field.optional && value[name] === undefined", "field.optional && !value[name]"},
		{"reverse-fields", "const entries = [];\n\t\t\t\t\tfor (const field of fields) {", "const entries = [];\n\t\t\t\t\tfor (const field of [...fields].reverse()) {"},
		{"replace-lone-surrogate", "return JSON.stringify(value);", "return JSON.stringify(typeof value === 'string' ? value.replace(/[\\uD800-\\uDFFF]/g, '\\uFFFD') : value);"},
		{"write-infinity", "return JSON.stringify(value);", "return typeof value === 'number' && !Number.isFinite(value) ? String(value) : JSON.stringify(value);"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			if !strings.Contains(string(original), mutant.before) {
				t.Fatal("mutant target missing")
			}
			write(t, runtimePath, strings.ReplaceAll(string(original), mutant.before, mutant.after))
			t.Cleanup(func() { write(t, runtimePath, string(original)) })
			command := exec.Command("node", filepath.Join(root, "internal/worker/json_decode_run.mjs"), filepath.Join(generated, "handler.mjs"), corpus, observations)
			output, err := command.CombinedOutput()
			if err == nil || !bytes.Contains(output, []byte("native disagreement")) {
				t.Fatalf("mutant escaped native comparison: %v\n%s", err, output)
			}
			t.Logf("caught %s by native and JSON.stringify corpus comparison", mutant.name)
		})
	}
}
