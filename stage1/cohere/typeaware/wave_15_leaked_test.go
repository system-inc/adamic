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
func TestWave15LeakedAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE15_LEAKED_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_15_leaked_suite.a")
	binary := h.build(stage0, "wave15", entry, archive, false)
	oracle := volumeOracle(h, "wave15-oracle", "oracle_wave_15_leaked.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	prelude := `declare const count:number;declare const optional:number|undefined;declare const big:bigint;declare const flag:boolean;declare const label:string;declare const mixed:string|number;declare const loose:any;declare const opaque:unknown;declare const zero:0|1;declare const nonzero:1|2;type Branded=number&{readonly unit:'pixels'};declare const width:Branded;enum Level{Off,Low,High};enum Size{Small=1,Large=2};declare const level:Level;declare const size:Size;`
	controls := []string{
		"count && 'child';optional && 'child';big && 'child';",
		"zero && 'child';nonzero && 'child';0n && 'child';1n && 'child';",
		"flag && count && 'child'; count && optional && 'child';",
		"(flag?count:false)&&'child'; (flag||count)&&'child'; (count||flag)&&'child';",
		"flag?(count&&'child'):null;label||(count&&'child');(count&&label)||'fallback';",
		"(optional??count)&&'child';(count&&'child')??(big&&'child');",
		"mixed && 'child';label && 'child';loose && 'child';opaque && 'child';",
		"width && 'child';level && 'child';size && 'child';",
		"function f<T extends number>(value:T){value&&'child';}function g<T>(value:T){value&&'child';}",
		"count>0&&'child';!!count&&'child';flag&&count;optional!==undefined&&optional&&'child';",
		"/* attribute */ count&&'child';",
		"/* spread */ count&&'child';",
		"/* outside */ count&&'child';",
		"/* fragment */ count&&'child';",
		"/* 世界 🌍 */\r\n(count)&&'child';\r\n",
	}
	paths := []string{}
	for i, s := range controls {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), prelude+s+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	compareControls := func(name, executable string) result {
		want := h.must(name+"-go", exec.Command(oracle, config, manifest, "--render-controls"))
		got := h.must(name+"-native", exec.Command(executable, config, manifest, "--render-controls"))
		if len(got.stderr) != 0 || !bytes.Equal(want.stdout, got.stdout) {
			t.Fatalf("%s differs at byte %d: %s", name, firstDifference(want.stdout, got.stdout), got.stderr)
		}
		t.Logf("%s: %d identical bytes; native=%s Go=%s", name, len(got.stdout), got.elapsed, want.elapsed)
		return want
	}
	truth := compareControls("controls", binary)
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave15-asan", entry, sanitized, true)
	compareControls("controls-asan", asan)
	for _, change := range []struct{ name, file, from, to string }{
		{"leaked", "no_leaked_number_render.a", "return false;", "return true;"},
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
			mutant := h.build(stage0, change.name+"-mutant", filepath.Join(scratch, "wave_15_leaked_suite.a"), archive, false)
			got := h.must(change.name+"-run", exec.Command(mutant, config, manifest, "--render-controls"))
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
console.log(tsgoInspect(program,file,0,1,'Identifier','numeric-literal'));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released handle: panic 70, invalid or released checker handle")
}
