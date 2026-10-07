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

func wave11Controls() []string {
	return []string{
		`declare const s:string; s.match(/x/); s.match(/x/g); s.match('a/b'); s.match(''); s.match('[a-z'); s.match(/x/, 1); s['match'](/x/);`,
		"declare const s:string; s.match(new RegExp('x')); s.match(RegExp('x','i')); s.match(RegExp('x',undefined)); s.match(`x`);",
		`declare const s:string; const r=/x/; let one=/x/; var once=/x/; s.match(r);s.match(one);s.match(once); const pattern='a'+'b'; s.match(pattern);`,
		"declare const s:string; declare const t:string; const dynamic=new RegExp(`^${t}$`); s.match(dynamic); s.match(new RegExp(`^${t}$`)); const r=new RegExp('test','');s.match(r);",
		`declare const s:string; let r=/x/;s.match(r);r=/x/g; let other=/x/; function nested(){other=/y/g;}s.match(other);`,
		`declare const s:string; let r=/x/; {let r=/x/;r=/x/g;} s.match(r); const p=/x/g; s.match(p); declare const flags:string;s.match(new RegExp('x',flags));`,
		`declare const s:string; s.match(/x/) || null; !s.match(/x/); (s.match(/x/)) || null; s.match(/x/).length; s?.match(/x/); s.match?.(/x/);`,
		"declare const s:string; declare const n:number; (`${s}`).match(/x/); (s+n).match(/x/); function generic<T extends string>(s:T){return s.match(/x/);}declare const branded:string & {brand:true};branded.match(/x/);",
		`declare const s:'a'|'b';s.match(/x/);declare const o:{match:(r:RegExp)=>unknown};o.match(/x/);declare const a:RegExp|string;declare const b:string;a.match(b);`,
		`declare const s:string; let r=/x/; ({r}={r:/x/g});s.match(r); let a=/x/;[a]=[/x/g];s.match(a);`,
		`declare const s:string; s.match('(?=x)x');s.match('(x)\\1');s.match('[^]');s.match('[]');s.match('[z-a]');s.match('[');s.match('a\\/b');s.match('a\n');`,
		`for(const key in [1,2]){} for(const key of [1,2]){} for(const key in {x:1}){} for(const key in 'text'){}`,
		`declare const t:[number,string];for(const key in t){}declare const rows:string[]|null;for(const key in rows){}function f<T extends any[]>(x:T){for(const k in x){}}`,
		`declare const a:{[n:number]:string;length:1};declare const b:{[n:number]:string};declare const c:{length:number};for(const k in a){}for(const k in b){}for(const k in c){}`,
		`function f(){for(const k in arguments){}}declare const a:readonly number[];for(const k in a){}declare const b:{[n:number]:string;length:string};for(const k in b){}`,
		"/* 世界 🌍 */\r\ndeclare const rows:number[];for (const k in\n ((rows))\n /* before close */ )\n /* before body */ {}\r\n",
		`declare const a:boolean;declare const b:boolean;const x=a?'pro':'pro';const y=a?'pro':'std';const z=a?'x':b?'x':'x';const w=a?'y':b?'x':'x';`,
		`declare const a:boolean;declare const b:boolean;declare function f():void;declare function g():void;if(a){f();}else f();if(a)f();else if(b)f();else f();if(a)g();else if(b)f();else f();`,
		`declare const a:boolean;if(a){}else{} if(a){/* intent */}else{/* other */}if(a);else;function f(){if(a)return;else return;}`,
		`declare const a:boolean;const x=a?([]):[ ];const y=a?'/* */':'/* */';const z=a?'a  b':'a b';const w=a?'x':"x";`,
		`declare function format(x:string):string;declare function format(x:number):number;declare const v:string|number;const x=typeof v==='string'?format(v):format(v);`,
		`declare const v:string|number;const x=typeof v==='string'?v+1:v+1;declare const a:boolean;declare const n:number;const y=a?n+1:n+1;`,
		`declare const v:{type:'a';source:string}|{type:'b';source:string};const x=v.type==='a'?v.source:v.source;declare const a:boolean;const y=a?(true?1:2):(true?1:2);`,
		`declare function f(...args:[number]|[string,number]):void;declare const v:[number]|[string,number];const x=v.length===1?f(...v):f(...v);`,
		"/* 世界 🌍 */\r\ndeclare const a:boolean;const 漢=a?42 /* first */:42;\r\n",
	}
}

func wave11SourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
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
	return h.build(stage0, name, filepath.Join(directory, "wave_11_suite.a"), archive, false)
}

// Not parallel: archive builds, sanitizer subprocesses and measurements share a machine.
func TestWave11AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE_11_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_11_suite.a")
	binary := h.build(stage0, "wave-11", entry, archive, false)
	oracle := volumeOracle(h, "wave-11-oracle", "oracle_wave_11.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for i, source := range wave11Controls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.ts", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"@typescript-eslint/prefer-regexp-exec", "@typescript-eslint/no-for-in-array", "nexus/correctness-no-identical-branches"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control: %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave-11-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"regexp", "prefer_regexp_exec.a", "finding.fixEnd = this.rules.byte(node.end);", "finding.fixEnd = this.rules.byte(node.end) + 1;"},
		{"array", "no_for_in_array.a", "this.rules.byte(this.rules.parser.node(body).pos))", "this.rules.byte(this.rules.parser.node(body).pos) + 1)"},
		{"branches", "correctness_no_identical_branches.a", "this.rules.byte(finding.end), '')", "this.rules.byte(finding.end) + 1, '')"},
		{"signature", "resolved_signature_equal.a", "return equal;", "return true;"},
		{"index", "number_index_type.a", "return present;", "return false;"},
		{"syntax", "regular_expression_syntax.a", "return valid;", "return true;"},
	} {
		mutant := wave11SourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
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
console.log(tsgoInspect(program,file,0,1,'Identifier','regular-expression-syntax\nx'));
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
