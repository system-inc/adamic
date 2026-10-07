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

// Take the production rule's table controls verbatim, with its Node declarations.
func wave29Controls(t *testing.T, repository string) (string, []string) {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/nexus/concurrency_no_check_then_write_test.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	types := ""
	var controls []string
	goast.Inspect(file, func(node goast.Node) bool {
		if declaration, ok := node.(*goast.ValueSpec); ok && len(declaration.Names) == 1 && declaration.Names[0].Name == "concurrencyNoCheckThenWriteNodeTypes" {
			types, err = strconv.Unquote(declaration.Values[0].(*goast.BasicLit).Value)
			if err != nil {
				t.Fatal(err)
			}
		}
		literal, ok := node.(*goast.CompositeLit)
		if !ok {
			return true
		}
		array, ok := literal.Type.(*goast.ArrayType)
		if !ok {
			return true
		}
		element, ok := array.Elt.(*goast.Ident)
		if !ok || element.Name != "string" || array.Len != nil {
			return true
		}
		var lines []string
		for _, item := range literal.Elts {
			text, ok := item.(*goast.BasicLit)
			if !ok || text.Kind != token.STRING {
				return true
			}
			value, e := strconv.Unquote(text.Value)
			if e != nil {
				t.Fatal(e)
			}
			lines = append(lines, value)
		}
		if len(lines) < 2 || strings.HasPrefix(strings.TrimSpace(lines[0]), "import ") {
			return true
		}
		header := "import * as NodeFileSystem from 'node:fs';\nimport * as NodeFileSystemPromises from 'node:fs/promises';\nimport {access,writeFile} from 'node:fs/promises';\nimport * as NodePath from 'node:path';\ndeclare function delay(milliseconds:number):Promise<void>;\ndeclare function later(callback:()=>void):void;\nexport async function save(directory:string,bytes:Uint8Array,options:{flag:string},items:string[]):Promise<void>{let index=0;\n"
		controls = append(controls, header+strings.Join(lines, "\n")+"\n}\n")
		return true
	})
	if types == "" || len(controls) != 31 {
		t.Fatalf("control extraction: types=%d cases=%d, want 31", len(types), len(controls))
	}
	return types, controls
}

func wave29Mutant(h *harness, stage0, archive, name, file, from, to string) string {
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	for _, extension := range []string{"*.a"} {
		files, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware", extension))
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
					h.t.Fatalf("nonunique %s mutation", name)
				}
				source = strings.Replace(source, from, to, 1)
			}
			legacy, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*.ts"))
			if err != nil {
				h.t.Fatal(err)
			}
			for _, dependency := range legacy {
				source = strings.ReplaceAll(source, "'./"+filepath.Base(dependency)+"'", "'"+dependency+"'")
			}
			source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
			source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
			if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0644); err != nil {
				h.t.Fatal(err)
			}
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_29.a"), archive, false)
}

