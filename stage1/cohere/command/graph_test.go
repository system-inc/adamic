package command

import (
	"bytes"
	"encoding/json"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func graphTests(t *testing.T, source, binary string) run {
	t.Helper()
	root := absolute(t, filepath.Join(repository, "cohere"))
	original := filepath.Join(root, "command/cohere/main.go")
	content, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	for _, name := range []string{"narrowToClosure", "rescopeAfterRebuild"} {
		from := "func " + name + "("
		if strings.Count(text, from) != 1 {
			t.Fatal("unknown graph boundary", name)
		}
		text = strings.Replace(text, from, "func adamicGo"+strings.ToUpper(name[:1])+name[1:]+"(", 1)
	}
	// Capture the actual Go format/write in a per-call builder. Returning the captured note
	// avoids reconstructing expected wording and avoids a global stderr race in parallel tests.
	start := strings.Index(text, "func adamicGoNarrowToClosure(")
	end := strings.Index(text[start:], "\nfunc reportForeignCheckers(") + start
	if start < 0 || end < start {
		t.Fatal("unknown narrow function extent")
	}
	body := text[start:end]
	for _, change := range []struct{ from, to string }{
		{") (formatScope, []*ast.SourceFile) {", ") (formatScope, []*ast.SourceFile, string) {\n var note strings.Builder"},
		{"accountOutput(os.Stderr)", "&note"},
		{"return scope, projectFiles", "return scope, projectFiles, note.String()"},
		{"return formatScope{Everything: true}, projectFiles", "return formatScope{Everything: true}, projectFiles, note.String()"},
		{"return scope, closure", "return scope, closure, note.String()"},
	} {
		if strings.Count(body, change.from) != 1 {
			t.Fatal("unknown note boundary", change.from)
		}
		body = strings.Replace(body, change.from, change.to, 1)
	}
	text = text[:start] + body + text[end:]
	replacement := filepath.Join(t.TempDir(), "main.go")
	write(t, replacement, text, 0644)
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	raw, err := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement, filepath.Join(root, "command/cohere/adamic_graph_test.go"): absolute(t, "testdata/graph_boundaries.go.txt")}})
	if err != nil {
		t.Fatal(err)
	}
	write(t, overlay, string(raw), 0644)
	command := bounded(t, "go", "test", "-count=1", "-v", "-overlay="+overlay, "-run=^(TestNarrowToClosure.*|TestAdamicGraphCyclesAndRebuild|TestAdamicGraphConstructionGap|TestAdamicExactClosureLimit|TestAdamicRunProjectBoundary)$", "./command/cohere")
	command.Dir = root
	command.Env = append(os.Environ(), "ADAMIC_NATIVE_PROBE="+binary, "ADAMIC_NODE_PROBE="+wrapper(t, source, false), "ADAMIC_BACKEND_PROBE="+wrapper(t, source, true))
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	_ = command.Run()
	if command.ProcessState == nil {
		t.Fatal("graph test process did not start")
	}
	return run{stdout.Bytes(), stderr.Bytes(), command.ProcessState.ExitCode()}
}
func TestOriginalGraphAndProcessBoundaries(t *testing.T) {
	source := absolute(t, "graph_probe.ts")
	program := lowered(t, source)
	binary := filepath.Join(t.TempDir(), "probe")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := graphTests(t, source, binary)
	if result.exitCode != 0 {
		t.Fatalf("Go graph and process tests: exit %d\n%s\n%s", result.exitCode, result.stdout, result.stderr)
	}
	t.Logf("%s", result.stdout)
}
