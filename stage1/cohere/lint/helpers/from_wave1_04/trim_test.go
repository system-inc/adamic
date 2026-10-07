package wave104

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func trimOracle(t *testing.T) string {
	t.Helper()
	root := filepath.Join(repository(t), "cohere")
	local, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	virtual := filepath.Join(root, "adamic_wave104_trim.go")
	replacements := map[string]string{virtual: filepath.Join(local, "trim_oracle.go.txt"), filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_wave104_trim_exports.go"): filepath.Join(local, "trim_exports.go.txt")}
	sourcePath := filepath.Join(root, "internal/lint/rules/tailwind/collapse/data_type.go")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	sourceAnchor := "func scanNumber(value string) int {"
	if strings.Count(string(source), sourceAnchor) != 1 {
		t.Fatal("Go scanner observer anchor changed")
	}
	observed := filepath.Join(t.TempDir(), "data_type.go")
	if err = os.WriteFile(observed, []byte(strings.Replace(string(source), sourceAnchor, sourceAnchor+"\n if wave104ObserveScan { wave104ScanTrace=append(wave104ScanTrace,value) }; if wave104ObserveFraction { wave104FractionTrace=append(wave104FractionTrace,\"N:\"+value) }", 1)), 0644); err != nil {
		t.Fatal(err)
	}
	replacements[sourcePath] = observed
	segmentPath := filepath.Join(root, "internal/lint/rules/tailwind/collapse/segment.go")
	segment, err := os.ReadFile(segmentPath)
	if err != nil {
		t.Fatal(err)
	}
	segmentText := string(segment)
	for _, hook := range []struct{ anchor, body string }{
		{"func hasMathFunction(input string) bool {", `if wave104ObserveFraction {wave104FractionTrace=append(wave104FractionTrace,"M:"+input)}`},
		{"func trimLeadingJavaScriptSpace(value string) string {", `if wave104ObserveFraction {wave104FractionTrace=append(wave104FractionTrace,"T:"+value)}`},
	} {
		if strings.Count(segmentText, hook.anchor) != 1 {
			t.Fatal("Go dependency observer anchor changed")
		}
		segmentText = strings.Replace(segmentText, hook.anchor, hook.anchor+"\n"+hook.body, 1)
	}
	observedSegment := filepath.Join(t.TempDir(), "segment.go")
	if err = os.WriteFile(observedSegment, []byte(segmentText), 0644); err != nil {
		t.Fatal(err)
	}
	replacements[segmentPath] = observedSegment
	data, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	if err = os.WriteFile(overlay, data, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "oracle")
	run(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	return binary
}
func trimCases(t *testing.T, oracle string) string {
	t.Helper()
	input, err := filepath.Abs("trim_fixture_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "cases.json")
	run(t, "", oracle, "--generate", input, path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Spaces []int
		Cases  []struct{ Rule, File, Value string }
	}
	if err = json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, c := range corpus.Cases {
		if c.Rule != "" {
			counts[c.Rule]++
		}
	}
	if len(counts) != 4 || len(corpus.Spaces) != 25 || len(corpus.Cases) < 63488 {
		t.Fatalf("coverage: consumers=%v spaces=%d cases=%d", counts, len(corpus.Spaces), len(corpus.Cases))
	}
	for name, count := range counts {
		t.Logf("%s: %d original fixture literal/field inputs", name, count)
	}
	t.Logf("%d total cases, all 63,488 BMP scalar leading points and 25 actual Go whitespace members", len(corpus.Cases))
	return path
}

// Not parallel: bound native sanitizer resources while comparing the same input population.
func TestTrimLeadingJavaScriptSpaceMatchesGo(t *testing.T) {
	oracle := trimOracle(t)
	path := trimCases(t, oracle)
	want := run(t, "", oracle, path)
	entry, js, binary := buildHelper(t, "trim_leading_javascript_space.a", "trim_main.a", "", "")
	for side, got := range observations(t, entry, js, binary, path) {
		if !bytes.Equal(got, want) {
			t.Fatalf("%s %s", side, firstDifference(got, want))
		}
	}
	t.Logf("%d exact Go bytes on source Node, emitted JavaScript and ASan/UBSan native", len(want))
}

// Not parallel: each semantic mutant must compile and complete without sanitizer findings.
func TestTrimLeadingJavaScriptSpaceMutants(t *testing.T) {
	oracle := trimOracle(t)
	path := trimCases(t, oracle)
	want := run(t, "", oracle, path)
	for _, change := range []struct{ name, from, to string }{{"omit leading BOM", "if(!isJavaScriptSpace(point))", "if(point === 0xfeff || !isJavaScriptSpace(point))"}, {"strip an extra suffix character", "return value.slice(index);", "return value.slice(index + 1);"}} {
		t.Run(change.name, func(t *testing.T) {
			entry, js, binary := buildHelper(t, "trim_leading_javascript_space.a", "trim_main.a", change.from, change.to)
			for side, got := range observations(t, entry, js, binary, path) {
				if bytes.Equal(got, want) {
					t.Fatal("mutant survived", side)
				}
				t.Logf("compiled and completed cleanly; only Go comparison caught %s on %s: %s", change.name, side, firstDifference(got, want))
			}
		})
	}
}
