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

func init() {
	for _, name := range []string{"qualified-live", "tracing-qualified"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"stage3/namespace-value/" + name + ".a", true, false})
	}
}

func TestNamespaceValueBoundary(t *testing.T) {
	for _, test := range []struct{ name, expression, stdout string }{
		{"stored", "const held = State;", "2\n7\n"},
		{"passed", "inspect(State)", "2\n"},
		{"returned", "return State;", "2\n7\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/namespace-value", test.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != test.stdout {
				t.Fatalf("source Node: %+v", truth)
			}
			program, err := lowered(t, path)
			var notYet *lower.NotYet
			if program != nil || !errors.As(err, &notYet) || !strings.Contains(notYet.What, "namespace object used as a value (State, declared at ") || !strings.Contains(notYet.What, "use State.member access, or pass fixed namespace functions and explicit state") {
				t.Fatalf("missing named capability boundary and fix: %v", err)
			}
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			before, after, found := strings.Cut(string(source), test.expression)
			if !found {
				t.Fatal("missing escape site")
			}
			// Replacing only the escaped value by current fields models the tempting snapshot implementation.
			snapshot := "{ value: State.value, advance: State.advance }"
			replacement := map[string]string{"stored": "const held = " + snapshot + ";", "passed": "inspect(" + snapshot + ")", "returned": "return " + snapshot + ";"}[test.name]
			mutantPath := filepath.Join(t.TempDir(), "snapshot.a")
			if err := os.WriteFile(mutantPath, []byte(before+replacement+after), 0600); err != nil {
				t.Fatal(err)
			}
			mutant, err := lowered(t, mutantPath)
			if err != nil {
				t.Fatal(err)
			}
			nodeMutant := onNode(t, mutantPath)
			js := onJavaScriptBackend(t, mutant)
			observed, binary := nativelyUncached(t, mutant)
			for backend, result := range map[string]run{"JavaScript": js, "native": observed} {
				if d := disagreement(nodeMutant, result); d != "" {
					t.Fatalf("mutant must match its own Node source, %s %s", backend, d)
				}
				if result.exitCode != 0 || len(result.stderr) != 0 || disagreement(truth, result) != "stdout differs" {
					t.Fatalf("snapshot must finish normally and lose the live binding, %s %+v", backend, result)
				}
			}
			if report := leaksUncached(t, mutant, binary); report != "" {
				t.Fatal(report)
			}
			// Pin the original expression position, not an unrelated later lowering stop.
			lines := strings.Split(string(source), "\n")
			line, column := 0, 0
			for i, text := range lines {
				if strings.Contains(text, test.expression) {
					line = i + 1
					column = strings.LastIndex(text, "State") + 1
					break
				}
			}
			wantSuffix := ":" + fmt.Sprint(line) + ":" + fmt.Sprint(column)
			if notYet.Where != path+wantSuffix {
				t.Fatalf("wrong refusal site: %s, want %s", notYet.Where, path+wantSuffix)
			}
			t.Logf("Node %q; refused %s; snapshot mutant %q caught in both backends", truth.stdout, notYet.Where, observed.stdout)
		})
	}
}

func TestTracingNamespaceValueBoundary(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/namespaces/10_tracing_escape.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != "1,3\nparse\n" {
		t.Fatalf("unchanged stock reduction Node: %+v", truth)
	}
	program, err := lowered(t, path)
	var notYet *lower.NotYet
	if program != nil || !errors.As(err, &notYet) || notYet.Where != path+":38:11" || !strings.Contains(notYet.What, "tracingEnabled") || !strings.Contains(notYet.What, "use tracingEnabled.member access") {
		t.Fatalf("original escape lost its located fix: %v", err)
	}
	t.Logf("source Node %q, capability stop %v", truth.stdout, err)
}
