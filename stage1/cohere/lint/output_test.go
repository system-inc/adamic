package lint

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

type outputAnswer struct {
	Mode          string
	Color, Phases bool
	Output        string
	Exit          int
}
type outputCase struct {
	Source  string
	Driver  string
	Answers []outputAnswer
}

func outputCases(t *testing.T, rows ...[]string) []outputCase {
	return outputCasesKind(t, false, rows...)
}

func outputCasesKind(t *testing.T, auxiliary bool, rows ...[]string) []outputCase {
	t.Helper()
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs("testdata/output_oracle_test.go")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	replacements := map[string]string{filepath.Join(root, "command/cohere/adamic_output_oracle_test.go"): side}
	testName := "^TestAdamicOutputCapture$"
	if auxiliary {
		testName = "^TestAdamicAuxiliaryCapture$"
		module, err := filepath.Abs("output")
		if err != nil {
			t.Fatal(err)
		}
		t.Setenv("ADAMIC_OUTPUT_MODULE", module)
	}
	if len(rows) > 0 {
		corpusSide, err := filepath.Abs("testdata/corpus_output_oracle_test.go")
		if err != nil {
			t.Fatal(err)
		}
		replacements[filepath.Join(root, "command/cohere/adamic_corpus_output_oracle_test.go")] = corpusSide
		data, err := os.ReadFile("testdata/oracle.go")
		if err != nil {
			t.Fatal(err)
		}
		source := strings.ReplaceAll(string(data), "func main()", "func adamicRulesMain()")
		source = strings.ReplaceAll(source, "run(", "adamicRulesRun(")
		rules := filepath.Join(directory, "rules_test.go")
		if err := os.WriteFile(rules, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
		replacements[filepath.Join(root, "command/cohere/adamic_rules_oracle_test.go")] = rules
		t.Setenv("ADAMIC_OUTPUT_MANIFEST", manifest(t, rows[0]))
		project, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		module, err := filepath.Abs("output")
		if err != nil {
			t.Fatal(err)
		}
		t.Setenv("ADAMIC_OUTPUT_ROOT", project)
		t.Setenv("ADAMIC_OUTPUT_MODULE", module)
		testName = "^TestAdamicCorpusOutputCapture$"
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": replacements})
	path := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(path, overlay, 0o644); err != nil {
		t.Fatal(err)
	}
	// Not parallel: this capture's environment is inherited by its Go subprocess.
	t.Setenv("ADAMIC_OUTPUT_CAPTURE", directory)
	captureOutputGo(t, root, directory, "test", "-overlay="+path, "./command/cohere", "-run", testName, "-count=1", "-timeout=10m")
	data, err := os.ReadFile(filepath.Join(directory, "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []outputCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	return cases
}

func captureOutputGo(t *testing.T, root, directory string, arguments ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", arguments...)
	command.Dir = root
	path := filepath.Join(directory, "capture.log")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	command.Stdout, command.Stderr = file, file
	err = command.Run()
	file.Close()
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if err != nil {
		t.Fatalf("Go output capture: %v\n%s", err, data)
	}
	t.Logf("Go output capture:\n%s", data)
}

func compareOutputCases(t *testing.T, cases []outputCase) {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for index, one := range cases {
		t.Run(string(rune('A'+index)), func(t *testing.T) {
			arguments := [][]string{}
			var expected strings.Builder
			status := one.Answers[0].Exit
			for _, answer := range one.Answers {
				color, phases := "false", "false"
				if answer.Color {
					color = "true"
				}
				if answer.Phases {
					phases = "true"
				}
				if answer.Exit != status {
					t.Fatal("formats disagree on exit status")
				}
				arguments = append(arguments, []string{answer.Mode, color, phases})
				expected.WriteString(answer.Output)
			}
			encoded, _ := json.Marshal(arguments)
			path, binary := outputProgram(t, one.Source, "", string(encoded), one.Driver)
			for _, command := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", runner, path}} {
				got, code := outputRun(t, command[0], command[1:]...)
				if code != status || got != expected.String() {
					t.Fatalf("%s: exit %d want %d\n%s", command[0], code, status, difference([]byte(got), []byte(expected.String())))
				}
			}
		})
	}
}

// Not parallel: the reference capture inherits its manifest environment.
func TestOutputLiveRulesAndFixes(t *testing.T) {
	compareOutputCases(t, outputCases(t, generated(t)))
}

// Not parallel: large corpus observations must run with the same manifest and captured clocks.
func TestOutputWholeCorpus(t *testing.T) {
	compiler := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if compiler == "" {
		t.Skip("set ADAMIC_TYPESCRIPT_SOURCE to the pinned compiler checkout")
	}
	pin := execute(t, "", "git", "-C", compiler, "rev-parse", "HEAD")
	if strings.TrimSpace(string(pin.output)) != compilerCommit {
		t.Fatal("compiler corpus pin differs")
	}
	rows := []string{}
	for _, root := range []string{filepath.Join(compiler, "src/compiler"), repository} {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "cohere" && path != filepath.Join(repository, "stage1/cohere")) {
				return filepath.SkipDir
			}
			if !entry.IsDir() && (strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".a")) {
				absolute, err := filepath.Abs(path)
				if err != nil {
					return err
				}
				rows = append(rows, absolute)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("TypeScript compiler and repository: %d source files", len(rows))
	compareOutputCases(t, outputCases(t, rows))
}

func outputProgram(t *testing.T, source string, modules ...string) (string, string) {
	t.Helper()
	module, err := filepath.Abs("output")
	if len(modules) > 0 && modules[0] != "" {
		module = modules[0]
	}
	if err != nil {
		t.Fatal(err)
	}
	quote := func(value string) string { data, _ := json.Marshal(value); return string(data) }
	header := "import { programArguments } from 'adamic';\nimport type { RunSummary, RunFinding } from " + quote(filepath.Join(module, "model.ts")) + ";\nimport { exitCode } from " + quote(filepath.Join(module, "model.ts")) + ";\nimport { changedFileLines, changedFilesJSON, renderFindings, renderSummary } from " + quote(filepath.Join(module, "render.ts")) + ";\n"
	driver := `const args = programArguments();
const mode = args[0] ?? 'human';
const color = args[1] === 'true';
const phases = args[2] === 'true';
if (mode === 'human') { for (const line of changedFileLines(summary.changed, color)) { console.log(line); } }
if (mode === 'json') { for (const line of changedFilesJSON(summary.changed)) { console.log(line); } }
for (const line of renderFindings(findings, mode, color)) { console.log(line); }
console.log(renderSummary(summary, mode, color, phases));
process.exitCode = exitCode(summary);
`
	if len(modules) > 2 && modules[2] != "" {
		driver = modules[2]
	}
	if len(modules) > 1 {
		driver = "for (const args of " + modules[1] + ") {\n" + strings.TrimPrefix(driver, "const args = programArguments();\n") + "}\n"
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "main.ts")
	if err := os.WriteFile(path, []byte(header+source+driver), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "program")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	return path, binary
}

func TestOutputRendererMutants(t *testing.T) {
	cases := outputCases(t)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []struct {
		name, file, from, to, mode string
		index                      int
	}{
		{"JSON HTML escapes", "render.ts", `.replaceAll('<', '\\u003c')`, ``, "json", 0},
		{"UTF-8 ordering", "model.ts", "const x = left.codePointAt(a) ?? 0;\n        const y = right.codePointAt(b) ?? 0;", "const x = left.charCodeAt(a);\n        const y = right.charCodeAt(b);", "human", 0},
		{"Go midpoint rounding", "footer.ts", "&& lower % 2 === 0", "&& lower % 2 === -1", "human", 3},
	} {
		t.Run(change.name, func(t *testing.T) {
			directory := t.TempDir()
			for _, file := range []string{"model.ts", "render.ts", "footer.ts"} {
				data, err := os.ReadFile(filepath.Join("output", file))
				if err != nil {
					t.Fatal(err)
				}
				source := string(data)
				if file == change.file {
					if strings.Count(source, change.from) != 1 {
						t.Fatal("mutant anchor changed")
					}
					source = strings.Replace(source, change.from, change.to, 1)
				}
				if err := os.WriteFile(filepath.Join(directory, file), []byte(source), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			one := cases[change.index]
			path, binary := outputProgram(t, one.Source, directory)
			for _, answer := range one.Answers {
				if answer.Mode != change.mode || answer.Color || answer.Phases {
					continue
				}
				got, status := outputRun(t, binary, answer.Mode, "false", "false")
				node, nodeStatus := outputRun(t, "node", "--disable-warning=ExperimentalWarning", runner, path, answer.Mode, "false", "false")
				if status != answer.Exit || nodeStatus != status || node != got {
					t.Fatalf("mutant must run cleanly and agree on Node/native: statuses %d/%d, want %d", status, nodeStatus, answer.Exit)
				}
				if got == answer.Output {
					t.Fatal("mutant survived Go comparison")
				}
				t.Log("compiled and ran cleanly on Node and native; caught only by Go output comparison")
			}
		})
	}
}

func outputRun(t *testing.T, name string, arguments ...string) (string, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, name, arguments...)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1")
	file, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	command.Stdout = file
	var stderr bytes.Buffer
	command.Stderr = &stderr
	err = command.Run()
	if ctx.Err() != nil {
		t.Fatalf("output driver timed out: %v", ctx.Err())
	}
	var exited *exec.ExitError
	if err != nil && !errors.As(err, &exited) {
		t.Fatal(err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr: %s", &stderr)
	}
	data, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(data), command.ProcessState.ExitCode()
}

func TestOutputRenderersAgree(t *testing.T) {
	cases := outputCases(t)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	comparisons := 0
	for index, one := range cases {
		t.Run(string(rune('A'+index)), func(t *testing.T) {
			path, binary := outputProgram(t, one.Source)
			for _, answer := range one.Answers {
				color, phases := "false", "false"
				if answer.Color {
					color = "true"
				}
				if answer.Phases {
					phases = "true"
				}
				for _, command := range [][]string{{binary, answer.Mode, color, phases}, {"node", "--disable-warning=ExperimentalWarning", runner, path, answer.Mode, color, phases}} {
					got, status := outputRun(t, command[0], command[1:]...)
					if status != answer.Exit || got != answer.Output {
						t.Fatalf("%s %s color=%t phases=%t: exit %d want %d\n%s", command[0], answer.Mode, answer.Color, answer.Phases, status, answer.Exit, difference([]byte(got), []byte(answer.Output)))
					}
					comparisons++
				}
			}
		})
	}
	if !t.Failed() {
		t.Logf("%d Node/native observations held byte for byte to unmodified Go cohere renderers", comparisons)
	}
}

// Not parallel: the Go capture inherits its module path.
func TestOutputAuxiliaryRenderers(t *testing.T) {
	compareOutputCases(t, outputCasesKind(t, true))
}
