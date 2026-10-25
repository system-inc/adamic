package worker

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Only find prepared products here; every owned case makes fresh backend observations.
func runWorkerEncodeJSONShard(t *testing.T, shard int) {
	family := "TestWorkerEncodeJSON"
	if !workerJSONShardHasCases(family, shard) {
		return
	}
	for _, tool := range []string{"go", "clang", "node"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Fatalf("missing %s: %v", tool, err)
		}
	}
	if err := exec.Command("node", "--input-type=module", "-e", "import {stripTypeScriptTypes,registerHooks} from 'node:module'; if(typeof stripTypeScriptTypes !== 'function'||typeof registerHooks !== 'function') process.exit(1)").Run(); err != nil {
		t.Fatalf("missing Node 24 type stripping and module hooks: %v", err)
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	jsonTypesPath := ""
	run := func(name string, args ...string) []byte {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Dir = root
		if name == "node" {
			for index, arg := range args {
				if arg == "oracle/node.mjs" && index+1 < len(args) {
					if jsonTypesPath == "" {
						jsonTypesPath = jsonTypes(t)
					}
					cmd.Env = append(os.Environ(), "ADAMIC_JSON_TYPES="+jsonTypesPath)
				}
			}
		}
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("%s %v: %v\nstdout: %s\nstderr: %s", name, args, err, &stdout, &stderr)
		}
		return stdout.Bytes()
	}
	fixtures, err := filepath.Glob(filepath.Join(root, "internal/oracle/testdata/json_encode_*.a"))
	if err != nil {
		t.Fatal(err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no encodeJson fixtures selected")
	}
	for _, fixture := range fixtures {
		if !workerJSONCaseOwns(family, shard, "fixture/"+filepath.Base(fixture)) {
			continue
		}
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
				compiler := workerCompilerProduct(t, nil, false)
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
			product := workerJSONProduct(t, "encode/"+filepath.Base(fixture), false)
			generated = filepath.Join(directory, "generated")
			if err := os.CopyFS(generated, os.DirFS(filepath.Join(product, "worker"))); err != nil {
				t.Fatal(err)
			}
			native := filepath.Join(product, "native")
			expected := run(native)
			oracle := run("node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", fixture)
			actual := run("node", "internal/worker/json_decode_run.mjs", filepath.Join(generated, "handler.mjs"))
			if !bytes.Equal(expected, oracle) || !bytes.Equal(expected, actual) {
				t.Fatalf("native:\n%s\nNode:\n%s\nWorker:\n%s", expected, oracle, actual)
			}
			t.Log("native, Node source and generated Worker output identical")
		})
	}

	inputs := workerJSONCorpusProduct(t, "encode", false)
	product := workerJSONProduct(t, "encode/corpus", false)
	native := filepath.Join(product, "native")
	for start := 0; start < 40000; start += 1000 {
		key := fmt.Sprintf("corpus/%05d", start)
		if !workerJSONCaseOwns(family, shard, key) {
			continue
		}
		t.Run(key, func(t *testing.T) {
			corpus := filepath.Join(inputs, fmt.Sprintf("%05d.json", start))
			offsets := workerJSONBlockOffsets(start)
			offset := workerJSONOffsetsArgument(offsets)
			var expected, oracle []byte
			for _, original := range offsets {
				part := filepath.Join(inputs, fmt.Sprintf("part-%05d.json", original))
				position := strconv.Itoa(original)
				expected = append(expected, run(native, part, position)...)
				oracle = append(oracle, run("node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", "internal/worker/testdata/json_encode/native.a", part, position)...)
			}
			if !bytes.Equal(expected, oracle) {
				t.Fatal("native corpus output differs from Node source")
			}
			stringify, err := os.ReadFile(filepath.Join(inputs, fmt.Sprintf("%05d.txt", start)))
			if err != nil || !bytes.Equal(expected, stringify) {
				t.Fatalf("native encoding differs from JSON.stringify: %v", err)
			}

			if workerJSONProbe(family, key) {
				expected = append([]byte("planted wrong case\n"), expected...)
			}
			observations := filepath.Join(t.TempDir(), "expected.txt")
			if err := os.WriteFile(observations, expected, 0600); err != nil {
				t.Fatal(err)
			}
			t.Log(string(run("node", "internal/worker/json_decode_run.mjs", filepath.Join(product, "worker/handler.mjs"), corpus, observations, offset)))
		})
	}
	// Mutants retain all 40,000 potential witnesses and fresh native observations.

	for _, mutant := range workerJSONEncodeMutants {
		if !workerJSONCaseOwns(family, shard, "mutant/"+mutant.name) {
			continue
		}
		t.Run(mutant.name, func(t *testing.T) {
			generated := filepath.Join(t.TempDir(), "worker")
			if err := os.CopyFS(generated, os.DirFS(filepath.Join(product, "worker"))); err != nil {
				t.Fatal(err)
			}
			runtimePath := filepath.Join(generated, "adamic.mjs")
			original, err := os.ReadFile(runtimePath)
			if err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(string(original), mutant.before) {
				t.Fatal("mutant target missing")
			}
			write(t, runtimePath, strings.ReplaceAll(string(original), mutant.before, mutant.after))
			t.Cleanup(func() { write(t, runtimePath, string(original)) })
			var output []byte
			var caught bool
			// The old comparison also stopped at its first disagreement. Use
			// bounded blocks so parsing the outer 40,000-text array is not an
			// expensive shared prelude. An escaped mutant visits every text.
			starts := workerJSONMutantBlocks()
			for _, start := range starts {
				corpus := filepath.Join(inputs, fmt.Sprintf("%05d.json", start))
				offsets := workerJSONBlockOffsets(start)
				offset := workerJSONOffsetsArgument(offsets)
				var expected []byte
				for _, original := range offsets {
					part := filepath.Join(inputs, fmt.Sprintf("part-%05d.json", original))
					expected = append(expected, run(native, part, strconv.Itoa(original))...)
				}
				observations := filepath.Join(t.TempDir(), "expected.txt")
				if err := os.WriteFile(observations, expected, 0600); err != nil {
					t.Fatal(err)
				}
				command := exec.CommandContext(t.Context(), "node", filepath.Join(root, "internal/worker/json_decode_run.mjs"), filepath.Join(generated, "handler.mjs"), corpus, observations, offset)
				var err error
				output, err = command.CombinedOutput()
				if err != nil {
					if !bytes.Contains(output, []byte("native disagreement")) {
						t.Fatalf("mutant failed outside its native comparison: %v\n%s", err, output)
					}
					caught = true
					break
				}
			}
			if !caught {
				t.Fatal("mutant escaped native comparison of all 40,000 texts")
			}
			t.Logf("caught %s by native and JSON.stringify corpus comparison", mutant.name)
		})
	}
}

var workerJSONEncodeMutants = []struct{ name, before, after string }{
	{"omit-falsy-optional", "field.optional && value[name] === undefined", "field.optional && !value[name]"},
	{"reverse-fields", "const entries = [];\n\t\t\t\t\tfor (const field of fields) {", "const entries = [];\n\t\t\t\t\tfor (const field of [...fields].reverse()) {"},
	{"replace-lone-surrogate", "return JSON.stringify(value);", "return JSON.stringify(typeof value === 'string' ? value.replace(/[\\uD800-\\uDFFF]/g, '\\uFFFD') : value);"},
	{"write-infinity", "return JSON.stringify(value);", "return typeof value === 'number' && !Number.isFinite(value) ? String(value) : JSON.stringify(value);"},
}
