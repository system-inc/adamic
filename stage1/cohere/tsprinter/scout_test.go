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
	runScoutFixtures(t, "scout/fixtures.json", 69, 30, []mutation{
		{"import keyword lost", "expressions.ts", "const parts: number[] = [this.docs.text('import')];", "const parts: number[] = [this.docs.text('export')];", "scoutMain.ts"},
		{"import separator lost", "expressions.ts", "parts.push(this.docs.text(' '));\n        parts.push(this.print(this.child(index, offset), index));", "parts.push(this.docs.text(''));\n        parts.push(this.print(this.child(index, offset), index));", "scoutMain.ts"},
		{"import semicolon lost", "expressions.ts", "parts.push(this.docs.text(';'));\n        return this.docs.concat(parts);", "parts.push(this.docs.text(''));\n        return this.docs.concat(parts);", "scoutMain.ts"},
		{"import attributes silently dropped", "expressions.ts", "if(this.node(index).children.length > offset + 1) {", "if(this.node(index).children.length > offset + 2) {", "scoutMain.ts"},
	})
}

func TestScoutImportClausesAndAttributes(t *testing.T) {
	t.Parallel()
	runScoutFixtures(t, "scout/import-fixtures.json", 61, 0, []mutation{
		{"default binding lost", "expressions.ts", "else standalone.push(this.print(child, index));", "else standalone.push(this.docs.text('lost'));", "scoutMain.ts"},
		{"namespace alias lost", "expressions.ts", "this.docs.text('* as ')", "this.docs.text('* ')", "scoutMain.ts"},
		{"named specifier alias lost", "expressions.ts", "parts.push(this.docs.text(' as '));", "parts.push(this.docs.text(' '));", "scoutMain.ts"},
		{"clause type modifier lost", "expressions.ts", "if(phase === 'TypeKeyword') parts.push(this.docs.text(' type'));", "if(phase === 'TypeKeyword') parts.push(this.docs.text(''));", "scoutMain.ts"},
		{"specifier type modifier lost", "expressions.ts", "if(node.semantic === '1') parts.push(this.docs.text('type '));", "if(node.semantic === '1') parts.push(this.docs.text(''));", "scoutMain.ts"},
		{"attribute value lost", "expressions.ts", "this.print(this.child(child, 1), child),", `this.docs.text("'lost'"),`, "scoutMain.ts"},
		{"single type attribute flattening lost", "expressions.ts", "? this.docs.removeLines(content)", "? content", "scoutMain.ts"},
	})
}

func runScoutFixtures(t *testing.T, fixtureFile string, count, realCount int, changes []mutation) {
	t.Helper()
	data, err := os.ReadFile(fixtureFile)
	if err != nil {
		t.Fatal(err)
	}
	var cases []scoutFixture
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != count {
		t.Fatalf("fixture count %d, want %d", len(cases), count)
	}
	directory := t.TempDir()
	root, _ := filepath.Abs(repository)
	fixtures, _ := filepath.Abs(fixtureFile)
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
	if real != realCount || accepted < 20 {
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
	t.Logf("%d fixtures, %d accepted, %d full pinned real sources; Go/npm/Node/native/backend and %d mutants", len(cases), accepted, real, len(changes))
}
func TestScoutSourceContext(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("scout/parser-blockers.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []scoutFixture
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 4 {
		t.Fatal("shortest regression evidence missing")
	}
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	var batch, want strings.Builder
	for _, c := range cases[:2] {
		batch.WriteString("0" + escape.Replace(c.Source) + "\n")
		want.WriteString("ok\t" + escape.Replace(c.Go) + "\n")
	}
	path := filepath.Join(t.TempDir(), "shortest.txt")
	if err = os.WriteFile(path, []byte(batch.String()), 0644); err != nil {
		t.Fatal(err)
	}
	entry, _ := filepath.Abs("main.ts")
	program := lowered(t, entry)
	for _, side := range []struct {
		name   string
		result run
	}{{"Node", onNode(t, entry, "--cases", path, "120")}, {"native", nativelyRun(t, program, "--cases", path, "120")}, {"backend", onJavaScriptBackend(t, program, "--cases", path, "120")}} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != want.String() {
			t.Fatalf("%s shortest method regressions: %d %s %s", side.name, side.result.exitCode, side.result.stderr, firstDifference(string(side.result.stdout), want.String()))
		}
	}
	change := mutation{"source context lost", "expressions.ts", "parser.beginList('source');\n    const root = parser.expression();\n    parser.endList('source');", "const root = parser.expression();", "main.ts"}
	mutated := mutatedPort(t, change)
	p := lowered(t, mutated)
	for _, side := range []struct {
		name   string
		result run
	}{{"Node", onNode(t, mutated, "--cases", path, "120")}, {"native", nativelyRun(t, p, "--cases", path, "120")}} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
			t.Fatalf("%s context mutant must finish normally: %d %s", side.name, side.result.exitCode, side.result.stderr)
		}
		if string(side.result.stdout) == want.String() {
			t.Fatalf("%s context mutant survived", side.name)
		}
		t.Logf("%s killed by bytes: %s", side.name, firstDifference(string(side.result.stdout), want.String()))
	}
}
