package typeaware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: native builds and sanitizer subprocesses share this machine.
func TestWave30ProcessGraphAgreement(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE_30_PROCESS_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_30_graph_suite.a")
	binary := h.build(stage0, "process-graph", entry, archive, false)
	oracle := volumeOracle(h, "process-graph-oracle", "oracle_wave_30_graph.go")
	controls := []string{
		"console.log('x');process.exit();",
		"function f(a:boolean,b:boolean){if(a||b){console.log('x')}else{process.exit()};return;console.log('dead')}",
		"function f(a:boolean){try{console.log('x');a&&process.exit()}catch(e){console.log(e);process.exit()}finally{console.log('final')}}",
		"function f(){try{try{return console.log('x')}finally{console.log('inner')}}finally{process.exit()}}",
		"function f(){try{throw Error('x')}finally{console.log('final')}console.log('dead')}",
		"function f(a:any){while(a){console.log('x');if(a.b)break;continue}do{if(a)continue;console.log('y')}while(a);for(;;){console.log('x');break}}",
		"function f(a:any){for(let x=console.log('start');a;console.log('increment')){if(a.b)break;console.log('body')}for(;a;){continue}for(;;console.log('next')){break}}",
		"function f(a:any){for(const {x=console.log('default')} of a){console.log(x)}for(let x in a){if(x)continue;console.log(x)}}",
		"function f(a:any){switch(a){case console.log(1):console.log(2);case 3:break;default:console.log(4)}console.log(5)}",
		"function f(a:any){switch(a){default:console.log(1);case 2:console.log(3);break;case 4:console.log(5)}}",
		"function f(a:any){try{const [x=console.log('default'),...rest]=a;({[console.log('key')]:a.x=console.log('v')}=a)}catch(e){console.log(e)}}",
		"function f(a:any){a?.b(console.log('arg'));a?.[console.log('key')]?.(console.log('arg2'));(a?.b).c();a?.b!.c();a?.(a?.b).c}",
		"function f(a:any){if(a&&console.log('x'))process.exit();if(a??console.log('y'))process.exit();if(a?console.log('z'):console.log('q'))process.exit()}",
		"async function f(a:any){await console.log('a');for await(const x of a){console.log(x)}}function* g(){try{yield console.log('x')}finally{console.log('f')}}",
		"class C{[console.log('key')]=console.log('initializer');m(x=console.log('arg')){console.log(x)}static{console.log('static')}}",
		"function f(a:any){outer:inner:while(a){if(a.b)continue outer;break inner}label:{console.log('x');break label}console.log('after')}",
		"function f(){try{const x:typeof console=console.log('x');(console.log)<typeof console>(x)}catch(e){console.log(e)}}",
		"function f(){while(0x0n){console.log('zero')}while(0x10n){console.log('one');break}while(0){console.log('zero')}while(1e400){break}}",
	}
	paths := []string{}
	for at, source := range controls {
		paths = append(paths, h.write(fmt.Sprintf("graph-%03d.a", at), source+"\nexport {};\n"))
	}
	config := h.write("graph-tsconfig.json", `{"files":["graph-000.a"],"compilerOptions":{"target":"ESNext","module":"ESNext","strict":true,"lib":["ESNext","DOM"]}}`)
	manifest := h.write("graph.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.must("graph-go", exec.Command(oracle, config, manifest))
	got := h.must("graph-native", exec.Command(binary, config, manifest))
	if len(got.stderr) != 0 || !bytes.Equal(got.stdout, truth.stdout) {
		t.Fatalf("graph differs at byte %d", firstDifference(got.stdout, truth.stdout))
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "process-graph-asan", entry, sanitized, true)
	got = h.must("graph-asan", exec.Command(asan, config, manifest))
	if len(got.stderr) != 0 || !bytes.Equal(got.stdout, truth.stdout) {
		t.Fatal("sanitized graph differs")
	}
	mutant := wave30ProcessMutant(h, stage0, archive, "graph-throw-fork", "wave_30_graph_suite.a", "process_control_flow.a", "if(frame.forked) { return; }", "if(!frame.forked) { return; }")
	got = h.must("graph-mutant-run", exec.Command(mutant, config, manifest))
	if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
		t.Fatal("graph mutant survived")
	}
	t.Logf("graph mutant exits 0; Go catches byte %d", firstDifference(got.stdout, truth.stdout))
	t.Logf("%d exact graph bytes agree", len(truth.stdout))
}

