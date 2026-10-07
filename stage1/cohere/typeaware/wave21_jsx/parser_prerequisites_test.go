package wave21jsx

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Prerequisite coverage only. A refused parse does not prove a rule port.
func TestJSXSourcePrerequisites(t *testing.T) {
	repository, e := filepath.Abs("../../../..")
	if e != nil {
		t.Fatal(e)
	}
	artifacts := os.Getenv("ADAMIC_WAVE21_JSX_ARTIFACTS")
	if artifacts == "" {
		t.Skip("run the prepared JSX suite first with an artifact directory")
	}
	directory := filepath.Join(artifacts, "parser-prerequisites")
	if e = os.MkdirAll(directory, 0755); e != nil {
		t.Fatal(e)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	source := `import {panic,readTextFile,programArguments} from 'adamic';import {Parser} from ` + strconv.Quote(filepath.Join(repository, "stage1/typescript/parser/parser.ts")) + `;const args=programArguments();const path=args[0]??panic('source');const input=readTextFile(path);if(input.kind==='Error'){panic(input.message);}const parser=new Parser(input.text,path);const root=parser.file();console.log(` + "`parsed ${parser.node(root).children.length}`" + `);`
	entry := h.write("parser.a", source)
	stage0 := filepath.Join(artifacts, "adamic")
	oracle := filepath.Join(artifacts, "oracle")
	config := filepath.Join(artifacts, "config.json")
	normal := h.build(stage0, "parser-normal", entry, filepath.Join(artifacts, "checker.a"), false)
	asan := h.build(stage0, "parser-asan", entry, filepath.Join(artifacts, "checker-asan.a"), true)
	for _, control := range []struct{ name, source, id string }{{"jsx-fragments", "<React.Fragment />;", "preferFragment"}, {"jsx-no-undef", "<Missing />;", "jsxIdentifierNotDefined"}, {"jsx-no-constructed-context-values", "function Component(){return <Ctx.Provider value={{}}/>;}", "defaultMsg"}} {
		input := h.write(control.name+".tsx", control.source+"\n")
		manifest := h.write(control.name+".manifest", input+"\n")
		want := h.must(control.name+"-go", exec.Command(oracle, config, manifest))
		if !bytes.Contains(want.stdout, []byte("\t"+control.id+"\t")) {
			t.Fatalf("Go positive absent: %s", want.stdout)
		}
		for _, binary := range []struct{ name, path string }{{"normal", normal}, {"asan", asan}} {
			got := h.run(control.name+"-"+binary.name, exec.Command(binary.path, input))
			if got.err != nil || len(got.stderr) != 0 || !strings.HasPrefix(string(got.stdout), "parsed ") || string(got.stdout) == "parsed 0\n" {
				t.Fatalf("source prerequisite changed: %v %s %s", got.err, got.stdout, got.stderr)
			}
			t.Logf("%s %s: %s", control.name, binary.name, strings.TrimSpace(string(got.stdout)))
		}
	}
	mutant := h.build(stage0, "parser-skip-mutant", h.write("parser_skip_mutant.a", strings.Replace(source, "const root=parser.file();console.log(`parsed ${parser.node(root).children.length}`);", "console.log('parsed 0');", 1)), filepath.Join(artifacts, "checker.a"), false)
	got := h.must("parser-skip-run", exec.Command(mutant, filepath.Join(directory, "jsx-fragments.tsx")))
	if string(got.stdout) != "parsed 0\n" || len(got.stderr) != 0 {
		t.Fatal("parser guard mutant failed outside the nonempty-tree check", got.stderr)
	}
	t.Log("parser-skip prerequisite mutant exits 0; nonempty parsed-tree assertion catches it; this is not a rule mutant")
}
