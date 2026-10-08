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
	"sort"
	"strconv"
	"strings"
	"testing"
)

func wave22FourthSources(h *harness) []string {
	unique := map[string]bool{}
	add := func(expr goast.Expr) {
		if literal, ok := expr.(*goast.BasicLit); ok && literal.Kind == token.STRING {
			value, err := strconv.Unquote(literal.Value)
			if err != nil {
				h.t.Fatal(err)
			}
			unique[value] = true
		}
	}
	for _, name := range []string{"prefer_numeric_literals", "prefer_object_has_own", "prefer_object_spread"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(h.repository, "cohere/internal/lint/rules/core", name+"_test.go"), nil, 0)
		if err != nil {
			h.t.Fatal(err)
		}
		goast.Inspect(tree, func(node goast.Node) bool {
			if value, ok := node.(*goast.ValueSpec); ok && len(value.Names) == 1 && value.Names[0].Name == "impliedEvalAmbientGlobals" && len(value.Values) == 1 {
				if literal, ok := value.Values[0].(*goast.BasicLit); ok {
					text, err := strconv.Unquote(literal.Value)
					if err != nil {
						h.t.Fatal(err)
					}
					h.write("ambient-globals.d.ts", text)
				}
			}

			if field, ok := node.(*goast.KeyValueExpr); ok {
				if key, ok := field.Key.(*goast.Ident); ok && (key.Name == "sourceText" || key.Name == "Code" || key.Name == "source") {
					add(field.Value)
				}
			}
			if assignment, ok := node.(*goast.AssignStmt); ok {
				for _, rhs := range assignment.Rhs {
					values, ok := rhs.(*goast.CompositeLit)
					if !ok {
						continue
					}
					array, ok := values.Type.(*goast.ArrayType)
					if !ok {
						continue
					}
					structure, ok := array.Elt.(*goast.StructType)
					if !ok {
						continue
					}
					sourceIndex := -1
					slot := 0
					for _, field := range structure.Fields.List {
						for _, fieldName := range field.Names {
							if fieldName.Name == "sourceText" || fieldName.Name == "source" {
								sourceIndex = slot
							}
							slot++
						}
					}
					if sourceIndex < 0 {
						continue
					}
					for _, entry := range values.Elts {
						if row, ok := entry.(*goast.CompositeLit); ok && sourceIndex < len(row.Elts) {
							if _, keyed := row.Elts[sourceIndex].(*goast.KeyValueExpr); !keyed {
								add(row.Elts[sourceIndex])
							}
						}
					}
				}
			}
			return true
		})
	}
	for _, source := range []string{
		"/* 世界 🌍 */\r\nparseInt('11',2);Number?.parseInt(`77`,8);Number['parseInt']('fF',0x10);",
		"parseInt('11111111111111111111111111111111111111111111111111111111111111111',2);parseInt('ffffffffffffffff',16);parseInt('10000000000000000',16);parseInt('1_0',2);",
		"parseInt('\\x31\\u0031',2);parseInt('00',8);parseInt('',16);parseInt(' +11',2);parseInt(('11'),2);parseInt('11',(2));",
		"function f(Object){return({}).hasOwnProperty.call(a,b);}function g(){return{}.hasOwnProperty.call(a,b);}({/*c*/}).hasOwnProperty.call(a,b);",
		"Object/*c*/.prototype.hasOwnProperty.call(a,b);((Object['prototype']['hasOwnProperty']['call']))(a,b);",
		"const A=Object;A.assign({},x);const {assign: B}=Object;B({},y);let C;C=Object.assign;C({},z);",
		"const A=globalThis;const {Object:{assign: B}}=A;B({},x);(true?Object:Object).assign({},y);",
		"Object.assign=custom;Object.assign({},x);",
		"Object.assign({},x);Object=custom;",
		"foo\nObject.assign({},x).f();foo\nObject.assign({},x),2;",
		"const A=Object;A=A;A.assign({},x);const {...B}=Object;B.assign({},x);",
		"Object.assign(({a:1,}),(((x))),{b:2});()=>Object.assign({},a=b);()=>Object.assign({},()=>x);",
		"Object['as'+'sign']({},x);globalThis[`Obj${'ect'}`][`as${'sign'}`]({},y);",
		"Object.assign({get a(){return 1}});Object.assign({},{get a(){return 1}});Object.assign({},...sources);",
		"Object.assign({a:1 //c\n},x);Object.assign({},/*c*/x);Object.assign<T>({},x);",
		"let A;({assign:A}=Object);A({},x);function f(A=Object){A.assign({},x)}",
	} {
		unique[source] = true
	}

	result := make([]string, 0, len(unique))
	for source := range unique {
		result = append(result, source)
	}
	sort.Strings(result)
	return result
}

