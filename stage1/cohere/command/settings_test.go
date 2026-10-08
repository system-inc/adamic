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
	properties := filepath.Join(root, "stage1/cohere/command/settings/properties_main.a")
	projection := settingsPropertyExpected(t, expected)
	nodeProperties := run(t, root, "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", properties, manifest)
	if !bytes.Equal(nodeProperties, projection) {
		t.Fatal("Node settings property projection differs from Go")
	}
	for _, sanitize := range []bool{false, true} {
		binary := build(t, root, properties, filepath.Join(temp, fmt.Sprintf("properties-%v", sanitize)), sanitize)
		if got := run(t, root, binary, manifest); !bytes.Equal(got, projection) {
			t.Fatal("native settings property projection differs from Go")
		}
	}
	t.Log("283 exact Go/Node/release/sanitized scalar property projections; original full-object JSON transport remains blocked")
	p, e := load.Load([]string{entry})
	if e != nil {
		t.Fatal(e)
	}
	_, e = lower.Lower(context.Background(), p)
	if e == nil || !strings.Contains(e.Error(), "JSON.stringify object references") || !strings.Contains(e.Error(), "settings/main.a") {
		t.Fatalf("expected retained structural JSON serializer refusal, got %v", e)
	}
	t.Logf("native JSON transport blocked: %v", e)
	gap := filepath.Join(root, "internal/lower/testdata/step42-command-settings/call_spread.a")
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
		{"return { ...answer, houseIgnoreDeclared: true };", "return { ...answer, options: prettier, houseIgnore: [], houseIgnoreDeclared: true };"},
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
		propertyEntry := filepath.Join(directory, "properties_main.a")
		nodeProperties := run(t, root, "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", propertyEntry, manifest)
		binary := build(t, root, propertyEntry, filepath.Join(directory, "native"), false)
		nativeProperties := run(t, root, binary, manifest)
		if !bytes.Equal(nativeProperties, nodeProperties) {
			t.Fatal("settings mutant Node/native property disagreement")
		}
		settingsCompare(t, nativeProperties, projection, cases, true)
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

