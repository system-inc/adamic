package guard

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/release/dispatch"
)

// The formatter's identity keys the format record, so it must move with everything that can change what
// the formatter prints and with nothing else (see internal/release/dispatch/formatter.go). The check below
// reads the real module's formatter sources, so it lives here with the other whole-source checks: in
// dispatch, every format edit reran dispatch's Swift and build tests with it (#nxgt2ca).

// TestFormatterInputsCoverWhatCanChangeTheOutput reads the real module's inputs and checks each class
// is there: the printers, the options they resolve, the typescript-go parser they parse with, and the
// pass loop, and that no lint rule is.
func TestFormatterInputsCoverWhatCanChangeTheOutput(t *testing.T) {
	t.Parallel()

	// This package sits two directories below the module root.
	module, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	inputs, err := dispatch.FormatterInputs(module, nil)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, input := range inputs {
		names[input.Name] = true
		if strings.Contains(input.Name, "/internal/lint/rules/") {
			t.Errorf("a lint rule is a formatter input, so a rule commit would throw the record away: %s", input.Name)
		}
		if len(input.Sum) != 64 {
			t.Errorf("%s has no content hash: %q", input.Name, input.Sum)
		}
	}
	for _, class := range []struct{ what, contains string }{
		{"the printers", "github.com/system-inc/cohere/internal/format/native/"},
		{"the options they resolve", "github.com/system-inc/cohere/internal/format/formatoptions/"},
		{"the parser they parse with", "github.com/microsoft/TypeScript/tsc/"},
		{"the pass loop", dispatch.FormatterPassLoop},
	} {
		present := false
		for name := range names {
			present = present || strings.Contains(name, class.contains)
		}
		if !present {
			t.Errorf("%s (%s) are not among the formatter's inputs", class.what, class.contains)
		}
	}
}
