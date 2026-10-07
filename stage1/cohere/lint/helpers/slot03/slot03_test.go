package slot03

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func command(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	log, err := os.CreateTemp(t.TempDir(), "command-")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	cmd.Stdout = log
	stderr, err := os.CreateTemp(t.TempDir(), "stderr-")
	if err != nil {
		t.Fatal(err)
	}
	defer stderr.Close()
	cmd.Stderr = stderr
	err = cmd.Run()
	data, readErr := os.ReadFile(log.Name())
	if readErr != nil {
		t.Fatal(readErr)
	}
	errorData, readErr := os.ReadFile(stderr.Name())
	if readErr != nil {
		t.Fatal(readErr)
	}
	if err != nil {
		t.Fatalf("%s %v: %v\n%s\n%s", name, args, err, data, errorData)
	}
	if len(errorData) != 0 {
		t.Fatalf("%s unexpectedly wrote stderr: %s", name, errorData)
	}
	return data
}
func oracle(t *testing.T) (string, []byte) {
	t.Helper()
	verifyCoverage(t)
	root, _ := filepath.Abs("../../../../../cohere")
	here, _ := filepath.Abs("testdata")
	virtual := filepath.Join(root, "adamic_slot03_oracle.go")
	raw, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(here, "oracle.go"), filepath.Join(root, "internal/lint/ecmascript/react/adamic_slot03.go"): filepath.Join(here, "react_export.go"), filepath.Join(root, "internal/lint/rules/tailwind/adamic_slot03.go"): filepath.Join(here, "tailwind_export.go")}})
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	overlay := filepath.Join(directory, "overlay.json")
	if err = os.WriteFile(overlay, raw, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "oracle")
	command(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	cases := filepath.Join(directory, "cases.json")
	want := command(t, "", binary, filepath.Join(here, "sources.jsonl.gz"), cases)
	return cases, want
}
func build(t *testing.T, entry string) string {
	t.Helper()
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "helper")
	if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	return binary
}

// Not parallel: million-line observations and native builds are reused serially to bound memory.
func TestSlot03HelpersMatchCohere(t *testing.T) {
	cases, want := oracle(t)
	entry, _ := filepath.Abs("main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	for _, got := range [][]byte{command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), command(t, "", build(t, entry), cases)} {
		if !bytes.Equal(got, want) {
			mismatch(t, got, want)
		}
	}
	t.Logf("%d helper output lines match Go, Node source and sanitized native", bytes.Count(want, []byte{'\n'}))
}
func mismatch(t *testing.T, got, want []byte) {
	t.Helper()
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			t.Fatalf("line %d: got %q Go %q", i+1, a[i], b[i])
		}
	}
	t.Fatalf("output sizes: got %d Go %d", len(got), len(want))
}

// Not parallel: the semantic mutants each compile and compare the same large observation log.
func TestSlot03HelperMutants(t *testing.T) {
	cases, want := oracle(t)
	for _, mutant := range []struct{ file, old, new, prefix string }{
		{"component_base_name.a", "name === 'Component' || name === 'PureComponent'", "name === 'Component'", ""},
		{"tailwind_space.a", " || codePoint === 11", "", ""},
		{"listener_kinds.a", "'VariableDeclaration'", "'StringLiteral'", ""},
		{"listener_kinds.a", "return ['JsxAttribute', 'CallExpression', 'VariableDeclaration'];", "return sharedKinds;", "const sharedKinds: string[] = ['JsxAttribute', 'CallExpression', 'VariableDeclaration'];\n"},
	} {
		name := mutant.file + "/value"
		if mutant.prefix != "" {
			name = mutant.file + "/shared-list"
		}
		t.Run(name, func(t *testing.T) {
			scratch := t.TempDir()
			for _, file := range []string{"component_base_name.a", "tailwind_space.a", "listener_kinds.a", "main.a"} {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				text := string(data)
				if file == mutant.file {
					if strings.Count(text, mutant.old) != 1 {
						t.Fatal("mutant anchor changed")
					}
					text = mutant.prefix + strings.Replace(text, mutant.old, mutant.new, 1)
				}
				if file == "main.a" {
					reader, _ := filepath.Abs("../options_json.ts")
					text = strings.ReplaceAll(text, "../options_json.ts", filepath.ToSlash(reader))
				}
				if err = os.WriteFile(filepath.Join(scratch, file), []byte(text), 0644); err != nil {
					t.Fatal(err)
				}
			}
			got := command(t, "", build(t, filepath.Join(scratch, "main.a")), cases)
			if bytes.Equal(got, want) {
				t.Fatal("compiling semantic mutant survived")
			}
			a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
			for i := 0; i < len(a) && i < len(b); i++ {
				if a[i] != b[i] {
					t.Logf("compiled semantic mutant caught at line %d: got %q Go %q", i+1, a[i], b[i])
					return
				}
			}
			t.Fatal("mutant must change a semantic output line")
		})
	}
}

