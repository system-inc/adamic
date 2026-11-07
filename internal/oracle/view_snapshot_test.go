package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"boolean", "number"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"stage3/interface-downcasts/v2/snapshot-maybe-" + name + ".ts", true, false})
	}
}

func checkViewSnapshot(test *testing.T, name string) {
	test.Helper()
	program, path := interfaceFixture(test, "v2/snapshot-maybe-"+name)
	node := onNode(test, path)
	test.Logf("Node: exit=%d stdout=%q stderr=%q", node.exitCode, node.stdout, node.stderr)
	sanitized, binary := nativelyUncached(test, program)
	for backend, got := range map[string]run{
		"native":           releasedUncached(test, program),
		"native-sanitized": sanitized,
		"javascript":       onJavaScriptBackend(test, program),
	} {
		test.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(node, got); difference != "" {
			test.Errorf("%s: %s", backend, difference)
		}
	}
	if !test.Failed() {
		if report := leaksUncached(test, program, binary); report != "" {
			test.Fatal(report)
		}
	}
}

func TestViewSnapshotMaybeBoolean(test *testing.T) {
	test.Parallel()
	checkViewSnapshot(test, "boolean")
}

func TestViewSnapshotMaybeNumber(test *testing.T) {
	test.Parallel()
	checkViewSnapshot(test, "number")
}

// Typed arrays retain their current mixed-union refusal while snapshots preserve their reference. The program is .ts,
// where the cast stays a checked view: in .a its tag test every Root passes is refused (#q97a1rr).
func TestViewSnapshotTypedArrayReference(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "typed-array-view.ts")
	source := `interface Root { readonly kind: "Holder" }
interface View extends Root { readonly value: string | number }
const raw = { kind: "Holder" as const, value: new Uint8Array(1) };
const base: Root = raw;
const viewed = base as View;
console.log(typeof viewed.value);
`
	if writeError := os.WriteFile(path, []byte(source), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	code := native.C(program)
	// Behavior alone cannot prove this refusal travels through the snapshot reader.
	if !strings.Contains(code, "adamic_object_view_union_snapshot(") {
		test.Fatal("view read bypassed snapshot")
	}
	node := onNode(test, path)
	if difference := disagreement(run{stdout: []byte("object\n")}, node); difference != "" {
		test.Fatal("Node control: " + difference)
	}
	want := run{exitCode: 70, stderr: []byte(
		"adamic: panic: field read failed: viewed.value matches no member of string | number; expected string | number, found unsupported representation\n",
	)}
	sanitized, _ := nativelyUncached(test, program)
	for backend, got := range map[string]run{"native": releasedUncached(test, program), "native-sanitized": sanitized} {
		test.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(want, got); difference != "" {
			test.Errorf("%s: %s", backend, difference)
		}
	}
	// The refusal above does not observe the payload. A fallback must still receive the actual reference.
	probe := `
#include "view_unions_mixed.h"
#include <stdio.h>
static const char *const names[]={"value"};
static const bool references[]={true};
static const adamic_field_kind kinds[]={adamic_field_reference};
static const adamic_shape shape={1,names,references,NULL,kinds};
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
	binary := filepath.Join(test.TempDir(), "snapshot-reference")
	if buildError := native.Build(probe, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	if difference := disagreement(
		run{stdout: []byte("references preserved\n")},
		execute(test, binary),
	); difference != "" {
		test.Fatal(difference)
	}
}
