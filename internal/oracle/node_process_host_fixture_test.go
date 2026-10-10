package oracle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestNodeProcessHostNewLine(t *testing.T) {
	t.Parallel()
	path, binary, script := sanitized(t, "stage3/fixtures/host/23_newLine.a")
	truth := onNode(t, path)
	data, err := os.ReadFile(filepath.Join(repository, "stage3/fixtures/host/status.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		File string `json:"file"`
		Node struct {
			Stdout string `json:"stdout"`
			Stderr string `json:"stderr"`
			Exit   int    `json:"exit"`
		} `json:"node"`
	}
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, row := range rows {
		if row.File == "23_newLine.a" {
			found = true
			expected := run{stdout: []byte(row.Node.Stdout), stderr: []byte(row.Node.Stderr), exitCode: row.Node.Exit}
			if difference := disagreement(expected, truth); difference != "" {
				t.Fatal("recorded Node: " + difference)
			}
		}
	}
	if !found {
		t.Fatal("missing recorded fixture")
	}
	for _, got := range []run{execute(t, binary), onNode(t, script)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Fatal(difference)
		}
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	if strings.Count(code, "adamic_node_eol()") != 1 {
		t.Fatal("native EOL anchor changed")
	}
	mutant := filepath.Join(t.TempDir(), "newline-mutant")
	if err := native.Build("#include \"adamic.h\"\nstatic adamic_string mutant_eol = ADAMIC_STRING(\"\\r\\n\");\n"+strings.Replace(code, "adamic_node_eol()", "&mutant_eol", 1), mutant, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	bad := execute(t, mutant)
	if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		t.Fatal("native mutant must fail only Node stdout comparison")
	}
	source, err := os.ReadFile(script)
	if err != nil {
		t.Fatal(err)
	}
	anchor := `"\n"`
	if strings.Count(string(source), anchor) != 1 {
		t.Fatal("JavaScript EOL anchor changed")
	}
	jsMutant := filepath.Join(t.TempDir(), "newline-mutant.mjs")
	if err := os.WriteFile(jsMutant, []byte(strings.Replace(string(source), anchor, `"\r\n"`, 1)), 0600); err != nil {
		t.Fatal(err)
	}
	bad = onNode(t, jsMutant)
	if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		t.Fatal("JavaScript mutant must fail only Node stdout comparison")
	}
	t.Log("23_newLine unchanged source matches status.json on both backends; CRLF mutants run cleanly and fail only Node bytes")
}
