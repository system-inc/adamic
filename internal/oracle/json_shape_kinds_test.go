package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Use the decoder's real schemas and allocations, then release the objects before
// checking the interned descriptors again. ASan checks the descriptors' ownership.
const jsonDecodeShapeProbe = `#include "json_decode.h"
#include <assert.h>
#include <stdio.h>
#include <string.h>
static const adamic_decode_field probe_fields[]={{"value",0,false}};
static const adamic_decode_node probe_boolean_nodes[]={
 {.kind="boolean",.representation=adamic_ir_Boolean,.expected="boolean"},
 {.kind="object",.representation=adamic_ir_Object,.expected="object",.field_count=1,.fields=probe_fields}
};
static const adamic_decode_node probe_number_nodes[]={
 {.kind="number",.representation=adamic_ir_Number,.expected="number"},
 {.kind="object",.representation=adamic_ir_Object,.expected="object",.field_count=1,.fields=probe_fields}
};
static const adamic_decode_schema probe_boolean={probe_boolean_nodes,1},probe_number={probe_number_nodes,1};
static const adamic_decode_node probe_null_nodes[]={
 {.kind="null",.representation=adamic_ir_NullRepresentation,.expected="null"}
};
static adamic_object *probe_decode(const adamic_decode_schema *schema,const char *json) {
 size_t length=strlen(json);
 adamic_string *text=adamic_string_allocate(length);
 memcpy((char *)text->bytes,json,length);
 adamic_object *result=adamic_json_decode(text,schema);
 adamic_release(text);return result;
}
int main(int argc,char **argv) {
 adamic_start(argc,argv);
 adamic_object *boolean=probe_decode(&probe_boolean,"{\"value\":true}");
 adamic_object *number=probe_decode(&probe_number,"{\"value\":7}");
 const adamic_shape *boolean_shape=((adamic_object *)boolean->slots[1].reference)->shape;
 const adamic_shape *number_shape=((adamic_object *)number->slots[1].reference)->shape;
 if(boolean_shape==number_shape) { fputs("decoder storage kinds share a shape\n",stderr);return 1; }
 assert(boolean_shape->kinds[0]==adamic_field_boolean && !boolean_shape->references[0]);
 assert(number_shape->kinds[0]==adamic_field_number && !number_shape->references[0]);
 adamic_release(boolean);adamic_release(number);
 for(size_t i=0;i<64;i++) {
  boolean=probe_decode(&probe_boolean,"{\"value\":false}");
  number=probe_decode(&probe_number,"{\"value\":8}");
  assert(((adamic_object *)boolean->slots[1].reference)->shape==boolean_shape);
  assert(((adamic_object *)number->slots[1].reference)->shape==number_shape);
  adamic_release(boolean);adamic_release(number);
 }
 assert(boolean_shape->kinds[0]==adamic_field_boolean);
 assert(number_shape->kinds[0]==adamic_field_number);
 const adamic_decode_schema scalar={probe_boolean_nodes,0};
 boolean=probe_decode(&scalar,"true");
 assert(boolean->shape->kinds[1]==adamic_field_boolean);
 adamic_release(boolean);
 // The C storage contract also covers unowned sentinel words; this does not
 // change the public library's existing nullable-type admission refusal.
 const adamic_decode_schema sentinel={probe_null_nodes,0};
 adamic_object *word=probe_decode(&sentinel,"null");
 assert(word->shape->kinds[1]==adamic_field_number && !word->shape->references[1]);
 adamic_release(word);
 puts("distinct decoder shapes");return 0;
}
`

func jsonDecodeShapeRun(t *testing.T, code string, sanitize bool) run {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "shape")
	if err := native.Build(code, binary, native.Options{Count: !sanitize, Sanitize: sanitize}); err != nil {
		t.Fatal(err)
	}
	return executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
}

func TestJSONDecodeShapeKinds(t *testing.T) {
	t.Parallel()
	for _, sanitize := range []bool{false, true} {
		got := jsonDecodeShapeRun(t, jsonDecodeShapeProbe, sanitize)
		stderrOK := strings.HasPrefix(string(got.stderr), "adamic: counts:")
		if sanitize {
			stderrOK = len(got.stderr) == 0
		}
		if got.exitCode != 0 || string(got.stdout) != "distinct decoder shapes\n" || !stderrOK {
			t.Fatalf("decoder shape control (sanitize=%t): exit %d stdout %s stderr %s", sanitize, got.exitCode, got.stdout, got.stderr)
		}
	}
}

func TestJSONDecodeShapeInternKindMutant(t *testing.T) {
	t.Parallel()
	original, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/json_decode.c"))
	if err != nil {
		t.Fatal(err)
	}
	before := "entry->shape.kinds[i]!=shape->kinds[i] || "
	if strings.Count(string(original), before) != 1 {
		t.Fatal("intern-kind mutant target changed")
	}
	mutant := strings.Replace(string(original), before, "", 1)
	got := jsonDecodeShapeRun(t, inlineDecodeMutant(mutant, jsonDecodeShapeProbe), false)
	if got.exitCode != 1 || !strings.Contains(string(got.stderr), "decoder storage kinds share a shape") {
		t.Fatalf("intern-kind mutant escaped TestJSONDecodeShapeKinds: exit %d stdout %s stderr %s", got.exitCode, got.stdout, got.stderr)
	}
}

func TestJSONDecodeShapeKindMutant(t *testing.T) {
	t.Parallel()
	original, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/json_decode.c"))
	if err != nil {
		t.Fatal(err)
	}
	before := "((adamic_field_kind *)o->shape->kinds)[written]=storage_kind(schema->nodes[f->node].representation);"
	if strings.Count(string(original), before) != 1 {
		t.Fatal("wrong-kind mutant target changed")
	}
	mutant := strings.Replace(string(original), before, "((adamic_field_kind *)o->shape->kinds)[written]=written==0?adamic_field_reference:storage_kind(schema->nodes[f->node].representation);", 1)
	path := filepath.Join(t.TempDir(), "kind.a")
	if err := os.WriteFile(path, []byte(`import {decodeJson} from 'adamic'; const value=decodeJson<{readonly value:boolean}>('{"value":true}'); console.log(value.kind);`), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	got := jsonDecodeShapeRun(t, inlineDecodeMutant(mutant, native.C(program)), false)
	if got.exitCode == 0 || len(got.stdout) != 0 || !strings.Contains(string(got.stderr), "adamic: inconsistent shape kind for field value") {
		t.Fatalf("wrong decoder kind escaped adamic_shape_check: exit %d stdout %s stderr %s", got.exitCode, got.stdout, got.stderr)
	}
}
