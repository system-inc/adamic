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

// Not parallel: compiler builds, sanitizers and measurements share the machine.
func TestWave15AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE15_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_15_suite.a")
	binary := h.build(stage0, "wave15", entry, archive, false)
	oracle := volumeOracle(h, "wave15-oracle", "oracle_wave_15.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	controls := []string{
		"declare const rows:number[]; rows.sort(); rows.toSorted(); rows.sort(undefined); declare const words:string[];words.sort();",
		"declare const rows:number[]|string[]; rows.sort(); declare const fake:{sort():void};fake.sort(); declare const tuple:[number,number];tuple.sort();",
		"declare const rows:number[]; rows.filter(x=>x>0)[0]; rows.filter(x=>x>0).at(0); rows.filter(x=>x>0)[1];rows.filter(x=>x>0).at(0.2);rows.filter(x=>x>0).at('a');",
		"declare const rows:number[];const zero=0;const key='filter';rows[key](x=>x>0)[zero];rows.filter(x=>x>0) /* [ . ? comment */ .at('0');",
		"declare const rows:number[];declare const other:number[];declare const yes:boolean;(yes?rows.filter(x=>x>0):other.filter(x=>x>1))[0];",
		"function f(a=undefined,b:number|undefined=undefined,c:string='ok'){return [a,b,c];}declare const object:{a:string};const {a='default'}=object;",
		"declare const tuple:[string,string?];const [a='default',b='live']=tuple;declare const rows:string[];const [c='live']=rows;",
		"declare const rows:number[];const h:(x:number)=>number=x=>x;rows.map((x=42)=>x);const f:(x:number)=>number=(x=42)=>x;const g:(x?:number)=>number=(x=42)=>x;",
		"declare const object:{nested:{a:string}};const {nested:{a='default'}}=object;declare const cond:boolean;const {z='live'}=cond?{z:'ok'}:{};",
		"/* 世界 🌍 */\r\ndeclare const rows:number[];rows['sort']();rows['filter'](x=>x>0)[0];\r\n",
	}
	paths := []string{}
	for i, s := range controls {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), s+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave15-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"default", "no_useless_default_assignment.a", "this.rules.byte(value.end)", "this.rules.byte(value.end) + 1"},
		{"find", "prefer_find.a", "'find'", "'filter'"},
		{"sort", "require_array_sort_compare.a", "(type.flags & (32 | 1024 | 4194304 | 8388608)) !== 0", "(type.flags & (32 | 1024 | 4194304 | 8388608)) === 0"},
	} {
		t.Run(change.name, func(t *testing.T) {
			previous := h.t
			h.t = t
			defer func() { h.t = previous }()
			scratch := filepath.Join(directory, change.name+"-source")
			if err := os.MkdirAll(scratch, 0755); err != nil {
				t.Fatal(err)
			}
			files, err := filepath.Glob(filepath.Join(repository, "stage1/cohere/typeaware/*"))
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range files {
				if filepath.Ext(file) != ".a" && filepath.Ext(file) != ".ts" {
					continue
				}
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				s := string(data)
				if filepath.Base(file) == change.file {
					if strings.Count(s, change.from) != 1 {
						t.Fatal("nonunique mutant")
					}
					s = strings.Replace(s, change.from, change.to, 1)
				}
				s = strings.ReplaceAll(s, "../../typescript/", filepath.Join(repository, "stage1/typescript")+"/")
				s = strings.ReplaceAll(s, "../lint/", filepath.Join(repository, "stage1/cohere/lint")+"/")
				h.write(filepath.Join(change.name+"-source", filepath.Base(file)), s)
			}
			mutant := h.build(stage0, change.name+"-mutant", filepath.Join(scratch, "wave_15_suite.a"), archive, false)
			got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
			if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
				t.Fatal("mutant survived")
			}
			t.Logf("%s: exit 0, independent Go bytes catch byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		})
	}
	for _, corpus := range []string{"repository", "compiler"} {
		manifest := os.Getenv("ADAMIC_WAVE15_" + strings.ToUpper(corpus) + "_MANIFEST")
		if manifest == "" {
			continue
		}
		config := filepath.Join(repository, "tsconfig.json")
		if corpus == "compiler" {
			config = filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json")
		}
		h.compare(corpus, oracle, binary, config, manifest)
		h.compare(corpus+"-asan", oracle, asan, config, manifest)
		goResult := h.must(corpus+"-timed-go", exec.Command(oracle, config, manifest))
		command := exec.Command(binary, config, manifest)
		command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		native := h.must(corpus+"-timed-native", command)
		if !bytes.Equal(goResult.stdout, native.stdout) {
			t.Fatal("timed bytes differ")
		}
		t.Logf("%s native=%s Go=%s; %s %s", corpus, native.elapsed, goResult.elapsed, native.stderr, goResult.stderr)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','declaration-contract'));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released handle: panic 70, invalid or released checker handle")
}
