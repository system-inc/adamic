package oracle

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

var jsonEncodeFixtures = []string{"json_encode_scalars.a", "json_encode_objects.a", "json_encode_unions.a", "json_encode_differences.a", "json_encode_depth.a", "json_schema_names.a"}

func init() {
	for _, name := range jsonEncodeFixtures {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}
func TestJSONEncode(t *testing.T) {
	for _, name := range jsonEncodeFixtures {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("source failed: %s", truth.stderr)
			}
			got, binary := natively(t, program)
			if diff := disagreement(truth, got); diff != "" {
				t.Fatalf("native %s: Node %q native %q stderr %s", diff, truth.stdout, got.stdout, got.stderr)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			if got := onJavaScriptBackend(t, program); disagreement(truth, got) != "" {
				t.Fatalf("JavaScript differs: %s %s", got.stdout, got.stderr)
			}
		})
	}
}
func TestJSONEncodeRefusals(t *testing.T) {
	cases := []struct{ name, source, want string }{
		{"missing", `encodeJson(1);`, "name the type: encodeJson<YourType>(value)"},
		{"closure_field", `encodeJson<{readonly fn:()=>number}>({fn:()=>1});`, "encodeJson cannot prove"},
		{"map", `encodeJson<Map<string,number>>(new Map<string,number>());`, "encodeJson cannot prove"},
		{"class", "", "encodeJson of a class instance is not yet supported; describe the data with an interface"},
		{"nested_class", `class Point {readonly x=1;} encodeJson<{readonly point:Point}>({point:new Point()});`, "encodeJson of a class instance is not yet supported; describe the data with an interface"},
		{"null", `encodeJson<null>(null);`, "containing null"},
		{"undefined", `encodeJson<undefined>(undefined);`, "encodeJson cannot prove"},
		{"optional_explicit_undefined", `encodeJson<{readonly n?:number|undefined}>({});`, "encodeJson cannot prove"},
		{"widened_union_slot", `const full={value:42};const view:{readonly value:number|string}=full;encodeJson<{readonly value:number|string}>(view);`, "complete runtime slot metadata"},
		{"widened_union_array", `const full:readonly number[]=[42];const view:readonly (number|string)[]=full;encodeJson<readonly (number|string)[]>(view);`, "complete runtime slot metadata"},
		{"ambiguous_union", `encodeJson<{readonly x:number}|{readonly y:number}>({x:1});`, "encodeJson cannot prove"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if test.name == "class" {
				data, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/json_encode_refused/class_notyet.a"))
				if err != nil {
					t.Fatal(err)
				}
				test.source = strings.TrimPrefix(string(data), "import {encodeJson} from 'adamic';\n")
			}
			path := filepath.Join(t.TempDir(), "probe.a")
			if err := os.WriteFile(path, []byte("import {encodeJson} from 'adamic';\n"+test.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), program)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("want refusal %q, got %v", test.want, err)
			}
			if strings.Contains(err.Error(), "decodeJson") {
				t.Fatalf("encoder diagnostic names decoder: %v", err)
			}
			if test.name == "class" || test.name == "nested_class" {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) {
					t.Fatalf("class refusal must be NotYet: %T", err)
				}
			}
		})
	}
}

