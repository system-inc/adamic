package command

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

type settingsCase struct {
	Name, Source string
	Files        map[string]string
	Queries      []string
}
type settingsSnapshot struct {
	CWD     string            `json:"cwd"`
	Files   map[string]string `json:"files"`
	Queries []string          `json:"queries"`
}
type settingsPin struct {
	Repo, Commit string
	Files        []struct{ Source, Fixture, SHA256 string }
}

func settingsFixture(t *testing.T, fixture, hash string) string {
	t.Helper()
	data := read(t, filepath.Join("testdata", fixture))
	if fmt.Sprintf("%x", sha256.Sum256(data)) != hash {
		t.Fatalf("hash %s", fixture)
	}
	return string(data)
}
func settingsInputs(t *testing.T, temp string) ([]settingsSnapshot, []settingsCase) {
	t.Helper()
	var cases []settingsCase
	if e := json.Unmarshal(read(t, "testdata/settings-cases.json"), &cases); e != nil {
		t.Fatal(e)
	}
	var pins, packages []settingsPin
	if e := json.Unmarshal(read(t, "testdata/settings-corpus.json"), &pins); e != nil {
		t.Fatal(e)
	}
	if e := json.Unmarshal(read(t, "testdata/corpus.json"), &packages); e != nil {
		t.Fatal(e)
	}
	if len(pins) != 23 || len(packages) != 23 {
		t.Fatal("all 23 pins are mandatory")
	}
	configs := 0
	for index, pin := range pins {
		if pin.Repo != packages[index].Repo || pin.Commit != packages[index].Commit {
			t.Fatal("pin mismatch")
		}
		files := map[string]string{}
		for _, file := range packages[index].Files {
			if filepath.Base(file.Fixture) == "package.json" {
				files["package.json"] = settingsFixture(t, file.Fixture, file.SHA256)
			}
		}
		if _, ok := files["package.json"]; !ok {
			t.Fatalf("missing manifest %s", pin.Repo)
		}
		queries := []string{"source/probe.css"}
		for _, file := range pin.Files {
			files[file.Source] = settingsFixture(t, file.Fixture, file.SHA256)
			configs++
			query := filepath.Join(filepath.Dir(file.Source), "source/probe.css")
			found := false
			for _, q := range queries {
				if q == query {
					found = true
				}
			}
			if !found {
				queries = append(queries, query)
			}
		}
		cases = append(cases, settingsCase{pin.Repo + "@" + pin.Commit, "testdata/settings-corpus.json", files, queries})
	}
	if configs != 20 {
		t.Fatalf("configs=%d want 20", configs)
	}
	snapshots := []settingsSnapshot{}
	for index, c := range cases {
		directory := filepath.Join(temp, "trees", fmt.Sprintf("%03d", index))
		if e := os.MkdirAll(directory, 0755); e != nil {
			t.Fatal(e)
		}
		snapshot := settingsSnapshot{directory, map[string]string{}, []string{}}
		for name, text := range c.Files {
			path := filepath.Join(directory, name)
			write(t, path, []byte(text))
			snapshot.Files[path] = text
		}
		for _, query := range c.Queries {
			path := filepath.Join(directory, query)
			snapshot.Queries = append(snapshot.Queries, path)
			if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
				t.Fatal(e)
			}
		}
		snapshots = append(snapshots, snapshot)
	}
	t.Logf("control sets=%d public pins=%d config blobs=%d total query sets=%d", len(cases)-23, len(pins), configs, len(cases))
	return snapshots, cases
}
func settingsCompare(t *testing.T, got, expected []byte, cases []settingsCase, requireDifference bool) int {
	t.Helper()
	g, e := bytes.Split(bytes.TrimSpace(got), []byte("\n")), bytes.Split(bytes.TrimSpace(expected), []byte("\n"))
	if len(g) != len(e) {
		t.Fatalf("answers=%d want %d", len(g), len(e))
	}
	index, differences := 0, 0
	for _, c := range cases {
		for range c.Queries {
			if !bytes.Equal(g[index], e[index]) {
				differences++
				if !requireDifference {
					t.Logf("%s (%s)\ngot %s\nwant %s", c.Name, c.Source, g[index], e[index])
				}
			}
			index++
		}
	}
	if !requireDifference && differences != 0 {
		t.Fatalf("settings differences=%d", differences)
	}
	if requireDifference && differences == 0 {
		t.Fatal("mutant survived")
	}
	t.Logf("exact answer lines=%d differences=%d", len(g), differences)
	return differences
}
func TestSettingsResolution(t *testing.T) {
	t.Parallel()
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	temp := t.TempDir()
	snapshots, cases := settingsInputs(t, temp)
	data, e := json.Marshal(snapshots)
	if e != nil {
		t.Fatal(e)
	}
	manifest := filepath.Join(temp, "manifest.json")
	write(t, manifest, data)
	overlayData, _ := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "cohere/command/formatter_comparison/main.go"): filepath.Join(root, "stage1/cohere/command/testdata/settings_oracle.go")}})
	overlay := filepath.Join(temp, "overlay.json")
	write(t, overlay, overlayData)
	oracle := filepath.Join(temp, "oracle")
	run(t, filepath.Join(root, "cohere"), "go", "build", "-overlay", overlay, "-o", oracle, "./command/formatter_comparison")
	expected := run(t, root, oracle, manifest)
	entry := filepath.Join(root, "stage1/cohere/command/settings/main.a")
	actual := run(t, root, "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", entry, manifest)
	settingsCompare(t, actual, expected, cases, false)
	p, e := load.Load([]string{entry})
	if e != nil {
		t.Fatal(e)
	}
	_, e = lower.Lower(context.Background(), p)
	if e == nil || !strings.Contains(e.Error(), "stage 0 can't lower a SpreadElement yet") || !strings.Contains(e.Error(), "settings/resolve.a") {
		t.Fatalf("expected retained shared compiler refusal, got %v", e)
	}
	t.Logf("native integration blocked: %v", e)
	gap := filepath.Join(root, "stage1/cohere/command/settings/gaps/call_spread.a")
	if got := run(t, root, "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", gap); string(got) != "x\n" {
		t.Fatalf("call-spread reduction: %s", got)
	}
	gp, ge := load.Load([]string{gap})
	if ge != nil {
		t.Fatal(ge)
	}
	_, ge = lower.Lower(context.Background(), gp)
	if ge == nil || !strings.Contains(ge.Error(), "SpreadElement") {
		t.Fatalf("call-spread reduction refusal: %v", ge)
	}
	t.Logf("shortest call-spread proof: %v", ge)
	// This is a proven boundary, not native settings parity credit.
	originals := string(read(t, "settings/resolve.a"))
	mutants := []struct{ from, to string }{
		{"let answer = this.house();", "let answer: Answer = { kind: 'Ok', options: prettier, source: '', houseIgnore: [], houseIgnoreDeclared: true, ignorePatterns: [] };"},
		{"if(leftover !== '') return refused", "if(leftover === 'unreachable') return refused"},
		{"return up === directory ? this.house() : this.resolve(up);", "return this.house();"},
		{"const sources = [...read.layers].reverse();", "const sources = [...read.layers];"},
	}
	for index, mutant := range mutants {
		if strings.Count(originals, mutant.from) != 1 {
			t.Fatalf("mutant %d target", index)
		}
		directory := filepath.Join(temp, fmt.Sprintf("mutant%d", index))
		files, e := filepath.Glob("settings/*.a")
		if e != nil {
			t.Fatal(e)
		}
		for _, file := range files {
			data := read(t, file)
			if filepath.Base(file) == "resolve.a" {
				data = []byte(strings.Replace(originals, mutant.from, mutant.to, 1))
			}
			write(t, filepath.Join(directory, filepath.Base(file)), data)
		}
		got := run(t, root, "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", filepath.Join(directory, "main.a"), manifest)
		t.Logf("source Node mutant %d", index)
		settingsCompare(t, got, expected, cases, true)
	}
	// Every input set is supplied to the upstream tree-comparison test: no skip.
	c := exec.Command("go", "test", "-v", "-count=1", "./internal/format/formatoptions")
	c.Dir = filepath.Join(root, "cohere")
	trees := []string{filepath.Join(temp, "trees")}
	c.Env = append(os.Environ(), "COHERE_RESOLVE_TREES="+strings.Join(trees, string(os.PathListSeparator)))
	upstream, e := c.CombinedOutput()
	if e != nil || bytes.Contains(upstream, []byte("--- SKIP:")) {
		t.Fatalf("upstream settings tests: %v\n%s", e, upstream)
	}
	t.Logf("upstream settings tests (no skips):\n%s", upstream)
	if keep := os.Getenv("ADAMIC_SETTINGS_KEEP"); keep != "" {
		write(t, filepath.Join(keep, "settings-go.jsonl"), expected)
		write(t, filepath.Join(keep, "settings-node.jsonl"), actual)
	}
}
func TestSettingsJSON(t *testing.T) {
	t.Parallel()
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	temp := t.TempDir()
	var cases []settingsCase
	if e = json.Unmarshal(read(t, "testdata/settings-cases.json"), &cases); e != nil {
		t.Fatal(e)
	}
	inputs := []string{"", " ", "[", "{\"a\":1", "{\"a\"", "[1 2]", "\"\\uD800\"", "\"\\uDC00\"", "\"\\uD83D\\uDE00\""}
	for _, c := range cases {
		for _, text := range c.Files {
			inputs = append(inputs, text)
		}
	}
	data, e := json.Marshal(inputs)
	if e != nil {
		t.Fatal(e)
	}
	manifest := filepath.Join(temp, "json.json")
	write(t, manifest, data)
	expected := []byte{}
	for _, text := range inputs {
		var raw json.RawMessage
		e = json.Unmarshal([]byte(text), &raw)
		message := ""
		if e != nil {
			message = e.Error()
		}
		expected = append(expected, []byte(message+"\n")...)
	}
	entry := filepath.Join(root, "stage1/cohere/command/settings/json_main.a")
	got := run(t, root, "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", entry, manifest)
	if !bytes.Equal(got, expected) {
		lines, want := bytes.Split(got, []byte("\n")), bytes.Split(expected, []byte("\n"))
		for i := range inputs {
			if !bytes.Equal(lines[i], want[i]) {
				t.Fatalf("source JSON input %q: got %q want %q", inputs[i], lines[i], want[i])
			}
		}
		t.Fatal("JSON framing mismatch")
	}
	for _, sanitize := range []bool{false, true} {
		binary := build(t, root, entry, filepath.Join(temp, fmt.Sprintf("json-%v", sanitize)), sanitize)
		got = run(t, root, binary, manifest)
		if !bytes.Equal(got, expected) {
			t.Fatal("native JSON mismatch")
		}
	}
	t.Logf("strict JSON inputs=%d Go/Node/release/sanitized exact", len(inputs))
	original := string(read(t, "settings/json.a"))
	mutations := []struct{ from, to string }{
		{"[' ', '\\t', '\\r', '\\n']", "[' ', '\\t', '\\r', '\\n', '\\ufeff']"},
		{"if(this.message === '' && this.position !== this.source.length) this.fail('after top-level value');", "if(false) this.fail('after top-level value');"},
	}
	for index, mutation := range mutations {
		if strings.Count(original, mutation.from) != 1 {
			t.Fatalf("JSON mutant %d target", index)
		}
		directory := filepath.Join(temp, fmt.Sprintf("json-mutant%d", index))
		write(t, filepath.Join(directory, "json.a"), []byte(strings.Replace(original, mutation.from, mutation.to, 1)))
		for _, file := range []string{"json_main.a", "printable.a"} {
			write(t, filepath.Join(directory, file), read(t, filepath.Join("settings", file)))
		}
		mutated := filepath.Join(directory, "json_main.a")
		node := run(t, root, "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", mutated, manifest)
		binary := build(t, root, mutated, filepath.Join(directory, "native"), false)
		native := run(t, root, binary, manifest)
		if !bytes.Equal(node, native) || bytes.Equal(native, expected) {
			t.Fatalf("JSON mutant %d must finish and disagree with Go identically on Node and native", index)
		}
		t.Logf("JSON mutant %d: successful wrong output caught on Node and native", index)
	}

}

