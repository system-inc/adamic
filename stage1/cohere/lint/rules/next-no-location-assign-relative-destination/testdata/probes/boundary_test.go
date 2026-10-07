package next

import (
	"github.com/system-inc/cohere/internal/lint/testing"
	"testing"
)

func TestAdamicFilenameBoundary(t *testing.T) {
	for _, subject := range []struct {
		name   string
		path   string
		source string
		next   bool
	}{
		{"document", "pages/_document.tsx", "import Head from 'next/head';\n", false},
		{"pages", "pages/home.tsx", "export const getStaticpaths = () => {};\n", true},
	} {
		selected := NoHeadImportInDocument
		if subject.next {
			selected = NoTypos
		}
		correct := rule_testing.Run(t, selected, subject.path, subject.source)
		renamed := rule_testing.Run(t, selected, "case-000.ts", subject.source)
		if len(correct.Diagnostics) != 1 || len(renamed.Diagnostics) != 0 {
			t.Fatalf("%s: correct=%d renamed=%d", subject.name, len(correct.Diagnostics), len(renamed.Diagnostics))
		}
		rule_testing.ExpectFindings(t, correct, correct.Diagnostics[0].Message.Id)
		rule_testing.ExpectClean(t, renamed)
		t.Logf("%s: original filename reports 1, shared harness filename reports 0", selected.Name)
	}
}
