package typeaware

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
	"sort"
	"strconv"
	"strings"
	"testing"
)

func wave11EighthUpstream(t *testing.T, repository string) []string {
	t.Helper()
	var sources []string
	for _, name := range []string{"no_danger_with_children", "no_namespace", "no_multi_comp"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/react", name+"_test.go"), nil, 0)
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

func TestWave11EighthAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE_11_EIGHTH_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err = os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave-11-eighth/main.a")
	binary := h.build(stage0, "eighth", entry, archive, false)
	oracle := volumeOracle(h, "eighth-oracle", "../wave-11-eighth/testdata/oracle.go")
	listenerOracle := volumeOracle(h, "eighth-listeners", "../wave-11-eighth/testdata/listeners.go")
	listenerConfig := h.write("listeners-config.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","jsx":"react"}}`)
	listenerSource := h.write("listeners.tsx", "export {};\n")
	listenerTruth := h.must("listeners", exec.Command(listenerOracle, listenerConfig, listenerSource))
	var listenerLines []string
	for _, directory := range []string{"react-no-danger-with-children", "react-no-namespace", "react-no-multi-comp"} {
		data, e := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/wave-11-eighth/rules", directory, "rule.json"))
		if e != nil {
			t.Fatal(e)
		}
		var declaration struct {
			Name  string `json:"name"`
			Kinds []int  `json:"kinds"`
		}
		if e = json.Unmarshal(data, &declaration); e != nil {
			t.Fatal(e)
		}
		sort.Ints(declaration.Kinds)
		var kinds []string
		for _, kind := range declaration.Kinds {
			kinds = append(kinds, strconv.Itoa(kind))
		}
		listenerLines = append(listenerLines, declaration.Name+"\t"+strings.Join(kinds, ","))
	}
	sort.Strings(listenerLines)
	if strings.Join(listenerLines, "\n")+"\n" != string(listenerTruth.stdout) {
		t.Fatalf("listener kinds mismatch: %s", listenerTruth.stdout)
	}
	t.Log("numeric manifests match live production Go listeners")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","jsx":"react","lib":["ES2022"],"noEmit":true},"include":["*.tsx"]}`)

	controls := []string{
		`<div dangerouslySetInnerHTML={{__html:'x'}}>child</div>; <div dangerouslySetInnerHTML={{__html:'x'}}/>; <div children="x" dangerouslySetInnerHTML={{__html:'x'}}/>;`,
		"<div dangerouslySetInnerHTML={{__html:'x'}}> </div>; <div dangerouslySetInnerHTML={{__html:'x'}}>\n </div>; <div dangerouslySetInnerHTML={{__html:'x'}}>\n <X/></div>;",
		`const a={children:1}; const b={...a}; const props={...b,dangerouslySetInnerHTML:{__html:'x'}}; <X {...props}/>; const c={...d}; const d={...c}; <X {...c}/>;`,
		`const a={children:1}; React.createElement('div',{...a,dangerouslySetInnerHTML:{__html:'x'}}); const p={...a,dangerouslySetInnerHTML:{__html:'x'}}; React.createElement('div',p);`,
		`document.createElement('div',{['dangerouslySetInnerHTML']:1},'x'); createElement('div',{dangerouslySetInnerHTML:1},'x'); React['createElement']('div',{dangerouslySetInnerHTML:1},'x');`,
		`<ns:x/>; <ns:x></ns:x>; <X/>; <X.Y/>; <this.foo/>; React.createElement('a:b'); Foo.createElement('a:b'); React.createElement('a:b',{});`,
		`import {createElement} from 'react'; createElement('a:b'); import {createElement as h} from 'react'; h('a:b');`,
		`const {createElement}=React; createElement('a:b'); const other=React.createElement; other('a:b');`,
		`function A(){return <div/>;} function B(){return null;} function c(){return <div/>;} const D=()=> <div/>; const e=()=> <div/>;`,
		`class A extends (React.Component) {} class B extends PureComponent {} const C=React.memo((p)=><A/>); const D=React.memo((p)=><div/>);`,
		`import {memo,forwardRef} from 'react'; const A=()=> <div/>; const B=memo(forwardRef((p,r)=><A/>));`,
		`const A=()=>null; export default ()=>null; const B=()=>x?<div/>:null; const C=()=>x&&<div/>; const D=()=> (0,<div/>);`,
		`function A(){try {return <div/>;}catch(e){}} function B(){while(x){return <div/>;}} function C(){for(const x of y){return <div/>;}} function D(){if(x){return <div/>;}}`,
		`const $foo=()=> <div/>; const _1=()=> <div/>; const 테스트=()=> <div/>; const ß=()=> <div/>; const 𐐨=()=> <div/>; const 𐐀=()=> <div/>;`,
		`const A=createReactClass({render:function(){return <div/>;}}); const B=createReactClass({render:function(){return <div/>;}});`,
		`const A=()=> <div/>; const o={Foo(){return <div/>;}, bar:function Bar(){return <div/>;}, low:()=>null}; module.exports=()=> <div/>; exports.foo=function Bar(){return <div/>;};`,
		`const A=()=> <div/>; demo=()=>()=>null; const B=(0,()=> <div/>); const C=()=> {function Nested(){return <div/>;}};`,
		"/* 世界 🌍 */\r\n<div dangerouslySetInnerHTML={{__html:'x'}}>text</div>; <svg:circle/>; function Émile(){return <div/>;} function Ωmega(){return <div/>;}",
	}

	var paths []string
	for i, source := range append(controls, wave11EighthUpstream(t, repository)...) {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.tsx", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("valid", exec.Command(oracle, config, manifest, "--valid-sources"))
	t.Logf("controls %d parse-clean %d", len(paths), len(strings.Fields(string(valid.stdout))))
	manifest = h.write("controls-valid.manifest", string(valid.stdout))
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"react/no-danger-with-children", "react/no-namespace", "react/no-multi-comp"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive: %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "eighth-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"danger-property", "rules/react-no-danger-with-children/rule.a", "includes(name.kind) && name.text === wanted", "includes(name.kind) && name.text === wanted + '!'"},
		{"namespace-colon", "rules/react-no-namespace/rule.a", "first.text.includes(':')", "first.text.includes('/')"},
		{"component-first", "rules/react-no-multi-comp/rule.a", "let at = 1", "let at = 2"},
	} {
		root := filepath.Join(directory, change.name+"-source")
		sourceRoot := filepath.Join(repository, "stage1/cohere/typeaware/wave-11-eighth")
		err = filepath.WalkDir(sourceRoot, func(path string, d os.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if d.IsDir() || !strings.HasSuffix(path, ".a") {
				return nil
			}
			relative, e := filepath.Rel(sourceRoot, path)
			if e != nil {
				return e
			}
			data, e := os.ReadFile(path)
			if e != nil {
				return e
			}
			source := string(data)
			if relative == change.file {
				if strings.Count(source, change.from) != 1 {
					t.Fatalf("nonunique %s", change.name)
				}
				source = strings.Replace(source, change.from, change.to, 1)
			}
			for _, external := range []string{"frames.ts", "diagnostic.ts", "unary_minus.ts"} {
				source = strings.ReplaceAll(source, "'../"+external+"'", "'"+filepath.Join(repository, "stage1/cohere/typeaware", external)+"'")
			}
			source = strings.ReplaceAll(source, "'../../../typescript/", "'"+filepath.Join(repository, "stage1/typescript")+"/")
			target := filepath.Join(root, relative)
			if e = os.MkdirAll(filepath.Dir(target), 0755); e != nil {
				return e
			}
			return os.WriteFile(target, []byte(source), 0644)
		})
		if err != nil {
			t.Fatal(err)
		}
		mutant := h.build(stage0, change.name, filepath.Join(root, "main.a"), archive, false)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("mutant survived %s", change.name)
		}
		t.Logf("%s mutant exit 0 empty stderr, Go bytes catch byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		os.Remove(mutant)
	}
	overlay := h.overlay("jsx-whitespace", "bridge/tsgo/checker/react_syntax_details.go", "out.yes(ast.IsWhitespaceOnlyJsxText(node))", "out.yes(false && ast.IsWhitespaceOnlyJsxText(node))")
	mutantArchive := h.archive("jsx-whitespace", overlay, false)
	mutant := h.build(stage0, "jsx-whitespace-mutant", entry, mutantArchive, false)
	got := h.must("jsx-whitespace-mutant-run", exec.Command(mutant, config, manifest))
	if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
		t.Fatal("raw whitespace mutant survived")
	}
	t.Logf("raw whitespace mutant exit 0 empty stderr, Go bytes catch byte %d", firstDifference(got.stdout, truth.stdout))
	os.Remove(mutantArchive)

	os.Remove(mutant)
	upperOverlay := h.overlay("unicode-upper", "bridge/tsgo/checker/react_syntax_details.go", "out.text(strings.Map(unicode.ToUpper, text(node)))", "out.text(strings.Map(func(r rune) rune { _ = unicode.ToUpper(r); return r }, text(node)))")
	upperArchive := h.archive("unicode-upper", upperOverlay, false)
	upperMutant := h.build(stage0, "unicode-upper-mutant", entry, upperArchive, false)
	upperResult := h.must("unicode-upper-mutant-run", exec.Command(upperMutant, config, manifest))
	if len(upperResult.stderr) != 0 || bytes.Equal(upperResult.stdout, truth.stdout) {
		t.Fatal("raw uppercase mutant survived")
	}
	t.Logf("raw uppercase mutant exit 0 empty stderr, Go bytes catch byte %d", firstDifference(upperResult.stdout, truth.stdout))
	os.Remove(upperArchive)
	os.Remove(upperMutant)
	for _, population := range []struct{ name, root, config, manifest string }{
		{"repository", repository, filepath.Join(repository, "tsconfig.json"), "repository.manifest"},
		{"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), "compiler.manifest"},
	} {
		if population.root == "" {
			t.Fatal("compiler source required")
		}
		data, e := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage", population.manifest))
		if e != nil {
			t.Fatal(e)
		}
		var roots []string
		for _, path := range strings.Split(string(data), "\n") {
			if path != "" {
				roots = append(roots, filepath.Join(population.root, path))
			}
		}
		m := h.write(population.name+".manifest", strings.Join(roots, "\n")+"\n")
		h.compare(population.name, oracle, binary, population.config, m)
		h.compare(population.name+"-asan", oracle, asan, population.config, m)
		for _, impl := range []struct{ name, path string }{{"go", oracle}, {"native", binary}} {
			command := exec.Command(impl.path, population.config, m)
			command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
			r := h.must(population.name+"-timed-"+impl.name, command)
			t.Logf("%s %s process %s %s", population.name, impl.name, r.elapsed, strings.TrimSpace(string(r.stderr)))
		}
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic'; const args=programArguments(); const file=args[1]??''; const p=tsgoProgram(args[0]??'',[file]); tsgoRelease(p); console.log(tsgoInspect(p,file,0,3,'SourceFile','react-syntax-details'));`)
	probe := h.write("probe.tsx", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got = h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released escaped %v %s", got.err, got.stderr)
	}
	t.Log("released handle rejected with panic 70")
	overlay = h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released program.")
	mutantArchive = h.archive("released-registry", overlay, false)
	mutant = h.build(stage0, "released-mutant", released, mutantArchive, false)
	h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released registry mutant exits 0, caught by required panic 70")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
