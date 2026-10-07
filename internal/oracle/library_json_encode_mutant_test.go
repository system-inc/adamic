package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// Mutate runtime writing, not source values or the Node schema. Each mutant must
// build with -Werror, exit zero and leak nothing; only the byte oracle may kill it.
func TestJSONEncodeNativeMutants(t *testing.T) {
	original, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/json_encode.c"))
	if err != nil {
		t.Fatal(err)
	}
	const field = `const adamic_decode_field *field=&type->fields[i];adamic_value child={0};`
	cases := []struct{ name, source, from, to string }{
		{"runtime_order", `interface Pair {readonly first:string;readonly second:number} const v:Pair={second:2,first:'one'};console.log(encodeJson<Pair>(v));`, field, `const adamic_decode_field *field=&type->fields[i];
if (!tuple) { for (size_t f=0;f<type->field_count;f++) { if (strcmp(type->fields[f].name,object->shape->names[i])==0) {field=&type->fields[f];break;} } }
adamic_value child={0};`},
		{"hidden_field", `interface Pair {readonly first:string;readonly second:number} const full={first:'one',second:2,hidden:'secret'.repeat(2)};const v:Pair=full;console.log(encodeJson<Pair>(v));`, field, `adamic_decode_field extra={"hidden",type->fields[0].node,false};
const adamic_decode_field *field=i<type->field_count?&type->fields[i]:&extra;adamic_value child={0};`},
		{"negative_zero", `console.log(encodeJson<number>(-0));`, `char bytes[ADAMIC_NUMBER_FORMAT_MAX];size_t length=adamic_number_format(value.number,bytes);`, `if (value.number==0.0 && signbit(value.number)) { ascii(builder,"-0");return; }
char bytes[ADAMIC_NUMBER_FORMAT_MAX];size_t length=adamic_number_format(value.number,bytes);`},
		{"nan", `console.log(encodeJson<number>(NaN));`, `if (!isfinite(value.number)) { ascii(builder,"null");return; }`, `if (isnan(value.number)) { ascii(builder,"NaN");return; }
if (!isfinite(value.number)) { ascii(builder,"null");return; }`},
		{"raw_surrogate", `console.log(encodeJson<string>('x'.repeat(3)+'\ud800'));`, `code < 0x20 || (code >= 0xd800 && code <= 0xdfff)`, `code < 0x20`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			mutated := strings.Replace(string(original), test.from, test.to, 1)
			if mutated == string(original) {
				t.Fatal("mutation target missing")
			}
			if test.name == "hidden_field" {
				mutated = strings.Replace(mutated, `for (size_t i=0;i<type->field_count;i++)`, `for (size_t i=0;i<object->shape->count;i++)`, 1)
			}
			path := filepath.Join(t.TempDir(), "mutant.a")
			if err := os.WriteFile(path, []byte("import {encodeJson} from 'adamic';\n"+test.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			code := strings.ReplaceAll(mutated, "adamic_json_encode(", "mutant_json_encode(") + "\n" + strings.ReplaceAll(native.C(program), "adamic_json_encode(", "mutant_json_encode(")
			got := runDecodeMutant(t, code, buildDecodeMutant(t, code))
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("caught outside byte oracle: exit %d stderr %s", got.exitCode, got.stderr)
			}
			if diff := disagreement(onNode(t, path), got); diff != "stdout differs" {
				t.Fatalf("mutant survived or wrong check caught it: %s", diff)
			}
			t.Logf("%s: compiled, exit 0, leak-free; caught only by Node stdout", test.name)
		})
	}
}
func TestJSONEncodePlainStringifyMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/json_encode_differences.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := javascript.JavaScript(program)
	mutant := strings.ReplaceAll(code, "encodeJson(", "JSON.stringify(")
	if code == mutant {
		t.Fatal("mutant changed nothing")
	}
	file := filepath.Join(t.TempDir(), "mutant.mjs")
	if err := os.WriteFile(file, []byte(mutant), 0600); err != nil {
		t.Fatal(err)
	}
	got := onNode(t, file)
	if got.exitCode != 70 || !strings.Contains(string(got.stderr), "declared fields differ") {
		t.Fatalf("plain stringify mutant escaped pinned assertion: %d %s", got.exitCode, got.stderr)
	}
	if disagreement(onNode(t, path), got) == "" {
		t.Fatal("mutant survived")
	}
	t.Log("plain JSON.stringify backend: pinned declared-fields assertion failed, exit 70")
}
