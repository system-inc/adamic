package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// FirstDifference records exact canonical rows, including diagnostic text and
// byte positions, rather than accepting any failure under a case's name.
type lintCaseDifference struct {
	Line         int    `json:"line"`
	Port         string `json:"port"`
	TypeScriptGo string `json:"typescript_go"`
}
type lintCaseExpectation struct {
	File            string             `json:"file"`
	FoundBy         string             `json:"found_by"`
	FirstDifference lintCaseDifference `json:"first_difference"`
}

func lintCaseFirstDifference(port, oracle []byte) *lintCaseDifference {
	if bytes.Equal(port, oracle) {
		return nil
	}
	a, b := strings.Split(string(port), "\n"), strings.Split(string(oracle), "\n")
	for i := 0; i < len(a) || i < len(b); i++ {
		left, right := "<EOF>", "<EOF>"
		if i < len(a) {
			left = a[i]
		}
		if i < len(b) {
			right = b[i]
		}
		if left != right {
			return &lintCaseDifference{i + 1, left, right}
		}
	}
	return nil
}

func lintCaseCheck(file string, expected *lintCaseExpectation, port, oracle []byte) error {
	diff := lintCaseFirstDifference(port, oracle)
	if expected == nil {
		if diff != nil {
			return fmt.Errorf("unlisted mismatch %s: %+v", file, *diff)
		}
		return nil
	}
	if diff == nil {
		return fmt.Errorf("listed case now matches %s: remove its expected.json entry", file)
	}
	if *diff != expected.FirstDifference {
		return fmt.Errorf("first difference changed %s: got %+v, recorded %+v", file, *diff, expected.FirstDifference)
	}
	return nil
}

// A parser refusal is an observable gap, not a skip. Only its known explicit
// exit-70 protocol may become a diagnostic row. Crashes, timeouts, sanitizer
// errors and other stderr always fail, even for a listed case.
func lintCasePort(t *testing.T, name string, args []string, path string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, name, args...)
	out, err := os.CreateTemp(t.TempDir(), "lint-case-stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	stderr, err := os.CreateTemp(t.TempDir(), "lint-case-stderr-")
	if err != nil {
		t.Fatal(err)
	}
	defer stderr.Close()
	command.Stdout, command.Stderr = out, stderr
	runErr := command.Run()
	got, err := os.ReadFile(out.Name())
	if err != nil {
		t.Fatal(err)
	}
	diagnostic, err := os.ReadFile(stderr.Name())
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Err() != nil {
		t.Fatalf("parser timed out for %s: %v", path, ctx.Err())
	}
	if runErr == nil && len(diagnostic) == 0 {
		return got
	}
	var exited *exec.ExitError
	message := strings.TrimSpace(string(diagnostic))
	prefix := "adamic: panic: parser slice unsupported "
	if errors.As(runErr, &exited) && exited.ExitCode() == 70 && len(got) == 0 && strings.HasPrefix(message, prefix) && !strings.Contains(message, "\n") {
		message = strings.TrimSuffix(message, " in "+path)
		return []byte("diagnostic refusal\t" + strings.TrimPrefix(message, "adamic: panic: ") + "\n")
	}
	t.Fatalf("parser execution %s: %v; stderr %q", path, runErr, diagnostic)
	return nil
}

func lintCaseSides(t *testing.T, directory, binary, path string) []struct {
	name   string
	output []byte
} {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	args := []string{path, "--whole", "--recovery"}
	return []struct {
		name   string
		output []byte
	}{
		{"Node", lintCasePort(t, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts")}, args...), path)},
		{"sanitized native", lintCasePort(t, binary, args, path)},
	}
}

func TestLintCases(t *testing.T) {
	data, err := os.ReadFile("testdata/lint_cases/expected.json")
	if err != nil {
		t.Fatal(err)
	}
	var entries []lintCaseExpectation
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&entries); err != nil {
		t.Fatal(err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		t.Fatalf("expected.json must contain exactly one JSON array: %v", err)
	}
	expected := map[string]lintCaseExpectation{}
	for _, entry := range entries {
		if filepath.Base(entry.File) != entry.File || entry.FoundBy == "" || entry.FirstDifference.Line < 1 {
			t.Fatalf("invalid expectation: %+v", entry)
		}
		if _, exists := expected[entry.File]; exists {
			t.Fatalf("duplicate expectation: %s", entry.File)
		}
		expected[entry.File] = entry
	}
	files, err := os.ReadDir("testdata/lint_cases")
	if err != nil {
		t.Fatal(err)
	}
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	binary := buildPort(t, directory, true)
	count := 0
	for _, file := range files {
		if file.IsDir() || !(strings.HasSuffix(file.Name(), ".ts") || strings.HasSuffix(file.Name(), ".tsx")) {
			continue
		}
		count++
		entry, listed := expected[file.Name()]
		delete(expected, file.Name())
		t.Run(file.Name(), func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join("testdata/lint_cases", file.Name()))
			if err != nil {
				t.Fatal(err)
			}
			want := execute(t, "", oracle, path, "--whole", "--recovery").output
			var expectation *lintCaseExpectation
			if listed {
				expectation = &entry
			}
			for _, side := range lintCaseSides(t, directory, binary, path) {
				if err := lintCaseCheck(file.Name(), expectation, side.output, want); err != nil {
					t.Errorf("%s: %v", side.name, err)
				}
			}
			t.Logf("listed=%t, oracle bytes=%d", listed, len(want))
		})
	}
	if count == 0 {
		t.Fatal("no lint parser cases")
	}
	if len(expected) != 0 {
		t.Fatalf("expectations without case files: %+v", expected)
	}
}

func TestLintCasesControls(t *testing.T) {
	oracle := goOracle(t)
	path := filepath.Join(t.TempDir(), "precedence.ts")
	if err := os.WriteFile(path, []byte("a + b * c;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	want := execute(t, "", oracle, path, "--whole", "--recovery").output
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	healthy := lintCaseSides(t, directory, buildPort(t, directory, true), path)
	for _, side := range healthy {
		if err := lintCaseCheck("precedence.ts", nil, side.output, want); err != nil {
			t.Fatal(err)
		}
	}
	mutant := copyPort(t, "grammar.ts", "return 14;", "return 12;")
	for _, side := range lintCaseSides(t, mutant, buildPort(t, mutant, true), path) {
		err := lintCaseCheck("precedence.ts", nil, side.output, want)
		if err == nil || !strings.Contains(err.Error(), "unlisted mismatch") {
			t.Fatalf("%s unlisted control survived: %v", side.name, err)
		}
		t.Logf("%s healthy precedence mutant rejected: %v", side.name, err)
		// Listing this actual mutant's first difference accepts only that gap.
		// Switching back to the healthy parser is a real matching observation,
		// which must flip the listed case red. This also works with no open gaps.
		entry := lintCaseExpectation{File: "precedence.ts", FoundBy: "planted control", FirstDifference: *lintCaseFirstDifference(side.output, want)}
		if err := lintCaseCheck(entry.File, &entry, side.output, want); err != nil {
			t.Fatal(err)
		}
		for _, repaired := range healthy {
			if repaired.name != side.name {
				continue
			}
			err := lintCaseCheck(entry.File, &entry, repaired.output, want)
			if err == nil || !strings.Contains(err.Error(), "listed case now matches") {
				t.Fatalf("%s listed-match control survived: %v", repaired.name, err)
			}
			t.Logf("%s listed gap repaired by healthy parser rejected: %v", repaired.name, err)
		}
	}
}