// A valid Go int64 beyond JavaScript's exact integers is an unresolved language
// representation boundary. This proves the mismatch; it grants no parity credit.
func TestSettingsIntegerRepresentationBoundary(t *testing.T) {
	t.Parallel()
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	temp := t.TempDir()
	path := filepath.Join(temp, "CohereSettings.json")
	source := string(read(t, "testdata/settings-integer-boundary.json"))
	write(t, path, []byte(source))
	snapshots := []settingsSnapshot{{temp, map[string]string{path: source}, []string{filepath.Join(temp, "probe.css")}}}
	data, e := json.Marshal(snapshots)
	if e != nil {
		t.Fatal(e)
	}
	manifest := filepath.Join(temp, "cases.json")
	write(t, manifest, data)
	overlayData, _ := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "cohere/command/formatter_comparison/main.go"): filepath.Join(root, "stage1/cohere/command/testdata/settings_oracle.go")}})
	overlay := filepath.Join(temp, "overlay.json")
	write(t, overlay, overlayData)
	oracle := filepath.Join(temp, "oracle")
	run(t, filepath.Join(root, "cohere"), "go", "build", "-overlay", overlay, "-o", oracle, "./command/formatter_comparison")
	goOutput := run(t, root, oracle, manifest)
	nodeOutput := run(t, root, "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", filepath.Join(root, "stage1/cohere/command/settings/main.a"), manifest)
	if !bytes.Contains(goOutput, []byte(`"tabWidth":9007199254740993`)) || !bytes.Contains(nodeOutput, []byte(`"tabWidth":9007199254740992`)) {
		t.Fatalf("integer boundary changed: Go=%s Node=%s", goOutput, nodeOutput)
	}
	t.Logf("boundary: Go accepts tabWidth=9007199254740993; Node number becomes 9007199254740992")
}

