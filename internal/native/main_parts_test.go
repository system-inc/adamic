package native

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"reflect"
	goruntime "runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

// A cut falls only where no name a statement declares is used past it, and never takes a statement
// that reads main's parameters out of main.
func TestMainPartsCutBetweenStatements(t *testing.T) {
	t.Parallel()
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
		t.Fatalf("parts:\n%q\nwant:\n%q", got, want)
	}
	if got := mainParts(statements, 100); len(got) != 1 || got[0] != strings.Join(statements, "") {
		t.Fatalf("under the limit main is one part: %q", got)
	}
	parameters := append([]string{"\tadamic_use(argc);\n"}, statements...)
	if got := mainParts(parameters, 1); len(got) != 1 {
		t.Fatalf("a statement reading argc was cut away from main: %q", got)
	}
}

// Every top-level statement its own part runs exactly as one main does: module order, an import
// cycle, a top-level throw and a read before initialization, under the sanitizers. Running the parts
// out of order changes what a program that prints from more than one of them prints, so the
// comparison sees order.
func TestMainPartsRunAsMain(t *testing.T) {
	t.Parallel()
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
		t.Run(filepath.Base(filepath.Dir(fixture.path))+"/"+filepath.Base(fixture.path), func(t *testing.T) {
			t.Parallel()
			program := loopProgram(t, fixture.path)
			whole, parted := cProgram(program, -1, mainLines), cProgram(program, -1, 1)
			if strings.Contains(whole, "adamic_main_0") {
				t.Fatal("a fixture under mainLines was cut")
			}
			if !strings.Contains(parted, "\tadamic_main_1();\n") {
				t.Fatalf("not cut into parts:\n%s", parted)
			}
			want := runMainParts(t, whole)
			if got := runMainParts(t, parted); !reflect.DeepEqual(got, want) {
				t.Fatalf("parts ran %+v, main ran %+v", got, want)
			}
			if !fixture.ordered {
				return
			}
			// The mutant: main calls its parts last to first.
			calls := []string{}
			for index := 0; strings.Contains(parted, fmt.Sprintf("\tadamic_main_%d();\n", index)); index++ {
				calls = append(calls, fmt.Sprintf("\tadamic_main_%d();\n", index))
			}
			reversed := slices.Clone(calls)
			slices.Reverse(reversed)
			if strings.Count(parted, strings.Join(calls, "")) != 1 {
				t.Fatal("the reversing mutant lost its anchor")
			}
			mutant := strings.Replace(parted, strings.Join(calls, ""), strings.Join(reversed, ""), 1)
			if got := runMainParts(t, mutant); reflect.DeepEqual(got, want) {
				t.Fatalf("parts run out of order still ran %+v", got)
			}
		})
	}
}

type mainPartsRun struct {
	Stdout, Stderr string
	Exit           int
}

func runMainParts(t *testing.T, source string) mainPartsRun {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "program")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := programCommand(ctx, "", binary)
	if goruntime.GOOS == "linux" {
		// A program an uncaught error ends lets nothing go, as any panic; main's leaks are the oracle's.
		command.Env = append(command.Environ(), "ASAN_OPTIONS=detect_leaks=0")
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	var exit *exec.ExitError
	if err != nil && !errors.As(err, &exit) {
		t.Fatal(err)
	}
	return mainPartsRun{stdout.String(), stderr.String(), command.ProcessState.ExitCode()}
}
