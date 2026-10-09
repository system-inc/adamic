package oracle

import (
	"context"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewV2ReadsAfterWrites(t *testing.T) {
	for _, name := range []string{"mixed-read-write", "null-read-write", "undefined-read-write", "representation-read-write", "fixed-tuple-good"} {
		t.Run(name, func(t *testing.T) {
			program, path := interfaceFixture(t, "v2/"+name)
			node := onNode(t, path)
			sanitized, binary := nativelyUncached(t, program)
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "native-sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s: %s; %#v", backend, difference, got)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}

func TestCheckedViewV2ArrayArmBoundary(t *testing.T) {
	path, _ := filepath.Abs("../../stage3/interface-downcasts/v2/array-arm.a")
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), loaded)
	if unsupported, ok := err.(*lower.NotYet); !ok || !strings.Contains(unsupported.What, "views-v3: array element kind") {
		t.Fatalf("expected named V3 arm, got %v", err)
	}
	if got := onNode(t, path); got.exitCode != 0 {
		t.Fatalf("Node: %#v", got)
	}
}

func TestCheckedViewV2WrongFamilyMutant(t *testing.T) {
	program, path := interfaceFixture(t, "untagged/fixtures/callable-union-good-number")
	want := onNode(t, path)
	for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatal(difference)
		}
	}
	wrong := ir.ViewContractID(0)
	for i, c := range program.ViewContracts {
		if c.Kind == ir.ViewObject && c.Name == "Receiver" {
			wrong = ir.ViewContractID(i + 1)
		}
	}
	if wrong == 0 {
		t.Fatal("missing wrong-family object descriptor")
	}
	changed := 0
	for i := range program.ViewContracts {
		if program.ViewContracts[i].Kind == ir.ViewUnion && program.ViewContracts[i].Of == ir.Closure {
			program.ViewContracts[i].Of = ir.Object
			program.ViewContracts[i].Members = []ir.ViewContractID{wrong}
			changed++
		}
	}
	if changed != 1 {
		t.Fatalf("wrong-family mutation count %d", changed)
	}
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "native-sanitized": func() run { got, _ := nativelyUncached(t, program); return got }(), "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || disagreement(want, got) == "" {
			t.Fatalf("wrong adapter survived %s: %#v", backend, got)
		}
		t.Logf("wrong family caught by %s: exit %d stderr %q", backend, got.exitCode, got.stderr)
	}
}

func TestCheckedViewV2RepresentationMutants(t *testing.T) {
	// The backend differential above checks the same five slot families after writes.
	// Here the native reader deliberately restores each old interpretation, one at a time.
	for _, sample := range []struct {
		name  string
		tag   int
		kind  string
		value string
	}{
		{"null", 12, "null", "NULL"}, {"undefined", 13, "undefined", "NULL"},
		{"uint8", 15, "uint8", "adamic_typed_array_new(adamic_typed_array_uint8,1)"},
		{"int32", 16, "int32", "adamic_typed_array_new(adamic_typed_array_int32,1)"},
		{"float64", 17, "float64", "adamic_typed_array_new(adamic_typed_array_float64,1)"},
	} {
		t.Run(sample.name, func(t *testing.T) {
			code := representationProbeC(sample.tag, sample.value, false)
			binary := filepath.Join(t.TempDir(), "correct")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte(sample.kind + "\n")}
			if difference := disagreement(want, execute(t, binary)); difference != "" {
				t.Fatal(difference)
			}
			binary = filepath.Join(t.TempDir(), "old-layout")
			if err := native.Build(representationProbeC(sample.tag, sample.value, true), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			got := execute(t, binary)
			if disagreement(want, got) == "" {
				t.Fatal("old-layout read survived")
			}
			if got.exitCode == 0 && len(got.stderr) != 0 {
				t.Fatalf("unexpected mutant result: %#v", got)
			}
			t.Logf("old-layout %s caught: exit %d stdout %q stderr %q", sample.name, got.exitCode, got.stdout, got.stderr)
		})
	}
}