// Not parallel: native builds and measured processes share this machine.
func TestWave30ProcessExitAgreement(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE_30_EXIT_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_30_exit_suite.a")
	binary := h.build(stage0, "process-exit", entry, archive, false)
	oracle := volumeOracle(h, "process-exit-oracle", "oracle_wave_30_exit.go")
	var inputs struct {
		Fixtures map[string]string
		Controls []string
	}
	data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/testdata/wave_30_exit_controls.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &inputs); err != nil {
		t.Fatal(err)
	}
	for name, source := range map[string]string{"node": inputs.Fixtures["NodeTypes"], "web-console": inputs.Fixtures["WebConsole"]} {
		path := h.write(name+".a", source)
		alias := filepath.Join(directory, name+".d.ts")
		if err = os.Symlink(filepath.Base(path), alias); err != nil {
			if target, e := os.Readlink(alias); e != nil || target != filepath.Base(path) {
				t.Fatal(err)
			}
		}
	}
	for name, source := range map[string]string{"Help": inputs.Fixtures["HelpModule"], "ScriptHelp": inputs.Fixtures["Script"]} {
		path := h.write(name+".a", source)
		alias := filepath.Join(directory, name+".ts")
		if err = os.Symlink(filepath.Base(path), alias); err != nil {
			if target, e := os.Readlink(alias); e != nil || target != filepath.Base(path) {
				t.Fatal(err)
			}
		}
	}
	paths := []string{}
	for at, source := range inputs.Controls {
		paths = append(paths, h.write(fmt.Sprintf("exit-%03d.a", at), source))
	}
	config := h.write("exit-tsconfig.json", `{"files":["node.d.ts","web-console.d.ts","ScriptHelp.ts","exit-000.a"],"compilerOptions":{"target":"ESNext","module":"ESNext","moduleDetection":"auto","strict":true,"lib":["ESNext","DOM"],"types":[]}}`)
	manifest := h.write("exit.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("exit-controls", oracle, binary, config, manifest)
	if !bytes.Contains(truth.stdout, []byte("\texitAfterOutput\t")) {
		t.Fatal("no process exit positive control")
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "process-exit-asan", entry, sanitized, true)
	h.compare("exit-controls-asan", oracle, asan, config, manifest)
	mutant := wave30ProcessMutant(h, stage0, archive, "exit-empty-write-state", "wave_30_exit_suite.a", "correctness_no_process_exit_after_output.a", "state.chains.length > 0", "state.chains.length === 0")
	got := h.must("exit-mutant-run", exec.Command(mutant, config, manifest))
	if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
		t.Fatal("exit mutant survived")
	}
	t.Logf("exit mutant exits 0; Go catches byte %d", firstDifference(got.stdout, truth.stdout))
	wave30ProcessCorpora(h, oracle, binary, asan)
	wave30ProcessReleasedHandles(h, stage0, archive, config, paths[0])

}

