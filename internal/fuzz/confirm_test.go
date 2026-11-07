package fuzz

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// All tools are fixtures: clang creates an executable script on success, so the
// test exercises Try's complete pipeline without building a real checkout.
func flakyCheckout(test *testing.T, stage string, always bool, silent bool) (*Checkout, string) {
	test.Helper()
	root := test.TempDir()
	tools := filepath.Join(root, "tools", "bin")
	if createDirectoryError := os.MkdirAll(tools, 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	test.Setenv("ADAMIC_TOOLS", filepath.Dir(tools))
	test.Setenv("FUZZ_FAIL_STAGE", stage)
	test.Setenv("FUZZ_FAIL_STATE", filepath.Join(root, "failed"))
	calls := filepath.Join(root, "calls")
	test.Setenv("FUZZ_FAIL_CALLS", calls)
	test.Setenv("FUZZ_FAIL_ALWAYS", fmt.Sprint(always))
	message := "printf 'panic: fixture failure under load\\n' >&2\n"
	if silent {
		message = ""
	}
	fail := func(stage, stop string) string {
		return fmt.Sprintf(`if [ "$FUZZ_FAIL_STAGE" = '%s' ]; then
printf 'attempt\n' >> "$FUZZ_FAIL_CALLS"
if [ "$FUZZ_FAIL_ALWAYS" = true ] || [ ! -e "$FUZZ_FAIL_STATE" ]; then
: > "$FUZZ_FAIL_STATE"
%s%s
fi
fi
`, stage, message, stop)
	}
	settings := fmt.Sprintf(
		"FUZZ_FAIL_STAGE=%q\nFUZZ_FAIL_STATE=%q\nFUZZ_FAIL_CALLS=%q\nFUZZ_FAIL_ALWAYS=%q\nPATH=%q\n",
		stage,
		filepath.Join(root, "failed"),
		calls,
		fmt.Sprint(always),
		tools+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	write := func(path, body string) {
		test.Helper()
		if writeError := os.WriteFile(path, []byte("#!/bin/sh\n"+settings+body), 0755); writeError != nil {
			test.Fatal(writeError)
		}
	}
	adamic := filepath.Join(tools, "adamic")
	write(adamic, "case \"$1\" in\nc)\n"+
		fail("c", "exit 2")+
		";;\njs)\n"+
		fail("js", "exit 2")+
		";;\nesac\nprintf 'fixture source\\n'\n")
	native := settings +
		"if [ \"$ASAN_OPTIONS\" = detect_leaks=1 ]; then\n" +
		fail("leak", "kill -TERM $$") +
		"else\n" +
		fail("native", "kill -TERM $$") +
		"fi\nprintf 'ok\\n'\n"
	write(filepath.Join(tools, "clang"), fail("clang", "exit 1")+`while [ "$#" -gt 0 ]; do
if [ "$1" = -o ]; then shift; binary=$1; break; fi
shift
done
cat > "$binary" <<'PROGRAM'
#!/bin/sh
`+native+"PROGRAM\nchmod +x \"$binary\"\n")
	write(filepath.Join(tools, "node"), `for last do :; done
case "$last" in
*.mjs)
`+fail("backend", "kill -TERM $$")+";;\n*)\n"+fail("node", "kill -TERM $$")+";;\nesac\nprintf 'ok\\n'\n")
	return &Checkout{Root: root, adamic: adamic, runtime: filepath.Join(root, "runtime.a")}, calls
}

// Not parallel: changes ADAMIC_TOOLS and FUZZ_FAIL_* process environment through flakyCheckout.
func TestFindingReproducesAlone(test *testing.T) {
	for _, testCase := range []struct {
		name, stage    string
		always, silent bool
		want           Verdict
	}{
		{"clang flakes", "clang", false, false, Flaked},
		{"clang always fails", "clang", true, false, Finding},
		{"silent clang death", "clang", false, true, Flaked},
		{"compiler crash", "c", false, false, Flaked},
		{"backend compiler crash", "js", false, false, Flaked},
		{"native death", "native", false, false, Flaked},
		{"node death", "node", false, false, Flaked},
		{"backend death", "backend", false, false, Flaked},
		{"leak runner death", "leak", false, false, Flaked},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			checkout, calls := flakyCheckout(test, testCase.stage, testCase.always, testCase.silent)
			outcome := checkout.Try("fixture", filepath.Join(checkout.Root, "program"))
			if outcome.Verdict != testCase.want {
				test.Fatalf("got %s %s: %s; want %s", outcome.Verdict, outcome.Key, outcome.Detail, testCase.want)
			}
			attempts, readError := os.ReadFile(calls)
			if readError != nil {
				test.Fatal(readError)
			}
			if count := strings.Count(string(attempts), "attempt\n"); count != 2 {
				test.Fatalf("got %d attempts, want exactly 2", count)
			}
			diagnostic := "fixture failure under load"
			if testCase.silent {
				diagnostic = "exit status 1"
			}
			if !strings.Contains(outcome.Detail, diagnostic) {
				test.Fatalf("first failure lost: %q", outcome.Detail)
			}
		})
	}
}

