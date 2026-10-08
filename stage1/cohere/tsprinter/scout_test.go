package tsprinter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type scoutFixture struct{ Label, Source, Path, Origin, Go, Prettier, Reason string }

func TestScoutSideEffectImports(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("scout/fixtures.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []scoutFixture
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 69 {
		t.Fatalf("fixture count %d, want 69", len(cases))
	}
	directory := t.TempDir()
	root, _ := filepath.Abs(repository)
	fixtures, _ := filepath.Abs("scout/fixtures.json")
	side, _ := filepath.Abs("testdata/scout_side_test.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{root + "/cohere/internal/format/javascript/scout_witnesses_test.go": side}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err = os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	command := bounded(t, "go", "test", "-count=1", "-overlay="+overlayPath, "-run=^TestScoutFormatterWitnesses$", "./internal/format/javascript")
	command.Dir = root + "/cohere"
	command.Env = append(os.Environ(), "ADAMIC_SCOUT_FIXTURES="+fixtures, "ADAMIC_SCOUT_ANSWERS="+directory+"/go.json")
	if output, err := combinedOutput(command); err != nil {
		t.Fatalf("Go oracle: %v %s", err, output)
	}
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		t.Fatal("ADAMIC_TS_PRETTIER must name prettier@3.9.6; no scout oracle skips")
	}
	script, _ := filepath.Abs("scout/prettier.mjs")
	result := execute(t, nil, "node", script, library, fixtures, directory+"/npm.json")
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("Prettier: exit %d stderr %s", result.exitCode, result.stderr)
	}
	npm, err := os.ReadFile(directory + "/npm.json")
	if err != nil {
		t.Fatal(err)
	}
	var fresh []scoutFixture
	if err = json.Unmarshal(npm, &fresh); err != nil {
		t.Fatal(err)
	}
	if len(fresh) != len(cases) {
		t.Fatal("npm lost cases")
	}
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	var batch, want strings.Builder
	accepted, real := 0, 0
	for i, c := range cases {
		if fresh[i].Prettier != c.Prettier {
			t.Fatalf("%s upstream bytes changed", c.Label)
		}
		if strings.HasPrefix(c.Label, "real-source-") {
			real++
		}
		batch.WriteString(c.Path + "\t" + escape.Replace(c.Source) + "\n")
		if c.Reason == "" {
			accepted++
			if c.Go != c.Prettier {
				t.Fatalf("accepted upstream mismatch %s", c.Label)
			}
			want.WriteString("ok\t" + escape.Replace(c.Go) + "\n")
		} else {
			want.WriteString("notyet\t" + c.Reason + "\n")
		}
	}
	if real != 30 || accepted < 20 {
		t.Fatalf("fixture families missing: real=%d accepted=%d", real, accepted)
	}
	path := filepath.Join(directory, "batch.txt")
	if err = os.WriteFile(path, []byte(batch.String()), 0644); err != nil {
		t.Fatal(err)
	}
	entry, _ := filepath.Abs("scoutMain.ts")
	program := lowered(t, entry)
	native, binary := natively(t, program, path)
	compare := func(name string, result run) {
		t.Helper()
		if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != want.String() {
			t.Fatalf("%s exit %d stderr %s difference %s", name, result.exitCode, result.stderr, firstDifference(string(result.stdout), want.String()))
		}
	}
	compare("Node", onNode(t, entry, path))
	compare("native", native)
	compare("JavaScript backend", onJavaScriptBackend(t, program, path))
	if report := leaks(t, program, binary, path); report != "" {
		t.Fatal(report)
	}
	changes := []mutation{
		{"import keyword lost", "expressions.ts", "this.docs.text('import '),", "this.docs.text('export '),", "scoutMain.ts"},
		{"import separator lost", "expressions.ts", "this.docs.text('import '),", "this.docs.text('import'),", "scoutMain.ts"},
		{"import semicolon lost", "expressions.ts", "this.print(this.child(index, 0)),\n                    this.docs.text(';'),", "this.print(this.child(index, 0)),\n                    this.docs.text(''),", "scoutMain.ts"},
		{"import attributes silently dropped", "syntax.ts", "if(node.children.length !== 1) return 'ImportDeclaration';", "if(node.children.length === 0) return 'ImportDeclaration';", "scoutMain.ts"},
	}

	for _, change := range changes {
		t.Run(change.name, func(t *testing.T) {
			mutated := mutatedPort(t, change)
			p := lowered(t, mutated)
			for _, side := range []struct {
				name   string
				result run
			}{{"Node", onNode(t, mutated, path)}, {"native", nativelyRun(t, p, path)}} {
				if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
					t.Fatalf("%s mutant must finish normally: %d %s", side.name, side.result.exitCode, side.result.stderr)
				}
				if string(side.result.stdout) == want.String() {
					t.Fatalf("%s mutant survived", side.name)
				}
				t.Logf("%s killed by bytes: %s", side.name, firstDifference(string(side.result.stdout), want.String()))
			}
		})
	}
	t.Logf("%d fixtures, %d accepted, %d full pinned real sources; Go/npm/Node/native/backend and four mutants", len(cases), accepted, real)
}
