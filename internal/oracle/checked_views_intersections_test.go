package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// Component only: source admission and shared propagation are not wired yet.
func TestCheckedViewIntersectionConjunction(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "intersections")
	if err := native.Build(intersectionProbeC, binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	for _, variant := range []string{"good", "wrong", "nested", "absent"} {
		t.Run(variant, func(t *testing.T) {
			path, err := filepath.Abs("../../stage3/interface-downcasts/lane7/" + variant + ".a")
			if err != nil {
				t.Fatal(err)
			}
			source := run{stdout: []byte("name:42\n")}
			want := source
			if variant == "wrong" {
				source.stdout = []byte("name:true\n")
			}
			if variant == "nested" {
				source.stdout = []byte("name:bad\n")
			}
			if variant == "wrong" || variant == "nested" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.value does not satisfy every member; expected Named & Counted, found object\n")}
			}
			if difference := disagreement(source, onNode(t, path)); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			jsPath := filepath.Join(t.TempDir(), "probe.mjs")
			code := "import {panic} from 'adamic';\n" + javascript.IntersectionRuntime() + intersectionProbeJS + "\nprobe(" + strconv.Quote(variant) + ");\n"
			if err := os.WriteFile(jsPath, []byte(code), 0600); err != nil {
				t.Fatal(err)
			}
			for _, got := range []run{execute(t, binary, variant), onNode(t, jsPath)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
		})
	}
}

// Mutations alter executable helper/adapter code, keeping valid release builds.
// The nested mutant holds the adapter seam, not frontend propagation.
func TestCheckedViewIntersectionComponentMutants(t *testing.T) {
	for _, mutation := range []struct{ name, variant, oldC, newC, oldJS, newJS string }{
		{"skip check", "wrong", "size_t members[]={1,2};", "size_t members[]={1};", "[1,2]", "[1]"},
		{"accept wrong shape", "wrong", "p->count_kind==adamic_view_union_number", "p->count_kind!=adamic_view_union_unknown", "typeof snapshot.value.child.count === 'number'", "snapshot.value.child.count !== undefined"},
		{"drop nested adapter", "nested", "p->count_kind==adamic_view_union_number", "true", "typeof snapshot.value.child.count === 'number'", "true"},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			if strings.Count(intersectionProbeC, mutation.oldC) != 1 || strings.Count(intersectionProbeJS, mutation.oldJS) != 1 {
				t.Fatal("mutation anchor changed")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(strings.Replace(intersectionProbeC, mutation.oldC, mutation.newC, 1), binary, native.Options{}); err != nil {
				t.Fatal(err)
			}
			jsPath := filepath.Join(t.TempDir(), "mutant.mjs")
			code := "import {panic} from 'adamic';\n" + javascript.IntersectionRuntime() + strings.Replace(intersectionProbeJS, mutation.oldJS, mutation.newJS, 1) + "\nprobe(" + strconv.Quote(mutation.variant) + ");\n"
			if err := os.WriteFile(jsPath, []byte(code), 0600); err != nil {
				t.Fatal(err)
			}
			for _, got := range []run{execute(t, binary, mutation.variant), onNode(t, jsPath)} {
				if got.exitCode != 0 || len(got.stdout) == 0 {
					t.Fatalf("mutant must execute valid release code: %#v", got)
				}
				t.Logf("component refusal pin catches mutant: exit %d, stdout %q", got.exitCode, got.stdout)
			}
		})
	}
}

const intersectionProbeJS = `
const probe = variant => {
 const value = {text:'name', child:{count:variant==='wrong'?true:variant==='nested'?'bad':42}};
 if(variant !== 'absent') value.label='ok';
 const match = (member,snapshot) => member===1 ? typeof snapshot.value.text==='string' : typeof snapshot.value.child.count === 'number';
 adamicViewIntersectionRequire({kind:'object',value},[1,2],match,'view.value','Named & Counted');
 console.log(value.text+':'+String(value.child.count));
};
`

const intersectionProbeC = `
#include "view_intersections.h"
#include <stdio.h>
#include <string.h>
typedef struct probe { adamic_view_union_kind count_kind; } probe;
static bool match(void *context,size_t member,const adamic_view_union_value *value) {
 (void)context;
 if(value->kind!=adamic_view_union_object || value->payload.reference==NULL)return false;
 const probe *p=value->payload.reference;
 return member==1 || (member==2 && p->count_kind==adamic_view_union_number);
}
int main(int argc,char **argv) {
 if(argc!=2)return 2;
 bool wrong=strcmp(argv[1],"wrong")==0;
 bool nested=strcmp(argv[1],"nested")==0;
 probe p={wrong?adamic_view_union_boolean:nested?adamic_view_union_string:adamic_view_union_number};
 adamic_view_union_value value={adamic_view_union_object,{.reference=&p}};
 size_t members[]={1,2};
 adamic_view_intersection_require(&value,members,sizeof members/sizeof members[0],match,NULL,"view.value","Named & Counted");
 printf("name:%s\n",wrong?"true":nested?"bad":"42");
 return 0;
}
`