func wave22FourthSourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	// The mutation bundle uses .a for every Adamic source, including unchanged
	// copies of older .ts helpers. External imports still name existing files.
	var paths []string
	for _, extension := range []string{"*.ts", "*.a"} {
		matches, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware", extension))
		if err != nil {
			h.t.Fatal(err)
		}
		paths = append(paths, matches...)
	}
	for _, path := range paths {
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
		for _, dependency := range paths {
			base := filepath.Base(dependency)
			if strings.HasSuffix(base, ".ts") {
				source = strings.ReplaceAll(source, "./"+base, "./"+strings.TrimSuffix(base, ".ts")+".a")
			}
		}
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
		base := filepath.Base(path)
		if strings.HasSuffix(base, ".ts") {
			base = strings.TrimSuffix(base, ".ts") + ".a"
		}
		if err := os.WriteFile(filepath.Join(directory, base), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_22_fourth_suite.a"), archive, false)
}

// Not parallel: native builds, sanitizers and timing share the worker.
func TestWave22FourthAgreement(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE22_FOURTH_ARTIFACTS")
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
	binary := h.build(stage0, "wave-22-fourth", filepath.Join(repository, "stage1/cohere/typeaware/wave_22_fourth_suite.a"), archive, false)
	oracle := volumeOracle(h, "wave-22-fourth-oracle", "oracle_wave_22_fourth.go")
	config := h.write("controls-tsconfig.json", `{ "compilerOptions": {"strict":true,"target":"ES2022","lib":["ES2022","DOM"],"moduleDetection":"auto","allowJs":true,"types":[]} }`)
	var paths []string
	for at, source := range wave22FourthSources(h) {
		for _, extension := range []string{"js", "a"} {
			paths = append(paths, h.write(fmt.Sprintf("control-%03d.%s", at, extension), source+"\nexport {};\n"))
		}
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("valid-controls", exec.Command(oracle, config, manifest, "--valid-sources"))
	accepted := strings.Fields(string(valid.stdout))
	t.Logf("controls: %d accepted, %d rejected by independent parser", len(accepted), len(paths)-len(accepted))
	manifest = h.write("valid-controls.manifest", strings.Join(accepted, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"prefer-numeric-literals", "prefer-object-has-own", "prefer-object-spread"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive controls for %s", name)
		}
	}
	if os.Getenv("ADAMIC_WAVE22_FOURTH_CONTROLS_ONLY") != "" {
		return
	}
	asanArchive := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave-22-fourth-asan", filepath.Join(repository, "stage1/cohere/typeaware/wave_22_fourth_suite.a"), asanArchive, true)
	h.compare("controls-asan", oracle, asan, config, manifest)

	for _, change := range []struct{ name, file, from, to string }{
		{"numeric-prefix", "prefer_numeric_literals.a", "radix === 2 ? '0b'", "radix === 2 ? '0o'"},
		{"has-own-fix-span", "prefer_object_has_own.a", "support.repair(start, target.end,", "support.repair(start, target.end + 1,"},
		{"spread-parens", "prefer_object_spread.a", "wrap ? '})' : '}'", "wrap ? '}' : '}'"},
	} {
		mutant := wave22FourthSourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s survived", change.name)
		}
		t.Logf("%s: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
	}

	for _, question := range []string{"global-binding\nnode", "resolved-name\nObject", "symbol-identities"} {
		probe := h.write("released-probe.a", "x;\n")
		kind, end := "Identifier", 1
		if question == "resolved-name\nObject" {
			kind, end = "SourceFile", 3
		}
		source := h.write("released-next.a", "import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';\nconst args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,"+strconv.Itoa(end)+","+strconv.Quote(kind)+","+strconv.Quote(question)+"));\n")
		stale := h.build(stage0, "released-next", source, archive, false)
		got := h.run("released-next-run", exec.Command(stale, config, probe))
		if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
			t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
		}
		if question == "global-binding\nnode" {
			overlay := h.overlay("next-released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released handle.")
			mutantArchive := h.archive("next-released-registry", overlay, false)
			mutant := h.build(stage0, "released-next-mutant", source, mutantArchive, false)
			h.must("released-next-mutant-run", exec.Command(mutant, config, probe))
		}
		t.Logf("%q released handle: panic 70; registry retention checked separately", question)
	}

	for _, corpus := range []struct{ name, root, config string }{{"repository", repository, filepath.Join(repository, "tsconfig.json")}, {"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), ""}} {
		if corpus.root == "" {
			t.Log("compiler omitted: ADAMIC_TYPESCRIPT_SOURCE unset")
			continue
		}
		corpusConfig := corpus.config
		if corpus.name == "compiler" {
			corpusConfig = filepath.Join(corpus.root, "src/compiler/tsconfig.json")
		}
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage", corpus.name+".manifest"))
		if err != nil {
			t.Fatal(err)
		}
		var paths []string
		for _, path := range strings.Split(string(data), "\n") {
			if path != "" {
				paths = append(paths, filepath.Join(corpus.root, path))
			}
		}
		list := h.write(corpus.name+".manifest", strings.Join(paths, "\n")+"\n")
		h.compare(corpus.name, oracle, binary, corpusConfig, list)
		h.compare(corpus.name+"-asan", oracle, asan, corpusConfig, list)
	}
}
