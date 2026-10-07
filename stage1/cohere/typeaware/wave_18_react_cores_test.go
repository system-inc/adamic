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

// This validates native cores on prepared Go HIR. Source lowering, unit gates
// and memo annotations remain test-provider dependencies, not native coverage.
func TestWave18ReactCores(t *testing.T) {
	repository, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	directory := os.Getenv("ADAMIC_WAVE18_REACT_CORE_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if e := os.MkdirAll(directory, 0755); e != nil {
		t.Fatal(e)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	oracle := volumeOracle(h, "react-core-oracle", "oracle_wave_18_react_cores.go")
	// Exactly the declaration shapes used by production Go's tests: Dispatch is
	// an alias; RefObject is an interface. Both module specifier forms resolve.
	declarations := `export type SetStateAction<S> = S | ((prev: S) => S);
export type Dispatch<A> = (value: A) => void;
export interface RefObject<T> { current: T }
export declare function useState<S>(initial?: S): [S, Dispatch<SetStateAction<S>>];
export declare function useEffect(callback: () => void, deps?: unknown[]): void;
export declare function useLayoutEffect(callback: () => void, deps?: unknown[]): void;
export declare function useInsertionEffect(callback: () => void, deps?: unknown[]): void;
export declare function useEffectEvent<T extends Function>(callback:T):T;
export declare function useRef<T>(initial: T): RefObject<T>;
export declare function useCallback<T>(callback: T, deps: unknown[]): T;
export declare function useMemo<T>(callback: () => T, deps: unknown[]): T;
export type ActionDispatch<A extends unknown[]> = (...args:A)=>void;
export declare function useReducer<S,A>(r:(s:S,a:A)=>S,i:S):[S,ActionDispatch<[A]>];
`
	h.write("react.d.ts", declarations)
	h.write("ambient.d.ts", "declare module 'react' {\n"+strings.ReplaceAll(declarations, "export declare ", "export ")+"}\n")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","moduleDetection":"force","lib":["ES2022"],"jsx":"preserve","noEmit":true},"include":["*.d.ts"]}`)
	texts := map[string]bool{}
	for _, name := range []string{"set_state_in_effect", "set_state_in_render", "static_components"} {
		tree, e := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/react", name+"_test.go"), nil, 0)
		if e != nil {
			t.Fatal(e)
		}
		goast.Inspect(tree, func(n goast.Node) bool {
			lit, ok := n.(*goast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			s, e := strconv.Unquote(lit.Value)
			if e == nil && strings.Contains(s, "function ") && (strings.Contains(s, "useState") || strings.Contains(s, "Component") || strings.Contains(s, "useEffect")) && !strings.Contains(s, "declare module") {
				texts[s] = true
			}
			return true
		})
	}
	// Independent controls for instruction operands, capture translation,
	// post-dominance and ref control/taint, without guessing expected findings.
	for _, body := range []string{
		"setS(1);return s;", "if(props.c)return null;setS(1);return s;", "if(props.c)setS(1);return s;", "try{setS(1);}catch(e){}return s;",
		"const f=()=>setS(1);const g=()=>f();g();return s;", "useMemo(()=>{if(props.c)setS(1);return s;},[]);return s;", "setS?.(1);return s;",
		"useEffect(()=>{setS(1);});return s;", "const f=()=>setS(1);const g=()=>f();useEffect(()=>g());return s;",
		"const r=useRef(0);useEffect(()=>{setS(r.current+1);});return s;", "const r=useRef(0);useEffect(()=>{if(r.current)setS(1);});return s;",
		"const r=useRef(0);useEffect(()=>{const x=r.current;setS(1);});return s;", "const r=useRef({height:0});useEffect(()=>{const {height}=r.current;setS(height);});return s;",
		"const f=useCallback(()=>setS(1),[]);useEffect(()=>f());return s;", "useEffect(()=>{const f=()=>setS(1);f();});return s;",
		"for(const x of props){if(props.c)break;else throw Error('x');}setS(1);return s;",
	} {
		texts["import {useState,useEffect,useMemo,useCallback,useRef} from './react';function Component(props){const [s,setS]=useState(0);"+body+"}\n"] = true
	}
	keys := make([]string, 0, len(texts))
	for s := range texts {
		keys = append(keys, s)
	}
	sort.Strings(keys)
	var paths []string
	var loopPath string
	for i, s := range keys {
		path := h.write(fmt.Sprintf("fixture-%03d.tsx", i), s+"\n")
		paths = append(paths, path)
		if strings.Contains(s, "for(const x of props)") {
			loopPath = path
		}
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("source-filter", exec.Command(oracle, config, manifest, "--valid-sources"))
	paths = strings.Split(strings.TrimSpace(string(valid.stdout)), "\n")
	manifest = h.write("valid.manifest", strings.Join(paths, "\n")+"\n")
	t.Logf("%d/%d valid production-fixture and independent controls", len(paths), len(keys))
	modules := filepath.Join(repository, "stage1/cohere/typeaware/wave_18_react_partial/cores")

	sanitized := h.archive("checker-asan", "", true)
	type batch struct {
		generated result
		want      result
	}
	var batches []batch
	var nativeSeconds, goSeconds, prepareSeconds float64
	for from := 0; from < len(paths); from += 8 {
		end := min(from+8, len(paths))
		name := fmt.Sprintf("batch-%02d", from/8)
		list := h.write(name+".manifest", strings.Join(paths[from:end], "\n")+"\n")
		generated := h.must(name+"-prepared-hir", exec.Command(oracle, config, list, "--native-graphs", modules))
		entry := h.write(name+".a", string(generated.stdout))
		binary := h.build(stage0, name+"-native", entry, archive, false)
		want := h.compare(name, oracle, binary, config, list)
		asan := h.build(stage0, name+"-asan", entry, sanitized, true)
		h.compare(name+"-asan", oracle, asan, config, list)
		timed := h.must(name+"-timed-native", exec.Command(binary))
		if !bytes.Equal(timed.stdout, want.stdout) {
			t.Fatal("timed findings differ")
		}
		nativeSeconds += timed.elapsed.Seconds()
		goSeconds += want.elapsed.Seconds()
		prepareSeconds += generated.elapsed.Seconds()
		batches = append(batches, batch{generated, want})
	}
	for _, change := range []struct{ name, file, from, to string }{
		{"render-unconditional", "set_state_in_render.a", "if(!unconditional.has(block.id))", "if(unconditional.has(block.id))"},
		{"effect-setter", "set_state_in_effect.a", "if(!local.has(i.p0.id) && !function_.setter(i.p0.id))", "if(!local.has(i.p0.id) && function_.setter(i.p0.id))"},
		{"static-creator", "static_components.a", "dynamic.set(target, i.p0.id);", "dynamic.delete(target);"},
	} {

		marker := ""
		if strings.HasPrefix(change.name, "render") {
			marker = "setStateInRender"
		} else if strings.HasPrefix(change.name, "effect") {
			marker = "setStateInEffect"
		} else {
			marker = "staticComponents"
		}
		var selected *batch
		for i := range batches {
			if bytes.Contains(batches[i].want.stdout, []byte("\t"+marker+"\t")) {
				selected = &batches[i]
				break
			}
		}
		if selected == nil {
			t.Fatal("no positive mutant control", change.name)
		}
		generated, want := selected.generated, selected.want
		data, e := os.ReadFile(filepath.Join(modules, change.file))
		if e != nil {
			t.Fatal(e)
		}
		if strings.Count(string(data), change.from) != 1 {
			t.Fatal("nonunique mutant", change.name)
		}
		source := strings.Replace(string(data), change.from, change.to, 1)
		source = strings.ReplaceAll(source, "'./", "'"+modules+"/")
		source = strings.ReplaceAll(source, "'../", "'"+filepath.Dir(modules)+"/")
		mutantFile := h.write(change.name+"_mutant.a", source)
		main := strings.Replace(string(generated.stdout), strconv.Quote(filepath.Join(modules, change.file)), strconv.Quote(mutantFile), 1)
		mutant := h.build(stage0, change.name+"-mutant", h.write(change.name+"_main.a", main), archive, false)
		got := h.must(change.name+"-mutant-run", exec.Command(mutant))
		if len(got.stderr) > 0 || bytes.Equal(got.stdout, want.stdout) {
			t.Fatal("core mutant survived or failed outside comparison", change.name)
		}
		t.Logf("%s: exit 0, empty stderr; independent complete-byte comparison catches byte %d", change.name, firstDifference(got.stdout, want.stdout))
	}

	t.Logf("91 controls: prepared-input native wall %.6fs; full Go wall %.6fs; Go fixture preparation %.6fs. Different input pipelines, not an end-to-end native speed comparison", nativeSeconds, goSeconds, prepareSeconds)
	// Recreate the discovered adjacency-row alias bug as a byte-only mutant.
	var loop *batch
	for i := range batches {
		if bytes.Contains(batches[i].want.stdout, []byte(loopPath)) {
			loop = &batches[i]
			break
		}
	}
	if loop == nil || loopPath == "" {
		t.Fatal("missing loop mutant control")
	}
	data, e := os.ReadFile(filepath.Join(modules, "postdominator.a"))
	if e != nil {
		t.Fatal(e)
	}
	if strings.Count(string(data), "const fresh: number[] = [];") != 1 {
		t.Fatal("postdominator mutation anchor")
	}
	source := strings.Replace(string(data), "const fresh: number[] = [];", "const fresh: number[] = this.empty;", 1)
	source = strings.ReplaceAll(source, "'./", "'"+modules+"/")
	postdom := h.write("postdominator_alias_mutant.a", source)
	main := string(loop.generated.stdout)
	for _, file := range []string{"set_state_in_render.a", "set_state_in_effect.a"} {
		data, e := os.ReadFile(filepath.Join(modules, file))
		if e != nil {
			t.Fatal(e)
		}
		source := strings.Replace(string(data), "'./postdominator.a'", strconv.Quote(postdom), 1)
		source = strings.ReplaceAll(source, "'./", "'"+modules+"/")
		source = strings.ReplaceAll(source, "'../", "'"+filepath.Dir(modules)+"/")
		copy := h.write("alias_mutant_"+file, source)
		main = strings.Replace(main, strconv.Quote(filepath.Join(modules, file)), strconv.Quote(copy), 1)
	}
	alias := h.build(stage0, "adjacency-alias-mutant", h.write("adjacency_alias_main.a", main), archive, false)
	wrong := h.must("adjacency-alias-mutant-run", exec.Command(alias))
	if len(wrong.stderr) > 0 || bytes.Equal(wrong.stdout, loop.want.stdout) {
		t.Fatal("adjacency alias mutant survived or failed outside comparison")
	}
	t.Logf("adjacency-row alias mutant: exit 0, empty stderr; byte comparison catches byte %d", firstDifference(wrong.stdout, loop.want.stdout))
	for _, corpus := range []struct{ name, config, manifest string }{{"compiler", os.Getenv("ADAMIC_WAVE18_COMPILER_CONFIG"), os.Getenv("ADAMIC_WAVE18_COMPILER_MANIFEST")}, {"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE18_REPOSITORY_MANIFEST")}} {
		if corpus.manifest != "" {
			prepared := h.must(corpus.name+"-prepared-hir", exec.Command(oracle, corpus.config, corpus.manifest, "--native-graphs", modules))
			entry := h.write(corpus.name+"_prepared_hir.a", string(prepared.stdout))
			binary := h.build(stage0, corpus.name+"-prepared-native", entry, archive, false)
			h.compare(corpus.name+"-prepared", oracle, binary, corpus.config, corpus.manifest)
			asan := h.build(stage0, corpus.name+"-prepared-asan", entry, sanitized, true)
			h.compare(corpus.name+"-prepared-asan", oracle, asan, corpus.config, corpus.manifest)
		}
	}
}
