package tsprinter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The mutation witnesses are generated families, fixed by their generator's
// seeds and loops. Repository files cannot change their membership. Reuse the
// independent Go oracle while avoiding parsing/formatting unrelated inputs.
func mutantOracle(t *testing.T, change mutation) (string, string) {
	t.Helper()
	directory := t.TempDir()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	gaps, err := filepath.Abs("testdata/notyet.json")
	if err != nil {
		t.Fatal(err)
	}
	request, err := json.Marshal(map[string]any{"Files": []string{}, "Directory": directory, "Gaps": gaps})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(directory+"/request.json", request, 0644); err != nil {
		t.Fatal(err)
	}
	families, ok := mutantFamilies[change.name]
	if !ok && change.name != "hashbang loses its refusal" {
		t.Fatalf("missing witness families for %q", change.name)
	}
	allowed := "map[string]bool{"
	for _, family := range families {
		allowed += fmt.Sprintf("%q:true,", family)
	}
	allowed += "}"
	replacements := map[string]string{}
	kind := "Expression"
	for _, name := range []string{"expressions", "statements"} {
		if name == "statements" && change.entry != "statementsMain.ts" {
			continue
		}
		side, err := os.ReadFile("testdata/" + name + "_side_test.go")
		if err != nil {
			t.Fatal(err)
		}
		anchor := "add := func(label, source string) { cases = append(cases, portExpressionCase{Label: label, Source: source}) }"
		if strings.Count(string(side), anchor) != 1 {
			t.Fatalf("%s oracle add hook moved", name)
		}
		filtered := strings.Replace(string(side), anchor, "allowed := "+allowed+"\nadd := func(label, source string) { if allowed[label] { cases = append(cases, portExpressionCase{Label: label, Source: source}) } }", 1)
		path := filepath.Join(directory, name+"_side_test.go")
		if err := os.WriteFile(path, []byte(filtered), 0644); err != nil {
			t.Fatal(err)
		}
		replacements[root+"/cohere/internal/format/javascript/adamic_"+name+"_test.go"] = path
	}
	if change.entry == "statementsMain.ts" {
		kind = "Statement"
	}
	overlay, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(directory+"/overlay.json", overlay, 0644); err != nil {
		t.Fatal(err)
	}
	command := bounded(t, "go", "test", "-count=1", "-timeout=0", "-overlay="+directory+"/overlay.json", "-run=^TestAdamic"+kind+"Corpus$", "./internal/format/javascript")
	command.Dir = root + "/cohere"
	command.Env = append(os.Environ(), "ADAMIC_TS_EXPRESSION_REQUEST="+directory+"/request.json", "ADAMIC_TS_STATEMENT_REQUEST="+directory+"/request.json")
	if output, err := combinedOutput(command); err != nil {
		t.Fatalf("Go mutation oracle: %v\n%s", err, output)
	}
	answers, err := os.ReadFile(directory + "/answers.txt")
	if err != nil {
		t.Fatal(err)
	}
	return directory + "/cases.txt", string(answers)
}
