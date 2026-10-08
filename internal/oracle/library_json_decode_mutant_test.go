package oracle

import (
	"fmt"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// buildDecodeMutant builds a mutated program, sanitized on Linux, where LeakSanitizer is part of
// the run, and plain on macOS, where the sanitizers have no leak check.
func buildDecodeMutant(t *testing.T, code string) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(code, binary, native.Options{Sanitize: runtime.GOOS == "linux"}); err != nil {
		t.Fatal(err)
	}
	return binary
}

// runDecodeMutant runs a mutant that must leak nothing, and makes a leak its failure: on Linux
// LeakSanitizer's, on macOS the counted build's report, so a leak is caught on either.
func runDecodeMutant(t *testing.T, code, binary string) run {
	t.Helper()
	switch runtime.GOOS {
	case "linux":
		return executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	case "darwin":
		got := execute(t, binary)
		if report := decodeMutantLeaks(t, code, binary); report != "" {
			got.exitCode, got.stderr = 1, []byte(report)
		}
		return got
	}
	t.Fatalf("no leak check for %s", runtime.GOOS)
	return run{}
}

// decodeMutantLeaks is the leak report for a finished mutant, or "" when it let go of everything.
// On macOS that's a counted build's: leaks --atExit can't see a value the size-class allocator
// holds, which is every value in a build without the sanitizers, so a leaked half-built array
// passes it. A finished program's allocations are its frees and its regions (runtime/count.h).
func decodeMutantLeaks(t *testing.T, code, binary string) string {
	t.Helper()
	switch runtime.GOOS {
	case "linux":
		return leakSanitizer(t, binary)
	case "darwin":
		counted := filepath.Join(t.TempDir(), "counted")
		if err := native.Build(code, counted, native.Options{Count: true}); err != nil {
			t.Fatal(err)
		}
		result := execute(t, counted)
		var allocations, frees, retains, releases, peak, regions int
		line := string(result.stderr)
		if at := strings.LastIndex(line, "adamic: counts:"); at >= 0 {
			line = line[at:]
		}
		if _, err := fmt.Sscanf(line, "adamic: counts: allocations %d frees %d retains %d releases %d peak %d regions %d", &allocations, &frees, &retains, &releases, &peak, &regions); err != nil {
			t.Fatalf("no counts from the counted build: %v\n%s", err, result.stderr)
		}
		if allocations != frees+regions {
			return fmt.Sprintf("allocations %d, frees %d, regions %d", allocations, frees, regions)
		}
		return ""
	}
	t.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
}

