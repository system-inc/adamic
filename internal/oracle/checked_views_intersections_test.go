package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
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

// These compile the source through production dispatch, without an overlay.
func TestCheckedViewIntersectionSource(t *testing.T) {
	for _, probe := range []struct{ name, node, diagnostic string }{
		{"emit-good", "1:42:p\n", ""}, {"emit-absent", "1:42:none\n", ""},
		{"emit-wrong", "true:42:p\n", "field read failed: node.emitNode.flags is not a number; expected number, found boolean"},
		{"emit-nested", "1:true:p\n", "field read failed: node.emitNode.autoGenerate.id is not a number; expected number, found boolean"},
		{"emit-root-wrong", "true\n", "field read failed: generated.emitNode.flags is not a number; expected number, found boolean"},
		{"emit-root-nested", "true\n", "field read failed: generated.emitNode.autoGenerate.id is not a number; expected number, found boolean"},
		{"brand-good", "name:42\n", ""},
		{"brand-wrong", "true:42\n", "field read failed: view.value.text is not a string; expected string, found boolean"},
		{"duplicate-good", "name:42:1\n", ""},
		{"duplicate-absent", "name:42:0\n", ""},
		{"duplicate-wrong", "name:true:1\n", "field read failed: view.value.child.count is not a number; expected number, found boolean"},
		{"duplicate-literal", "name:42:2\n", "field read failed: view.value.maybe expected 1 | undefined, found number 2"},
		{"good", "name:42\n", ""}, {"absent", "name:42\n", ""},
		{"wrong", "name:true\n", "field read failed: view.value.child.count is not a number; expected number, found boolean"},
		{"nested", "name:bad\n", "field read failed: view.value.child.count is not a number; expected number, found string"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane7/"+probe.name)
			if difference := disagreement(run{stdout: []byte(probe.node)}, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			kind := os.Getenv("ADAMIC_INTERSECTION_SOURCE_MUTANT")
			if probe.name == "emit-root-wrong" && (kind == "skip" || kind == "shape") || probe.name == "emit-root-nested" && kind == "nested" {
				intersectionSourceMutant(t, program, kind)
			}
			want := run{stdout: []byte(probe.node)}
			if probe.diagnostic != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: " + probe.diagnostic + "\n")}
			}
			actual, binary := nativelyUncached(t, program)
			if want.exitCode == 0 {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s; got %#v", difference, got)
				}
			}
		})
	}
}

// Each mutation changes the production contract consumed by both emitters.
// Root-only fixtures prevent a redundant later field check from hiding the gap.
func intersectionSourceMutant(t *testing.T, program *ir.Program, kind string) {
	t.Helper()
	changed := 0
	for i := range program.ViewContracts {
		contract := &program.ViewContracts[i]
		if kind == "skip" && contract.Intersection {
			contract.Intersection = false
			changed++
		}
		if kind == "shape" && contract.Intersection {
			for j, field := range contract.Fields {
				if field.Name == "flags" {
					child := program.ViewContracts[field.Contract-1]
					child.Of = ir.Boolean
					child.Name = "boolean"
					contract.Fields[j].Contract = ir.ViewContractID(len(program.ViewContracts) + 1)
					program.ViewContracts = append(program.ViewContracts, child)
					changed++
				}
			}
		}
		if kind == "nested" {
			for j, field := range contract.Fields {
				if field.Name == "id" {
					contract.Fields = append(contract.Fields[:j:j], contract.Fields[j+1:]...)
					changed++
					break
				}
			}
		}
	}
	if changed == 0 {
		t.Fatal("source mutant found no production contract")
	}
}

func TestCheckedViewIntersectionProductionMutants(t *testing.T) {
	for _, test := range []struct{ kind, fixture string }{{"skip", "emit-root-wrong"}, {"shape", "emit-root-wrong"}, {"nested", "emit-root-nested"}} {
		t.Run(test.kind, func(t *testing.T) {
			program, _ := interfaceFixture(t, "lane7/"+test.fixture)
			intersectionSourceMutant(t, program, test.kind)
			for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if got.exitCode != 0 || string(got.stdout) != "true\n" {
					t.Fatalf("mutant must run valid release code: %#v", got)
				}
				t.Logf("production refusal pin catches %s mutant: exit=0 stdout=%q", test.kind, got.stdout)
			}
		})
	}
}

