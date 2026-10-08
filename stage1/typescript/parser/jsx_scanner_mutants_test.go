package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: each compiled scanner mutant is checked after its positive control.
func TestJsxScannerMutants(t *testing.T) {
	t.Parallel()
	manifest, _ := jsxManifest(t)
	oracle := goOracle(t)
	want := execute(t, "", oracle, "--manifest", manifest, "--whole").output
	directory, _ := filepath.Abs(".")
	for _, got := range []execution{wholeNode(t, directory, manifest, false), execute(t, "", buildPort(t, directory, true), "--manifest", manifest, "--whole")} {
		if diff := difference(got.output, want); diff != "" {
			t.Fatal(diff)
		}
	}
	changes := []struct{ name, from, to string }{
		{"raw attribute value", "this.value = this.text.slice(start, this.pos);\n        if(this.pos === this.text.length)", "this.value = this.text.slice(start, this.pos) + '!';\n        if(this.pos === this.text.length)"},
		{"JSX name payload", "this.value += this.text.slice(start, this.pos);", "this.value += this.text.slice(start, this.pos) + '!';"},
	}
	for _, change := range changes {
		t.Run(change.name, func(t *testing.T) {
			// Copy the parser unchanged, then give it its own scanner module.
			mutant := copyPort(t, "", "", "")
			scannerDirectory := filepath.Join(mutant, "scanner")
			if err := os.Mkdir(scannerDirectory, 0755); err != nil {
				t.Fatal(err)
			}
			originalScanner, _ := filepath.Abs("../scanner/scanner.ts")
			for _, file := range []string{"scanner.ts", "characters.ts", "tokens.ts"} {
				data, err := os.ReadFile(filepath.Join("../scanner", file))
				if err != nil {
					t.Fatal(err)
				}
				source := string(data)
				if file == "scanner.ts" {
					if strings.Count(source, change.from) != 1 {
						t.Fatal("scanner mutant must change exactly one site")
					}
					source = strings.Replace(source, change.from, change.to, 1)
				}
				if err := os.WriteFile(filepath.Join(scannerDirectory, file), []byte(source), 0644); err != nil {
					t.Fatal(err)
				}
			}
			for _, file := range portFiles {
				path := filepath.Join(mutant, file)
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				source := strings.ReplaceAll(string(data), originalScanner, filepath.Join(scannerDirectory, "scanner.ts"))
				if err := os.WriteFile(path, []byte(source), 0644); err != nil {
					t.Fatal(err)
				}
			}
			node := wholeNode(t, mutant, manifest, false)
			native := execute(t, "", buildPort(t, mutant, true), "--manifest", manifest, "--whole")
			for _, side := range []struct {
				name string
				got  execution
			}{{"Node", node}, {"native", native}} {
				diff := difference(side.got.output, want)
				if diff == "" {
					t.Fatalf("%s scanner mutant survived", side.name)
				}
				t.Logf("%s compiled mutant caught: %s", side.name, strings.Split(diff, "\n")[0])
			}
		})
	}
}
