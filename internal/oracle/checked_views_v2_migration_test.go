package oracle

import (
	"context"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewV2ReadsAfterWrites(test *testing.T) {
	test.Parallel()
	for _, name := range []string{
		"mixed-read-write",
		"null-read-write",
		"undefined-read-write",
		"representation-read-write",
		"fixed-tuple-good",
		"fs-option-boxing",
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			program, path := interfaceFixture(test, "v2/"+name)
			node := onNode(test, path)
			sanitized, binary := nativelyUncached(test, program)
			for backend, got := range map[string]run{
				"native":           releasedUncached(test, program),
				"native-sanitized": sanitized,
				"javascript":       onJavaScriptBackend(test, program),
			} {
				if difference := disagreement(node, got); difference != "" {
					test.Fatalf("%s: %s; %#v", backend, difference, got)
				}
			}
			if report := leaks(test, program, binary); report != "" {
				test.Fatal(report)
			}
		})
	}
}

// In .a the cast is refused first, since its tag test, kind === 'Holder', is one every Root passes (#q97a1rr); its .ts
// twin keeps the cast a checked view and stops at the V3 arm.
func TestCheckedViewV2ArrayArmBoundary(test *testing.T) {
	test.Parallel()
	path, _ := filepath.Abs("../../stage3/interface-downcasts/v2/array-arm.a")
	_, loweredError := lowered(test, path)
	if refused, ok := loweredError.(*lower.Refused); !ok ||
		!strings.Contains(refused.What, "a downcast to View whose tag test is vacuous") {
		test.Fatalf("expected the vacuous downcast refusal, got %v", loweredError)
	}
	_, loweredError = lowered(test, typeScriptTwin(test, path))
	if unsupported, ok := loweredError.(*lower.NotYet); !ok ||
		!strings.Contains(unsupported.What, "views-v3: array element kind") {
		test.Fatalf("twin: expected named V3 arm, got %v", loweredError)
	}
	if got := onNode(test, path); got.exitCode != 0 {
		test.Fatalf("Node: %#v", got)
	}
}

func TestCheckedViewV2WrongFamilyMutant(test *testing.T) {
	test.Parallel()
	program, path := interfaceFixture(test, "untagged/fixtures/callable-union-good-number")
	want := onNode(test, path)
	for _, got := range []run{releasedUncached(test, program), onJavaScriptBackend(test, program)} {
		if difference := disagreement(want, got); difference != "" {
			test.Fatal(difference)
		}
	}
	wrong := ir.ViewContractID(0)
	for i, c := range program.ViewContracts {
		if c.Kind == ir.ViewObject && c.Name == "Receiver" {
			wrong = ir.ViewContractID(i + 1)
		}
	}
	if wrong == 0 {
		test.Fatal("missing wrong-family object descriptor")
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
		test.Fatalf("wrong-family mutation count %d", changed)
	}
	for backend, got := range map[string]run{
		"native":           releasedUncached(test, program),
		"native-sanitized": func() run { got, _ := nativelyUncached(test, program); return got }(),
		"javascript":       onJavaScriptBackend(test, program),
	} {
		if got.exitCode != 70 || disagreement(want, got) == "" {
			test.Fatalf("wrong adapter survived %s: %#v", backend, got)
		}
		test.Logf("wrong family caught by %s: exit %d stderr %q", backend, got.exitCode, got.stderr)
	}
}

