package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

func v4IdentityProgram(t *testing.T, source, output string) (*ir.Program, run) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "identity.ts")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != output || len(truth.stderr) != 0 {
		t.Fatalf("source Node: %#v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	return program, truth
}

func v4IdentityJavaScript(t *testing.T, code string) run {
	t.Helper()
	path := filepath.Join(t.TempDir(), "identity.mjs")
	if err := os.WriteFile(path, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	return onNode(t, path)
}

func v4IdentityCounted(t *testing.T, code string) (run, [6]int) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "counted")
	if err := native.Build(code, binary, native.Options{Count: true, Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	match := countsLine.FindSubmatch(got.stderr)
	if got.exitCode != 0 || match == nil || strings.Contains(string(got.stderr), "Sanitizer") {
		t.Fatalf("counted fixture must finish sanitizer-clean: %#v", got)
	}
	var counts [6]int
	for i := range counts {
		value, err := strconv.Atoi(string(match[i+1]))
		if err != nil {
			t.Fatal(err)
		}
		counts[i] = value
	}
	if counts[0] != counts[1]+counts[5] {
		t.Fatalf("counted fixture leaked: %v", counts)
	}
	got.stderr = got.stderr[:len(got.stderr)-len(match[0])]
	return got, counts
}

func v4EscapeAdapterIdentity(t *testing.T) {
	t.Helper()
	source := `interface Base { readonly kind:'Receiver'; }
interface Target extends Base { readonly run:(value:string)=>string; }
interface Narrow extends Base { readonly run:(value:string)=>'ok'; }
const producer=(value:string):string=>'ok';
const raw={kind:'Receiver' as const,run:producer};
function probe(base:Base):void {
 const view=base as Target; const other=base as Narrow;
 const first=view.run; const second=view.run; const third=other.run;
 console.log(` + "`${first===second}:${first===third}:${first===producer}:${first!==third}`" + `);
 console.log(` + "`${Object.is(first,second)}:${Object.is(first,third)}:${Object.is(first,producer)}`" + `);
 const map=new Map<(value:string)=>string,string>(); map.set(first,'first'); map.set(third,'updated');
 console.log(` + "`${map.size}:${map.has(second)}:${map.has(producer)}:${map.get(third)}`" + `);
 const set=new Set<(value:string)=>string>(); set.add(first); set.add(third);
 console.log(` + "`${set.size}:${set.has(second)}:${set.has(producer)}`" + `);
 for (const key of map.keys()) { console.log(key('ok')); }
 for (const key of set) { console.log(key('ok')); }
 console.log(` + "`${map.delete(producer)}:${set.delete(third)}:${map.size}:${set.size}`" + `);
}
probe(raw);`
	output := "true:true:true:false\ntrue:true:true\n1:true:true:updated\n1:true:true\nok\nok\ntrue:true:0:0\n"
	program, truth := v4IdentityProgram(t, source, output)
	sanitized, binary := nativelyUncached(t, program)
	for name, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("%s: %s; %#v", name, difference, got)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
	// Omit root normalization in emitted native comparisons. The two distinct
	// view-type adapters now compare unequal, with all checks/calls still present.
	code := native.C(program)
	include := "#include \"adamic.h\"\n"
	if strings.Count(code, include) != 1 {
		t.Fatal("identity mutant include moved")
	}
	code = strings.Replace(code, include, include+"#define adamic_view_adapter_underlying(value) (value)\n", 1)
	got, _ := v4IdentityCounted(t, code)
	if got.exitCode != 0 || !strings.HasPrefix(string(got.stdout), "true:false:false:true\n") || disagreement(truth, got) == "" {
		t.Fatalf("native unequal-read mutant must fail semantically: %#v", got)
	}
	t.Logf("native/sanitized unequal-read mutant caught: %q", got.stdout)
	code = javascript.JavaScript(program)
	before := "const adamicViewAdapterUnderlying = value => adamicViewAdapterRoots.get(value) ?? value;"
	if strings.Count(code, before) != 1 {
		t.Fatal("JavaScript identity mutation site moved")
	}
	got = v4IdentityJavaScript(t, strings.Replace(code, before, "const adamicViewAdapterUnderlying = value => value;", 1))
	if got.exitCode != 0 || !strings.HasPrefix(string(got.stdout), "true:false:false:true\n") || disagreement(truth, got) == "" {
		t.Fatalf("JavaScript unequal-read mutant must fail semantically: %#v", got)
	}
	t.Logf("JavaScript unequal-read mutant caught: %q", got.stdout)
}

func v4ReViewSource(count int) string {
	return v4EscapeDeclarations + fmt.Sprintf(`
const raw={kind:'Receiver' as const,run:(value:string):string=>'ok'};
function read(base:Base):(value:string)=>string { const view=base as Target; return view.run; }
let held=read(raw);
for(let i=0;i<%d;i++) { const next={kind:'Receiver' as const,run:held}; held=read(next); }
console.log(held('ok'));`, count)
}

func v4EscapeAdapterNoStacking(t *testing.T) {
	t.Helper()
	baseline, _ := v4IdentityProgram(t, v4ReViewSource(0), "ok\n")
	program, truth := v4IdentityProgram(t, v4ReViewSource(1000), "ok\n")
	larger, _ := v4IdentityProgram(t, v4ReViewSource(2000), "ok\n")
	_, baseCounts := v4IdentityCounted(t, native.C(baseline))
	got, counts := v4IdentityCounted(t, native.C(program))
	_, largerCounts := v4IdentityCounted(t, native.C(larger))
	if disagreement(truth, got) != "" || counts[0]-baseCounts[0] != 1000 || counts[4] > baseCounts[4]+3 {
		t.Fatalf("bounded allocation exceeded: base=%v loop=%v", baseCounts, counts)
	}
	if largerCounts[0]-counts[0] != 1000 || largerCounts[4] != counts[4] || largerCounts[2]-counts[2] != counts[2]-baseCounts[2] || largerCounts[3]-counts[3] != counts[3]-baseCounts[3] {
		t.Fatalf("re-view retain/allocation counts must stay steady per iteration: %v %v %v", baseCounts, counts, largerCounts)
	}
	t.Logf("bounded native allocation/retain rows: zero=%v 1000=%v 2000=%v", baseCounts, counts, largerCounts)
	if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatal(difference)
	}
	// A clean uncached adapter, with exactly the real adapter metadata and free
	// path. This changes generated calls, never production runtime source.
	code := native.C(program)
	include := "#include \"adamic.h\"\n"
	mutant := `#include <stdlib.h>
static adamic_closure *adamic_unconditional_adapter(adamic_closure *root,const void *key,adamic_closure *(*make)(adamic_closure *,const void *)) {
 root=adamic_view_adapter_underlying(root);
 adamic_closure *fresh=make(root,key);
 fresh->view=malloc(sizeof *fresh->view);
 if(fresh->view==NULL) adamic_panic("out of memory",13);
 *fresh->view=(adamic_view_adapter){adamic_retain(root),key,SIZE_MAX};
 return fresh;
}
#define adamic_view_adapter_intern(root,key,make) adamic_unconditional_adapter(root,key,make)
`
	code = strings.Replace(code, include, include+mutant, 1)
	got, mutantCounts := v4IdentityCounted(t, code)
	if disagreement(truth, got) != "" || mutantCounts[0]-counts[0] != 1000 {
		t.Fatalf("unconditional wrap must only fail allocation bound: got=%#v counts=%v baseline=%v", got, mutantCounts, counts)
	}
	t.Logf("native/sanitized unconditional-wrap caught by bounded allocation: %v vs %v", mutantCounts, counts)
	// Count actual factory executions, without changing calls or their result.
	code = javascript.JavaScript(program)
	code = "let adamicViewAdapterMade=0;\n" + strings.Replace(code, "const adapter = make(underlying);", "adamicViewAdapterMade++; const adapter = make(underlying);", 1) + "\nconsole.log('adapter-makes:'+adamicViewAdapterMade);\n"
	got = v4IdentityJavaScript(t, code)
	if got.exitCode != 0 || string(got.stdout) != "ok\nadapter-makes:1\n" || len(got.stderr) != 0 {
		t.Fatalf("JavaScript adapter factory count: %#v", got)
	}
	code = strings.Replace(code, "const held = views.get(key)?.deref();", "const held = undefined;", 1)
	got = v4IdentityJavaScript(t, code)
	if got.exitCode != 0 || string(got.stdout) != "ok\nadapter-makes:1001\n" || len(got.stderr) != 0 {
		t.Fatalf("unconditional-wrap must only fail JavaScript factory bound: %#v", got)
	}
	t.Logf("JavaScript unconditional-wrap caught by bounded allocation: %q", got.stdout)
}

func TestV4EscapeAdapterConstructorIdentity(t *testing.T) {
	t.Parallel()
	v4DirectSource(t, `console.log(`+"`${Array===Array}`"+`);`, "true\n", "")
}

func TestV4EscapeAdapterCollectionKeyContract(t *testing.T) {
	t.Parallel()
	source := v4EscapeDeclarations + `
const raw={kind:'Receiver' as const,run:(value:'ok'):string=>'producer'};
function argument():string { console.log('argument'); return 'bad'; }
function probe(base:Base):void { const view=base as Target; const escaped=view.run; const set=new Set<(value:string)=>string>(); set.add(escaped); for(const key of set) { console.log(key(argument())); } }
probe(raw);`
	program, truth := v4IdentityProgram(t, source, "argument\nproducer\n")
	sanitized, _ := nativelyUncached(t, program)
	for name, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || string(got.stdout) != "argument\n" || !strings.Contains(string(got.stderr), "argument 1 expected producer \"ok\", view string") || strings.Contains(string(got.stderr), "Sanitizer") {
			t.Fatalf("%s iterated key lost its checking contract: %#v; source Node %#v", name, got, truth)
		}
	}
}

func TestV4EscapeAdapterScalarIdentityControls(t *testing.T) {
	t.Parallel()
	for _, fixture := range []string{"library_object_same.a", "library_map_set_keys.a"} {
		path, err := filepath.Abs(filepath.Join(repository, "internal", "oracle", "testdata", fixture))
		if err != nil {
			t.Fatal(err)
		}
		truth := onNode(t, path)
		if truth.exitCode != 0 {
			t.Fatalf("Node %s: %#v", fixture, truth)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		v4EscapeAdmitted(t, program, truth)
	}
}
