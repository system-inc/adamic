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

func wave11SeventhUpstream(t *testing.T, repository string) []string {
	t.Helper()
	var sources []string
	for _, name := range []string{"jsx_fragments", "jsx_no_undef", "no_array_index_key"} {
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

func TestWave11SeventhAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE_11_SEVENTH_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave-11-seventh/main.a")
	binary := h.build(stage0, "seventh", entry, archive, false)
	oracle := volumeOracle(h, "seventh-oracle", "../wave-11-seventh/testdata/oracle.go")
	listenerOracle := volumeOracle(h, "seventh-listeners", "../wave-11-seventh/testdata/listeners.go")
	listenerConfig := h.write("listeners-config.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","jsx":"react"}}`)
	listenerSource := h.write("listeners.tsx", "export {};\n")
	listenerTruth := h.must("listeners", exec.Command(listenerOracle, listenerConfig, listenerSource))
	var listenerLines []string
	for _, directory := range []string{"react-jsx-fragments", "react-jsx-no-undef", "react-no-array-index-key"} {
		data, e := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/wave-11-seventh/rules", directory, "rule.json"))
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
		`declare const React:any; <React.Fragment>text</React.Fragment>; <React.Fragment> </React.Fragment>;`,
		`declare const React:any; <React.Fragment/>; <React.Fragment key="x"/>; <><Missing/></>; <div/>; <a/>; <Map/>; <Foo.Bar/>;`,
		`import {Fragment as F} from 'react'; <F/>; <F key="x"/>; import {Fragment as Other} from 'other'; <Other/>;`,
		`const F=React; <F/>; const G=React.Fragment; <G/>; const {Fragment:H}=React; <H/>; const I=require('react'); <I/>;`,
		`declare const React:any; declare const X:any; declare const xs:any; xs.map((x:any,i:number)=><X key={i}/>); xs.map((x:any,i:number)=><X key={String(i)}/>); xs.map((x:any,i:number)=><X key={i.toString()}/>);`,
		"declare const xs:any; xs.map((x:any,i:number)=><X key={`${i}${i}`}/>); xs.map((x:any,i:number)=><X key={i+i}/>); xs.map((x:any,i:number)=><X key={i||x.id}/>);",
		`declare const xs:any; xs.reduce((acc:any,x:any,i:number)=><X key={i}/>); xs.map((x:any,i=0)=><X key={i}/>); xs.map((x:any,...i:any)=><X key={i}/>);`,
		`declare const React:any; declare const xs:any; xs.map((x:any,i:number)=>React.createElement('div',{key:i})); React.Children.map(xs,(x:any,i:number)=><X key={i}/>); Children.forEach(xs,(x:any,i:number)=><X key={i}/>);`,
		`import {createElement as h} from 'react'; declare const xs:any; xs.map((x:any,i:number)=>h('div',{key:i}));`,
		"/* 世界 🌍 */\r\n<Missing/>; <é/>; <_Missing/>; <x-widget/>; <ns:tag/>;",
		`declare const Foo:any; <Foo.Bar/>; <foo.Bar/>; declare const foo:any; <foo.Bar/>;`,
	}
	var paths []string
	for i, source := range append(controls, wave11SeventhUpstream(t, repository)...) {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.tsx", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("valid", exec.Command(oracle, config, manifest, "--valid-sources"))
	t.Logf("controls %d parse-clean %d", len(paths), len(strings.Fields(string(valid.stdout))))
	manifest = h.write("controls-valid.manifest", string(valid.stdout))
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"react/jsx-fragments", "react/jsx-no-undef", "react/no-array-index-key"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive: %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "seventh-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"fragment-attributes", "rules/react-jsx-fragments/rule.a", "attributes.children.length > 0", "attributes.children.length > 1"},
		{"undefined-intrinsic", "rules/react-jsx-no-undef/rule.a", "first >= 97", "first >= 98"},
		{"index-name", "rules/react-no-array-index-key/rule.a", "this.indices.includes(node.text)", "this.indices.includes(node.text + '!')"},
	} {
		root := filepath.Join(directory, change.name+"-source")
		sourceRoot := filepath.Join(repository, "stage1/cohere/typeaware/wave-11-seventh")
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
	overlay := h.overlay("foreign-binding", "bridge/tsgo/checker/numeric_syntax_bindings.go", "out.text(source.FileName().AsString())", "out.text(source.FileName().AsString()[:0] + root.AsSourceFile().FileName().AsString())")
	mutantArchive := h.archive("foreign-binding", overlay, false)
	mutant := h.build(stage0, "foreign-binding-mutant", entry, mutantArchive, false)
	got := h.must("foreign-binding-mutant-run", exec.Command(mutant, config, manifest))
	if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
		t.Fatal("raw binding mutant survived")
	}
	t.Logf("raw binding mutant exit 0 empty stderr, Go bytes catch byte %d", firstDifference(got.stdout, truth.stdout))
	os.Remove(mutantArchive)
	os.Remove(mutant)
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
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic'; const args=programArguments(); const file=args[1]??''; const p=tsgoProgram(args[0]??'',[file]); tsgoRelease(p); console.log(tsgoInspect(p,file,0,3,'SourceFile','numeric-syntax-bindings'));`)
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
