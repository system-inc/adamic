package oracle

import (
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Helpers receive viewed and ordinary objects of the same static interface.
// A guard cannot depend on a cast appearing lexically inside the helper.
func TestCheckedViewLane4HelperReads(t *testing.T) {
	for _, fixture := range []string{"helper-good", "helper-wrong"} {
		t.Run(fixture, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane4/read-fixtures/"+fixture)
			sourceWant := run{stdout: []byte("true\ntrue\n")}
			want := sourceWant
			if fixture == "helper-wrong" {
				sourceWant.stdout = []byte("true\n42\n")
				want = run{stdout: []byte("true\n"), exitCode: 70, stderr: []byte("adamic: panic: field read failed: value.unsupported is not a boolean; expected boolean, found number\n")}
			}
			if difference := disagreement(sourceWant, onNode(t, path)); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
			if fixture == "helper-wrong" {
				dropped := dropLane4HelperView(program)
				if dropped != 1 {
					t.Fatalf("want exactly one helper read mutation, got %d", dropped)
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if got.exitCode != 0 {
						t.Fatalf("mutant must run valid release code, got %#v", got)
					}
					if disagreement(want, got) == "" {
						t.Fatal("unchecked helper mutant escaped pinned read refusal")
					}
					t.Logf("unchecked helper mutant caught: exit %d stdout %q", got.exitCode, got.stdout)
				}
			}
		})
	}
}

// Primitive unions are admitted lazily; the original wrong boolean must still
// fail at the helper read even when an ordinary valid object reaches that helper.
func TestCheckedViewLane4UnsupportedHelper(t *testing.T) {
	input, err := os.ReadFile(checkedViewFixturePath("../../stage3/interface-downcasts/lane4/read-fixtures/helper-unsupported.a"))
	if err != nil {
		t.Fatal(err)
	}
	for _, probe := range []struct{ name, value, output string }{
		{"string", "'word'", "word\n"}, {"number", "42", "42\n"}, {"wrong-boolean", "true", "true\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			source := strings.Replace(string(input), "unsupported: true", "unsupported: "+probe.value, 1)
			source = strings.Replace(source, "helper(base as Box);", "const ordinary: Box = {kind: 'box', unsupported: 'ordinary'}; helper(ordinary); helper(base as Box);", 1)
			path := filepath.Join(t.TempDir(), "helper.a")
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			sourceWant := run{stdout: []byte("ordinary\n" + probe.output)}
			if difference := disagreement(sourceWant, onNode(t, path)); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := sourceWant
			if probe.name == "wrong-boolean" {
				want = run{stdout: []byte("ordinary\n"), exitCode: 70, stderr: []byte("adamic: panic: field read failed: value.unsupported matches no member of string | number; expected string | number, found boolean\n")}
			}
			actual, binary := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s: stdout %q stderr %q", difference, got.stdout, got.stderr)
				}
			}
			if probe.name != "wrong-boolean" {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
				return
			}
			if count := skipPrimitiveMemberChecks(program, "unsupported"); count != 1 {
				t.Fatalf("want one helper member-check mutation, got %d", count)
			}
			for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(sourceWant, got); difference != "" {
					t.Fatalf("mutant must execute the wrong source value: %s, stderr %q", difference, got.stderr)
				}
				if disagreement(want, got) == "" {
					t.Fatal("primitive helper mutant escaped the refusal pin")
				}
				t.Logf("primitive helper member-check mutant caught: exit %d stdout %q", got.exitCode, got.stdout)
			}
		})
	}
}

func dropLane4HelperView(program *ir.Program, replacement ...ir.Expression) int {
	return dropLane4NamedHelperView(program, "unsupported", replacement...)
}

func dropLane4NamedHelperView(program *ir.Program, name string, replacement ...ir.Expression) int {
	dropped := 0
	var rewrite func(reflect.Value) reflect.Value
	rewrite = func(v reflect.Value) reflect.Value {
		switch v.Kind() {
		case reflect.Interface:
			if v.IsNil() {
				return v
			}
			out := reflect.New(v.Type()).Elem()
			out.Set(rewrite(v.Elem()))
			return out
		case reflect.Struct:
			out := reflect.New(v.Type()).Elem()
			for i := 0; i < v.NumField(); i++ {
				out.Field(i).Set(rewrite(v.Field(i)))
			}
			if field, ok := out.Interface().(ir.Property); ok && field.Name == name && field.View != "" {
				dropped++
				if len(replacement) != 0 {
					return reflect.ValueOf(replacement[0])
				}
				field.View = ""
				out.Set(reflect.ValueOf(field))
			}
			return out
		case reflect.Slice:
			if v.IsNil() {
				return v
			}
			out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
			for i := 0; i < v.Len(); i++ {
				out.Index(i).Set(rewrite(v.Index(i)))
			}
			return out
		default:
			return v
		}
	}
	for i := range program.Functions {
		program.Functions[i].Body = rewrite(reflect.ValueOf(program.Functions[i].Body)).Interface().([]ir.Statement)
	}
	return dropped
}