// A root-only read must check the conjunction before any descendant read.
func TestCheckedViewIntersectionRootConjunctionProbe(t *testing.T) {
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

// Lazy admission must stay intact while compound runtime selection is unwired.
func TestCheckedViewIntersectionCompoundDemand(t *testing.T) {
	program, path := interfaceFixture(t, "lane7/compound-unread")
	want := run{stdout: []byte("ok\n")}
	for _, got := range []run{onNode(t, path), releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatal(difference)
		}
	}
	readPath, pathErr := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/lane7/compound-untagged-read.a"))
	if pathErr != nil {
		t.Fatal(pathErr)
	}
	_, err := lowered(t, readPath)
	if err == nil || !strings.Contains(err.Error(), "field value with unsupported union intersection") {
		t.Fatalf("compound demand must refuse, got %v", err)
	}
	// The runtime control is valid JavaScript; the compiler's refusal prevents a
	// silent success until selection validates the matching intersection arm.
	if got := onNode(t, readPath); got.exitCode != 0 || string(got.stdout) != "true\n" {
		t.Fatalf("Node: %#v", got)
	}
}

func TestCheckedViewIntersectionRecursiveDemand(t *testing.T) {
	program, path := interfaceFixture(t, "lane7/recursive-unread")
	want := run{stdout: []byte("ok\n")}
	for _, got := range []run{onNode(t, path), releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatal(difference)
		}
	}
	for _, test := range []struct{ name, diagnostic string }{
		{"recursive-good", ""}, {"recursive-absent", ""}, {"recursive-optional-absent", ""}, {"recursive-null-root-absent", ""},
		{"recursive-null-root-wrong", "field read failed: view.value.next.count is not a number; expected number, found boolean"},
		{"recursive-optional-root", "field read failed: view.value.next.count is not a number; expected number, found boolean"}, {"recursive-literal-good", ""},
		{"recursive-literal-wrong", "field read failed: view.value.next.mode expected \"a\\0b\" | \"c\", found string a"},
		{"recursive-number-literal", "field read failed: view.value.next.code expected 1 | undefined, found number 2"},
		{"recursive-boolean-literal", "field read failed: view.value.next.enabled expected true | undefined, found boolean false"},
		{"recursive-read", "field read failed: view.value.next.count is not a number; expected number, found boolean"},
		{"recursive-deep-wrong", "field read failed: view.value.next.next.next.count is not a number; expected number, found boolean"},
		{"recursive-missing", "field read failed: view.value.next.count is not initialized; expected number, found missing"},
		{"recursive-object-wrong", "field read failed: view.value.next is not a Link | undefined; expected Link | undefined, found boolean"},
	} {
		t.Run(test.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane7/"+test.name)
			if got := onNode(t, path); got.exitCode != 0 || string(got.stdout) != "true\n" {
				t.Fatalf("Node: %#v", got)
			}
			if kind := os.Getenv("ADAMIC_INTERSECTION_RECURSIVE_MUTANT"); kind != "" && (test.name == "recursive-read" || kind == "optional" && test.name == "recursive-optional-root" || kind == "literal-mode" && test.name == "recursive-literal-wrong" || kind == "literal-code" && test.name == "recursive-number-literal" || kind == "literal-enabled" && test.name == "recursive-boolean-literal") {
				intersectionRecursiveMutant(t, program, kind)
			}
			want := run{stdout: []byte("true\n")}
			if test.diagnostic != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: " + test.diagnostic + "\n")}
			}
			actual, binary := nativelyUncached(t, program)
			if want.exitCode == 0 {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s; got %#v", difference, got)
				}
			}
		})
	}

}

