package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"missing_message", "empty_message", "empty_name", "empty_name_and_message", "named_message"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			"internal/oracle/testdata/step21_error_" + name + ".a", true, false,
		})
	}
}

// Node's actual Error.prototype.toString decides the name/message composition;
// the compiler's owned Uncaught prefix and single diagnostic line are explicit.
func checkUncaughtErrorText(t *testing.T, name, text string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/step21_error_"+name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(source), "throw error;") != 1 {
		t.Fatal("missing Node observation seam")
	}
	observation := strings.Replace(string(source), "throw error;", "console.log(Error.prototype.toString.call(error));", 1)
	truth := execute(t, "node", "--input-type=module", "-e", observation)
	if truth.exitCode != 0 || string(truth.stdout) != text+"\n" || len(truth.stderr) != 0 {
		t.Fatalf("Node Error.prototype.toString: %+v, want %q", truth, text)
	}
	step21Uncaught(t, path, "Uncaught "+text+"\n")
}

func TestUncaughtErrorMissingMessage(t *testing.T) {
	t.Parallel()
	checkUncaughtErrorText(t, "missing_message", "Error")
}
func TestUncaughtErrorEmptyMessage(t *testing.T) {
	t.Parallel()
	checkUncaughtErrorText(t, "empty_message", "Error")
}
func TestUncaughtErrorEmptyName(t *testing.T) {
	t.Parallel()
	checkUncaughtErrorText(t, "empty_name", "message")
}
func TestUncaughtErrorEmptyNameAndMessage(t *testing.T) {
	t.Parallel()
	checkUncaughtErrorText(t, "empty_name_and_message", "")
}
func TestUncaughtErrorNamedMessage(t *testing.T) {
	t.Parallel()
	checkUncaughtErrorText(t, "named_message", "Failure: message")
}

func TestUncaughtErrorSeparatorMutantNative(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/step21_error_empty_message.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/exceptions.c"))
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(string(data), "static void uncaught_text(")
	if start < 0 {
		t.Fatal("missing uncaught renderer")
	}
	renderer := string(data[start:])
	const guard = `if (adamic_string_length(adamic_object_field(error, "name", &name_cache)->reference) != 0 && adamic_string_length(adamic_object_field(error, "message", &message_cache)->reference) != 0) { fputs(": ", stderr); }`
	if strings.Count(renderer, guard) != 1 {
		t.Fatal("separator guard moved")
	}
	// Copy the real renderer into the program's translation unit under a private
	// symbol. Both control and mutant call it, with the ordinary sanitized runtime.
	build := func(renderer, name string) run {
		t.Helper()
		source := "#include \"adamic.h\"\n#include <stdio.h>\n#include <stdlib.h>\n" + renderer + "\n" + native.C(program)
		source = strings.ReplaceAll(source, "adamic_uncaught", "adamic_uncaught_separator_probe")
		binary := filepath.Join(t.TempDir(), name)
		if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		return executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"}, binary)
	}
	control := build(renderer, "control")
	if control.exitCode != 1 || len(control.stdout) != 0 || string(control.stderr) != "Uncaught Error\n" {
		t.Fatalf("copied renderer control: %+v", control)
	}
	mutant := build(strings.Replace(renderer, guard, `fputs(": ", stderr);`, 1), "mutant")
	if mutant.exitCode != 1 || len(mutant.stdout) != 0 || string(mutant.stderr) != "Uncaught Error: \n" {
		t.Fatalf("mutant must reach only the wrong diagnostic with sanitizers clean: %+v", mutant)
	}
	if string(mutant.stderr) == string(control.stderr) {
		t.Fatal("always-separator mutant survived")
	}
	t.Log("always-separator native mutant caught by exact stderr; ASan, UBSan and leak detection clean")
}

func TestUncaughtErrorSeparatorMutantJavaScript(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/step21_error_empty_name.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	control := onJavaScriptBackend(t, program)
	if control.exitCode != 1 || len(control.stdout) != 0 || string(control.stderr) != "Uncaught message\n" {
		t.Fatalf("JavaScript control: %+v", control)
	}
	source := javascript.JavaScript(program)
	const guard = `name === '' ? message : message === '' ? name : name + ': ' + message`
	if strings.Count(source, guard) != 1 {
		t.Fatal("JavaScript separator guard moved")
	}
	source = strings.Replace(source, guard, `name + ': ' + message`, 1)
	mutantPath := filepath.Join(t.TempDir(), "mutant.mjs")
	if err := os.WriteFile(mutantPath, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	mutant := onNode(t, mutantPath)
	if mutant.exitCode != 1 || len(mutant.stdout) != 0 || string(mutant.stderr) != "Uncaught : message\n" {
		t.Fatalf("mutant must reach only the wrong diagnostic: %+v", mutant)
	}
	if string(mutant.stderr) == string(control.stderr) {
		t.Fatal("always-separator mutant survived")
	}
	t.Log("always-separator JavaScript mutant caught by exact stderr")
}
