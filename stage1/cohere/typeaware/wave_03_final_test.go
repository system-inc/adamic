package typeaware

import (
	"bytes"
	"fmt"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func wave03FinalBodies(t *testing.T, repository string) []string {
	t.Helper()
	var bodies []string
	seen := map[string]bool{}
	for _, path := range []string{"core/require_await_test.go", "core/symbol_description_test.go", "structure/react_hook_no_any_type_test.go"} {
		tree, err := goparser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules", path), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		goast.Inspect(tree, func(node goast.Node) bool {
			literal, ok := node.(*goast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			text, err := strconv.Unquote(literal.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !seen[text] && (strings.Contains(text, "async ") || strings.Contains(text, "Symbol(") || strings.Contains(text, "useState(") || strings.Contains(text, "use2Things(") || strings.Contains(text, "useX(")) {
				seen[text] = true
				bodies = append(bodies, text)
			}
			return true
		})
	}
	return bodies
}
func wave03FinalCompare(h *harness, name, oracle, binary, config, manifest, catalog string) result {
	h.t.Helper()
	want := h.must(name+"-go", exec.Command(oracle, config, manifest))
	got := h.must(name+"-native", exec.Command(binary, config, manifest, catalog))
	if len(got.stderr) != 0 || !bytes.Equal(want.stdout, got.stdout) {
		h.t.Fatalf("%s differs at byte %d; native stderr %s", name, firstDifference(want.stdout, got.stdout), got.stderr)
	}
	h.t.Logf("%s: %d identical bytes", name, len(got.stdout))
	return want
}
func wave03FinalMutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	root := filepath.Join(h.repository, "stage1/cohere/typeaware/wave_03_final")
	destination := filepath.Join(h.directory, name+"-source")
	imports := regexp.MustCompile(`from '([^']+)'`)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".a") {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(data)
		if relative == file {
			if strings.Count(text, from) != 1 {
				h.t.Fatalf("nonunique mutant %s", name)
			}
			text = strings.Replace(text, from, to, 1)
		}
		text = imports.ReplaceAllStringFunc(text, func(match string) string {
			parts := imports.FindStringSubmatch(match)
			target := parts[1]
			if !strings.HasPrefix(target, ".") {
				return match
			}
			absolute := filepath.Clean(filepath.Join(filepath.Dir(path), target))
			if strings.HasPrefix(absolute, root+string(filepath.Separator)) {
				r, _ := filepath.Rel(root, absolute)
				absolute = filepath.Join(destination, r)
			}
			return "from '" + absolute + "'"
		})
		target := filepath.Join(destination, relative)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		return os.WriteFile(target, []byte(text), 0600)
	})
	if err != nil {
		h.t.Fatal(err)
	}
	return h.build(stage0, name, filepath.Join(destination, "main.a"), archive, false)
}