func representationProbeC(tag int, value string, old bool) string {
	return fmt.Sprintf(`
#include "view_unions_mixed.h"
#include <stdio.h>
static const char *const names[]={"value"};
static const bool references[]={true};
static const adamic_shape shape={1,names,references,NULL};
int main(void){
 adamic_object *object=adamic_object_new(&shape);
 object->slots[0].reference=%s;
 adamic_object_initialized(object)[0]=1;
 adamic_object_field_types(object)[0]=%d;
 unsigned char storage=adamic_object_field_types(object)[0];
 if(storage>=15 && storage<=17) adamic_typed_array_set(object->slots[0].reference,0,41);
 adamic_slot_cache cache={0};
 adamic_value *slot=adamic_object_field(object,"value",&cache);
 if(%t){
  if(storage==12 || storage==13){adamic_maybe_number result=adamic_typed_array_get(slot->reference,0);printf("old-typed:%%g\n",result.number);}
  else {static const char text[]="old layout has no slot representation";adamic_panic(text,sizeof text-1);}
 } else {
  if(storage>=15 && storage<=17){adamic_maybe_number result=adamic_typed_array_get(slot->reference,0);if(!result.present || result.number!=41)return 3;puts(storage==15?"uint8":storage==16?"int32":"float64");}
  else {adamic_view_union_value snapshot=adamic_object_view_union_snapshot(object,"value",&cache,"object.value","null | undefined",false);puts(adamic_view_union_kind_name(snapshot.kind));}
 }
 adamic_release(object);
 return 0;
}
`, value, tag, old)
}

func TestCheckedViewV2MembershipMutant(t *testing.T) {
	program, path := interfaceFixture(t, "untagged/fixtures/structural-source-wrong")
	node := onNode(t, path)
	if node.exitCode != 0 {
		t.Fatalf("Node: %#v", node)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.value matches no member of Target; expected Target, found object\n")}
	for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatal(difference)
		}
	}
	changed := 0
	for i, c := range program.ViewContracts {
		if c.Kind == ir.ViewObject && (c.Name == "Left" || c.Name == "Right") {
			program.ViewContracts[i].Fields = nil
			changed++
		}
	}
	if changed != 2 {
		t.Fatalf("membership mutation count %d", changed)
	}
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "native-sanitized": func() run { got, _ := nativelyUncached(t, program); return got }(), "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || disagreement(want, got) == "" {
			t.Fatalf("membership check mutant survived %s: %#v", backend, got)
		}
		t.Logf("membership check removal caught by %s: exit %d stdout %q", backend, got.exitCode, got.stdout)
	}
}

func TestCheckedViewV2TupleIdentityMutant(t *testing.T) {
	program, path := interfaceFixture(t, "v2/tuple-identity-wrong")
	node := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("first\n")}, node); difference != "" {
		t.Fatal(difference)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: viewed.value matches no member of readonly [string, number]; expected readonly [string, number], found object\n")}
	for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatal(difference)
		}
	}
	changed := 0
	for i := range program.ViewContracts {
		if program.ViewContracts[i].FixedTuple {
			program.ViewContracts[i].FixedTuple = false
			changed++
		}
	}
	if changed != 1 {
		t.Fatalf("tuple mutation count %d", changed)
	}
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "native-sanitized": func() run { got, _ := nativelyUncached(t, program); return got }(), "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(node, got); difference != "" {
			t.Fatalf("tuple mutant must execute the Node counterfactual in %s: %s", backend, difference)
		}
		t.Logf("tuple flag removal caught by %s: exit %d stdout %q", backend, got.exitCode, got.stdout)
	}
}

func TestCheckedViewV2CallableProducerMutant(t *testing.T) {
	program, path := interfaceFixture(t, "v2/callable-producer-wrong")
	node := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("true\n")}, node); difference != "" {
		t.Fatal(difference)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.value expected Target, found function with unknown signature\n")}
	for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatal(difference)
		}
	}
	changed := 0
	for i := range program.ViewContracts {
		if program.ViewContracts[i].Kind == ir.ViewCallable && program.ViewContracts[i].ProducerCertified {
			program.ViewContracts[i].ProducerCertified = false
			changed++
		}
	}
	if changed != 2 {
		t.Fatalf("producer mutation count %d", changed)
	}
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "native-sanitized": func() run { got, _ := nativelyUncached(t, program); return got }(), "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(node, got); difference != "" {
			t.Fatalf("producer mutant must execute Node in %s: %s", backend, difference)
		}
		t.Logf("producer certificate removal caught by %s: exit %d stdout %q", backend, got.exitCode, got.stdout)
	}
}

func TestCheckedViewV2MovedStage3Results(t *testing.T) {
	for _, name := range []string{"taste/refused/21_truthy_loops.a", "predicates/03_void_zero.a"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs("../../stage3/fixtures/" + name)
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.Lower(context.Background(), loaded)
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			sanitized, binary := nativelyUncached(t, program)
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "native-sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s: %s; %#v", backend, difference, got)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}
