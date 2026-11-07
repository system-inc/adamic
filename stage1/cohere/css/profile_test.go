package css

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/css/oracles"
)

// Opt-in immutable artifacts let profiling and comparison share the same source.
// Not parallel: writes artifacts and binaries in the fixed ADAMIC_CSS_PROFILE_DIR directory.
func TestCSSProfileArtifacts(test *testing.T) {
	directory := os.Getenv("ADAMIC_CSS_PROFILE_DIR")
	if directory == "" {
		test.Skip("set ADAMIC_CSS_PROFILE_DIR")
	}
	if createDirectoryError := os.MkdirAll(directory, 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	cases, _ := askedCases(test)
	expected := printerAnswers(test, cases, "default")
	narrow := printerAnswers(test, cases, "narrow")
	script, _ := filepath.Abs("testdata/print_library.mjs")
	// Reads: internal/native/runtime/*.c
	repo, _ := filepath.Abs(repository)
	original := execute(
		test,
		nil,
		"node",
		script,
		filepath.Join(repo, "cohere/internal/format/prettier/bundles"),
		cases,
		"default",
		"fork",
	)
	if original.exitCode != 0 || len(original.stderr) != 0 {
		test.Fatal(string(original.stderr))
	}
	data, _ := os.ReadFile(cases)
	inputs := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	goLines := strings.Split(strings.TrimSuffix(expected, "\n"), "\n")
	jsLines := strings.Split(strings.TrimSuffix(string(original.stdout), "\n"), "\n")
	var shared []string
	var units []int
	var manifest strings.Builder
	largest, largestSize := 0, 0
	for index, input := range inputs {
		var a, b string
		if json.Unmarshal([]byte(goLines[index*2+1]), &a) == nil &&
			json.Unmarshal([]byte(jsLines[index*2+1]), &b) == nil &&
			a == b {
			if len(input) > largestSize {
				largest, largestSize = len(shared), len(input)
			}
			shared = append(shared, input)
			units = append(units, len(utf16.Encode([]rune(a))))
		}
	}
	var sample []string
	totalUnits := 0
	for index, input := range shared {
		if index < 32 || index%16 == 0 || index == largest {
			sample = append(sample, input)
			totalUnits += units[index]
			fmt.Fprintf(&manifest, "%d\t%d\t%d\n", index, len(input), units[index])
		}
	}
	for name, data := range map[string][]byte{
		"cases.txt":           data,
		"expected.txt":        []byte(expected),
		"narrow-expected.txt": []byte(narrow),
		"shared.txt":          []byte(strings.Join(shared, "\n") + "\n"),
		"sample.txt":          []byte(strings.Join(sample, "\n") + "\n"),
		"manifest.tsv":        []byte(manifest.String()),
		"sample-expected.txt": []byte(
			fmt.Sprintf("%d of %d stylesheets formatted, %d units\n", len(sample), len(sample), totalUnits),
		),
	} {
		if writeError := os.WriteFile(filepath.Join(directory, name), data, 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	copied := portDirectory(test, nil)
	snapshot := filepath.Join(directory, "source")
	if createDirectoryError := os.MkdirAll(snapshot, 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	parent := filepath.Dir(copied)
	for _, slice := range []string{"css", "selector", "values", "mediaquery", "cssstrings", "cssnumbers"} {
		entries, readDirectoryError := os.ReadDir(filepath.Join(parent, slice))
		if readDirectoryError != nil {
			test.Fatal(readDirectoryError)
		}
		if createDirectoryError := os.MkdirAll(filepath.Join(snapshot, slice), 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		for _, entry := range entries {
			data, readError := os.ReadFile(filepath.Join(parent, slice, entry.Name()))
			if readError != nil {
				test.Fatal(readError)
			}
			if writeError := os.WriteFile(filepath.Join(snapshot, slice, entry.Name()), data, 0644); writeError != nil {
				test.Fatal(writeError)
			}
		}
	}
	program := lowered(test, filepath.Join(snapshot, "css/print_main.ts"))
	code := native.C(program)
	source := filepath.Join(directory, "main.c")
	if writeError := os.WriteFile(source, []byte(code), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	for _, build := range []struct {
		name    string
		options native.Options
	}{{"printer", native.Options{}}, {"counted", native.Options{Count: true}}} {
		if buildError := native.Build(code, filepath.Join(directory, build.name), build.options); buildError != nil {
			test.Fatal(buildError)
		}
	}
	runtimeDirectory := filepath.Join(repo, "internal/native/runtime")
	files, _ := filepath.Glob(filepath.Join(runtimeDirectory, "*.c"))
	// The runtime's units are compiled with the program's features as flags, as main.c's #defines say.
	args := append(
		native.SourceFlags(code, native.Options{}),
		"-g",
		"-I",
		runtimeDirectory,
		"-o",
		filepath.Join(directory, "profiled"),
		source,
	)
	args = append(args, files...)
	args = append(args, "-lm")
	result := execute(test, nil, "clang", args...)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		test.Fatalf("profile build %d %s", result.exitCode, result.stderr)
	}

	request, _ := json.Marshal(
		map[string]any{"Cases": filepath.Join(directory, "shared.txt"), "Repeat": "once", "Rounds": 1},
	)
	if writeError := os.WriteFile(filepath.Join(directory, "go-request.json"), request, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	side, _ := filepath.Abs("testdata/print_side_test.go")
	overlay, _ := json.Marshal(map[string]any{
		"Replace": map[string]string{filepath.Join(repo, "cohere/internal/format/css/adamic_print_side_test.go"): side},
	})
	overlayPath := filepath.Join(directory, "go-overlay.json")
	if writeError := os.WriteFile(overlayPath, overlay, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	command := bounded(
		test,
		"go",
		"test",
		"-c",
		"-overlay="+overlayPath,
		"-o",
		filepath.Join(directory, "go-printer"),
		"./internal/format/css",
	)
	command.Dir = filepath.Join(repo, "cohere")
	if output, combinedOutputError := childguard.CombinedOutput(
		command,
		childguard.Options{Stall: oracles.ChildStallLimit},
	); combinedOutputError != nil {
		test.Fatalf("Go benchmark build: %v %s", combinedOutputError, output)
	}
	test.Logf("saved %d shared inputs and %d profile inputs, %d output units", len(shared), len(sample), totalUnits)
}

// Not parallel: askedCases writes the shared ADAMIC_CSS_KEEP and ADAMIC_CSS_KEEP_RAW paths when configured.
func TestCSSProfileSnapshotsAgree(test *testing.T) {
	asked := os.Getenv("ADAMIC_CSS_PROFILE_SNAPSHOTS")
	if asked == "" {
		test.Skip("set ADAMIC_CSS_PROFILE_SNAPSHOTS")
	}
	cases, _ := askedCases(test)
	for _, mode := range []string{"default", "narrow"} {
		expected := printerAnswers(test, cases, mode)
		for _, directory := range filepath.SplitList(asked) {
			js := onNode(test, filepath.Join(directory, "source/css/print_main.ts"), cases, "output", "once", mode)
			if js.exitCode != 0 || len(js.stderr) != 0 {
				test.Fatalf("snapshot Node: %d %s", js.exitCode, js.stderr)
			}
			if difference := firstDifference(string(js.stdout), expected); difference != "" {
				test.Fatalf("snapshot Node %s: %s", mode, difference)
			}
			test.Logf("%s source Node %s: %d exact Go answer bytes", filepath.Base(directory), mode, len(expected))
			for _, name := range []string{"printer", "profiled", "counted"} {
				result := execute(test, nil, filepath.Join(directory, name), cases, "output", "once", mode)
				if result.exitCode != 0 || (name != "counted" && len(result.stderr) != 0) {
					test.Fatalf("%s %s %d %s", directory, name, result.exitCode, result.stderr)
				}
				if difference := firstDifference(string(result.stdout), expected); difference != "" {
					test.Fatalf("%s %s %s: %s", directory, name, mode, difference)
				}
				test.Logf("%s %s %s: %d exact Go answer bytes", filepath.Base(directory), name, mode, len(expected))
			}
		}
	}
}
