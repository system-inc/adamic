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

func wave26FourthControls() []string {
	controls := []string{
		"Function('x');",
		"new Function('x');",
		"(Function)('x');new ((Function))('x');",
		"Function.apply(null,[]);Function.bind(null,'x');Function.call(null,'x');",
		"Function['bind'](null,'x');Function[`call`](null,'x');",
		"(Function)['apply'](null,[]);",
		"(Function.bind(null,'x'))();",
		"new Function.call(null,'x');",
		"Function.other();Function[key]();",
		"(Function as any)('x');new (Function as any)('x');",
		"function f(Function:any){Function('x');new Function('x');Function.bind(null);}",
		"const Function=()=>0;Function();new Function();",
		"const alias=Function;alias('x');new alias('x');",
		"globalThis.Function('x');new globalThis.Function('x');",
		"Function?.('x');Function?.bind(null,'x');",
		"new Symbol('x');new BigInt(1);",
		"new (Symbol)('x');new ((BigInt))(1);",
		"new Symbol;new BigInt;",
		"Symbol('x');BigInt(1);",
		"function f(Symbol:any,BigInt:any){new Symbol();new BigInt();}",
		"const Symbol=class {};const BigInt=class {};new Symbol();new BigInt();",
		"new globalThis.Symbol('x');new globalThis.BigInt(1);",
		"new (Symbol as any)('x');new (BigInt as any)(1);",
		"new String('x');new Number(1);new Boolean(false);",
		"new ((String))('x');new (Number)(1);new (Boolean)(false);",
		"new String;new Number;new Boolean;",
		"String('x');Number(1);Boolean(false);",
		"function f(String:any,Number:any,Boolean:any){new String();new Number();new Boolean();}",
		"const String=class {};const Number=class {};const Boolean=class {};new String();new Number();new Boolean();",
		"new globalThis.String('x');new globalThis.Number(1);new globalThis.Boolean(false);",
		"new (String as any)('x');new (Number!)(1);",
		"const alias=String;new alias('x');",
		"const key='bind';Function[key](null,'x');Function['b'+'ind'](null,'x');",
		"new String(local);new Function(local);new Symbol(local);",
		"/* 世界 🌍 */\r\nnew String('x');new Symbol('x');Function.bind(null,'x');",
		"import {Function,String,Symbol} from './exports';Function('x');new String('x');new Symbol('x');",
	}
	for i := range controls {
		controls[i] += "\nexport {};\n"
	}
	return controls
}

func wave26FourthMutant(h *harness, stage0, archive, name, file, from, to string) string {
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	for _, extension := range []string{"*.ts", "*.a"} {
		paths, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware", extension))
		if err != nil {
			h.t.Fatal(err)
		}
		for _, path := range paths {
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
	return h.build(stage0, name, filepath.Join(directory, "wave_26_fourth.a"), archive, false)
}

// Not parallel: builds, sanitizers and cost observations share a machine.
func TestWave26FourthAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE26_FOURTH_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_26_fourth.a")
	binary := h.build(stage0, "wave26", entry, archive, false)
	oracle := volumeOracle(h, "wave26-oracle", "oracle_wave_26_fourth.go")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"jsx":"preserve","target":"ES2022","module":"ESNext","lib":["ES2022","DOM"],"noEmit":true},"files":["control-000.a","configured.d.ts"]}`)
	h.write("configured.d.ts", "")
	h.write("exports.d.ts", "export declare const Function:any,String:any,Symbol:any;\n")
	var paths []string
	for i, source := range wave26FourthControls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d%s", i, func() string {
			if strings.Contains(source, "<div") {
				return ".tsx"
			}
			return ".a"
		}()), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-new-func", "no-new-native-nonconstructor", "no-new-wrappers"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave26-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"function", "no_new_func.a", "['apply', 'bind', 'call']", "['apply', 'call']"},
		{"nonconstructor", "no_new_native_nonconstructor.a", "this.tree.add(callee,", "this.tree.add(id,"},
		{"wrapper", "no_new_wrappers.a", "this.tree.skip(this.constructors.callee(id))", "this.constructors.callee(id)"},
	} {
		mutant := wave26FourthMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		os.Remove(mutant)
	}
	for _, change := range []struct{ name, file, from, to string }{
		{"constructor", "bridge/tsgo/checker/constructor_expression.go", "expression := node.AsNewExpression().Expression", "expression := node.AsNewExpression().Expression; if args := node.Arguments(); len(args) > 0 { expression = args[0] }"},
		{"origin", "bridge/tsgo/checker/binding_origin.go", "out.yes(f.IsDeclarationFile)", "out.yes(false)"},
	} {
		overlay := h.overlay(change.name, change.file, change.from, change.to)
		mutatedArchive := h.archive(change.name, overlay, false)
		mutant := h.build(stage0, change.name+"-question", entry, mutatedArchive, false)
		got := h.must(change.name+"-question-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s question mutant survived", change.name)
		}
		t.Logf("%s question mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		os.Remove(mutatedArchive)
		os.Remove(mutant)
	}
	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE26_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE26_COMPILER_MANIFEST")},
	} {
		if corpus.manifest == "" {
			continue
		}
		h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
		h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		want := h.must(corpus.name+"-timed-go", exec.Command(oracle, corpus.config, corpus.manifest))
		native := exec.Command(binary, corpus.config, corpus.manifest)
		native.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		got := h.must(corpus.name+"-timed-native", native)
		if !bytes.Equal(want.stdout, got.stdout) {
			t.Fatal("timed output mismatch")
		}
		t.Logf("%s whole process native %s Go %s; native phases %s; Go phases %s", corpus.name, got.elapsed, want.elapsed, got.stderr, want.stderr)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);console.log(tsgoInspect(program,file,0,5,'NewExpression','constructor-expression'));tsgoRelease(program);
console.log(tsgoInspect(program,file,0,5,'NewExpression','constructor-expression'));`)
	probe := h.write("probe.a", "new X;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released handle: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released-registry mutant: exit 0, required panic 70 catches it")
}
