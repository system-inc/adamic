package typeaware

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: this native archive/sanitizer test uses the shared scratch budget.
func TestWave21Enums(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE21_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_21_suite.a")
	binary := h.build(stage0, "wave21", entry, archive, false)
	oracle := volumeOracle(h, "wave21-oracle", "oracle_wave_21.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for i, source := range []string{
		"enum E { A, B='b', C='c' }",
		"enum E { A='a', B=(1) }",
		"enum E { A=1, B=false, C='c' }",
		"enum E { A='a', B=(false) }",
		"enum E { A='a', B }",
		"enum E { A=1 } enum E { B='b' } enum E { C=2 }",
		"enum E { Z=false } enum E { A=1, B='b' }",
		"enum E {} enum E { A=1, B='b' }",
		"namespace N { export enum E { A=1 } } namespace N { export enum E { B='b' } }",
		"namespace N { enum E { A=1 } } namespace N { enum E { B='b' } }",
		"declare const s:string; enum E { A=1, B=s }",
		"declare const f:()=>any; enum E { A=f(), B=1, C='c' }",
		"enum E { A='a', B=`b${1}` }",
		"enum E { A=1, B=null, C='c' }",
		"/* 世界 🌍 */\r\nenum 漢 { A='a', B=(1) }\r\n",
	} {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	helper := h.write("helper.a", "export enum Imported { A='a' }\n")
	paths = append(paths, helper, h.write("augmentation.a", "import {Imported} from './helper.a'; declare module './helper.a' { enum Imported { B=1 } }\n"))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	if !bytes.Contains(truth.stdout, []byte("\tno-mixed-enums\t")) && !bytes.Contains(truth.stdout, []byte("\t@typescript-eslint/no-mixed-enums\t")) {
		t.Fatal("no positive enum control")
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave21-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	// Reversing the enum comparison compiles and exits normally; only diagnostic bytes catch it.
	original, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/no_mixed_enums.a"))
	if err != nil {
		t.Fatal(err)
	}
	mutantSource := strings.Replace(string(original), "current !== desired", "current === desired", 1)
	// Keep original sibling imports absolute when building the scratch mutant.
	for _, name := range []string{"rules.ts", "enum_declarations.a", "facts.ts", "flags.ts"} {
		mutantSource = strings.ReplaceAll(mutantSource, "'./"+name+"'", "'"+filepath.Join(repository, "stage1/cohere/typeaware", name)+"'")
	}
	mutantFile := h.write("no_mixed_enums_mutant.a", mutantSource)
	main, err := os.ReadFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	mainSource := strings.Replace(string(main), "'./no_mixed_enums.a'", "'"+mutantFile+"'", 1)
	for _, name := range []string{"unary_minus.ts", "rules.ts"} {
		mainSource = strings.ReplaceAll(mainSource, "'./"+name+"'", "'"+filepath.Join(repository, "stage1/cohere/typeaware", name)+"'")
	}
	mainSource = strings.ReplaceAll(mainSource, "'../../typescript/", "'"+filepath.Join(repository, "stage1/typescript")+"/")
	mutant := h.build(stage0, "enum-mutant", h.write("enum_mutant_main.a", mainSource), archive, false)
	got := h.must("enum-mutant-run", exec.Command(mutant, config, manifest))
	if len(got.stderr) > 0 || bytes.Equal(got.stdout, truth.stdout) {
		t.Fatal("enum mutant survived or failed outside comparison")
	}
	t.Logf("enum mutant: exit 0, byte oracle catches byte %d", firstDifference(got.stdout, truth.stdout))
	releasedSource := h.write("released_enum.a", `import {panic,programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
 const a=programArguments();const file=a[1]??panic('file');const p=tsgoProgram(a[0]??panic('config'),[file]);tsgoRelease(p);console.log(tsgoInspect(p,file,0,14,'EnumDeclaration','enum-declarations'));`)
	probe := h.write("released_probe.a", "enum E { A=1 }")
	stale := h.build(stage0, "released-enum", releasedSource, archive, false)
	observed := h.run("released-enum-run", exec.Command(stale, config, probe))
	if code, ok := observed.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(observed.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released enum question escaped: %v %s", observed.err, observed.stderr)
	}
	t.Log("released enum query: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-enum-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released handle.")
	staleArchive := h.archive("released-enum-registry", overlay, false)
	staleMutant := h.build(stage0, "released-enum-mutant", releasedSource, staleArchive, false)
	observed = h.must("released-enum-mutant-run", exec.Command(staleMutant, config, probe))
	t.Log("released registry mutant: exit 0 caught by required panic 70")

	for _, corpus := range []struct{ name, config, manifest string }{
		{"compiler", os.Getenv("ADAMIC_WAVE21_COMPILER_CONFIG"), os.Getenv("ADAMIC_WAVE21_COMPILER_MANIFEST")},
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE21_REPOSITORY_MANIFEST")},
	} {
		if corpus.manifest != "" {
			h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
			h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		}
	}
}
