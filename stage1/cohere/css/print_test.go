package css

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/native"
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

// Not parallel: run only on request, with the semantic corpus checked first.
func TestCSSPrinterThroughput(t *testing.T) {
	if os.Getenv("ADAMIC_CSS_PRINTER_BENCH") == "" {
		t.Skip("set ADAMIC_CSS_PRINTER_BENCH=1")
	}
	library := os.Getenv("ADAMIC_CSS_PRINTER_LIBRARY")
	if library == "" {
		t.Fatal("set ADAMIC_CSS_PRINTER_LIBRARY")
	}
	cases, _ := askedCases(t)
	expected := printerAnswers(t, cases, "default")
	script, _ := filepath.Abs("testdata/print_library.mjs")
	repo, _ := filepath.Abs(repository)
	fork := filepath.Join(repo, "cohere", "internal", "format", "prettier", "bundles")
	original := execute(t, nil, "node", script, fork, cases, "default", "fork")
	if original.exitCode != 0 || len(original.stderr) != 0 {
		t.Fatalf("fork: %d %s", original.exitCode, original.stderr)
	}
	goLines := strings.Split(strings.TrimSuffix(expected, "\n"), "\n")
	jsLines := strings.Split(strings.TrimSuffix(string(original.stdout), "\n"), "\n")
	data, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	inputs := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	var shared strings.Builder
	var expectedUnits, count int
	for i, input := range inputs {
		var a, b string
		if json.Unmarshal([]byte(goLines[i*2+1]), &a) == nil && json.Unmarshal([]byte(jsLines[i*2+1]), &b) == nil && a == b {
			shared.WriteString(input + "\n")
			count++
			expectedUnits += len(utf16.Encode([]rune(a)))
		}
	}
	sharedPath := filepath.Join(t.TempDir(), "shared.txt")
	if err := os.WriteFile(sharedPath, []byte(shared.String()), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("shared exact-output throughput corpus: %d successful inputs, %d UTF-16 output units per round", count, expectedUnits)
	repetitions, repetitionMode := 10, "repeat"
	if os.Getenv("ADAMIC_CSS_PRINTER_BENCH_ONCE") != "" {
		repetitions, repetitionMode = 1, "once"
	}
	want := fmt.Sprintf("%d of %d stylesheets formatted, %d units\n", count*repetitions, count*repetitions, expectedUnits*repetitions)
	runner, _ := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	main, _ := filepath.Abs("print_main.ts")
	program := lowered(t, main)
	binary := filepath.Join(t.TempDir(), "printer")
	if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	for round := 0; round < 3; round++ {
		for _, side := range []struct {
			name    string
			command string
			args    []string
		}{
			{"native", binary, []string{sharedPath, "count", repetitionMode}},
			{"Adamic source on Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, main, sharedPath, "count", repetitionMode}},
			{"Prettier fork on Node", "node", []string{script, fork, sharedPath, "default", "fork", "count", repetitionMode}},
			{"Prettier npm on Node", "node", []string{script, library, sharedPath, "default", "npm", "count", repetitionMode}},
		} {
			start := time.Now()
			output := execute(t, nil, side.command, side.args...)
			elapsed := time.Since(start)
			if output.exitCode != 0 || len(output.stderr) != 0 || string(output.stdout) != want {
				t.Fatalf("%s: %d %q %s, want %q", side.name, output.exitCode, output.stdout, output.stderr, want)
			}
			t.Logf("%s round %d: %d stylesheets in %s, %.0f stylesheets/s; %s", side.name, round+1, count*repetitions, elapsed, float64(count*repetitions)/elapsed.Seconds(), output.stdout)
		}
	}
	directory := t.TempDir()
	request, _ := json.Marshal(map[string]string{"Cases": sharedPath, "Repeat": repetitionMode})
	requestPath := filepath.Join(directory, "request.json")
	if err := os.WriteFile(requestPath, request, 0644); err != nil {
		t.Fatal(err)
	}
	side, _ := filepath.Abs("testdata/print_side_test.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(repo, "cohere", "internal", "format", "css", "adamic_print_side_test.go"): side}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	cmd := bounded(t, "go", "test", "-timeout=0", "-v", "-count=1", "-overlay="+overlayPath, "-run=^TestAdamicPrinterThroughput$", "./internal/format/css")
	cmd.Dir = filepath.Join(repo, "cohere")
	cmd.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	output, err := childguard.CombinedOutput(cmd, childguard.Options{Stall: childStall})
	if err != nil {
		t.Fatalf("Go throughput: %v %s", err, output)
	}
	if !strings.Contains(string(output), fmt.Sprintf("%d units", expectedUnits*repetitions)) {
		t.Fatal("Go throughput checksum changed")
	}
	t.Logf("%s", output)
}

func TestCSSPrinterBoundaryProofs(t *testing.T) {
	t.Parallel()
	cases := filepath.Join(t.TempDir(), "cases.txt")
	inputs := "\ufeffa{b:c}"
	encoded := ">C" + inputs + "\n>C// x\\ra{}\n>C\u00a0\n>C---\\na:     b\\n---\\na{}\n"
	if err := os.WriteFile(cases, []byte(encoded), 0644); err != nil {
		t.Fatal(err)
	}
	expected := printerAnswers(t, cases, "default")
	lines := strings.Split(strings.TrimSuffix(expected, "\n"), "\n")
	var answers strings.Builder
	for i := 1; i < len(lines); i += 2 {
		answers.WriteString(lines[i] + "\n")
	}
	path, _ := filepath.Abs("gaps/printer_boundaries.ts")
	result := onNode(t, path)
	if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != answers.String() {
		t.Fatalf("boundary proof: %d %q %s; Go %q", result.exitCode, result.stdout, result.stderr, answers.String())
	}
	t.Logf("Go and Node direct-printer proofs:\n%s", answers.String())
	library := os.Getenv("ADAMIC_CSS_PRINTER_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_CSS_PRINTER_LIBRARY")
	}
	script, _ := filepath.Abs("gaps/printer_library.mjs")
	result = execute(t, nil, "node", script, library)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("Prettier proof: %d %s", result.exitCode, result.stderr)
	}
	t.Logf("original Prettier proofs:\n%s", result.stdout)
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

// This corpus isolates boolean flags and namespace choices without the full printer corpus.
// Not parallel: native.Build writes the shared adamic/runtime cache.
func TestOptionalBooleanPrinterMatchesGo(t *testing.T) {
	cases := filepath.Join(t.TempDir(), "cases.txt")
	inputs := ">Ca{b:c}\n>Ca{b:c!important}\n>C*|a{b:c}\n>C|a{b:c}\n>Csvg|a{b:c}\n>S$x:1!default;a{b:$x!important}\n"
	if err := os.WriteFile(cases, []byte(inputs), 0644); err != nil {
		t.Fatal(err)
	}
	expected := printerAnswers(t, cases, "default")
	path, err := filepath.Abs("print_main.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, path)
	arguments := []string{cases, "output", "once", "default"}
	nativeRun, binary := natively(t, program, arguments...)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"Node", onNode(t, path, arguments...)}, {"native ASan/UBSan", nativeRun}, {"JavaScript backend", onJavaScriptBackend(t, program, arguments...)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != expected {
			t.Fatalf("%s: %d %s %s", side.name, side.result.exitCode, side.result.stderr, firstDifference(string(side.result.stdout), expected))
		}
	}
	if report := leaks(t, program, binary, arguments...); report != "" {
		t.Fatal(report)
	}
	t.Log("six boolean-flag and namespace cases agree with Go, Node and both backends; leak clean")
}
