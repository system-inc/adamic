package tsprinter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mutation struct{ name, file, from, to, entry string }

var mutations = []mutation{
	{"group ignores width", "doc.ts", "!document.broken &&\n                        this.fits(flat", "!document.broken ||\n                        this.fits(flat", "docMain.ts"},
	{"fill never packs a pair", "doc.ts", "separatorMode = fitsPair ? 2 : 1;", "separatorMode = 1;", "docMain.ts"},
	{"expression loses required parentheses", "expressions.ts", "return this.parenthesize(id, parent, role)", "return parent < -1 && this.parenthesize(id, parent, role)", "main.ts"},
}

func mutatedPort(t *testing.T, change mutation) string {
	t.Helper()
	directory := t.TempDir()
	parser, err := filepath.Abs("../../typescript/parser")
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob("*.ts")
	if err != nil {
		t.Fatal(err)
	}
	applied := false
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		source := string(data)
		if file == change.file {
			if strings.Count(source, change.from) != 1 {
				t.Fatalf("mutant %q must change one place", change.name)
			}
			source = strings.Replace(source, change.from, change.to, 1)
			applied = true
		}
		source = strings.ReplaceAll(source, "../../typescript/parser/", parser+"/")
		if err = os.WriteFile(filepath.Join(directory, file), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if !applied {
		t.Fatal("mutant not applied")
	}
	return filepath.Join(directory, change.entry)
}

func TestMutants(t *testing.T) {
	t.Parallel()
	documents, documentWant, _ := documentCorpus(t)
	expressions, expressionWant, _ := expressionCorpus(t)
	for _, change := range mutations {
		t.Run(change.name, func(t *testing.T) {
			t.Parallel()
			cases, want := documents, documentWant
			arguments := []string{cases}
			if change.entry == "main.ts" {
				cases, want = expressions, expressionWant
				arguments = []string{"--cases", cases, "80"}
			}
			path := mutatedPort(t, change)
			node := onNode(t, path, arguments...)
			program := lowered(t, path)
			native, _ := natively(t, program, arguments...)
			for _, side := range []struct {
				name   string
				result run
			}{{"Node", node}, {"native", native}} {
				if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
					t.Fatalf("%s mutant must finish normally: exit %d stderr %s", side.name, side.result.exitCode, side.result.stderr)
				}
				if string(side.result.stdout) == want {
					t.Fatalf("%s mutant escaped the byte comparison", side.name)
				}
				t.Logf("%s caught by successful-run output mismatch: %s", side.name, firstDifference(string(side.result.stdout), want))
			}
		})
	}
}
