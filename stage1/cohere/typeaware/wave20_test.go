package typeaware

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func wave20Controls() []string {
	return []string{
		"declare const p:Promise<number>;p;void p;await p;const held=p;p.catch(()=>{});p.then(()=>{},()=>{});p.then(()=>{});p.catch(null);p.catch();p.finally(()=>{});\nexport {};\n",
		"declare const p:Promise<number>;[p];declare const ps:Promise<number>[];ps;declare const tuple:readonly [number,Promise<string>];tuple;declare const mixed:number[]|Promise<number>[];mixed;\nexport {};\n",
		"declare const p:Promise<number>;p,p;true?p:p;p||p;p&&p;p??p;(p.catch)(()=>{});(p.catch(()=>{}));(p).catch(()=>{});enum E{Catch='catch'}p[E.Catch](()=>{});\nexport {};\n",
		"declare const p:PromiseLike<number>;p;declare const q:{then(cb:(x:number)=>void):void};q;void q;\nexport {};\n",
		"declare const p:Promise<number>;function f<T extends Promise<number>>(x:T){x;}class P extends Promise<number>{};declare const sub:P;sub;declare const branded:Promise<number>&{brand:true};branded;\nexport {};\n",
		"declare function setTimeout(x:unknown,n:number):number;setTimeout('bad',0);setTimeout(()=>{},0);\nexport {};\n",
		"globalThis.setTimeout('bad',0);window.setInterval('bad',0);window['setTimeout']('bad',0);Function('return 1');new Function('return 1');\nexport {};\n",
		"declare const setTimeout:(x:unknown,n:number)=>number;setTimeout('bad',0);\nexport {};\n",
		"declare const fn:()=>void;declare const anyValue:any;declare const unknownValue:unknown;window.setTimeout(fn,0);window.setTimeout(anyValue,0);window.setTimeout(unknownValue,0);window.setTimeout((()=>{}),0);window.setTimeout(fn.bind(null),0);(window.setTimeout)('bad',0);\nexport {};\n",
		"declare function f():void;declare function u():undefined;declare function v():void|undefined;void f();void /* comment */ u();void(v());const a=void f();void (()=>{})();\nexport {};\n",
		"declare const x:number;void x;(void x);const a=void x;void 0;void 0x0;void 0.0;void -0;let y=0;void(y=1);void(y+=1);void ((0,x));void (x as number);\nexport {};\n",
		"declare function n():never;declare function f():number;void n();void f();declare const p:Promise<void>;void p;void import('x');\nexport {};\n",
		"interface Then {then(...cb:((x:number)=>void)[]):void};declare const q:Then;void q;interface NotThen {then(...cb:number[]):void};declare const r:NotThen;void r;export {};\n",
		"declare const p:Promise<number>;p!;p as Promise<number>;p satisfies Promise<number>;<Promise<number>>p;declare const ps:Promise<number>[];await ps;export {};\n",
		"declare const x:number;void <number>x;declare function f():void;void <void>f();void (0, f());export {};\n",
		"declare const f:Function;window.setTimeout(f,0);self.setTimeout('bad',0);window[`setTimeout`]('bad',0);window.setTimeout(123,0);export {};\n",
		"/* 世界 🌍 */\ndeclare function f():void;void /*漢*/ f();declare const p:Promise<void>; (p);\n\nexport {};\n",
	}
}

func wave20Mutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*.a"))
	if err != nil {
		h.t.Fatal(err)
	}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			h.t.Fatal(err)
		}
		source := string(data)
		if filepath.Base(path) == file {
			if strings.Count(source, from) != 1 {
				h.t.Fatal("nonunique mutant", name)
			}
			source = strings.Replace(source, from, to, 1)
		}
		// Existing .ts dependencies stay pinned; new .a dependencies are copied together.
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		for _, dep := range []string{"rules", "facts", "frames", "types", "type_fact", "unary_minus", "diagnostic", "repair", "suggestion", "caller", "bindings"} {
			source = strings.ReplaceAll(source, "./"+dep+".ts", filepath.Join(h.repository, "stage1/cohere/typeaware", dep+".ts"))
		}
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0644); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave20_suite.a"), archive, false)
}

