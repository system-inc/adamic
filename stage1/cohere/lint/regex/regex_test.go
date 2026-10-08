package regex

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func run(t *testing.T, cwd, command string, args ...string) []byte {
	t.Helper()
	c := exec.Command(command, args...)
	c.Dir = cwd
	var out, err bytes.Buffer
	c.Stdout = &out
	c.Stderr = &err
	if e := c.Run(); e != nil || err.Len() != 0 {
		t.Fatalf("%s: %v stdout=%s stderr=%s", command, e, out.String(), err.String())
	}
	return out.Bytes()
}
func backends(t *testing.T, root string) [][]byte {
	t.Helper()
	entry := filepath.Join(root, "testdata/runner.a")
	p, e := load.Load([]string{entry})
	if e != nil {
		t.Fatal(e)
	}
	ir, e := lower.Lower(context.Background(), p)
	if e != nil {
		t.Fatal(e)
	}
	temp := t.TempDir()
	bin := filepath.Join(temp, "runner")
	if e := native.Build(native.C(ir), bin, native.Options{Sanitize: true}); e != nil {
		t.Fatal(e)
	}
	js := filepath.Join(temp, "runner.js")
	if e := os.WriteFile(js, []byte(javascript.JavaScript(ir)), 0644); e != nil {
		t.Fatal(e)
	}
	repository, _ := filepath.Abs("../../../..")
	return [][]byte{run(t, root, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), entry), run(t, root, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), js), run(t, root, bin)}
}
func firstRow(a, b []byte) int {
	left := strings.Split(string(a), "\n")
	right := strings.Split(string(b), "\n")
	for i, s := range left {
		if i >= len(right) || s != right[i] {
			n, _ := strconv.Atoi(strings.Split(s, ":")[0])
			return n
		}
	}
	return 0
}
func TestFixedPatterns(t *testing.T) {
	// Not parallel: recompiles the same sizable generated corpus for a semantic mutant.
	root, _ := filepath.Abs(".")
	cohere, _ := filepath.Abs("../../../../cohere")
	expected := run(t, cohere, "go", "run", filepath.Join(root, "testdata/oracle.go"), filepath.Join(root, "table.json"), filepath.Join(root, "testdata/corpus.json"))
	var rows []struct {
		ID      string  `json:"id"`
		Pattern *string `json:"go_pattern"`
	}
	data, _ := os.ReadFile("table.json")
	if e := json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	ids := []string{}
	for _, r := range rows {
		if r.Pattern != nil {
			ids = append(ids, r.ID)
		}
	}
	for backend, actual := range backends(t, root) {
		if !bytes.Equal(expected, actual) {
			t.Fatalf("backend %d differs: %s", backend, ids[firstRow(expected, actual)])
		}
	}
	t.Logf("%d fixed patterns, %d match records, %d identical bytes on Go, Node, emitted JS and sanitized native", len(ids), bytes.Count(expected, []byte("\n")), len(expected))
	temp := t.TempDir()
	if e := os.MkdirAll(filepath.Join(temp, "testdata"), 0755); e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"patterns.a", "testdata/runner.a", "testdata/corpus.a"} {
		content, e := os.ReadFile(name)
		if e != nil {
			t.Fatal(e)
		}
		if name == "patterns.a" {
			content = bytes.Replace(content, []byte("/PaginationInput("), []byte("/PaginationInputBROKEN("), 1)
		}
		if e := os.WriteFile(filepath.Join(temp, name), content, 0644); e != nil {
			t.Fatal(e)
		}
	}
	for backend, actual := range backends(t, temp) {
		if bytes.Equal(expected, actual) {
			t.Fatalf("translation mutant escaped backend %d", backend)
		}
		t.Logf("translation mutant caught by byte comparison on backend %d: %s", backend, ids[firstRow(expected, actual)])
	}
}
func TestDynamicPatternGap(t *testing.T) {
	entry, _ := filepath.Abs("testdata/dynamic_gap.a")
	repository, _ := filepath.Abs("../../../..")
	out := run(t, ".", "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), entry, "TODO")
	if string(out) != "true\n" {
		t.Fatalf("Node %q", out)
	}
	p, e := load.Load([]string{entry})
	if e != nil {
		t.Fatal(e)
	}
	lowered, e := lower.Lower(context.Background(), p)
	if e != nil {
		t.Fatalf("runtime RegExp acceptance blocked on merged area/library: %v", e)
	}
	directory := t.TempDir()
	js := filepath.Join(directory, "dynamic.js")
	binary := filepath.Join(directory, "dynamic")
	if e := os.WriteFile(js, []byte(javascript.JavaScript(lowered)), 0644); e != nil {
		t.Fatal(e)
	}
	if e := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); e != nil {
		t.Fatal(e)
	}
	for _, actual := range [][]byte{run(t, ".", "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), js, "TODO"), run(t, ".", binary, "TODO")} {
		if !bytes.Equal(actual, out) {
			t.Fatalf("dynamic constructor differs from Node: %q", actual)
		}
	}
	t.Log("dynamic constructor emitted JS and sanitized native acceptance now green")
}

func optionAnswers(t *testing.T, cases string) ([]byte, []byte) {
	t.Helper()
	root, _ := filepath.Abs(".")
	cohere, _ := filepath.Abs("../../../../cohere")
	directory := t.TempDir()
	virtual := filepath.Join(cohere, "adamic_regex_options_oracle.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(root, "testdata/option_dialects.go")}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "options-oracle")
	run(t, cohere, "go", "build", "-overlay="+overlayPath, "-o", binary, virtual)
	corpus := filepath.Join(root, "testdata", cases)
	return run(t, cohere, binary, corpus), run(t, root, "node", filepath.Join(root, "testdata/option_dialects.mjs"), corpus)
}
func TestOptionPatternsAgreeWithESRegexp(t *testing.T) {
	goAnswer, nodeAnswer := optionAnswers(t, "option_cases.json")
	if !bytes.Equal(goAnswer, nodeAnswer) {
		t.Fatalf("esregexp/Node option comparison differs: Go %q Node %q", goAnswer, nodeAnswer)
	}
	if bytes.Count(goAnswer, []byte("\n")) != 9 {
		t.Fatal("option controls missing")
	}
	t.Logf("9 unchanged JavaScript option sources agree against cohere esregexp Compile(source, u), including syntax rejection: %q", goAnswer)
}
func TestOptionPropertyGap(t *testing.T) {
	goAnswer, nodeAnswer := optionAnswers(t, "option_property_gap.json")
	if string(goAnswer) != "\\p{Script=Greek}\tSyntaxError\n" || string(nodeAnswer) != "\\p{Script=Greek}\ttrue\n" {
		t.Fatalf("property gap changed; remove the exclusion and extend agreement cases: Go %q Node %q", goAnswer, nodeAnswer)
	}
	t.Logf("named pinned esregexp property gap, excluded from agreement controls: Go %q Node %q", goAnswer, nodeAnswer)
}

func TestInventoryMatchesPinnedSource(t *testing.T) {
	root, _ := filepath.Abs(".")
	cohere, _ := filepath.Abs("../../../../cohere")
	actual := run(t, cohere, "go", "run", filepath.Join(root, "testdata/inventory.go"), filepath.Join(cohere, "internal/lint/rules"))
	expected, err := os.ReadFile("sites.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(expected, actual) {
		t.Fatal("regexp census drifted from pinned cohere AST; regenerate and review the table")
	}
}

func TestOptionInventoryMatchesPinnedSource(t *testing.T) {
	root, _ := filepath.Abs(".")
	cohere, _ := filepath.Abs("../../../../cohere")
	actual := run(t, cohere, "go", "run", filepath.Join(root, "testdata/inventory.go"), filepath.Join(cohere, "internal/lint/rules"), "--esregexp")
	expected, err := os.ReadFile("option-sites.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(expected, actual) {
		t.Fatal("esregexp option census drifted from pinned AST")
	}
	t.Log("34 esregexp sites retained separately from the 89 Go-regexp sites; unchanged JavaScript sources")
}
func TestCompiledOptionCaptureSources(t *testing.T) {
	root, _ := filepath.Abs(".")
	cohere, _ := filepath.Abs("../../../../cohere")
	directory := t.TempDir()
	helper, err := os.ReadFile("testdata/capture_options.go")
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("ADAMIC_CAPTURE_SOURCE_MUTANT") == "1" {
		helper = []byte(strings.Replace(string(helper), "return pattern.Source()", "return map[string]any{}", 1))
		helper = []byte(strings.Replace(string(helper), "if pattern, ok :=", "if _, ok :=", 1))
	}
	side := filepath.Join(directory, "capture_options.go")
	if err := os.WriteFile(side, helper, 0644); err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(cohere, "internal/lint/testing")
	data, _ := json.Marshal(map[string]any{"Replace": map[string]string{
		filepath.Join(base, "adamic_capture_options.go"):      side,
		filepath.Join(base, "adamic_capture_options_test.go"): filepath.Join(root, "testdata/capture_options_test.go"),
	}})
	overlay := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlay, data, 0644); err != nil {
		t.Fatal(err)
	}
	actual := run(t, cohere, "go", "test", "-overlay="+overlay, "./internal/lint/testing", "-run", "^TestAdamicCompiledOptionSources$", "-v", "-count=1")
	t.Logf("source-preserving port capture: %s", actual)
}

func TestNativeSplitAttributeGap(t *testing.T) {
	entry, _ := filepath.Abs("testdata/split_gap.a")
	repository, _ := filepath.Abs("../../../..")
	expected := run(t, ".", "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), entry)
	if string(expected) != "true\ntrue\n" {
		t.Fatalf("shortest split witness: %q", expected)
	}
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	source := native.C(lowered)
	if err := os.WriteFile(filepath.Join(directory, "split_gap.c"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	err = native.Build(source, filepath.Join(directory, "split-gap"), native.Options{Sanitize: true, Split: true, Jobs: 1})
	if err == nil || !strings.Contains(err.Error(), "duplicate definition __attribute__") {
		t.Fatalf("split attribute gap changed; remove blocker and require native finding legs: %v", err)
	}
	t.Logf("shortest native split gap: two RegExp literals: %v", err)
}

func TestNativeCheckerRegexLinkGap(t *testing.T) {
	entry, _ := filepath.Abs("testdata/dynamic_gap.a")
	repository, _ := filepath.Abs("../../../..")
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "dynamic-with-checker")
	err = native.BuildTSGo(native.C(lowered), binary, regexCheckerArchive(t, repository), native.Options{Sanitize: true})
	if err != nil {
		t.Fatal(err)
	}
	actual := run(t, ".", binary)
	if string(actual) != "true\n" {
		t.Fatalf("checker runtime constructor: %q", actual)
	}
	t.Log("dynamic_gap.a passes canonical checker-linked sanitized native")
}