// Not parallel: C archives, sanitizer processes and corpus timing share resources.
func TestWave29AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE29_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_29.a")
	binary := h.build(stage0, "wave29", entry, archive, false)
	oracle := volumeOracle(h, "wave29-oracle", "oracle_wave_29.go")
	types, controls := wave29Controls(t, repository)
	declaration := h.write("node.d.ts", types)
	config := h.write("controls.json", fmt.Sprintf(`{"compilerOptions":{"strict":true,"target":"ESNext","module":"NodeNext","moduleResolution":"NodeNext"},"files":[%q]}`, declaration))
	var paths []string
	for i, source := range controls {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source))
	}

	if err := os.MkdirAll(filepath.Join(directory, "node_modules/sharp"), 0755); err != nil {
		t.Fatal(err)
	}
	h.write("node_modules/sharp/package.json", `{"name":"sharp","version":"0.35.3","types":"index.d.ts"}`)
	h.write("node_modules/sharp/index.d.ts", `export interface Sharp {toFile(path:string):Promise<void>;}declare function sharp():Sharp;export default sharp;`)
	for i, body := range []string{
		"let name='';do{name=`${index}.png`;index++;}while(NodeFileSystem.existsSync(name));NodeFileSystem.writeFileSync(name,bytes);",
		"while(true){const candidate=NodePath.join(directory,`${index}.png`);try{await access(candidate);index++;}catch{break;}}const chosen=NodePath.join(directory,`${index}.png`);await writeFile(chosen,bytes);",
		"while(NodeFileSystem.existsSync(`${index}.png`))index++;await sharp().toFile(`${index}.png`);",
		"while(NodeFileSystem.existsSync(`${index}.png`))index++;await image.toFile(`${index}.png`);",
		"while(NodeFileSystem.existsSync(`${index}.png`))index++;const f=()=>{index++;};NodeFileSystem.writeFileSync(`${index}.png`,bytes);",
		"while(NodeFileSystem.existsSync(`${index}.png`))index++;NodeFileSystem.writeFileSync(`${index}.png`,bytes,{['flag']:'wx'});",
		"for(let unused=0;NodeFileSystem.existsSync(`${index}.png`);index++){}NodeFileSystem.writeFileSync(`${index}.png`,bytes);",
		"while(NodeFileSystem.existsSync(`${index}.png`))index++;NodeFileSystem.writeFileSync<string>(`${index}.png`,bytes);",
	} {
		source := "import * as NodeFileSystem from 'node:fs';import {access,writeFile} from 'node:fs/promises';import * as NodePath from 'node:path';import sharp from 'sharp';declare const image:{toFile(path:string):Promise<void>};export async function save(directory:string,bytes:Uint8Array){let index=0;" + body + "}"
		paths = append(paths, h.write(fmt.Sprintf("extra-%03d.a", i), source))
	}
	paths = append(paths, h.write("unicode.a", "/* 世界 🌍 */\r\nimport {existsSync,writeFileSync} from 'node:fs';\r\nlet é=0;while(existsSync(`${é}.png`)){é++;}writeFileSync(`${é}.png`,'');\r\n"))
	paths = append(paths, h.write("names.a", "export const data=1;export class Fields{#under_score=1;method_name(){return this.#under_score;}}\n"))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	if !bytes.Contains(truth.stdout, []byte("\tnexus/concurrency-no-check-then-write\t")) || bytes.Contains(truth.stdout, []byte("\tid-denylist\t")) || bytes.Contains(truth.stdout, []byte("\tid-match\t")) {
		t.Fatal("controls did not exercise expected default rule behavior")
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave29-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"denylist", "id_denylist.a", "if(names.length === 0) { return; }", "if(names.length === 0) { this.rules.add('id-denylist', 'restricted', 'wrong default finding', 0); return; }"},
		{"match", "id_match.a", "if(pattern === '') { return; }", "if(pattern === '') { this.rules.add('id-match', 'notMatch', 'wrong default finding', 0); return; }"},
		{"race", "concurrency_no_check_then_write.a", "this.rules.byte(check.end), 'nexus/'", "this.rules.byte(check.end) + 1, 'nexus/'"},
	} {
		mutant := wave29Mutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s mutant exits 0, empty stderr, byte oracle catches offset %d", change.name, firstDifference(got.stdout, truth.stdout))
	}
	provenanceOverlay := h.overlay("provenance-mutant", "bridge/tsgo/checker/symbol_provenance.go", "out.text(module)", `out.text(module[:0])`)
	provenanceArchive := h.archive("provenance-mutant", provenanceOverlay, false)
	provenanceBinary := h.build(stage0, "provenance-mutant", entry, provenanceArchive, false)
	provenanceResult := h.must("provenance-mutant-run", exec.Command(provenanceBinary, config, manifest))
	if len(provenanceResult.stderr) != 0 || bytes.Equal(provenanceResult.stdout, truth.stdout) {
		t.Fatal("provenance mutant survived")
	}
	t.Logf("provenance mutant exits 0, empty stderr, byte oracle catches offset %d", firstDifference(provenanceResult.stdout, truth.stdout))
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,1,'Identifier','symbol-provenance'));`)
	releasedBinary := h.build(stage0, "released", released, archive, false)
	anchor := h.write("anchor.a", "x;")
	got := h.run("released-run", exec.Command(releasedBinary, config, anchor))
	exit, exited := got.err.(*exec.ExitError)
	if !exited || exit.ExitCode() != 70 || !bytes.Contains(got.stderr, []byte("invalid or released checker handle")) {
		t.Fatalf("released handle accepted: %v %s", got.err, got.stderr)
	}
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released handle.")
	mutatedArchive := h.archive("released-registry", overlay, false)
	mutatedRelease := h.build(stage0, "released-mutant", released, mutatedArchive, false)
	got = h.must("released-mutant-run", exec.Command(mutatedRelease, config, anchor))
	if len(got.stderr) != 0 {
		t.Fatal("released handle mutant did not exit cleanly")
	}
	t.Log("released handle mutant exits 0 and is caught by required panic")
	if root := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"); root != "" {
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-volume/compiler.manifest"))
		if err != nil {
			t.Fatal(err)
		}
		var paths []string
		for _, relative := range strings.Split(string(data), "\n") {
			if relative != "" {
				paths = append(paths, filepath.Join(root, relative))
			}
		}
		if len(paths) != 77 {
			t.Fatalf("compiler roots=%d want77", len(paths))
		}
		mf := h.write("compiler.manifest", strings.Join(paths, "\n")+"\n")
		cfg := filepath.Join(root, "src/compiler/tsconfig.json")
		h.compare("compiler", oracle, binary, cfg, mf)
		h.compare("compiler-asan", oracle, asan, cfg, mf)
		wave29Timing(h, "compiler", oracle, binary, cfg, mf)
	} else {
		t.Log("compiler corpus not requested")
	}
	var repositoryPaths []string
	data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-volume/repository.manifest"))
	if err != nil {
		t.Fatal(err)
	}
	for _, relative := range strings.Split(string(data), "\n") {
		if relative != "" {
			repositoryPaths = append(repositoryPaths, filepath.Join(repository, relative))
		}
	}
	mf := h.write("repository.manifest", strings.Join(repositoryPaths, "\n")+"\n")
	cfg := filepath.Join(repository, "tsconfig.json")
	h.compare("repository", oracle, binary, cfg, mf)
	h.compare("repository-asan", oracle, asan, cfg, mf)
	wave29Timing(h, "repository", oracle, binary, cfg, mf)
}
func wave29Timing(h *harness, name, oracle, binary, config, manifest string) {
	goRun := h.must(name+"-timed-go", exec.Command(oracle, config, manifest))
	command := exec.Command(binary, config, manifest)
	command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
	native := h.must(name+"-timed-native", command)
	if !bytes.Equal(goRun.stdout, native.stdout) {
		h.t.Fatal("timed byte mismatch")
	}
	h.t.Logf("%s timing native=%s Go=%s; native %s; Go %s", name, native.elapsed, goRun.elapsed, native.stderr, goRun.stderr)
}
