package oracle

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// Component seam tests complement the compiler-generated source controls.
func TestCheckedViewDictionaryComponents(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(t.TempDir(), "dictionary")
	if err := native.Build(dictionaryProbeC, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for _, variant := range []string{"options-good", "options-wrong", "nested-good", "nested-wrong", "array-wrong"} {
		t.Run(variant, func(t *testing.T) {
			path, err := filepath.Abs(checkedViewFixturePath("../../stage3/interface-downcasts/dictionaries/components/" + variant + ".a"))
			if err != nil {
				t.Fatal(err)
			}
			source := run{stdout: []byte("42\nundefined\n")}
			want := source
			switch variant {
			case "options-wrong":
				source.stdout = []byte("[object Object]\nundefined\n")
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.options['value']; expected string | number | boolean | undefined, found object\n")}
			case "nested-good":
				source.stdout = []byte("name\n")
				want = source
			case "nested-wrong":
				source.stdout = []byte("42\n")
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: entry.name is not a string; expected string, found number\n")}
			case "array-wrong":
				source.stdout = []byte("undefined\n")
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.symbols['item']; expected Entry | undefined, found array\n")}
			}
			if difference := disagreement(source, onNode(t, path)); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			js := filepath.Join(t.TempDir(), "dictionary.mjs")
			code := "import {panic} from 'adamic';\n" + javascript.DictionaryRuntime() + dictionaryProbeJS + "\nprobe(" + strconv.Quote(variant) + ");\n"
			if err := os.WriteFile(js, []byte(code), 0600); err != nil {
				t.Fatal(err)
			}
			for _, got := range []run{execute(t, binary, variant), onNode(t, js)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
		})
	}
}

// Every mutant runs valid release code and loses a pinned runtime refusal.
// The nested mutant drops the result contract at the adapter seam; it does not
// certify frontend aliases while source dispatch remains absent.
func TestCheckedViewDictionaryComponentMutants(t *testing.T) {
	t.Parallel()
	production, err := os.ReadFile(checkedViewFixturePath("../native/runtime/view_dictionaries.c"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []struct{ name, variant, oldC, newC, oldJS, newJS string }{
		{"skip-check", "options-wrong", "value.kind == adamic_view_union_unknown || (kinds & (1u << value.kind)) == 0", "kinds == 0 && value.kind == adamic_view_union_unknown", "!kinds.includes(kind)", "kinds.length === 0"},
		{"accept-wrong-shape", "array-wrong", "(kinds & (1u << value.kind)) == 0", "((kinds & (1u << value.kind)) == 0 && value.kind != adamic_view_union_array)", "!kinds.includes(kind)", "!kinds.includes(kind) && kind !== 'array'"},
		{"drop-transitive-contract", "nested-wrong", "reference ? child_contract : 0", "reference ? 0 : child_contract", "reference ? childContract : 0", "reference ? 0 : childContract"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			c := strings.Split(string(production), "\nstatic adamic_view_union_value dictionary_boxed_value")[0]
			jsRuntime := javascript.DictionaryRuntime()
			if strings.Count(c, mutant.oldC) != 1 || strings.Count(jsRuntime, mutant.oldJS) != 1 {
				t.Fatal("mutation anchor changed")
			}
			c = strings.Replace(c, mutant.oldC, mutant.newC, 1)
			c = strings.ReplaceAll(c, "adamic_view_dictionary_read(", "adamic_view_dictionary_read_mutant(")
			harness := strings.ReplaceAll(dictionaryProbeC, "adamic_view_dictionary_read(", "adamic_view_dictionary_read_mutant(")
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(c+harness, binary, native.Options{}); err != nil {
				t.Fatal("mutant must compile: ", err)
			}
			js := filepath.Join(t.TempDir(), "mutant.mjs")
			code := "import {panic} from 'adamic';\n" + strings.Replace(jsRuntime, mutant.oldJS, mutant.newJS, 1) + dictionaryProbeJS + "\nprobe(" + strconv.Quote(mutant.variant) + ");\n"
			if err := os.WriteFile(js, []byte(code), 0600); err != nil {
				t.Fatal(err)
			}
			for _, got := range []run{execute(t, binary, mutant.variant), onNode(t, js)} {
				if got.exitCode != 0 || len(got.stdout) == 0 {
					t.Fatalf("mutant must finish with a lost refusal: %#v", got)
				}
				t.Logf("pinned exit-70 diagnostic catches mutant: exit=%d stdout=%q", got.exitCode, got.stdout)
			}
		})
	}
}

const dictionaryProbeJS = `
const probe = variant => {
 const options = variant.startsWith('options');
 const value = options ? variant==='options-wrong' ? {} : 42 : variant==='array-wrong' ? ['name'] : {name:variant==='nested-wrong'?42:'name'};
 const record = {[options?'value':'item']:value};
 const expression = options ? "view.options['value']" : "view.symbols['item']";
 const declared = options ? 'string | number | boolean | undefined' : 'Entry | undefined';
 const kinds = options ? ['string','number','boolean','undefined'] : ['object','undefined'];
 const entry = adamicViewDictionaryRead(record,options?'value':'item',kinds,41,expression,declared);
 if(options) {
  console.log(String(entry.value));
  console.log(String(adamicViewDictionaryRead(record,'missing',kinds,41,"view.options['missing']",declared).value));
 } else if(Array.isArray(entry.value)) {
  console.log('undefined');
 } else {
  if(entry.contract && typeof entry.value.name !== 'string') panic('field read failed: entry.name is not a string; expected string, found '+typeof entry.value.name);
  console.log(String(entry.value.name));
 }
};
`

const dictionaryProbeC = `
#include "view_dictionaries.h"
#include <stdio.h>
#include <string.h>
static const char *const names[]={"name"};
static const bool refs[]={true};
static const adamic_shape shape={1,names,refs,NULL};
static adamic_string *text(const char *s) {
 size_t length=strlen(s);
 adamic_string *r=adamic_string_allocate(length);
 memcpy((char *)r->bytes,s,length);
 return r;
}
int main(int argc,char **argv) {
 if(argc!=2)return 2;
 const char *variant=argv[1];
 bool options=strncmp(variant,"options",7)==0;
 bool wrong=strcmp(variant,"options-wrong")==0;
 bool nested=strcmp(variant,"nested-wrong")==0;
 bool array=strcmp(variant,"array-wrong")==0;
 adamic_record *record=adamic_record_new(true);
 adamic_heap *payload;
 if(options && !wrong)payload=adamic_box_number(42);
 else if(array)payload=(adamic_heap *)adamic_array_new(0,true);
 else {
  adamic_object *child=adamic_object_new(&shape);
  child->slots[0].reference=nested?(void *)adamic_box_number(42):(void *)text("name");
  adamic_object_field_types(child)[0]=nested?10:3;
  adamic_object_initialized(child)[0]=true;
  payload=(adamic_heap *)child;
 }
 adamic_string *key=text(options?"value":"item");
 adamic_record_define(record,adamic_retain(key),(adamic_value){.reference=payload});
 unsigned int kinds=options?(1u<<adamic_view_union_string)|(1u<<adamic_view_union_number)|(1u<<adamic_view_union_boolean)|(1u<<adamic_view_union_undefined):(1u<<adamic_view_union_object)|(1u<<adamic_view_union_undefined);
 const char *declared=options?"string | number | boolean | undefined":"Entry | undefined";
 adamic_view_dictionary_result result=adamic_view_dictionary_read(record,true,key,kinds,41,options?"view.options['value']":"view.symbols['item']",declared);
 if(options) {
  if(result.value.kind==adamic_view_union_number)printf("%.0f\n",result.value.payload.number);else puts("[object Object]");
  adamic_string *missing=text("missing");
  (void)adamic_view_dictionary_read(record,true,missing,kinds,41,"view.options['missing']",declared);
  puts("undefined");
  adamic_release(missing);
 } else if(array) puts("undefined");
 else {
  adamic_object *child=result.value.payload.reference;
  adamic_slot_cache cache={NULL,0};
  if(result.contract!=0) {
   adamic_value name=adamic_object_view(child,"name",&cache,3,"string","entry.name");
   adamic_string *s=name.reference;
   fwrite(s->bytes,1,s->length,stdout);putchar('\n');
  } else puts(nested?"42":"name");
 }
 adamic_release(key);adamic_release(record);
 return 0;
}
`

func TestCheckedViewDictionaryJavaScriptGuards(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, input, kinds, contract, found, old, replacement string }{
		{"container", "42", "['undefined']", "41", "number", "record === null || typeof record !== 'object' || Array.isArray(record) || record instanceof Map", "false"},
		{"accessor", "{get item(){console.log('getter ran');return 42}}", "['undefined']", "41", "accessor", "slot !== undefined && !Object.hasOwn(slot, 'value')", "false"},
		{"missing-child", "{item:{}}", "['object']", "0", "unsupported representation", "reference && !childContract", "false"},
		{"required-missing", "{}", "['string']", "41", "undefined", "!kinds.includes(kind)", "false"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			body := "\nadamicViewDictionaryRead(" + probe.input + ",'item'," + probe.kinds + "," + probe.contract + ",'view.item','TestContract');console.log('passed');\n"
			want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.item; expected TestContract, found " + probe.found + "\n")}
			for _, mutated := range []bool{false, true} {
				runtime := javascript.DictionaryRuntime()
				if mutated {
					runtime = strings.Replace(runtime, probe.old, probe.replacement, 1)
				}
				js := filepath.Join(t.TempDir(), "guard.mjs")
				if err := os.WriteFile(js, []byte("import {panic} from 'adamic';\n"+runtime+body), 0600); err != nil {
					t.Fatal(err)
				}
				got := onNode(t, js)
				if !mutated {
					if difference := disagreement(want, got); difference != "" {
						t.Fatalf("%s; got %#v", difference, got)
					}
				} else if got.exitCode != 0 || string(got.stdout) != "passed\n" || len(got.stderr) != 0 {
					t.Fatalf("guard mutant must lose only the pinned refusal: %#v", got)
				}
			}
		})
	}
}

func TestCheckedViewDictionaryNativeStorageGuard(t *testing.T) {
	t.Parallel()
	production, err := os.ReadFile(checkedViewFixturePath("../native/runtime/view_dictionaries.c"))
	if err != nil {
		t.Fatal(err)
	}
	harness := strings.ReplaceAll(dictionaryProbeC, "adamic_view_dictionary_read(", "adamic_view_dictionary_read_guard(")
	harness = strings.ReplaceAll(harness, "record,true,key,", "record,false,key,")
	want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.options['value']; expected string | number | boolean | undefined, found unsupported representation\n")}
	for _, mutated := range []bool{false, true} {
		source := strings.ReplaceAll(strings.Split(string(production), "\nstatic adamic_view_union_value dictionary_boxed_value")[0], "adamic_view_dictionary_read(", "adamic_view_dictionary_read_guard(")
		if mutated {
			source = strings.Replace(source, "!boxed_storage || record == NULL || key == NULL", "!boxed_storage && record == NULL && key == NULL", 1)
		}
		binary := filepath.Join(t.TempDir(), "guard")
		if err := native.Build(source+harness, binary, native.Options{}); err != nil {
			t.Fatal("guard mutant must compile: ", err)
		}
		got := execute(t, binary, "options-good")
		if !mutated {
			if difference := disagreement(want, got); difference != "" {
				t.Fatalf("%s; got %#v", difference, got)
			}
		} else if got.exitCode != 0 || string(got.stdout) != "42\nundefined\n" {
			t.Fatalf("source-certificate mutant must lose the pinned refusal: %#v", got)
		}
	}
}
