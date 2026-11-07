package native

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	goruntime "runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

// A cut falls only where no name a statement declares is used past it, and never takes a statement
// that reads main's parameters out of main.
func TestMainPartsCutBetweenStatements(test *testing.T) {
	test.Parallel()
	statements := []string{
		"\tdouble adamic_temporary_1 = 1;\n",
		"\tadamic_use(adamic_temporary_1);\n",
		"\tadamic_use(2);\n",
		"\tadamic_temporary_2:;\n",
		"\tgoto adamic_temporary_2;\n",
		"\tdouble held = 3;\n",
		"\tif (true) {\n\t\tdouble inner = 4;\n\t}\n",
		"\tadamic_use(held);\n",
		"\tif (true) {\n\t\tdouble inner = 5;\n\t}\n",
		"\tif (adamic_thrown != NULL) {\n\t\tgoto adamic_done_9;\n\t}\n",
		"\tadamic_use(6);\n",
		"\tadamic_done_9:;\n",
		"",
	}
	got := mainParts(statements, 1)
	want := []string{
		statements[0] + statements[1],
		statements[2],
		statements[3] + statements[4],
		statements[5] + statements[6] + statements[7],
		statements[8],
		// A goto ahead of its label keeps everything between them with it.
		statements[9] + statements[10] + statements[11],
	}
	if !reflect.DeepEqual(got, want) {
		test.Fatalf("parts:\n%q\nwant:\n%q", got, want)
	}
	if got := mainParts(statements, 100); len(got) != 1 || got[0] != strings.Join(statements, "") {
		test.Fatalf("under the limit main is one part: %q", got)
	}
	parameters := append([]string{"\tadamic_use(argc);\n"}, statements...)
	if got := mainParts(parameters, 1); len(got) != 1 {
		test.Fatalf("a statement reading argc was cut away from main: %q", got)
	}
}

// Every top-level statement its own part runs exactly as one main does: module order, an import
// cycle, a top-level throw and a read before initialization, under the sanitizers. Running the parts
// out of order changes what a program that prints from more than one of them prints, so the
// comparison sees order.
func TestMainPartsRunAsMain(test *testing.T) {
	test.Parallel()
	for _, fixture := range []struct {
		path    string
		ordered bool
	}{
		{"../oracle/testdata/modules/main.a", true},
		{"../oracle/testdata/import_cycles/order/main.a", true},
		{"../oracle/testdata/import_cycles/classes/a.a", true},
		{"../oracle/testdata/import_cycles/read/a.a", false},
		{"../oracle/testdata/exceptions_uncaught.a", true},
		{"../oracle/testdata/adversarial_order.a", true},
	} {
		test.Run(filepath.Base(filepath.Dir(fixture.path))+"/"+filepath.Base(fixture.path), func(test *testing.T) {
			test.Parallel()
			program := loopProgram(test, fixture.path)
			whole, parted := cProgram(program, -1, mainLines), cProgram(program, -1, 1)
			if strings.Contains(whole, "static void adamic_main_") {
				test.Fatal("a fixture under mainLines was cut")
			}
			if len(regexp.MustCompile(`(?m)^\tadamic_main_[A-Za-z0-9_]+\(\);\n`).FindAllString(parted, -1)) < 2 {
				test.Fatalf("not cut into parts:\n%s", parted)
			}
			want := runMainParts(test, whole)
			if got := runMainParts(test, parted); !reflect.DeepEqual(got, want) {
				test.Fatalf("parts ran %+v, main ran %+v", got, want)
			}
			if !fixture.ordered {
				return
			}
			// The mutant: main calls its parts last to first.
			calls := regexp.MustCompile(`(?m)^\tadamic_main_[A-Za-z0-9_]+\(\);\n`).FindAllString(parted, -1)
			reversed := slices.Clone(calls)
			slices.Reverse(reversed)
			if strings.Count(parted, strings.Join(calls, "")) != 1 {
				test.Fatal("the reversing mutant lost its anchor")
			}
			mutant := strings.Replace(parted, strings.Join(calls, ""), strings.Join(reversed, ""), 1)
			if got := runMainParts(test, mutant); reflect.DeepEqual(got, want) {
				test.Fatalf("parts run out of order still ran %+v", got)
			}
		})
	}
}

type mainPartsRun struct {
	Stdout, Stderr string
	Exit           int
}

func runMainParts(test *testing.T, source string) mainPartsRun {
	test.Helper()
	binary := filepath.Join(test.TempDir(), "program")
	if buildError := Build(source, binary, Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := programCommand(timeoutContext, "", binary)
	if goruntime.GOOS == "linux" {
		// A program an uncaught error ends lets nothing go, as any panic; main's leaks are the oracle's.
		command.Env = append(command.Environ(), "ASAN_OPTIONS=detect_leaks=0")
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	runError := command.Run()
	var exit *exec.ExitError
	if runError != nil && !errors.As(runError, &exit) {
		test.Fatal(runError)
	}
	return mainPartsRun{stdout.String(), stderr.String(), command.ProcessState.ExitCode()}
}
