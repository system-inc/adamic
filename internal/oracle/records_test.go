package oracle

import (
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/records_refuse/r24.a", false, false})
	for _, name := range []string{"operations", "ownership", "for_in", "census", "scalars", "optional", "narrowed_number", "narrowed_reference", "prototype_read", "prototype_in", "prototype_set"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/records_" + name + ".a", true, name == "prototype_read" || name == "prototype_in" || name == "prototype_set" || name == "narrowed_number"})
	}
}

// The native oracle also consumes these fixtures. This separate check allows the
// JavaScript lowering to be held to Node while checked native declarations arrive.
func TestRecordJavaScriptAgreesWithNode(t *testing.T) {
	for _, name := range []string{"operations", "ownership", "for_in", "census", "scalars", "optional"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			p, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want, got := onNode(t, path), onJavaScriptBackend(t, p)
			if diff := disagreement(want, got); diff != "" {
				t.Fatalf("%s: Node %+v; lowered %+v", diff, want, got)
			}
		})
	}
}

func TestRecordOperationMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range []struct{ name, fixture, find, replace, helper string }{
		{"read", "operations", "adamic_record_get(", "record_mutant_get(", `static adamic_value *record_mutant_get(const adamic_record *r,const adamic_string *k) { static adamic_value copy; adamic_value *slot=adamic_record_get_own(r,k); if(slot==NULL)return NULL; copy=*slot;copy.number+=1;return &copy;}`},
		{"write", "operations", "adamic_record_set(", "record_mutant_set(", `static void record_mutant_set(adamic_record *r,adamic_string *k,adamic_value v) {v.number+=1;adamic_record_set(r,k,v);}`},
		{"delete", "operations", "adamic_record_delete(", "record_mutant_delete(", `static bool record_mutant_delete(adamic_record *r,const adamic_string *k) {(void)r;(void)k;return true;}`},
		{"in", "operations", "adamic_record_has(", "record_mutant_has(", `static bool record_mutant_has(const adamic_record *r,const adamic_string *k) {return !adamic_record_has(r,k);}`},
		{"has_own", "operations", "adamic_record_has_own(", "record_mutant_has_own(", `static bool record_mutant_has_own(const adamic_record *r,const adamic_string *k) {return !adamic_record_has_own(r,k);}`},
		{"keys", "operations", "adamic_record_keys(", "record_mutant_keys(", `static adamic_array *record_mutant_keys(const adamic_record *r) {adamic_array *a=adamic_record_keys(r);if(a->length>1){adamic_value first=a->elements[0];a->elements[0]=a->elements[1];a->elements[1]=first;}return a;}`},
		{"values", "operations", "adamic_record_iterator_next(", "record_mutant_next(", `static bool record_mutant_next(adamic_record_iterator *it,adamic_string **k,adamic_value *v) {bool more=adamic_record_iterator_next(it,k,v);if(more)v->number+=1;return more;}`},
		{"entries", "operations", "adamic_record_iterator_next(", "record_mutant_next(", `static bool record_mutant_next(adamic_record_iterator *it,adamic_string **k,adamic_value *v) {bool more=adamic_record_iterator_next(it,k,v);if(more)*k=&adamic_string_empty;return more;}`},
		{"spread", "operations", "adamic_record_define(", "record_mutant_define(", `static void record_mutant_define(adamic_record *r,adamic_string *k,adamic_value v) {v.number+=1;adamic_record_define(r,k,v);}`},
		{"for_in", "for_in", "adamic_record_keys(", "record_mutant_keys(", `static adamic_array *record_mutant_keys(const adamic_record *r) {(void)r;return adamic_array_new(0,true);}`},
		{"stringify", "operations", "adamic_json_record", "adamic_json_map", ""},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_"+mutant.fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			p, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			code := native.C(p)
			changed := strings.ReplaceAll(code, mutant.find, mutant.replace)
			if code == changed {
				t.Fatal("mutant changed no code")
			}
			if mutant.helper != "" {
				changed = insertCollectionMutant(changed, mutant.helper)
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(changed, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if got.exitCode != 0 || len(got.stderr) > 0 {
				t.Fatalf("mutant must run cleanly: %+v", got)
			}
			if diff := disagreement(onNode(t, path), got); diff != "stdout differs" {
				t.Fatalf("mutant survived: %q", diff)
			}
			t.Log("caught by Node stdout comparison, clean exit and sanitizers")
		})
	}
}

func TestRecordPrototypeMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_prototype_read.a"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := javascript.JavaScript(p)
	// This is JavaScript's actual inherited function, rather than an invented sentinel.
	start := strings.Index(code, "const adamicRecordGet =")
	end := strings.Index(code[start:], "const adamicRecordHas =") + start
	mutated := code[:start] + "const adamicRecordGet = (record,key) => record[key];\n" + code[end:]
	file := filepath.Join(t.TempDir(), "mutant.mjs")
	if err := os.WriteFile(file, []byte(mutated), 0644); err != nil {
		t.Fatal(err)
	}
	got, want := onNode(t, file), onJavaScriptBackend(t, p)
	if got.exitCode != 0 || want.exitCode != 70 || disagreement(want, got) == "" {
		t.Fatalf("prototype mutant survived: mutant %+v; checked %+v", got, want)
	}
	t.Log("prototype read-through caught by checked exit 70; mutant runs as Node")
}

func TestRecordOwnershipMutant(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("this mutant uses LeakSanitizer")
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_ownership.a"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(p)
	// Keeping the iterator forever preserves all output, but leaks its record and keys.
	changed := strings.ReplaceAll(code, "adamic_release(", "(void)(")
	if changed == code {
		t.Fatal("mutant changed no releases")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(changed, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if diff := disagreement(onNode(t, path), got); diff != "" {
		t.Fatalf("mutant should preserve output: %s", diff)
	}
	report := leakSanitizer(t, binary)
	if !strings.Contains(report, "LeakSanitizer") {
		t.Fatalf("ownership mutant survived: %s", report)
	}
	t.Log("lost releases caught by LeakSanitizer alone")
}

func TestRecordCoalesceMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_census.a"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := javascript.JavaScript(p)
	changed := strings.ReplaceAll(code, "adamicRecordGet(r,k) ?? adamicRecordSet(r,k,make())", "adamicRecordSet(r,k,make())")
	if code == changed {
		t.Fatal("mutant changed no coalescing write")
	}
	file := filepath.Join(t.TempDir(), "mutant.mjs")
	if err := os.WriteFile(file, []byte(changed), 0644); err != nil {
		t.Fatal(err)
	}
	got := onNode(t, file)
	if got.exitCode != 0 || disagreement(onNode(t, path), got) != "stdout differs" {
		t.Fatalf("eager fallback mutant survived: %+v", got)
	}
	t.Log("eager/repeated ??= fallback caught by Node stdout comparison")
}

func TestRecordNarrowingMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_narrowed_number.a"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := javascript.JavaScript(p)
	changed := strings.ReplaceAll(code, "value === undefined ? panic(message) : value", "value")
	if code == changed {
		t.Fatal("mutant changed no narrowed read check")
	}
	file := filepath.Join(t.TempDir(), "mutant.mjs")
	if err := os.WriteFile(file, []byte(changed), 0644); err != nil {
		t.Fatal(err)
	}
	got, want := onNode(t, file), onJavaScriptBackend(t, p)
	if got.exitCode != 0 || want.exitCode != 70 || disagreement(want, got) == "" {
		t.Fatalf("stale narrowing mutant survived: mutant %+v; checked %+v", got, want)
	}
	t.Log("unchecked deleted scalar read caught by checked exit 70")
}
