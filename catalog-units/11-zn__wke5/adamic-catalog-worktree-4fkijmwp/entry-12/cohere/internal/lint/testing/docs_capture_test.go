package rule_testing

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/docsdata/capture"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// captureProbe reports every debugger statement and proposes deleting it, a rule written to exercise
// the docs capture rather than to ship.
var captureProbe = rule.Rule{
	Name: "probe-docs-capture",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindDebuggerStatement: func(node *ast.Node) {
				ctx.ReportNodeWithFixes(node, rule.Message{Id: "debugger", Description: "A debugger statement."},
					rule.Fix{Range: rule.TokenRange(ctx.SourceFile, node)})
			},
		}
	},
}

// captureHelperCase names the case TestDocsCaptureHelper runs in a subprocess; unset, it runs nothing.
const captureHelperCase = "COHERE_DOCS_CAPTURE_HELPER_CASE"

const captureProbeSource = "let a = 1;\n  debugger;\n"

// TestDocsCaptureHelper is the body of each case TestDocsCaptureRecordsOnlyAssertedCases runs. It runs in
// a subprocess because capture.Directory reads the environment once per process, so only a process
// started with COHERE_DOCS_CAPTURE set can show what the hook does with it.
func TestDocsCaptureHelper(t *testing.T) {
	t.Parallel()
	switch os.Getenv(captureHelperCase) {
	case "":
		t.Skip("run by TestDocsCaptureRecordsOnlyAssertedCases")
	case "Findings":
		ExpectFindings(t, Run(t, captureProbe, "subject.ts", captureProbeSource), "debugger")
	case "Clean":
		ExpectClean(t, Run(t, captureProbe, "subject.ts", "let a = 1;\n"))
	case "FixedMismatch":
		ExpectFixedSource(t, Run(t, captureProbe, "subject.ts", captureProbeSource), "not what the fix writes")
	}
}

// TestDocsCaptureRecordsOnlyAssertedCases: a passing assertion writes one record of what it asserted,
// a failing one writes nothing, and with capture off no file is created.
func TestDocsCaptureRecordsOnlyAssertedCases(t *testing.T) {
	t.Parallel()
	run := func(t *testing.T, helperCase string, captureOn bool) ([]capture.Record, string, error) {
		t.Helper()
		directory := t.TempDir()
		command := exec.Command(os.Args[0], "-test.run=^TestDocsCaptureHelper$", "-test.count=1")
		command.Env = append(os.Environ(), captureHelperCase+"="+helperCase)
		if captureOn {
			command.Env = append(command.Env, capture.Variable+"="+directory)
		} else {
			command.Env = append(command.Env, capture.Variable+"=")
		}
		output, err := command.CombinedOutput()
		paths, globError := filepath.Glob(filepath.Join(directory, "*"))
		if globError != nil {
			t.Fatal(globError)
		}
		var records []capture.Record
		for _, path := range paths {
			file, openError := os.Open(path)
			if openError != nil {
				t.Fatal(openError)
			}
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				var record capture.Record
				if decodeError := json.Unmarshal(scanner.Bytes(), &record); decodeError != nil {
					t.Fatalf("%s: %v", path, decodeError)
				}
				records = append(records, record)
			}
			file.Close()
		}
		if !captureOn && len(paths) > 0 {
			t.Errorf("capture off created %v", paths)
		}
		return records, string(output), err
	}

	t.Run("a passing ExpectFindings writes one record of the finding", func(t *testing.T) {
		t.Parallel()
		records, output, err := run(t, "Findings", true)
		if err != nil {
			t.Fatalf("the helper failed: %v\n%s", err, output)
		}
		if len(records) != 1 {
			t.Fatalf("%d records, want 1: %+v", len(records), records)
		}
		record := records[0]
		if record.Rule != captureProbe.Name || record.File != "subject.ts" || record.Source != captureProbeSource ||
			record.Outcome != capture.OutcomeFindings {
			t.Errorf("record %+v", record)
		}
		// The statement opens line 2 at its third byte, and its range ends after the semicolon, so the end
		// column is the twelfth: one past the last byte, as a range end is.
		want := capture.Finding{Line: 2, Column: 3, EndLine: 2, EndColumn: 12, MessageId: "debugger", Message: "A debugger statement.", Fix: true}
		if len(record.Findings) != 1 || record.Findings[0] != want {
			t.Errorf("findings %+v, want [%+v]", record.Findings, want)
		}
	})

	t.Run("a passing ExpectClean writes a Clean record", func(t *testing.T) {
		t.Parallel()
		records, output, err := run(t, "Clean", true)
		if err != nil {
			t.Fatalf("the helper failed: %v\n%s", err, output)
		}
		if len(records) != 1 || records[0].Outcome != capture.OutcomeClean || len(records[0].Findings) != 0 {
			t.Errorf("records %+v", records)
		}
	})

	t.Run("a failing ExpectFixedSource writes nothing", func(t *testing.T) {
		t.Parallel()
		records, output, err := run(t, "FixedMismatch", true)
		if err == nil {
			t.Fatalf("the helper passed, so the mismatch was never asserted:\n%s", output)
		}
		if len(records) != 0 {
			t.Errorf("a failed assertion recorded %+v", records)
		}
	})

	t.Run("with capture off nothing is written", func(t *testing.T) {
		t.Parallel()
		records, output, err := run(t, "Findings", false)
		if err != nil {
			t.Fatalf("the helper failed: %v\n%s", err, output)
		}
		if len(records) != 0 {
			t.Errorf("capture off recorded %+v", records)
		}
	})
}
