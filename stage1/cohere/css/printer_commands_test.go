package css

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Printer-owned runners retain the existing checks with context-bounded commands.
func cssPrinterExecute(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	command := cssCommand(t, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := childguard.Run(command, childguard.Options{Stall: childStall})
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatalf("running %s: %v", name, err)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
}

// cssPrinterOnNode runs source or emitted JavaScript through the oracle runner.
func cssPrinterOnNode(t *testing.T, path string, arguments ...string) run {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return cssPrinterExecute(t, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...)
}

// cssPrinterCompareLibrary retains every recorded-gap and exact-answer check.
func cssPrinterCompareLibrary(t *testing.T, cases, expected, library, mode, variant string) {
	t.Helper()
	script, _ := filepath.Abs("testdata/print_library.mjs")
	output := cssPrinterExecute(t, nil, "node", script, library, cases, mode, variant)
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

// cssPrinterLeaks retains the original platform leak check.
func cssPrinterLeaks(t *testing.T, binary string, arguments ...string) string {
	t.Helper()
	switch runtime.GOOS {
	case "darwin":
		report := cssPrinterExecute(t, nil, "leaks", append([]string{"--atExit", "--", binary}, arguments...)...)
		if report.exitCode == 0 {
			return ""
		}
		return string(report.stdout)
	case "linux":
		report := cssPrinterExecute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, arguments...)
		if report.exitCode == 0 {
			return ""
		}
		return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)
	}
	t.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
}

// Cancellation must close pipes inherited by grandchildren, not only the child.
// Not parallel: cancellation timing shares the machine's CPU budget.
func TestCSSPrinterCommandsCancelProcessGroup(t *testing.T) {
	command := cssCommandLimit(t, 2*time.Second, "node", "-e", `const {spawn}=require('node:child_process'); const child=spawn(process.execPath,['-e','setInterval(()=>{},1000)'],{stdio:'inherit'}); console.log(child.pid); setInterval(()=>{},1000);`)
	command.WaitDelay = time.Second
	var output bytes.Buffer
	command.Stdout = &output
	started := time.Now()
	err := command.Run()
	if err == nil {
		t.Fatal("command survived deadline")
	}
	if _, err := strconv.Atoi(strings.TrimSpace(output.String())); err != nil {
		t.Fatalf("grandchild did not start: %q", output.String())
	}
	if elapsed := time.Since(started); elapsed > 2750*time.Millisecond {
		t.Fatalf("grandchild kept inherited pipe open after cancellation: %s", elapsed)
	}
}

// The self child puts native.Build and any compilers it spawns in the same
// context-bounded process group; it materializes an input, not a timed test.
type cssPrinterNativeRequest struct {
	Source, Output string
	Sanitize       bool
}

func cssPrinterNativeBuild(t cssTesting, dir string, sanitize bool) error {
	request, err := json.Marshal(cssPrinterNativeRequest{filepath.Join(dir, "main.c"), filepath.Join(dir, "native"), sanitize})
	if err != nil {
		return err
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	command := cssCommand(t, executable)
	command.Env = append(os.Environ(), "ADAMIC_CSS_PRINTER_NATIVE_BUILD="+string(request))
	output, err := childguard.CombinedOutput(command, childguard.Options{Ceiling: 90 * time.Second})
	if err != nil {
		return fmt.Errorf("native product: %w\n%s", err, output)
	}
	return nil
}
func cssPrinterNativeBuildMode() (int, bool) {
	input := os.Getenv("ADAMIC_CSS_PRINTER_NATIVE_BUILD")
	if input == "" {
		return 0, false
	}
	var request cssPrinterNativeRequest
	if err := json.Unmarshal([]byte(input), &request); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1, true
	}
	source, err := os.ReadFile(request.Source)
	if err == nil {
		err = native.Build(string(source), request.Output, native.Options{Sanitize: request.Sanitize})
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1, true
	}
	return 0, true
}
