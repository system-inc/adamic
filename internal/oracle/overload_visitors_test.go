package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/overload_visitors/instantiations.ts", true, false})
	for _, name := range []string{"original", "helper", "overloaded-helper"} {
		for _, extension := range []string{".a", ".ts"} {
			fixtures = append(fixtures, struct {
				path            string
				lowers, checked bool
			}{"internal/oracle/testdata/overload_visitors/" + name + extension, true, false})
		}
	}
}

func TestOverloadVisitors(t *testing.T) {
	for _, name := range []string{"original", "helper", "overloaded-helper", "default"} {
		variants := []string{"valid", "liar"}
		if name == "default" {
			variants = []string{"liar"}
		}
		for _, variant := range variants {
			t.Run(name+"/"+variant, func(t *testing.T) {
				suffix := ""
				if variant == "liar" {
					suffix = "-liar"
				}
				path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/overload_visitors", name+suffix+".ts"))
				if err != nil {
					t.Fatal(err)
				}
				truth := onNode(t, path)
				if truth.exitCode != 0 || len(truth.stderr) != 0 {
					t.Fatalf("Node: %+v", truth)
				}
				program, err := lowered(t, path)
				if err != nil {
					t.Fatal(err)
				}
				native, sanitized := natively(t, program)
				observations := map[string]run{"javascript": onJavaScriptBackend(t, program), "native": native, "release": released(t, program)}
				for _, backend := range []string{"javascript", "native", "release"} {
					actual := observations[backend]
					if variant == "valid" {
						if difference := disagreement(truth, actual); difference != "" {
							t.Fatalf("%s: %s; actual=%+v truth=%+v", backend, difference, actual, truth)
						}
					} else {
						if actual.exitCode != 70 {
							t.Fatalf("%s unproven visitor input admitted: %+v", backend, actual)
						}
						for _, part := range []string{"visitor visitor at ", ".ts:", "argument Base cannot serve Named"} {
							if !strings.Contains(string(actual.stderr), part) {
								t.Fatalf("%s missing %q: %s", backend, part, actual.stderr)
							}
						}
					}
				}
				want := 0
				if variant == "liar" {
					want = 1
				}
				if program.PredicateChecks.Checked != want {
					t.Fatalf("checked invocations: %+v", program.PredicateChecks)
				}
				if variant == "valid" {
					if report := leaks(t, program, sanitized); report != "" {
						t.Fatal(report)
					}
					adam, err := lowered(t, strings.TrimSuffix(path, ".ts")+".a")
					if err != nil {
						t.Fatal(err)
					}
					actual, safe := natively(t, adam)
					if difference := disagreement(truth, actual); difference != "" {
						t.Fatal(difference)
					}
					if difference := disagreement(truth, onJavaScriptBackend(t, adam)); difference != "" {
						t.Fatal(difference)
					}
					if report := leaks(t, adam, safe); report != "" {
						t.Fatal(report)
					}
				} else {
					_, err := lowered(t, strings.TrimSuffix(path, ".ts")+".a")
					var refused *lower.Refused
					if !errors.As(err, &refused) || !strings.Contains(refused.What, "visitor argument path is unproven") {
						t.Fatalf(".a unproven invocation must refuse: %v", err)
					}
				}
			})
		}
	}
}

func TestOverloadVisitorInstantiations(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/overload_visitors/instantiations.ts"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	native, sanitized := natively(t, program)
	for backend, actual := range map[string]run{"javascript": onJavaScriptBackend(t, program), "native": native, "release": released(t, program)} {
		if difference := disagreement(truth, actual); difference != "" {
			t.Fatalf("%s: %s", backend, difference)
		}
	}
	if program.PredicateChecks.Checked != 2 || len(program.PredicateChecks.Sites) != 2 {
		t.Fatalf("concrete invocation records: %+v", program.PredicateChecks)
	}
	reasons := program.PredicateChecks.Sites[0].Directions[0].Reason + program.PredicateChecks.Sites[1].Directions[0].Reason
	if !strings.Contains(reasons, "Named") || !strings.Contains(reasons, "Numbered") {
		t.Fatal(reasons)
	}
	if report := leaks(t, program, sanitized); report != "" {
		t.Fatal(report)
	}
	_, err = lowered(t, strings.TrimSuffix(path, ".ts")+".a")
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "visitor argument path is unproven") {
		t.Fatalf(".a must retain input proof: %v", err)
	}
}
