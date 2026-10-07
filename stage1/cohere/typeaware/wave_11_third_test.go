package typeaware

import (
	"bytes"
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func wave11ThirdControls() []string {
	return []string{
		`eval('x');eval(eval);eval?.(eval);eval!('x');(eval as any)('x');(eval satisfies Function)('x');(<any>eval)('x');(eval)('x');(0,eval)('x');const run=eval;eval.toString();`,
		`function f(eval:(x:string)=>unknown){eval('x');const v=eval;eval?.('x')}function g(){var eval=(x:string)=>x;const v=eval;}function h(){const v=eval;function eval(x:string){return x}}`,
		"window.eval('x');window.window.eval('x');window['eval']('x');global.global[`eval`]('x');globalThis.globalThis['eval']('x');(window?.window).eval('x');window[('eval')]('x');",
		`global.window.eval('x');window.global.eval('x');other.window.eval('x');this.eval('x');class C{eval(){}method(){this.eval()}}const window={eval:(x:string)=>x};window.eval('x');`,
		"declare const key:string;window[key]('x');window[`ev${key}al`]('x');window[eval]('x');window['\\u0065val']('x');\u0065val('x');",
		`const evalValue=eval;function f(x=eval){}class C{x=eval;eval=1;method(){return eval}}const object={eval:eval};const short={eval};enum E{x=eval}`,
		`interface I{eval():void};type T=typeof eval;declare const other:{eval:unknown};other.eval;({eval:0});declare const box:{[eval]:number};`,
		`Object.prototype.p=0;Function.prototype['p']=0;String['prototype'].p=0;Number['prototype']['p']=0;Object.defineProperty(Array.prototype,'p',{value:0});Object.defineProperties(Array.prototype,{p:{value:0}});`,
		`Object.prototype.p++;delete Object.prototype.p;Object.prototype=0;Object.prototype.p.q=0;Object.freeze(Array.prototype);Object.defineProperty({},Array.prototype);Object.defineProperty();`,
		`Array.prototype.p&&=0;Array.prototype.p||=0;Array.prototype.p??=0;Array.prototype.p+=1;Array.prototype.p**=2;(Object?.prototype).p=0;Object?.defineProperty(Object.prototype);`,
		"Object[`prototype`]['p']=1;Object[('prototype')].p=1;(Object)['prototype'].p=1;Object['defineProperty'](Array.prototype);Object[`defineProperties`](Array.prototype);",
		`function f(){const Object={prototype:{},defineProperty(...args:unknown[]){}};Object.prototype.p=1;Object.defineProperty(Array.prototype)}{let Array={prototype:{}};Array.prototype.p=1}const own=Object;own.prototype.p=1;globalThis.Object.prototype.p=1;`,
		`parseFloat.prototype.p=1;parseInt.prototype.p=1;undefined.prototype.p=1;globalThis.prototype.p=1;Foo.prototype.p=1;`,
		`function foo(){}foo=1;foo+=1;foo++;--foo;foo&&=1;foo||=1;foo??=1;foo();foo.x=1;object.foo=1;`,
		`foo=1;function foo(){}function f(){function g(){}g=1;}const named=function inner(){inner=1;};const anonymous=function(){anonymous=1};const arrow=()=>{};arrow=1;`,
		`function foo(foo:unknown){foo=1}function other(){var other;other=1}function outer(){}{let outer;outer=1}{function outer(){}outer=1}`,
		`function foo(){}[foo]=[];[...foo]=[];[[...foo]]=[];({foo}={});({x:foo=0}={});({...foo}={});({x:{...foo}}={});[...(foo)]=[];({... (foo)}={});`,
		`function foo(){}({files=foo}={});({[foo]:other}={});const arr=[...foo];foo(...foo);function f(){for(const foo of []){foo=1}}`,
		`function foo(){}for(foo of []){}for(foo in {}){}for([...foo] of []){}for({...foo} of []){}(foo)=1;(foo as unknown)=1;`,
		`function foo(){}function foo(x:number):void;foo=1;function foo(){var foo=1;foo=2}`,
		"/* 世界 🌍 */\r\nfunction é(){}\r\né=1;\r\nwindow['eval']('x');\r\nObject.prototype.p=0;\r\n",
		`function \u0066oo(){}foo=1;function f(){function f(){}f=1}declare const other:unknown;eval(other);`,
	}
}

// Read source inputs from the pinned upstream-derived fixture tables. Options
// remain the production defaults on both sides; expectations are never copied.
func wave11ThirdUpstream(t *testing.T, repository string) []string {
	t.Helper()
	var sources []string
	for _, name := range []string{"no_eval", "no_extend_native", "no_func_assign"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/core", name+"_test.go"), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		goast.Inspect(tree, func(node goast.Node) bool {
			table, ok := node.(*goast.CompositeLit)
			if !ok {
				return true
			}
			array, ok := table.Type.(*goast.ArrayType)
			if !ok {
				return true
			}
			fields, ok := array.Elt.(*goast.StructType)
			if !ok {
				return true
			}
			sourceIndex := -1
			offset := 0
			for _, field := range fields.Fields.List {
				for _, name := range field.Names {
					if name.Name == "sourceText" || name.Name == "source" {
						sourceIndex = offset
					}
					offset++
				}
			}
			if sourceIndex < 0 {
				return true
			}
			for _, element := range table.Elts {
				row, ok := element.(*goast.CompositeLit)
				if !ok || sourceIndex >= len(row.Elts) {
					continue
				}
				value := row.Elts[sourceIndex]
				for _, entry := range row.Elts {
					if pair, ok := entry.(*goast.KeyValueExpr); ok {
						if key, ok := pair.Key.(*goast.Ident); ok && (key.Name == "sourceText" || key.Name == "source") {
							value = pair.Value
						}
					}
				}
				literal, ok := value.(*goast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					continue
				}
				source, err := strconv.Unquote(literal.Value)
				if err != nil {
					t.Fatal(err)
				}
				sources = append(sources, source)
			}
			return true
		})
	}
	return sources
}