func TestSettingsOutputInterface(t *testing.T) {
	t.Parallel()
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	entry := filepath.Join(root, "stage1/cohere/command/settings/report_main.a")
	expected := []byte{}
	for _, message := range []string{"", "bad configuration", "a\nb", "\ufeffé😀"} {
		value, e := json.Marshal(struct {
			Status int    `json:"status"`
			Stdout string `json:"stdout"`
			Stderr string `json:"stderr"`
		}{1, "", message + "\n"})
		if e != nil {
			t.Fatal(e)
		}
		expected = append(expected, append(value, '\n')...)
	}
	got := run(t, root, "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", entry)
	if !bytes.Equal(got, expected) {
		t.Fatalf("source output interface %s want %s", got, expected)
	}
	p, e := load.Load([]string{entry})
	if e != nil {
		t.Fatal(e)
	}
	_, e = lower.Lower(context.Background(), p)
	if e == nil || !strings.Contains(e.Error(), "JSON.stringify a literal with spread, shorthand or methods") || !strings.Contains(e.Error(), "settings/report_main.a") {
		t.Fatalf("retained report compiler boundary: %v", e)
	}
	t.Logf("four captured Node report strings and returned status 1 agree; native reporting proof blocked: %v", e)
	gap := filepath.Join(root, "stage1/cohere/command/settings/gaps/json_shorthand.a")
	if got := run(t, root, "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", gap); string(got) != "{\"status\":1}\n" {
		t.Fatalf("JSON shorthand reduction: %s", got)
	}
	gp, ge := load.Load([]string{gap})
	if ge != nil {
		t.Fatal(ge)
	}
	_, ge = lower.Lower(context.Background(), gp)
	if ge == nil || !strings.Contains(ge.Error(), "JSON.stringify a literal with spread, shorthand or methods") {
		t.Fatalf("JSON shorthand reduction refusal: %v", ge)
	}
	t.Logf("shortest shorthand proof: %v", ge)
}
