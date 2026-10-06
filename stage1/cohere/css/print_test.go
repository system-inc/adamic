package css

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf16"
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
	cmd := bounded(t, "go", "test", "-v", "-count=1", "-overlay="+overlayPath, "-run=^TestAdamicPrinterCases$", "./internal/format/css")
	cmd.Dir = filepath.Join(repo, "cohere")
	cmd.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	output, err := cmd.CombinedOutput()
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
func TestCSSPrinterAgreesWithGo(t *testing.T) {
	cases, _ := askedCases(t)
	for _, mode := range []string{"default", "narrow"} {
		t.Run(mode, func(t *testing.T) {
			expected := printerAnswers(t, cases, mode)
			directory := portDirectory(t, nil)
			output := onNode(t, filepath.Join(directory, "print_main.ts"), cases, "output", "once", mode)
			if output.exitCode != 0 || len(output.stderr) != 0 {
				t.Fatalf("Node: %d %s", output.exitCode, output.stderr)
			}
			if difference := firstDifference(string(output.stdout), expected); difference != "" {
				t.Fatal(difference)
			}
			t.Logf("%d stylesheet formats and refusals agree byte for byte", strings.Count(expected, "\n")/2)
			for _, mutation := range printerMutants {
				t.Run("catches "+mutation.name, func(t *testing.T) {
					mutated := portDirectory(t, &mutation)
					output := onNode(t, filepath.Join(mutated, "print_main.ts"), cases, "output", "once", mode)
					if output.exitCode != 0 || len(output.stderr) != 0 {
						t.Fatalf("mutant must run: %d %s", output.exitCode, output.stderr)
					}
					difference := firstDifference(string(output.stdout), expected)
					if difference == "" {
						t.Fatal("printer mutant survived")
					}
					t.Logf("Node caught: %s", difference)
				})
			}
			if library := os.Getenv("ADAMIC_CSS_PRINTER_LIBRARY"); library != "" {
				comparePrinterLibrary(t, cases, expected, library, mode, "npm")
				repo, _ := filepath.Abs(repository)
				comparePrinterLibrary(t, cases, expected, filepath.Join(repo, "cohere", "internal", "format", "prettier", "bundles"), mode, "fork")
			} else {
				t.Skip("set ADAMIC_CSS_PRINTER_LIBRARY to scratch Prettier 3.9.6")
			}
		})
	}
}
func TestNativeCSSPrinterHasTheRecordedCompositionGap(t *testing.T) {
	for _, name := range []string{"gaps/6_printer_regex_tree.ts", "print_main.ts"} {
		t.Run(name, func(t *testing.T) {
			path, _ := filepath.Abs(name)
			if strings.HasPrefix(name, "gaps/") {
				output := onNode(t, path)
				if output.exitCode != 0 || len(output.stderr) != 0 || string(output.stdout) != "2\nOk\n" {
					t.Fatalf("Node proof: %d %q %s", output.exitCode, output.stdout, output.stderr)
				}
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), program)
			if err == nil {
				t.Fatal("gap closed: enable native, JS backend, sanitizers, leaks and throughput")
			}
			if !strings.Contains(err.Error(), "ir.RegExpNew is a node the cycle finder doesn't know") {
				t.Fatalf("new blocker: %v", err)
			}
			t.Logf("%v", err)
		})
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
	want := fmt.Sprintf("%d of %d stylesheets formatted, %d units\n", count*10, count*10, expectedUnits*10)
	runner, _ := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	main, _ := filepath.Abs("print_main.ts")
	for round := 0; round < 3; round++ {
		for _, side := range []struct {
			name string
			args []string
		}{
			{"Adamic source on Node", []string{"--disable-warning=ExperimentalWarning", runner, main, sharedPath, "count", "repeat"}},
			{"Prettier fork on Node", []string{script, fork, sharedPath, "default", "fork", "count"}},
		} {
			start := time.Now()
			output := execute(t, nil, "node", side.args...)
			elapsed := time.Since(start)
			if output.exitCode != 0 || len(output.stderr) != 0 || string(output.stdout) != want {
				t.Fatalf("%s: %d %q %s, want %q", side.name, output.exitCode, output.stdout, output.stderr, want)
			}
			t.Logf("%s round %d: %d stylesheets in %s, %.0f stylesheets/s; %s", side.name, round+1, count*10, elapsed, float64(count*10)/elapsed.Seconds(), output.stdout)
		}
	}
	directory := t.TempDir()
	request, _ := json.Marshal(map[string]string{"Cases": sharedPath})
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
	cmd := bounded(t, "go", "test", "-v", "-count=1", "-overlay="+overlayPath, "-run=^TestAdamicPrinterThroughput$", "./internal/format/css")
	cmd.Dir = filepath.Join(repo, "cohere")
	cmd.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Go throughput: %v %s", err, output)
	}
	if !strings.Contains(string(output), fmt.Sprintf("%d units", expectedUnits*10)) {
		t.Fatal("Go throughput checksum changed")
	}
	t.Logf("%s", output)
}

func TestCSSPrinterBoundaryProofs(t *testing.T) {
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
