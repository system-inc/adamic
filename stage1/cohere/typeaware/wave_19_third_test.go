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
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// Not parallel: native/sanitized builds and mutants share the machine.
func TestWave19ThirdSimpleRules(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE19_THIRD_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err = os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	overlay := h.overlay("ancestry-registration", "bridge/tsgo/checker/facts.go", "switch mode {", "switch mode {\ncase \"declaration-ancestry\": return p.declarationAncestry(out,c,node,question)\ncase \"wave19-type-signatures\": return p.wave19TypeSignatures(out,c,node,question)\ncase \"wave19-generic-call\": return p.wave19GenericCall(out,c,node,question)\ncase \"wave19-type-members\": return p.wave19TypeMembers(out,c,node,question)\ncase \"wave19-heritage-members\": return p.wave19HeritageMembers(out,c,node,question)")
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("ancestry", overlay, false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_19_third/main.a")
	binary := h.build(stage0, "third", entry, archive, false)
	oracle := volumeOracle(h, "third-oracle", "oracle_wave_19_third.go")
	var paths []string
	sources := []string{"/* 世界 🌍 */\r\nSymbol(); (Symbol)(); ((Symbol))(); Symbol(undefined); Symbol(...[]); new Symbol; globalThis.Symbol();", "declare const value:unknown;typeof value === 'array';'null' != typeof value;typeof value!==undefined;typeof value===null;typeof value!==true;typeof value==1n;typeof value==/a/;typeof value===`object`;", "function f(undefined:any,Symbol:any){typeof value!==undefined;Symbol();} declare const value:unknown;"}
	unique := map[string]bool{}
	for _, name := range []string{"symbol_description_test.go", "valid_typeof_test.go"} {
		tree, e := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/core", name), nil, 0)
		if e != nil {
			t.Fatal(e)
		}
		goast.Inspect(tree, func(node goast.Node) bool {
			row, ok := node.(*goast.CompositeLit)
			if !ok || row.Type != nil || len(row.Elts) < 2 || len(row.Elts) > 3 {
				return true
			}
			label, ok := row.Elts[0].(*goast.BasicLit)
			if !ok || label.Kind != token.STRING {
				return true
			}
			literal, ok := row.Elts[1].(*goast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			source, e := strconv.Unquote(literal.Value)
			if e != nil {
				t.Fatal(e)
			}
			if !unique[source] {
				sources = append(sources, source)
				unique[source] = true
			}
			return true
		})
	}
	if len(unique) < 35 {
		t.Fatalf("only %d production sources", len(unique))
	}
	for i, source := range sources {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["control-000.a"]}`)
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("parse-filter", exec.Command(oracle, config, manifest, "--valid-sources"))
	manifest = h.write("valid.manifest", string(valid.stdout))
	t.Logf("%d controls before independent parse filtering; %d retained", len(paths), len(strings.Fields(string(valid.stdout))))
	truth := h.compare("controls", oracle, binary, config, manifest)
	if !bytes.Contains(truth.stdout, []byte("expected")) || !bytes.Contains(truth.stdout, []byte("suggestString")) {
		t.Fatal("missing positives/suggestions")
	}
	sanitizedArchive := h.archive("ancestry-asan", overlay, true)
	sanitized := h.build(stage0, "third-asan", entry, sanitizedArchive, true)
	h.compare("controls-asan", oracle, sanitized, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{{"symbol", "symbol-description/rule.a", "callee.text!=='Symbol'", "callee.text!=='SymbolWrong'"}, {"typeof", "valid-typeof/rule.a", "'symbol','bigint']", "'symbol','bigint','array']"}} {
		root := filepath.Join(repository, "stage1/cohere/typeaware/wave_19_third")
		scratch := filepath.Join(directory, change.name+"-source")
		err = filepath.WalkDir(root, func(path string, d os.DirEntry, e error) error {
			if e != nil {
				return e
			}
			relative, e := filepath.Rel(root, path)
			if e != nil {
				return e
			}
			target := filepath.Join(scratch, relative)
			if d.IsDir() {
				return os.MkdirAll(target, 0755)
			}
			if !strings.HasSuffix(path, ".a") {
				return nil
			}
			data, e := os.ReadFile(path)
			if e != nil {
				return e
			}
			source := string(data)
			if relative == change.file {
				if strings.Count(source, change.from) != 1 {
					t.Fatal("nonunique mutation")
				}
				source = strings.Replace(source, change.from, change.to, 1)
			}
			// Only imports outside our owned tree are rewritten; owned modules remain together.
			for _, prefix := range []string{"../../../typescript/", "../unary_minus.ts", "../frames.ts", "../diagnostic.ts", "../../suggestion.ts", "../../repair.ts"} {
				if strings.HasSuffix(prefix, "/") {
					source = strings.ReplaceAll(source, prefix, filepath.Join(repository, "stage1/typescript")+"/")
				} else {
					source = strings.ReplaceAll(source, prefix, filepath.Join(repository, "stage1/cohere/typeaware", filepath.Base(prefix)))
				}
			}
			return os.WriteFile(target, []byte(source), 0600)
		})
		if err != nil {
			t.Fatal(err)
		}
		mutant := h.build(stage0, change.name+"-mutant", filepath.Join(scratch, "main.a"), archive, false)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived")
		}
		t.Logf("%s compiled mutant exits 0 with empty stderr; only Go bytes catch byte %d", change.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, population := range []struct{ name, root, config string }{{"repository", repository, filepath.Join(repository, "tsconfig.json")}, {"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json")}} {
		if population.root == "" {
			t.Fatal("ADAMIC_TYPESCRIPT_SOURCE required")
		}
		data, e := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage", population.name+".manifest"))
		if e != nil {
			t.Fatal(e)
		}
		var roots []string
		for _, line := range strings.Split(string(data), "\n") {
			if line != "" {
				roots = append(roots, filepath.Join(population.root, line))
			}
		}
		manifest := h.write(population.name+".manifest", strings.Join(roots, "\n")+"\n")
		h.compare(population.name, oracle, binary, population.config, manifest)
		h.compare(population.name+"-asan", oracle, sanitized, population.config, manifest)
	}
}

func cloneWave19Third(h *harness, name, file, from, to string) string {
	h.t.Helper()
	root := filepath.Join(h.repository, "stage1/cohere/typeaware/wave_19_third")
	scratch := filepath.Join(h.directory, name+"-source")
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		relative, e := filepath.Rel(root, path)
		if e != nil {
			return e
		}
		target := filepath.Join(scratch, relative)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		if !strings.HasSuffix(path, ".a") {
			return nil
		}
		data, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		source := string(data)
		if relative == file {
			if strings.Count(source, from) != 1 {
				h.t.Fatal("nonunique mutation")
			}
			source = strings.Replace(source, from, to, 1)
		}
		expression := regexp.MustCompile(`from '([^']+)'`)
		source = expression.ReplaceAllStringFunc(source, func(match string) string {
			parts := expression.FindStringSubmatch(match)
			if !strings.HasPrefix(parts[1], ".") {
				return match
			}
			resolved := filepath.Clean(filepath.Join(filepath.Dir(path), parts[1]))
			if strings.HasPrefix(resolved, root+string(os.PathSeparator)) {
				return match
			}
			return "from '" + resolved + "'"
		})
		return os.WriteFile(target, []byte(source), 0600)
	})
	if err != nil {
		h.t.Fatal(err)
	}
	return scratch
}

