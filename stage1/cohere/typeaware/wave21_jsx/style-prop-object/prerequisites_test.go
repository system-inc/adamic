package wave21style

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestStyleSourcePrerequisite(t *testing.T) {
	directory := os.Getenv("ADAMIC_WAVE21_STYLE_ARTIFACTS")
	if directory == "" {
		t.Skip("run the style rule suite with retained artifacts first")
	}
	repository, e := filepath.Abs("../../../../..")
	if e != nil {
		t.Fatal(e)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	input := h.write("jsx-source-witness.tsx", "<div style=\"bad\" />;\n")
	manifest := h.write("jsx-source-witness.manifest", input+"\n")
	truth := h.must("jsx-source-go", exec.Command(filepath.Join(directory, "oracle"), filepath.Join(directory, "config.json"), manifest))
	if !bytes.Contains(truth.stdout, []byte("\tstylePropNotObject\t")) {
		t.Fatal("missing Go source positive", truth.stdout)
	}
	source := "import {Parser} from " + strconv.Quote(filepath.Join(repository, "stage1/typescript/parser/parser.ts")) + ";\nimport {readTextFile,panic,programArguments} from 'adamic';\nconst path=programArguments()[0]??panic('missing input');const data=readTextFile(path);if(data.kind==='Error'){panic(data.message);}const parser=new Parser(data.text,path);parser.file();console.log('parsed');\n"
	for _, kind := range []struct {
		name, archive string
		sanitize      bool
	}{{"source-normal", filepath.Join(directory, "checker.a"), false}, {"source-asan", filepath.Join(directory, "checker-asan.a"), true}} {
		binary := h.build(stage0, kind.name, h.write(kind.name+".a", source), kind.archive, kind.sanitize)
		got := h.run(kind.name+"-run", exec.Command(binary, input))
		code, ok := got.err.(*exec.ExitError)
		if !ok || code.ExitCode() != 70 || !strings.HasPrefix(string(got.stderr), "adamic: panic: parser slice expected GreaterThanToken, got Identifier at 5 in ") {
			t.Fatal("JSX source prerequisite", got.err, got.stderr)
		}
		t.Logf("Go positive, %s native refusal: %s", kind.name, strings.TrimSpace(string(got.stderr)))
	}
	mutated := strings.Replace(source, "parser.file();", "", 1)
	binary := h.build(stage0, "source-skip-mutant", h.write("source_skip_mutant.a", mutated), filepath.Join(directory, "checker.a"), false)
	got := h.must("source-skip-mutant-run", exec.Command(binary, input))
	if len(got.stderr) != 0 || string(got.stdout) != "parsed\n" {
		t.Fatal("skip mutant failed outside assertion", got.stderr)
	}
	t.Log("source-skip prerequisite mutant exits 0; required refusal assertion catches it; separate from the rule mutant")
}
