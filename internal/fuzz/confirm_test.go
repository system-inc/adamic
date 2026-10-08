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
func flakyCheckout(t *testing.T, stage string, always bool, silent bool) (*Checkout, string) {
	t.Helper()
	root := t.TempDir()
	tools := filepath.Join(root, "tools")
	if err := os.Mkdir(tools, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", tools+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FUZZ_FAIL_STAGE", stage)
	t.Setenv("FUZZ_FAIL_STATE", filepath.Join(root, "failed"))
	calls := filepath.Join(root, "calls")
	t.Setenv("FUZZ_FAIL_CALLS", calls)
	t.Setenv("FUZZ_FAIL_ALWAYS", fmt.Sprint(always))
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
	write := func(path, body string) {
		t.Helper()
		if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0755); err != nil {
			t.Fatal(err)
		}
	}
	adamic := filepath.Join(tools, "adamic")
	write(adamic, "case \"$1\" in\nc)\n"+fail("c", "exit 2")+";;\njs)\n"+fail("js", "exit 2")+";;\nesac\nprintf 'fixture source\\n'\n")
	native := "if [ \"$ASAN_OPTIONS\" = detect_leaks=1 ]; then\n" + fail("leak", "kill -TERM $$") + "else\n" + fail("native", "kill -TERM $$") + "fi\nprintf 'ok\\n'\n"
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

func TestFindingReproducesAlone(t *testing.T) {
	for _, test := range []struct {
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
		t.Run(test.name, func(t *testing.T) {
			checkout, calls := flakyCheckout(t, test.stage, test.always, test.silent)
			outcome := checkout.Try("fixture", filepath.Join(checkout.Root, "program"))
			if outcome.Verdict != test.want {
				t.Fatalf("got %s %s: %s; want %s", outcome.Verdict, outcome.Key, outcome.Detail, test.want)
			}
			attempts, err := os.ReadFile(calls)
			if err != nil {
				t.Fatal(err)
			}
			if count := strings.Count(string(attempts), "attempt\n"); count != 2 {
				t.Fatalf("got %d attempts, want exactly 2", count)
			}
			diagnostic := "fixture failure under load"
			if test.silent {
				diagnostic = "exit status 1"
			}
			if !strings.Contains(outcome.Detail, diagnostic) {
				t.Fatalf("first failure lost: %q", outcome.Detail)
			}
		})
	}
}

func TestConfirmationWaitsForOtherPrograms(t *testing.T) {
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
		t.Fatal("exclusive rerun overlapped another program")
	}
	if outcome.Verdict != Flaked {
		t.Fatalf("got %s, want flaked", outcome.Verdict)
	}
}

func TestObserveConfirmsCompilerCrash(t *testing.T) {
	checkout, _ := flakyCheckout(t, "c", false, false)
	outcome := checkout.Observe("fixture", "program.a", filepath.Join(checkout.Root, "program"), "crash")
	if outcome.Verdict != Flaked {
		t.Fatalf("got %s, want flaked", outcome.Verdict)
	}
	if outcome.Has(Signature{Kind: "crash", Text: "fixture failure"}) {
		t.Fatal("transient crash kept a reducible signature")
	}
}
