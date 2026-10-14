package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestSequenceArrayNodeMutant(t *testing.T) {
	t.Parallel()
	sequenceNextNodeMutant(t, "library_iterator_array_checked.a", "skip")
}
func TestSequenceStringNodeMutant(t *testing.T) {
	t.Parallel()
	sequenceNextNodeMutant(t, "library_iterator_string_next.a", "skip")
}
func TestSequenceMapNodeMutant(t *testing.T) {
	t.Parallel()
	sequenceNextNodeMutant(t, "library_iterator_map_set_live.a", "map_skip")
}
func TestSequenceSetNodeMutant(t *testing.T) {
	t.Parallel()
	sequenceNextNodeMutant(t, "library_iterator_map_set_live.a", "set_skip")
}
func TestSequenceStringCodePointNodeMutant(t *testing.T) {
	t.Parallel()
	sequenceNextNodeMutant(t, "library_iterator_string_next.a", "unit")
}
func TestSequenceCallbackNodeMutant(t *testing.T) {
	t.Parallel()
	sequenceNextNodeMutant(t, "library_iterator_tsc_helpers.a", "skip")
}
func TestSequenceToArrayNodeMutant(t *testing.T) {
	t.Parallel()
	sequenceNextNodeMutant(t, "library_iterator_to_array.a", "skip")
}
func TestSequenceExhaustionNodeMutant(t *testing.T) {
	t.Parallel()
	sequenceNextNodeMutant(t, "library_iterator_array_checked.a", "revive")
}

