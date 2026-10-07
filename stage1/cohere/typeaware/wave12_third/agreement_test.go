package wave12third

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

func ownControls() []string {
	return []string{
		"new Symbol;new BigInt(1);new String('x');new Number(1);new Boolean(false);Function('return 1');",
		"new (Symbol)('x');new ((BigInt))(1);new (String)('x');new ((Boolean))(0);(Function)('x');",
		"Function['call'](null);Function[`bind`](null);(Function?.apply)(null,[]);Function?.();",
		"class Function{};Function.call(null);class Symbol{};new Symbol();class String{};new String();",
		"function f(Function:any,Symbol:any,String:any){Function.call(null);new Symbol;new String;}",
		"import {Function,Symbol,String} from './helper';Function();new Symbol();new String();",
		"import {Function,Symbol,String} from './ambient';Function();new Symbol();new String();",
		"export {}; /* 世界 🌍 */\r\nnew String('é');Function.bind(null,'é')();\r\n",
		"var String:any;new String('x');var Function:any;Function('x');",
		"Function.bind;Function.toString();Function[0]();globalThis.Function();new globalThis.String();",
		"Function[('call')](null);Function[((`bind`))](null);",
	}
}

func oracle(h *harness) string {
	virtual := filepath.Join(h.repository, "cohere/adamic_wave12_third_oracle.go")
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(h.repository, "stage1/cohere/typeaware/wave12_third/oracle.go")}})
	if err != nil {
		h.t.Fatal(err)
	}
	overlay := h.write("oracle-overlay.json", string(data))
	binary := filepath.Join(h.directory, "oracle")
	command := exec.Command("go", "build", "-overlay", overlay, "-o", binary, virtual)
	command.Dir = filepath.Join(h.repository, "cohere")
	h.must("oracle-build", command)
	return binary
}

// Not parallel: native archives, sanitizer subprocesses, and timings share resources.
func TestAgreement(t *testing.T) {
	repository, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE12_THIRD_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave12_third/suite.a")
	binary := h.build(stage0, "native", entry, archive, false)
	truth := oracle(h)
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","lib":["ES2022"],"noEmit":true}}`)
	modules := filepath.Join(directory, "modules")
	if err := os.MkdirAll(modules, 0755); err != nil {
		t.Fatal(err)
	}
	for name, source := range map[string]string{"helper.ts": "export class Function{};export class Symbol{};export class String{};", "ambient.d.ts": "export declare class Function{};export declare class Symbol{};export declare class String{};"} {
		if err := os.WriteFile(filepath.Join(modules, name), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	var paths []string
	for at, source := range append(ownControls(), referenceControls(t, repository)...) {
		path := filepath.Join(modules, fmt.Sprintf("control-%03d.a", at))
		if err := os.WriteFile(path, []byte("export {};\n"+source+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	want := h.compare("controls", truth, binary, config, manifest)
	for _, rule := range []string{"no-new-func", "no-new-native-nonconstructor", "no-new-wrappers"} {
		if !bytes.Contains(want.stdout, []byte(rule+"\t")) {
			t.Fatal("missing positive", rule)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "native-asan", entry, sanitized, true)
	h.compare("controls-asan", truth, asan, config, manifest)
	for _, mutant := range []struct{ file, from, to string }{
		{"no_new_func.a", "this.files.global(receiver)", "!this.files.global(receiver)"},
		{"no_new_native_nonconstructor.a", "this.files.global(callee)", "!this.files.global(callee)"},
		{"no_new_wrappers.a", "this.files.global(callee)", "!this.files.global(callee)"},
	} {
		own := filepath.Join(directory, mutant.file+"-source")
		if err := os.MkdirAll(own, 0755); err != nil {
			t.Fatal(err)
		}
		files, err := filepath.Glob(filepath.Join(repository, "stage1/cohere/typeaware/wave12_third/*.a"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range files {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			source := string(data)
			if filepath.Base(path) == mutant.file {
				if strings.Count(source, mutant.from) != 1 {
					t.Fatal("nonunique mutant", mutant.file)
				}
				source = strings.Replace(source, mutant.from, mutant.to, 1)
			}
			source = strings.ReplaceAll(source, "'../", "'"+filepath.Join(repository, "stage1/cohere/typeaware")+"/")
			source = strings.ReplaceAll(source, filepath.Join(repository, "stage1/cohere/typeaware")+"/../../typescript/", filepath.Join(repository, "stage1/typescript")+"/")
			if err := os.WriteFile(filepath.Join(own, filepath.Base(path)), []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
		}
		changed := h.build(stage0, mutant.file+"-mutant", filepath.Join(own, "suite.a"), archive, false)
		got := h.must(mutant.file+"-run", exec.Command(changed, config, manifest))
		if len(got.stderr) > 0 || bytes.Equal(got.stdout, want.stdout) {
			t.Fatal("mutant survived", mutant.file)
		}
		t.Logf("%s mutant exits 0, empty stderr, byte comparison catches byte %d", mutant.file, firstDifference(got.stdout, want.stdout))
		os.Remove(changed)
	}
	for _, mutation := range []struct{ name, file, from, to string }{
		{"declaration-file", "bridge/tsgo/checker/symbol_declaration_files.go", "out.yes(file.IsDeclarationFile)", "out.yes(false)"},
	} {
		overlay := h.overlay(mutation.name, mutation.file, mutation.from, mutation.to)
		mutatedArchive := h.archive(mutation.name, overlay, false)
		changed := h.build(stage0, mutation.name+"-native", entry, mutatedArchive, false)
		got := h.must(mutation.name+"-run", exec.Command(changed, config, manifest))
		if len(got.stderr) > 0 || bytes.Equal(got.stdout, want.stdout) {
			t.Fatal("question mutant survived", mutation.name)
		}
		t.Logf("%s raw question mutant exits 0, empty stderr, byte comparison catches byte %d", mutation.name, firstDifference(got.stdout, want.stdout))
		os.Remove(changed)
		os.Remove(mutatedArchive)
	}
	probe := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,1,'Identifier',args[2]??''));`)
	sample := h.write("probe.a", "x;\n")
	released := h.build(stage0, "released", probe, archive, false)
	for _, question := range []string{"symbol-declaration-files"} {
		got := h.run("released-"+question, exec.Command(released, config, sample, question))
		code, ok := got.err.(*exec.ExitError)
		if !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
			t.Fatalf("released %s: %v %s", question, got.err, got.stderr)
		}
	}
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "")
	retainedArchive := h.archive("released-registry", overlay, false)
	retained := h.build(stage0, "retained", probe, retainedArchive, false)
	h.must("retained-run", exec.Command(retained, config, sample, "symbol-declaration-files"))
	t.Log("new question rejects released handles with panic 70; registry retention mutant exits 0 and is caught")
	os.Remove(retained)
	os.Remove(retainedArchive)

	for _, corpus := range []string{"COMPILER", "REPOSITORY"} {
		roots := os.Getenv("ADAMIC_WAVE12_THIRD_" + corpus + "_MANIFEST")
		if roots == "" {
			continue
		}
		cfg := filepath.Join(repository, "tsconfig.json")
		if corpus == "COMPILER" {
			cfg = filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json")
		}
		h.compare(strings.ToLower(corpus), truth, binary, cfg, roots)
		h.compare(strings.ToLower(corpus)+"-asan", truth, asan, cfg, roots)
	}
}
