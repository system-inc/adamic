package probes

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

type result struct {
	output []byte
	stderr string
	err    error
}

func execute(t *testing.T, cwd string, argv ...string) result {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = cwd
	output, err := os.CreateTemp(t.TempDir(), "output-")
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stdout = output
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if err = output.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return result{data, stderr.String(), runErr}
}
func clean(t *testing.T, r result) []byte {
	t.Helper()
	if r.err != nil || r.stderr != "" {
		t.Fatalf("%v\n%s", r.err, r.stderr)
	}
	return r.output
}
func build(t *testing.T, entry string) (string, string) {
	t.Helper()
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "probe")
	if err = native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	js := filepath.Join(dir, "probe.mjs")
	if err = os.WriteFile(js, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	return binary, js
}
func goOracle(t *testing.T, repo string) string {
	t.Helper()
	root := filepath.Join(repo, "cohere")
	side, _ := filepath.Abs("oracle.go.txt")
	virtual := filepath.Join(root, "adamic_wave104_gap_probe.go")
	encoded, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
	dir := t.TempDir()
	overlay := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(overlay, encoded, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(dir, "oracle")
	clean(t, execute(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual))
	return binary
}
func TestSelectedRuleParserSubstrate(t *testing.T) {
	repo, _ := filepath.Abs("../../../../..")
	entry, _ := filepath.Abs("parser_gap.a")
	runner := filepath.Join(repo, "oracle/node.mjs")
	binary, js := build(t, entry)
	oracle := goOracle(t, repo)
	data, err := os.ReadFile("witnesses.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]string
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	commands := [][]string{
		{"node", "--disable-warning=ExperimentalWarning", runner, entry},
		{"node", "--disable-warning=ExperimentalWarning", runner, js},
		{binary},
	}
	names := []string{"Node source", "emitted JavaScript", "ASan UBSan native"}
	for i, row := range rows {
		encoded, _ := json.Marshal([]any{row})
		corpus := filepath.Join(t.TempDir(), "row.json")
		if err = os.WriteFile(corpus, encoded, 0644); err != nil {
			t.Fatal(err)
		}
		want := clean(t, execute(t, "", oracle, corpus))
		facts := clean(t, execute(t, "", oracle, corpus, "--facts"))
		var fact struct {
			ParseDiagnostics int
			FindingIDs       []string
		}
		if err = json.Unmarshal(facts, &fact); err != nil {
			t.Fatal(err)
		}
		if fact.ParseDiagnostics != 0 {
			t.Fatal("gap witness is not valid Go TypeScript")
		}
		if i != 4 && len(fact.FindingIDs) != 1 {
			t.Fatal("Go rule must fire on the witness or its plain-source control")
		}
		t.Logf("row %d Go facts: %s", i, bytes.TrimSpace(facts))
		for j, argv := range commands {
			got := execute(t, "", append(append([]string{}, argv...), corpus)...)
			if i%2 == 0 {
				if !bytes.Equal(clean(t, got), want) {
					t.Fatalf("%s control %d differs: port %s Go %s", names[j], i, got.output, want)
				}
				t.Logf("%s control %d agrees: %s", names[j], i, bytes.TrimSpace(want))
			} else {
				if i == 3 {
					clean(t, got)
					if !bytes.Equal(got.output, []byte("case 0\nstrings 0 jsx 0 comments 1\n")) {
						t.Fatalf("%s changed the known JSX misparse: %q", names[j], got.output)
					}
				} else {
					exit, ok := got.err.(*exec.ExitError)
					if !ok || exit.ExitCode() != 70 || !strings.Contains(got.stderr, "parser slice expected GreaterThanToken, got Identifier") {
						t.Fatalf("%s failed differently from the measured parser refusal: %v %s", names[j], got.err, got.stderr)
					}
				}
				if bytes.Equal(got.output, want) {
					t.Fatalf("row %d %s no longer reproduces the JSX gap", i, names[j])
				}
				t.Logf("%s gap %d: exit=%v stderr=%q port=%q Go=%q", names[j], i, got.err, got.stderr, got.output, want)
			}
		}
	}
	// Instrumentation mutant: it must compile and finish all three plain-source
	// controls, then fail only the comparison with the independent Go parser.
	// It is not a mutant of any selected rule and is not credited as one.
	source, err := os.ReadFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	old := "strings++;"
	if strings.Count(string(source), old) != 1 {
		t.Fatal("mutant anchor drift")
	}
	dir := t.TempDir()
	mutated := strings.Replace(string(source), old, "strings += 2;", 1)
	mutated = strings.ReplaceAll(mutated, "'../../../../typescript/", "'"+filepath.ToSlash(filepath.Join(repo, "stage1/typescript"))+"/")
	mutated = strings.ReplaceAll(mutated, "'../../helpers/", "'"+filepath.ToSlash(filepath.Join(repo, "stage1/cohere/lint/helpers"))+"/")
	mutant := filepath.Join(dir, "probe.a")
	if err = os.WriteFile(mutant, []byte(mutated), 0644); err != nil {
		t.Fatal(err)
	}
	mutatedBinary, mutatedJS := build(t, mutant)
	controls, _ := json.Marshal([]any{rows[0], rows[2], rows[4]})
	corpus := filepath.Join(t.TempDir(), "controls.json")
	if err = os.WriteFile(corpus, controls, 0644); err != nil {
		t.Fatal(err)
	}
	want := clean(t, execute(t, "", oracle, corpus))
	for i, argv := range [][]string{{"node", "--disable-warning=ExperimentalWarning", runner, mutant}, {"node", "--disable-warning=ExperimentalWarning", runner, mutatedJS}, {mutatedBinary}} {
		got := clean(t, execute(t, "", append(argv, corpus)...))
		if bytes.Equal(got, want) {
			t.Fatalf("%s compiling count mutant survived", names[i])
		}
		t.Logf("%s compiling instrumentation mutant caught by Go comparison: port=%q Go=%q", names[i], got, want)
	}
}