// Change the real runtime closure's behavior without rebuilding unrelated archives.
// The wrapper forwards the existing closure and its actual counted state.
func sequenceNextNodeMutant(t *testing.T, fixture, mode string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/sequence_iterators", fixture))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	changed := strings.ReplaceAll(source, "adamic_collection_iterator(", "sequence_mutant_factory(")
	if source == changed {
		t.Fatal("mutant changed nothing")
	}
	body := `adamic_value skipped = original(self, arguments); adamic_release(skipped.reference); return original(self, arguments);`
	if mode == "map_skip" || mode == "set_skip" {
		condition := "!state->slots[4].boolean"
		if mode == "set_skip" {
			condition = "state->slots[4].boolean"
		}
		body = `adamic_object *state = self->cells[0]->value.reference; if (` + condition + `) { adamic_value skipped=original(self,arguments); adamic_release(skipped.reference); } return original(self,arguments);`
	}
	if mode == "unit" {
		body = `adamic_object *state=self->cells[0]->value.reference; double previous=state->slots[2].number; adamic_value result=original(self,arguments); adamic_object *step=result.reference; if (!step->slots[0].boolean && state->slots[2].number>previous+1) state->slots[2].number=previous+1; return result;`
	}
	if mode == "revive" {
		body = `adamic_object *state = self->cells[0]->value.reference; if (((adamic_heap *)state->slots[0].reference)->kind == adamic_kind_array) state->slots[4].boolean = false; return original(self, arguments);`
	}
	prefix := `#include "adamic.h"
static adamic_code original;
static adamic_value sequence_mutant_next(adamic_closure *self, adamic_value *arguments) { ` + body + ` }
static adamic_object *sequence_mutant_factory(void *collection, int part, int key, int value, bool set) {
 adamic_object *object = adamic_collection_iterator(collection,part,key,value,set);
 adamic_closure *next = object->slots[0].reference;
 original = next->code;
 next->code = sequence_mutant_next;
 return object;
}
`
	sequenceNodeOnly(t, prefix+changed, path, false)
}

func TestSequenceDoneNodeMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/sequence_iterators/library_overlay_done.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	mutant := regexp.MustCompile(`(} else \{\n\s*adamic_temporary_[0-9]+ = )false;`).ReplaceAllString(source, "${1}true;")
	if mutant == source {
		t.Fatal("mutant changed nothing")
	}
	sequenceNodeOnly(t, mutant, path, false)
}

func TestRuntimeMapTupleFlagNodeMutant(t *testing.T) {
	t.Parallel()
	runtimeTupleFlagNodeMutant(t, "map_entries")
}
func TestRuntimeArrayTupleFlagNodeMutant(t *testing.T) {
	t.Parallel()
	runtimeTupleFlagNodeMutant(t, "array_entries")
}
func TestRuntimeObjectTupleFlagNodeMutant(t *testing.T) {
	t.Parallel()
	runtimeTupleFlagNodeMutant(t, "object_entries")
}
func TestRuntimeCheckedObjectTupleFlagNodeMutant(t *testing.T) {
	t.Parallel()
	runtimeTupleFlagNodeMutant(t, "object_entries_checked")
}

func runtimeTupleFlagNodeMutant(t *testing.T, family string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/sequence_iterators/runtime_tuple_"+family+".a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	var prefix, mutant string
	if family == "map_entries" || family == "array_entries" {
		mutant = strings.ReplaceAll(source, "adamic_collection_iterator(", "tuple_mutant_factory(")
		prefix = `#include "adamic.h"
static adamic_code original;
static adamic_value tuple_mutant_next(adamic_closure *self, adamic_value *arguments) {
 adamic_value result = original(self,arguments);
 adamic_object *step = result.reference;
 if (!step->slots[0].boolean) ((adamic_object *)step->slots[1].reference)->tuple = false;
 return result;
}
static adamic_object *tuple_mutant_factory(void *collection,int part,int key,int value,bool set) {
 adamic_object *object=adamic_collection_iterator(collection,part,key,value,set);
 adamic_closure *next=object->slots[0].reference; original=next->code; next->code=tuple_mutant_next; return object;
}
`
	} else {
		prefix = `#include "adamic.h"
static adamic_array *clear_tuple_flags(adamic_array *pairs) {
 for (size_t i=0;i<pairs->length;i++) ((adamic_object *)pairs->elements[i].reference)->tuple=false;
 return pairs;
}
`
		if family == "object_entries" {
			mutant = strings.ReplaceAll(source, "adamic_object_values(", "tuple_mutant_values(")
			prefix += `static adamic_array *tuple_mutant_values(const adamic_object *object,bool references,bool entries) {
 return clear_tuple_flags(adamic_object_values(object,references,entries));
}
`
		} else {
			mutant = strings.ReplaceAll(source, "adamic_object_values_checked(", "tuple_mutant_checked(")
			prefix += `extern adamic_array *adamic_object_values_checked(adamic_object *,int,const char *,bool,const adamic_value *,size_t);
static adamic_array *tuple_mutant_checked(adamic_object *object,int expected,const char *name,bool entries,const adamic_value *allowed,size_t count) {
 return clear_tuple_flags(adamic_object_values_checked(object,expected,name,entries,allowed,count));
}
`
		}
	}
	if source == mutant {
		t.Fatal("mutant changed nothing")
	}
	sequenceNodeOnly(t, prefix+mutant, path, true)
}

func sequenceNodeOnly(t *testing.T, source, path string, checked bool) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	truth := onNode(t, path)
	if checked {
		if got.exitCode != 70 || !strings.Contains(string(got.stderr), "field read failed") || strings.Contains(string(got.stderr), "Sanitizer") || disagreement(truth, got) != "exit codes differ" {
			t.Fatalf("tuple flag mutant failed outside Node checked-view comparison: truth %+v got %+v", truth, got)
		}
		t.Log("cleared tuple flag rejected by checked view; Node comparison catches exit 70 versus exit 0; no sanitizer finding")
	} else {
		if got.exitCode != 0 || len(got.stderr) != 0 || disagreement(truth, got) != "stdout differs" {
			t.Fatalf("mutant failed outside Node stdout: truth %+v got %+v", truth, got)
		}
		t.Log("Node stdout alone catches mutation; exit 0, empty stderr, ASan/UBSan/LSan clean")
	}
}

