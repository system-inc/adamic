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

func TestNodeProcessHostNewLine(test *testing.T) {
	test.Parallel()
	nodeProcessHostNewLine(test, filepath.Join(repository, "stage3/fixtures/host/status.json"))
}

func nodeProcessHostNewLine(test *testing.T, statusPath string) {
	test.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/host/23_newLine.a"))
	if err != nil {
		test.Fatal(err)
	}
	var binary, script string
	data, err := os.ReadFile(statusPath)
	if err != nil {
		test.Fatal(err)
	}
	var rows []hostMeasurement
	if err := json.Unmarshal(data, &rows); err != nil {
		test.Fatal(err)
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
			test.Run("node", func(test *testing.T) {
				if why := hostconfiguration.Skip(index+1, row.Configuration); why != "" {
					test.Skipf(
						"node process host row %d: the node step needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR",
						index+1,
					)
				}
				nodeBinary := hostconfiguration.Require(test, index+1, row.Configuration)
				actualNode := hostNodeWith(test, inputRun{}, nodeBinary, path)
				if difference := disagreement(truth, actualNode); difference != "" {
					test.Fatal("recorded Node: " + difference)
				}
			})
			_, binary, script = sanitized(test, "stage3/fixtures/host/23_newLine.a")
		}
	}
	if !found {
		test.Fatal("missing recorded fixture")
	}
	test.Run("native", func(test *testing.T) {
		if difference := disagreement(truth, execute(test, binary)); difference != "" {
			test.Fatal("native versus recorded Node: " + difference)
		}
	})
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	code := native.C(program)
	if strings.Count(code, "adamic_node_eol()") != 1 {
		test.Fatal("native EOL anchor changed")
	}
	mutant := filepath.Join(test.TempDir(), "newline-mutant")
	if err := native.Build(
		"#include \"adamic.h\"\nstatic adamic_string mutant_eol = ADAMIC_STRING(\"\\r\\n\");\n"+
			strings.Replace(code, "adamic_node_eol()", "&mutant_eol", 1),
		mutant,
		native.Options{Sanitize: true},
	); err != nil {
		test.Fatal(err)
	}
	bad := execute(test, mutant)
	if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		test.Fatal("native mutant must fail only Node stdout comparison")
	}
	test.Run("javascript", func(test *testing.T) {
		if why := hostconfiguration.Skip(rowNumber, configuration); why != "" {
			test.Skipf(
				"node process host row %d: the javascript step needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR",
				rowNumber,
			)
		}
		nodeBinary := hostconfiguration.Require(test, rowNumber, configuration)
		if difference := disagreement(truth, hostNodeWith(test, inputRun{}, nodeBinary, script)); difference != "" {
			test.Fatal("JavaScript versus recorded Node: " + difference)
		}
		source, err := os.ReadFile(script)
		if err != nil {
			test.Fatal(err)
		}
		anchor := `"\n"`
		if strings.Count(string(source), anchor) != 1 {
			test.Fatal("JavaScript EOL anchor changed")
		}
		jsMutant := filepath.Join(test.TempDir(), "newline-mutant.mjs")
		if err := os.WriteFile(
			jsMutant,
			[]byte(strings.Replace(string(source), anchor, `"\r\n"`, 1)),
			0600,
		); err != nil {
			test.Fatal(err)
		}
		bad := hostNodeWith(test, inputRun{}, nodeBinary, jsMutant)
		if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
			test.Fatal("JavaScript mutant must fail only Node stdout comparison")
		}
	})

	test.Log(
		"23_newLine native checks match recorded bytes; JavaScript and its CRLF mutant require the row's Node build",
	)
}
