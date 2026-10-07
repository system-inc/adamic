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

func wave26ThirdControls() []string {
	controls := []string{
		`eval('1');`, `(eval as any)('1');`, `(0,eval)('1');`, `const alias=eval;`, `eval?.('1');`,
		`function f(eval:any){eval('1');}`, `function f(eval:any){const x=eval;}`,
		"window.eval('1');window['eval']('1');window[`eval`]('1');",
		`window.window.eval('1');globalThis.globalThis['eval']('1');`,
		`foo.window.eval('1');window.globalThis.eval('1');this.eval('1');`,
		`const x={eval:1};x.eval;const y={eval};`,
		`function f(x=eval){return x;}class C {field=eval;}const [x=eval]=[];`,
		`const eval=1;const alias=eval;`, `/* 世界 🌍 */` + "\r\n" + `window.eval('1');`,
		`Object.prototype.p=0;`, `Array.prototype['p']=0;`,
		`Object.defineProperty(Array.prototype,'p',{value:0});`,
		`Object['defineProperties'](String['prototype'],{p:{value:0}});`,
		`Object.prototype.p&&=0;Object.prototype.p||=0;Object.prototype.p??=0;`,
		`Object.prototype.p++;delete Object.prototype.p;Object.prototype=0;Object.prototype.p.q=0;`,
		`Object.freeze(Array.prototype);Object.defineProperty(x,Array.prototype);Object.defineProperty();`,
		`function f(Object:any){Object.prototype.p=0;Object.defineProperty(Array.prototype,'p',{});}`,
		`function f(Array:any){Array.prototype.p=0;}`,
		`globalThis.Object.prototype.p=0;parseFloat.prototype.p=0;`,
		`(Object?.prototype).p=0;Object.defineProperty((Array.prototype));`,
		`function f(){}f=1;`, `function f(){f=1;}f=2;`, `f=1;function f(){}`,
		`const a=function f(){f=1;};`, `const f=function(){f=1;};const g=()=>{};g=1;`,
		`function f(f:any){f=1;}function g(){var g;g=1;}`,
		`function f(){}{function f(){}f=1;}`, `function f(){}f.x=0;f();`,
		`function f(){}({f}={});`, `function f(){}({x:f=0}={});`,
		`function f(){}({files=f}={});`, `function f(){}[...f]=[];({ ...f }={});`,
		`function f(){}[...(f)]=[];({...(f)}={});`,
		`function f(){}for(f of []){}for(f in {}){}`,
		`function f(){}for([...(f)] of []){}for({...f} of []){}`,
		`function f(){}++f;f--;f+=1;f&&=1;`, `function f(){}const values=[...f];f(...f);`,
		`function f(){}({[f]:x}={});const {x=f}={};`,
		`function méthoδ(){}/* 世界 🌍 */` + "\r\n" + `méthoδ=1;`,
	}
	for i := range controls {
		controls[i] += "\nexport {};\n"
	}
	return controls
}

func wave26ThirdMutant(h *harness, stage0, archive, name, file, from, to string) string {
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	for _, extension := range []string{"*.ts", "*.a"} {
		paths, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware", extension))
		if err != nil {
			h.t.Fatal(err)
		}
		for _, path := range paths {
			data, err := os.ReadFile(path)
			if err != nil {
				h.t.Fatal(err)
			}
			source := string(data)
			if filepath.Base(path) == file {
				if strings.Count(source, from) != 1 {
					h.t.Fatalf("nonunique mutant %s", name)
				}
				source = strings.Replace(source, from, to, 1)
			}
			source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
			source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
			if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0644); err != nil {
				h.t.Fatal(err)
			}
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_26_third.a"), archive, false)
}

// Not parallel: builds, sanitizers and cost observations share a machine.
func TestWave26ThirdAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE26_THIRD_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_26_third.a")
	binary := h.build(stage0, "wave26", entry, archive, false)
	oracle := volumeOracle(h, "wave26-oracle", "oracle_wave_26_third.go")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"jsx":"preserve","target":"ES2022","module":"ESNext","lib":["ES2022","DOM"],"noEmit":true},"files":["control-000.a","configured.d.ts"]}`)
	h.write("configured.d.ts", "")
	h.write("exports.d.ts", "export const value:number; export default {value:1};\n")
	var paths []string
	for i, source := range wave26ThirdControls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d%s", i, func() string {
			if strings.Contains(source, "<div") {
				return ".tsx"
			}
			return ".a"
		}()), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-eval", "no-extend-native", "no-func-assign"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave26-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"eval", "no_eval.a", "['global', 'window', 'globalThis']", "['global', 'globalThis']"},
		{"extend", "no_extend_native.a", "this.access.assignment(node.operator)", "node.operator === 'EqualsToken'"},
		{"function", "no_func_assign.a", "origin.local(this.tree.path, shorthand)", "origin.local(this.tree.path)"},
	} {
		mutant := wave26ThirdMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		os.Remove(mutant)
	}
	for _, change := range []struct{ name, file, from, to string }{
		{"syntax", "bridge/tsgo/checker/binding_structure.go", `operator = strings.TrimPrefix(b.OperatorToken.Kind.String(), "Kind")`, `operator = "BarBarToken"`},
		{"origin", "bridge/tsgo/checker/binding_origin.go", "out.yes(f.IsDeclarationFile)", "out.yes(false)"},
	} {
		overlay := h.overlay(change.name, change.file, change.from, change.to)
		mutatedArchive := h.archive(change.name, overlay, false)
		mutant := h.build(stage0, change.name+"-question", entry, mutatedArchive, false)
		got := h.must(change.name+"-question-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s question mutant survived", change.name)
		}
		t.Logf("%s question mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		os.Remove(mutatedArchive)
		os.Remove(mutant)
	}
	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE26_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE26_COMPILER_MANIFEST")},
	} {
		if corpus.manifest == "" {
			continue
		}
		h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
		h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		want := h.must(corpus.name+"-timed-go", exec.Command(oracle, corpus.config, corpus.manifest))
		native := exec.Command(binary, corpus.config, corpus.manifest)
		native.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		got := h.must(corpus.name+"-timed-native", native)
		if !bytes.Equal(want.stdout, got.stdout) {
			t.Fatal("timed output mismatch")
		}
		t.Logf("%s whole process native %s Go %s; native phases %s; Go phases %s", corpus.name, got.elapsed, want.elapsed, got.stderr, want.stderr)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);console.log(tsgoInspect(program,file,0,1,'Identifier','raw-shape'));tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','binding-origin'));`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released handle: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released-registry mutant: exit 0, required panic 70 catches it")
}
