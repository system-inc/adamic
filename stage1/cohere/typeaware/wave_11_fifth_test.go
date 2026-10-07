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

func wave11FifthControls() []string {
	return []string{
		`RegExp('\\1(a)');({RegExp}={});RegExp('\\1(a)');`,
		`const object={RegExp};RegExp('\\1(a)');const alias=RegExp;({alias}=other);alias('\\1(a)');`,
		`throw 'x';throw undefined;throw new Error();throw (new Error());function f(undefined:any){throw undefined}throw x;throw true;throw null;`,
		`throw a='x';throw a=new Error();throw a&&='x';throw a||='x';throw a??='x';throw a+='x';throw a&&'x';throw a||'x';throw a??'x';throw a?'x':'y';throw a?'x':new Error();throw new Error(),1;`,
		`const a=/\1(a)/;const b=/(a\1)/;const c=/(a|\1b)/;const d=/(?<=(a)\1)b/;const e=/(?!(a))\1/;const f=/(a)\1/;`,
		`RegExp('\\1(a)');new RegExp('\\1(a)','u');const p='\\1(a)';RegExp(p);let q='\\1(a)';RegExp(q);let r='\\1(a)';r='(a)\\1';RegExp(r);`,
		`const make=RegExp;make('\\1(a)');new make('\\1(a)');const {RegExp:R}=globalThis;R('\\1(a)');globalThis['Reg'+'Exp']('\\1(a)');`,
		`function f(RegExp:any){RegExp('\\1(a)')}RegExp('\\1(a)');function g(globalThis:any){globalThis.RegExp('\\1(a)')}`,
		`RegExp('\\1(a)');RegExp=other;window.RegExp('\\1(a)');const w=window;w.RegExp('\\1(a)');`,
		`let r; r=RegExp;r('\\1(a)');r=other;r('\\1(a)');const a=(x?RegExp:RegExp);a('\\1(a)');`,
		`foo(function(){});new Foo(function(){});foo(function bar(){bar()});foo(function bar(){const bar=1;bar});foo(function bar(){function bar(){}bar()});`,
		`foo(function(){arguments});foo(function(arguments:any){arguments});foo(function(){()=>arguments});foo(function(){function inner(){arguments}});`,
		`foo(function(){this.x});foo(function(){this.x}.bind(this));foo((function(){this.x}).bind(this));foo((function(){}?.bind)(this));foo(function(){}.bind(this).bind(other));`,
		`foo(function(a,a){});foo(function(this:any){});foo(function*(){});foo(function(){new.target});foo(function(){()=>this});foo(async function(){});`,
		"foo(async function\n(){});foo(function /*keep*/ named (){});foo(function named /*keep*/ (){});foo(function(){}.bind(/*keep*/this));",
		`foo(native || function(){});foo(x?function(){}:function(){});foo((function(){}));foo(function(){}['bind'](this));`,
		"/* 世界 🌍 */\r\nfoo(function é(){return 1});\r\nthrow 'é';\r\nconst re=/\\1(é)/;",
		`const p='\\1'+'(a)';RegExp(p);const s='a';RegExp(s+'\\1(a)');RegExp('\\1(a){', 'u');RegExp('\\1(a){', flags);`,
		`let root=globalThis;const {RegExp}=root;RegExp('\\1(a)');`,
		`foo(function(){class C{m(){this.x;arguments;new.target}}});foo(function(){const arguments=1;arguments});`,
	}
}

// Read source inputs from the pinned upstream-derived fixture tables. Options
// remain the production defaults on both sides; expectations are never copied.
func wave11FifthUpstream(t *testing.T, repository string) []string {
	t.Helper()
	var sources []string
	for _, name := range []string{"no_throw_literal", "no_useless_backreference", "prefer_arrow_callback"} {
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

func wave11FifthSourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
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
	return h.build(stage0, name, filepath.Join(directory, "wave_11_fifth_suite.a"), archive, false)
}

// Not parallel: archive builds, sanitizer subprocesses and measurements share a machine.
func TestWave11FifthAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE_11_FIFTH_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_11_fifth_suite.a")
	binary := h.build(stage0, "wave-11-fifth", entry, archive, false)
	oracle := volumeOracle(h, "wave-11-fifth-oracle", "oracle_wave_11_fifth.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for i, source := range append(wave11FifthControls(), wave11FifthUpstream(t, repository)...) {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.ts", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("controls-valid", exec.Command(oracle, config, manifest, "--valid-sources"))
	validPaths := strings.Fields(string(valid.stdout))
	t.Logf("controls: %d sources, %d parse-clean, %d excluded parse diagnostics", len(paths), len(validPaths), len(paths)-len(validPaths))
	for _, path := range paths {
		if !strings.Contains(string(valid.stdout), path+"\n") {
			t.Logf("excluded parse diagnostics: %s", filepath.Base(path))
		}
	}
	manifest = h.write("controls-valid.manifest", string(valid.stdout))
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-throw-literal", "no-useless-backreference", "prefer-arrow-callback"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control: %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave-11-fifth-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"throw", "no_throw_literal.a", "this.rules.byte(node.end), '')", "this.rules.byte(node.end) + 1, '')"},
		{"backreference", "no_useless_backreference.a", "this.rules.byte(node.end), '')", "this.rules.byte(node.end) + 1, '')"},
		{"arrow-fix", "prefer_arrow_callback.a", "this.rules.byte(insertion), ' =>'))", "this.rules.byte(insertion), ' ->'))"},
		{"provenance", "wave_11_fifth_support.a", "(declaration) => !declaration.declarationFile", "(declaration) => declaration.declarationFile"},
		{"regex-path", "no_useless_backreference.a", "return 'nested';", "return 'forward';"},
		{"self-resolution", "prefer_arrow_callback.a", "this.facts.symbol(current) === symbol", "this.rules.parser.node(current).text === text"},
	} {
		mutant := wave11FifthSourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, independent Go bytes catch byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		if err := os.Remove(mutant); err != nil {
			t.Fatal(err)
		}
	}
	provenanceOverlay := h.overlay("symbol-provenance", "bridge/tsgo/checker/symbol_declaration_provenance.go", "out.yes(source.IsDeclarationFile)", "out.yes(!source.IsDeclarationFile)")
	provenanceArchive := h.archive("symbol-provenance", provenanceOverlay, false)
	provenanceMutant := h.build(stage0, "symbol-provenance-mutant", entry, provenanceArchive, false)
	provenanceResult := h.must("symbol-provenance-mutant-run", exec.Command(provenanceMutant, config, manifest))
	if len(provenanceResult.stderr) != 0 || bytes.Equal(provenanceResult.stdout, truth.stdout) {
		t.Fatal("symbol provenance question mutant survived")
	}
	t.Logf("symbol provenance question mutant: exit 0, empty stderr, independent Go bytes catch byte %d", firstDifference(provenanceResult.stdout, truth.stdout))
	os.Remove(provenanceArchive)
	os.Remove(provenanceMutant)
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
console.log(tsgoInspect(program,file,0,1,'Identifier','symbol-declaration-provenance\n0'));
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