func verifyCoverage(t *testing.T) {
	t.Helper()
	var ledger struct {
		Remaining []struct {
			Rule             string
			RemainingHelpers []string `json:"remaining_helpers"`
		}
	}
	data, err := os.ReadFile("../readiness.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &ledger); err != nil {
		t.Fatal(err)
	}
	expected := map[string]bool{}
	for _, rule := range ledger.Remaining {
		for _, helper := range rule.RemainingHelpers {
			for _, suffix := range []string{"/react.isComponentBaseName", "/tailwind.isSpace", "/tailwind.ListenerKinds"} {
				if strings.HasSuffix(helper, suffix) {
					expected[rule.Rule] = true
				}
			}
		}
	}
	var coverage struct {
		Rules        []string
		Sources      int
		CohereCommit string `json:"cohere_commit"`
	}
	data, err = os.ReadFile("testdata/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &coverage); err != nil {
		t.Fatal(err)
	}
	observed := map[string]bool{}
	f, err := os.Open("testdata/sources.jsonl.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zip, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer zip.Close()
	decoder := json.NewDecoder(zip)
	count := 0
	for {
		var row struct{ Rule string }
		err = decoder.Decode(&row)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		observed[row.Rule] = true
		count++
	}
	if err = coverageVerdict(expected, observed); err != nil {
		t.Fatal(err)
	}
	recorded := map[string]bool{}
	for _, rule := range coverage.Rules {
		recorded[rule] = true
	}
	if err = coverageVerdict(expected, recorded); err != nil {
		t.Fatal(err)
	}
	if count != coverage.Sources {
		t.Fatalf("captured inputs: %d, recorded %d", count, coverage.Sources)
	}
	commit := strings.TrimSpace(string(command(t, "../../../../../cohere", "git", "rev-parse", "HEAD")))
	if commit != coverage.CohereCommit {
		t.Fatalf("cohere pin changed: %s recorded %s", commit, coverage.CohereCommit)
	}
}
func coverageVerdict(expected, observed map[string]bool) error {
	missing, extra := []string{}, []string{}
	for rule := range expected {
		if !observed[rule] {
			missing = append(missing, rule)
		}
	}
	for rule := range observed {
		if !expected[rule] {
			extra = append(extra, rule)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	if len(missing) > 0 || len(extra) > 0 {
		return fmt.Errorf("consumer capture mismatch: missing %v extra %v", missing, extra)
	}
	return nil
}
func TestConsumerCoverageRejectsMutant(t *testing.T) {
	expected := map[string]bool{"react/no-direct-mutation-state": true, "better-tailwindcss/no-unknown-classes": true}
	observed := map[string]bool{"react/no-direct-mutation-state": true}
	if err := coverageVerdict(expected, observed); err == nil {
		t.Fatal("dropping the externally skipped consumer survived")
	} else {
		t.Logf("missing-consumer mutant caught: %v", err)
	}
}