// Not parallel: native builds and measured subprocesses share this machine.
func TestWave30BlockingStreamsAgreement(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE_30_BLOCKING_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_30_blocking_suite.a")
	binary := h.build(stage0, "blocking-streams", entry, archive, false)
	oracle := volumeOracle(h, "blocking-streams-oracle", "oracle_wave_30_blocking.go")
	data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/testdata/wave_30_blocking_controls.json"))
	if err != nil {
		t.Fatal(err)
	}
	var controls []map[string]string
	if err = json.Unmarshal(data, &controls); err != nil {
		t.Fatal(err)
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "blocking-streams-asan", entry, sanitized, true)
	mutant := wave30ProcessMutant(h, stage0, archive, "blocking-empty-write-state", "wave_30_blocking_suite.a", "correctness_require_blocking_standard_streams.a", "state.chains.length > 0", "state.chains.length === 0")
	mutantCaught := false
	positives := 0
	for at, files := range controls {
		root := filepath.Join(directory, fmt.Sprintf("case-%03d", at))
		paths := []string{}
		for name, source := range files {
			path := filepath.Join(root, strings.TrimPrefix(name, "/repository/"))
			authored := strings.TrimSuffix(path, ".ts") + ".a"
			if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(authored, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			if err = os.Symlink(filepath.Base(authored), path); err != nil {
				if target, e := os.Readlink(path); e != nil || target != filepath.Base(authored) {
					t.Fatal(err)
				}
			}
			paths = append(paths, path)
		}
		config := filepath.Join(root, "tsconfig.json")
		configData, err := json.Marshal(map[string]any{"files": paths, "compilerOptions": map[string]any{"target": "ESNext", "module": "ESNext", "moduleDetection": "force", "moduleResolution": "Bundler", "strict": true, "lib": []string{"ESNext", "DOM"}, "types": []string{}}})
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(config, configData, 0600); err != nil {
			t.Fatal(err)
		}
		manifest := h.write(fmt.Sprintf("blocking-%03d.manifest", at), filepath.Join(root, "modules/subject/Subject.ts")+"\n")
		truth := h.compare(fmt.Sprintf("blocking-%03d", at), oracle, binary, config, manifest)
		h.compare(fmt.Sprintf("blocking-%03d-asan", at), oracle, asan, config, manifest)
		if !mutantCaught {
			got := h.must(fmt.Sprintf("blocking-%03d-mutant", at), exec.Command(mutant, config, manifest))
			if len(got.stderr) != 0 {
				t.Fatal("mutant stderr")
			}
			if !bytes.Equal(got.stdout, truth.stdout) {
				mutantCaught = true
				t.Logf("blocking mutant exits 0; Go catches byte %d", firstDifference(got.stdout, truth.stdout))
			}
		}

		if bytes.Contains(truth.stdout, []byte("\texitBeforeBlockingStandardStreams\t")) {
			positives++
		}
	}
	if !mutantCaught {
		t.Fatal("blocking mutant survived")
	}
	wave30ProcessCorpora(h, oracle, binary, asan)
	if positives == 0 {
		t.Fatal("no blocking positive controls")
	}
	t.Logf("%d programs agree; %d report", len(controls), positives)
}

func wave30ProcessMutant(h *harness, stage0, archive, name, entry, file, from, to string) string {
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	for _, pattern := range []string{"*.a", "*.ts"} {
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
					h.t.Fatalf("nonunique %s mutant", name)
				}
				source = strings.Replace(source, from, to, 1)
			}
			source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
			source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
			if err = os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0600); err != nil {
				h.t.Fatal(err)
			}
		}
	}
	return h.build(stage0, name, filepath.Join(directory, entry), archive, false)
}
func wave30ProcessCorpora(h *harness, oracle, binary, asan string) {
	for _, corpus := range []struct{ name, config, manifest string }{{"repository", filepath.Join(h.repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE_30_PROCESS_REPOSITORY_MANIFEST")}, {"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE_30_PROCESS_COMPILER_MANIFEST")}} {
		if corpus.manifest == "" {
			continue
		}
		h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
		h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
	}
}
func wave30ProcessReleasedHandles(h *harness, stage0, archive, config, file string) {
	source := h.write("released-process.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,1,'Identifier',args[2]??''));`)
	binary := h.build(stage0, "released-process", source, archive, false)
	for _, question := range []string{"process-node-fields", "symbol-declaration-paths", "resolved-call-target", "program-module-resolution", "source-parse-context"} {
		got := h.run("released-"+question, exec.Command(binary, config, file, question))
		if err, ok := got.err.(*exec.ExitError); !ok || err.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
			h.t.Fatalf("released %s escaped: %v %s", question, got.err, got.stderr)
		}
		h.t.Logf("released %s refused with panic 70", question)
	}
}