// These changes touch the implementation's checks, not the fixture or its oracle descriptor.
// Each validation mutant must compile, exit zero and leak nothing before Node can kill it.
func TestJSONDecodeValidationMutants(t *testing.T) {
	original, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/json_decode.c"))
	if err != nil {
		t.Fatal(err)
	}
	const mismatch = `if (!match) { mismatch(error,path,type->expected,value->kind); return false; }`
	cases := []struct{ name, source, from, to string }{
		{"number_scalar", `console.log(decodeJson<number>('"wrong"').kind);`, mismatch, `if (!match && strcmp(kind,"number")==0) { out->number=0; return true; } ` + mismatch},
		{"string_scalar", `console.log(decodeJson<string>('42').kind);`, mismatch, `if (!match && strcmp(kind,"string")==0) { out->reference=text_of(""); return true; } ` + mismatch},
		{"boolean_scalar", `console.log(decodeJson<boolean>('42').kind);`, mismatch, `if (!match && strcmp(kind,"boolean")==0) { out->boolean=false; return true; } ` + mismatch},
		{"null_rejected", `console.log(decodeJson<number>('null').kind);`, mismatch, `if (!match && value->kind==json_null && strcmp(kind,"number")==0) { out->number=0; return true; } ` + mismatch},
		{"literal_string", `console.log(decodeJson<'Morning'|'Evening'>('"Noon"').kind);`, `return value->kind==json_string && adamic_string_equal(value->string,type->literal);`, `return value->kind==json_string;`},
		{"literal_number", `console.log(decodeJson<1|2>('3').kind);`, `value->number==type->number`, `(value->number==type->number || true)`},
		{"literal_boolean", `console.log(decodeJson<true>('false').kind);`, `value->boolean==type->boolean`, `(value->boolean==type->boolean || true)`},
		{"array_kind", `console.log(decodeJson<readonly number[]>('{}').kind);`, mismatch, `if (!match && strcmp(kind,"array")==0) { out->reference=adamic_array_new(0,false); return true; } ` + mismatch},
		{"object_kind", `console.log(decodeJson<{readonly x:number}>('[]').kind);`, mismatch, `if (!match && strcmp(kind,"object")==0) { out->reference=make_object(0); return true; } ` + mismatch},
		{"tuple_kind", `console.log(decodeJson<readonly [number]>('{}').kind);`, mismatch, `if (!match && strcmp(kind,"tuple")==0) { out->reference=make_object(0); return true; } ` + mismatch},
		{"tuple_length", `console.log(decodeJson<readonly [number]>('[]').kind);`, `if (tuple && value->count!=type->field_count) { mismatch(error,path,type->expected,value->kind); return false; }`, `if (tuple && value->count!=type->field_count) { out->reference=make_object(0); return true; }`},
		{"required_field", `console.log(decodeJson<{readonly name:string}>('{}').kind);`, `if (f->optional) { continue; }`, `if (f->optional || strcmp(f->name,"name")==0) { continue; }`},
		{"missing_discriminant", `type S={readonly kind:'A';readonly n:number}|{readonly kind:'B';readonly n:number}; console.log(decodeJson<S>('{}').kind);`, `if (tag==NULL) { missing(error,path,type->discriminant); return false; }`, `if (tag==NULL) { out->reference=make_object(0); return true; }`},
		{"discriminant", `type S={readonly kind:'A';readonly n:number}|{readonly kind:'B';readonly n:number}; console.log(decodeJson<S>('{"kind":"Bad","n":1}').kind);`, `if (strcmp(member->fields[f].name,type->discriminant)==0) { match=literal(tag,&schema->nodes[member->fields[f].node]); break; }`, `if (strcmp(member->fields[f].name,type->discriminant)==0) { match=literal(tag,&schema->nodes[member->fields[f].node]) || tag->kind==json_string; break; }`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mutated := strings.Replace(string(original), c.from, c.to, 1)
			if mutated == string(original) {
				t.Fatal("mutant changed nothing")
			}
			if c.name == "discriminant" {
				mutated = strings.Replace(mutated, `return value->kind==json_string && adamic_string_equal(value->string,type->literal);`, `return value->kind==json_string;`, 1)
			}
			path := filepath.Join(t.TempDir(), "mutant.a")
			if err := os.WriteFile(path, []byte("import { decodeJson } from 'adamic';\n"+c.source), 0600); err != nil {
				t.Fatal(err)
			}
			p, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			code := inlineDecodeMutant(mutated, native.C(p))
			got := runDecodeMutant(t, code, buildDecodeMutant(t, code))
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("caught outside Node comparison: exit %d stderr %s", got.exitCode, got.stderr)
			}
			requireJSONMutantCaught(t, onNode(t, path), got)
			t.Logf("%s: caught only by Node stdout comparison", c.name)
		})
	}
}
func inlineDecodeMutant(runtime, code string) string {
	runtime = strings.ReplaceAll(runtime, "adamic_json_decode(", "json_mutant_decode(")
	code = strings.ReplaceAll(code, "adamic_json_decode(", "json_mutant_decode(")
	return runtime + "\n" + code
}
func TestJSONDecodeParserMutants(t *testing.T) {
	original, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/json_decode.c"))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ name, fixture, from, to string }{
		{"rounding", "json_decode_scalars.a", "n->number=strtod(bytes,NULL);", "n->number=strtod(bytes,NULL); if (n->number==0.1) { n->number=nextafter(n->number,INFINITY); }"},
		{"negative_zero", "json_decode_scalars.a", "n->number=strtod(bytes,NULL);", "n->number=strtod(bytes,NULL); if (n->number==0.0) { n->number=0.0; }"},
		{"lone_surrogate", "json_decode_scalars.a", "units[count++]=(double)c;", "units[count++]=(double)(c==0xd800?0xfffd:c);"},
		{"depth", "json_decode_depth.a", "if (depth>=ADAMIC_JSON_DEPTH)", "if (depth>ADAMIC_JSON_DEPTH)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", c.fixture))
			if err != nil {
				t.Fatal(err)
			}
			p, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			mutated := strings.Replace(string(original), c.from, c.to, 1)
			if mutated == string(original) {
				t.Fatal("mutant changed nothing")
			}
			code := inlineDecodeMutant(mutated, native.C(p))
			got := runDecodeMutant(t, code, buildDecodeMutant(t, code))
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("caught outside comparison: %d %s", got.exitCode, got.stderr)
			}
			requireJSONMutantCaught(t, onNode(t, path), got)
			t.Logf("%s: caught only by Node stdout comparison", c.name)
		})
	}
}
func TestJSONDecodeErrorLeakMutant(t *testing.T) {
	original, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/json_decode.c"))
	if err != nil {
		t.Fatal(err)
	}
	from := `if (!ok) { adamic_release(a); return false; }`
	mutated := strings.Replace(string(original), from, `if (!ok) { return false; }`, 1)
	if mutated == string(original) {
		t.Fatal("mutant changed nothing")
	}
	path := filepath.Join(t.TempDir(), "error.a")
	if err := os.WriteFile(path, []byte(`import {decodeJson} from 'adamic'; const r=decodeJson<readonly string[]>('["half built",42]'); console.log(r.kind);`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := inlineDecodeMutant(mutated, native.C(p))
	binary := buildDecodeMutant(t, code)
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	if diff := disagreement(onNode(t, path), executeWith(t, environment, binary)); diff != "" {
		t.Fatalf("caught outside leak check: %s", diff)
	}
	if report := decodeMutantLeaks(t, code, binary); report == "" {
		t.Fatal("leak mutant survived")
	}
	t.Log("half-array error cleanup: caught only by the leak check")
}

func TestJSONDecodeFieldPresenceMutants(t *testing.T) {
	original, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/json_decode.c"))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, source string
		mutate       func(string) string
	}{
		{"extra_field_kept", `const r=decodeJson<{readonly first:string}>('{"first":"one","extra":1}'); if(r.kind==='Ok') console.log(Object.keys(r.value).join(',')); else console.log(r.message);`, func(s string) string {
			s = strings.Replace(s, `adamic_object *o=make_object(count); size_t written=0;`, `adamic_object *o=make_object(count+1); size_t written=0;`, 1)
			return strings.Replace(s, `o->shape=canonical_shape(o->shape);`, `((const char **)o->shape->names)[written]="extra"; ((bool *)o->shape->references)[written]=false; o->slots[written].number=1; o->shape=canonical_shape(o->shape);`, 1)
		}},
		{"optional_field_filled", `const r=decodeJson<{readonly first:string;readonly note?:string}>('{"first":"one"}'); if(r.kind==='Ok') console.log(Object.keys(r.value).join(',')); else console.log(r.message);`, func(s string) string {
			s = strings.Replace(s, `if (tuple || field(value,type->fields[i].name)!=NULL) { count++; }`, `if (tuple || field(value,type->fields[i].name)!=NULL || type->fields[i].optional) { count++; }`, 1)
			return strings.Replace(s, `if (f->optional) { continue; }`, `if (f->optional) { ((const char **)o->shape->names)[written]=f->name; ((bool *)o->shape->references)[written]=true; o->slots[written++].reference=text_of("filled"); continue; }`, 1)
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "mutant.a")
			if err := os.WriteFile(path, []byte("import {decodeJson} from 'adamic';\n"+c.source), 0600); err != nil {
				t.Fatal(err)
			}
			p, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			mutated := c.mutate(string(original))
			if mutated == string(original) {
				t.Fatal("mutant changed nothing")
			}
			code := inlineDecodeMutant(mutated, native.C(p))
			got := runDecodeMutant(t, code, buildDecodeMutant(t, code))
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("caught outside comparison: %d %s", got.exitCode, got.stderr)
			}
			requireJSONMutantCaught(t, onNode(t, path), got)
			t.Logf("%s: caught only by Node stdout comparison", c.name)
		})
	}
}

// Restore the old omission in the emitted descriptor, leaving valid JavaScript and the source
// oracle untouched. Each empty-shape fixture must catch its validator's TypeError at runtime.
func TestJSONDecodeEmptyFieldsMutant(t *testing.T) {
	for _, name := range []string{"json_decode_empty_tuple.a", "json_decode_empty_object.a"} {
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
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("reference failed: %s", truth.stderr)
			}
			code := javascript.JavaScript(program)
			mutant := strings.ReplaceAll(code, `,"fields":[]`, "")
			if mutant == code {
				t.Fatal("mutant changed nothing")
			}
			mutantPath := filepath.Join(t.TempDir(), "omitted-fields.mjs")
			if err := os.WriteFile(mutantPath, []byte(mutant), 0600); err != nil {
				t.Fatal(err)
			}
			got := onNode(t, mutantPath)
			if got.exitCode != 70 || !strings.Contains(string(got.stderr), "TypeError") || disagreement(truth, got) == "" {
				t.Fatalf("omission survived or was caught outside runtime comparison: exit %d stderr %s", got.exitCode, got.stderr)
			}
			t.Log("empty fields omitted: caught by fixture's Node comparison (TypeError, exit 70)")
		})
	}
}
