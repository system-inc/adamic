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
func runWorkerDecodeJSONShard(t *testing.T, shard int) {
	family := "TestWorkerDecodeJSON"
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
	// Reads: internal/oracle/testdata/json_*.a
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
	fixtures, err := filepath.Glob(filepath.Join(root, "internal/oracle/testdata/json_decode_*.a"))
	if err != nil {
		t.Fatal(err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no decodeJson fixtures selected")
	}
	fixtures = append(fixtures, filepath.Join(root, "internal/worker/testdata/json_decode/fast_risks.a"))
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
			// Main's enumeration proof rejects these empty-object fixtures. Pin
			// that same NotYet on both targets rather than bypassing its rule.
			if name := filepath.Base(fixture); name == "json_decode_empty_object.a" || name == "json_decode_coverage_empty.a" {
				compiler := workerCompilerProduct(t, nil, false)
				for _, args := range [][]string{{"worker", entry, "--out", generated}, {"build", fixture, "-o", filepath.Join(directory, "native")}} {
					command := exec.Command(compiler, args...)
					command.Dir = root
					output, err := command.CombinedOutput()
					if err == nil || !bytes.Contains(output, []byte("stage 0 can't lower Object.keys on a non-object yet")) {
						t.Fatalf("want main's empty-object enumeration NotYet for %v: %v\n%s", args, err, output)
					}
				}
				t.Log("Worker and native retain main's empty-object enumeration NotYet")
				return
			}
			product := workerJSONProduct(t, "decode/"+filepath.Base(fixture), false)
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

	inputs := workerJSONCorpusProduct(t, "decode", false)
	product := workerJSONProduct(t, "decode/corpus", false)
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
				oracle = append(oracle, run("node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", "internal/worker/testdata/json_decode/native.a", part, position)...)
			}
			if !bytes.Equal(expected, oracle) {
				t.Fatal("native corpus output differs from Node source")
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

	for _, mutant := range workerJSONDecodeMutants {
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

var workerJSONDecodeMutants = []struct{ name, before, after string }{
	{"duplicate-first", "object[key] = value(depth + 1);", "const parsed = value(depth + 1); if (!Object.hasOwn(object, key)) object[key] = parsed;"},
	{"wrong-descriptor-cache", "workerDecodeWalkers.get(descriptor)", "workerDecodeWalkers.get(workerDecodeWalkers)"},
	{"skip-depth-scan", "if (!withinWorkerJsonDepth(text)) return parseWorkerJson(text);", ""},
	{"depth-128-accepted", "if (depth >= 128)", "if (depth >= 129)"},
	{"missing-required", "if (field.optional) continue;", "if (field.optional || name === 'required') continue;"},
	{"reject-extra", "const result = {};", "const result = {}; if (Object.keys(value).some(name => !fields.some(field => field.name === name))) workerJsonMismatch(expected, path, value);"},
	{"reword-error", "missing field ${name}", "required field ${name} is missing"},
	// Failure-path-only mutant; direct pin: TestWorkerDecodeJSONLazyScalarFailurePaths.
	{"lazy-scalar-index-plus-one", "index === undefined ? parentPath : `${parentPath}[${index}]`", "index === undefined ? parentPath : `${parentPath}[${index + 1}]`"},
}

// Pin scalar failure locations separately from the corpus, and retain projection
// and copying while optimizing only successful scalar-array reads.
func TestWorkerDecodeJSONLazyScalarFailurePaths(t *testing.T) {
	t.Parallel()
	runtime := filepath.Join(t.TempDir(), "decoder.mjs")
	write(t, runtime, jsonDecodeRuntime+"\nexport { compileWorkerDecode };\n")
	script := `import assert from 'node:assert/strict';
import {pathToFileURL} from 'node:url';
const {decodeJson,compileWorkerDecode}=await import(pathToFileURL(process.argv[1]));
for(const [kind,valid,invalid] of [['number',1,'bad'],['string','ok',7],['boolean',true,7],['literal','yes','no']]){
 const expected=kind==='literal'?'"yes"':kind;
 const leaf={kind,expected,...(kind==='literal'?{of:3,literalUnits:[121,101,115]}:{})};
 const schema={root:0,nodes:[{kind:'object',expected:'Envelope',fields:[{name:'values',node:1},{name:'label',node:3,optional:true}]},{kind:'array',expected:'readonly '+expected+'[]',children:[2]},leaf,{kind:'string',expected:'string'}]};
 for(const index of [0,1]){
  const values=[valid,valid];values[index]=invalid;
  const found=invalid===null?'null':typeof invalid;
  assert.deepEqual(decodeJson(JSON.stringify({values}),schema),{kind:'Error',message:'at $.values['+index+']: expected '+expected+', found '+found});
 }
 const parsed={label:'first',values:[valid,valid],extra:'hidden'};
 const projected=compileWorkerDecode(schema)(parsed,'$');
 assert.notStrictEqual(projected,parsed);
 assert.notStrictEqual(projected.values,parsed.values);
 assert.deepEqual(projected.values,parsed.values);
 assert.deepEqual(Object.keys(projected),['values','label']);
 assert(!Object.hasOwn(projected,'extra'));
 assert.deepEqual(decodeJson('{}',schema),{kind:'Error',message:'at $: missing field values'});
}`
	command := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", script, runtime)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("lazy scalar paths, projection and copy: %v\n%s", err, output)
	}
}

func TestWorkerDecodeJSONNestedArrayScalarFailurePaths(t *testing.T) {
	t.Parallel()
	schema := `{"root":0,"nodes":[{"kind":"array","expected":"number[][][]","children":[1]},{"kind":"array","expected":"number[][]","children":[2]},{"kind":"array","expected":"number[]","children":[3]},{"kind":"number","expected":"number"}]}`
	checkWorkerDecodeJSONNestedScalarPath(t, schema, `[[[0]],[[1,2,3]]]`, `[[[0]],[[1,2,"bad"]]]`, "at $[1][0][2]: expected number, found string")
}

func TestWorkerDecodeJSONRecordNestedArrayScalarFailurePaths(t *testing.T) {
	t.Parallel()
	schema := `{"root":0,"nodes":[{"kind":"object","expected":"Rows","fields":[{"name":"rows","node":1}]},{"kind":"array","expected":"number[][]","children":[2]},{"kind":"array","expected":"number[]","children":[3]},{"kind":"number","expected":"number"}]}`
	checkWorkerDecodeJSONNestedScalarPath(t, schema, `{"rows":[[0],[1],[2,3]]}`, `{"rows":[[0],[1],[2,false]]}`, "at $.rows[2][1]: expected number, found boolean")
}

func checkWorkerDecodeJSONNestedScalarPath(t *testing.T, schema, valid, invalid, message string) {
	t.Helper()
	runtime := filepath.Join(t.TempDir(), "decoder.mjs")
	write(t, runtime, jsonDecodeRuntime+"\nexport { compileWorkerDecode };\n")
	script := `import assert from 'node:assert/strict';
import {pathToFileURL} from 'node:url';
const {decodeJson,compileWorkerDecode}=await import(pathToFileURL(process.argv[1]));
const schema=JSON.parse(process.argv[2]),valid=JSON.parse(process.argv[3]);
assert.deepEqual(decodeJson(process.argv[4],schema),{kind:'Error',message:process.argv[5]});
const decoded=decodeJson(process.argv[3],schema);
assert.deepEqual(decoded,{kind:'Ok',value:valid});
function freshCopy(actual,source){
 if(source!==null&&typeof source==='object'){
  assert.notStrictEqual(actual,source);
  for(const key of Object.keys(source))freshCopy(actual[key],source[key]);
 }
}
const projected=compileWorkerDecode(schema)(valid,'$');
assert.deepEqual(projected,valid);
freshCopy(projected,valid);
const second=decodeJson(process.argv[3],schema);
assert.deepEqual(second,decoded);
freshCopy(second.value,decoded.value);`
	command := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", script, runtime, schema, valid, invalid, message)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("nested scalar error path and fresh copy: %v\n%s", err, output)
	}
}
