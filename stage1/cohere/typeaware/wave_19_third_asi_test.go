package typeaware

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWave19AwaitInitializerASI(t *testing.T) {
	directory := os.Getenv("ADAMIC_WAVE19_AWAIT_ARTIFACTS")
	if directory == "" {
		t.Skip("run after TestWave19AwaitRule with persistent artifacts")
	}
	repository, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	var paths []string
	for i, source := range []string{"class C{private a\nasync [b](){return 1;}}", "class C{private 'a'\nasync [b](){return 1;}}", "class C{private a?\nasync [b](){return 1;}}", "class C{private a=0\nasync [b](){return 1;}}"} {
		paths = append(paths, h.write(string(rune('a'+i))+"-asi.a", "declare const b:string;"+source+"\nexport {};\n"))
	}
	manifest := h.write("asi.manifest", strings.Join(paths, "\n")+"\n")
	config := filepath.Join(directory, "tsconfig.json")
	oracle := filepath.Join(directory, "await-oracle")
	stage0 := filepath.Join(directory, "adamic")
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_19_third/await_main.a")
	binary := h.build(stage0, "await-asi", entry, filepath.Join(directory, "ancestry.a"), false)
	want := h.compare("asi", oracle, binary, config, manifest)
	sanitized := h.build(stage0, "await-asi-asan", entry, filepath.Join(directory, "ancestry-asan.a"), true)
	h.compare("asi-asan", oracle, sanitized, config, manifest)
	root := cloneWave19Third(h, "initializer", "await_main.a", "flags[index]=initialized?1:0;", "flags[index]=1;")
	mutant := h.build(stage0, "initializer-mutant", filepath.Join(root, "await_main.a"), filepath.Join(directory, "ancestry.a"), false)
	got := h.must("initializer-mutant-run", exec.Command(mutant, config, manifest))
	if len(got.stderr) != 0 || bytes.Equal(got.stdout, want.stdout) {
		t.Fatal("initializer mutant survived")
	}
	t.Logf("initializer mutant exits 0 with empty stderr; only Go bytes catch byte %d", firstDifference(got.stdout, want.stdout))
}
