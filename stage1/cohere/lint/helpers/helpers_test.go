package helpers

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/stage1/cohere/lint/helpers/oracles"

	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func run(test *testing.T, dir, name string, args ...string) []byte {
	test.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if !strings.ContainsRune(name, filepath.Separator) {
		// A bare name (node) is a tool: the one internal/toolchain resolves, with a tool's environment, not PATH's, run
		// in dir with PWD to match. A program a test built (the Go oracle, the native build) runs as named, with this
		// process's environment.
		cmd = toolchain.CommandIn(dir, name, args...)
	}
	f, temporaryFileError := os.CreateTemp(test.TempDir(), "output-")
	if temporaryFileError != nil {
		test.Fatal(temporaryFileError)
	}
	defer f.Close()
	cmd.Stdout = f
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if runError := childguard.Run(cmd, childguard.Options{}); runError != nil {
		test.Fatalf("%s %v: %v\n%s", name, args, runError, &stderr)
	}
	if stderr.Len() != 0 {
		test.Fatalf("%s stderr: %s", name, &stderr)
	}
	data, readError := os.ReadFile(f.Name())
	if readError != nil {
		test.Fatal(readError)
	}
	return data
}
func oracle(test *testing.T) string {
	test.Helper()
	return oracles.Helpers(test)
}
func TestProduct_HelpersFindsGoOracle(test *testing.T) {
	test.Parallel()
	oracle(test)
}
func build(test *testing.T, directory string) string {
	test.Helper()
	program, loadError := load.Load([]string{filepath.Join(directory, "main.ts")})
	if loadError != nil {
		test.Fatal(loadError)
	}
	ir, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	binary := filepath.Join(test.TempDir(), "helpers")
	if buildError := native.Build(native.C(ir), binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	return binary
}

func TestHelpersMatchCohere(test *testing.T) {
	test.Parallel()
	goOracle := oracle(test)
	catalog := run(test, "", goOracle, "catalog")
	descriptor := run(test, "", goOracle, "descriptors")
	pinned, readError := os.ReadFile("testdata/descriptors.json")
	if readError != nil || !bytes.Equal(descriptor, pinned) {
		test.Fatal("regenerate target descriptors against Go cohere")
	}
	current, readError := os.ReadFile("testdata/catalog.json")
	if readError != nil {
		test.Fatal(readError)
	}
	if !bytes.Equal(catalog, current) {
		test.Fatal("regenerate catalog against pinned Go cohere")
	}
	cases := fixture(test)
	catalogPath, _ := filepath.Abs("testdata/catalog.json")
	want := run(test, "", goOracle, cases)
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.ts")
	node := run(test, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases, catalogPath)
	compare(test, node, want)
	compare(test, run(test, "", helperMutantProduct(test, helperClean), cases, catalogPath), want)
	test.Logf(
		"%d oracle output lines matched Go, Node and sanitized native",
		len(strings.Split(strings.TrimSpace(string(want)), "\n")),
	)
}
func compare(test *testing.T, got, want []byte) {
	test.Helper()
	if bytes.Equal(got, want) {
		return
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			test.Fatalf("output line %d: got %q, Go %q", i+1, a[i], b[i])
		}
	}
	test.Fatalf("output size: got %d Go %d", len(got), len(want))
}

func fixture(test *testing.T) string {
	test.Helper()
	f, openError := os.Open("testdata/cases.json.gz")
	if openError != nil {
		test.Fatal(openError)
	}
	defer f.Close()
	reader, gzipError := gzip.NewReader(f)
	if gzipError != nil {
		test.Fatal(gzipError)
	}
	defer reader.Close()
	path := filepath.Join(test.TempDir(), "cases.json")
	out, createError := os.Create(path)
	if createError != nil {
		test.Fatal(createError)
	}
	if _, copyError := io.Copy(out, reader); copyError != nil {
		test.Fatal(copyError)
	}
	if closeError := out.Close(); closeError != nil {
		test.Fatal(closeError)
	}
	return path
}

