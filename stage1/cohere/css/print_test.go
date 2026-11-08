package css

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/stage1/cohere/css/oracles"
)

var printerMutants = []mutant{
	{
		"declarations lose their semicolons",
		"print.ts",
		"let ending = d.text(';');\n        if(n.truth('nodes'))\n            ending",
		"let ending = d.text('');\n        if(n.truth('nodes'))\n            ending",
	},
	{
		"rule bodies lose indentation",
		"print.ts",
		"d.indent(d.concat([d.hard(), this.sequence(path)]))",
		"d.concat([d.hard(), this.sequence(path)])",
	},
	{
		"groups ignore the remaining width",
		"print_doc.ts",
		"flat =\n                        !n.broken &&\n                        this.fits(\n                            command(n.parts[0] ?? 0, c.indent, true),\n                            stack,\n                            width - column,\n                            suffix.length > 0,\n                            false,\n                        );",
		"flat = !n.broken;",
	},
}

func printerAnswers(test *testing.T, cases, mode string) string {
	test.Helper()
	repo, _ := filepath.Abs(repository)
	directory := test.TempDir()
	answers := filepath.Join(directory, "answers.txt")
	request, _ := json.Marshal(map[string]string{"Cases": cases, "Answers": answers, "Mode": mode})
	requestPath := filepath.Join(directory, "request.json")
	if writeError := os.WriteFile(requestPath, request, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	// The Go printer is css-printer-oracle-1, cohere's css package with testdata/print_side_test.go laid over it, built
	// ahead; the unit only runs it, in the package's directory (#5qykzj5).
	command := bounded(
		test,
		oracles.ComparisonOracle(test, true),
		"-test.timeout=0",
		"-test.v",
		"-test.count=1",
		"-test.run=^TestAdamicPrinterCases$",
	)
	command.Dir = filepath.Join(repo, "cohere", "internal", "format", "css")
	command.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	output, combinedOutputError := childguard.CombinedOutput(
		command,
		childguard.Options{Stall: oracles.ChildStallLimit},
	)
	if combinedOutputError != nil {
		test.Fatalf("Go printer: %v\n%s", combinedOutputError, output)
	}
	test.Logf("%s", output)
	data, readError := os.ReadFile(answers)
	if readError != nil {
		test.Fatal(readError)
	}
	return string(data)
}

// Not parallel: native.Build writes the shared adamic/runtime cache.
func TestClosedPrinterRegexGap(test *testing.T) {
	path, _ := filepath.Abs("gaps/6_printer_regex_tree.ts")
	program := lowered(test, path)
	binary := filepath.Join(programSanitized(test, "gaps/6_printer_regex_tree.ts"), "port")
	nativeRun := sanitizedRun(test, binary)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"native", nativeRun}, {"Node", onNode(test, path)}, {"JavaScript backend", onJavaScriptBackend(test, program)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != "2\nOk\n" {
			test.Fatalf("%s: %d %q %s", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
		}
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
}

type printerGap struct {
	Input    string `json:"input"`
	Go       string `json:"go"`
	Prettier string `json:"prettier"`
	Mode     string `json:"mode"`
	Variant  string `json:"variant"`
	Class    string `json:"class"`
}

func comparePrinterLibrary(test *testing.T, cases, expected, library, mode, variant string) {
	test.Helper()
	script, _ := filepath.Abs("testdata/print_library.mjs")
	output := execute(test, nil, "node", script, library, cases, mode, variant)
	if output.exitCode != 0 || len(output.stderr) != 0 {
		test.Fatalf("Prettier %s: %d %s", variant, output.exitCode, output.stderr)
	}
	data, readError := os.ReadFile(cases)
	if readError != nil {
		test.Fatal(readError)
	}
	inputs := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	goLines := strings.Split(strings.TrimSuffix(expected, "\n"), "\n")
	jsLines := strings.Split(strings.TrimSuffix(string(output.stdout), "\n"), "\n")
	if len(goLines) != len(jsLines) || len(goLines) != len(inputs)*2 {
		test.Fatal("original formatter returned a different number of cases")
	}
	data, readError = os.ReadFile("testdata/printer_library_gaps.json")
	if readError != nil {
		test.Fatal(readError)
	}
	var recorded []printerGap
	if decodeError := json.Unmarshal(data, &recorded); decodeError != nil {
		test.Fatal(decodeError)
	}
	agrees, refuses, gaps := 0, 0, 0
	classes := map[string]int{}
	for index, input := range inputs {
		if goLines[index*2] != jsLines[index*2] {
			test.Fatal("case markers differ")
		}
		goAnswer, jsAnswer := goLines[index*2+1], jsLines[index*2+1]
		if strings.HasPrefix(goAnswer, "error ") && strings.HasPrefix(jsAnswer, "error ") {
			refuses++
			continue
		}
		var goText, jsText string
		goOk := json.Unmarshal([]byte(goAnswer), &goText) == nil
		jsOk := json.Unmarshal([]byte(jsAnswer), &jsText) == nil
		if goOk && jsOk && goText == jsText {
			agrees++
			continue
		}
		found := false
		for _, gap := range recorded {
			if gap.Mode == mode &&
				gap.Variant == variant &&
				gap.Input == input &&
				gap.Go == goAnswer &&
				gap.Prettier == jsAnswer {
				found = true
				classes[gap.Class]++
				break
			}
		}
		if !found {
			test.Errorf(
				"case %d %q: unrecorded Prettier %s difference: %s",
				index,
				input,
				variant,
				firstDifference(jsAnswer, goAnswer),
			)
		} else {
			gaps++
		}
	}
	test.Logf(
		"Prettier %s %s: %d byte-identical formats, %d shared refusals, %d exact recorded discrepancy occurrences: %v",
		variant,
		mode,
		agrees,
		refuses,
		gaps,
		classes,
	)
}

// printerLeaks runs the same platform leak check against a prebuilt printer.
func printerLeaks(test *testing.T, binary string, arguments ...string) string {
	test.Helper()
	switch runtime.GOOS {
	case "darwin":
		report := execute(test, nil, "leaks", append([]string{"--atExit", "--", binary}, arguments...)...)
		if report.exitCode == 0 {
			return ""
		}
		return string(report.stdout)
	case "linux":
		report := execute(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, arguments...)
		if report.exitCode == 0 {
			return ""
		}
		return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)
	}
	test.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
}