func TestCheckedViewV2RepresentationMutants(test *testing.T) {
	test.Parallel()
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
		test.Run(sample.name, func(test *testing.T) {
			test.Parallel()
			code := representationProbeC(sample.tag, sample.value, false)
			binary := filepath.Join(test.TempDir(), "correct")
			if buildError := native.Build(code, binary, native.Options{Sanitize: true}); buildError != nil {
				test.Fatal(buildError)
			}
			want := run{stdout: []byte(sample.kind + "\n")}
			if difference := disagreement(want, execute(test, binary)); difference != "" {
				test.Fatal(difference)
			}
			binary = filepath.Join(test.TempDir(), "old-layout")
			if buildError := native.Build(
				representationProbeC(sample.tag, sample.value, true),
				binary,
				native.Options{Sanitize: true},
			); buildError != nil {
				test.Fatal(buildError)
			}
			got := execute(test, binary)
			if disagreement(want, got) == "" {
				test.Fatal("old-layout read survived")
			}
			if got.exitCode == 0 && len(got.stderr) != 0 {
				test.Fatalf("unexpected mutant result: %#v", got)
			}
			test.Logf(
				"old-layout %s caught: exit %d stdout %q stderr %q",
				sample.name,
				got.exitCode,
				got.stdout,
				got.stderr,
			)
		})
	}
}

