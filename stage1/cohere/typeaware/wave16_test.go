package typeaware

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var wave16RuleNames = []string{"no-import-assign", "prefer-exponentiation-operator", "use-isnan"}

func wave16Controls() []string {
	return []string{
		"import {value as renamed} from './helper.a'; renamed=2; renamed++; ({renamed}=obj); [renamed]=rows; declare const obj:any,rows:any;",
		"import * as ns from './helper.a'; (ns.value)=3; ((ns['value']))++; [ns.value]=[]; ({x:ns.value}=obj); declare const obj:any; ns.value=2; ns['value']++; delete (ns.value); Object.assign(ns,{}); Reflect.set(ns,'value',2); Object.freeze(ns);",
		"import * as ns from './helper.a'; ns.object.x=2; Object.assign(ns.object,{}); Object.seal(ns); const o={}; Object.assign(o,ns); o[ns as any]=2;",
		"import {object} from './helper.a'; object.x=2; function f(object:number){object=2;} import * as ns from './helper.a'; function g(ns:any){ns.value=2;} function h(Object:any){Object.assign(ns,{});}",
		"import {value as a} from './helper.a'; import {value as b} from './helper.a'; b=2; ({x:a}=obj); declare const obj:any;",
		"declare const x:number; x===NaN; NaN!==x; x<Number.NaN; x>=Number['NaN']; x==(1,NaN); switch(NaN){case NaN:break;case Number.NaN:break;}",
		"function f(NaN:number,Number:{NaN:number},x:number){return [x===NaN,x===Number.NaN];} const a=[NaN].indexOf(NaN);",
		"declare const x:number; x===Number[`NaN`]; x!==((NaN)); switch(x){case (1,NaN):break;} x===Number[(1,'NaN')];",
		"Math.pow(2,3); Math.pow(-2,2); Math.pow(2,-2); Math.pow(2**3,4); Math.pow(2,3**4); Math.pow(1+2,3+4); Math.pow((2),(3));",
		"declare let a:number,b:number; a+Math.pow(++b,2); Math.pow(a?b:2,3); Math.pow(a,b)in {}; Math.pow(a,b).toFixed(); typeof Math.pow(a,b); Math.pow(a,b) as number;",
		"Math.pow(); Math.pow(2); Math.pow(2,3,4); Math.pow(...[2,3]); Math.pow(2,/*keep*/3); Math.pow(/*keep*/2,3); Math./*keep*/pow(2,3);",
		"Math['pow'](2,3); Math[`pow`](2,3); Math['p'+'ow'](2,3); Math[`p${'ow'}`](2,3); globalThis.Math.pow(2,3); Math?.pow(2,3); (Math.pow)(2,3);",
		"function f(Math:{pow:(a:number,b:number)=>number}){return Math.pow(2,3);} const pow=Math.pow; pow(2,3); const M=Math; M.pow(2,3);",
		"Math.pow('/* string */' as any,2); Math.pow(/a/.test('a') as any,2); Math.pow({x:1} as any,2);",
		"declare let a:number,b:number; Math.pow(a,b)(); new (Math.pow(a,b))(); 2**Math.pow(a,b); [Math.pow(a,b)]; a[Math.pow(a,b)];",
		"/* 世界 🌍 */\r\nimport * as 漢 from './helper.a';Object.assign(漢,{});Math.pow(-2,3);2===NaN;\r\n",
	}
}

func wave16SourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
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
				h.t.Fatalf("nonunique %s mutant", name)
			}
			source = strings.Replace(source, from, to, 1)
		}
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		source = regexp.MustCompile(`'\./([^']+\.ts)'`).ReplaceAllString(source, "'"+filepath.Join(h.repository, "stage1/cohere/typeaware")+"/$1'")
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave16_suite.a"), archive, false)
}

// All output, including failures and mutants, goes to files. This suite is serial
// because sanitizer archives and corpus runs share the limited scratch volume.
func TestWave16AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_WAVE16_ARTIFACTS"); path != "" {
		directory, err = filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave16_suite.a")
	binary := h.build(stage0, "wave16", entry, archive, false)
	oracle := volumeOracle(h, "wave16-oracle", "oracle_wave16.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	paths := []string{}
	for i, source := range wave16Controls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	paths = append(paths, h.write("helper.a", "export let value=1; export const object={x:1};\n"))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range wave16RuleNames {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave16-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	// Each rule mutant changes a real judgment or report range, finishes normally,
	// and is killed exclusively by the independent cohere diagnostic bytes.
	for _, change := range []struct{ name, file, from, to string }{
		{"import-write", "no_import_assign.a", "const direct = reassign.writes(index);", "const direct = false;"},
		{"exponent-base", "prefer_exponentiation_operator.a", "const baseParens = this.baseParens(base);", "const baseParens = false;"},
		{"nan-comma", "use_isnan.a", "this.rules.parser.node(node.children[1] ?? -1).kind === 'CommaToken'", "this.rules.parser.node(node.children[1] ?? -1).kind === 'PlusToken'"},
	} {
		t.Run(change.name, func(t *testing.T) {
			previous := h.t
			h.t = t
			defer func() { h.t = previous }()
			mutant := wave16SourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
			got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
			if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
				t.Fatalf("%s mutant survived", change.name)
			}
			t.Logf("%s: exit 0, Go byte oracle catches byte %d; %s", change.name, firstDifference(got.stdout, truth.stdout), summary(got.stdout))
			if err := os.Remove(mutant); err != nil {
				t.Fatal(err)
			}
		})
	}
	// The existing fact API is independently held by its direct checker tests.
	if manifest := os.Getenv("ADAMIC_WAVE16_REPOSITORY_MANIFEST"); manifest != "" {
		h.compare("repository", oracle, binary, filepath.Join(repository, "tsconfig.json"), manifest)
		h.compare("repository-asan", oracle, asan, filepath.Join(repository, "tsconfig.json"), manifest)
	}
	if manifest := os.Getenv("ADAMIC_WAVE16_COMPILER_MANIFEST"); manifest != "" {
		corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
		if corpus == "" {
			t.Fatal("compiler source is required")
		}
		h.compare("compiler", oracle, binary, filepath.Join(corpus, "src/compiler/tsconfig.json"), manifest)
		h.compare("compiler-asan", oracle, asan, filepath.Join(corpus, "src/compiler/tsconfig.json"), manifest)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','node-symbol-details'));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released program queried using the existing fact API: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	got = h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released-registry mutant exits 0, caught by the required panic 70")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
