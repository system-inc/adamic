package oracle

import (
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"strings"
	"testing"
)

func v4EscapeSurface(t *testing.T, source, output string, mutant bool) {
	t.Helper()
	program, truth := v4IdentityProgram(t, source, output)
	got, counts := v4IdentityCounted(t, native.C(program))
	if difference := disagreement(truth, got); difference != "" {
		t.Fatal(difference)
	}
	t.Logf("native sanitized surface counts: %v", counts)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("%s: %s", backend, difference)
		}
	}
	if !mutant {
		return
	}
	code := native.C(program)
	before := "value = adamic_view_adapter_underlying(value);"
	if strings.Count(code, before) == 0 {
		t.Fatal("native surface mutation site moved")
	}
	got, _ = v4IdentityCounted(t, strings.ReplaceAll(code, before, "/* mutant: read the adapter's own surface */"))
	if got.exitCode != 0 || disagreement(truth, got) == "" {
		t.Fatalf("native own-surface mutant survived: %#v", got)
	}
	t.Logf("native/ASan/UBSan/LSan own-surface mutant caught: expected %q, got %q", truth.stdout, got.stdout)
	code = javascript.JavaScript(program)
	before = "value = adamicViewAdapterUnderlying(value);"
	if strings.Count(code, before) == 0 {
		t.Fatal("JavaScript surface mutation site moved")
	}
	got = v4IdentityJavaScript(t, strings.ReplaceAll(code, before, "/* mutant: read the adapter's own surface */"))
	if got.exitCode != 0 || len(got.stderr) != 0 || disagreement(truth, got) == "" {
		t.Fatalf("JavaScript own-surface mutant survived: %#v", got)
	}
	t.Logf("JavaScript own-surface mutant caught: expected %q, got %q", truth.stdout, got.stdout)
}

func TestV4EscapeAdapterLength(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(value:string)=>'good'}
function producer(value:'ok',second?:string,third?:string):string {console.log('called');return 'bad'}
const raw={kind:'Receiver' as const,run:producer};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log(` + "`${escaped.length}`" + `);}
probe(raw);`
	v4EscapeSurface(t, source, "3\n", true)
}

func TestV4EscapeAdapterName(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(value:string)=>'good'}
function producer(value:'ok'):string {console.log('called');return 'bad'}
const raw={kind:'Receiver' as const,run:producer};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log(escaped.name);}
probe(raw);`
	v4EscapeSurface(t, source, "producer\n", true)
}

func TestV4EscapeAdapterSurfaceDefaults(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(value:string)=>string}
const raw={kind:'Receiver' as const,run:(value:'ok',other:string='default',third?:string):string=>'bad'};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log(` + "`${escaped.length}:${escaped.name}`" + `);}
probe(raw);`
	v4EscapeSurface(t, source, "1:run\n", false)
}

func TestV4EscapeAdapterSurfaceReturned(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(value:string)=>string}
const producer=(value:'ok'):string=>'bad';
const raw={kind:'Receiver' as const,run:producer};
function read(base:Base):(value:string)=>string {const view=base as Target;return view.run;}
function again(base:Base):(value:string)=>string {console.log('read');return read(base);}
const escaped=again(raw);const next={kind:'Receiver' as const,run:escaped};const revisited=read(next);
console.log(` + "`${revisited.length}:${revisited.name}:${producer.length}:${producer.name}`" + `);`
	v4EscapeSurface(t, source, "read\n1:producer:1:producer\n", false)
}

func TestV4EscapeAdapterSurfaceLengthRefusal(t *testing.T) {
	t.Parallel()
	v4EscapeRefusal(t, v4EscapeDeclarations+`
const raw={kind:'Receiver' as const,run:(value:'ok'):string=>'bad'};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log(`+"`${escaped.length}`"+`);}
probe(raw);`, "1\n")
}

func TestV4EscapeAdapterSurfaceNameRefusal(t *testing.T) {
	t.Parallel()
	v4EscapeRefusal(t, v4EscapeDeclarations+`
const raw={kind:'Receiver' as const,run:(value:'ok'):string=>'bad'};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log(escaped.name);}
probe(raw);`, "run\n")
}

func TestV4EscapeAdapterSurfaceComputedName(t *testing.T) {
	t.Parallel()
	source := v4EscapeDeclarations + `
const raw={kind:'Receiver' as const,['run']:(value:'ok'):string=>'bad'};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log(escaped.name);}
probe(raw);`
	v4EscapeSurface(t, source, "run\n", false)
}

func TestV4EscapeAdapterSurfaceAnonymousName(t *testing.T) {
	t.Parallel()
	source := v4EscapeDeclarations + `
function make():(value:'ok')=>string {return (value:'ok'):string=>'bad';}
const raw={kind:'Receiver' as const,run:make()};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log(escaped.name);}
probe(raw);`
	v4EscapeSurface(t, source, "\n", false)
}
