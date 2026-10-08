package printer

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestPrinterConstructorGap(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs("gaps/nestedConstructor.ts")
	result := onNode(t, path)
	if result.exitCode != 0 || string(result.stdout) != "ready\n" {
		t.Fatalf("Node: %+v", result)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "this escaping a constructor before every field is set") {
		t.Fatalf("gap changed; update GAPS.md and remove workaround: %v", err)
	}
	t.Log(err)
}

// This independent input file proves the recorded upstream acceptance difference.
func TestPrinterWhitespaceGap(t *testing.T) {
	directory := os.Getenv("ADAMIC_GRAPHQL_PRETTIER")
	if directory == "" {
		// census: required-input ADAMIC_GRAPHQL_PRETTIER (prettier 3.9.6 and graphql 17.0.2), provided by cloud/setup.sh --gate-inputs (#xq2ecw6).
		t.Skip("set ADAMIC_GRAPHQL_PRETTIER to an npm install of prettier@3.9.6 and graphql@17.0.2; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	proof, err := os.ReadFile("gaps/whitespace-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var texts []string
	if err = json.Unmarshal(proof, &texts); err != nil {
		t.Fatal(err)
	}
	var transport strings.Builder
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	for _, text := range texts {
		transport.WriteString(">" + escape.Replace(text) + "\n")
	}
	path := filepath.Join(t.TempDir(), "whitespace-cases.txt")
	if err = os.WriteFile(path, []byte(transport.String()), 0644); err != nil {
		t.Fatal(err)
	}
	port, _ := filepath.Abs("main.ts")
	cases, answers := printerCases(t, "defaults")
	source, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	input, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	inputs, outputs := strings.Split(string(source), "\n"), strings.Split(answers, "\n")
	byInput := map[string]string{}
	for index, text := range inputs {
		if text != "" {
			byInput[text] = outputs[index]
		}
	}
	var want strings.Builder
	for _, text := range strings.Split(string(input), "\n") {
		if text == "" {
			continue
		}
		answer, exists := byInput[text]
		if !exists || !strings.HasPrefix(answer, "error\tSyntax Error:") {
			t.Fatalf("Go whitespace refusal missing: %q", text)
		}
		want.WriteString(answer + "\n")
	}
	native := nativelyRun(t, lowered(t, port), "--cases", path)
	for name, result := range map[string]run{"native": native, "Node": onNode(t, port, "--cases", path)} {
		if result.exitCode != 0 || len(result.stderr) > 0 || string(result.stdout) != want.String() {
			t.Fatalf("%s: %+v", name, result)
		}
	}
	script, _ := filepath.Abs("testdata/prettier.mjs")
	embedded, _ := filepath.Abs(filepath.Join(repository, "cohere/internal/format/prettier/bundles"))
	for _, side := range []struct{ directory, engine string }{{directory, "npm"}, {embedded, "embedded"}} {
		result := execute(t, nil, "node", script, side.directory, path, "defaults", side.engine)
		if result.exitCode != 0 || len(result.stderr) > 0 || string(result.stdout) != strings.Repeat("ok\t\n", 5) {
			t.Fatalf("Prettier %s: %+v", side.engine, result)
		}
	}
	t.Log("five whitespace texts: native and Node equal Go refusals; npm and embedded Prettier return empty text")
}
