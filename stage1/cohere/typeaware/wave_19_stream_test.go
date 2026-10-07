package typeaware

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

func wave19LiteralLines(call *goast.CallExpr) (string, bool) {
	var lines []string
	for _, argument := range call.Args {
		literal, ok := argument.(*goast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			return "", false
		}
		value, err := strconv.Unquote(literal.Value)
		if err != nil {
			return "", false
		}
		lines = append(lines, value)
	}
	return strings.Join(lines, "\n") + "\n", len(lines) > 0
}

// Not parallel: native compiler and sanitizer builds share the machine.
func TestWave19ProcessPendingRegistration(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE19_STREAM_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	overlay := h.overlay("stream-registration", "bridge/tsgo/checker/facts.go", "switch mode {", "switch mode {\ncase \"declaration-ancestry\": return p.declarationAncestry(out,c,node,question)\ncase \"resolved-callee\": return p.resolvedCallee(out,c,node,question)\ncase \"program-modules\": return p.programModules(out,node,question)")
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("stream", overlay, false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_19_process.a")
	binary := h.build(stage0, "process", entry, archive, false)
	oracle := volumeOracle(h, "process-oracle", "oracle_wave_19_process.go")
	tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output_test.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	prelude := ""
	goast.Inspect(tree, func(node goast.Node) bool {
		spec, ok := node.(*goast.ValueSpec)
		if !ok || len(spec.Names) != 1 || len(spec.Values) != 1 {
			return true
		}
		call, ok := spec.Values[0].(*goast.CallExpr)
		if !ok {
			return true
		}
		value, ok := wave19LiteralLines(call)
		if !ok {
			return true
		}
		switch spec.Names[0].Name {
		case "correctnessNoProcessExitAfterOutputNodeTypes":
			h.write("node.d.ts", value)
		case "correctnessNoProcessExitAfterOutputWebConsole":
			h.write("web-console.d.ts", value)
		case "correctnessNoProcessExitAfterOutputPrelude":
			prelude = value
		}
		return true
	})
	if prelude == "" {
		t.Fatal("missing production prelude")
	}
	var paths []string
	goast.Inspect(tree, func(node goast.Node) bool {
		call, ok := node.(*goast.CallExpr)
		if !ok {
			return true
		}
		name, ok := call.Fun.(*goast.Ident)
		if !ok || name.Name != "correctnessNoProcessExitAfterOutputSource" {
			return true
		}
		value, ok := wave19LiteralLines(call)
		if !ok {
			return true
		}
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", len(paths)), prelude+value))
		return true
	})
	goast.Inspect(tree, func(node goast.Node) bool {
		row, ok := node.(*goast.CompositeLit)
		if !ok || len(row.Elts) < 2 || len(row.Elts) > 3 {
			return true
		}
		label, ok := row.Elts[0].(*goast.BasicLit)
		if !ok || label.Kind != token.STRING {
			return true
		}
		lines, ok := row.Elts[1].(*goast.CompositeLit)
		if !ok {
			return true
		}
		call := &goast.CallExpr{Args: lines.Elts}
		value, ok := wave19LiteralLines(call)
		if ok {
			paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", len(paths)), prelude+value))
		}
		return true
	})
	for _, body := range []string{"console.log(output);process.exit(0);", "process.exit(0);console.log(output);process.exit(1);", "try{console.log(output);}catch{process.exit(1);}", "console.log(output);try{use(output);}catch{process.exit(1);}", "while(flag){process.exit(0);console.log(output);}", "while(flag){if(flag)process.exit(0);console.log(output);}", "console.log(process.exit(0));process.exit(1);", "try{console.log(output);return;}finally{process.exit(0);}", "for(;;){console.log(output);break;}process.exit(0);", "const console={log(x:unknown){}};console.log(output);process.exit(0);"} {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", len(paths)), prelude+"export function probe(){"+body+"}"))
	}
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022","DOM"]},"files":["control-000.a","node.d.ts","web-console.d.ts"]}`)
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	t.Logf("%d controls", len(paths))
	truth := h.compare("process-controls", oracle, binary, config, manifest)
	if !bytes.Contains(truth.stdout, []byte("exitAfterOutput")) {
		t.Fatal("missing positive")
	}
	if os.Getenv("ADAMIC_WAVE19_STREAM_FAST") == "1" {
		return
	}
	wave19StreamMutant(h, stage0, archive, "wave_19_process.a", "no_process_exit_after_output.a", "if(reported.has(exit))", "if(false)", config, manifest, truth)
	sanitized := h.archive("stream-asan", overlay, true)
	asan := h.build(stage0, "process-asan", entry, sanitized, true)
	h.compare("process-controls-asan", oracle, asan, config, manifest)
	for _, population := range []struct{ name, root, config string }{{"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"}, {"repository", repository, "tsconfig.json"}} {
		if population.root == "" {
			t.Log("compiler corpus absent")
			continue
		}
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage", population.name+".manifest"))
		if err != nil {
			t.Fatal(err)
		}
		var paths []string
		for _, path := range strings.Split(string(data), "\n") {
			if path != "" {
				paths = append(paths, filepath.Join(population.root, path))
			}
		}
		manifest := h.write(population.name+".manifest", strings.Join(paths, "\n")+"\n")
		config := filepath.Join(population.root, population.config)
		h.compare(population.name, oracle, binary, config, manifest)
		h.compare(population.name+"-asan", oracle, asan, config, manifest)
	}
}

// Not parallel: archive and native builds share the machine with the process test.
func TestWave19BlockingPendingRegistration(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE19_BLOCKING_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	overlay := h.overlay("stream-registration", "bridge/tsgo/checker/facts.go", "switch mode {", "switch mode {\ncase \"declaration-ancestry\": return p.declarationAncestry(out,c,node,question)\ncase \"resolved-callee\": return p.resolvedCallee(out,c,node,question)\ncase \"program-modules\": return p.programModules(out,node,question)")
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("stream", overlay, false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_19_blocking.a")
	binary := h.build(stage0, "blocking", entry, archive, false)
	oracle := volumeOracle(h, "blocking-oracle", "oracle_wave_19_blocking.go")
	tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output_test.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	goast.Inspect(tree, func(node goast.Node) bool {
		spec, ok := node.(*goast.ValueSpec)
		if !ok || len(spec.Names) != 1 || len(spec.Values) != 1 {
			return true
		}
		call, ok := spec.Values[0].(*goast.CallExpr)
		if !ok {
			return true
		}
		value, ok := wave19LiteralLines(call)
		if !ok {
			return true
		}
		switch spec.Names[0].Name {
		case "correctnessNoProcessExitAfterOutputNodeTypes":
			h.write("node.d.ts", value)
		case "correctnessNoProcessExitAfterOutputWebConsole":
			h.write("web-console.d.ts", value)
		}
		return true
	})
	h.write("prelude.d.ts", "declare const flag:boolean;declare const output:string;declare const rows:string[];declare function use(x:unknown):void;declare function work():Promise<void>;declare function require(name:string):unknown;")
	// These .ts files are generated TypeScript oracle inputs, never Adamic modules.
	if err := os.MkdirAll(filepath.Join(directory, "source/system"), 0755); err != nil {
		t.Fatal(err)
	}
	h.write("source/system/StandardStreams.ts", "export function blockStandardStreams():void {Reflect.get(process.stdout,'_handle');}")
	var paths []string
	bodies := []string{
		"console.log(output);process.exit(0);",
		"process.exit(0);console.log(output);process.exit(1);",
		"if(flag){console.error(output);process.exit(1);}console.log(output);process.exit(0);",
		"function main(){console.log(output);process.exit(0);}main();",
		"function unused(){console.log(output);process.exit(0);}",
		"const main=()=>{console.log(output);process.exit(0);};main();",
		"function onError(){console.error(output);process.exit(1);}process.on('error',onError);",
		"function* generate(){console.log(output);process.exit(0);}generate();",
		"process.stdout.write(output,()=>process.exit(0));",
		"try{console.log(output);}catch{process.exit(1);}",
		"console.log(output);try{use(output);}catch{process.exit(1);}",
		"for(const row of rows){console.log(row);}process.exit(0);",
		"while(flag){if(flag)process.exit(0);console.log(output);}",
		"try{console.log(output);throw Error();}finally{process.exit(0);}",
		"async function main(){console.log(output);await work();process.exit(0);}main();",
		"function helper(){console.log(output);}helper();process.exit(0);",
		"declare const resourceValue:{[Symbol.asyncDispose]():Promise<void>};async function main(){console.log(output);await /* disposal */ using resource=resourceValue;process.exit(0);}main();",
	}
	for _, body := range bodies {
		for _, ordered := range []bool{false, true} {
			prefix := "export {};"
			if ordered {
				prefix = "import {blockStandardStreams} from './source/system/StandardStreams';"
			}
			paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", len(paths)), prefix+body))
		}
	}
	for _, body := range []string{"blockStandardStreams();console.log(output);process.exit(0);", "console.log(output);process.exit(0);blockStandardStreams();", "function main(){console.log(output);process.exit(0);blockStandardStreams();}main();", "function main(){blockStandardStreams();console.log(output);process.exit(0);}main();", "if(flag){console.error(output);process.exit(1);}blockStandardStreams();", "function recursively(){if(flag)recursively();else blockStandardStreams();}recursively();console.log(output);process.exit(0);", "console.log(output);use(()=>blockStandardStreams());process.exit(0);", "console.log(output);(()=>{})(blockStandardStreams());process.exit(0);"} {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", len(paths)), "import {blockStandardStreams} from './source/system/StandardStreams';"+body))
	}
	paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", len(paths)), "#!/usr/bin/env tsx\nconsole.log(output);process.exit(0);"))
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022","DOM"]},"files":["control-000.a","node.d.ts","web-console.d.ts","prelude.d.ts"]}`)
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	t.Logf("%d controls", len(paths))
	truth := h.compare("blocking-controls", oracle, binary, config, manifest)
	if !bytes.Contains(truth.stdout, []byte("exitBeforeBlockingStandardStreams")) {
		t.Fatal("missing positive")
	}
	if os.Getenv("ADAMIC_WAVE19_STREAM_FAST") == "1" {
		return
	}
	wave19StreamMutant(h, stage0, archive, "wave_19_blocking.a", "require_blocking_standard_streams.a", "if(found.length === 0)", "if(found.length >= 0)", config, manifest, truth)
	wave19StreamMutant(h, stage0, archive, "wave_19_blocking.a", "require_blocking_standard_streams.a", "semantic === '6'", "semantic === 'disabled'", config, manifest, truth)
	sanitized := h.archive("stream-asan", overlay, true)
	asan := h.build(stage0, "blocking-asan", entry, sanitized, true)
	h.compare("blocking-controls-asan", oracle, asan, config, manifest)
	for _, population := range []struct{ name, root, config string }{{"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"}, {"repository", repository, "tsconfig.json"}} {
		if population.root == "" {
			t.Log("compiler corpus absent")
			continue
		}
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage", population.name+".manifest"))
		if err != nil {
			t.Fatal(err)
		}
		var paths []string
		for _, path := range strings.Split(string(data), "\n") {
			if path != "" {
				paths = append(paths, filepath.Join(population.root, path))
			}
		}
		manifest := h.write(population.name+".manifest", strings.Join(paths, "\n")+"\n")
		config := filepath.Join(population.root, population.config)
		h.compare(population.name, oracle, binary, config, manifest)
		h.compare(population.name+"-asan", oracle, asan, config, manifest)
	}
}

func wave19StreamMutant(h *harness, stage0, archive, entry, ruleFile, from, to, config, manifest string, truth result) {
	h.t.Helper()
	directory := filepath.Join(h.directory, "mutant-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	for _, name := range []string{entry, "no_process_exit_after_output.a", "require_blocking_standard_streams.a", "stream_graph.a", "declaration_ancestry.a", "resolved_callee.a", "program_modules.a"} {
		data, err := os.ReadFile(filepath.Join(h.repository, "stage1/cohere/typeaware", name))
		if err != nil {
			h.t.Fatal(err)
		}
		source := string(data)
		for _, helper := range []string{"rules.ts", "bindings.ts", "diagnostic.ts", "frames.ts", "unary_minus.ts"} {
			source = strings.ReplaceAll(source, "'./"+helper+"'", "'"+filepath.ToSlash(filepath.Join(h.repository, "stage1/cohere/typeaware", helper))+"'")
		}
		source = strings.ReplaceAll(source, "'../../typescript/", "'"+filepath.ToSlash(filepath.Join(h.repository, "stage1/typescript"))+"/")
		if name == ruleFile {
			if strings.Count(source, from) != 1 {
				h.t.Fatal("mutant site changed")
			}
			source = strings.Replace(source, from, to, 1)
		}
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0644); err != nil {
			h.t.Fatal(err)
		}
	}
	mutant := h.build(stage0, "rule-mutant", filepath.Join(directory, entry), archive, false)
	got := h.must("rule-mutant-run", exec.Command(mutant, config, manifest))
	if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
		h.t.Fatal("mutant survived")
	}
	h.t.Logf("%s mutant exit 0, empty stderr, only Go bytes catch byte %d", ruleFile, firstDifference(got.stdout, truth.stdout))
}