// Node executes the fixture values and checks its own JSON.stringify, outside both
// implementations of JSON writing. The source descriptors contain types only.
func encodeWitnessSource(t *testing.T, name string, witness string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	sources, err := lower.DecodeJsonSources(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(sources[path], "encodeJson<", "witnessEncodeJson<")
	source = strings.ReplaceAll(source, "encodeJson(", "witnessEncodeJson(")
	source = "import assert from 'node:assert/strict';\nimport {encodeJson as actualEncodeJson, decodeJson as actualDecodeJson} from 'adamic';\n" + witness + "\n" + source
	encoded, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	script := `import {stripTypeScriptTypes} from 'node:module';
const source=stripTypeScriptTypes(` + string(encoded) + `);
await import('data:text/javascript;base64,'+Buffer.from(source).toString('base64'));
`
	file := filepath.Join(t.TempDir(), "witness.mjs")
	if err := os.WriteFile(file, []byte(script), 0600); err != nil {
		t.Fatal(err)
	}
	return file
}
func TestJSONEncodeStringifyWitness(t *testing.T) {
	for _, name := range jsonEncodeFixtures {
		if name == "json_encode_differences.a" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			path := encodeWitnessSource(t, name, `let witnessed=0;
process.on('beforeExit',()=>assert.ok(witnessed>0));
function witnessEncodeJson(value,schema) {
 const encoded=actualEncodeJson(value,schema);
 assert.equal(encoded,JSON.stringify(value),'independent JSON.stringify witness');
 witnessed++;
 if (typeof value !== 'number' || Number.isFinite(value)) {
  const decoded=actualDecodeJson(encoded,schema);
  assert.equal(decoded.kind,'Ok','round trip');
  assert.equal(actualEncodeJson(decoded.value,schema),encoded,'stable re-encode');
 }
 return encoded;
}`)
			got := onNode(t, path)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("Node witness failed: %s", got.stderr)
			}
		})
	}
}
func TestJSONEncodePinnedDifferences(t *testing.T) {
	path := encodeWitnessSource(t, "json_encode_differences.a", `let reorderedSeen=false,hiddenSeen=false;
function witnessEncodeJson(value,schema) {
 const encoded=actualEncodeJson(value,schema);
 if (Object.hasOwn(value,'hidden')) {
  assert.equal(encoded,'{"first":"one","second":2}');
  assert.equal(JSON.stringify(value),'{"first":"one","second":2,"hidden":"secretsecret"}');
  assert.notEqual(encoded,JSON.stringify(value));hiddenSeen=true;
 } else if (Object.keys(value).join(',')==='second,first') {
  assert.equal(encoded,'{"first":"one","second":2}');
  assert.equal(JSON.stringify(value),'{"second":2,"first":"one"}');
  assert.notEqual(encoded,JSON.stringify(value));reorderedSeen=true;
 }
 return encoded;
}
process.on('beforeExit',()=>{assert.ok(reorderedSeen);assert.ok(hiddenSeen)});`)
	got := onNode(t, path)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("pinned differences failed: %s", got.stderr)
	}
	want := "{\"first\":\"one\",\"second\":2}\n{\"first\":\"one\",\"second\":2}\n{\"2\":2,\"1\":1,\"z\":3}\n"
	if string(got.stdout) != want {
		t.Fatalf("pinned encode text: want %q got %q", want, got.stdout)
	}
}

// Pin all three nonfinite spellings against V8, independently of our schema runtime.
// They encode as null, which deliberately cannot round-trip through decodeJson<number>.
func TestJSONEncodeNonfinitePinned(t *testing.T) {
	dir := t.TempDir()
	witness := filepath.Join(dir, "witness.mjs")
	if err := os.WriteFile(witness, []byte(`for (const value of [NaN, Infinity, -Infinity]) console.log(JSON.stringify(value));`), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, witness)
	if truth.exitCode != 0 || string(truth.stdout) != "null\nnull\nnull\n" {
		t.Fatalf("Node nonfinite pin differs: %d %q %s", truth.exitCode, truth.stdout, truth.stderr)
	}
	path := filepath.Join(dir, "nonfinite.a")
	if err := os.WriteFile(path, []byte(`import {encodeJson,decodeJson,panic} from 'adamic';
for (const value of [NaN, Infinity, -Infinity]) {
 const text=encodeJson<number>(value);
 if (text !== 'null') panic('nonfinite encoding must be null');
 if (decodeJson<number>(text).kind !== 'Error') panic('numeric decode must reject null');
 console.log(text);
}`), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	if diff := disagreement(truth, source); diff != "" {
		t.Fatalf("source nonfinite pin: %s %s", diff, source.stderr)
	}
	got, binary := natively(t, program)
	if diff := disagreement(truth, got); diff != "" {
		t.Fatalf("native nonfinite pin: %s %s", diff, got.stderr)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if got := onJavaScriptBackend(t, program); disagreement(truth, got) != "" {
		t.Fatalf("JavaScript nonfinite pin differs: %q %s", got.stdout, got.stderr)
	}
}
