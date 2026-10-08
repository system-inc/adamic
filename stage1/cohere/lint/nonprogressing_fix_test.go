package lint

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func nonprogressingTypeDefinitionsPort(t *testing.T) string {
	t.Helper()
	directory := mutant(t, "", "")
	source := filepath.Join("testdata", "nonprogressing-type-definitions")
	destination := filepath.Join(directory, "rules", "typescript-eslint-consistent-type-definitions")
	err := filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if filepath.Ext(relative) == ".a" {
			data = []byte(rewritePortImports(t, filepath.Join("rules", "typescript-eslint-consistent-type-definitions", relative), string(data)))
		}
		return os.WriteFile(target, data, 0644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return directory
}

func nonprogressingTypeDefinitionsManifest(t *testing.T) string {
	t.Helper()
	source := filepath.Join(t.TempDir(), "Shape.ts")
	if err := os.WriteFile(source, []byte("type Shape = { value: string }\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return manifest(t, []string{source + "\t@typescript-eslint/consistent-type-definitions\t\t\tfalse\t\"interface\""})
}

func TestNonprogressingFix(t *testing.T) {
	directory := nonprogressingTypeDefinitionsPort(t)
	path := nonprogressingTypeDefinitionsManifest(t)
	want := compare(t, goOracleFrom(t, directory), buildPort(t, directory, true), directory, path)
	for _, field := range []string{
		"fix-edit\t30 30\t\n",
		"rejected @typescript-eslint/consistent-type-definitions 30 30  the fix replaces text with itself\n",
		"fixed\tinterface Shape { value: string }\\u000a\n",
	} {
		if !bytes.Contains(want, []byte(field)) {
			t.Fatalf("missing %q: %s", field, want)
		}
	}
	t.Log("no-progress sibling refused; the other automatic edits apply byte-identically on all runtimes")
}

// This mutant remains a valid program. Only executing the regression exposes the restored panic.
func TestNonprogressingFixPanicMutant(t *testing.T) {
	directory := nonprogressingTypeDefinitionsPort(t)
	path := nonprogressingTypeDefinitionsManifest(t)
	before := []byte("                    this.rejected.push(\n" +
		"                        `rejected ${finding.rule} ${utf8Length(current.slice(0, finding.editStart))} ${utf8Length(current.slice(0, finding.editEnd))}  the fix replaces text with itself`,\n" +
		"                    );\n                    continue;")
	module := filepath.Join(directory, "lint.ts")
	data, err := os.ReadFile(module)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Count(data, before) != 1 {
		t.Fatal("no-progress mutant anchor changed")
	}
	if err := os.WriteFile(module, bytes.Replace(data, before, []byte("                    panic('nonprogressing fix');"), 1), 0644); err != nil {
		t.Fatal(err)
	}
	runner := filepath.Join(repository, "oracle", "node.mjs")
	commands := []struct {
		name string
		args []string
	}{
		{"Node", []string{"node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", path}},
		{"emitted JavaScript", []string{"node", "--disable-warning=ExperimentalWarning", runner, emittedJavaScript(t, directory), "--manifest", path}},
		{"sanitized native", []string{buildPort(t, directory, true), "--manifest", path}},
	}
	for _, side := range commands {
		t.Run(side.name, func(t *testing.T) {
			log := filepath.Join(t.TempDir(), "panic-mutant.log")
			output, err := os.Create(log)
			if err != nil {
				t.Fatal(err)
			}
			command := exec.Command(side.args[0], side.args[1:]...)
			command.Stdout, command.Stderr = output, output
			runError := command.Run()
			if err := output.Close(); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			exit, ok := runError.(*exec.ExitError)
			if !ok || exit.ExitCode() != 70 || !bytes.Contains(data, []byte("adamic: panic: nonprogressing fix")) {
				t.Fatalf("panic mutant escaped the regression: %v\n%s", runError, data)
			}
			t.Logf("restored panic caught at runtime on %s: %s", side.name, data)
		})
	}
}

// Independent Go edit planning decides reservation, refusal order and UTF-8 positions.
func TestNonprogressingFixPlanOrder(t *testing.T) {
	directory := mutant(t, "", "")
	for _, name := range []string{"rule.a", "oracle.go"} {
		data, err := os.ReadFile(filepath.Join("testdata", "nonprogressing-plan", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "rules", "no-debugger", name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Remove(filepath.Join(directory, "rules", "no-debugger", "rule.ts")); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "plan.ts")
	if err := os.WriteFile(source, []byte("/*😀*/debugger;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	path := manifest(t, []string{source + "\tno-debugger"})
	answer := compare(t, goOracleFrom(t, directory), buildPort(t, directory, true), directory, path)
	expected := "rejected no-debugger 8 16 no-debugger overlaps another fix\n" +
		"rejected no-debugger 8 9  the fix replaces text with itself\n" +
		"rejected no-debugger 17 17  the fix replaces text with itself\n" +
		"fixed\t/*ok*/debugger;\\u000a\n"
	if !bytes.Contains(answer, []byte(expected)) {
		t.Fatalf("wrong refusal order, reservation or progress: %s", answer)
	}
	t.Log("overlap refusals precede no-progress refusals; rejected survivor reserves its span; independent edit applies")
}