// Not parallel: native builds and sanitizer archives share the machine and disk.
func TestWave20AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_WAVE20_ARTIFACTS"); path != "" {
		directory = path
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave20_suite.a")
	binary := h.build(stage0, "wave20", entry, archive, false)
	oracle := volumeOracle(h, "wave20-oracle", "oracle_wave20.go")
	h.write("fixture.d.ts", "export {};\n")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","lib":["ES2022","DOM"],"noEmit":true},"sourceExtensions":[".a"]}`)
	var paths []string
	for i, source := range wave20Controls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-floating-promises", "no-implied-eval", "no-meaningless-void-operator"} {
		if !bytes.Contains(truth.stdout, []byte("\t@typescript-eslint/"+name+"\t")) {
			t.Fatal("missing positive control", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave20-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"floating", "no_floating_promises.a", "if(!state.unhandled)", "if(state.unhandled)"},
		{"eval", "no_implied_eval.a", "if(!this.declaredHere(callee))", "if(this.declaredHere(callee))"},
		{"void", "no_meaningless_void_operator.a", "facts.type(part).flags & (16 | 4)", "facts.type(part).flags & (64 | 4)"},
	} {
		t.Run(change.name, func(t *testing.T) {
			previous := h.t
			h.t = t
			defer func() { h.t = previous }()
			mutant := wave20Mutant(h, stage0, archive, change.name, change.file, change.from, change.to)
			got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
			if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
				t.Fatal("mutant survived")
			}
			t.Logf("%s: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
			os.Remove(mutant)
		})
	}
	for _, change := range []struct{ name, file, from, to string }{
		{"accessed-property", "bridge/tsgo/checker/accessed_property.go", "out.text(name)", "out.text(name + \"-mutant\")"},
		{"callback-parameters", "bridge/tsgo/checker/callback_parameters.go", "t = checker.Checker_getIndexTypeOfType(c, t, checker.Checker_numberType(c))", "t = checker.Checker_numberType(c)"},
	} {
		overlay := h.overlay(change.name, change.file, change.from, change.to)
		mutatedArchive := h.archive(change.name+"-checker", overlay, false)
		mutant := h.build(stage0, change.name+"-mutant", entry, mutatedArchive, false)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("question mutant survived", change.name)
		}
		t.Logf("%s question mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		os.Remove(mutant)
		os.Remove(mutatedArchive)
	}
	for _, corpus := range []struct{ name, root, config string }{
		{"repository", repository, filepath.Join(repository, "tsconfig.json")},
		{"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json")},
	} {
		if corpus.root == "" {
			t.Log("compiler corpus not requested")
			continue
		}
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage", corpus.name+".manifest"))
		if err != nil {
			t.Fatal(err)
		}
		var roots []string
		for _, path := range strings.Split(string(data), "\n") {
			if path != "" {
				roots = append(roots, filepath.Join(corpus.root, path))
			}
		}
		manifest := h.write(corpus.name+".manifest", strings.Join(roots, "\n")+"\n")
		h.compare(corpus.name, oracle, binary, corpus.config, manifest)
		h.compare(corpus.name+"-asan", oracle, asan, corpus.config, manifest)
		// Isolated complete-stream process timings, separate from sanitizer runs.
		goRun := h.must(corpus.name+"-timed-go", exec.Command(oracle, corpus.config, manifest))
		command := exec.Command(binary, corpus.config, manifest)
		command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		native := h.must(corpus.name+"-timed-native", command)
		if !bytes.Equal(goRun.stdout, native.stdout) {
			t.Fatal("timed bytes differ")
		}
		t.Logf("%s whole process native %.6fs Go %.6fs; %s; native phases %s; Go phases %s", corpus.name, native.elapsed.Seconds(), goRun.elapsed.Seconds(), summary(native.stdout), native.stderr, goRun.stderr)
	}
	probe := h.write("probe.a", "x.y;\n")
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,3,'PropertyAccessExpression','accessed-property'));
`)
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released accessed-property query: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant keeps the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	got = h.must("released-mutant-run", exec.Command(mutant, config, probe))
	if len(got.stderr) != 0 {
		t.Fatal("registry mutant stderr")
	}
	t.Log("released-registry mutant: exit 0, required panic catches it")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