// Not parallel: asserts exclusive ordering through the process-wide programRuns confirmation lock.
func TestConfirmationWaitsForOtherPrograms(test *testing.T) {
	started, release, otherDone := make(chan struct{}), make(chan struct{}), make(chan struct{})
	go func() {
		confirm(func() Outcome { close(started); <-release; return Outcome{Verdict: Agreed} })
		close(otherDone)
	}()
	<-started
	first, exclusive := make(chan struct{}), make(chan struct{})
	result := make(chan Outcome, 1)
	go func() {
		attempts := 0
		result <- confirm(func() Outcome {
			attempts++
			if attempts == 1 {
				close(first)
				return Outcome{Verdict: Finding, Key: "clang refused the C", retry: true}
			}
			close(exclusive)
			return Outcome{Verdict: Agreed}
		})
	}()
	<-first
	enteredEarly := false
	select {
	case <-exclusive:
		enteredEarly = true
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	<-otherDone
	outcome := <-result
	if enteredEarly {
		test.Fatal("exclusive rerun overlapped another program")
	}
	if outcome.Verdict != Flaked {
		test.Fatalf("got %s, want flaked", outcome.Verdict)
	}
}

// Not parallel: changes ADAMIC_TOOLS and FUZZ_FAIL_* process environment through flakyCheckout.
func TestObserveConfirmsCompilerCrash(test *testing.T) {
	checkout, _ := flakyCheckout(test, "c", false, false)
	outcome := checkout.Observe("fixture", "program.a", filepath.Join(checkout.Root, "program"), "crash")
	if outcome.Verdict != Flaked {
		test.Fatalf("got %s, want flaked", outcome.Verdict)
	}
	if outcome.Has(Signature{Kind: "crash", Text: "fixture failure"}) {
		test.Fatal("transient crash kept a reducible signature")
	}
}

// Not parallel: inspects the process-wide confirmation lock.
func TestCompileFindingConfirmation(test *testing.T) {
	for _, testCase := range []struct {
		name, body, key string
		want            Verdict
		attempts        int
	}{
		{
			"refusal mismatch",
			"printf 'Adamic 0.1 refuses another path\\n' >&2; exit 1",
			"refusal did not name the path",
			Finding,
			1,
		},
		{
			"not yet mismatch",
			"printf \"can't lower fixture yet\\n\" >&2; exit 2",
			"expected a refusal, stage 0 said not yet",
			Finding,
			1,
		},
		{"panic", "printf 'panic: x\\n goroutine 1 [running]:\\n' >&2; exit 2", "compiler crashed", Flaked, 2},
		{"timeout", "kill -XCPU $$", "compiler never finished", Flaked, 2},
		{"death", "kill -TERM $$", "refusal did not name the path", Flaked, 2},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			directory := test.TempDir()
			compiler := filepath.Join(directory, "compiler")
			if writeError := os.WriteFile(compiler, []byte("#!/bin/sh\n"+testCase.body+"\n"), 0755); writeError != nil {
				test.Fatal(writeError)
			}
			path := filepath.Join(directory, "program.a")
			if writeError := os.WriteFile(
				path,
				[]byte("// parallel-refuse: expected path\n"),
				0644,
			); writeError != nil {
				test.Fatal(writeError)
			}
			checkout := &Checkout{adamic: compiler}
			attempts := 0
			outcome := confirm(func() Outcome {
				attempts++
				// TryRLock succeeds during the shared attempt and fails during the exclusive one.
				shared := programRuns.TryRLock()
				if shared {
					programRuns.RUnlock()
				}
				if shared != (attempts == 1) {
					test.Fatalf("attempt %d: shared lock available = %v", attempts, shared)
				}
				if attempts > 1 {
					return Outcome{Verdict: Agreed}
				}
				result := checkout.tryFile(path, directory)
				if result.Verdict != Finding || result.Key != testCase.key {
					test.Fatalf("got %+v, want finding", result)
				}
				return result
			})
			if outcome.Verdict != testCase.want || attempts != testCase.attempts {
				test.Fatalf(
					"got %s after %d attempts, want %s after %d",
					outcome.Verdict,
					attempts,
					testCase.want,
					testCase.attempts,
				)
			}
		})
	}
}
