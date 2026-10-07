package typeaware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func wave18TitleBuild(h *harness, stage0, archive, slice, name, from, to string, sanitize bool) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*"))
	if err != nil {
		h.t.Fatal(err)
	}
	changed := from == ""
	for _, path := range files {
		extension := filepath.Ext(path)
		if extension != ".a" && extension != ".ts" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			h.t.Fatal(err)
		}
		source := string(data)
		if from != "" && filepath.Base(path) == "no_title_in_document_head.a" {
			if strings.Count(source, from) != 1 {
				h.t.Fatalf("nonunique mutant %s", name)
			}
			source = strings.Replace(source, from, to, 1)
			changed = true
		}
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(slice, "stage1/typescript")+"/")
		source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	if !changed {
		h.t.Fatal("mutant did not change a source")
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_18_title_suite.a"), archive, sanitize)
}

// Not parallel: archive and sanitizer builds use bounded scratch space.
func TestWave18TitleWithJsxSlice(t *testing.T) {
	slice := os.Getenv("ADAMIC_WAVE18_JSX_SLICE")
	if slice == "" {
		t.Skip("requires isolated JSX parser dependency; no shared parser files are modified")
	}
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE18_TITLE_ARTIFACTS")
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
	binary := wave18TitleBuild(h, stage0, archive, slice, "title", "", "", false)
	oracle := volumeOracle(h, "title-oracle", "oracle_wave_18_title.go")
	h.must("production-tests", exec.Command("go", "-C", "cohere", "test", "./internal/lint/rules/next", "-run", "^TestNoTitleInDocumentHead", "-count=1", "-v"))
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ESNext","module":"NodeNext","jsx":"preserve","lib":["ESNext"],"types":[],"moduleDetection":"force"},"files":["anchor.d.ts"]}`)
	h.write("anchor.d.ts", "export {};\n")
	data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/testdata/title_controls.json"))
	if err != nil {
		t.Fatal(err)
	}
	var controls []string
	if err = json.Unmarshal(data, &controls); err != nil {
		t.Fatal(err)
	}
	var paths []string
	for i, source := range controls {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.tsx", i), source))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	if !bytes.Contains(truth.stdout, []byte("\t@next/next/no-title-in-document-head\t")) {
		t.Fatal("missing positive control")
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := wave18TitleBuild(h, stage0, sanitized, slice, "title-asan", "", "", true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	mutant := wave18TitleBuild(h, stage0, archive, slice, "title-mutant", "title.text !== 'title'", "title.text === 'title'", false)
	got := h.must("title-mutant-run", exec.Command(mutant, config, manifest))
	if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
		t.Fatal("title mutant survived")
	}
	t.Logf("title mutant exit 0 empty stderr: byte oracle caught byte %d", firstDifference(got.stdout, truth.stdout))
	os.Remove(mutant)
	for _, name := range []string{"repository", "compiler"} {
		corpusManifest := os.Getenv("ADAMIC_WAVE18_" + strings.ToUpper(name) + "_MANIFEST")
		if corpusManifest == "" {
			continue
		}
		corpusConfig := filepath.Join(repository, "tsconfig.json")
		if name == "compiler" {
			corpusConfig = filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json")
		}
		h.compare(name, oracle, binary, corpusConfig, corpusManifest)
		h.compare(name+"-asan", oracle, asan, corpusConfig, corpusManifest)
		for round := 0; round < 3; round++ {
			for _, command := range []struct{ name, binary string }{{"native", binary}, {"go", oracle}} {
				cmd := exec.Command(command.binary, corpusConfig, corpusManifest, "--count")
				cmd.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
				got := h.must(fmt.Sprintf("%s-%d-%s-timing", name, round, command.name), cmd)
				t.Logf("%s round %d %s %.6fs %s", name, round, command.name, got.elapsed.Seconds(), strings.TrimSpace(string(got.stderr)))
			}
		}
	}
}
