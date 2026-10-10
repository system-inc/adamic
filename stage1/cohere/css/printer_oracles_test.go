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
)

var printerMutants = []mutant{
	{"declarations lose their semicolons", "print.ts", "let ending = d.text(';');\n        if(n.truth('nodes')) ending", "let ending = d.text('');\n        if(n.truth('nodes')) ending"},
	{"rule bodies lose indentation", "print.ts", "d.indent(d.concat([d.hard(), this.sequence(path)]))", "d.concat([d.hard(), this.sequence(path)])"},
	{"groups ignore the remaining width", "print_doc.ts", "flat = !n.broken && this.fits(command(n.parts[0] ?? 0, c.indent, true), stack, width - column, suffix.length > 0, false);", "flat = !n.broken;"},
}

func printerAnswers(t *testing.T, cases, mode string) string {
	t.Helper()
	repo, _ := filepath.Abs(repository)
	side, _ := filepath.Abs("testdata/print_side_test.go")
	directory := t.TempDir()
	answers := filepath.Join(directory, "answers.txt")
	request, _ := json.Marshal(map[string]string{"Cases": cases, "Answers": answers, "Mode": mode})
	requestPath := filepath.Join(directory, "request.json")
	if err := os.WriteFile(requestPath, request, 0644); err != nil {
		t.Fatal(err)
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(repo, "cohere", "internal", "format", "css", "adamic_print_side_test.go"): side}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	cmd := bounded(t, "go", "test", "-timeout=0", "-v", "-count=1", "-overlay="+overlayPath, "-run=^TestAdamicPrinterCases$", "./internal/format/css")
	cmd.Dir = filepath.Join(repo, "cohere")
	cmd.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	output, err := childguard.CombinedOutput(cmd, childguard.Options{Stall: childStall})
	if err != nil {
		t.Fatalf("Go printer: %v\n%s", err, output)
	}
	t.Logf("%s", output)
	data, err := os.ReadFile(answers)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// Not parallel: native.Build writes the shared adamic/runtime cache.
func TestClosedPrinterRegexGap(t *testing.T) {
	path, _ := filepath.Abs("gaps/6_printer_regex_tree.ts")
	program := lowered(t, path)
	nativeRun, binary := natively(t, program)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"native", nativeRun}, {"Node", onNode(t, path)}, {"JavaScript backend", onJavaScriptBackend(t, program)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != "2\nOk\n" {
			t.Fatalf("%s: %d %q %s", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
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

func comparePrinterLibrary(t *testing.T, cases, expected, library, mode, variant string) {
	t.Helper()
	script, _ := filepath.Abs("testdata/print_library.mjs")
	output := execute(t, nil, "node", script, library, cases, mode, variant)
	if output.exitCode != 0 || len(output.stderr) != 0 {
		t.Fatalf("Prettier %s: %d %s", variant, output.exitCode, output.stderr)
	}
	data, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	inputs := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	goLines := strings.Split(strings.TrimSuffix(expected, "\n"), "\n")
	jsLines := strings.Split(strings.TrimSuffix(string(output.stdout), "\n"), "\n")
	if len(goLines) != len(jsLines) || len(goLines) != len(inputs)*2 {
		t.Fatal("original formatter returned a different number of cases")
	}
	data, err = os.ReadFile("testdata/printer_library_gaps.json")
	if err != nil {
		t.Fatal(err)
	}
	var recorded []printerGap
	if err := json.Unmarshal(data, &recorded); err != nil {
		t.Fatal(err)
	}
	agrees, refuses, gaps := 0, 0, 0
	classes := map[string]int{}
	for index, input := range inputs {
		if goLines[index*2] != jsLines[index*2] {
			t.Fatal("case markers differ")
		}
		goAnswer, jsAnswer := goLines[index*2+1], jsLines[index*2+1]
		if strings.HasPrefix(goAnswer, "error ") && strings.HasPrefix(jsAnswer, "error ") {
			refuses++
			continue
		}
		var goText, jsText string
		goOK := json.Unmarshal([]byte(goAnswer), &goText) == nil
		jsOK := json.Unmarshal([]byte(jsAnswer), &jsText) == nil
		if goOK && jsOK && goText == jsText {
			agrees++
			continue
		}
		found := false
		for _, gap := range recorded {
			if gap.Mode == mode && gap.Variant == variant && gap.Input == input && gap.Go == goAnswer && gap.Prettier == jsAnswer {
				found = true
				classes[gap.Class]++
				break
			}
		}
		if !found {
			t.Errorf("case %d %q: unrecorded Prettier %s difference: %s", index, input, variant, firstDifference(jsAnswer, goAnswer))
		} else {
			gaps++
		}
	}
	t.Logf("Prettier %s %s: %d byte-identical formats, %d shared refusals, %d exact recorded discrepancy occurrences: %v", variant, mode, agrees, refuses, gaps, classes)
}

// printerLeaks runs the same platform leak check against a prebuilt printer.
func printerLeaks(t *testing.T, binary string, arguments ...string) string {
	t.Helper()
	switch runtime.GOOS {
	case "darwin":
		report := execute(t, nil, "leaks", append([]string{"--atExit", "--", binary}, arguments...)...)
		if report.exitCode == 0 {
			return ""
		}
		return string(report.stdout)
	case "linux":
		report := execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, arguments...)
		if report.exitCode == 0 {
			return ""
		}
		return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)
	}
	t.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
}