// Not parallel: checker archives, sanitizers and timings share machine resources.
func TestWave03FinalAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_WAVE03_FINAL_ARTIFACTS"); path != "" {
		directory = path
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("final-checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_03_final/main.a")
	binary := h.build(stage0, "final", entry, archive, false)
	oracle := volumeOracle(h, "final-oracle", "oracle_wave_03_final.go")
	config := h.write("final-config.json", `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022","DOM"],"moduleDetection":"auto","jsx":"preserve"}}`)
	bodies := append(wave03FinalBodies(t, repository), "Symbol(); (Symbol)(); Symbol(undefined); globalThis.Symbol(); new Symbol();", "declare function useState<T>(x:T):any;useState(1);", "declare const React:{use2():any;useÉ():any};(React).use2();React.useÉ();", "async function missing(){return 1;} async function real(){await foo();} async function nested(){async()=>await foo();}", "interface I {m():Promise<number>}; class C implements I {async m(){return 1;}}", "const f:()=>number|Promise<number>=async()=>1;const g:()=>Promise<number>=async()=>1;", "[1].map(async x=>x+1);", "declare function accept<T>(x:()=>T):T;const x:Promise<number>=accept(async()=>1);", "// 世界🌍\nasync function é(){return 1;}", "class C{a=0\nasync [key](){return 1;}}", "async function f(){await using x=foo();}", "declare class Box<T>{constructor(f:()=>T)};new Box(async()=>1);", "declare class Box<T>{constructor(f:()=>T)};new Box<Promise<number>>(async()=>1);", "declare function pick<T>(first:T,second:T):void;declare const first:()=>Promise<number>;pick(first,async()=>1);", "declare function f<T>(items:[()=>T]):void;f([async()=>1]);", "declare function f<T>(items:{value:()=>T}):void;f({value:async()=>1});", "declare function f<T>(factory:()=>{value:()=>T}):void;f(()=>({value:async()=>1}));", "declare function f<T>(value:T):void;f(async()=>1);", "async function f(){for await(const x of xs){foo(x)}}")
	var paths []string
	for index, source := range bodies {
		paths = append(paths, h.write(fmt.Sprintf("final-control-%03d.tsx", index), source+"\nexport {};\n"))
	}
	candidates := h.write("final-candidates.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("parse-filter", exec.Command(oracle, config, candidates, "--valid-sources"))
	manifest := h.write("final-controls.manifest", string(valid.stdout))
	metadata := h.must("linked-metadata", exec.Command(oracle, config, manifest, "--linked-metadata"))
	catalog := h.write("linked-rules.txt", string(metadata.stdout))
	t.Logf("controls: %d candidates %d parseable; linked metadata %q", len(paths), len(strings.Fields(string(valid.stdout))), string(metadata.stdout))
	truth := wave03FinalCompare(h, "final-controls", oracle, binary, config, manifest, catalog)
	for _, name := range []string{"require-await", "symbol-description", "structure/react-hook-no-any-type"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	if os.Getenv("ADAMIC_WAVE03_FINAL_QUICK") != "" {
		return
	}
	sanitized := h.archive("final-checker-asan", "", true)
	asan := h.build(stage0, "final-asan", entry, sanitized, true)
	wave03FinalCompare(h, "final-controls-asan", oracle, asan, config, manifest, catalog)
	for _, change := range []struct{ name, file, from, to string }{{"await", "rules/require-await/rule.a", "obliges every caller", "obliges  every caller"}, {"symbol", "rules/symbol-description/rule.a", "Pass one.", "Pass one. "}, {"hook", "rules/react-hook-no-any-type/rule.a", "here resolves to", "here  resolves to"}} {
		mutant := wave03FinalMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest, catalog))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s mutant compiled, exit 0, byte comparison catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, population := range []struct{ name, config, manifest string }{{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE03_REPOSITORY_MANIFEST")}, {"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE03_COMPILER_MANIFEST")}} {
		if population.manifest == "" {
			continue
		}
		wave03FinalCompare(h, population.name, oracle, binary, population.config, population.manifest, catalog)
		wave03FinalCompare(h, population.name+"-asan", oracle, asan, population.config, population.manifest, catalog)
		want := h.must(population.name+"-timed-go", exec.Command(oracle, population.config, population.manifest))
		command := exec.Command(binary, population.config, population.manifest, catalog)
		command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		got := h.must(population.name+"-timed-native", command)
		if !bytes.Equal(want.stdout, got.stdout) {
			t.Fatal("timed bytes differ")
		}
		t.Logf("%s native %s Go %s; %s", population.name, got.elapsed, want.elapsed, got.stderr)
	}
	probe := h.write("final-released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const p=tsgoProgram(args[0]??'',[file]);tsgoInspect(p,file,0,1,'Identifier','raw-shape');tsgoRelease(p);console.log(tsgoInspect(p,file,0,1,'Identifier','callable-return-shape\n1'));`)
	source := h.write("final-probe.ts", "x;")
	stale := h.build(stage0, "final-released", probe, archive, false)
	got := h.run("final-released-run", exec.Command(stale, config, source))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped %v %s", got.err, got.stderr)
	}
	t.Log("released query: required panic 70")
	overlay := h.overlay("final-released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released checker.")
	mutantArchive := h.archive("final-released-registry", overlay, false)
	mutant := h.build(stage0, "final-released-mutant", probe, mutantArchive, false)
	h.must("final-released-mutant-run", exec.Command(mutant, config, source))
	t.Log("released registry mutant exits 0 and fails required panic 70")
}
