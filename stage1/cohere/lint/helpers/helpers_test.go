package helpers

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func run(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	f, err := os.CreateTemp(t.TempDir(), "output-")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cmd.Stdout = f
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = childguard.Run(cmd, childguard.Options{}); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	if stderr.Len() != 0 {
		t.Fatalf("%s stderr: %s", name, &stderr)
	}
	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func oracle(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs("../../../../cohere")
	side, _ := filepath.Abs("testdata/oracle.go")
	catalog, _ := filepath.Abs("testdata/catalog.go")
	descriptors, _ := filepath.Abs("testdata/descriptors.go")
	virtual := filepath.Join(root, "adamic_helper_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side, filepath.Join(root, "policy/adamic_helpers.go"): catalog, filepath.Join(root, "adamic_helper_descriptors.go"): descriptors}})
	path := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "go-oracle")
	run(t, root, "go", "build", "-overlay="+path, "-o", binary, virtual, filepath.Join(root, "adamic_helper_descriptors.go"))
	return binary
}
func build(t *testing.T, directory string) string {
	t.Helper()
	program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "helpers")
	if err := native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	return binary
}
// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestHelpersMatchCohere(t *testing.T) {
	goOracle := oracle(t)
	catalog := run(t, "", goOracle, "catalog")
	descriptor := run(t, "", goOracle, "descriptors")
	pinned, err := os.ReadFile("testdata/descriptors.json")
	if err != nil || !bytes.Equal(descriptor, pinned) {
		t.Fatal("regenerate target descriptors against Go cohere")
	}
	current, err := os.ReadFile("testdata/catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(catalog, current) {
		t.Fatal("regenerate catalog against pinned Go cohere")
	}
	cases := fixture(t)
	catalogPath, _ := filepath.Abs("testdata/catalog.json")
	want := run(t, "", goOracle, cases)
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.ts")
	node := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases, catalogPath)
	compare(t, node, want)
	compare(t, run(t, "", build(t, "."), cases, catalogPath), want)
	t.Logf("%d oracle output lines matched Go, Node and sanitized native", len(strings.Split(strings.TrimSpace(string(want)), "\n")))
}
func compare(t *testing.T, got, want []byte) {
	t.Helper()
	if bytes.Equal(got, want) {
		return
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			t.Fatalf("output line %d: got %q, Go %q", i+1, a[i], b[i])
		}
	}
	t.Fatalf("output size: got %d Go %d", len(got), len(want))
}
// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestHelperMutants(t *testing.T) {
	cases := fixture(t)
	catalog, _ := filepath.Abs("testdata/catalog.json")
	want := run(t, "", oracle(t), cases)
	for _, m := range []struct{ file, old, new string }{{"options_json.ts", "if(char.charCodeAt(0) < 32)", "if(false)"}, {"option_schema.ts", "matched !== 1", "matched === 0"}, {"policy_message.ts", "text = text.split(`{{${name}}}`).join(value);", "text = text;"}, {"strict_options.ts", "if(field < 0) { return false; }", "if(field < 0) { continue; }"}} {
		t.Run(m.file, func(t *testing.T) {
			directory := t.TempDir()
			for _, file := range []string{"main.ts", "options_json.ts", "option_schema.ts", "policy_message.ts", "strict_options.ts"} {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if file == m.file {
					if strings.Count(string(data), m.old) != 1 {
						t.Fatal("mutant anchor changed")
					}
					data = []byte(strings.Replace(string(data), m.old, m.new, 1))
				}
				if err := os.WriteFile(filepath.Join(directory, file), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			got := run(t, "", build(t, directory), cases, catalog)
			if bytes.Equal(got, want) {
				t.Fatal("compiled mutant survived")
			}
			a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
			for i := 0; i < len(a) && i < len(b); i++ {
				if a[i] != b[i] {
					t.Logf("compiled semantic mutant caught at output line %d: got %q, Go %q", i+1, a[i], b[i])
					break
				}
			}
		})
	}
}

func fixture(t *testing.T) string {
	t.Helper()
	f, err := os.Open("testdata/cases.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	reader, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	path := filepath.Join(t.TempDir(), "cases.json")
	out, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.Copy(out, reader); err != nil {
		t.Fatal(err)
	}
	if err = out.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func smallFixture(t *testing.T, corpus any) string {
	t.Helper()
	data, err := json.Marshal(corpus)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "cases.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestMessageRefusalsMatchGo(t *testing.T) {
	binary := build(t, ".")
	goOracle := oracle(t)
	catalog, _ := filepath.Abs("testdata/catalog.json")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.ts")
	cases := []map[string]any{
		{"Kind": "message", "Rule": "missing/rule", "Id": "missing", "Values": map[string]string{}},
		{"Kind": "message", "Rule": "nexus/consistency-no-for-in", "Id": "missing", "Values": map[string]string{}},
		{"Kind": "message", "Rule": "nexus/consistency-no-stuttering-name", "Id": "stutteringName", "Values": map[string]string{}},
		{"Kind": "message", "Rule": "nexus/consistency-no-stuttering-name", "Id": "stutteringName", "Values": map[string]string{"other": "x"}},
		{"Kind": "message", "Rule": "nexus/consistency-no-for-in", "Id": "forIn", "Values": map[string]string{"extra": "x"}},
		{"Kind": "message", "Rule": "nexus/consistency-no-for-in", "Id": "forIn", "Values": map[string]string{}, "Choices": []map[string]string{{"Rule": "nexus/consistency-no-for-in", "Id": "forIn", "Phrase": "reason", "Name": "missing"}}},
	}
	phraseRule := "structure/consistency-require-matching-file-name"
	phraseId := "requireMatchingFileName"
	choice := map[string]string{"Rule": phraseRule, "Id": phraseId, "Phrase": "alternative", "Name": "none"}
	for _, choices := range [][]map[string]string{nil, {choice, choice}, {{"Rule": "wrong/rule", "Id": phraseId, "Phrase": "alternative", "Name": "none"}}, {{"Rule": phraseRule, "Id": "wrong", "Phrase": "alternative", "Name": "none"}}} {
		cases = append(cases, map[string]any{"Kind": "message", "Rule": phraseRule, "Id": phraseId, "Values": map[string]string{}, "Choices": choices})
	}
	for i, c := range cases {
		path := smallFixture(t, map[string]any{"Definitions": map[string]string{}, "Cases": []any{c}})
		want := strings.TrimSpace(string(run(t, "", goOracle, path)))
		if !strings.HasPrefix(want, "panic: ") {
			t.Fatalf("guard %d: Go did not refuse", i)
		}
		expected := "adamic: " + want + "\n"
		for _, command := range [][]string{{binary, path, catalog}, {"node", "--disable-warning=ExperimentalWarning", runner, entry, path, catalog}} {
			// Refusals are silent until their final panic; protect startup with FirstOutput.
			cmd := exec.Command(command[0], command[1:]...)
			output, err := os.CreateTemp(t.TempDir(), "stdout-")
			if err != nil {
				t.Fatal(err)
			}
			cmd.Stdout = output
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			err = childguard.Run(cmd, childguard.Options{})
			output.Close()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 70 || stderr.String() != expected {
				t.Fatalf("guard %d %s: %v stderr %q, Go %q", i, command[0], err, stderr.String(), expected)
			}
		}
	}
	t.Logf("%d message refusal cases match Go on Node and sanitized native", len(cases))
}
// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestKnownGapsAreExplicit(t *testing.T) {
	corpus := map[string]any{"Definitions": map[string]string{
		"regex":  `{"type":"array","items":{"type":"string","pattern":"x"}}`,
		"custom": `{"kind":"unsupported"}`,
		"fold":   `{"kind":"object","fields":{"Ignore":{"tagged":false,"shape":{"kind":"boolean"}}}}`,
	}, "Cases": []map[string]string{
		{"Kind": "schema", "Schema": "regex", "Input": "[]"},
		{"Kind": "decode", "Schema": "custom", "Input": "null"},
		{"Kind": "decode", "Schema": "fold", "Input": "{\"ſgnore\":true}"},
		{"Kind": "decode", "Schema": "fold", "Input": "{}"},
	}}
	path := smallFixture(t, corpus)
	catalog, _ := filepath.Abs("testdata/catalog.json")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.ts")
	want := []byte("NotYet: unsupported schema keyword pattern\nNotYet: custom or unsupported Go option type\nNotYet: Unicode fold for untagged option keys\nvalid\n")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path, catalog), want)
	compare(t, run(t, "", build(t, "."), path, catalog), want)
}
