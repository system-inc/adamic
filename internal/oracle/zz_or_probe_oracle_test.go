package oracle

// Throwaway probe for #xqq507s, never committed: each program in $OR_PROBES lowered as main lowers
// it, and, when it lowers, run on Node, the JavaScript backend, and natively under ASan, then leak
// checked (LeakSanitizer on Linux).

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/lower"
)

func TestZzOrProbeRuns(t *testing.T) {
	directory := os.Getenv("OR_PROBES")
	if directory == "" {
		t.Skip("OR_PROBES unset")
	}
	paths, _ := filepath.Glob(filepath.Join(directory, "*"))
	sort.Strings(paths)
	for _, path := range paths {
		if ext := filepath.Ext(path); ext != ".a" && ext != ".ts" {
			continue
		}
		t.Run(filepath.Base(path), func(t *testing.T) {
			out := &strings.Builder{}
			defer func() { fmt.Print(out.String()) }()
			fmt.Fprintf(out, "\n######## %s\n", filepath.Base(path))
			show := func(label string, result run) {
				fmt.Fprintf(out, "%s: exit=%d stdout=%q stderr=%q\n", label, result.exitCode, result.stdout, truncate(string(result.stderr)))
			}
			show("node", executeWith(t, jsonTypesEnvironment(t, path), "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), path))
			program, err := lowered(t, path)
			var refused *lower.Refused
			if errors.As(err, &refused) {
				fmt.Fprintf(out, "REFUSED: %v\n", refused)
				return
			}
			if err != nil {
				fmt.Fprintf(out, "ERROR (%T): %v\n", err, err)
				return
			}
			fmt.Fprintf(out, "LOWERED: graphTypes=%d\n", len(program.GraphTypes))
			script := filepath.Join(t.TempDir(), "program.mjs")
			if err := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0o644); err != nil {
				t.Fatal(err)
			}
			show("javascript", executeWith(t, jsonTypesEnvironment(t, script), "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), script))
			native, binary := nativelyUncached(t, program)
			show("native-asan", native)
			report := leaksUncached(t, program, binary)
			if report == "" {
				fmt.Fprintf(out, "leaks: none\n")
			} else {
				fmt.Fprintf(out, "leaks: %s\n", truncate(report))
			}
		})
	}
}

func truncate(text string) string {
	if len(text) > 3000 {
		return text[:3000] + "...[truncated]"
	}
	return text
}
