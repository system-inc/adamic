package css

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/system-inc/adamic/internal/native"
)

// Opt-in immutable artifacts let profiling and comparison share the same source.
func TestCSSProfileArtifacts(t *testing.T) {
	directory := os.Getenv("ADAMIC_CSS_PROFILE_DIR")
	if directory == "" {
		// census: measurement Opt-in timing or profile artifact comparison; does not replace correctness verification.
		t.Skip("set ADAMIC_CSS_PROFILE_DIR")
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	cases, _ := askedCases(t)
	expected := printerAnswers(t, cases, "default")
	narrow := printerAnswers(t, cases, "narrow")
	script, _ := filepath.Abs("testdata/print_library.mjs")
	repo, _ := filepath.Abs(repository)
	original := execute(t, nil, "node", script, filepath.Join(repo, "cohere/internal/format/prettier/bundles"), cases, "default", "fork")
	if original.exitCode != 0 || len(original.stderr) != 0 {
		t.Fatal(string(original.stderr))
	}
	data, _ := os.ReadFile(cases)
	inputs := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	goLines := strings.Split(strings.TrimSuffix(expected, "\n"), "\n")
	jsLines := strings.Split(strings.TrimSuffix(string(original.stdout), "\n"), "\n")
	var shared []string
	var units []int
	var manifest strings.Builder
	largest, largestSize := 0, 0
	for i, input := range inputs {
		var a, b string
		if json.Unmarshal([]byte(goLines[i*2+1]), &a) == nil && json.Unmarshal([]byte(jsLines[i*2+1]), &b) == nil && a == b {
			if len(input) > largestSize {
				largest, largestSize = len(shared), len(input)
			}
			shared = append(shared, input)
			units = append(units, len(utf16.Encode([]rune(a))))
		}
	}
	var sample []string
	totalUnits := 0
	for i, input := range shared {
		if i < 32 || i%16 == 0 || i == largest {
			sample = append(sample, input)
			totalUnits += units[i]
			fmt.Fprintf(&manifest, "%d\t%d\t%d\n", i, len(input), units[i])
		}
	}
	for name, data := range map[string][]byte{"cases.txt": data, "expected.txt": []byte(expected), "narrow-expected.txt": []byte(narrow), "shared.txt": []byte(strings.Join(shared, "\n") + "\n"), "sample.txt": []byte(strings.Join(sample, "\n") + "\n"), "manifest.tsv": []byte(manifest.String()), "sample-expected.txt": []byte(fmt.Sprintf("%d of %d stylesheets formatted, %d units\n", len(sample), len(sample), totalUnits))} {
		if err := os.WriteFile(filepath.Join(directory, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	copied := portDirectory(t, nil)
	snapshot := filepath.Join(directory, "source")
	if err := os.MkdirAll(snapshot, 0755); err != nil {
		t.Fatal(err)
	}
	parent := filepath.Dir(copied)
	for _, slice := range []string{"css", "selector", "values", "mediaquery", "cssstrings", "cssnumbers"} {
		entries, err := os.ReadDir(filepath.Join(parent, slice))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(snapshot, slice), 0755); err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			data, err := os.ReadFile(filepath.Join(parent, slice, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(snapshot, slice, entry.Name()), data, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	program := lowered(t, filepath.Join(snapshot, "css/print_main.ts"))
	code := native.C(program)
	source := filepath.Join(directory, "main.c")
	if err := os.WriteFile(source, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	for _, build := range []struct {
		name    string
		options native.Options
	}{{"printer", native.Options{}}, {"counted", native.Options{Count: true}}} {
		if err := native.Build(code, filepath.Join(directory, build.name), build.options); err != nil {
			t.Fatal(err)
		}
	}
	runtimeDir := filepath.Join(repo, "internal/native/runtime")
	files, _ := filepath.Glob(filepath.Join(runtimeDir, "*.c"))
	args := append(native.Flags(native.Options{}), "-g", "-I", runtimeDir, "-o", filepath.Join(directory, "profiled"), source)
	args = append(args, files...)
	args = append(args, "-lm")
	result := execute(t, nil, "clang", args...)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("profile build %d %s", result.exitCode, result.stderr)
	}

	request, _ := json.Marshal(map[string]any{"Cases": filepath.Join(directory, "shared.txt"), "Repeat": "once", "Rounds": 1})
	if err := os.WriteFile(filepath.Join(directory, "go-request.json"), request, 0644); err != nil {
		t.Fatal(err)
	}
	side, _ := filepath.Abs("testdata/print_side_test.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(repo, "cohere/internal/format/css/adamic_print_side_test.go"): side}})
	overlayPath := filepath.Join(directory, "go-overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	cmd := bounded(t, "go", "test", "-c", "-overlay="+overlayPath, "-o", filepath.Join(directory, "go-printer"), "./internal/format/css")
	cmd.Dir = filepath.Join(repo, "cohere")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Go benchmark build: %v %s", err, output)
	}
	t.Logf("saved %d shared inputs and %d profile inputs, %d output units", len(shared), len(sample), totalUnits)
}

func TestCSSProfileSnapshotsAgree(t *testing.T) {
	asked := os.Getenv("ADAMIC_CSS_PROFILE_SNAPSHOTS")
	if asked == "" {
		// census: measurement Opt-in timing or profile artifact comparison; does not replace correctness verification.
		t.Skip("set ADAMIC_CSS_PROFILE_SNAPSHOTS")
	}
	cases, _ := askedCases(t)
	for _, mode := range []string{"default", "narrow"} {
		expected := printerAnswers(t, cases, mode)
		for _, directory := range filepath.SplitList(asked) {
			js := onNode(t, filepath.Join(directory, "source/css/print_main.ts"), cases, "output", "once", mode)
			if js.exitCode != 0 || len(js.stderr) != 0 {
				t.Fatalf("snapshot Node: %d %s", js.exitCode, js.stderr)
			}
			if difference := firstDifference(string(js.stdout), expected); difference != "" {
				t.Fatalf("snapshot Node %s: %s", mode, difference)
			}
			t.Logf("%s source Node %s: %d exact Go answer bytes", filepath.Base(directory), mode, len(expected))
			for _, name := range []string{"printer", "profiled", "counted"} {
				result := execute(t, nil, filepath.Join(directory, name), cases, "output", "once", mode)
				if result.exitCode != 0 || (name != "counted" && len(result.stderr) != 0) {
					t.Fatalf("%s %s %d %s", directory, name, result.exitCode, result.stderr)
				}
				if difference := firstDifference(string(result.stdout), expected); difference != "" {
					t.Fatalf("%s %s %s: %s", directory, name, mode, difference)
				}
				t.Logf("%s %s %s: %d exact Go answer bytes", filepath.Base(directory), name, mode, len(expected))
			}
		}
	}
}