func representationProbeC(tag int, value string, old bool) string {
	return fmt.Sprintf(`
#include "view_unions_mixed.h"
#include <stdio.h>
static const char *const names[]={"value"};
static const bool references[]={true};
static const adamic_field_kind shape_kinds[] = {adamic_field_reference};
static const adamic_shape shape={1,names,references,NULL, shape_kinds};
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

func TestCheckedViewV2MembershipMutant(test *testing.T) {
	test.Parallel()
	program, path := interfaceFixture(test, "untagged/fixtures/structural-source-wrong")
	node := onNode(test, path)
	if node.exitCode != 0 {
		test.Fatalf("Node: %#v", node)
	}
	want := run{exitCode: 70, stderr: []byte(
		"adamic: panic: field read failed: view.value matches no member of Target; expected Target, found object\n",
	)}
	for _, got := range []run{releasedUncached(test, program), onJavaScriptBackend(test, program)} {
		if difference := disagreement(want, got); difference != "" {
			test.Fatal(difference)
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
		test.Fatalf("membership mutation count %d", changed)
	}
	for backend, got := range map[string]run{
		"native":           releasedUncached(test, program),
		"native-sanitized": func() run { got, _ := nativelyUncached(test, program); return got }(),
		"javascript":       onJavaScriptBackend(test, program),
	} {
		if got.exitCode != 0 || disagreement(want, got) == "" {
			test.Fatalf("membership check mutant survived %s: %#v", backend, got)
		}
		test.Logf("membership check removal caught by %s: exit %d stdout %q", backend, got.exitCode, got.stdout)
	}
}

func TestCheckedViewV2TupleIdentityMutant(test *testing.T) {
	test.Parallel()
	program, path := interfaceFixture(test, "v2/tuple-identity-wrong")
	node := onNode(test, path)
	if difference := disagreement(run{stdout: []byte("first\n")}, node); difference != "" {
		test.Fatal(difference)
	}
	want := run{exitCode: 70, stderr: []byte(
		"adamic: panic: field read failed: viewed.value matches no member of readonly [string, number]; expected readonly [string, number], found object\n",
	)}
	for _, got := range []run{releasedUncached(test, program), onJavaScriptBackend(test, program)} {
		if difference := disagreement(want, got); difference != "" {
			test.Fatal(difference)
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
		test.Fatalf("tuple mutation count %d", changed)
	}
	for backend, got := range map[string]run{
		"native":           releasedUncached(test, program),
		"native-sanitized": func() run { got, _ := nativelyUncached(test, program); return got }(),
		"javascript":       onJavaScriptBackend(test, program),
	} {
		if difference := disagreement(node, got); difference != "" {
			test.Fatalf("tuple mutant must execute the Node counterfactual in %s: %s", backend, difference)
		}
		test.Logf("tuple flag removal caught by %s: exit %d stdout %q", backend, got.exitCode, got.stdout)
	}
}

func TestCheckedViewV2CallableProducerMutant(test *testing.T) {
	test.Parallel()
	program, path := interfaceFixture(test, "v2/callable-producer-wrong")
	node := onNode(test, path)
	if difference := disagreement(run{stdout: []byte("true\n")}, node); difference != "" {
		test.Fatal(difference)
	}
	// Despite its historical name, the number producer accepts the member's
	// literal 1 parameter. It is assignable and must agree with Node.
	for _, got := range []run{releasedUncached(test, program), onJavaScriptBackend(test, program)} {
		if difference := disagreement(node, got); difference != "" {
			test.Fatal(difference)
		}
	}
	changed := 0
	for i := range program.ViewContracts {
		if program.ViewContracts[i].Kind == ir.ViewCallable && program.ViewContracts[i].ProducerCertified {
			program.ViewContracts[i].Functions = nil
			changed++
		}
	}
	if changed != 2 {
		test.Fatalf("producer mutation count %d", changed)
	}
	want := run{exitCode: 70, stderr: []byte(
		"adamic: panic: field read failed: view.value expected Target, found function with unknown signature\n",
	)}
	for backend, got := range map[string]run{
		"native":           releasedUncached(test, program),
		"native-sanitized": func() run { got, _ := nativelyUncached(test, program); return got }(),
		"javascript":       onJavaScriptBackend(test, program),
	} {
		if difference := disagreement(want, got); difference != "" {
			test.Fatalf("producer certificate removal survived %s: %s", backend, difference)
		}
		test.Logf("producer certificate removal caught by %s: exit %d stdout %q", backend, got.exitCode, got.stdout)
	}
}

func TestCheckedViewV2MovedStage3Results(test *testing.T) {
	test.Parallel()
	// #80cs7rz: predicates/03_void_zero.a is NotYet now, isNumericLiteral reading kind through Node of node.expression, a
	// viewed value inside a viewed one; the per-read check wins it back. Pinned in TestCheckedViewV2MovedStage3NotYet.
	for _, name := range []string{
		"taste/refused/21_truthy_loops.ts",
		"assertions/11_parenthesized_kind.ts",
		"assertions/20_type_flag_mask.ts",
		"assertions/13_flag_downcast.ts",
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs("../../stage3/fixtures/" + name)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			loaded, loadError := load.Load([]string{path})
			if loadError != nil {
				test.Fatal(loadError)
			}
			program, lowerError := lower.Lower(context.Background(), loaded)
			if lowerError != nil {
				test.Fatal(lowerError)
			}
			node := onNode(test, path)
			sanitized, binary := nativelyUncached(test, program)
			for backend, got := range map[string]run{
				"native":           releasedUncached(test, program),
				"native-sanitized": sanitized,
				"javascript":       onJavaScriptBackend(test, program),
			} {
				if difference := disagreement(node, got); difference != "" {
					test.Fatalf("%s: %s; %#v", backend, difference, got)
				}
			}
			if report := leaks(test, program, binary); report != "" {
				test.Fatal(report)
			}
		})
	}
}

// The stage 3 result 03_void_zero.a moved to NotYet with #80cs7rz (its status.json row says so): isVoidZero passes
// node.expression, which a checked read takes out of the viewed VoidExpression, to isNumericLiteral, which reads its
// kind through Node. Node prints the result; stage 0 stops at the read until the per-read check wins it back.
func TestCheckedViewV2MovedStage3NotYet(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("../../stage3/fixtures/predicates/03_void_zero.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	if difference := disagreement(run{stdout: []byte("true\nfalse\nfalse\n")}, onNode(test, path)); difference != "" {
		test.Fatal("Node: " + difference)
	}
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	_, lowerError := lower.Lower(context.Background(), loaded)
	var notYet *lower.NotYet
	if !errors.As(lowerError, &notYet) || !strings.Contains(
		lowerError.Error(),
		"03_void_zero.a:23:12: stage 0 can't lower a read of field kind through type Node, which may hold a checked view's value",
	) {
		test.Fatalf("want the widened read NotYet at 23:12, got %v", lowerError)
	}
}
