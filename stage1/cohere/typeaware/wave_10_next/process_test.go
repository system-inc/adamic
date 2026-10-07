package wave10next_test

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

func processSources(t *testing.T, repository string) (map[string]string, []string) {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output_test.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	constants := map[string]string{}
	for _, declaration := range file.Decls {
		general, ok := declaration.(*goast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range general.Specs {
			value, ok := spec.(*goast.ValueSpec)
			if !ok || len(value.Names) != 1 || len(value.Values) != 1 {
				continue
			}
			call, ok := value.Values[0].(*goast.CallExpr)
			if !ok {
				continue
			}
			var lines []string
			all := true
			for _, argument := range call.Args {
				literal, ok := argument.(*goast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					all = false
					break
				}
				line, err := strconv.Unquote(literal.Value)
				if err != nil {
					t.Fatal(err)
				}
				lines = append(lines, line)
			}
			if all {
				constants[value.Names[0].Name] = strings.Join(lines, "\n") + "\n"
			}
		}
	}
	var sources []string
	goast.Inspect(file, func(node goast.Node) bool {
		literal, ok := node.(*goast.CompositeLit)
		if !ok {
			return true
		}
		array, ok := literal.Type.(*goast.ArrayType)
		if !ok {
			return true
		}
		element, ok := array.Elt.(*goast.Ident)
		if !ok || element.Name != "string" {
			return true
		}
		var lines []string
		for _, element := range literal.Elts {
			literal, ok := element.(*goast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			text, err := strconv.Unquote(literal.Value)
			if err != nil {
				t.Fatal(err)
			}
			lines = append(lines, text)
		}
		source := strings.Join(lines, "\n")
		if len(lines) > 1 && strings.Contains(source, "exit(") {
			sources = append(sources, source)
		}
		return true
	})
	sources = append(sources,
		"console.log('a');if(flag){process.exit(0)}else{process.exit(1)}",
		"while(true){process.exit(0);console.log('dead')}process.exit(1);",
		"for(;;){if(flag)break;console.log('write')}process.exit(0);",
		"for(;flag;){console.log('write');break;}process.exit(0);",
		"for(let i=0;;i++){console.log('write');break;}process.exit(0);",
		"outer:while(flag){console.log('write');continue outer;}process.exit(0);",
		"do{console.log('write');continue;}while(flag);process.exit(0);",
		"console.log('a');try{returner()}catch({code=process.exit(1)}){process.exit(2)}function returner(){throw new Error('a')}",
		"function f(){console.log('a');try{return 1;}finally{process.exit(1)}}f();",
		"function f(){console.log('a');try{try{return 1;}finally{use(1)}}finally{process.exit(1)}}f();",
		"function f(){console.log('a');try{throw new Error('a')}finally{process.exit(1)}}f();",
		"if(flag&&console.log('a'))process.exit(0);else process.exit(1);",
		"if(flag||console.log('a'))process.exit(0);else process.exit(1);",
		"console.log?.(process.exit(1));process.exit(2);",
		"const {a=console.log('a')}={};process.exit(0);",
		"class A{[console.log('name')](){process.exit(0)}field=(console.log('f'),process.exit(1));static{console.log('s');process.exit(2)}}",
		"/* 世界 🌍 */\r\nconsole.log('漢');process.exit(0);\r\n",
	)
	return constants, sources
}

func TestProcessNativeAgreement(t *testing.T) {
	if os.Getenv("ADAMIC_WAVE10_PROCESS_VALIDATE") == "" {
		t.Skip("set ADAMIC_WAVE10_PROCESS_VALIDATE for native checks")
	}
	repository, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE10_PROCESS_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_10_next/process_suite.a")
	binary := filepath.Join(directory, "process")
	virtual := filepath.Join(repository, "cohere/adamic_wave10_process_oracle.go")
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(repository, "stage1/cohere/typeaware/wave_10_next/testdata/oracle_process.go")}})
	if err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(directory, "oracle")
	cmd := exec.Command("go", "build", "-overlay", h.write("oracle.json", string(data)), "-o", oracle, virtual)
	cmd.Dir = filepath.Join(repository, "cohere")
	h.must("oracle-build", cmd)
	constants, sources := processSources(t, repository)
	h.write("node.d.ts", constants["correctnessNoProcessExitAfterOutputNodeTypes"])
	h.write("web-console.d.ts", constants["correctnessNoProcessExitAfterOutputWebConsole"])
	h.write("Help.ts", constants["correctnessNoProcessExitAfterOutputHelpModule"])
	h.write("ScriptHelp.d.ts", constants["correctnessNoProcessExitAfterOutputScript"])
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"],"moduleDetection":"auto","types":[]},"files":["node.d.ts","web-console.d.ts","ScriptHelp.d.ts"]}`)
	prelude := constants["correctnessNoProcessExitAfterOutputPrelude"] + "declare function returner():void;\n"
	var paths []string
	for i, source := range sources {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), prelude+source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	h.must("process", exec.Command("timeout", "--signal=QUIT", "45", stage0, "build", entry, "-o", binary, "--tsgo", archive))
	truth := h.compare("controls", oracle, binary, config, manifest)
	if !bytes.Contains(truth.stdout, []byte("\texitAfterOutput\t")) {
		t.Fatal("no process positive finding")
	}
	t.Logf("%d controls", len(paths))
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "process-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE10_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE10_COMPILER_MANIFEST")},
	} {
		if corpus.manifest != "" {
			h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
			h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		}
	}
}