// Not parallel: native builds and sanitizer runs share the machine.
func TestWave19AwaitRule(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE19_AWAIT_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err = os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	overlay := h.overlay("ancestry-registration", "bridge/tsgo/checker/facts.go", "switch mode {", "switch mode {\ncase \"declaration-ancestry\": return p.declarationAncestry(out,c,node,question)\ncase \"wave19-type-signatures\": return p.wave19TypeSignatures(out,c,node,question)\ncase \"wave19-generic-call\": return p.wave19GenericCall(out,c,node,question)\ncase \"wave19-type-members\": return p.wave19TypeMembers(out,c,node,question)\ncase \"wave19-heritage-members\": return p.wave19HeritageMembers(out,c,node,question)")
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("ancestry", overlay, false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_19_third/await_main.a")
	binary := h.build(stage0, "await", entry, archive, false)
	oracle := volumeOracle(h, "await-oracle", "oracle_wave_19_await.go")
	var paths []string
	sources := []string{
		"async function f(){return 1;}", "export async function f(){return 1;}", "const f=async function named(){return 1;};", "const f=async ()=>1;", "const f=async (a:number)=>a;",
		"const o={run:async function named(){return 1;}};", "const o={async run(){return 1;}};", "class C{async run(){return 1;}}", "class C{static async run(){return 1;}}",
		"declare const b:string;class C{a=0\nasync [b](){return 1;}}", "declare const foo:unknown;foo\nasync ()=>1;", "async function f(){}", "async function f(){;}",
		"declare function g():Promise<number>;async function f(){async ()=>{await g();};}", "async function* f(){yield 1;}", "declare function g():AsyncIterable<number>;async function f(){for await(const x of g()){x;}}",
		"declare const resource:any;async function f(){await using x=resource;}", "declare function g():Promise<number>;const f=async ()=>await g();", "class C{async ''(){return 1;}}", "async /*keep*/ function f(){return 1;}",
	}
	unique := map[string]bool{}
	tree, e := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/core/require_await_test.go"), nil, 0)
	if e != nil {
		t.Fatal(e)
	}
	goast.Inspect(tree, func(node goast.Node) bool {
		row, ok := node.(*goast.CompositeLit)
		if !ok || row.Type != nil || len(row.Elts) < 2 || len(row.Elts) > 5 {
			return true
		}
		label, ok := row.Elts[0].(*goast.BasicLit)
		if !ok || label.Kind != token.STRING {
			return true
		}
		literal, ok := row.Elts[1].(*goast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			return true
		}
		source, e := strconv.Unquote(literal.Value)
		if e != nil {
			t.Fatal(e)
		}
		if !unique[source] {
			sources = append(sources, source)
			unique[source] = true
		}
		return true
	})
	t.Logf("%d await controls including %d production table sources", len(sources), len(unique))
	for i, source := range sources {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022","ESNext.Disposable"]},"files":["control-000.a"]}`)
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("parse-filter", exec.Command(oracle, config, manifest, "--valid-sources"))
	manifest = h.write("valid.manifest", string(valid.stdout))
	t.Logf("%d parse-valid controls", len(strings.Fields(string(valid.stdout))))
	truth := h.compare("controls", oracle, binary, config, manifest)
	if !bytes.Contains(truth.stdout, []byte("removeAsync")) {
		t.Fatal("no positive/suggestion")
	}
	sanitizedArchive := h.archive("ancestry-asan", overlay, true)
	sanitized := h.build(stage0, "await-asan", entry, sanitizedArchive, true)
	h.compare("controls-asan", oracle, sanitized, config, manifest)
	scratch := cloneWave19Third(h, "await", "require-await/rule.a", "if(this.containsAwait(body,context,context.flags))", "if(false && this.containsAwait(body,context,context.flags))")
	mutant := h.build(stage0, "await-mutant", filepath.Join(scratch, "await_main.a"), archive, false)
	got := h.must("await-mutant-run", exec.Command(mutant, config, manifest))
	if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
		t.Fatal("await mutant survived")
	}
	t.Logf("await mutant exits 0 with empty stderr; only Go bytes catch byte %d", firstDifference(got.stdout, truth.stdout))
	for i, source := range []string{"const f:()=>Promise<number>=async ()=>1;", "interface I{f():Promise<number>}class C implements I{async f(){return 1;}}", "[1].map(async x=>x);"} {
		path := h.write(fmt.Sprintf("contract-%d.a", i), source+"\nexport {};\n")
		contract := h.write(fmt.Sprintf("contract-%d.manifest", i), path+"\n")
		h.compare("contract", oracle, binary, config, contract)
		h.compare("contract-asan", oracle, sanitized, config, contract)
	}
	for _, population := range []struct{ name, root, config string }{{"repository", repository, filepath.Join(repository, "tsconfig.json")}, {"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json")}} {
		if population.root == "" {
			t.Fatal("ADAMIC_TYPESCRIPT_SOURCE required")
		}
		data, e := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage", population.name+".manifest"))
		if e != nil {
			t.Fatal(e)
		}
		var roots []string
		for _, line := range strings.Split(string(data), "\n") {
			if line != "" {
				roots = append(roots, filepath.Join(population.root, line))
			}
		}
		manifest := h.write(population.name+".manifest", strings.Join(roots, "\n")+"\n")
		h.compare(population.name, oracle, binary, population.config, manifest)
		h.compare(population.name+"-asan", oracle, sanitized, population.config, manifest)
	}

}
