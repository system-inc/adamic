package typeaware

import (
	"bytes"
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

func wave11FourthControls() []string {
	return []string{
		`new Function('return 1');Function('return 2');Function.call(null,'x');Function.apply(null,['x']);Function.bind(null,'x')();new Function.call(null,'x');`,
		"(Function)('x');new ((Function))('x');(Function?.call)(null,'x');(Function)['bind'](null,'x');Function[`apply`](null,[]);Function[('call')](null,'x');",
		`Function.toString();Function.bind;const f=Function;new f('x');globalThis.Function('x');Function.call.call(null,'x');Function[key](null,'x');`,
		`function f(Function:any){new Function('x');Function.call(null,'x')}class Function{};new Function('x');`,
		`function f(){class Function{};new Function()}new Function('x');{const Function=()=>{};Function('x')}Function('x');`,
		`new Symbol('x');new BigInt(1);Symbol('x');BigInt(1);new (Symbol)('x');new ((BigInt))(1);new Symbol;new BigInt;`,
		`function f(Symbol:any,BigInt:any){new Symbol;new BigInt}class Symbol{};new Symbol;const BigInt=()=>1;new BigInt;`,
		`new String('x');new Number(0);new Boolean(false);new (String)('x');new ((Number))(0);new Boolean;String('x');Number(0);Boolean(false);`,
		`function f(String:any,Number:any,Boolean:any){new String;new Number;new Boolean}class String{};new String;`,
		`function f(){const Number=()=>1;new Number}new Number;{const Boolean=()=>false;new Boolean}new Boolean;`,
		`new (Function as any)('x');(Function as any)('x');Function!('x');new (String as any)('x');new (Symbol as any)('x');`,
		"/* 世界 🌍 */\r\nnew Function('x');\r\nnew Symbol;\r\nnew String('é');\r\n",
		`new \u0053ymbol;new \u0053tring;new \u0046unction('x');`,
		`declare function Function():void;Function();declare const Symbol:unknown;new Symbol;declare class String{};new String;`,
		`import Function from './missing';Function('x');import Symbol from './missing';new Symbol;import String from './missing';new String;`,
	}
}

// Read source inputs from the pinned upstream-derived fixture tables. Options
// remain the production defaults on both sides; expectations are never copied.
func wave11FourthUpstream(t *testing.T, repository string) []string {
	t.Helper()
	var sources []string
	for _, name := range []string{"no_new_func", "no_new_native_nonconstructor", "no_new_wrappers"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/core", name+"_test.go"), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		goast.Inspect(tree, func(node goast.Node) bool {
			table, ok := node.(*goast.CompositeLit)
			if !ok {
				return true
			}
			array, ok := table.Type.(*goast.ArrayType)
			if !ok {
				return true
			}
			fields, ok := array.Elt.(*goast.StructType)
			if !ok {
				return true
			}
			sourceIndex := -1
			offset := 0
			for _, field := range fields.Fields.List {
				for _, name := range field.Names {
					if name.Name == "sourceText" || name.Name == "source" {
						sourceIndex = offset
					}
					offset++
				}
			}
			if sourceIndex < 0 {
				return true
			}
			for _, element := range table.Elts {
				row, ok := element.(*goast.CompositeLit)
				if !ok || sourceIndex >= len(row.Elts) {
					continue
				}
				value := row.Elts[sourceIndex]
				for _, entry := range row.Elts {
					if pair, ok := entry.(*goast.KeyValueExpr); ok {
						if key, ok := pair.Key.(*goast.Ident); ok && (key.Name == "sourceText" || key.Name == "source") {
							value = pair.Value
						}
					}
				}
				literal, ok := value.(*goast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					continue
				}
				source, err := strconv.Unquote(literal.Value)
				if err != nil {
					t.Fatal(err)
				}
				sources = append(sources, source)
			}
			return true
		})
	}
	return sources
}

func wave11FourthSourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
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
	return h.build(stage0, name, filepath.Join(directory, "wave_11_fourth_suite.a"), archive, false)
}

// Not parallel: archive builds, sanitizer subprocesses and measurements share a machine.
func TestWave11FourthAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE_11_FOURTH_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_11_fourth_suite.a")
	binary := h.build(stage0, "wave-11-fourth", entry, archive, false)
	oracle := volumeOracle(h, "wave-11-fourth-oracle", "oracle_wave_11_fourth.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for i, source := range append(wave11FourthControls(), wave11FourthUpstream(t, repository)...) {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.ts", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-new-func", "no-new-native-nonconstructor", "no-new-wrappers"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control: %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave-11-fourth-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"func", "no_new_func.a", "this.rules.byte(node.end), '')", "this.rules.byte(node.end) + 1, '')"},
		{"nonconstructor", "no_new_native_nonconstructor.a", "this.rules.byte(target.end), '')", "this.rules.byte(target.end) + 1, '')"},
		{"wrappers", "no_new_wrappers.a", "this.rules.byte(node.end), '')", "this.rules.byte(node.end) + 1, '')"},
		{"global", "wave_11_fourth_support.a", "declaration !== undefined && declaration.declarationFile", "declaration !== undefined && !declaration.declarationFile"},
	} {
		mutant := wave11FourthSourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
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
console.log(tsgoInspect(program,file,0,1,'Identifier','declaration-lineage'));
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
