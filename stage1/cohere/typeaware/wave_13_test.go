package typeaware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func wave13Controls() []string {
	return []string{
		"let x; console.log(x);",
		"let x; function f(x:number){x=1;}console.log(x);",
		"let x; {let x;x=1;console.log(x);}console.log(x);",
		"let x;function f(){x=1;} console.log(x);",
		"let x; x=undefined;console.log(x);let unused;let initialized=undefined;console.log(initialized);",
		"declare let ambient:number;console.log(ambient);namespace N{let x;console.log(x);}for(let y;;){console.log(y);break;}",
		"let x; ({x}= {x:1});console.log(x);let y;[...y]=[];console.log(y);",
		"let x;({files=x}={});console.log(x);",
		"let x;const obj={x};console.log(obj);",
		"try{}catch(e){throw new Error('m');}",
		"try{}catch(e){throw new Error;}",
		"try{}catch(e){throw new (Error);}",
		"try{}catch(e){throw new Error/* ( */();}",
		"try{}catch(e){throw (new (Error)('m'));}",
		"try{}catch(e){throw Error('m',{});}",
		"try{}catch(e){throw new Error('m',{foo:1,});}",
		"try{}catch(e){throw new Error('m',{cause:e});}",
		"try{}catch(e){throw new Error('m',{cause:e.message});}",
		"try{}catch(cause){throw new Error('m',{cause});}",
		"try{}catch(e){throw new Error('m',{cause});}",
		"try{}catch(e){const e=1;throw new Error('m',{cause:e});}",
		"try{}catch({message}){throw new Error(message);}",
		"try{}catch{throw new Error('m');}",
		"try{}catch(e){const f=()=>{throw new Error('m');};}",
		"function f(Error:any){try{}catch(e){throw new Error('m');}}",
		"try{}catch(e){throw new AggregateError;}",
		"try{}catch(e){throw new AggregateError([]);}",
		"try{}catch(e){throw new Error('m',{['cause']:e});}",
		"const cause='cause';try{}catch(e){throw new Error('m',{[cause]:e});}",
		"try{}catch(e){throw new Error('m',{[('cause')]:e});}",
		"try{}catch(e){throw new Error('m',{[`cause`]:e});}",
		"try{}catch(e){throw new Error('m',{cause:e,cause:0});}",
		"try{}catch(e){throw new Error('m',{cause(){return 0;}});}",
		"try{}catch(e){throw new Error('m',{get cause(){return 0;}});}",
		"try{}catch(e){throw new Error('m',{...opts});}",
		"try{}catch(e){throw new Error(...args);}",
		"try{}catch(e){throw Error?.('m');}",
		"interface T{x:number};export {T};",
		"interface T{x:number};interface U{y:string};const v=1;export {type U,T,v};",
		"interface T{x:number};export {T as '世界'};",
		"interface T{x:number};interface U{y:string};interface V{z:boolean};const v=1;export {T,U,V,v};",
		"export {T,v} from './helper.js';",
		"export /* * in comment */ * from './types_only.js';",
		"export * as ns from './types_only.js';",
		"export * from './helper.js';",
		"export * from './type_star.js';",
		"export {missing};",
		"/* 世界🌍 */\r\nlet é;console.log(é);try{}catch(漢){throw new Error('🌍');}\r\n",
	}
}

func wave13Mutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	paths, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*.a"))
	if err != nil {
		h.t.Fatal(err)
	}
	imports := regexp.MustCompile(`'\./([^']+\.ts)'`)
	for _, path := range paths {
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
		source = imports.ReplaceAllString(source, "'"+filepath.Join(h.repository, "stage1/cohere/typeaware")+"/$1'")
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		h.write(filepath.Join(name+"-source", filepath.Base(path)), source)
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_13_suite.a"), archive, false)
}