func TestCheckedViewIntersectionSelectedArms(t *testing.T) {
	for _, test := range []struct{ name, diagnostic string }{
		{"leading-access-identifier", ""}, {"leading-access-element", ""}, {"leading-access-property", ""},
		{"leading-access-wrong", "field read failed: node.expression.argumentExpression.text is not a string; expected string, found boolean"},
		{"leading-access-unknown-tag", "field read failed: node.expression.kind expected SyntaxKind, found number 999"},
		{"leading-access-missing", "field read failed: node.expression.escapedText is not initialized; expected string, found missing"},
		{"compound-read", "field read failed: view.value.common is not a number; expected number, found boolean"},
	} {
		t.Run(test.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane7/"+test.name)
			if difference := disagreement(run{stdout: []byte("true\n")}, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			if kind := os.Getenv("ADAMIC_INTERSECTION_ARM_MUTANT"); kind != "" && (test.name == "leading-access-wrong" || kind == "tag" && test.name == "leading-access-unknown-tag") {
				intersectionArmMutant(t, program, kind)
			}
			want := run{stdout: []byte("true\n")}
			if test.diagnostic != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: " + test.diagnostic + "\n")}
			}
			actual, binary := nativelyUncached(t, program)
			if want.exitCode == 0 {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s; got %#v", difference, got)
				}
			}
		})
	}
}

// Root-only demand makes each deleted obligation observable in every backend.
func intersectionArmMutant(t *testing.T, program *ir.Program, kind string) {
	t.Helper()
	changed := false
	for i := range program.ViewContracts {
		c := &program.ViewContracts[i]
		if kind == "tag" && c.IntersectionTag != "" {
			arm := &program.ViewContracts[c.Members[0]-1]
			for j, f := range arm.Fields {
				if f.Name == c.IntersectionTag {
					child := program.ViewContracts[f.Contract-1]
					child.Allowed = append(append([]ir.ViewLiteral(nil), child.Allowed...), ir.ViewLiteral{Of: ir.Number, Number: 999})
					arm.Fields[j].Contract = ir.ViewContractID(len(program.ViewContracts) + 1)
					program.ViewContracts = append(program.ViewContracts, child)
					return
				}
			}
		}
		if kind == "skip" && c.IntersectionTag != "" {
			for _, member := range c.Members {
				arm := &program.ViewContracts[member-1]
				fields := []ir.ViewFieldContract{}
				for _, f := range arm.Fields {
					if f.Name == c.IntersectionTag {
						fields = append(fields, f)
					}
				}
				arm.Fields = fields
			}
			changed = true
		}
		if kind == "nested" {
			for _, f := range c.Fields {
				if f.Name == "argumentExpression" {
					program.ViewContracts[f.Contract-1].Fields = nil
					changed = true
				}
			}
		}
		if kind == "shape" {
			for j, f := range c.Fields {
				if f.Name == "text" {
					child := program.ViewContracts[f.Contract-1]
					child.Of = ir.Boolean
					child.Name = "boolean"
					c.Fields[j].Contract = ir.ViewContractID(len(program.ViewContracts) + 1)
					program.ViewContracts = append(program.ViewContracts, child)
					changed = true
				}
			}
		}
	}
	if !changed {
		t.Fatal("arm mutant found no obligation")
	}
}

func intersectionRecursiveMutant(t *testing.T, program *ir.Program, kind string) {
	t.Helper()
	changed := false
	for i := range program.ViewContracts {
		c := &program.ViewContracts[i]
		if kind == "optional" && c.Intersection {
			c.Intersection = false
			changed = true
		}
		if kind == "canonical" && c.ObjectPresent != 0 {
			c.ObjectPresent = 0
			changed = true
		}
		if kind == "skip" && c.IntersectionRecursive {
			c.Fields = nil
			changed = true
		}
		if strings.HasPrefix(kind, "literal-") {
			for _, field := range c.Fields {
				if field.Name == strings.TrimPrefix(kind, "literal-") {
					program.ViewContracts[field.Contract-1].Allowed = nil
					changed = true
				}
			}
		}
		if c.ObjectPresent == 0 {
			continue
		}
		target := &program.ViewContracts[c.ObjectPresent-1]
		if kind == "nested" {
			target.Fields = nil
			changed = true
		}
		if kind == "shape" {
			for j, f := range target.Fields {
				if f.Name == "count" {
					child := program.ViewContracts[f.Contract-1]
					child.Of = ir.Boolean
					child.Name = "boolean"
					target.Fields[j].Contract = ir.ViewContractID(len(program.ViewContracts) + 1)
					program.ViewContracts = append(program.ViewContracts, child)
					changed = true
				}
			}
		}
	}
	if !changed {
		t.Fatal("recursive mutant found no obligation")
	}
}