func smallFixture(test *testing.T, corpus any) string {
	test.Helper()
	data, encodeError := json.Marshal(corpus)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	path := filepath.Join(test.TempDir(), "cases.json")
	if writeError := os.WriteFile(path, data, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	return path
}

func TestMessageRefusalsMatchGo(test *testing.T) {
	test.Parallel()
	binary := helperMutantProduct(test, helperClean)
	goOracle := oracle(test)
	catalog, _ := filepath.Abs("testdata/catalog.json")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.ts")
	cases := []map[string]any{
		{"Kind": "message", "Rule": "missing/rule", "Id": "missing", "Values": map[string]string{}},
		{"Kind": "message", "Rule": "nexus/consistency-no-for-in", "Id": "missing", "Values": map[string]string{}},
		{
			"Kind":   "message",
			"Rule":   "nexus/consistency-no-stuttering-name",
			"Id":     "stutteringName",
			"Values": map[string]string{},
		},
		{
			"Kind":   "message",
			"Rule":   "nexus/consistency-no-stuttering-name",
			"Id":     "stutteringName",
			"Values": map[string]string{"other": "x"},
		},
		{
			"Kind":   "message",
			"Rule":   "nexus/consistency-no-for-in",
			"Id":     "forIn",
			"Values": map[string]string{"extra": "x"},
		},
		{
			"Kind":   "message",
			"Rule":   "nexus/consistency-no-for-in",
			"Id":     "forIn",
			"Values": map[string]string{},
			"Choices": []map[string]string{
				{"Rule": "nexus/consistency-no-for-in", "Id": "forIn", "Phrase": "reason", "Name": "missing"},
			},
		},
	}
	phraseRule := "structure/consistency-require-matching-file-name"
	phraseId := "requireMatchingFileName"
	choice := map[string]string{"Rule": phraseRule, "Id": phraseId, "Phrase": "alternative", "Name": "none"}
	for _, choices := range [][]map[string]string{
		nil,
		{choice, choice},
		{{"Rule": "wrong/rule", "Id": phraseId, "Phrase": "alternative", "Name": "none"}},
		{{"Rule": phraseRule, "Id": "wrong", "Phrase": "alternative", "Name": "none"}},
	} {
		cases = append(cases, map[string]any{
			"Kind":    "message",
			"Rule":    phraseRule,
			"Id":      phraseId,
			"Values":  map[string]string{},
			"Choices": choices,
		})
	}
	for i, c := range cases {
		path := smallFixture(test, map[string]any{"Definitions": map[string]string{}, "Cases": []any{c}})
		want := strings.TrimSpace(string(run(test, "", goOracle, path)))
		if !strings.HasPrefix(want, "panic: ") {
			test.Fatalf("guard %d: Go did not refuse", i)
		}
		expected := "adamic: " + want + "\n"
		for _, command := range [][]string{
			{binary, path, catalog},
			{"node", "--disable-warning=ExperimentalWarning", runner, entry, path, catalog},
		} {
			// Refusals are silent until their final panic; protect startup with FirstOutput.
			cmd := exec.Command(command[0], command[1:]...)
			if !strings.ContainsRune(
				command[0],
				filepath.Separator,
			) { // node, through internal/toolchain; the native build runs as named
				cmd = toolchain.Command(command[0], command[1:]...)
			}
			output, temporaryFileError := os.CreateTemp(test.TempDir(), "stdout-")
			if temporaryFileError != nil {
				test.Fatal(temporaryFileError)
			}
			cmd.Stdout = output
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			runError := childguard.Run(cmd, childguard.Options{})
			output.Close()
			exit, ok := runError.(*exec.ExitError)
			if !ok || exit.ExitCode() != 70 || stderr.String() != expected {
				test.Fatalf("guard %d %s: %v stderr %q, Go %q", i, command[0], runError, stderr.String(), expected)
			}
		}
	}
	test.Logf("%d message refusal cases match Go on Node and sanitized native", len(cases))
}

func TestKnownGapsAreExplicit(test *testing.T) {
	test.Parallel()
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
	path := smallFixture(test, corpus)
	catalog, _ := filepath.Abs("testdata/catalog.json")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.ts")
	want := []byte(
		"NotYet: unsupported schema keyword pattern\nNotYet: custom or unsupported Go option type\nNotYet: Unicode fold for untagged option keys\nvalid\n",
	)
	compare(test, run(test, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path, catalog), want)
	compare(test, run(test, "", helperMutantProduct(test, helperClean), path, catalog), want)
}