// Not parallel: native builds, sanitizer probes and timing observations share the machine.
func TestWave13AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE13_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := os.Getenv("ADAMIC_WAVE13_STAGE0")
	if stage0 == "" {
		stage0 = filepath.Join(directory, "adamic")
		h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	}
	archive := os.Getenv("ADAMIC_WAVE13_ARCHIVE")
	if archive == "" {
		archive = h.archive("checker", "", false)
	}
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_13_suite.a")
	binary := h.build(stage0, "wave13", entry, archive, false)
	oracle := volumeOracle(h, "wave13-oracle", "oracle_wave_13.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for i, source := range wave13Controls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	paths = append(paths, h.write("helper.ts", "export interface T{x:number};export const v=1;\n"), h.write("types_only.ts", "export interface T{x:number};\n"), h.write("type_star.ts", "export type * from './helper.js';\n"))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-unassigned-vars", "preserve-caught-error", "@typescript-eslint/consistent-type-exports"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave13-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"unassigned", "no_unassigned_vars.a", "if(read && !write)", "if(read || write)"},
		{"caught", "preserve_caught_error.a", "this.suggest(finding, end, end, `, ${cause}`);", "this.suggest(finding, end, end, `, ${cause} `);"},
		{"exports", "consistent_type_exports.a", "new Repair(position, position, ' type')", "new Repair(position, position, ' type ')"},
	} {
		mutant := wave13Mutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s: exit 0, empty stderr, only Go bytes catch offset %d", change.name, firstDifference(got.stdout, truth.stdout))
		if err := os.Remove(mutant); err != nil {
			t.Fatal(err)
		}
	}
	for _, change := range []struct{ name, file, from, to string }{
		{"export-chain-flags", "export_symbol_chain.go", "out.number(uint64(symbol.Flags))", "out.number(uint64(ast.SymbolFlagsProperty))"},
		{"export-module-lookup", "export_module_properties.go", "out.yes(checker.Checker_getPropertyOfType(c, t, property.Name) != nil)", "out.yes(checker.Checker_getPropertyOfType(c, t, property.Name) == nil)"},
	} {
		overlay := h.overlay(change.name, "bridge/tsgo/checker/"+change.file, change.from, change.to)
		mutantArchive := h.archive(change.name, overlay, false)
		mutant := h.build(stage0, change.name, entry, mutantArchive, false)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s question mutant survived", change.name)
		}
		t.Logf("%s: exit 0, empty stderr, only Go bytes catch offset %d", change.name, firstDifference(got.stdout, truth.stdout))
		os.Remove(mutantArchive)
		os.Remove(mutant)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','export-symbol-chain'));
`)
	probe := h.write("probe.a", "x;\n")
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
	t.Log("released registry mutant exits 0, required panic catches it")
	os.Remove(mutantArchive)
	os.Remove(mutant)
	measurements := map[string]any{}
	for _, population := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE13_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE13_COMPILER_MANIFEST")},
	} {
		if population.manifest == "" {
			t.Logf("%s not configured", population.name)
			continue
		}
		h.compare(population.name, oracle, binary, population.config, population.manifest)
		h.compare(population.name+"-asan", oracle, asan, population.config, population.manifest)
		want := h.must(population.name+"-timed-go", exec.Command(oracle, population.config, population.manifest))
		cmd := exec.Command(binary, population.config, population.manifest)
		cmd.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		got := h.must(population.name+"-timed-native", cmd)
		if !bytes.Equal(want.stdout, got.stdout) {
			t.Fatal("timed bytes differ")
		}
		measurements[population.name] = map[string]any{"go_seconds": want.elapsed.Seconds(), "native_seconds": got.elapsed.Seconds(), "bytes": len(got.stdout), "summary": summary(got.stdout), "go_stderr": string(want.stderr), "native_stderr": string(got.stderr)}
		t.Logf("%s whole process: native %.6fs Go %.6fs; %s", population.name, got.elapsed.Seconds(), want.elapsed.Seconds(), summary(got.stdout))
	}
	data, err := json.MarshalIndent(measurements, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	h.write("measurements.json", string(data)+"\n")
}
