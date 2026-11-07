package oracle

import (
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// buildDecodeMutant builds a mutated program, sanitized on Linux, where LeakSanitizer is part of
// the run, and plain on macOS, where the sanitizers have no leak check.
func buildDecodeMutant(test *testing.T, code string) string {
	test.Helper()
	binary := filepath.Join(test.TempDir(), "mutant")
	if buildError := native.Build(code, binary, native.Options{Sanitize: runtime.GOOS == "linux"}); buildError != nil {
		test.Fatal(buildError)
	}
	return binary
}

// runDecodeMutant runs a mutant that must leak nothing, and makes a leak its failure: on Linux
// LeakSanitizer's, on macOS the counted build's report, so a leak is caught on either.
func runDecodeMutant(test *testing.T, code, binary string) run {
	test.Helper()
	switch runtime.GOOS {
	case "linux":
		return executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	case "darwin":
		got := execute(test, binary)
		if report := leakcheck.Report(test, code, binary); report != "" {
			got.exitCode, got.stderr = 1, []byte(report)
		}
		return got
	}
	test.Fatalf("no leak check for %s", runtime.GOOS)
	return run{}
}

// These changes touch the implementation's checks, not the fixture or its oracle descriptor.
// Each validation mutant must compile, exit zero and leak nothing before Node can kill it.
type jsonDecodeValidationCase struct{ name, source, from, to string }

func jsonDecodeValidationCases() []jsonDecodeValidationCase {
	const mismatch = `if (!match) { mismatch(error,path,type->expected,value->kind); return false; }`
	return []jsonDecodeValidationCase{
		{
			"number_scalar",
			`console.log(decodeJson<number>('"wrong"').kind);`,
			mismatch,
			`if (!match && strcmp(kind,"number")==0) { out->number=0; return true; } ` + mismatch,
		},
		{
			"string_scalar",
			`console.log(decodeJson<string>('42').kind);`,
			mismatch,
			`if (!match && strcmp(kind,"string")==0) { out->reference=text_of(""); return true; } ` + mismatch,
		},
		{
			"boolean_scalar",
			`console.log(decodeJson<boolean>('42').kind);`,
			mismatch,
			`if (!match && strcmp(kind,"boolean")==0) { out->boolean=false; return true; } ` + mismatch,
		},
		{
			"null_rejected",
			`console.log(decodeJson<number>('null').kind);`,
			mismatch,
			`if (!match && value->kind==json_null && strcmp(kind,"number")==0) { out->number=0; return true; } ` +
				mismatch,
		},
		{
			"literal_string",
			`console.log(decodeJson<'Morning'|'Evening'>('"Noon"').kind);`,
			`return value->kind==json_string && adamic_string_equal(value->string,type->literal);`,
			`return value->kind==json_string;`,
		},
		{
			"literal_number",
			`console.log(decodeJson<1|2>('3').kind);`,
			`value->number==type->number`,
			`(value->number==type->number || true)`,
		},
		{
			"literal_boolean",
			`console.log(decodeJson<true>('false').kind);`,
			`value->boolean==type->boolean`,
			`(value->boolean==type->boolean || true)`,
		},
		{
			"array_kind",
			`console.log(decodeJson<readonly number[]>('{}').kind);`,
			mismatch,
			`if (!match && strcmp(kind,"array")==0) { out->reference=adamic_array_make(0,adamic_ir_Number); return true; } ` +
				mismatch,
		},
		{
			"object_kind",
			`console.log(decodeJson<{readonly x:number}>('[]').kind);`,
			mismatch,
			`if (!match && strcmp(kind,"object")==0) { out->reference=make_object(0); return true; } ` + mismatch,
		},
		{
			"tuple_kind",
			`console.log(decodeJson<readonly [number]>('{}').kind);`,
			mismatch,
			`if (!match && strcmp(kind,"tuple")==0) { out->reference=make_object(0); return true; } ` + mismatch,
		},
		{
			"tuple_length",
			`console.log(decodeJson<readonly [number]>('[]').kind);`,
			`if (tuple && value->count!=type->field_count) { mismatch(error,path,type->expected,value->kind); return false; }`,
			`if (tuple && value->count!=type->field_count) { out->reference=make_object(0); return true; }`,
		},
		{
			"required_field",
			`console.log(decodeJson<{readonly name:string}>('{}').kind);`,
			`if (f->optional) { continue; }`,
			`if (f->optional || strcmp(f->name,"name")==0) { continue; }`,
		},
		{
			"missing_discriminant",
			`type S={readonly kind:'A';readonly n:number}|{readonly kind:'B';readonly n:number}; console.log(decodeJson<S>('{}').kind);`,
			`if (tag==NULL) { missing(error,path,type->discriminant); return false; }`,
			`if (tag==NULL) { out->reference=make_object(0); return true; }`,
		},
		{
			"discriminant",
			`type S={readonly kind:'A';readonly n:number}|{readonly kind:'B';readonly n:number}; console.log(decodeJson<S>('{"kind":"Bad","n":1}').kind);`,
			`if (strcmp(member->fields[f].name,type->discriminant)==0) { match=literal(tag,&schema->nodes[member->fields[f].node]); break; }`,
			`if (strcmp(member->fields[f].name,type->discriminant)==0) { match=literal(tag,&schema->nodes[member->fields[f].node]) || tag->kind==json_string; break; }`,
		},
	}
}

func jsonDecodeValidationKeys() []string {
	var keys []string
	for _, validationCase := range jsonDecodeValidationCases() {
		keys = append(keys, validationCase.name)
	}
	return keys
}

func runJSONDecodeValidationMutantsShard(test *testing.T, shard int) {
	original, readError := os.ReadFile(filepath.Join(repository, "internal/native/runtime/json_decode.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	cases := jsonDecodeValidationCases()
	for _, index := range familyCases(test, "TestJSONDecodeValidationMutants", shard, jsonDecodeValidationKeys()) {
		validationCase := cases[index]
		test.Run(validationCase.name, func(test *testing.T) {
			mutated := strings.Replace(string(original), validationCase.from, validationCase.to, 1)
			if _, probing := familyProbe("TestJSONDecodeValidationMutants"); probing {
				mutated = string(original)
			}
			if mutated == string(original) {
				test.Fatal("mutant changed nothing")
			}
			if validationCase.name == "discriminant" {
				mutated = strings.Replace(
					mutated,
					`return value->kind==json_string && adamic_string_equal(value->string,type->literal);`,
					`return value->kind==json_string;`,
					1,
				)
			}
			path := filepath.Join(test.TempDir(), "mutant.a")
			if writeError := os.WriteFile(
				path,
				[]byte("import { decodeJson } from 'adamic';\n"+validationCase.source),
				0600,
			); writeError != nil {
				test.Fatal(writeError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			code := inlineDecodeMutant(mutated, native.C(program))
			got := runDecodeMutant(test, code, buildDecodeMutant(test, code))
			if got.exitCode != 0 || len(got.stderr) != 0 {
				test.Fatalf("caught outside Node comparison: exit %d stderr %s", got.exitCode, got.stderr)
			}
			requireJSONMutantCaught(test, onNode(test, path), got)
			test.Logf("%s: caught only by Node stdout comparison", validationCase.name)
		})
	}
}
func inlineDecodeMutant(runtime, code string) string {
	runtime = strings.ReplaceAll(runtime, "adamic_json_decode(", "json_mutant_decode(")
	code = strings.ReplaceAll(code, "adamic_json_decode(", "json_mutant_decode(")
	return runtime + "\n" + code
}
func TestJSONDecodeParserMutants(test *testing.T) {
	test.Parallel()
	original, readError := os.ReadFile(filepath.Join(repository, "internal/native/runtime/json_decode.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	cases := []struct{ name, fixture, from, to string }{
		{
			"rounding",
			"json_decode_scalars.a",
			"n->number=strtod(bytes,NULL);",
			"n->number=strtod(bytes,NULL); if (n->number==0.1) { n->number=nextafter(n->number,INFINITY); }",
		},
		{
			"negative_zero",
			"json_decode_scalars.a",
			"n->number=strtod(bytes,NULL);",
			"n->number=strtod(bytes,NULL); if (n->number==0.0) { n->number=0.0; }",
		},
		{
			"lone_surrogate",
			"json_decode_scalars.a",
			"units[count++]=(double)c;",
			"units[count++]=(double)(c==0xd800?0xfffd:c);",
		},
		{"depth", "json_decode_depth.a", "if (depth>=ADAMIC_JSON_DEPTH)", "if (depth>ADAMIC_JSON_DEPTH)"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata", testCase.fixture),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			mutated := strings.Replace(string(original), testCase.from, testCase.to, 1)
			if mutated == string(original) {
				test.Fatal("mutant changed nothing")
			}
			code := inlineDecodeMutant(mutated, native.C(program))
			got := runDecodeMutant(test, code, buildDecodeMutant(test, code))
			if got.exitCode != 0 || len(got.stderr) != 0 {
				test.Fatalf("caught outside comparison: %d %s", got.exitCode, got.stderr)
			}
			requireJSONMutantCaught(test, onNode(test, path), got)
			test.Logf("%s: caught only by Node stdout comparison", testCase.name)
		})
	}
}
func TestJSONDecodeErrorLeakMutant(test *testing.T) {
	test.Parallel()
	original, readError := os.ReadFile(filepath.Join(repository, "internal/native/runtime/json_decode.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	from := `if (!ok) { adamic_release(a); return false; }`
	mutated := strings.Replace(string(original), from, `if (!ok) { return false; }`, 1)
	if mutated == string(original) {
		test.Fatal("mutant changed nothing")
	}
	path := filepath.Join(test.TempDir(), "error.a")
	if writeError := os.WriteFile(path, []byte(
		`import {decodeJson} from 'adamic'; const r=decodeJson<readonly string[]>('["half built",42]'); console.log(r.kind);`,
	), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	code := inlineDecodeMutant(mutated, native.C(program))
	binary := buildDecodeMutant(test, code)
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	if diff := disagreement(onNode(test, path), executeWith(test, environment, binary)); diff != "" {
		test.Fatalf("caught outside leak check: %s", diff)
	}
	if report := leakcheck.Report(test, code, binary); report == "" {
		test.Fatal("leak mutant survived")
	}
	test.Log("half-array error cleanup: caught only by the leak check")
}

func TestJSONDecodeFieldPresenceMutants(test *testing.T) {
	test.Parallel()
	original, readError := os.ReadFile(filepath.Join(repository, "internal/native/runtime/json_decode.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	cases := []struct {
		name, source string
		mutate       func(string) string
	}{
		{
			"extra_field_kept",
			`const r=decodeJson<{readonly first:string}>('{"first":"one","extra":1}'); if(r.kind==='Ok') console.log(Object.keys(r.value).join(',')); else console.log(r.message);`,
			func(
				text string,
			) string {
				text = strings.Replace(
					text,
					`adamic_object *o=make_object(count);`,
					`adamic_object *o=make_object(count+1);`,
					1,
				)
				return strings.Replace(
					text,
					`o->shape=canonical_shape(o->shape);`,
					`((const char **)o->shape->names)[written]="extra"; ((bool *)o->shape->references)[written]=false; o->slots[written].number=1; o->shape=canonical_shape(o->shape);`,
					1,
				)
			},
		},
		{
			"optional_field_filled",
			`const r=decodeJson<{readonly first:string;readonly note?:string}>('{"first":"one"}'); if(r.kind==='Ok') console.log(Object.keys(r.value).join(',')); else console.log(r.message);`,
			func(
				text string,
			) string {
				text = strings.Replace(
					text,
					`if (tuple || field(value,type->fields[i].name)!=NULL) { count++; }`,
					`if (tuple || field(value,type->fields[i].name)!=NULL || type->fields[i].optional) { count++; }`,
					1,
				)
				return strings.Replace(
					text,
					`if (f->optional) { continue; }`,
					`if (f->optional) { ((const char **)o->shape->names)[written]=f->name; ((bool *)o->shape->references)[written]=true; o->slots[written++].reference=text_of("filled"); continue; }`,
					1,
				)
			},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			path := filepath.Join(test.TempDir(), "mutant.a")
			if writeError := os.WriteFile(
				path,
				[]byte("import {decodeJson} from 'adamic';\n"+testCase.source),
				0600,
			); writeError != nil {
				test.Fatal(writeError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			mutated := testCase.mutate(string(original))
			if mutated == string(original) {
				test.Fatal("mutant changed nothing")
			}
			code := inlineDecodeMutant(mutated, native.C(program))
			got := runDecodeMutant(test, code, buildDecodeMutant(test, code))
			if got.exitCode != 0 || len(got.stderr) != 0 {
				test.Fatalf("caught outside comparison: %d %s", got.exitCode, got.stderr)
			}
			requireJSONMutantCaught(test, onNode(test, path), got)
			test.Logf("%s: caught only by Node stdout comparison", testCase.name)
		})
	}
}

// Restore the old omission in the emitted descriptor, leaving valid JavaScript and the source
// oracle untouched. Each empty-shape fixture must catch its validator's TypeError at runtime.
func TestJSONDecodeEmptyFieldsMutant(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"json_decode_empty_tuple.a", "json_decode_empty_object.a"} {
		test.Run(name, func(test *testing.T) {
			path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			if jsonDecodeEnumerationRefused(name) {
				_, loweredError := lowered(test, path)
				assertJSONDecodeEnumerationRefusal(test, loweredError)
				// Retain the empty-object descriptor witness without Object.keys, which main refuses.
				path = filepath.Join(test.TempDir(), "empty-object.a")
				source := "import { decodeJson } from 'adamic'; console.log(decodeJson<{}>('{}').kind);"
				if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
					test.Fatal(writeError)
				}
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			truth := onNode(test, path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				test.Fatalf("reference failed: %s", truth.stderr)
			}
			code := javascript.JavaScript(program)
			mutant := strings.ReplaceAll(code, `,"fields":[]`, "")
			if mutant == code {
				test.Fatal("mutant changed nothing")
			}
			mutantPath := filepath.Join(test.TempDir(), "omitted-fields.mjs")
			if writeError := os.WriteFile(mutantPath, []byte(mutant), 0600); writeError != nil {
				test.Fatal(writeError)
			}
			got := onNode(test, mutantPath)
			if got.exitCode != 70 ||
				!strings.Contains(string(got.stderr), "TypeError") ||
				disagreement(truth, got) == "" {
				test.Fatalf(
					"omission survived or was caught outside runtime comparison: exit %d stderr %s",
					got.exitCode,
					got.stderr,
				)
			}
			test.Log("empty fields omitted: caught by fixture's Node comparison (TypeError, exit 70)")
		})
	}
}