// Enable with the lane-owned Go overlay until the integrator applies shared-hooks.patch.
func TestCheckedViewIntersectionSource(t *testing.T) {
	if os.Getenv("ADAMIC_INTERSECTION_HOOKS") != "1" {
		t.Skip("shared intersection hooks await integrator; run lane7 overlay")
	}
	for _, probe := range []struct{ name, node, diagnostic string }{
		{"emit-good", "1:42:p\n", ""},
		{"emit-absent", "1:42:none\n", ""},
		{"emit-wrong", "true:42:p\n", "field read failed: node.emitNode.flags is not a number; expected number, found boolean"},
		{"emit-nested", "1:true:p\n", "field read failed: node.emitNode.autoGenerate.id is not a number; expected number, found boolean"},
		{"good", "name:42\n", ""}, {"absent", "name:42\n", ""},
		{"wrong", "name:true\n", "field read failed: view.value.child.count is not a number; expected number, found boolean"},
		{"nested", "name:bad\n", "field read failed: view.value.child.count is not a number; expected number, found string"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane7/"+probe.name)
			kind := os.Getenv("ADAMIC_INTERSECTION_SOURCE_MUTANT")
			if (probe.name == "emit-wrong" && (kind == "skip" || kind == "shape")) || (probe.name == "wrong" && kind == "nested") {
				changed := 0
				rewrite := func(expression ir.Expression) ir.Expression {
					property, ok := expression.(ir.Property)
					if !ok || property.View == "" {
						return expression
					}
					if kind == "skip" || kind == "nested" && property.Name == "count" {
						property.View = ""
						changed++
						return property
					}
					if kind == "shape" && property.Name == "flags" {
						property.Of = ir.Boolean
						property.ViewType = "boolean"
						changed++
						return property
					}
					return expression
				}
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), rewrite)
				mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), rewrite)
				if changed == 0 {
					t.Fatal("requested source mutant found no checked read")
				}
			}

			if difference := disagreement(run{stdout: []byte(probe.node)}, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			want := run{stdout: []byte(probe.node)}
			if probe.diagnostic != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: " + probe.diagnostic + "\n")}
			}
			for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s; got %#v", difference, got)
				}
			}
			if kind == "" && probe.name == "emit-wrong" {
				for _, kind := range []string{"skip check", "accept wrong shape"} {
					mutated, _ := interfaceFixture(t, "lane7/emit-wrong")
					changed := 0
					rewrite := func(expression ir.Expression) ir.Expression {
						property, ok := expression.(ir.Property)
						if !ok || property.View == "" {
							return expression
						}
						if kind == "skip check" {
							property.View = ""
							changed++
							return property
						}
						if property.Name == "flags" {
							property.Of = ir.Boolean
							property.ViewType = "boolean"
							changed++
							return property
						}
						return expression
					}
					mutateStringExpressions(reflect.ValueOf(&mutated.Main).Elem(), rewrite)
					mutateStringExpressions(reflect.ValueOf(&mutated.Functions).Elem(), rewrite)
					if changed == 0 {
						t.Fatal("source mutant found no checked read")
					}
					for _, got := range []run{releasedUncached(t, mutated), onJavaScriptBackend(t, mutated)} {
						if got.exitCode != 0 {
							t.Fatalf("%s must execute valid release code: %#v", kind, got)
						}
						if disagreement(want, got) == "" {
							t.Fatalf("%s mutant survived pinned refusal", kind)
						}
						t.Logf("%s source mutant caught: exit=%d stdout=%q", kind, got.exitCode, got.stdout)
					}
				}
			}
			if kind == "" && probe.name == "wrong" {
				changed := 0
				mutate := func(expression ir.Expression) ir.Expression {
					if property, ok := expression.(ir.Property); ok && property.Name == "count" && property.View != "" {
						property.View = ""
						changed++
						return property
					}
					return expression
				}
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
				mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
				if changed == 0 {
					t.Fatal("nested source mutant found no checked read")
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if disagreement(want, got) == "" {
						t.Fatal("drop transitive source check mutant survived")
					}
					t.Logf("transitive source mutant caught: exit=%d stdout=%q", got.exitCode, got.stdout)
				}
			}
		})
	}
}

// This red probe prevents treating flattened field metadata as runtime conjunction.
// Run explicitly when reviewing the shared hooks; it remains a known blocker.
func TestCheckedViewIntersectionRootConjunctionProbe(t *testing.T) {
	if os.Getenv("ADAMIC_INTERSECTION_CONJUNCTION_PROBE") != "1" {
		t.Skip("explicit negative probe for unfinished all-members runtime dispatch")
	}
	program, path := interfaceFixture(t, "lane7/root-only-wrong")
	if difference := disagreement(run{stdout: []byte("true\n")}, onNode(t, path)); difference != "" {
		t.Fatal(difference)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.value.child.count is not a number; expected number, found boolean\n")}
	for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s; got %#v", difference, got)
		}
	}
}