// The selector is held to source Node here independently of frontend admission.
// These normalized snapshots are a component oracle, not unlocked pair claims.
func TestCheckedViewMixedSelection(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "mixed-selection")
	if err := native.Build(mixedSelectionC, binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	sanitized := filepath.Join(t.TempDir(), "mixed-selection-sanitized")
	if err := native.Build(mixedSelectionC, sanitized, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for _, sample := range []struct{ name, stdout, nodeWrong, declared, found string }{
		{"string-number-undefined", "string:word\nnumber:42\nundefined:undefined\n", "boolean:true\n", "string | number | undefined", "boolean"},
		{"string-object", "string:word\nobject:name\n", "object:42\n", "string | Identifier", "object"},
		{"untagged-objects", "Left:name\nRight:42\n", "Right:true\n", "Left | Right", "object"},
		{"false-string-undefined", "boolean:false\nstring:word\nundefined:undefined\n", "boolean:true\n", "false | string | undefined", "boolean"},
	} {
		t.Run(sample.name, func(t *testing.T) {
			path, err := filepath.Abs(checkedViewFixturePath("../../stage3/interface-downcasts/lane4/selection-fixtures/" + sample.name + "-good.a"))
			if err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte(sample.stdout)}
			if difference := disagreement(want, onNode(t, path)); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			for _, got := range []run{execute(t, binary, sample.name, "good"), execute(t, sanitized, sample.name, "good"), mixedSelectionJavaScript(t, sample.name, "good")} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
			path = strings.Replace(path, "-good.a", "-wrong.a", 1)
			if difference := disagreement(run{stdout: []byte(sample.nodeWrong)}, onNode(t, path)); difference != "" {
				t.Fatal("wrong source Node: " + difference)
			}
			want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.value matches no member of " + sample.declared + "; expected " + sample.declared + ", found " + sample.found + "\n")}
			for _, got := range []run{execute(t, binary, sample.name, "wrong"), executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, sanitized, sample.name, "wrong"), mixedSelectionJavaScript(t, sample.name, "wrong")} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
		})
	}
}

func mixedSelectionJavaScript(t *testing.T, family, variant string) run {
	t.Helper()
	path := filepath.Join(t.TempDir(), "selection.mjs")
	source := "import {panic} from 'adamic';\n" + javascript.MixedUnionRuntime() + mixedSelectionJavaScriptCases + "\nprobe(" + strconv.Quote(family) + ", " + strconv.Quote(variant) + ");\n"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return onNode(t, path)
}

const mixedSelectionJavaScriptCases = `
const kind = value => value === undefined ? 'undefined' : value === null ? 'null' : typeof value;
const match = (member, snapshot) => { const descriptor=Object.getOwnPropertyDescriptor(snapshot.value,'text'); return descriptor !== undefined && Object.hasOwn(descriptor,'value') && typeof descriptor.value === (member.contract === 1 ? 'string' : 'number'); };
const probe = (family, variant) => {
 let members, values, declared;
 if (family === 'string-number-undefined') {
  members=[{kind:'string'},{kind:'number'},{kind:'undefined'}]; values=variant==='good'?['word',42,undefined]:[true];declared='string | number | undefined';
 } else if (family === 'string-object') {
  members=[{kind:'string'},{kind:'object',contract:1}]; values=variant==='good'?['word',{text:'name'}]:[{text:42}];declared='string | Identifier';
 } else if (family === 'untagged-objects') {
  members=[{kind:'object',contract:1},{kind:'object',contract:2}];values=variant==='good'?[{text:'name'},{text:42}]:[{text:true}];declared='Left | Right';
 } else {
  members=[{kind:'boolean',literal:true,value:false},{kind:'string'},{kind:'undefined'}];values=variant==='good'?[false,'word',undefined]:[true];declared='false | string | undefined';
 }
 for (const value of values) {
  const selected=adamicViewMixedUnionSelect({kind:kind(value),value},members,match,'view.value',declared);
  const prefix=family==='untagged-objects'?(selected===0?'Left':'Right'):members[selected].kind;
  console.log(prefix+':'+String(typeof value==='object'?value.text:value));
 }
};
`

