package wave28next

import (
	"bytes"
	"encoding/json"
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

// Not parallel: native builds, sanitizers and timings share this machine.
func TestContinuation(t *testing.T) {
	repository, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE28_NEXT_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave28_next/suite.a")
	binary := h.build(stage0, "next", entry, archive, false)
	virtual := filepath.Join(repository, "cohere/adamic_wave28_next.go")
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(repository, "stage1/cohere/typeaware/wave28_next/oracle.go.txt")}})
	if err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(directory, "oracle")
	build := exec.Command("go", "build", "-overlay", h.write("oracle-overlay.json", string(data)), "-o", oracle, virtual)
	build.Dir = filepath.Join(repository, "cohere")
	h.must("oracle-build", build)
	node := h.write("node.d.ts", fixtureDeclaration(t, repository, "correctness_no_process_exit_after_output_test.go", "correctnessNoProcessExitAfterOutputNodeTypes"))
	console := h.write("web-console.d.ts", fixtureDeclaration(t, repository, "correctness_no_process_exit_after_output_test.go", "correctnessNoProcessExitAfterOutputWebConsole"))
	timer := h.write("timers.d.ts", fixtureDeclaration(t, repository, "correctness_no_uncleared_race_timeout_test.go", "correctnessNoUnclearedRaceTimeoutNodeTimers"))
	config := h.write("tsconfig.json", fmt.Sprintf(`{"compilerOptions":{"target":"ES2022","module":"ESNext","strict":true,"lib":["ES2022","DOM"]},"files":[%q,%q,%q,%q]}`, filepath.Join(repository, "stage1/cohere/typeaware/testdata/fact_cost.ts"), node, console, timer))
	var paths []string
	for i, source := range continuationControls(t, repository) {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("valid", exec.Command(oracle, config, manifest, "--valid-sources"))
	manifest = h.write("valid-controls.manifest", string(valid.stdout))
	t.Logf("accepted controls %d", len(strings.Split(strings.TrimSpace(string(valid.stdout)), "\n")))
	truth := h.compare("controls", oracle, binary, config, manifest)
	_ = truth
	domConfig := h.write("dom-tsconfig.json", fmt.Sprintf(`{"compilerOptions":{"target":"ES2022","module":"ESNext","strict":true,"lib":["ES2022","DOM"]},"files":[%q,%q,%q]}`, filepath.Join(repository, "stage1/cohere/typeaware/testdata/fact_cost.ts"), node, console))
	timerManifest := h.write("timer-controls.manifest", strings.Join(append(append([]string{}, paths[:17]...), paths[len(paths)-9]), "\n")+"\n")
	dom := h.compare("dom-controls", oracle, binary, domConfig, timerManifest)
	for _, rule := range []string{"no-process-exit-after-output", "no-uncleared-race-timeout", "require-blocking-standard-streams"} {
		n := bytes.Count(dom.stdout, []byte("\tnexus/correctness-"+rule+"\t"))
		if n == 0 {
			t.Fatalf("no positive controls for %s", rule)
		}
		t.Logf("%s positive findings %d", rule, n)
	}
	h.continueChecks(stage0, entry, archive, oracle, binary, domConfig, timerManifest, dom.stdout)
	h.releasedChecks()
	h.compare("expanded-controls-final", oracle, binary, config, manifest)
	h.compare("expanded-controls-asan", oracle, filepath.Join(directory, "next-asan"), config, manifest)

}
func continuationControls(t *testing.T, repository string) []string {
	sources := []string{
		"console.log('x');process.exit(0);",
		"function f(){console.log('x');process.exit(0);}",
		"function f(){process.exit(0);console.log('x');process.exit(1);}",
		"function f(flag:boolean){if(flag)console.log('x');process.exit(0);}",
		"function f(){try{console.log('x');}catch{process.exit(0);}}",
		"function f(){console.log('x');try{unknown();}catch{process.exit(0);}}",
		"function show(){console.log('x');} show();process.exit(0);",
		"Promise.race([new Promise((_r,reject)=>{setTimeout(reject,10);})]);",
		"const timeout=new Promise((_r,reject)=>{const timer=setTimeout(reject,10);}); Promise.race([timeout]);",
		"Promise.race([new Promise((_r,reject)=>{let timer;timer=setTimeout(reject,10);})]);",
		"Promise.race([new Promise((_r,reject)=>{const timer=setTimeout(reject,10);console.log(timer);})]);",
		"Promise.race([new Promise((_r,reject)=>setTimeout(reject,10))]);",
		"Promise.race([new Promise((_r,reject)=>{setTimeout(reject,10).unref();})]);",
		"Promise.race([new Promise((_r,reject)=>{window.setTimeout(reject,10);})]);",
		"Promise.race([new Promise((_r,reject)=>{globalThis.setTimeout(reject,10);})]);",
		"function setTimeout(x:unknown,y:number){return 0;} Promise.race([new Promise((_r,reject)=>{setTimeout(reject,10);})]);",
		"const pool={race(x:unknown){}};pool.race([new Promise((_r,reject)=>{setTimeout(reject,10);})]);",
	}
	for _, name := range []string{"correctness_no_process_exit_after_output_test.go", "correctness_no_uncleared_race_timeout_test.go", "correctness_require_blocking_standard_streams_test.go"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/nexus", name), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		goast.Inspect(tree, func(node goast.Node) bool {
			var elements []goast.Expr
			if literal, ok := node.(*goast.CompositeLit); ok {
				if array, ok := literal.Type.(*goast.ArrayType); ok {
					if element, ok := array.Elt.(*goast.Ident); ok && element.Name == "string" {
						elements = literal.Elts
					}
				}
			}
			if call, ok := node.(*goast.CallExpr); ok {
				if id, ok := call.Fun.(*goast.Ident); ok && strings.HasSuffix(id.Name, "Source") {
					elements = call.Args
				}
			}
			var lines []string
			for _, element := range elements {
				literal, ok := element.(*goast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					return true
				}
				value, err := strconv.Unquote(literal.Value)
				if err != nil {
					t.Fatal(err)
				}
				lines = append(lines, value)
			}
			source := strings.Join(lines, "\n")
			if len(lines) > 1 && strings.ContainsAny(source, "{};") {
				sources = append(sources, source)
			}
			return true
		})
	}
	sources = append(sources,
		"async function main(){await work();console.log('x');process.exit(0);}main();",
		"#!/usr/bin/env node\nconsole.log('世界 🌍'); process.exit(1);",
		"function f(){outer:while(true){process.exit(0);console.log('x');continue outer;}}",
		"function f(){for(let i=0;i<3;i++){console.log('x');}process.exit(0);}",
		"function f(){console.log('x');try{return;}finally{process.exit(0);}}",
		"function f(){for(;;){process.exit(0);console.log('x');}}",
		"Promise.race([new Promise<never>((_r,reject)=>{const timer=setTimeout(reject,10); const held={timer};})]);",
		"Promise.race([new Promise<never>((_r,reject)=>{let timer; (timer)=setTimeout(reject,10);})]);",
		"Promise.race([new Promise<never>((_r,reject)=>{void setTimeout(reject,10);})]);",
	)
	return sources
}
func fixtureDeclaration(t *testing.T, repository, file, variable string) string {
	tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/nexus", file), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var result string
	goast.Inspect(tree, func(node goast.Node) bool {
		value, ok := node.(*goast.ValueSpec)
		if !ok || len(value.Names) != 1 || value.Names[0].Name != variable {
			return true
		}
		var lines []string
		goast.Inspect(value.Values[0], func(child goast.Node) bool {
			if literal, ok := child.(*goast.BasicLit); ok && literal.Kind == token.STRING {
				text, err := strconv.Unquote(literal.Value)
				if err != nil {
					t.Fatal(err)
				}
				lines = append(lines, text)
			}
			return true
		})
		// strings.Join's final separator is not part of the declaration.
		if len(lines) > 0 && lines[len(lines)-1] == "\n" {
			lines = lines[:len(lines)-1]
		}
		result = strings.Join(lines, "\n")
		return false
	})
	if result == "" {
		t.Fatalf("missing fixture %s", variable)
	}
	return result
}
