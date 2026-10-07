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