const mixedSelectionC = `
#include "view_unions_mixed.h"
#include <stdio.h>
#include <string.h>
static adamic_string word = ADAMIC_STRING("word");
static adamic_string name = ADAMIC_STRING("name");
typedef struct sample_object { adamic_view_union_value text; } sample_object;
static bool matches(void *context, const adamic_view_union_member *member, const adamic_view_union_value *value) {
 (void)context;
 const sample_object *object=value->payload.reference;
 return object->text.kind==(member->contract==1?adamic_view_union_string:adamic_view_union_number);
}
static void show(const char *prefix, adamic_view_union_value value) {
 if(value.kind==adamic_view_union_object){const sample_object *object=value.payload.reference;value=object->text;}
 printf("%s:",prefix);
 switch(value.kind){
 case adamic_view_union_string: {const adamic_string *s=value.payload.reference;printf("%.*s",(int)s->length,s->bytes);break;}
 case adamic_view_union_number:printf("%g",value.payload.number);break;
 case adamic_view_union_boolean:printf("%s",value.payload.boolean?"true":"false");break;
 case adamic_view_union_undefined:printf("undefined");break;
 default:printf("unsupported");break;
 }
 printf("\n");
}
int main(int argc,char **argv){
 if(argc!=3)return 2;
 const char *family=argv[1];bool wrong=strcmp(argv[2],"wrong")==0;
 sample_object text={{adamic_view_union_string,{.reference=&name}}};
 sample_object number={{adamic_view_union_number,{.number=42}}};
 sample_object boolean={{adamic_view_union_boolean,{.boolean=true}}};
 adamic_view_union_member members[3]={0};adamic_view_union_value values[3]={0};size_t count=0,amount=0;const char *declared=NULL;
 if(strcmp(family,"string-number-undefined")==0){
  members[0].kind=adamic_view_union_string;members[1].kind=adamic_view_union_number;members[2].kind=adamic_view_union_undefined;count=3;
  values[0]=(adamic_view_union_value){adamic_view_union_string,{.reference=&word}};values[1]=(adamic_view_union_value){adamic_view_union_number,{.number=42}};values[2]=(adamic_view_union_value){adamic_view_union_undefined,{.reference=NULL}};amount=3;declared="string | number | undefined";
  if(wrong){values[0]=(adamic_view_union_value){adamic_view_union_boolean,{.boolean=true}};amount=1;}
 }else if(strcmp(family,"string-object")==0){
  members[0].kind=adamic_view_union_string;members[1].kind=adamic_view_union_object;members[1].contract=1;count=2;
  values[0]=(adamic_view_union_value){adamic_view_union_string,{.reference=&word}};values[1]=(adamic_view_union_value){adamic_view_union_object,{.reference=&text}};amount=2;declared="string | Identifier";
  if(wrong){values[0]=(adamic_view_union_value){adamic_view_union_object,{.reference=&number}};amount=1;}
 }else if(strcmp(family,"untagged-objects")==0){
  members[0].kind=adamic_view_union_object;members[0].contract=1;members[1].kind=adamic_view_union_object;members[1].contract=2;count=2;
  values[0]=(adamic_view_union_value){adamic_view_union_object,{.reference=&text}};values[1]=(adamic_view_union_value){adamic_view_union_object,{.reference=&number}};amount=2;declared="Left | Right";
  if(wrong){values[0]=(adamic_view_union_value){adamic_view_union_object,{.reference=&boolean}};amount=1;}
 }else{
  members[0]=(adamic_view_union_member){adamic_view_union_boolean,true,{.boolean=false},0};members[1].kind=adamic_view_union_string;members[2].kind=adamic_view_union_undefined;count=3;
  values[0]=(adamic_view_union_value){adamic_view_union_boolean,{.boolean=false}};values[1]=(adamic_view_union_value){adamic_view_union_string,{.reference=&word}};values[2]=(adamic_view_union_value){adamic_view_union_undefined,{.reference=NULL}};amount=3;declared="false | string | undefined";
  if(wrong){values[0].payload.boolean=true;amount=1;}
 }
 for(size_t i=0;i<amount;i++){
  size_t selected=adamic_view_mixed_union_select(&values[i],members,count,matches,NULL,"view.value",declared);
  const char *prefix=strcmp(family,"untagged-objects")==0?(selected==0?"Left":"Right"):adamic_view_union_kind_name(members[selected].kind);
  show(prefix,values[i]);
 }
 return 0;
}
`