func wave11ThirdSourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	for _, pattern := range []string{"*.ts", "*.a"} {
		files, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware", pattern))
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
	return h.build(stage0, name, filepath.Join(directory, "wave_11_third_suite.a"), archive, false)
}

// Not parallel: archive builds, sanitizer subprocesses and measurements share a machine.
func TestWave11ThirdAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE_11_THIRD_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_11_third_suite.a")
	binary := h.build(stage0, "wave-11-third", entry, archive, false)
	oracle := volumeOracle(h, "wave-11-third-oracle", "oracle_wave_11_third.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for i, source := range append(wave11ThirdControls(), wave11ThirdUpstream(t, repository)...) {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.ts", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-eval", "no-extend-native", "no-func-assign"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control: %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave-11-third-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"eval", "no_eval.a", "this.rules.byte(node.end), '')", "this.rules.byte(node.end) + 1, '')"},
		{"extend", "no_extend_native.a", "this.rules.byte(node.end), '')", "this.rules.byte(node.end) + 1, '')"},
		{"assign", "no_func_assign.a", "this.rules.byte(node.end), '')", "this.rules.byte(node.end) + 1, '')"},
		{"global", "wave_11_third_support.a", "declaration !== undefined && declaration.declarationFile", "declaration !== undefined && !declaration.declarationFile"},
		{"anchor", "no_func_assign.a", "!this.anchors.has(this.declaration(index))", "!this.names.has(node.text)"},
	} {
		mutant := wave11ThirdSourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, independent Go bytes catch byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		if err := os.Remove(mutant); err != nil {
			t.Fatal(err)
		}
	}
	for _, population := range []struct{ name, root, config, manifest string }{
		{"repository", repository, filepath.Join(repository, "tsconfig.json"), "repository.manifest"},
		{"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), "compiler.manifest"},
	} {
		if population.root == "" {
			t.Log("compiler corpus not supplied")
			continue
		}
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage", population.manifest))
		if err != nil {
			t.Fatal(err)
		}
		var roots []string
		for _, path := range strings.Split(string(data), "\n") {
			if path != "" {
				roots = append(roots, filepath.Join(population.root, path))
			}
		}
		rootsManifest := h.write(population.name+".manifest", strings.Join(roots, "\n")+"\n")
		h.compare(population.name, oracle, binary, population.config, rootsManifest)
		h.compare(population.name+"-asan", oracle, asan, population.config, rootsManifest)
		for _, impl := range []struct{ name, path string }{{"go", oracle}, {"native", binary}} {
			command := exec.Command(impl.path, population.config, rootsManifest)
			command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
			got := h.must(population.name+"-timed-"+impl.name, command)
			t.Logf("%s %s process=%s %s", population.name, impl.name, got.elapsed, strings.TrimSpace(string(got.stderr)))
		}
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','declaration-lineage'));
`)
	probe := h.write("probe.ts", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released query: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released program.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	got = h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released-registry mutant exits 0, caught by required panic 70")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
