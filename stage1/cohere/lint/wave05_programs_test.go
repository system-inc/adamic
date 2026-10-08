package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// This owned check preserves the program that cohere actually asserted. The ordinary capture
// contains just its subject source, which loses module fixtures and compiler options for typed rules.
func TestWave05OriginalProgramsAgree(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	capture := filepath.Join(scratch, "projects")
	if err := os.MkdirAll(capture, 0755); err != nil {
		t.Fatal(err)
	}
	harness := filepath.Join(root, "internal/lint/testing/program.go")
	data, err := os.ReadFile(harness)
	if err != nil {
		t.Fatal(err)
	}
	anchor := "return Result{\n\t\tDiagnostics: diagnostics,\n\t\tSourceFile:  sourceFile,\n\t\tcapture:     newCapturedRun(subject, subjectFileName, len(files)-1, options),\n\t}"
	if strings.Count(string(data), anchor) != 1 {
		t.Fatal("original-program capture anchor changed")
	}
	source := strings.Replace(string(data), "import (", "import (\n\"encoding/json\"", 1)
	source = strings.Replace(source, anchor, "wave05RecordProgram(t, subject.Name, directory, subjectPath, options, files, verbatim)\n\t"+anchor, 1) + wave05ProgramCapture
	side := filepath.Join(scratch, "program.go")
	if err := os.WriteFile(side, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{harness: side}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(scratch, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	packages := map[string]string{
		"core":       "Test(RequireAwait|SymbolDescription|ValidTypeof|NoUselessAssignment)",
		"typescript": "Test(NoUnnecessaryTypeParameters|RestrictTemplateExpressions)",
		"nexus":      "TestCorrectness(NoProcessExitAfterOutput|NoUnclearedRaceTimeout|RequireBlockingStandardStreams)",
	}
	for _, name := range []string{"core", "typescript", "nexus"} {
		if _, err := wave05CaptureRun(root, scratch, []string{"WAVE05_ORIGINAL_PROGRAMS=" + capture}, "go", "test", "-overlay="+overlayPath, "./internal/lint/rules/"+name, "-run", "^("+packages[name]+")", "-count=1", "-timeout=10m"); err != nil {
			t.Fatal(err)
		}
	}
	paths, err := filepath.Glob(filepath.Join(capture, "*", "wave05-case.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no original typed programs captured")
	}
	sort.Strings(paths)
	oracle := goOracle(t)
	binary := buildPort(t, directory, true)
	module := emittedJavaScript(t, directory)
	counts := map[string]int{}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var item struct {
			Rule, Subject string
			Options       json.RawMessage
			Files         int
		}
		if err := json.Unmarshal(data, &item); err != nil {
			t.Fatal(err)
		}
		counts[item.Rule]++
		project := filepath.Dir(path)
		t.Run(fmt.Sprintf("%s/%d", strings.ReplaceAll(item.Rule, "/", "-"), counts[item.Rule]), func(t *testing.T) {
			config := filepath.Join(project, "tsconfig.json")
			if _, err := os.Stat(config); err != nil {
				t.Fatal(err)
			}
			subject := filepath.Join(project, item.Subject)
			row := subject + "\t" + item.Rule + "\t\t\tfalse\t" + string(item.Options)
			compareWithJavaScript(t, oracle, binary, directory, manifest(t, []string{"program " + config, row}), module)
		})
	}
	t.Logf("original complete programs: %d; cases per rule: %v", len(paths), counts)
}

const wave05ProgramCapture = `
func wave05RecordProgram(t *testing.T, ruleName, directory, subjectPath string, options any, fixture map[string]string, verbatim bool) {
 t.Helper()
 destination := os.Getenv("WAVE05_ORIGINAL_PROGRAMS")
 if destination == "" { t.Fatal("original-program capture destination missing") }
 project, err := os.MkdirTemp(destination,"program-"); if err != nil { t.Fatal(err) }
 files := 0
 if strings.HasPrefix(directory, memoryFixtureRoot+"/") {
  for name,contents := range fixture {
   if !verbatim { contents = FixtureText(contents) }
   path := filepath.Join(project,name)
   if err := os.MkdirAll(filepath.Dir(path),0755); err != nil { t.Fatal(err) }
   if err := os.WriteFile(path,[]byte(contents),0644); err != nil { t.Fatal(err) }
   files++
  }
  if err := os.WriteFile(filepath.Join(project,"tsconfig.json"),[]byte(defaultTsConfig),0644); err != nil { t.Fatal(err) }
  files++
 } else {
 err = filepath.Walk(directory,func(path string,info os.FileInfo,walkErr error) error {
  if walkErr != nil { return walkErr }
  relative,err := filepath.Rel(directory,path); if err != nil { return err }
  copied := filepath.Join(project,relative)
  if info.IsDir() { return os.MkdirAll(copied,0755) }
  data,err := os.ReadFile(path); if err != nil { return err }
  if filepath.Base(path)=="tsconfig.json" { data = []byte(strings.ReplaceAll(string(data),directory,project)) }
  files++
  return os.WriteFile(copied,data,0644)
 }); if err != nil { t.Fatal(err) }
 }
 relative,err := filepath.Rel(directory,subjectPath); if err != nil { t.Fatal(err) }
 decoded,err := json.Marshal(options); if err != nil { t.Fatal(err) }
 decoded = []byte(strings.ReplaceAll(string(decoded),directory,project))
 data,err := json.Marshal(struct{Rule,Subject string; Options json.RawMessage; Files int}{ruleName,relative,decoded,files}); if err != nil { t.Fatal(err) }
 if err := os.WriteFile(filepath.Join(project,"wave05-case.json"),data,0644); err != nil { t.Fatal(err) }
}
`

// Not parallel: each upstream process owns its capture destination and finishes before replay.
func wave05CaptureRun(directory, scratch string, environment []string, name string, arguments ...string) ([]byte, error) {
	context, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	command := exec.CommandContext(context, name, arguments...)
	command.Dir = directory
	command.Env = append(os.Environ(), environment...)
	command.Env = append(command.Env, "PWD="+directory)
	output, err := os.CreateTemp(scratch, "upstream-")
	if err != nil {
		return nil, err
	}
	defer output.Close()
	command.Stdout = output
	command.Stderr = output
	executionError := command.Run()
	data, err := os.ReadFile(output.Name())
	if err != nil {
		return nil, err
	}
	if executionError != nil {
		return nil, fmt.Errorf("%s %v: %v\n%s", name, arguments, executionError, data)
	}
	return data, nil
}

// The original-program hook sees typed runs. These core corpora also contain
// source-only runs, so replay them separately instead of losing them after an
// unrelated typed parser refusal in the shared aggregate comparison.
func TestWave05SourceOnlyCorporaAgree(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	binary := buildPort(t, directory, true)
	module := emittedJavaScript(t, directory)
	selected := map[string]bool{
		"require-await":      true,
		"symbol-description": true,
		"valid-typeof":       true,
		"@typescript-eslint/adjacent-overload-signatures": true,
	}
	counts := map[string]int{}
	for _, row := range upstream(t) {
		fields := strings.Split(row, "\t")
		if len(fields) < 2 || !selected[fields[1]] {
			continue
		}
		name := fields[1]
		counts[name]++
		t.Run(fmt.Sprintf("%s/%d", strings.ReplaceAll(name, "/", "-"), counts[name]), func(t *testing.T) {
			if name == "@typescript-eslint/adjacent-overload-signatures" {
				compareWithJavaScript(t, oracle, binary, directory, manifest(t, []string{row}), module)
				return
			}
			config := filepath.Join(t.TempDir(), "tsconfig.json")
			options := fmt.Sprintf(`{"compilerOptions":{"strict":true},"files":[%q]}`, fields[0])
			if err := os.WriteFile(config, []byte(options), 0644); err != nil {
				t.Fatal(err)
			}
			compareWithJavaScript(t, oracle, binary, directory, manifest(t, []string{"program " + config, row}), module)
		})
	}
	for name := range selected {
		if counts[name] == 0 {
			t.Fatalf("source corpus missing for %s", name)
		}
	}
	t.Logf("complete source-only corpus counts: %v", counts)
}
