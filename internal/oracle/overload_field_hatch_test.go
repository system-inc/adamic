package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	for _, name := range []string{"kind-valid", "value-valid", "value-undefined"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/overload_field_hatch/" + name + ".ts", true, false})
	}
}
func checkOverloadFieldHatch(t *testing.T, selectedName, selectedVariant string) {
	t.Helper()
	for _, field := range []string{"kind", "value"} {
		variants := []string{"valid", "liar"}
		if field == "value" {
			variants = append(variants, "undefined")
		}
		for _, variant := range variants {
			func() {
				if field != selectedName || variant != selectedVariant {
					return
				}
				path, pathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/overload_field_hatch", field+"-"+variant+".ts"))
				if pathError != nil {
					t.Fatal(pathError)
				}
				truth := onNode(t, path)
				if truth.exitCode != 0 || len(truth.stderr) != 0 {
					t.Fatalf("Node: %+v", truth)
				}
				program, err := lowered(t, path)
				if err != nil {
					t.Fatal(err)
				}
				wantChecks := 2
				if program.PredicateChecks.Checked != wantChecks || len(program.PredicateChecks.Sites) != wantChecks {
					t.Fatalf("checked sites: %+v", program.PredicateChecks)
				}
				for _, site := range program.PredicateChecks.Sites {
					if len(site.Directions) != 1 || site.Directions[0].Direction != "result."+field {
						t.Fatalf("site: %+v", site)
					}
				}
				native, sanitized := natively(t, program)
				for backend, actual := range map[string]run{"javascript": onJavaScriptBackend(t, program), "native": native, "release": released(t, program)} {
					if variant != "liar" {
						if difference := disagreement(truth, actual); difference != "" {
							t.Fatalf("%s: %s; actual=%+v truth=%+v", backend, difference, actual, truth)
						}
					} else {
						if actual.exitCode != 70 || len(actual.stdout) != 0 {
							t.Fatalf("%s unchecked wrong result: %+v", backend, actual)
						}
						for _, part := range []string{"overload 1 of ", " at ", ".ts:", "field result." + field, "implementation result ", "cannot serve overload result "} {
							if !strings.Contains(string(actual.stderr), part) {
								t.Fatalf("%s missing %q: %s", backend, part, actual.stderr)
							}
						}
					}
				}
				if variant != "liar" {
					if report := leaks(t, program, sanitized); report != "" {
						t.Fatal(report)
					}
				}
				refused, err := lowered(t, strings.TrimSuffix(path, ".ts")+".a")
				var refusal *lower.Refused
				if refused != nil || !errors.As(err, &refusal) || !strings.Contains(refusal.What, "result."+field) {
					t.Fatalf("Adamic must retain field proof refusal: %v", err)
				}
			}()
		}
	}
}

func TestOverloadFieldHatchKindValid(t *testing.T) {
	t.Parallel()
	checkOverloadFieldHatch(t, "kind", "valid")
}

func TestOverloadFieldHatchKindLiar(t *testing.T) {
	t.Parallel()
	checkOverloadFieldHatch(t, "kind", "liar")
}

func TestOverloadFieldHatchValueValid(t *testing.T) {
	t.Parallel()
	checkOverloadFieldHatch(t, "value", "valid")
}

func TestOverloadFieldHatchValueLiar(t *testing.T) {
	t.Parallel()
	checkOverloadFieldHatch(t, "value", "liar")
}

func TestOverloadFieldHatchValueUndefined(t *testing.T) {
	t.Parallel()
	checkOverloadFieldHatch(t, "value", "undefined")
}
