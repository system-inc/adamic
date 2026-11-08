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

var jsonEncodeFixtures = []string{
	"json_encode_scalars.a",
	"json_encode_objects.a",
	"json_encode_unions.a",
	"json_encode_differences.a",
	"json_encode_depth.a",
	"json_schema_names.a",
	"json_encode_own_data.a",
	"json_encode_method.a",
	"json_encode_last_use.a",
}

func init() {
	for _, name := range jsonEncodeFixtures {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}
func jsonEncodeKeys() []string { return jsonEncodeFixtures }

func runJSONEncodeShard(test *testing.T, shard int) {
	for _, index := range familyCases(test, "TestJSONEncode", shard, jsonEncodeKeys()) {
		name := jsonEncodeFixtures[index]
		test.Run(name, func(test *testing.T) {
			path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			truth := onNode(test, path)
			if _, probing := familyProbe("TestJSONEncode"); probing {
				truth.stdout = append(truth.stdout, '!')
			}
			if truth.exitCode != 0 {
				test.Fatalf("source failed: %s", truth.stderr)
			}
			got, binary := natively(test, program)
			if diff := disagreement(truth, got); diff != "" {
				test.Fatalf("native %s: Node %q native %q stderr %s", diff, truth.stdout, got.stdout, got.stderr)
			}
			if report := leaks(test, program, binary); report != "" {
				test.Fatal(report)
			}
			if got := onJavaScriptBackend(test, program); disagreement(truth, got) != "" {
				test.Fatalf("JavaScript differs: %s %s", got.stdout, got.stderr)
			}
		})
	}
}
func TestJsonEncodeRefusals(test *testing.T) {
	test.Parallel()
	cases := []struct{ name, source, want string }{
		{"missing", `encodeJson(1);`, "name the type: encodeJson<YourType>(value)"},
		{"closure_field", `encodeJson<{readonly fn:()=>number}>({fn:()=>1});`, "encodeJson cannot prove"},
		{"map", `encodeJson<Map<string,number>>(new Map<string,number>());`, "encodeJson cannot prove"},
		{"class", "", "encodeJson of a class instance is not yet supported; describe the data with an interface"},
		{
			"nested_class",
			`class Point {readonly x=1;} encodeJson<{readonly point:Point}>({point:new Point()});`,
			"encodeJson of a class instance is not yet supported; describe the data with an interface",
		},
		{"null", `encodeJson<null>(null);`, "containing null"},
		{"undefined", `encodeJson<undefined>(undefined);`, "encodeJson cannot prove"},
		{"optional_explicit_undefined", `encodeJson<{readonly n?:number|undefined}>({});`, "encodeJson cannot prove"},
		{
			"widened_union_slot",
			`const full={value:42};const view:{readonly value:number|string}=full;encodeJson<{readonly value:number|string}>(view);`,
			"complete runtime slot metadata",
		},
		// A widened array whose elements read as the view holds them, strings read as a union, reaches
		// encodeJson's own refusal. A number[] widened so stops first, where the view is made, since its
		// numbers would be read as boxes (element_fit.go, #3vsp1ds).
		{
			"widened_union_array",
			`const full:readonly string[]=['a'];const view:readonly (number|string)[]=full;encodeJson<readonly (number|string)[]>(view);`,
			"complete runtime slot metadata",
		},
		{
			"widened_number_array",
			`const full:readonly number[]=[42];const view:readonly (number|string)[]=full;encodeJson<readonly (number|string)[]>(view);`,
			"held as a boxed union",
		},
		{"ambiguous_union", `encodeJson<{readonly x:number}|{readonly y:number}>({x:1});`, "encodeJson cannot prove"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if testCase.name == "class" {
				data, readError := os.ReadFile(
					filepath.Join(repository, "internal/oracle/testdata/json_encode_refused/class_notyet.a"),
				)
				if readError != nil {
					test.Fatal(readError)
				}
				testCase.source = strings.TrimPrefix(string(data), "import {encodeJson} from 'adamic';\n")
			}
			path := filepath.Join(test.TempDir(), "probe.a")
			if writeError := os.WriteFile(
				path,
				[]byte("import {encodeJson} from 'adamic';\n"+testCase.source),
				0600,
			); writeError != nil {
				test.Fatal(writeError)
			}
			program, loadError := load.Load([]string{path})
			if loadError != nil {
				test.Fatal(loadError)
			}
			_, lowerError := lower.Lower(context.Background(), program)
			if lowerError == nil || !strings.Contains(lowerError.Error(), testCase.want) {
				test.Fatalf("want refusal %q, got %v", testCase.want, lowerError)
			}
			if strings.Contains(lowerError.Error(), "decodeJson") {
				test.Fatalf("encoder diagnostic names decoder: %v", lowerError)
			}
			if testCase.name == "class" || testCase.name == "nested_class" {
				var notYet *lower.NotYet
				if !errors.As(lowerError, &notYet) {
					test.Fatalf("class refusal must be NotYet: %T", lowerError)
				}
			}
		})
	}
}