// Keep the prototype layout's observable facts held to Node's own built-ins.
func TestSequenceIteratorPrototypeNode(t *testing.T) {
	t.Parallel()
	const truth = `const a=[].values(),b=[].keys(),m=new Map().keys(),s=new Set().values(),t=''[Symbol.iterator](); console.log([Object.hasOwn(a,'next'),Object.getPrototypeOf(a)===Object.getPrototypeOf(b),a.next===b.next,a.next===m.next,m.next===s.next,a[Symbol.iterator]===m[Symbol.iterator],a[Symbol.iterator]()===a,t[Symbol.iterator]()===t].map(Number).join(' '));`
	source := `#include "adamic.h"
#include <stdio.h>
#ifdef ADAMIC_CLOSURE_CONVENTION
#define METHOD(entry) (entry).code
#else
#define METHOD(entry) (entry)
#endif
int main(void) {
 adamic_array *array=adamic_array_new(0,false);adamic_map *map=adamic_map_new(false,false);adamic_map *set=adamic_map_new(false,false);
 static adamic_string empty=ADAMIC_STRING("");static adamic_string next=ADAMIC_STRING("next");
 adamic_object *a=adamic_collection_iterator(array,2,1,1,false),*b=adamic_collection_iterator(array,1,1,1,false),*m=adamic_collection_iterator(map,1,1,1,false),*s=adamic_collection_iterator(set,2,1,1,true),*text=adamic_collection_iterator(&empty,1,1,3,false);
 adamic_value identity=METHOD(a->shape->methods->code[1])(a,NULL), string_identity=METHOD(text->shape->methods->code[1])(text,NULL);
 printf("%d %d %d %d %d %d %d %d\n",adamic_object_has(a,&next),a->shape->methods==b->shape->methods,METHOD(a->shape->methods->code[0])==METHOD(b->shape->methods->code[0]),METHOD(a->shape->methods->code[0])==METHOD(m->shape->methods->code[0]),METHOD(m->shape->methods->code[0])==METHOD(s->shape->methods->code[0]),METHOD(a->shape->methods->code[1])==METHOD(m->shape->methods->code[1]),identity.reference==a,string_identity.reference==text);
 adamic_release(identity.reference);adamic_release(string_identity.reference);adamic_release(a);adamic_release(b);adamic_release(m);adamic_release(s);adamic_release(text);adamic_release(array);adamic_release(map);adamic_release(set);return 0;
}
`
	binary := filepath.Join(t.TempDir(), "prototype")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	expected := executeWith(t, nil, "node", "--input-type=module", "-e", truth)
	if difference := disagreement(expected, actual); difference != "" {
		t.Fatalf("Node prototype: %s, got %+v", difference, actual)
	}
}

// Project loading is exercised with authored .a sources and a .a project root.
func TestSequenceProjectNode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"iterable", "callback", "entries", "nested_entries", "done"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			data, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/sequence_iterators/library_overlay_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			directory := t.TempDir()
			path := filepath.Join(directory, "main.a")
			if err := os.WriteFile(path, data, 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(`{"compilerOptions":{"strict":true,"target":"es2024","lib":["es2024"],"types":[],"noEmit":true},"files":["main.a"]}`), 0644); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			expected := onNode(t, path)
			actual, binary := natively(t, program)
			if disagreement(expected, actual) != "" || disagreement(expected, onJavaScriptBackend(t, program)) != "" || disagreement(expected, released(t, program)) != "" {
				t.Fatal("project backend differs from Node")
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}

func TestSequenceIteratorTagNodeMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/sequence_iterators/library_iterator_tags.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	index := len(program.Strings)
	program.Strings = append(program.Strings, "Object")
	changed := 0
	mutate := func(value ir.Expression) ir.Expression {
		if call, ok := value.(ir.CallClosure); ok {
			if property, ok := call.Closure.(ir.Property); ok && property.Name == "__adamic_iterator_tag" {
				changed++
				return ir.StringConstant{Index: index}
			}
		}
		return value
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
	for i := range program.Functions {
		mutateStringExpressions(reflect.ValueOf(&program.Functions[i].Body).Elem(), mutate)
	}
	if changed == 0 {
		t.Fatal("mutant changed nothing")
	}
	sequenceNodeOnly(t, native.C(program), path, false)
}
