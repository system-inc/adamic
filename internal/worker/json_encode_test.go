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
func runWorkerEncodeJSONShard(test *testing.T, shard int) {
	family := "TestWorkerEncodeJSON"
	if !workerJSONShardHasCases(family, shard) {
		return
	}
	for _, tool := range []string{"go", "clang", "node"} {
		if _, lookPathError := exec.LookPath(tool); lookPathError != nil {
			test.Fatalf("missing %s: %v", tool, lookPathError)
		}
	}
	if runError := exec.Command(
		"node",
		"--input-type=module",
		"-e",
		"import {stripTypeScriptTypes,registerHooks} from 'node:module'; if(typeof stripTypeScriptTypes !== 'function'||typeof registerHooks !== 'function') process.exit(1)",
	).Run(); runError != nil {
		test.Fatalf("missing Node 24 type stripping and module hooks: %v", runError)
	}
	// Reads: internal/oracle/testdata/json_*.a
	root, absolutePathError := filepath.Abs("../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	jsonTypesPath := ""
	run := func(name string, args ...string) []byte {
		test.Helper()
		timeoutContext, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		command := exec.CommandContext(timeoutContext, name, args...)
		command.Dir = root
		if name == "node" {
			for index, arg := range args {
				if arg == "oracle/node.mjs" && index+1 < len(args) {
					if jsonTypesPath == "" {
						jsonTypesPath = jsonTypes(test)
					}
					command.Env = append(os.Environ(), "ADAMIC_JSON_TYPES="+jsonTypesPath)
				}
			}
		}
		var stdout, stderr bytes.Buffer
		command.Stdout = &stdout
		command.Stderr = &stderr
		if runError := command.Run(); runError != nil {
			test.Fatalf("%s %v: %v\nstdout: %s\nstderr: %s", name, args, runError, &stdout, &stderr)
		}
		return stdout.Bytes()
	}
	fixtures, globError := filepath.Glob(filepath.Join(root, "internal/oracle/testdata/json_encode_*.a"))
	if globError != nil {
		test.Fatal(globError)
	}
	if len(fixtures) == 0 {
		test.Fatal("no encodeJson fixtures selected")
	}
	for _, fixture := range fixtures {
		if !workerJSONCaseOwns(family, shard, "fixture/"+filepath.Base(fixture)) {
			continue
		}
		test.Run(filepath.Base(fixture), func(test *testing.T) {
			directory := test.TempDir()
			source, readError := os.ReadFile(fixture)
			if readError != nil {
				test.Fatal(readError)
			}
			entry := filepath.Join(directory, "entry.a")
			write(test, entry, string(source)+
				"\nimport type { HttpRequest, HttpResponse } from 'adamic/http';\nexport function handle(request: HttpRequest): HttpResponse { return {status:200,headers:[],body:request.body}; }\n")
			generated := filepath.Join(directory, "worker")
			if strings.HasSuffix(fixture, "_notyet.a") {
				// The merged encoder includes a refusal fixture alongside successes.
				compiler := workerCompilerProduct(test, nil, false)
				for _, args := range [][]string{
					{"worker", entry, "--out", generated},
					{"build", fixture, "-o", filepath.Join(directory, "native")},
				} {
					command := exec.Command(compiler, args...)
					command.Dir = root
					output, combinedOutputError := command.CombinedOutput()
					if combinedOutputError == nil || !bytes.Contains(output, []byte(
						"encodeJson of a class instance is not yet supported; describe the data with an interface",
					)) {
						test.Fatalf("expected class-instance NotYet for %v: %v\n%s", args, combinedOutputError, output)
					}
				}
				test.Log("Worker and native both retain the class-instance NotYet refusal")
				return
			}
			product := workerJSONProduct(test, "encode/"+filepath.Base(fixture), false)
			generated = filepath.Join(directory, "generated")
			if copyFsError := os.CopyFS(generated, os.DirFS(filepath.Join(product, "worker"))); copyFsError != nil {
				test.Fatal(copyFsError)
			}
			native := filepath.Join(product, "native")
			expected := run(native)
			oracle := run("node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", fixture)
			actual := run("node", "internal/worker/json_decode_run.mjs", filepath.Join(generated, "handler.mjs"))
			if !bytes.Equal(expected, oracle) || !bytes.Equal(expected, actual) {
				test.Fatalf("native:\n%s\nNode:\n%s\nWorker:\n%s", expected, oracle, actual)
			}
			test.Log("native, Node source and generated Worker output identical")
		})
	}

	inputs := workerJSONCorpusProduct(test, "encode", false)
	product := workerJSONProduct(test, "encode/corpus", false)
	native := filepath.Join(product, "native")
	for start := 0; start < 40000; start += 1000 {
		key := fmt.Sprintf("corpus/%05d", start)
		if !workerJSONCaseOwns(family, shard, key) {
			continue
		}
		test.Run(key, func(test *testing.T) {
			corpus := filepath.Join(inputs, fmt.Sprintf("%05d.json", start))
			offsets := workerJSONBlockOffsets(start)
			offset := workerJSONOffsetsArgument(offsets)
			var expected, oracle []byte
			for _, original := range offsets {
				part := filepath.Join(inputs, fmt.Sprintf("part-%05d.json", original))
				position := strconv.Itoa(original)
				expected = append(expected, run(native, part, position)...)
				oracle = append(oracle, run(
					"node",
					"--disable-warning=ExperimentalWarning",
					"oracle/node.mjs",
					"internal/worker/testdata/json_encode/native.a",
					part,
					position,
				)...)
			}
			if !bytes.Equal(expected, oracle) {
				test.Fatal("native corpus output differs from Node source")
			}
			stringify, readError := os.ReadFile(filepath.Join(inputs, fmt.Sprintf("%05d.txt", start)))
			if readError != nil || !bytes.Equal(expected, stringify) {
				test.Fatalf("native encoding differs from JSON.stringify: %v", readError)
			}

			if workerJSONProbe(family, key) {
				expected = append([]byte("planted wrong case\n"), expected...)
			}
			observations := filepath.Join(test.TempDir(), "expected.txt")
			if writeError := os.WriteFile(observations, expected, 0600); writeError != nil {
				test.Fatal(writeError)
			}
			test.Log(string(run(
				"node",
				"internal/worker/json_decode_run.mjs",
				filepath.Join(product, "worker/handler.mjs"),
				corpus,
				observations,
				offset,
			)))
		})
	}
	// Mutants retain all 40,000 potential witnesses and fresh native observations.

	for _, mutant := range workerJSONEncodeMutants {
		if !workerJSONCaseOwns(family, shard, "mutant/"+mutant.name) {
			continue
		}
		test.Run(mutant.name, func(test *testing.T) {
			generated := filepath.Join(test.TempDir(), "worker")
			if copyFsError := os.CopyFS(generated, os.DirFS(filepath.Join(product, "worker"))); copyFsError != nil {
				test.Fatal(copyFsError)
			}
			runtimePath := filepath.Join(generated, "adamic.mjs")
			original, readError := os.ReadFile(runtimePath)
			if readError != nil {
				test.Fatal(readError)
			}

			if !strings.Contains(string(original), mutant.before) {
				test.Fatal("mutant target missing")
			}
			write(test, runtimePath, strings.ReplaceAll(string(original), mutant.before, mutant.after))
			test.Cleanup(func() { write(test, runtimePath, string(original)) })
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
				observations := filepath.Join(test.TempDir(), "expected.txt")
				if writeError := os.WriteFile(observations, expected, 0600); writeError != nil {
					test.Fatal(writeError)
				}
				command := exec.CommandContext(
					test.Context(),
					"node",
					filepath.Join(root, "internal/worker/json_decode_run.mjs"),
					filepath.Join(generated, "handler.mjs"),
					corpus,
					observations,
					offset,
				)
				var combinedOutputError error
				output, combinedOutputError = command.CombinedOutput()
				if combinedOutputError != nil {
					if !bytes.Contains(output, []byte("native disagreement")) {
						test.Fatalf("mutant failed outside its native comparison: %v\n%s", combinedOutputError, output)
					}
					caught = true
					break
				}
			}
			if !caught {
				test.Fatal("mutant escaped native comparison of all 40,000 texts")
			}
			test.Logf("caught %s by native and JSON.stringify corpus comparison", mutant.name)
		})
	}
}

var workerJSONEncodeMutants = []struct{ name, before, after string }{
	{"omit-falsy-optional", "field.optional && value[name] === undefined", "field.optional && !value[name]"},
	{
		"reverse-fields",
		"const entries = [];\n                    for(const field of fields) {",
		"const entries = [];\n\t\t\t\t\tfor (const field of [...fields].reverse()) {",
	},
	{
		"replace-lone-surrogate",
		"return JSON.stringify(value);",
		"return JSON.stringify(typeof value === 'string' ? value.replace(/[\\uD800-\\uDFFF]/g, '\\uFFFD') : value);",
	},
	{
		"write-infinity",
		"return JSON.stringify(value);",
		"return typeof value === 'number' && !Number.isFinite(value) ? String(value) : JSON.stringify(value);",
	},
}
