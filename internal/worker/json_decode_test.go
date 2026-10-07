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
func TestWorkerDecodeJSON(t *testing.T) {
	if os.Getenv("ADAMIC_TEST_WORKER_JSON") != "1" {
		t.Skip("set ADAMIC_TEST_WORKER_JSON=1 for native/Node/Worker JSON decoding comparisons")
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
	fixtures, err := filepath.Glob(filepath.Join(root, "internal/oracle/testdata/json_decode_*.a"))
	if err != nil {
		t.Fatal(err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no decodeJson fixtures selected")
	}
	fixtures = append(fixtures, filepath.Join(root, "internal/worker/testdata/json_decode/fast_risks.a"))
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
			run(compiler, "worker", entry, "--out", generated)
			native := filepath.Join(directory, "native")
			run(compiler, "build", fixture, "-o", native)
			expected := run(native)
			oracle := run("node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", fixture)
			actual := run("node", "internal/worker/json_decode_run.mjs", filepath.Join(generated, "handler.mjs"))
			if !bytes.Equal(expected, oracle) || !bytes.Equal(expected, actual) {
				t.Fatalf("native:\n%s\nNode:\n%s\nWorker:\n%s", expected, oracle, actual)
			}
			if filepath.Base(fixture) == "fast_risks.a" {
				// Exercise the same explicit risks through the original parser too.
				runtimePath := filepath.Join(generated, "adamic.mjs")
				runtime, err := os.ReadFile(runtimePath)
				if err != nil {
					t.Fatal(err)
				}
				before := "const value = parseWorkerJsonFast(text);"
				if strings.Count(string(runtime), before) != 1 {
					t.Fatal("fast parser call target changed")
				}
				write(t, runtimePath, strings.Replace(string(runtime), before, "const value = parseWorkerJson(text);", 1))
				legacy := run("node", "internal/worker/json_decode_run.mjs", filepath.Join(generated, "handler.mjs"))
				if !bytes.Equal(expected, legacy) {
					t.Fatalf("legacy parser differs from native:\nnative: %s\nlegacy: %s", expected, legacy)
				}
				t.Log("explicit duplicates, depth and UTF-16 risks agree across legacy parser, fast path, native and Node source")
			}
			t.Log("native, Node source and generated Worker output identical")
		})
	}
	corpus := filepath.Join(scratch, "corpus.json")
	run("node", "internal/worker/json_decode_corpus.mjs", corpus)
	native := filepath.Join(scratch, "native")
	run(compiler, "build", "internal/worker/testdata/json_decode/native.a", "-o", native)
	expected := run(native, corpus)
	oracle := run("node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", "internal/worker/testdata/json_decode/native.a", corpus)
	if !bytes.Equal(expected, oracle) {
		t.Fatal("native corpus output differs from Node source")
	}
	observations := filepath.Join(scratch, "expected.txt")
	if err := os.WriteFile(observations, expected, 0644); err != nil {
		t.Fatal(err)
	}
	generated := filepath.Join(scratch, "worker")
	run(compiler, "worker", "internal/worker/testdata/json_decode/worker.a", "--out", generated)
	t.Log(string(run("node", "internal/worker/json_decode_run.mjs", filepath.Join(generated, "handler.mjs"), corpus, observations)))
	// Mutate the assembled runtime, keeping compiler, corpus and native observations fixed.
	runtimePath := filepath.Join(generated, "adamic.mjs")
	original, err := os.ReadFile(runtimePath)
	if err != nil {
		t.Fatal(err)
	}
	mutants := []struct{ name, before, after string }{
		{"duplicate-first", "object[key] = value(depth + 1);", "const parsed = value(depth + 1); if (!Object.hasOwn(object, key)) object[key] = parsed;"},
		{"wrong-descriptor-cache", "workerDecodeWalkers.get(descriptor)", "workerDecodeWalkers.get(workerDecodeWalkers)"},
		{"skip-depth-scan", "if (!withinWorkerJsonDepth(text)) return parseWorkerJson(text);", ""},
		{"depth-128-accepted", "if (depth >= 128)", "if (depth >= 129)"},
		{"missing-required", "if (field.optional) continue;", "if (field.optional || name === 'required') continue;"},
		{"reject-extra", "const result = {};", "const result = {}; if (Object.keys(value).some(name => !fields.some(field => field.name === name))) workerJsonMismatch(expected, path, value);"},
		{"reword-error", "missing field ${name}", "required field ${name} is missing"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			if !strings.Contains(string(original), mutant.before) {
				t.Fatal("mutant target missing")
			}
			mutated := strings.ReplaceAll(string(original), mutant.before, mutant.after)
			// V8 owns duplicate selection on success. Retain the old duplicate
			// mutant as an explicit fallback-parser regression check.
			if mutant.name == "wrong-descriptor-cache" {
				mutated = strings.Replace(mutated, "workerDecodeWalkers.set(descriptor, walker)", "workerDecodeWalkers.set(workerDecodeWalkers, walker)", 1)
			}
			if mutant.name == "duplicate-first" {
				mutated = strings.Replace(mutated, "parsed = JSON.parse(text);", "parsed = parseWorkerJson(text);", 1)
			}
			write(t, runtimePath, mutated)
			t.Cleanup(func() { write(t, runtimePath, string(original)) })
			command := exec.Command("node", filepath.Join(root, "internal/worker/json_decode_run.mjs"), filepath.Join(generated, "handler.mjs"), corpus, observations)
			output, err := command.CombinedOutput()
			if err == nil || !bytes.Contains(output, []byte("native disagreement")) {
				t.Fatalf("mutant escaped its native comparison: %v\n%s", err, output)
			}
			if mutant.name == "skip-depth-scan" {
				depth129 := strings.Repeat("[", 129) + "0" + strings.Repeat("]", 129)
				if !bytes.Contains(output, []byte(depth129)) {
					t.Fatalf("depth scan mutant must fail on the depth-129 text: %s", output)
				}
				t.Log("fast depth scan omission caught on the 129-container text")
			}
			t.Logf("caught %s by native corpus comparison", mutant.name)
		})
	}
}
