package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Required fields retain presence obligations even when their values may be undefined.
func TestCheckedViewIntersectionRequiredUndefinedPresence(t *testing.T) {
	for _, sample := range []struct{ name, actual, field string }{
		{"finite-missing", "{count: 42}", "view.value.label"},
		{"finite-undefined", "{count: 42, label: undefined}", ""},
		{"recursive-missing", "{count: 42, label: undefined, next: {count: 8}}", "view.value.next.label"},
		{"recursive-undefined", "{count: 42, label: undefined, next: {count: 8, label: undefined}}", ""},
		{"bounded-missing", "{count: 42, label: undefined, items: [], next: {count: 8, items: []}}", "view.value.next.label"},
		{"bounded-undefined", "{count: 42, label: undefined, items: [], next: {count: 8, label: undefined, items: []}}", ""},
	} {
		t.Run(sample.name, func(t *testing.T) {
			next := ""
			if strings.HasPrefix(sample.name, "recursive") || strings.HasPrefix(sample.name, "bounded") {
				next = "; readonly next?: Link"
			}
			if strings.HasPrefix(sample.name, "bounded") {
				next += "; readonly items: readonly number[]"
			}
			source := "interface Named {readonly label: string | undefined}\ninterface Link {readonly count: number" + next + "}\ninterface Box {readonly value: object}\ninterface Target extends Box {readonly value: Named & Link}\nconst actual=" + sample.actual + ";const box:Box={value:actual};const view=box as Target;const root=view.value;console.log(`${root===root}`);\n"
			if next != "" {
				source = strings.Replace(source, "readonly count: number; readonly next?: Link", "readonly count: number; readonly label: string | undefined; readonly next?: Link", 1)
			}
			path := filepath.Join(t.TempDir(), sample.name+".a")
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if difference := disagreement(run{stdout: []byte("true\n")}, truth); difference != "" {
				t.Fatal("Node: " + difference)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if os.Getenv("ADAMIC_INTERSECTION_PRESENCE_MUTANT") != "" && sample.field != "" {
				changed := 0
				for i := range program.ViewContracts {
					for j := range program.ViewContracts[i].Fields {
						if program.ViewContracts[i].Fields[j].Name == "label" {
							program.ViewContracts[i].Fields[j].Optional = true
							changed++
						}
					}
				}
				if changed == 0 {
					t.Fatal("presence mutant found no required label")
				}
			}
			actual, binary := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if sample.field == "" {
					if difference := disagreement(truth, got); difference != "" {
						t.Errorf("%s; got %#v", difference, got)
					}
					continue
				}
				if got.exitCode != 70 || len(got.stdout) != 0 || !strings.Contains(string(got.stderr), "cast failed: field read failed: "+sample.field+" is not initialized;") || !strings.Contains(string(got.stderr), "found missing") {
					t.Errorf("required undefined field must retain presence: got %#v", got)
				}
			}
			if sample.field == "" {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}
