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
func runWorkerDecodeJSONShard(test *testing.T, shard int) {
	family := "TestWorkerDecodeJSON"
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
	fixtures, globError := filepath.Glob(filepath.Join(root, "internal/oracle/testdata/json_decode_*.a"))
	if globError != nil {
		test.Fatal(globError)
	}
	if len(fixtures) == 0 {
		test.Fatal("no decodeJson fixtures selected")
	}
	fixtures = append(fixtures, filepath.Join(root, "internal/worker/testdata/json_decode/fast_risks.a"))
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
			// Main's enumeration proof rejects these empty-object fixtures. Pin
			// that same NotYet on both targets rather than bypassing its rule.
			if name := filepath.Base(
				fixture,
			); name == "json_decode_empty_object.a" || name == "json_decode_coverage_empty.a" {
				compiler := workerCompilerProduct(test, nil, false)
				for _, args := range [][]string{
					{"worker", entry, "--out", generated},
					{"build", fixture, "-o", filepath.Join(directory, "native")},
				} {
					command := exec.Command(compiler, args...)
					command.Dir = root
					output, combinedOutputError := command.CombinedOutput()
					if combinedOutputError == nil ||
						!bytes.Contains(output, []byte("stage 0 can't lower Object.keys on a non-object yet")) {
						test.Fatalf(
							"want main's empty-object enumeration NotYet for %v: %v\n%s",
							args,
							combinedOutputError,
							output,
						)
					}
				}
				test.Log("Worker and native retain main's empty-object enumeration NotYet")
				return
			}
			product := workerJSONProduct(test, "decode/"+filepath.Base(fixture), false)
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
			if filepath.Base(fixture) == "fast_risks.a" {
				// Exercise the same explicit risks through the original parser too.
				runtimePath := filepath.Join(generated, "adamic.mjs")
				runtime, readError := os.ReadFile(runtimePath)
				if readError != nil {
					test.Fatal(readError)
				}
				before := "const value = parseWorkerJsonFast(text);"
				if strings.Count(string(runtime), before) != 1 {
					test.Fatal("fast parser call target changed")
				}
				write(
					test,
					runtimePath,
					strings.Replace(string(runtime), before, "const value = parseWorkerJson(text);", 1),
				)
				legacy := run("node", "internal/worker/json_decode_run.mjs", filepath.Join(generated, "handler.mjs"))
				if !bytes.Equal(expected, legacy) {
					test.Fatalf("legacy parser differs from native:\nnative: %s\nlegacy: %s", expected, legacy)
				}
				test.Log(
					"explicit duplicates, depth and UTF-16 risks agree across legacy parser, fast path, native and Node source",
				)
			}
			test.Log("native, Node source and generated Worker output identical")
		})
	}

	inputs := workerJSONCorpusProduct(test, "decode", false)
	product := workerJSONProduct(test, "decode/corpus", false)
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
					"internal/worker/testdata/json_decode/native.a",
					part,
					position,
				)...)
			}
			if !bytes.Equal(expected, oracle) {
				test.Fatal("native corpus output differs from Node source")
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

	for _, mutant := range workerJSONDecodeMutants {
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
			mutated := strings.ReplaceAll(string(original), mutant.before, mutant.after)
			// V8 owns duplicate selection on success. Retain the old duplicate
			// mutant as an explicit fallback-parser regression check.
			if mutant.name == "wrong-descriptor-cache" {
				mutated = strings.Replace(
					mutated,
					"workerDecodeWalkers.set(descriptor, walker)",
					"workerDecodeWalkers.set(workerDecodeWalkers, walker)",
					1,
				)
			}
			if mutant.name == "duplicate-first" {
				mutated = strings.Replace(mutated, "parsed = JSON.parse(text);", "parsed = parseWorkerJson(text);", 1)
			}
			write(test, runtimePath, mutated)
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
			if mutant.name == "skip-depth-scan" {
				depth129 := strings.Repeat("[", 129) + "0" + strings.Repeat("]", 129)
				if !bytes.Contains(output, []byte(depth129)) {
					test.Fatalf("depth scan mutant must fail on the depth-129 text: %s", output)
				}
				test.Log("fast depth scan omission caught on the 129-container text")
			}
			test.Logf("caught %s by native corpus comparison", mutant.name)
		})
	}
}

var workerJSONDecodeMutants = []struct{ name, before, after string }{
	{
		"duplicate-first",
		"object[key] = value(depth + 1);",
		"const parsed = value(depth + 1); if (!Object.hasOwn(object, key)) object[key] = parsed;",
	},
	{"wrong-descriptor-cache", "workerDecodeWalkers.get(descriptor)", "workerDecodeWalkers.get(workerDecodeWalkers)"},
	{"skip-depth-scan", "if(!withinWorkerJsonDepth(text)) return parseWorkerJson(text);", ""},
	{"depth-128-accepted", "if(depth >= 128)", "if (depth >= 129)"},
	{"missing-required", "if(field.optional) continue;", "if (field.optional || name === 'required') continue;"},
	{
		"reject-extra",
		"const result = {};",
		"const result = {}; if (Object.keys(value).some(name => !fields.some(field => field.name === name))) workerJsonMismatch(expected, path, value);",
	},
	{"reword-error", "missing field ${name}", "required field ${name} is missing"},
}
