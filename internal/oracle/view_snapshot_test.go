package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	additionalFixtureCounts = append(additionalFixtureCounts, func(t *testing.T) []string {
		return []string{counted(t, interfaceFixturePath("v2/snapshot-maybe-boolean"), false, nil, false, false), counted(t, interfaceFixturePath("v2/snapshot-maybe-number"), false, nil, false, false)}
	})
}

func checkViewSnapshot(t *testing.T, name string) {
	t.Helper()
	program, path := interfaceFixture(t, "v2/snapshot-maybe-"+name)
	node := onNode(t, path)
	t.Logf("Node: exit=%d stdout=%q stderr=%q", node.exitCode, node.stdout, node.stderr)
	sanitized, binary := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "native-sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(node, got); difference != "" {
			t.Errorf("%s: %s", backend, difference)
		}
	}
	if !t.Failed() {
		if report := leaksUncached(t, program, binary); report != "" {
			t.Fatal(report)
		}
	}
}

func TestViewSnapshotMaybeBoolean(t *testing.T) {
	t.Parallel()
	checkViewSnapshot(t, "boolean")
}

func TestViewSnapshotMaybeNumber(t *testing.T) {
	t.Parallel()
	checkViewSnapshot(t, "number")
}

// Typed arrays retain their current mixed-union refusal while snapshots preserve their reference.
func TestViewSnapshotTypedArrayReference(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "typed-array-view.a")
	source := `interface Root { readonly kind: "Holder" }
interface View extends Root { readonly value: string | number }
const raw = { kind: "Holder" as const, value: new Uint8Array(1) };
const base: Root = raw;
const viewed = base as View;
console.log(typeof viewed.value);
`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	// Behavior alone cannot prove this refusal travels through the snapshot reader.
	if !strings.Contains(code, "adamic_object_view_union_snapshot(") {
		t.Fatal("view read bypassed snapshot")
	}
	node := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("object\n")}, node); difference != "" {
		t.Fatal("Node control: " + difference)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: viewed.value matches no member of string | number; expected string | number, found unsupported representation\n")}
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "native-sanitized": sanitized} {
		t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s", backend, difference)
		}
	}
	// The refusal above does not observe the payload. A fallback must still receive the actual reference.
	probe := `
#include "view_unions_mixed.h"
#include <stdio.h>
static const char *const names[]={"value"};
static const bool references[]={true};
static const adamic_shape shape={1,names,references,NULL};
int main(void){
 adamic_object *object=adamic_object_new(&shape);
 adamic_typed_array *array=adamic_typed_array_new(adamic_typed_array_uint8,1);
 object->slots[0].reference=array;
 for(unsigned char storage=adamic_rep_record;storage<=adamic_rep_float64_array;storage++){
  adamic_object_field_types(object)[0]=storage;
  adamic_slot_cache cache={0};
  adamic_view_union_value value=adamic_object_view_union_snapshot(object,"value",&cache,"view.value","string | number",false);
  if(value.kind!=adamic_view_union_unknown || value.payload.reference!=array){fprintf(stderr,"lost reference for storage %u\n",storage);adamic_release(object);return 3;}
 }
 adamic_release(object);
 puts("references preserved");
 return 0;
}
`
	binary := filepath.Join(t.TempDir(), "snapshot-reference")
	if err := native.Build(probe, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(run{stdout: []byte("references preserved\n")}, execute(t, binary)); difference != "" {
		t.Fatal(difference)
	}
}