// #hnm8t56: the ruled checked refusal differs from the current Go oracle.
// Keep this named divergence until that task lands at our cohere pin.
func TestSettingsIntegerNamedDivergence(t *testing.T) {
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
	if !bytes.Contains(goOutput, []byte(`"tabWidth":9007199254740993`)) || !bytes.Contains(nodeOutput, []byte(`"kind":"Refused"`)) || !bytes.Contains(nodeOutput, []byte(`format option \"tabWidth\" value 9007199254740993`)) {
		t.Fatalf("integer boundary changed: Go=%s Node=%s", goOutput, nodeOutput)
	}
	var divergence struct{ Task string }
	if e = json.Unmarshal(read(t, "testdata/settings-integer-divergences.json"), &divergence); e != nil || divergence.Task != "#hnm8t56" {
		t.Fatal("named divergence task missing")
	}
	t.Log("#hnm8t56 named divergence: Go accepts tabWidth=9007199254740993 exactly; Adamic refuses with key/value, status 1")
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
	temp := t.TempDir()
	binary := build(t, root, entry, filepath.Join(temp, "report"), true)
	if native := run(t, root, binary); !bytes.Equal(native, expected) {
		t.Fatalf("native output interface %s want %s", native, expected)
	}
	t.Log("four captured report strings and returned status 1 agree on Node and sanitized native; #kqxkvg0 written-property workaround active")
	gap := filepath.Join(root, "internal/lower/testdata/step42-command-settings/json_shorthand.a")
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

func settingsPropertyExpected(t *testing.T, data []byte) []byte {
	t.Helper()
	out := []byte{}
	for _, line := range bytes.Split(bytes.TrimSpace(data), []byte("\n")) {
		var answer struct {
			Kind, Source string
			Options      struct {
				TabWidth, PrintWidth                                        int
				UseTabs, Semi, SingleQuote, BracketSpacing, BracketSameLine bool
				TrailingComma, ArrowParens, EndOfLine                       string
			}
			HouseIgnore, IgnorePatterns []string
			HouseIgnoreDeclared         bool
		}
		if e := json.Unmarshal(line, &answer); e != nil {
			t.Fatal(e)
		}
		if answer.Kind != "Ok" {
			out = append(out, append(line, '\n')...)
			continue
		}
		v := answer.Options
		payload := struct {
			Kind                string `json:"kind"`
			TabWidth            int    `json:"tabWidth"`
			UseTabs             bool   `json:"useTabs"`
			Semi                bool   `json:"semi"`
			SingleQuote         bool   `json:"singleQuote"`
			PrintWidth          int    `json:"printWidth"`
			TrailingComma       string `json:"trailingComma"`
			BracketSpacing      bool   `json:"bracketSpacing"`
			BracketSameLine     bool   `json:"bracketSameLine"`
			ArrowParens         string `json:"arrowParens"`
			EndOfLine           string `json:"endOfLine"`
			Source              string `json:"source"`
			HouseIgnore         string `json:"houseIgnore"`
			HouseIgnoreCount    int    `json:"houseIgnoreCount"`
			HouseIgnoreDeclared bool   `json:"houseIgnoreDeclared"`
			IgnorePatterns      string `json:"ignorePatterns"`
			IgnorePatternsCount int    `json:"ignorePatternsCount"`
		}{answer.Kind, v.TabWidth, v.UseTabs, v.Semi, v.SingleQuote, v.PrintWidth, v.TrailingComma, v.BracketSpacing, v.BracketSameLine, v.ArrowParens, v.EndOfLine, answer.Source, strings.Join(answer.HouseIgnore, "\x00"), len(answer.HouseIgnore), answer.HouseIgnoreDeclared, strings.Join(answer.IgnorePatterns, "\x00"), len(answer.IgnorePatterns)}
		encoded, e := json.Marshal(payload)
		if e != nil {
			t.Fatal(e)
		}
		out = append(out, append(encoded, '\n')...)
	}
	return out
}

func TestSettingsCheckedIntegers(t *testing.T) {
	t.Parallel()
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	temp := t.TempDir()
	var cases []struct {
		Key, Value, GoDivergenceTask string
		CheckedRefusal               bool
	}
	fixture := read(t, "testdata/settings-integer-cases.json")
	if e = json.Unmarshal(fixture, &cases); e != nil {
		t.Fatal(e)
	}
	if len(cases) != 36 {
		t.Fatal("mandatory integer matrix")
	}
	snapshots := []settingsSnapshot{}
	for index, c := range cases {
		directory := filepath.Join(temp, fmt.Sprintf("case%d", index))
		path := filepath.Join(directory, "CohereSettings.json")
		source := fmt.Sprintf(`{"format":{"%s":%s}}`, c.Key, c.Value)
		write(t, path, []byte(source))
		snapshots = append(snapshots, settingsSnapshot{directory, map[string]string{path: source}, []string{filepath.Join(directory, "probe.css")}})
	}
	data, e := json.Marshal(snapshots)
	if e != nil {
		t.Fatal(e)
	}
	manifest := filepath.Join(temp, "go.json")
	write(t, manifest, data)
	overlayData, _ := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "cohere/command/formatter_comparison/main.go"): filepath.Join(root, "stage1/cohere/command/testdata/settings_oracle.go")}})
	overlay := filepath.Join(temp, "overlay.json")
	write(t, overlay, overlayData)
	oracle := filepath.Join(temp, "oracle")
	run(t, filepath.Join(root, "cohere"), "go", "build", "-overlay", overlay, "-o", oracle, "./command/formatter_comparison")
	goLines := bytes.Split(bytes.TrimSpace(run(t, root, oracle, manifest)), []byte("\n"))
	if len(goLines) != len(cases) {
		t.Fatal("Go integer outcomes")
	}
	expected := []byte{}
	divergences := 0
	for index, c := range cases {
		var answer struct {
			Kind, Message string
			Options       map[string]json.RawMessage
		}
		if e = json.Unmarshal(goLines[index], &answer); e != nil {
			t.Fatal(e)
		}
		line := ""
		if c.CheckedRefusal {
			if answer.Kind != "Ok" || c.GoDivergenceTask != "#hnm8t56" {
				t.Fatalf("named divergence %s=%s changed: %s", c.Key, c.Value, goLines[index])
			}
			divergences++
			message := fmt.Sprintf(`/CohereSettings.json: format option %q value %s is outside the safe integer range [-9007199254740991, 9007199254740991]`, c.Key, c.Value)
			encoded, e := json.Marshal(message)
			if e != nil {
				t.Fatal(e)
			}
			line = "Refused:" + string(encoded)
		} else if answer.Kind == "Ok" {
			line = "Ok:" + string(answer.Options[c.Key])
		} else {
			message := strings.ReplaceAll(answer.Message, filepath.Join(snapshots[index].CWD, "CohereSettings.json"), "/CohereSettings.json")
			encoded, e := json.Marshal(message)
			if e != nil {
				t.Fatal(e)
			}
			line = "Refused:" + string(encoded)
		}
		expected = append(expected, []byte(line+"\n")...)
	}
	input := filepath.Join(temp, "integers.json")
	write(t, input, fixture)
	entry := filepath.Join(root, "stage1/cohere/command/settings/integer_main.a")
	node := run(t, root, "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", entry, input)
	if !bytes.Equal(node, expected) {
		t.Fatalf("checked integer Node\ngot %s\nwant %s", node, expected)
	}
	for _, sanitize := range []bool{false, true} {
		binary := build(t, root, entry, filepath.Join(temp, fmt.Sprintf("integers-%v", sanitize)), sanitize)
		if got := run(t, root, binary, input); !bytes.Equal(got, expected) {
			t.Fatal("checked integer native mismatch")
		}
	}
	// Removing the check must finish with wrong output; it cannot satisfy the ruling by silently rounding.
	directory := filepath.Join(temp, "mutant")
	files, e := filepath.Glob("settings/*.a")
	if e != nil {
		t.Fatal(e)
	}
	original := string(read(t, "settings/resolve.a"))
	target := "if(number && !safeInteger(value.raw))"
	if strings.Count(original, target) != 1 {
		t.Fatal("integer mutant target")
	}
	for _, file := range files {
		contents := read(t, file)
		if filepath.Base(file) == "resolve.a" {
			contents = []byte(strings.Replace(original, target, "if(false)", 1))
		}
		write(t, filepath.Join(directory, filepath.Base(file)), contents)
	}
	mutated := filepath.Join(directory, "integer_main.a")
	wrongNode := run(t, root, "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", mutated, input)
	binary := build(t, root, mutated, filepath.Join(directory, "native"), false)
	wrongNative := run(t, root, binary, input)
	if !bytes.Equal(wrongNode, wrongNative) || bytes.Equal(wrongNode, expected) {
		t.Fatal("unchecked rounding mutant must fail identically on Node/native")
	}
	t.Logf("36 integer decisions exact on Node/release/sanitized; %d named Go divergences #hnm8t56; native/Node rounding mutant caught", divergences)
}

// This selector must fail when the ruled host API lands, forcing a real adapter
// instead of allowing the UTF-8 fixture host to stand in for branded Path I/O.
func TestSettingsPathAPISelector(t *testing.T) {
	t.Parallel()
	temp := t.TempDir()
	path := filepath.Join(temp, "host_path.a")
	write(t, path, read(t, "testdata/host_path.a.txt"))
	_, e := load.Load([]string{path})
	if e == nil || !strings.Contains(e.Error(), `has no exported member 'Path'`) {
		t.Fatalf("Path host boundary changed; implement the real adapter: %v", e)
	}
	t.Logf("real Path-backed adapter blocked at shared prelude: %v", e)
}
