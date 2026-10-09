package oracle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Every way to sneak a cycle past the relaxation for fresh writes (docs/memory.md) is a probe that
// must stay refused, naming the write marked "closes the cycle". Each is a whole program Node runs,
// whose marked write really closes a cycle when it runs: if the relaxation ever lets one through,
// the test builds it and reports what LeakSanitizer says, which is how each mutant of the proof
// shows it was caught.
func TestFreshWriteProbesStayRefused(t *testing.T) {
	t.Parallel()
	probes, err := filepath.Glob(filepath.Join(repository, "internal", "oracle", "testdata", "fresh_refused", "*.a"))
	if err != nil {
		t.Fatal(err)
	}
	if len(probes) < 20 {
		t.Fatalf("found only %d probes", len(probes))
	}
	for _, probe := range probes {
		t.Run(filepath.Base(probe), func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(probe)
			if err != nil {
				t.Fatal(err)
			}
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			marked := 0
			for index, line := range strings.Split(string(source), "\n") {
				if strings.HasSuffix(line, "// closes the cycle") {
					marked = index + 1
				}
			}
			if marked == 0 {
				t.Fatal("no line is marked // closes the cycle")
			}
			program, err := lowered(t, path)
			var refused *lower.Refused
			if errors.As(err, &refused) {
				want := fmt.Sprintf("the write at %s:%d:", path, marked)
				if !strings.Contains(refused.Error(), want) || !strings.Contains(refused.Error(), "(adamic/cycle-capable)") {
					t.Errorf("refused, but not naming the marked write (%q):\n%v", want, refused)
				}
				return
			}
			if err != nil {
				t.Fatalf("want the cycle refused, got %v", err)
			}
			// Accepted: the hole is real when the program leaks the cycle it closed.
			oracle := onNode(t, path)
			native, sanitized := natively(t, program)
			report := "nothing"
			if leaked := leaks(t, program, sanitized); leaked != "" {
				report = leaked
			}
			t.Errorf("accepted a program that closes a cycle at line %d; Node exit %d %q, native exit %d %q; the leak check reports:\n%s",
				marked, oracle.exitCode, oracle.stdout, native.exitCode, native.stdout, report)
		})
	}
}
