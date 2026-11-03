package oracle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/hostconfiguration"
	"github.com/system-inc/adamic/internal/native"
)

func TestNodeProcessHostNewLine(t *testing.T) {
	t.Parallel()
	nodeProcessHostNewLine(t, filepath.Join(repository, "stage3/fixtures/host/status.json"))
}

func nodeProcessHostNewLine(t *testing.T, statusPath string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/host/23_newLine.a"))
	if err != nil {
		t.Fatal(err)
	}
	var binary, script string
	data, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatal(err)
	}
	var rows []hostMeasurement
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	found := false
	var truth run
	var configuration *hostconfiguration.Configuration
	var rowNumber int
	for index, row := range rows {
		if row.File == "23_newLine.a" {
			found = true
			configuration, rowNumber = row.Configuration, index+1
			truth = run{stdout: []byte(row.Node.Stdout), stderr: []byte(row.Node.Stderr), exitCode: row.Node.Exit}
			t.Run("node", func(t *testing.T) {
				nodeBinary := hostconfiguration.Require(t, index+1, row.Configuration)
				actualNode := hostNodeWith(t, inputRun{}, nodeBinary, path)
				if difference := disagreement(truth, actualNode); difference != "" {
					t.Fatal("recorded Node: " + difference)
				}
			})
			_, binary, script = sanitized(t, "stage3/fixtures/host/23_newLine.a")
		}
	}
	if !found {
		t.Fatal("missing recorded fixture")
	}
	t.Run("native", func(t *testing.T) {
		if difference := disagreement(truth, execute(t, binary)); difference != "" {
			t.Fatal("native versus recorded Node: " + difference)
		}
	})
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
	t.Run("javascript", func(t *testing.T) {
		nodeBinary := hostconfiguration.Require(t, rowNumber, configuration)
		if difference := disagreement(truth, hostNodeWith(t, inputRun{}, nodeBinary, script)); difference != "" {
			t.Fatal("JavaScript versus recorded Node: " + difference)
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
		bad := hostNodeWith(t, inputRun{}, nodeBinary, jsMutant)
		if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
			t.Fatal("JavaScript mutant must fail only Node stdout comparison")
		}
	})

	t.Log("23_newLine native checks match recorded bytes; JavaScript and its CRLF mutant require the row's Node build")
}
