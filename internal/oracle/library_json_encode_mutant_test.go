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
	const field = "adamic_encode_field *field = &type->fields[i];\n\t\t\tadamic_value child = {0};"
	cases := []struct{ name, source, from, to string }{
		{"runtime_order", `interface Pair {readonly first:string;readonly second:number} const v:Pair={second:2,first:'one'};console.log(encodeJson<Pair>(v));`, field, `adamic_encode_field *field=&type->fields[i];
if (!tuple) { for (size_t f=0;f<type->field_count;f++) { if (strcmp(type->fields[f].name,object->shape->names[i])==0) {field=&type->fields[f];break;} } }
adamic_value child={0};`},
		{"hidden_field", `interface Pair {readonly first:string;readonly second:number} const full={first:'one',second:2,hidden:'secret'.repeat(2)};const v:Pair=full;console.log(encodeJson<Pair>(v));`, field, `adamic_encode_field extra={.name="hidden",.node=type->fields[0].node,.optional=false};
adamic_encode_field *field=i<type->field_count?&type->fields[i]:&extra;adamic_value child={0};`},
		{"negative_zero", `console.log(encodeJson<number>(-0));`, "char bytes[ADAMIC_NUMBER_FORMAT_MAX];\n\t\tsize_t length = adamic_number_format(value.number, bytes);", `if (value.number==0.0 && signbit(value.number)) { ascii(builder,"-0");return; }
char bytes[ADAMIC_NUMBER_FORMAT_MAX];size_t length=adamic_number_format(value.number,bytes);`},
		{"nan", `console.log(encodeJson<number>(NaN));`, "if (!isfinite(value.number)) {\n\t\t\tascii(builder, \"null\");\n\t\t\treturn;\n\t\t}", `if (isnan(value.number)) { ascii(builder,"NaN");return; }
if (!isfinite(value.number)) { ascii(builder,"null");return; }`},
		{"raw_surrogate", `console.log(encodeJson<string>('x'.repeat(3)+'\ud800'));`, `width == 3 && code == 0xed && here[1] >= 0xa0`, `width == 3 && code == 0xed && here[1] >= 0xa0 && false`},
		{"split_byte_run", `console.log(encodeJson<string>('ordinary'.repeat(3)));`, `append(w, bytes + run, length - run, units);`, `append(w, bytes + run, length - run > 0 ? length - run - 1 : 0, units);`},
		{"cache_invalidation", `interface Pair {readonly first:string;readonly second:string} const values:readonly Pair[]=[{first:"a",second:"b"},{second:"d",first:"c"},{second:"f",first:"e"},{first:"g",second:"h"}];console.log(encodeJson<readonly Pair[]>(values));`, `field->cache.shape != object->shape`, `field->cache.shape == NULL`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			mutated := strings.Replace(string(original), test.from, test.to, 1)
			if mutated == string(original) {
				t.Fatal("mutation target missing")
			}
			if test.name == "hidden_field" {
				mutated = strings.Replace(mutated, `for (size_t i = 0; i < type->field_count; i++)`, `for (size_t i = 0; i < object->shape->count; i++)`, 1)
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
			requireJSONMutantCaught(t, onNode(t, path), got)
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

func TestJSONEncodeNativeStringMetadata(t *testing.T) {
	runtime, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/json_encode.c"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "metadata.a")
	if err := os.WriteFile(path, []byte("import {encodeJson} from 'adamic'; console.log(encodeJson<string>('x🌍é\\ud800')); console.log(encodeJson<string>('abc'));"), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	const wrapper = `static adamic_string *checked_encode(adamic_value value, const adamic_encode_schema *schema) {
 adamic_string *result=metadata_target(value,schema);
 if (!((result->length==15 && result->units==13) || (result->length==5 && result->units==6))) {
  adamic_panic("encode metadata differs",sizeof "encode metadata differs"-1);
 }
 return result;
}
`
	for _, mutant := range []bool{false, true} {
		source := string(runtime)
		if mutant {
			source = strings.Replace(source, "result->units = builder.units + 1;", "result->units = 0;", 1)
			if source == string(runtime) {
				t.Fatal("metadata mutation target missing")
			}
		}
		code := strings.ReplaceAll(source, "adamic_json_encode(", "metadata_target(") + "\n" + wrapper + strings.ReplaceAll(native.C(program), "adamic_json_encode(", "checked_encode(")
		binary := buildDecodeMutant(t, code)
		// The mutant panics with its result still allocated, by design. This test is about the
		// result's metadata, not leaks, so the mutant runs uncounted: on macOS runDecodeMutant's
		// counted build would report that allocation, exit 1 and hide the panic being asserted.
		// The baseline keeps the leak check.
		if mutant {
			got := execute(t, binary)
			if got.exitCode != 70 || !strings.Contains(string(got.stderr), "encode metadata differs") {
				t.Fatalf("metadata mutant survived: %d %s", got.exitCode, got.stderr)
			}
		} else {
			got := runDecodeMutant(t, code, binary)
			if diff := disagreement(onNode(t, path), got); diff != "" {
				t.Fatalf("metadata baseline: %s %s", diff, got.stderr)
			}
		}
	}
}