// Node executes the fixture values and checks its own JSON.stringify, outside both
// implementations of JSON writing. The source descriptors contain types only.
func encodeWitnessSource(test *testing.T, name string, witness string, selectSource ...func(string) string) string {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	sources, decodeJsonSourcesError := lower.DecodeJsonSources(context.Background(), program)
	if decodeJsonSourcesError != nil {
		test.Fatal(decodeJsonSourcesError)
	}
	selected := sources[path]
	for _, selectOne := range selectSource {
		selected = selectOne(selected)
	}
	source := strings.ReplaceAll(selected, "encodeJson<", "witnessEncodeJson<")
	source = strings.ReplaceAll(source, "encodeJson(", "witnessEncodeJson(")
	source = "import assert from 'node:assert/strict';\nimport {encodeJson as actualEncodeJson, decodeJson as actualDecodeJson} from 'adamic';\n" +
		witness +
		"\n" +
		source
	encoded, encodeError := json.Marshal(source)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	script := `import {stripTypeScriptTypes} from 'node:module';
const source=stripTypeScriptTypes(` + string(encoded) + `);
await import('data:text/javascript;base64,'+Buffer.from(source).toString('base64'));
`
	file := filepath.Join(test.TempDir(), "witness.mjs")
	if writeError := os.WriteFile(file, []byte(script), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	return file
}
func TestJsonEncodeStringifyWitness(test *testing.T) {
	test.Parallel()
	for _, name := range jsonEncodeFixtures {
		if name == "json_encode_differences.a" {
			continue
		}
		test.Run(name, func(test *testing.T) {
			path := encodeWitnessSource(test, name, `let witnessed=0;
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
			got := onNode(test, path)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				test.Fatalf("Node witness failed: %s", got.stderr)
			}
		})
	}
}
func jsonPinnedDifferenceKeys() []string {
	return []string{"runtime_order", "hidden_field", "integer_keys"}
}
func runJSONEncodePinnedDifferencesShard(test *testing.T, shard int) {
	for _, index := range familyCases(test, "TestJSONEncodePinnedDifferences", shard, jsonPinnedDifferenceKeys()) {
		key := jsonPinnedDifferenceKeys()[index]
		test.Run(key, func(test *testing.T) { jsonEncodePinnedDifferences(test, key) })
	}
}
func jsonEncodePinnedDifferences(test *testing.T, key string) {
	path := encodeWitnessSource(test, "json_encode_differences.a", `let reorderedSeen=false,hiddenSeen=false;
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
process.on('beforeExit',()=>{const key='`+key+`';if(key==='runtime_order') assert.ok(reorderedSeen);if(key==='hidden_field') assert.ok(hiddenSeen)});`, func(source string) string {
		start := strings.Index(source, "for (const value of [reordered, hidden])")
		numeric := strings.Index(source, "// Numeric keys")
		if start < 0 || numeric < start {
			test.Fatal("pinned fixture split targets changed")
		}
		if key == "integer_keys" {
			return source[:start] + source[numeric:]
		}
		source = source[:numeric]
		value := "reordered"
		if key == "hidden_field" {
			value = "hidden"
		}
		return strings.Replace(source, "[reordered, hidden]", "["+value+"]", 1)
	})
	got := onNode(test, path)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		test.Fatalf("pinned differences failed: %s", got.stderr)
	}
	want := "{\"first\":\"one\",\"second\":2}\n"
	if key == "integer_keys" {
		want = "{\"2\":2,\"1\":1,\"z\":3}\n"
	}
	if _, probing := familyProbe("TestJSONEncodePinnedDifferences"); probing {
		want += "!"
	}
	if string(got.stdout) != want {
		test.Fatalf("pinned encode text: want %q got %q", want, got.stdout)
	}

}

// Pin all three nonfinite spellings against V8, independently of our schema runtime.
// They encode as null, which deliberately cannot round-trip through decodeJson<number>.
func jsonNonfiniteKeys() []string { return []string{"NaN", "Infinity", "-Infinity"} }
func runJSONEncodeNonfinitePinnedShard(test *testing.T, shard int) {
	for _, index := range familyCases(test, "TestJSONEncodeNonfinitePinned", shard, jsonNonfiniteKeys()) {
		key := jsonNonfiniteKeys()[index]
		test.Run(key, func(test *testing.T) { jsonEncodeNonfinitePinned(test, key) })
	}
}
func jsonEncodeNonfinitePinned(test *testing.T, key string) {
	dir := test.TempDir()
	witness := filepath.Join(dir, "witness.mjs")
	if writeError := os.WriteFile(witness, []byte(`console.log(JSON.stringify(`+key+`));`), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	truth := onNode(test, witness)
	if truth.exitCode != 0 || string(truth.stdout) != "null\n" {
		test.Fatalf("Node nonfinite pin differs: %d %q %s", truth.exitCode, truth.stdout, truth.stderr)
	}
	path := filepath.Join(dir, "nonfinite.a")
	if writeError := os.WriteFile(path, []byte(`import {encodeJson,decodeJson,panic} from 'adamic';
for (const value of [`+key+`]) {
 const text=encodeJson<number>(value);
 if (text !== 'null') panic('nonfinite encoding must be null');
 if (decodeJson<number>(text).kind !== 'Error') panic('numeric decode must reject null');
 console.log(text);
}`), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	source := onNode(test, path)
	if _, probing := familyProbe("TestJSONEncodeNonfinitePinned"); probing {
		source.stdout = append(source.stdout, '!')
	}
	if diff := disagreement(truth, source); diff != "" {
		test.Fatalf("source nonfinite pin: %s %s", diff, source.stderr)
	}
	got, binary := natively(test, program)
	if diff := disagreement(truth, got); diff != "" {
		test.Fatalf("native nonfinite pin: %s %s", diff, got.stderr)
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
	if got := onJavaScriptBackend(test, program); disagreement(truth, got) != "" {
		test.Fatalf("JavaScript nonfinite pin differs: %q %s", got.stdout, got.stderr)
	}
}

// This is red if encodeJson drops the shared descriptor guard: the accessor is
// hidden by Data, and native serialization would read a nonexistent data slot.
func TestJsonEncodeAccessorRefusal(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/json_encode_refused/getter.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	source, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	_, lowerError := lower.Lower(context.Background(), source)
	var refused *lower.Refused
	if !errors.As(lowerError, &refused) || !strings.Contains(
		lowerError.Error(),
		"Object enumeration with getter or accessor descriptors; use own data properties for this first reflection implementation",
	) {
		test.Fatalf("want shared accessor descriptor refusal, got %v", lowerError)
	}
}