// Not parallel: builds normal, sanitized and deliberately broken registry archives.
func TestWave19StreamReleasedHandles(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE19_RELEASE_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	overlay := h.overlay("stream-registration", "bridge/tsgo/checker/facts.go", "switch mode {", "switch mode {\ncase \"declaration-ancestry\": return p.declarationAncestry(out,c,node,question)\ncase \"resolved-callee\": return p.resolvedCallee(out,c,node,question)\ncase \"program-modules\": return p.programModules(out,node,question)")
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("stream", overlay, false)
	sanitized := h.archive("stream-asan", overlay, true)
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022"},"files":["probe.a"]}`)
	probe := h.write("probe.a", "work();")
	entry := h.write("released.a", `import {panic,programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const path=args[1]??panic('path');const program=tsgoProgram(args[0]??panic('config'),[path]);tsgoRelease(program);console.log(tsgoInspect(program,path,0,Number.parseInt(args[4]??'4',10),args[3]??'Identifier',args[2]??'declaration-ancestry'));`)
	normal := h.build(stage0, "released", entry, archive, false)
	asan := h.build(stage0, "released-asan", entry, sanitized, true)
	questions := []struct{ question, kind, end string }{{"declaration-ancestry\nalias", "Identifier", "4"}, {"resolved-callee", "CallExpression", "6"}, {"program-modules", "SourceFile", "7"}}
	for _, binary := range []string{normal, asan} {
		for _, question := range questions {
			got := h.run("released-run", exec.Command(binary, config, probe, question.question, question.kind, question.end))
			if got.err == nil || !strings.Contains(string(got.stderr), "invalid or released checker handle") {
				t.Fatal("released handle accepted")
			}
			if bytes.Contains(got.stderr, []byte("ERROR: AddressSanitizer")) || bytes.Contains(got.stderr, []byte("runtime error:")) || bytes.Contains(got.stderr, []byte("LeakSanitizer")) {
				t.Fatal("sanitizer failed")
			}
		}
	}
	t.Log("all three questions refuse released handles in normal and ASan/UBSan/LSan builds")
	registryOverlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant: retain the released handle.")
	var combined, registry map[string]map[string]string
	data, err := os.ReadFile(overlay)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &combined); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(registryOverlay)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &registry); err != nil {
		t.Fatal(err)
	}
	for original, replacement := range registry["Replace"] {
		combined["Replace"][original] = replacement
	}
	data, err = json.Marshal(combined)
	if err != nil {
		t.Fatal(err)
	}
	combinedOverlay := h.write("combined-release.json", string(data))
	mutantArchive := h.archive("released-mutant-checker", combinedOverlay, false)
	mutant := h.build(stage0, "released-mutant", entry, mutantArchive, false)
	for _, question := range questions {
		got := h.must("released-mutant-run", exec.Command(mutant, config, probe, question.question, question.kind, question.end))
		if len(got.stderr) != 0 {
			t.Fatal("registry mutant failed outside the refusal check")
		}
	}
	t.Log("retained-registry mutant exits 0 with empty stderr; all three required-panic checks kill it")
	// The real production archive still lacks these dispatch lines.
	unregisteredArchive := h.archive("unregistered-checker", "", false)
	source, err := os.ReadFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	liveEntry := h.write("live.a", strings.Replace(string(source), "tsgoRelease(program);", "", 1)+"tsgoRelease(program);")
	unregistered := h.build(stage0, "unregistered", liveEntry, unregisteredArchive, false)
	for _, question := range questions {
		got := h.run("unregistered-run", exec.Command(unregistered, config, probe, question.question, question.kind, question.end))
		exit, ok := got.err.(*exec.ExitError)
		if !ok || exit.ExitCode() != 70 || !strings.Contains(string(got.stderr), "unsupported checker question: "+strings.Split(question.question, "\n")[0]) {
			t.Fatalf("pending registration was not explicitly refused: %s", got.stderr)
		}
	}
	t.Log("unmodified production archive refuses all three unregistered questions with panic 70")
}
