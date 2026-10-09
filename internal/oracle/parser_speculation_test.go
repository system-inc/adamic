package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"strings"
	"testing"
)

var parserSpeculationNames = []string{"named-lookahead", "named-tryparse", "arrow-tryparse", "truthiness-rewind", "misfit"}

func init() {
	for _, name := range parserSpeculationNames {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"stage3/parser-next/speculation/" + name + ".a", true, name == "misfit"})
	}
}
func parserSpeculationPath(t *testing.T, name string) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join(repository, "stage3/parser-next/speculation", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestParserSpeculationSources(t *testing.T) {
	for _, name := range parserSpeculationNames {
		if name == "misfit" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			path := parserSpeculationPath(t, name)
			truth := onNode(t, path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("Node %+v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			actual, binary := nativelyUncached(t, program)
			for backend, got := range map[string]run{"native": actual, "JavaScript": onJavaScriptBackend(t, program)} {
				if d := disagreement(truth, got); d != "" {
					t.Fatalf("%s: %s", backend, d)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Logf("Node stdout %q; native sanitizers/leaks and JavaScript agree", truth.stdout)
		})
	}
}
func TestParserSpeculationMisfit(t *testing.T) {
	path := parserSpeculationPath(t, "misfit")
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "wrong\n" {
		t.Fatalf("Node %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	message := "adamic: panic: speculative result failed: callback misfit at " + path + ":3:14; expected number\n"
	actual, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": actual, "JavaScript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != message {
			t.Errorf("%s: %+v, want %q", backend, got, message)
		}
	}
}
func TestParserSpeculationRewindMutant(t *testing.T) {
	path := parserSpeculationPath(t, "truthiness-rewind")
	truth := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	for i := range program.Functions {
		if program.Functions[i].Name != "lookAhead" {
			continue
		}
		for j, statement := range program.Functions[i].Body {
			if assign, ok := statement.(ir.Assign); ok && program.Locals[assign.Local].Name == "position" {
				assign.Value = ir.Read{Local: assign.Local, Of: program.Locals[assign.Local].Type}
				program.Functions[i].Body[j] = assign
				changed++
			}
		}
	}
	if changed != 1 {
		t.Fatalf("rewind sites %d", changed)
	}
	actual, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": actual, "JavaScript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 || disagreement(truth, got) == "" || !strings.Contains(string(got.stdout), "99|7|3|false") {
			t.Fatalf("%s: rewind mutant not executed/caught %+v", backend, got)
		}
	}
	t.Log("skipped scanner-position rewind caught by Node stdout in both backends")
}
