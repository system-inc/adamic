package rule_testing

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/docsdata/capture"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// capturedRun is what a Run hands the docs capture, kept on its Result so the assertion that passes on
// it can record the case. Nil unless COHERE_DOCS_CAPTURE is set, which is every ordinary run.
//
// The website's examples are the cases these tests assert, rather than examples written for the site,
// so an example can only show what a passing test has shown (internal/docsdata).
type capturedRun struct {
	rule       string
	file       string
	otherFiles int
	options    any
}

// newCapturedRun starts a record for a Run, or returns nil when capture is off.
func newCapturedRun(subject rule.Rule, fileName string, otherFiles int, options any) *capturedRun {
	if capture.Directory() == "" {
		return nil
	}
	return &capturedRun{rule: subject.Name, file: fileName, otherFiles: otherFiles, options: options}
}

// RecordAssertedCase records a case a test has just asserted by its own comparison rather than through
// an Expect helper: a replayed upstream corpus that checks each finding's span, text and message, which
// ExpectFindings cannot express. Call it only after that comparison passed. Without it the docs capture
// never sees the case, and rules.json never learns a message id only such a corpus asserts.
func RecordAssertedCase(t *testing.T, result Result) {
	t.Helper()
	outcome := capture.OutcomeFindings
	if len(result.Diagnostics) == 0 {
		outcome = capture.OutcomeClean
	}
	recordCase(t, result, outcome, "")
}

// recordCase writes the case an assertion just passed on. Called at the end of each Expect, after every
// check, so only an asserted outcome is ever recorded.
func recordCase(t *testing.T, result Result, outcome string, fixedSource string) {
	t.Helper()
	if result.capture == nil || result.SourceFile == nil {
		return
	}
	record := capture.Record{
		Rule:        result.capture.rule,
		File:        result.capture.file,
		Source:      result.SourceFile.Text(),
		OtherFiles:  result.capture.otherFiles,
		Outcome:     outcome,
		FixedSource: fixedSource,
	}
	if result.capture.options != nil {
		encoded, err := json.Marshal(result.capture.options)
		if err != nil {
			// A case whose options cannot be shown cannot be an example either; the test itself passed.
			t.Logf("docs capture: skipping a case whose options do not encode: %v", err)
			return
		}
		record.Options = encoded
	}
	for _, diagnostic := range result.Diagnostics {
		sourceFile := result.SourceFile
		if diagnostic.SourceFile != nil {
			sourceFile = diagnostic.SourceFile
		}
		// Positioned as cohere prints a finding: one-based, the column a byte offset into the line.
		line, column := scanner.GetECMALineAndByteOffsetOfPosition(sourceFile, diagnostic.Range.Pos())
		endLine, endColumn := scanner.GetECMALineAndByteOffsetOfPosition(sourceFile, diagnostic.Range.End())
		record.Findings = append(record.Findings, capture.Finding{
			Line:        line + 1,
			Column:      column + 1,
			EndLine:     endLine + 1,
			EndColumn:   endColumn + 1,
			MessageId:   diagnostic.Message.Id,
			Message:     strings.ReplaceAll(diagnostic.Message.Description, "\n", " "),
			Fix:         len(diagnostic.Fixes) > 0,
			Suggestions: len(diagnostic.Suggestions),
		})
	}
	if err := capture.Write(record); err != nil {
		t.Fatalf("docs capture: %v", err)
	}
}
