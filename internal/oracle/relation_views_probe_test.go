package oracle

import (
	"path/filepath"
	"testing"
)

// A probe, never committed: each whole-object reader of a relation-viewed array's element, stopped or
// lowered, and when lowered what Node, release native and the JavaScript backend print.
func TestRelationViewsProbe(t *testing.T) {
	for _, name := range []string{"element_in", "element_has_own", "element_console_log", "element_string", "element_template", "element_structured_clone", "element_for_in", "element_values", "element_entries"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			absolute, err := filepath.Abs(filepath.Join(repository, relationViewsDirectory+"/"+name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, absolute)
			if err != nil {
				t.Logf("PROBE %s stops: %v", name, err)
				return
			}
			truth := onNode(t, absolute)
			native := releasedUncached(t, program)
			javascript := onJavaScriptBackend(t, program)
			t.Logf("PROBE %s lowers: node exit %d stdout %q; native exit %d stdout %q stderr %q; javascript exit %d stdout %q stderr %q", name, truth.exitCode, truth.stdout, native.exitCode, native.stdout, native.stderr, javascript.exitCode, javascript.stdout, javascript.stderr)
		})
	}
}
