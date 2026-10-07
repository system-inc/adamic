package typeaware

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// This checks prerequisites, not native rule agreement. No React rule is wired
// into the parser probe; its successful parse must never count as a rule pass.
func TestWave21ReactPrerequisites(t *testing.T) {
	repository, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	directory := os.Getenv("ADAMIC_WAVE21_REACT_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/testdata/wave_21_react_parser.a")
	probe := h.build(stage0, "parser-prerequisite", entry, archive, false)
	asanArchive := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "parser-prerequisite-asan", entry, asanArchive, true)
	oracle := volumeOracle(h, "react-prerequisite-oracle", "oracle_wave_21_react.go")
	h.write("react.d.ts", `export type SetStateAction<S> = S | ((prev: S) => S);
export type Dispatch<A> = (value: A) => void;
export interface RefObject<T> { current: T }
export declare function useState<S>(initial: S): [S, Dispatch<SetStateAction<S>>];
export declare function useEffect(callback: () => void, deps?: unknown[]): void;
export declare function useRef<T>(initial: T): RefObject<T>;
export declare function useCallback<T>(callback: T, deps: unknown[]): T;
export declare function useMemo<T>(callback: () => T, deps: unknown[]): T;
`)
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","moduleDetection":"force","lib":["ES2022"],"jsx":"preserve","noEmit":true},"include":["react.d.ts"]}`)
	var staticPath string
	for _, control := range []struct {
		name, source, id string
		jsx              bool
	}{
		{"effect-direct", `import {useState,useEffect} from './react';function Component(){const [state,setState]=useState(0);useEffect(()=>{setState(1);});return state;}`, "setStateInEffect", false},
		{"effect-captures", `import {useState,useEffect} from './react';function Component(){const [state,setState]=useState(0);const f=()=>{setState(1);};const g=()=>{f();};useEffect(()=>{g();});return state;}`, "setStateInEffect", false},
		{"render-direct", `import {useState} from './react';function Component(){const [state,setState]=useState(0);setState(1);return state;}`, "setStateInRender", false},
		{"render-captures", `import {useState} from './react';function Component(){const [state,setState]=useState(0);const f=()=>{setState(1);};const g=()=>{f();};g();return state;}`, "setStateInRender", false},
		{"static-direct", `function Example(props){const Component=createComponent();return <Component />;}`, "staticComponents", true},
		{"static-phi", `function Example(props){let Component;if(props.cond){Component=createComponent();}else{Component=DefaultComponent;}return <Component />;}`, "staticComponents", true},
	} {
		ext := ".a"
		if control.jsx {
			ext = ".tsx"
		}
		path := h.write(control.name+ext, control.source+"\n")
		list := h.write(control.name+".manifest", path+"\n")
		truth := h.must(control.name+"-go", exec.Command(oracle, config, list))
		if !bytes.Contains(truth.stdout, []byte("\t"+control.id+"\t")) {
			t.Fatalf("Go positive control did not fire: %s", truth.stdout)
		}
		if control.name == "static-direct" {
			staticPath = path
		}
		t.Logf("%s production Go: %s", control.name, summary(truth.stdout))
		for _, binary := range []struct{ name, path string }{{"normal", probe}, {"asan", asan}} {
			got := h.run(control.name+"-parser-"+binary.name, exec.Command(binary.path, path))
			if control.jsx {
				code, ok := got.err.(*exec.ExitError)
				if !ok || code.ExitCode() != 70 || !strings.Contains(string(got.stderr), "adamic: panic: parser slice expected") {
					t.Fatalf("JSX parser prerequisite changed: %v %s", got.err, got.stderr)
				}
				t.Logf("%s %s parser refusal: %s", control.name, binary.name, strings.TrimSpace(string(got.stderr)))
			} else if got.err != nil || string(got.stdout) != "parsed SourceFile\n" || len(got.stderr) > 0 {
				t.Fatalf("parser prerequisite: %v %s %s", got.err, got.stdout, got.stderr)
			}
		}
	}

	// Prerequisite guard mutants only; these do not count as native rule mutants.
	data, e := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/testdata/oracle_wave_21_react.go"))
	if e != nil {
		t.Fatal(e)
	}
	for _, change := range []struct{ name, rule, control, id string }{
		{"effect-positive", "react-hooks/set-state-in-effect", "effect-direct", "setStateInEffect"},
		{"render-positive", "react-hooks/set-state-in-render", "render-direct", "setStateInRender"},
		{"static-positive", "react-hooks/static-components", "static-direct", "staticComponents"},
	} {
		from := `"` + change.rule + `": true`
		if strings.Count(string(data), from) != 1 {
			t.Fatal("nonunique coverage mutant", change.name)
		}
		mutantSource := h.write(change.name+"_mutant.go", strings.Replace(string(data), from, `"`+change.rule+`": false`, 1))
		virtual := filepath.Join(repository, "cohere/adamic_react_prerequisite_mutant.go")
		overlay, e := json.Marshal(map[string]any{"Replace": map[string]string{virtual: mutantSource}})
		if e != nil {
			t.Fatal(e)
		}
		binary := filepath.Join(directory, change.name+"-mutant")
		build := exec.Command("go", "build", "-overlay", h.write(change.name+"-overlay.json", string(overlay)), "-o", binary, virtual)
		build.Dir = filepath.Join(repository, "cohere")
		h.must(change.name+"-mutant-build", build)
		result := h.must(change.name+"-mutant-run", exec.Command(binary, config, filepath.Join(directory, change.control+".manifest")))
		if bytes.Contains(result.stdout, []byte("\t"+change.id+"\t")) {
			t.Fatal("positive guard mutant survived", change.name)
		}
		t.Logf("%s prerequisite guard mutant: exit 0; required Go message disappeared", change.name)
	}
	parserSource, e := os.ReadFile(entry)
	if e != nil {
		t.Fatal(e)
	}
	// Skipping parsing would pretend the unsupported JSX input was accepted.
	mutantSource := strings.Replace(string(parserSource), "const root = parser.file();", "const root = 0;", 1)
	mutantSource = strings.Replace(mutantSource, "console.log(`parsed ${parser.node(root).kind}`);", "console.log('parsed SourceFile');", 1)
	mutantSource = strings.ReplaceAll(mutantSource, "'../../../typescript/", "'"+filepath.Join(repository, "stage1/typescript")+"/")
	mutant := h.build(stage0, "parser-guard-mutant", h.write("parser_guard_mutant.a", mutantSource), archive, false)
	skipped := h.must("parser-guard-mutant-run", exec.Command(mutant, staticPath))
	if len(skipped.stderr) != 0 {
		t.Fatal("parser guard mutant failed outside the refusal assertion")
	}
	t.Log("parser guard mutant: exit 0; required JSX refusal disappeared")
	for _, corpus := range []struct{ name, config, manifest string }{{"compiler", os.Getenv("ADAMIC_WAVE21_COMPILER_CONFIG"), os.Getenv("ADAMIC_WAVE21_COMPILER_MANIFEST")}, {"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE21_REPOSITORY_MANIFEST")}} {
		if corpus.manifest != "" {
			got := h.must(corpus.name+"-go", exec.Command(oracle, corpus.config, corpus.manifest))
			t.Logf("%s Go baseline only: %s; %d canonical bytes", corpus.name, summary(got.stdout), len(got.stdout))
		}
	}
}