// Primitive phantom brands erase only after their approved primitive base is known.
func TestCheckedViewBrandString(t *testing.T) {
	for _, fixture := range []string{"good", "wrong", "literal-good", "literal-wrong"} {
		t.Run(fixture, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane4/read-fixtures/brand-string-"+fixture)
			sourceWant := run{stdout: []byte("word\n")}
			want := sourceWant
			if strings.HasSuffix(fixture, "wrong") {
				sourceWant.stdout = []byte("42\n")
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: value.unsupported is not a __String; expected __String, found number\n")}
			}
			if fixture == "literal-wrong" {
				sourceWant.stdout = []byte("different\n")
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: value.unsupported expected __String, found string different\n")}
			}
			if difference := disagreement(sourceWant, onNode(t, path)); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
			if strings.HasSuffix(fixture, "good") {
				got, _ := nativelyUncached(t, program)
				if difference := disagreement(want, got); difference != "" {
					t.Fatal(difference)
				}
			} else {
				index := len(program.Strings)
				program.Strings = append(program.Strings, "unchecked")
				if count := dropLane4HelperView(program, ir.StringConstant{Index: index}); count != 1 {
					t.Fatalf("want one mutated read, got %d", count)
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if got.exitCode != 0 {
						t.Fatalf("mutant must execute valid release code: %#v", got)
					}
					if disagreement(want, got) == "" {
						t.Fatal("brand read mutant escaped refusal")
					}
					t.Logf("brand read mutant caught: exit %d stdout %q", got.exitCode, got.stdout)
				}
			}
		})
	}
}

func TestCheckedViewCompleteBrand(t *testing.T) {
	for _, fixture := range []string{"good", "undefined", "wrong", "null", "missing", "internal", "object", "symbol-good", "symbol-undefined", "symbol-wrong", "symbol-null"} {
		t.Run(fixture, func(t *testing.T) {
			prefix, variant, field := "brand-complete-", fixture, "escapedText"
			if strings.HasPrefix(fixture, "symbol-") {
				prefix, variant, field = "brand-symbol-", strings.TrimPrefix(fixture, "symbol-"), "escapedName"
			}
			program, path := interfaceFixture(t, "lane4/read-fixtures/"+prefix+variant)
			text := map[string]string{"good": "word", "undefined": "undefined", "wrong": "42", "null": "null", "missing": "undefined", "internal": "__call", "object": "[object Object]"}[variant]
			if difference := disagreement(run{stdout: []byte(text + "\n")}, onNode(t, path)); difference != "" {
				t.Fatal(difference)
			}
			want := run{stdout: []byte(text + "\n")}
			if variant == "wrong" || variant == "null" || variant == "object" {
				found := map[string]string{"wrong": "number", "null": "null", "object": "object"}[variant]
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: value." + field + " is not a __String; expected __String, found " + found + "\n")}
			}
			if variant == "missing" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: value." + field + " is not initialized; expected __String, found missing\n")}
			}
			for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
			if want.exitCode == 0 {
				got, _ := nativelyUncached(t, program)
				if difference := disagreement(want, got); difference != "" {
					t.Fatal(difference)
				}
			}
			if want.exitCode != 0 {
				index := len(program.Strings)
				program.Strings = append(program.Strings, "unchecked")
				if count := dropLane4NamedHelperView(program, field, ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: index}, ir.StringConstant{Index: index}}}); count != 1 {
					t.Fatalf("expected one helper field mutation, got %d", count)
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if got.exitCode != 0 {
						t.Fatalf("read-removal mutant must run valid release code: %#v", got)
					}
					if disagreement(want, got) == "" {
						t.Fatal("brand read bypass escaped refusal pin")
					}
					t.Logf("member-check removal caught: exit %d stdout %q", got.exitCode, got.stdout)
				}
			}
			if variant == "undefined" {
				first := len(program.Strings)
				program.Strings = append(program.Strings, "word")
				if count := dropLane4NamedHelperView(program, field, ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: first}, ir.StringConstant{Index: first}}}); count != 1 {
					t.Fatalf("expected one first-member mutation, got %d", count)
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if got.exitCode != 0 {
						t.Fatalf("first-member mutant must execute valid release code: %#v", got)
					}
					if disagreement(want, got) == "" {
						t.Fatal("untested first-member substitution escaped Node control")
					}
					t.Logf("first-member substitution caught: exit %d stdout %q", got.exitCode, got.stdout)
				}
				return
			}
		})
	}
}
