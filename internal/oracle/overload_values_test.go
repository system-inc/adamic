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
	}{"internal/oracle/testdata/overload_values/proven.a", true, false})
	for _, name := range []string{"returned", "narrow", "order", "module"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/overload_values/" + name + "-valid.ts", true, false})
	}
}

func TestOverloadValues(t *testing.T) {
	for _, name := range []string{"returned", "narrow", "order", "module"} {
		for _, variant := range []string{"valid", "liar"} {
			t.Run(name+"/"+variant, func(t *testing.T) {
				path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/overload_values", name+"-"+variant+".ts"))
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
				wantChecks := map[string]int{"returned": 2, "narrow": 2, "order": 1, "module": 1}[name]
				if program.PredicateChecks.Checked != wantChecks || len(program.PredicateChecks.Sites) != wantChecks {
					t.Fatalf("checks: %+v", program.PredicateChecks)
				}
				native, sanitized := natively(t, program)
				for backend, actual := range map[string]run{"javascript": onJavaScriptBackend(t, program), "native": native, "release": released(t, program)} {
					if variant == "valid" {
						if difference := disagreement(truth, actual); difference != "" {
							t.Fatalf("%s: %s; actual=%+v truth=%+v", backend, difference, actual, truth)
						}
					} else {
						if actual.exitCode != 70 {
							t.Fatalf("%s unchecked wrong result: %+v", backend, actual)
						}
						for _, part := range []string{"overload 1 of evaluate", " at ", ".ts:", "field result.value", "implementation result ", "cannot serve overload result "} {
							if !strings.Contains(string(actual.stderr), part) {
								t.Fatalf("%s missing %q: %s", backend, part, actual.stderr)
							}
						}
					}
				}
				if variant == "valid" {
					if report := leaks(t, program, sanitized); report != "" {
						t.Fatal(report)
					}
				}
				_, err = lowered(t, strings.TrimSuffix(path, ".ts")+".a")
				var refused *lower.Refused
				if !errors.As(err, &refused) || !strings.Contains(refused.What, "result.value") {
					t.Fatalf(".a needs result proof: %v", err)
				}
			})
		}
	}
}

func TestOverloadValueProof(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/overload_values/proven.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if program.PredicateChecks.Checked != 0 {
		t.Fatalf("proven result retained a check: %+v", program.PredicateChecks)
	}
	native, sanitized := natively(t, program)
	for backend, actual := range map[string]run{"javascript": onJavaScriptBackend(t, program), "native": native, "release": released(t, program)} {
		if difference := disagreement(truth, actual); difference != "" {
			t.Fatalf("%s: %s", backend, difference)
		}
	}
	if report := leaks(t, program, sanitized); report != "" {
		t.Fatal(report)
	}
}
